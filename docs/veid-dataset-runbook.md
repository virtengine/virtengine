# VEID Dataset Build and Approval Runbook

This document describes the procedures for building, validating, and approving VEID training datasets.

## Overview

The VEID dataset pipeline provides a production-grade system for:

- Ingesting identity verification data from multiple sources
- Anonymizing PII before processing
- Applying labels (human review + heuristics)
- Creating deterministic train/val/test splits
- Generating signed manifests for audit trails
- Tracking complete dataset lineage

## Quick Start

### Generate Synthetic Dataset (CI/Development)

```bash
# Minimal synthetic dataset for CI tests
python -m ml.training.build_dataset synthetic \
    --output data/synthetic/ci \
    --profile ci_minimal

# Standard development dataset
python -m ml.training.build_dataset synthetic \
    --output data/synthetic/dev \
    --profile dev_medium \
    --seed 42
```

### Build Production Dataset

```bash
python -m ml.training.build_dataset build \
    --source /data/veid/raw \
    --output /data/veid/v1.0.0 \
    --version 1.0.0 \
    --sign \
    --signer-id production-builder
```

### Validate Existing Dataset

```bash
python -m ml.training.build_dataset validate \
    --dataset /data/veid/v1.0.0 \
    --report validation_report.json \
    --fail-on-error
```

## Dataset Build Workflow

### 1. Data Collection

Data can be ingested from multiple sources:

| Source Type | URI Format | Example |
|-------------|------------|---------|
| Local Files | `/path/to/data` | `/data/veid/batch_001` |
| AWS S3 | `s3://bucket/prefix` | `s3://veid-data/raw/2024` |
| Google Cloud Storage | `gs://bucket/prefix` | `gs://veid-data/raw/2024` |
| HTTP API | `https://api.example.com` | `https://api.veid.io/v1/data` |

Each source must contain samples in the expected format:

```
data/
├── manifest.json      # Optional manifest file
├── sample_001/
│   ├── metadata.json  # Sample metadata
│   ├── document.png   # Document image
│   └── selfie.png     # Selfie image
├── sample_002/
│   ├── metadata.json
│   ├── document.jpg
│   └── selfie.jpg
...
```

### 2. PII Anonymization

All PII is anonymized during processing:

- Sample IDs are hashed with SHA-256 + salt
- Document IDs are hashed
- Raw images are encrypted with X25519-XSalsa20-Poly1305
- Derived features are stored separately (non-PII)

The anonymization salt is generated per build and stored in the lineage record.

### 3. Labeling Pipeline

Labels are applied in two stages:

**Heuristic Auto-Labels:**
- Applied to all samples automatically
- Based on signal quality (face confidence, OCR, document quality)
- Default rules detect common fraud patterns

**Human Review Labels:**
- Import from CSV files
- Override heuristic labels
- Require annotator_id and review timestamp

```bash
# Export samples for human review
python -m ml.training.build_dataset label \
    --dataset /data/veid/raw \
    --export-for-review /reviews/pending.csv

# Apply reviewed labels
python -m ml.training.build_dataset label \
    --dataset /data/veid/raw \
    --labels /reviews/completed.csv \
    --output /data/veid/labeled
```

### 4. Deterministic Splitting

Datasets are split with guaranteed reproducibility:

- Fixed random seed (default: 42)
- Samples sorted by ID before splitting
- Stratified by document type and genuine/fraud status
- Split hashes computed for verification

Split configuration:
```yaml
train_ratio: 0.7
val_ratio: 0.15
test_ratio: 0.15
strategy: stratified
stratify_by:
  - doc_type
  - is_genuine
random_seed: 42
```

### 5. Validation

Datasets are validated for:

**Schema Compliance:**
- Required fields present
- Values within valid ranges
- Correct data types

**Label Quality:**
- Score-label consistency
- Outlier detection
- Distribution checks

**Data Quality:**
- Image presence
- Quality score thresholds
- OCR success rates

**Split Integrity:**
- No sample overlap between splits
- No duplicate sample IDs
- Minimum samples per split

### 6. Manifest Signing

Production datasets must be signed:

```bash
python -m ml.training.build_dataset build \
    --source /data/veid/raw \
    --output /data/veid/v1.0.0 \
    --sign \
    --signer-id production-builder
```

The manifest contains:
- Content hashes for all samples
- Build configuration hash
- Signature timestamp
- Signer identity

## Output Artifacts

A complete build produces:

```
output/
├── manifest.json          # Signed manifest with all content hashes
├── lineage.json           # Complete lineage record
├── validation_report.json # Validation results
├── dataset.json           # Dataset samples and metadata
└── raw/                   # Encrypted raw images (if stored)
    └── sample_xxx_document.enc
```

## Approval Process

### Pre-Approval Checklist

- [ ] Validation report shows no errors
- [ ] Manifest is signed by authorized signer
- [ ] Lineage record is complete
- [ ] Split hashes match expected values (if known)
- [ ] Sample counts meet minimum requirements
- [ ] Genuine/fraud ratios are within expected range

### Approval Steps

1. **Generate Dataset**
   ```bash
   python -m ml.training.build_dataset build \
       --source /data/veid/raw \
       --output /data/veid/pending/v1.0.0 \
       --version 1.0.0 \
       --sign
   ```

2. **Review Validation Report**
   ```bash
   python -m ml.training.build_dataset validate \
       --dataset /data/veid/pending/v1.0.0 \
       --report /reviews/validation.json
   ```

3. **Verify Manifest Signature**
   ```python
   from ml.training.dataset.manifest import DatasetManifest, ManifestVerifier
   
   manifest = DatasetManifest.load("manifest.json")
   verifier = ManifestVerifier(trusted_signers={"production-builder": key})
   result = verifier.verify(manifest)
   
   assert result["valid"]
   ```

4. **Record Approval**
   - Add approval signature to manifest
   - Move to production storage
   - Update dataset registry

### Post-Approval Verification

After moving to production:

```bash
# Verify hash matches
python -m ml.training.build_dataset validate \
    --dataset /data/veid/production/v1.0.0 \
    --expected-hash <original-hash> \
    --fail-on-error
```

## Synthetic Data for CI

CI pipelines should use synthetic datasets:

```bash
# In CI pipeline
python -m ml.training.build_dataset synthetic \
    --output /tmp/test_data \
    --profile ci_minimal \
    --seed 42

# Run training tests
python -m pytest ml/training/tests/ -v
```

Synthetic profiles:

| Profile | Samples | Images | Use Case |
|---------|---------|--------|----------|
| ci_minimal | 30 | No | Fast CI tests |
| ci_standard | 100 | Yes | Standard CI |
| dev_small | 500 | Yes | Local development |
| dev_medium | 2000 | Yes | Full development |
| dev_large | 10000 | Yes | Performance testing |
| benchmark | 5000 | Yes | Benchmarking |

## Troubleshooting

### Common Issues

**Validation Errors:**
```
ERROR: Required field 'trust_score' is missing
```
→ Ensure all samples have required fields in metadata.json

**Split Hash Mismatch:**
```
ERROR: Hash mismatch: expected abc123, got xyz789
```
→ Check that the same seed and split ratios are used

**Signature Verification Failed:**
```
ERROR: Unknown signer: build-server-1
```
→ Add the signer to trusted_signers in verifier config

### Debug Mode

Enable verbose logging:
```bash
python -m ml.training.build_dataset build \
    --verbose \
    --source /data/veid/raw \
    --output /data/veid/debug
```

## Security Considerations

1. **Encryption Keys:**
   - Master keys must be stored in secure key management
   - Key rotation every 90 days (configurable)
   - Never commit keys to source control

2. **PII Handling:**
   - Raw images are encrypted at rest
   - Anonymization salt is unique per build
   - Audit logs track all data access

3. **Signing Keys:**
   - Use hardware security modules (HSM) for production signing
   - Separate keys for development and production
   - Key access requires multi-party authorization

## API Reference

### Python API

```python
from ml.training.dataset import (
    # Ingestion
    DatasetIngestion,
    ConnectorRegistry,
    
    # Synthetic
    generate_synthetic_dataset,
    SyntheticConfig,
    
    # Labeling
    LabelingPipeline,
    HeuristicLabeler,
    
    # Splitting
    DeterministicSplitter,
    SplitConfig,
    
    # Validation
    DatasetValidator,
    validate_dataset,
    
    # Manifest
    ManifestBuilder,
    ManifestSigner,
    
    # Lineage
    LineageTracker,
)

# Example: Full pipeline
tracker = LineageTracker("veid_trust", "1.0.0")

# Load data
connector = ConnectorRegistry.from_uri("/data/raw")
samples = list(connector.iter_records())
tracker.add_source("/data/raw", record_count=len(samples))

# Label
labeler = HeuristicLabeler()
labels = labeler.label_batch(samples)

# Split
splitter = DeterministicSplitter(SplitConfig(random_seed=42))
result = splitter.split(samples)

# Validate
report = validate_dataset(result.dataset)

# Create manifest
builder = ManifestBuilder("veid_trust")
for s in result.dataset.train:
    builder.add_sample(s.sample_id, s.to_bytes(), "train")
manifest = builder.build("1.0.0")

# Sign
signer = ManifestSigner("builder-1")
signed_manifest = signer.sign(manifest)

# Save
signed_manifest.save("manifest.json")
tracker.finalize(sample_count=len(samples)).save("lineage.json")
```

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-01-30 | Initial release with full pipeline |

## Support

For questions or issues:
- Create an issue in the repository
- Contact the ML team on Slack #ml-veid
- Email: ml-support@virtengine.io
