# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" -o /out/slnxsync .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/slnxsync /slnxsync

USER 65532:65532
ENTRYPOINT ["/slnxsync"]
