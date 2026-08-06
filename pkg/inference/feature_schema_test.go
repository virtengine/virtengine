package inference

import "testing"

func TestCanonicalFeatureSchemaLayout(t *testing.T) {
	schema, err := LoadFeatureSchema()
	if err != nil {
		t.Fatalf("LoadFeatureSchema() error = %v", err)
	}

	expected := map[string]struct {
		offset int
		width  int
	}{
		"selfie_embedding":       {0, 512},
		"face_confidence":        {512, 1},
		"document_quality":       {513, 1},
		"document_sharpness":     {514, 1},
		"document_brightness":    {515, 1},
		"document_contrast":      {516, 1},
		"document_noise_quality": {517, 1},
		"document_blur_quality":  {518, 1},
		"ocr_name":               {519, 2},
		"ocr_date_of_birth":      {521, 2},
		"ocr_document_number":    {523, 2},
		"ocr_expiry_date":        {525, 2},
		"ocr_nationality":        {527, 2},
		"metadata":               {529, 16},
		"reserved":               {545, 223},
	}
	for name, want := range expected {
		segment, ok := featureSegment(schema, name)
		if !ok {
			t.Fatalf("schema missing segment %q", name)
		}
		if segment.Offset != want.offset || segment.Dimensions != want.width {
			t.Errorf("segment %q = [%d,%d), want [%d,%d)", name, segment.Offset, segment.Offset+segment.Dimensions, want.offset, want.offset+want.width)
		}
	}
}
