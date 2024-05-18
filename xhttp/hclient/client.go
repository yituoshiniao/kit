package hclient

import (
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(opts ...Option) *sling.Sling {
	o := &defaultServerOptions
	opts = append(opts, checkServiceName())
	for _, opt := range opts {
		opt(o)
	}

	o.transport.(*http.Transport).TLSClientConfig = o.tlsConfig

	client := &http.Client{Transport: o.transport, Timeout: o.timeout}

	var doer sling.Doer

	// 先日志 修复 cancel 无法被记录情况
	doer = LogDoer{doer: client, durationFunc: o.durationFunc}
	doer = TraceDoer{doer: doer, operationName: o.serviceName}

	if o.metrics {
		doer = MetricsDoer{doer: doer}
	}

	if o.statusCodeGuard {
		doer = StatusCodeGuardDoer{doer: doer}
	}

	// return sling.New().ResponseDecoder(JsonDecoder{logicCodeGuard: o.logicCodeGuard}).Base(o.target).Doer(doer)
	return sling.New().Base(o.target).Doer(doer)
}

// Deprecated
// 请使用 hclient.New()
func NewClient(opts ...Option) *sling.Sling {
	return New(opts...)
}

func DurationToTimeMillisField(duration time.Duration) zapcore.Field {
	return zap.Float32("http.timeMs", durationToMilliseconds(duration))
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}
