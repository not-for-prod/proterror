package proterrors

import (
	"errors"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/not-for-prod/proterror/proterror"
)

func WriteHTTPResponse(w http.ResponseWriter, err error) {
	var protErr ProtError

	ok := errors.As(err, &protErr)
	if !ok {
		protErr = &proterror.Internal{}
	}

	st := protErr.Status()
	marshaler := &runtime.JSONPb{} // same marshaler used by gRPC-Gateway
	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.WriteHeader(httpStatus)

	// Use same structure as gRPC-Gateway (code, message, details)
	err = marshaler.NewEncoder(w).Encode(
		map[string]any{
			"code":    st.Code(),
			"message": st.Message(),
			"details": st.Proto().GetDetails(),
		},
	)
	if err != nil {
		w.WriteHeader(httpStatus)
	}
}
