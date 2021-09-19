PROJECT_ID:=fersonality-1
REPO_NAME:=jaeger-gcp-proxy

.DEFAULT_GOAL := help
.PHONY: help buf-ls buf-gen buf-install test build clean submit submit-test-job

help:
	@printf "### Makefile Commands\n";
	@cat Makefile | grep .PHONY | head -n1


# for development
# list protobuf specs
buf-ls:
	buf ls-files "https://github.com/fersonality/proto.git#branch=master"

# generate protobuf codes
buf-gen: buf-ls
	rm -rf pkg/proto
	buf generate "https://github.com/fersonality/proto.git#branch=master" \
		--include-imports --path src/third_party/jaeger

# install grpc plugins
buf-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

# for build phase
# run unit test
test:
	go vet ./...
	go test ./...

dev:
	ISTIO_MESH_ID=proj-815946781175 GOOGLE_CLOUD_PROJECT=fersonality-1 GOOGLE_APPLICATION_CREDENTIALS=$(shell pwd)/gcp-sa.json air

dev-install:
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

# build local binary
build: test clean
	go build -o build/ ./...

clean:
	rm -rf build/*


# for CI/CD configuration test
docker:
	docker build --progress=plain . -t gcr.io/${PROJECT_ID}/${REPO_NAME}:$(shell git rev-parse --short HEAD)

submit:
	gcloud builds submit . --substitutions="REPO_NAME=${REPO_NAME},SHORT_SHA=$(shell git rev-parse --short HEAD),BRANCH_NAME=$(shell git rev-parse --abbrev-ref HEAD)"
