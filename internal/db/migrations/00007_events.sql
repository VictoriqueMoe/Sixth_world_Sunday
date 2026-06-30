-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id            uuid PRIMARY KEY,
    title         text NOT NULL,
    description   text NOT NULL DEFAULT '',
    cover_url     text NOT NULL DEFAULT '',
    location_type text NOT NULL CHECK (location_type IN ('voice', 'external')),
    voice_room_id uuid REFERENCES chat_rooms(id) ON DELETE SET NULL,
    external_url  text NOT NULL DEFAULT '',
    start_at      timestamptz NOT NULL,
    frequency     text NOT NULL DEFAULT 'none' CHECK (frequency IN ('none', 'weekly', 'biweekly', 'monthly', 'annually')),
    created_by    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    canceled_at   timestamptz
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_events_start_at ON events(start_at);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_events_created_by ON events(created_by);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE event_rsvps (
    event_id   uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, user_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_event_rsvps_event_id ON event_rsvps(event_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event_rsvps;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
