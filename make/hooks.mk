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

# ── Windows Firewall Setup ──────────────────────────────────────────────────
#
# Go test binaries are compiled to unique paths under the build cache.
# On Windows, each new binary may trigger a Firewall popup. These targets
# configure Windows Firewall to suppress those popups.
#
# Usage:
#   make setup-firewall       Install firewall rules (prompts for admin)
#   make check-firewall       Check if rules are installed
#   make remove-firewall      Remove firewall rules

.PHONY: setup-firewall
setup-firewall:
	@if command -v pwsh >/dev/null 2>&1; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/setup-firewall.ps1; \
	else \
		echo "setup-firewall is only needed on Windows (pwsh not found)"; \
	fi

.PHONY: check-firewall
check-firewall:
	@if command -v pwsh >/dev/null 2>&1; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/setup-firewall.ps1 -Check; \
	else \
		echo "Firewall check is only needed on Windows"; \
	fi

.PHONY: remove-firewall
remove-firewall:
	@if command -v pwsh >/dev/null 2>&1; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/setup-firewall.ps1 -Remove; \
	else \
		echo "Firewall removal is only needed on Windows"; \
	fi
