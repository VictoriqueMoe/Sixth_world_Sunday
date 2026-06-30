-- +goose Up
-- +goose StatementBegin
CREATE TABLE weather_locations (
    id         uuid PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label      text NOT NULL DEFAULT '',
    place_name text NOT NULL,
    country    text NOT NULL DEFAULT '',
    admin1     text NOT NULL DEFAULT '',
    latitude   double precision NOT NULL,
    longitude  double precision NOT NULL,
    timezone   text NOT NULL DEFAULT '',
    is_default boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_weather_locations_user_id ON weather_locations(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS weather_locations;
-- +goose StatementEnd
