// vendored from github.com/containerd/containerd/errdefs
package errdefs

import (
	"github.com/pkg/errors"
)

var (
	ErrUnknown            = errors.New("unknown") // used internally to represent a missed mapping.
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrUnavailable        = errors.New("unavailable")
	ErrNotImplemented     = errors.New("not implemented") // represents not supported and unimplemented
)

// IsInvalidArgument returns true if the error is due to an invalid argument
func IsInvalidArgument(err error) bool {
	return errors.Cause(err) == ErrInvalidArgument
}

// IsNotFound returns true if the error is due to a missing object
func IsNotFound(err error) bool {
	return errors.Cause(err) == ErrNotFound
}

// IsAlreadyExists returns true if the error is due to an already existing
// metadata item
func IsAlreadyExists(err error) bool {
	return errors.Cause(err) == ErrAlreadyExists
}

// IsFailedPrecondition returns true if an operation could not proceed to the
// lack of a particular condition
func IsFailedPrecondition(err error) bool {
	return errors.Cause(err) == ErrFailedPrecondition
}

// IsUnavailable returns true if the error is due to a resource being unavailable
func IsUnavailable(err error) bool {
	return errors.Cause(err) == ErrUnavailable
}

// IsNotImplemented returns true if the error is due to not being implemented
func IsNotImplemented(err error) bool {
	return errors.Cause(err) == ErrNotImplemented
}
