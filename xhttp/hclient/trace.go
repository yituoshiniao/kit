package hclient

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type TraceDoer struct {
	doer          sling.Doer
	operationName string
	tracer        trace.Tracer
	propagator    propagation.TextMapPropagator
}

func (t TraceDoer) Do(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()
	var span trace.Span

	// 从上下文中开始一个新的 Span
	ctx, span = t.tracer.Start(ctx, t.operationName)
	defer span.End()

	// 设置 HTTP 请求相关的属性
	span.SetAttributes(
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.method", req.Method),
		attribute.String("span.kind", "client"),
	)

	// 注入当前 Span 的上下文到 HTTP 请求头中
	t.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// 执行请求
	resp, err = t.doer.Do(req)

	// 设置 HTTP 状态码
	if resp != nil {
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	}

	// 错误处理
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}

	if resp != nil && (resp.StatusCode < 200 || resp.StatusCode > 299) {
		span.SetAttributes(attribute.Bool("error", true))
		tmpErr := errors.New(fmt.Sprintf("HTTP 错误码: %d", resp.StatusCode))
		span.RecordError(tmpErr)
		
		span.SetAttributes(attribute.String("err", tmpErr.Error()))

	}

	return resp, err
}

// NewTraceDoer 创建一个新的 TraceDoer，初始化 tracer 和 propagator
func NewTraceDoer(doer sling.Doer, operationName string, tracer trace.Tracer, propagator propagation.TextMapPropagator) TraceDoer {
	return TraceDoer{
		doer:          doer,
		operationName: operationName,
		tracer:        tracer,
		propagator:    propagator,
	}
}
