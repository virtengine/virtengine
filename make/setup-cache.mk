$(VE_DEVCACHE):
	@echo "creating .cache dir structure..."
	mkdir -p $@
	mkdir -p $(VE_DEVCACHE_BIN)
	mkdir -p $(VE_DEVCACHE_INCLUDE)
	mkdir -p $(VE_DEVCACHE_VERSIONS)
	mkdir -p $(VE_DEVCACHE_NODE_MODULES)
	mkdir -p $(VE_RUN_BIN)
cache: $(VE_DEVCACHE)

$(GIT_CHGLOG_VERSION_FILE): $(VE_DEVCACHE)
	@echo "installing git-chglog $(GIT_CHGLOG_VERSION) ..."
	rm -f $(GIT_CHGLOG)
	GOBIN=$(VE_DEVCACHE_BIN) go install github.com/git-chglog/git-chglog/cmd/git-chglog@$(GIT_CHGLOG_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GIT_CHGLOG): $(GIT_CHGLOG_VERSION_FILE)

MOCKERY_MAJOR=$(shell $(SEMVER) get major $(MOCKERY_VERSION))
$(MOCKERY_VERSION_FILE): $(VE_DEVCACHE)
	@echo "installing mockery $(MOCKERY_VERSION) ..."
	rm -f $(MOCKERY)
	GOBIN=$(VE_DEVCACHE_BIN) go install -ldflags '-s -w -X github.com/vektra/mockery/v$(MOCKERY_MAJOR)/pkg/config.SemVer=$(MOCKERY_VERSION)' github.com/vektra/mockery/v$(MOCKERY_MAJOR)@v$(MOCKERY_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MOCKERY): $(MOCKERY_VERSION_FILE)

GOLANGCI_LINT_MAJOR=$(shell $(SEMVER) get major $(GOLANGCI_LINT_VERSION))
$(GOLANGCI_LINT_VERSION_FILE): $(VE_DEVCACHE)
	@echo "installing golangci-lint $(GOLANGCI_LINT_VERSION) ..."
	rm -f $(GOLANGCI_LINT)
	GOBIN=$(VE_DEVCACHE_BIN) go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GOLANGCI_LINT): $(GOLANGCI_LINT_VERSION_FILE)

$(STATIK_VERSION_FILE): $(VE_DEVCACHE)
	@echo "Installing statik $(STATIK_VERSION) ..."
	rm -f $(STATIK)
	GOBIN=$(VE_DEVCACHE_BIN) $(GO) install github.com/rakyll/statik@$(STATIK_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(STATIK): $(STATIK_VERSION_FILE)

$(COSMOVISOR_VERSION_FILE): $(VE_DEVCACHE)
	@echo "installing cosmovisor $(COSMOVISOR_VERSION) ..."
	rm -f $(COSMOVISOR)
	GOBIN=$(VE_DEVCACHE_BIN) $(GO) install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@$(COSMOVISOR_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(COSMOVISOR): $(COSMOVISOR_VERSION_FILE)

$(GITLEAKS_VERSION_FILE): $(VE_DEVCACHE)
	@echo "installing gitleaks $(GITLEAKS_VERSION) ..."
	rm -f $(GITLEAKS)
ifeq ($(OS),Windows_NT)
	powershell -Command "$$url = 'https://github.com/gitleaks/gitleaks/releases/download/v$(GITLEAKS_VERSION)/gitleaks_$(GITLEAKS_VERSION)_windows_x64.zip'; \
		Invoke-WebRequest -Uri $$url -OutFile $(VE_DEVCACHE)/gitleaks.zip; \
		Expand-Archive -Path $(VE_DEVCACHE)/gitleaks.zip -DestinationPath $(VE_DEVCACHE_BIN) -Force; \
		Remove-Item $(VE_DEVCACHE)/gitleaks.zip"
else ifeq ($(UNAME_OS),Darwin)
	curl -sSfL "https://github.com/gitleaks/gitleaks/releases/download/v$(GITLEAKS_VERSION)/gitleaks_$(GITLEAKS_VERSION)_darwin_$(UNAME_ARCH).tar.gz" | tar -xz -C $(VE_DEVCACHE_BIN) gitleaks
else
	curl -sSfL "https://github.com/gitleaks/gitleaks/releases/download/v$(GITLEAKS_VERSION)/gitleaks_$(GITLEAKS_VERSION)_linux_$(UNAME_ARCH).tar.gz" | tar -xz -C $(VE_DEVCACHE_BIN) gitleaks
endif
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GITLEAKS): $(GITLEAKS_VERSION_FILE)

cache-clean:
	rm -rf $(VE_DEVCACHE)
