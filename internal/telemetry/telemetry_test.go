package telemetry

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
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
	restoreOverride := setEndpointOverrideForTest("")
	defer restoreOverride()

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

func TestInitFallsBackToConsoleExporterWhenOTLPUnavailable(t *testing.T) {
	restoreOverride := setEndpointOverrideForTest("")
	defer restoreOverride()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://collector:4318")
	restoreFactory := setExporterFactoryForTest(func(_ context.Context, _ string) (sdktrace.SpanExporter, error) {
		return nil, errors.New("dial failed")
	})
	defer restoreFactory()

	stderr := captureTelemetryStderr(t, func() {
		shutdown, err := Init(context.Background())
		if err != nil {
			t.Fatalf("init telemetry: %v", err)
		}
		_, span := otel.Tracer("telemetry-test").Start(context.Background(), "startup")
		span.End()
		shutdown()
	})

	if !strings.Contains(stderr, "falling back to console exporter") {
		t.Fatalf("expected fallback warning on stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "[SPAN] startup") {
		t.Fatalf("expected console span output on stderr, got: %q", stderr)
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

func TestResolveEndpointPriorityOrder(t *testing.T) {
	restoreOverride := setEndpointOverrideForTest("")
	defer restoreOverride()

	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	writeTelemetryConfig(t, filepath.Join(home, ".sc3", "config.toml"), "[otel]\nendpoint = \"http://home:4318\"\n")
	writeTelemetryConfig(t, filepath.Join(work, ".sc3", "config.toml"), "[otel]\nendpoint = \"http://project:4318\"\n")

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(originalWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})
	if err := os.Chdir(work); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	SetEndpointOverride("https://flag:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://env:4318")
	if got := resolveEndpoint(); got != "https://flag:4318" {
		t.Fatalf("flag endpoint priority = %q, want https://flag:4318", got)
	}

	SetEndpointOverride("")
	if got := resolveEndpoint(); got != "http://env:4318" {
		t.Fatalf("env endpoint priority = %q, want http://env:4318", got)
	}

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	if got := resolveEndpoint(); got != "http://project:4318" {
		t.Fatalf("config endpoint priority = %q, want http://project:4318", got)
	}
}

func TestTLSConfigFromCertificate(t *testing.T) {
	_, err := tlsConfigFromCertificate(filepath.Join(t.TempDir(), "missing.pem"))
	if err == nil {
		t.Fatal("expected error for missing certificate path")
	}

	certPEM := generateTestCertificatePEM(t)
	certPath := filepath.Join(t.TempDir(), "otel-ca.pem")
	if writeErr := os.WriteFile(certPath, certPEM, 0o600); writeErr != nil {
		t.Fatalf("write cert: %v", writeErr)
	}

	cfg, err := tlsConfigFromCertificate(certPath)
	if err != nil {
		t.Fatalf("tls config from certificate: %v", err)
	}
	if cfg.RootCAs == nil {
		t.Fatal("expected RootCAs to be configured")
	}
}

func captureTelemetryStderr(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	originalStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = originalStderr
	})

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if closeErr := reader.Close(); closeErr != nil {
		t.Fatalf("close stderr reader: %v", closeErr)
	}
	return string(data)
}

func writeTelemetryConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func generateTestCertificatePEM(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "otel-test-ca",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
