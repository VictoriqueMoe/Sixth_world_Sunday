package filehost

import (
	"context"
	"strings"
	"testing"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/repository"
	"Sixth_world_Sunday/internal/role"
	"Sixth_world_Sunday/internal/settings"
	"Sixth_world_Sunday/internal/ws"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T, vaultDir string) (
	*service,
	*repository.MockVaultRepository,
	*authz.MockService,
	*settings.MockService,
) {
	repo := repository.NewMockVaultRepository(t)
	authzSvc := authz.NewMockService(t)
	settingsSvc := settings.NewMockService(t)
	svc := NewService(repo, authzSvc, settingsSvc, ws.NewHub(), vaultDir).(*service)
	return svc, repo, authzSvc, settingsSvc
}

func TestBrowse_Root_NonPrivileged_HidesLocked(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	repo.EXPECT().ListChildFolders(mock.Anything, mock.Anything, false).Return(nil, nil)
	repo.EXPECT().ListFolderFiles(mock.Anything, mock.Anything, false).Return(nil, nil)

	// when
	res, err := svc.Browse(context.Background(), userID, nil)

	// then
	require.NoError(t, err)
	assert.False(t, res.CanManageLocks)
}

func TestBrowse_Root_Privileged_IncludesLocked(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(true)
	repo.EXPECT().ListChildFolders(mock.Anything, mock.Anything, true).Return(nil, nil)
	repo.EXPECT().ListFolderFiles(mock.Anything, mock.Anything, true).Return(nil, nil)

	// when
	res, err := svc.Browse(context.Background(), userID, nil)

	// then
	require.NoError(t, err)
	assert.True(t, res.CanManageLocks)
}

func TestBrowse_LockedFolder_NonPrivileged_NotFound(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	repo.EXPECT().GetFolder(mock.Anything, folderID).Return(&repository.VaultFolder{ID: folderID, Name: "secret"}, nil)
	repo.EXPECT().ChainLocked(mock.Anything, folderID).Return(true, nil)

	// when
	_, err := svc.Browse(context.Background(), userID, &folderID)

	// then
	require.ErrorIs(t, err, ErrFolderNotFound)
}

func TestBrowse_LockedFolder_Privileged_Visible(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(true)
	repo.EXPECT().GetFolder(mock.Anything, folderID).Return(&repository.VaultFolder{ID: folderID, Name: "secret", Locked: true}, nil)
	repo.EXPECT().FolderChain(mock.Anything, folderID).Return([]repository.VaultFolder{{ID: folderID, Name: "secret"}}, nil)
	repo.EXPECT().ListChildFolders(mock.Anything, mock.Anything, true).Return(nil, nil)
	repo.EXPECT().ListFolderFiles(mock.Anything, mock.Anything, true).Return(nil, nil)

	// when
	res, err := svc.Browse(context.Background(), userID, &folderID)

	// then
	require.NoError(t, err)
	require.NotNil(t, res.Folder)
	assert.Equal(t, folderID.String(), res.Folder.ID)
	require.Len(t, res.Breadcrumbs, 1)
}

func TestOpenForDownload_UnlockedFile_OK(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	fileID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	repo.EXPECT().GetFile(mock.Anything, fileID).Return(&repository.VaultFile{
		ID: fileID, OriginalName: "report.pdf", StoredName: fileID.String() + ".pdf", Mime: "application/pdf",
	}, nil)

	// when
	info, err := svc.OpenForDownload(context.Background(), userID, fileID)

	// then
	require.NoError(t, err)
	assert.Equal(t, "report.pdf", info.Name)
	assert.Contains(t, info.Path, fileID.String())
}

func TestOpenForDownload_FileInLockedFolder_NonPrivileged_NotFound(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	fileID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	repo.EXPECT().GetFile(mock.Anything, fileID).Return(&repository.VaultFile{ID: fileID, FolderID: &folderID}, nil)
	repo.EXPECT().ChainLocked(mock.Anything, folderID).Return(true, nil)

	// when
	_, err := svc.OpenForDownload(context.Background(), userID, fileID)

	// then
	require.ErrorIs(t, err, ErrFileNotFound)
}

func TestOpenForDownload_LockedFile_NonPrivileged_NotFound(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	fileID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	repo.EXPECT().GetFile(mock.Anything, fileID).Return(&repository.VaultFile{ID: fileID, Locked: true}, nil)

	// when
	_, err := svc.OpenForDownload(context.Background(), userID, fileID)

	// then
	require.ErrorIs(t, err, ErrFileNotFound)
}

func TestSetFolderLocked_NonPrivileged_Forbidden(t *testing.T) {
	// given
	svc, _, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)

	// when
	err := svc.SetFolderLocked(context.Background(), userID, folderID, true)

	// then
	require.ErrorIs(t, err, ErrForbidden)
}

func TestSetFolderLocked_Privileged_Persists(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(true)
	repo.EXPECT().GetFolder(mock.Anything, folderID).Return(&repository.VaultFolder{ID: folderID}, nil)
	repo.EXPECT().SetFolderLocked(mock.Anything, folderID, true).Return(nil)

	// when
	err := svc.SetFolderLocked(context.Background(), userID, folderID, true)

	// then
	require.NoError(t, err)
}

func TestCreateFolder_EmptyName_Invalid(t *testing.T) {
	// given
	svc, _, _, _ := newTestService(t, t.TempDir())

	// when
	_, err := svc.CreateFolder(context.Background(), uuid.New(), "   ", nil)

	// then
	require.ErrorIs(t, err, ErrInvalidName)
}

func TestDeleteFolder_NotOwnerNotStaff_Forbidden(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	authzSvc.EXPECT().GetRole(mock.Anything, userID).Return(role.Role(""), nil)
	repo.EXPECT().GetFolder(mock.Anything, folderID).Return(&repository.VaultFolder{ID: folderID, CreatedBy: uuid.New()}, nil)
	repo.EXPECT().ChainLocked(mock.Anything, folderID).Return(false, nil)

	// when
	err := svc.DeleteFolder(context.Background(), userID, folderID)

	// then
	require.ErrorIs(t, err, ErrForbidden)
}

func TestDeleteFolder_Staff_RemovesBytes(t *testing.T) {
	// given
	svc, repo, authzSvc, _ := newTestService(t, t.TempDir())
	userID := uuid.New()
	folderID := uuid.New()
	authzSvc.EXPECT().Can(mock.Anything, userID, authz.PermLockFiles).Return(false)
	authzSvc.EXPECT().GetRole(mock.Anything, userID).Return(authz.RoleModerator, nil)
	repo.EXPECT().GetFolder(mock.Anything, folderID).Return(&repository.VaultFolder{ID: folderID, CreatedBy: uuid.New()}, nil)
	repo.EXPECT().ChainLocked(mock.Anything, folderID).Return(false, nil)
	repo.EXPECT().DescendantStoredNames(mock.Anything, folderID).Return(nil, nil)
	repo.EXPECT().DeleteFolder(mock.Anything, folderID).Return(nil)

	// when
	err := svc.DeleteFolder(context.Background(), userID, folderID)

	// then
	require.NoError(t, err)
}

func TestUpload_TooLarge(t *testing.T) {
	// given
	svc, _, _, settingsSvc := newTestService(t, t.TempDir())
	settingsSvc.EXPECT().GetInt(mock.Anything, mock.Anything).Return(10)

	// when
	_, err := svc.Upload(context.Background(), uuid.New(), nil, "big.bin", 99, strings.NewReader("x"))

	// then
	require.ErrorIs(t, err, ErrFileTooLarge)
}

func TestUpload_Root_WritesAndRecords(t *testing.T) {
	// given
	svc, repo, _, settingsSvc := newTestService(t, t.TempDir())
	userID := uuid.New()
	settingsSvc.EXPECT().GetInt(mock.Anything, mock.Anything).Return(50 * 1024 * 1024)
	repo.EXPECT().CreateFile(mock.Anything, mock.Anything).Return(nil)

	// when
	res, err := svc.Upload(context.Background(), userID, nil, "notes.txt", 5, strings.NewReader("hello"))

	// then
	require.NoError(t, err)
	assert.Equal(t, "notes.txt", res.Name)
	assert.Equal(t, int64(5), res.Size)
}

func TestSanitizeName_StripsPathSeparators(t *testing.T) {
	assert.Equal(t, "..etcpasswd", sanitizeName("../etc/passwd"))
	assert.Equal(t, "windowsevil", sanitizeName("windows\\evil"))
	assert.Equal(t, "", sanitizeName("   "))
}

func TestSafeExt_RejectsWeirdExtensions(t *testing.T) {
	assert.Equal(t, ".pdf", safeExt("report.pdf"))
	assert.Equal(t, ".gz", safeExt("archive.tar.gz"))
	assert.Equal(t, "", safeExt("weird.p$p"))
	assert.Equal(t, "", safeExt("noextension"))
}
