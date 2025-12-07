package proterrors

import (
	"context"
	"errors"

	"google.golang.org/grpc"
)

// UnaryInterceptorOptions holds configuration for unary interceptors.
type UnaryInterceptorOptions struct{}

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

// UnaryServerInterceptor converts returned ProtError values into gRPC
// Status errors using Converter.AsStatus.
//
// Behavior:
//   - Handler returns (resp, nil) → response passed through
//   - Handler returns (nil, err) → err converted via AsStatus
//
// This interceptor should be installed on your gRPC server to ensure
// all application errors are encoded into structured protobuf details.
func UnaryServerInterceptor(_ ...UnaryInterceptorOption) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			// Convert domain error → gRPC status.
			return nil, AsStatus(err).Err()
		}

		return resp, nil
	}
}

// UnaryClientInterceptor converts received gRPC Status errors into ProtError
// domain errors via Converter.AsProtError.
//
// Behavior:
//   - Call succeeds → nil returned
//   - Call fails → gRPC Status decoded to the appropriate ProtError
//
// This interceptor ensures clients always receive structured domain
// errors instead of raw gRPC Status values.
func UnaryClientInterceptor(_ ...UnaryInterceptorOption) grpc.UnaryClientInterceptor {
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
			protError, ok := AsProtError(err)
			if ok {
				return errors.Join(protError, err)
			}

			return err
		}

		return nil
	}
}
