package maps

import (
	"context"
	"testing"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validMapURL = "https://www.google.com/maps/d/u/1/edit?mid=1tBY_ooVYpgtN7ucNmm2KtXTjIPCj5H8&ll=41.79780952184883,-87.80200720012422&z=11"

func TestCreateValidMyMapsURL(t *testing.T) {
	repo := repository.NewMockMapRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := NewService(repo, authzSvc)
	user := uuid.New()

	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)
	authzSvc.EXPECT().Can(mock.Anything, user, authz.PermManageMaps).Return(true)

	resp, err := svc.Create(context.Background(), user, dto.SaveMapRequest{Title: "Chicago", SourceURL: validMapURL})

	require.NoError(t, err)
	assert.True(t, resp.CanManage)
	assert.Contains(t, resp.EmbedURL, "https://www.google.com/maps/d/embed?")
	assert.Contains(t, resp.EmbedURL, "mid=1tBY_ooVYpgtN7ucNmm2KtXTjIPCj5H8")
	assert.Contains(t, resp.EmbedURL, "z=11")
}

func TestCreateRejectsBadURLs(t *testing.T) {
	cases := map[string]string{
		"empty":           "",
		"wrong host":      "https://evil.example/maps/d/edit?mid=1tBY_ooVYpgtN7ucNmm2KtXTjIPCj5H8",
		"not a my-map":    "https://www.google.com/search?q=maps",
		"missing mid":     "https://www.google.com/maps/d/edit?mid=",
		"mid too short":   "https://www.google.com/maps/d/edit?mid=short",
		"javascript href": "javascript:alert(1)",
	}

	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			repo := repository.NewMockMapRepository(t)
			authzSvc := authz.NewMockService(t)
			svc := NewService(repo, authzSvc)

			_, err := svc.Create(context.Background(), uuid.New(), dto.SaveMapRequest{SourceURL: raw})

			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestUpdateMissingReturnsNotFound(t *testing.T) {
	repo := repository.NewMockMapRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := NewService(repo, authzSvc)
	id := uuid.New()

	repo.EXPECT().Get(mock.Anything, id).Return(nil, nil)

	_, err := svc.Update(context.Background(), uuid.New(), id, dto.SaveMapRequest{SourceURL: validMapURL})

	assert.ErrorIs(t, err, ErrMapNotFound)
}

func TestDeleteMissingReturnsNotFound(t *testing.T) {
	repo := repository.NewMockMapRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := NewService(repo, authzSvc)
	id := uuid.New()

	repo.EXPECT().Get(mock.Anything, id).Return(nil, nil)

	err := svc.Delete(context.Background(), uuid.New(), id)

	assert.ErrorIs(t, err, ErrMapNotFound)
}
