package inference

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

type featureParityFixture struct {
	Schema              string `json:"schema"`
	Version             int    `json:"version"`
	Contract            string `json:"contract"`
	FeatureSchemaSHA256 string `json:"feature_schema_sha256"`
	Layout              struct {
		TotalDimension int `json:"total_dimension"`
		Components     []struct {
			Name      string   `json:"name"`
			Offset    int      `json:"offset"`
			Dimension int      `json:"dimension"`
			Fields    []string `json:"fields"`
		} `json:"components"`
		Encoding struct {
			ValueType            string `json:"value_type"`
			ByteOrder            string `json:"byte_order"`
			PreHashDecimalPlaces int    `json:"pre_hash_decimal_places"`
			Rounding             string `json:"rounding"`
			HashAlgorithm        string `json:"hash_algorithm"`
		} `json:"encoding"`
	} `json:"layout"`
	Cases []struct {
		Name    string `json:"name"`
		Profile string `json:"profile"`
		Source  struct {
			FaceEmbedding struct {
				Kind      string  `json:"kind"`
				Dimension int     `json:"dimension"`
				Index     int     `json:"index"`
				Value     float32 `json:"value"`
				Low       float32 `json:"low"`
				High      float32 `json:"high"`
				Start     float32 `json:"start"`
				Step      float32 `json:"step"`
			} `json:"face_embedding"`
			FaceConfidence  float32 `json:"face_confidence"`
			DocQualityScore float32 `json:"doc_quality_score"`
			DocumentQuality struct {
				Sharpness  float32 `json:"sharpness"`
				Brightness float32 `json:"brightness"`
				Contrast   float32 `json:"contrast"`
				NoiseLevel float32 `json:"noise_level"`
				BlurScore  float32 `json:"blur_score"`
			} `json:"document_quality"`
			OCRConfidences     map[string]float32 `json:"ocr_confidences"`
			OCRFieldValidation map[string]bool    `json:"ocr_field_validation"`
			ScopeTypes         []string           `json:"scope_types"`
			ScopeCount         int                `json:"scope_count"`
			BlockHeight        int64              `json:"block_height"`
		} `json:"source"`
		ExpectedVectorHash       string             `json:"expected_vector_hash"`
		ExpectedRawVectorHash    string             `json:"expected_raw_vector_hash"`
		ExpectedNonzeroPositions map[string]float32 `json:"expected_nonzero_positions"`
	} `json:"cases"`
}

func TestCanonicalFeatureParityFixture(t *testing.T) {
	data, err := os.ReadFile("conformance/testdata/feature_parity_v1.json")
	if err != nil {
		t.Fatalf("read feature parity fixture: %v", err)
	}
	var fixture featureParityFixture
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode feature parity fixture: %v", err)
	}
	validateFeatureParityLayout(t, &fixture)

	hasher := NewDeterminismController(42, true)
	for index := range fixture.Cases {
		testCase := &fixture.Cases[index]
		t.Run(testCase.Name, func(t *testing.T) {
			config := DefaultFeatureExtractorConfig()
			if testCase.Profile == "production" {
				config = ProductionFeatureExtractorConfig()
			} else if testCase.Profile != "development" {
				t.Fatalf("unsupported fixture profile %q", testCase.Profile)
			}
			features, err := NewFeatureExtractor(config).ExtractFeatures(featureParityInputs(t, testCase))
			if err != nil {
				t.Fatalf("extract canonical features: %v", err)
			}
			if got := hasher.ComputeFeatureHash(features); got != testCase.ExpectedVectorHash {
				t.Fatalf("feature hash: got %s, want %s", got, testCase.ExpectedVectorHash)
			}
			if got := rawFeatureVectorHash(features); got != testCase.ExpectedRawVectorHash {
				t.Fatalf("raw feature hash: got %s, want %s", got, testCase.ExpectedRawVectorHash)
			}
			for position, value := range features {
				expected, nonzero := testCase.ExpectedNonzeroPositions[strconv.Itoa(position)]
				if nonzero && value != expected {
					t.Errorf("position %d: got %v, want %v", position, value, expected)
				} else if !nonzero && value != 0 {
					t.Errorf("position %d: unexpected nonzero %v", position, value)
				}
			}
		})
	}
}

func rawFeatureVectorHash(features []float32) string {
	hasher := sha256.New()
	var encoded [4]byte
	for _, value := range features {
		binary.BigEndian.PutUint32(encoded[:], math.Float32bits(value))
		_, _ = hasher.Write(encoded[:])
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func validateFeatureParityLayout(t *testing.T, fixture *featureParityFixture) {
	t.Helper()
	if fixture.Schema != "virtengine.inference.feature_parity" || fixture.Version != 1 || fixture.Contract != "fixture_only" {
		t.Fatalf("unsupported feature parity contract")
	}
	schemaBytes, err := os.ReadFile("schema/trust_score_features_v1.json")
	if err != nil {
		t.Fatalf("read canonical schema: %v", err)
	}
	if !bytes.Equal(schemaBytes, canonicalFeatureSchemaJSON) {
		t.Fatal("embedded schema bytes differ from repository artifact")
	}
	digest := sha256.Sum256(schemaBytes)
	if got := hex.EncodeToString(digest[:]); got != fixture.FeatureSchemaSHA256 {
		t.Fatalf("schema hash: got %s, want %s", got, fixture.FeatureSchemaSHA256)
	}
	expected := []struct {
		name              string
		offset, dimension int
	}{
		{"selfie_embedding", SelfieEmbeddingOffset, FaceEmbeddingDim},
		{"face_confidence", FaceConfidenceOffset, FaceConfidenceDim},
		{"document_quality", DocQualityOffset, DocQualityDim},
		{"ocr", OCROffset, OCRFieldsDim},
		{"metadata", MetadataOffset, MetadataDim},
		{"reserved", ReservedOffset, PaddingDim},
	}
	if fixture.Layout.TotalDimension != TotalFeatureDim || len(fixture.Layout.Components) != len(expected) {
		t.Fatal("fixture layout dimensions do not match serving")
	}
	for index, want := range expected {
		got := fixture.Layout.Components[index]
		if got.Name != want.name || got.Offset != want.offset || got.Dimension != want.dimension {
			t.Errorf("component %d: got %s/%d/%d, want %s/%d/%d", index, got.Name, got.Offset, got.Dimension, want.name, want.offset, want.dimension)
		}
	}
	encoding := fixture.Layout.Encoding
	if encoding.ValueType != "ieee754-float32" || encoding.ByteOrder != "big-endian" || encoding.PreHashDecimalPlaces != 6 || encoding.Rounding != "half_away_from_zero" || encoding.HashAlgorithm != "sha256" {
		t.Fatalf("unsupported fixture encoding: %+v", encoding)
	}
}

func featureParityInputs(t *testing.T, testCase *struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
	Source  struct {
		FaceEmbedding struct {
			Kind      string  `json:"kind"`
			Dimension int     `json:"dimension"`
			Index     int     `json:"index"`
			Value     float32 `json:"value"`
			Low       float32 `json:"low"`
			High      float32 `json:"high"`
			Start     float32 `json:"start"`
			Step      float32 `json:"step"`
		} `json:"face_embedding"`
		FaceConfidence  float32 `json:"face_confidence"`
		DocQualityScore float32 `json:"doc_quality_score"`
		DocumentQuality struct {
			Sharpness  float32 `json:"sharpness"`
			Brightness float32 `json:"brightness"`
			Contrast   float32 `json:"contrast"`
			NoiseLevel float32 `json:"noise_level"`
			BlurScore  float32 `json:"blur_score"`
		} `json:"document_quality"`
		OCRConfidences     map[string]float32 `json:"ocr_confidences"`
		OCRFieldValidation map[string]bool    `json:"ocr_field_validation"`
		ScopeTypes         []string           `json:"scope_types"`
		ScopeCount         int                `json:"scope_count"`
		BlockHeight        int64              `json:"block_height"`
	} `json:"source"`
	ExpectedVectorHash       string             `json:"expected_vector_hash"`
	ExpectedRawVectorHash    string             `json:"expected_raw_vector_hash"`
	ExpectedNonzeroPositions map[string]float32 `json:"expected_nonzero_positions"`
}) *ScoreInputs {
	t.Helper()
	source := testCase.Source
	var embedding []float32
	switch source.FaceEmbedding.Kind {
	case "missing":
	case "single_index":
		embedding = make([]float32, source.FaceEmbedding.Dimension)
		embedding[source.FaceEmbedding.Index] = source.FaceEmbedding.Value
	case "alternating_bounds":
		embedding = make([]float32, source.FaceEmbedding.Dimension)
		for index := range embedding {
			if index%2 == 0 {
				embedding[index] = source.FaceEmbedding.Low
			} else {
				embedding[index] = source.FaceEmbedding.High
			}
		}
	case "asymmetric_sequence":
		embedding = make([]float32, source.FaceEmbedding.Dimension)
		for index := range embedding {
			embedding[index] = source.FaceEmbedding.Start + float32(index)*source.FaceEmbedding.Step
		}
	default:
		t.Fatalf("unsupported face source kind %q", source.FaceEmbedding.Kind)
	}
	return &ScoreInputs{
		FaceEmbedding: embedding, FaceConfidence: source.FaceConfidence, DocQualityScore: source.DocQualityScore,
		DocQualityFeatures: DocQualityFeatures{Sharpness: source.DocumentQuality.Sharpness, Brightness: source.DocumentQuality.Brightness, Contrast: source.DocumentQuality.Contrast, NoiseLevel: source.DocumentQuality.NoiseLevel, BlurScore: source.DocumentQuality.BlurScore},
		OCRConfidences:     source.OCRConfidences, OCRFieldValidation: source.OCRFieldValidation,
		ScopeTypes: source.ScopeTypes, ScopeCount: source.ScopeCount, Metadata: InferenceMetadata{BlockHeight: source.BlockHeight},
	}
}

func TestProductionFeatureExtractorRejectsIncompleteOrInvalidInputs(t *testing.T) {
	valid := func() *ScoreInputs {
		inputs := &ScoreInputs{
			FaceEmbedding: make([]float32, FaceEmbeddingDim), FaceConfidence: 0.5, DocQualityScore: 0.5,
			DocQualityFeatures: DocQualityFeatures{Sharpness: 0.5, Brightness: 0.5, Contrast: 0.5, NoiseLevel: 0.5, BlurScore: 0.5},
			OCRConfidences:     make(map[string]float32), OCRFieldValidation: make(map[string]bool), ScopeTypes: []string{}, ScopeCount: 1,
		}
		inputs.FaceEmbedding[0] = 1
		for _, field := range OCRFieldNames {
			inputs.OCRConfidences[field] = 0.5
			inputs.OCRFieldValidation[field] = true
		}
		return inputs
	}
	tests := []struct {
		name, want string
		mutate     func(*ScoreInputs)
	}{
		{"missing face", "missing face embedding", func(inputs *ScoreInputs) { inputs.FaceEmbedding = nil }},
		{"zero norm face", "nonzero norm", func(inputs *ScoreInputs) { inputs.FaceEmbedding[0] = 0 }},
		{"missing OCR field", "missing OCR confidence", func(inputs *ScoreInputs) { delete(inputs.OCRConfidences, "name") }},
		{"nonfinite", "must be finite", func(inputs *ScoreInputs) { inputs.FaceEmbedding[3] = float32(math.NaN()) }},
		{"invalid scalar", "blur score out of range", func(inputs *ScoreInputs) { inputs.DocQualityFeatures.BlurScore = 1.1 }},
		{"invalid scope count", "scope count out of range", func(inputs *ScoreInputs) { inputs.ScopeCount = 11 }},
		{"missing scope types", "missing scope types", func(inputs *ScoreInputs) { inputs.ScopeTypes = nil }},
		{"negative block height", "negative block height", func(inputs *ScoreInputs) { inputs.Metadata.BlockHeight = -1 }},
	}
	extractor := NewFeatureExtractor(ProductionFeatureExtractorConfig())
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			inputs := valid()
			testCase.mutate(inputs)
			if _, err := extractor.ExtractFeatures(inputs); err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error = %v, want containing %q", err, testCase.want)
			}
		})
	}
	if _, err := extractor.ExtractFeatures(nil); err == nil {
		t.Fatal("nil inputs must fail")
	}

	config := ProductionFeatureExtractorConfig()
	config.NormalizeFeatures = true
	config.FeatureMean = make([]float32, TotalFeatureDim-1)
	config.FeatureStd = make([]float32, TotalFeatureDim-1)
	if _, err := NewFeatureExtractor(config).ExtractFeatures(valid()); err == nil || !strings.Contains(err.Error(), "normalization dimensions") {
		t.Fatalf("normalization mismatch error = %v", err)
	}
}
