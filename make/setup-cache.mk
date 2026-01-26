$(VIRTENGINE_DEVCACHE):
	@echo "creating .cache dir structure..."
	mkdir -p $@
	mkdir -p $(VIRTENGINE_DEVCACHE_BIN)
	mkdir -p $(VIRTENGINE_DEVCACHE_INCLUDE)
	mkdir -p $(VIRTENGINE_DEVCACHE_VERSIONS)
	mkdir -p $(VIRTENGINE_DEVCACHE_NODE_MODULES)
	mkdir -p $(VIRTENGINE_RUN_BIN)
cache: $(VIRTENGINE_DEVCACHE)

$(GIT_CHGLOG_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing git-chglog $(GIT_CHGLOG_VERSION) ..."
	rm -f $(GIT_CHGLOG)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install github.com/git-chglog/git-chglog/cmd/git-chglog@$(GIT_CHGLOG_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GIT_CHGLOG): $(GIT_CHGLOG_VERSION_FILE)

MOCKERY_MAJOR=$(shell $(SEMVER) get major $(MOCKERY_VERSION))
$(MOCKERY_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing mockery $(MOCKERY_VERSION) ..."
	rm -f $(MOCKERY)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install -ldflags '-s -w -X github.com/vektra/mockery/v$(MOCKERY_MAJOR)/pkg/config.SemVer=$(MOCKERY_VERSION)' github.com/vektra/mockery/v$(MOCKERY_MAJOR)@v$(MOCKERY_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MOCKERY): $(MOCKERY_VERSION_FILE)

GOLANGCI_LINT_MAJOR=$(shell $(SEMVER) get major $(GOLANGCI_LINT_VERSION))
$(GOLANGCI_LINT_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing golangci-lint $(GOLANGCI_LINT_VERSION) ..."
	rm -f $(GOLANGCI_LINT)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GOLANGCI_LINT): $(GOLANGCI_LINT_VERSION_FILE)

$(STATIK_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "Installing statik $(STATIK_VERSION) ..."
	rm -f $(STATIK)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install github.com/rakyll/statik@$(STATIK_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(STATIK): $(STATIK_VERSION_FILE)

$(COSMOVISOR_VERSION_FILE): $(VIRTENGINE_DEVCACHE)
	@echo "installing cosmovisor $(COSMOVISOR_VERSION) ..."
	rm -f $(COSMOVISOR)
	GOBIN=$(VIRTENGINE_DEVCACHE_BIN) $(GO) install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@$(COSMOVISOR_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(COSMOVISOR): $(COSMOVISOR_VERSION_FILE)

cache-clean:
	rm -rf $(VIRTENGINE_DEVCACHE)
