package xtrace

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/uber/jaeger-client-go/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

const KeyName = "traceId"

var randomNumber func() uint64

func init() {
	seedGenerator := utils.NewRand(time.Now().UnixNano())
	pool := sync.Pool{
		New: func() interface{} {
			return rand.NewSource(seedGenerator.Int63())
		},
	}
	randomNumber = func() uint64 {
		generator := pool.Get().(rand.Source)
		number := uint64(generator.Int63())
		pool.Put(generator)
		return number
	}
}

// TraceId represents unique 128bit identifier of a trace
type TraceId struct {
	High, Low uint64
}

func NewTraceId() string {
	return TraceId{High: randomNumber(), Low: randomNumber()}.String()
}

func (t TraceId) String() string {
	if t.High == 0 {
		return fmt.Sprintf("%x", t.Low)
	}
	return fmt.Sprintf("%x%016x", t.High, t.Low)
}

// IsValid checks if the TraceId is valid, i.e. not zero.
func (t TraceId) IsValid() bool {
	return t.High != 0 || t.Low != 0
}

// TraceIdFromContext 从context中获取TraceId
func TraceIdFromContext(ctx context.Context) (traceId string) {
	span := trace.SpanFromContext(ctx)
	// 检查是否存在有效 Span
	if !span.SpanContext().IsValid() {
		return ""
	}
	// 返回 TraceID
	return span.SpanContext().TraceID().String()
}

func NewCtxWithTraceId(ctx context.Context) context.Context {
	traceId := TraceIdFromContext(ctx)
	if traceId != "" {
		return ctx
	}

	tracer := otel.Tracer("NewCtxWithTraceId")
	ctx, _ = tracer.Start(context.Background(), "parent-operation")
	return ctx
}

func WithSubTraceId(ctx context.Context) context.Context {
	traceId := TraceIdFromContext(ctx)
	traceId = traceId + "." + NewTraceId()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.Pairs(KeyName, traceId)
	} else {
		md = md.Copy()
		md.Set(KeyName, traceId)
	}

	return metadata.NewIncomingContext(ctx, md)
}

// GetTraceID 获取ctx中的traceID
func GetTraceID(ctx context.Context) (traceID string) {
	return TraceIdFromContext(ctx)
}
