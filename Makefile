APP_DIR               := ./app

GOBIN                 ?= $(shell go env GOPATH)/bin

ifneq (,$(wildcard _run/kube))
KIND_APP_IP           ?= $(shell make -sC _run/kube kind-k8s-ip)
KIND_APP_PORT         ?= $(shell make -sC _run/kube app-http-port)
else
KIND_APP_IP           ?=
KIND_APP_PORT         ?=
endif
KIND_VARS             ?= KUBE_INGRESS_IP="$(KIND_APP_IP)" KUBE_INGRESS_PORT="$(KIND_APP_PORT)"

include make/init.mk

.DEFAULT_GOAL         := bins

ifeq ($(OS),Windows_NT)
DOCKER_RUN            := docker run --rm -v $(CURDIR):/workspace -w /workspace
else
DOCKER_RUN            := docker run --rm -v $(shell pwd):/workspace -w /workspace
endif
GOLANGCI_LINT_RUN     := $(GOLANGCI_LINT) run
LINT                   = $(GOLANGCI_LINT_RUN) ./... --disable-all --deadline=5m --enable

GORELEASER_CONFIG     ?= .goreleaser.yaml

GIT_HEAD_COMMIT_LONG  := $(shell git log -1 --format='%H')
GIT_HEAD_COMMIT_SHORT := $(shell git rev-parse --short HEAD)
GIT_HEAD_ABBREV       := $(shell git rev-parse --abbrev-ref HEAD)

ifeq ($(OS),Windows_NT)
IS_PREREL             := false
IS_MAINNET            := false
else
IS_PREREL             := $(shell $(ROOT_DIR)/script/is_prerelease.sh "$(RELEASE_TAG)" && echo "true" || echo "false")
IS_MAINNET            := $(shell $(ROOT_DIR)/script/mainnet-from-tag.sh "$(RELEASE_TAG)" && echo "true" || echo "false")
endif
IS_STABLE             ?= false

GO_LINKMODE            ?= external
CGO_ENABLED            ?= $(shell go env CGO_ENABLED)
ifeq ($(CGO_ENABLED),0)
	ifeq ($(GO_LINKMODE),external)
		GO_LINKMODE := internal
	endif
endif
GOMOD                  ?= readonly
BUILD_TAGS             ?= osusergo,netgo,hidraw,ledger
GORELEASER_STRIP_FLAGS ?=

ifeq ($(OS),Windows_NT)
GO_LINKMODE            := internal
endif

ifeq ($(IS_MAINNET), true)
	ifeq ($(IS_PREREL), false)
		IS_STABLE                  := true
	endif
endif

ifneq (,$(findstring cgotrace,$(BUILD_OPTIONS)))
	BUILD_TAGS := $(BUILD_TAGS),cgotrace
endif

GORELEASER_BUILD_VARS := \
-X github.com/cosmos/cosmos-sdk/version.Name=virtengine \
-X github.com/cosmos/cosmos-sdk/version.AppName=virtengine \
-X github.com/cosmos/cosmos-sdk/version.BuildTags=\"$(BUILD_TAGS)\" \
-X github.com/cosmos/cosmos-sdk/version.Version=$(RELEASE_TAG) \
-X github.com/cosmos/cosmos-sdk/version.Commit=$(GIT_HEAD_COMMIT_LONG)

ifeq ($(OS),Windows_NT)
GIT_VERSION_RAW := $(shell git describe --tags 2>NUL || echo 0.0.0)
GIT_VERSION := $(patsubst v%,%,$(GIT_VERSION_RAW))
else
GIT_VERSION := $(shell git describe --tags 2>/dev/null | sed 's/^v//' || echo "0.0.0")
endif

ldflags = -linkmode=$(GO_LINKMODE) -X github.com/cosmos/cosmos-sdk/version.Name=virtengine \
-X github.com/cosmos/cosmos-sdk/version.AppName=virtengine \
-X github.com/cosmos/cosmos-sdk/version.BuildTags="$(BUILD_TAGS)" \
-X github.com/cosmos/cosmos-sdk/version.Version=$(GIT_VERSION) \
-X github.com/cosmos/cosmos-sdk/version.Commit=$(GIT_HEAD_COMMIT_LONG)

# check for nostrip option
ifeq (,$(findstring nostrip,$(BUILD_OPTIONS)))
	ldflags                += -s -w
	GORELEASER_STRIP_FLAGS += -s -w
endif

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -mod=$(GOMOD) -tags='$(BUILD_TAGS)' -ldflags '$(ldflags)'

.PHONY: all
all: build bins

.PHONY: clean
clean: cache-clean
	rm -f $(BINS)

# Report the linkmode this Makefile actually builds with, so a CI job can be
# checked against it instead of against a hardcoded expectation. `scripts/ci/
# check-build-linkmode.mjs` compares every `.github/actions/*` that exports
# GO_LINKMODE against this value under the SAME environment; the two disagreeing
# is how `Build (macOS)` stayed red on every run since #967 (2026-09-27), because
# setup-macos forced `internal` on a CGO_ENABLED=1 runner where the `hidraw`
# build tag pulls a cgo-only package that cannot be internally linked.
#
# `internal` here is Make's own variable expansion inside a recipe, so it reports
# the post-conditional value (the CGO_ENABLED=0 downgrade, the Windows_NT
# override, and any `?=` that yields to the environment) rather than a
# restatement of the default.
.PHONY: print-linkmode
print-linkmode:
	@echo "$(GO_LINKMODE)"

# Report the linkmode a LINUX OR MACOS runner would derive, which is what every
# composite CI action actually builds with.
#
# `make print-linkmode` alone cannot answer that from a Windows host: line 53
# above is `GO_LINKMODE := internal` under `ifeq ($(OS),Windows_NT)`, an
# unconditional `:=` assignment applied AFTER the CGO_ENABLED rule, so it
# overrides the environment no matter what is exported. On Windows the variable is
# therefore always `internal`, every candidate override equals it, and a guard
# comparing the two can never disagree with anything — it would report OK on the
# exact defect it exists to catch.
#
# `OS` cannot be cleared for the whole Makefile, because `make/init.mk:12-29`
# branches on it to decide whether to demand direnv. So this target delegates to
# make/linkmode.mk, which carries the same two rules minus the Windows override.
# That file is a verbatim copy of lines 41-47 and cannot drift silently: the
# guard asserts its answer agrees with the real Makefile wherever the real
# Makefile can be interrogated (see check-build-linkmode.mjs `selftestAgreement`).
print-linkmode-for-runner:
	@$(MAKE) --no-print-directory -f make/linkmode.mk print-linkmode

.PHONY: proto-generate
proto-generate:
	./scripts/proto-generate.sh all

.PHONY: proto-verify
proto-verify:
	./scripts/verify-proto-generation.sh

.PHONY: modules-verify
modules-verify:
	./scripts/verify-modules.sh

include make/releasing.mk
include make/mod.mk
include make/lint.mk
include make/test-integration.mk
include make/test-simulation.mk
include make/tools.mk
include make/codegen.mk
include make/hooks.mk
include make/supply-chain.mk
