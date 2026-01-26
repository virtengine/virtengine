package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockPipelineRunner implements PipelineRunner for testing
type MockPipelineRunner struct {
	version    string
	modelHashes map[string]string
	
	// Configurable outputs for testing
	shouldFail    bool
	outputHash    string
	score         uint32
	intermediates map[string]string
}

func NewMockPipelineRunner(version string) *MockPipelineRunner {
	return &MockPipelineRunner{
		version: version,
		modelHashes: map[string]string{
			"deepface": "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			"craft":    "sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
			"unet":     "sha256:c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
		},
		outputHash:    "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
		score:         75,
		intermediates: make(map[string]string),
	}
}

func (m *MockPipelineRunner) RunPipeline(inputData []byte) (*PipelineOutput, error) {
	if m.shouldFail {
		return nil, &mockError{"pipeline execution failed"}
	}
	
	return &PipelineOutput{
		Score:              m.score,
		OutputHash:         m.outputHash,
		IntermediateHashes: m.intermediates,
		RawOutput:          inputData,
	}, nil
}

func (m *MockPipelineRunner) RunStage(stage string, inputData []byte) ([]byte, error) {
	if m.shouldFail {
		return nil, &mockError{"stage execution failed"}
	}
	return inputData, nil
}

func (m *MockPipelineRunner) GetVersion() string {
	return m.version
}

func (m *MockPipelineRunner) GetModelHashes() map[string]string {
	return m.modelHashes
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// TestTestVectorCreation tests creating test vectors
func TestTestVectorCreation(t *testing.T) {
	score := uint32(85)
	vector := TestVector{
		ID:                 "test_001",
		Name:               "Test Vector",
		Description:        "A test vector",
		Category:           TestCategoryFaceDetection,
		InputHash:          "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		ExpectedOutputHash: "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
		ExpectedScore:      &score,
		PipelineVersion:    "1.0.0",
	}
	
	if vector.ID != "test_001" {
		t.Errorf("expected ID test_001, got %s", vector.ID)
	}
	
	if vector.Category != TestCategoryFaceDetection {
		t.Errorf("expected category %s, got %s", TestCategoryFaceDetection, vector.Category)
	}
	
	if *vector.ExpectedScore != 85 {
		t.Errorf("expected score 85, got %d", *vector.ExpectedScore)
	}
}

// TestTestSuiteCreation tests creating a test suite
func TestTestSuiteCreation(t *testing.T) {
	suite := GetDefaultTestSuite("1.0.0")
	
	if suite.Name != "VirtEngine VEID Conformance Suite" {
		t.Errorf("unexpected suite name: %s", suite.Name)
	}
	
	if suite.PipelineVersion != "1.0.0" {
		t.Errorf("expected pipeline version 1.0.0, got %s", suite.PipelineVersion)
	}
	
	if len(suite.Vectors) == 0 {
		t.Error("expected vectors in suite")
	}
	
	if suite.SuiteHash == "" {
		t.Error("expected suite hash to be computed")
	}
}

// TestTestRunnerCreation tests creating a test runner
func TestTestRunnerCreation(t *testing.T) {
	runner := NewMockPipelineRunner("1.0.0")
	
	validatorInfo := ValidatorInfo{
		ValidatorAddress:  "validator1",
		PipelineVersion:   "1.0.0",
		PipelineImageHash: "sha256:abc123",
		ModelManifestHash: "manifesthash",
	}
	
	testRunner := NewTestRunner("1.0.0", validatorInfo, "/tmp/test", runner)
	
	if testRunner.pipelineVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", testRunner.pipelineVersion)
	}
	
	if testRunner.validatorInfo.ValidatorAddress != "validator1" {
		t.Errorf("expected validator1, got %s", testRunner.validatorInfo.ValidatorAddress)
	}
}

// TestVersionMismatch tests that version mismatch is detected
func TestVersionMismatch(t *testing.T) {
	runner := NewMockPipelineRunner("1.0.0")
	
	validatorInfo := ValidatorInfo{
		ValidatorAddress: "validator1",
		PipelineVersion:  "1.0.0",
	}
	
	testRunner := NewTestRunner("1.0.0", validatorInfo, "/tmp/test", runner)
	
	// Create suite with different version
	suite := &TestSuite{
		Name:            "Test Suite",
		Version:         "1.0.0",
		PipelineVersion: "2.0.0", // Different version
		Vectors:         []TestVector{},
	}
	
	_, err := testRunner.RunSuite(suite)
	if err == nil {
		t.Error("expected error for version mismatch")
	}
}

// TestHashComputation tests hash computation is deterministic
func TestHashComputation(t *testing.T) {
	data := []byte("test data for hashing")
	
	hash1 := computeHash(data)
	hash2 := computeHash(data)
	
	if hash1 != hash2 {
		t.Error("hash computation should be deterministic")
	}
	
	// Different data should produce different hash
	differentData := []byte("different data")
	hash3 := computeHash(differentData)
	
	if hash1 == hash3 {
		t.Error("different data should produce different hash")
	}
}

// TestSuiteResultAllPassed tests AllPassed helper
func TestSuiteResultAllPassed(t *testing.T) {
	result := &SuiteResult{
		TotalTests: 5,
		Passed:     5,
		Failed:     0,
	}
	
	if !result.AllPassed() {
		t.Error("expected AllPassed to return true")
	}
	
	result.Failed = 1
	if result.AllPassed() {
		t.Error("expected AllPassed to return false")
	}
}

// TestSuiteResultToJSON tests JSON serialization
func TestSuiteResultToJSON(t *testing.T) {
	result := &SuiteResult{
		SuiteName:       "Test Suite",
		SuiteVersion:    "1.0.0",
		PipelineVersion: "1.0.0",
		TotalTests:      3,
		Passed:          2,
		Failed:          1,
		Timestamp:       time.Now().UTC(),
	}
	
	jsonData, err := result.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize to JSON: %v", err)
	}
	
	// Verify it can be deserialized
	var parsed SuiteResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	
	if parsed.SuiteName != result.SuiteName {
		t.Errorf("expected suite name %s, got %s", result.SuiteName, parsed.SuiteName)
	}
}

// TestLoadTestSuite tests loading a test suite from file
func TestLoadTestSuite(t *testing.T) {
	// Create a temporary test suite file
	tmpDir := t.TempDir()
	suitePath := filepath.Join(tmpDir, "test_suite.json")
	
	suite := GetDefaultTestSuite("1.0.0")
	suiteData, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal suite: %v", err)
	}
	
	if err := os.WriteFile(suitePath, suiteData, 0644); err != nil {
		t.Fatalf("failed to write suite file: %v", err)
	}
	
	// Load the suite
	loaded, err := LoadTestSuite(suitePath)
	if err != nil {
		t.Fatalf("failed to load suite: %v", err)
	}
	
	if loaded.Name != suite.Name {
		t.Errorf("expected name %s, got %s", suite.Name, loaded.Name)
	}
	
	if len(loaded.Vectors) != len(suite.Vectors) {
		t.Errorf("expected %d vectors, got %d", len(suite.Vectors), len(loaded.Vectors))
	}
}

// TestLoadTestSuiteNotFound tests loading non-existent suite
func TestLoadTestSuiteNotFound(t *testing.T) {
	_, err := LoadTestSuite("/nonexistent/path/suite.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// TestTestCategories tests all test categories are defined
func TestTestCategories(t *testing.T) {
	categories := []TestCategory{
		TestCategoryFaceDetection,
		TestCategoryFaceRecognition,
		TestCategoryFaceVerification,
		TestCategoryTextDetection,
		TestCategoryOCR,
		TestCategoryDocumentQuality,
		TestCategoryFaceExtraction,
		TestCategoryIdentityScoring,
		TestCategoryEndToEnd,
	}
	
	if len(categories) != 9 {
		t.Errorf("expected 9 categories, got %d", len(categories))
	}
	
	// Verify each category is a non-empty string
	for _, cat := range categories {
		if string(cat) == "" {
			t.Error("category should not be empty")
		}
	}
}

// TestValidatorInfo tests ValidatorInfo struct
func TestValidatorInfo(t *testing.T) {
	info := ValidatorInfo{
		ValidatorAddress:  "cosmosvaloper1...",
		PipelineVersion:   "1.0.0",
		PipelineImageHash: "sha256:abc123def456",
		ModelManifestHash: "manifesthash123",
		Hostname:          "validator-node-1",
		Platform:          "linux/amd64",
	}
	
	if info.ValidatorAddress == "" {
		t.Error("validator address should not be empty")
	}
	
	if info.Platform != "linux/amd64" {
		t.Errorf("expected platform linux/amd64, got %s", info.Platform)
	}
}

// TestTestResultDifferences tests recording differences
func TestTestResultDifferences(t *testing.T) {
	result := &TestResult{
		VectorID:    "test_001",
		Passed:      false,
		Differences: []string{
			"output hash mismatch",
			"score mismatch",
		},
	}
	
	if len(result.Differences) != 2 {
		t.Errorf("expected 2 differences, got %d", len(result.Differences))
	}
	
	if result.Passed {
		t.Error("result with differences should not pass")
	}
}

// TestPipelineOutputHashing tests that pipeline outputs are hashable
func TestPipelineOutputHashing(t *testing.T) {
	output := &PipelineOutput{
		Score:      85,
		OutputHash: "abc123",
		IntermediateHashes: map[string]string{
			"stage1": "hash1",
			"stage2": "hash2",
		},
		RawOutput: []byte("raw data"),
	}
	
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal output: %v", err)
	}
	
	hash1 := computeHash(data)
	hash2 := computeHash(data)
	
	if hash1 != hash2 {
		t.Error("output hashing should be deterministic")
	}
}

// TestDefaultTestSuiteVectors tests default vectors are valid
func TestDefaultTestSuiteVectors(t *testing.T) {
	suite := GetDefaultTestSuite("1.0.0")
	
	for _, vector := range suite.Vectors {
		// Verify required fields
		if vector.ID == "" {
			t.Error("vector ID should not be empty")
		}
		
		if vector.Name == "" {
			t.Error("vector Name should not be empty")
		}
		
		if vector.InputHash == "" {
			t.Error("vector InputHash should not be empty")
		}
		
		if vector.ExpectedOutputHash == "" {
			t.Error("vector ExpectedOutputHash should not be empty")
		}
		
		if vector.PipelineVersion == "" {
			t.Error("vector PipelineVersion should not be empty")
		}
		
		// Verify hash format (64 hex chars)
		if len(vector.InputHash) != 64 {
			t.Errorf("invalid input hash length for %s: %d", vector.ID, len(vector.InputHash))
		}
		
		if len(vector.ExpectedOutputHash) != 64 {
			t.Errorf("invalid expected output hash length for %s: %d", 
				vector.ID, len(vector.ExpectedOutputHash))
		}
	}
}
