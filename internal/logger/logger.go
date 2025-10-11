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
	log := ctx.Value(ctxVal).(*logger)
	if log == nil {
		zlog, _ := zap.NewDevelopment()
		log = &logger{z: zlog}
		ctx = context.WithValue(ctx, ctxVal, log)
	}

	return log, ctx
}

func F(key string, val any) Field {
	return Field{key: key, val: val}
}

type Field struct {
	key string
	val any
}

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	With(ctx context.Context, fields ...Field) (Logger, context.Context)
}

func (l *logger) Debug(msg string, fields ...Field) {
	l.z.Debug(msg, convFields(fields...)...)
}
func (l *logger) Info(msg string, fields ...Field) {
	l.z.Info(msg, convFields(fields...)...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	l.z.Warn(msg, convFields(fields...)...)
}

func (l *logger) Error(msg string, fields ...Field) {
	l.z.Error(msg, convFields(fields...)...)
}

func (l *logger) With(ctx context.Context, fields ...Field) (Logger, context.Context) {
	child := l.z.With(convFields(fields...)...)
	logger := &logger{z: child}
	ctx = context.WithValue(ctx, ctxVal, logger)
	return logger, ctx

}

func convFields(fields ...Field) []zap.Field {
	var conv []zap.Field
	for _, f := range fields {
		conv = append(conv, zap.Any(f.key, f.val))
	}

	return conv
}
