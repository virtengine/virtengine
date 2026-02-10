package ctxlog

import (
	"context"

	"cosmossdk.io/log"
	"github.com/go-logr/logr"
)

type CtxKey string

const (
	CtxKeyLog = CtxKey("log")
)

type options struct {
	logName CtxKey
}

type LogOption func(*options) error

// WithLogName set custom name of the log object
func WithLogName(val string) LogOption {
	return func(t *options) error {
		t.logName = CtxKey(val)
		return nil
	}
}

type dummyLogger struct{}

var _ log.Logger = (*dummyLogger)(nil)

// WithLogger add logger object to the context
// key defaults to the "log"
// use WithLogName("<custom name>") to set custom key
func WithLogger(ctx context.Context, lg log.Logger, opts ...LogOption) context.Context {
	opt, _ := applyOptions(opts...)

	ctx = context.WithValue(ctx, opt.logName, lg)

	return ctx
}

// WithLogc add logger object to the context
// key defaults to the "log"
// use WithLogName("<custom name>") to set custom key
func WithLogc(ctx context.Context, lg log.Logger, opts ...LogOption) context.Context {
	opt, _ := applyOptions(opts...)

	ctx = context.WithValue(ctx, opt.logName, lg)

	return ctx
}

func LogcFromCtx(ctx context.Context, opts ...LogOption) log.Logger {
	opt, _ := applyOptions(opts...)

	var logger log.Logger
	if lg, valid := ctx.Value(opt.logName).(log.Logger); valid {
		logger = lg
	} else {
		logger = &dummyLogger{}
	}

	return logger
}

func LogrFromCtx(ctx context.Context) logr.Logger {
	lg, _ := logr.FromContext(ctx)
	return lg
}

// Logger get logger from the context.
// By default, it uses "log" key. use WithLogName("<custom name>") to set custom key
// If logger not found dummyLogger is returned
func Logger(ctx context.Context, opts ...LogOption) log.Logger {
	opt, _ := applyOptions(opts...)

	var logger log.Logger
	if lg, valid := ctx.Value(opt.logName).(log.Logger); valid {
		logger = lg
	} else {
		logger = &dummyLogger{}
	}

	return logger
}

func (l *dummyLogger) Warn(_ string, _ ...any) {
}

func (l *dummyLogger) Impl() any {
	return l
}

func (l *dummyLogger) Debug(_ string, _ ...interface{}) {}
func (l *dummyLogger) Info(_ string, _ ...interface{})  {}
func (l *dummyLogger) Error(_ string, _ ...interface{}) {}
func (l *dummyLogger) With(_ ...interface{}) log.Logger {
	return &dummyLogger{}
}

func applyOptions(opts ...LogOption) (options, error) {
	obj := &options{}
	for _, opt := range opts {
		if err := opt(obj); err != nil {
			return options{}, err
		}
	}

	if obj.logName == "" {
		obj.logName = CtxKeyLog
	}

	return *obj, nil
}
