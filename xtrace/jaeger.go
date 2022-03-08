package xtrace

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"log"
	"os"
)

func Set(conf Config) func() {
	_, closer := New(conf)
	return func() {
		_ = closer.Close()
	}
}

// NewTracer 使用配置信息初始化Jaeger Tracer，如果初始化失败会返回NoopTracer，避免出现空指针
func New(conf Config) (tracer opentracing.Tracer, closer io.Closer) {
	cfg := (config.Configuration)(conf)
	cfg.Tags = append(cfg.Tags, hostname())

	//使用zipkin 日志追踪
	//zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	//tracer, closer, err := cfg.NewTracer(
	//	config.Injector(opentracing.HTTPHeaders, zipkinPropagator),
	//	config.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
	//
	//	config.Injector(opentracing.TextMap, zipkinPropagator),
	//	config.Extractor(opentracing.TextMap, zipkinPropagator),
	//
	//	config.ZipkinSharedRPCSpan(true),
	//)



	//时候用 jaeger 日志追踪
	tracer, closer, err := cfg.NewTracer()

	if err != nil {
		closer = &NullCloser{}
		log.Printf("初始化 Jaeger Tracer 失败 err:%s", err)
		return
	}

	opentracing.SetGlobalTracer(tracer)

	return tracer, closer
}

// Deprecated
// 请使用 xtrace.New()
// NewTracer 使用配置信息初始化Jaeger Tracer，如果初始化失败会返回NoopTracer，避免出现空指针
func NewTracer(conf Config) (tracer opentracing.Tracer, closer io.Closer) {
	return New(conf)
}

func hostname() opentracing.Tag {
	hostname, _ := os.Hostname()

	return opentracing.Tag{Key: "hostname", Value: hostname}
}

type NullCloser struct {
}

func (*NullCloser) Close() error {
	return nil
}

//TraceIdSpanFromContext 从context中获取traceId
func TraceIdSpanFromContext(ctx context.Context) (traceId string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		traceId = span.Context().(jaeger.SpanContext).TraceID().String()
	}
	return
}
