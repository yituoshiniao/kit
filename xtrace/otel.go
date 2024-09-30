package xtrace

import (
	"context"
	"log"
	"os"
	// "go.opentelemetry.io/otel/exporters/jaeger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(conf OtelConfig) (*sdktrace.TracerProvider, error) {
	// // 初始化 OpenTelemetry 追踪导出器,可以是 jaeger、zipkin、signoz 等数据平台
	// // "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	// exporter, err := stdouttrace.New(
	// 	stdouttrace.WithPrettyPrint(),
	// )

	SamplerRate := conf.SamplerParam
	ctx := context.Background()

	// 默认使用jaeger
	exporterObj := JaegerExporter{}
	exporter, res, err := exporterObj.New(ctx, conf)

	if conf.ExporterType == SignozExporterType {
		exporterObj := SignozExporter{}
		exporter, res, err = exporterObj.New(ctx, conf)
	}

	if err != nil {
		log.Printf("初始化  TracerProvider 失败 err:%s", err)
		return nil, err
	}

	// 设置追踪提供者（Tracer Provider）
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		// sdktrace.WithSampler(sdktrace.AlwaysSample()), // 采样率配置,全部
		// sdktrace.WithSampler(sdktrace.NeverSample()),  // 采样率配置,全部不采样

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
