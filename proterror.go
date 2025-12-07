package proterrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProtError is implemented by generated proto errors that can produce a gRPC status.
type ProtError interface {
	Code() codes.Code
	Error() string
	Is(err error) bool
	Status() *status.Status
}
