package controllers

import (
	"github.com/gofiber/fiber/v3"

	"Sixth_world_Suday/internal/controllers/utils"
	"Sixth_world_Suday/internal/middleware"
)

func (s *Service) getAllUserPreferencesRoutes() []FSetupRoute {
	return []FSetupRoute{
		s.setupUpdateAppearance,
	}
}

func (s *Service) setupUpdateAppearance(r fiber.Router) {
	r.Put("/preferences/appearance", middleware.RequireAuth(s.AuthSession, s.AuthzService), s.updateAppearance)
}

func (s *Service) updateAppearance(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)
	var req struct {
		Theme      string `json:"theme"`
		Font       string `json:"font"`
		WideLayout bool   `json:"wide_layout"`
	}
	if err := ctx.Bind().JSON(&req); err != nil {
		return utils.BadRequest(ctx, "invalid request")
	}
	if err := s.UserService.UpdateAppearance(ctx.Context(), userID, req.Theme, req.Font, req.WideLayout); err != nil {
		return utils.InternalError(ctx, "failed to save")
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
