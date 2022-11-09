package xrds

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

// Trace 为redis.client 增加 trace 功能 ，返回 cloned client.
func Trace(ctx context.Context, client *redis.Client) *redis.Client {
	if ctx == nil {
		return client
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		xlog.S(ctx).Debugw("parentSpan err", "err", "parentSpan nil")
		return client
	}

	ctxClient := client.WithContext(ctx)
	opts := ctxClient.Options()
	ctxClient.WrapProcess(process(ctx, parentSpan, opts))
	ctxClient.WrapProcessPipeline(processPipeline(ctx, parentSpan, opts))

	if MetricsEnable {
		ctxClient.WrapProcess(processMetrics(ctx))
		ctxClient.WrapProcessPipeline(processMetricsPipeline(ctx))
	}
	return ctxClient
}

// process 原生 process 包装器，增加trace span 埋点功能.
func process(ctx context.Context, parentSpan opentracing.Span, opts *redis.Options) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			sTime := time.Now()
			span, tmpCtx := startSpan(ctx, parentSpan, opts, "redis", cmd.Name())
			span.SetTag("cmd.Arsg", cmd.Args())
			defer span.Finish()
			defer func() {
				fields := []zap.Field{
					zap.String("cmd.Name", cmd.Name()),
					zap.Any("cmd.Args", cmd.Args()),
					zap.String("rds耗时", time.Now().Sub(sTime).String()),
				}
				xlog.L(tmpCtx).Debug("process redis 执行命令", fields...)
			}()

			obj := oldProcess(cmd)
			if cmd.Err() != nil {
				ext.Error.Set(span, true)
				span.SetTag("cmd.Err", cmd.Err())
			}
			return obj
		}
	}
}

// processPipeline 原生 processPipeline 包装器，增加trace span 埋点功能.
func processPipeline(ctx context.Context, parentSpan opentracing.Span, opts *redis.Options) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			commands := cmdsName(cmds)
			span, tmpCtx := startSpan(ctx, parentSpan, opts, "redis", commands)
			defer span.Finish()
			defer func() {
				xlog.L(tmpCtx).Debug("processPipeline redis 执行命令", zap.String("commands", commands))
			}()
			return oldProcess(cmds)
		}
	}
}

// cmdsName 转换 Pipeline 的命令为 string
func cmdsName(cmds []redis.Cmder) string {
	names := make([]string, len(cmds))
	for i, cmd := range cmds {
		//names[i] = cmd.Name()
		names[i] = fmt.Sprintf("cmd.Name:%s; cmd.args:%v", cmd.Name(), cmd.Args())
	}
	return strings.Join(names, " -> ")
}

// startSpan 开启并返回 ChildSpan，记得在调用方执行 defer span.Finish().
func startSpan(ctx context.Context, parentSpan opentracing.Span, opts *redis.Options, operationName, method string) (opentracing.Span, context.Context) {
	//tr := parentSpan.Tracer()
	//sp := tr.StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
	sp, tmpCtx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("%s-%s", operationName, method))
	ext.DBType.Set(sp, "redis")
	ext.PeerAddress.Set(sp, opts.Addr)
	ext.DBInstance.Set(sp, strconv.Itoa(opts.DB))
	ext.SpanKind.Set(sp, "client")
	sp.SetTag("db.op", method)
	return sp, tmpCtx
}
