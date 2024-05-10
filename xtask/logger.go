package xtask

import (
	"context"

	"github.com/yituoshiniao/kit/xlog"
)

type Logger struct {
	ctx context.Context
}

func NewLogger(ctx context.Context) *Logger {
	return &Logger{
		ctx: ctx,
	}
}

func (l *Logger) Info(args ...interface{}) {
	xlog.S(l.ctx).Info(args)
}

func (l *Logger) Debug(args ...interface{}) {
	xlog.S(l.ctx).Debug(args)
}

func (l *Logger) Warn(args ...interface{}) {
	xlog.S(l.ctx).Warn(args)
}

func (l *Logger) Error(args ...interface{}) {
	xlog.S(l.ctx).Error(args)
}

func (l *Logger) Fatal(args ...interface{}) {
	xlog.S(context.Background()).Fatal(args)
}
