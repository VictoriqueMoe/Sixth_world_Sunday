package main

import (
	"context"
	"os"

	"Sixth_world_Suday/internal/config"
	"Sixth_world_Suday/internal/db"
	"Sixth_world_Suday/internal/logger"
	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/settings"
	"Sixth_world_Suday/internal/telemetry"
)

func initDatabase() (*repository.Repositories, settings.Service) {
	if err := telemetry.Init(
		context.Background(),
		"sixth-world-sunday",
		config.Version,
		"",
	); err != nil {
		logger.Log.Warn().Err(err).Msg("otel init failed; traces disabled")
	}

	database, err := db.Open(config.Cfg.PostgresDSN())
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to open database")
	}

	if err := db.Migrate(database); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to run migrations")
	}

	repos := repository.New(database)

	settingsSvc := settings.NewService(repos.Settings)
	if err := settingsSvc.Refresh(context.Background()); err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load settings")
	}

	logger.Init(settingsSvc.Get(context.Background(), config.SettingLogLevel))
	logger.ApplyDSN(settingsSvc.Get(context.Background(), config.SettingSentryDSN))

	if err := telemetry.Apply(settingsSvc.Get(context.Background(), config.SettingOTLPEndpoint)); err != nil {
		logger.Log.Warn().Err(err).Msg("otel apply failed")
	}

	hostname, _ := os.Hostname()
	if err := telemetry.InitProfiling(
		"sixth-world-sunday",
		hostname,
		settingsSvc.Get(context.Background(), config.SettingPyroscopeURL),
	); err != nil {
		logger.Log.Warn().Err(err).Msg("pyroscope init failed; profiling disabled")
	}

	return repos, settingsSvc
}
