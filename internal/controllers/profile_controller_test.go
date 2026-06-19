package controllers

import (
	"errors"
	"net/http"
	"testing"

	"Sixth_world_Suday/internal/controllers/utils/testutil"
	"Sixth_world_Suday/internal/dto"
	"Sixth_world_Suday/internal/profile"
	"Sixth_world_Suday/internal/ws"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type profileDeps struct {
	profileSvc *profile.MockService
}

func newProfileHarness(t *testing.T) (*testutil.Harness, profileDeps) {
	h := testutil.NewHarness(t)
	deps := profileDeps{
		profileSvc: profile.NewMockService(t),
	}

	s := &Service{
		ProfileService:  deps.profileSvc,
		SettingsService: h.SettingsService,
		AuthSession:     h.SessionManager,
		AuthzService:    h.AuthzService,
		Hub:             ws.NewHub(),
	}
	for _, setup := range s.getAllProfileRoutes() {
		setup(h.App)
	}
	return h, deps
}

func TestGetProfile_Anonymous_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	expected := &dto.UserProfileResponse{UserResponse: dto.UserResponse{ID: userID, Username: "nightjar"}}
	deps.profileSvc.EXPECT().GetProfile(mock.Anything, "nightjar", uuid.Nil).Return(expected, nil)

	// when
	status, body := h.NewRequest("GET", "/users/nightjar").Do()

	// then
	require.Equal(t, http.StatusOK, status)
	got := testutil.UnmarshalJSON[dto.UserProfileResponse](t, body)
	assert.Equal(t, "nightjar", got.Username)
	assert.False(t, got.Online)
}

func TestGetProfile_Authenticated_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	viewerID := uuid.New()
	profileID := uuid.New()
	h.ExpectValidSession("valid-cookie", viewerID)
	expected := &dto.UserProfileResponse{UserResponse: dto.UserResponse{ID: profileID, Username: "nightjar"}}
	deps.profileSvc.EXPECT().GetProfile(mock.Anything, "nightjar", viewerID).Return(expected, nil)

	// when
	status, _ := h.NewRequest("GET", "/users/nightjar").WithCookie("valid-cookie").Do()

	// then
	require.Equal(t, http.StatusOK, status)
}

func TestGetProfile_NotFound(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	deps.profileSvc.EXPECT().GetProfile(mock.Anything, "ghost", uuid.Nil).Return(nil, profile.ErrUserNotFound)

	// when
	status, body := h.NewRequest("GET", "/users/ghost").Do()

	// then
	require.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, string(body), "user not found")
}

func TestGetProfile_InternalError(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	deps.profileSvc.EXPECT().GetProfile(mock.Anything, "nightjar", uuid.Nil).Return(nil, errors.New("boom"))

	// when
	status, body := h.NewRequest("GET", "/users/nightjar").Do()

	// then
	require.Equal(t, http.StatusInternalServerError, status)
	assert.Contains(t, string(body), "failed to get profile")
}

func TestUpdateProfile_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newProfileHarness, "PUT", "/auth/profile",
		dto.UpdateProfileRequest{DisplayName: "Nightjar"})
}

func TestUpdateProfile_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.UpdateProfileRequest{DisplayName: "Nightjar", Bio: "runner"}
	deps.profileSvc.EXPECT().UpdateProfile(mock.Anything, userID, req).Return(nil)

	// when
	status, _ := h.NewRequest("PUT", "/auth/profile").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
}

func TestUpdateProfile_MissingDisplayName(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("PUT", "/auth/profile").
		WithCookie("valid-cookie").
		WithJSONBody(dto.UpdateProfileRequest{}).
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "display name is required")
}

func TestUpdateProfile_BadJSON(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("PUT", "/auth/profile").
		WithCookie("valid-cookie").
		WithRawBody("not json", "application/json").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid request body")
}

func TestUpdateProfile_ServiceError(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.UpdateProfileRequest{DisplayName: "Nightjar"}
	deps.profileSvc.EXPECT().UpdateProfile(mock.Anything, userID, req).Return(errors.New("boom"))

	// when
	status, body := h.NewRequest("PUT", "/auth/profile").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusInternalServerError, status)
	assert.Contains(t, string(body), "failed to update profile")
}

func TestUpdateProfile_InvalidDOB(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.UpdateProfileRequest{DisplayName: "Nightjar", DOB: "15-04-2000"}
	deps.profileSvc.EXPECT().UpdateProfile(mock.Anything, userID, req).Return(profile.ErrInvalidDOB)

	// when
	status, body := h.NewRequest("PUT", "/auth/profile").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), profile.ErrInvalidDOB.Error())
}

func TestUploadAvatar_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newProfileHarness, "POST", "/auth/avatar", nil)
}

func TestUploadAvatar_MissingFile(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("POST", "/auth/avatar").
		WithCookie("valid-cookie").
		WithRawBody("", "multipart/form-data; boundary=----xxx").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "avatar file is required")
}

func TestUploadBanner_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newProfileHarness, "POST", "/auth/banner", nil)
}

func TestUploadBanner_MissingFile(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("POST", "/auth/banner").
		WithCookie("valid-cookie").
		WithRawBody("", "multipart/form-data; boundary=----xxx").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "banner file is required")
}

func TestChangePassword_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newProfileHarness, "PUT", "/auth/password",
		dto.ChangePasswordRequest{OldPassword: "old", NewPassword: "newnewnew"})
}

func TestChangePassword_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.ChangePasswordRequest{OldPassword: "old", NewPassword: "newnewnew"}
	deps.profileSvc.EXPECT().ChangePassword(mock.Anything, userID, req).Return(nil)

	// when
	status, _ := h.NewRequest("PUT", "/auth/password").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
}

func TestChangePassword_BadJSON(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("PUT", "/auth/password").
		WithCookie("valid-cookie").
		WithRawBody("not json", "application/json").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid request body")
}

func TestChangePassword_TooShort(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.ChangePasswordRequest{OldPassword: "old", NewPassword: "x"}
	deps.profileSvc.EXPECT().ChangePassword(mock.Anything, userID, req).Return(profile.ErrPasswordTooShort)

	// when
	status, body := h.NewRequest("PUT", "/auth/password").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "new password must be at least")
	assert.Contains(t, string(body), "characters")
}

func TestChangePassword_GenericError(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.ChangePasswordRequest{OldPassword: "bad", NewPassword: "newnewnew"}
	deps.profileSvc.EXPECT().ChangePassword(mock.Anything, userID, req).Return(errors.New("wrong old password"))

	// when
	status, body := h.NewRequest("PUT", "/auth/password").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "wrong old password")
}

func TestDeleteAccount_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, newProfileHarness, "DELETE", "/auth/account",
		dto.DeleteAccountRequest{Password: "pw"})
}

func TestDeleteAccount_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.DeleteAccountRequest{Password: "pw"}
	deps.profileSvc.EXPECT().DeleteAccount(mock.Anything, userID, req).Return(nil)

	// when
	status, _ := h.NewRequest("DELETE", "/auth/account").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusOK, status)
}

func TestDeleteAccount_BadJSON(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)

	// when
	status, body := h.NewRequest("DELETE", "/auth/account").
		WithCookie("valid-cookie").
		WithRawBody("not json", "application/json").
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid request body")
}

func TestDeleteAccount_ServiceError(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	userID := uuid.New()
	h.ExpectValidSession("valid-cookie", userID)
	req := dto.DeleteAccountRequest{Password: "wrong"}
	deps.profileSvc.EXPECT().DeleteAccount(mock.Anything, userID, req).Return(errors.New("invalid password"))

	// when
	status, body := h.NewRequest("DELETE", "/auth/account").
		WithCookie("valid-cookie").
		WithJSONBody(req).
		Do()

	// then
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid password")
}

func TestGetOnlineStatus_EmptyIDs(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)

	// when
	status, body := h.NewRequest("GET", "/users/online").Do()

	// then
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{}`, string(body))
}

func TestGetOnlineStatus_WithIDs(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)
	id1 := uuid.New()

	// when
	status, body := h.NewRequest("GET", "/users/online?ids="+id1.String()+",not-a-uuid,").Do()

	// then
	require.Equal(t, http.StatusOK, status)
	got := testutil.UnmarshalJSON[map[string]bool](t, body)
	_, ok := got[id1.String()]
	assert.True(t, ok)
	assert.Len(t, got, 1)
}

func TestSearchUsers_EmptyQuery(t *testing.T) {
	// given
	h, _ := newProfileHarness(t)

	// when
	status, body := h.NewRequest("GET", "/users/search").Do()

	// then
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `[]`, string(body))
}

func TestSearchUsers_Anonymous_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	u := dto.UserResponse{ID: uuid.New(), Username: "nightjar"}
	deps.profileSvc.EXPECT().SearchUsers(mock.Anything, "be", 10).Return([]dto.UserResponse{u}, nil)

	// when
	status, body := h.NewRequest("GET", "/users/search?q=be").Do()

	// then
	require.Equal(t, http.StatusOK, status)
	got := testutil.UnmarshalJSON[[]map[string]any](t, body)
	require.Len(t, got, 1)
	assert.Equal(t, "nightjar", got[0]["username"])
}

func TestSearchUsers_InternalError(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	deps.profileSvc.EXPECT().SearchUsers(mock.Anything, "be", 10).Return(nil, errors.New("boom"))

	// when
	status, body := h.NewRequest("GET", "/users/search?q=be").Do()

	// then
	require.Equal(t, http.StatusInternalServerError, status)
	assert.Contains(t, string(body), "search failed")
}

func TestListUsersPublic_OK(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	users := []dto.UserResponse{{ID: uuid.New(), Username: "nightjar"}}
	deps.profileSvc.EXPECT().ListPublicUsers(mock.Anything).Return(users, nil)

	// when
	status, body := h.NewRequest("GET", "/users").Do()

	// then
	require.Equal(t, http.StatusOK, status)
	got := testutil.UnmarshalJSON[[]map[string]any](t, body)
	require.Len(t, got, 1)
	assert.Equal(t, "nightjar", got[0]["username"])
	assert.Equal(t, false, got[0]["online"])
}

func TestListUsersPublic_InternalError(t *testing.T) {
	// given
	h, deps := newProfileHarness(t)
	deps.profileSvc.EXPECT().ListPublicUsers(mock.Anything).Return(nil, errors.New("boom"))

	// when
	status, body := h.NewRequest("GET", "/users").Do()

	// then
	require.Equal(t, http.StatusInternalServerError, status)
	assert.Contains(t, string(body), "failed to list users")
}
