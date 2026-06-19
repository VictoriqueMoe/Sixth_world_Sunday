-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_rooms ADD COLUMN channel_kind text NOT NULL DEFAULT 'text';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE chat_rooms ADD CONSTRAINT chat_rooms_channel_kind_check CHECK (channel_kind IN ('text', 'voice'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_rooms DROP CONSTRAINT IF EXISTS chat_rooms_channel_kind_check;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE chat_rooms DROP COLUMN IF EXISTS channel_kind;
-- +goose StatementEnd
