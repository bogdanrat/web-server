package main

import (
	"github.com/bogdanrat/web-server/service/auth/handler"
	"github.com/bogdanrat/web-server/service/auth/interceptor"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.RequestDurationInterceptor))
	pb.RegisterAuthServer(grpcServer, &handler.AuthServer{})

	log.Printf("Starting gRPC listener on port: %s\n", ":50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
