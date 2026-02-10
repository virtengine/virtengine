#!/usr/bin/env python3
"""
VEID Dataset Build CLI

Command-line tool for building production VEID training datasets.

Usage:
    python -m ml.training.build_dataset --help
    python -m ml.training.build_dataset build --source /path/to/data --output /path/to/output
    python -m ml.training.build_dataset validate --dataset /path/to/dataset
    python -m ml.training.build_dataset synthetic --output /path/to/output --samples 1000
"""

import argparse
import hashlib
import json
import logging
import os
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

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
        description="VEID Dataset Build Tool",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # Build dataset from local data
    python -m ml.training.build_dataset build \\
        --source /data/veid/raw \\
        --output /data/veid/processed \\
        --version 1.0.0

    # Build from multiple sources
    python -m ml.training.build_dataset build \\
        --source /data/source1 /data/source2 s3://bucket/prefix \\
        --output /data/output

    # Generate synthetic dataset for CI
    python -m ml.training.build_dataset synthetic \\
        --output /data/synthetic \\
        --samples 500 \\
        --profile ci_standard

    # Validate existing dataset
    python -m ml.training.build_dataset validate \\
        --dataset /data/veid/processed
        
    # Apply labels from human review
    python -m ml.training.build_dataset label \\
        --dataset /data/veid/processed \\
        --labels /reviews/labels.csv
""",
    )
    
    # Global options
    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Enable verbose logging",
    )
    
    parser.add_argument(
        "--config",
        type=str,
        help="Path to configuration file (YAML or JSON)",
    )
    
    parser.add_argument(
        "--seed",
        type=int,
        default=42,
        help="Random seed for reproducibility (default: 42)",
    )
    
    subparsers = parser.add_subparsers(dest="command", help="Commands")
    
    # Build command
    build_parser = subparsers.add_parser(
        "build",
        help="Build a dataset from source data",
    )
    build_parser.add_argument(
        "--source", "-s",
        type=str,
        nargs="+",
        required=True,
        help="Source data paths (local paths or URIs)",
    )
    build_parser.add_argument(
        "--output", "-o",
        type=str,
        required=True,
        help="Output directory for processed dataset",
    )
    build_parser.add_argument(
        "--version",
        type=str,
        default=None,
        help="Dataset version (default: auto-generated)",
    )
    build_parser.add_argument(
        "--schema-version",
        type=str,
        default="1.0.0",
        help="Schema version (default: 1.0.0)",
    )
    build_parser.add_argument(
        "--train-split",
        type=float,
        default=0.7,
        help="Training set ratio (default: 0.7)",
    )
    build_parser.add_argument(
        "--val-split",
        type=float,
        default=0.15,
        help="Validation set ratio (default: 0.15)",
    )
    build_parser.add_argument(
        "--test-split",
        type=float,
        default=0.15,
        help="Test set ratio (default: 0.15)",
    )
    build_parser.add_argument(
        "--no-anonymize",
        action="store_true",
        help="Skip PII anonymization",
    )
    build_parser.add_argument(
        "--no-encrypt",
        action="store_true",
        help="Skip encryption of raw assets",
    )
    build_parser.add_argument(
        "--labels",
        type=str,
        help="Path to human labels CSV file",
    )
    build_parser.add_argument(
        "--sign",
        action="store_true",
        help="Sign the output manifest",
    )
    build_parser.add_argument(
        "--signer-id",
        type=str,
        default="dataset-builder",
        help="Signer ID for manifest signing",
    )
    
    # Synthetic command
    synth_parser = subparsers.add_parser(
        "synthetic",
        help="Generate synthetic dataset for testing",
    )
    synth_parser.add_argument(
        "--output", "-o",
        type=str,
        required=True,
        help="Output directory for synthetic dataset",
    )
    synth_parser.add_argument(
        "--samples", "-n",
        type=int,
        default=100,
        help="Number of samples to generate (default: 100)",
    )
    synth_parser.add_argument(
        "--profile",
        type=str,
        choices=["ci_minimal", "ci_standard", "dev_small", "dev_medium", "dev_large", "benchmark"],
        help="Predefined profile for generation",
    )
    synth_parser.add_argument(
        "--no-images",
        action="store_true",
        help="Skip image generation",
    )
    synth_parser.add_argument(
        "--fraud-ratio",
        type=float,
        default=0.1,
        help="Ratio of fraud samples (default: 0.1)",
    )
    
    # Validate command
    val_parser = subparsers.add_parser(
        "validate",
        help="Validate an existing dataset",
    )
    val_parser.add_argument(
        "--dataset", "-d",
        type=str,
        required=True,
        help="Path to dataset to validate",
    )
    val_parser.add_argument(
        "--report",
        type=str,
        help="Path to save validation report",
    )
    val_parser.add_argument(
        "--fail-on-error",
        action="store_true",
        help="Exit with error code if validation fails",
    )
    val_parser.add_argument(
        "--expected-hash",
        type=str,
        help="Expected dataset hash for verification",
    )
    
    # Label command
    label_parser = subparsers.add_parser(
        "label",
        help="Apply labels to dataset",
    )
    label_parser.add_argument(
        "--dataset", "-d",
        type=str,
        required=True,
        help="Path to dataset",
    )
    label_parser.add_argument(
        "--labels", "-l",
        type=str,
        required=True,
        help="Path to labels CSV file",
    )
    label_parser.add_argument(
        "--output", "-o",
        type=str,
        help="Output path (default: update in place)",
    )
    label_parser.add_argument(
        "--export-for-review",
        type=str,
        help="Export unlabeled samples for human review",
    )
    
    # Info command
    info_parser = subparsers.add_parser(
        "info",
        help="Display dataset information",
    )
    info_parser.add_argument(
        "--dataset", "-d",
        type=str,
        required=True,
        help="Path to dataset",
    )
    info_parser.add_argument(
        "--format",
        type=str,
        choices=["text", "json"],
        default="text",
        help="Output format (default: text)",
    )
    
    return parser.parse_args()


def cmd_build(args: argparse.Namespace) -> int:
    """Execute build command."""
    from ml.training.config import DatasetConfig, TrainingConfig
    from ml.training.dataset.connectors import ConnectorRegistry
    from ml.training.dataset.ingestion import DatasetIngestion, IdentitySample
    from ml.training.dataset.splits import DeterministicSplitter, SplitConfig, SplitStrategy
    from ml.training.dataset.labeling import LabelingPipeline
    from ml.training.dataset.validation import DatasetValidator
    from ml.training.dataset.manifest import ManifestBuilder, ManifestSigner
    from ml.training.dataset.lineage import LineageTracker, TransformType, SourceType
    from ml.training.dataset.anonymization import PIIAnonymizer
    
    logger.info("=" * 60)
    logger.info("VEID DATASET BUILD")
    logger.info("=" * 60)
    
    start_time = time.time()
    
    # Create output directory
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Generate version if not provided
    version = args.version or datetime.now().strftime("%Y%m%d-%H%M%S")
    
    # Initialize lineage tracking
    tracker = LineageTracker(
        dataset_name="veid_trust",
        dataset_version=version,
        schema_version=args.schema_version,
    )
    
    # Step 1: Load data from sources
    logger.info("\n[Step 1/6] Loading data from sources...")
    
    all_samples = []
    
    for source_path in args.source:
        logger.info(f"  Loading from: {source_path}")
        
        # Determine source type
        if source_path.startswith("s3://"):
            source_type = SourceType.OBJECT_STORE
        elif source_path.startswith("gs://"):
            source_type = SourceType.OBJECT_STORE
        elif source_path.startswith("http://") or source_path.startswith("https://"):
            source_type = SourceType.API
        else:
            source_type = SourceType.LOCAL_FILE
        
        try:
            connector = ConnectorRegistry.from_uri(source_path)
            result = connector.connect()
            
            if not result.success:
                logger.warning(f"  Failed to connect: {result.message}")
                continue
            
            # Load samples
            config = DatasetConfig(
                data_paths=[source_path],
                random_seed=args.seed,
                anonymize=False,  # We'll anonymize separately
            )
            ingestion = DatasetIngestion(config)
            dataset = ingestion.load_dataset()
            
            samples = list(dataset.train) + list(dataset.validation) + list(dataset.test)
            all_samples.extend(samples)
            
            # Track source
            tracker.add_source(
                location=source_path,
                source_type=source_type,
                record_count=len(samples),
            )
            
            logger.info(f"  Loaded {len(samples)} samples")
            
        except Exception as e:
            logger.error(f"  Error loading from {source_path}: {e}")
    
    if len(all_samples) == 0:
        logger.error("No samples loaded from any source")
        return 1
    
    logger.info(f"  Total samples loaded: {len(all_samples)}")
    
    # Step 2: Anonymize data
    if not args.no_anonymize:
        logger.info("\n[Step 2/6] Anonymizing PII...")
        
        anonymizer = PIIAnonymizer()
        
        # Create temporary dataset for anonymization
        from ml.training.dataset.ingestion import Dataset, DatasetSplit, SplitType
        temp_dataset = Dataset(
            train=DatasetSplit(SplitType.TRAIN, all_samples),
            validation=DatasetSplit(SplitType.VALIDATION, []),
            test=DatasetSplit(SplitType.TEST, []),
        )
        
        anon_result = anonymizer.anonymize_dataset(temp_dataset)
        logger.info(f"  Anonymized {anon_result.fields_anonymized} fields")
        
        tracker.add_transform(
            transform_type=TransformType.ANONYMIZATION,
            description="PII anonymization with SHA256 hashing",
            input_count=len(all_samples),
            output_count=len(all_samples),
        )
    else:
        logger.info("\n[Step 2/6] Skipping anonymization (--no-anonymize)")
    
    # Step 3: Apply labels
    logger.info("\n[Step 3/6] Applying labels...")
    
    pipeline = LabelingPipeline()
    
    # Create temp dataset for labeling
    from ml.training.dataset.ingestion import Dataset, DatasetSplit, SplitType
    temp_dataset = Dataset(
        train=DatasetSplit(SplitType.TRAIN, all_samples),
        validation=DatasetSplit(SplitType.VALIDATION, []),
        test=DatasetSplit(SplitType.TEST, []),
    )
    
    temp_dataset, label_report = pipeline.label_dataset(
        temp_dataset,
        human_labels_path=args.labels,
    )
    
    all_samples = list(temp_dataset.train)
    
    human_count = sum(1 for s in all_samples if hasattr(s, '_human_labeled'))
    heuristic_count = len(all_samples) - human_count
    logger.info(f"  Applied labels: {human_count} human, {heuristic_count} heuristic")
    
    tracker.add_transform(
        transform_type=TransformType.LABELING,
        description="Label application (human + heuristic)",
        input_count=len(all_samples),
        output_count=len(all_samples),
    )
    
    # Step 4: Split dataset
    logger.info("\n[Step 4/6] Splitting dataset...")
    
    split_config = SplitConfig(
        train_ratio=args.train_split,
        val_ratio=args.val_split,
        test_ratio=args.test_split,
        random_seed=args.seed,
        strategy=SplitStrategy.STRATIFIED,
    )
    
    splitter = DeterministicSplitter(split_config)
    split_result = splitter.split(all_samples)
    dataset = split_result.dataset
    
    logger.info(f"  Train: {len(dataset.train)} samples")
    logger.info(f"  Validation: {len(dataset.validation)} samples")
    logger.info(f"  Test: {len(dataset.test)} samples")
    logger.info(f"  Split hash: {split_result.combined_hash}")
    
    tracker.add_transform(
        transform_type=TransformType.SPLITTING,
        description=f"Stratified split (seed={args.seed})",
        config={"train": args.train_split, "val": args.val_split, "test": args.test_split},
        input_count=len(all_samples),
        output_count=len(all_samples),
        output_hash=split_result.combined_hash,
    )
    
    # Step 5: Validate dataset
    logger.info("\n[Step 5/6] Validating dataset...")
    
    validator = DatasetValidator(fail_on_error=False)
    validation_report = validator.validate(dataset)
    
    if validation_report.valid:
        logger.info("  Validation: PASSED")
    else:
        logger.warning(f"  Validation: FAILED ({validation_report.error_count} errors)")
    
    # Step 6: Save outputs
    logger.info("\n[Step 6/6] Saving outputs...")
    
    # Save manifest
    manifest_builder = ManifestBuilder(
        dataset_name="veid_trust",
        schema_version=args.schema_version,
    )
    
    for sample in list(dataset.train):
        manifest_builder.add_sample(
            sample.sample_id,
            json.dumps(sample.to_dict()).encode(),
            "train",
        )
    
    for sample in list(dataset.validation):
        manifest_builder.add_sample(
            sample.sample_id,
            json.dumps(sample.to_dict()).encode(),
            "validation",
        )
    
    for sample in list(dataset.test):
        manifest_builder.add_sample(
            sample.sample_id,
            json.dumps(sample.to_dict()).encode(),
            "test",
        )
    
    manifest = manifest_builder.build(version)
    
    # Sign manifest if requested
    if args.sign:
        signer = ManifestSigner(args.signer_id)
        manifest = signer.sign(manifest)
        logger.info(f"  Manifest signed by: {args.signer_id}")
    
    manifest.save(str(output_dir / "manifest.json"))
    logger.info(f"  Saved manifest: {output_dir / 'manifest.json'}")
    
    # Save lineage
    lineage = tracker.finalize(
        final_hash=split_result.combined_hash,
        sample_count=len(all_samples),
    )
    lineage.save(str(output_dir / "lineage.json"))
    logger.info(f"  Saved lineage: {output_dir / 'lineage.json'}")
    
    # Save validation report
    validation_report.save(str(output_dir / "validation_report.json"))
    logger.info(f"  Saved validation report: {output_dir / 'validation_report.json'}")
    
    # Save dataset manifest for ingestion
    dataset_manifest = {
        "version": version,
        "schema_version": args.schema_version,
        "splits": {
            "train": len(dataset.train),
            "validation": len(dataset.validation),
            "test": len(dataset.test),
        },
        "hash": split_result.combined_hash,
        "samples": [],
    }
    
    for split_name, split in [
        ("train", dataset.train),
        ("validation", dataset.validation),
        ("test", dataset.test),
    ]:
        for sample in split:
            sample_data = sample.to_dict()
            sample_data["split"] = split_name
            dataset_manifest["samples"].append(sample_data)
    
    with open(output_dir / "dataset.json", "w") as f:
        json.dump(dataset_manifest, f, indent=2)
    logger.info(f"  Saved dataset: {output_dir / 'dataset.json'}")
    
    # Summary
    elapsed = time.time() - start_time
    
    logger.info("\n" + "=" * 60)
    logger.info("BUILD COMPLETE")
    logger.info("=" * 60)
    logger.info(f"  Version: {version}")
    logger.info(f"  Total samples: {len(all_samples)}")
    logger.info(f"  Dataset hash: {split_result.combined_hash}")
    logger.info(f"  Output directory: {output_dir}")
    logger.info(f"  Elapsed time: {elapsed:.2f}s")
    
    return 0 if validation_report.valid else 1


def cmd_synthetic(args: argparse.Namespace) -> int:
    """Execute synthetic data generation command."""
    from ml.training.dataset.synthetic import (
        SyntheticDataGenerator,
        SyntheticConfig,
        SyntheticProfile,
    )
    
    logger.info("=" * 60)
    logger.info("SYNTHETIC DATASET GENERATION")
    logger.info("=" * 60)
    
    # Create config
    if args.profile:
        profile = SyntheticProfile(args.profile)
        config = SyntheticConfig.from_profile(profile)
        config.random_seed = args.seed
        logger.info(f"Using profile: {args.profile}")
    else:
        config = SyntheticConfig(
            num_samples=args.samples,
            random_seed=args.seed,
            generate_images=not args.no_images,
            fraud_ratio=args.fraud_ratio,
            save_to_disk=True,
            output_dir=args.output,
        )
    
    config.output_dir = args.output
    config.save_to_disk = True
    
    # Generate
    generator = SyntheticDataGenerator(config)
    dataset = generator.generate_dataset()
    
    logger.info("\n" + "=" * 60)
    logger.info("GENERATION COMPLETE")
    logger.info("=" * 60)
    logger.info(f"  Total samples: {len(dataset)}")
    logger.info(f"  Train: {len(dataset.train)}")
    logger.info(f"  Validation: {len(dataset.validation)}")
    logger.info(f"  Test: {len(dataset.test)}")
    logger.info(f"  Dataset hash: {dataset.dataset_hash}")
    logger.info(f"  Output directory: {args.output}")
    
    return 0


def cmd_validate(args: argparse.Namespace) -> int:
    """Execute validation command."""
    from ml.training.dataset.validation import DatasetValidator, ValidationReport
    from ml.training.dataset.ingestion import DatasetIngestion
    from ml.training.config import DatasetConfig
    
    logger.info("=" * 60)
    logger.info("DATASET VALIDATION")
    logger.info("=" * 60)
    
    dataset_path = Path(args.dataset)
    
    if not dataset_path.exists():
        logger.error(f"Dataset not found: {dataset_path}")
        return 1
    
    # Load dataset
    logger.info(f"Loading dataset from: {dataset_path}")
    
    config = DatasetConfig(data_paths=[str(dataset_path)])
    ingestion = DatasetIngestion(config)
    dataset = ingestion.load_dataset()
    
    logger.info(f"  Loaded {len(dataset)} samples")
    
    # Validate
    validator = DatasetValidator(fail_on_error=False)
    report = validator.validate(dataset)
    
    # Check expected hash if provided
    if args.expected_hash:
        # Compute current hash
        from ml.training.dataset.splits import DeterministicSplitter
        sample_ids = sorted([
            s.sample_id
            for s in list(dataset.train) + list(dataset.validation) + list(dataset.test)
        ])
        current_hash = hashlib.sha256("|".join(sample_ids).encode()).hexdigest()[:16]
        
        if current_hash != args.expected_hash:
            report.add_issue(type("ValidationIssue", (), {
                "level": type("ValidationLevel", (), {"value": "error"})(),
                "validation_type": type("ValidationType", (), {"value": "integrity"})(),
                "message": f"Hash mismatch: expected {args.expected_hash}, got {current_hash}",
                "sample_id": None,
                "field": None,
                "value": current_hash,
                "expected": args.expected_hash,
                "to_dict": lambda self: {
                    "level": "error",
                    "type": "integrity",
                    "message": f"Hash mismatch: expected {args.expected_hash}, got {current_hash}",
                },
            })())
            report.valid = False
    
    # Save report if requested
    if args.report:
        report.save(args.report)
        logger.info(f"  Saved report to: {args.report}")
    
    # Print summary
    print("\n" + report.print_summary())
    
    if args.fail_on_error and not report.valid:
        return 1
    
    return 0


def cmd_label(args: argparse.Namespace) -> int:
    """Execute label application command."""
    from ml.training.dataset.labeling import (
        LabelingPipeline,
        CSVLabelReader,
        CSVLabelWriter,
    )
    from ml.training.dataset.ingestion import DatasetIngestion
    from ml.training.config import DatasetConfig
    
    logger.info("=" * 60)
    logger.info("APPLY LABELS")
    logger.info("=" * 60)
    
    # Load dataset
    logger.info(f"Loading dataset from: {args.dataset}")
    
    config = DatasetConfig(data_paths=[args.dataset])
    ingestion = DatasetIngestion(config)
    dataset = ingestion.load_dataset()
    
    logger.info(f"  Loaded {len(dataset)} samples")
    
    # Apply labels
    pipeline = LabelingPipeline()
    dataset, report = pipeline.label_dataset(
        dataset,
        human_labels_path=args.labels,
        export_path=args.export_for_review,
    )
    
    logger.info(f"  Applied labels from: {args.labels}")
    
    if args.export_for_review:
        logger.info(f"  Exported samples for review: {args.export_for_review}")
    
    # Save updated dataset
    output_path = args.output or args.dataset
    logger.info(f"  Saving to: {output_path}")
    
    # Save as manifest
    import json
    output_dir = Path(output_path)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    manifest = {
        "samples": [],
    }
    
    for split_name, split in [
        ("train", dataset.train),
        ("validation", dataset.validation),
        ("test", dataset.test),
    ]:
        for sample in split:
            sample_data = sample.to_dict()
            sample_data["split"] = split_name
            manifest["samples"].append(sample_data)
    
    with open(output_dir / "manifest.json", "w") as f:
        json.dump(manifest, f, indent=2)
    
    logger.info("  Labels applied successfully")
    
    return 0


def cmd_info(args: argparse.Namespace) -> int:
    """Execute info command."""
    from ml.training.dataset.ingestion import DatasetIngestion
    from ml.training.config import DatasetConfig
    
    dataset_path = Path(args.dataset)
    
    if not dataset_path.exists():
        logger.error(f"Dataset not found: {dataset_path}")
        return 1
    
    # Load dataset
    config = DatasetConfig(data_paths=[str(dataset_path)])
    ingestion = DatasetIngestion(config)
    dataset = ingestion.load_dataset()
    
    # Compute statistics
    stats = dataset.get_statistics()
    
    # Check for lineage file
    lineage_path = dataset_path / "lineage.json"
    lineage_info = None
    if lineage_path.exists():
        with open(lineage_path) as f:
            lineage_info = json.load(f)
    
    # Check for manifest
    manifest_path = dataset_path / "manifest.json"
    manifest_info = None
    if manifest_path.exists():
        with open(manifest_path) as f:
            manifest_info = json.load(f)
    
    if args.format == "json":
        output = {
            "statistics": stats,
            "lineage": lineage_info,
            "manifest": manifest_info,
        }
        print(json.dumps(output, indent=2))
    else:
        print("\n" + "=" * 60)
        print("DATASET INFORMATION")
        print("=" * 60)
        
        if manifest_info:
            print(f"\nVersion: {manifest_info.get('dataset_version', 'unknown')}")
            print(f"Hash: {manifest_info.get('manifest_hash', 'unknown')}")
            if manifest_info.get("signature"):
                sig = manifest_info["signature"]
                print(f"Signed by: {sig.get('signer_id', 'unknown')}")
        
        print(f"\nTotal samples: {stats['total_samples']}")
        
        for split_name in ["train", "validation", "test"]:
            split_stats = stats[split_name]
            print(f"\n{split_name.upper()}:")
            print(f"  Samples: {split_stats['num_samples']}")
            print(f"  Genuine: {split_stats['genuine_count']}")
            print(f"  Fraud: {split_stats['fraud_count']}")
            print(f"  Mean score: {split_stats['label_mean']:.2f}")
            print(f"  Std score: {split_stats['label_std']:.2f}")
            
            if split_stats.get('doc_type_distribution'):
                print(f"  Doc types: {split_stats['doc_type_distribution']}")
        
        if lineage_info:
            print(f"\nLineage ID: {lineage_info.get('lineage_id', 'unknown')}")
            print(f"Build timestamp: {lineage_info.get('created_at_iso', 'unknown')}")
            
            if lineage_info.get('sources'):
                print(f"Sources: {len(lineage_info['sources'])}")
                for src in lineage_info['sources'][:3]:
                    print(f"  - {src.get('location', 'unknown')}")
            
            if lineage_info.get('transforms'):
                print(f"Transforms: {len(lineage_info['transforms'])}")
                for t in lineage_info['transforms'][:5]:
                    print(f"  - {t.get('description', 'unknown')}")
    
    return 0


def main() -> int:
    """Main entry point."""
    args = parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    if not args.command:
        print("No command specified. Use --help for usage information.")
        return 1
    
    # Dispatch command
    commands = {
        "build": cmd_build,
        "synthetic": cmd_synthetic,
        "validate": cmd_validate,
        "label": cmd_label,
        "info": cmd_info,
    }
    
    cmd_func = commands.get(args.command)
    if cmd_func is None:
        print(f"Unknown command: {args.command}")
        return 1
    
    try:
        return cmd_func(args)
    except KeyboardInterrupt:
        print("\nInterrupted by user")
        return 130
    except Exception as e:
        logger.exception(f"Command failed: {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
