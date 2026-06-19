package main

import (
	"os"

	"Sixth_world_Suday/internal/config"
	"Sixth_world_Suday/internal/logger"
	"Sixth_world_Suday/internal/telemetry"
	"Sixth_world_Suday/internal/utils"
)

func main() {
	logger.Init(config.SettingLogLevel.Default)
	defer logger.Shutdown()
	defer telemetry.Shutdown()
	defer telemetry.ShutdownProfiling()

	logger.Log.Info().
		Str("db_host", config.Cfg.Postgres.Host).
		Str("db_name", config.Cfg.Postgres.DB).
		Msg("starting")

	app := initServer()

	addr := ":4323"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	logger.Log.Info().Str("addr", addr).Msg("starting server")
	utils.StartServerWithGracefulShutdown(app, addr)
}
