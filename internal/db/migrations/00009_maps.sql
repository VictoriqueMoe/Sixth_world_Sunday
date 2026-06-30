-- +goose Up
-- +goose StatementBegin
CREATE TABLE maps (
    id          uuid PRIMARY KEY,
    title       text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    source_url  text NOT NULL,
    mid         text NOT NULL,
    ll          text NOT NULL DEFAULT '',
    zoom        text NOT NULL DEFAULT '',
    created_by  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS maps;
-- +goose StatementEnd
