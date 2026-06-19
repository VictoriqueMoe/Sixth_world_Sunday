package main

import (
	"context"
	"time"

	"Sixth_world_Suday/internal/email"
	"Sixth_world_Suday/internal/logger"
	"Sixth_world_Suday/internal/middleware"
	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/settings"
	"Sixth_world_Suday/internal/telemetry"
	"Sixth_world_Suday/internal/upload"

	"github.com/gofiber/fiber/v3"
)

func registerListeners(settingsSvc settings.Service, app *fiber.App, svc *services, repos *repository.Repositories) {
	settingsSvc.Subscribe(logger.NewSettingsListener())
	settingsSvc.Subscribe(telemetry.NewSettingsListener())
	settingsSvc.Subscribe(telemetry.NewProfilingSettingsListener())
	settingsSvc.Subscribe(middleware.NewBodyLimitListener(app))
	settingsSvc.Subscribe(email.NewMailSettingListener(svc.email))

	if err := svc.chat.EnsureSystemRooms(context.Background()); err != nil {
		logger.Log.Error().Err(err).Msg("ensure system chat rooms at startup")
	}

	uploadDir := svc.upload.GetUploadDir()

	scheduleJob("clean orphaned uploads", "cleaned orphaned upload files", 24*time.Hour, func() (int, error) {
		return upload.CleanOrphanedFiles(repos.Upload, uploadDir), nil
	})
	scheduleJob("archive stale chat rooms", "archived stale chat rooms", time.Hour, func() (int, error) {
		return svc.chat.ArchiveStale(context.Background())
	})
	scheduleJob("reconcile voice presence", "reconciled voice presence", 30*time.Second, func() (int, error) {
		return svc.chat.ReconcilePresence(context.Background())
	})
}

func scheduleJob(name string, successMsg string, interval time.Duration, fn func() (int, error)) {
	logger.Log.Info().Str("interval", interval.String()).Msgf("registered job: %s", name)
	go func() {
		run := func() {
			n, err := fn()
			if err != nil {
				logger.Log.Error().Err(err).Msgf("%s failed", name)
				return
			}
			if n > 0 {
				logger.Log.Info().Int("count", n).Msg(successMsg)
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}
