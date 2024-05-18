package hserver

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/yituoshiniao/kit/xlog"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "http")

	// ServerField is used in every server-side log statement made through grpc_zap.Can be overwritten before initialization.
	ServerField = zap.String("span.kind", "server")
)

// LogMiddleware is a middleware interceptor that logs the request as it goes in and the response as it goes out.
type LogMiddleware struct {
}

// NewLogger returns a new LogMiddleware instance
func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (s *LogMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	startTime := time.Now()
	reqFs := []zap.Field{
		SystemField,
		ServerField,
		zap.String("uri", r.URL.String()),
		zap.Object("header", &xlog.JsonMarshaler{Key: "header", Data: r.Header}),
		zap.Object("body", &xlog.JsonMarshaler{Key: "body", Data: r.Body}),
	}

	if len(r.PostForm) > 0 {
		reqFs = append(reqFs, zap.Object("form", &xlog.JsonMarshaler{Key: "form", Data: r.PostForm}))
	}

	xlog.L(r.Context()).Check(zap.InfoLevel, "接收请求[http.server]").Write(reqFs...)

	next(rw, r)

	resp := r.Context().Value(RespKey)
	err, _ := r.Context().Value(ErrKey).(error)

	// code := status.Code(err)
	// level := grpc_zap.DefaultCodeToLevel(code)

	level := zapcore.InfoLevel

	respFs := []zap.Field{
		zap.Error(err),
		DurationToTimeMillisField(time.Since(startTime)),
	}

	if resp != nil {
		respFs = append(respFs, zap.Object("resp", &xlog.JsonMarshaler{Key: "resp", Data: resp}))
	}

	xlog.L(r.Context()).Check(level, "发送响应[http.server]").Write(respFs...)
}

func DurationToTimeMillisField(duration time.Duration) zapcore.Field {
	return zap.Float32("http.timeMs", durationToMilliseconds(duration))
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}
