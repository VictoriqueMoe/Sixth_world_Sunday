package repository_test

import (
	"context"
	"testing"

	"Sixth_world_Suday/internal/repository/repotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVanityRoleRepository_List_EmptyByDefault(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)

	// when
	roles, err := repos.VanityRole.List(context.Background())

	// then
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestVanityRoleRepository_Create_AndGetByID(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()

	// when
	err := repos.VanityRole.Create(ctx, "vip", "VIP", "#ff00ff", 5)

	// then
	require.NoError(t, err)
	row, err := repos.VanityRole.GetByID(ctx, "vip")
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, "vip", row.ID)
	assert.Equal(t, "VIP", row.Label)
	assert.Equal(t, "#ff00ff", row.Color)
	assert.False(t, row.IsSystem)
	assert.Equal(t, 5, row.SortOrder)
}

func TestVanityRoleRepository_GetByID_NotFound(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)

	// when
	row, err := repos.VanityRole.GetByID(context.Background(), "missing")

	// then
	require.NoError(t, err)
	assert.Nil(t, row)
}

func TestVanityRoleRepository_Update(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	require.NoError(t, repos.VanityRole.Create(ctx, "mod", "Mod", "#111111", 1))

	// when
	err := repos.VanityRole.Update(ctx, "mod", "Moderator", "#222222", 9)

	// then
	require.NoError(t, err)
	row, err := repos.VanityRole.GetByID(ctx, "mod")
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, "Moderator", row.Label)
	assert.Equal(t, "#222222", row.Color)
	assert.Equal(t, 9, row.SortOrder)
}

func TestVanityRoleRepository_Delete(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	require.NoError(t, repos.VanityRole.Create(ctx, "temp", "Temp", "#000000", 0))

	// when
	err := repos.VanityRole.Delete(ctx, "temp")

	// then
	require.NoError(t, err)
	row, err := repos.VanityRole.GetByID(ctx, "temp")
	require.NoError(t, err)
	assert.Nil(t, row)
}

func TestVanityRoleRepository_List_OrdersBySortOrderThenLabel(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	require.NoError(t, repos.VanityRole.Create(ctx, "b", "Beta", "#000000", 10))
	require.NoError(t, repos.VanityRole.Create(ctx, "a", "Alpha", "#000000", 10))
	require.NoError(t, repos.VanityRole.Create(ctx, "c", "Charlie", "#000000", 5))

	// when
	roles, err := repos.VanityRole.List(ctx)

	// then
	require.NoError(t, err)
	require.Len(t, roles, 3)
	assert.Equal(t, "c", roles[0].ID)
	assert.Equal(t, "a", roles[1].ID)
	assert.Equal(t, "b", roles[2].ID)
}

func TestVanityRoleRepository_AssignToUser_AndGetRolesForUser(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	user := repotest.CreateUser(t, repos)
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))
	require.NoError(t, repos.VanityRole.Create(ctx, "mod", "Mod", "#000", 2))

	// when
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, user.ID, "vip"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, user.ID, "mod"))

	// then
	roles, err := repos.VanityRole.GetRolesForUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 2)
	ids := []string{roles[0].ID, roles[1].ID}
	assert.ElementsMatch(t, []string{"vip", "mod"}, ids)
}

func TestVanityRoleRepository_AssignToUser_DuplicateIgnored(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	user := repotest.CreateUser(t, repos)
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))

	// when
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, user.ID, "vip"))
	err := repos.VanityRole.AssignToUser(ctx, user.ID, "vip")

	// then
	require.NoError(t, err)
	roles, err := repos.VanityRole.GetRolesForUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
}

func TestVanityRoleRepository_UnassignFromUser(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	user := repotest.CreateUser(t, repos)
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))
	require.NoError(t, repos.VanityRole.Create(ctx, "mod", "Mod", "#000", 2))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, user.ID, "vip"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, user.ID, "mod"))

	// when
	err := repos.VanityRole.UnassignFromUser(ctx, user.ID, "vip")

	// then
	require.NoError(t, err)
	roles, err := repos.VanityRole.GetRolesForUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, "mod", roles[0].ID)
}

func TestVanityRoleRepository_GetRolesForUser_Empty(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	user := repotest.CreateUser(t, repos)

	// when
	roles, err := repos.VanityRole.GetRolesForUser(context.Background(), user.ID)

	// then
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestVanityRoleRepository_GetUsersForRole_PaginatesAndOrders(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	alice := repotest.CreateUser(t, repos, repotest.WithUsername("alice"), repotest.WithDisplayName("Alice"))
	bob := repotest.CreateUser(t, repos, repotest.WithUsername("bob"), repotest.WithDisplayName("Bob"))
	carol := repotest.CreateUser(t, repos, repotest.WithUsername("carol"), repotest.WithDisplayName("Carol"))
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, alice.ID, "vip"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, bob.ID, "vip"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, carol.ID, "vip"))

	// when
	page1, total, err := repos.VanityRole.GetUsersForRole(ctx, "vip", "", 2, 0)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	require.Len(t, page1, 2)
	assert.Equal(t, "Alice", page1[0].DisplayName)
	assert.Equal(t, "Bob", page1[1].DisplayName)

	page2, _, err := repos.VanityRole.GetUsersForRole(ctx, "vip", "", 2, 2)
	require.NoError(t, err)
	require.Len(t, page2, 1)
	assert.Equal(t, "Carol", page2[0].DisplayName)
}

func TestVanityRoleRepository_GetUsersForRole_SearchFilter(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	alice := repotest.CreateUser(t, repos, repotest.WithUsername("alice123"), repotest.WithDisplayName("Alice"))
	bob := repotest.CreateUser(t, repos, repotest.WithUsername("bob"), repotest.WithDisplayName("Bobson"))
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, alice.ID, "vip"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, bob.ID, "vip"))

	// when
	users, total, err := repos.VanityRole.GetUsersForRole(ctx, "vip", "ali", 10, 0)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, users, 1)
	assert.Equal(t, alice.ID, users[0].UserID)
	assert.Equal(t, "alice123", users[0].Username)
}

func TestVanityRoleRepository_GetUsersForRole_NoResults(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))

	// when
	users, total, err := repos.VanityRole.GetUsersForRole(ctx, "vip", "", 10, 0)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, users)
}

func TestVanityRoleRepository_GetAllAssignments(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)
	ctx := context.Background()
	a := repotest.CreateUser(t, repos)
	b := repotest.CreateUser(t, repos)
	require.NoError(t, repos.VanityRole.Create(ctx, "vip", "VIP", "#fff", 1))
	require.NoError(t, repos.VanityRole.Create(ctx, "mod", "Mod", "#000", 2))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, a.ID, "vip"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, a.ID, "mod"))
	require.NoError(t, repos.VanityRole.AssignToUser(ctx, b.ID, "mod"))

	// when
	assignments, err := repos.VanityRole.GetAllAssignments(ctx)

	// then
	require.NoError(t, err)
	require.Len(t, assignments, 2)
	assert.ElementsMatch(t, []string{"vip", "mod"}, assignments[a.ID.String()])
	assert.ElementsMatch(t, []string{"mod"}, assignments[b.ID.String()])
}

func TestVanityRoleRepository_GetAllAssignments_Empty(t *testing.T) {
	// given
	repos := repotest.NewRepos(t)

	// when
	assignments, err := repos.VanityRole.GetAllAssignments(context.Background())

	// then
	require.NoError(t, err)
	assert.Empty(t, assignments)
}
