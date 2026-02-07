#!/usr/bin/env python3
"""
Trust Score Model Training Script

Main entry point for training the trust score model used in
VEID identity verification.

Usage:
    python -m ml.training.train --config config.yaml
    python -m ml.training.train --data-path /path/to/data --output-dir output
"""

import argparse
import json
import logging
import os
import random
import sys
import time
from pathlib import Path
from typing import Optional

import numpy as np

from ml.training.config import TrainingConfig

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)


def configure_determinism(
    seed: int,
    force_cpu: bool = True,
    inter_op_threads: int = 1,
    intra_op_threads: int = 1,
) -> None:
    """Configure deterministic environment and random seeds."""
    os.environ["PYTHONHASHSEED"] = str(seed)
    os.environ["TF_DETERMINISTIC_OPS"] = "1"
    os.environ["TF_CUDNN_DETERMINISTIC"] = "1"
    os.environ["TF_NUM_INTEROP_THREADS"] = str(inter_op_threads)
    os.environ["TF_NUM_INTRAOP_THREADS"] = str(intra_op_threads)
    os.environ["OMP_NUM_THREADS"] = str(intra_op_threads)
    os.environ["MKL_NUM_THREADS"] = str(intra_op_threads)
    os.environ["OPENBLAS_NUM_THREADS"] = str(intra_op_threads)
    os.environ["TF_CPP_MIN_LOG_LEVEL"] = "2"
    if force_cpu:
        os.environ["CUDA_VISIBLE_DEVICES"] = ""
    
    random.seed(seed)
    np.random.seed(seed)
    
    try:
        import tensorflow as tf
        tf.random.set_seed(seed)
        try:
            tf.config.experimental.enable_op_determinism()
        except Exception:
            logger.warning("Could not enable TensorFlow determinism")
        try:
            tf.config.threading.set_inter_op_parallelism_threads(inter_op_threads)
            tf.config.threading.set_intra_op_parallelism_threads(intra_op_threads)
        except Exception:
            logger.warning("Could not set TensorFlow threading config")
    except ImportError:
        logger.warning("TensorFlow not available, determinism setup skipped")


def parse_args() -> argparse.Namespace:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="Train the trust score model for VEID identity verification"
    )
    
    # Config file
    parser.add_argument(
        "--config", "-c",
        type=str,
        help="Path to YAML or JSON configuration file"
    )
    
    # Data paths
    parser.add_argument(
        "--data-path", "-d",
        type=str,
        nargs="+",
        help="Path(s) to training data directories"
    )
    
    # Output
    parser.add_argument(
        "--output-dir", "-o",
        type=str,
        default="output",
        help="Output directory for model and artifacts"
    )
    
    # Training parameters
    parser.add_argument(
        "--epochs",
        type=int,
        default=None,
        help="Number of training epochs"
    )
    
    parser.add_argument(
        "--batch-size",
        type=int,
        default=None,
        help="Training batch size"
    )
    
    parser.add_argument(
        "--learning-rate",
        type=float,
        default=None,
        help="Learning rate"
    )
    
    # Options
    parser.add_argument(
        "--no-augmentation",
        action="store_true",
        help="Disable data augmentation"
    )
    
    parser.add_argument(
        "--no-anonymize",
        action="store_true",
        help="Disable PII anonymization"
    )
    
    parser.add_argument(
        "--export-only",
        action="store_true",
        help="Only export model from checkpoint, don't train"
    )
    
    parser.add_argument(
        "--checkpoint",
        type=str,
        help="Path to checkpoint to resume from or export"
    )
    
    parser.add_argument(
        "--version",
        type=str,
        help="Model version string for export"
    )
    
    parser.add_argument(
        "--seed",
        type=int,
        default=42,
        help="Random seed for reproducibility"
    )
    
    parser.add_argument(
        "--cpu-only",
        action="store_true",
        default=True,
        help="Force CPU-only execution for determinism (default)"
    )
    
    parser.add_argument(
        "--allow-gpu",
        action="store_true",
        help="Allow GPU execution (may reduce determinism)"
    )
    
    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Enable verbose output"
    )
    
    return parser.parse_args()


def load_config(args: argparse.Namespace) -> TrainingConfig:
    """Load and merge configuration from file and CLI args."""
    from ml.training.config import TrainingConfig
    # Start with defaults
    config = TrainingConfig()
    
    # Load from file if provided
    if args.config:
        config_path = Path(args.config)
        if config_path.suffix == ".yaml" or config_path.suffix == ".yml":
            config = TrainingConfig.from_yaml(args.config)
        elif config_path.suffix == ".json":
            config = TrainingConfig.from_json(args.config)
        else:
            logger.warning(f"Unknown config format: {config_path.suffix}")
    
    # Override with CLI args
    if args.data_path:
        config.dataset.data_paths = args.data_path
    
    if args.epochs:
        config.model.epochs = args.epochs
    
    if args.batch_size:
        config.model.batch_size = args.batch_size
    
    if args.learning_rate:
        config.model.learning_rate = args.learning_rate
    
    if args.no_augmentation:
        config.augmentation.enabled = False
    
    if args.no_anonymize:
        config.dataset.anonymize = False
    
    if args.seed:
        config.random_seed = args.seed
        config.dataset.random_seed = args.seed
    
    # Set output paths
    output_dir = Path(args.output_dir)
    config.model.checkpoint_dir = str(output_dir / "checkpoints")
    config.model.tensorboard_log_dir = str(output_dir / "logs")
    config.export.export_dir = str(output_dir / "exported_models")
    
    return config


def main():
    """Main training entry point."""
    start_time = time.time()
    
    # Parse arguments
    args = parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    logger.info("=" * 60)
    logger.info("VIRTENGINE TRUST SCORE MODEL TRAINING")
    logger.info("=" * 60)
    
    # Load configuration
    config = load_config(args)
    logger.info(f"Configuration loaded: {config.experiment_name}")
    
    # Determinism configuration
    force_cpu = args.cpu_only and not args.allow_gpu
    configure_determinism(
        seed=config.random_seed,
        force_cpu=force_cpu,
        inter_op_threads=1,
        intra_op_threads=1,
    )
    
    # Import training components after determinism is configured
    from ml.training.dataset.ingestion import DatasetIngestion
    from ml.training.dataset.preprocessing import DatasetPreprocessor
    from ml.training.dataset.augmentation import DataAugmentation
    from ml.training.features.feature_combiner import FeatureExtractor
    from ml.training.model.training import ModelTrainer
    from ml.training.model.evaluation import ModelEvaluator
    from ml.training.model.export import ModelExporter
    
    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Save configuration
    config.save_json(str(output_dir / "config.json"))
    
    # Export only mode
    if args.export_only:
        if not args.checkpoint:
            logger.error("--checkpoint required for export-only mode")
            sys.exit(1)
        
        export_from_checkpoint(args.checkpoint, config, args.version)
        return
    
    # Check for data paths
    if not config.dataset.data_paths:
        logger.error("No data paths specified. Use --data-path or config file.")
        sys.exit(1)
    
    try:
        # Step 1: Load dataset
        logger.info("")
        logger.info("STEP 1: Dataset Ingestion")
        logger.info("-" * 40)
        
        ingestion = DatasetIngestion(config.dataset)
        dataset = ingestion.load_dataset()
        
        logger.info(f"Dataset loaded: {len(dataset)} samples")
        logger.info(f"  Train: {len(dataset.train)} samples")
        logger.info(f"  Validation: {len(dataset.validation)} samples")
        logger.info(f"  Test: {len(dataset.test)} samples")
        
        # Step 2: Preprocessing
        logger.info("")
        logger.info("STEP 2: Data Preprocessing")
        logger.info("-" * 40)
        
        preprocessor = DatasetPreprocessor(config.preprocessing)
        preprocessed = preprocessor.preprocess_dataset(dataset)
        
        # Step 3: Augmentation (training set only)
        if config.augmentation.enabled:
            logger.info("")
            logger.info("STEP 3: Data Augmentation")
            logger.info("-" * 40)
            
            augmentor = DataAugmentation(config.augmentation)
            train_augmented = augmentor.augment_batch(
                preprocessed.train,
                seed=config.random_seed
            )
            logger.info(f"Training samples after augmentation: {len(train_augmented)}")
        else:
            train_augmented = None
        
        # Step 4: Feature Extraction
        logger.info("")
        logger.info("STEP 4: Feature Extraction")
        logger.info("-" * 40)
        
        feature_extractor = FeatureExtractor(config.features)
        
        if train_augmented:
            train_features = feature_extractor.extract_from_augmented(train_augmented)
        else:
            train_features = feature_extractor.extract_from_preprocessed(preprocessed.train)
        
        val_features = feature_extractor.extract_from_preprocessed(preprocessed.validation)
        test_features = feature_extractor.extract_from_preprocessed(preprocessed.test)
        
        logger.info(f"Features extracted: {len(train_features)} train, "
                   f"{len(val_features)} val, {len(test_features)} test")
        
        # Convert to arrays
        train_X = np.stack([f.combined_vector for f in train_features])
        train_y = np.array([f.trust_score for f in train_features])
        val_X = np.stack([f.combined_vector for f in val_features])
        val_y = np.array([f.trust_score for f in val_features])
        test_X = np.stack([f.combined_vector for f in test_features])
        test_y = np.array([f.trust_score for f in test_features])
        
        # Step 5: Training
        logger.info("")
        logger.info("STEP 5: Model Training")
        logger.info("-" * 40)
        
        trainer = ModelTrainer(config)
        training_result = trainer.train_from_arrays(
            train_X, train_y,
            val_X, val_y,
        )
        
        logger.info(f"Training completed in {training_result.training_time_seconds:.2f}s")
        logger.info(f"Best validation loss: {training_result.best_val_loss:.4f}")
        
        # Step 6: Evaluation
        logger.info("")
        logger.info("STEP 6: Model Evaluation")
        logger.info("-" * 40)
        
        evaluator = ModelEvaluator(config.model)
        metrics = evaluator.evaluate(training_result.model, test_X, test_y)
        latency = evaluator.benchmark_latency(
            training_result.model,
            input_dim=config.model.input_dim,
            batch_size=1,
            warmup_runs=10,
            timed_runs=50,
            seed=config.random_seed,
        )
        
        report_path = str(output_dir / "evaluation_report.txt")
        report = evaluator.generate_report(metrics, report_path)
        print(report)
        
        # Save metrics
        metrics.save(str(output_dir / "evaluation_metrics.json"))
        
        # Step 7: Export
        logger.info("")
        logger.info("STEP 7: Model Export")
        logger.info("-" * 40)
        
        exporter = ModelExporter(config.export)
        version = args.version or f"v{config.experiment_name}"
        export_result = exporter.export_savedmodel(
            training_result.model,
            config.export.export_dir,
            version
        )
        
        if export_result.success:
            logger.info(f"Model exported to: {export_result.model_path}")
            logger.info(f"Model hash: {export_result.model_hash}")
            
            # Verify export
            if exporter.verify_export(export_result.model_path):
                logger.info("Export verification: PASSED")
            else:
                logger.warning("Export verification: FAILED")
            
            # Generate Go example
            go_example_path = str(output_dir / "go_inference_example.go")
            with open(go_example_path, 'w') as f:
                f.write(exporter.get_go_inference_example(export_result))
            
            # Save metrics report alongside export
            metrics_report = {
                "model_name": "trust_score",
                "version": export_result.version,
                "training_date": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "framework": "tensorflow",
                "framework_version": export_result.tensorflow_version,
                "metrics": {
                    "mae": metrics.mae,
                    "rmse": metrics.rmse,
                    "r2": metrics.r2,
                    "accuracy_5": metrics.accuracy_5,
                    "accuracy_10": metrics.accuracy_10,
                    "accuracy_20": metrics.accuracy_20,
                    "p95_error": metrics.p95_error,
                },
                "inference_latency_ms": latency,
                "training_config": {
                    "epochs": config.model.epochs,
                    "batch_size": config.model.batch_size,
                    "learning_rate": config.model.learning_rate,
                    "optimizer": config.model.optimizer,
                    "seed": config.random_seed,
                },
                "dataset": {
                    "train_samples": int(train_X.shape[0]),
                    "validation_samples": int(val_X.shape[0]),
                    "test_samples": int(test_X.shape[0]),
                },
                "determinism": {
                    "tf_deterministic_ops": config.deterministic,
                    "random_seed": config.random_seed,
                    "cpu_only": force_cpu,
                },
            }
            
            metrics_path = Path(export_result.model_path).parent / "metrics.json"
            with open(metrics_path, "w") as f:
                json.dump(metrics_report, f, indent=2)
        else:
            logger.error(f"Export failed: {export_result.error_message}")
        
        # Summary
        total_time = time.time() - start_time
        logger.info("")
        logger.info("=" * 60)
        logger.info("TRAINING COMPLETE")
        logger.info("=" * 60)
        logger.info(f"Total time: {total_time:.2f}s")
        logger.info(f"Output directory: {output_dir}")
        logger.info(f"Model version: {export_result.version if export_result.success else 'N/A'}")
        logger.info(f"Model hash: {export_result.model_hash if export_result.success else 'N/A'}")
        logger.info("")
        logger.info("Key metrics:")
        logger.info(f"  MAE: {metrics.mae:.4f}")
        logger.info(f"  RMSE: {metrics.rmse:.4f}")
        logger.info(f"  R²: {metrics.r2:.4f}")
        logger.info(f"  Accuracy@10: {metrics.accuracy_10:.1%}")
        
    except Exception as e:
        logger.exception(f"Training failed: {e}")
        sys.exit(1)


def export_from_checkpoint(
    checkpoint_path: str,
    config: TrainingConfig,
    version: Optional[str] = None
) -> None:
    """Export model from a checkpoint."""
    logger.info(f"Loading model from checkpoint: {checkpoint_path}")
    from ml.training.model.architecture import TrustScoreModel
    from ml.training.model.export import ModelExporter
    
    model = TrustScoreModel.load(checkpoint_path, config.model)
    exporter = ModelExporter(config.export)
    export_result = exporter.export_savedmodel(
        model,
        config.export.export_dir,
        version
    )
    
    if export_result.success:
        logger.info(f"Model exported to: {export_result.model_path}")
        logger.info(f"Model hash: {export_result.model_hash}")
    else:
        logger.error(f"Export failed: {export_result.error_message}")


if __name__ == "__main__":
    main()
