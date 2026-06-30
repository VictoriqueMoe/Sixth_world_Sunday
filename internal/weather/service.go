package weather

import (
	"context"
	"strings"
	"time"

	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/repository"

	"github.com/google/uuid"
)

const maxSavedLocations = 12

type (
	Service interface {
		List(ctx context.Context, userID uuid.UUID) (*dto.WeatherLocationListResponse, error)
		Save(ctx context.Context, userID uuid.UUID, req dto.SaveWeatherLocationRequest) (*dto.WeatherLocationResponse, error)
		Rename(ctx context.Context, userID, id uuid.UUID, label string) (*dto.WeatherLocationResponse, error)
		SetDefault(ctx context.Context, userID, id uuid.UUID) error
		Delete(ctx context.Context, userID, id uuid.UUID) error
	}

	service struct {
		repo repository.WeatherRepository
	}
)

func NewService(repo repository.WeatherRepository) Service {
	return &service{repo: repo}
}

func (s *service) owned(ctx context.Context, userID, id uuid.UUID) (*repository.WeatherLocation, error) {
	loc, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, ErrLocationNotFound
	}

	if loc.UserID != userID {
		return nil, ErrForbidden
	}

	return loc, nil
}

func (s *service) List(ctx context.Context, userID uuid.UUID) (*dto.WeatherLocationListResponse, error) {
	locations, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.WeatherLocationResponse, len(locations))
	for i := 0; i < len(locations); i++ {
		out[i] = locationToDTO(locations[i])
	}

	return &dto.WeatherLocationListResponse{Locations: out}, nil
}

func (s *service) Save(ctx context.Context, userID uuid.UUID, req dto.SaveWeatherLocationRequest) (*dto.WeatherLocationResponse, error) {
	placeName := strings.TrimSpace(req.PlaceName)
	if placeName == "" {
		return nil, ErrInvalidInput
	}

	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return nil, ErrInvalidInput
	}

	count, err := s.repo.CountForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= maxSavedLocations {
		return nil, ErrTooMany
	}

	location := &repository.WeatherLocation{
		ID:        uuid.New(),
		UserID:    userID,
		Label:     strings.TrimSpace(req.Label),
		PlaceName: placeName,
		Country:   strings.TrimSpace(req.Country),
		Admin1:    strings.TrimSpace(req.Admin1),
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timezone:  strings.TrimSpace(req.Timezone),
		IsDefault: count == 0,
	}
	if err := s.repo.Create(ctx, location); err != nil {
		return nil, err
	}

	return new(locationToDTO(*location)), nil
}

func (s *service) Rename(ctx context.Context, userID, id uuid.UUID, label string) (*dto.WeatherLocationResponse, error) {
	location, err := s.owned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	clean := strings.TrimSpace(label)
	if err := s.repo.UpdateLabel(ctx, id, clean); err != nil {
		return nil, err
	}

	location.Label = clean
	return new(locationToDTO(*location)), nil
}

func (s *service) SetDefault(ctx context.Context, userID, id uuid.UUID) error {
	if _, err := s.owned(ctx, userID, id); err != nil {
		return err
	}

	return s.repo.SetDefault(ctx, userID, id)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	location, err := s.owned(ctx, userID, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if !location.IsDefault {
		return nil
	}

	remaining, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return s.repo.SetDefault(ctx, userID, remaining[0].ID)
	}

	return nil
}

func locationToDTO(l repository.WeatherLocation) dto.WeatherLocationResponse {
	return dto.WeatherLocationResponse{
		ID:        l.ID.String(),
		Label:     l.Label,
		PlaceName: l.PlaceName,
		Country:   l.Country,
		Admin1:    l.Admin1,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Timezone:  l.Timezone,
		IsDefault: l.IsDefault,
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
	}
}
