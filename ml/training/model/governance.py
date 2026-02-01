"""
Governance proposal generator for trust score model updates.

VE-3A: Generates on-chain governance proposals for updating the
trust_score_model version reference in the VEID module.
"""

import hashlib
import json
import logging
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List, Optional

logger = logging.getLogger(__name__)


@dataclass
class GovernanceProposal:
    """
    On-chain governance proposal for model version update.
    
    This proposal follows the Cosmos SDK governance module format
    for x/gov parameter change proposals.
    """
    
    # Proposal metadata
    title: str = ""
    description: str = ""
    proposal_type: str = "UpdateTrustScoreModel"
    
    # Model information
    model_version: str = ""
    model_hash: str = ""
    model_url: str = ""
    
    # Evaluation metrics (for voter information)
    metrics: Dict[str, float] = field(default_factory=dict)
    
    # Previous version (for rollback reference)
    previous_version: Optional[str] = None
    previous_hash: Optional[str] = None
    
    # Governance parameters
    voting_period: str = "604800s"  # 7 days
    deposit: str = "1000000uvirt"  # 1 VIRT
    
    # Timestamps
    created_at: str = ""
    
    # Config hash for reproducibility
    config_hash: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "title": self.title,
            "description": self.description,
            "proposal_type": self.proposal_type,
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "model_url": self.model_url,
            "metrics": self.metrics,
            "previous_version": self.previous_version,
            "previous_hash": self.previous_hash,
            "voting_period": self.voting_period,
            "deposit": self.deposit,
            "created_at": self.created_at,
            "config_hash": self.config_hash,
        }
    
    def to_cosmos_proposal(self) -> Dict[str, Any]:
        """
        Convert to Cosmos SDK proposal format.
        
        Returns format suitable for `virtengine tx gov submit-proposal`
        """
        return {
            "@type": "/virtengine.veid.v1.MsgUpdateTrustScoreModel",
            "authority": "virtengine10d07y265gmmuvt4z0w9aw880jnsr700jdufnyd",  # x/gov module account
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "model_url": self.model_url,
            "metadata": json.dumps({
                "title": self.title,
                "description": self.description,
                "metrics": self.metrics,
                "previous_version": self.previous_version,
                "config_hash": self.config_hash,
            }),
        }
    
    def to_proposal_json(self) -> str:
        """Generate JSON file for proposal submission."""
        proposal = {
            "messages": [self.to_cosmos_proposal()],
            "metadata": json.dumps({
                "title": self.title,
                "authors": ["VirtEngine Core Team"],
                "summary": self.description[:200],
                "details": self.description,
                "proposal_forum_url": "",
                "vote_option_context": "YES to approve model update, NO to reject",
            }),
            "deposit": self.deposit,
            "title": self.title,
            "summary": self.description[:200],
        }
        return json.dumps(proposal, indent=2)


class GovernanceProposalGenerator:
    """
    Generates governance proposals for trust score model updates.
    
    The generated proposals can be submitted via:
    ```
    virtengine tx gov submit-proposal proposal.json --from validator
    ```
    """
    
    def __init__(self):
        """Initialize the generator."""
        pass
    
    def generate(
        self,
        manifest,  # ModelManifest from manifest.py
        model_url: str,
        previous_manifest: Optional[Any] = None,
    ) -> GovernanceProposal:
        """
        Generate a governance proposal from a model manifest.
        
        Args:
            manifest: The model manifest
            model_url: URL where model artifact is published
            previous_manifest: Previous model manifest (optional)
            
        Returns:
            GovernanceProposal ready for submission
        """
        # Build description
        description = self._build_description(manifest, previous_manifest)
        
        # Build title
        title = f"Update Trust Score Model to {manifest.model_version}"
        
        proposal = GovernanceProposal(
            title=title,
            description=description,
            proposal_type="UpdateTrustScoreModel",
            model_version=manifest.model_version,
            model_hash=manifest.model_hash,
            model_url=model_url,
            metrics=manifest.evaluation_metrics,
            previous_version=manifest.previous_version,
            previous_hash=manifest.previous_hash,
            created_at=datetime.utcnow().isoformat(),
            config_hash=manifest.config_hash,
        )
        
        return proposal
    
    def _build_description(
        self,
        manifest,
        previous_manifest: Optional[Any] = None,
    ) -> str:
        """Build proposal description with metrics comparison."""
        lines = [
            f"# Trust Score Model Update: {manifest.model_version}",
            "",
            "## Summary",
            f"This proposal updates the VEID trust score model to version "
            f"`{manifest.model_version}`.",
            "",
            "## Model Details",
            f"- **Version**: {manifest.model_version}",
            f"- **Hash**: `{manifest.model_hash}`",
            f"- **TensorFlow Version**: {manifest.tensorflow_version}",
            f"- **Export Timestamp**: {manifest.export_timestamp}",
            "",
            "## Evaluation Metrics",
        ]
        
        # Add metrics
        metrics = manifest.evaluation_metrics
        lines.extend([
            f"- **MAE**: {metrics.get('mae', 0):.4f}",
            f"- **RMSE**: {metrics.get('rmse', 0):.4f}",
            f"- **R²**: {metrics.get('r2', 0):.4f}",
            f"- **Accuracy@5**: {metrics.get('accuracy_5', 0):.1%}",
            f"- **Accuracy@10**: {metrics.get('accuracy_10', 0):.1%}",
            f"- **Accuracy@20**: {metrics.get('accuracy_20', 0):.1%}",
            f"- **Test Samples**: {metrics.get('num_samples', 0):,}",
            "",
        ])
        
        # Add comparison with previous version if available
        if previous_manifest:
            prev_metrics = previous_manifest.evaluation_metrics
            lines.extend([
                "## Comparison with Previous Version",
                f"Previous version: `{previous_manifest.model_version}`",
                "",
                "| Metric | Previous | New | Change |",
                "|--------|----------|-----|--------|",
            ])
            
            for key in ['mae', 'rmse', 'r2', 'accuracy_10']:
                prev = prev_metrics.get(key, 0)
                new = metrics.get(key, 0)
                if key in ['mae', 'rmse']:
                    # Lower is better
                    change = prev - new
                    indicator = "✓" if change > 0 else "✗" if change < 0 else "="
                else:
                    # Higher is better
                    change = new - prev
                    indicator = "✓" if change > 0 else "✗" if change < 0 else "="
                
                lines.append(
                    f"| {key} | {prev:.4f} | {new:.4f} | {change:+.4f} {indicator} |"
                )
            
            lines.append("")
        
        # Add determinism info
        lines.extend([
            "## Determinism Settings",
            f"- **Deterministic**: {manifest.determinism_settings.get('deterministic', True)}",
            f"- **Random Seed**: {manifest.determinism_settings.get('random_seed', 42)}",
            f"- **Force CPU**: {manifest.determinism_settings.get('force_cpu', True)}",
            "",
            "## Verification",
            "Validators can verify the model by:",
            "1. Downloading the model from the published URL",
            "2. Computing SHA256 hash and comparing with the proposal hash",
            "3. Running inference tests with known inputs",
            "",
            "## Rollback",
            f"If this proposal passes and issues are discovered, a rollback "
            f"proposal can be submitted to revert to version "
            f"`{manifest.previous_version or 'N/A'}`.",
        ])
        
        return "\n".join(lines)
    
    def save_proposal(self, proposal: GovernanceProposal, filepath: str) -> None:
        """Save proposal to JSON file."""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        
        with open(filepath, 'w') as f:
            f.write(proposal.to_proposal_json())
        
        logger.info(f"Proposal saved to {filepath}")
    
    def generate_cli_commands(self, proposal: GovernanceProposal) -> str:
        """
        Generate CLI commands for proposal submission.
        
        Returns shell commands for submitting the proposal.
        """
        commands = [
            "# Trust Score Model Update - Governance Proposal",
            f"# Version: {proposal.model_version}",
            f"# Hash: {proposal.model_hash}",
            "",
            "# 1. Submit the proposal",
            "virtengine tx gov submit-proposal proposal.json \\",
            "  --from=<validator-key> \\",
            "  --chain-id=<chain-id> \\",
            "  --gas=auto \\",
            "  --gas-adjustment=1.5 \\",
            "  --fees=5000uvirt \\",
            "  -y",
            "",
            "# 2. Query the proposal (replace PROPOSAL_ID)",
            "virtengine query gov proposal <PROPOSAL_ID>",
            "",
            "# 3. Vote on the proposal",
            "virtengine tx gov vote <PROPOSAL_ID> yes \\",
            "  --from=<validator-key> \\",
            "  --chain-id=<chain-id> \\",
            "  --gas=auto \\",
            "  -y",
            "",
            "# 4. Query voting results",
            "virtengine query gov votes <PROPOSAL_ID>",
        ]
        
        return "\n".join(commands)


def generate_governance_proposal(
    manifest,
    model_url: str,
    output_path: str,
    previous_manifest: Optional[Any] = None,
) -> GovernanceProposal:
    """
    Convenience function to generate and save a governance proposal.
    
    Args:
        manifest: Model manifest
        model_url: URL where model is published
        output_path: Path to save proposal JSON
        previous_manifest: Previous model manifest (optional)
        
    Returns:
        Generated GovernanceProposal
    """
    generator = GovernanceProposalGenerator()
    proposal = generator.generate(manifest, model_url, previous_manifest)
    generator.save_proposal(proposal, output_path)
    
    # Also save CLI commands
    cli_path = output_path.replace('.json', '_commands.sh')
    with open(cli_path, 'w') as f:
        f.write(generator.generate_cli_commands(proposal))
    
    return proposal


# ============================================================================
# Governance Workflow Templates (VE-11B)
# ============================================================================

class GovernanceWorkflow:
    """
    Complete governance workflow for model updates.
    
    Provides templates and utilities for:
    - Pre-proposal checklist validation
    - Proposal submission automation
    - Validator notification generation
    - Post-approval activation workflow
    """
    
    def __init__(self):
        """Initialize the workflow manager."""
        self.generator = GovernanceProposalGenerator()
    
    def generate_pre_proposal_checklist(
        self,
        manifest,
        metrics,
    ) -> Dict[str, Any]:
        """
        Generate pre-proposal checklist for governance submission.
        
        Returns checklist with pass/fail status for each requirement.
        """
        checklist = {
            "timestamp": datetime.utcnow().isoformat(),
            "model_version": manifest.model_version,
            "model_hash": manifest.model_hash,
            "checks": [],
            "all_passed": True,
        }
        
        # Check 1: Evaluation passed
        check1 = {
            "name": "Evaluation Thresholds",
            "description": "Model meets all evaluation thresholds",
            "passed": manifest.evaluation_passed,
            "details": f"R²: {metrics.get('r2', 0):.4f}, MAE: {metrics.get('mae', 0):.4f}",
        }
        checklist["checks"].append(check1)
        checklist["all_passed"] &= check1["passed"]
        
        # Check 2: Model hash computed
        check2 = {
            "name": "Model Hash",
            "description": "SHA-256 hash computed for model files",
            "passed": len(manifest.model_hash) == 64,
            "details": f"Hash: {manifest.model_hash[:16]}...",
        }
        checklist["checks"].append(check2)
        checklist["all_passed"] &= check2["passed"]
        
        # Check 3: Determinism settings
        det_settings = manifest.determinism_settings
        check3 = {
            "name": "Determinism Configuration",
            "description": "CPU-only, fixed seed, deterministic ops enabled",
            "passed": (
                det_settings.get('force_cpu', False) and
                det_settings.get('deterministic', False) and
                det_settings.get('tf_deterministic_ops', False)
            ),
            "details": f"CPU: {det_settings.get('force_cpu')}, Seed: {det_settings.get('random_seed')}",
        }
        checklist["checks"].append(check3)
        checklist["all_passed"] &= check3["passed"]
        
        # Check 4: TensorFlow version
        check4 = {
            "name": "TensorFlow Version",
            "description": "TensorFlow version matches requirements",
            "passed": manifest.tensorflow_version.startswith("2.15"),
            "details": f"Version: {manifest.tensorflow_version}",
        }
        checklist["checks"].append(check4)
        checklist["all_passed"] &= check4["passed"]
        
        # Check 5: Minimum test samples
        min_samples = 100  # Default threshold
        num_samples = metrics.get('num_samples', 0)
        check5 = {
            "name": "Test Sample Count",
            "description": f"At least {min_samples} test samples evaluated",
            "passed": num_samples >= min_samples,
            "details": f"Samples: {num_samples}",
        }
        checklist["checks"].append(check5)
        checklist["all_passed"] &= check5["passed"]
        
        return checklist
    
    def generate_validator_notification(
        self,
        proposal: GovernanceProposal,
    ) -> str:
        """
        Generate notification message for validators about model update.
        """
        notification = f"""
================================================================================
VIRTENGINE MODEL UPDATE NOTIFICATION
================================================================================

A new trust score model update proposal has been submitted for governance voting.

PROPOSAL DETAILS
----------------
Title: {proposal.title}
Model Version: {proposal.model_version}
Model Hash: {proposal.model_hash}
Voting Period: {proposal.voting_period}

KEY METRICS
-----------
- MAE: {proposal.metrics.get('mae', 'N/A'):.4f}
- RMSE: {proposal.metrics.get('rmse', 'N/A'):.4f}
- R²: {proposal.metrics.get('r2', 'N/A'):.4f}
- Accuracy@10: {proposal.metrics.get('accuracy_10', 0) * 100:.1f}%

VALIDATOR ACTIONS REQUIRED
--------------------------
1. Review the proposal on-chain
2. Download and verify the model hash
3. Test model inference on your node (optional)
4. Cast your vote before voting period ends

VERIFICATION COMMANDS
---------------------
# Query the proposal
virtengine query gov proposal <PROPOSAL_ID>

# Download and verify model hash
curl -O <MODEL_URL>
sha256sum saved_model/* | sha256sum

# Vote on the proposal
virtengine tx gov vote <PROPOSAL_ID> yes --from <VALIDATOR_KEY>

================================================================================
"""
        return notification
    
    def generate_activation_workflow(
        self,
        proposal: GovernanceProposal,
    ) -> str:
        """
        Generate post-approval activation workflow instructions.
        """
        workflow = f"""
# Trust Score Model Activation Workflow
# Model Version: {proposal.model_version}
# Model Hash: {proposal.model_hash}

## Prerequisites
- Proposal has been approved via governance
- Activation delay blocks have passed
- Node is synced to latest block height

## Activation Steps

### 1. Download the model artifact
```bash
# From approved storage location
wget -O trust_score_model.tar.gz <APPROVED_MODEL_URL>

# Extract
tar -xzf trust_score_model.tar.gz -C ~/.virtengine/models/
```

### 2. Verify the model hash
```bash
# Compute hash of downloaded model
cd ~/.virtengine/models/trust_score/{proposal.model_version}/model
find . -type f | sort | xargs sha256sum | sha256sum

# Compare with approved hash
# Expected: {proposal.model_hash}
```

### 3. Update node configuration
```bash
# Update models.toml
cat >> ~/.virtengine/config/models.toml << EOF
[trust_score]
version = "{proposal.model_version}"
hash = "{proposal.model_hash}"
path = "models/trust_score/{proposal.model_version}/model"
EOF
```

### 4. Restart the node
```bash
systemctl restart virtengined
```

### 5. Verify model is loaded
```bash
virtengine query veid model-version trust_score

# Should show:
# model_id: <computed_model_id>
# version: {proposal.model_version}
# sha256_hash: {proposal.model_hash}
# status: active
```

### 6. Report model sync status
```bash
virtengine tx veid report-model-versions --from <VALIDATOR_KEY>
```

## Rollback Procedure

If issues are discovered after activation:

```bash
# 1. Stop the node
systemctl stop virtengined

# 2. Restore previous model
mv ~/.virtengine/models/trust_score/{proposal.model_version} ~/.virtengine/models/trust_score/{proposal.model_version}.failed
cp -r ~/.virtengine/models/trust_score/{proposal.previous_version or 'previous'} ~/.virtengine/models/trust_score/current

# 3. Update configuration to previous version
# 4. Restart node
# 5. Submit rollback governance proposal
```

## Support

For issues during activation:
- GitHub Issues: https://github.com/virtengine/virtengine/issues
- Discord: #validator-support
"""
        return workflow
    
    def save_workflow_artifacts(
        self,
        proposal: GovernanceProposal,
        manifest,
        metrics: Dict[str, float],
        output_dir: str,
    ) -> Dict[str, str]:
        """
        Save all governance workflow artifacts to output directory.
        
        Returns dict of artifact paths.
        """
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)
        
        artifacts = {}
        
        # Save proposal JSON
        proposal_path = output_path / "proposal.json"
        self.generator.save_proposal(proposal, str(proposal_path))
        artifacts["proposal"] = str(proposal_path)
        
        # Save CLI commands
        cli_path = output_path / "submit_proposal.sh"
        with open(cli_path, 'w') as f:
            f.write(self.generator.generate_cli_commands(proposal))
        artifacts["cli_commands"] = str(cli_path)
        
        # Save pre-proposal checklist
        checklist = self.generate_pre_proposal_checklist(manifest, metrics)
        checklist_path = output_path / "pre_proposal_checklist.json"
        with open(checklist_path, 'w') as f:
            json.dump(checklist, f, indent=2)
        artifacts["checklist"] = str(checklist_path)
        
        # Save validator notification
        notification = self.generate_validator_notification(proposal)
        notification_path = output_path / "validator_notification.txt"
        with open(notification_path, 'w') as f:
            f.write(notification)
        artifacts["notification"] = str(notification_path)
        
        # Save activation workflow
        activation = self.generate_activation_workflow(proposal)
        activation_path = output_path / "activation_workflow.md"
        with open(activation_path, 'w') as f:
            f.write(activation)
        artifacts["activation_workflow"] = str(activation_path)
        
        logger.info(f"Governance workflow artifacts saved to: {output_dir}")
        
        return artifacts


def generate_complete_governance_package(
    manifest,
    model_url: str,
    output_dir: str,
    previous_manifest: Optional[Any] = None,
) -> Dict[str, Any]:
    """
    Generate complete governance package with all artifacts.
    
    Args:
        manifest: Model manifest
        model_url: URL where model is published
        output_dir: Directory to save artifacts
        previous_manifest: Previous model manifest (optional)
        
    Returns:
        Dictionary with paths to all generated artifacts
    """
    workflow = GovernanceWorkflow()
    
    # Generate proposal
    proposal = workflow.generator.generate(manifest, model_url, previous_manifest)
    
    # Get metrics from manifest
    metrics = manifest.evaluation_metrics if hasattr(manifest, 'evaluation_metrics') else {}
    
    # Save all artifacts
    artifacts = workflow.save_workflow_artifacts(
        proposal=proposal,
        manifest=manifest,
        metrics=metrics,
        output_dir=output_dir,
    )
    
    return {
        "proposal": proposal,
        "artifacts": artifacts,
        "checklist_passed": workflow.generate_pre_proposal_checklist(manifest, metrics)["all_passed"],
    }
