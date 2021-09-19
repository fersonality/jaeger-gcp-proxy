package gcp

import (
	"cloud.google.com/go/trace/apiv1"
	"context"
	"errors"
	"google.golang.org/api/iterator"
	pb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"strings"
	"time"
)

type CloudTraceAPIClient struct {
	projectId string
	istioMeshId string
	client *trace.Client
}

type CloudTraceAPIResponse struct {
	Trace *pb.Trace
	Error error
}

func NewCloudTraceAPIClient() (*CloudTraceAPIClient, error) {
	ctx := context.Background()
	client, err := trace.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	projectId := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if len(projectId) == 0 {
		return nil, errors.New("GOOGLE_CLOUD_PROJECT env var is not configured")
	}
	istioMeshId := os.Getenv("ISTIO_MESH_ID")
	return &CloudTraceAPIClient{
		projectId,
		istioMeshId,
		client,
	}, nil
}

func (api *CloudTraceAPIClient) ListTraces(ctx context.Context, serviceName string, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) chan *CloudTraceAPIResponse {
	channel := make(chan *CloudTraceAPIResponse)
	var filters []string
	if len(api.istioMeshId) != 0 {
		filters = append(filters, "istio.mesh_id:" + api.istioMeshId)
	}
	if len(serviceName) != 0 {
		// eg. serviceName = istio-ingressgateway.istio-system
		serviceNameTokens := strings.Split(serviceName, ".")
		filters = append(filters, "istio.canonical_service:" + serviceNameTokens[0])
		if len(serviceNameTokens) > 1 {
			filters = append(filters, "istio.namespace:" + serviceNameTokens[1])
		}
	}
	if startTime == nil {
		endTime = timestamppb.Now()
		startTime = timestamppb.New(endTime.AsTime().Add(-24 * 1 * time.Hour))
	}
	req := &pb.ListTracesRequest{
		ProjectId: api.projectId,
		View: pb.ListTracesRequest_ROOTSPAN,
		StartTime: startTime,
		EndTime: endTime,
		Filter: strings.Join(filters, " "),
	}
	log.Printf("handle requests: %v", req)
	go func() {
		it := api.client.ListTraces(ctx, req)
		for {
			res, err := it.Next()
			if err != nil {
				if err == iterator.Done {
					close(channel)
					break
				} else {
					channel <- &CloudTraceAPIResponse{Error: err}
					time.Sleep(1 * time.Second)
				}
			} else {
				channel <- &CloudTraceAPIResponse{Trace: res}
			}
		}
	}()
	return channel
}

func (api *CloudTraceAPIClient) GetTrace(ctx context.Context, traceId string) *CloudTraceAPIResponse {
	req := &pb.GetTraceRequest{
		ProjectId: api.projectId,
		TraceId: traceId,
	}
	log.Printf("handle requests: %v", req)
	res, err := api.client.GetTrace(ctx, req)
	return &CloudTraceAPIResponse{Error: err, Trace: res}
}