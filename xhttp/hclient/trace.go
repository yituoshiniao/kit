package hclient

import (
	"gitlab.intsig.net/cs-server2/kit-test-cb/xlog"
	"github.com/dghubble/sling"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

type TraceDoer struct {
	doer          sling.Doer
	operationName string
}

func (t TraceDoer) Do(req *http.Request) (resp *http.Response, err error) {
	xlog.L(req.Context()).Info("t  TraceDoer")
	parentSpan := opentracing.SpanFromContext(req.Context())
	if parentSpan != nil {
		span := opentracing.StartSpan(
			t.operationName,
			opentracing.ChildOf(parentSpan.Context()))

		//body, _ := ioutil.ReadAll(req.Body)
		//str := fmt.Sprintf( "%s--- reqbody--%s", req.URL.String(), ext.SpanKindEnum(body))
		//ext.HTTPUrl.Set(span, str)

		ext.HTTPUrl.Set(span, req.URL.String())
		ext.HTTPMethod.Set(span, req.Method)
		ext.SpanKind.Set(span, "client")
		ext.SpanKindRPCClient.Set(span)


		defer span.Finish()


		//注入日志追踪信息
		// Transmit the span's TraceContext as HTTP headers on our
		// outbound request.
		err = opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))

		if err != nil {
			return nil, err
		}

		resp, err = t.doer.Do(req)

		if resp != nil {
			ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
		}

		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("error.msg", err.Error())
			return nil, err
		}

	} else {
		resp, err = t.doer.Do(req)
	}

	return
}
