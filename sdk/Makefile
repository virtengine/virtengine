UNAME_OS              := $(shell uname -s)
UNAME_ARCH            := $(shell uname -m)

ifeq (0, $(shell id -u))
$(warning "make was started with superuser privileges. it may cause issues with direnv")
endif

ifeq (, $(shell which direnv))
$(error "No direnv in $(PATH), consider installing. https://direnv.net")
endif

ifneq (1, $(VIRTENGINE_DIRENV_SET))
$(error "no envrc detected. might need to run \"direnv allow\"")
endif

# VIRTENGINE_ROOT may not be set if environment does not support/use direnv
# in this case define it manually as well as all required env variables
ifndef VIRTENGINE_ROOT
$(error "VIRTENGINE_ROOT is not set. might need to run \"direnv allow\"")
endif

ifeq (, $(GOTOOLCHAIN))
$(error "GOTOOLCHAIN is not set")
endif

ifeq ($(GO111MODULE),off)
else
	GOMOD=readonly
endif

ifneq ($(GOWORK),off)
	ifneq ($(shell test -e "$(GOWORK)" && echo yes),yes)
		GOWORK=off
	endif
endif

ifneq ($(GOWORK),off)
	ifeq ($(GOMOD),$(filter $(GOMOD),mod ""))
$(error '-mod may only be set to readonly or vendor when in workspace mode, but it is set to ""')
	endif
endif

ifeq ($(GOMOD),vendor)
	ifneq ($(wildcard ./vendor/.),)
$(error "go -mod is in vendor mode but vendor dir has not been found. consider to run go mod vendor")
	endif
endif

GO_ROOT                       := go
TS_ROOT                       := $(VIRTENGINE_TS_ROOT)

BUMP_MOD                      ?=

GO                           := GO111MODULE=$(GO111MODULE) go
GO_MOD_NAME                  := $(shell cd $(GO_ROOT); GOWORK=off go list -m | head -n 1)

BUF_VERSION                     ?= 1.47.2
PROTOC_VERSION                  ?= 29.1
GOGOPROTO_VERSION               ?= $(shell cd $(GO_ROOT); $(GO) list -mod=readonly -m -f '{{ .Version }}' github.com/cosmos/gogoproto)
PROTOC_GEN_GOCOSMOS_VERSION     ?= $(GOGOPROTO_VERSION)
PROTOC_GEN_GO_PULSAR_VERSION    ?= $(shell cd $(GO_ROOT); $(GO) list -mod=readonly -m -f '{{ .Version }}' github.com/cosmos/cosmos-proto)
PROTOC_GEN_GO_VERSION           ?= $(shell cd $(GO_ROOT); $(GO) list -mod=readonly -m -f '{{ .Version }}' google.golang.org/protobuf)
PROTOC_GEN_PROST_VERSION        := v1.62.0
PROTOC_GEN_GRPC_GATEWAY_VERSION := $(shell cd $(GO_ROOT); $(GO) list -mod=readonly -m -f '{{ .Version }}' github.com/grpc-ecosystem/grpc-gateway)
PROTOC_GEN_DOC_VERSION          := $(shell cd $(GO_ROOT); $(GO) list -mod=readonly -m -f '{{ .Version }}' github.com/pseudomuto/protoc-gen-doc)
PROTOC_GEN_SWAGGER_VERSION      := $(PROTOC_GEN_GRPC_GATEWAY_VERSION)
MODVENDOR_VERSION               ?= v0.5.0
MOCKERY_VERSION                 ?= 2.52.2
GOLANGCI_LINT_VERSION           ?= v2.3.0
SEMVER_VERSION                  ?= v1.3.0

BUF_VERSION_FILE                     := $(VIRTENGINE_DEVCACHE_VERSIONS)/buf/$(BUF_VERSION)
PROTOC_VERSION_FILE                  := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc/$(PROTOC_VERSION)
GOGOPROTO_VERSION_FILE               := $(VIRTENGINE_DEVCACHE_VERSIONS)/gogoproto/$(GOGOPROTO_VERSION)
PROTOC_GEN_GOCOSMOS_VERSION_FILE     := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-gocosmos/$(PROTOC_GEN_GOCOSMOS_VERSION)
PROTOC_GEN_GO_PULSAR_VERSION_FILE    := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-go-pulsar/$(PROTOC_GEN_GO_PULSAR_VERSION)
PROTOC_GEN_GO_VERSION_FILE           := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-go/$(PROTOC_GEN_GO_VERSION)
PROTOC_GEN_PROST_VERSION_FILE        := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-prost/$(PROTOC_GEN_PROST_VERSION)
PROTOC_GEN_GOGO_VERSION_FILE         := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-gogo/$(GOGOPROTO_VERSION)
PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-grpc-gateway/$(PROTOC_GEN_GRPC_GATEWAY_VERSION)
PROTOC_GEN_SWAGGER_VERSION_FILE      := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-swagger/$(PROTOC_GEN_SWAGGER_VERSION)
PROTOC_GEN_DOC_VERSION_FILE          := $(VIRTENGINE_DEVCACHE_VERSIONS)/protoc-gen-doc/$(PROTOC_GEN_DOC_VERSION)
MODVENDOR_VERSION_FILE               := $(VIRTENGINE_DEVCACHE_VERSIONS)/modvendor/$(MODVENDOR_VERSION)
GIT_CHGLOG_VERSION_FILE              := $(VIRTENGINE_DEVCACHE_VERSIONS)/git-chglog/$(GIT_CHGLOG_VERSION)
MOCKERY_VERSION_FILE                 := $(VIRTENGINE_DEVCACHE_VERSIONS)/mockery/v$(MOCKERY_VERSION)
GOLANGCI_LINT_VERSION_FILE           := $(VIRTENGINE_DEVCACHE_VERSIONS)/golangci-lint/$(GOLANGCI_LINT_VERSION)
SEMVER_VERSION_FILE                  := $(VIRTENGINE_DEVCACHE_VERSIONS)/semver/$(SEMVER_VERSION)

BUF                              := $(VIRTENGINE_DEVCACHE_BIN)/buf
PROTOC                           := $(VIRTENGINE_DEVCACHE_BIN)/protoc
PROTOC_GEN_GOCOSMOS              := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-gocosmos
GOGOPROTO                        := $(VIRTENGINE_DEVCACHE_BIN)/gogoproto
PROTOC_GEN_GO_PULSAR             := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-go-pulsar
PROTOC_GEN_GO                    := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-go
PROTOC_GEN_PROST                 := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-prost
PROTOC_GEN_GOGO                  := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-gogo
PROTOC_GEN_GRPC_GATEWAY          := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-grpc-gateway
PROTOC_GEN_SWAGGER               := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-swagger
PROTOC_GEN_DOC                   := $(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-doc
MODVENDOR                        := $(VIRTENGINE_DEVCACHE_BIN)/modvendor
GIT_CHGLOG                       := $(VIRTENGINE_DEVCACHE_BIN)/git-chglog
SWAGGER_COMBINE                  := $(VIRTENGINE_DEVCACHE_NODE_BIN)/swagger-combine
MOCKERY                          := $(VIRTENGINE_DEVCACHE_BIN)/mockery
GOLANGCI_LINT                    := $(VIRTENGINE_DEVCACHE_BIN)/golangci-lint
SEMVER                           := $(VIRTENGINE_DEVCACHE_BIN)/semver

GOLANGCI_LINT_RUN                := $(GOLANGCI_LINT) run
GOLINT                           := $(GOLANGCI_LINT_RUN) ./... --disable-all --timeout=5m --enable

DOCKER_RUN            := docker run --rm -v $(shell pwd):/workspace -w /workspace
DOCKER_BUF            := $(DOCKER_RUN) bufbuild/buf:$(BUF_VERSION)

GO_MODULES            ?= $(shell find * -name go.mod -not -path "specs/*" -not -path "testdata/*" -not -path "mock/*" -exec dirname {} \;)
GO_TEST_DIRS          ?= ./...

# Use SDK-local make files (not root make/)
SDK_MAKE_DIR          := $(dir $(lastword $(MAKEFILE_LIST)))make

include $(SDK_MAKE_DIR)/setup-cache.mk
include $(SDK_MAKE_DIR)/mod.mk
include $(SDK_MAKE_DIR)/test.mk
include $(SDK_MAKE_DIR)/codegen.mk
include $(SDK_MAKE_DIR)/lint.mk
include $(SDK_MAKE_DIR)/release-ts.mk
include $(SDK_MAKE_DIR)/code-style.mk

.PHONY: bump-%
bump-%:
	@./script/tools.sh bump "$*" "$(BUMP_MOD)"

.PHONY: clean
clean:
	rm -rf $(VIRTENGINE_DEVCACHE)
	rm -rf $(VIRTENGINE_TS_ROOT)/node_modules
	rm -rf $(VIRTENGINE_TS_ROOT)/dist
