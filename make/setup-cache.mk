# ── VE_DEVCACHE guard ─────────────────────────────────────────────────────────
# Every rule in this file derives its paths from VE_DEVCACHE, which direnv
# supplies through .envrc -> .env. When that variable is empty (a bare terminal
# with no direnv) these targets fail silently: `make cache` reports success
# having done nothing, `make cache-clean` expands to a bare `rm -rf`, and the
# version-marker rules write to root-relative paths (e.g. `touch /gitleaks/8.22.1`
# -> C:\gitleaks\8.22.1 on Windows).
#
# Deliberately NOT conditional on VE_DIRENV_SET: make/init.mk:19-21 force-sets
# that variable on Windows, so it cannot be used to detect a missing dev cache.
# Also deliberately NOT a parse-time $(error): this file is included by
# make/init.mk:134 for every invocation, so a parse-time error would break
# unrelated targets (`make bins`, `make proto-generate`) in an unconfigured shell.
#
# The guard is applied two ways, because the two kinds of target need different
# treatment:
#   * phony targets (`cache`, `cache-clean`) take the phony `require-devache`
#     prerequisite;
#   * file targets (the version markers) run $(DEVACHE_GUARD) as their first
#     recipe line -- a phony prerequisite there would mark the target
#     permanently out of date and re-install every tool on every run.
DEVACHE_ERROR = VE_DEVCACHE is empty - run "direnv allow" or export VE_DEVCACHE
DEVACHE_GUARD = $(if $(VE_DEVCACHE),@:,$(error $(DEVACHE_ERROR)))

.PHONY: require-devache
require-devache:
	@$(if $(VE_DEVCACHE),,$(error $(DEVACHE_ERROR)))

# PowerShell does not understand MSYS-style paths (/c/Users/...): it resolves
# them relative to the current drive (C:\c\Users\...), creating a stray tree.
# Convert the cache binary path to native form once, at parse time.
ifeq ($(OS),Windows_NT)
VE_DEVCACHE_BIN_NATIVE := $(shell cygpath -w "$(VE_DEVCACHE_BIN)" 2>/dev/null || echo "$(VE_DEVCACHE_BIN)")
endif

.PHONY: cache cache-clean
$(VE_DEVCACHE):
	$(DEVACHE_GUARD)
	@echo "creating .cache dir structure..."
	mkdir -p $(VE_DEVCACHE)
	mkdir -p $(VE_DEVCACHE_BIN)
	mkdir -p $(VE_DEVCACHE_INCLUDE)
	mkdir -p $(VE_DEVCACHE_VERSIONS)
	mkdir -p $(VE_DEVCACHE_NODE_MODULES)
	mkdir -p $(VE_RUN_BIN)
cache: require-devache $(VE_DEVCACHE)

# The version markers below take the cache directory as an ORDER-ONLY
# prerequisite (`| $(VE_DEVCACHE)`): it has to exist, but its mtime must never
# make a marker look stale. As a normal prerequisite it did exactly that -- the
# gitleaks recipe writes and deletes .cache/gitleaks.zip, bumping .cache's mtime
# past the five markers that were touched earlier in the same run, so every
# second `make tools` re-installed all five Go tools.
$(GIT_CHGLOG_VERSION_FILE): | $(VE_DEVCACHE)
	$(DEVACHE_GUARD)
	@echo "installing git-chglog $(GIT_CHGLOG_VERSION) ..."
	rm -f $(GIT_CHGLOG)
	GOBIN=$(VE_DEVCACHE_BIN) go install github.com/git-chglog/git-chglog/cmd/git-chglog@$(GIT_CHGLOG_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GIT_CHGLOG): $(GIT_CHGLOG_VERSION_FILE)

# A module path carrying a major (`github.com/vektra/mockery/v3`, `.../golangci-lint/v2`)
# derives that major from the tool version via $(SEMVER), which make/init.mk defaults
# when the environment (i.e. direnv) does not supply it. A SEMVER that exists but does
# not resolve -- a stale path exported by hand, a script that is not runnable -- still
# yields an empty major, and `go install` would then fail on a malformed module path
# (github.com/vektra/mockery/v@v3.5.0) that names neither SEMVER nor the tool. Check it
# here instead, before the recipe's `rm -f` can remove a working binary.
# $(1) = tool name, $(2) = derived major.
define require-module-major
case "$(2)" in ''|*[!0-9]*) echo "ERROR: $(1): could not resolve the module major from SEMVER='$(SEMVER)' (got '$(2)'). Check that script/semver.sh exists and is runnable." >&2; exit 2;; esac
endef

MOCKERY_MAJOR=$(shell $(SEMVER) get major $(MOCKERY_VERSION))
$(MOCKERY_VERSION_FILE): | $(VE_DEVCACHE)
	$(DEVACHE_GUARD)
	@$(call require-module-major,mockery,$(MOCKERY_MAJOR))
	@echo "installing mockery $(MOCKERY_VERSION) ..."
	rm -f $(MOCKERY)
	GOBIN=$(VE_DEVCACHE_BIN) go install -ldflags '-s -w -X github.com/vektra/mockery/v$(MOCKERY_MAJOR)/pkg/config.SemVer=$(MOCKERY_VERSION)' github.com/vektra/mockery/v$(MOCKERY_MAJOR)@v$(MOCKERY_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(MOCKERY): $(MOCKERY_VERSION_FILE)

GOLANGCI_LINT_MAJOR=$(shell $(SEMVER) get major $(GOLANGCI_LINT_VERSION))
$(GOLANGCI_LINT_VERSION_FILE): | $(VE_DEVCACHE)
	$(DEVACHE_GUARD)
	@$(call require-module-major,golangci-lint,$(GOLANGCI_LINT_MAJOR))
	@echo "installing golangci-lint $(GOLANGCI_LINT_VERSION) ..."
	rm -f $(GOLANGCI_LINT)
	GOBIN=$(VE_DEVCACHE_BIN) go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GOLANGCI_LINT): $(GOLANGCI_LINT_VERSION_FILE)

$(STATIK_VERSION_FILE): | $(VE_DEVCACHE)
	$(DEVACHE_GUARD)
	@echo "Installing statik $(STATIK_VERSION) ..."
	rm -f $(STATIK)
	GOBIN=$(VE_DEVCACHE_BIN) $(GO) install github.com/rakyll/statik@$(STATIK_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(STATIK): $(STATIK_VERSION_FILE)

$(COSMOVISOR_VERSION_FILE): | $(VE_DEVCACHE)
	$(DEVACHE_GUARD)
	@echo "installing cosmovisor $(COSMOVISOR_VERSION) ..."
	rm -f $(COSMOVISOR)
	GOBIN=$(VE_DEVCACHE_BIN) $(GO) install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@$(COSMOVISOR_VERSION)
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(COSMOVISOR): $(COSMOVISOR_VERSION_FILE)

$(GITLEAKS_VERSION_FILE): | $(VE_DEVCACHE)
	$(DEVACHE_GUARD)
	@echo "installing gitleaks $(GITLEAKS_VERSION) ..."
	rm -f $(GITLEAKS)
ifeq ($(OS),Windows_NT)
	# \$$url (not $$url): the shell must pass a literal $url through to
	# PowerShell. With $$url the shell expands the unset $url to empty and
	# PowerShell receives `= 'https://...'`. Native paths for the same reason:
	# Invoke-WebRequest/Expand-Archive do not understand /c/Users/...
	# The archive is staged in $$env:TEMP, not in the cache root: a transient zip
	# there changed .cache's mtime, which is what used to invalidate every
	# version marker touched earlier in the same run.
	powershell -Command "\$$url = 'https://github.com/gitleaks/gitleaks/releases/download/v$(GITLEAKS_VERSION)/gitleaks_$(GITLEAKS_VERSION)_windows_x64.zip'; \
		\$$zip = Join-Path \$$env:TEMP 'gitleaks.zip'; \
		Invoke-WebRequest -Uri \$$url -OutFile \$$zip; \
		Expand-Archive -Path \$$zip -DestinationPath '$(VE_DEVCACHE_BIN_NATIVE)' -Force; \
		Remove-Item \$$zip"
else ifeq ($(UNAME_OS),Darwin)
	curl -sSfL "https://github.com/gitleaks/gitleaks/releases/download/v$(GITLEAKS_VERSION)/gitleaks_$(GITLEAKS_VERSION)_darwin_$(UNAME_ARCH).tar.gz" | tar -xz -C $(VE_DEVCACHE_BIN) gitleaks
else
	curl -sSfL "https://github.com/gitleaks/gitleaks/releases/download/v$(GITLEAKS_VERSION)/gitleaks_$(GITLEAKS_VERSION)_linux_$(UNAME_ARCH).tar.gz" | tar -xz -C $(VE_DEVCACHE_BIN) gitleaks
endif
	rm -rf "$(dir $@)"
	mkdir -p "$(dir $@)"
	touch $@
$(GITLEAKS): $(GITLEAKS_VERSION_FILE)

# Install every build tool declared in make/init.mk as a <TOOL> = <TOOL>_VERSION_FILE
# pair. Nothing used to depend on $(GITLEAKS), so gitleaks had no install path at
# all and .githooks/pre-commit:299 pointed at the non-existent `make setup-cache`.
.PHONY: tools
tools: $(GIT_CHGLOG) $(MOCKERY) $(GOLANGCI_LINT) $(STATIK) $(COSMOVISOR) $(GITLEAKS)

cache-clean: require-devache
	rm -rf $(VE_DEVCACHE)
