ARG BASE_IMAGE

FROM golang:1.22.7 AS builder

WORKDIR /workspace

RUN dpkg --add-architecture amd64 \
    && apt-get update \
    && apt-get install -y --no-install-recommends 

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/loggen/main.go main.go
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o loggen main.go

FROM ${BASE_IMAGE}
WORKDIR /
COPY --from=builder /workspace/loggen /bin

ENTRYPOINT ["/bin/loggen"]
