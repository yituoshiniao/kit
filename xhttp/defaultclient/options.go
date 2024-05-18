package defaultclient

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultServerOptions = options{
	timeout:         3 * time.Second,
	statusCodeGuard: false,
	logicCodeGuard:  true,
	durationFunc:    DurationToTimeMillisField,
	transport:       http.DefaultTransport,
	tlsConfig:       &tls.Config{},
}

type Option func(*options)

type DurationToField func(duration time.Duration) zapcore.Field

type options struct {
	serviceName     string
	target          string
	timeout         time.Duration
	statusCodeGuard bool
	logicCodeGuard  bool
	tracer          opentracing.Tracer
	logger          *zap.Logger
	durationFunc    DurationToField
	tlsConfig       *tls.Config
	transport       http.RoundTripper
	metrics         bool
}

func WithTarget(target string) Option {
	return func(o *options) {
		o.target = target
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

func EnableStatusCodeGuard() Option {
	return func(o *options) {
		o.statusCodeGuard = true
	}
}

func DisableStatusCodeGuard() Option {
	return func(o *options) {
		o.logicCodeGuard = false
	}
}

func checkServiceName() Option {
	return func(o *options) {
		if o.serviceName == "" {
			panic("http client 没有设置 serviceName，请使用 hclient.New(hclient.WithServiceName(\"passport\")) 设置")
		}
	}
}

func WithServiceName(serviceName string) Option {
	return func(o *options) {
		o.serviceName = serviceName
	}
}

func WithInsecure() Option {
	return func(o *options) {
		o.tlsConfig.InsecureSkipVerify = true
	}
}

// WithMetrics 是否采集接口请求
func WithMetrics(isMetrics bool) Option {
	return func(o *options) {
		o.metrics = isMetrics
	}
}
