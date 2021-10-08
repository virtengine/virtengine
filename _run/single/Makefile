include ../common.mk
include ../common-commands.mk
include ../common-kind.mk

KUSTOMIZE_ROOT           ?= $(VIRTENGINE_ROOT)/_docs/kustomize
KUSTOMIZE_DIR            ?= $(VIRTENGINE_ROOT)/_run/single/kustomize
KUSTOMIZE_PROVIDER_DIR   ?= $(KUSTOMIZE_DIR)/virtengine-provider
KUSTOMIZE_PROVIDER_CACHE ?= $(KUSTOMIZE_PROVIDER_DIR)/cache
KUSTOMIZE_VIRTENGINED_DIR     ?= $(KUSTOMIZE_DIR)/virtengine-node
KUSTOMIZE_VIRTENGINED_CACHE   ?= $(KUSTOMIZE_VIRTENGINED_DIR)/cache
CLIENT_EXPORT_PASSWORD   ?= 12345678

PROVIDER_HOSTNAME  = virtengine-provider.localhost
export VIRTENGINE_NODE  = tcp://virtengine.localhost:$(KIND_PORT_BINDINGS)
GATEWAY_ENDPOINT  ?= https://virtengine-provider.localhost

.PHONY: kind-namespace-setup
kind-namespace-setup:
	kubectl apply -f "$(KUSTOMIZE_ROOT)/networking"

.PHONY: kustomize-init
kustomize-init: kustomize-init-virtengine-node kustomize-init-provider kind-namespace-setup

.PHONY: kustomize-init-virtengine-node
kustomize-init-virtengine-node:
	mkdir -p "$(KUSTOMIZE_VIRTENGINED_CACHE)"
	cp -r "$(VIRTENGINE_HOME)/"* "$(KUSTOMIZE_VIRTENGINED_CACHE)/"

.PHONY: kustomize-init-provider
kustomize-init-provider:
	mkdir -p "$(KUSTOMIZE_PROVIDER_CACHE)"
	cp -r "$(VIRTENGINE_HOME)/config" "$(KUSTOMIZE_PROVIDER_CACHE)/"
	echo "$(CLIENT_EXPORT_PASSWORD)" > "$(KUSTOMIZE_PROVIDER_CACHE)/key-pass.txt"
	cat "$(VIRTENGINE_HOME)/$(PROVIDER_ADDRESS).pem" > "$(KUSTOMIZE_PROVIDER_CACHE)/provider-cert.pem"
	( \
		cat "$(KUSTOMIZE_PROVIDER_CACHE)/key-pass.txt" ; \
		cat "$(KUSTOMIZE_PROVIDER_CACHE)/key-pass.txt"   \
	) | $(VIRTENGINE) keys export provider 1> "$(KUSTOMIZE_PROVIDER_CACHE)/key.txt"

.PHONY: kustomize-install-node
kustomize-install-node:
	kubectl kustomize kustomize/virtengine-node | kubectl apply -f-

.PHONY: kustomize-install-provider
kustomize-install-provider:
	kubectl kustomize kustomize/virtengine-provider | kubectl apply -f-

.PHONY: provider-lease-ping
provider-lease-ping:
	curl -sIH "Host: hello.localhost" localhost:$(KIND_HTTP_PORT)

PHONY: provider-lease-status
provider-lease-status:
	$(VIRTENGINE) provider lease-status \
		--dseq      "$(DSEQ)"        \
		--gseq      "$(GSEQ)"        \
		--oseq      "$(OSEQ)"        \
		--from      "$(KEY_NAME)" \
		--provider  "$(PROVIDER_ADDRESS)"

PHONY: clean-$(VIRTENGINE_RUN_NAME)
clean-$(VIRTENGINE_RUN_NAME):
	rm -rf $(CURDIR)/kustomize/virtengine-node/cache
	rm -f  $(CURDIR)/kustomize/virtengine-node/docker-image.yaml
	rm -rf $(CURDIR)/kustomize/virtengine-provider/cache
	rm -f  $(CURDIR)/kustomize/virtengine-provider/docker-image.yaml
