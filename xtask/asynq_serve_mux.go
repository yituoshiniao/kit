package xtask

import (
	"context"
	log "log"
	"time"

	"github.com/hibiken/asynq"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"gitlab.intsig.net/cs-server2/kit/xtrace"
)

func NewAsynqServeMux() (serveMux *asynq.ServeMux) {
	serveMux = asynq.NewServeMux()
	serveMux.Use(loggingMiddleware)
	return serveMux
}

func loggingMiddleware1(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("开始处理 %q", t.Type())
		err := h.ProcessTask(ctx, t)
		if err != nil {
			return err
		}
		log.Printf("完成处理 %q: 经过时间 = %v", t.Type(), time.Since(start))
		return nil
	})
}

// loggingMiddleware 记录任务日志 中间件
func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		ctx = xtrace.NewCtxWithTraceId(ctx)
		start := time.Now()
		xlog.S(ctx).Infow("task任务处理开始Start", "task", string(task.Payload()))
		err := h.ProcessTask(ctx, task)
		if err != nil {
			return err
		}
		xlog.S(ctx).Infow("task任务处理结束Finished", "task.Type()", task.Type(), "task耗时", time.Since(start).String())
		return nil
	})
}
