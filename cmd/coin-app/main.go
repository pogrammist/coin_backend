package main

import (
	"coin-app/internal/config"
)

func main() {
	// Init config: cleanenv
	cfg := config.MustLoad()
	_ = cfg

	// TODO: Init logger: slog

	// TODO: Init storage: postgresql

	// TODO: Init router: chi, "chi render"

	// TODO: Run server
}
