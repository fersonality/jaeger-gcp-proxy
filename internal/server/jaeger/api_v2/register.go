package api_v2

import (
	"github.com/fersonality/jaeger-gcp-proxy/internal/gcp"
	jaegerpb "github.com/fersonality/jaeger-gcp-proxy/pkg/proto/third_party/jaeger/api_v2"
	"google.golang.org/grpc"
)

func RegisterQueryServiceServer(registrar grpc.ServiceRegistrar) error {
	client, err := gcp.NewCloudTraceAPIClient()
	if err != nil {
		return err
	}
	service := &queryServiceServer{
		client,
	}
	jaegerpb.RegisterQueryServiceServer(registrar, service)
	return nil
}
