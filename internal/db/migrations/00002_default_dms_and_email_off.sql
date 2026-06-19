-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN dms_enabled SET DEFAULT false;
ALTER TABLE users ALTER COLUMN email_notifications SET DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN dms_enabled SET DEFAULT true;
ALTER TABLE users ALTER COLUMN email_notifications SET DEFAULT true;
-- +goose StatementEnd
