package defaultclient

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HeaderJSON      = "json"
	ContentTypeJson = "Content-Type"
)

// RoundTrip implements a transport that will count requests.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	startTime := time.Now()
	sTime := startTime.UnixNano() / 1e6
	var respBody []byte
	atomic.AddInt64(&t.N, 1)
	span, ctx := opentracing.StartSpanFromContext(req.Context(), t.serverName)
	req = req.WithContext(ctx)
	defer span.Finish()

	newReq := req
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		defer req.Body.Close()
		newReq = req
		newBody := make([]byte, len(body))
		copy(newBody, body)
		buf := gbytes.NewBuffer()
		_, err := buf.Write(newBody)
		if err != nil {
			xlog.S(ctx).Warnw("buf.Write--错误", "err", err)
		}
		newReq.Body = buf
	}

	tmpFields := []zap.Field{
		zap.String(xlog.MethodPath, newReq.URL.Path),
	}
	xlog.LE(newReq.Context(), tmpFields).Debug("["+t.serverName+"]"+"发送请求",
		zap.String("Host", req.URL.Host),
		zap.String("Method", req.Method),
		zap.String("Scheme", req.URL.Scheme),
		zap.String("Req-host", req.Host),
		zap.String("Url", req.URL.String()),
		zap.String("Uri", req.URL.Path),
		zap.String("Url.RawQuery", req.URL.RawQuery),
		zap.Reflect("Header", req.Header),
		zap.Reflect("Form", req.Form),
		zap.Reflect("PostForm", req.PostForm),
		zap.ByteString("Body", body),
	)

	//trace
	ext.Component.Set(span, "http-client")
	ext.HTTPUrl.Set(span, req.URL.String())
	ext.HTTPMethod.Set(span, req.Method)
	ext.PeerHostname.Set(span, req.URL.Hostname())
	ext.PeerPort.Set(span, atouint16(req.URL.Port()))

	//发送请求
	if t.rt != nil {
		resp, err = t.rt.RoundTrip(newReq)
	} else {
		resp, err = http.DefaultTransport.RoundTrip(newReq)
	}
	if err != nil {
		ext.Error.Set(span, true)
		//span.LogKV("error-信息", err)
		span.LogFields(log.String("[http client] error错误信息", err.Error()))
	}
	if resp != nil {
		ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	}

	if resp != nil && resp.Body != nil {
		b, bErr := io.ReadAll(resp.Body)
		if bErr != nil {
			zap.L().Error("读取req.body失败", zap.Error(bErr))
			return nil, bErr
		} else {
			respBody = b
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		}
	}

	level := zap.DebugLevel
	if err != nil || (resp != nil && (resp.StatusCode < 200 || 299 < resp.StatusCode)) {
		level = zap.ErrorLevel
	}
	statusF := zap.Skip()
	statusCodeF := zap.Skip()
	contentLengthF := zap.Skip()
	respF := zap.Skip()
	path := zap.String("path", req.URL.Path)
	rawQuery := zap.String("rawQuery", req.URL.RawQuery)
	tmpRawQuery, errUrl := url.QueryUnescape(req.URL.RawQuery)
	if errUrl != nil {
		xlog.S(req.Context()).Errorw("url.QueryUnescape错误", "err", err)
	}
	if errUrl == nil {
		rawQuery = zap.String("rawQuery", tmpRawQuery)
	}

	if resp != nil {
		statusF = zap.String("status", resp.Status)
		statusCodeF = zap.Int("statusCode", resp.StatusCode)
		contentLengthF = zap.Int64("contentLength", resp.ContentLength)

	}

	if resp != nil && strings.Contains(resp.Header.Get(ContentTypeJson), HeaderJSON) {
		respF = zap.Object("resp", &jsonMarshaller{b: respBody})
	} else if resp != nil && strings.Contains(resp.Header.Get(ContentTypeJson), "text") {
		respF = zap.ByteString("respString", respBody)
	} else {
		respF = zap.String("respBody", "respBody")
	}

	respFs := xlog.ExtFields(req.Context())
	timeMs := (time.Now().UnixNano() / 1e6) - sTime
	timeField := zap.Int64("Http.timeMs", timeMs)

	respFs = append(respFs,
		zap.Error(err),
		statusF,
		statusCodeF,
		contentLengthF,
		timeField,
		//l.durationFunc(time.Since(startTime)),
		respF,
		path,
		rawQuery,
	)
	if resp != nil {
		respFs = append(respFs, zap.Reflect("header", resp.Header))
	}
	zap.L().Check(level, "接收响应[http.client]").Write(respFs...)
	return
}

func atouint16(s string) uint16 {
	v, _ := strconv.ParseUint(s, 10, 16)
	return uint16(v)
}

type jsonMarshaller struct {
	b []byte
}

func (j *jsonMarshaller) MarshalLogObject(e zapcore.ObjectEncoder) error {
	return e.AddReflected("body", j)
}

func (j *jsonMarshaller) MarshalJSON() ([]byte, error) {
	return j.b, nil
}

func DurationToTimeMillisField(duration time.Duration) zapcore.Field {
	return zap.Float32("http.timeMs", durationToMilliseconds(duration))
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}
