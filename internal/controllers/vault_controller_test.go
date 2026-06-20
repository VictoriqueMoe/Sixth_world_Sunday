package controllers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/controllers/utils/testutil"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/filehost"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newVaultHarness(t *testing.T) (*testutil.Harness, *filehost.MockService) {
	h := testutil.NewHarness(t)
	vs := filehost.NewMockService(t)

	s := &Service{
		FileVaultService: vs,
		AuthSession:      h.SessionManager,
		AuthzService:     h.AuthzService,
	}
	for _, setup := range s.getAllVaultRoutes() {
		setup(h.App)
	}
	return h, vs
}

func TestVaultBrowse_OK(t *testing.T) {
	// given
	h, vs := newVaultHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	vs.EXPECT().Browse(mock.Anything, userID, mock.Anything).Return(&dto.VaultBrowseResponse{CanManageLocks: true}, nil)

	// when
	status, body := h.NewRequest("GET", "/files/contents").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(body), `"canManageLocks":true`)
}

func TestVaultBrowse_InvalidFolderQuery(t *testing.T) {
	// given
	h, _ := newVaultHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("GET", "/files/contents?folder=not-a-uuid").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid folder id")
}

func TestVaultBrowse_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newVaultHarness, "GET", "/files/contents", nil)
}

func TestVaultCreateFolder_OK(t *testing.T) {
	// given
	h, vs := newVaultHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	vs.EXPECT().CreateFolder(mock.Anything, userID, "docs", mock.Anything).Return(&dto.VaultFolderResponse{Name: "docs"}, nil)

	// when
	status, body := h.NewRequest("POST", "/files/folders").
		WithCookie("valid-cookie").
		WithJSONBody(map[string]any{"name": "docs"}).
		Do()

	// then
	require.Equal(t, http.StatusCreated, status)
	assert.Contains(t, string(body), `"name":"docs"`)
}

func TestVaultCreateFolder_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newVaultHarness, "POST", "/files/folders", map[string]any{"name": "docs"})
}

func TestVaultCreateFolder_ServiceErrors(t *testing.T) {
	cases := []struct {
		name     string
		svcErr   error
		wantCode int
		wantBody string
	}{
		{"parent not found", filehost.ErrFolderNotFound, http.StatusNotFound, "folder not found"},
		{"invalid name", filehost.ErrInvalidName, http.StatusBadRequest, "invalid name"},
		{"internal", errors.New("boom"), http.StatusInternalServerError, "file vault error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			h, vs := newVaultHarness(t)
			userID := uuid.New()
			h.ExpectValidSession("valid-cookie", userID)
			vs.EXPECT().CreateFolder(mock.Anything, userID, "docs", mock.Anything).Return(nil, tc.svcErr)

			// when
			status, body := h.NewRequest("POST", "/files/folders").
				WithCookie("valid-cookie").
				WithJSONBody(map[string]any{"name": "docs"}).
				Do()

			// then
			require.Equal(t, tc.wantCode, status)
			assert.Contains(t, string(body), tc.wantBody)
		})
	}
}

func TestVaultLockFolder_OK(t *testing.T) {
	// given
	h, vs := newVaultHarness(t)
	userID := uuid.New()
	folderID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	h.ExpectHasPermission(userID, authz.PermLockFiles, true)
	vs.EXPECT().SetFolderLocked(mock.Anything, userID, folderID, true).Return(nil)

	// when
	status, _ := h.NewRequest("POST", "/files/folders/"+folderID.String()+"/lock").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
}

func TestVaultLockFolder_PermissionFailures(t *testing.T) {
	testutil.RunPermissionFailureSuite(t, newVaultHarness, "POST", "/files/folders/"+uuid.New().String()+"/lock", nil, authz.PermLockFiles)
}

func TestVaultUnlockFile_OK(t *testing.T) {
	// given
	h, vs := newVaultHarness(t)
	userID := uuid.New()
	fileID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	h.ExpectHasPermission(userID, authz.PermLockFiles, true)
	vs.EXPECT().SetFileLocked(mock.Anything, userID, fileID, false).Return(nil)

	// when
	status, _ := h.NewRequest("POST", "/files/items/"+fileID.String()+"/unlock").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
}

func TestVaultUpload_MissingFile(t *testing.T) {
	// given
	h, _ := newVaultHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("POST", "/files/upload").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "file is required")
}

func TestVaultDownload_OK(t *testing.T) {
	// given
	h, vs := newVaultHarness(t)
	userID := uuid.New()
	fileID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	path := filepath.Join(t.TempDir(), "stored.bin")
	require.NoError(t, os.WriteFile(path, []byte("payload"), 0o644))
	vs.EXPECT().OpenForDownload(mock.Anything, userID, fileID).Return(&filehost.DownloadInfo{
		Path: path,
		Name: "report.pdf",
		Mime: "application/pdf",
		Size: 7,
	}, nil)

	// when
	status, body := h.NewRequest("GET", "/files/items/"+fileID.String()+"/download").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "payload", string(body))
}

func TestVaultDownload_ServiceErrors(t *testing.T) {
	cases := []struct {
		name     string
		svcErr   error
		wantCode int
		wantBody string
	}{
		{"not found", filehost.ErrFileNotFound, http.StatusNotFound, "file not found"},
		{"internal", errors.New("boom"), http.StatusInternalServerError, "file vault error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			h, vs := newVaultHarness(t)
			userID := uuid.New()
			fileID := uuid.New()
			h.ExpectValidSession("valid-cookie", userID)
			vs.EXPECT().OpenForDownload(mock.Anything, userID, fileID).Return(nil, tc.svcErr)

			// when
			status, body := h.NewRequest("GET", "/files/items/"+fileID.String()+"/download").
				WithCookie("valid-cookie").
				Do()

			// then
			require.Equal(t, tc.wantCode, status)
			assert.Contains(t, string(body), tc.wantBody)
		})
	}
}

func TestVaultDeleteFile_ServiceErrors(t *testing.T) {
	cases := []struct {
		name     string
		svcErr   error
		wantCode int
		wantBody string
	}{
		{"forbidden", filehost.ErrForbidden, http.StatusForbidden, "cannot modify"},
		{"not found", filehost.ErrFileNotFound, http.StatusNotFound, "file not found"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			h, vs := newVaultHarness(t)
			userID := uuid.New()
			fileID := uuid.New()
			h.ExpectValidSession("valid-cookie", userID)
			vs.EXPECT().DeleteFile(mock.Anything, userID, fileID).Return(tc.svcErr)

			// when
			status, body := h.NewRequest("DELETE", "/files/items/"+fileID.String()).
				WithCookie("valid-cookie").
				Do()

			// then
			require.Equal(t, tc.wantCode, status)
			assert.Contains(t, string(body), tc.wantBody)
		})
	}
}

func TestVaultDeleteFile_InvalidID(t *testing.T) {
	// given
	h, _ := newVaultHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("DELETE", "/files/items/not-a-uuid").
		WithCookie("valid-cookie").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid id")
}
