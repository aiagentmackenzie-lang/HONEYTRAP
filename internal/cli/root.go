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

	"github.com/aiagentmackenzie-lang/HONEYTRAP/internal/config"
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
	case "profiles":
		return r.listProfiles()
	case "status":
		return r.status()
	case "sessions":
		limit := parseLimit(args[1:])
		return r.sessions(ctx, limit)
	case "events":
		limit := parseLimit(args[1:])
		return r.events(ctx, limit)
	case "version":
		fmt.Println("honeytrap v0.4.0")
		return nil
	case "help", "--help", "-h":
		return r.help()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (r *Runner) deploy(ctx context.Context, profileName string) error {
	// Try to load and apply deploy profile
	profile, err := config.LoadProfile(profileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "honeytrap: warning: %v (using env config)\n", err)
	} else {
		fmt.Printf("Deploying HONEYTRAP profile %q\n", profileName)
		// Log profile details
		for name, svc := range profile.Services {
			if svc.Enabled {
				fmt.Printf("  ✓ %s (port %d)\n", name, svc.Port)
			}
		}
		if profile.AI.Enabled {
			fmt.Printf("  ✓ AI emulation (model=%s)\n", profile.AI.Model)
		}
		if profile.Logging.PCAPCapture {
			fmt.Printf("  ✓ PCAP capture enabled\n")
		}
	}

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	return r.engine.Run(runCtx)
}

func (r *Runner) listProfiles() error {
	names, err := config.ListProfiles()
	if err != nil {
		return err
	}
	fmt.Println("Available profiles:")
	for _, name := range names {
		fmt.Printf("  %s\n", name)
	}
	return nil
}

func (r *Runner) status() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tPROTOCOL\tADDRESS\tENABLED\tACTIVE")
	for _, status := range r.engine.Status() {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%d\n", status.Name, status.Protocol, status.Address, status.Enabled, status.ActiveConns)
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
	_, err := fmt.Fprintln(os.Stdout, `HONEYTRAP — AI-Powered Deception Framework

Commands:
  deploy [profile]   Start the core honeypot engine
  profiles           List available deploy profiles
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