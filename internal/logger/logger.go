package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

var ctxVal ctxKey

type logger struct {
	z *zap.Logger
}

func FromCtx(ctx context.Context) (Logger, context.Context) {
	log, _ := ctx.Value(ctxVal).(*logger)
	if log == nil {
		zlog, _ := zap.NewDevelopment()
		log = &logger{z: zlog}
		ctx = context.WithValue(ctx, ctxVal, log)
	}

	return log, ctx
}

func F(key string, val any) zap.Field {
	return zap.Any(key, val)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)

	With(ctx context.Context, fields ...zap.Field) (Logger, context.Context)
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.z.Debug(msg, fields...)
}
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.z.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.z.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	l.z.Error(msg, fields...)
}

func (l *logger) With(ctx context.Context, fields ...zap.Field) (Logger, context.Context) {
	child := l.z.With(fields...)
	logger := &logger{z: child}
	ctx = context.WithValue(ctx, ctxVal, logger)
	return logger, ctx

}
