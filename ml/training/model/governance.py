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
