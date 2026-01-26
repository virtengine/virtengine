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
import logging
import sys
import time
from pathlib import Path
from typing import Optional

import numpy as np

from ml.training.config import TrainingConfig, DatasetConfig
from ml.training.dataset.ingestion import DatasetIngestion
from ml.training.dataset.preprocessing import DatasetPreprocessor
from ml.training.dataset.augmentation import DataAugmentation
from ml.training.dataset.anonymization import PIIAnonymizer
from ml.training.features.feature_combiner import FeatureExtractor
from ml.training.model.architecture import TrustScoreModel
from ml.training.model.training import ModelTrainer
from ml.training.model.evaluation import ModelEvaluator
from ml.training.model.export import ModelExporter

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)


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
        "--verbose", "-v",
        action="store_true",
        help="Enable verbose output"
    )
    
    return parser.parse_args()


def load_config(args: argparse.Namespace) -> TrainingConfig:
    """Load and merge configuration from file and CLI args."""
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
    
    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Save configuration
    config.saVIRTENGINE_json(str(output_dir / "config.json"))
    
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
