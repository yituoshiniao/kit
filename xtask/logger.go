package xtask

import (
	"context"

	"gitlab.intsig.net/cs-server2/kit/xlog"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(args ...interface{}) {
	xlog.S(context.Background()).Info(args)
}

func (l *Logger) Debug(args ...interface{}) {
	xlog.S(context.Background()).Debug(args)
}

func (l *Logger) Warn(args ...interface{}) {
	xlog.S(context.Background()).Warn(args)
}

func (l *Logger) Error(args ...interface{}) {
	xlog.S(context.Background()).Error(args)
}

func (l *Logger) Fatal(args ...interface{}) {
	xlog.S(context.Background()).Fatal(args)
}
