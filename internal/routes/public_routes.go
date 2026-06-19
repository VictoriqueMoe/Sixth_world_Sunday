package routes

import (
	"Sixth_world_Suday/internal/controllers"
	"Sixth_world_Suday/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func PublicRoutes(service controllers.Service, app *fiber.App) {
	apiRoutes := service.GetAPIRoutes()
	api := app.Group("/api/v1")
	api.Use(middleware.RequireAuthGate(service.AuthSession, service.AuthzService))
	for i := 0; i < len(apiRoutes); i++ {
		apiRoutes[i](api)
	}

	app.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status":  "ok",
			"service": "sixth-world-sunday",
		})
	})

	pageRoutes := service.GetPageRoutes()
	for i := 0; i < len(pageRoutes); i++ {
		pageRoutes[i](app)
	}
}
