package main

import (
	"coin-app/internal/config"
	"coin-app/internal/lib/logger/handlers/slogpretty"
	"coin-app/internal/lib/logger/sl"
	"coin-app/internal/storage/postgresql"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Init config: cleanenv
	cfg := config.MustLoad()

	// Init logger: slog
	log := setupLogger(cfg.Env)
	log.Info("starting driver server", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// Init storage: postgresql
	storage, err := postgresql.New()
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	// TODO: Init router: chi, "chi render"

	// TODO: Run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
