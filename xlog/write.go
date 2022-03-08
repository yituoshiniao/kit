package xlog

import (
	"context"
	"encoding/json"
	"strings"
)


import (
	"gitlab.intsig.net/cs-server2/kit/xtrace"
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)


//兼容以前老格式--不带ctx 没有trace-id日志追踪信息
func Info(args ...interface{})  {
	sugaredLogger := zap.L().Sugar()
	sugaredLogger.Info(args)
}

func Error(args ...interface{})  {
	sugaredLogger := zap.L().Sugar()
	sugaredLogger.Info(args)
}


func S(ctx context.Context) *zap.SugaredLogger {
	return zap.L().With(ExtFields(ctx)...).Sugar()
}

func L(ctx context.Context) *zap.Logger {
	return zap.L().With(ExtFields(ctx)...)
}

// Deprecated
// 请使用 xlog.S() ,通过 ctx获取 taceId
func Ctx(ctx context.Context) *zap.SugaredLogger {
	return zap.L().With(ExtFields(ctx)...).Sugar()
}

// Deprecated
// 请使用 xlog.L()
func Ctx2(ctx context.Context) *zap.Logger {
	return zap.L().With(ExtFields(ctx)...)
}

func ExtFields(ctx context.Context) (fs []zap.Field) {
	fs = append(fs, TraceIdField(ctx), BaggageFlowField(ctx))
	return fs
}

//写入 taceId 到日志组件中
func TraceIdField(ctx context.Context) (f zap.Field) {
	if id := xtrace.TraceIdFromContext(ctx); id != "" {
		return zap.String("traceId", xtrace.TraceIdFromContext(ctx))
	}
	return zap.Skip()

}

func BaggageFlowField(ctx context.Context) (f zap.Field) {
	meta := metautils.ExtractIncoming(ctx)
	flow := meta.Get(xtrace.BaggageFlow)
	if flow != "" {
		return zap.String("baggageFlow", flow)
	}
	return zap.Skip()
}

type JsonMarshaler struct {
	Key  string
	Data interface{}
}

func (j *JsonMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected(j.Key, j)
}

func (j *JsonMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

func (j *JsonMarshaler) NeedKeepSecrecy() bool {
	b, err := j.MarshalJSON()
	if err != nil {
		return false
	}

	return IsSecrecyMsg(string(b))
}

func (j JsonMarshaler) NotNeedKeepSecrecy() bool {
	return !j.NeedKeepSecrecy()
}

type ByteMarshaler struct {
	Key  string
	Data []byte
}

func (j *ByteMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected(j.Key, j)
}

func (j *ByteMarshaler) MarshalJSON() ([]byte, error) {
	return j.Data, nil
}

func IsSecrecyMsg(msg string) bool {
	for _, s := range []string{"password", "passWord", "pass_word"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}
