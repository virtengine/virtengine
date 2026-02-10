MODULES ?= go \
ts

.PHONY: code-style-staged
code-style-staged: $(patsubst %, code-style-staged-%,$(MODULES))

.PHONY: code-style-staged-ts
code-style-staged-ts: $(VIRTENGINE_TS_NODE_MODULES)
	ts/node_modules/.bin/lint-staged --cwd ts

.PHONY: code-style-staged-go
code-style-staged-go:
	script/gofmt-staged.sh
