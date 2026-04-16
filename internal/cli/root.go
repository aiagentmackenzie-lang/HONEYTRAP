package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/engine"
)

type Runner struct {
	engine *engine.Engine
}

func NewRunner(engine *engine.Engine) *Runner {
	return &Runner{engine: engine}
}

func (r *Runner) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return r.help()
	}

	switch args[0] {
	case "deploy":
		profile := "default"
		if len(args) > 1 {
			profile = args[1]
		}
		return r.deploy(ctx, profile)
	case "status":
		return r.status()
	case "sessions":
		limit := parseLimit(args[1:])
		return r.sessions(ctx, limit)
	case "events":
		limit := parseLimit(args[1:])
		return r.events(ctx, limit)
	case "version":
		fmt.Println("honeytrap phase1")
		return nil
	case "help", "--help", "-h":
		return r.help()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (r *Runner) deploy(ctx context.Context, profile string) error {
	fmt.Printf("Deploying HONEYTRAP profile %q\n", profile)
	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	return r.engine.Run(runCtx)
}

func (r *Runner) status() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tPROTOCOL\tADDRESS\tENABLED")
	for _, status := range r.engine.Status() {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\n", status.Name, status.Protocol, status.Address, status.Enabled)
	}
	return w.Flush()
}

func (r *Runner) sessions(ctx context.Context, limit int) error {
	sessions, err := r.engine.Repository().ListSessions(ctx, limit)
	if err != nil {
		return err
	}
	return printJSON(sessions)
}

func (r *Runner) events(ctx context.Context, limit int) error {
	events, err := r.engine.Repository().ListEvents(ctx, limit)
	if err != nil {
		return err
	}
	return printJSON(events)
}

func (r *Runner) help() error {
	_, err := fmt.Fprintln(os.Stdout, `HONEYTRAP Phase 1 CLI

Commands:
  deploy [profile]   Start the core honeypot engine
  status             Show configured listeners
  sessions [limit]   Print recent captured sessions as JSON
  events [limit]     Print recent captured events as JSON
  version            Print CLI version`)
	return err
}

func parseLimit(args []string) int {
	if len(args) == 0 {
		return 50
	}
	limit, err := strconv.Atoi(args[0])
	if err != nil || limit <= 0 {
		return 50
	}
	return limit
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return errors.New("encode output")
	}
	return nil
}
