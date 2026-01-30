# ML Training Pipeline

TensorFlow-based training pipeline for the VEID Trust Score Model used in
VirtEngine identity verification.

## Overview

This module provides a complete ML training pipeline for building trust score
models that evaluate identity verification submissions. The model takes face
embeddings, document OCR signals, and capture metadata as inputs and outputs
a trust score (0-100).

## Architecture

```
ml/training/
├── __init__.py           # Module exports
├── config.py             # Configuration classes
├── train.py              # Main training script
├── requirements.txt      # Python dependencies
├── README.md             # This file
├── build_dataset.py      # CLI for dataset build pipeline
├── dataset/              # Data loading and preprocessing
│   ├── __init__.py
│   ├── ingestion.py      # Dataset loading from various formats
│   ├── preprocessing.py  # Image preprocessing pipeline
│   ├── augmentation.py   # Data augmentation
│   ├── anonymization.py  # PII anonymization
│   ├── connectors.py     # Data source connectors (local, S3, GCS, API)
│   ├── synthetic.py      # Synthetic data generator for CI testing
│   ├── storage.py        # PII-safe encrypted storage
│   ├── labeling.py       # Labeling pipeline (heuristic + human review)
│   ├── manifest.py       # Signed dataset manifests
│   ├── splits.py         # Deterministic train/val/test splitting
│   ├── lineage.py        # Dataset lineage tracking
│   └── validation.py     # Schema and label validation
├── features/             # Feature extraction
│   ├── __init__.py
│   ├── face_features.py      # Face embedding extraction
│   ├── doc_features.py       # Document quality features
│   ├── ocr_features.py       # OCR-based features
│   ├── metadata_features.py  # Device/session metadata
│   └── feature_combiner.py   # Unified feature extractor
├── model/                # Model training and export
│   ├── __init__.py
│   ├── architecture.py   # TrustScoreModel (MLP)
│   ├── training.py       # Training loop with callbacks
│   ├── evaluation.py     # Metrics and evaluation
│   └── export.py         # TensorFlow SavedModel export
└── tests/                # Unit tests
    ├── __init__.py
    ├── conftest.py       # Pytest fixtures
    ├── test_dataset.py
    ├── test_features.py
    ├── test_training.py
    └── test_pipeline.py  # Tests for dataset pipeline components
```

## Installation

```bash
cd ml/training
pip install -r requirements.txt
```

## Usage

### Training from Command Line

```bash
# Basic training
python -m ml.training.train --data-path /path/to/data --output-dir output

# With configuration file
python -m ml.training.train --config config.yaml

# Override parameters
python -m ml.training.train \
    --data-path /data/veid \
    --output-dir ./models \
    --epochs 100 \
    --batch-size 32 \
    --learning-rate 0.001
```

### Training from Python

```python
from ml.training.config import TrainingConfig
from ml.training.dataset.ingestion import DatasetIngestion
from ml.training.features.feature_combiner import FeatureExtractor
from ml.training.model.training import ModelTrainer
from ml.training.model.evaluation import ModelEvaluator
from ml.training.model.export import ModelExporter

# Load configuration
config = TrainingConfig.from_yaml("config.yaml")

# Ingest dataset
ingestion = DatasetIngestion(config.dataset)
dataset = ingestion.load_dataset()

# Extract features
extractor = FeatureExtractor(config.features)
features = extractor.extract_dataset(dataset)

# Train model
trainer = ModelTrainer(config)
result = trainer.train(features)

# Evaluate
evaluator = ModelEvaluator(config.model)
metrics = evaluator.evaluate(result.model, features.test)

# Export for Go inference
exporter = ModelExporter(config.export)
export_result = exporter.export_savedmodel(result.model, "models/", "v1.0.0")
```

## Model Architecture

The trust score model is a Multi-Layer Perceptron (MLP) with:

- **Input**: 768-dimensional feature vector
  - Face embeddings (512 dims × 2)
  - Document quality features (17 dims)
  - OCR features (~37 dims)
  - Metadata features (24 dims)
  - Additional signals

- **Hidden Layers**: [512, 256, 128, 64] with:
  - Batch normalization
  - Dropout (0.3)
  - ReLU activation
  - L2 regularization

- **Output**: Single trust score (0-100)
  - Sigmoid activation × 100

## Feature Pipeline

### Face Features
- Document face embedding (512-dim)
- Selfie face embedding (512-dim)
- Face similarity score
- Liveness detection signals
- Quality metrics

### Document Features
- Sharpness score
- Brightness/contrast
- Noise level
- Edge density
- Corner/region detection
- Overall quality score

### OCR Features
- Field extraction confidence (name, DOB, doc number, etc.)
- MRZ validation status
- Format validation scores
- Character confidence

### Metadata Features
- Device type (mobile/tablet/desktop)
- OS encoding
- Camera facing (front/back)
- Light level normalization
- GPS availability
- Session duration

## Export Format

The model is exported in TensorFlow SavedModel format for Go inference:

```
exported_models/
└── v1.0.0/
    ├── saved_model.pb
    ├── variables/
    │   ├── variables.data-00000-of-00001
    │   └── variables.index
    ├── metadata.json
    └── model_hash.txt
```

### Go Integration

```go
package main

import (
    tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

func LoadTrustScoreModel(modelPath string) (*tf.SavedModel, error) {
    return tf.LoadSavedModel(modelPath, []string{"serve"}, nil)
}

func Predict(model *tf.SavedModel, features []float32) (float32, error) {
    inputTensor, _ := tf.NewTensor([][]float32{features})
    
    result, err := model.Session.Run(
        map[tf.Output]*tf.Tensor{
            model.Graph.Operation("features").Output(0): inputTensor,
        },
        []tf.Output{
            model.Graph.Operation("trust_score").Output(0),
        },
        nil,
    )
    
    return result[0].Value().([][]float32)[0][0], err
}
```

## Evaluation Metrics

The pipeline computes comprehensive metrics:

- **MAE**: Mean Absolute Error
- **MSE**: Mean Squared Error
- **RMSE**: Root Mean Squared Error
- **R²**: Coefficient of determination
- **Accuracy@k**: Predictions within k points of ground truth
  - @5, @10, @15, @20 thresholds
- **Percentile errors**: P50, P90, P95, P99
- **Confusion matrix**: For score bin classification

## Configuration

```yaml
# config.yaml
experiment_name: "veid_trust_v1"
random_seed: 42

dataset:
  data_paths:
    - "/data/veid/train"
  document_types:
    - id_card
    - passport
    - drivers_license
  train_split: 0.8
  val_split: 0.1
  test_split: 0.1
  anonymize: true
  min_face_confidence: 0.7
  min_doc_quality: 0.5

preprocessing:
  image_size: [224, 224]
  normalize_images: true
  correct_orientation: true
  correct_perspective: true

augmentation:
  enabled: true
  num_augmented_copies: 2
  brightness_range: [0.8, 1.2]
  contrast_range: [0.8, 1.2]
  rotation_range: [-5, 5]

features:
  face_embedding_dim: 512
  combined_feature_dim: 768
  normalize_features: true

model:
  input_dim: 768
  hidden_layers: [512, 256, 128, 64]
  dropout_rate: 0.3
  activation: "relu"
  output_activation: "sigmoid"
  learning_rate: 0.001
  batch_size: 32
  epochs: 100
  early_stopping: true
  patience: 10

export:
  export_dir: "./exported_models"
  include_version: true
  compute_hash: true
```

## Testing

```bash
# Run all tests
pytest ml/training/tests/

# Run with coverage
pytest ml/training/tests/ --cov=ml.training --cov-report=html

# Run specific test file
pytest ml/training/tests/test_training.py -v
```

## Dataset Pipeline

The module includes a production-grade dataset pipeline for building training datasets.
See [docs/veid-dataset-runbook.md](../../docs/veid-dataset-runbook.md) for full documentation.

### Quick Start

```bash
# Generate synthetic dataset for CI tests
python -m ml.training.build_dataset synthetic \
    --output data/synthetic/ci \
    --profile ci_minimal

# Build production dataset
python -m ml.training.build_dataset build \
    --source /data/veid/raw \
    --output /data/veid/v1.0.0 \
    --version 1.0.0 \
    --sign

# Validate dataset
python -m ml.training.build_dataset validate \
    --dataset /data/veid/v1.0.0 \
    --report validation_report.json
```

### Pipeline Components

- **Connectors**: Ingest data from local files, S3, GCS, or HTTP APIs
- **Synthetic Generator**: Generate synthetic data for CI/dev testing
- **Labeling**: Apply heuristic auto-labels and import human review labels
- **Manifests**: Create signed manifests with content hashes
- **Splits**: Deterministic train/val/test splitting with verification
- **Validation**: Schema validation, label anomaly detection, quality checks
- **Lineage**: Track data sources, transforms, and build information

### Synthetic Data Profiles

| Profile | Samples | Use Case |
|---------|---------|----------|
| ci_minimal | 30 | Fast CI tests |
| ci_standard | 100 | Standard CI |
| dev_small | 500 | Local development |
| dev_medium | 2000 | Full development |
| benchmark | 5000 | Benchmarking |

## License

See [LICENSE](../../LICENSE) in the project root.
