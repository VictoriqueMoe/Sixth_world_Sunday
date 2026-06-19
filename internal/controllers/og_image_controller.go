package controllers

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"Sixth_world_Suday/internal/logger"
	"Sixth_world_Suday/internal/media"

	"github.com/gofiber/fiber/v3"
)

type (
	OGImageHandler struct {
		uploadDir string
	}
)

func NewOGImageHandler(uploadDir string) *OGImageHandler {
	return &OGImageHandler{uploadDir: uploadDir}
}

func (h *OGImageHandler) Register(app fiber.Router) {
	app.Get("/og-image/*", h.serve)
}

func (h *OGImageHandler) serve(ctx fiber.Ctx) error {
	rel := ctx.Params("*")
	if !strings.HasSuffix(strings.ToLower(rel), ".jpg") {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	webpRel := rel[:len(rel)-len(".jpg")] + ".webp"
	webpPath := filepath.Join(h.uploadDir, filepath.FromSlash(path.Clean("/"+webpRel)))

	if _, err := os.Stat(webpPath); err == nil {
		data, err := media.WebPToJPEG(ctx.Context(), webpPath)
		if err != nil {
			logger.Log.Warn().Err(err).Str("path", webpPath).Msg("og image conversion failed, serving original webp")
			return ctx.SendFile(webpPath)
		}
		return ctx.Type("jpg").Send(data)
	}

	jpgPath := filepath.Join(h.uploadDir, filepath.FromSlash(path.Clean("/"+rel)))
	if _, err := os.Stat(jpgPath); err == nil {
		return ctx.Type("jpg").SendFile(jpgPath)
	}

	return ctx.SendStatus(fiber.StatusNotFound)
}
