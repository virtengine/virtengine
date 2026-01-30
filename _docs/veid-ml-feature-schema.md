# VEID ML Feature Schema Specification

**Version:** 1.0.0  
**Status:** Draft  
**Last Updated:** 2026-01-30  
**Authors:** VirtEngine Team

## Table of Contents

1. [Overview](#overview)
2. [Scope Types](#scope-types)
3. [Feature Groups](#feature-groups)
4. [Feature Definitions](#feature-definitions)
5. [Normalization Rules](#normalization-rules)
6. [Missing Data Handling](#missing-data-handling)
7. [Schema Versioning](#schema-versioning)
8. [Training and Inference Mapping](#training-and-inference-mapping)
9. [Consent and Retention Constraints](#consent-and-retention-constraints)
10. [Migration Policy](#migration-policy)
11. [Examples](#examples)

---

## Overview

This document defines the canonical ML feature schema and data contract for the VirtEngine Identity (VEID) verification system. The schema ensures deterministic feature extraction across training and inference pipelines, enabling consensus-safe ML scoring on the blockchain.

### Design Principles

1. **Determinism**: All feature extraction and normalization must be deterministic across validators
2. **Privacy-by-Design**: Features are derived from encrypted data; raw PII is never stored on-chain
3. **Backward Compatibility**: Schema changes must not break existing verification records
4. **Consent-Aware**: Feature groups are mapped to consent requirements

### Feature Vector Dimensions

| Component | Dimensions | Description |
|-----------|------------|-------------|
| Face Embedding | 512 | Normalized face embedding from selfie/document |
| Document Quality | 5 | Image quality metrics |
| OCR Features | 10 | Field confidence and validation (5 fields × 2) |
| Metadata Features | 16 | Scope types, counts, and contextual data |
| Padding/Reserved | 225 | Reserved for future expansion |
| **Total** | **768** | Combined feature vector dimension |

---

## Scope Types

The VEID system supports the following scope types, each contributing specific features:

| Scope Type | Code | Weight | Required Features | Consent Category |
|------------|------|--------|-------------------|------------------|
| ID Document | `id_document` | 30 | Document quality, OCR, face extraction | biometric_pii |
| Selfie | `selfie` | 20 | Face embedding, face quality | biometric |
| Face Video | `face_video` | 25 | Liveness features, face embedding | biometric |
| Biometric | `biometric` | 20 | Biometric hash, template features | biometric |
| SSO Metadata | `sso_metadata` | 5 | Provider metadata, claims | identity_attestation |
| Email Proof | `email_proof` | 10 | Verification status | contact_verification |
| SMS Proof | `sms_proof` | 10 | Verification status | contact_verification |
| Domain Verify | `domain_verify` | 15 | DNS record validation | domain_ownership |
| AD SSO | `ad_sso` | 12 | Enterprise claims, tenant info | enterprise_identity |

### Scope Type Definitions

```go
// From x/veid/types/scope.go
const (
    ScopeTypeIDDocument   ScopeType = "id_document"
    ScopeTypeSelfie       ScopeType = "selfie"
    ScopeTypeFaceVideo    ScopeType = "face_video"
    ScopeTypeBiometric    ScopeType = "biometric"
    ScopeTypeSSOMetadata  ScopeType = "sso_metadata"
    ScopeTypeEmailProof   ScopeType = "email_proof"
    ScopeTypeSMSProof     ScopeType = "sms_proof"
    ScopeTypeDomainVerify ScopeType = "domain_verify"
    ScopeTypeADSSO        ScopeType = "ad_sso"
)
```

---

## Feature Groups

### Group 1: Face Features (512 + 7 = 519 dimensions in training)

Features extracted from facial verification pipeline.

| Feature Name | Type | Dimensions | Unit | Range | Description |
|--------------|------|------------|------|-------|-------------|
| `face_embedding` | float32[] | 512 | - | [-1.0, 1.0] | L2-normalized face embedding vector |
| `document_face_embedding` | float32[] | 512 | - | [-1.0, 1.0] | Face embedding from ID document |
| `selfie_face_embedding` | float32[] | 512 | - | [-1.0, 1.0] | Face embedding from selfie |
| `face_similarity` | float32 | 1 | cosine | [0.0, 1.0] | Cosine similarity between document and selfie |
| `document_face_confidence` | float32 | 1 | probability | [0.0, 1.0] | Detection confidence for document face |
| `selfie_face_confidence` | float32 | 1 | probability | [0.0, 1.0] | Detection confidence for selfie face |
| `document_face_detected` | bool | 1 | binary | {0, 1} | Whether face was detected in document |
| `selfie_face_detected` | bool | 1 | binary | {0, 1} | Whether face was detected in selfie |
| `document_face_quality` | float32 | 1 | score | [0.0, 1.0] | Quality score for document face |
| `selfie_face_quality` | float32 | 1 | score | [0.0, 1.0] | Quality score for selfie face |

**Inference Mapping**: Combined into 512-dimensional embedding for model input.

### Group 2: Document Quality Features (17 dimensions in training, 5 in inference)

Features extracted from document image quality analysis.

| Feature Name | Type | Unit | Range | Description |
|--------------|------|------|-------|-------------|
| `sharpness_score` | float32 | score | [0.0, 1.0] | Laplacian variance-based sharpness |
| `brightness_score` | float32 | score | [0.0, 1.0] | Deviation from ideal brightness (0.5) |
| `contrast_score` | float32 | score | [0.0, 1.0] | Standard deviation-based contrast |
| `noise_level` | float32 | ratio | [0.0, 1.0] | Estimated noise level (lower is better) |
| `blur_score` | float32 | score | [0.0, 1.0] | Blur detection score (lower is better) |
| `saturation_score` | float32 | score | [0.0, 1.0] | Average color saturation |
| `color_uniformity` | float32 | score | [0.0, 1.0] | Inverse of color variance |
| `edge_density` | float32 | ratio | [0.0, 1.0] | Ratio of edge pixels |
| `corner_count` | int | count | [0, ∞) | Number of detected corners |
| `text_region_ratio` | float32 | ratio | [0.0, 1.0] | Estimated text region coverage |
| `has_photo_region` | bool | binary | {0, 1} | Photo region detected |
| `has_mrz_region` | bool | binary | {0, 1} | MRZ region detected |
| `has_barcode` | bool | binary | {0, 1} | Barcode detected |
| `document_bounds_confidence` | float32 | probability | [0.0, 1.0] | Document boundary detection confidence |
| `orientation_corrected` | bool | binary | {0, 1} | Whether orientation was corrected |
| `perspective_corrected` | bool | binary | {0, 1} | Whether perspective was corrected |
| `overall_quality_score` | float32 | score | [0.0, 1.0] | Composite quality score |

**Inference Mapping**: Condensed to 5 core metrics: sharpness, brightness, contrast, noise_level, blur_score.

### Group 3: OCR Features (5 fields × 5 = 25 dimensions in training, 10 in inference)

Features extracted from OCR field extraction.

#### Per-Field Features

For each OCR field (`name`, `date_of_birth`, `document_number`, `expiry_date`, `nationality`):

| Feature Name | Type | Unit | Range | Description |
|--------------|------|------|-------|-------------|
| `{field}_extracted` | bool | binary | {0, 1} | Whether field was extracted |
| `{field}_confidence` | float32 | probability | [0.0, 1.0] | OCR confidence for field |
| `{field}_validated` | bool | binary | {0, 1} | Whether field passed validation |
| `{field}_validation_score` | float32 | score | [0.0, 1.0] | Field validation score |
| `{field}_char_count` | int | count | [0, 50] | Character count (normalized) |

#### Aggregate OCR Features

| Feature Name | Type | Unit | Range | Description |
|--------------|------|------|-------|-------------|
| `overall_ocr_confidence` | float32 | probability | [0.0, 1.0] | Mean confidence of extracted fields |
| `fields_extracted_ratio` | float32 | ratio | [0.0, 1.0] | Ratio of successfully extracted fields |
| `fields_validated_ratio` | float32 | ratio | [0.0, 1.0] | Ratio of validated fields |
| `average_character_confidence` | float32 | probability | [0.0, 1.0] | Mean character-level confidence |
| `text_density` | float32 | ratio | [0.0, 1.0] | Text coverage estimation |
| `ocr_success` | bool | binary | {0, 1} | Overall OCR success flag |
| `ocr_timeout` | bool | binary | {0, 1} | OCR timeout flag (inverted) |
| `has_name` | bool | binary | {0, 1} | Name field extracted |
| `has_dob` | bool | binary | {0, 1} | Date of birth extracted |
| `has_doc_number` | bool | binary | {0, 1} | Document number extracted |
| `has_expiry` | bool | binary | {0, 1} | Expiry date extracted |
| `has_nationality` | bool | binary | {0, 1} | Nationality extracted |

**Inference Mapping**: Condensed to 10 dimensions (5 fields × 2: confidence + validation).

### Group 4: Metadata Features (16 dimensions)

Contextual features from verification request.

| Feature Name | Type | Unit | Range | Description |
|--------------|------|------|-------|-------------|
| `scope_count` | float32 | count | [0.0, 1.0] | Normalized scope count (÷10) |
| `has_id_document` | bool | binary | {0, 1} | ID document scope present |
| `has_selfie` | bool | binary | {0, 1} | Selfie scope present |
| `has_face_video` | bool | binary | {0, 1} | Face video scope present |
| `has_biometric` | bool | binary | {0, 1} | Biometric scope present |
| `has_sso_metadata` | bool | binary | {0, 1} | SSO metadata present |
| `has_email_proof` | bool | binary | {0, 1} | Email proof present |
| `has_sms_proof` | bool | binary | {0, 1} | SMS proof present |
| `has_domain_verify` | bool | binary | {0, 1} | Domain verify present |
| `face_confidence` | float32 | probability | [0.0, 1.0] | Best face confidence |
| `block_height_normalized` | float32 | ratio | [0.0, 1.0] | Block height mod 1M / 1M |
| `reserved_11` | float32 | - | - | Reserved for future use |
| `reserved_12` | float32 | - | - | Reserved for future use |
| `reserved_13` | float32 | - | - | Reserved for future use |
| `reserved_14` | float32 | - | - | Reserved for future use |
| `reserved_15` | float32 | - | - | Reserved for future use |

---

## Normalization Rules

All normalization operations must be deterministic and reproducible across validators.

### Face Embedding Normalization

```python
# L2 normalization to unit length
def normalize_embedding(embedding: np.ndarray) -> np.ndarray:
    norm = np.linalg.norm(embedding)
    if norm > 1e-10:
        return embedding / norm
    return np.zeros_like(embedding)
```

```go
// Go implementation in pkg/inference/feature_extractor.go
func (fe *FeatureExtractor) normalizeEmbedding(embedding []float32) {
    var sumSquares float64
    for _, v := range embedding {
        sumSquares += float64(v) * float64(v)
    }
    norm := math.Sqrt(sumSquares)
    if norm > 1e-10 {
        for i := range embedding {
            embedding[i] = float32(float64(embedding[i]) / norm)
        }
    }
}
```

### Score Normalization

| Feature Type | Normalization | Formula |
|--------------|---------------|---------|
| Probability scores | Clamp | `clamp(value, 0.0, 1.0)` |
| Count features | Scale | `min(1.0, count / max_count)` |
| Quality scores | Identity | Already in [0, 1] range |
| Boolean features | Binary | `1.0 if true else 0.0` |

### Inverted Features

Some features are inverted so that higher values are always better:

| Feature | Inversion Formula |
|---------|-------------------|
| `noise_level` | `1.0 - noise_level` |
| `blur_score` | `1.0 - blur_score` (in inference) |
| `ocr_timeout` | `1.0 if not timeout else 0.0` |

### Z-Score Normalization (Optional)

For training pipelines with feature scaling enabled:

```python
def zscore_normalize(features: np.ndarray, mean: np.ndarray, std: np.ndarray) -> np.ndarray:
    # Avoid division by zero
    std_safe = np.where(std > 1e-8, std, 1.0)
    return (features - mean) / std_safe
```

**Note**: Z-score normalization parameters (mean, std) must be computed from training data and frozen during inference.

---

## Missing Data Handling

### Default Values by Feature Type

| Feature Type | Default Value | Rationale |
|--------------|---------------|-----------|
| Face embedding | zeros(512) | No face detected |
| Confidence scores | 0.0 | No detection = zero confidence |
| Quality scores | 0.5 | Neutral quality assumption |
| Boolean indicators | 0.0 | Feature absent |
| Count features | 0.0 | No items counted |

### Missing Scope Handling

When a scope type is missing from the verification request:

1. All features from that scope receive default values
2. The corresponding `has_{scope_type}` metadata feature is set to 0
3. `scope_count` is decremented

### Partial Feature Handling

When a scope is present but feature extraction fails:

| Scenario | Handling |
|----------|----------|
| Face not detected | `face_confidence = 0.0`, embedding = zeros |
| OCR timeout | `ocr_timeout = 1.0`, all field features = 0.0 |
| Document quality failure | All quality features = 0.5 |
| Validation failure | `*_validated = 0.0`, `*_validation_score = 0.0` |

---

## Schema Versioning

### Version Format

Schema versions follow semantic versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes to feature dimensions or semantics
- **MINOR**: New features added (backward compatible)
- **PATCH**: Bug fixes, documentation updates

### Current Version

```
Schema Version: 1.0.0
Feature Vector Version: 768
Model Compatibility: v1.x
```

### Version Constants

```go
// x/veid/types/scope.go
const ScopeSchemaVersion uint32 = 1

// pkg/inference/types.go
const (
    FaceEmbeddingDim = 512
    DocQualityDim    = 5
    OCRFieldsDim     = 10
    MetadataDim      = 16
    TotalFeatureDim  = 768
    PaddingDim       = 225
)
```

### Backward Compatibility Rules

1. **Dimension Stability**: `TotalFeatureDim` (768) must remain constant within a major version
2. **Feature Semantics**: Existing feature meanings cannot change within a major version
3. **Default Values**: Default values for new features must ensure backward compatibility
4. **Padding Usage**: New features should consume padding dimensions before requiring a major version bump

### Version Migration

When upgrading schema versions:

1. Old verification records retain their original schema version
2. New verifications use the current schema version
3. Inference service must support all schema versions within the same major version

---

## Training and Inference Mapping

### Training Pipeline Input (Python)

```python
# ml/training/features/feature_combiner.py

@dataclass
class TrainingFeatures:
    # Face features (1031 dims in training, 512 combined for inference)
    face_features: FaceFeatures      # 512*2 + 7 = 1031
    
    # Document features (17 dims in training, 5 for inference)
    doc_features: DocumentFeatures   # 17
    
    # OCR features (25 + 12 = 37 dims in training, 10 for inference)
    ocr_features: OCRFeatures        # 37
    
    # Metadata features (16 dims)
    metadata_features: np.ndarray    # 16

def combine_for_inference(features: TrainingFeatures) -> np.ndarray:
    """Combine training features into 768-dim inference vector."""
    combined = np.zeros(768, dtype=np.float32)
    
    # Face embedding (512)
    combined[0:512] = features.face_features.combined_embedding
    
    # Document quality (5)
    combined[512:517] = features.doc_features.to_inference_vector()
    
    # OCR features (10)
    combined[517:527] = features.ocr_features.to_inference_vector()
    
    # Metadata (16)
    combined[527:543] = features.metadata_features
    
    # Padding (225) - zeros
    return combined
```

### Inference Pipeline Input (Go)

```go
// pkg/inference/feature_extractor.go

type ScoreInputs struct {
    FaceEmbedding      []float32            // 512 dims
    FaceConfidence     float32
    DocQualityScore    float32
    DocQualityFeatures DocQualityFeatures   // 5 dims
    OCRConfidences     map[string]float32   // 5 fields
    OCRFieldValidation map[string]bool      // 5 fields
    ScopeTypes         []string
    ScopeCount         int
    Metadata           InferenceMetadata
}

// Extracted feature vector layout:
// [0:512]   - Face embedding
// [512:517] - Document quality (sharpness, brightness, contrast, 1-noise, 1-blur)
// [517:527] - OCR features (5 fields × 2)
// [527:543] - Metadata features
// [543:768] - Padding (zeros)
```

### Feature Consistency Validation

Training and inference must produce identical features for the same input:

```python
def validate_consistency(
    training_features: np.ndarray,
    inference_features: np.ndarray,
    tolerance: float = 1e-6
) -> bool:
    return np.allclose(training_features, inference_features, atol=tolerance)
```

---

## Consent and Retention Constraints

### Consent Categories

| Category | Scope Types | Required Consent |
|----------|-------------|------------------|
| `biometric_pii` | id_document | Explicit, purpose-specific |
| `biometric` | selfie, face_video, biometric | Explicit |
| `identity_attestation` | sso_metadata | Implicit via SSO flow |
| `contact_verification` | email_proof, sms_proof | Implicit via verification |
| `domain_ownership` | domain_verify | Implicit via DNS record |
| `enterprise_identity` | ad_sso | Delegated via enterprise |

### Feature-to-Consent Mapping

| Feature Group | Consent Category | Can Share | Retention Policy |
|---------------|------------------|-----------|------------------|
| Face Embedding | biometric | With consent | Hash: indefinite; Raw: 7 days |
| Document Quality | biometric_pii | With consent | 30 days max |
| OCR Features | biometric_pii | Field hashes only | Hash: indefinite; Raw: 7 days |
| Metadata Features | identity_attestation | Anonymized only | Indefinite |

### Retention Constraints by Feature Type

From `x/veid/types/data_lifecycle.go`:

| Artifact Type | On-Chain | Encryption Required | Default Retention | Max Retention |
|---------------|----------|---------------------|-------------------|---------------|
| Raw Image | ❌ | ✅ | 7 days | 30 days |
| Processed Image | ❌ | ✅ | 1 day | 7 days |
| Face Embedding | ✅ (hash) | ✅ (raw) | 365 days | Indefinite |
| Document Hash | ✅ | ❌ | 365 days | Indefinite |
| OCR Data | ❌ | ✅ | 7 days | 30 days |
| Verification Record | ✅ | ❌ | Indefinite | Indefinite |

### Derived Feature Sharing

Derived feature hashes can be shared when:

1. `AllowDerivedFeatureSharing` consent is granted
2. Features are cryptographically hashed before sharing
3. Original data has been deleted per retention policy

---

## Migration Policy

### Schema Migration Process

1. **Proposal**: Submit schema change via governance proposal
2. **Review**: Security and privacy impact assessment
3. **Soft Fork**: Deploy new schema version with backward compatibility
4. **Migration Window**: Allow validators to update (14-day minimum)
5. **Hard Fork**: If breaking changes, require network upgrade

### Feature Addition Process

To add a new feature within a major version:

1. Allocate dimensions from padding (`PaddingDim`)
2. Define default value for backward compatibility
3. Update training pipeline to extract new feature
4. Deploy model trained with new feature
5. Update inference pipeline
6. Bump MINOR version

### Breaking Change Process

Breaking changes require:

1. MAJOR version bump
2. Governance approval
3. Model retraining with version migration
4. Verification record migration plan
5. Minimum 30-day notice period

### Model Versioning

Models must be versioned alongside schema:

```go
type ModelMetadata struct {
    Version           string  // e.g., "v1.2.0"
    SchemaVersion     string  // e.g., "1.0.0"
    Hash              string  // SHA256 of weights
    InputShape        []int64 // [1, 768]
    OutputShape       []int64 // [1, 1]
}
```

---

## Examples

### Example 1: Full Verification Request Features

```json
{
  "schema_version": "1.0.0",
  "scope_types": ["id_document", "selfie"],
  "features": {
    "face_embedding": [0.123, -0.456, ...],  // 512 values
    "face_confidence": 0.95,
    "doc_quality": {
      "sharpness": 0.85,
      "brightness": 0.72,
      "contrast": 0.68,
      "noise_level": 0.12,
      "blur_score": 0.08
    },
    "ocr": {
      "name": {"confidence": 0.92, "validated": true},
      "date_of_birth": {"confidence": 0.88, "validated": true},
      "document_number": {"confidence": 0.95, "validated": true},
      "expiry_date": {"confidence": 0.90, "validated": true},
      "nationality": {"confidence": 0.85, "validated": true}
    },
    "metadata": {
      "scope_count": 0.2,
      "has_id_document": 1.0,
      "has_selfie": 1.0,
      "face_confidence": 0.95,
      "block_height_normalized": 0.123456
    }
  }
}
```

### Example 2: Minimal Verification (Email Only)

```json
{
  "schema_version": "1.0.0",
  "scope_types": ["email_proof"],
  "features": {
    "face_embedding": [0.0, 0.0, ...],  // 512 zeros
    "face_confidence": 0.0,
    "doc_quality": {
      "sharpness": 0.5,
      "brightness": 0.5,
      "contrast": 0.5,
      "noise_level": 0.5,
      "blur_score": 0.5
    },
    "ocr": {
      "name": {"confidence": 0.0, "validated": false},
      "date_of_birth": {"confidence": 0.0, "validated": false},
      "document_number": {"confidence": 0.0, "validated": false},
      "expiry_date": {"confidence": 0.0, "validated": false},
      "nationality": {"confidence": 0.0, "validated": false}
    },
    "metadata": {
      "scope_count": 0.1,
      "has_email_proof": 1.0,
      "block_height_normalized": 0.234567
    }
  }
}
```

### Example 3: Feature Vector Reconstruction

```go
// Reconstruct 768-dim feature vector from ScoreInputs
func (fe *FeatureExtractor) ExtractFeatures(inputs *ScoreInputs) ([]float32, error) {
    features := make([]float32, TotalFeatureDim)  // 768
    
    // [0:512] Face embedding
    copy(features[0:512], inputs.FaceEmbedding)
    fe.normalizeEmbedding(features[0:512])
    
    // [512:517] Document quality
    features[512] = inputs.DocQualityScore
    features[513] = inputs.DocQualityFeatures.Sharpness
    features[514] = inputs.DocQualityFeatures.Brightness
    features[515] = inputs.DocQualityFeatures.Contrast
    features[516] = 1.0 - inputs.DocQualityFeatures.NoiseLevel
    
    // [517:527] OCR features
    for i, field := range OCRFieldNames {
        features[517+i*2] = inputs.OCRConfidences[field]
        if inputs.OCRFieldValidation[field] {
            features[517+i*2+1] = 1.0
        }
    }
    
    // [527:543] Metadata
    features[527] = float32(inputs.ScopeCount) / 10.0
    // ... scope type indicators
    
    return features, nil
}
```

---

## Appendix A: Feature Dimension Map

| Offset | End | Dimension | Feature Group | Description |
|--------|-----|-----------|---------------|-------------|
| 0 | 511 | 512 | Face | Face embedding vector |
| 512 | 516 | 5 | DocQuality | Document quality metrics |
| 517 | 526 | 10 | OCR | OCR field features |
| 527 | 542 | 16 | Metadata | Contextual metadata |
| 543 | 767 | 225 | Padding | Reserved for future use |

## Appendix B: OCR Field Names

```go
var OCRFieldNames = []string{
    "name",           // Index 0
    "date_of_birth",  // Index 1
    "document_number", // Index 2
    "expiry_date",    // Index 3
    "nationality",    // Index 4
}
```

## Appendix C: Scope Type Weights

```go
func ScopeTypeWeight(scopeType ScopeType) uint32 {
    switch scopeType {
    case ScopeTypeIDDocument:  return 30
    case ScopeTypeFaceVideo:   return 25
    case ScopeTypeSelfie:      return 20
    case ScopeTypeBiometric:   return 20
    case ScopeTypeDomainVerify: return 15
    case ScopeTypeADSSO:       return 12
    case ScopeTypeEmailProof:  return 10
    case ScopeTypeSMSProof:    return 10
    case ScopeTypeSSOMetadata: return 5
    default:                   return 0
    }
}
```
