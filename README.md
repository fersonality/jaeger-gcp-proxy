#  Jaeger GCP proxy 

A gRPC proxy server which proxies GCP Cloud Trace API gRPC v1 (`cloud.google.com/go/trace/apiv1`) as like Jaeger Query Service API gRPC v2 (`jager.api_v2.QueryService`)

Developed for [Kiali](https://kiali.io/) integration with Google Cloud Trace in ACM (Anthos Service Mesh) / Istio Service Mesh.

## Installation
- Build docker image with `docker build . -t yourdomain.com/registry/jaeger-gcp-proxy`.
- Run with below env vars configurations
  - GRPC_HOST: `0.0.0.0:9000` by default.
  - HTTP_HOST: `0.0.0.0:8080` by default, http server works as just a redirector from jaeger URL toward GCP tracing API list page.
  - GOOGLE_CLOUD_PROJECT: GCP project name
  - GOOGLE_APPLICATION_CREDENTIALS: google api service account json credential file path. Or omit it if run on GKE with service account IAM role binding.
  - ISTIO_MESH_ID: `istio.mesh_id` filter will be applied when provieded. (optional) 

## Development
Local testing for cloud API connection 
- `GOOGLE_CLOUD_PROJECT=fersonality-1 GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/gcp-sa.json go test ./internal/gcp/trace_test.go -test.v`

