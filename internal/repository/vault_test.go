package repository_test

import (
	"context"
	"testing"

	"Sixth_world_Sunday/internal/repository"
	"Sixth_world_Sunday/internal/repository/repotest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultRepository_CreateAndGetFolder(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	folder := &repository.VaultFolder{ID: uuid.New(), Name: "root", CreatedBy: user.ID}

	// when
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), folder))
	got, err := repos.Vault.GetFolder(context.Background(), folder.ID)

	// then
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "root", got.Name)
	assert.Nil(t, got.ParentID)
	assert.False(t, got.Locked)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestVaultRepository_GetFolder_Missing(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)

	// when
	got, err := repos.Vault.GetFolder(context.Background(), uuid.New())

	// then
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestVaultRepository_ChainLockedCascadesToDescendants(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	root := &repository.VaultFolder{ID: uuid.New(), Name: "root", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), root))
	child := &repository.VaultFolder{ID: uuid.New(), ParentID: &root.ID, Name: "child", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), child))

	// when
	lockedBefore, err := repos.Vault.ChainLocked(context.Background(), child.ID)
	require.NoError(t, err)
	require.NoError(t, repos.Vault.SetFolderLocked(context.Background(), root.ID, true))
	lockedAfter, err := repos.Vault.ChainLocked(context.Background(), child.ID)
	require.NoError(t, err)

	// then
	assert.False(t, lockedBefore)
	assert.True(t, lockedAfter, "locking the root must cascade to the child via the ancestor chain")
}

func TestVaultRepository_FolderChain_RootFirst(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	root := &repository.VaultFolder{ID: uuid.New(), Name: "root", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), root))
	child := &repository.VaultFolder{ID: uuid.New(), ParentID: &root.ID, Name: "child", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), child))

	// when
	chain, err := repos.Vault.FolderChain(context.Background(), child.ID)

	// then
	require.NoError(t, err)
	require.Len(t, chain, 2)
	assert.Equal(t, root.ID, chain[0].ID)
	assert.Equal(t, child.ID, chain[1].ID)
}

func TestVaultRepository_ListChildFolders_FiltersLocked(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	parent := &repository.VaultFolder{ID: uuid.New(), Name: "parent", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), parent))
	open := &repository.VaultFolder{ID: uuid.New(), ParentID: &parent.ID, Name: "open", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), open))
	secret := &repository.VaultFolder{ID: uuid.New(), ParentID: &parent.ID, Name: "secret", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), secret))
	require.NoError(t, repos.Vault.SetFolderLocked(context.Background(), secret.ID, true))

	// when
	visible, err := repos.Vault.ListChildFolders(context.Background(), &parent.ID, false)
	require.NoError(t, err)
	all, err := repos.Vault.ListChildFolders(context.Background(), &parent.ID, true)
	require.NoError(t, err)

	// then
	require.Len(t, visible, 1)
	assert.Equal(t, open.ID, visible[0].ID)
	assert.Len(t, all, 2)
}

func TestVaultRepository_CreateGetAndListFiles(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	folder := &repository.VaultFolder{ID: uuid.New(), Name: "docs", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), folder))
	open := &repository.VaultFile{ID: uuid.New(), FolderID: &folder.ID, OriginalName: "open.txt", StoredName: uuid.NewString() + ".txt", Mime: "text/plain", Size: 5, UploadedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFile(context.Background(), open))
	secret := &repository.VaultFile{ID: uuid.New(), FolderID: &folder.ID, OriginalName: "secret.txt", StoredName: uuid.NewString() + ".txt", Mime: "text/plain", Size: 7, UploadedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFile(context.Background(), secret))
	require.NoError(t, repos.Vault.SetFileLocked(context.Background(), secret.ID, true))

	// when
	got, err := repos.Vault.GetFile(context.Background(), open.ID)
	require.NoError(t, err)
	visible, err := repos.Vault.ListFolderFiles(context.Background(), &folder.ID, false)
	require.NoError(t, err)
	all, err := repos.Vault.ListFolderFiles(context.Background(), &folder.ID, true)
	require.NoError(t, err)

	// then
	require.NotNil(t, got)
	assert.Equal(t, "open.txt", got.OriginalName)
	assert.Equal(t, int64(5), got.Size)
	require.Len(t, visible, 1)
	assert.Equal(t, open.ID, visible[0].ID)
	assert.Len(t, all, 2)
}

func TestVaultRepository_DescendantStoredNames(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	root := &repository.VaultFolder{ID: uuid.New(), Name: "root", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), root))
	child := &repository.VaultFolder{ID: uuid.New(), ParentID: &root.ID, Name: "child", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), child))
	stored := uuid.NewString() + ".bin"
	file := &repository.VaultFile{ID: uuid.New(), FolderID: &child.ID, OriginalName: "f.bin", StoredName: stored, Mime: "application/octet-stream", Size: 3, UploadedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFile(context.Background(), file))

	// when
	names, err := repos.Vault.DescendantStoredNames(context.Background(), root.ID)

	// then
	require.NoError(t, err)
	assert.Contains(t, names, stored)
}

func TestVaultRepository_DeleteFolderCascades(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	root := &repository.VaultFolder{ID: uuid.New(), Name: "root", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), root))
	child := &repository.VaultFolder{ID: uuid.New(), ParentID: &root.ID, Name: "child", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), child))
	file := &repository.VaultFile{ID: uuid.New(), FolderID: &child.ID, OriginalName: "f.bin", StoredName: uuid.NewString() + ".bin", Mime: "application/octet-stream", Size: 3, UploadedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFile(context.Background(), file))

	// when
	require.NoError(t, repos.Vault.DeleteFolder(context.Background(), root.ID))

	// then
	gotChild, err := repos.Vault.GetFolder(context.Background(), child.ID)
	require.NoError(t, err)
	assert.Nil(t, gotChild, "child folder should be cascade-deleted")
	gotFile, err := repos.Vault.GetFile(context.Background(), file.ID)
	require.NoError(t, err)
	assert.Nil(t, gotFile, "file under deleted subtree should be cascade-deleted")
}

func TestVaultRepository_RenameFolder(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	folder := &repository.VaultFolder{ID: uuid.New(), Name: "before", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), folder))

	// when
	require.NoError(t, repos.Vault.RenameFolder(context.Background(), folder.ID, "after"))
	got, err := repos.Vault.GetFolder(context.Background(), folder.ID)

	// then
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "after", got.Name)
}

func TestVaultRepository_ListRootFolders(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)
	rootA := &repository.VaultFolder{ID: uuid.New(), Name: "alpha", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), rootA))
	nested := &repository.VaultFolder{ID: uuid.New(), ParentID: &rootA.ID, Name: "nested", CreatedBy: user.ID}
	require.NoError(t, repos.Vault.CreateFolder(context.Background(), nested))

	// when
	roots, err := repos.Vault.ListChildFolders(context.Background(), nil, true)

	// then
	require.NoError(t, err)
	ids := make([]uuid.UUID, len(roots))
	for i := range roots {
		ids[i] = roots[i].ID
	}
	assert.Contains(t, ids, rootA.ID)
	assert.NotContains(t, ids, nested.ID, "nested folder must not appear at root level")
}
