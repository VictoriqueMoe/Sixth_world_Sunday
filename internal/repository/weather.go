package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type (
	WeatherLocation struct {
		ID        uuid.UUID
		UserID    uuid.UUID
		Label     string
		PlaceName string
		Country   string
		Admin1    string
		Latitude  float64
		Longitude float64
		Timezone  string
		IsDefault bool
		CreatedAt time.Time
	}

	WeatherRepository interface {
		ListByUser(ctx context.Context, userID uuid.UUID) ([]WeatherLocation, error)
		Get(ctx context.Context, id uuid.UUID) (*WeatherLocation, error)
		Create(ctx context.Context, l *WeatherLocation) error
		UpdateLabel(ctx context.Context, id uuid.UUID, label string) error
		Delete(ctx context.Context, id uuid.UUID) error
		SetDefault(ctx context.Context, userID, id uuid.UUID) error
		CountForUser(ctx context.Context, userID uuid.UUID) (int, error)
	}

	weatherRepository struct {
		db *sql.DB
	}
)

const weatherLocationColumns = `id, user_id, label, place_name, country, admin1, latitude, longitude, timezone, is_default, created_at`

func scanWeatherLocation(s interface{ Scan(...any) error }) (WeatherLocation, error) {
	var l WeatherLocation
	if err := s.Scan(&l.ID, &l.UserID, &l.Label, &l.PlaceName, &l.Country, &l.Admin1,
		&l.Latitude, &l.Longitude, &l.Timezone, &l.IsDefault, &l.CreatedAt); err != nil {
		return WeatherLocation{}, err
	}

	return l, nil
}

func (r *weatherRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]WeatherLocation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+weatherLocationColumns+` FROM weather_locations
		 WHERE user_id = $1 ORDER BY is_default DESC, created_at ASC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list weather locations: %w", err)
	}
	defer rows.Close()

	var out []WeatherLocation
	for rows.Next() {
		l, err := scanWeatherLocation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan weather location: %w", err)
		}
		out = append(out, l)
	}

	return out, rows.Err()
}

func (r *weatherRepository) Get(ctx context.Context, id uuid.UUID) (*WeatherLocation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+weatherLocationColumns+` FROM weather_locations WHERE id = $1`, id,
	)

	l, err := scanWeatherLocation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get weather location: %w", err)
	}

	return &l, nil
}

func (r *weatherRepository) Create(ctx context.Context, l *WeatherLocation) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO weather_locations (id, user_id, label, place_name, country, admin1, latitude, longitude, timezone, is_default)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING created_at`,
		l.ID, l.UserID, l.Label, l.PlaceName, l.Country, l.Admin1, l.Latitude, l.Longitude, l.Timezone, l.IsDefault,
	).Scan(&l.CreatedAt)
	if err != nil {
		return fmt.Errorf("create weather location: %w", err)
	}

	return nil
}

func (r *weatherRepository) UpdateLabel(ctx context.Context, id uuid.UUID, label string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE weather_locations SET label = $2 WHERE id = $1`, id, label)
	if err != nil {
		return fmt.Errorf("update weather location label: %w", err)
	}

	return nil
}

func (r *weatherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM weather_locations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete weather location: %w", err)
	}

	return nil
}

func (r *weatherRepository) SetDefault(ctx context.Context, userID, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE weather_locations SET is_default = (id = $2) WHERE user_id = $1`, userID, id,
	)
	if err != nil {
		return fmt.Errorf("set default weather location: %w", err)
	}

	return nil
}

func (r *weatherRepository) CountForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM weather_locations WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count weather locations: %w", err)
	}

	return count, nil
}
