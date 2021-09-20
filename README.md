#  Jaeger GCP proxy 

A gRPC proxy server which proxies GCP Cloud Trace API gRPC v1 (`cloud.google.com/go/trace/apiv1`) as like Jaeger Query Service API gRPC v2 (`jager.api_v2.QueryService`)

For [Kiali](https://kiali.io/) integration with Google Cloud Trace in ACM (Anthos Service Mesh) / Istio Service Mesh.

![Demo](./demo.png)

## Installation
- Build docker image with `docker build . -t yourdomain.com/registry/jaeger-gcp-proxy`.
- Run with below env vars configurations
  - GOOGLE_CLOUD_PROJECT: GCP project name (mandatory)
  - GOOGLE_APPLICATION_CREDENTIALS: google api service account json credential file path. Or omit it if run on GKE with workload identity & service account IAM role binding. (optional)
  - ISTIO_MESH_ID: `istio.mesh_id` filter will be applied when provieded. (optional) 
  - GRPC_HOST: `0.0.0.0:9000` by default. (optional)
  - HTTP_HOST: `0.0.0.0:8080` by default, http server works as just a redirector from jaeger URL toward GCP tracing API list page. (optional)

## Development
Local testing for cloud API connection.
- Download SA credential locally from GCP Console.
- `GOOGLE_CLOUD_PROJECT=fersonality-1 GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/gcp-sa.json go run ./cmd/server`

