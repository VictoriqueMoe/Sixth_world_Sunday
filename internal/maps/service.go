package maps

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/repository"

	"github.com/google/uuid"
)

var (
	allowedHosts = map[string]bool{
		"www.google.com":   true,
		"google.com":       true,
		"maps.google.com":  true,
		"www.google.co.uk": true,
	}

	midPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{8,128}$`)
	llPattern   = regexp.MustCompile(`^-?\d{1,3}(\.\d+)?,-?\d{1,3}(\.\d+)?$`)
	zoomPattern = regexp.MustCompile(`^\d{1,2}$`)
)

type (
	Service interface {
		List(ctx context.Context, viewerID uuid.UUID) (*dto.MapListResponse, error)
		Create(ctx context.Context, userID uuid.UUID, req dto.SaveMapRequest) (*dto.MapResponse, error)
		Update(ctx context.Context, userID, id uuid.UUID, req dto.SaveMapRequest) (*dto.MapResponse, error)
		Delete(ctx context.Context, userID, id uuid.UUID) error
	}

	service struct {
		repo  repository.MapRepository
		authz authz.Service
	}
)

func NewService(repo repository.MapRepository, authzSvc authz.Service) Service {
	return &service{repo: repo, authz: authzSvc}
}

func (s *service) canManage(ctx context.Context, userID uuid.UUID) bool {
	return s.authz.Can(ctx, userID, authz.PermManageMaps)
}

func (s *service) List(ctx context.Context, viewerID uuid.UUID) (*dto.MapListResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	canManage := s.canManage(ctx, viewerID)

	out := make([]dto.MapResponse, len(items))
	for i := 0; i < len(items); i++ {
		out[i] = mapToDTO(items[i], canManage)
	}

	return &dto.MapListResponse{Maps: out}, nil
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req dto.SaveMapRequest) (*dto.MapResponse, error) {
	item, err := buildMap(uuid.New(), userID, req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	return new(mapToDTO(*item, s.canManage(ctx, userID))), nil
}

func (s *service) Update(ctx context.Context, userID, id uuid.UUID, req dto.SaveMapRequest) (*dto.MapResponse, error) {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrMapNotFound
	}

	updated, err := buildMap(id, existing.CreatedBy, req)
	if err != nil {
		return nil, err
	}
	updated.CreatedAt = existing.CreatedAt

	if err := s.repo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return new(mapToDTO(*updated, s.canManage(ctx, userID))), nil
}

func (s *service) Delete(ctx context.Context, _ uuid.UUID, id uuid.UUID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrMapNotFound
	}

	return s.repo.Delete(ctx, id)
}

func buildMap(id, createdBy uuid.UUID, req dto.SaveMapRequest) (*repository.Map, error) {
	mid, ll, zoom, err := parseMyMapsURL(req.SourceURL)
	if err != nil {
		return nil, err
	}

	return &repository.Map{
		ID:          id,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		SourceURL:   strings.TrimSpace(req.SourceURL),
		Mid:         mid,
		LL:          ll,
		Zoom:        zoom,
		CreatedBy:   createdBy,
	}, nil
}

func parseMyMapsURL(raw string) (mid, ll, zoom string, err error) {
	u, parseErr := url.Parse(strings.TrimSpace(raw))
	if parseErr != nil {
		return "", "", "", ErrInvalidInput
	}

	if !allowedHosts[strings.ToLower(u.Hostname())] || !strings.Contains(u.Path, "/maps/d/") {
		return "", "", "", ErrInvalidInput
	}

	q := u.Query()

	mid = q.Get("mid")
	if !midPattern.MatchString(mid) {
		return "", "", "", ErrInvalidInput
	}

	if v := q.Get("ll"); llPattern.MatchString(v) {
		ll = v
	}
	if v := q.Get("z"); zoomPattern.MatchString(v) {
		zoom = v
	}

	return mid, ll, zoom, nil
}

func buildEmbedURL(mid, ll, zoom string) string {
	q := url.Values{}
	q.Set("mid", mid)
	if ll != "" {
		q.Set("ll", ll)
	}
	if zoom != "" {
		q.Set("z", zoom)
	}

	return "https://www.google.com/maps/d/embed?" + q.Encode()
}

func mapToDTO(m repository.Map, canManage bool) dto.MapResponse {
	return dto.MapResponse{
		ID:          m.ID.String(),
		Title:       m.Title,
		Description: m.Description,
		EmbedURL:    buildEmbedURL(m.Mid, m.LL, m.Zoom),
		SourceURL:   m.SourceURL,
		CanManage:   canManage,
		CreatedBy:   m.CreatedBy.String(),
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}
