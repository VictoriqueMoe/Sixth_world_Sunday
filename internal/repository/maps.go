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
	Map struct {
		ID          uuid.UUID
		Title       string
		Description string
		SourceURL   string
		Mid         string
		LL          string
		Zoom        string
		CreatedBy   uuid.UUID
		CreatedAt   time.Time
	}

	MapRepository interface {
		List(ctx context.Context) ([]Map, error)
		Get(ctx context.Context, id uuid.UUID) (*Map, error)
		Create(ctx context.Context, m *Map) error
		Update(ctx context.Context, m *Map) error
		Delete(ctx context.Context, id uuid.UUID) error
	}

	mapRepository struct {
		db *sql.DB
	}
)

const mapColumns = `id, title, description, source_url, mid, ll, zoom, created_by, created_at`

func scanMap(s interface{ Scan(...any) error }) (Map, error) {
	var m Map
	if err := s.Scan(&m.ID, &m.Title, &m.Description, &m.SourceURL, &m.Mid, &m.LL, &m.Zoom, &m.CreatedBy, &m.CreatedAt); err != nil {
		return Map{}, err
	}

	return m, nil
}

func (r *mapRepository) List(ctx context.Context) ([]Map, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+mapColumns+` FROM maps ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list maps: %w", err)
	}
	defer rows.Close()

	var out []Map
	for rows.Next() {
		m, err := scanMap(rows)
		if err != nil {
			return nil, fmt.Errorf("scan map: %w", err)
		}
		out = append(out, m)
	}

	return out, rows.Err()
}

func (r *mapRepository) Get(ctx context.Context, id uuid.UUID) (*Map, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+mapColumns+` FROM maps WHERE id = $1`, id)

	m, err := scanMap(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get map: %w", err)
	}

	return &m, nil
}

func (r *mapRepository) Create(ctx context.Context, m *Map) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO maps (id, title, description, source_url, mid, ll, zoom, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING created_at`,
		m.ID, m.Title, m.Description, m.SourceURL, m.Mid, m.LL, m.Zoom, m.CreatedBy,
	).Scan(&m.CreatedAt)
	if err != nil {
		return fmt.Errorf("create map: %w", err)
	}

	return nil
}

func (r *mapRepository) Update(ctx context.Context, m *Map) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE maps SET title = $2, description = $3, source_url = $4, mid = $5, ll = $6, zoom = $7 WHERE id = $1`,
		m.ID, m.Title, m.Description, m.SourceURL, m.Mid, m.LL, m.Zoom,
	)
	if err != nil {
		return fmt.Errorf("update map: %w", err)
	}

	return nil
}

func (r *mapRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM maps WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete map: %w", err)
	}

	return nil
}
