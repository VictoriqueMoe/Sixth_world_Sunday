package repository_test

import (
	"context"
	"testing"

	"Sixth_world_Suday/internal/repository/repotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsRepository_GetOverview_EmptyDB(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)

	// when
	stats, err := repos.Stats.GetOverview(context.Background())

	// then
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalUsers)
	assert.Equal(t, 0, stats.TotalMessages)
	assert.Equal(t, 0, stats.TotalRooms)
	assert.Equal(t, 0, stats.NewUsers24h)
}

func TestStatsRepository_GetOverview_CountsUsers(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	repotest.CreateUser(t, repos)
	repotest.CreateUser(t, repos)
	repotest.CreateUser(t, repos)

	// when
	stats, err := repos.Stats.GetOverview(context.Background())

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalUsers)
	assert.Equal(t, 3, stats.NewUsers24h)
	assert.Equal(t, 3, stats.NewUsers7d)
	assert.Equal(t, 3, stats.NewUsers30d)
}

func TestStatsRepository_GetMostActiveUsers_Empty(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)

	// when
	users, err := repos.Stats.GetMostActiveUsers(context.Background(), 10)

	// then
	require.NoError(t, err)
	assert.Empty(t, users)
}
