package xrds

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"strings"
)

// Trace 为redis.client 增加 trace 功能 ，返回 cloned client.
func Trace(ctx context.Context, client *redis.Client) *redis.Client {
	if ctx == nil {
		return client
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		return client
	}

	ctxClient := client.WithContext(ctx)
	opts := ctxClient.Options()
	ctxClient.WrapProcess(process(parentSpan, opts))
	ctxClient.WrapProcessPipeline(processPipeline(parentSpan, opts))
	return ctxClient
}

// process 原生 process 包装器，增加trace span 埋点功能.
func process(parentSpan opentracing.Span, opts *redis.Options) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			span := startSpan(parentSpan, opts, "redis", cmd.Name())
			defer span.Finish()
			return oldProcess(cmd)
		}
	}
}

// processPipeline 原生 processPipeline 包装器，增加trace span 埋点功能.
func processPipeline(parentSpan opentracing.Span, opts *redis.Options) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			span := startSpan(parentSpan, opts, "redis", cmdsName(cmds))
			defer span.Finish()
			return oldProcess(cmds)
		}
	}
}

// cmdsName 转换 Pipeline 的命令为 string
func cmdsName(cmds []redis.Cmder) string {
	names := make([]string, len(cmds))
	for i, cmd := range cmds {
		names[i] = cmd.Name()
	}
	return strings.Join(names, " -> ")
}

// startSpan 开启并返回 ChildSpan，记得在调用方执行 defer span.Finish().
func startSpan(parentSpan opentracing.Span, opts *redis.Options, operationName, method string) opentracing.Span {
	tr := parentSpan.Tracer()
	sp := tr.StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
	ext.DBType.Set(sp, "redis")
	ext.PeerAddress.Set(sp, opts.Addr)
	ext.SpanKind.Set(sp, "client")
	sp.SetTag("db.op", method)

	return sp
}
