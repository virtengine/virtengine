GON_CONFIGFILE           ?= gon.json

GORELEASER_VERBOSE       ?= false
GORELEASER_IMAGE         := ghcr.io/goreleaser/goreleaser-cross:$(GOTOOLCHAIN_SEMVER)
GORELEASER_RELEASE       ?= false
GORELEASER_MOUNT_CONFIG  ?= false
GORELEASER_SKIP          := $(subst $(COMMA),$(SPACE),$(GORELEASER_SKIP))
RELEASE_DOCKER_IMAGE     ?= ghcr.io/virtengine/virtengine
#GORELEASER_MOD_MOUNT     ?= $(shell git config --get remote.origin.url | sed -r 's/.*(\@|\/\/)(.*)(\:|\/)([^:\/]*)\/([^\/\.]*)\.git/\2\/\4\/\5/' | tr -d '\n')
ifeq ($(OS),Windows_NT)
GORELEASER_MOD_MOUNT     ?= $(shell powershell -NoProfile -Command "(Get-Content -Path '$(ROOT_DIR)/.github/repo' -Raw).Trim()")
else
GORELEASER_MOD_MOUNT     ?= $(shell cat $(ROOT_DIR)/.github/repo | tr -d '\n')
endif

RELEASE_DOCKER_IMAGE     ?= ghcr.io/virtengine/virtengine

GORELEASER_GOWORK        := $(GOWORK)

ifneq ($(GOWORK), off)
	GORELEASER_GOWORK    := /go/src/$(GORELEASER_MOD_MOUNT)/go.work
endif

ifneq ($(GORELEASER_RELEASE),true)
	ifeq (,$(findstring publish,$(GORELEASER_SKIP)))
		GORELEASER_SKIP += publish
	endif

	GITHUB_TOKEN=
endif

ifneq (,$(GORELEASER_SKIP))
	GORELEASER_SKIP := --skip=$(subst $(SPACE),$(COMMA),$(strip $(GORELEASER_SKIP)))
endif

ifeq ($(GORELEASER_MOUNT_CONFIG),true)
	GORELEASER_IMAGE := -v $(HOME)/.docker/config.json:/root/.docker/config.json $(GORELEASER_IMAGE)
endif

.PHONY: bins
bins: $(BINS)

.PHONY: build
build:
	$(GO_BUILD) -a  ./...

.PHONY: $(VIRTENGINE)
$(VIRTENGINE):
	$(GO_BUILD) -o $@ $(BUILD_FLAGS) ./cmd/virtengine

.PHONY: virtengine
virtengine: $(VIRTENGINE)

.PHONY: virtengine_docgen
virtengine_docgen: $(VE_DEVCACHE)
	$(GO_BUILD) -o $(VE_DEVCACHE_BIN)/virtengine_docgen $(BUILD_FLAGS) ./docgen

.PHONY: install
install:
	@echo installing virtengine
	$(GO) install $(BUILD_FLAGS) ./cmd/virtengine

.PHONY: image-minikube
image-minikube:
	eval $$(minikube docker-env) && docker-image

.PHONY: test-bins
test-bins:
	docker run \
		--rm \
		-e STABLE=$(IS_STABLE) \
		-e MOD="$(GOMOD)" \
		-e BUILD_TAGS="$(BUILD_TAGS)" \
		-e BUILD_VARS="$(GORELEASER_BUILD_VARS)" \
		-e STRIP_FLAGS="$(GORELEASER_STRIP_FLAGS)" \
		-e LINKMODE="$(GO_LINKMODE)" \
		-e DOCKER_IMAGE=$(RELEASE_DOCKER_IMAGE) \
		-e GOPATH=/go \
		-e GOTOOLCHAIN="$(GOTOOLCHAIN)" \
		-e GOWORK="$(GORELEASER_GOWORK)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH):/go \
		-v $(VE_ROOT):/go/src/$(GORELEASER_MOD_MOUNT) \
		-w /go/src/$(GORELEASER_MOD_MOUNT) \
		$(GORELEASER_IMAGE) \
		-f .goreleaser-test-bins.yaml \
		--verbose=$(GORELEASER_VERBOSE) \
		--clean \
		--skip=publish,validate \
		--snapshot

.PHONY: docker-image
docker-image:
	docker run \
		--rm \
		-e STABLE=$(IS_STABLE) \
		-e MOD="$(GOMOD)" \
		-e BUILD_TAGS="$(BUILD_TAGS)" \
		-e BUILD_VARS="$(GORELEASER_BUILD_VARS)" \
		-e STRIP_FLAGS="$(GORELEASER_STRIP_FLAGS)" \
		-e LINKMODE="$(GO_LINKMODE)" \
		-e DOCKER_IMAGE=$(RELEASE_DOCKER_IMAGE) \
		-e GOPATH=/go \
		-e GOTOOLCHAIN="$(GOTOOLCHAIN)" \
		-e GOWORK="$(GORELEASER_GOWORK)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH):/go \
		-v $(VE_ROOT):/go/src/$(GORELEASER_MOD_MOUNT) \
		-w /go/src/$(GORELEASER_MOD_MOUNT) \
		$(GORELEASER_IMAGE) \
		-f .goreleaser-docker.yaml \
		--verbose=$(GORELEASER_VERBOSE) \
		--clean \
		--skip=publish,validate \
		--snapshot

.PHONY: gen-changelog
gen-changelog: release-tag-check $(GIT_CHGLOG)
	@echo "generating changelog for $(RELEASE_TAG) to .cache/changelog.md"
	./script/genchangelog.sh "$(RELEASE_TAG)" .cache/changelog.md

# Preview release notes for a tag that has not been created yet. Nothing is tagged and
# nothing is published: git-chglog is told about the tag with --next-tag, which is the
# only way to review notes before the tag exists (release plan defect D5).
.PHONY: gen-changelog-preview
gen-changelog-preview: release-tag-check $(GIT_CHGLOG)
	@echo "previewing changelog for $(RELEASE_TAG) (tag need not exist)"
	./script/genchangelog.sh --next-tag "$(RELEASE_TAG)" .cache/changelog.md

# ---------------------------------------------------------------------------
# RELEASE_TAG guard (release plan defect D4).
#
# The repository carries ~31 checkpoint/* automation tags against a single release tag
# (v0.1.0). A non-semver RELEASE_TAG used to flow silently into the release path and
# produce a 0-byte notes file with exit 1 from git-chglog. Fail loudly and early instead.
# ---------------------------------------------------------------------------
RELEASE_TAG_SEMVER_RE := ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$$

.PHONY: release-tag-check
release-tag-check:
	@printf '%s' "$(RELEASE_TAG)" | grep -Eq '$(RELEASE_TAG_SEMVER_RE)' || { \
		echo "ERROR: RELEASE_TAG '$(RELEASE_TAG)' is not a semver release tag."; \
		echo "       Expected v<major>.<minor>.<patch>[-prerelease][+build]."; \
		echo "       checkpoint/* tags are automation checkpoints, not releases."; \
		echo "       Pass the intended tag explicitly, e.g. make gen-changelog RELEASE_TAG=v0.3.0"; \
		exit 1; \
	}

.PHONY: release
release: release-tag-check gen-changelog
	docker run \
		--rm \
		-e STABLE=$(IS_STABLE) \
		-e MOD="$(GOMOD)" \
		-e BUILD_TAGS="$(BUILD_TAGS)" \
		-e BUILD_VARS="$(GORELEASER_BUILD_VARS)" \
		-e STRIP_FLAGS="$(GORELEASER_STRIP_FLAGS)" \
		-e LINKMODE="$(GO_LINKMODE)" \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-e GORELEASER_CURRENT_TAG="$(RELEASE_TAG)" \
		-e DOCKER_IMAGE=$(RELEASE_DOCKER_IMAGE) \
		-e GOTOOLCHAIN="$(GOTOOLCHAIN)" \
		-e GOWORK="$(GORELEASER_GOWORK)" \
		-e GOPATH=/go \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH):/go \
		-v $(VE_ROOT):/go/src/$(GORELEASER_MOD_MOUNT) \
		-w /go/src/$(GORELEASER_MOD_MOUNT) \
		$(GORELEASER_IMAGE) \
		-f "$(GORELEASER_CONFIG)" \
		release \
		$(GORELEASER_SKIP) \
		--verbose=$(GORELEASER_VERBOSE) \
		--clean \
		--release-notes=/go/src/$(GORELEASER_MOD_MOUNT)/.cache/changelog.md
