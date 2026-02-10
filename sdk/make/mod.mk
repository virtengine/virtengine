# Golang modules and vendoring

define VENDOR_BUF
version: v1
lint:
  use:
    - DEFAULT
  ignore:
    - ibc-go
    - cosmos
    - tendermint
    - gogoproto
    - cosmos_proto
    - google
    - confio
    - k8s.io
endef

.PHONY: buf-deps
buf-deps:
	$(BUF) dep update

.PHONY: mod-tidy
mod-tidy:
	@$(TOOLS) go-mod-tidy "$(GO_MODULES)"

.PHONY: deps-tidy
deps-tidy:
	(cd $(GO_ROOT); $(GO) mod tidy)

.PHONY: deps-vendor
deps-vendor:
	(cd $(GO_ROOT); GOWORK=off go mod vendor)

.PHONY: modsensure
modsensure: deps-tidy deps-vendor

.PHONY: modvendor
modvendor: export VENDOR_BUF:=$(VENDOR_BUF)
modvendor: $(MODVENDOR) modsensure
	@echo "vendoring non-go files..."
	@(cd $(GO_ROOT); \
		$(MODVENDOR) -copy="**/*.proto" -include=k8s.io/apimachinery; \
		$(MODVENDOR) -copy="**/*.proto" -include=github.com/cosmos/cosmos-sdk/proto; \
		$(MODVENDOR) -copy="**/*.proto" -include=github.com/cosmos/ibc-go/v10/proto; \
		$(MODVENDOR) -copy="**/swagger.yaml" -include=github.com/cosmos/cosmos-sdk/client/docs/swagger-ui \
	)
	@(cd .cache/include; \
		mkdir -p k8s; \
		ln -snf ../../../$(GO_ROOT)/vendor/k8s.io k8s/k8s.io; \
	)
#	echo "$${VENDOR_BUF}" > $(GO_ROOT)/vendor/k8s.io/buf.yaml
