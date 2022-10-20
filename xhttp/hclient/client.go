package hclient

import (
	"github.com/dghubble/sling"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
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
	doer = TraceDoer{doer: client, operationName: o.serviceName}
	doer = LogDoer{doer: doer, durationFunc: o.durationFunc}

	if o.metrics {
		doer = MetricsDoer{doer: doer}
	}

	if o.statusCodeGuard {
		doer = StatusCodeGuardDoer{doer: doer}
	}

	//return sling.New().ResponseDecoder(JsonDecoder{logicCodeGuard: o.logicCodeGuard}).Base(o.target).Doer(doer)
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
