package xtrace

import (
	"log"
	"os"
	// "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	// "go.opentelemetry.io/otel/exporters/jaeger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewTracerProvider(conf OtelConfig) (*sdktrace.TracerProvider, error) {
	// // 初始化 OpenTelemetry 追踪导出器,可以是 jaeger、zipkin、signoz 等数据平台
	// // "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	// exporter, err := stdouttrace.New(
	// 	stdouttrace.WithPrettyPrint(),
	// )

	// 创建 Jaeger exporter
	// ep := "http://localhost:14268/api/traces"
	ep := conf.ReporterLocalAgentHostPort

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(ep)))

	serviceName := conf.ServerName
	SamplerRate := conf.SamplerParam

	// Record information about this application in a Resource.
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			// semconv.ServiceVersionKey.String("v0.1.0"),
			// attribute.String("environment", "test"),
			hostnameKeyValue(),
		),
	)

	// serviceName := conf.ServerName
	// SamplerRate := conf.SamplerParam
	// insecure := conf.Insecure
	// collectorURL := conf.ReporterLocalAgentHostPort
	//
	// var secureOption otlptracegrpc.Option
	//
	// if strings.ToLower(insecure) == "false" || insecure == "0" || strings.ToLower(insecure) == "f" {
	// 	secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	// } else {
	// 	secureOption = otlptracegrpc.WithInsecure()
	// }
	//
	// ctx := context.Background()
	// exporter, err := otlptrace.New(
	// 	ctx,
	// 	otlptracegrpc.NewClient(
	// 		secureOption,
	// 		otlptracegrpc.WithEndpoint(collectorURL),
	// 	),
	// )
	//
	// res := resource.NewWithAttributes(
	// 	"example-service",
	// 	attribute.String("service.name", serviceName),
	// 	attribute.String("library.language", "go"),
	// 	hostnameKeyValue(),
	// )

	if err != nil {
		log.Printf("初始化  TracerProvider 失败 err:%s", err)
		return nil, err
	}

	// 设置追踪提供者（Tracer Provider）
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		// sdktrace.WithSampler(sdktrace.AlwaysSample()), // 采样率配置,全部

		// 头部采样, 根据parent span来决定是否被采样
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(SamplerRate))),

		// sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.8)), // 采样率配置
		sdktrace.WithResource(res),
	)

	// 设置全局 Tracer 提供者
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}

func hostnameKeyValue() attribute.KeyValue {
	hostname, _ := os.Hostname()
	return attribute.String("hostname", hostname)
}
