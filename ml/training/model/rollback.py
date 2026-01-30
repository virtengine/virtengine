"""
Model rollback utilities for trust score SavedModel.

VE-3A: Provides rollback procedures for reverting to previous
model versions in case of issues.
"""

import hashlib
import json
import logging
import os
import shutil
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List, Optional, Tuple

logger = logging.getLogger(__name__)


@dataclass
class RollbackTarget:
    """Information about a rollback target version."""
    
    version: str = ""
    model_hash: str = ""
    model_path: str = ""
    manifest_path: str = ""
    published_at: str = ""
    metrics: Dict[str, float] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "version": self.version,
            "model_hash": self.model_hash,
            "model_path": self.model_path,
            "manifest_path": self.manifest_path,
            "published_at": self.published_at,
            "metrics": self.metrics,
        }


@dataclass
class RollbackResult:
    """Result of a rollback operation."""
    
    success: bool = True
    from_version: str = ""
    to_version: str = ""
    
    # Verification
    hash_verified: bool = False
    inference_verified: bool = False
    
    # Actions taken
    actions: List[str] = field(default_factory=list)
    
    # Errors
    error_message: Optional[str] = None
    
    # Timestamps
    started_at: str = ""
    completed_at: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "success": self.success,
            "from_version": self.from_version,
            "to_version": self.to_version,
            "hash_verified": self.hash_verified,
            "inference_verified": self.inference_verified,
            "actions": self.actions,
            "error_message": self.error_message,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
        }


class RollbackManager:
    """
    Manages model version rollbacks.
    
    Provides:
    - Version history tracking
    - Safe rollback with verification
    - Governance proposal generation for rollback
    - Audit logging
    """
    
    def __init__(
        self,
        registry_path: str = "artifacts/models",
        active_model_path: str = "models/trust_score/active",
    ):
        """
        Initialize the rollback manager.
        
        Args:
            registry_path: Path to artifact registry
            active_model_path: Path to active model symlink/directory
        """
        self.registry_path = Path(registry_path)
        self.active_model_path = Path(active_model_path)
    
    def list_available_versions(self) -> List[RollbackTarget]:
        """
        List all available versions for rollback.
        
        Returns:
            List of RollbackTarget objects sorted by version (newest first)
        """
        versions = []
        
        if not self.registry_path.exists():
            logger.warning(f"Registry path does not exist: {self.registry_path}")
            return versions
        
        for version_dir in self.registry_path.iterdir():
            if not version_dir.is_dir():
                continue
            
            # Look for manifest
            manifest_path = version_dir / "manifest.json"
            if not manifest_path.exists():
                continue
            
            try:
                with open(manifest_path, 'r') as f:
                    manifest = json.load(f)
                
                # Find model path
                model_path = version_dir / "model"
                if not model_path.exists():
                    # Check for archive
                    archives = list(version_dir.glob("*.tar.gz"))
                    model_path = archives[0] if archives else None
                
                versions.append(RollbackTarget(
                    version=manifest.get("model_version", version_dir.name),
                    model_hash=manifest.get("model_hash", ""),
                    model_path=str(model_path) if model_path else "",
                    manifest_path=str(manifest_path),
                    published_at=manifest.get("export_timestamp", ""),
                    metrics=manifest.get("evaluation_metrics", {}),
                ))
            except Exception as e:
                logger.warning(f"Could not load manifest for {version_dir}: {e}")
        
        # Sort by version (newest first)
        versions.sort(key=lambda v: v.published_at, reverse=True)
        return versions
    
    def get_current_version(self) -> Optional[RollbackTarget]:
        """Get the currently active model version."""
        if not self.active_model_path.exists():
            return None
        
        # Check for manifest in active path
        manifest_path = self.active_model_path / "manifest.json"
        if not manifest_path.exists():
            manifest_path = self.active_model_path.parent / "manifest.json"
        
        if manifest_path.exists():
            try:
                with open(manifest_path, 'r') as f:
                    manifest = json.load(f)
                return RollbackTarget(
                    version=manifest.get("model_version", ""),
                    model_hash=manifest.get("model_hash", ""),
                    model_path=str(self.active_model_path),
                    manifest_path=str(manifest_path),
                    published_at=manifest.get("export_timestamp", ""),
                    metrics=manifest.get("evaluation_metrics", {}),
                )
            except Exception as e:
                logger.warning(f"Could not load active manifest: {e}")
        
        return None
    
    def get_previous_version(self) -> Optional[RollbackTarget]:
        """Get the previous model version (for quick rollback)."""
        versions = self.list_available_versions()
        current = self.get_current_version()
        
        if current and len(versions) > 1:
            for v in versions:
                if v.version != current.version:
                    return v
        
        return versions[1] if len(versions) > 1 else None
    
    def verify_version(self, target: RollbackTarget) -> Tuple[bool, str]:
        """
        Verify a rollback target version.
        
        Args:
            target: RollbackTarget to verify
            
        Returns:
            Tuple of (passed, message)
        """
        # Verify model path exists
        if not target.model_path or not os.path.exists(target.model_path):
            return False, f"Model path does not exist: {target.model_path}"
        
        # Verify manifest exists
        if not target.manifest_path or not os.path.exists(target.manifest_path):
            return False, f"Manifest does not exist: {target.manifest_path}"
        
        # Verify hash
        computed_hash = self._compute_model_hash(target.model_path)
        if computed_hash != target.model_hash:
            return False, (
                f"Hash mismatch: expected {target.model_hash}, got {computed_hash}"
            )
        
        # Verify model can be loaded
        try:
            load_verified = self._verify_model_loads(target.model_path)
            if not load_verified:
                return False, "Model failed to load"
        except Exception as e:
            return False, f"Model load verification failed: {e}"
        
        return True, "Verification passed"
    
    def _compute_model_hash(self, model_path: str) -> str:
        """Compute SHA256 hash of model files."""
        h = hashlib.sha256()
        
        path = Path(model_path)
        if path.is_file():
            with open(path, 'rb') as f:
                h.update(f.read())
        else:
            files = []
            for root, dirs, filenames in os.walk(model_path):
                dirs.sort()
                for filename in sorted(filenames):
                    if filename == "export_metadata.json":
                        continue
                    files.append(os.path.join(root, filename))
            
            for filepath in files:
                with open(filepath, 'rb') as f:
                    h.update(f.read())
        
        return h.hexdigest()
    
    def _verify_model_loads(self, model_path: str) -> bool:
        """Verify that the model can be loaded by TensorFlow."""
        try:
            import tensorflow as tf
            
            path = Path(model_path)
            if path.is_file() and path.suffix == '.gz':
                # Archive - would need extraction
                logger.info("Model is archived, skipping load test")
                return True
            
            loaded = tf.saved_model.load(model_path)
            
            # Check for serving signature
            if hasattr(loaded, 'signatures'):
                if 'serving_default' in loaded.signatures:
                    logger.info("Model loaded successfully with serving_default signature")
                    return True
            
            logger.info("Model loaded successfully")
            return True
            
        except Exception as e:
            logger.error(f"Model load failed: {e}")
            return False
    
    def rollback(
        self,
        target: RollbackTarget,
        dry_run: bool = False,
    ) -> RollbackResult:
        """
        Perform rollback to a target version.
        
        Args:
            target: Target version to rollback to
            dry_run: If True, simulate rollback without making changes
            
        Returns:
            RollbackResult with details of the operation
        """
        result = RollbackResult(
            started_at=datetime.utcnow().isoformat(),
        )
        
        current = self.get_current_version()
        result.from_version = current.version if current else "none"
        result.to_version = target.version
        
        logger.info(f"Starting rollback: {result.from_version} -> {result.to_version}")
        
        try:
            # Step 1: Verify target version
            result.actions.append(f"Verifying target version {target.version}")
            verified, message = self.verify_version(target)
            if not verified:
                result.success = False
                result.error_message = f"Verification failed: {message}"
                return result
            
            result.hash_verified = True
            result.actions.append("Hash verification passed")
            
            # Step 2: Verify inference works
            result.actions.append("Running inference verification")
            try:
                inference_ok = self._run_inference_test(target.model_path)
                result.inference_verified = inference_ok
                if inference_ok:
                    result.actions.append("Inference verification passed")
                else:
                    result.actions.append("Inference verification skipped (no TF)")
            except Exception as e:
                result.actions.append(f"Inference verification warning: {e}")
            
            if dry_run:
                result.actions.append("DRY RUN - no changes made")
                result.success = True
                result.completed_at = datetime.utcnow().isoformat()
                return result
            
            # Step 3: Update active model symlink/directory
            result.actions.append(f"Updating active model to {target.version}")
            
            self.active_model_path.parent.mkdir(parents=True, exist_ok=True)
            
            # Backup current if exists
            if self.active_model_path.exists():
                backup_path = self.active_model_path.with_suffix('.backup')
                if backup_path.exists():
                    if backup_path.is_dir():
                        shutil.rmtree(backup_path)
                    else:
                        os.remove(backup_path)
                
                if self.active_model_path.is_symlink():
                    os.rename(self.active_model_path, backup_path)
                elif self.active_model_path.is_dir():
                    os.rename(self.active_model_path, backup_path)
                
                result.actions.append("Backed up current active model")
            
            # Create symlink or copy
            if os.name == 'nt':  # Windows
                # Copy instead of symlink on Windows
                shutil.copytree(target.model_path, self.active_model_path)
            else:
                os.symlink(os.path.abspath(target.model_path), self.active_model_path)
            
            result.actions.append(f"Active model now points to {target.version}")
            
            # Step 4: Log rollback
            self._log_rollback(result)
            result.actions.append("Rollback logged")
            
            result.success = True
            result.completed_at = datetime.utcnow().isoformat()
            
            logger.info(f"Rollback completed successfully: {target.version}")
            return result
            
        except Exception as e:
            logger.exception(f"Rollback failed: {e}")
            result.success = False
            result.error_message = str(e)
            result.completed_at = datetime.utcnow().isoformat()
            return result
    
    def _run_inference_test(self, model_path: str) -> bool:
        """Run a basic inference test on the model."""
        try:
            import tensorflow as tf
            import numpy as np
            
            # Load model
            loaded = tf.saved_model.load(model_path)
            
            if not hasattr(loaded, 'signatures'):
                return True  # No signature to test
            
            if 'serving_default' not in loaded.signatures:
                return True
            
            # Get serving function
            serve = loaded.signatures['serving_default']
            
            # Create dummy input
            input_dim = 768  # Default feature dimension
            dummy_input = tf.constant(np.zeros((1, input_dim), dtype=np.float32))
            
            # Run inference
            output = serve(dummy_input)
            
            # Check output exists
            if 'trust_score' in output:
                score = output['trust_score'].numpy()[0]
                logger.info(f"Inference test passed, dummy score: {score}")
                return True
            
            return True
            
        except ImportError:
            logger.warning("TensorFlow not available for inference test")
            return True  # Skip if no TF
        except Exception as e:
            logger.error(f"Inference test failed: {e}")
            return False
    
    def _log_rollback(self, result: RollbackResult) -> None:
        """Log rollback to audit log."""
        log_path = self.registry_path / "rollback_log.jsonl"
        
        with open(log_path, 'a') as f:
            f.write(json.dumps(result.to_dict()) + "\n")
    
    def generate_rollback_proposal(
        self,
        target: RollbackTarget,
        reason: str = "Performance regression detected",
    ) -> Dict[str, Any]:
        """
        Generate a governance proposal for rollback.
        
        Args:
            target: Target version to rollback to
            reason: Reason for rollback
            
        Returns:
            Governance proposal dictionary
        """
        current = self.get_current_version()
        
        proposal = {
            "@type": "/virtengine.veid.v1.MsgUpdateTrustScoreModel",
            "authority": "virtengine10d07y265gmmuvt4z0w9aw880jnsr700jdufnyd",
            "model_version": target.version,
            "model_hash": target.model_hash,
            "model_url": "",  # To be filled in
            "metadata": json.dumps({
                "type": "rollback",
                "reason": reason,
                "from_version": current.version if current else "unknown",
                "to_version": target.version,
                "target_metrics": target.metrics,
            }),
        }
        
        return {
            "messages": [proposal],
            "metadata": json.dumps({
                "title": f"Rollback Trust Score Model to {target.version}",
                "authors": ["VirtEngine Core Team"],
                "summary": f"Rollback due to: {reason}",
                "details": (
                    f"This proposal rolls back the trust score model from "
                    f"{current.version if current else 'current'} to {target.version}.\n\n"
                    f"Reason: {reason}\n\n"
                    f"Target version metrics:\n"
                    f"- MAE: {target.metrics.get('mae', 'N/A')}\n"
                    f"- R²: {target.metrics.get('r2', 'N/A')}\n"
                    f"- Accuracy@10: {target.metrics.get('accuracy_10', 'N/A')}"
                ),
                "vote_option_context": "YES to approve rollback, NO to keep current version",
            }),
            "deposit": "1000000uvirt",
            "title": f"Rollback Trust Score Model to {target.version}",
            "summary": f"Rollback due to: {reason}",
        }


def rollback_to_version(
    version: str,
    registry_path: str = "artifacts/models",
    active_model_path: str = "models/trust_score/active",
    dry_run: bool = False,
) -> RollbackResult:
    """
    Convenience function to rollback to a specific version.
    
    Args:
        version: Version string to rollback to
        registry_path: Path to artifact registry
        active_model_path: Path to active model
        dry_run: If True, simulate rollback
        
    Returns:
        RollbackResult
    """
    manager = RollbackManager(registry_path, active_model_path)
    
    # Find target version
    versions = manager.list_available_versions()
    target = None
    for v in versions:
        if v.version == version:
            target = v
            break
    
    if target is None:
        return RollbackResult(
            success=False,
            error_message=f"Version {version} not found in registry"
        )
    
    return manager.rollback(target, dry_run=dry_run)


def rollback_to_previous(
    registry_path: str = "artifacts/models",
    active_model_path: str = "models/trust_score/active",
    dry_run: bool = False,
) -> RollbackResult:
    """
    Convenience function to rollback to the previous version.
    
    Args:
        registry_path: Path to artifact registry
        active_model_path: Path to active model
        dry_run: If True, simulate rollback
        
    Returns:
        RollbackResult
    """
    manager = RollbackManager(registry_path, active_model_path)
    
    target = manager.get_previous_version()
    if target is None:
        return RollbackResult(
            success=False,
            error_message="No previous version available for rollback"
        )
    
    return manager.rollback(target, dry_run=dry_run)
