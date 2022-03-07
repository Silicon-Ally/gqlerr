package gqlerr

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Silicon-Ally/gqlerr/codes"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorID represents a type of error specific to the domain of the caller of
// this package, like admin_only or too_many_muffins.
type ErrorID string

type logLevel int

// These log levels mirror the levels available in zapcore.Level, since that's
// the structured logger we've chosen.
const (
	unsetLevel logLevel = iota
	debugLevel
	infoLevel
	warnLevel
	errorLevel
)

var (
	defaultLevelForCode = map[codes.Code]logLevel{
		codes.InvalidArgument:    warnLevel,
		codes.NotFound:           warnLevel,
		codes.AlreadyExists:      warnLevel,
		codes.PermissionDenied:   warnLevel,
		codes.ResourceExhausted:  warnLevel,
		codes.FailedPrecondition: warnLevel,
		codes.Unimplemented:      warnLevel,
		codes.Internal:           errorLevel,
		codes.Unauthenticated:    warnLevel,
	}

	defaultMessageForCode = map[codes.Code]string{
		codes.InvalidArgument:    "invalid argument",
		codes.NotFound:           "not found",
		codes.AlreadyExists:      "already exists",
		codes.PermissionDenied:   "permission denied",
		codes.ResourceExhausted:  "resource exhausted",
		codes.FailedPrecondition: "failed precondition",
		codes.Unimplemented:      "unimplemented",
		codes.Internal:           "internal error",
		codes.Unauthenticated:    "unauthenticated",
	}
)

type Error struct {
	// These must be set for all errors
	code codes.Code
	msg  string
	path ast.Path

	// A default log level will be chosen based on
	// code if none is provided.
	level logLevel
	// A default client message will be chosen based
	// on the code if none is provided.
	clientMsg string

	// These are additional metadata that do not need
	// to be set.
	fields  []zap.Field
	errorID ErrorID
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	err := e.err()
	if err == nil {
		// We only write the code and message for now, the actual logger should log
		// the fields.
		return fmt.Sprintf("[%q] %s", e.code, e.msg)
	}
	return fmt.Sprintf("[%q] %s: %v", e.code, e.msg, err)
}

func (e *Error) err() error {
	for _, f := range e.fields {
		if f.Key != "error" || f.Type != zapcore.ErrorType {
			continue
		}
		errVal, ok := f.Interface.(error)
		if !ok {
			continue
		}
		return errVal
	}
	return nil
}

func (e *Error) toGQLError() *gqlerror.Error {
	if e == nil {
		return nil
	}

	return &gqlerror.Error{
		Message:    e.clientMessage(),
		Path:       e.path,
		Extensions: e.extensions(),
	}
}

func (e *Error) extensions() map[string]interface{} {
	ext := map[string]interface{}{
		"code": string(e.code),
	}

	if e.errorID != "" {
		ext["error_reason"] = string(e.errorID)
	}

	return ext
}

func (e *Error) logLevel() logLevel {
	// Return a level if one was explicitly set
	if e.level != unsetLevel {
		return e.level
	}

	// Otherwise, return whatever the default level is for that code.
	level, ok := defaultLevelForCode[e.code]
	if !ok {
		// We didn't have a default for it, which seems pretty exceptional.
		return errorLevel
	}

	return level
}

func (e *Error) clientMessage() string {
	// Return the client message if one was explicitly set.
	if e.clientMsg != "" {
		return e.clientMsg
	}

	// Otherwise, return whatever the default is for that code.
	return defaultMessageForCode[e.code]
}

// WithMessage adds an error intended for clients to see, and returns the error
// for chaining purposes.
func (e *Error) WithMessage(msg string) *Error {
	e.clientMsg = msg
	return e
}

// WithErrorID adds an error intended for client apps to use, and returns the
// error for chaining purposes.
func (e *Error) WithErrorID(errID ErrorID) *Error {
	e.errorID = errID
	return e
}

func (e *Error) AtDebug() *Error {
	e.level = debugLevel
	return e
}

func (e *Error) AtInfo() *Error {
	e.level = infoLevel
	return e
}

func (e *Error) AtWarn() *Error {
	e.level = warnLevel
	return e
}

func (e *Error) AtError() *Error {
	e.level = errorLevel
	return e
}

func New(ctx context.Context, code codes.Code, msg string, fields ...zap.Field) *Error {
	return &Error{
		code:   code,
		msg:    msg,
		path:   graphql.GetPath(ctx),
		fields: fields,
	}
}

func Internal(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.Internal, msg, fields...)
}

func InvalidArgument(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.InvalidArgument, msg, fields...)
}

func NotFound(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.NotFound, msg, fields...)
}

func AlreadyExists(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.AlreadyExists, msg, fields...)
}

func PermissionDenied(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.PermissionDenied, msg, fields...)
}

func ResourceExhausted(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.ResourceExhausted, msg, fields...)
}

func FailedPrecondition(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.FailedPrecondition, msg, fields...)
}

func Unimplemented(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.Unimplemented, msg, fields...)
}

func Unauthenticated(ctx context.Context, msg string, fields ...zap.Field) *Error {
	return New(ctx, codes.Unauthenticated, msg, fields...)
}

func ErrorPresenter(logger *zap.Logger) func(context.Context, error) *gqlerror.Error {
	return func(ctx context.Context, err error) *gqlerror.Error {
		if err == nil {
			return nil
		}

		e := &Error{}
		if errors.As(err, &e) {
			logError(logger, e)
			return e.toGQLError()
		}

		logger.Error(
			"received error that was not of type *gqlerr.Error",
			zap.String("type", fmt.Sprintf("%T", err)),
			zap.Error(err),
		)
		return Internal(ctx, err.Error(), zap.Error(err)).toGQLError()
	}
}

func logError(logger *zap.Logger, err *Error) {
	var logFn func(msg string, fields ...zap.Field)
	switch err.logLevel() {
	case debugLevel:
		logFn = logger.Debug
	case infoLevel:
		logFn = logger.Info
	case warnLevel:
		logFn = logger.Warn
	case errorLevel:
		logFn = logger.Error
	default:
		// If something went wrong finding the log level, log at errorLevel.
		logFn = logger.Error
	}

	logFn(err.msg, err.fields...)
}
