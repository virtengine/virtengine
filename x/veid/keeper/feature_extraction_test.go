package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/inference"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Media Parser Tests
// ============================================================================

func TestMediaParser_ParseImagePayload(t *testing.T) {
	parser := NewMediaParser()

	tests := []struct {
		name       string
		payload    []byte
		scopeType  types.ScopeType
		wantType   MediaType
		wantFormat ImageFormat
		wantErr    bool
	}{
		{
			name:       "valid JPEG",
			payload:    makeJPEGPayload(),
			scopeType:  types.ScopeTypeSelfie,
			wantType:   MediaTypeImage,
			wantFormat: ImageFormatJPEG,
			wantErr:    false,
		},
		{
			name:       "valid PNG",
			payload:    makePNGPayload(),
			scopeType:  types.ScopeTypeIDDocument,
			wantType:   MediaTypeImage,
			wantFormat: ImageFormatPNG,
			wantErr:    false,
		},
		{
			name:      "empty payload",
			payload:   []byte{},
			scopeType: types.ScopeTypeSelfie,
			wantErr:   true,
		},
		{
			name:      "invalid image format",
			payload:   []byte("not an image"),
			scopeType: types.ScopeTypeSelfie,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted := DecryptedScope{
				ScopeID:     "test-scope",
				ScopeType:   tt.scopeType,
				Plaintext:   tt.payload,
				ContentHash: sha256Sum(tt.payload),
			}

			result, err := parser.Parse(decrypted)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, result.MediaType)
			if result.ImageData != nil {
				assert.Equal(t, tt.wantFormat, result.ImageData.Format)
			}
		})
	}
}

func TestMediaParser_ParseVideoPayload(t *testing.T) {
	parser := NewMediaParser()

	tests := []struct {
		name       string
		payload    []byte
		wantFormat VideoFormat
		wantErr    bool
	}{
		{
			name:       "valid MP4",
			payload:    makeMP4Payload(),
			wantFormat: VideoFormatMP4,
			wantErr:    false,
		},
		{
			name:       "valid WebM",
			payload:    makeWebMPayload(),
			wantFormat: VideoFormatWebM,
			wantErr:    false,
		},
		{
			name:    "too small video",
			payload: []byte{0x00, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted := DecryptedScope{
				ScopeID:     "test-video",
				ScopeType:   types.ScopeTypeFaceVideo,
				Plaintext:   tt.payload,
				ContentHash: sha256Sum(tt.payload),
			}

			result, err := parser.Parse(decrypted)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, MediaTypeVideo, result.MediaType)
			assert.Equal(t, tt.wantFormat, result.VideoData.Format)
		})
	}
}

func TestMediaParser_ParseJSONPayload(t *testing.T) {
	parser := NewMediaParser()

	decrypted := DecryptedScope{
		ScopeID:     "test-json",
		ScopeType:   types.ScopeTypeSSOMetadata,
		Plaintext:   []byte(`{"provider": "azure", "tenant_id": "abc123"}`),
		ContentHash: []byte{0x01, 0x02, 0x03, 0x04},
	}

	result, err := parser.Parse(decrypted)
	require.NoError(t, err)

	assert.Equal(t, MediaTypeJSON, result.MediaType)
	assert.NotNil(t, result.JSONData)
	assert.Equal(t, "azure", result.JSONData["provider"])
}

// ============================================================================
// Feature Extraction Pipeline Tests
// ============================================================================

func TestFeatureExtractionPipeline_ExtractFeatures(t *testing.T) {
	config := DefaultFeatureExtractionConfig()
	pipeline := NewFeatureExtractionPipeline(config)

	// Create test scopes
	selfiePayload := makeJPEGPayload()
	idDocPayload := makePNGPayload()
	videoPayload := makeMP4Payload()

	scopes := []DecryptedScope{
		{
			ScopeID:     "selfie-1",
			ScopeType:   types.ScopeTypeSelfie,
			Plaintext:   selfiePayload,
			ContentHash: sha256Sum(selfiePayload),
		},
		{
			ScopeID:     "id-doc-1",
			ScopeType:   types.ScopeTypeIDDocument,
			Plaintext:   idDocPayload,
			ContentHash: sha256Sum(idDocPayload),
		},
		{
			ScopeID:     "video-1",
			ScopeType:   types.ScopeTypeFaceVideo,
			Plaintext:   videoPayload,
			ContentHash: sha256Sum(videoPayload),
		},
	}

	features, err := pipeline.ExtractFeatures(
		scopes,
		"cosmos1test",
		12345,
		time.Now(),
	)

	require.NoError(t, err)
	assert.NotNil(t, features)

	// Verify face embedding
	assert.Equal(t, types.FaceEmbeddingDim, len(features.FaceEmbedding))
	assert.True(t, features.FaceConfidence >= 0.0 && features.FaceConfidence <= 1.0)

	// Verify embedding is L2 normalized
	var sumSquares float64
	for _, v := range features.FaceEmbedding {
		sumSquares += float64(v) * float64(v)
	}
	norm := math.Sqrt(sumSquares)
	assert.InDelta(t, 1.0, norm, 0.01, "face embedding should be L2 normalized")

	// Verify OCR features
	assert.Equal(t, 5, len(features.OCRConfidences))
	for field, conf := range features.OCRConfidences {
		assert.True(t, conf >= 0.0 && conf <= 1.0, "OCR confidence for %s should be in [0,1]", field)
	}

	// Verify liveness features
	assert.True(t, features.LivenessScore >= 0.0 && features.LivenessScore <= 1.0)
	assert.Contains(t, []string{"live", "uncertain", "spoof"}, features.LivenessDecision)

	// Verify metadata
	assert.NotNil(t, features.Metadata)
	assert.Equal(t, types.MLFeatureSchemaVersion, features.Metadata.SchemaVersion)
}

func TestFeatureExtractionPipeline_Determinism(t *testing.T) {
	config := DefaultFeatureExtractionConfig()
	pipeline := NewFeatureExtractionPipeline(config)

	// Create a deterministic test scope
	selfiePayload := makeJPEGPayload()
	scope := DecryptedScope{
		ScopeID:     "selfie-determinism",
		ScopeType:   types.ScopeTypeSelfie,
		Plaintext:   selfiePayload,
		ContentHash: sha256Sum(selfiePayload),
	}

	blockHeight := int64(12345)
	blockTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Extract features twice
	features1, err := pipeline.ExtractFeatures(
		[]DecryptedScope{scope},
		"cosmos1test",
		blockHeight,
		blockTime,
	)
	require.NoError(t, err)

	features2, err := pipeline.ExtractFeatures(
		[]DecryptedScope{scope},
		"cosmos1test",
		blockHeight,
		blockTime,
	)
	require.NoError(t, err)

	// Verify embeddings are identical
	assert.Equal(t, len(features1.FaceEmbedding), len(features2.FaceEmbedding))
	for i := range features1.FaceEmbedding {
		assert.Equal(t, features1.FaceEmbedding[i], features2.FaceEmbedding[i],
			"embedding[%d] should be deterministic", i)
	}

	// Verify other features are identical
	assert.Equal(t, features1.FaceConfidence, features2.FaceConfidence)
	assert.Equal(t, features1.DocQualityScore, features2.DocQualityScore)
}

func TestFeatureExtractionPipeline_NormalizeToMLVector(t *testing.T) {
	config := DefaultFeatureExtractionConfig()
	pipeline := NewFeatureExtractionPipeline(config)

	features := NewRealExtractedFeatures()
	features.FaceEmbedding = make([]float32, types.FaceEmbeddingDim)
	for i := range features.FaceEmbedding {
		features.FaceEmbedding[i] = 0.01 * float32(i%100-50)
	}
	// Normalize
	var sumSq float64
	for _, v := range features.FaceEmbedding {
		sumSq += float64(v) * float64(v)
	}
	norm := float32(math.Sqrt(sumSq))
	for i := range features.FaceEmbedding {
		features.FaceEmbedding[i] /= norm
	}

	features.DocQualityScore = 0.85
	features.DocQualityFeatures = inference.DocQualityFeatures{
		Sharpness:  0.9,
		Brightness: 0.7,
		Contrast:   0.8,
		NoiseLevel: 0.1,
		BlurScore:  0.15,
	}
	features.OCRConfidences = map[string]float32{
		"name":            0.95,
		"date_of_birth":   0.90,
		"document_number": 0.85,
		"expiry_date":     0.80,
		"nationality":     0.75,
	}
	features.OCRFieldValidation = map[string]bool{
		"name":            true,
		"date_of_birth":   true,
		"document_number": true,
		"expiry_date":     true,
		"nationality":     true,
	}

	vector, err := pipeline.NormalizeToMLVector(features)
	require.NoError(t, err)
	assert.NotNil(t, vector)

	// Validate the vector
	err = vector.Validate()
	assert.NoError(t, err)

	// Check dimensions
	assert.Equal(t, types.TotalFeatureDim, len(vector.Features))
}

// ============================================================================
// Quality Gates Tests
// ============================================================================

func TestQualityGates_Check(t *testing.T) {
	config := DefaultQualityGatesConfig()
	gates := NewQualityGates(config)

	tests := []struct {
		name           string
		features       *RealExtractedFeatures
		expectAllPass  bool
		expectFailures int
	}{
		{
			name: "all gates pass",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.75,
					Brightness: 0.6,
					FaceSize:   0.3,
				},
				DocQualityScore: 0.8,
				DocQualityFeatures: inference.DocQualityFeatures{
					Sharpness:  0.8,
					Brightness: 0.6,
					Contrast:   0.7,
					NoiseLevel: 0.1,
					BlurScore:  0.2,
				},
				OCRConfidences: map[string]float32{
					"name": 0.9, "date_of_birth": 0.85,
				},
				OCRFieldValidation: map[string]bool{
					"name": true, "date_of_birth": true,
				},
				LivenessScore:    0.8,
				LivenessDecision: "live",
			},
			expectAllPass:  true,
			expectFailures: 0,
		},
		{
			name: "face confidence too low",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.5, // Below 0.7 threshold
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.75,
					Brightness: 0.6,
					FaceSize:   0.3,
				},
			},
			expectAllPass:  false,
			expectFailures: 1,
		},
		{
			name: "document too blurry",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness: 0.75, Brightness: 0.6, FaceSize: 0.3,
				},
				DocQualityScore: 0.5,
				DocQualityFeatures: inference.DocQualityFeatures{
					Sharpness:  0.2, // Below 0.4 threshold
					Brightness: 0.6,
					Contrast:   0.7,
					NoiseLevel: 0.1,
					BlurScore:  0.8, // Above 0.5 threshold
				},
			},
			expectAllPass:  false,
			expectFailures: 2,
		},
		{
			name: "liveness check failed",
			features: &RealExtractedFeatures{
				FaceEmbedding:    make([]float32, types.FaceEmbeddingDim),
				FaceConfidence:   0.85,
				FaceQuality:      FaceQualityMetrics{Sharpness: 0.75, Brightness: 0.6, FaceSize: 0.3},
				LivenessScore:    0.3, // Below 0.5 threshold
				LivenessDecision: "spoof",
			},
			expectAllPass:  false,
			expectFailures: 2, // score and decision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := gates.Check(tt.features)

			allPassed := AllQualityGatesPassed(results)
			assert.Equal(t, tt.expectAllPass, allPassed)

			if tt.expectFailures > 0 {
				failed := GetFailedGates(results)
				assert.GreaterOrEqual(t, len(failed), tt.expectFailures)
			}
		})
	}
}

func TestQualityGates_GetFailureReasonCodes(t *testing.T) {
	results := []QualityGateResult{
		{Gate: QualityGateFaceConfidence, Passed: false, ReasonCode: types.ReasonCodeFaceMismatch},
		{Gate: QualityGateFaceSharpness, Passed: true, ReasonCode: types.ReasonCodeSuccess},
		{Gate: QualityGateDocBlur, Passed: false, ReasonCode: types.ReasonCodeLowDocQuality},
		{Gate: QualityGateLivenessScore, Passed: false, ReasonCode: types.ReasonCodeLivenessCheckFailed},
	}

	codes := GetFailureReasonCodes(results)
	assert.Len(t, codes, 3)
	assert.Contains(t, codes, types.ReasonCodeFaceMismatch)
	assert.Contains(t, codes, types.ReasonCodeLowDocQuality)
	assert.Contains(t, codes, types.ReasonCodeLivenessCheckFailed)
}

// ============================================================================
// Feature Extraction Metadata Tests
// ============================================================================

func TestFeatureExtractionMetadata_ComputeFeatureHash(t *testing.T) {
	meta := types.NewFeatureExtractionMetadata("1.0.0", "cosmos1test", 12345)

	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = 0.01 * float32(i)
	}

	meta.ComputeFeatureHash(embedding, 0.85)

	assert.NotEmpty(t, meta.FeatureHash)
	assert.Len(t, meta.FeatureHash, 64) // SHA256 hex

	// Same inputs should produce same hash
	meta2 := types.NewFeatureExtractionMetadata("1.0.0", "cosmos1test", 12345)
	meta2.ComputeFeatureHash(embedding, 0.85)

	assert.Equal(t, meta.FeatureHash, meta2.FeatureHash)
}

func TestFeatureExtractionMetadata_AddModelUsed(t *testing.T) {
	meta := types.NewFeatureExtractionMetadata("1.0.0", "cosmos1test", 12345)

	meta.AddModelUsed("face_embedding", "FaceNet512", "1.0.0")
	meta.AddModelUsed("ocr", "Tesseract", "4.1.1")

	assert.Len(t, meta.ModelsUsed, 2)

	versions := meta.GetModelVersions()
	assert.Equal(t, "1.0.0", versions["face_embedding"])
	assert.Equal(t, "4.1.1", versions["ocr"])
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestFeatureExtraction_ToScoreInputs(t *testing.T) {
	config := DefaultFeatureExtractionConfig()
	pipeline := NewFeatureExtractionPipeline(config)

	// Create features
	features := &RealExtractedFeatures{
		FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
		FaceConfidence: 0.9,
		DocQualityScore: 0.85,
		DocQualityFeatures: inference.DocQualityFeatures{
			Sharpness: 0.9, Brightness: 0.7, Contrast: 0.8,
		},
		OCRConfidences: map[string]float32{"name": 0.95},
		OCRFieldValidation: map[string]bool{"name": true},
		LivenessScore: 0.85,
	}

	scopes := []DecryptedScope{
		{ScopeID: "scope-1", ScopeType: types.ScopeTypeSelfie},
		{ScopeID: "scope-2", ScopeType: types.ScopeTypeIDDocument},
	}

	inputs := pipeline.ToScoreInputs(features, scopes, "cosmos1test", 12345, time.Now())

	assert.Equal(t, features.FaceEmbedding, inputs.FaceEmbedding)
	assert.Equal(t, features.FaceConfidence, inputs.FaceConfidence)
	assert.Equal(t, features.DocQualityScore, inputs.DocQualityScore)
	assert.Equal(t, 2, inputs.ScopeCount)
	assert.Contains(t, inputs.ScopeTypes, "selfie")
	assert.Contains(t, inputs.ScopeTypes, "id_document")
}

// ============================================================================
// Test Helpers
// ============================================================================

func makeJPEGPayload() []byte {
	// Minimal valid JPEG header + padding to meet size requirements
	header := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	padding := make([]byte, 2048)
	return append(header, padding...)
}

func makePNGPayload() []byte {
	// PNG magic bytes + padding
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	padding := make([]byte, 2048)
	return append(header, padding...)
}

func makeMP4Payload() []byte {
	// MP4 ftyp box header + padding
	header := make([]byte, 12)
	binary.BigEndian.PutUint32(header[0:4], 20) // Box size
	copy(header[4:8], []byte("ftyp"))
	copy(header[8:12], []byte("mp42"))
	padding := make([]byte, 20000) // Enough for estimated frame count
	return append(header, padding...)
}

func makeWebMPayload() []byte {
	// WebM EBML header + padding
	header := []byte{0x1A, 0x45, 0xDF, 0xA3}
	padding := make([]byte, 15000)
	return append(header, padding...)
}

func sha256Sum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
