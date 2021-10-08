# VIRTENGINE_ROOT may not be set if environment does not support/use direnv
# in this case define it manually as well as all required env variables
ifndef VIRTENGINE_ROOT
VIRTENGINE_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../)
include $(VIRTENGINE_ROOT)/.env
endif

VIRTENGINE  			   := $(VIRTENGINE_DEVCACHE_BIN)/virtengine

BINS                       := $(VIRTENGINE)

GO                         := GO111MODULE=$(GO111MODULE) go

# setup .cache bins first in paths to have precedence over already installed same tools for system wide use
PATH                       := "$(PATH):$(VIRTENGINE_DEVCACHE_BIN):$(VIRTENGINE_DEVCACHE_NODE_BIN)"

BUF_VERSION                ?= 0.35.1
PROTOC_VERSION             ?= 3.13.0
PROTOC_GEN_COSMOS_VERSION  ?= v0.3.1
GRPC_GATEWAY_VERSION       := $(shell $(GO) list -mod=readonly -m -f '{{ .Version }}' github.com/grpc-ecosystem/grpc-gateway)
PROTOC_SWAGGER_GEN_VERSION := $(GRPC_GATEWAY_VERSION)
GOLANGCI_LINT_VERSION      ?= v1.38.0
GOLANG_VERSION             ?= 1.16.6
GOLANG_CROSS_VERSION       := v$(GOLANG_VERSION)
STATIK_VERSION             ?= v0.1.7
GIT_CHGLOG_VERSION         ?= v0.10.0
MODVENDOR_VERSION          ?= v0.3.0
MOCKERY_VERSION            ?= 2.5.1
K8S_CODE_GEN_VERSION       ?= v0.19.3

# <TOOL>_VERSION_FILE points to the marker file for the installed version.
# If <TOOL>_VERSION_FILE is changed, the binary will be re-downloaded.
PROTOC_VERSION_FILE            := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc/$(PROTOC_VERSION)
GRPC_GATEWAY_VERSION_FILE      := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-grpc-gateway/$(GRPC_GATEWAY_VERSION)
PROTOC_GEN_COSMOS_VERSION_FILE := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-cosmos/$(PROTOC_GEN_COSMOS_VERSION)
STATIK_VERSION_FILE            := $(VIRTENGINE_DEVCACHE_VERSIONS)/statik/$(STATIK_VERSION)
MODVENDOR_VERSION_FILE         := $(VIRTENGINE_DEVCACHE_VERSIONS)/modvendor/$(MODVENDOR_VERSION)
GIT_CHGLOG_VERSION_FILE        := $(VIRTENGINE_DEVCACHE_VERSIONS)/git-chglog/$(GIT_CHGLOG_VERSION)
MOCKERY_VERSION_FILE           := $(VIRTENGINE_DEVCACHE_VERSIONS)/mockery/v$(MOCKERY_VERSION)
K8S_CODE_GEN_VERSION_FILE      := $(VIRTENGINE_DEVCACHE_VERSIONS)/k8s-codegen/$(K8S_CODE_GEN_VERSION)

MODVENDOR                       = $(VIRTENGINE_DEVCACHE_BIN)/modvendor
SWAGGER_COMBINE                 = $(VIRTENGINE_DEVCACHE_NODE_BIN)/swagger-combine
PROTOC_SWAGGER_GEN             := $(VIRTENGINE_DEVCACHE_BIN)/protoc-swagger-gen
PROTOC                         := $(VIRTENGINE_DEVCACHE_BIN)/protoc
STATIK                         := $(VIRTENGINE_DEVCACHE_BIN)/statik
PROTOC_GEN_COSMOS              := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-cosmos
GRPC_GATEWAY                   := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-grpc-gateway
GIT_CHGLOG                     := $(VIRTENGINE_DEVCACHE_BIN)/git-chglog
MOCKERY                        := $(VIRTENGINE_DEVCACHE_BIN)/mockery
K8S_GENERATE_GROUPS            := $(VIRTENGINE_ROOT)/vendor/k8s.io/code-generator/generate-groups.sh
K8S_GO_TO_PROTOBUF             := $(VIRTENGINE_DEVCACHE_BIN)/go-to-protobuf
KIND                           := kind
NPM                            := npm

include $(VIRTENGINE_ROOT)/make/setup-cache.mk
