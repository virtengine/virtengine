# syntax=docker/dockerfile:1.7
FROM golang:1.25.5-bookworm@sha256:d9132cce84391efab786495288756d60e1da215b1f94e87860aeefc3d4c45b6d

ARG TARGETARCH
ARG PROTOC_VERSION=29.1
ARG PROTOC_AMD64_SHA256=00c83fe9722d85e96c81b941b29f17a744b33b4ce66e0f18009fd8937de22c60
ARG PROTOC_ARM64_SHA256=1f74a3f3355de7c0666bc125611c13532c2598f853521d0d3e621a5b09f24799
ARG NODE_VERSION=24.12.0
ARG NODE_AMD64_SHA256=bdebee276e58d0ef5448f3d5ac12c67daa963dd5e0a9bb621a53d1cefbc852fd
ARG NODE_ARM64_SHA256=a06d42807fb500f7459e5f3fa6cb431447352826ee6f07e14adfeec58a1b3210

RUN apt-get update \
    && apt-get install --yes --no-install-recommends ca-certificates curl unzip xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN case "${TARGETARCH}" in \
      amd64) protoc_arch=x86_64; protoc_sha="${PROTOC_AMD64_SHA256}"; node_arch=x64; node_sha="${NODE_AMD64_SHA256}" ;; \
      arm64) protoc_arch=aarch_64; protoc_sha="${PROTOC_ARM64_SHA256}"; node_arch=arm64; node_sha="${NODE_ARM64_SHA256}" ;; \
      *) echo "unsupported TARGETARCH=${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && curl --fail --location --silent --show-error \
      "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-${protoc_arch}.zip" \
      --output /tmp/protoc.zip \
    && echo "${protoc_sha}  /tmp/protoc.zip" | sha256sum --check --strict \
    && unzip -q /tmp/protoc.zip -d /usr/local bin/protoc include \
    && rm /tmp/protoc.zip \
    && curl --fail --location --silent --show-error \
      "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-${node_arch}.tar.xz" \
      --output /tmp/node.tar.xz \
    && echo "${node_sha}  /tmp/node.tar.xz" | sha256sum --check --strict \
    && tar -xJf /tmp/node.tar.xz --strip-components=1 -C /usr/local \
    && rm /tmp/node.tar.xz

WORKDIR /src/sdk/generation
COPY sdk/generation/go.mod sdk/generation/go.sum sdk/generation/tools.go ./
RUN GOWORK=off go mod verify \
    && GOWORK=off go install github.com/bufbuild/buf/cmd/buf \
    && GOWORK=off go install github.com/cosmos/gogoproto/protoc-gen-gocosmos \
    && GOWORK=off go install github.com/goware/modvendor \
    && GOWORK=off go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
    && GOWORK=off go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

ENV PATH="/go/bin:/usr/local/bin:${PATH}"
WORKDIR /src
ENTRYPOINT ["/src/sdk/generation/generate.sh"]
