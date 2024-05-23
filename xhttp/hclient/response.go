package hclient

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// jsonDecoder decodes http response JSON into a JSON-tagged struct value.
type JsonDecoder struct {
	logicCodeGuard bool
}

// 验证 返回code是正常
// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d JsonDecoder) Decode(resp *http.Response, v interface{}) error {
	err := errors.WithStack(json.NewDecoder(resp.Body).Decode(v))
	if err != nil {
		return err
	}

	if d.logicCodeGuard {
		ret, ok := v.(Response)
		if ok {
			if ret.GetCode() != 0 {
				return errors.New(ret.GetMsg())
			}
		}
	}

	return nil
}

type Response interface {
	GetCode() int32
	GetMsg() string
}
