package hserver

import (
	"github.com/yituoshiniao/kit/xtrace"
	"net/http"
)

type TraceIdMiddleware struct{}

func NewTraceIdMiddleware() *TraceIdMiddleware {
	return &TraceIdMiddleware{}
}

func (l *TraceIdMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	newCtx := xtrace.NewCtxWithTraceId(r.Context())
	*r = *r.WithContext(newCtx)
	next(rw, r)
	if traceId := xtrace.TraceIdFromContext(newCtx); traceId != "" {
		rw.Header().Set("trace-id", traceId)
	}
}