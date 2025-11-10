// Package tracer ...
package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	t "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// TManager ...
type TManager struct {
	Tracer         t.Tracer
	TracerProvider *trace.TracerProvider
}

// NewTracer ...
func NewTracer(ctx context.Context) (*TManager, error) {
	var serviceName string

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		serviceName = md.Get("service-name")[0]
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL("http://jaeger:4318"))
	if err != nil {
		return nil, err
	}

	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironmentName("development"),
			semconv.URLFull("jaeger"),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)

	tracer := otel.GetTracerProvider().Tracer(serviceName)
	return &TManager{
		Tracer:         tracer,
		TracerProvider: tracerProvider,
	}, nil
}
