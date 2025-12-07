package proterrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/not-for-prod/proterror/proterror"
)

// ProtError is implemented by generated proto errors that can produce a gRPC status.
type ProtError interface {
	Code() codes.Code
	Error() string
	Is(err error) bool
	Status() *status.Status
}

var defaultProtErrors = []any{
	&proterror.Unknown{},
	&proterror.InvalidArgument{},
	&proterror.NotFound{},
	&proterror.AlreadyExists{},
	&proterror.PermissionDenied{},
	&proterror.Unauthenticated{},
	&proterror.Internal{},
	&proterror.Unavailable{},
	&proterror.DeadlineExceeded{},
	&proterror.Unimplemented{},
	&proterror.FailedPrecondition{},
	&proterror.Aborted{},
	&proterror.OutOfRange{},
	&proterror.ResourceExhausted{},
	&proterror.Cancelled{},
}
