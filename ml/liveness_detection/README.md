# VirtEngine Liveness Detection Module

## Overview

VE-901: Liveness detection module for anti-spoofing in the VEID identity verification pipeline.

This module provides comprehensive liveness detection to prevent presentation attacks during facial biometric capture.

## Features

### Active Liveness Detection
- **Blink Detection**: Eye Aspect Ratio (EAR) analysis
- **Smile Detection**: Mouth geometry analysis
- **Head Turn Detection**: Pose estimation for left/right turns
- **Head Nod Detection**: Pitch angle analysis
- **Eyebrow Raise Detection**: Landmark distance analysis

### Passive Liveness Detection
- **Texture Analysis**: Local Binary Patterns (LBP) for skin texture
- **Depth Analysis**: Gradient-based monocular depth estimation
- **Motion Analysis**: Optical flow and frame differencing
- **Reflection Analysis**: Specular highlight detection
- **Moire Pattern Detection**: FFT-based screen artifact detection
- **Frequency Analysis**: High/low frequency distribution

### Spoof Detection
- **Photo Print Detection**: Paper/ink texture, color saturation
- **Screen Display Detection**: Moire patterns, color banding
- **Video Replay Detection**: Temporal consistency, compression artifacts
- **2D Mask Detection**: Edge analysis, texture uniformity
- **3D Mask Detection**: Depth uniformity, material analysis
- **Deepfake Detection**: Boundary artifacts, temporal coherence

## Usage

```python
from ml.liveness_detection import LivenessDetector, LivenessConfig, ChallengeType

# Create detector with default config
config = LivenessConfig()
detector = LivenessDetector(config)

# Analyze frame sequence
result = detector.detect(
    frames=frame_sequence,          # List of BGR numpy arrays
    face_regions=detected_faces,    # List of (x, y, w, h) tuples
    landmarks=landmark_sequence,    # List of LandmarkData objects
    required_challenges=[ChallengeType.BLINK],
)

if result.is_live:
    print(f"Liveness confirmed: score={result.liveness_score:.2f}")
    print(f"Confidence: {result.confidence:.2f}")
else:
    print(f"Liveness failed: {result.decision}")
    print(f"Reason codes: {result.reason_codes}")

# Get VEID-compatible record
veid_record = result.to_veid_record()
```

## Configuration

### Default Configuration
```python
from ml.liveness_detection import get_default_config

config = get_default_config()
# Required challenges: blink
# Pass threshold: 0.75
# Spoof threshold: 0.50
```

### Strict Configuration
```python
from ml.liveness_detection.config import get_strict_config

config = get_strict_config()
# Required challenges: blink, smile, head_turn_left
# Pass threshold: 0.85
# Spoof threshold: 0.40
```

### Permissive Configuration
```python
from ml.liveness_detection.config import get_permissive_config

config = get_permissive_config()
# Required challenges: blink only
# Pass threshold: 0.60
# Spoof threshold: 0.60
```

## VEID Integration

The liveness score is integrated into the VEID scoring model:

```go
// In x/veid/types/scoring_model.go
type ScoringWeights struct {
    // ... other weights ...
    LivenessCheckWeight uint32 `json:"liveness_check_weight"` // 10% = 1000 basis points
}
```

The liveness score contributes to the overall identity score:
- **10% weight** in the default scoring model
- Minimum threshold of **75%** (7500 basis points) required to pass
- Scores are in basis points (0-10000 = 0-100%)

## Reason Codes

### Success Codes
- `LIVENESS_CONFIRMED` - Liveness verified
- `HIGH_CONFIDENCE_LIVE` - High confidence verification
- `ALL_CHALLENGES_PASSED` - All challenges completed

### Challenge Failure Codes
- `BLINK_NOT_DETECTED`, `BLINK_TOO_FAST`, `BLINK_TOO_SLOW`
- `SMILE_NOT_DETECTED`, `SMILE_INSUFFICIENT`
- `HEAD_TURN_NOT_DETECTED`, `HEAD_TURN_WRONG_DIRECTION`
- `CHALLENGE_TIMEOUT`, `CHALLENGE_INCOMPLETE`

### Spoof Detection Codes
- `PHOTO_PRINT_DETECTED` - Printed photo attack
- `SCREEN_DISPLAY_DETECTED` - Phone/tablet/monitor attack
- `VIDEO_REPLAY_DETECTED` - Recorded video attack
- `MASK_2D_DETECTED`, `MASK_3D_DETECTED` - Mask attacks
- `DEEPFAKE_DETECTED` - Synthetic face attack

## Determinism

The module is designed for deterministic execution:
- Fixed random seeds
- Reproducible model hashes
- Result hashes for consensus verification

```python
# Result hash for consensus
result = detector.detect(frames, ...)
print(result.result_hash)  # SHA256 of result for verification
print(result.model_hash)   # SHA256 of model config
```

## Security Considerations

1. **No Raw Biometrics Logging**: The module never logs raw facial data
2. **Score Only Output**: Only scores and reason codes are persisted
3. **Deterministic Verification**: Results can be verified across validators
4. **Multi-layer Detection**: Combines active, passive, and spoof detection

## Testing

Run the test suite:
```bash
cd ml/liveness_detection
pytest tests/ -v
```

## Dependencies

- numpy >= 1.20.0
- OpenCV (cv2) >= 4.5.0 (optional, for advanced features)

See `requirements.txt` for full dependency list.
