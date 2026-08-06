# syntax=docker/dockerfile:1.7
# The deployed image is always referenced by digest in Kubernetes. Builders
# should pass a digest-pinned Go base via GO_BUILDER_IMAGE in release CI.
ARG GO_BUILDER_IMAGE=golang:1.25.8-alpine
FROM ${GO_BUILDER_IMAGE} AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY go.mod go.sum ./
COPY sdk/go/go.mod sdk/go/go.sum sdk/go/
COPY sdk/specs/go.mod sdk/specs/
RUN go mod download
COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown
ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}
RUN go build -trimpath -buildvcs=false \
  -ldflags "-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}" \
  -o /out/veid-inference ./cmd/inference-sidecar

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.title="VirtEngine VEID Inference" \
      org.opencontainers.image.description="Deterministic VEID inference sidecar" \
      org.opencontainers.image.source="https://github.com/virtengine-network/node"
COPY --from=build /out/veid-inference /usr/local/bin/veid-inference
USER 65532:65532
EXPOSE 50051 9090
ENTRYPOINT ["/usr/local/bin/veid-inference"]
