ARG BASE_IMAGE

FROM golang:1.24.1 AS builder

WORKDIR /workspace

RUN dpkg --add-architecture amd64 \
    && apt-get update \
    && apt-get install -y --no-install-recommends 

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/lobster-query/main.go main.go
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster main.go

FROM ${BASE_IMAGE} 
WORKDIR /
COPY --from=builder /workspace/lobster /bin

USER root
RUN mkdir -p /web/static/css
RUN mkdir -p /web/static/js
RUN mkdir -p /web/static/docs
RUN apk upgrade && apk add --no-cache libc6-compat git tcpdump bash vim iptables curl tzdata libcap iproute2 && cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime && echo "Asia/Seoul" > /etc/timezone
RUN /usr/sbin/setcap cap_net_bind_service=+ep /bin/lobster
RUN addgroup -g 10001 lobster && adduser -u 10001 -S lobster -G lobster
USER lobster

COPY web/* /web
COPY web/static/css/*.css /web/static/css
COPY web/static/js/*.js /web/static/js
COPY web/static/docs/query/*.json /web/static/docs

ENTRYPOINT ["/bin/lobster"]
