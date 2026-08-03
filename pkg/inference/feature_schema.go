package inference

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

const FeatureSchemaVersion = "virtengine.trust-score-features/v1"

//go:embed schema/trust_score_features_v1.json
var canonicalFeatureSchemaJSON []byte

type FeatureSchema struct {
	SchemaVersion   string                 `json:"schema_version"`
	DType           string                 `json:"dtype"`
	ByteOrder       string                 `json:"byte_order"`
	TotalDimensions int                    `json:"total_dimensions"`
	Segments        []FeatureSchemaSegment `json:"segments"`
}

type FeatureSchemaSegment struct {
	Name       string   `json:"name"`
	Offset     int      `json:"offset"`
	Dimensions int      `json:"dimensions"`
	Required   bool     `json:"required"`
	Transform  string   `json:"transform,omitempty"`
	Values     []string `json:"values,omitempty"`
	Value      *float32 `json:"value,omitempty"`
}

func LoadFeatureSchema() (FeatureSchema, error) {
	var schema FeatureSchema
	if err := json.Unmarshal(canonicalFeatureSchemaJSON, &schema); err != nil {
		return FeatureSchema{}, fmt.Errorf("decode canonical feature schema: %w", err)
	}
	if err := validateFeatureSchema(schema); err != nil {
		return FeatureSchema{}, err
	}
	return schema, nil
}

func validateFeatureSchema(schema FeatureSchema) error {
	if schema.SchemaVersion != FeatureSchemaVersion {
		return fmt.Errorf("feature schema version mismatch: %q", schema.SchemaVersion)
	}
	if schema.DType != "float32" || schema.ByteOrder != "big-endian" {
		return fmt.Errorf("unsupported feature encoding %s/%s", schema.DType, schema.ByteOrder)
	}
	if schema.TotalDimensions != TotalFeatureDim {
		return fmt.Errorf("feature dimension mismatch: schema=%d serving=%d", schema.TotalDimensions, TotalFeatureDim)
	}

	zero := float32(0)
	expected := []FeatureSchemaSegment{
		{Name: "selfie_embedding", Offset: 0, Dimensions: 512, Required: true, Transform: "l2_normalize"},
		{Name: "face_confidence", Offset: 512, Dimensions: 1, Required: true},
		{Name: "document_quality", Offset: 513, Dimensions: 1, Required: true},
		{Name: "document_sharpness", Offset: 514, Dimensions: 1, Required: true},
		{Name: "document_brightness", Offset: 515, Dimensions: 1, Required: true},
		{Name: "document_contrast", Offset: 516, Dimensions: 1, Required: true},
		{Name: "document_noise_quality", Offset: 517, Dimensions: 1, Required: true, Transform: "one_minus"},
		{Name: "document_blur_quality", Offset: 518, Dimensions: 1, Required: true, Transform: "one_minus"},
		{Name: "ocr_name", Offset: 519, Dimensions: 2, Required: true, Values: []string{"confidence", "validated"}},
		{Name: "ocr_date_of_birth", Offset: 521, Dimensions: 2, Required: true, Values: []string{"confidence", "validated"}},
		{Name: "ocr_document_number", Offset: 523, Dimensions: 2, Required: true, Values: []string{"confidence", "validated"}},
		{Name: "ocr_expiry_date", Offset: 525, Dimensions: 2, Required: true, Values: []string{"confidence", "validated"}},
		{Name: "ocr_nationality", Offset: 527, Dimensions: 2, Required: true, Values: []string{"confidence", "validated"}},
		{Name: "metadata", Offset: 529, Dimensions: 16, Required: true, Values: []string{"scope_count", "scope_id_document", "scope_selfie", "scope_face_video", "scope_biometric", "scope_sso_metadata", "scope_email_proof", "scope_sms_proof", "scope_domain_verify", "block_height_mod_1000000", "reserved_0", "reserved_1", "reserved_2", "reserved_3", "reserved_4", "reserved_5"}},
		{Name: "reserved", Offset: 545, Dimensions: 223, Required: false, Value: &zero},
	}
	if len(schema.Segments) != len(expected) {
		return fmt.Errorf("feature segment count mismatch: got %d, expected %d", len(schema.Segments), len(expected))
	}
	for i := range expected {
		got, want := schema.Segments[i], expected[i]
		if got.Name != want.Name || got.Offset != want.Offset || got.Dimensions != want.Dimensions || got.Required != want.Required || got.Transform != want.Transform || !equalStrings(got.Values, want.Values) || !equalOptionalFloat32(got.Value, want.Value) {
			return fmt.Errorf("feature segment %d (%q) does not match canonical contract", i, got.Name)
		}
	}
	return nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func equalOptionalFloat32(left, right *float32) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func featureSegment(schema FeatureSchema, name string) (FeatureSchemaSegment, bool) {
	for _, segment := range schema.Segments {
		if segment.Name == name {
			return segment, true
		}
	}
	return FeatureSchemaSegment{}, false
}
