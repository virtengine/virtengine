GORELEASER_SKIP_VALIDATE ?= false

GON_CONFIGFILE ?= gon.json

ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
else
    DETECTED_OS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

.PHONY: bins
bins: $(BINS)

.PHONY: build
build:
	$(GO) build -a  ./...

$(VIRTENGINE): modvendor
	$(GO) build -o $@ $(BUILD_FLAGS) ./cmd/virtengine

.PHONY: virtengine
virtengine: $(VIRTENGINE)

.PHONY: virtengine_docgen
virtengine_docgen: $(VIRTENGINE_DEVCACHE)
	$(GO) build -o $(VIRTENGINE_DEVCACHE_BIN)/virtengine_docgen $(BUILD_FLAGS) ./docgen

.PHONY: install
install:
	$(GO) install $(BUILD_FLAGS) ./cmd/virtengine

.PHONY: image-minikube
image-minikube:
	eval $$(minikube docker-env) && docker-image

.PHONY: docker-image
docker-image:
	docker run \
		--rm \
		--privileged \
		-e MAINNET=$(MAINNET) \
		-e BUILD_FLAGS="$(GORELEASER_FLAGS)" \
		-e LD_FLAGS="$(GORELEASER_LD_FLAGS)" \
		-e GOLANG_VERSION="$(GOLANG_VERSION)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/virtengine/virtengine \
		-w /go/src/github.com/virtengine/virtengine \
		troian/golang-cross:${GOLANG_CROSS_VERSION}-linux-amd64 \
		-f .goreleaser-docker.yaml --rm-dist --skip-validate --skip-publish --snapshot

.PHONY: gen-changelog
gen-changelog: $(GIT_CHGLOG)
	@echo "generating changelog to .cache/changelog"
	./script/genchangelog.sh "$(GORELEASER_TAG)" .cache/changelog.md

.PHONY: release-dry-run
release-dry-run: modvendor gen-changelog
	docker run \
		--rm \
		--privileged \
		-e MAINNET=$(MAINNET) \
		-e BUILD_FLAGS="$(GORELEASER_FLAGS)" \
		-e LD_FLAGS="$(GORELEASER_LD_FLAGS)" \
		-e HOMEBREW_NAME="$(GORELEASER_HOMEBREW_NAME)" \
		-e HOMEBREW_CUSTOM="$(GORELEASER_HOMEBREW_CUSTOM)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/virtengine/virtengine \
		-w /go/src/github.com/virtengine/virtengine \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		-f "$(GORELEASER_CONFIG)" --skip-validate=$(GORELEASER_SKIP_VALIDATE) --rm-dist --skip-publish --release-notes=/go/src/github.com/virtengine/virtengine/.cache/changelog.md

.PHONY: release
release: modvendor gen-changelog
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e MAINNET=$(MAINNET) \
		-e BUILD_FLAGS="$(GORELEASER_FLAGS)" \
		-e LD_FLAGS="$(GORELEASER_LD_FLAGS)" \
		-e HOMEBREW_NAME="$(GORELEASER_HOMEBREW_NAME)" \
		-e HOMEBREW_CUSTOM="$(GORELEASER_HOMEBREW_CUSTOM)" \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/virtengine/virtengine \
		-w /go/src/github.com/virtengine/virtengine \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		-f "$(GORELEASER_CONFIG)" release --rm-dist --release-notes=/go/src/github.com/virtengine/virtengine/.cache/changelog.md
