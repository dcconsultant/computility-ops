package main

import (
	"log"

	"computility-ops/backend/internal/app"
	"computility-ops/backend/internal/config"
)

func main() {
	cfg := config.Load()
	r, err := app.Build(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}
	log.Printf("starting server at %s (storage=%s)", cfg.Addr, cfg.StorageDriver)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
