from __future__ import annotations

import importlib.util
import sys
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "validate_security_policies.py"


def load_validator_module():
    spec = importlib.util.spec_from_file_location("validate_security_policies", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


class SecurityPolicyIntegrationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.validator = load_validator_module()

    def assert_clean(self, path: Path, validator_name: str) -> None:
        validator = getattr(self.validator, validator_name)
        errors = validator(path)
        self.assertEqual(errors, [], f"{path} returned validation errors: {errors}")

    def test_security_workflow_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / ".github" / "workflows" / "security.yaml", "validate_workflow")

    def test_supply_chain_workflow_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / ".github" / "workflows" / "supply-chain.yaml", "validate_workflow")

    def test_license_workflow_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / ".github" / "workflows" / "license-compliance.yaml", "validate_workflow")

    def test_pr_security_workflow_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / ".github" / "workflows" / "pr-security-check.yaml", "validate_workflow")

    def test_gitleaks_policy_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / ".gitleaks.toml", "validate_gitleaks")

    def test_vulnerability_allowlist_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / ".vulnerability-allowlist.yaml", "validate_allowlist")

    def test_supply_chain_doc_is_clean(self) -> None:
        self.assert_clean(REPO_ROOT / "SUPPLY_CHAIN_SECURITY.md", "validate_supply_chain_doc")


if __name__ == "__main__":
    unittest.main()
