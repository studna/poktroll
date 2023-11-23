package polylog

import "context"

const CtxKey = "polylog/context"

// DefaultContextLogger is the default logger implementation used when no logger
// is associated with a context. It is assigned in the implementation package's
// init() function to avoid potentially creating import cycles.
// The default logger implementation is zerolog (i.e. pkg/polylog/polyzero).
var DefaultContextLogger Logger

func Ctx(ctx context.Context) Logger {
	logger, ok := ctx.Value(CtxKey).(Logger)
	if !ok {
		// TODO_IMPROVE: support configuration of default logger implementation.
		// TODO_TECHDEBT: return disabled logger once available.
		panic("no logger associated with context; disabled logger not yet supported")
	}
	return logger
}

func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, CtxKey, logger)
}
