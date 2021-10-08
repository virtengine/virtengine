KEY_NAME   ?= main


KEY_ADDRESS          ?= $(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(KEY_NAME)" -a)
PROVIDER_KEY_NAME    ?= provider
PROVIDER_ADDRESS     ?= $(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(PROVIDER_KEY_NAME)" -a)
PROVIDER_CONFIG_PATH ?= provider.yaml

SDL_PATH ?= deployment.yaml

DSEQ           ?= 1
GSEQ           ?= 1
OSEQ           ?= 1
PRICE          ?= 10uve
CERT_HOSTNAME  ?= localhost
LEASE_SERVICES ?= web

.PHONY: multisig-send
multisig-send:
	$(VIRTENGINE) tx send \
		"$(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(MULTISIG_KEY)" -a)" \
		"$(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(KEY_NAME)"     -a)" \
		1000000uve \
		--generate-only \
		> "$(VIRTENGINE_HOME)/multisig-tx.json"
	$(VIRTENGINE) tx sign \
		"$(VIRTENGINE_HOME)/multisig-tx.json" \
		--multisig "$(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(MULTISIG_KEY)" -a)" \
		--from "main" \
		> "$(VIRTENGINE_HOME)/multisig-sig-main.json"
	$(VIRTENGINE) tx sign \
		"$(VIRTENGINE_HOME)/multisig-tx.json" \
		--multisig "$(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(MULTISIG_KEY)" -a)" \
		--from "other" \
		> "$(VIRTENGINE_HOME)/multisig-sig-other.json"
	$(VIRTENGINE) tx multisign \
		"$(VIRTENGINE_HOME)/multisig-tx.json" \
		"$(MULTISIG_KEY)" \
		"$(VIRTENGINE_HOME)/multisig-sig-main.json" \
		"$(VIRTENGINE_HOME)/multisig-sig-other.json" \
		> "$(VIRTENGINE_HOME)/multisig-final.json"
	$(VIRTENGINE) "$(CHAIN_OPTS)" tx broadcast "$(VIRTENGINE_HOME)/multisig-final.json"

.PHONY: provider-create
provider-create:
	$(VIRTENGINE) tx provider create "$(PROVIDER_CONFIG_PATH)" \
		--from "$(PROVIDER_KEY_NAME)"

.PHONY: provider-update
provider-update:
	$(VIRTENGINE) tx provider update "$(PROVIDER_CONFIG_PATH)" \
		--from "$(PROVIDER_KEY_NAME)"

.PHONY: provider-status
provider-status:
	$(VIRTENGINE) provider status $(PROVIDER_ADDRESS)

.PHONY: send-manifest
send-manifest:
	$(VIRTENGINE) provider send-manifest "$(SDL_PATH)" \
		--dseq "$(DSEQ)"     \
		--from "$(KEY_NAME)" \
		--provider "$(PROVIDER_ADDRESS)"

.PHONY: deployment-create
deployment-create:
	$(VIRTENGINE) tx deployment create "$(SDL_PATH)" \
		--dseq "$(DSEQ)" \
		--from "$(KEY_NAME)"

.PHONY: deploy-create
deploy-create:
	$(VIRTENGINE) deploy create "$(SDL_PATH)" \
		--dseq "$(DSEQ)" \
		--from "$(KEY_NAME)"

.PHONY: deployment-deposit
deployment-deposit:
	$(VIRTENGINE) tx deployment deposit "$(PRICE)" \
		--dseq "$(DSEQ)" \
		--from "$(KEY_NAME)"

.PHONY: deployment-update
deployment-update:
	$(VIRTENGINE) tx deployment update "$(SDL_PATH)" \
		--dseq "$(DSEQ)" \
		--from "$(KEY_NAME)"

.PHONY: deployment-close
deployment-close:
	$(VIRTENGINE) tx deployment close \
		--owner "$(MAIN_ADDR)" \
		--dseq "$(DSEQ)"       \
		--from "$(KEY_NAME)"

.PHONY: group-close
group-close:
	$(VIRTENGINE) tx deployment group close \
		--owner "$(KEY_ADDRESS)"       \
		--dseq  "$(DSEQ)"              \
		--gseq  "$(GSEQ)"              \
		--from  "$(KEY_NAME)"

.PHONY: group-pause
group-pause:
	$(VIRTENGINE) tx deployment group pause \
		--owner "$(KEY_ADDRESS)"       \
		--dseq  "$(DSEQ)"              \
		--gseq  "$(GSEQ)"              \
		--from  "$(KEY_NAME)"

.PHONY: group-start
group-start:
	$(VIRTENGINE) tx deployment group start \
		--owner "$(KEY_ADDRESS)"       \
		--dseq  "$(DSEQ)"              \
		--gseq  "$(GSEQ)"              \
		--from  "$(KEY_NAME)"

.PHONY: bid-create
bid-create:
	$(VIRTENGINE) tx market bid create \
		--owner "$(KEY_ADDRESS)"       \
		--dseq  "$(DSEQ)"              \
		--gseq  "$(GSEQ)"              \
		--oseq  "$(OSEQ)"              \
		--from  "$(PROVIDER_KEY_NAME)" \
		--price "$(PRICE)"

.PHONY: bid-close
bid-close:
	$(VIRTENGINE) tx market bid close \
		--owner "$(KEY_ADDRESS)"       \
		--dseq  "$(DSEQ)"              \
		--gseq  "$(GSEQ)"              \
		--oseq  "$(OSEQ)"              \
		--from  "$(PROVIDER_KEY_NAME)"

.PHONY: lease-create
lease-create:
	$(VIRTENGINE) tx market lease create \
		--owner "$(KEY_ADDRESS)"         \
		--dseq  "$(DSEQ)"                \
		--gseq  "$(GSEQ)"                \
		--oseq  "$(OSEQ)"                \
		--provider "$(PROVIDER_ADDRESS)" \
		--from  "$(KEY_NAME)"

.PHONY: lease-withdraw
lease-withdraw:
	$(VIRTENGINE) tx market lease withdraw \
		--owner "$(KEY_ADDRESS)"         \
		--dseq  "$(DSEQ)"                \
		--gseq  "$(GSEQ)"                \
		--oseq  "$(OSEQ)"                \
		--provider "$(PROVIDER_ADDRESS)" \
		--from  "$(PROVIDER_KEY_NAME)"

.PHONY: lease-close
lease-close:
	$(VIRTENGINE) tx market lease close \
		--owner "$(KEY_ADDRESS)"         \
		--dseq  "$(DSEQ)"                \
		--gseq  "$(GSEQ)"                \
		--oseq  "$(OSEQ)"                \
		--provider "$(PROVIDER_ADDRESS)" \
		--from  "$(KEY_NAME)"

.PHONY: query-accounts
query-accounts: $(patsubst %, query-account-%,$(GENESIS_ACCOUNTS))

.PHONY: query-account-%
query-account-%:
	$(VIRTENGINE) query bank balances "$(shell $(VIRTENGINE) $(KEY_OPTS) keys show -a "$(@:query-account-%=%)")"
	$(VIRTENGINE) query account       "$(shell $(VIRTENGINE) $(KEY_OPTS) keys show -a "$(@:query-account-%=%)")"

.PHONY: query-provider
query-provider:
	$(VIRTENGINE) query provider get "$(PROVIDER_ADDRESS)"

.PHONY: query-providers
query-providers:
	$(VIRTENGINE) query provider list

.PHONY: query-deployment
query-deployment:
	$(VIRTENGINE) query deployment get \
		--owner "$(KEY_ADDRESS)" \
		--dseq  "$(DSEQ)"

.PHONY: query-deployments
query-deployments:
	$(VIRTENGINE) query deployment list

.PHONY: query-order
query-order:
	$(VIRTENGINE) query market order get \
		--owner "$(KEY_ADDRESS)" \
		--dseq  "$(DSEQ)"        \
		--gseq  "$(GSEQ)"        \
		--oseq  "$(OSEQ)"

.PHONY: query-orders
query-orders:
	$(VIRTENGINE) query market order list

.PHONY: query-bid
query-bid:
	$(VIRTENGINE) query market bid get \
		--owner     "$(KEY_ADDRESS)" \
		--dseq      "$(DSEQ)"        \
		--gseq      "$(GSEQ)"        \
		--oseq      "$(OSEQ)"        \
		--provider  "$(PROVIDER_ADDRESS)"

.PHONY: query-bids
query-bids:
	$(VIRTENGINE) query market bid list

.PHONY: query-lease
query-lease:
	$(VIRTENGINE) query market lease get \
		--owner     "$(KEY_ADDRESS)" \
		--dseq      "$(DSEQ)"        \
		--gseq      "$(GSEQ)"        \
		--oseq      "$(OSEQ)"        \
		--provider  "$(PROVIDER_ADDRESS)"

.PHONY: query-leases
query-leases:
	$(VIRTENGINE) query market lease list

.PHONY: query-certificates
query-certificates:
	$(VIRTENGINE) query cert list

.PHONY: query-account-certificates
query-account-certificates:
	$(VIRTENGINE) query cert list --owner="$(KEY_ADDRESS)" --state="valid"

.PHONY: create-server-certificate
create-server-certificate:
	$(VIRTENGINE) tx cert create server $(CERT_HOSTNAME) --from=$(KEY_NAME) --rie

.PHONY: revoke-certificate
revoke-certificate:
	$(VIRTENGINE) tx cert revoke --from=$(KEY_NAME)

.PHONY: events-run
events-run:
	$(VIRTENGINE) events

.PHONY: provider-lease-logs
provider-lease-logs:
	$(VIRTENGINE) provider lease-logs \
		-f \
		--service="$(LEASE_SERVICES)" \
		--dseq "$(DSEQ)"     \
		--from "$(KEY_NAME)" \
		--provider "$(PROVIDER_ADDRESS)"

.PHONY: provider-lease-events
provider-lease-events:
	$(VIRTENGINE) provider lease-events \
		-f \
		--dseq "$(DSEQ)"     \
		--from "$(KEY_NAME)" \
		--provider "$(PROVIDER_ADDRESS)"
