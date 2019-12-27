//
// The interceptor package stay here for testing purpose
// Second step was to move this package in more appropriate code base
//

package logger_grpc

import (
	"context"
	"path"
	"time"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Options struct {
	LoggerContextProvider LoggerContextProvider
	LevelFunc             CodeToLevel
	CodeFunc              func(error) codes.Code
}

// LoggerContextProvider function defines the default logger context values
type LoggerContextProvider func(fullMethodName string) *logger.Context

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) logger.Level

func DefaultCodeToLevel(code codes.Code) logger.Level {
	switch code {
	case codes.OK, codes.Canceled, codes.NotFound, codes.AlreadyExists:
		return logger.InfoLevel
	case codes.InvalidArgument, codes.PermissionDenied, codes.Unauthenticated:
		return logger.NoticeLevel
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return logger.WarningLevel
		//case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
	}
	return logger.ErrorLevel
}

func newDefaultOptions() *Options {
	return &Options{
		LoggerContextProvider: func(fullMethodName string) *logger.Context {
			return nil
		},
		LevelFunc: DefaultCodeToLevel,
		CodeFunc:  status.Code,
	}
}

func EvaluateServerOpt(opts []Option) *Options {
	optCopy := newDefaultOptions()
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func EvaluateClientOpt(opts []Option) *Options {
	optCopy := newDefaultOptions()
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*Options)

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithLoggerContext(f LoggerContextProvider) Option {
	return func(o *Options) {
		o.LoggerContextProvider = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *Options) {
		o.LevelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f func(error) codes.Code) Option {
	return func(o *Options) {
		o.CodeFunc = f
	}
}

func FeedContext(loggerContext *logger.Context, ctx context.Context, fullMethod string, startTime time.Time) *logger.Context {
	if loggerContext == nil {
		loggerContext = logger.NewContext()
	}
	loggerContext.
		Add("grpc_service", path.Dir(fullMethod)[1:]).
		Add("grpc_method", path.Base(fullMethod)).
		Add("grpc_start_time", startTime.Format(time.RFC3339))

	if d, ok := ctx.Deadline(); ok {
		loggerContext.Add("grpc_request_deadline", d.Format(time.RFC3339))
	}
	return loggerContext
}
