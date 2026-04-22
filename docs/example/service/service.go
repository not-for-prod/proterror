package service

import (
	"context"

	examplev1 "github.com/not-for-prod/proterror/docs/example/pkg/example/v1"
)

var _ examplev1.ExampleServiceServer = &Service{}

type Service struct{}

func (s Service) Test(ctx context.Context, request *examplev1.TestRequest) (*examplev1.TestResponse, error) {
	if request.Internal {
		return nil, &examplev1.InternalError{}
	} else {
		return nil, &examplev1.PublicError{
			Text: "sample text",
		}
	}
}
