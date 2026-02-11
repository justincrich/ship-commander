package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	defaultLogMaxSizeBytes    = 10 * 1024 * 1024
	defaultLogMaxFiles        = 5
)

// Config stores runtime settings loaded from TOML files.
type Config struct {
	DefaultHarness        string
	DefaultModel          string
	Roles                 map[string]RoleHarnessConfig
	WIPLimit              int
	MaxRevisions          int
	PlanningMaxIterations int
	StuckTimeout          time.Duration
	HeartbeatInterval     time.Duration
	GateTimeout           time.Duration
	LogMaxSizeBytes       int64
	LogMaxFiles           int
}

// RoleHarnessConfig stores role-level and domain-level harness/model overrides.
type RoleHarnessConfig struct {
	Harness string
	Model   string
	Domains map[string]HarnessModelConfig
}

// HarnessModelConfig stores one harness/model pair.
type HarnessModelConfig struct {
	Harness string
	Model   string
}

type fileConfig struct {
	DefaultHarness        *string         `toml:"default_harness"`
	DefaultModel          *string         `toml:"default_model"`
	Defaults              *defaultsConfig `toml:"defaults"`
	WIPLimit              *int            `toml:"wip_limit"`
	MaxRevisions          *int            `toml:"max_revisions"`
	PlanningMaxIterations *int            `toml:"planning_max_iterations"`
	StuckTimeout          *string         `toml:"stuck_timeout"`
	HeartbeatInterval     *string         `toml:"heartbeat_interval"`
	GateTimeout           *string         `toml:"gate_timeout"`
	LogMaxSizeMB          *int            `toml:"log_max_size_mb"`
	LogMaxFiles           *int            `toml:"log_max_files"`
}

type defaultsConfig struct {
	Harness *string `toml:"harness"`
	Model   *string `toml:"model"`
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
		Roles:                 map[string]RoleHarnessConfig{},
		WIPLimit:              defaultWIPLimit,
		MaxRevisions:          defaultMaxRevisions,
		PlanningMaxIterations: defaultPlanningIterations,
		StuckTimeout:          defaultStuckTimeout,
		HeartbeatInterval:     defaultHeartbeatInterval,
		GateTimeout:           defaultGateTimeout,
		LogMaxSizeBytes:       defaultLogMaxSizeBytes,
		LogMaxFiles:           defaultLogMaxFiles,
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
	var raw map[string]any
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return fmt.Errorf("decode config roles in %q: %w", path, err)
	}

	if err := applyScalarOverrides(cfg, decoded); err != nil {
		return err
	}
	if err := applyDurationOverrides(cfg, decoded, path); err != nil {
		return err
	}
	if err := applyLogOverrides(cfg, decoded, path); err != nil {
		return err
	}
	if err := overlayRoleConfigs(cfg, raw, path); err != nil {
		return err
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

// ResolveHarnessModel resolves harness/model with this precedence:
// role+domain specific > role specific > defaults.
//
// When availability information is provided and the selected harness is unavailable,
// the resolver falls back to an available harness and returns a warning.
func (c *Config) ResolveHarnessModel(
	role string,
	domain string,
	availability map[string]bool,
) (string, string, []string, error) {
	if c == nil {
		return "", "", nil, errors.New("config must not be nil")
	}

	selectedHarness := normalizeHarness(c.DefaultHarness)
	if selectedHarness == "" {
		selectedHarness = defaultHarness
	}
	selectedModel := strings.TrimSpace(c.DefaultModel)
	if selectedModel == "" {
		selectedModel = defaultModel
	}

	if roleConfig, ok := c.Roles[normalizeKey(role)]; ok {
		if roleHarness := normalizeHarness(roleConfig.Harness); roleHarness != "" {
			selectedHarness = roleHarness
		}
		if roleModel := strings.TrimSpace(roleConfig.Model); roleModel != "" {
			selectedModel = roleModel
		}
		if domainConfig, ok := roleConfig.Domains[normalizeKey(domain)]; ok {
			if domainHarness := normalizeHarness(domainConfig.Harness); domainHarness != "" {
				selectedHarness = domainHarness
			}
			if domainModel := strings.TrimSpace(domainConfig.Model); domainModel != "" {
				selectedModel = domainModel
			}
		}
	}

	warnings := []string{}
	if len(availability) == 0 {
		return selectedHarness, selectedModel, warnings, nil
	}
	if available, ok := availability[selectedHarness]; ok && available {
		return selectedHarness, selectedModel, warnings, nil
	}

	fallback := fallbackHarness(availability)
	if fallback == "" {
		return "", "", warnings, fmt.Errorf("configured harness %q unavailable and no fallback harness available", selectedHarness)
	}

	warnings = append(
		warnings,
		fmt.Sprintf("configured harness %q unavailable; falling back to %q", selectedHarness, fallback),
	)
	return fallback, selectedModel, warnings, nil
}

func overlayRoleConfigs(cfg *Config, raw map[string]any, path string) error {
	rolesRaw, ok := raw["roles"]
	if !ok {
		return nil
	}

	rolesMap, ok := rolesRaw.(map[string]any)
	if !ok {
		return fmt.Errorf("parse roles in %q: expected table", path)
	}
	if cfg.Roles == nil {
		cfg.Roles = map[string]RoleHarnessConfig{}
	}

	for roleName, roleValue := range rolesMap {
		if err := overlaySingleRoleConfig(cfg, roleName, roleValue, path); err != nil {
			return err
		}
	}

	return nil
}

func overlaySingleRoleConfig(cfg *Config, roleName string, roleValue any, path string) error {
	roleMap, ok := roleValue.(map[string]any)
	if !ok {
		return fmt.Errorf("parse roles.%s in %q: expected table", roleName, path)
	}
	normalizedRole := normalizeKey(roleName)
	roleConfig := cfg.Roles[normalizedRole]
	if roleConfig.Domains == nil {
		roleConfig.Domains = map[string]HarnessModelConfig{}
	}

	for key, value := range roleMap {
		if err := applyRoleOrDomainEntry(&roleConfig, roleName, key, value, path); err != nil {
			return err
		}
	}
	inheritRoleDefaults(&roleConfig)
	cfg.Roles[normalizedRole] = roleConfig
	return nil
}

func applyRoleOrDomainEntry(
	roleConfig *RoleHarnessConfig,
	roleName string,
	key string,
	value any,
	path string,
) error {
	switch normalizeKey(key) {
	case "harness":
		text, err := stringValue(value, fmt.Sprintf("roles.%s.harness", roleName), path)
		if err != nil {
			return err
		}
		roleConfig.Harness = normalizeHarness(text)
		return nil
	case "model":
		text, err := stringValue(value, fmt.Sprintf("roles.%s.model", roleName), path)
		if err != nil {
			return err
		}
		roleConfig.Model = strings.TrimSpace(text)
		return nil
	default:
		return overlayDomainConfig(roleConfig, roleName, key, value, path)
	}
}

func overlayDomainConfig(
	roleConfig *RoleHarnessConfig,
	roleName string,
	domainName string,
	value any,
	path string,
) error {
	domainMap, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("parse roles.%s.%s in %q: expected table", roleName, domainName, path)
	}

	domainConfig := roleConfig.Domains[normalizeKey(domainName)]
	for domainKey, domainValue := range domainMap {
		switch normalizeKey(domainKey) {
		case "harness":
			text, err := stringValue(
				domainValue,
				fmt.Sprintf("roles.%s.%s.harness", roleName, domainName),
				path,
			)
			if err != nil {
				return err
			}
			domainConfig.Harness = normalizeHarness(text)
		case "model":
			text, err := stringValue(
				domainValue,
				fmt.Sprintf("roles.%s.%s.model", roleName, domainName),
				path,
			)
			if err != nil {
				return err
			}
			domainConfig.Model = strings.TrimSpace(text)
		default:
			return fmt.Errorf(
				"parse roles.%s.%s.%s in %q: unsupported key",
				roleName,
				domainName,
				domainKey,
				path,
			)
		}
	}

	roleConfig.Domains[normalizeKey(domainName)] = domainConfig
	return nil
}

func inheritRoleDefaults(roleConfig *RoleHarnessConfig) {
	for domainName, domainConfig := range roleConfig.Domains {
		if normalizeHarness(domainConfig.Harness) == "" {
			domainConfig.Harness = roleConfig.Harness
		}
		if strings.TrimSpace(domainConfig.Model) == "" {
			domainConfig.Model = roleConfig.Model
		}
		roleConfig.Domains[domainName] = domainConfig
	}
}

func applyScalarOverrides(cfg *Config, decoded fileConfig) error {
	if decoded.DefaultHarness != nil {
		cfg.DefaultHarness = normalizeHarness(*decoded.DefaultHarness)
	}
	if decoded.DefaultModel != nil {
		cfg.DefaultModel = strings.TrimSpace(*decoded.DefaultModel)
	}
	if decoded.Defaults != nil {
		if decoded.Defaults.Harness != nil {
			cfg.DefaultHarness = normalizeHarness(*decoded.Defaults.Harness)
		}
		if decoded.Defaults.Model != nil {
			cfg.DefaultModel = strings.TrimSpace(*decoded.Defaults.Model)
		}
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
	return nil
}

func applyDurationOverrides(cfg *Config, decoded fileConfig, path string) error {
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

func applyLogOverrides(cfg *Config, decoded fileConfig, path string) error {
	if decoded.LogMaxSizeMB != nil {
		if *decoded.LogMaxSizeMB <= 0 {
			return fmt.Errorf("parse log_max_size_mb in %q: must be > 0", path)
		}
		cfg.LogMaxSizeBytes = int64(*decoded.LogMaxSizeMB) * 1024 * 1024
	}
	if decoded.LogMaxFiles != nil {
		if *decoded.LogMaxFiles <= 0 {
			return fmt.Errorf("parse log_max_files in %q: must be > 0", path)
		}
		cfg.LogMaxFiles = *decoded.LogMaxFiles
	}
	return nil
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeHarness(value string) string {
	return normalizeKey(value)
}

func stringValue(value any, key string, path string) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("parse %s in %q: must be string", key, path)
	}
	return text, nil
}

func fallbackHarness(availability map[string]bool) string {
	for _, preferred := range []string{defaultHarness, "claude"} {
		if availability[preferred] {
			return preferred
		}
	}

	keys := make([]string, 0, len(availability))
	for key := range availability {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if availability[key] {
			return key
		}
	}
	return ""
}
