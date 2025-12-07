package proterrors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryInterceptorOptions holds configuration for unary interceptors.
type UnaryInterceptorOptions struct {
	errorConverterOptions []ErrorConverterOption
}

// UnaryInterceptorOption configures UnaryInterceptorOptions.
type UnaryInterceptorOption func(*UnaryInterceptorOptions)

// NewUnaryInterceptorOptions constructs a UnaryInterceptorOptions instance
// and applies all provided options in order.
func NewUnaryInterceptorOptions(opts ...UnaryInterceptorOption) *UnaryInterceptorOptions {
	o := &UnaryInterceptorOptions{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithConverterOptions injects one or more ErrorConverterOption values into
// the interceptor configuration. These options control how errors are
// translated by the underlying ErrorConverter.
//
// Example:
//
//	UnaryServerInterceptor(
//	    WithConverterOptions(
//	        WithDetailType(&proterror.Unknown{}),
//	    ),
//	)
func WithConverterOptions(opts ...ErrorConverterOption) UnaryInterceptorOption {
	return func(o *UnaryInterceptorOptions) {
		o.errorConverterOptions = append(o.errorConverterOptions, opts...)
	}
}

// UnaryServerInterceptor converts returned ProtError values into gRPC
// Status errors using ErrorConverter.ToStatusError.
//
// Behavior:
//   - Handler returns (resp, nil) → response passed through
//   - Handler returns (nil, err) → err converted via ToStatusError
//
// This interceptor should be installed on your gRPC server to ensure
// all application errors are encoded into structured protobuf details.
func UnaryServerInterceptor(opts ...UnaryInterceptorOption) grpc.UnaryServerInterceptor {
	o := NewUnaryInterceptorOptions(opts...)
	converter := NewErrorConverter(o.errorConverterOptions...)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			// Convert domain error → gRPC status.
			return nil, converter.ToStatusError(err)
		}

		return resp, nil
	}
}

// UnaryClientInterceptor converts received gRPC Status errors into ProtError
// domain errors via ErrorConverter.ToProtError.
//
// Behavior:
//   - Call succeeds → nil returned
//   - Call fails → gRPC Status decoded to the appropriate ProtError
//
// This interceptor ensures clients always receive structured domain
// errors instead of raw gRPC Status values.
func UnaryClientInterceptor(opts ...UnaryInterceptorOption) grpc.UnaryClientInterceptor {
	o := NewUnaryInterceptorOptions(opts...)
	converter := NewErrorConverter(o.errorConverterOptions...)

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, callOpts...)
		if err != nil {
			// Convert gRPC status → domain error.
			return converter.ToProtError(err)
		}

		return nil
	}
}
