package hserver

import (
	"github.com/yituoshiniao/kit/xlog"
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type errKey struct{}
type respKey struct{}

var ErrKey = errKey{}
var RespKey = respKey{}

type Server struct {
	options    *Options
	middleware *negroni.Negroni
	router     *httprouter.Router
}

func New(opts ...Option) *Server {
	o := &defaultOptions
	for _, opt := range opts {
		opt(o)
	}

	middleware := o.MiddlewareFactory(o)
	router := httprouter.New()
	router.NotFound = NotFound(o.ErrFactory)
	router.MethodNotAllowed = MethodNotAllowed(o.ErrFactory)

	s := &Server{options: o, middleware: middleware, router: router}
	s.HandlerFunc(http.MethodGet, "/health", func(rw http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(rw, "ok")
	})

	return s
}

func (s *Server) GET(path string, handler HandlerFunc) {
	s.Handle(http.MethodGet, path, handler)
}

func (s *Server) HEAD(path string, handler HandlerFunc) {
	s.Handle(http.MethodHead, path, handler)
}

func (s *Server) OPTIONS(path string, handler HandlerFunc) {
	s.Handle(http.MethodOptions, path, handler)
}

func (s *Server) POST(path string, handler HandlerFunc) {
	s.Handle(http.MethodPost, path, handler)
}

func (s *Server) PUT(path string, handler HandlerFunc) {
	s.Handle(http.MethodPut, path, handler)
}

func (s *Server) PATCH(path string, handler HandlerFunc) {
	s.Handle(http.MethodPatch, path, handler)
}

func (s *Server) DELETE(path string, handler HandlerFunc) {
	s.Handle(http.MethodDelete, path, handler)
}

func (s *Server) Handle(method, path string, handler HandlerFunc) {
	xlog.S(context.Background()).Infof("添加 http 路由 %s %s", method, path)
	s.router.Handle(method, path, s.warp(handler))
}

func (s *Server) Handler(method, path string, handler http.Handler) {
	xlog.S(context.Background()).Infof("添加 http 路由 %s %s", method, path)
	s.router.Handler(method, path, handler)
}

func (s *Server) HandlerFunc(method, path string, handler http.HandlerFunc) {
	xlog.S(context.Background()).Infof("添加 http 路由 %s %s", method, path)
	s.router.HandlerFunc(method, path, handler)
}

func (s *Server) ListenAndServe(addr string) error {
	s.middleware.UseHandler(s.router)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.middleware,
		ReadTimeout:  s.options.ReadTimeout,
		WriteTimeout: s.options.WriteTimeout,
		IdleTimeout:  s.options.IdleTimeout,
	}

	return server.ListenAndServe()
}

type Handler interface {
	ServeHTTP(ctx context.Context, req *http.Request) (resp interface{}, err error)
}

type HandlerFunc func(ctx context.Context, req *http.Request) (resp interface{}, err error)

func (f HandlerFunc) ServeHTTP(ctx context.Context, req *http.Request) (resp interface{}, err error) {
	return f(ctx, req)
}

func (s *Server) warp(handler HandlerFunc) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		resp, err := handler.ServeHTTP(r.Context(), r)
		if err != nil {
			body, ct := s.options.ErrFactory.Handle(err, r)
			rw.Header().Set("Content-Type", ct)
			st, _ := status.FromError(err)
			rw.WriteHeader(HTTPStatusFromCode(st.Code()))
			_, _ = rw.Write(body)
		} else {
			body, ct, err := s.options.SuccFactory.Handle(resp)
			if err != nil {
				body, ct := s.options.ErrFactory.Handle(err, r)
				rw.Header().Set("Content-Type", ct)
				rw.WriteHeader(http.StatusInternalServerError)
				_, _ = rw.Write(body)
			} else {
				rw.Header().Set("Content-Type", ct)
				_, _ = rw.Write(body)
			}
		}

		*r = *r.WithContext(context.WithValue(r.Context(), ErrKey, err))
		*r = *r.WithContext(context.WithValue(r.Context(), RespKey, resp))
	}
}

func NotFound(trans ErrRespFactory) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		err := status.Error(codes.NotFound, "404 page not found")
		body, ct := trans.Handle(err, r)

		*r = *r.WithContext(context.WithValue(r.Context(), ErrKey, err))

		rw.Header().Set("Content-Type", ct)
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write(body)
	})
}

func MethodNotAllowed(trans ErrRespFactory) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		err := status.Error(codes.FailedPrecondition, "405 Method Not Allowed")
		body, ct := trans.Handle(err, r)

		*r = *r.WithContext(context.WithValue(r.Context(), ErrKey, err))

		rw.Header().Set("Content-Type", ct)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = rw.Write(body)
	})
}

func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
