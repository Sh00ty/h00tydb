package internal

import "context"

type Logger interface {
	Infof(ctx context.Context, message string, args ...interface{})
	Debugf(ctx context.Context, message string, args ...interface{})
	ErrorMessage(ctx context.Context, message string, args ...interface{})
	Warnf(ctx context.Context, message string, args ...interface{})
	Errorf(ctx context.Context, err error, message string, args ...interface{})
	Fatalf(ctx context.Context, message string, args ...interface{})
}
