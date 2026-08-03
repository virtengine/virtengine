package inference

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
)

type featureParityFixture struct {
	Schema  string `json:"schema"`
	Version int    `json:"version"`
	Layout  struct {
		TotalDimension int `json:"total_dimension"`
		Components     []struct {
			Name                  string   `json:"name"`
			Offset                int      `json:"offset"`
			Dimension             int      `json:"dimension"`
			PositionName          string   `json:"position_name"`
			PositionNames         []string `json:"position_names"`
			Fields                []string `json:"fields"`
			PositionNamesPerField []string `json:"position_names_per_field"`
		} `json:"components"`
		Encoding struct {
			ValueType            string `json:"value_type"`
			ByteOrder            string `json:"byte_order"`
			PreHashDecimalPlaces int    `json:"pre_hash_decimal_places"`
			HashAlgorithm        string `json:"hash_algorithm"`
		} `json:"encoding"`
	} `json:"layout"`
	Cases []struct {
		Name   string `json:"name"`
		Source struct {
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
			} `json:"document_quality"`
			OCRConfidences     map[string]float32 `json:"ocr_confidences"`
			OCRFieldValidation map[string]bool    `json:"ocr_field_validation"`
			ScopeTypes         []string           `json:"scope_types"`
			ScopeCount         int                `json:"scope_count"`
			BlockHeight        int64              `json:"block_height"`
		} `json:"source"`
		ExpectedVectorHash       string             `json:"expected_vector_hash"`
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
	if len(fixture.Cases) == 0 {
		t.Fatal("feature parity fixture has zero cases")
	}

	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
	hasher := NewDeterminismController(42, true)
	for _, testCase := range fixture.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			inputs := featureParityInputs(t, &testCase)
			features, err := extractor.ExtractFeatures(inputs)
			if err != nil {
				t.Fatalf("extract canonical features: %v", err)
			}
			if len(features) != TotalFeatureDim {
				t.Fatalf("feature dimension: got %d, want %d", len(features), TotalFeatureDim)
			}
			if got := hasher.ComputeFeatureHash(features); got != testCase.ExpectedVectorHash {
				t.Fatalf("feature hash: got %s, want %s", got, testCase.ExpectedVectorHash)
			}
			for index, value := range features {
				expected, nonzero := testCase.ExpectedNonzeroPositions[strconv.Itoa(index)]
				if nonzero {
					if value != expected {
						t.Errorf("position %d: got %v, want %v", index, value, expected)
					}
				} else if value != 0 {
					t.Errorf("position %d: got unexpected nonzero value %v", index, value)
				}
			}
		})
	}
}

func validateFeatureParityLayout(t *testing.T, fixture *featureParityFixture) {
	t.Helper()
	if fixture.Schema != "virtengine.inference.feature_parity" || fixture.Version != 1 {
		t.Fatalf("unsupported feature parity schema %q version %d", fixture.Schema, fixture.Version)
	}
	if fixture.Layout.TotalDimension != TotalFeatureDim {
		t.Fatalf("fixture dimension: got %d, want %d", fixture.Layout.TotalDimension, TotalFeatureDim)
	}
	expected := []struct {
		name              string
		offset, dimension int
	}{
		{"face_embedding", 0, FaceEmbeddingDim},
		{"document_quality", FaceEmbeddingDim, DocQualityDim},
		{"ocr", FaceEmbeddingDim + DocQualityDim, OCRFieldsDim},
		{"metadata", FaceEmbeddingDim + DocQualityDim + OCRFieldsDim, MetadataDim},
		{"padding", FaceEmbeddingDim + DocQualityDim + OCRFieldsDim + MetadataDim, PaddingDim},
	}
	if len(fixture.Layout.Components) != len(expected) {
		t.Fatalf("layout components: got %d, want %d", len(fixture.Layout.Components), len(expected))
	}
	for index, want := range expected {
		got := fixture.Layout.Components[index]
		if got.Name != want.name || got.Offset != want.offset || got.Dimension != want.dimension {
			t.Errorf("layout component %d: got %s/%d/%d, want %s/%d/%d",
				index, got.Name, got.Offset, got.Dimension, want.name, want.offset, want.dimension)
		}
	}
	if got := fixture.Layout.Components[1].PositionNames; len(got) != DocQualityDim {
		t.Fatalf("document quality position names: got %d, want %d", len(got), DocQualityDim)
	}
	ocr := fixture.Layout.Components[2]
	if len(ocr.Fields) != len(OCRFieldNames) {
		t.Fatalf("OCR fields: got %d, want %d", len(ocr.Fields), len(OCRFieldNames))
	}
	for index, want := range OCRFieldNames {
		if ocr.Fields[index] != want {
			t.Fatalf("OCR field %d: got %q, want %q", index, ocr.Fields[index], want)
		}
	}
	if len(ocr.PositionNamesPerField) != 2 || ocr.PositionNamesPerField[0] != "confidence" || ocr.PositionNamesPerField[1] != "validated" {
		t.Fatalf("unexpected OCR position names: %v", ocr.PositionNamesPerField)
	}
	if got := fixture.Layout.Components[3].PositionNames; len(got) != MetadataDim {
		t.Fatalf("metadata position names: got %d, want %d", len(got), MetadataDim)
	}
	encoding := fixture.Layout.Encoding
	if encoding.ValueType != "ieee754-float32" || encoding.ByteOrder != "big-endian" ||
		encoding.PreHashDecimalPlaces != 6 || encoding.HashAlgorithm != "sha256" {
		t.Fatalf("unsupported feature hash encoding: %+v", encoding)
	}
}

func featureParityInputs(t *testing.T, testCase *struct {
	Name   string `json:"name"`
	Source struct {
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
		} `json:"document_quality"`
		OCRConfidences     map[string]float32 `json:"ocr_confidences"`
		OCRFieldValidation map[string]bool    `json:"ocr_field_validation"`
		ScopeTypes         []string           `json:"scope_types"`
		ScopeCount         int                `json:"scope_count"`
		BlockHeight        int64              `json:"block_height"`
	} `json:"source"`
	ExpectedVectorHash       string             `json:"expected_vector_hash"`
	ExpectedNonzeroPositions map[string]float32 `json:"expected_nonzero_positions"`
}) *ScoreInputs {
	t.Helper()
	source := testCase.Source
	var embedding []float32
	switch source.FaceEmbedding.Kind {
	case "missing":
	case "zeros":
		embedding = make([]float32, source.FaceEmbedding.Dimension)
	case "single_index":
		embedding = make([]float32, source.FaceEmbedding.Dimension)
		if source.FaceEmbedding.Index < 0 || source.FaceEmbedding.Index >= len(embedding) {
			t.Fatalf("face sentinel index %d outside dimension %d", source.FaceEmbedding.Index, len(embedding))
		}
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
		FaceEmbedding:   embedding,
		FaceConfidence:  source.FaceConfidence,
		DocQualityScore: source.DocQualityScore,
		DocQualityFeatures: DocQualityFeatures{
			Sharpness: source.DocumentQuality.Sharpness, Brightness: source.DocumentQuality.Brightness,
			Contrast: source.DocumentQuality.Contrast, NoiseLevel: source.DocumentQuality.NoiseLevel,
		},
		OCRConfidences: source.OCRConfidences, OCRFieldValidation: source.OCRFieldValidation,
		ScopeTypes: source.ScopeTypes, ScopeCount: source.ScopeCount,
		Metadata: InferenceMetadata{BlockHeight: source.BlockHeight},
	}
}

func TestFeatureExtractorRejectsNonfiniteInputs(t *testing.T) {
	inputs := &ScoreInputs{FaceEmbedding: make([]float32, FaceEmbeddingDim)}
	inputs.FaceEmbedding[7] = float32(math.NaN())
	_, err := NewFeatureExtractor(DefaultFeatureExtractorConfig()).ExtractFeatures(inputs)
	if err == nil {
		t.Fatal("expected nonfinite feature input to fail")
	}
	if got := err.Error(); got != "face embedding[7] must be finite" {
		t.Fatalf("unexpected error: %s", got)
	}
}
