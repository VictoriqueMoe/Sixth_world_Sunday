package weather

import (
	"context"
	"testing"

	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func saveReq(name string, lat, lon float64) dto.SaveWeatherLocationRequest {
	return dto.SaveWeatherLocationRequest{PlaceName: name, Latitude: lat, Longitude: lon}
}

func TestSaveAutoDefaultsFirstLocation(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	user := uuid.New()

	repo.EXPECT().CountForUser(mock.Anything, user).Return(0, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(l *repository.WeatherLocation) bool { return l.IsDefault })).
		Return(nil)

	resp, err := svc.Save(context.Background(), user, saveReq("Home", 51.5, -0.12))

	require.NoError(t, err)
	assert.True(t, resp.IsDefault)
}

func TestSaveSecondLocationNotDefault(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	user := uuid.New()

	repo.EXPECT().CountForUser(mock.Anything, user).Return(1, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(l *repository.WeatherLocation) bool { return !l.IsDefault })).
		Return(nil)

	resp, err := svc.Save(context.Background(), user, saveReq("Work", 52, 1))

	require.NoError(t, err)
	assert.False(t, resp.IsDefault)
}

func TestSaveEnforcesCap(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	user := uuid.New()

	repo.EXPECT().CountForUser(mock.Anything, user).Return(maxSavedLocations, nil)

	_, err := svc.Save(context.Background(), user, saveReq("Nope", 10, 10))

	assert.ErrorIs(t, err, ErrTooMany)
}

func TestSaveRejectsInvalidInput(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	user := uuid.New()

	_, blankErr := svc.Save(context.Background(), user, saveReq("  ", 10, 10))
	assert.ErrorIs(t, blankErr, ErrInvalidInput)

	_, rangeErr := svc.Save(context.Background(), user, saveReq("X", 200, 10))
	assert.ErrorIs(t, rangeErr, ErrInvalidInput)
}

func TestDeleteRejectsNonOwner(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	owner := uuid.New()
	other := uuid.New()
	id := uuid.New()

	repo.EXPECT().Get(mock.Anything, id).Return(&repository.WeatherLocation{ID: id, UserID: owner}, nil)

	err := svc.Delete(context.Background(), other, id)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestDeleteMissingReturnsNotFound(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	user := uuid.New()
	id := uuid.New()

	repo.EXPECT().Get(mock.Anything, id).Return(nil, nil)

	err := svc.Delete(context.Background(), user, id)

	assert.ErrorIs(t, err, ErrLocationNotFound)
}

func TestDeletingDefaultPromotesAnother(t *testing.T) {
	repo := repository.NewMockWeatherRepository(t)
	svc := NewService(repo)
	user := uuid.New()
	deleted := uuid.New()
	promoted := uuid.New()

	repo.EXPECT().Get(mock.Anything, deleted).
		Return(&repository.WeatherLocation{ID: deleted, UserID: user, IsDefault: true}, nil)
	repo.EXPECT().Delete(mock.Anything, deleted).Return(nil)
	repo.EXPECT().ListByUser(mock.Anything, user).
		Return([]repository.WeatherLocation{{ID: promoted, UserID: user}}, nil)
	repo.EXPECT().SetDefault(mock.Anything, user, promoted).Return(nil)

	err := svc.Delete(context.Background(), user, deleted)

	require.NoError(t, err)
}
