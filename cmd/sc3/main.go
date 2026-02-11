package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/ship-commander/sc3/internal/config"
	"github.com/ship-commander/sc3/internal/logging"
	"github.com/spf13/cobra"
)

// Version is set at build time.
var Version = "dev"

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.New(ctx)
	if err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer func() {
		if closeErr := logger.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", closeErr)
		}
	}()

	cmd := newRootCommand(ctx, cfg, logger.Logger)
	cmd.SetArgs(args)
	if err := cmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

func newRootCommand(ctx context.Context, cfg *config.Config, logger *log.Logger) *cobra.Command {
	root := &cobra.Command{
		Use:           "sc3",
		Short:         "Ship Commander 3 orchestration runtime",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	root.SetVersionTemplate("{{printf \"%s\\n\" .Version}}")
	root.AddCommand(
		newLeafCommand("init", "Initialize Ship Commander 3 project state", logger),
		newLeafCommand("plan", "Run Ready Room mission planning", logger),
		newLeafCommand("execute", "Execute approved missions", logger),
		newLeafCommand("tui", "Launch terminal dashboard", logger),
		newLeafCommand("status", "Show commission and mission status", logger),
	)

	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}
		if logger == nil {
			return errors.New("logger is required")
		}
		if cfg == nil {
			return errors.New("config is required")
		}
		logger.With("command", cmd.Name()).Debug("command invocation")
		return nil
	}

	_ = ctx
	return root
}

func newLeafCommand(name, short string, logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if logger != nil {
				logger.With("command", cmd.Name()).Info("command scaffold executed")
			}
			return nil
		},
	}
}
