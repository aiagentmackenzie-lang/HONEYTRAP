package app

import (
	"context"
	"fmt"
	"os"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/cli"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/engine"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

type App struct {
	runner *cli.Runner
}

func New(profileName string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if profileName != "" {
		profile, err := config.LoadProfile(profileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "honeytrap: warning: %v\n", err)
		} else {
			cfg = *config.ApplyProfile(&cfg, profile)
		}
	}

	repo, err := storage.NewMemoryRepository(cfg.SessionLogPath(), cfg.EventLogPath())
	if err != nil {
		return nil, err
	}

	core := engine.New(cfg, repo)
	return &App{runner: cli.NewRunner(core)}, nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	return a.runner.Run(ctx, args)
}
