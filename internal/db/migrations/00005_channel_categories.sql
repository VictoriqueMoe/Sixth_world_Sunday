-- +goose Up
-- +goose StatementBegin
CREATE TABLE chat_categories (
    id         uuid PRIMARY KEY,
    name       text NOT NULL,
    position   integer NOT NULL DEFAULT 0,
    is_builtin boolean NOT NULL DEFAULT false,
    kind       text,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO chat_categories (id, name, position, is_builtin, kind) VALUES
    ('11111111-1111-1111-1111-111111111111', 'text channels', 0, true, 'text'),
    ('22222222-2222-2222-2222-222222222222', 'voice channels', 1, true, 'voice');
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE chat_rooms ADD COLUMN category_id uuid REFERENCES chat_categories(id) ON DELETE SET NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE chat_rooms ADD COLUMN position integer NOT NULL DEFAULT 0;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE chat_rooms cr
SET category_id = '11111111-1111-1111-1111-111111111111',
    position = sub.rn
FROM (
    SELECT id, (row_number() OVER (ORDER BY created_at ASC) - 1) AS rn
    FROM chat_rooms
    WHERE type = 'group' AND channel_kind = 'text'
) sub
WHERE cr.id = sub.id;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE chat_rooms cr
SET category_id = '22222222-2222-2222-2222-222222222222',
    position = sub.rn
FROM (
    SELECT id, (row_number() OVER (ORDER BY created_at ASC) - 1) AS rn
    FROM chat_rooms
    WHERE type = 'group' AND channel_kind = 'voice'
) sub
WHERE cr.id = sub.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_rooms DROP COLUMN IF EXISTS position;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE chat_rooms DROP COLUMN IF EXISTS category_id;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS chat_categories;
-- +goose StatementEnd
