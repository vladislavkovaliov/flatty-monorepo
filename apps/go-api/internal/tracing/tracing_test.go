package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// NOTE: these tests intentionally do NOT use t.Parallel() — Init() mutates
// global OTel state and t.Setenv panics under t.Parallel.

func TestInit_SucceedsWithoutJaeger(t *testing.T) {
	err := Init()
	assert.NoError(t, err)
}

func TestInit_SetsGlobalTracerProvider(t *testing.T) {
	err := Init()
	assert.NoError(t, err)

	_, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	assert.True(t, ok, "expected *sdktrace.TracerProvider, got %T", otel.GetTracerProvider())
}

func TestInit_DefaultServiceName(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")

	err := Init()
	assert.NoError(t, err)

	tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	assert.True(t, ok)
	assert.Equal(t, "go-api", serviceNameFromResource(tp))
}

func TestInit_ServiceNameFromEnv(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "test-svc")

	err := Init()
	assert.NoError(t, err)

	tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	assert.True(t, ok)
	assert.Equal(t, "test-svc", serviceNameFromResource(tp))
}

func TestInit_EndpointFromEnv(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://example.test:4318/v1/traces")

	err := Init()
	assert.NoError(t, err)
}

func TestShutdown_FlushesWithoutError(t *testing.T) {
	err := Init()
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assert.NoError(t, Shutdown(ctx))
}

func TestShutdown_NoopWithoutInit(t *testing.T) {
	otel.SetTracerProvider(noop.NewTracerProvider())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assert.NoError(t, Shutdown(ctx))
}

// serviceNameFromResource reads the provider's resource via a span. The SDK
// v1.45.0 TracerProvider has no public Resource() getter, but a recording
// span exposes the resource of the tracer that created it.
func serviceNameFromResource(tp *sdktrace.TracerProvider) string {
	_, span := tp.Tracer("tracing-test").Start(context.Background(), "test-span")
	defer span.End()

	ro, ok := span.(sdktrace.ReadOnlySpan)
	if !ok {
		return ""
	}
	for _, kv := range ro.Resource().Attributes() {
		if string(kv.Key) == "service.name" {
			return kv.Value.AsString()
		}
	}
	return ""
}