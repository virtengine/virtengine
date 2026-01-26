package mobile

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	testDeviceID      = "device-123"
	testAppInstanceID = "instance-456"
	testSessionID     = "session-123"
	testFingerprint   = "fingerprint-abc"
	testImageData     = "test image data"
)

// ============================================================================
// Types Tests
// ============================================================================

func TestResolutionMegapixels(t *testing.T) {
	tests := []struct {
		name       string
		resolution Resolution
		expected   float64
	}{
		{
			name:       "1080p",
			resolution: Resolution{Width: 1920, Height: 1080},
			expected:   2.0736,
		},
		{
			name:       "4K",
			resolution: Resolution{Width: 3840, Height: 2160},
			expected:   8.2944,
		},
		{
			name:       "720p",
			resolution: Resolution{Width: 1280, Height: 720},
			expected:   0.9216,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.resolution.Megapixels()
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestDefaultDocumentCameraConfig(t *testing.T) {
	config := DefaultDocumentCameraConfig()

	assert.Equal(t, CameraPositionBack, config.Position)
	assert.Equal(t, 3840, config.TargetResolution.Width)
	assert.Equal(t, 2160, config.TargetResolution.Height)
	assert.True(t, config.AutoFocusEnabled)
	assert.Equal(t, FlashModeAuto, config.FlashMode)
	assert.True(t, config.ContinuousCapture)
	assert.Equal(t, 30, config.FrameRate)
}

func TestDefaultSelfieCameraConfig(t *testing.T) {
	config := DefaultSelfieCameraConfig()

	assert.Equal(t, CameraPositionFront, config.Position)
	assert.Equal(t, 1920, config.TargetResolution.Width)
	assert.Equal(t, 1080, config.TargetResolution.Height)
	assert.True(t, config.AutoFocusEnabled)
	assert.Equal(t, FlashModeOff, config.FlashMode)
}

func TestDeviceFingerprintComputeHash(t *testing.T) {
	fp := &DeviceFingerprint{
		Platform:      PlatformIOS,
		DeviceID:      testDeviceID,
		AppInstanceID: testAppInstanceID,
		HardwareID:    "test-hardware-id",
		CreatedAt:     time.Now(),
	}

	hash := fp.ComputeFingerprintHash()
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hex = 64 chars

	// Same input should produce same hash
	hash2 := fp.ComputeFingerprintHash()
	assert.Equal(t, hash, hash2)

	// Different input should produce different hash
	fp.DeviceID = "different-id"
	hash3 := fp.ComputeFingerprintHash()
	assert.NotEqual(t, hash, hash3)
}

func TestDeviceFingerprintValidate(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint DeviceFingerprint
		expectError bool
	}{
		{
			name: "valid fingerprint",
			fingerprint: DeviceFingerprint{
				Platform:      PlatformIOS,
				DeviceID:      testDeviceID,
				AppInstanceID: testAppInstanceID,
			},
			expectError: false,
		},
		{
			name: "missing platform",
			fingerprint: DeviceFingerprint{
				DeviceID:      testDeviceID,
				AppInstanceID: testAppInstanceID,
			},
			expectError: true,
		},
		{
			name: "missing device ID",
			fingerprint: DeviceFingerprint{
				Platform:      PlatformAndroid,
				AppInstanceID: testAppInstanceID,
			},
			expectError: true,
		},
		{
			name: "missing app instance ID",
			fingerprint: DeviceFingerprint{
				Platform: PlatformIOS,
				DeviceID: testDeviceID,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fingerprint.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Quality Tests
// ============================================================================

func TestDefaultDocumentQualityConfig(t *testing.T) {
	config := DefaultDocumentQualityConfig()

	assert.Equal(t, 1024, config.MinResolution.Width)
	assert.Equal(t, 768, config.MinResolution.Height)
	assert.Equal(t, 40, config.MinBrightness)
	assert.Equal(t, 220, config.MaxBrightness)
	assert.Equal(t, 70, config.MinQualityScore)
	assert.True(t, config.RequireDocumentDetection)
	assert.False(t, config.RequireFaceDetection)
}

func TestDefaultSelfieQualityConfig(t *testing.T) {
	config := DefaultSelfieQualityConfig()

	assert.Equal(t, 640, config.MinResolution.Width)
	assert.Equal(t, 480, config.MinResolution.Height)
	assert.Equal(t, 75, config.MinQualityScore)
	assert.False(t, config.RequireDocumentDetection)
	assert.True(t, config.RequireFaceDetection)
}

func TestDefaultEnglishGuidance(t *testing.T) {
	guidance := DefaultEnglishGuidance()

	assert.NotEmpty(t, guidance.PositionDocument)
	assert.NotEmpty(t, guidance.HoldSteady)
	assert.NotEmpty(t, guidance.MoreLight)
	assert.NotEmpty(t, guidance.Ready)
	assert.NotEmpty(t, guidance.Capturing)
}

// ============================================================================
// Liveness Tests
// ============================================================================

func TestDefaultLivenessConfig(t *testing.T) {
	config := DefaultLivenessConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, LivenessModeHybrid, config.Mode)
	assert.Len(t, config.ChallengeTypes, 2)
	assert.Contains(t, config.ChallengeTypes, ChallengeBlink)
	assert.Contains(t, config.ChallengeTypes, ChallengeSmile)
	assert.Equal(t, 1, config.MinChallenges)
	assert.Equal(t, 0.85, config.MinConfidence)
	assert.True(t, config.UseDepthSensor)
	assert.True(t, config.RequireTexture)
}

// ============================================================================
// Gallery Prevention Tests
// ============================================================================

func TestDefaultGalleryPreventionConfig(t *testing.T) {
	config := DefaultGalleryPreventionConfig()

	assert.True(t, config.StrictMode)
	assert.Equal(t, 30*time.Second, config.MaxImageAge)
	assert.True(t, config.RequireTimestampBinding)
	assert.True(t, config.ValidateEXIF)
	assert.True(t, config.RejectEditedImages)
}

func TestCaptureOriginProofHash(t *testing.T) {
	now := time.Now()
	imageHash := sha256.Sum256([]byte(testImageData))

	proof := &CaptureOriginProof{
		ProofVersion:       1,
		CaptureTimestamp:   now,
		SystemTimestamp:    now,
		MonotonicTimestamp: now.UnixNano(),
		CameraSessionID:    testSessionID,
		FrameNumber:        42,
		DeviceFingerprint:  testFingerprint,
		ImageMetadata: CaptureImageMetadata{
			ImageHash: imageHash[:],
		},
	}

	t.Run("compute hash", func(t *testing.T) {
		hash := proof.ComputeProofHash()
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32) // SHA256 = 32 bytes

		// Same proof should produce same hash
		hash2 := proof.ComputeProofHash()
		assert.Equal(t, hash, hash2)
	})

	t.Run("verify hash", func(t *testing.T) {
		// Set correct proof hash
		proof.ProofHash = proof.ComputeProofHash()
		assert.True(t, proof.Verify())

		// Tamper with proof
		proof.FrameNumber = 999
		assert.False(t, proof.Verify())
	})
}

func TestGalleryPreventionValidator(t *testing.T) {
	validator := NewGalleryPreventionValidator(DefaultGalleryPreventionConfig())

	imageData := []byte(testImageData)
	imageHash := sha256.Sum256(imageData)
	now := time.Now()

	t.Run("valid origin proof", func(t *testing.T) {
		proof := &CaptureOriginProof{
			ProofVersion:       1,
			CaptureTimestamp:   now,
			SystemTimestamp:    now,
			MonotonicTimestamp: now.UnixNano(),
			CameraSessionID:    testSessionID,
			FrameNumber:        1,
			DeviceFingerprint:  testFingerprint,
			ImageMetadata: CaptureImageMetadata{
				Width:        1920,
				Height:       1080,
				ImageHash:    imageHash[:],
				EXIFStripped: true,
			},
		}
		proof.ProofHash = proof.ComputeProofHash()

		result, err := validator.ValidateCaptureOrigin(proof, imageData)
		require.NoError(t, err)
		assert.True(t, result.IsLiveCapture)
		assert.False(t, result.IsGalleryUpload)
		assert.False(t, result.Blocked)
	})

	t.Run("missing proof in strict mode", func(t *testing.T) {
		result, err := validator.ValidateCaptureOrigin(nil, imageData)
		require.NoError(t, err)
		assert.False(t, result.IsLiveCapture)
		assert.True(t, result.IsGalleryUpload)
		assert.True(t, result.Blocked)
		assert.Equal(t, "Missing capture origin proof", result.BlockReason)
	})

	t.Run("invalid proof hash", func(t *testing.T) {
		proof := &CaptureOriginProof{
			ProofVersion:       1,
			CaptureTimestamp:   now,
			SystemTimestamp:    now,
			MonotonicTimestamp: now.UnixNano(),
			CameraSessionID:    testSessionID,
			FrameNumber:        1,
			DeviceFingerprint:  testFingerprint,
			ImageMetadata: CaptureImageMetadata{
				ImageHash: imageHash[:],
			},
			ProofHash: []byte("invalid hash"),
		}

		result, err := validator.ValidateCaptureOrigin(proof, imageData)
		require.NoError(t, err)
		assert.False(t, result.IsLiveCapture)
		assert.True(t, result.Blocked)
	})

	t.Run("expired timestamp", func(t *testing.T) {
		oldTime := now.Add(-2 * time.Minute) // Older than 30 seconds
		proof := &CaptureOriginProof{
			ProofVersion:       1,
			CaptureTimestamp:   oldTime,
			SystemTimestamp:    oldTime,
			MonotonicTimestamp: oldTime.UnixNano(),
			CameraSessionID:    testSessionID,
			FrameNumber:        1,
			DeviceFingerprint:  testFingerprint,
			ImageMetadata: CaptureImageMetadata{
				ImageHash: imageHash[:],
			},
		}
		proof.ProofHash = proof.ComputeProofHash()

		result, err := validator.ValidateCaptureOrigin(proof, imageData)
		require.NoError(t, err)
		assert.True(t, result.Blocked)
		assert.Contains(t, result.BlockReason, "too old")
	})

	t.Run("future timestamp", func(t *testing.T) {
		futureTime := now.Add(5 * time.Minute)
		proof := &CaptureOriginProof{
			ProofVersion:       1,
			CaptureTimestamp:   futureTime,
			SystemTimestamp:    futureTime,
			MonotonicTimestamp: futureTime.UnixNano(),
			CameraSessionID:    testSessionID,
			FrameNumber:        1,
			DeviceFingerprint:  testFingerprint,
			ImageMetadata: CaptureImageMetadata{
				ImageHash: imageHash[:],
			},
		}
		proof.ProofHash = proof.ComputeProofHash()

		result, err := validator.ValidateCaptureOrigin(proof, imageData)
		require.NoError(t, err)
		assert.True(t, result.Blocked)
		assert.Contains(t, result.BlockReason, "future timestamp")
	})

	t.Run("image hash mismatch", func(t *testing.T) {
		wrongHash := sha256.Sum256([]byte("different data"))
		proof := &CaptureOriginProof{
			ProofVersion:       1,
			CaptureTimestamp:   now,
			SystemTimestamp:    now,
			MonotonicTimestamp: now.UnixNano(),
			CameraSessionID:    testSessionID,
			FrameNumber:        1,
			DeviceFingerprint:  testFingerprint,
			ImageMetadata: CaptureImageMetadata{
				ImageHash: wrongHash[:],
			},
		}
		proof.ProofHash = proof.ComputeProofHash()

		result, err := validator.ValidateCaptureOrigin(proof, imageData)
		require.NoError(t, err)
		assert.True(t, result.Blocked)
		assert.Contains(t, result.BlockReason, "hash mismatch")
	})
}

func TestGalleryDetectionResultIsBlocked(t *testing.T) {
	tests := []struct {
		name     string
		result   GalleryDetectionResult
		expected bool
	}{
		{
			name: "blocked explicitly",
			result: GalleryDetectionResult{
				Blocked:         true,
				IsGalleryUpload: false,
			},
			expected: true,
		},
		{
			name: "gallery upload",
			result: GalleryDetectionResult{
				Blocked:         false,
				IsGalleryUpload: true,
			},
			expected: true,
		},
		{
			name: "live capture",
			result: GalleryDetectionResult{
				Blocked:         false,
				IsGalleryUpload: false,
				IsLiveCapture:   true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsBlocked())
		})
	}
}

// ============================================================================
// Signing Tests
// ============================================================================

func TestMobileSaltBindingHash(t *testing.T) {
	binding := &MobileSaltBinding{
		Salt:              []byte("random-salt-bytes"),
		DeviceFingerprint: "device-fp-123",
		SessionID:         "session-456",
		Timestamp:         time.Now().Unix(),
	}

	t.Run("compute hash", func(t *testing.T) {
		hash := binding.ComputeBindingHash()
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32) // SHA256

		// Same binding produces same hash
		hash2 := binding.ComputeBindingHash()
		assert.Equal(t, hash, hash2)

		// Different timestamp produces different hash
		binding.Timestamp++
		hash3 := binding.ComputeBindingHash()
		assert.NotEqual(t, hash, hash3)
	})

	t.Run("verify hash", func(t *testing.T) {
		// Reset to known state
		binding.Timestamp = time.Now().Unix()
		// Set correct binding hash
		binding.BindingHash = binding.ComputeBindingHash()
		assert.True(t, binding.Verify())

		// Tamper with binding
		binding.SessionID = "tampered"
		assert.False(t, binding.Verify())
	})
}

func TestDefaultSignatureBuilderConfig(t *testing.T) {
	config := DefaultSignatureBuilderConfig()

	assert.Equal(t, uint32(1), config.ProtocolVersion)
	assert.True(t, config.RequireClientSignature)
	assert.True(t, config.RequireUserSignature)
	assert.False(t, config.RequireHardwareBackedKeys)
}

func TestComputeSigningData(t *testing.T) {
	salt := []byte("salt-data")
	payloadHash := []byte("payload-hash-data")
	clientSig := []byte("client-signature")

	t.Run("client signing data", func(t *testing.T) {
		data := computeClientSigningData(salt, payloadHash)
		assert.Len(t, data, len(salt)+len(payloadHash))
		assert.Equal(t, salt, data[:len(salt)])
		assert.Equal(t, payloadHash, data[len(salt):])
	})

	t.Run("user signing data", func(t *testing.T) {
		data := computeUserSigningData(salt, payloadHash, clientSig)
		assert.Len(t, data, len(salt)+len(payloadHash)+len(clientSig))
		assert.Equal(t, salt, data[:len(salt)])
		assert.Equal(t, payloadHash, data[len(salt):len(salt)+len(payloadHash)])
		assert.Equal(t, clientSig, data[len(salt)+len(payloadHash):])
	})
}

// ============================================================================
// Error Tests
// ============================================================================

func TestMobileError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &MobileError{
			Code:    ErrCodeCameraPermissionDenied,
			Message: "Camera permission was denied",
		}

		assert.Contains(t, err.Error(), ErrCodeCameraPermissionDenied)
		assert.Contains(t, err.Error(), "Camera permission was denied")
	})

	t.Run("wrap error", func(t *testing.T) {
		originalErr := assert.AnError
		err := ErrCameraPermissionDenied.Wrap(originalErr)

		assert.Equal(t, ErrCodeCameraPermissionDenied, err.Code)
		assert.Equal(t, originalErr, err.Unwrap())
		assert.Contains(t, err.Error(), originalErr.Error())
	})

	t.Run("with details", func(t *testing.T) {
		err := ErrCaptureQualityFailed.WithDetails(
			"score", 45,
			"required", 70,
		)

		assert.Equal(t, ErrCodeCaptureQualityFailed, err.Code)
		assert.Equal(t, 45, err.Details["score"])
		assert.Equal(t, 70, err.Details["required"])
	})

	t.Run("with field", func(t *testing.T) {
		err := ErrQualityResolutionLow.WithField("width")

		assert.Equal(t, ErrCodeQualityResolutionLow, err.Code)
		assert.Equal(t, "width", err.Field)
	})
}

func TestErrorTypeChecks(t *testing.T) {
	t.Run("camera errors", func(t *testing.T) {
		assert.True(t, IsCameraError(ErrCameraPermissionDenied))
		assert.True(t, IsCameraError(ErrCameraNotFound))
		assert.True(t, IsCameraError(ErrCameraSessionFailed))
		assert.False(t, IsCameraError(ErrCaptureTimeout))
		assert.False(t, IsCameraError(ErrLivenessTimeout))
	})

	t.Run("gallery errors", func(t *testing.T) {
		assert.True(t, IsGalleryError(ErrGalleryUploadBlocked))
		assert.True(t, IsGalleryError(ErrOriginProofMissing))
		assert.True(t, IsGalleryError(ErrOriginProofInvalid))
		assert.False(t, IsGalleryError(ErrCameraPermissionDenied))
	})

	t.Run("liveness errors", func(t *testing.T) {
		assert.True(t, IsLivenessError(ErrLivenessTimeout))
		assert.True(t, IsLivenessError(ErrLivenessChallengeFailed))
		assert.True(t, IsLivenessError(ErrSpoofingDetected))
		assert.False(t, IsLivenessError(ErrCaptureTimeout))
	})

	t.Run("quality errors", func(t *testing.T) {
		assert.True(t, IsQualityError(ErrQualityResolutionLow))
		assert.True(t, IsQualityError(ErrQualityTooDark))
		assert.True(t, IsQualityError(ErrQualityTooBlurry))
		assert.False(t, IsQualityError(ErrCameraPermissionDenied))
	})

	t.Run("retryable errors", func(t *testing.T) {
		// Retryable errors
		assert.True(t, IsRetryable(ErrCaptureQualityFailed))
		assert.True(t, IsRetryable(ErrCaptureNoDocument))
		assert.True(t, IsRetryable(ErrCaptureNoFace))
		assert.True(t, IsRetryable(ErrQualityTooDark))

		// Non-retryable errors
		assert.False(t, IsRetryable(ErrCameraPermissionDenied))
		assert.False(t, IsRetryable(ErrGalleryUploadBlocked))
		assert.False(t, IsRetryable(ErrSpoofingDetected))
	})
}

// ============================================================================
// Interface Tests
// ============================================================================

func TestDefaultMobileLibraryConfig(t *testing.T) {
	config := DefaultMobileLibraryConfig("test-client", "1.0.0")

	assert.Equal(t, "test-client", config.ClientID)
	assert.Equal(t, "1.0.0", config.ClientVersion)
	assert.Equal(t, EnvironmentProduction, config.Environment)
	assert.False(t, config.EnableDebugLogging)
	assert.NotEmpty(t, config.GuidanceMessages.Ready)
}

func TestCurrentSDKInfo(t *testing.T) {
	info := CurrentSDKInfo()

	assert.Equal(t, "1.0.0", info.SDKVersion)
	assert.Equal(t, MobileProtocolVersion, info.ProtocolVersion)
	assert.Contains(t, info.Features, "document_capture")
	assert.Contains(t, info.Features, "selfie_capture")
	assert.Contains(t, info.Features, "liveness_detection")
	assert.Contains(t, info.Features, "gallery_prevention")
	assert.Contains(t, info.Features, "quality_feedback")
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestInt64ToBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{256, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
		{-1, []byte{255, 255, 255, 255, 255, 255, 255, 255}},
	}

	for _, tt := range tests {
		result := int64ToBytes(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestConstantTimeEqual(t *testing.T) {
	a := []byte("test data")
	b := []byte("test data")
	c := []byte("different")
	d := []byte("test dat") // Different length

	assert.True(t, constantTimeEqual(a, b))
	assert.False(t, constantTimeEqual(a, c))
	assert.False(t, constantTimeEqual(a, d))
	assert.True(t, constantTimeEqual([]byte{}, []byte{}))
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCaptureOriginProofHash(b *testing.B) {
	now := time.Now()
	imageHash := sha256.Sum256([]byte(testImageData))

	proof := &CaptureOriginProof{
		ProofVersion:       1,
		CaptureTimestamp:   now,
		SystemTimestamp:    now,
		MonotonicTimestamp: now.UnixNano(),
		CameraSessionID:    testSessionID,
		FrameNumber:        42,
		DeviceFingerprint:  testFingerprint,
		ImageMetadata: CaptureImageMetadata{
			ImageHash: imageHash[:],
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof.ComputeProofHash()
	}
}

func BenchmarkMobileSaltBindingHash(b *testing.B) {
	binding := &MobileSaltBinding{
		Salt:              []byte("random-salt-bytes-32-characters!"),
		DeviceFingerprint: "device-fingerprint-hash",
		SessionID:         "session-id-12345",
		Timestamp:         time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binding.ComputeBindingHash()
	}
}

func BenchmarkGalleryPreventionValidation(b *testing.B) {
	validator := NewGalleryPreventionValidator(DefaultGalleryPreventionConfig())
	imageData := []byte("test image data for benchmark")
	imageHash := sha256.Sum256(imageData)
	now := time.Now()

	proof := &CaptureOriginProof{
		ProofVersion:       1,
		CaptureTimestamp:   now,
		SystemTimestamp:    now,
		MonotonicTimestamp: now.UnixNano(),
		CameraSessionID:    testSessionID,
		FrameNumber:        1,
		DeviceFingerprint:  testFingerprint,
		ImageMetadata: CaptureImageMetadata{
			ImageHash: imageHash[:],
		},
	}
	proof.ProofHash = proof.ComputeProofHash()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateCaptureOrigin(proof, imageData)
	}
}
