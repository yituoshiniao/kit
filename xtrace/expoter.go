package xtrace

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc/credentials"
)

type Exporter interface {
	New(ctx context.Context, config OtelConfig) (exporter sdktrace.SpanExporter, resource *resource.Resource, err error)
}

type JaegerExporter struct{}

func (j JaegerExporter) New(ctx context.Context, conf OtelConfig) (exporter sdktrace.SpanExporter, res *resource.Resource, err error) {
	ep := conf.ReporterLocalAgentHostPort
	exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(ep)))
	serviceName := conf.ServerName

	res, err = resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			// semconv.ServiceVersionKey.String("v0.1.0"),
			// attribute.String("environment", "test"),
			hostnameKeyValue(),
		),
	)

	return
}

type SignozExporter struct{}

func (j SignozExporter) New(ctx context.Context, conf OtelConfig) (exporter sdktrace.SpanExporter, res *resource.Resource, err error) {
	serviceName := conf.ServerName
	// SamplerRate := conf.SamplerParam
	insecure := conf.Insecure
	collectorURL := conf.ReporterLocalAgentHostPort

	var secureOption otlptracegrpc.Option

	if strings.ToLower(insecure) == "false" || insecure == "0" || strings.ToLower(insecure) == "f" {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err = otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	res = resource.NewWithAttributes(
		"example-service",
		attribute.String("service.name", serviceName),
		attribute.String("library.language", "go"),
		hostnameKeyValue(),
	)

	return
}
