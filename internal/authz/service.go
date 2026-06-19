package authz

import (
	"context"

	"Sixth_world_Suday/internal/logger"
	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/role"

	"github.com/google/uuid"
)

type (
	Service interface {
		Can(ctx context.Context, userID uuid.UUID, perm Permission) bool
		GetRole(ctx context.Context, userID uuid.UUID) (role.Role, error)
		GetRoles(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]role.Role, error)
		IsBanned(ctx context.Context, userID uuid.UUID) bool
		IsLocked(ctx context.Context, userID uuid.UUID) bool
		RequiresEmailVerification(ctx context.Context, userID uuid.UUID) bool
	}

	service struct {
		roleRepo repository.RoleRepository
		userRepo repository.UserRepository
	}
)

func NewService(roleRepo repository.RoleRepository, userRepo repository.UserRepository) Service {
	return &service{roleRepo: roleRepo, userRepo: userRepo}
}

func (s *service) IsBanned(ctx context.Context, userID uuid.UUID) bool {
	banned, err := s.userRepo.IsBanned(ctx, userID)
	if err != nil {
		logger.Log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to check ban status")
		return false
	}
	return banned
}

func (s *service) IsLocked(ctx context.Context, userID uuid.UUID) bool {
	locked, err := s.userRepo.IsLocked(ctx, userID)
	if err != nil {
		logger.Log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to check lock status")
		return false
	}
	return locked
}

func (s *service) RequiresEmailVerification(ctx context.Context, userID uuid.UUID) bool {
	blocked, err := s.userRepo.RequiresEmailVerification(ctx, userID)
	if err != nil {
		logger.Log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to check email verification status")
		return false
	}
	return blocked
}

func (s *service) Can(ctx context.Context, userID uuid.UUID, perm Permission) bool {
	if userID == uuid.Nil {
		return false
	}

	r, err := s.roleRepo.GetRole(ctx, userID)
	if err != nil {
		logger.Log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get role for permission check")
		return false
	}
	if r == "" {
		return false
	}

	perms, ok := rolePermissions[r]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == PermAll || p == perm {
			return true
		}
	}
	return false
}

func (s *service) GetRole(ctx context.Context, userID uuid.UUID) (role.Role, error) {
	return s.roleRepo.GetRole(ctx, userID)
}

func (s *service) GetRoles(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]role.Role, error) {
	return s.roleRepo.GetRoles(ctx, userIDs)
}
