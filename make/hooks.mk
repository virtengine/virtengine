# Git hooks management
#
# Usage:
#   make setup-hooks    Install pre-push quality gate
#   make remove-hooks   Remove hook configuration
#   make check-hooks    Check if hooks are installed

GITHOOKS_DIR := $(VE_ROOT)/.githooks

.PHONY: setup-hooks
setup-hooks:
	@chmod +x $(GITHOOKS_DIR)/* 2>/dev/null || true
	@git config core.hooksPath .githooks
	@echo ""
	@echo "Git hooks installed."
	@echo ""
	@echo "  Pre-push checks:"
	@echo "    1. go vet"
	@echo "    2. go mod tidy (sync check)"
	@echo "    3. go mod vendor (sync check)"
	@echo "    4. golangci-lint"
	@echo "    5. Proto generation (staleness check)"
	@echo "    6. Build binaries"
	@echo "    7. Unit tests"
	@echo ""
	@echo "  Bypass:      git push --no-verify"
	@echo "  Quick mode:  VE_HOOK_QUICK=1 git push"
	@echo "  Remove:      make remove-hooks"

.PHONY: remove-hooks
remove-hooks:
	@git config --unset core.hooksPath 2>/dev/null || true
	@echo "Git hooks removed."

.PHONY: check-hooks
check-hooks:
	@if [ "$$(git config core.hooksPath 2>/dev/null)" = ".githooks" ]; then \
		echo "Git hooks are installed."; \
	else \
		echo "Git hooks are NOT installed. Run: make setup-hooks"; \
	fi
