package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	Event struct {
		ID            uuid.UUID
		Title         string
		Description   string
		CoverURL      string
		LocationType  string
		VoiceRoomID   *uuid.UUID
		VoiceRoomName string
		ExternalURL   string
		StartAt       time.Time
		Frequency     string
		CreatedBy     uuid.UUID
		CreatedAt     time.Time
		UpdatedAt     time.Time
		CanceledAt    *time.Time
	}

	EventRepository interface {
		Create(ctx context.Context, e *Event) error
		GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
		List(ctx context.Context, includeCanceled bool) ([]Event, error)
		Update(ctx context.Context, e *Event) error
		Cancel(ctx context.Context, id uuid.UUID) error
		Delete(ctx context.Context, id uuid.UUID) error

		AddRSVP(ctx context.Context, eventID, userID uuid.UUID) error
		RemoveRSVP(ctx context.Context, eventID, userID uuid.UUID) error
		RSVPCounts(ctx context.Context, eventIDs []uuid.UUID) (map[uuid.UUID]int, error)
		ViewerRSVPed(ctx context.Context, eventIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error)
		RSVPAvatars(ctx context.Context, eventIDs []uuid.UUID, limit int) (map[uuid.UUID][]string, error)
	}

	eventRepository struct {
		db *sql.DB
	}
)

const eventSelect = `SELECT e.id, e.title, e.description, e.cover_url, e.location_type, e.voice_room_id,
	COALESCE(cr.name, ''), e.external_url, e.start_at, e.frequency,
	e.created_by, e.created_at, e.updated_at, e.canceled_at
	FROM events e
	LEFT JOIN chat_rooms cr ON cr.id = e.voice_room_id`

func scanEvent(s interface{ Scan(...any) error }) (Event, error) {
	var e Event
	var voiceRoom uuid.NullUUID
	var canceled sql.NullTime

	if err := s.Scan(&e.ID, &e.Title, &e.Description, &e.CoverURL, &e.LocationType, &voiceRoom,
		&e.VoiceRoomName, &e.ExternalURL, &e.StartAt, &e.Frequency,
		&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt, &canceled); err != nil {
		return Event{}, err
	}

	if voiceRoom.Valid {
		id := voiceRoom.UUID
		e.VoiceRoomID = &id
	}

	if canceled.Valid {
		t := canceled.Time
		e.CanceledAt = &t
	}

	return e, nil
}

func (r *eventRepository) Create(ctx context.Context, e *Event) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO events (id, title, description, cover_url, location_type, voice_room_id, external_url, start_at, frequency, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING created_at, updated_at`,
		e.ID, e.Title, e.Description, e.CoverURL, e.LocationType, nullableUUIDValue(e.VoiceRoomID),
		e.ExternalURL, e.StartAt, e.Frequency, e.CreatedBy,
	).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	return nil
}

func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	row := r.db.QueryRowContext(ctx, eventSelect+` WHERE e.id = $1`, id)

	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	return &e, nil
}

func (r *eventRepository) List(ctx context.Context, includeCanceled bool) ([]Event, error) {
	query := eventSelect
	if !includeCanceled {
		query += ` WHERE e.canceled_at IS NULL`
	}
	query += ` ORDER BY e.start_at`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var out []Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

func (r *eventRepository) Update(ctx context.Context, e *Event) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE events SET title = $2, description = $3, cover_url = $4, location_type = $5,
		 voice_room_id = $6, external_url = $7, start_at = $8, frequency = $9, updated_at = now()
		 WHERE id = $1`,
		e.ID, e.Title, e.Description, e.CoverURL, e.LocationType, nullableUUIDValue(e.VoiceRoomID),
		e.ExternalURL, e.StartAt, e.Frequency,
	)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

func (r *eventRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE events SET canceled_at = now(), updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("cancel event: %w", err)
	}

	return nil
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	return nil
}

func (r *eventRepository) AddRSVP(ctx context.Context, eventID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO event_rsvps (event_id, user_id) VALUES ($1, $2) ON CONFLICT (event_id, user_id) DO NOTHING`,
		eventID, userID,
	)
	if err != nil {
		return fmt.Errorf("add rsvp: %w", err)
	}

	return nil
}

func (r *eventRepository) RemoveRSVP(ctx context.Context, eventID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM event_rsvps WHERE event_id = $1 AND user_id = $2`, eventID, userID)
	if err != nil {
		return fmt.Errorf("remove rsvp: %w", err)
	}

	return nil
}

func (r *eventRepository) RSVPCounts(ctx context.Context, eventIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	out := make(map[uuid.UUID]int, len(eventIDs))
	if len(eventIDs) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(eventIDs))
	args := make([]any, len(eventIDs))
	for i := 0; i < len(eventIDs); i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = eventIDs[i]
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT event_id, COUNT(*) FROM event_rsvps WHERE event_id IN (`+strings.Join(placeholders, ",")+`) GROUP BY event_id`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("rsvp counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, fmt.Errorf("scan rsvp count: %w", err)
		}
		out[id] = count
	}

	return out, rows.Err()
}

func (r *eventRepository) ViewerRSVPed(ctx context.Context, eventIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(eventIDs))
	if len(eventIDs) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(eventIDs))
	args := make([]any, 0, len(eventIDs)+1)
	args = append(args, userID)
	for i := 0; i < len(eventIDs); i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, eventIDs[i])
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT event_id FROM event_rsvps WHERE user_id = $1 AND event_id IN (`+strings.Join(placeholders, ",")+`)`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("viewer rsvped: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan viewer rsvp: %w", err)
		}
		out[id] = true
	}

	return out, rows.Err()
}

func (r *eventRepository) RSVPAvatars(ctx context.Context, eventIDs []uuid.UUID, limit int) (map[uuid.UUID][]string, error) {
	out := make(map[uuid.UUID][]string, len(eventIDs))
	if len(eventIDs) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(eventIDs))
	args := make([]any, 0, len(eventIDs)+1)
	for i := 0; i < len(eventIDs); i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, eventIDs[i])
	}
	args = append(args, limit)

	query := `SELECT event_id, avatar_url FROM (
		SELECT er.event_id, u.avatar_url,
			ROW_NUMBER() OVER (PARTITION BY er.event_id ORDER BY er.created_at) AS rn
		FROM event_rsvps er JOIN users u ON u.id = er.user_id
		WHERE er.event_id IN (` + strings.Join(placeholders, ",") + `) AND u.avatar_url <> ''
	) t WHERE rn <= $` + fmt.Sprintf("%d", len(eventIDs)+1) + ` ORDER BY event_id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("rsvp avatars: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var avatar string
		if err := rows.Scan(&id, &avatar); err != nil {
			return nil, fmt.Errorf("scan rsvp avatar: %w", err)
		}
		out[id] = append(out[id], avatar)
	}

	return out, rows.Err()
}
