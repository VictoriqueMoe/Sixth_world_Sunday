package controllers

import (
	"net/http"
	"testing"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/controllers/utils/testutil"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/maps"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newMapHarness(t *testing.T) (*testutil.Harness, *maps.MockService) {
	h := testutil.NewHarness(t)
	mapMock := maps.NewMockService(t)

	s := &Service{
		MapService:   mapMock,
		AuthSession:  h.SessionManager,
		AuthzService: h.AuthzService,
	}
	for _, setup := range s.getAllMapRoutes() {
		setup(h.App)
	}

	return h, mapMock
}

func mapFactory(t *testing.T) (*testutil.Harness, *maps.MockService) {
	return newMapHarness(t)
}

func TestMaps_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, mapFactory, "GET", "/maps", nil)
}

func TestMapCreate_PermissionFailures(t *testing.T) {
	testutil.RunPermissionFailureSuite(
		t,
		mapFactory,
		"POST",
		"/maps",
		map[string]any{"source_url": "https://www.google.com/maps/d/edit?mid=x"},
		authz.PermManageMaps,
	)
}

func TestMapListAnyAuthedUser(t *testing.T) {
	h, m := newMapHarness(t)
	user := uuid.New()
	h.ExpectValidSession("c", user)
	m.EXPECT().List(mock.Anything, user).Return(&dto.MapListResponse{}, nil)

	status, body := h.NewRequest("GET", "/maps").WithCookie("c").Do()

	require.Equal(t, http.StatusOK, status)
	res := testutil.UnmarshalJSON[dto.MapListResponse](t, body)
	require.Empty(t, res.Maps)
}

func TestMapCreateInvalidURLMapsTo400(t *testing.T) {
	h, m := newMapHarness(t)
	user := uuid.New()
	h.ExpectValidSession("c", user)
	h.ExpectHasPermission(user, authz.PermManageMaps, true)
	m.EXPECT().Create(mock.Anything, user, mock.Anything).Return(nil, maps.ErrInvalidInput)

	status, _ := h.NewRequest("POST", "/maps").
		WithCookie("c").
		WithJSONBody(map[string]any{"source_url": "not a url"}).
		Do()

	require.Equal(t, http.StatusBadRequest, status)
}
