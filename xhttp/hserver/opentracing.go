package hserver

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	"github.com/yituoshiniao/kit/xlog"
)

var (
	httpTag = opentracing.Tag{Key: string(ext.Component), Value: "http"}
)

type OpentracingMiddleware struct{}

func NewOpentracingMiddleware() *OpentracingMiddleware {
	return &OpentracingMiddleware{}
}

func (m *OpentracingMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	ctx, sp := newServerSpanFromInbound(r)
	*r = *r.WithContext(ctx)

	next(rw, r)

	err, _ := r.Context().Value(ErrKey).(error)

	finishServerSpan(sp, err)

}

func newServerSpanFromInbound(r *http.Request) (context.Context, opentracing.Span) {
	parentSpanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		xlog.S(r.Context()).Error("http_opentracing: failed parsing trace information: %v", err)
	}

	serverSpan := opentracing.GlobalTracer().StartSpan(
		r.URL.Path,
		opentracing.ChildOf(parentSpanContext),
		httpTag,
	)

	ext.HTTPMethod.Set(serverSpan, r.Method)
	ext.HTTPUrl.Set(serverSpan, r.URL.String())

	return opentracing.ContextWithSpan(r.Context(), serverSpan), serverSpan
}

func finishServerSpan(serverSpan opentracing.Span, err error) {
	if err != nil {
		ext.Error.Set(serverSpan, true)
		serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
	}
	serverSpan.Finish()
}
