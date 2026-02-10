.PHONY: release-ts
release-ts: $(VIRTENGINE_TS_NODE_MODULES) $(VIRTENGINE_TS_ROOT)/dist $(BUF) modvendor
	cd $(VIRTENGINE_TS_ROOT) && npm run release

ifdef VIRTENGINE_TS_ROOT
TS_SRC_FILES := $(shell find $(VIRTENGINE_TS_ROOT)/src -type f 2>/dev/null || true)
$(VIRTENGINE_TS_ROOT)/dist: $(TS_SRC_FILES)
	cd $(VIRTENGINE_TS_ROOT) && npm run build
else
$(VIRTENGINE_TS_ROOT)/dist:
	@echo "Warning: VIRTENGINE_TS_ROOT not set, skipping TypeScript build"
endif
