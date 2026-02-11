package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// ServiceName is the canonical telemetry service name.
	ServiceName = "ship-commander-3"
	// DefaultEnvironment is used when no environment variable is configured.
	DefaultEnvironment = "dev"
	// DefaultEndpoint is used when OTEL_EXPORTER_OTLP_ENDPOINT is unset.
	DefaultEndpoint = "http://localhost:4318"
	// BatchTimeout configures batch span processor flush interval.
	BatchTimeout = 5 * time.Second
	// BatchSize configures batch span processor max export batch size.
	BatchSize = 512
)

var (
	// ServiceVersion is set at build time via ldflags when available.
	ServiceVersion = "dev"

	exporterFactory = func(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(endpoint)}
		certPath := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_CERTIFICATE"))
		if certPath != "" {
			tlsConfig, err := tlsConfigFromCertificate(certPath)
			if err != nil {
				return nil, err
			}
			opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsConfig))
		}
		return otlptracehttp.New(ctx, opts...)
	}

	endpointOverrideMu sync.RWMutex
	endpointOverride   string
)

// Init configures OpenTelemetry with OTLP HTTP exporter, resource attributes, and batch processing.
func Init(ctx context.Context) (func(), error) {
	endpoint := resolveEndpoint()
	exporter, err := exporterFactory(ctx, endpoint)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"warning: OTLP exporter unavailable for %s (%v); falling back to console exporter\n",
			endpoint,
			err,
		)
		exporter = &stderrSpanExporter{out: os.Stderr}
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", ServiceName),
			attribute.String("service.version", resolveServiceVersion()),
			attribute.String("environment", resolveEnvironment()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create telemetry resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(BatchTimeout),
			sdktrace.WithMaxExportBatchSize(BatchSize),
		),
	)
	otel.SetTracerProvider(provider)

	var once sync.Once
	shutdown := func() {
		once.Do(func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), BatchTimeout)
			defer cancel()
			if err := provider.Shutdown(shutdownCtx); err != nil {
				otel.Handle(err)
			}
		})
	}

	return shutdown, nil
}

func resolveEndpoint() string {
	endpointOverrideMu.RLock()
	override := endpointOverride
	endpointOverrideMu.RUnlock()
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}

	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		endpoint = endpointFromConfig()
	}
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return endpoint
}

func endpointFromConfig() string {
	homeDir, homeErr := os.UserHomeDir()
	workDir, cwdErr := os.Getwd()
	if homeErr != nil && cwdErr != nil {
		return ""
	}

	paths := make([]string, 0, 2)
	if homeErr == nil {
		paths = append(paths, filepath.Join(homeDir, ".sc3", "config.toml"))
	}
	if cwdErr == nil {
		paths = append(paths, filepath.Join(workDir, ".sc3", "config.toml"))
	}

	candidate := ""
	for _, path := range paths {
		value, err := endpointFromConfigPath(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: unable to read telemetry endpoint from %s: %v\n", path, err)
			continue
		}
		if value != "" {
			candidate = value
		}
	}
	return candidate
}

type telemetryFileConfig struct {
	OTEL struct {
		Endpoint *string `toml:"endpoint"`
	} `toml:"otel"`
	OTLPEndpoint *string `toml:"otel_endpoint"`
}

func endpointFromConfigPath(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("stat config path: %w", err)
	}
	var decoded telemetryFileConfig
	if _, err := toml.DecodeFile(path, &decoded); err != nil {
		return "", fmt.Errorf("decode config file: %w", err)
	}
	if decoded.OTEL.Endpoint != nil {
		return strings.TrimSpace(*decoded.OTEL.Endpoint), nil
	}
	if decoded.OTLPEndpoint != nil {
		return strings.TrimSpace(*decoded.OTLPEndpoint), nil
	}
	return "", nil
}

func resolveEnvironment() string {
	for _, key := range []string{"SC3_ENV", "ENVIRONMENT", "ENV"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return strings.ToLower(value)
		}
	}
	return DefaultEnvironment
}

func resolveServiceVersion() string {
	version := strings.TrimSpace(ServiceVersion)
	if version == "" {
		return "dev"
	}
	return version
}

// SetEndpointOverride sets a process-local endpoint override (used by CLI flag precedence).
func SetEndpointOverride(endpoint string) {
	endpointOverrideMu.Lock()
	defer endpointOverrideMu.Unlock()
	endpointOverride = strings.TrimSpace(endpoint)
}

func tlsConfigFromCertificate(path string) (*tls.Config, error) {
	// #nosec G304 -- certificate path is explicitly provided by OTEL_EXPORTER_OTLP_CERTIFICATE configuration.
	certPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read OTEL certificate %q: %w", path, err)
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(certPEM); !ok {
		return nil, fmt.Errorf("parse OTEL certificate %q: no certificates found", path)
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool}, nil
}

type stderrSpanExporter struct {
	out io.Writer
}

func (e *stderrSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	if e == nil || e.out == nil {
		return nil
	}
	for _, span := range spans {
		duration := span.EndTime().Sub(span.StartTime()).Round(time.Millisecond)
		if _, err := fmt.Fprintf(e.out, "[SPAN] %s %s %v\n", span.Name(), duration, span.Status().Code); err != nil {
			return err
		}
		for _, event := range span.Events() {
			if _, err := fmt.Fprintf(e.out, "  [EVENT] %s\n", event.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *stderrSpanExporter) Shutdown(_ context.Context) error {
	return nil
}

func setExporterFactoryForTest(factory func(context.Context, string) (sdktrace.SpanExporter, error)) func() {
	previous := exporterFactory
	exporterFactory = factory
	return func() {
		exporterFactory = previous
	}
}

func setEndpointOverrideForTest(value string) func() {
	endpointOverrideMu.RLock()
	previous := endpointOverride
	endpointOverrideMu.RUnlock()
	SetEndpointOverride(value)
	return func() {
		SetEndpointOverride(previous)
	}
}
