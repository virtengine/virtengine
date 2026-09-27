AP_RUN_DIR               := $(VE_RUN)/upgrade

VE_INIT               := $(AP_RUN_DIR)/.virtengine-init

export VE_HOME
export VE_KEYRING_BACKEND    = test
export VE_GAS_ADJUSTMENT     = 2
export VE_CHAIN_ID           = localvirtengine
export VE_YES                = true
export VE_GAS_PRICES         = 0.025uve
export VE_GAS                = auto
export VE_STATESYNC_ENABLE   = false
export VE_LOG_COLOR          = true

STATE_CONFIG            ?= $(ROOT_DIR)/tests/upgrade/testnet.json
TEST_CONFIG             ?= test-config.json
KEY_OPTS                := --keyring-backend=$(VE_KEYRING_BACKEND)
KEY_NAME                ?= validator
UPGRADE_TO              ?= $(shell $(ROOT_DIR)/script/upgrades.sh upgrade-from-release $(RELEASE_TAG))
UPGRADE_FROM            := $(shell cat $(ROOT_DIR)/meta.json | jq -r --arg name $(UPGRADE_TO) '.upgrades[$$name].from_version' | tr -d '\n')
GENESIS_BINARY_VERSION  := $(shell cat $(ROOT_DIR)/meta.json | jq -r --arg name $(UPGRADE_TO) '.upgrades[$$name].from_binary' | tr -d '\n')
UPGRADE_BINARY_VERSION  ?= local

# Valid values are mainnet, sandbox and sandbox1 — the same set emitted by
# `script/upgrades.sh snapshot-source`. The previous default, "sandbox-2", matched
# no branch and made the $(error) below fire on every invocation of this makefile.
SNAPSHOT_SOURCE         ?= sandbox

# CHAIN_METADATA_URL points at the per-network chain metadata document used to
# locate the genesis download for the upgrade test. The previous default,
# https://raw.githubusercontent.com/virtengine-network/net/master/<net>/meta.json,
# is retired: the virtengine-network org no longer resolves (404 on
# raw.githubusercontent.com, `gh repo view` cannot resolve the repository), and no
# equivalent path exists in this fork. The value is therefore unset by default
# and must be injected by the caller, e.g.
#   make test CHAIN_METADATA_URL=https://<host>/<net>/meta.json
# An empty value leaves the genesis download skipped rather than fetching a URL
# that cannot resolve.
SNAPSHOT_NETWORK_MAINNET  := virtenginenet-2
SNAPSHOT_NETWORK_SANDBOX  := sandbox-2
SNAPSHOT_NETWORK_SANDBOX1 := sandbox-01

ifeq ($(SNAPSHOT_SOURCE),mainnet)
	SNAPSHOT_NETWORK    := $(SNAPSHOT_NETWORK_MAINNET)
else ifeq ($(SNAPSHOT_SOURCE),sandbox)
	SNAPSHOT_NETWORK    := $(SNAPSHOT_NETWORK_SANDBOX)
else ifeq ($(SNAPSHOT_SOURCE),sandbox1)
	SNAPSHOT_NETWORK    := $(SNAPSHOT_NETWORK_SANDBOX1)
else
$(error "invalid snapshot source $(SNAPSHOT_SOURCE)")
endif

CHAIN_METADATA_URL      ?=

SNAPSHOT_URL            ?= https://snapshots.VirtEngine.network/$(SNAPSHOT_NETWORK)/latest
REMOTE_TEST_WORKDIR     ?= ~/go/src/github.com/virtengine/virtengine
REMOTE_TEST_HOST        ?=

MAX_VALIDATORS          := $(shell cat $(TEST_CONFIG) | jq -r '.validators | length' | tr -d '\n')

$(VE_INIT):
	$(ROOT_DIR)/script/upgrades.sh \
		--workdir=$(AP_RUN_DIR) \
		--gbv=$(GENESIS_BINARY_VERSION) \
		--ufrom=$(UPGRADE_FROM) \
		--uto=$(UPGRADE_TO) \
		--config="$(PWD)/config.json" \
		--chain-meta=$(CHAIN_METADATA_URL) \
		--state-config=$(STATE_CONFIG) \
		--snapshot-url=$(SNAPSHOT_URL) \
		--max-validators=$(MAX_VALIDATORS) \
		init
	touch $@

.PHONY: init
init: $(COSMOVISOR) $(VE_INIT)

.PHONY: genesis
genesis: $(GENESIS_DEST)

.PHONY: test
test: init
	$(GO_TEST) -run "^\QTestUpgrade\E$$" -tags e2e.upgrade -timeout 180m -v -args \
		-cosmovisor=$(COSMOVISOR) \
		-workdir=$(AP_RUN_DIR)/validators \
		-sourcesdir=$(VE_ROOT) \
		-config=$(TEST_CONFIG) \
		-upgrade-name=$(UPGRADE_TO) \
		-upgrade-version="$(UPGRADE_BINARY_VERSION)" \
		-test-cases=test-cases.json

.PHONY: test-reset
test-reset:
	$(ROOT_DIR)/script/upgrades.sh --workdir=$(AP_RUN_DIR) --config="$(PWD)/config.json" --uto=$(UPGRADE_TO) --snapshot-url=$(SNAPSHOT_URL) --chain-meta=$(CHAIN_METADATA_URL) --max-validators=$(MAX_VALIDATORS) clean
	$(ROOT_DIR)/script/upgrades.sh --workdir=$(AP_RUN_DIR) --config="$(PWD)/config.json" --uto=$(UPGRADE_TO) --snapshot-url=$(SNAPSHOT_URL) --gbv=$(GENESIS_BINARY_VERSION) --chain-meta=$(CHAIN_METADATA_URL) bins
	$(ROOT_DIR)/script/upgrades.sh --workdir=$(AP_RUN_DIR) --config="$(PWD)/config.json" --uto=$(UPGRADE_TO) --snapshot-url=$(SNAPSHOT_URL) --chain-meta=$(CHAIN_METADATA_URL) keys
	$(ROOT_DIR)/script/upgrades.sh --workdir=$(AP_RUN_DIR) --config="$(PWD)/config.json" --state-config=$(STATE_CONFIG) --snapshot-url=$(SNAPSHOT_URL) --chain-meta=$(CHAIN_METADATA_URL) --max-validators=$(MAX_VALIDATORS) prepare-state

.PHONY: prepare-state
prepare-state:
	$(ROOT_DIR)/script/upgrades.sh --workdir=$(AP_RUN_DIR) --config="$(PWD)/config.json" --state-config=$(STATE_CONFIG) --chain-meta=$(CHAIN_METADATA_URL) --max-validators=$(MAX_VALIDATORS) prepare-state

.PHONY: bins
bins:
ifneq ($(findstring build,$(SKIP)),build)
bins:
	$(ROOT_DIR)/script/upgrades.sh --workdir=$(AP_RUN_DIR) --config="$(PWD)/config.json" --uto=$(UPGRADE_TO) --gbv=$(GENESIS_BINARY_VERSION) --chain-meta=$(CHAIN_METADATA_URL) bins
endif

.PHONY: clean
clean:
	rm -rf $(AP_RUN_DIR)
