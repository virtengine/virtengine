# Feature Extraction Pipeline

## Overview

The feature extraction pipeline transforms decrypted identity scopes into ML feature vectors for the VEID identity scoring system. This document describes the architecture, components, and usage of the feature extraction subsystem.

## Architecture

```
┌─────────────────────┐
│  Decrypted Scopes   │
│  (Selfie, ID Doc,   │
│   Video, etc.)      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│    Media Parser     │
│  - Format detection │
│  - Image/Video/JSON │
│  - Validation       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Feature Extractor  │
│  - Face embedding   │
│  - Liveness signals │
│  - OCR features     │
│  - Quality metrics  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Quality Gates     │
│  - Blur detection   │
│  - Brightness check │
│  - OCR confidence   │
│  - Face confidence  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   ML Feature Vector │
│   (768 dimensions)  │
└─────────────────────┘
```

## Components

### Media Parser (`x/veid/keeper/media_parser.go`)

Parses decrypted scope payloads into typed media content.

**Supported Formats:**
- **Images**: JPEG, PNG, WebP
- **Videos**: MP4, WebM, AVI
- **JSON**: SSO metadata, email/SMS proofs

**Usage:**
```go
parser := NewMediaParser()
payload, err := parser.Parse(decryptedScope)
if err != nil {
    return err
}

switch payload.MediaType {
case MediaTypeImage:
    // Process image
    width := payload.ImageData.Width
    height := payload.ImageData.Height
case MediaTypeVideo:
    // Process video
    frames := payload.VideoData.EstimatedFrameCount
case MediaTypeJSON:
    // Process structured data
    data := payload.JSONData
}
```

### Feature Extraction Pipeline (`x/veid/keeper/feature_extraction.go`)

Orchestrates extraction of all ML features from parsed media.

**Extracted Features:**

| Feature Group | Dimensions | Description |
|---------------|------------|-------------|
| Face Embedding | 512 | L2-normalized face representation |
| Doc Quality | 5 | Sharpness, brightness, contrast, noise, blur |
| OCR Features | 10 | Confidence + validation for 5 fields |
| Metadata | 16 | Scope counts, types, timestamps |
| Padding | 225 | Reserved for future use |
| **Total** | **768** | Combined feature vector |

**Usage:**
```go
config := DefaultFeatureExtractionConfig()
pipeline := NewFeatureExtractionPipeline(config)

features, err := pipeline.ExtractFeatures(
    decryptedScopes,
    accountAddress,
    blockHeight,
    blockTime,
)
if err != nil {
    return err
}

// Convert to ML vector
vector, err := pipeline.NormalizeToMLVector(features)
```

### Quality Gates (`x/veid/keeper/quality_gates.go`)

Enforces minimum quality thresholds for feature extraction.

**Default Thresholds:**

| Gate | Threshold | Description |
|------|-----------|-------------|
| Face Confidence | ≥ 0.7 | Face detection confidence |
| Face Sharpness | ≥ 0.4 | Face image sharpness |
| Face Brightness | 0.2-0.9 | Face region brightness |
| Face Size | ≥ 0.1 | Relative face size in image |
| Doc Sharpness | ≥ 0.4 | Document image sharpness |
| Doc Brightness | 0.2-0.9 | Document image brightness |
| Doc Contrast | ≥ 0.3 | Document image contrast |
| Doc Noise | ≤ 0.4 | Document noise level |
| Doc Blur | ≤ 0.5 | Document blur score |
| OCR Confidence | ≥ 0.6 | Average OCR confidence |
| OCR Validation | ≥ 40% | Validated field ratio |
| Liveness Score | ≥ 0.5 | Liveness detection score |

**Quality Gate Results:**
```go
gates := NewQualityGates(DefaultQualityGatesConfig())
results := gates.Check(features)

if !AllQualityGatesPassed(results) {
    failed := GetFailedGates(results)
    codes := GetFailureReasonCodes(results)
    // Handle quality failures
}
```

## Feature Schema

### Version: 1.0.0

The feature schema is defined in `x/veid/types/ml_feature_schema.go`.

**Feature Vector Layout:**

| Index Range | Group | Description |
|-------------|-------|-------------|
| 0-511 | Face Embedding | 512-dim L2-normalized embedding |
| 512-516 | Doc Quality | 5 quality metrics |
| 517-526 | OCR Features | 5 fields × 2 (confidence + valid) |
| 527-542 | Metadata | Scope info, timestamps |
| 543-767 | Padding | Reserved |

### OCR Fields

The following OCR fields are extracted from ID documents:

1. `name` - Full name
2. `date_of_birth` - Date of birth
3. `document_number` - ID/passport number
4. `expiry_date` - Document expiration
5. `nationality` - Country/nationality

## Determinism

All feature extraction is deterministic for blockchain consensus:

1. **Random Seed**: Fixed seed (42) for any pseudo-random operations
2. **CPU-Only**: GPU operations disabled to avoid floating-point variance
3. **Pinned Versions**: All ML libraries use pinned versions (see `ml/requirements-deterministic.txt`)
4. **Hash Verification**: Feature hashes included in verification records

**Verifying Determinism:**
```go
// Extract features twice with same inputs
features1, _ := pipeline.ExtractFeatures(scopes, addr, height, time)
features2, _ := pipeline.ExtractFeatures(scopes, addr, height, time)

// Embeddings must be identical
for i := range features1.FaceEmbedding {
    assert.Equal(t, features1.FaceEmbedding[i], features2.FaceEmbedding[i])
}
```

## Audit Metadata

Feature extraction produces audit metadata for on-chain storage:

```go
type FeatureExtractionMetadata struct {
    SchemaVersion        string
    ExtractionTimestamp  time.Time
    AccountAddress       string
    BlockHeight          int64
    ProcessingDurationMs int64
    ModelsUsed           []ModelUsageInfo
    FeatureHash          string
    QualityGatesPassed   bool
    Errors               []ExtractionError
}
```

## Integration with Scoring

The extracted features integrate with the TensorFlow-based scoring pipeline:

```go
// In TensorFlowScorerAdapter.Score()
features, err := pipeline.ExtractFeatures(input.DecryptedScopes, ...)

// Check quality gates
if failedGates := GetFailedGates(features.QualityGateResults); len(failedGates) > 0 {
    // Apply score penalties
}

// Convert to inference inputs
inferInputs := pipeline.ToScoreInputs(features, ...)

// Run TensorFlow inference
result, err := scorer.ComputeScore(inferInputs)
```

## Error Handling

### Quality Gate Failures

When quality gates fail, the system returns specific reason codes:

| Reason Code | Description |
|-------------|-------------|
| `FACE_MISMATCH` | Face detection/confidence issues |
| `LOW_DOC_QUALITY` | Document quality below threshold |
| `LOW_OCR_CONFIDENCE` | OCR extraction confidence low |
| `LIVENESS_CHECK_FAILED` | Liveness detection failed |

### Extraction Errors

Non-fatal errors are recorded in metadata:

```go
if features.Metadata.HasErrors() {
    for _, err := range features.Metadata.Errors {
        log.Printf("Scope %s: %s", err.ScopeID, err.ErrorMessage)
    }
}
```

## Configuration

### Feature Extraction Config

```go
config := FeatureExtractionConfig{
    FaceEmbeddingDim:     512,
    EnableLivenessCheck:  true,
    MinFaceConfidence:    0.7,
    MinOCRConfidence:     0.6,
    UseDeterministicMode: true,
    RandomSeed:           42,
}
```

### Quality Gates Config

```go
config := QualityGatesConfig{
    MinFaceConfidence:    0.7,
    MinFaceSharpness:     0.4,
    MinDocSharpness:      0.4,
    MinOCRConfidence:     0.6,
    MinLivenessScore:     0.5,
    EnableFaceGates:      true,
    EnableDocGates:       true,
    EnableOCRGates:       true,
    EnableLivenessGates:  true,
}
```

## Testing

Run feature extraction tests:

```bash
go test -v ./x/veid/keeper/... -run TestFeature
go test -v ./x/veid/keeper/... -run TestQualityGates
```

Golden test fixtures are in `testutil/fixtures/`:
- `golden_face_embedding.json` - Expected face embedding properties
- `golden_ocr_features.json` - Expected OCR feature properties

## Future Work

- [ ] Integration with production Python ML pipelines via gRPC
- [ ] Hardware Security Module (HSM) support for model weights
- [ ] Model versioning and hot-swap support
- [ ] Enhanced liveness detection (active challenges)
- [ ] Multi-language OCR support
