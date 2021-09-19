FROM golang:1.17
WORKDIR /workspace

# make deps cache
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

# run unit test and build
COPY . ./
RUN make build

EXPOSE 9000
ENV APP_PROFILE=dev
ENTRYPOINT ["/bin/bash", "-l", "-c"]
CMD ["/workspace/build/server"]
