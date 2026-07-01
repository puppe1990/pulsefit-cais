package main

import (
	"log"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/console"
	"github.com/puppe1990/pulsefit/internal/store"
)

func main() {
	cfg := cais.Load()
	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	if err := s.SeedDemo(); err != nil {
		log.Fatal(err)
	}

	if err := console.Run(console.Options{
		AppName: "PulseFit",
		Config:  cfg,
		Bindings: map[string]any{
			"store": s,
			"db":    s.DB(),
		},
	}); err != nil {
		log.Fatal(err)
	}
}