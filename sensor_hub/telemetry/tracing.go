package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// initTracerProvider creates a TracerProvider with optional OTLP gRPC export.
func initTracerProvider(ctx context.Context, res *resource.Resource, exportEnabled bool) (*sdktrace.TracerProvider, error) {
	var opts []sdktrace.TracerProviderOption
	opts = append(opts, sdktrace.WithResource(res))

	if exportEnabled {
		exporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	return tp, nil
}

// Tracer returns a named tracer from the global TracerProvider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
