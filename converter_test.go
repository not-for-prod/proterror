package proterrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/not-for-prod/proterror/proterror"
)

func TestErrorConverter_ToProtError(t *testing.T) {
	t.Parallel()

	protError := &proterror.Unknown{}
	err := NewErrorConverter().ToProtError(protError.Status().Err())

	var unknown *proterror.Unknown
	ok := errors.As(err, &unknown)
	require.True(t, ok)
}

func TestErrorConverter_ToStatusError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   error
		allowed []any
		code    codes.Code
	}{
		{
			name:    "allowed",
			input:   &proterror.Internal{},
			allowed: []any{&proterror.Internal{}},
			code:    codes.Internal,
		},
		{
			name:    "not allowed",
			input:   &proterror.Internal{},
			allowed: []any{&proterror.Cancelled{}},
			code:    codes.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				err := NewErrorConverter(WithDetailType(tt.allowed...)).ToStatusError(tt.input)
				code := status.Code(err)
				require.Equal(t, tt.code, code)
			},
		)
	}
}
