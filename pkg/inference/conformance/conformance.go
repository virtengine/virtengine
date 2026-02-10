// Package conformance provides a deterministic conformance test framework
// for verifying that the identity verification pipeline produces identical
// outputs across different validator machines.
//
// VE-219: Deterministic identity verification runtime
package conformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ============================================================================
// Test Vector Types
// ============================================================================

// TestVector represents a single conformance test case
type TestVector struct {
	// ID is the unique identifier for this test vector
	ID string `json:"id"`

	// Name is a human-readable name for the test
	Name string `json:"name"`

	// Description describes what this test verifies
	Description string `json:"description"`

	// Category is the test category (face_detection, ocr, scoring, etc.)
	Category TestCategory `json:"category"`

	// InputHash is the SHA256 hash of the test input
	InputHash string `json:"input_hash"`

	// ExpectedOutputHash is the expected SHA256 hash of the output
	ExpectedOutputHash string `json:"expected_output_hash"`

	// ExpectedScore is the expected identity score (if applicable)
	ExpectedScore *uint32 `json:"expected_score,omitempty"`

	// IntermediateHashes maps pipeline stage to expected hash
	IntermediateHashes map[string]string `json:"intermediate_hashes,omitempty"`

	// PipelineVersion is the version this test is valid for
	PipelineVersion string `json:"pipeline_version"`

	// Tolerance is the allowed tolerance for floating-point comparisons
	Tolerance float64 `json:"tolerance,omitempty"`
}

// TestCategory represents the category of a conformance test
type TestCategory string

const (
	TestCategoryFaceDetection    TestCategory = "face_detection"
	TestCategoryFaceRecognition  TestCategory = "face_recognition"
	TestCategoryFaceVerification TestCategory = "face_verification"
	TestCategoryTextDetection    TestCategory = "text_detection"
	TestCategoryOCR              TestCategory = "ocr"
	TestCategoryDocumentQuality  TestCategory = "document_quality"
	TestCategoryFaceExtraction   TestCategory = "face_extraction"
	TestCategoryIdentityScoring  TestCategory = "identity_scoring"
	TestCategoryEndToEnd         TestCategory = "end_to_end"
)

// ============================================================================
// Test Result Types
// ============================================================================

// TestResult represents the result of running a conformance test
type TestResult struct {
	// VectorID is the ID of the test vector
	VectorID string `json:"vector_id"`

	// Passed indicates if the test passed
	Passed bool `json:"passed"`

	// ActualOutputHash is the actual output hash produced
	ActualOutputHash string `json:"actual_output_hash"`

	// ActualScore is the actual score produced (if applicable)
	ActualScore *uint32 `json:"actual_score,omitempty"`

	// ActualIntermediateHashes maps stage to actual hash
	ActualIntermediateHashes map[string]string `json:"actual_intermediate_hashes,omitempty"`

	// Differences contains descriptions of any differences
	Differences []string `json:"differences,omitempty"`

	// ExecutionTimeMs is how long the test took in milliseconds
	ExecutionTimeMs int64 `json:"execution_time_ms"`

	// ValidatorInfo contains information about the validator
	ValidatorInfo ValidatorInfo `json:"validator_info"`

	// Timestamp is when the test was run
	Timestamp time.Time `json:"timestamp"`
}

// ValidatorInfo contains information about the validator running tests
type ValidatorInfo struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// PipelineVersion is the pipeline version being used
	PipelineVersion string `json:"pipeline_version"`

	// PipelineImageHash is the OCI image hash
	PipelineImageHash string `json:"pipeline_image_hash"`

	// ModelManifestHash is the model manifest hash
	ModelManifestHash string `json:"model_manifest_hash"`

	// Hostname is the machine hostname
	Hostname string `json:"hostname,omitempty"`

	// Platform is the OS/architecture
	Platform string `json:"platform,omitempty"`
}

// ============================================================================
// Conformance Test Suite
// ============================================================================

// TestSuite represents a collection of conformance tests
type TestSuite struct {
	// Name is the name of the test suite
	Name string `json:"name"`

	// Version is the version of the test suite
	Version string `json:"version"`

	// PipelineVersion is the pipeline version these tests are for
	PipelineVersion string `json:"pipeline_version"`

	// Vectors contains all test vectors
	Vectors []TestVector `json:"vectors"`

	// CreatedAt is when the suite was created
	CreatedAt time.Time `json:"created_at"`

	// SuiteHash is the SHA256 hash of the entire suite for integrity
	SuiteHash string `json:"suite_hash"`
}

// ============================================================================
// Test Suite Runner
// ============================================================================

// TestRunner runs conformance tests
type TestRunner struct {
	// pipelineVersion is the pipeline version being tested
	pipelineVersion string

	// validatorInfo contains validator metadata
	validatorInfo ValidatorInfo

	// testDataDir is the directory containing test data
	testDataDir string

	// pipelineRunner executes the ML pipeline
	pipelineRunner PipelineRunner
}

// PipelineRunner is the interface for running the ML pipeline
type PipelineRunner interface {
	// RunPipeline runs the full pipeline on input data
	RunPipeline(inputData []byte) (*PipelineOutput, error)

	// RunStage runs a specific pipeline stage
	RunStage(stage string, inputData []byte) ([]byte, error)

	// GetVersion returns the pipeline version
	GetVersion() string

	// GetModelHashes returns all model weight hashes
	GetModelHashes() map[string]string
}

// PipelineOutput contains the output from running the pipeline
type PipelineOutput struct {
	// Score is the computed identity score
	Score uint32 `json:"score"`

	// OutputHash is the hash of the final output
	OutputHash string `json:"output_hash"`

	// IntermediateHashes contains hashes for each stage
	IntermediateHashes map[string]string `json:"intermediate_hashes"`

	// RawOutput contains the raw output data
	RawOutput []byte `json:"raw_output,omitempty"`
}

// NewTestRunner creates a new test runner
func NewTestRunner(
	pipelineVersion string,
	validatorInfo ValidatorInfo,
	testDataDir string,
	pipelineRunner PipelineRunner,
) *TestRunner {
	return &TestRunner{
		pipelineVersion: pipelineVersion,
		validatorInfo:   validatorInfo,
		testDataDir:     testDataDir,
		pipelineRunner:  pipelineRunner,
	}
}

// ============================================================================
// Test Execution
// ============================================================================

// RunSuite runs all tests in a test suite
func (tr *TestRunner) RunSuite(suite *TestSuite) (*SuiteResult, error) {
	if suite.PipelineVersion != tr.pipelineVersion {
		return nil, fmt.Errorf(
			"pipeline version mismatch: suite requires %s, running %s",
			suite.PipelineVersion,
			tr.pipelineVersion,
		)
	}

	results := make([]TestResult, 0, len(suite.Vectors))
	passed := 0
	failed := 0

	for _, vector := range suite.Vectors {
		result := tr.RunTest(&vector)
		results = append(results, *result)

		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	return &SuiteResult{
		SuiteName:       suite.Name,
		SuiteVersion:    suite.Version,
		PipelineVersion: suite.PipelineVersion,
		TotalTests:      len(suite.Vectors),
		Passed:          passed,
		Failed:          failed,
		Results:         results,
		ValidatorInfo:   tr.validatorInfo,
		Timestamp:       time.Now().UTC(),
	}, nil
}

// RunTest runs a single test vector
func (tr *TestRunner) RunTest(vector *TestVector) *TestResult {
	startTime := time.Now()

	result := &TestResult{
		VectorID:                 vector.ID,
		Passed:                   true,
		ActualIntermediateHashes: make(map[string]string),
		Differences:              make([]string, 0),
		ValidatorInfo:            tr.validatorInfo,
		Timestamp:                startTime,
	}

	// Load test input data
	inputData, err := tr.loadTestInput(vector.ID)
	if err != nil {
		result.Passed = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("failed to load test input: %v", err))
		result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Verify input hash
	inputHash := computeHash(inputData)
	if inputHash != vector.InputHash {
		result.Passed = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("input hash mismatch: expected %s, got %s",
				vector.InputHash, inputHash))
		result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Run the pipeline
	output, err := tr.pipelineRunner.RunPipeline(inputData)
	if err != nil {
		result.Passed = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("pipeline execution failed: %v", err))
		result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return result
	}

	result.ActualOutputHash = output.OutputHash
	result.ActualScore = &output.Score
	result.ActualIntermediateHashes = output.IntermediateHashes

	// Compare output hash
	if output.OutputHash != vector.ExpectedOutputHash {
		result.Passed = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("output hash mismatch: expected %s, got %s",
				vector.ExpectedOutputHash, output.OutputHash))
	}

	// Compare score if expected
	if vector.ExpectedScore != nil && output.Score != *vector.ExpectedScore {
		result.Passed = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("score mismatch: expected %d, got %d",
				*vector.ExpectedScore, output.Score))
	}

	// Compare intermediate hashes
	for stage, expectedHash := range vector.IntermediateHashes {
		actualHash, ok := output.IntermediateHashes[stage]
		if !ok {
			result.Passed = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("missing intermediate hash for stage: %s", stage))
			continue
		}

		if actualHash != expectedHash {
			result.Passed = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("intermediate hash mismatch for %s: expected %s, got %s",
					stage, expectedHash, actualHash))
		}
	}

	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return result
}

// loadTestInput loads test input data from the test data directory
func (tr *TestRunner) loadTestInput(vectorID string) ([]byte, error) {
	path := filepath.Join(tr.testDataDir, vectorID+".bin")
	//nolint:gosec // G304: path is constructed from trusted test data directory
	return os.ReadFile(path)
}

// computeHash computes SHA256 hash of data
func computeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ============================================================================
// Suite Result
// ============================================================================

// SuiteResult contains the results of running a test suite
type SuiteResult struct {
	// SuiteName is the name of the test suite
	SuiteName string `json:"suite_name"`

	// SuiteVersion is the version of the test suite
	SuiteVersion string `json:"suite_version"`

	// PipelineVersion is the pipeline version tested
	PipelineVersion string `json:"pipeline_version"`

	// TotalTests is the total number of tests
	TotalTests int `json:"total_tests"`

	// Passed is the number of passed tests
	Passed int `json:"passed"`

	// Failed is the number of failed tests
	Failed int `json:"failed"`

	// Results contains individual test results
	Results []TestResult `json:"results"`

	// ValidatorInfo contains validator metadata
	ValidatorInfo ValidatorInfo `json:"validator_info"`

	// Timestamp is when the suite was run
	Timestamp time.Time `json:"timestamp"`
}

// AllPassed returns true if all tests passed
func (sr *SuiteResult) AllPassed() bool {
	return sr.Failed == 0
}

// ToJSON returns the result as JSON
func (sr *SuiteResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sr, "", "  ")
}

// ============================================================================
// Test Suite Loading
// ============================================================================

// LoadTestSuite loads a test suite from a JSON file
func LoadTestSuite(path string) (*TestSuite, error) {
	//nolint:gosec // G304: path is provided by trusted caller/configuration
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite: %w", err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse test suite: %w", err)
	}

	// Verify suite integrity
	expectedHash := suite.SuiteHash
	suite.SuiteHash = "" // Clear for hash computation
	suiteData, err := json.Marshal(suite)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suite for hash: %w", err)
	}
	actualHash := computeHash(suiteData)

	if expectedHash != "" && actualHash != expectedHash {
		return nil, fmt.Errorf("suite integrity check failed: hash mismatch")
	}

	suite.SuiteHash = expectedHash
	return &suite, nil
}

// ============================================================================
// Default Test Vectors
// ============================================================================

// GetDefaultTestSuite returns the default conformance test suite
func GetDefaultTestSuite(pipelineVersion string) *TestSuite {
	score75 := uint32(75)
	score85 := uint32(85)
	score50 := uint32(50)

	vectors := []TestVector{
		{
			ID:                 "face_detect_001",
			Name:               "Standard Face Detection",
			Description:        "Verify face detection on a clear portrait image",
			Category:           TestCategoryFaceDetection,
			InputHash:          "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			ExpectedOutputHash: "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
			PipelineVersion:    pipelineVersion,
		},
		{
			ID:                 "face_embed_001",
			Name:               "Face Embedding Generation",
			Description:        "Verify face embedding is deterministic",
			Category:           TestCategoryFaceRecognition,
			InputHash:          "c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			ExpectedOutputHash: "d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5",
			PipelineVersion:    pipelineVersion,
			IntermediateHashes: map[string]string{
				"face_detect":    "e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
				"face_align":     "f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1",
				"face_embedding": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			},
		},
		{
			ID:                 "ocr_extract_001",
			Name:               "OCR Field Extraction",
			Description:        "Verify OCR produces identical text extraction",
			Category:           TestCategoryOCR,
			InputHash:          "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ExpectedOutputHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			PipelineVersion:    pipelineVersion,
		},
		{
			ID:                 "text_detect_001",
			Name:               "Text Region Detection (CRAFT)",
			Description:        "Verify CRAFT text detection is deterministic",
			Category:           TestCategoryTextDetection,
			InputHash:          "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			ExpectedOutputHash: "0987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba",
			PipelineVersion:    pipelineVersion,
		},
		{
			ID:                 "unet_extract_001",
			Name:               "U-Net Face Extraction",
			Description:        "Verify U-Net face extraction from ID document",
			Category:           TestCategoryFaceExtraction,
			InputHash:          "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00",
			ExpectedOutputHash: "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899",
			PipelineVersion:    pipelineVersion,
		},
		{
			ID:                 "e2e_score_001",
			Name:               "End-to-End Scoring - High Quality",
			Description:        "Full pipeline with high-quality inputs should produce score 85",
			Category:           TestCategoryEndToEnd,
			InputHash:          "deadbeef00000000deadbeef00000000deadbeef00000000deadbeef00000000",
			ExpectedOutputHash: "00000000deadbeef00000000deadbeef00000000deadbeef00000000deadbeef",
			ExpectedScore:      &score85,
			PipelineVersion:    pipelineVersion,
		},
		{
			ID:                 "e2e_score_002",
			Name:               "End-to-End Scoring - Medium Quality",
			Description:        "Full pipeline with medium-quality inputs should produce score 75",
			Category:           TestCategoryEndToEnd,
			InputHash:          "cafebabe11111111cafebabe11111111cafebabe11111111cafebabe11111111",
			ExpectedOutputHash: "11111111cafebabe11111111cafebabe11111111cafebabe11111111cafebabe",
			ExpectedScore:      &score75,
			PipelineVersion:    pipelineVersion,
		},
		{
			ID:                 "e2e_score_003",
			Name:               "End-to-End Scoring - Low Quality",
			Description:        "Full pipeline with low-quality inputs should produce score 50",
			Category:           TestCategoryEndToEnd,
			InputHash:          "badf00d022222222badf00d022222222badf00d022222222badf00d022222222",
			ExpectedOutputHash: "22222222badf00d022222222badf00d022222222badf00d022222222badf00d0",
			ExpectedScore:      &score50,
			PipelineVersion:    pipelineVersion,
		},
	}

	suite := &TestSuite{
		Name:            "VirtEngine VEID Conformance Suite",
		Version:         "1.0.0",
		PipelineVersion: pipelineVersion,
		Vectors:         vectors,
		CreatedAt:       time.Now().UTC(),
	}

	// Compute suite hash
	suiteData, err := json.Marshal(suite)
	if err == nil {
		suite.SuiteHash = computeHash(suiteData)
	}

	return suite
}
