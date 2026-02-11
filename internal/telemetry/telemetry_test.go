package telemetry

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type fakeExporter struct {
	exported []sdktrace.ReadOnlySpan
	shutdown bool
}

func (f *fakeExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	f.exported = append(f.exported, spans...)
	return nil
}

func (f *fakeExporter) Shutdown(_ context.Context) error {
	f.shutdown = true
	return nil
}

func TestInitUsesConfiguredEndpointAndResourceAttributes(t *testing.T) {
	originalVersion := ServiceVersion
	ServiceVersion = "v1.2.3-test"
	defer func() { ServiceVersion = originalVersion }()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://collector:4318")
	t.Setenv("SC3_ENV", "prod")

	fake := &fakeExporter{}
	capturedEndpoint := ""
	restoreFactory := setExporterFactoryForTest(func(_ context.Context, endpoint string) (sdktrace.SpanExporter, error) {
		capturedEndpoint = endpoint
		return fake, nil
	})
	defer restoreFactory()

	shutdown, err := Init(context.Background())
	if err != nil {
		t.Fatalf("init telemetry: %v", err)
	}

	if capturedEndpoint != "http://collector:4318" {
		t.Fatalf("endpoint = %q, want collector endpoint", capturedEndpoint)
	}

	_, span := otel.Tracer("telemetry-test").Start(context.Background(), "startup")
	span.End()

	shutdown()
	if !fake.shutdown {
		t.Fatal("expected exporter shutdown on telemetry shutdown")
	}
	if len(fake.exported) == 0 {
		t.Fatal("expected at least one exported span")
	}

	attrs := fake.exported[0].Resource().Attributes()
	assertResourceAttribute(t, attrs, "service.name", ServiceName)
	assertResourceAttribute(t, attrs, "service.version", "v1.2.3-test")
	assertResourceAttribute(t, attrs, "environment", "prod")
}

func TestInitUsesDefaultEndpointWhenUnset(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	fake := &fakeExporter{}
	capturedEndpoint := ""
	restoreFactory := setExporterFactoryForTest(func(_ context.Context, endpoint string) (sdktrace.SpanExporter, error) {
		capturedEndpoint = endpoint
		return fake, nil
	})
	defer restoreFactory()

	shutdown, err := Init(context.Background())
	if err != nil {
		t.Fatalf("init telemetry: %v", err)
	}
	defer shutdown()

	if capturedEndpoint != DefaultEndpoint {
		t.Fatalf("endpoint = %q, want %q", capturedEndpoint, DefaultEndpoint)
	}
}

func TestInitReturnsExporterErrors(t *testing.T) {
	restoreFactory := setExporterFactoryForTest(func(_ context.Context, _ string) (sdktrace.SpanExporter, error) {
		return nil, errors.New("dial failed")
	})
	defer restoreFactory()

	shutdown, err := Init(context.Background())
	if err == nil {
		t.Fatal("expected exporter error")
	}
	if shutdown != nil {
		t.Fatal("shutdown must be nil when init fails")
	}
}

func TestBatchConfigConstants(t *testing.T) {
	if BatchSize != 512 {
		t.Fatalf("BatchSize = %d, want 512", BatchSize)
	}
	if BatchTimeout != 5*time.Second {
		t.Fatalf("BatchTimeout = %s, want 5s", BatchTimeout)
	}
}

func assertResourceAttribute(t *testing.T, attrs []attribute.KeyValue, key, want string) {
	t.Helper()
	for _, attr := range attrs {
		if string(attr.Key) == key {
			if attr.Value.AsString() != want {
				t.Fatalf("resource attr %s = %q, want %q", key, attr.Value.AsString(), want)
			}
			return
		}
	}
	t.Fatalf("resource attribute %q not found", key)
}

func TestResolveEnvironmentFallback(t *testing.T) {
	t.Setenv("SC3_ENV", "")
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("ENV", "dev")

	if got := resolveEnvironment(); got != "dev" {
		t.Fatalf("environment = %q, want dev", got)
	}
}
