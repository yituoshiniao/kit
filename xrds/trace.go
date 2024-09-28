package xrds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/yituoshiniao/kit/xlog"
	"go.opentelemetry.io/otel/trace"
	// "github.com/go-redis/redis/v8" // 确保使用 v8 或相应版本
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// otel.Tracer中的redis-client 设置的是 otel.library.name
var tracer = otel.Tracer("redis-client") // Initialize your tracer

// Trace 为redis.client 增加 trace 功能 ，返回 cloned client.
func Trace(ctx context.Context, client *redis.Client) *redis.Client {
	if ctx == nil {
		return client
	}

	ctxClient := client.WithContext(ctx)
	opts := ctxClient.Options()

	ctxClient.WrapProcess(process(ctx, opts))
	ctxClient.WrapProcessPipeline(processPipeline(ctx, opts))

	if MetricsEnable {
		ctxClient.WrapProcess(processMetrics(ctx))
		ctxClient.WrapProcessPipeline(processMetricsPipeline(ctx))
	}
	return ctxClient
}

// process 原生 process 包装器，增加trace span 埋点功能.
func process(ctx context.Context, opts *redis.Options) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			sTime := time.Now()

			// Start a new span
			ctx, span := tracer.Start(ctx, "redis "+cmd.Name())
			defer span.End()

			// // Set attributes
			// span.SetAttributes(
			// 	attribute.String("cmd.args", fmt.Sprintf("%v", cmd.Args())),
			// )

			span.AddEvent(
				"Redis executed",
				trace.WithAttributes(
					attribute.String("cmd.args", fmt.Sprintf("%v", cmd.Args())),
				),
			)

			defer func() {
				xlog.L(ctx).Debug("process redis 执行命令", zap.String("cmd.Name", cmd.Name()), zap.Any("cmd.Args", cmd.Args()), zap.String("rds耗时", time.Since(sTime).String()))
			}()

			obj := oldProcess(cmd)
			if cmd.Err() != nil {
				// 记录错误在log中
				span.RecordError(cmd.Err())
				// tag 标签错误
				span.SetAttributes(attribute.String("cmd.Err", cmd.Err().Error()))
				// 错误 告警表示，红色感叹号
				span.SetAttributes(attribute.Bool("error", true))

			}
			return obj
		}
	}
}

// processPipeline 原生 processPipeline 包装器，增加trace span 埋点功能.
func processPipeline(ctx context.Context, opts *redis.Options) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			commands := cmdsName(cmds)
			ctx, span := tracer.Start(ctx, "redis "+commands)
			defer span.End()

			xlog.L(ctx).Debug("processPipeline redis 执行命令", zap.String("commands", commands))
			return oldProcess(cmds)
		}
	}
}

// cmdsName 转换 Pipeline 的命令为 string
func cmdsName(cmds []redis.Cmder) string {
	names := make([]string, len(cmds))
	for i, cmd := range cmds {
		names[i] = fmt.Sprintf("cmd.Name:%s; cmd.args:%v", cmd.Name(), cmd.Args())
	}
	return strings.Join(names, " -> ")
}
