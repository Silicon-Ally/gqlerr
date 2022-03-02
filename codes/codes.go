// Package codes contains error codes for working with the gqlerr package. For
// the most part, we copy gRPC error codes from [1] for consistency across the
// stack.
// [1] https://pkg.go.dev/google.golang.org/grpc@v1.44.0/codes#Code
package codes

type Code string

const (
	// InvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	InvalidArgument = Code("invalid_argument")

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	NotFound = Code("not_found")

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	AlreadyExists = Code("already_exists")

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	PermissionDenied = Code("permission_denied")

	// ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	ResourceExhausted = Code("resource_exhausted")

	// FailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	FailedPrecondition = Code("failed_precondition")

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	Unimplemented = Code("unimplemented")

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	Internal = Code("internal")

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	Unauthenticated = Code("unauthenticated")
)
