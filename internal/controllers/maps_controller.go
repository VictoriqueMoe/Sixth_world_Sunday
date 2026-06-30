package controllers

import (
	"errors"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/controllers/utils"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/maps"

	"github.com/gofiber/fiber/v3"
)

func (s *Service) getAllMapRoutes() []FSetupRoute {
	return []FSetupRoute{
		s.setupMapList,
		s.setupMapCreate,
		s.setupMapUpdate,
		s.setupMapDelete,
	}
}

func (s *Service) setupMapList(r fiber.Router) {
	r.Get("/maps", s.mapList)
}

func (s *Service) setupMapCreate(r fiber.Router) {
	r.Post("/maps", s.requirePerm(authz.PermManageMaps), s.mapCreate)
}

func (s *Service) setupMapUpdate(r fiber.Router) {
	r.Put("/maps/:id", s.requirePerm(authz.PermManageMaps), s.mapUpdate)
}

func (s *Service) setupMapDelete(r fiber.Router) {
	r.Delete("/maps/:id", s.requirePerm(authz.PermManageMaps), s.mapDelete)
}

func (s *Service) mapList(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	res, err := s.MapService.List(ctx.Context(), userID)
	if err != nil {
		return handleMapError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) mapCreate(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	req, ok := utils.BindJSON[dto.SaveMapRequest](ctx)
	if !ok {
		return nil
	}

	res, err := s.MapService.Create(ctx.Context(), userID, req)
	if err != nil {
		return handleMapError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (s *Service) mapUpdate(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	req, ok := utils.BindJSON[dto.SaveMapRequest](ctx)
	if !ok {
		return nil
	}

	res, err := s.MapService.Update(ctx.Context(), userID, id, req)
	if err != nil {
		return handleMapError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) mapDelete(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.MapService.Delete(ctx.Context(), userID, id); err != nil {
		return handleMapError(ctx, err)
	}

	return utils.OK(ctx)
}

func handleMapError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, maps.ErrMapNotFound):
		return utils.NotFound(ctx, "map not found")
	case errors.Is(err, maps.ErrInvalidInput):
		return utils.BadRequest(ctx, "that doesn't look like a Google My Maps link")
	default:
		return utils.InternalError(ctx, "map error", err)
	}
}
