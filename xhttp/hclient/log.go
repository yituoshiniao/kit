package hclient

import (
	"bytes"
	"github.com/dghubble/sling"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type LogDoer struct {
	doer         sling.Doer
	durationFunc DurationToField
}

const (
	HeaderJSON      = "json"
	ContentTypeJson = "Content-Type"
)

func (l LogDoer) Do(req *http.Request) (resp *http.Response, err error) {
	startTime := time.Now()
	var reqBody []byte
	var respBody []byte
	if req.Body != nil {
		if b, bErr := ioutil.ReadAll(req.Body); bErr != nil {
			zap.L().Error("读取req.body失败", zap.Error(bErr))
		} else {
			reqBody = b
			req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		}
	}

	reqFs := xlog.ExtFields(req.Context())
	reqFs = append(reqFs, zap.String("method", req.Method), zap.String("url", req.URL.String()))
	if len(reqBody) > 0 && !xlog.IsSecrecyMsg(string(reqBody)) {
		reqFs = append(reqFs, zap.Object("req", &jsonMarshaler{b: reqBody}))
	}

	zap.L().Debug("发送请求[http.client]", reqFs...)

	resp, err = l.doer.Do(req)

	if resp != nil && resp.Body != nil {
		if b, bErr := ioutil.ReadAll(resp.Body); bErr != nil {
			zap.L().Error("读取req.body失败", zap.Error(bErr))
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
	path := zap.Skip()

	if resp != nil {
		statusF = zap.String("status", resp.Status)
		statusCodeF = zap.Int("statusCode", resp.StatusCode)
		contentLengthF = zap.Int64("contentLength", resp.ContentLength)
		path = zap.String("path", req.URL.Path)

	}

	if resp != nil && (200 <= resp.StatusCode && resp.StatusCode <= 299) &&
		strings.Contains(resp.Header.Get(ContentTypeJson), HeaderJSON) {
		respF = zap.Object("resp", &jsonMarshaler{b: respBody})
	} else {
		//respF = zap.Object("resp", &jsonMarshaler{b: respBody})
		//错误响应处理
		respF = zap.ByteString("resp", respBody)
	}

	respFs := xlog.ExtFields(req.Context())
	respFs = append(respFs,
		zap.Error(err),
		statusF,
		statusCodeF,
		contentLengthF,
		zap.Reflect("header", resp.Header),
		l.durationFunc(time.Since(startTime)),
		respF,
		path,
	)

	zap.L().Check(level, "接收响应[http.client]").Write(respFs...)

	return
}

type jsonMarshaler struct {
	b []byte
}

func (j *jsonMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	return e.AddReflected("body", j)
}

func (j *jsonMarshaler) MarshalJSON() ([]byte, error) {
	return j.b, nil
}
