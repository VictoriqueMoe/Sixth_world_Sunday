package controllers

import (
	"net/http"
	"testing"

	"Sixth_world_Suday/internal/controllers/utils/testutil"
	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/search"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type searchDeps struct {
	svc *search.MockService
}

func newSearchHarness(t *testing.T) (*testutil.Harness, searchDeps) {
	h := testutil.NewHarness(t)
	deps := searchDeps{
		svc: search.NewMockService(t),
	}
	s := &Service{
		SearchService:   deps.svc,
		SettingsService: h.SettingsService,
		AuthSession:     h.SessionManager,
		AuthzService:    h.AuthzService,
	}
	for _, setup := range s.getAllSearchRoutes() {
		setup(h.App)
	}
	return h, deps
}

func TestSearchController_EmptyQuery_ReturnsEmpty(t *testing.T) {
	// given
	h, _ := newSearchHarness(t)

	// when
	status, body := h.NewRequest(http.MethodGet, "/search?q=").Do()

	// then
	assert.Equal(t, http.StatusOK, status)
	parsed := testutil.UnmarshalJSON[map[string]any](t, body)
	assert.Equal(t, float64(0), parsed["total"])
	results, _ := parsed["results"].([]any)
	assert.Empty(t, results)
}

func TestSearchController_FullSearch_PassesParamsAndShapesResponse(t *testing.T) {
	// given
	h, deps := newSearchHarness(t)
	deps.svc.EXPECT().ParseTypes("user").
		Return([]repository.SearchEntityType{repository.SearchEntityUser})
	deps.svc.EXPECT().
		Search(mock.Anything, "wraith", []repository.SearchEntityType{repository.SearchEntityUser}, 10, 5, uuid.Nil, uuid.Nil).
		Return([]search.Result{
			{
				SearchResult: repository.SearchResult{
					EntityType:        repository.SearchEntityUser,
					ID:                "user-id",
					Title:             "Wraith",
					Snippet:           "matched <mark>wraith</mark>",
					AuthorUsername:    "wraith",
					AuthorDisplayName: "Wraith",
					CreatedAt:         "2026-04-28T00:00:00Z",
				},
				URL: "/user/wraith",
			},
		}, 1, nil)

	// when
	status, body := h.NewRequest(http.MethodGet, "/search?q=wraith&types=user&limit=10&offset=5").Do()

	// then
	assert.Equal(t, http.StatusOK, status)
	parsed := testutil.UnmarshalJSON[map[string]any](t, body)
	assert.Equal(t, float64(1), parsed["total"])
	results, _ := parsed["results"].([]any)
	require.Len(t, results, 1)
	first := results[0].(map[string]any)
	assert.Equal(t, "user", first["type"])
	assert.Equal(t, "/user/wraith", first["url"])
	author := first["author"].(map[string]any)
	assert.Equal(t, "wraith", author["username"])
}

func TestSearchController_QuickSearch_DelegatesToService(t *testing.T) {
	// given
	h, deps := newSearchHarness(t)
	deps.svc.EXPECT().QuickSearch(mock.Anything, "wraith", 5, uuid.Nil).
		Return([]search.Result{
			{
				SearchResult: repository.SearchResult{EntityType: repository.SearchEntityUser, ID: "user-id", Title: "Wraith", AuthorUsername: "wraith"},
				URL:          "/user/wraith",
			},
		}, nil)

	// when
	status, body := h.NewRequest(http.MethodGet, "/search/quick?q=wraith&perType=5").Do()

	// then
	assert.Equal(t, http.StatusOK, status)
	parsed := testutil.UnmarshalJSON[map[string]any](t, body)
	results := parsed["results"].([]any)
	require.Len(t, results, 1)
	first := results[0].(map[string]any)
	assert.Equal(t, "/user/wraith", first["url"])
}

func TestSearchController_QuickSearch_EmptyQ_NoCall(t *testing.T) {
	// given
	h, _ := newSearchHarness(t)

	// when
	status, body := h.NewRequest(http.MethodGet, "/search/quick?q=").Do()

	// then
	assert.Equal(t, http.StatusOK, status)
	parsed := testutil.UnmarshalJSON[map[string]any](t, body)
	results, _ := parsed["results"].([]any)
	assert.Empty(t, results)
}
