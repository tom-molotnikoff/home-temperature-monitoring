package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// initMeterProvider creates a MeterProvider with Prometheus and optional OTLP export.
func initMeterProvider(ctx context.Context, res *resource.Resource, exportEnabled bool) (*sdkmetric.MeterProvider, http.Handler, error) {
	promExp, err := promexporter.New()
	if err != nil {
		return nil, nil, err
	}

	var opts []sdkmetric.Option
	opts = append(opts, sdkmetric.WithResource(res))
	opts = append(opts, sdkmetric.WithReader(promExp))

	if exportEnabled {
		otlpExp, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, nil, err
		}
		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExp)))
	}

	mp := sdkmetric.NewMeterProvider(opts...)
	otel.SetMeterProvider(mp)

	return mp, promhttp.Handler(), nil
}

// Meter returns a named meter from the global MeterProvider.
func Meter(name string) otelmetric.Meter {
	return otel.Meter(name)
}
