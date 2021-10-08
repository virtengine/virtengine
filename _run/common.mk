include ../common-base.mk

# https://stackoverflow.com/a/7531247
# https://www.gnu.org/software/make/manual/make.html#Flavors
null  := 
space := $(null) #
comma := ,

ifndef VIRTENGINE_HOME
$(error VIRTENGINE_HOME is not set)
endif

export VIRTENGINE_KEYRING_BACKEND = test
export VIRTENGINE_GAS_ADJUSTMENT  = 2
export VIRTENGINE_CHAIN_ID        = local
export VIRTENGINE_YES             = true

VIRTENGINE        := $(VIRTENGINE) --home $(VIRTENGINE_HOME)

KEY_OPTS     := --keyring-backend=$(VIRTENGINE_KEYRING_BACKEND)
GENESIS_PATH := $(VIRTENGINE_HOME)/config/genesis.json

CHAIN_MIN_DEPOSIT     := 10000000000000
CHAIN_ACCOUNT_DEPOSIT := $(shell echo $$(($(CHAIN_MIN_DEPOSIT) * 10)))
CHAIN_TOKEN_DENOM     := uve

KEY_NAMES := main provider validator other

MULTISIG_KEY     := msig
MULTISIG_SIGNERS := main other

GENESIS_ACCOUNTS := $(KEY_NAMES) $(MULTISIG_KEY)

CLIENT_CERTS := main validator other
SERVER_CERTS := provider


.PHONY: init
init: bins client-init node-init

.PHONY: client-init
client-init: init-dirs client-init-keys

.PHONY: init-dirs
init-dirs: 
	mkdir -p "$(VIRTENGINE_HOME)"

.PHONY: client-init-keys
client-init-keys: $(patsubst %,client-init-key-%,$(KEY_NAMES)) client-init-multisig-key

.PHONY: client-init-key-%
client-init-key-%:
	$(VIRTENGINE) keys add "$(@:client-init-key-%=%)"

.PHONY: client-init-multisig-key
client-init-multisig-key:
	$(VIRTENGINE) keys add \
		"$(MULTISIG_KEY)" \
		--multisig "$(subst $(space),$(comma),$(strip $(MULTISIG_SIGNERS)))" \
		--multisig-threshold 2

.PHONY: node-init
node-init: node-init-genesis node-init-genesis-accounts node-init-genesis-certs node-init-gentx node-init-finalize

.PHONY: node-init-genesis
node-init-genesis: init-dirs
	$(VIRTENGINE) init node0
	cp "$(GENESIS_PATH)" "$(GENESIS_PATH).orig"
	cat "$(GENESIS_PATH).orig" | \
		jq -rM '(..|objects|select(has("denom"))).denom           |= "$(CHAIN_TOKEN_DENOM)"' | \
		jq -rM '(..|objects|select(has("bond_denom"))).bond_denom |= "$(CHAIN_TOKEN_DENOM)"' | \
		jq -rM '(..|objects|select(has("mint_denom"))).mint_denom |= "$(CHAIN_TOKEN_DENOM)"' > \
		"$(GENESIS_PATH)"

.PHONY: node-init-genesis-certs
node-init-genesis-certs: $(patsubst %,node-init-genesis-client-cert-%,$(CLIENT_CERTS)) $(patsubst %,node-init-genesis-server-cert-%,$(SERVER_CERTS))

.PHONY: node-init-genesis-client-cert-%
node-init-genesis-client-cert-%:
	$(VIRTENGINE) tx cert create client --to-genesis=true --from=$(@:node-init-genesis-client-cert-%=%)

.PHONY: node-init-genesis-server-cert-%
node-init-genesis-server-cert-%:
	$(VIRTENGINE) tx cert create server localhost virtengine-provider.localhost --to-genesis=true --from=$(@:node-init-genesis-server-cert-%=%)

.PHONY: node-init-genesis-accounts
node-init-genesis-accounts: $(patsubst %,node-init-genesis-account-%,$(GENESIS_ACCOUNTS))
	$(VIRTENGINE) validate-genesis

.PHONY: node-init-genesis-account-%
node-init-genesis-account-%:
	$(VIRTENGINE) add-genesis-account \
		"$(shell $(VIRTENGINE) $(KEY_OPTS) keys show "$(@:node-init-genesis-account-%=%)" -a)" \
		"$(CHAIN_MIN_DEPOSIT)$(CHAIN_TOKEN_DENOM)"

.PHONY: node-init-gentx
node-init-gentx:
	$(VIRTENGINE) gentx validator \
		"$(CHAIN_MIN_DEPOSIT)$(CHAIN_TOKEN_DENOM)"

.PHONY: node-init-finalize
node-init-finalize:
	$(VIRTENGINE) collect-gentxs
	$(VIRTENGINE) validate-genesis

.PHONY: node-run
node-run:
	$(VIRTENGINE) start

.PHONY: node-status
node-status:
	$(VIRTENGINE) status

.PHONY: rest-server-run
rest-server-run:
	$(VIRTENGINE) rest-server

.PHONY: clean
clean: clean-$(VIRTENGINE_RUN_NAME)
	rm -rf "$(VIRTENGINE_HOME)"
