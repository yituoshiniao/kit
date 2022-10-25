package hclient

import (
	"github.com/dghubble/sling"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/pkg/errors"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"net/http"
)

type MetricsDoer struct {
	doer sling.Doer
}

var HttpClientAPICounter *kitprometheus.Counter

const (
	HttpClientAPICounterMethod string = "method"
	HttpClientAPICounterHost   string = "host"
	HttpClientAPICounterPath   string = "path"
	HttpClientAPICounterProto  string = "proto"
	HttpClientAPICounterStatus string = "status"
)

func InitHttpClientAPICounterMetrics() {
	HttpClientAPICounter = kitprometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Namespace: "http_client",
			Name:      "api_count",
			Help:      "http client  count of Counter metrics",
		},
		[]string{
			HttpClientAPICounterMethod,
			HttpClientAPICounterHost,
			HttpClientAPICounterPath,
			HttpClientAPICounterProto,
			HttpClientAPICounterStatus,
		})
}

//Do 普罗米修斯监控
func (t MetricsDoer) Do(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.doer.Do(req)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				xlog.S(req.Context()).Errorw("MetricsDoer 错误", "err", errors.WithStack(errors.Errorf("%v", err)))
			}
		}()

		HttpClientAPICounter.With(
			HttpClientAPICounterMethod, req.Method,
			HttpClientAPICounterHost, req.Host,
			HttpClientAPICounterPath, req.URL.Path,
			HttpClientAPICounterProto, req.Proto,
			HttpClientAPICounterStatus, resp.Status,
		).Add(1)
	}()
	return
}
