PROTO_GEN_MODS ?= go \
ts \
doc

RUST_SDK_DIR        := $(VIRTENGINE_ROOT)/rs
RUST_GEN_DIR        := $(RUST_SDK_DIR)/src/gen
BUF_GEN_RUST_CONFIG := buf.gen.rs.yaml

.PHONY: proto-gen
proto-gen: $(patsubst %, proto-gen-%,$(PROTO_GEN_MODS))

.PHONY: proto-gen-go
proto-gen-go: $(BUF) $(PROTOC) $(GOGOPROTO) $(PROTOC_GEN_GOCOSMOS) $(PROTOC_GEN_GRPC_GATEWAY) $(PROTOC_GEN_GO) $(PROTOC_GEN_GOGO) modvendor
	./script/protocgen.sh go $(GO_MOD_NAME) $(GO_ROOT)

.PHONY: proto-gen-rust
proto-gen-rust: proto-gen-rust-clean proto-gen-rust-buf proto-gen-rust-fmt

.PHONY: proto-gen-rust-clean
proto-gen-rust-clean:
	rm -rf $(RUST_GEN_DIR)
	mkdir -p $(RUST_GEN_DIR)

.PHONY: proto-gen-rust-buf
proto-gen-rust-buf: $(BUF) $(PROTOC) $(GOGOPROTO) $(PROTOC_GEN_PROST)
	$(BUF) generate --template $(BUF_GEN_RUST_CONFIG)

.PHONY: proto-gen-rust-fmt
proto-gen-rust-fmt:
	@echo "Formatting generated Rust code..."
	@if command -v $(RUSTFMT) > /dev/null 2>&1; then \
		find $(RUST_GEN_DIR) -name "*.rs" -exec $(RUSTFMT) {} \; 2>/dev/null || true; \
	else \
		echo "Warning: rustfmt not found, skipping formatting"; \
	fi

.PHONY: proto-gen-pulsar
proto-gen-pulsar: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_PULSAR)
	./script/protocgen.sh pulsar $(GO_MOD_NAME)

.PHONY: proto-gen-ts
proto-gen-ts: $(BUF) $(VIRTENGINE_TS_NODE_MODULES) modvendor
	./script/protocgen.sh ts

.PHONY: proto-gen-doc
proto-gen-doc: $(BUF) $(SWAGGER_COMBINE) $(PROTOC_GEN_DOC) $(PROTOC_GEN_SWAGGER)
	./script/protocgen.sh doc $(GO_MOD_NAME)

mocks: $(MOCKERY)
	(cd $(GO_ROOT); $(MOCKERY))

.PHONY: codegen
codegen: proto-gen mocks

.PHONY: changelog
changelog: $(GIT_CHGLOG)
	@echo "generating changelog to changelog"
	./script/changelog.sh $(shell git describe --tags --abbrev=0) changelog.md
