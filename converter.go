package proterrors

import (
	"errors"
	"reflect"

	"google.golang.org/grpc/status"

	"github.com/not-for-prod/proterror/proterror"
)

// ErrorConverterOption configures ErrorConverter.
type ErrorConverterOption func(*ErrorConverter)

// WithDetailType registers one or more gRPC detail message types that the
// converter should recognize and return directly.
func WithDetailType(values ...any) ErrorConverterOption {
	if len(values) == 0 {
		return nil
	}

	types := make([]reflect.Type, 0, len(values))

	for _, v := range values {
		if v == nil {
			continue // avoid nil TypeOf panic
		}

		types = append(types, reflect.TypeOf(v))
	}

	return func(ec *ErrorConverter) {
		if ec.useDefault {
			ec.useDefault = false
			ec.allowedTypes = types

			return
		}

		ec.allowedTypes = append(ec.allowedTypes, types...)
	}
}

type ErrorConverter struct {
	useDefault   bool
	allowedTypes []reflect.Type
}

// NewErrorConverter creates an ErrorConverter with default error types.
func NewErrorConverter(opts ...ErrorConverterOption) *ErrorConverter {
	allowed := make([]reflect.Type, 0, len(defaultProtErrors))

	for _, def := range defaultProtErrors {
		if def != nil {
			allowed = append(allowed, reflect.TypeOf(def))
		}
	}

	ec := &ErrorConverter{
		useDefault:   true,
		allowedTypes: allowed,
	}

	for _, opt := range opts {
		opt(ec)
	}

	return ec
}

// ToProtError tries to turn an error into a ProtError if possible.
func (ec *ErrorConverter) ToProtError(err error) error {
	var pt ProtError
	if errors.As(err, &pt) {
		return pt
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	for _, detail := range st.Details() {
		dt := reflect.TypeOf(detail)

		for _, allowed := range ec.allowedTypes {
			if dt == allowed {
				err, ok = detail.(error)
				if ok {
					return err
				}
			}
		}
	}

	return err
}

// ToStatusError tries to turn a ProtError into a domain error if possible.
func (ec *ErrorConverter) ToStatusError(err error) error {
	var allowed bool

	for _, allowedType := range ec.allowedTypes {
		if reflect.TypeOf(err) == allowedType {
			allowed = true
		}
	}

	if allowed {
		var val ProtError
		if errors.As(err, &val) {
			return val.Status().Err()
		}
	}

	return (&proterror.Unknown{}).Status().Err()
}
