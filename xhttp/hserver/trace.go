package hserver

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OTelMiddleware struct{}

func NewOTelMiddleware() *OTelMiddleware {
	return &OTelMiddleware{}
}

func (m *OTelMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// 从传入请求中创建新的 Span
	ctx, sp := newServerSpanFromInbound(r)
	defer sp.End()

	*r = *r.WithContext(ctx)

	next(rw, r)

	err, _ := r.Context().Value(ErrKey).(error)

	// 结束 Span 并记录错误（如果有）
	finishServerSpan(sp, err)
}

func newServerSpanFromInbound(r *http.Request) (context.Context, trace.Span) {
	tracer := otel.Tracer("http-server")

	// 提取传入请求中的上下文信息
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, r.URL.Path,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)

	return ctx, span
}

func finishServerSpan(sp trace.Span, err error) {
	if err != nil {
		sp.SetAttributes(attribute.String("error", err.Error()))
		sp.RecordError(err)
	}
}
