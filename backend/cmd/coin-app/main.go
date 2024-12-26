package main

import (
	"coin-app/internal/config"
	"coin-app/internal/http-server/handlers/wallet/create"
	"coin-app/internal/http-server/handlers/wallet/transaction"
	"coin-app/internal/http-server/handlers/wallet/wallet"
	"coin-app/internal/lib/logger/handlers/slogpretty"
	"coin-app/internal/lib/logger/sl"
	"coin-app/internal/storage/postgres"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mwLogger "coin-app/internal/http-server/middleware/logger"
	walletService "coin-app/internal/services/wallet"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
	storage, err := postgres.New()
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	walletService := walletService.New(log, storage, storage)

	// Init router: chi, "chi render"
	router := setupRouter(log, walletService)

	// Init server
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Run server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", sl.Err(err))
			os.Exit(1)
		}
	}()
	log.Info("server started")

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping server", slog.String("signal", sign.String()))

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: Добавить отдельную остановку для SQLite сервера

	log.Info("server gracefully stopped")
}

func setupRouter(log *slog.Logger, walletService *walletService.Wallet) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Post("/wallet/create", create.New(log, walletService))
	r.Post("/wallet", transaction.New(log, walletService))
	r.Get("/wallet/{walletId}", wallet.New(log, walletService))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome anonymous"))
	})

	return r
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
