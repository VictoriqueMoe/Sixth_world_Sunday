package event

import (
	"context"
	"io"
	"sort"
	"strings"
	"time"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/config"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/repository"
	"Sixth_world_Sunday/internal/settings"
	"Sixth_world_Sunday/internal/upload"
	"Sixth_world_Sunday/internal/ws"

	"github.com/google/uuid"
)

const (
	occurrenceCount = 5
	avatarStackSize = 3
)

var validFrequencies = map[string]bool{
	"none":     true,
	"weekly":   true,
	"biweekly": true,
	"monthly":  true,
	"annually": true,
}

type (
	Service interface {
		List(ctx context.Context, viewerID uuid.UUID) (*dto.EventListResponse, error)
		Create(ctx context.Context, userID uuid.UUID, req dto.CreateEventRequest) (*dto.EventResponse, error)
		Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateEventRequest) (*dto.EventResponse, error)
		Cancel(ctx context.Context, userID, id uuid.UUID) error
		Delete(ctx context.Context, userID, id uuid.UUID) error
		SetRSVP(ctx context.Context, userID, id uuid.UUID, interested bool) error
		UploadCover(ctx context.Context, userID uuid.UUID, size int64, reader io.Reader) (string, error)
	}

	service struct {
		repo        repository.EventRepository
		authz       authz.Service
		uploadSvc   upload.Service
		settingsSvc settings.Service
		hub         *ws.Hub
	}
)

func NewService(repo repository.EventRepository, authzSvc authz.Service, uploadSvc upload.Service, settingsSvc settings.Service, hub *ws.Hub) Service {
	return &service{
		repo:        repo,
		authz:       authzSvc,
		uploadSvc:   uploadSvc,
		settingsSvc: settingsSvc,
		hub:         hub,
	}
}

func (s *service) broadcast(action string) {
	s.hub.Broadcast(ws.Message{
		Type: "events_changed",
		Data: map[string]any{"action": action},
	})
}

func (s *service) canManage(ctx context.Context, userID uuid.UUID, e *repository.Event) bool {
	if e.CreatedBy == userID {
		return true
	}

	return s.authz.Can(ctx, userID, authz.PermManageEvents)
}

func (s *service) List(ctx context.Context, viewerID uuid.UUID) (*dto.EventListResponse, error) {
	events, err := s.repo.List(ctx, false)
	if err != nil {
		return nil, err
	}

	responses, err := s.buildResponses(ctx, viewerID, events)
	if err != nil {
		return nil, err
	}

	return &dto.EventListResponse{Events: responses}, nil
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req dto.CreateEventRequest) (*dto.EventResponse, error) {
	event, err := buildEvent(uuid.New(), userID, fieldsFromRequest(req.EventFields))
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	s.broadcast("create")

	return s.singleResponse(ctx, userID, event.ID)
}

func (s *service) Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateEventRequest) (*dto.EventResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrEventNotFound
	}

	if !s.canManage(ctx, userID, existing) {
		return nil, ErrForbidden
	}

	updated, err := buildEvent(id, existing.CreatedBy, fieldsFromRequest(req.EventFields))
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, updated); err != nil {
		return nil, err
	}

	s.broadcast("update")

	return s.singleResponse(ctx, userID, id)
}

func (s *service) Cancel(ctx context.Context, userID, id uuid.UUID) error {
	return s.mutateManaged(ctx, userID, id, "cancel", s.repo.Cancel)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.mutateManaged(ctx, userID, id, "delete", s.repo.Delete)
}

func (s *service) mutateManaged(ctx context.Context, userID, id uuid.UUID, action string, fn func(context.Context, uuid.UUID) error) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrEventNotFound
	}

	if !s.canManage(ctx, userID, existing) {
		return ErrForbidden
	}

	if err := fn(ctx, id); err != nil {
		return err
	}

	s.broadcast(action)
	return nil
}

func (s *service) SetRSVP(ctx context.Context, userID, id uuid.UUID, interested bool) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil || existing.CanceledAt != nil {
		return ErrEventNotFound
	}

	if interested {
		err = s.repo.AddRSVP(ctx, id, userID)
	} else {
		err = s.repo.RemoveRSVP(ctx, id, userID)
	}
	if err != nil {
		return err
	}

	s.broadcast("rsvp")
	return nil
}

func (s *service) UploadCover(ctx context.Context, _ uuid.UUID, size int64, reader io.Reader) (string, error) {
	maxSize := int64(s.settingsSvc.GetInt(ctx, config.SettingMaxImageSize))

	url, err := s.uploadSvc.SaveImage(ctx, "events", uuid.New(), size, maxSize, reader)
	if err != nil {
		return "", ErrInvalidInput
	}

	return url, nil
}

func (s *service) singleResponse(ctx context.Context, viewerID, id uuid.UUID) (*dto.EventResponse, error) {
	full, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if full == nil {
		return nil, ErrEventNotFound
	}

	responses, err := s.buildResponses(ctx, viewerID, []repository.Event{*full})
	if err != nil {
		return nil, err
	}

	return &responses[0], nil
}

func (s *service) buildResponses(ctx context.Context, viewerID uuid.UUID, events []repository.Event) ([]dto.EventResponse, error) {
	ids := make([]uuid.UUID, len(events))
	for i := 0; i < len(events); i++ {
		ids[i] = events[i].ID
	}

	counts, err := s.repo.RSVPCounts(ctx, ids)
	if err != nil {
		return nil, err
	}

	viewerRSVP, err := s.repo.ViewerRSVPed(ctx, ids, viewerID)
	if err != nil {
		return nil, err
	}

	avatars, err := s.repo.RSVPAvatars(ctx, ids, avatarStackSize)
	if err != nil {
		return nil, err
	}

	staff := s.authz.Can(ctx, viewerID, authz.PermManageEvents)
	now := time.Now().UTC()

	out := make([]dto.EventResponse, len(events))
	for i := 0; i < len(events); i++ {
		e := events[i]
		occurrences := nextOccurrences(e.StartAt, e.Frequency, now, occurrenceCount)

		nextStart := e.StartAt
		if len(occurrences) > 0 {
			nextStart = occurrences[0]
		}

		occStrings := make([]string, len(occurrences))
		for j := 0; j < len(occurrences); j++ {
			occStrings[j] = occurrences[j].Format(time.RFC3339)
		}

		var voiceRoomID *string
		if e.VoiceRoomID != nil {
			id := e.VoiceRoomID.String()
			voiceRoomID = &id
		}

		avatarList := avatars[e.ID]
		if avatarList == nil {
			avatarList = []string{}
		}

		out[i] = dto.EventResponse{
			ID:               e.ID.String(),
			Title:            e.Title,
			Description:      e.Description,
			CoverURL:         e.CoverURL,
			LocationType:     e.LocationType,
			VoiceRoomID:      voiceRoomID,
			VoiceRoomName:    e.VoiceRoomName,
			ExternalURL:      e.ExternalURL,
			StartAt:          e.StartAt.Format(time.RFC3339),
			Frequency:        e.Frequency,
			NextStartAt:      nextStart.Format(time.RFC3339),
			NextOccurrences:  occStrings,
			RSVPCount:        counts[e.ID],
			ViewerInterested: viewerRSVP[e.ID],
			RSVPAvatars:      avatarList,
			CanManage:        e.CreatedBy == viewerID || staff,
			CreatedBy:        e.CreatedBy.String(),
			CreatedAt:        e.CreatedAt.Format(time.RFC3339),
		}
	}

	sortResponses(out)

	return out, nil
}

// sortResponses orders upcoming events soonest-first, with fully-past events
// after them ordered most-recent-first.
func sortResponses(events []dto.EventResponse) {
	sort.SliceStable(events, func(i, j int) bool {
		iUpcoming := len(events[i].NextOccurrences) > 0
		jUpcoming := len(events[j].NextOccurrences) > 0
		if iUpcoming != jUpcoming {
			return iUpcoming
		}

		if iUpcoming {
			return events[i].NextStartAt < events[j].NextStartAt
		}

		return events[i].StartAt > events[j].StartAt
	})
}

type eventFields struct {
	title        string
	description  string
	coverURL     string
	locationType string
	voiceRoomID  *string
	externalURL  string
	startAt      string
	frequency    string
}

func fieldsFromRequest(f dto.EventFields) eventFields {
	return eventFields{
		title:        f.Title,
		description:  f.Description,
		coverURL:     f.CoverURL,
		locationType: f.LocationType,
		voiceRoomID:  f.VoiceRoomID,
		externalURL:  f.ExternalURL,
		startAt:      f.StartAt,
		frequency:    f.Frequency,
	}
}

func buildEvent(id, createdBy uuid.UUID, f eventFields) (*repository.Event, error) {
	title := strings.TrimSpace(f.title)
	if title == "" {
		return nil, ErrInvalidInput
	}

	frequency := f.frequency
	if frequency == "" {
		frequency = "none"
	}
	if !validFrequencies[frequency] {
		return nil, ErrInvalidInput
	}

	startAt, err := parseStart(f.startAt)
	if err != nil {
		return nil, err
	}

	event := &repository.Event{
		ID:           id,
		Title:        title,
		Description:  strings.TrimSpace(f.description),
		CoverURL:     strings.TrimSpace(f.coverURL),
		LocationType: f.locationType,
		StartAt:      startAt,
		Frequency:    frequency,
		CreatedBy:    createdBy,
	}

	switch f.locationType {
	case "voice":
		if f.voiceRoomID == nil || strings.TrimSpace(*f.voiceRoomID) == "" {
			return nil, ErrInvalidInput
		}
		roomID, err := uuid.Parse(strings.TrimSpace(*f.voiceRoomID))
		if err != nil {
			return nil, ErrInvalidInput
		}
		event.VoiceRoomID = &roomID
	case "external":
		external := strings.TrimSpace(f.externalURL)
		if external == "" {
			return nil, ErrInvalidInput
		}
		event.ExternalURL = external
	default:
		return nil, ErrInvalidInput
	}

	return event, nil
}

func parseStart(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, ErrInvalidInput
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, ErrInvalidInput
}

// nextOccurrences returns up to count occurrence times at or after now, derived
// from the event's start and recurrence frequency.
func nextOccurrences(start time.Time, frequency string, now time.Time, count int) []time.Time {
	if frequency == "none" || frequency == "" {
		if !start.Before(now) {
			return []time.Time{start}
		}
		return nil
	}

	occ := start
	guard := 0
	for occ.Before(now) {
		occ = advance(occ, frequency)
		guard++
		if guard > 100000 {
			return nil
		}
	}

	out := make([]time.Time, 0, count)
	for len(out) < count {
		out = append(out, occ)
		occ = advance(occ, frequency)
	}

	return out
}

func advance(t time.Time, frequency string) time.Time {
	switch frequency {
	case "weekly":
		return t.AddDate(0, 0, 7)
	case "biweekly":
		return t.AddDate(0, 0, 14)
	case "monthly":
		return nextMonthlyNthWeekday(t)
	case "annually":
		return t.AddDate(1, 0, 0)
	default:
		return t.AddDate(0, 0, 7)
	}
}

// nextMonthlyNthWeekday returns the same nth-weekday-of-month (e.g. "first
// Sunday") in the month after t. Months without that nth occurrence fall back
// to the last matching weekday.
func nextMonthlyNthWeekday(t time.Time) time.Time {
	weekday := t.Weekday()
	nth := (t.Day()-1)/7 + 1

	year, month := t.Year(), t.Month()
	month++
	if month > time.December {
		month = time.January
		year++
	}

	first := time.Date(year, month, 1, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	offset := (int(weekday) - int(first.Weekday()) + 7) % 7
	day := 1 + offset + (nth-1)*7

	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()
	for day > daysInMonth {
		day -= 7
	}

	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}
