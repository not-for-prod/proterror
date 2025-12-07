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
	err, ok := AsProtError(protError.Status().Err())
	require.True(t, ok)

	var unknown *proterror.Unknown
	ok = errors.As(err, &unknown)
	require.True(t, ok)

	err, ok = AsProtError(protError)
	require.True(t, ok)

	ok = errors.As(err, &unknown)
	require.True(t, ok)
}

func TestErrorConverter_ToStatusError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input      error
		outputCode codes.Code
	}{
		{
			input:      &proterror.Internal{},
			outputCode: codes.Internal,
		},
		{
			input:      &proterror.NotFound{},
			outputCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input.Error(), func(t *testing.T) {
				t.Parallel()

				st := AsStatus(tt.input)
				code := status.Code(st.Err())
				require.Equal(t, tt.outputCode, code)
			},
		)
	}
}
