package xlog

import (
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"gitlab.intsig.net/cs-server2/kit/xtrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
	"strconv"
	"strings"
)

const (
	//LogField 日志字段
	LogField = "xlog"
	//MethodPath 请求方法
	MethodPath = "method_path"
	//TimeMs 请求时间 单位毫秒
	TimeMs = "time_ms"
)

func S(ctx context.Context) *zap.SugaredLogger {
	return zap.L().With(ExtFields(ctx)...).Sugar()
}

func L(ctx context.Context) *zap.Logger {
	return zap.L().With(ExtFields(ctx)...)
}

func ExtFields(ctx context.Context) (fs []zap.Field) {
	fs = append(
		fs,
		TraceIdField(ctx),
		BaggageFlowField(ctx),
		GidField(),
		zap.Namespace(LogField),
	)
	return fs
}

//Sn 未默认加载 zap.Namespace(LogField),
func Sn(ctx context.Context) *zap.SugaredLogger {
	return zap.L().With(ExtFieldsNotNamespace(ctx)...).Sugar()
}

//Ln 未默认加载 zap.Namespace(LogField),
func Ln(ctx context.Context) *zap.Logger {
	return zap.L().With(ExtFieldsNotNamespace(ctx)...)
}

func ExtFieldsNotNamespace(ctx context.Context) (fs []zap.Field) {
	fs = append(
		fs,
		TraceIdField(ctx),
		BaggageFlowField(ctx),
		GidField(),
	)
	return fs
}

//TraceIdField 写入 taceId 到日志组件中
func TraceIdField(ctx context.Context) (f zap.Field) {
	if id := xtrace.TraceIdFromContext(ctx); id != "" {
		return zap.String("traceId", xtrace.TraceIdFromContext(ctx))
	}
	return zap.Skip()

}

//GidField ...
func GidField() (f zap.Field) {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)
	id, err := strconv.Atoi(strings.Fields(stk)[0])
	if err != nil {
		id = 0
	}
	return zap.Int64("gid", int64(id))
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
