package server

import (
	jaeger "github.com/fersonality/jaeger-gcp-proxy/internal/server/jaeger/api_v2"
	"google.golang.org/grpc"
	"log"
	"net"
)

func Run(grpcHost string, httpHost string) {
	opts := []grpc.ServerOption{
	}
	grpcServer := grpc.NewServer(opts...)
	if err := jaeger.RegisterQueryServiceServer(grpcServer); err != nil {
		log.Fatalf("Failed to register jaeger query service: %v", err)
	}

	listener, err := net.Listen("tcp", grpcHost)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Running server: %s", grpcHost)
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed serve grpc services: %v", err)
	}
}