package hclient

import (
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"context"
	"fmt"
	"github.com/dghubble/sling"
	"github.com/pkg/errors"
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

	xlog.S(context.Background()).Infow("msg", "key", resp)
	xlog.S(context.Background()).Infow("msg", "key", 111111199)

	if err == nil && resp.StatusCode >= 400 {
		err = errors.Wrap(errors.New(fmt.Sprintf("StatusCode:%d", resp.StatusCode)), "http状态码异常")
		xlog.S(req.Context()).Errorw("http-code错误", "err", err)
	}

	return
}
