package engine

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/models"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/services"
	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/storage"
)

type Engine struct {
	cfg         config.Config
	repo        storage.Repository
	sessions    *SessionManager
	services    map[string]services.Service
	listeners   []io.Closer
	packetConns []io.Closer
}

func New(cfg config.Config, repo storage.Repository) *Engine {
	return &Engine{
		cfg:      cfg,
		repo:     repo,
		sessions: NewSessionManager(repo),
		services: map[string]services.Service{
			"ssh":       services.NewSSHService(),
			"http":      services.NewHTTPService(),
			"ftp":       services.NewFTPService(),
			"udp-decoy": services.NewUDPDecoyService(),
		},
	}
}

func (e *Engine) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(e.cfg.ActiveServices()))

	for _, svcCfg := range e.cfg.ActiveServices() {
		service, ok := e.services[svcCfg.Name]
		if !ok {
			return fmt.Errorf("service %s is not registered", svcCfg.Name)
		}

		wg.Add(1)
		go func(cfg config.ServiceConfig, service services.Service) {
			defer wg.Done()
			switch cfg.Protocol {
			case "tcp":
				errCh <- e.serveTCP(ctx, cfg, service)
			case "udp":
				errCh <- e.serveUDP(ctx, cfg, service)
			default:
				errCh <- fmt.Errorf("unsupported protocol %s", cfg.Protocol)
			}
		}(svcCfg, service)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) Status() []models.ServiceStatus {
	statuses := make([]models.ServiceStatus, 0, len(e.cfg.Services))
	for _, svc := range e.cfg.Services {
		statuses = append(statuses, models.ServiceStatus{
			Name:     svc.Name,
			Protocol: svc.Protocol,
			Address:  svc.Address,
			Enabled:  svc.Enabled,
		})
	}
	return statuses
}

func (e *Engine) Repository() storage.Repository {
	return e.repo
}
