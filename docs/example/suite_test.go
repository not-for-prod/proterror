package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	proterrors "github.com/not-for-prod/proterror"
	examplev1 "github.com/not-for-prod/proterror/docs/example/pkg/example/v1"
	"github.com/not-for-prod/proterror/docs/example/service"
	"github.com/not-for-prod/proterror/proterror"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type TestSuite struct {
	suite.Suite

	grpcServer  *grpc.Server
	grpcClient  *grpc.ClientConn
	serverError chan error

	exampleClient examplev1.ExampleServiceClient
}

func (suite *TestSuite) SetupSuite() {
	netListener, err := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(err)

	suite.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			proterrors.UnaryServerInterceptor(),
		),
	)
	reflection.Register(suite.grpcServer)

	svc := &service.Service{}
	examplev1.RegisterExampleServiceServer(suite.grpcServer, svc)

	suite.serverError = make(chan error)

	suite.grpcClient, err = grpc.NewClient(
		netListener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			proterrors.UnaryClientInterceptor(),
		),
	)
	suite.Require().NoError(err)

	suite.exampleClient = examplev1.NewExampleServiceClient(suite.grpcClient)

	go func() {
		if err = suite.grpcServer.Serve(netListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			suite.serverError <- err
		}
	}()
}

func (suite *TestSuite) TearDownSuite() {
	suite.Require().NoError(suite.grpcClient.Close())
	suite.grpcServer.GracefulStop()

	select {
	case err := <-suite.serverError:
		suite.Require().NoError(err)
	default:
	}
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestPublicError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := suite.exampleClient.Test(ctx, &examplev1.TestRequest{Internal: false})
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, &examplev1.PublicError{}))

	var publicError *examplev1.PublicError

	suite.Require().True(errors.As(err, &publicError))
	suite.Require().Equal("sample text", publicError.Text)
}

func (suite *TestSuite) TestInternalError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := suite.exampleClient.Test(ctx, &examplev1.TestRequest{Internal: true})
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, &proterror.Unknown{}))
}
