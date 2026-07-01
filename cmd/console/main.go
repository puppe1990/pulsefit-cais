package main

import (
	"context"
	"log"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/console"

	"github.com/puppe1990/pulsefit/internal/store"
)

func openStore(cfg cais.Config) (*store.SQLiteStore, error) {
	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, err
	}
	if err := s.SeedDemo(); err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

func bindings(s *store.SQLiteStore) map[string]any {
	return map[string]any{
		"store": s,
		"db":    s.DB(),
		"ctx":   context.Background(),
	}
}

func main() {
	cfg := cais.Load()
	s, err := openStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	active := s
	runErr := console.Run(console.Options{
		AppName:  "PulseFit",
		Config:   cfg,
		Bindings: bindings(active),
		Reload: func() (map[string]any, error) {
			_ = active.Close()
			next, err := openStore(cfg)
			if err != nil {
				return nil, err
			}
			active = next
			return bindings(active), nil
		},
	})
	_ = active.Close()
	if runErr != nil {
		log.Fatal(runErr)
	}
}
