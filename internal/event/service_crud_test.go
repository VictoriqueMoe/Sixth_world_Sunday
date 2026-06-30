package event

import (
	"context"
	"testing"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/repository"
	"Sixth_world_Sunday/internal/ws"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newEventTestService(repo repository.EventRepository, authzSvc authz.Service) Service {
	return NewService(repo, authzSvc, nil, nil, ws.NewHub())
}

func voiceFields(roomID string) dto.EventFields {
	return dto.EventFields{
		Title:        "Mob War",
		LocationType: "voice",
		VoiceRoomID:  &roomID,
		StartAt:      "2026-07-05T20:00:00Z",
		Frequency:    "biweekly",
	}
}

func expectRSVPLookups(repo *repository.MockEventRepository) {
	repo.EXPECT().RSVPCounts(mock.Anything, mock.Anything).Return(map[uuid.UUID]int{}, nil)
	repo.EXPECT().ViewerRSVPed(mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]bool{}, nil)
	repo.EXPECT().RSVPAvatars(mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID][]string{}, nil)
}

func TestCreateValidation(t *testing.T) {
	user := uuid.New()
	roomID := uuid.New().String()

	cases := []struct {
		name   string
		fields dto.EventFields
	}{
		{"blank title", func() dto.EventFields { f := voiceFields(roomID); f.Title = "  "; return f }()},
		{"voice without room", dto.EventFields{Title: "X", LocationType: "voice", StartAt: "2026-07-05T20:00:00Z"}},
		{"external without url", dto.EventFields{Title: "X", LocationType: "external", StartAt: "2026-07-05T20:00:00Z"}},
		{"unknown location", dto.EventFields{Title: "X", LocationType: "moon", StartAt: "2026-07-05T20:00:00Z"}},
		{"bad frequency", func() dto.EventFields { f := voiceFields(roomID); f.Frequency = "hourly"; return f }()},
		{"bad start", func() dto.EventFields { f := voiceFields(roomID); f.StartAt = "nope"; return f }()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repository.NewMockEventRepository(t)
			authzSvc := authz.NewMockService(t)
			svc := newEventTestService(repo, authzSvc)

			_, err := svc.Create(context.Background(), user, dto.CreateEventRequest{EventFields: tc.fields})

			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestCreateValidVoiceEvent(t *testing.T) {
	repo := repository.NewMockEventRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := newEventTestService(repo, authzSvc)
	user := uuid.New()
	roomID := uuid.New().String()

	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&repository.Event{
		ID:           uuid.New(),
		Title:        "Mob War",
		LocationType: "voice",
		Frequency:    "biweekly",
		StartAt:      mustTime(t, "2026-07-05T20:00:00Z"),
		CreatedBy:    user,
	}, nil)
	expectRSVPLookups(repo)
	authzSvc.EXPECT().Can(mock.Anything, user, authz.PermManageEvents).Return(false)

	resp, err := svc.Create(context.Background(), user, dto.CreateEventRequest{EventFields: voiceFields(roomID)})

	require.NoError(t, err)
	assert.Equal(t, "voice", resp.LocationType)
	assert.Equal(t, "biweekly", resp.Frequency)
	assert.True(t, resp.CanManage)
}

func TestUpdateForbiddenForNonStaff(t *testing.T) {
	repo := repository.NewMockEventRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := newEventTestService(repo, authzSvc)
	owner := uuid.New()
	other := uuid.New()
	id := uuid.New()

	repo.EXPECT().GetByID(mock.Anything, id).Return(&repository.Event{ID: id, CreatedBy: owner}, nil)
	authzSvc.EXPECT().Can(mock.Anything, other, authz.PermManageEvents).Return(false)

	_, err := svc.Update(context.Background(), other, id, dto.UpdateEventRequest{EventFields: voiceFields(uuid.New().String())})

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUpdateAllowedForStaff(t *testing.T) {
	repo := repository.NewMockEventRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := newEventTestService(repo, authzSvc)
	owner := uuid.New()
	staff := uuid.New()
	id := uuid.New()

	stored := &repository.Event{
		ID:           id,
		CreatedBy:    owner,
		Title:        "Renamed",
		LocationType: "voice",
		Frequency:    "biweekly",
		StartAt:      mustTime(t, "2026-07-05T20:00:00Z"),
	}
	repo.EXPECT().GetByID(mock.Anything, id).Return(stored, nil)
	authzSvc.EXPECT().Can(mock.Anything, staff, authz.PermManageEvents).Return(true)
	repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
	expectRSVPLookups(repo)

	fields := voiceFields(uuid.New().String())
	fields.Title = "Renamed"
	resp, err := svc.Update(context.Background(), staff, id, dto.UpdateEventRequest{EventFields: fields})

	require.NoError(t, err)
	assert.Equal(t, "Renamed", resp.Title)
}

func TestDeleteMissingReturnsNotFound(t *testing.T) {
	repo := repository.NewMockEventRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := newEventTestService(repo, authzSvc)
	user := uuid.New()
	id := uuid.New()

	repo.EXPECT().GetByID(mock.Anything, id).Return(nil, nil)

	err := svc.Delete(context.Background(), user, id)

	assert.ErrorIs(t, err, ErrEventNotFound)
}

func TestDeleteForbiddenForNonOwner(t *testing.T) {
	repo := repository.NewMockEventRepository(t)
	authzSvc := authz.NewMockService(t)
	svc := newEventTestService(repo, authzSvc)
	owner := uuid.New()
	other := uuid.New()
	id := uuid.New()

	repo.EXPECT().GetByID(mock.Anything, id).Return(&repository.Event{ID: id, CreatedBy: owner}, nil)
	authzSvc.EXPECT().Can(mock.Anything, other, authz.PermManageEvents).Return(false)

	err := svc.Delete(context.Background(), other, id)

	assert.ErrorIs(t, err, ErrForbidden)
}
