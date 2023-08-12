package hclient

import (
	"bytes"
	"github.com/dghubble/sling"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/http"
	"net/url"
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
		b, bErr := ioutil.ReadAll(req.Body)
		if bErr != nil {
			zap.L().Error("读取req.body失败", zap.Error(bErr))
			return nil, bErr
		} else {
			reqBody = b
			req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		}
	}

	reqFs := xlog.ExtFields(req.Context())
	reqFs = append(reqFs, zap.String("method", req.Method), zap.String("url", req.URL.String()), zap.Reflect("header", req.Header))
	if len(reqBody) > 0 && !xlog.IsSecrecyMsg(string(reqBody)) {
		if strings.Contains(req.Header.Get(ContentTypeJson), HeaderJSON) {
			reqFs = append(reqFs, zap.Object("reqBody", &jsonMarshaler{b: reqBody}))
		} else {
			reqBody, err := url.QueryUnescape(string(reqBody))
			if err != nil {
				zap.L().Error("QueryUnescape-错误", zap.Error(err))
			}
			reqFs = append(reqFs, zap.String("reqBody", reqBody))
		}
	}

	zap.L().Debug("发送请求[http.client]", reqFs...)

	resp, err = l.doer.Do(req)

	if resp != nil && resp.Body != nil {
		b, bErr := ioutil.ReadAll(resp.Body)
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
	//path := zap.Skip()
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
		respF = zap.Object("resp", &jsonMarshaler{b: respBody})
	} else {
		//respF = zap.Object("resp", &jsonMarshaler{b: respBody})
		//错误响应处理
		//respF = zap.ByteString("resp", respBody)

		if isExcludeRoutePath(req.URL.Path, excludeRoutePath) {
			respF = zap.Int("respLen", len(respBody))
		} else {
			respF = zap.ByteString("respString", respBody)
		}
	}

	respFs := xlog.ExtFields(req.Context())

	respFs = append(respFs,
		zap.Error(err),
		statusF,
		statusCodeF,
		contentLengthF,
		l.durationFunc(time.Since(startTime)),
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

var (
	excludeRoutePath []string = []string{
		"/file/public/download",
	}
)

func isExcludeRoutePath(path string, excludeRoutePath []string) bool {
	if path == "" {
		return false
	}
	for _, v := range excludeRoutePath {
		if v == path {
			return true
		}
	}
	return false
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
