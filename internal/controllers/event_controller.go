package controllers

import (
	"errors"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/controllers/utils"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/event"

	"github.com/gofiber/fiber/v3"
)

func (s *Service) getAllEventRoutes() []FSetupRoute {
	return []FSetupRoute{
		s.setupEventList,
		s.setupEventCreate,
		s.setupEventCover,
		s.setupEventUpdate,
		s.setupEventDelete,
		s.setupEventCancel,
		s.setupEventRSVP,
		s.setupEventUnRSVP,
	}
}

func (s *Service) setupEventList(r fiber.Router) {
	r.Get("/events", s.eventList)
}

func (s *Service) setupEventCreate(r fiber.Router) {
	r.Post("/events", s.requirePerm(authz.PermManageEvents), s.eventCreate)
}

func (s *Service) setupEventCover(r fiber.Router) {
	r.Post("/events/cover", s.requirePerm(authz.PermManageEvents), s.eventCover)
}

func (s *Service) setupEventUpdate(r fiber.Router) {
	r.Put("/events/:id", s.eventUpdate)
}

func (s *Service) setupEventDelete(r fiber.Router) {
	r.Delete("/events/:id", s.eventDelete)
}

func (s *Service) setupEventCancel(r fiber.Router) {
	r.Post("/events/:id/cancel", s.eventCancel)
}

func (s *Service) setupEventRSVP(r fiber.Router) {
	r.Post("/events/:id/rsvp", s.eventRSVP)
}

func (s *Service) setupEventUnRSVP(r fiber.Router) {
	r.Delete("/events/:id/rsvp", s.eventUnRSVP)
}

func (s *Service) eventList(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	res, err := s.EventService.List(ctx.Context(), userID)
	if err != nil {
		return handleEventError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) eventCreate(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	req, ok := utils.BindJSON[dto.CreateEventRequest](ctx)
	if !ok {
		return nil
	}

	res, err := s.EventService.Create(ctx.Context(), userID, req)
	if err != nil {
		return handleEventError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (s *Service) eventCover(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	fileHeader, err := ctx.FormFile("cover")
	if err != nil {
		return utils.BadRequest(ctx, "cover is required")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return utils.BadRequest(ctx, "failed to read cover")
	}
	defer src.Close()

	url, err := s.EventService.UploadCover(ctx.Context(), userID, fileHeader.Size, src)
	if err != nil {
		return handleEventError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"cover_url": url})
}

func (s *Service) eventUpdate(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	req, ok := utils.BindJSON[dto.UpdateEventRequest](ctx)
	if !ok {
		return nil
	}

	res, err := s.EventService.Update(ctx.Context(), userID, id, req)
	if err != nil {
		return handleEventError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) eventDelete(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.EventService.Delete(ctx.Context(), userID, id); err != nil {
		return handleEventError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) eventCancel(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.EventService.Cancel(ctx.Context(), userID, id); err != nil {
		return handleEventError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) eventRSVP(ctx fiber.Ctx) error {
	return s.setEventRSVP(ctx, true)
}

func (s *Service) eventUnRSVP(ctx fiber.Ctx) error {
	return s.setEventRSVP(ctx, false)
}

func (s *Service) setEventRSVP(ctx fiber.Ctx, interested bool) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.EventService.SetRSVP(ctx.Context(), userID, id, interested); err != nil {
		return handleEventError(ctx, err)
	}

	return utils.OK(ctx)
}

func handleEventError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, event.ErrEventNotFound):
		return utils.NotFound(ctx, "event not found")
	case errors.Is(err, event.ErrForbidden):
		return utils.Forbidden(ctx, "you cannot modify this event")
	case errors.Is(err, event.ErrInvalidInput):
		return utils.BadRequest(ctx, "invalid event details")
	default:
		return utils.InternalError(ctx, "event error", err)
	}
}
