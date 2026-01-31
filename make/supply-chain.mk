# SECURITY-9C: Supply Chain Security Make Targets
# 
# This file provides make targets for supply chain security operations
# including dependency verification, SBOM generation, and vulnerability scanning.

.PHONY: supply-chain-verify
supply-chain-verify: ## Verify all dependencies for integrity and security
	@echo "Running dependency verification..."
	@chmod +x ./scripts/supply-chain/verify-dependencies.sh 2>/dev/null || true
	@./scripts/supply-chain/verify-dependencies.sh --all

.PHONY: supply-chain-detect
supply-chain-detect: ## Detect potential supply chain attacks
	@echo "Running supply chain attack detection..."
	@chmod +x ./scripts/supply-chain/detect-supply-chain-attacks.sh 2>/dev/null || true
	@./scripts/supply-chain/detect-supply-chain-attacks.sh --all

.PHONY: supply-chain-risk
supply-chain-risk: ## Assess third-party dependency risk
	@echo "Running dependency risk assessment..."
	@cd scripts/supply-chain && go run assess-dependencies.go

.PHONY: supply-chain-risk-report
supply-chain-risk-report: ## Generate detailed dependency risk report
	@echo "Generating dependency risk report..."
	@mkdir -p .cache
	@cd scripts/supply-chain && go run assess-dependencies.go --report --json > ../../.cache/risk-assessment.json
	@cd scripts/supply-chain && go run assess-dependencies.go --report

.PHONY: sbom
sbom: ## Generate Software Bill of Materials (SBOM)
	@echo "Generating SBOM..."
	@chmod +x ./scripts/supply-chain/generate-sbom.sh 2>/dev/null || true
	@./scripts/supply-chain/generate-sbom.sh --format all

.PHONY: sbom-verify
sbom-verify: sbom ## Generate and verify SBOM for vulnerabilities
	@echo "Verifying SBOM for vulnerabilities..."
	@./scripts/supply-chain/generate-sbom.sh --format cyclonedx --verify

.PHONY: supply-chain-audit
supply-chain-audit: supply-chain-verify supply-chain-detect supply-chain-risk sbom ## Run full supply chain security audit
	@echo ""
	@echo "=========================================="
	@echo "  Supply Chain Audit Complete"
	@echo "=========================================="

.PHONY: deps-verify
deps-verify: ## Verify Go module checksums
	@echo "Verifying Go module checksums..."
	@go mod verify
	@echo "✓ All module checksums verified"

.PHONY: deps-vendor
deps-vendor: ## Vendor all dependencies (uses deps-tidy from mod.mk)
	@echo "Vendoring dependencies..."
	@$(MAKE) deps-tidy
	@go mod vendor
	@echo "✓ Dependencies vendored"

.PHONY: deps-update-check
deps-update-check: ## Check for available dependency updates
	@echo "Checking for dependency updates..."
	@go list -u -m all 2>/dev/null | grep '\[' || echo "All dependencies are up to date"

.PHONY: vuln-check
vuln-check: ## Run govulncheck for Go vulnerability scanning
	@echo "Running govulncheck..."
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

.PHONY: vuln-check-json
vuln-check-json: ## Run govulncheck with JSON output
	@echo "Running govulncheck (JSON output)..."
	@mkdir -p .cache
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck -format json ./... > .cache/govulncheck-report.json 2>/dev/null || true
	@echo "Report saved to .cache/govulncheck-report.json"

.PHONY: sign-artifact
sign-artifact: ## Sign an artifact with cosign (requires ARTIFACT env var)
ifndef ARTIFACT
	$(error ARTIFACT is required. Usage: make sign-artifact ARTIFACT=path/to/file)
endif
	@echo "Signing $(ARTIFACT)..."
	@command -v cosign >/dev/null 2>&1 || (echo "cosign not installed" && exit 1)
	@cosign sign-blob --yes \
		--output-signature "$(ARTIFACT).sig" \
		--output-certificate "$(ARTIFACT).pem" \
		"$(ARTIFACT)"
	@echo "✓ Signed: $(ARTIFACT).sig"

.PHONY: verify-signature
verify-signature: ## Verify an artifact signature (requires ARTIFACT env var)
ifndef ARTIFACT
	$(error ARTIFACT is required. Usage: make verify-signature ARTIFACT=path/to/file)
endif
	@echo "Verifying $(ARTIFACT)..."
	@command -v cosign >/dev/null 2>&1 || (echo "cosign not installed" && exit 1)
	@cosign verify-blob \
		--signature "$(ARTIFACT).sig" \
		--certificate "$(ARTIFACT).pem" \
		--certificate-identity-regexp ".*@virtengine.io" \
		--certificate-oidc-issuer https://token.actions.githubusercontent.com \
		"$(ARTIFACT)"
	@echo "✓ Signature verified"

# Help target for supply chain commands
.PHONY: help-supply-chain
help-supply-chain: ## Show supply chain security targets
	@echo "Supply Chain Security Targets:"
	@echo ""
	@echo "  Verification:"
	@echo "    supply-chain-verify   - Verify dependency integrity"
	@echo "    supply-chain-detect   - Detect supply chain attacks"
	@echo "    supply-chain-risk     - Assess dependency risk"
	@echo "    supply-chain-audit    - Full security audit"
	@echo ""
	@echo "  SBOM:"
	@echo "    sbom                  - Generate SBOM (all formats)"
	@echo "    sbom-verify           - Generate and scan SBOM"
	@echo ""
	@echo "  Dependencies:"
	@echo "    deps-verify           - Verify module checksums"
	@echo "    deps-vendor           - Vendor dependencies"
	@echo "    deps-update-check     - Check for updates"
	@echo ""
	@echo "  Vulnerabilities:"
	@echo "    vuln-check            - Run govulncheck"
	@echo "    vuln-check-json       - Run govulncheck (JSON)"
	@echo ""
	@echo "  Signing:"
	@echo "    sign-artifact         - Sign artifact with cosign"
	@echo "    verify-signature      - Verify artifact signature"
