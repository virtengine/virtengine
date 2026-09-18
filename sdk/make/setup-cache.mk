###############################################################################
###                           Protobuf                                    ###
###############################################################################
ifeq ($(UNAME_OS),Linux)
ifeq ($(UNAME_ARCH),aarch64)
	PROTOC_ZIP ?= protoc-${PROTOC_VERSION}-linux-aarch_64.zip
else
	PROTOC_ZIP ?= protoc-${PROTOC_VERSION}-linux-$(UNAME_ARCH).zip
endif
endif
ifeq ($(UNAME_OS),Darwin)
	PROTOC_ZIP ?= protoc-${PROTOC_VERSION}-osx-universal_binary.zip
endif

$(VIRTENGINE_DEVCACHE):
	@echo "creating .cache dir structure..."
	mkdir -p $@
	mkdir -p $(VIRTENGINE_DEVCACHE_BIN)
	mkdir -p $(VIRTENGINE_DEVCACHE_INCLUDE)
	mkdir -p $(VIRTENGINE_DEVCACHE_VERSIONS)
	mkdir -p $(VIRTENGINE_DEVCACHE_NODE_MODULES)
	mkdir -p $(VIRTENGINE_DEVCACHE)/run

cache: $(VIRTENGINE_DEVCACHE)

$(BUF_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing buf v$(BUF_VERSION) ..."
	rm -f $(BUF)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(BUF): $(BUF_VERSION_FILE)

$(PROTOC_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc compiler v$(PROTOC_VERSION) ..."
	rm -f $(PROTOC)
	(cd /tmp; \
	curl -sOL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}"; \
	unzip -oq ${PROTOC_ZIP} -d $(VIRTENGINE_DEVCACHE) bin/protoc; \
	rm -f ${PROTOC_ZIP})
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC): $(PROTOC_VERSION_FILE)

$(PROTOC_GEN_GOCOSMOS_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-gocosmos $(PROTOC_GEN_GOCOSMOS_VERSION) ..."
	rm -f $(PROTOC_GEN_GOCOSMOS)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/cosmos/gogoproto/protoc-gen-gocosmos)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_GOCOSMOS): $(PROTOC_GEN_GOCOSMOS_VERSION_FILE) #modvendor

$(PROTOC_GEN_GOGO_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-gogo $(PROTOC_GEN_GOCOSMOS_VERSION) ..."
	rm -f $(PROTOC_GEN_GOGO)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/cosmos/gogoproto/protoc-gen-gogo)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_GOGO): $(PROTOC_GEN_GOGO_VERSION_FILE) #modvendor

$(GOGOPROTO_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing gogoproto $(GOGOPROTO_VERSION) ..."
	rm -f $(GOGOPROTO)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/cosmos/gogoproto/gogoproto)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GOGOPROTO): $(GOGOPROTO_VERSION_FILE)

$(PROTOC_GEN_GO_PULSAR_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-go-pulsar $(PROTOC_GEN_GO_PULSAR_VERSION) ..."
	rm -f $(PROTOC_GEN_GO_PULSAR)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@$(PROTOC_GEN_GO_PULSAR_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_GO_PULSAR): $(PROTOC_GEN_GO_PULSAR_VERSION_FILE)

$(PROTOC_GEN_GO_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-go $(PROTOC_GEN_GO_VERSION) ..."
	rm -f $(PROTOC_GEN_GO)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_GO): $(PROTOC_GEN_GO_VERSION_FILE)

$(PROTOC_GEN_PROST_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-prost $(PROTOC_GEN_PROST_VERSION) ..."
	rm -f $(PROTOC_GEN_PROST)
	cargo install protoc-gen-prost --root $(VIRTENGINE_DEVCACHE)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_PROST): $(PROTOC_GEN_PROST_VERSION_FILE)

$(PROTOC_GEN_DOC_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-doc $(PROTOC_GEN_DOC_VERSION) ..."
	rm -f $(PROTOC_GEN_DOC)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@$(PROTOC_GEN_DOC_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_DOC): $(PROTOC_GEN_DOC_VERSION_FILE)

$(PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "Installing protoc-gen-grpc-gateway $(PROTOC_GEN_GRPC_GATEWAY_VERSION) ..."
	rm -f $(PROTOC_GEN_GRPC_GATEWAY)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_GRPC_GATEWAY): $(PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE)

$(PROTOC_GEN_SWAGGER_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "Installing protoc-gen-swagger $(PROTOC_GEN_SWAGGER_VERSION) ..."
	rm -f $(PROTOC_GEN_SWAGGER)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@$(PROTOC_GEN_SWAGGER_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_SWAGGER): $(PROTOC_GEN_SWAGGER_VERSION_FILE)

$(MODVENDOR_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing modvendor $(MODVENDOR_VERSION) ..."
	rm -f $(MODVENDOR)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/goware/modvendor@$(MODVENDOR_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MODVENDOR): $(MODVENDOR_VERSION_FILE)

$(GIT_CHGLOG_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing git-chglog $(GIT_CHGLOG_VERSION) ..."
	rm -f $(GIT_CHGLOG)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install github.com/git-chglog/git-chglog/cmd/git-chglog@$(GIT_CHGLOG_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GIT_CHGLOG): $(GIT_CHGLOG_VERSION_FILE)

# A module path carrying a major (`github.com/vektra/mockery/v<N>`,
# `.../golangci-lint/v<N>`) derives that major from $(SEMVER), which sdk/Makefile
# defaults to the SDK's own script/semver.sh when the environment (i.e. direnv) does
# not supply it. A SEMVER that exists but does not resolve -- a stale path, a script
# that is not runnable -- still yields an empty major, and `go install` would then
# fail on a malformed module path (github.com/vektra/mockery/v@vX.Y.Z,
# github.com/golangci/golangci-lint/v/cmd/...) that names neither SEMVER nor the
# tool. Check it in the recipe instead, before `rm -f` can remove a working binary.
# Recipe-time, never a parse-time $(error): sdk/Makefile includes this file for every
# target, so a parse-time error would break unrelated targets.
# $(1) = tool name, $(2) = derived major.
define require-module-major
case "$(2)" in ''|*[!0-9]*) echo "ERROR: $(1): could not resolve the module major from SEMVER='$(SEMVER)' (got '$(2)'). Check that script/semver.sh exists and is runnable." >&2; exit 2;; esac
endef

MOCKERY_MAJOR=$(shell $(SEMVER) get major $(MOCKERY_VERSION))

$(MOCKERY_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@$(call require-module-major,mockery,$(MOCKERY_MAJOR))
	@echo "installing mockery $(MOCKERY_VERSION) ..."
	rm -f $(MOCKERY)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install -ldflags '-s -w -X github.com/vektra/mockery/v$(MOCKERY_MAJOR)/pkg/config.SemVer=$(MOCKERY_VERSION)' github.com/vektra/mockery/v$(MOCKERY_MAJOR)@v$(MOCKERY_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MOCKERY): $(MOCKERY_VERSION_FILE)

GOLANGCI_LINT_MAJOR=$(shell $(SEMVER) get major $(GOLANGCI_LINT_VERSION))

$(GOLANGCI_LINT_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@$(call require-module-major,golangci-lint,$(GOLANGCI_LINT_MAJOR))
	@echo "installing golangci-lint $(GOLANGCI_LINT_VERSION) ..."
	rm -f $(GOLANGCI_LINT)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GOLANGCI_LINT): $(GOLANGCI_LINT_VERSION_FILE)

$(SEMVER_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing semver $(SEMVER_VERSION) ..."
	rm -f $(SEMVER_BIN)
	(cd $(GO_ROOT); GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install github.com/troian/semver/cmd/semver@$(SEMVER_VERSION))
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(SEMVER_BIN): $(SEMVER_VERSION_FILE)

semver: $(SEMVER_BIN)

$(NPM):
ifeq (, $(shell which $(NPM) 2>/dev/null))
	$(error "npm installation required")
endif

$(SWAGGER_COMBINE): $(VIRTENGINE_DEVCACHE) $(NPM)
ifeq (, $(shell which swagger-combine 2>/dev/null))
	@echo "Installing swagger-combine..."
	npm install swagger-combine --prefix $(VIRTENGINE_DEVCACHE_NODE_MODULES)
	chmod +x $(SWAGGER_COMBINE)
else
	@echo "swagger-combine already installed; skipping..."
endif

$(VIRTENGINE_TS_NODE_MODULES): $(VIRTENGINE_TS_PACKAGE_FILE)
	@echo "installing node modules..."
	cd $(VIRTENGINE_TS_ROOT) && npm ci
	@echo "node modules installed."
