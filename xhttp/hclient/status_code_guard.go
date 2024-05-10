package hclient

import (
	"fmt"
	"github.com/dghubble/sling"
	"github.com/pkg/errors"
	"github.com/yituoshiniao/kit/xlog"
	"net/http"
)

type ErrStatus struct {
	StatusCode int
	Message    string
}

func (e *ErrStatus) Error() string {
	return e.Message
}

type StatusCodeGuardDoer struct {
	doer sling.Doer
}

//判断http-code是否正常
func (t StatusCodeGuardDoer) Do(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.doer.Do(req)

	if err == nil && resp.StatusCode >= 400 {
		err = errors.Wrap(errors.New(fmt.Sprintf("StatusCode:%d", resp.StatusCode)), "http状态码异常")
		xlog.S(req.Context()).Errorw("http-code错误", "err", err)
	}

	return
}
