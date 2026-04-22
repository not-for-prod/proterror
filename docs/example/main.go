package main

import (
	"net"

	proterrors "github.com/not-for-prod/proterror"
	examplev1 "github.com/not-for-prod/proterror/docs/example/pkg/example/v1"
	"github.com/not-for-prod/proterror/docs/example/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	netListener, err := net.Listen("tcp", ":5555")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			proterrors.UnaryServerInterceptor(),
		),
	)
	reflection.Register(grpcServer)

	svc := &service.Service{}
	examplev1.RegisterExampleServiceServer(grpcServer, svc)

	grpcServer.Serve(netListener)
}
