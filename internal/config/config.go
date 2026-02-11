package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	defaultHarness            = "codex"
	defaultModel              = "gpt-5-codex"
	defaultWIPLimit           = 3
	defaultMaxRevisions       = 3
	defaultPlanningIterations = 5
	defaultStuckTimeout       = 5 * time.Minute
	defaultHeartbeatInterval  = 30 * time.Second
	defaultGateTimeout        = 120 * time.Second
)

// Config stores runtime settings loaded from TOML files.
type Config struct {
	DefaultHarness        string
	DefaultModel          string
	WIPLimit              int
	MaxRevisions          int
	PlanningMaxIterations int
	StuckTimeout          time.Duration
	HeartbeatInterval     time.Duration
	GateTimeout           time.Duration
}

type fileConfig struct {
	DefaultHarness        *string `toml:"default_harness"`
	DefaultModel          *string `toml:"default_model"`
	WIPLimit              *int    `toml:"wip_limit"`
	MaxRevisions          *int    `toml:"max_revisions"`
	PlanningMaxIterations *int    `toml:"planning_max_iterations"`
	StuckTimeout          *string `toml:"stuck_timeout"`
	HeartbeatInterval     *string `toml:"heartbeat_interval"`
	GateTimeout           *string `toml:"gate_timeout"`
}

// Load reads config from ~/.sc3/config.toml and overlays a project-local .sc3/config.toml.
func Load(ctx context.Context) (*Config, error) {
	cfg := defaults()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}

	paths := []string{
		filepath.Join(homeDir, ".sc3", "config.toml"),
		filepath.Join(workingDir, ".sc3", "config.toml"),
	}

	for _, path := range paths {
		if err := overlayFromFile(&cfg, path); err != nil {
			return nil, err
		}
	}

	_ = ctx
	return &cfg, nil
}

func defaults() Config {
	return Config{
		DefaultHarness:        defaultHarness,
		DefaultModel:          defaultModel,
		WIPLimit:              defaultWIPLimit,
		MaxRevisions:          defaultMaxRevisions,
		PlanningMaxIterations: defaultPlanningIterations,
		StuckTimeout:          defaultStuckTimeout,
		HeartbeatInterval:     defaultHeartbeatInterval,
		GateTimeout:           defaultGateTimeout,
	}
}

func overlayFromFile(cfg *Config, path string) error {
	if cfg == nil {
		return errors.New("config must not be nil")
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat config file %q: %w", path, err)
	}

	var decoded fileConfig
	if _, err := toml.DecodeFile(path, &decoded); err != nil {
		return fmt.Errorf("decode config file %q: %w", path, err)
	}

	if decoded.DefaultHarness != nil {
		cfg.DefaultHarness = *decoded.DefaultHarness
	}
	if decoded.DefaultModel != nil {
		cfg.DefaultModel = *decoded.DefaultModel
	}
	if decoded.WIPLimit != nil {
		cfg.WIPLimit = *decoded.WIPLimit
	}
	if decoded.MaxRevisions != nil {
		cfg.MaxRevisions = *decoded.MaxRevisions
	}
	if decoded.PlanningMaxIterations != nil {
		cfg.PlanningMaxIterations = *decoded.PlanningMaxIterations
	}

	if decoded.StuckTimeout != nil {
		value, err := parseDuration(*decoded.StuckTimeout, "stuck_timeout", path)
		if err != nil {
			return err
		}
		cfg.StuckTimeout = value
	}
	if decoded.HeartbeatInterval != nil {
		value, err := parseDuration(*decoded.HeartbeatInterval, "heartbeat_interval", path)
		if err != nil {
			return err
		}
		cfg.HeartbeatInterval = value
	}
	if decoded.GateTimeout != nil {
		value, err := parseDuration(*decoded.GateTimeout, "gate_timeout", path)
		if err != nil {
			return err
		}
		cfg.GateTimeout = value
	}

	return nil
}

func parseDuration(value, key, path string) (time.Duration, error) {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s in %q: %w", key, path, err)
	}
	return parsed, nil
}
