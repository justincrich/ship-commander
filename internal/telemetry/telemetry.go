package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

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
		return otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	}
)

// Init configures OpenTelemetry with OTLP HTTP exporter, resource attributes, and batch processing.
func Init(ctx context.Context) (func(), error) {
	exporter, err := exporterFactory(ctx, resolveEndpoint())
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
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
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		return DefaultEndpoint
	}
	return endpoint
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

func setExporterFactoryForTest(factory func(context.Context, string) (sdktrace.SpanExporter, error)) func() {
	previous := exporterFactory
	exporterFactory = factory
	return func() {
		exporterFactory = previous
	}
}
