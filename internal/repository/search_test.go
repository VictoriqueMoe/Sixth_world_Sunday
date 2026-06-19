package repository_test

import (
	"context"
	"testing"

	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/repository/repotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func searchRepoOnce(t *testing.T, repos *repository.Repositories, query string, types []repository.SearchEntityType) []repository.SearchResult {
	t.Helper()
	results, _, err := repos.Search.Search(context.Background(), query, types, 20, 0)
	require.NoError(t, err)
	return results
}

func TestSearchRepository_User_TrigramOnUsername(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	repotest.CreateUser(t, repos, repotest.WithUsername("hatchet1986"), repotest.WithDisplayName("Random Display"))

	// when
	results := searchRepoOnce(t, repos, "hatchet", []repository.SearchEntityType{repository.SearchEntityUser})

	// then
	require.NotEmpty(t, results)
	assert.Equal(t, "hatchet1986", results[0].AuthorUsername)
}

func TestSearchRepository_AllRegisteredEntitiesRoundTrip(t *testing.T) {
	// given
	registered := []repository.SearchEntityType{
		repository.SearchEntityUser,
	}

	// when / then - just confirms each entity has a valid registry entry
	for _, typ := range registered {
		_, ok := repository.SearchSourceFor(typ)
		require.Truef(t, ok, "missing registry entry for %s", typ)
	}
}

func TestSearchRepository_SearchSources_RegistryIntegrity(t *testing.T) {
	// given / when
	srcs := repository.SearchSources()

	// then
	assert.NotEmpty(t, srcs)
	for _, s := range srcs {
		assert.NotEmptyf(t, s.From, "%s missing From", s.Type)
		assert.NotEmptyf(t, s.IDExpr, "%s missing IDExpr", s.Type)
		assert.NotEmptyf(t, s.SearchVector, "%s missing SearchVector", s.Type)
	}
}
