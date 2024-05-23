package hserver

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecoveryMiddleware struct {
	trans ErrRespFactory
}

func NewRecoveryMiddleware(errBodyTransform ErrRespFactory) *RecoveryMiddleware {
	return &RecoveryMiddleware{trans: errBodyTransform}
}

func (m *RecoveryMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if rec := recover(); rec != nil {
			err := recoverFrom(rec)
			*r = *r.WithContext(context.WithValue(r.Context(), ErrKey, errors.WithStack(err)))

			body, ct := m.trans.Handle(err, r)

			rw.Header().Set("Content-Type", ct)
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write(body)
		}
	}()

	next(rw, r)
}

func recoverFrom(p interface{}) error {
	return status.Errorf(codes.Internal, "%s", p)
}
