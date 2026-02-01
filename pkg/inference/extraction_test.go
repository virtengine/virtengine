package inference

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Face Extractor Tests
// ============================================================================

func TestFaceExtractor_ExtractFromSelfie_EmptyData(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	result, err := extractor.ExtractFromSelfie(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FaceDetected {
		t.Error("expected FaceDetected to be false for empty data")
	}
	if len(result.ReasonCodes) == 0 {
		t.Error("expected reason codes for empty data")
	}
}

func TestFaceExtractor_ExtractFromSelfie_StubMode(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config) // No client = stub mode

	imageData := []byte("test image data for face extraction")

	result, err := extractor.ExtractFromSelfie(context.Background(), imageData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.FaceDetected {
		t.Error("expected FaceDetected to be true in stub mode")
	}

	if len(result.Embedding) != config.EmbeddingDim {
		t.Errorf("expected embedding dimension %d, got %d", config.EmbeddingDim, len(result.Embedding))
	}

	if result.Confidence <= 0 || result.Confidence > 1 {
		t.Errorf("confidence out of range: %f", result.Confidence)
	}

	if result.EmbeddingHash == "" {
		t.Error("expected embedding hash to be computed")
	}
}

func TestFaceExtractor_ExtractFromSelfie_Deterministic(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	imageData := []byte("deterministic test data")

	result1, _ := extractor.ExtractFromSelfie(context.Background(), imageData)
	result2, _ := extractor.ExtractFromSelfie(context.Background(), imageData)

	if result1.EmbeddingHash != result2.EmbeddingHash {
		t.Error("expected deterministic embedding hash for same input")
	}

	// Compare embeddings
	for i := range result1.Embedding {
		if result1.Embedding[i] != result2.Embedding[i] {
			t.Errorf("embedding mismatch at index %d: %f != %f", i, result1.Embedding[i], result2.Embedding[i])
			break
		}
	}
}

func TestFaceExtractor_ExtractFromVideo_StubMode(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	videoData := []byte("test video data")
	maxFrames := 10

	result, err := extractor.ExtractFromVideo(context.Background(), videoData, maxFrames)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.BestEmbedding) != config.EmbeddingDim {
		t.Errorf("expected embedding dimension %d, got %d", config.EmbeddingDim, len(result.BestEmbedding))
	}

	if result.ConsistencyScore <= 0 || result.ConsistencyScore > 1 {
		t.Errorf("consistency score out of range: %f", result.ConsistencyScore)
	}
}

func TestFaceExtractor_ValidateEmbedding_Valid(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	// Create valid embedding
	embedding := make([]float32, config.EmbeddingDim)
	for i := range embedding {
		embedding[i] = float32(i) / float32(config.EmbeddingDim)
	}

	issues := extractor.ValidateEmbedding(embedding)
	if len(issues) > 0 {
		t.Errorf("expected no issues for valid embedding, got: %v", issues)
	}
}

func TestFaceExtractor_ValidateEmbedding_WrongDimension(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	embedding := make([]float32, 128) // Wrong dimension

	issues := extractor.ValidateEmbedding(embedding)
	if len(issues) == 0 {
		t.Error("expected issues for wrong dimension embedding")
	}
}

func TestFaceExtractor_ValidateEmbedding_AllZeros(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	embedding := make([]float32, config.EmbeddingDim) // All zeros

	issues := extractor.ValidateEmbedding(embedding)
	if len(issues) == 0 {
		t.Error("expected issues for all-zero embedding")
	}
}

func TestFaceExtractor_SanitizeEmbedding(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	embedding := make([]float32, config.EmbeddingDim)
	embedding[0] = 100.0   // Out of range
	embedding[1] = -100.0  // Out of range
	embedding[2] = 0.5     // Valid

	sanitized := extractor.SanitizeEmbedding(embedding)

	if sanitized[0] > 2.0 || sanitized[0] < -2.0 {
		t.Errorf("expected clamped value, got: %f", sanitized[0])
	}
	if sanitized[1] > 2.0 || sanitized[1] < -2.0 {
		t.Errorf("expected clamped value, got: %f", sanitized[1])
	}
	if sanitized[2] != 0.5 {
		t.Errorf("expected unchanged valid value, got: %f", sanitized[2])
	}
}

func TestFaceExtractor_IsHealthy_StubMode(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	if !extractor.IsHealthy() {
		t.Error("expected stub mode to be healthy")
	}
}

func TestFaceExtractor_GetStats(t *testing.T) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)

	// Run some extractions
	imageData := []byte("test")
	_, _ = extractor.ExtractFromSelfie(context.Background(), imageData)
	_, _ = extractor.ExtractFromSelfie(context.Background(), imageData)

	stats := extractor.GetStats()
	if stats.ExtractionCount != 2 {
		t.Errorf("expected extraction count 2, got %d", stats.ExtractionCount)
	}
	if !stats.UsingStub {
		t.Error("expected UsingStub to be true")
	}
}

// ============================================================================
// OCR Extractor Tests
// ============================================================================

func TestOCRExtractor_Extract_EmptyData(t *testing.T) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)

	result, err := extractor.Extract(context.Background(), nil, "passport")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Success {
		t.Error("expected Success to be false for empty data")
	}
}

func TestOCRExtractor_Extract_StubMode(t *testing.T) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)

	imageData := []byte("test document image")

	result, err := extractor.Extract(context.Background(), imageData, "passport")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected Success to be true in stub mode")
	}

	// Check expected fields
	for _, fieldName := range config.ExpectedFields {
		if _, exists := result.Fields[fieldName]; !exists {
			t.Errorf("expected field %s to exist", fieldName)
		}
	}

	if result.OverallConfidence <= 0 || result.OverallConfidence > 1 {
		t.Errorf("confidence out of range: %f", result.OverallConfidence)
	}

	if result.IdentityHash == "" {
		t.Error("expected identity hash to be computed")
	}
}

func TestOCRExtractor_Extract_Deterministic(t *testing.T) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)

	imageData := []byte("deterministic document")

	result1, _ := extractor.Extract(context.Background(), imageData, "passport")
	result2, _ := extractor.Extract(context.Background(), imageData, "passport")

	if result1.IdentityHash != result2.IdentityHash {
		t.Error("expected deterministic identity hash for same input")
	}
}

func TestOCRExtractor_DocumentQuality(t *testing.T) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)

	imageData := []byte("test document")

	result, _ := extractor.Extract(context.Background(), imageData, "id_card")

	quality := result.DocumentQuality
	if quality.Sharpness < 0 || quality.Sharpness > 1 {
		t.Errorf("sharpness out of range: %f", quality.Sharpness)
	}
	if quality.Brightness < 0 || quality.Brightness > 1 {
		t.Errorf("brightness out of range: %f", quality.Brightness)
	}
	if quality.OverallScore < 0 || quality.OverallScore > 1 {
		t.Errorf("overall score out of range: %f", quality.OverallScore)
	}
}

func TestOCRExtractor_ToFeatureInputs(t *testing.T) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)

	imageData := []byte("test document")
	result, _ := extractor.Extract(context.Background(), imageData, "passport")

	docScore, docFeatures, ocrConf, ocrValid := extractor.ToFeatureInputs(result)

	if docScore < 0 || docScore > 1 {
		t.Errorf("document score out of range: %f", docScore)
	}

	if docFeatures.Sharpness < 0 || docFeatures.Sharpness > 1 {
		t.Errorf("sharpness out of range: %f", docFeatures.Sharpness)
	}

	if len(ocrConf) != len(config.ExpectedFields) {
		t.Errorf("expected %d OCR confidences, got %d", len(config.ExpectedFields), len(ocrConf))
	}

	if len(ocrValid) != len(config.ExpectedFields) {
		t.Errorf("expected %d OCR validations, got %d", len(config.ExpectedFields), len(ocrValid))
	}
}

func TestOCRExtractor_ValidateResult(t *testing.T) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)

	// Valid result
	result := &OCRExtractionResult{
		OverallConfidence: 0.85,
		DocumentQuality: DocumentQualityResult{
			OverallScore: 0.9,
		},
		Fields: make(map[string]*ExtractedField),
	}

	issues := extractor.ValidateResult(result)
	if len(issues) > 0 {
		t.Errorf("expected no issues for valid result, got: %v", issues)
	}

	// Invalid result
	result.OverallConfidence = 1.5 // Out of range
	issues = extractor.ValidateResult(result)
	if len(issues) == 0 {
		t.Error("expected issues for invalid confidence")
	}
}

// ============================================================================
// Liveness Scorer Tests
// ============================================================================

func TestLivenessScorer_CheckLiveness_EmptyData(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	result, err := scorer.CheckLiveness(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsLive {
		t.Error("expected IsLive to be false for empty data")
	}
}

func TestLivenessScorer_CheckLiveness_StubMode(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	videoData := []byte("test video data for liveness")

	result, err := scorer.CheckLiveness(context.Background(), videoData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.LivenessScore < 0 || result.LivenessScore > 1 {
		t.Errorf("liveness score out of range: %f", result.LivenessScore)
	}

	if result.Decision != "live" && result.Decision != "spoof" && result.Decision != "uncertain" {
		t.Errorf("invalid decision: %s", result.Decision)
	}

	if result.ResultHash == "" {
		t.Error("expected result hash to be computed")
	}
}

func TestLivenessScorer_CheckLiveness_Deterministic(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	videoData := []byte("deterministic video")

	result1, _ := scorer.CheckLiveness(context.Background(), videoData)
	result2, _ := scorer.CheckLiveness(context.Background(), videoData)

	if result1.LivenessScore != result2.LivenessScore {
		t.Error("expected deterministic liveness score for same input")
	}
	if result1.ResultHash != result2.ResultHash {
		t.Error("expected deterministic result hash for same input")
	}
}

func TestLivenessScorer_ChallengeResults(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	videoData := []byte("video with challenges")

	result, _ := scorer.CheckLiveness(context.Background(), videoData)

	// Check that required challenges have results
	for _, challenge := range config.RequiredChallenges {
		if _, exists := result.ChallengeResults[challenge]; !exists {
			t.Errorf("expected result for required challenge: %s", challenge)
		}
	}
}

func TestLivenessScorer_GetChallengeScoreVector(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	videoData := []byte("video")
	result, _ := scorer.CheckLiveness(context.Background(), videoData)

	scores := scorer.GetChallengeScoreVector(result)
	if len(scores) != 6 { // 6 challenge types
		t.Errorf("expected 6 challenge scores, got %d", len(scores))
	}
}

func TestLivenessScorer_ValidateResult(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	// Valid result
	result := &LivenessResult{
		LivenessScore: 0.85,
		Confidence:    0.9,
		Decision:      "live",
	}

	issues := scorer.ValidateResult(result)
	if len(issues) > 0 {
		t.Errorf("expected no issues for valid result, got: %v", issues)
	}

	// Invalid result
	result.LivenessScore = 1.5 // Out of range
	issues = scorer.ValidateResult(result)
	if len(issues) == 0 {
		t.Error("expected issues for invalid liveness score")
	}
}

func TestLivenessScorer_SanitizeResult(t *testing.T) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)

	result := &LivenessResult{
		LivenessScore: 1.5, // Out of range
		Confidence:    -0.5, // Out of range
		Decision:      "invalid_decision",
	}

	scorer.SanitizeResult(result)

	if result.LivenessScore > 1.0 || result.LivenessScore < 0.0 {
		t.Errorf("expected clamped liveness score, got: %f", result.LivenessScore)
	}
	if result.Confidence > 1.0 || result.Confidence < 0.0 {
		t.Errorf("expected clamped confidence, got: %f", result.Confidence)
	}
	if result.Decision != "uncertain" {
		t.Errorf("expected sanitized decision 'uncertain', got: %s", result.Decision)
	}
}

// ============================================================================
// Feature Pipeline Tests
// ============================================================================

func TestFeaturePipeline_Extract_EmptyScopes(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes:         []ScopeData{},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "test-request",
	}

	output, err := pipeline.Extract(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected Success to be true even with empty scopes")
	}

	if len(output.Features) != TotalFeatureDim {
		t.Errorf("expected feature dimension %d, got %d", TotalFeatureDim, len(output.Features))
	}

	// Check for warnings about missing scopes
	if len(output.Warnings) == 0 {
		t.Error("expected warnings for missing scope types")
	}
}

func TestFeaturePipeline_Extract_WithScopes(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie-1", ScopeType: "selfie", Data: []byte("selfie image")},
			{ScopeID: "doc-1", ScopeType: "id_document", Data: []byte("document image"), Metadata: map[string]string{"document_type": "passport"}},
			{ScopeID: "video-1", ScopeType: "face_video", Data: []byte("video data")},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "test-request",
	}

	output, err := pipeline.Extract(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected Success to be true")
	}

	if len(output.Features) != TotalFeatureDim {
		t.Errorf("expected feature dimension %d, got %d", TotalFeatureDim, len(output.Features))
	}

	// Check face results
	if _, exists := output.FaceResults["selfie-1"]; !exists {
		t.Error("expected face result for selfie-1")
	}

	// Check OCR results
	if _, exists := output.OCRResults["doc-1"]; !exists {
		t.Error("expected OCR result for doc-1")
	}

	// Check liveness result
	if output.LivenessResult == nil {
		t.Error("expected liveness result")
	}

	// Check hashes
	if output.InputHash == "" {
		t.Error("expected input hash")
	}
	if output.FeatureHash == "" {
		t.Error("expected feature hash")
	}
}

func TestFeaturePipeline_Extract_Deterministic(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie-1", ScopeType: "selfie", Data: []byte("deterministic selfie")},
			{ScopeID: "doc-1", ScopeType: "id_document", Data: []byte("deterministic doc")},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Unix(1000000, 0),
		RequestID:      "test-request",
	}

	output1, _ := pipeline.Extract(context.Background(), input)
	output2, _ := pipeline.Extract(context.Background(), input)

	// Same inputs should produce same outputs
	if output1.FeatureHash != output2.FeatureHash {
		t.Error("expected deterministic feature hash")
	}

	// Compare feature vectors
	for i := range output1.Features {
		if output1.Features[i] != output2.Features[i] {
			t.Errorf("feature mismatch at index %d: %f != %f", i, output1.Features[i], output2.Features[i])
			break
		}
	}
}

func TestFeaturePipeline_ParallelExtraction(t *testing.T) {
	config := DefaultPipelineConfig()
	config.ParallelExtraction = true
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie-1", ScopeType: "selfie", Data: []byte("selfie")},
			{ScopeID: "doc-1", ScopeType: "id_document", Data: []byte("doc")},
			{ScopeID: "video-1", ScopeType: "face_video", Data: []byte("video")},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "test-parallel",
	}

	output, err := pipeline.Extract(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected parallel extraction to succeed")
	}
}

func TestFeaturePipeline_SequentialExtraction(t *testing.T) {
	config := DefaultPipelineConfig()
	config.ParallelExtraction = false
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie-1", ScopeType: "selfie", Data: []byte("selfie")},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "test-sequential",
	}

	output, err := pipeline.Extract(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected sequential extraction to succeed")
	}
}

func TestFeaturePipeline_GetStats(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "test", ScopeType: "selfie", Data: []byte("test")},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "test",
	}

	_, _ = pipeline.Extract(context.Background(), input)

	stats := pipeline.GetStats()
	if stats.PipelineRunCount != 1 {
		t.Errorf("expected pipeline run count 1, got %d", stats.PipelineRunCount)
	}
}

func TestFeaturePipeline_IsHealthy(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	if !pipeline.IsHealthy() {
		t.Error("expected stub mode pipeline to be healthy")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestPipeline_ScoreInputsGeneration(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie", ScopeType: "selfie", Data: []byte("face image")},
			{ScopeID: "doc", ScopeType: "id_document", Data: []byte("document")},
		},
		AccountAddress:   "cosmos1abc",
		BlockHeight:      12345,
		BlockTime:        time.Now(),
		RequestID:        "req-123",
		ValidatorAddress: "cosmosvaloper1xyz",
	}

	output, err := pipeline.Extract(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scoreInputs := output.ScoreInputs
	if scoreInputs == nil {
		t.Fatal("expected ScoreInputs to be set")
	}

	// Validate face embedding
	if len(scoreInputs.FaceEmbedding) != FaceEmbeddingDim {
		t.Errorf("expected face embedding dim %d, got %d", FaceEmbeddingDim, len(scoreInputs.FaceEmbedding))
	}

	// Validate metadata
	if scoreInputs.Metadata.AccountAddress != input.AccountAddress {
		t.Errorf("expected account address %s, got %s", input.AccountAddress, scoreInputs.Metadata.AccountAddress)
	}
	if scoreInputs.Metadata.BlockHeight != input.BlockHeight {
		t.Errorf("expected block height %d, got %d", input.BlockHeight, scoreInputs.Metadata.BlockHeight)
	}

	// Validate scope info
	if scoreInputs.ScopeCount != 2 {
		t.Errorf("expected scope count 2, got %d", scoreInputs.ScopeCount)
	}
}

func TestPipeline_FeatureVectorCompatibility(t *testing.T) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "test", ScopeType: "selfie", Data: []byte("test data")},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "test",
	}

	output, _ := pipeline.Extract(context.Background(), input)

	// Feature vector should be exactly TotalFeatureDim
	if len(output.Features) != TotalFeatureDim {
		t.Errorf("expected %d features, got %d", TotalFeatureDim, len(output.Features))
	}

	// First 512 should be face embedding (normalized)
	var faceNorm float64
	for i := 0; i < FaceEmbeddingDim; i++ {
		faceNorm += float64(output.Features[i]) * float64(output.Features[i])
	}
	// Face embedding should be normalized (norm ~= 1) or all zeros
	if faceNorm > 0 && (faceNorm < 0.9 || faceNorm > 1.1) {
		t.Errorf("face embedding not properly normalized, norm = %f", faceNorm)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkFaceExtractor_ExtractFromSelfie(b *testing.B) {
	config := DefaultFaceExtractorConfig()
	extractor := NewFaceExtractor(config)
	imageData := make([]byte, 1024*1024) // 1MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFromSelfie(context.Background(), imageData)
	}
}

func BenchmarkOCRExtractor_Extract(b *testing.B) {
	config := DefaultOCRExtractorConfig()
	extractor := NewOCRExtractor(config)
	imageData := make([]byte, 1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.Extract(context.Background(), imageData, "passport")
	}
}

func BenchmarkLivenessScorer_CheckLiveness(b *testing.B) {
	config := DefaultLivenessScorerConfig()
	scorer := NewLivenessScorer(config)
	videoData := make([]byte, 1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scorer.CheckLiveness(context.Background(), videoData)
	}
}

func BenchmarkFeaturePipeline_Extract(b *testing.B) {
	config := DefaultPipelineConfig()
	pipeline := NewFeaturePipeline(config)

	input := &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie", ScopeType: "selfie", Data: make([]byte, 1024*100)},
			{ScopeID: "doc", ScopeType: "id_document", Data: make([]byte, 1024*100)},
			{ScopeID: "video", ScopeType: "face_video", Data: make([]byte, 1024*100)},
		},
		AccountAddress: "cosmos1test",
		BlockHeight:    100,
		BlockTime:      time.Now(),
		RequestID:      "bench",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pipeline.Extract(context.Background(), input)
	}
}
