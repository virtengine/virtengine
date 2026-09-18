ifeq ($(OS),Windows_NT)
UNAME_OS              := Windows
UNAME_ARCH            := $(PROCESSOR_ARCHITECTURE)
else
UNAME_OS              := $(shell uname -s)
UNAME_ARCH            := $(shell uname -m)
endif

# certain targets need to use bash
# detect where bash is installed
# use virtengine-node-ready target as example
ifeq ($(OS),Windows_NT)
BASH_PATH := $(shell where bash 2>nul)
else
BASH_PATH := $(shell which bash)
endif

# On Windows, skip direnv validation and fall back to local defaults.
ifeq ($(OS),Windows_NT)
VE_DIRENV_SET := 1
endif

# On non-Windows or when VE_DIRENV_SET is already exported, require direnv.
ifneq (1, $(VE_DIRENV_SET))
  ifeq (, $(shell which direnv))
    $(error "No direnv in $(PATH) and VE_DIRENV_SET not set. Install direnv (https://direnv.net) or export VE_DIRENV_SET=1 with required env vars.")
  endif
  $(error "no envrc detected. might need to run \"direnv allow\"")
endif

# VE_ROOT may not be set if environment does not support/use direnv.
ifndef VE_ROOT
ifeq ($(OS),Windows_NT)
VE_ROOT := $(CURDIR)
else
$(error "VE_ROOT is not set. Export VE_ROOT or run \"direnv allow\"")
endif
endif

ifndef ROOT_DIR
ROOT_DIR := $(VE_ROOT)
endif

# SEMVER is exported by .envrc:72 for direnv shells, but nothing else defines it:
# a shell that only carries the devcache env (CI's non-direnv fallback in
# .github/actions/setup-ubuntu, or a hand-exported VE_DEVCACHE as
# _docs/development-environment.md instructs) leaves it empty. make/setup-cache.mk
# then derives an empty major for mockery/golangci-lint and hands `go install` a
# malformed module path (github.com/vektra/mockery/v@v3.5.0) that surfaces as a Go
# module error, not as a missing-env diagnostic. Same fallback shape as VE_ROOT above.
ifndef SEMVER
SEMVER := $(ROOT_DIR)/script/semver.sh
endif

ifeq (, $(GOTOOLCHAIN))
ifeq ($(OS),Windows_NT)
GOTOOLCHAIN := auto
else
$(error "GOTOOLCHAIN is not set")
endif
endif

NULL  :=
SPACE := $(NULL) #
COMMA := ,

BINS := $(VIRTENGINE)

ifeq ($(GO111MODULE),off)
else
	GOMOD=readonly
endif

ifneq ($(GOWORK),off)
#	ifeq ($(shell test -e $(VE_ROOT)/go.work && echo -n yes),yes)
#		GOWORK=${VE_ROOT}/go.work
#	else
#		GOWORK=off
#	endif

	ifeq ($(GOMOD),$(filter $(GOMOD),mod ""))
$(error '-mod may only be set to readonly or vendor when in workspace mode, but it is set to ""')
	endif
endif

ifeq ($(GOMOD),vendor)
	ifneq ($(wildcard ./vendor/.),)
$(error "go -mod is in vendor mode but vendor dir has not been found. consider to run go mod vendor")
	endif
endif

GO                           := GO111MODULE=$(GO111MODULE) go
GO_BUILD                     := $(GO) build -mod=$(GOMOD)
GO_TEST                      := $(GO) test -mod=$(GOMOD)
GO_VET                       := $(GO) vet -mod=$(GOMOD)
ifeq ($(OS),Windows_NT)
GO_MOD_NAME                  := $(shell go list -m 2>nul)
else
GO_MOD_NAME                  := $(shell go list -m 2>/dev/null)
endif

ifeq ($(OS),Windows_NT)
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

# ==== Build tools versions ====
# Format <TOOL>_VERSION
GOLANGCI_LINT_VERSION        ?= v2.13.2
STATIK_VERSION               ?= v0.1.7
GIT_CHGLOG_VERSION           ?= v0.15.1
MOCKERY_VERSION              ?= 3.5.0
COSMOVISOR_VERSION           ?= v1.7.1
GITLEAKS_VERSION             ?= 8.22.1

# ==== Build tools version tracking ====
# <TOOL>_VERSION_FILE points to the marker file for the installed version.
# If <TOOL>_VERSION_FILE is changed, the binary will be re-downloaded.
GIT_CHGLOG_VERSION_FILE          := $(VE_DEVCACHE_VERSIONS)/git-chglog/$(GIT_CHGLOG_VERSION)
MOCKERY_VERSION_FILE             := $(VE_DEVCACHE_VERSIONS)/mockery/v$(MOCKERY_VERSION)
GOLANGCI_LINT_VERSION_FILE       := $(VE_DEVCACHE_VERSIONS)/golangci-lint/$(GOLANGCI_LINT_VERSION)
STATIK_VERSION_FILE              := $(VE_DEVCACHE_VERSIONS)/statik/$(STATIK_VERSION)
COSMOVISOR_VERSION_FILE          := $(VE_DEVCACHE_VERSIONS)/cosmovisor/$(COSMOVISOR_VERSION)
COSMOVISOR_DEBUG_VERSION_FILE    := $(VE_DEVCACHE_VERSIONS)/cosmovisor/debug/$(COSMOVISOR_VERSION)
GITLEAKS_VERSION_FILE            := $(VE_DEVCACHE_VERSIONS)/gitleaks/$(GITLEAKS_VERSION)

# ==== Build tools executables ====
GIT_CHGLOG                       := $(VE_DEVCACHE_BIN)/git-chglog
MOCKERY                          := $(VE_DEVCACHE_BIN)/mockery
NPM                              := npm
GOLANGCI_LINT                    := $(VE_DEVCACHE_BIN)/golangci-lint
STATIK                           := $(VE_DEVCACHE_BIN)/statik
COSMOVISOR                       := $(VE_DEVCACHE_BIN)/cosmovisor
COSMOVISOR_DEBUG                 := $(VE_RUN_BIN)/cosmovisor
GITLEAKS                         := $(VE_DEVCACHE_BIN)/gitleaks


# ==== Release tag ====
# `git describe --tags --abbrev=0` is NOT usable as a release-tag default here: 31 of the
# 32 remote tags are checkpoint/* automation tags and only v0.1.0 is a release tag, so
# git describe resolves to `checkpoint/stable-virtengine-beta/consolidated-v2`. Every
# consumer (GORELEASER_BUILD_VARS, IS_PREREL, IS_MAINNET, script/genchangelog.sh) then
# received a non-semver tag, which made `make gen-changelog` fail with exit 1 and a
# 0-byte notes file (release plan defect D4).
#
# Default to the newest semver (v*) tag; a caller-supplied RELEASE_TAG always wins.
# `make release-tag-check` (make/releasing.mk) rejects a non-semver value loudly.
ifeq ($(OS),Windows_NT)
RELEASE_TAG           ?= $(shell git tag --list "v[0-9]*" --sort=-v:refname 2>NUL | head -n 1)
else
RELEASE_TAG           ?= $(shell git tag --list "v[0-9]*" --sort=-v:refname 2>/dev/null | head -n 1)
endif
RELEASE_TAG           := $(strip $(RELEASE_TAG))
ifeq ($(RELEASE_TAG),)
RELEASE_TAG           := v0.0.0
endif

include $(VE_ROOT)/make/setup-cache.mk
