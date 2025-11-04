ARG BASE_IMAGE

# Build the manager binary
FROM golang:1.24.1 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/lobster-operator/main.go main.go
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM ${BASE_IMAGE}
WORKDIR /
COPY --from=builder /workspace/manager .
USER root
RUN mkdir -p /web/static/docs
RUN apk update && apk upgrade && apk add --no-cache libc6-compat git tcpdump bash vim iptables curl tzdata libcap && cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime && echo "Asia/Seoul" > /etc/timezone
RUN /usr/sbin/setcap cap_net_bind_service=+ep /manager
RUN addgroup -g 10000 lobster && adduser -u 10000 -S lobster -G lobster
USER lobster

COPY web/static/docs/operator/*.json /web/static/docs

ENTRYPOINT ["/manager"]
