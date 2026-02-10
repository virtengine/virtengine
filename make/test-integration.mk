COVER_PACKAGES = $(shell go list ./... | grep -v mock | paste -sd, -)

TEST_MODULES ?= $(shell $(GO) list ./... | grep -v '/mocks')

###############################################################################
###                           Misc tests                                    ###
###############################################################################

.PHONY: test
test:
	$(GO_TEST) -v -timeout 600s $(TEST_MODULES)

.PHONY: test-nocache
test-nocache:
	$(GO_TEST) -count=1 $(TEST_MODULES)

.PHONY: test-full
test-full:
	$(GO_TEST) -v -tags=$(BUILD_TAGS) $(TEST_MODULES)

.PHONY: test-integration
test-integration:
	$(GO_TEST) -v -tags="e2e.integration" $(TEST_MODULES)

.PHONY: test-coverage
test-coverage:
	CGO_ENABLED=1 $(GO_TEST) -tags=$(BUILD_TAGS) -coverprofile=coverage.txt \
		-covermode=atomic \
		-timeout=20m \
		$(TEST_MODULES)

.PHONY: test-vet
test-vet:
	$(GO_VET) ./...

###############################################################################
###                     Compatibility tests                                 ###
###############################################################################

.PHONY: test-compatibility
test-compatibility:
	@echo "Running compatibility tests..."
	$(GO_TEST) -v -tags="e2e.compatibility" ./tests/compatibility/...
	$(GO_TEST) -v ./pkg/compatibility/...

.PHONY: test-compatibility-full
test-compatibility-full:
	@echo "Running full compatibility test suite..."
	$(GO_TEST) -v -tags="e2e.compatibility" -coverprofile=coverage-compatibility.txt ./tests/compatibility/...
	$(GO_TEST) -v -coverprofile=coverage-pkg-compatibility.txt ./pkg/compatibility/...
