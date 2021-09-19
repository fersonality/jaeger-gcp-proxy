package gcp_test

import (
	"context"
	"github.com/fersonality/jaeger-gcp-proxy/internal/gcp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestCloudTraceAPIClient(t *testing.T) {
	t.Logf("creating cloud trace client using GOOGLE_CLOUD_PROJECT, GOOGLE_APPLICATION_CREDENTIALS env vars...")
	client, err := gcp.NewCloudTraceAPIClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Logf("created client...")
	endTime := timestamppb.Now()
	startTime := timestamppb.New(endTime.AsTime().Add(-24 * time.Hour))
	channel := client.ListTraces(context.Background(), "istio-ingressgateway.istio-system", startTime, endTime)

	for res := range channel {
		if res.Error != nil {
			t.Logf("list api error: %v\n", res.Error)
		} else {
			t.Logf("list response: %v\n", res.Trace)
		}
	}
}
