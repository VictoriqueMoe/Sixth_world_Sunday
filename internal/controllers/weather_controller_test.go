package controllers

import (
	"net/http"
	"testing"

	"Sixth_world_Sunday/internal/controllers/utils/testutil"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/weather"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newWeatherHarness(t *testing.T) (*testutil.Harness, *weather.MockService) {
	h := testutil.NewHarness(t)
	weatherMock := weather.NewMockService(t)

	s := &Service{
		WeatherService: weatherMock,
		AuthSession:    h.SessionManager,
		AuthzService:   h.AuthzService,
	}
	for _, setup := range s.getAllWeatherRoutes() {
		setup(h.App)
	}

	return h, weatherMock
}

func weatherFactory(t *testing.T) (*testutil.Harness, *weather.MockService) {
	return newWeatherHarness(t)
}

func TestWeatherLocations_AuthFailures(t *testing.T) {
	testutil.RunAuthFailureSuite(t, weatherFactory, "GET", "/weather/locations", nil)
}

func TestWeatherListEmpty(t *testing.T) {
	h, m := newWeatherHarness(t)
	user := uuid.New()
	h.ExpectValidSession("c", user)
	m.EXPECT().List(mock.Anything, user).Return(&dto.WeatherLocationListResponse{}, nil)

	status, body := h.NewRequest("GET", "/weather/locations").WithCookie("c").Do()

	require.Equal(t, http.StatusOK, status)
	res := testutil.UnmarshalJSON[dto.WeatherLocationListResponse](t, body)
	assert.Empty(t, res.Locations)
}

func TestWeatherSaveLocation(t *testing.T) {
	h, m := newWeatherHarness(t)
	user := uuid.New()
	h.ExpectValidSession("c", user)
	m.EXPECT().
		Save(mock.Anything, user, mock.Anything).
		Return(&dto.WeatherLocationResponse{PlaceName: "London", IsDefault: true}, nil)

	body := map[string]any{"place_name": "London", "latitude": 51.5, "longitude": -0.12}
	status, resp := h.NewRequest("POST", "/weather/locations").WithCookie("c").WithJSONBody(body).Do()

	require.Equal(t, http.StatusCreated, status)
	loc := testutil.UnmarshalJSON[dto.WeatherLocationResponse](t, resp)
	assert.Equal(t, "London", loc.PlaceName)
	assert.True(t, loc.IsDefault)
}

func TestWeatherUpdateNotFoundMapsTo404(t *testing.T) {
	h, m := newWeatherHarness(t)
	user := uuid.New()
	id := uuid.New()
	h.ExpectValidSession("c", user)
	m.EXPECT().Rename(mock.Anything, user, id, mock.Anything).Return(nil, weather.ErrLocationNotFound)

	status, _ := h.NewRequest("PUT", "/weather/locations/"+id.String()).
		WithCookie("c").
		WithJSONBody(map[string]any{"label": "x"}).
		Do()

	require.Equal(t, http.StatusNotFound, status)
}

func TestWeatherDeleteForbiddenMapsTo403(t *testing.T) {
	h, m := newWeatherHarness(t)
	user := uuid.New()
	id := uuid.New()
	h.ExpectValidSession("c", user)
	m.EXPECT().Delete(mock.Anything, user, id).Return(weather.ErrForbidden)

	status, _ := h.NewRequest("DELETE", "/weather/locations/"+id.String()).WithCookie("c").Do()

	require.Equal(t, http.StatusForbidden, status)
}
