package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type (
	VaultFolder struct {
		ID        uuid.UUID
		ParentID  *uuid.UUID
		Name      string
		Locked    bool
		CreatedBy uuid.UUID
		CreatedAt time.Time
	}

	VaultFile struct {
		ID           uuid.UUID
		FolderID     *uuid.UUID
		OriginalName string
		StoredName   string
		Mime         string
		Size         int64
		Locked       bool
		UploadedBy   uuid.UUID
		CreatedAt    time.Time
	}

	VaultRepository interface {
		CreateFolder(ctx context.Context, f *VaultFolder) error
		GetFolder(ctx context.Context, id uuid.UUID) (*VaultFolder, error)
		ListChildFolders(ctx context.Context, parentID *uuid.UUID, includeLocked bool) ([]VaultFolder, error)
		RenameFolder(ctx context.Context, id uuid.UUID, name string) error
		SetFolderLocked(ctx context.Context, id uuid.UUID, locked bool) error
		DeleteFolder(ctx context.Context, id uuid.UUID) error
		FolderChain(ctx context.Context, id uuid.UUID) ([]VaultFolder, error)
		ChainLocked(ctx context.Context, id uuid.UUID) (bool, error)
		DescendantStoredNames(ctx context.Context, id uuid.UUID) ([]string, error)

		CreateFile(ctx context.Context, f *VaultFile) error
		GetFile(ctx context.Context, id uuid.UUID) (*VaultFile, error)
		ListFolderFiles(ctx context.Context, folderID *uuid.UUID, includeLocked bool) ([]VaultFile, error)
		RenameFile(ctx context.Context, id uuid.UUID, name string) error
		SetFileLocked(ctx context.Context, id uuid.UUID, locked bool) error
		DeleteFile(ctx context.Context, id uuid.UUID) error
	}

	vaultRepository struct {
		db *sql.DB
	}
)

func nullableUUIDValue(p *uuid.UUID) any {
	if p == nil {
		return nil
	}
	return *p
}

func scanVaultFolder(s interface{ Scan(...any) error }) (VaultFolder, error) {
	var f VaultFolder
	var parent uuid.NullUUID

	if err := s.Scan(&f.ID, &parent, &f.Name, &f.Locked, &f.CreatedBy, &f.CreatedAt); err != nil {
		return VaultFolder{}, err
	}

	if parent.Valid {
		p := parent.UUID
		f.ParentID = &p
	}

	return f, nil
}

func scanVaultFile(s interface{ Scan(...any) error }) (VaultFile, error) {
	var f VaultFile
	var folder uuid.NullUUID

	if err := s.Scan(&f.ID, &folder, &f.OriginalName, &f.StoredName, &f.Mime, &f.Size, &f.Locked, &f.UploadedBy, &f.CreatedAt); err != nil {
		return VaultFile{}, err
	}

	if folder.Valid {
		fid := folder.UUID
		f.FolderID = &fid
	}

	return f, nil
}

func (r *vaultRepository) CreateFolder(ctx context.Context, f *VaultFolder) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO vault_folders (id, parent_id, name, created_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING locked, created_at`,
		f.ID, nullableUUIDValue(f.ParentID), f.Name, f.CreatedBy,
	).Scan(&f.Locked, &f.CreatedAt)
	if err != nil {
		return fmt.Errorf("create folder: %w", err)
	}

	return nil
}

func (r *vaultRepository) GetFolder(ctx context.Context, id uuid.UUID) (*VaultFolder, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, parent_id, name, locked, created_by, created_at FROM vault_folders WHERE id = $1`, id,
	)

	f, err := scanVaultFolder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get folder: %w", err)
	}

	return &f, nil
}

func (r *vaultRepository) ListChildFolders(ctx context.Context, parentID *uuid.UUID, includeLocked bool) ([]VaultFolder, error) {
	query := `SELECT id, parent_id, name, locked, created_by, created_at FROM vault_folders WHERE parent_id IS NOT DISTINCT FROM $1`
	if !includeLocked {
		query += ` AND locked = false`
	}
	query += ` ORDER BY lower(name)`

	rows, err := r.db.QueryContext(ctx, query, nullableUUIDValue(parentID))
	if err != nil {
		return nil, fmt.Errorf("list child folders: %w", err)
	}
	defer rows.Close()

	var out []VaultFolder
	for rows.Next() {
		f, err := scanVaultFolder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		out = append(out, f)
	}

	return out, rows.Err()
}

func (r *vaultRepository) RenameFolder(ctx context.Context, id uuid.UUID, name string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vault_folders SET name = $2 WHERE id = $1`, id, name)
	if err != nil {
		return fmt.Errorf("rename folder: %w", err)
	}

	return nil
}

func (r *vaultRepository) SetFolderLocked(ctx context.Context, id uuid.UUID, locked bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vault_folders SET locked = $2 WHERE id = $1`, id, locked)
	if err != nil {
		return fmt.Errorf("set folder locked: %w", err)
	}

	return nil
}

func (r *vaultRepository) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM vault_folders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}

	return nil
}

func (r *vaultRepository) FolderChain(ctx context.Context, id uuid.UUID) ([]VaultFolder, error) {
	rows, err := r.db.QueryContext(ctx,
		`WITH RECURSIVE chain AS (
			SELECT id, parent_id, name, locked, created_by, created_at, 0 AS depth
			FROM vault_folders WHERE id = $1
			UNION ALL
			SELECT f.id, f.parent_id, f.name, f.locked, f.created_by, f.created_at, c.depth + 1
			FROM vault_folders f JOIN chain c ON f.id = c.parent_id
		)
		SELECT id, parent_id, name, locked, created_by, created_at FROM chain ORDER BY depth DESC`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("folder chain: %w", err)
	}
	defer rows.Close()

	var out []VaultFolder
	for rows.Next() {
		f, err := scanVaultFolder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan chain folder: %w", err)
		}
		out = append(out, f)
	}

	return out, rows.Err()
}

func (r *vaultRepository) ChainLocked(ctx context.Context, id uuid.UUID) (bool, error) {
	var locked bool
	err := r.db.QueryRowContext(ctx,
		`WITH RECURSIVE chain AS (
			SELECT id, parent_id, locked FROM vault_folders WHERE id = $1
			UNION ALL
			SELECT f.id, f.parent_id, f.locked FROM vault_folders f JOIN chain c ON f.id = c.parent_id
		)
		SELECT COALESCE(bool_or(locked), false) FROM chain`, id,
	).Scan(&locked)
	if err != nil {
		return false, fmt.Errorf("chain locked: %w", err)
	}

	return locked, nil
}

func (r *vaultRepository) DescendantStoredNames(ctx context.Context, id uuid.UUID) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`WITH RECURSIVE subtree AS (
			SELECT id FROM vault_folders WHERE id = $1
			UNION ALL
			SELECT f.id FROM vault_folders f JOIN subtree s ON f.parent_id = s.id
		)
		SELECT stored_name FROM vault_files WHERE folder_id IN (SELECT id FROM subtree)`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("descendant stored names: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan stored name: %w", err)
		}
		out = append(out, name)
	}

	return out, rows.Err()
}

func (r *vaultRepository) CreateFile(ctx context.Context, f *VaultFile) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO vault_files (id, folder_id, original_name, stored_name, mime, size, uploaded_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING locked, created_at`,
		f.ID, nullableUUIDValue(f.FolderID), f.OriginalName, f.StoredName, f.Mime, f.Size, f.UploadedBy,
	).Scan(&f.Locked, &f.CreatedAt)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	return nil
}

func (r *vaultRepository) GetFile(ctx context.Context, id uuid.UUID) (*VaultFile, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, folder_id, original_name, stored_name, mime, size, locked, uploaded_by, created_at
		 FROM vault_files WHERE id = $1`, id,
	)

	f, err := scanVaultFile(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	return &f, nil
}

func (r *vaultRepository) ListFolderFiles(ctx context.Context, folderID *uuid.UUID, includeLocked bool) ([]VaultFile, error) {
	query := `SELECT id, folder_id, original_name, stored_name, mime, size, locked, uploaded_by, created_at
		 FROM vault_files WHERE folder_id IS NOT DISTINCT FROM $1`
	if !includeLocked {
		query += ` AND locked = false`
	}
	query += ` ORDER BY lower(original_name)`

	rows, err := r.db.QueryContext(ctx, query, nullableUUIDValue(folderID))
	if err != nil {
		return nil, fmt.Errorf("list folder files: %w", err)
	}
	defer rows.Close()

	var out []VaultFile
	for rows.Next() {
		f, err := scanVaultFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		out = append(out, f)
	}

	return out, rows.Err()
}

func (r *vaultRepository) RenameFile(ctx context.Context, id uuid.UUID, name string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vault_files SET original_name = $2 WHERE id = $1`, id, name)
	if err != nil {
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}

func (r *vaultRepository) SetFileLocked(ctx context.Context, id uuid.UUID, locked bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vault_files SET locked = $2 WHERE id = $1`, id, locked)
	if err != nil {
		return fmt.Errorf("set file locked: %w", err)
	}

	return nil
}

func (r *vaultRepository) DeleteFile(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM vault_files WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}
