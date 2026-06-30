package controllers

import (
	"errors"

	"Sixth_world_Sunday/internal/controllers/utils"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/weather"

	"github.com/gofiber/fiber/v3"
)

func (s *Service) getAllWeatherRoutes() []FSetupRoute {
	return []FSetupRoute{
		s.setupWeatherListLocations,
		s.setupWeatherSaveLocation,
		s.setupWeatherRenameLocation,
		s.setupWeatherDeleteLocation,
		s.setupWeatherSetDefaultLocation,
	}
}

func (s *Service) setupWeatherListLocations(r fiber.Router) {
	r.Get("/weather/locations", s.weatherListLocations)
}

func (s *Service) setupWeatherSaveLocation(r fiber.Router) {
	r.Post("/weather/locations", s.weatherSaveLocation)
}

func (s *Service) setupWeatherRenameLocation(r fiber.Router) {
	r.Put("/weather/locations/:id", s.weatherRenameLocation)
}

func (s *Service) setupWeatherDeleteLocation(r fiber.Router) {
	r.Delete("/weather/locations/:id", s.weatherDeleteLocation)
}

func (s *Service) setupWeatherSetDefaultLocation(r fiber.Router) {
	r.Post("/weather/locations/:id/default", s.weatherSetDefaultLocation)
}

func (s *Service) weatherListLocations(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	res, err := s.WeatherService.List(ctx.Context(), userID)
	if err != nil {
		return handleWeatherError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) weatherSaveLocation(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	req, ok := utils.BindJSON[dto.SaveWeatherLocationRequest](ctx)
	if !ok {
		return nil
	}

	res, err := s.WeatherService.Save(ctx.Context(), userID, req)
	if err != nil {
		return handleWeatherError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (s *Service) weatherRenameLocation(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	req, ok := utils.BindJSON[dto.RenameWeatherLocationRequest](ctx)
	if !ok {
		return nil
	}

	res, err := s.WeatherService.Rename(ctx.Context(), userID, id, req.Label)
	if err != nil {
		return handleWeatherError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) weatherDeleteLocation(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.WeatherService.Delete(ctx.Context(), userID, id); err != nil {
		return handleWeatherError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) weatherSetDefaultLocation(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.WeatherService.SetDefault(ctx.Context(), userID, id); err != nil {
		return handleWeatherError(ctx, err)
	}

	return utils.OK(ctx)
}

func handleWeatherError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, weather.ErrLocationNotFound):
		return utils.NotFound(ctx, "weather location not found")
	case errors.Is(err, weather.ErrForbidden):
		return utils.Forbidden(ctx, "you cannot modify this location")
	case errors.Is(err, weather.ErrInvalidInput):
		return utils.BadRequest(ctx, "invalid location details")
	case errors.Is(err, weather.ErrTooMany):
		return utils.BadRequest(ctx, "you have reached the maximum number of saved locations")
	default:
		return utils.InternalError(ctx, "weather error", err)
	}
}
