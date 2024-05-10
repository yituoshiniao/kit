package xtask

import (
	"context"
	log "log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yituoshiniao/kit/xlog"
	"github.com/yituoshiniao/kit/xtrace"
)

// NewAsynqServeMux asynq 注册调度任务路由
func NewAsynqServeMux() (serveMux *asynq.ServeMux) {
	serveMux = asynq.NewServeMux()
	serveMux.Use(loggingMiddleware, metricsMiddleware)
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

// 指标变量。
var (
	processedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processed_tasks_total",
			Help: "处理任务的总数",
		},
		[]string{"task_type"},
	)

	failedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "failed_tasks_total",
			Help: "处理失败的总次数",
		},
		[]string{"task_type"},
	)

	inProgressGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "in_progress_tasks",
			Help: "当前正在处理的任务数",
		},
		[]string{"task_type"},
	)
)

func metricsMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		inProgressGauge.WithLabelValues(t.Type()).Inc()
		err := next.ProcessTask(ctx, t)
		inProgressGauge.WithLabelValues(t.Type()).Dec()
		if err != nil {
			failedCounter.WithLabelValues(t.Type()).Inc()
		}
		processedCounter.WithLabelValues(t.Type()).Inc()
		return err
	})
}

// loggingMiddleware 记录任务日志 中间件
func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		ctx = xtrace.NewCtxWithTraceId(ctx)
		start := time.Now()
		xlog.S(ctx).Infow("task任务处理开始Start", "task", string(task.Payload()), "task.Type()", task.Type())
		err := h.ProcessTask(ctx, task)
		if err != nil {
			xlog.S(ctx).Errorw("任务执行错误", "err", err)
			xlog.S(ctx).Infow("task任务处理结束Finished", "task耗时", time.Since(start).String())
			return err
		}
		xlog.S(ctx).Infow("task任务处理结束Finished", "task耗时", time.Since(start).String())
		return nil
	})
}
