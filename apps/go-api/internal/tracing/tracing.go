package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init configures the global OpenTelemetry tracer provider and text map
// propagator. It is safe to call once at startup; a failure to create the
// exporter is returned to the caller, which may choose to continue without
// tracing.
func Init() error {
	endpoint := resolveEndpoint()
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-api"
	}

	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return err
	}

	res := resource.NewSchemaless(attribute.String("service.name", serviceName))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// Shutdown flushes and shuts down the global tracer provider. It is a no-op
// if the global provider is not an SDK tracer provider (e.g. Init was never
// called or failed).
func Shutdown(ctx context.Context) error {
	tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	if !ok {
		return nil
	}
	return tp.Shutdown(ctx)
}

// resolveEndpoint returns the OTLP HTTP traces endpoint using the first
// match: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT (full URL incl. path), then
// OTEL_EXPORTER_OTLP_ENDPOINT + "/v1/traces", then the default.
func resolveEndpoint() string {
	if v := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); v != "" {
		return v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		return v + "/v1/traces"
	}
	return "http://localhost:4318/v1/traces"
}