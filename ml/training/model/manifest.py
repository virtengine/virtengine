"""
Model manifest generation for trust score SavedModel.

VE-3A: Generates signed manifests with model hash, version, and governance metadata
for on-chain model version tracking.
"""

import hashlib
import json
import logging
import os
import time
from dataclasses import dataclass, field, asdict
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List, Optional

logger = logging.getLogger(__name__)


@dataclass
class ModelManifest:
    """
    Complete manifest for a trained trust score model.
    
    This manifest is used for:
    - On-chain governance proposals for model updates
    - Validator hash verification during inference
    - Rollback tracking and audit logging
    """
    
    # Model identity
    model_name: str = "trust_score"
    model_version: str = ""
    model_hash: str = ""
    
    # Export information
    export_timestamp: str = ""
    export_format: str = "SavedModel"
    tensorflow_version: str = ""
    
    # Model architecture
    input_signature: Dict[str, Any] = field(default_factory=dict)
    output_signature: Dict[str, Any] = field(default_factory=dict)
    operations: List[str] = field(default_factory=list)
    
    # Training configuration
    config_version: str = ""
    config_hash: str = ""
    feature_schema_version: str = ""
    dataset_version: str = ""
    
    # Training parameters
    training_parameters: Dict[str, Any] = field(default_factory=dict)
    
    # Determinism settings
    determinism_settings: Dict[str, Any] = field(default_factory=dict)
    
    # Evaluation metrics
    evaluation_metrics: Dict[str, float] = field(default_factory=dict)
    evaluation_passed: bool = False
    
    # System information
    system_info: Dict[str, str] = field(default_factory=dict)
    
    # File information
    model_path: str = ""
    model_size_bytes: int = 0
    
    # Governance
    governance: Dict[str, Any] = field(default_factory=dict)
    
    # Previous version (for rollback)
    previous_version: Optional[str] = None
    previous_hash: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return asdict(self)
    
    def to_json(self, indent: int = 2) -> str:
        """Convert to JSON string."""
        return json.dumps(self.to_dict(), indent=indent, default=str)
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ModelManifest":
        """Create manifest from dictionary."""
        return cls(**{k: v for k, v in data.items() if k in cls.__dataclass_fields__})
    
    @classmethod
    def from_json(cls, json_str: str) -> "ModelManifest":
        """Create manifest from JSON string."""
        data = json.loads(json_str)
        return cls.from_dict(data)
    
    def compute_manifest_hash(self) -> str:
        """Compute hash of the manifest content (excluding signature)."""
        # Create a copy without governance signature
        data = self.to_dict()
        if 'governance' in data:
            data['governance'].pop('signature', None)
        
        content = json.dumps(data, sort_keys=True, default=str)
        return hashlib.sha256(content.encode()).hexdigest()


class ManifestGenerator:
    """
    Generates model manifests for trust score SavedModels.
    
    The manifest includes all information required for:
    - Determinism verification
    - On-chain governance
    - Model rollback
    """
    
    def __init__(self):
        """Initialize the manifest generator."""
        pass
    
    def generate(
        self,
        export_result,  # ExportResult from export.py
        config,         # TrainingConfig
        config_hash: str,
        metrics,        # EvaluationMetrics
        system_info: Dict[str, str],
        previous_version: Optional[str] = None,
        previous_hash: Optional[str] = None,
    ) -> ModelManifest:
        """
        Generate a complete model manifest.
        
        Args:
            export_result: Result from model export
            config: Training configuration
            config_hash: SHA256 hash of config file
            metrics: Evaluation metrics
            system_info: System information dictionary
            previous_version: Previous model version (for rollback)
            previous_hash: Previous model hash
            
        Returns:
            ModelManifest with all fields populated
        """
        # Extract TensorFlow operations from the model
        operations = self._extract_operations(export_result.model_path)
        
        manifest = ModelManifest(
            model_name="trust_score",
            model_version=export_result.version,
            model_hash=export_result.model_hash,
            export_timestamp=export_result.export_timestamp or datetime.utcnow().isoformat(),
            export_format="SavedModel",
            tensorflow_version=export_result.tensorflow_version,
            input_signature=export_result.input_signature,
            output_signature=export_result.output_signature,
            operations=operations,
            config_version=getattr(config, 'config_version', '1.0.0'),
            config_hash=config_hash,
            feature_schema_version=getattr(config, 'feature_schema_version', '1.0.0'),
            dataset_version=getattr(config, 'dataset_version', 'unknown'),
            training_parameters=self._extract_training_params(config),
            determinism_settings=self._extract_determinism_settings(config),
            evaluation_metrics=self._extract_metrics(metrics),
            evaluation_passed=self._check_evaluation_passed(metrics, config),
            system_info=system_info,
            model_path=export_result.model_path,
            model_size_bytes=export_result.model_size_bytes,
            governance=self._create_governance_metadata(),
            previous_version=previous_version,
            previous_hash=previous_hash,
        )
        
        return manifest
    
    def _extract_operations(self, model_path: str) -> List[str]:
        """Extract TensorFlow operation names from SavedModel."""
        operations = []
        
        try:
            import tensorflow as tf
            
            # Load the saved model
            loaded_model = tf.saved_model.load(model_path)
            
            # Get operation names from the graph
            if hasattr(loaded_model, 'signatures'):
                for sig_name, sig in loaded_model.signatures.items():
                    # Get the concrete function
                    if hasattr(sig, 'graph'):
                        for op in sig.graph.get_operations():
                            if op.name not in operations:
                                operations.append(op.type)
            
            # Remove duplicates and sort
            operations = sorted(set(operations))
            
        except Exception as e:
            logger.warning(f"Could not extract operations: {e}")
        
        return operations
    
    def _extract_training_params(self, config) -> Dict[str, Any]:
        """Extract relevant training parameters from config."""
        model_cfg = config.model if hasattr(config, 'model') else {}
        
        return {
            "epochs": getattr(model_cfg, 'epochs', 100),
            "batch_size": getattr(model_cfg, 'batch_size', 64),
            "learning_rate": getattr(model_cfg, 'learning_rate', 0.001),
            "optimizer": getattr(model_cfg, 'optimizer', 'adam'),
            "hidden_layers": getattr(model_cfg, 'hidden_layers', [512, 256, 128, 64]),
            "dropout_rate": getattr(model_cfg, 'dropout_rate', 0.3),
            "l2_regularization": getattr(model_cfg, 'l2_regularization', 0.01),
            "random_seed": getattr(config, 'random_seed', 42),
        }
    
    def _extract_determinism_settings(self, config) -> Dict[str, Any]:
        """Extract determinism settings from config."""
        return {
            "deterministic": getattr(config, 'deterministic', True),
            "random_seed": getattr(config, 'random_seed', 42),
            "force_cpu": True,  # Always true for consensus
            "tf_deterministic_ops": True,
            "inter_op_parallelism": 1,
            "intra_op_parallelism": 1,
        }
    
    def _extract_metrics(self, metrics) -> Dict[str, float]:
        """Extract evaluation metrics."""
        return {
            "mae": float(getattr(metrics, 'mae', 0)),
            "mse": float(getattr(metrics, 'mse', 0)),
            "rmse": float(getattr(metrics, 'rmse', 0)),
            "r2": float(getattr(metrics, 'r2', 0)),
            "accuracy_5": float(getattr(metrics, 'accuracy_5', 0)),
            "accuracy_10": float(getattr(metrics, 'accuracy_10', 0)),
            "accuracy_20": float(getattr(metrics, 'accuracy_20', 0)),
            "p50_error": float(getattr(metrics, 'p50_error', 0)),
            "p95_error": float(getattr(metrics, 'p95_error', 0)),
            "num_samples": int(getattr(metrics, 'num_samples', 0)),
        }
    
    def _check_evaluation_passed(self, metrics, config) -> bool:
        """Check if evaluation metrics meet thresholds."""
        # Default thresholds
        min_r2 = 0.85
        max_mae = 8.0
        min_accuracy_10 = 0.80
        
        # Get thresholds from config if available
        if hasattr(config, 'evaluation'):
            eval_cfg = config.evaluation
            min_r2 = getattr(eval_cfg, 'min_r2', min_r2)
            max_mae = getattr(eval_cfg, 'max_mae', max_mae)
            min_accuracy_10 = getattr(eval_cfg, 'min_accuracy_10', min_accuracy_10)
        
        # Check thresholds
        r2_pass = getattr(metrics, 'r2', 0) >= min_r2
        mae_pass = getattr(metrics, 'mae', float('inf')) <= max_mae
        acc_pass = getattr(metrics, 'accuracy_10', 0) >= min_accuracy_10
        
        return r2_pass and mae_pass and acc_pass
    
    def _create_governance_metadata(self) -> Dict[str, Any]:
        """Create governance metadata for on-chain proposal."""
        return {
            "proposal_type": "UpdateTrustScoreModel",
            "requires_governance": True,
            "min_approval_percentage": 0.67,  # 2/3 majority
            "voting_period_blocks": 100800,   # ~7 days at 6s blocks
            "created_at": datetime.utcnow().isoformat(),
            "signature": None,  # To be signed by release key
        }
    
    def save(self, manifest: ModelManifest, filepath: str) -> None:
        """Save manifest to JSON file."""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        
        with open(filepath, 'w') as f:
            f.write(manifest.to_json())
        
        logger.info(f"Manifest saved to {filepath}")
    
    def load(self, filepath: str) -> ModelManifest:
        """Load manifest from JSON file."""
        with open(filepath, 'r') as f:
            return ModelManifest.from_json(f.read())
    
    def verify_manifest(self, manifest: ModelManifest, model_path: str) -> bool:
        """
        Verify that a manifest matches the model on disk.
        
        Args:
            manifest: The manifest to verify
            model_path: Path to the SavedModel
            
        Returns:
            True if verification passes
        """
        # Compute model hash
        computed_hash = self._compute_model_hash(model_path)
        
        if computed_hash != manifest.model_hash:
            logger.error(f"Model hash mismatch: expected {manifest.model_hash}, got {computed_hash}")
            return False
        
        logger.info("Manifest verification: PASSED")
        return True
    
    def _compute_model_hash(self, model_path: str) -> str:
        """Compute SHA256 hash of model files."""
        h = hashlib.sha256()
        
        # Walk through all files in the model directory
        files = []
        for root, dirs, filenames in os.walk(model_path):
            dirs.sort()
            for filename in sorted(filenames):
                # Skip metadata files
                if filename == "export_metadata.json":
                    continue
                files.append(os.path.join(root, filename))
        
        for filepath in files:
            with open(filepath, 'rb') as f:
                h.update(f.read())
        
        return h.hexdigest()


def generate_manifest(
    export_result,
    config,
    config_hash: str,
    metrics,
    system_info: Dict[str, str],
    output_path: str,
) -> ModelManifest:
    """
    Convenience function to generate and save a manifest.
    
    Args:
        export_result: Result from model export
        config: Training configuration
        config_hash: SHA256 hash of config file
        metrics: Evaluation metrics
        system_info: System information
        output_path: Path to save manifest
        
    Returns:
        Generated ModelManifest
    """
    generator = ManifestGenerator()
    manifest = generator.generate(
        export_result=export_result,
        config=config,
        config_hash=config_hash,
        metrics=metrics,
        system_info=system_info,
    )
    generator.save(manifest, output_path)
    return manifest
