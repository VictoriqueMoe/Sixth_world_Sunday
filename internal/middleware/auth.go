package middleware

import (
	"strings"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/session"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func sessionToken(ctx fiber.Ctx) string {
	if bearer := session.BearerToken(ctx.Get("Authorization")); bearer != "" {
		return bearer
	}

	return ctx.Cookies(session.CookieName)
}

func isWriteMethod(method string) bool {
	switch method {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
		return true
	default:
		return false
	}
}

func apiPath(path string) string {
	return strings.TrimPrefix(path, "/api/v1")
}

func isVerifyExemptPath(method, path string) bool {
	if method != fiber.MethodPost {
		return false
	}

	path = apiPath(path)
	switch path {
	case "/auth/set-email", "/auth/verify-email", "/auth/resend-verification":
		return true
	case "/notifications/read":
		return true
	}
	if strings.HasPrefix(path, "/notifications/") && strings.HasSuffix(path, "/read") {
		return true
	}
	if strings.HasPrefix(path, "/chat/rooms/") && strings.HasSuffix(path, "/read") {
		return true
	}
	return false
}

func isLockExemptPath(method, path string) bool {
	if method != fiber.MethodPost {
		return false
	}

	path = apiPath(path)
	if strings.HasPrefix(path, "/notifications/") && strings.HasSuffix(path, "/read") {
		return true
	}
	if path == "/notifications/read" {
		return true
	}
	if strings.HasPrefix(path, "/chat/rooms/") && strings.HasSuffix(path, "/read") {
		return true
	}
	if strings.HasPrefix(path, "/chat/rooms/") && strings.HasSuffix(path, "/messages") {
		return true
	}
	return false
}

func RequirePermission(mgr *session.Manager, authzSvc authz.Service, perm authz.Permission) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userID, _, ok := authenticateAndCheckBan(ctx, mgr, authzSvc)
		if !ok {
			return nil
		}

		if !authzSvc.Can(ctx.Context(), userID, perm) {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		ctx.Locals("userID", userID)
		return ctx.Next()
	}
}

func isPublicAPIPath(method, path string) bool {
	switch apiPath(path) {
	case "/site-info":
		return method == fiber.MethodGet
	case "/livekit/webhook":
		return method == fiber.MethodPost
	case "/auth/register",
		"/auth/login",
		"/auth/logout",
		"/auth/forgot-password",
		"/auth/reset-password",
		"/auth/verify-email":
		return method == fiber.MethodPost
	}

	return false
}

func RequireAuthGate(mgr *session.Manager, authzSvc authz.Service) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if isPublicAPIPath(ctx.Method(), ctx.Path()) {
			return ctx.Next()
		}

		userID, _, ok := authenticateAndCheckBan(ctx, mgr, authzSvc)
		if !ok {
			return nil
		}

		ctx.Locals("userID", userID)
		return ctx.Next()
	}
}

// authenticateAndCheckBan validates the session cookie and ban status. On any
// failure it writes the appropriate response and returns ok=false; callers
// must then `return nil` so fiber does not run subsequent handlers.
func authenticateAndCheckBan(ctx fiber.Ctx, mgr *session.Manager, authzSvc authz.Service) (uuid.UUID, string, bool) {
	token := sessionToken(ctx)
	if token == "" {
		_ = ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
		return uuid.Nil, "", false
	}

	userID, err := mgr.Validate(ctx.Context(), token)
	if err != nil {
		_ = ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired session",
		})
		return uuid.Nil, "", false
	}

	if authzSvc.IsBanned(ctx.Context(), userID) {
		mgr.Delete(ctx.Context(), token)
		_ = ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "your account has been banned",
		})
		return uuid.Nil, "", false
	}

	if isWriteMethod(ctx.Method()) && !isLockExemptPath(ctx.Method(), ctx.Path()) {
		if authzSvc.IsLocked(ctx.Context(), userID) {
			_ = ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "your account is locked",
			})
			return uuid.Nil, "", false
		}
	}

	if isWriteMethod(ctx.Method()) && !isVerifyExemptPath(ctx.Method(), ctx.Path()) {
		if authzSvc.RequiresEmailVerification(ctx.Context(), userID) {
			_ = ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "please verify your email address to continue",
				"code":  "email_unverified",
			})
			return uuid.Nil, "", false
		}
	}

	return userID, token, true
}
