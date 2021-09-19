package main

import (
	"flag"
	"github.com/fersonality/jaeger-gcp-proxy/internal/server"
	"os"
)

func main() {
	grpcHost := flag.String("grpc-host", getEnv("GRPC_HOST", "0.0.0.0:9000"), "gRPC server host")
	httpHost := flag.String("http-host", getEnv("HTTP_HOST", "0.0.0.0:8080"), "HTTP server host")
	flag.Parse()
	server.Run(*grpcHost, *httpHost)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}