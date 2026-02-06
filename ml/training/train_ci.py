#!/usr/bin/env python3
"""
CI Training Pipeline for VEID Trust Score Model.

Generates a deterministic synthetic SavedModel for CI/CD validation.
This script produces model artifacts without requiring real training data,
using synthetic correlated features to create a verifiable model.

Usage:
    python -m ml.training.train_ci
    python -m ml.training.train_ci --output models/trust_score/v1.0.0
    python -m ml.training.train_ci --samples 1000 --epochs 5
"""

import argparse
import hashlib
import json
import logging
import os
import random
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, Optional

import numpy as np

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

# Default feature dimension matching MODEL_CARD.md
DEFAULT_FEATURE_DIM = 768
DEFAULT_HIDDEN_LAYERS = [512, 256, 128, 64]
DEFAULT_SEED = 42
DEFAULT_EPOCHS = 10
DEFAULT_SAMPLES = 5000


def setup_determinism(seed: int = DEFAULT_SEED) -> None:
    """Configure environment for deterministic execution."""
    os.environ["PYTHONHASHSEED"] = str(seed)
    os.environ["TF_DETERMINISTIC_OPS"] = "1"
    os.environ["TF_CUDNN_DETERMINISTIC"] = "1"
    os.environ["CUDA_VISIBLE_DEVICES"] = ""
    os.environ["TF_NUM_INTEROP_THREADS"] = "1"
    os.environ["TF_NUM_INTRAOP_THREADS"] = "1"
    os.environ["TF_CPP_MIN_LOG_LEVEL"] = "2"
    os.environ["OMP_NUM_THREADS"] = "1"
    os.environ["MKL_NUM_THREADS"] = "1"
    os.environ["OPENBLAS_NUM_THREADS"] = "1"

    np.random.seed(seed)
    random.seed(seed)


def generate_synthetic_dataset(
    num_samples: int = DEFAULT_SAMPLES,
    feature_dim: int = DEFAULT_FEATURE_DIM,
    seed: int = DEFAULT_SEED,
):
    """Generate correlated synthetic dataset for training."""
    np.random.seed(seed)
    random.seed(seed)

    X = np.random.randn(num_samples, feature_dim).astype(np.float32)

    face_end = min(feature_dim, int(feature_dim * 0.66))
    doc_end = min(feature_dim, face_end + int(feature_dim * 0.10))
    ocr_end = min(feature_dim, doc_end + int(feature_dim * 0.10))

    face_quality = np.mean(X[:, :face_end], axis=1) * 0.5 + 0.5
    doc_quality = np.mean(X[:, face_end:doc_end], axis=1) * 0.3 + 0.7 if doc_end > face_end else np.full(num_samples, 0.7, dtype=np.float32)
    ocr_conf = np.mean(X[:, doc_end:ocr_end], axis=1) * 0.3 + 0.6 if ocr_end > doc_end else np.full(num_samples, 0.6, dtype=np.float32)
    metadata_score = np.mean(X[:, ocr_end:], axis=1) * 0.2 + 0.8

    base_trust = face_quality * 35 + doc_quality * 30 + ocr_conf * 20 + metadata_score * 15
    noise = np.random.randn(num_samples).astype(np.float32) * 5
    y = np.clip(base_trust + noise, 0, 100).astype(np.float32)

    indices = np.random.permutation(num_samples)
    X, y = X[indices], y[indices]

    n_train = int(num_samples * 0.8)
    n_val = int(num_samples * 0.1)

    return (
        X[:n_train], y[:n_train],
        X[n_train:n_train + n_val], y[n_train:n_train + n_val],
        X[n_train + n_val:], y[n_train + n_val:],
    )


def compute_model_hash(model_path: str) -> str:
    """Compute SHA-256 hash of SavedModel files."""
    hasher = hashlib.sha256()
    model_dir = Path(model_path)

    for filepath in sorted(model_dir.rglob("*")):
        if filepath.is_file() and filepath.name not in (
            "export_metadata.json", "MODEL_HASH.txt", "manifest.json",
        ):
            rel_path = filepath.relative_to(model_dir)
            hasher.update(str(rel_path).encode())
            with open(filepath, "rb") as f:
                for chunk in iter(lambda: f.read(8192), b""):
                    hasher.update(chunk)

    return hasher.hexdigest()


def train_ci_model(
    output_dir: str = "models/trust_score/v1.0.0",
    num_samples: int = DEFAULT_SAMPLES,
    epochs: int = DEFAULT_EPOCHS,
    seed: int = DEFAULT_SEED,
    version: str = "v1.0.0",
) -> Dict[str, Any]:
    """
    Train a deterministic model from synthetic data and export SavedModel.

    Returns dictionary with model path, hash, and metrics.
    """
    start_time = time.time()

    logger.info("=" * 60)
    logger.info("VEID CI Training Pipeline")
    logger.info("=" * 60)

    # Step 0: Determinism
    setup_determinism(seed)
    logger.info("Deterministic environment configured (seed=%d)", seed)

    # Step 1: Import TensorFlow
    try:
        import tensorflow as tf
        tf.random.set_seed(seed)
        try:
            tf.config.experimental.enable_op_determinism()
        except Exception:
            pass
        logger.info("TensorFlow %s loaded (CPU-only)", tf.__version__)
    except ImportError:
        logger.error("TensorFlow is required. Install tensorflow-cpu==2.15.0")
        return {"success": False, "error": "TensorFlow not installed"}

    # Step 2: Generate dataset
    logger.info("Generating synthetic dataset (%d samples, %d dims)", num_samples, DEFAULT_FEATURE_DIM)
    train_X, train_y, val_X, val_y, test_X, test_y = generate_synthetic_dataset(
        num_samples=num_samples, feature_dim=DEFAULT_FEATURE_DIM, seed=seed,
    )
    logger.info(
        "Dataset: %d train, %d val, %d test",
        len(train_y), len(val_y), len(test_y),
    )

    # Step 3: Build model (4-layer MLP per MODEL_CARD.md)
    logger.info("Building model: input=%d, layers=%s", DEFAULT_FEATURE_DIM, DEFAULT_HIDDEN_LAYERS)

    model = tf.keras.Sequential([
        tf.keras.layers.InputLayer(input_shape=(DEFAULT_FEATURE_DIM,)),
    ])
    for units in DEFAULT_HIDDEN_LAYERS:
        model.add(tf.keras.layers.Dense(units, activation="relu",
                                         kernel_initializer=tf.keras.initializers.GlorotUniform(seed=seed)))
        model.add(tf.keras.layers.BatchNormalization())
        model.add(tf.keras.layers.Dropout(0.2, seed=seed))
    model.add(tf.keras.layers.Dense(1, activation="sigmoid",
                                     kernel_initializer=tf.keras.initializers.GlorotUniform(seed=seed)))

    model.compile(
        optimizer=tf.keras.optimizers.Adam(learning_rate=0.001),
        loss="mse",
        metrics=["mae"],
    )

    # Step 4: Train
    logger.info("Training for %d epochs...", epochs)
    history = model.fit(
        train_X, train_y / 100.0,  # Normalize to 0-1 for sigmoid
        validation_data=(val_X, val_y / 100.0),
        epochs=epochs,
        batch_size=64,
        verbose=0,
        callbacks=[
            tf.keras.callbacks.EarlyStopping(
                monitor="val_loss", patience=3, restore_best_weights=True,
            ),
        ],
    )
    logger.info("Training complete (%d epochs)", len(history.history["loss"]))

    # Step 5: Evaluate
    test_pred = model.predict(test_X, verbose=0).flatten() * 100.0
    mae = float(np.mean(np.abs(test_pred - test_y)))
    rmse = float(np.sqrt(np.mean((test_pred - test_y) ** 2)))
    ss_res = np.sum((test_y - test_pred) ** 2)
    ss_tot = np.sum((test_y - np.mean(test_y)) ** 2)
    r2 = float(1 - ss_res / ss_tot) if ss_tot > 0 else 0.0
    acc_10 = float(np.mean(np.abs(test_pred - test_y) <= 10))

    logger.info("Evaluation: MAE=%.2f, RMSE=%.2f, R²=%.4f, Acc@10=%.1f%%",
                mae, rmse, r2, acc_10 * 100)

    # Step 6: Export SavedModel
    output_path = Path(output_dir)
    model_export_path = output_path / "model"
    model_export_path.mkdir(parents=True, exist_ok=True)

    tf.saved_model.save(model, str(model_export_path))
    logger.info("SavedModel exported to: %s", model_export_path)

    # Step 7: Compute hash
    model_hash = compute_model_hash(str(model_export_path))
    logger.info("Model hash: %s", model_hash)

    # Step 8: Write MODEL_HASH.txt
    hash_content = f"""# VirtEngine Trust Score Model Hash
# Generated: {datetime.utcnow().isoformat()}Z
# Version: {version}
# Pipeline: CI synthetic training
#
# Algorithm: SHA-256
# Scope: All model files (saved_model.pb, variables/*)

SHA256={model_hash}
VERSION={version}
TIMESTAMP={datetime.utcnow().isoformat()}Z
"""
    hash_path = output_path / "MODEL_HASH.txt"
    with open(hash_path, "w") as f:
        f.write(hash_content)

    # Step 9: Write export metadata
    metadata = {
        "version": version,
        "model_hash": model_hash,
        "feature_dim": DEFAULT_FEATURE_DIM,
        "hidden_layers": DEFAULT_HIDDEN_LAYERS,
        "training_samples": num_samples,
        "epochs_trained": len(history.history["loss"]),
        "seed": seed,
        "tensorflow_version": tf.__version__,
        "numpy_version": np.__version__,
        "metrics": {
            "mae": round(mae, 4),
            "rmse": round(rmse, 4),
            "r2": round(r2, 4),
            "accuracy_at_10": round(acc_10, 4),
        },
        "deterministic": True,
        "force_cpu": True,
        "exported_at": datetime.utcnow().isoformat() + "Z",
        "model_card": "models/trust_score/MODEL_CARD.md",
    }

    metadata_path = output_path / "export_metadata.json"
    with open(metadata_path, "w") as f:
        json.dump(metadata, f, indent=2)

    total_time = time.time() - start_time

    logger.info("")
    logger.info("=" * 60)
    logger.info("CI TRAINING COMPLETE")
    logger.info("=" * 60)
    logger.info("Version: %s", version)
    logger.info("Hash: %s", model_hash)
    logger.info("Path: %s", model_export_path)
    logger.info("Time: %.1fs", total_time)

    return {
        "success": True,
        "model_path": str(model_export_path),
        "model_hash": model_hash,
        "version": version,
        "metrics": metadata["metrics"],
        "metadata_path": str(metadata_path),
        "hash_path": str(hash_path),
        "training_time_seconds": total_time,
    }


def main():
    parser = argparse.ArgumentParser(description="VEID CI Training Pipeline")
    parser.add_argument("--output", default="models/trust_score/v1.0.0",
                        help="Output directory for model artifacts")
    parser.add_argument("--samples", type=int, default=DEFAULT_SAMPLES,
                        help="Number of synthetic samples")
    parser.add_argument("--epochs", type=int, default=DEFAULT_EPOCHS,
                        help="Training epochs")
    parser.add_argument("--seed", type=int, default=DEFAULT_SEED,
                        help="Random seed for determinism")
    parser.add_argument("--version", default="v1.0.0",
                        help="Model version string")
    args = parser.parse_args()

    result = train_ci_model(
        output_dir=args.output,
        num_samples=args.samples,
        epochs=args.epochs,
        seed=args.seed,
        version=args.version,
    )

    if not result["success"]:
        logger.error("Pipeline failed: %s", result.get("error", "unknown"))
        sys.exit(1)

    sys.exit(0)


if __name__ == "__main__":
    main()
