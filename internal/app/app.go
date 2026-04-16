package app

import (
	"context"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/cli"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/engine"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

type App struct {
	runner *cli.Runner
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
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
