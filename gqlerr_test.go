package gqlerr

import (
	"context"
	"errors"
	"testing"

	"github.com/Silicon-Ally/gqlerr/codes"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestHelpers(t *testing.T) {
	tests := []struct {
		desc     string
		in       func(context.Context, string, ...zap.Field) *Error
		wantCode codes.Code
	}{
		{
			desc:     "InvalidArgument",
			in:       InvalidArgument,
			wantCode: codes.InvalidArgument,
		},
		{
			desc:     "Internal",
			in:       Internal,
			wantCode: codes.Internal,
		},
		{
			desc:     "AlreadyExists",
			in:       AlreadyExists,
			wantCode: codes.AlreadyExists,
		},
		{
			desc:     "NotFound",
			in:       NotFound,
			wantCode: codes.NotFound,
		},
		{
			desc:     "FailedPrecondition",
			in:       FailedPrecondition,
			wantCode: codes.FailedPrecondition,
		},
		{
			desc:     "PermissionDenied",
			in:       PermissionDenied,
			wantCode: codes.PermissionDenied,
		},
		{
			desc:     "Unauthenticated",
			in:       Unauthenticated,
			wantCode: codes.Unauthenticated,
		},
		{
			desc:     "Unimplemented",
			in:       Unimplemented,
			wantCode: codes.Unimplemented,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			e := test.in(context.Background(), "some error occurred")
			if e.code != test.wantCode {
				t.Errorf("%s returned code %q, want %q", test.desc, e.code, test.wantCode)
			}
		})
	}
}

func TestPresenter_NoError(t *testing.T) {
	err, logs := runPresenterWithError(t, nil /* no error */)
	if err != nil {
		t.Fatalf("presenter returned an error even though no error occurred: %v", err)
	}

	if n := len(logs.All()); n > 0 {
		t.Fatalf("got %d unexpected logs, wanted none for no error", n)
	}
}

type randomError struct{}

func (randomError) Error() string { return "a random error" }

func TestPresenter_ErrorOfWrongType(t *testing.T) {
	err, logs := runPresenterWithError(t, randomError{})
	if err == nil {
		t.Fatal("presenter did not return an error, but one was expected")
	}

	// For a random error, we should turn it into an internal error.
	wantErr := &gqlerror.Error{
		Message:    "internal error",
		Extensions: map[string]interface{}{"code": "internal"},
	}
	if diff := cmp.Diff(wantErr, err, errOpts()); diff != "" {
		t.Errorf("unexpected GQL error returned (-want +got)\n%s", diff)
	}

	gotLogs := logs.AllUntimed()
	wantLogs := []observer.LoggedEntry{
		observer.LoggedEntry{
			Entry: zapcore.Entry{
				Level:   zapcore.ErrorLevel,
				Message: "received error that was not of type *gqlerr.Error",
			},
			Context: []zapcore.Field{
				{Key: "type", Type: zapcore.StringType, String: "gqlerr.randomError"},
				{Key: "error", Type: zapcore.ErrorType, Interface: randomError{}},
			},
		},
	}

	if diff := cmp.Diff(wantLogs, gotLogs); diff != "" {
		t.Errorf("unexpected logs written (-want +got)\n%s", diff)
	}
}

func TestPresenter(t *testing.T) {
	// In this test, we're simulating a common scenario. We've received an error
	// from an input validation function, so we're returing an InvalidArgument
	// error.
	underlyingError := errors.New("muffin validation failed")

	handlerErr := InvalidArgument(context.Background(),
		"user entered a bad number of muffins",
		zap.String("muffin_type", "blueberry"),
		zap.Int("muffin_count", -123),
		zap.Error(underlyingError),
	).
		AtError().
		WithMessage("bad input given").
		WithErrorID("muffins_must_be_positive")

	err, logs := runPresenterWithError(t, handlerErr)
	if err == nil {
		t.Fatal("presenter did not return an error, but one was expected")
	}

	// For a *gqlerr.Error, we should convert it to a gqlgen-supported
	// gqlerror.Error with all the trimmings.
	wantErr := &gqlerror.Error{
		Message: "bad input given",
		Extensions: map[string]interface{}{
			"code":         "invalid_argument",
			"error_reason": "muffins_must_be_positive",
		},
	}
	if diff := cmp.Diff(wantErr, err, errOpts()); diff != "" {
		t.Errorf("unexpected GQL error returned (-want +got)\n%s", diff)
	}

	gotLogs := logs.AllUntimed()
	wantLogs := []observer.LoggedEntry{
		observer.LoggedEntry{
			Entry: zapcore.Entry{
				Level:   zapcore.ErrorLevel,
				Message: "user entered a bad number of muffins",
			},
			Context: []zapcore.Field{
				{Key: "muffin_type", Type: zapcore.StringType, String: "blueberry"},
				{Key: "muffin_count", Type: zapcore.Int64Type, Integer: -123},
				{Key: "error", Type: zapcore.ErrorType, Interface: underlyingError},
			},
		},
	}

	if diff := cmp.Diff(wantLogs, gotLogs, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("unexpected logs written (-want +got)\n%s", diff)
	}
}

func errOpts() cmp.Option {
	return cmp.Options{
		cmpopts.IgnoreUnexported(gqlerror.Error{}),
	}
}

func runPresenterWithError(t *testing.T, handlerErr error) (*gqlerror.Error, *observer.ObservedLogs) {
	core, logs := observer.New(zap.LevelEnablerFunc(func(_ zapcore.Level) bool {
		return true
	}))
	// Instantiate a test logger, and swap out the core with a test core that
	// stores all the structured logs in-memory for later inspection.
	logger := zaptest.NewLogger(t,
		zaptest.WrapOptions(
			zap.WrapCore(func(zapcore.Core) zapcore.Core {
				return core
			}),
		),
	)

	err := ErrorPresenter(logger)(context.Background(), handlerErr)
	return err, logs
}
