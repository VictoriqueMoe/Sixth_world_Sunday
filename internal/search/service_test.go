package search_test

import (
	"context"
	"errors"
	"testing"

	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/search"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	viewer = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	room   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func newSvc(t *testing.T) (search.Service, *repository.MockSearchRepository, *repository.MockChatRepository) {
	repo := repository.NewMockSearchRepository(t)
	chat := repository.NewMockChatRepository(t)
	return search.NewService(repo, chat), repo, chat
}

func TestService_Search_DelegatesAndDecoratesURLs(t *testing.T) {
	// given
	svc, repo, _ := newSvc(t)
	repo.EXPECT().
		Search(mock.Anything, "wraith", []repository.SearchEntityType{repository.SearchEntityUser}, 20, 0).
		Return([]repository.SearchResult{
			{EntityType: repository.SearchEntityUser, ID: "u1", AuthorUsername: "wraith"},
		}, 1, nil)

	// when
	results, total, err := svc.Search(context.Background(), "wraith",
		[]repository.SearchEntityType{repository.SearchEntityUser}, 20, 0, uuid.Nil, uuid.Nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, results, 1)
	assert.Equal(t, "/user/wraith", results[0].URL)
}

func TestService_Search_EmptyQuery_NoRepoCall(t *testing.T) {
	// given
	svc, _, _ := newSvc(t)

	// when
	results, total, err := svc.Search(context.Background(), "  ", nil, 20, 0, uuid.Nil, uuid.Nil)

	// then
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 0, total)
}

func TestService_Search_ClampsLimit(t *testing.T) {
	// given
	svc, repo, _ := newSvc(t)
	repo.EXPECT().Search(mock.Anything, "x", mock.Anything, 100, 0).Return(nil, 0, nil)

	// when
	_, _, err := svc.Search(context.Background(), "x", nil, 9999, 0, uuid.Nil, uuid.Nil)

	// then
	require.NoError(t, err)
}

func TestService_Search_AppliesDefaults(t *testing.T) {
	// given
	svc, repo, _ := newSvc(t)
	repo.EXPECT().Search(mock.Anything, "x", mock.Anything, 20, 0).Return(nil, 0, nil)

	// when
	_, _, err := svc.Search(context.Background(), "x", nil, 0, -5, uuid.Nil, uuid.Nil)

	// then
	require.NoError(t, err)
}

func TestService_Search_PropagatesError(t *testing.T) {
	// given
	svc, repo, _ := newSvc(t)
	repo.EXPECT().Search(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, 0, errors.New("boom"))

	// when
	_, _, err := svc.Search(context.Background(), "x", nil, 20, 0, uuid.Nil, uuid.Nil)

	// then
	assert.Error(t, err)
}

func TestService_Search_MergesChatForViewerSortedByRank(t *testing.T) {
	// given
	svc, repo, chat := newSvc(t)
	repo.EXPECT().Search(mock.Anything, "knox", []repository.SearchEntityType(nil), 20, 0).
		Return([]repository.SearchResult{
			{EntityType: repository.SearchEntityUser, ID: "u1", AuthorUsername: "knox", Rank: 0.4},
		}, 1, nil)
	chat.EXPECT().SearchMessagesForViewer(mock.Anything, viewer, uuid.Nil, "knox", 20, 0).
		Return([]repository.SearchResult{
			{EntityType: repository.SearchEntityChatMessage, ID: "m1", ParentID: new("room1"), Rank: 0.9},
		}, 1, nil)

	// when
	results, total, err := svc.Search(context.Background(), "knox", nil, 20, 0, viewer, uuid.Nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, results, 2)
	assert.Equal(t, "/rooms/room1#msg-m1", results[0].URL)
	assert.Equal(t, "/user/knox", results[1].URL)
}

func TestService_Search_OnlyChatType_SkipsRepo(t *testing.T) {
	// given
	svc, _, chat := newSvc(t)
	chat.EXPECT().SearchMessagesForViewer(mock.Anything, viewer, uuid.Nil, "knox", 20, 0).
		Return([]repository.SearchResult{
			{EntityType: repository.SearchEntityChatMessage, ID: "m1", ParentID: new("room1")},
		}, 1, nil)

	// when
	results, total, err := svc.Search(context.Background(), "knox",
		[]repository.SearchEntityType{repository.SearchEntityChatMessage}, 20, 0, viewer, uuid.Nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, results, 1)
	assert.Equal(t, "/rooms/room1#msg-m1", results[0].URL)
}

func TestService_Search_ScopesChatToRoom(t *testing.T) {
	// given
	svc, _, chat := newSvc(t)
	chat.EXPECT().SearchMessagesForViewer(mock.Anything, viewer, room, "knox", 20, 0).
		Return([]repository.SearchResult{
			{EntityType: repository.SearchEntityChatMessage, ID: "m1", ParentID: new("room1")},
		}, 1, nil)

	// when
	results, total, err := svc.Search(context.Background(), "knox",
		[]repository.SearchEntityType{repository.SearchEntityChatMessage}, 20, 0, viewer, room)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, results, 1)
	assert.Equal(t, "/rooms/room1#msg-m1", results[0].URL)
}

func TestService_Search_ChatTypeAnonymous_ReturnsNothing(t *testing.T) {
	// given
	svc, _, _ := newSvc(t)

	// when
	results, total, err := svc.Search(context.Background(), "knox",
		[]repository.SearchEntityType{repository.SearchEntityChatMessage}, 20, 0, uuid.Nil, uuid.Nil)

	// then
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 0, total)
}

func TestService_QuickSearch_DelegatesAndDecoratesURL(t *testing.T) {
	// given
	svc, repo, _ := newSvc(t)
	repo.EXPECT().QuickSearch(mock.Anything, "x", 3).Return([]repository.SearchResult{
		{EntityType: repository.SearchEntityUser, ID: "m1", AuthorUsername: "neo"},
	}, nil)

	// when
	results, err := svc.QuickSearch(context.Background(), "x", 3, uuid.Nil)

	// then
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "/user/neo", results[0].URL)
}

func TestService_QuickSearch_MergesChatForViewer(t *testing.T) {
	// given
	svc, repo, chat := newSvc(t)
	repo.EXPECT().QuickSearch(mock.Anything, "x", 3).Return([]repository.SearchResult{
		{EntityType: repository.SearchEntityUser, ID: "m1", AuthorUsername: "neo", Rank: 0.2},
	}, nil)
	chat.EXPECT().SearchMessagesForViewer(mock.Anything, viewer, uuid.Nil, "x", 3, 0).Return([]repository.SearchResult{
		{EntityType: repository.SearchEntityChatMessage, ID: "c1", ParentID: new("room1"), Rank: 0.8},
	}, 1, nil)

	// when
	results, err := svc.QuickSearch(context.Background(), "x", 3, viewer)

	// then
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "/rooms/room1#msg-c1", results[0].URL)
}

func TestService_QuickSearch_EmptyQuery_NoRepoCall(t *testing.T) {
	// given
	svc, _, _ := newSvc(t)

	// when
	results, err := svc.QuickSearch(context.Background(), " ", 3, uuid.Nil)

	// then
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestService_QuickSearch_ClampsPerTypeLimit(t *testing.T) {
	// given
	svc, repo, _ := newSvc(t)
	repo.EXPECT().QuickSearch(mock.Anything, "x", 10).Return(nil, nil)

	// when
	_, err := svc.QuickSearch(context.Background(), "x", 999, uuid.Nil)

	// then
	require.NoError(t, err)
}

func TestService_ParseTypes_AllReturnsNil(t *testing.T) {
	// given
	svc, _, _ := newSvc(t)

	// when / then
	assert.Nil(t, svc.ParseTypes(""))
	assert.Nil(t, svc.ParseTypes("all"))
}

func TestService_ParseTypes_CommaList(t *testing.T) {
	// given
	svc, _, _ := newSvc(t)

	// when
	got := svc.ParseTypes("user,chat_message")

	// then
	assert.Equal(t, []repository.SearchEntityType{
		repository.SearchEntityUser,
		repository.SearchEntityChatMessage,
	}, got)
}

func TestService_ChildEntityTypes(t *testing.T) {
	// given
	svc, _, _ := newSvc(t)

	// when
	children := svc.ChildEntityTypes()

	// then
	assert.NotContains(t, children, repository.SearchEntityUser)
}

func TestBuildURL(t *testing.T) {
	// given
	cases := []struct {
		name string
		r    repository.SearchResult
		want string
	}{
		{"user", repository.SearchResult{EntityType: repository.SearchEntityUser, AuthorUsername: "wraith"}, "/user/wraith"},
		{"chat message", repository.SearchResult{EntityType: repository.SearchEntityChatMessage, ID: "m1", ParentID: new("room1")}, "/rooms/room1#msg-m1"},
		{"chat message with timestamp", repository.SearchResult{EntityType: repository.SearchEntityChatMessage, ID: "m1", ParentID: new("room1"), CreatedAt: "2026-05-29T19:00:00.5Z"}, "/rooms/room1?at=2026-05-29T19%3A00%3A00.5Z#msg-m1"},
		{"chat message without room", repository.SearchResult{EntityType: repository.SearchEntityChatMessage, ID: "m1", ParentID: nil}, ""},
		{"unknown", repository.SearchResult{EntityType: "nonsense"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when / then
			assert.Equal(t, tc.want, search.BuildURL(tc.r))
		})
	}
}
