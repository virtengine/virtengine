#!/usr/bin/env python3
"""
Reproducible Training Orchestration Script

VE-3A: Orchestrates deterministic training with proper seed setting,
environment configuration, and artifact generation.

Usage:
    python -m ml.training.run_training --config configs/trust_score_v1.yaml
    python -m ml.training.run_training --config configs/trust_score_v1.yaml --dry-run
"""

import argparse
import hashlib
import json
import logging
import os
import platform
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, Optional, Tuple

# Configure logging before other imports
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)


def setup_deterministic_environment(seed: int = 42, force_cpu: bool = True) -> Dict[str, str]:
    """
    Configure environment for deterministic execution.
    
    Returns:
        Dictionary of environment variables that were set
    """
    env_vars = {
        # Python hash seed for deterministic string hashing
        "PYTHONHASHSEED": str(seed),
        
        # TensorFlow determinism
        "TF_DETERMINISTIC_OPS": "1",
        "TF_CUDNN_DETERMINISTIC": "1",
        
        # Disable GPU if force_cpu
        "CUDA_VISIBLE_DEVICES": "" if force_cpu else os.environ.get("CUDA_VISIBLE_DEVICES", ""),
        
        # TensorFlow threading
        "TF_NUM_INTEROP_THREADS": "1",
        "TF_NUM_INTRAOP_THREADS": "1",
        
        # Disable TensorFlow warnings
        "TF_CPP_MIN_LOG_LEVEL": "2",
        
        # NumPy threading
        "OMP_NUM_THREADS": "1",
        "MKL_NUM_THREADS": "1",
        "OPENBLAS_NUM_THREADS": "1",
    }
    
    for key, value in env_vars.items():
        os.environ[key] = value
        
    return env_vars


def get_system_info() -> Dict[str, Any]:
    """Collect system information for reproducibility tracking."""
    info = {
        "platform": platform.platform(),
        "python_version": platform.python_version(),
        "machine": platform.machine(),
        "processor": platform.processor(),
        "timestamp": datetime.utcnow().isoformat(),
    }
    
    # Add TensorFlow version if available
    try:
        import tensorflow as tf
        info["tensorflow_version"] = tf.__version__
    except ImportError:
        info["tensorflow_version"] = "not installed"
    
    # Add NumPy version
    try:
        import numpy as np
        info["numpy_version"] = np.__version__
    except ImportError:
        info["numpy_version"] = "not installed"
    
    return info


def compute_config_hash(config_path: str) -> str:
    """Compute SHA256 hash of the configuration file."""
    with open(config_path, 'rb') as f:
        return hashlib.sha256(f.read()).hexdigest()


def load_config(config_path: str) -> Dict[str, Any]:
    """Load configuration from YAML file."""
    import yaml
    
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)
    
    # Expand environment variables in data paths
    if 'dataset' in config and 'data_paths' in config['dataset']:
        expanded_paths = []
        for path in config['dataset']['data_paths']:
            if path.startswith("${") and ":-" in path:
                # Handle default value syntax: ${VAR:-default}
                var_default = path[2:-1]  # Remove ${ and }
                var_name, default = var_default.split(":-", 1)
                expanded = os.environ.get(var_name, default)
            else:
                expanded = os.path.expandvars(path)
            expanded_paths.append(expanded)
        config['dataset']['data_paths'] = expanded_paths
    
    return config


def validate_config(config: Dict[str, Any]) -> Tuple[bool, list]:
    """
    Validate configuration for training.
    
    Returns:
        Tuple of (is_valid, list of errors)
    """
    errors = []
    
    # Check required fields
    required_fields = ['experiment_name', 'random_seed', 'dataset', 'model', 'export']
    for field in required_fields:
        if field not in config:
            errors.append(f"Missing required field: {field}")
    
    # Validate dataset paths
    if 'dataset' in config:
        data_paths = config['dataset'].get('data_paths', [])
        for path in data_paths:
            if not os.path.exists(path):
                errors.append(f"Dataset path does not exist: {path}")
    
    # Validate feature dimensions match
    if 'features' in config and 'model' in config:
        feature_dim = config['features'].get('combined_feature_dim', 768)
        input_dim = config['model'].get('input_dim', 768)
        if feature_dim != input_dim:
            errors.append(
                f"Feature dim ({feature_dim}) does not match model input dim ({input_dim})"
            )
    
    # Validate evaluation thresholds
    if 'evaluation' in config:
        eval_cfg = config['evaluation']
        if eval_cfg.get('min_r2', 0) > 1.0:
            errors.append("min_r2 cannot exceed 1.0")
        if eval_cfg.get('max_mae', 0) < 0:
            errors.append("max_mae cannot be negative")
    
    return len(errors) == 0, errors


def run_training(
    config_path: str,
    output_dir: Optional[str] = None,
    dry_run: bool = False,
    version: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Run the complete training pipeline.
    
    Args:
        config_path: Path to configuration file
        output_dir: Override output directory
        dry_run: If True, validate config but don't train
        version: Model version string
        
    Returns:
        Dictionary with training results
    """
    start_time = time.time()
    
    logger.info("=" * 70)
    logger.info("VIRTENGINE TRUST SCORE MODEL - REPRODUCIBLE TRAINING")
    logger.info("=" * 70)
    
    # Load configuration
    logger.info(f"Loading configuration from: {config_path}")
    config = load_config(config_path)
    config_hash = compute_config_hash(config_path)
    logger.info(f"Configuration hash: {config_hash[:16]}...")
    
    # Validate configuration
    is_valid, errors = validate_config(config)
    if not is_valid:
        logger.error("Configuration validation failed:")
        for error in errors:
            logger.error(f"  - {error}")
        return {"success": False, "errors": errors}
    
    logger.info("Configuration validation: PASSED")
    
    # Setup deterministic environment
    seed = config.get('random_seed', 42)
    force_cpu = config.get('determinism', {}).get('force_cpu', True)
    env_vars = setup_deterministic_environment(seed, force_cpu)
    logger.info(f"Deterministic environment configured (seed={seed}, cpu_only={force_cpu})")
    
    # Collect system info
    system_info = get_system_info()
    logger.info(f"Platform: {system_info['platform']}")
    logger.info(f"Python: {system_info['python_version']}")
    logger.info(f"TensorFlow: {system_info['tensorflow_version']}")
    
    if dry_run:
        logger.info("")
        logger.info("DRY RUN - Training would proceed with the following configuration:")
        logger.info(f"  Experiment: {config.get('experiment_name')}")
        logger.info(f"  Epochs: {config.get('model', {}).get('epochs', 100)}")
        logger.info(f"  Batch size: {config.get('model', {}).get('batch_size', 64)}")
        logger.info(f"  Learning rate: {config.get('model', {}).get('learning_rate', 0.001)}")
        return {"success": True, "dry_run": True}
    
    # Import training modules (after environment setup)
    import numpy as np
    np.random.seed(seed)
    
    import random
    random.seed(seed)
    
    # TensorFlow setup
    import tensorflow as tf
    tf.random.set_seed(seed)
    
    # Enable deterministic ops
    if config.get('determinism', {}).get('tf_deterministic_ops', True):
        try:
            tf.config.experimental.enable_op_determinism()
            logger.info("TensorFlow deterministic ops: ENABLED")
        except Exception as e:
            logger.warning(f"Could not enable TF determinism: {e}")
    
    # Import training components
    from ml.training.config import TrainingConfig
    from ml.training.dataset.ingestion import DatasetIngestion
    from ml.training.dataset.preprocessing import DatasetPreprocessor
    from ml.training.dataset.augmentation import DataAugmentation
    from ml.training.features.feature_combiner import FeatureExtractor
    from ml.training.model.architecture import TrustScoreModel
    from ml.training.model.training import ModelTrainer
    from ml.training.model.evaluation import ModelEvaluator
    from ml.training.model.export import ModelExporter
    from ml.training.model.manifest import ManifestGenerator
    
    # Create TrainingConfig from loaded config
    training_config = TrainingConfig.from_dict(config)
    
    # Override output directory if specified
    if output_dir:
        output_path = Path(output_dir)
    else:
        output_path = Path(config.get('model', {}).get('checkpoint_dir', 'output')).parent
    
    output_path.mkdir(parents=True, exist_ok=True)
    
    # Update config paths
    training_config.model.checkpoint_dir = str(output_path / "checkpoints")
    training_config.model.tensorboard_log_dir = str(output_path / "logs" / "tensorboard")
    training_config.export.export_dir = str(output_path / "exported_models")
    
    # Save configuration snapshot
    config_snapshot_path = output_path / "training_config.json"
    training_config.save_json(str(config_snapshot_path))
    logger.info(f"Configuration snapshot saved to: {config_snapshot_path}")
    
    try:
        # Step 1: Load dataset
        logger.info("")
        logger.info("STEP 1: Dataset Ingestion")
        logger.info("-" * 50)
        
        ingestion = DatasetIngestion(training_config.dataset)
        dataset = ingestion.load_dataset()
        
        logger.info(f"Dataset loaded: {len(dataset)} samples")
        
        # Step 2: Preprocessing
        logger.info("")
        logger.info("STEP 2: Data Preprocessing")
        logger.info("-" * 50)
        
        preprocessor = DatasetPreprocessor(training_config.preprocessing)
        preprocessed = preprocessor.preprocess_dataset(dataset)
        
        # Step 3: Augmentation
        if training_config.augmentation.enabled:
            logger.info("")
            logger.info("STEP 3: Data Augmentation")
            logger.info("-" * 50)
            
            augmentor = DataAugmentation(training_config.augmentation)
            train_augmented = augmentor.augment_batch(
                preprocessed.train,
                seed=training_config.random_seed
            )
        else:
            train_augmented = None
        
        # Step 4: Feature Extraction
        logger.info("")
        logger.info("STEP 4: Feature Extraction")
        logger.info("-" * 50)
        
        feature_extractor = FeatureExtractor(training_config.features)
        
        if train_augmented:
            train_features = feature_extractor.extract_from_augmented(train_augmented)
        else:
            train_features = feature_extractor.extract_from_preprocessed(preprocessed.train)
        
        val_features = feature_extractor.extract_from_preprocessed(preprocessed.validation)
        test_features = feature_extractor.extract_from_preprocessed(preprocessed.test)
        
        # Convert to arrays
        train_X = np.stack([f.combined_vector for f in train_features])
        train_y = np.array([f.trust_score for f in train_features])
        val_X = np.stack([f.combined_vector for f in val_features])
        val_y = np.array([f.trust_score for f in val_features])
        test_X = np.stack([f.combined_vector for f in test_features])
        test_y = np.array([f.trust_score for f in test_features])
        
        logger.info(f"Features: {train_X.shape[0]} train, {val_X.shape[0]} val, {test_X.shape[0]} test")
        
        # Step 5: Training
        logger.info("")
        logger.info("STEP 5: Model Training")
        logger.info("-" * 50)
        
        trainer = ModelTrainer(training_config)
        training_result = trainer.train_from_arrays(train_X, train_y, val_X, val_y)
        
        logger.info(f"Training completed in {training_result.training_time_seconds:.2f}s")
        logger.info(f"Best validation loss: {training_result.best_val_loss:.4f}")
        
        # Step 6: Evaluation
        logger.info("")
        logger.info("STEP 6: Model Evaluation")
        logger.info("-" * 50)
        
        evaluator = ModelEvaluator(training_config.model)
        metrics = evaluator.evaluate(training_result.model, test_X, test_y)
        
        # Check against thresholds
        eval_thresholds = config.get('evaluation', {})
        passed, threshold_results = check_evaluation_thresholds(metrics, eval_thresholds)
        
        # Generate report
        report_path = str(output_path / "evaluation_report.txt")
        report = evaluator.generate_report(metrics, report_path)
        logger.info(report)
        
        # Save metrics
        metrics.save(str(output_path / "evaluation_metrics.json"))
        
        # Log threshold results
        logger.info("")
        logger.info("EVALUATION THRESHOLDS")
        logger.info("-" * 50)
        for check, result in threshold_results.items():
            status = "✓ PASS" if result['passed'] else "✗ FAIL"
            logger.info(f"  {check}: {result['value']:.4f} (threshold: {result['threshold']}) {status}")
        
        if not passed:
            logger.error("Model FAILED to meet evaluation thresholds")
            return {
                "success": False,
                "error": "Model did not meet evaluation thresholds",
                "metrics": metrics.to_dict(),
                "threshold_results": threshold_results,
            }
        
        logger.info("Model PASSED all evaluation thresholds")
        
        # Step 7: Export
        logger.info("")
        logger.info("STEP 7: Model Export")
        logger.info("-" * 50)
        
        exporter = ModelExporter(training_config.export)
        model_version = version or f"v{datetime.utcnow().strftime('%Y%m%d_%H%M%S')}"
        export_result = exporter.export_savedmodel(
            training_result.model,
            training_config.export.export_dir,
            model_version
        )
        
        if not export_result.success:
            logger.error(f"Export failed: {export_result.error_message}")
            return {"success": False, "error": export_result.error_message}
        
        logger.info(f"Model exported to: {export_result.model_path}")
        logger.info(f"Model hash: {export_result.model_hash}")
        
        # Step 8: Generate Manifest
        logger.info("")
        logger.info("STEP 8: Manifest Generation")
        logger.info("-" * 50)
        
        manifest_generator = ManifestGenerator()
        manifest = manifest_generator.generate(
            export_result=export_result,
            config=training_config,
            config_hash=config_hash,
            metrics=metrics,
            system_info=system_info,
        )
        
        manifest_path = Path(export_result.model_path).parent / "manifest.json"
        manifest_generator.save(manifest, str(manifest_path))
        logger.info(f"Manifest saved to: {manifest_path}")
        
        # Summary
        total_time = time.time() - start_time
        logger.info("")
        logger.info("=" * 70)
        logger.info("TRAINING COMPLETE")
        logger.info("=" * 70)
        logger.info(f"Total time: {total_time:.2f}s")
        logger.info(f"Output directory: {output_path}")
        logger.info(f"Model version: {model_version}")
        logger.info(f"Model hash: {export_result.model_hash}")
        logger.info("")
        logger.info("Key metrics:")
        logger.info(f"  MAE: {metrics.mae:.4f}")
        logger.info(f"  RMSE: {metrics.rmse:.4f}")
        logger.info(f"  R²: {metrics.r2:.4f}")
        logger.info(f"  Accuracy@10: {metrics.accuracy_10:.1%}")
        
        return {
            "success": True,
            "model_path": export_result.model_path,
            "model_hash": export_result.model_hash,
            "model_version": model_version,
            "metrics": metrics.to_dict(),
            "manifest_path": str(manifest_path),
            "training_time_seconds": total_time,
        }
        
    except Exception as e:
        logger.exception(f"Training failed: {e}")
        return {"success": False, "error": str(e)}


def check_evaluation_thresholds(
    metrics,
    thresholds: Dict[str, Any]
) -> Tuple[bool, Dict[str, Dict[str, Any]]]:
    """
    Check if metrics meet evaluation thresholds.
    
    Returns:
        Tuple of (all_passed, detailed_results)
    """
    results = {}
    all_passed = True
    
    # R² threshold (higher is better)
    if 'min_r2' in thresholds:
        passed = metrics.r2 >= thresholds['min_r2']
        results['r2'] = {
            'value': metrics.r2,
            'threshold': f">= {thresholds['min_r2']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    # MAE threshold (lower is better)
    if 'max_mae' in thresholds:
        passed = metrics.mae <= thresholds['max_mae']
        results['mae'] = {
            'value': metrics.mae,
            'threshold': f"<= {thresholds['max_mae']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    # RMSE threshold (lower is better)
    if 'max_rmse' in thresholds:
        passed = metrics.rmse <= thresholds['max_rmse']
        results['rmse'] = {
            'value': metrics.rmse,
            'threshold': f"<= {thresholds['max_rmse']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    # Accuracy thresholds
    if 'min_accuracy_5' in thresholds:
        passed = metrics.accuracy_5 >= thresholds['min_accuracy_5']
        results['accuracy_5'] = {
            'value': metrics.accuracy_5,
            'threshold': f">= {thresholds['min_accuracy_5']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    if 'min_accuracy_10' in thresholds:
        passed = metrics.accuracy_10 >= thresholds['min_accuracy_10']
        results['accuracy_10'] = {
            'value': metrics.accuracy_10,
            'threshold': f">= {thresholds['min_accuracy_10']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    if 'min_accuracy_20' in thresholds:
        passed = metrics.accuracy_20 >= thresholds['min_accuracy_20']
        results['accuracy_20'] = {
            'value': metrics.accuracy_20,
            'threshold': f">= {thresholds['min_accuracy_20']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    # P95 error threshold
    if 'max_p95_error' in thresholds:
        passed = metrics.p95_error <= thresholds['max_p95_error']
        results['p95_error'] = {
            'value': metrics.p95_error,
            'threshold': f"<= {thresholds['max_p95_error']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    # Mean bias threshold
    if 'max_mean_bias' in thresholds:
        passed = abs(metrics.mean_error) <= thresholds['max_mean_bias']
        results['mean_bias'] = {
            'value': abs(metrics.mean_error),
            'threshold': f"<= {thresholds['max_mean_bias']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    # Sample count threshold
    if 'min_test_samples' in thresholds:
        passed = metrics.num_samples >= thresholds['min_test_samples']
        results['test_samples'] = {
            'value': metrics.num_samples,
            'threshold': f">= {thresholds['min_test_samples']}",
            'passed': passed
        }
        all_passed = all_passed and passed
    
    return all_passed, results


def parse_args() -> argparse.Namespace:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="Run reproducible training for trust score model"
    )
    
    parser.add_argument(
        "--config", "-c",
        type=str,
        required=True,
        help="Path to training configuration YAML file"
    )
    
    parser.add_argument(
        "--output-dir", "-o",
        type=str,
        help="Override output directory"
    )
    
    parser.add_argument(
        "--version", "-v",
        type=str,
        help="Model version string (default: auto-generated)"
    )
    
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Validate configuration without training"
    )
    
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose output"
    )
    
    return parser.parse_args()


def main():
    """Main entry point."""
    args = parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    result = run_training(
        config_path=args.config,
        output_dir=args.output_dir,
        dry_run=args.dry_run,
        version=args.version,
    )
    
    if result.get("success"):
        logger.info("Training completed successfully")
        sys.exit(0)
    else:
        logger.error(f"Training failed: {result.get('error', 'Unknown error')}")
        sys.exit(1)


if __name__ == "__main__":
    main()
