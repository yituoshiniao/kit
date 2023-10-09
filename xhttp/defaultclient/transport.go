package defaultclient

import (
	"net/http"
)

// Transport will count requests.
type Transport struct {
	N          int64             // number of requests passing this transport
	rt         http.RoundTripper // next round-tripper or http.DefaultTransport if nil
	serverName string
}

var serverName = ""

//type RoundTripper interface {
//	RoundTrip(*http.Request) (*http.Response, error)
//}

func New(opts ...Option) *Transport {
	o := &defaultServerOptions
	opts = append(opts, checkServiceName())
	for _, opt := range opts {
		opt(o)
	}
	serverName = o.serviceName

	//var transport RoundTripper
	transport := &Transport{}
	transport.serverName = serverName
	if transport.serverName == "" {
		serverName = "http-client-api"
	}

	return transport
}
