$(VIRTENGINE_DEVCACHE):
	@echo "creating .cache dir structure..."
	mkdir -p $@
	mkdir -p $(VIRTENGINE_DEVCACHE_BIN)
	mkdir -p $(VIRTENGINE_DEVCACHE_INCLUDE)
	mkdir -p $(VIRTENGINE_DEVCACHE_VERSIONS)
	mkdir -p $(VIRTENGINE_DEVCACHE_NODE_MODULES)
	mkdir -p $(VIRTENGINE_DEVCACHE)/run
cache: $(VIRTENGINE_DEVCACHE)

$(PROTOC_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc compiler v$(PROTOC_VERSION) ..."
	rm -f $(PROTOC)
	(cd /tmp; \
	curl -sOL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}"; \
	unzip -oq ${PROTOC_ZIP} -d $(VIRTENGINE_DEVCACHE) bin/protoc; \
	unzip -oq ${PROTOC_ZIP} -d $(VIRTENGINE_DEVCACHE) 'include/*'; \
	rm -f ${PROTOC_ZIP})
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC): $(PROTOC_VERSION_FILE)

$(PROTOC_GEN_COSMOS_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing protoc-gen-cosmos $(PROTOC_GEN_COSMOS_VERSION) ..."
	rm -f $(PROTOC_GEN_COSMOS)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@$(PROTOC_GEN_COSMOS_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(PROTOC_GEN_COSMOS): $(PROTOC_GEN_COSMOS_VERSION_FILE)

$(GRPC_GATEWAY_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "Installing protoc-gen-grpc-gateway $(GRPC_GATEWAY_VERSION) ..."
	rm -f $(GRPC_GATEWAY)
	curl -o "${VIRTENGINE_DEVCACHE_BIN}/protoc-gen-grpc-gateway" -L \
	"https://github.com/grpc-ecosystem/grpc-gateway/releases/download/${GRPC_GATEWAY_VERSION}/${GRPC_GATEWAY_BIN}"
	chmod +x "$(VIRTENGINE_DEVCACHE_BIN)/protoc-gen-grpc-gateway"
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GRPC_GATEWAY): $(GRPC_GATEWAY_VERSION_FILE)

$(STATIK_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "Installing statik $(STATIK_VERSION) ..."
	rm -f $(STATIK)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/rakyll/statik@$(STATIK_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(STATIK): $(STATIK_VERSION_FILE)

$(MODVENDOR_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing modvendor $(MODVENDOR_VERSION) ..."
	rm -f $(MODVENDOR)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/goware/modvendor@$(MODVENDOR_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MODVENDOR): $(MODVENDOR_VERSION_FILE)

$(GIT_CHGLOG_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing git-chglog $(GIT_CHGLOG_VERSION) ..."
	rm -f $(GIT_CHGLOG)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install github.com/git-chglog/git-chglog/cmd/git-chglog@$(GIT_CHGLOG_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GIT_CHGLOG): $(GIT_CHGLOG_VERSION_FILE)

$(MOCKERY_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing mockery $(MOCKERY_VERSION) ..."
	rm -f $(MOCKERY)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install -ldflags '-s -w -X github.com/vektra/mockery/v2/pkg/config.SemVer=$(MOCKERY_VERSION)' github.com/vektra/mockery/v2@v$(MOCKERY_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MOCKERY): $(MOCKERY_VERSION_FILE)

$(K8S_CODE_GEN_VERSION_FILE): $(VIRTENGINE_DEVCACHE) modvendor
	@echo "installing k8s code-generator $(K8S_CODE_GEN_VERSION) ..."
	rm -f $(K8S_GO_TO_PROTOBUF)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install $(ROOT_DIR)/vendor/k8s.io/code-generator/...
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
	chmod +x $(K8S_GENERATE_GROUPS)
$(K8S_GO_TO_PROTOBUF): $(K8S_CODE_GEN_VERSION_FILE)
$(K8S_GENERATE_GROUPS): $(K8S_CODE_GEN_VERSION_FILE)

.PHONY: $(KIND)
$(KIND):
	@echo "installing kind ..."
	$(GO) install sigs.k8s.io/kind

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

$(PROTOC_SWAGGER_GEN): $(VIRTENGINE_DEVCACHE)
ifeq (, $(shell which protoc-gen-swagger 2>/dev/null))
	@echo "installing protoc-gen-swagger $(PROTOC_SWAGGER_GEN_VERSION) ..."
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@$(PROTOC_SWAGGER_GEN_VERSION)
endif

cache-clean:
	rm -rf $(VIRTENGINE_DEVCACHE)
