package api_v2

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/fersonality/jaeger-gcp-proxy/internal/gcp"
	jaegerpb "github.com/fersonality/jaeger-gcp-proxy/pkg/proto/third_party/jaeger/api_v2"
	jaeger "github.com/jaegertracing/jaeger/model"
	"google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"log"
	"runtime/debug"
)

var unimplementedError = status.Error(codes.Unimplemented, "Unimplemented")

// ref: https://github.com/kiali/kiali/blob/v1.40.1/jaeger/client.go
type queryServiceServer struct {
	client *gcp.CloudTraceAPIClient
}

// GetTrace SHOULD BE IMPLEMENTED FOR KIALI
func (q *queryServiceServer) GetTrace(request *jaegerpb.GetTraceRequest, server jaegerpb.QueryService_GetTraceServer) error {
	defer recoverHandler()
	log.Printf("Handle GetTrace request: %s", request.String())
	ctx := server.Context()
	traceId := request.GetTraceId()

	res := q.client.GetTrace(ctx, string(traceId))
	if res.Error != nil {
		log.Printf("error while fetching trace: %v", res.Error)
		return res.Error
	}
	log.Printf("Sending trace: %s", res.Trace.TraceId)
	chunk, err := transform(res.Trace)
	if err != nil {
		log.Printf("error while decoding traceId: %v", err)
		return err
	}
	return server.Send(chunk)
}

// FindTraces SHOULD BE IMPLEMENTED FOR KIALI
func (q *queryServiceServer) FindTraces(request *jaegerpb.FindTracesRequest, server jaegerpb.QueryService_FindTracesServer) error {
	defer recoverHandler()
	log.Printf("Handle FindTraces request: %s", request.String())
	ctx := server.Context()
	query := request.GetQuery()

	channel := q.client.ListTraces(ctx, query.ServiceName, query.StartTimeMin, query.StartTimeMax)
	for res := range channel {
		if res.Error != nil {
			close(channel)
			log.Printf("error while fetching traces: %v", res.Error)
			return res.Error
		}

		log.Printf("Sending trace: %s", res.Trace.TraceId)
		chunk, err := transform(res.Trace)
		if err != nil {
			log.Printf("error while decoding traceId: %v", err)
			continue
		}
		server.Send(chunk)
	}
	log.Printf("request finished...")
	return nil
}

// GetServices SHOULD BE IMPLEMENTED FOR KIALI
func (q *queryServiceServer) GetServices(ctx context.Context, request *jaegerpb.GetServicesRequest) (*jaegerpb.GetServicesResponse, error) {
	log.Printf("Handle GetServices request: %s", request.String())
	res := &jaegerpb.GetServicesResponse{
		Services: []string{},
	}
	return res, nil
}

func (q *queryServiceServer) ArchiveTrace(ctx context.Context, request *jaegerpb.ArchiveTraceRequest) (*jaegerpb.ArchiveTraceResponse, error) {
	log.Printf("Handle ArchiveTrace request: %s", request.String())
	return nil, unimplementedError
}

func (q *queryServiceServer) GetOperations(ctx context.Context, request *jaegerpb.GetOperationsRequest) (*jaegerpb.GetOperationsResponse, error) {
	log.Printf("Handle GetOperations request: %s", request.String())
	return nil, unimplementedError
}

func (q *queryServiceServer) GetDependencies(ctx context.Context, request *jaegerpb.GetDependenciesRequest) (*jaegerpb.GetDependenciesResponse, error) {
	log.Printf("Handle GetDependencies request: %s", request.String())
	return nil, unimplementedError
}

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

func i64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & 0xff)
	}
	return r
}

func recoverHandler() {
	if r := recover(); r != nil {
		log.Println("Recovered...", r)
		debug.PrintStack()
	}
}

func transform(gcpTrace *cloudtrace.Trace) (*jaegerpb.SpansResponseChunk, error) {
	var chunk = &jaegerpb.SpansResponseChunk{
		Spans: []*jaegerpb.Span{},
	}
	traceId, err := hex.DecodeString(gcpTrace.TraceId)
	if err != nil {
		return nil, err
	}
	for _, gcpSpan := range gcpTrace.Spans {
		tags := []*jaegerpb.KeyValue{}
		operationName := "Unspecified"
		for k, v := range gcpSpan.Labels {
			if k == "OperationName" {
				operationName = v
			} else {
				tags = append(tags, &jaegerpb.KeyValue{
					Key:  k,
					VStr: v,
				})
			}
		}
		// ref: https://github.com/jaegertracing/jaeger/blob/v1.26.0/model/json/model.go
		processId := new(bytes.Buffer)
		jaeger.NewProcess(gcpSpan.Name, []jaeger.KeyValue{}).Hash(processId)
		refs := []*jaegerpb.SpanRef{}
		if gcpSpan.GetParentSpanId() != 0 {
			refs = append(refs, &jaegerpb.SpanRef{
				TraceId: traceId,
				SpanId:  i64tob(gcpSpan.ParentSpanId),
				RefType: jaegerpb.SpanRefType_CHILD_OF,
			})
		}
		span := &jaegerpb.Span{
			TraceId:       traceId,
			SpanId:        i64tob(gcpSpan.SpanId),
			OperationName: operationName,
			References:    refs,
			Flags:         uint32(1), // denote it is sampled
			StartTime:     gcpSpan.StartTime,
			Duration:      durationpb.New(gcpSpan.EndTime.AsTime().Sub(gcpSpan.StartTime.AsTime())),
			Tags:          tags,
			//Logs:	   []*jaegerpb.Log{},
			Process: &jaegerpb.Process{
				ServiceName: gcpSpan.Name,
				Tags:        []*jaegerpb.KeyValue{},
			},
			//ProcessId: processId.String(),
			//Warnings: []string{},
		}
		chunk.Spans = append(chunk.Spans, span)
	}
	return chunk, nil
}