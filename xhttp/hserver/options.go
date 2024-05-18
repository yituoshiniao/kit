package hserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type ErrRespFactory func(err error, r *http.Request) (body []byte, contentType string)

func (t ErrRespFactory) Handle(err error, r *http.Request) (body []byte, contentType string) {
	return t(err, r)
}

type SuccRespFactory func(data interface{}) (body []byte, contentType string, err error)

func (t SuccRespFactory) Handle(data interface{}) (body []byte, contentType string, err error) {
	return t(data)
}

func defaultSuccFactory(data interface{}) (body []byte, contentType string, err error) {
	body, err = json.Marshal(map[string]interface{}{
		"code": 0,
		"msg":  "succ",
		"data": data,
	})
	return body, "application/json; charset=utf-8", err
}

func defaultErrFactory(err error, _ *http.Request) (body []byte, contentType string) {
	body, _ = json.Marshal(map[string]interface{}{
		"code": HTTPStatusFromCode(status.Code(err)),
		"msg":  err.Error(),
		"data": nil,
	})

	return body, "application/json; charset=utf-8"
}

type MiddlewareFactory func(o *Options) *negroni.Negroni

func defaultMiddlewareFactory(o *Options) *negroni.Negroni {
	middleware := negroni.New()
	middleware.Use(NewOpentracingMiddleware())
	middleware.Use(NewTraceIdMiddleware())
	middleware.Use(NewLogMiddleware())
	middleware.Use(NewRecoveryMiddleware(o.ErrFactory))

	return middleware
}

var defaultOptions = Options{
	SuccFactory:       defaultSuccFactory,
	ErrFactory:        defaultErrFactory,
	MiddlewareFactory: defaultMiddlewareFactory,
	ReadTimeout:       5 * time.Second,
	WriteTimeout:      5 * time.Second,
	IdleTimeout:       30 * time.Second,
}

type Option func(*Options)

type Options struct {
	MiddlewareFactory MiddlewareFactory
	SuccFactory       SuccRespFactory
	ErrFactory        ErrRespFactory
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// Deprecated
// 不再需要，调用的地方直接使用 zap.L()
func WithLogger(_ *zap.Logger) Option {
	return func(o *Options) {
	}
}

// Deprecated
// 不再需要，调用的地方直接使用 opentracing.GlobalTracer()
func WithTracer(_ opentracing.Tracer) Option {
	return func(o *Options) {
	}
}

func WithReadTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = t
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = t
	}
}

func WithIdleTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = t
	}
}

func WithMiddlewareFactory(factory MiddlewareFactory) Option {
	return func(o *Options) {
		o.MiddlewareFactory = factory
	}
}

func WithSuccFactory(factory SuccRespFactory) Option {
	return func(o *Options) {
		o.SuccFactory = factory
	}
}

func WithErrFactory(factory ErrRespFactory) Option {
	return func(o *Options) {
		o.ErrFactory = factory
	}
}
