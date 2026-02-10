package types

import (
	"math"
	"testing"
)

func TestMLFeatureSchemaConstants(t *testing.T) {
	// Verify dimension constants sum correctly
	t.Run("dimension constants sum to total", func(t *testing.T) {
		sum := FaceEmbeddingDim + DocQualityDim + OCRFeaturesDim + MetadataFeaturesDim + PaddingDim
		if sum != TotalFeatureDim {
			t.Errorf("dimension constants sum to %d, expected %d", sum, TotalFeatureDim)
		}
	})

	t.Run("total dimension is 768", func(t *testing.T) {
		if TotalFeatureDim != 768 {
			t.Errorf("TotalFeatureDim = %d, expected 768", TotalFeatureDim)
		}
	})

	t.Run("face embedding dimension is 512", func(t *testing.T) {
		if FaceEmbeddingDim != 512 {
			t.Errorf("FaceEmbeddingDim = %d, expected 512", FaceEmbeddingDim)
		}
	})

	t.Run("OCR features dimension is 10", func(t *testing.T) {
		if OCRFeaturesDim != 10 {
			t.Errorf("OCRFeaturesDim = %d, expected 10", OCRFeaturesDim)
		}
	})

	t.Run("OCR field count is 5", func(t *testing.T) {
		if OCRFieldCount != 5 {
			t.Errorf("OCRFieldCount = %d, expected 5", OCRFieldCount)
		}
	})

	t.Run("metadata dimension is 16", func(t *testing.T) {
		if MetadataFeaturesDim != 16 {
			t.Errorf("MetadataFeaturesDim = %d, expected 16", MetadataFeaturesDim)
		}
	})

	t.Run("padding dimension is 225", func(t *testing.T) {
		if PaddingDim != 225 {
			t.Errorf("PaddingDim = %d, expected 225", PaddingDim)
		}
	})
}

func TestFeatureOffsets(t *testing.T) {
	offsets := FeatureOffsets()

	t.Run("has all feature groups", func(t *testing.T) {
		groups := AllFeatureGroups()
		if len(offsets) != len(groups) {
			t.Errorf("got %d offsets, expected %d", len(offsets), len(groups))
		}
	})

	t.Run("offsets are contiguous", func(t *testing.T) {
		expectedStart := 0
		for _, offset := range offsets {
			if offset.StartIndex != expectedStart {
				t.Errorf("group %s starts at %d, expected %d", offset.Group, offset.StartIndex, expectedStart)
			}
			expectedStart = offset.EndIndex + 1
		}
		if expectedStart != TotalFeatureDim {
			t.Errorf("final offset ends at %d, expected %d", expectedStart, TotalFeatureDim)
		}
	})

	t.Run("face offset is correct", func(t *testing.T) {
		offset, found := GetFeatureOffset(FeatureGroupFace)
		if !found {
			t.Fatal("face feature group not found")
		}
		if offset.StartIndex != 0 || offset.EndIndex != 511 || offset.Dimension != 512 {
			t.Errorf("face offset = %+v, expected start=0, end=511, dim=512", offset)
		}
	})

	t.Run("doc quality offset is correct", func(t *testing.T) {
		offset, found := GetFeatureOffset(FeatureGroupDocQuality)
		if !found {
			t.Fatal("doc quality feature group not found")
		}
		if offset.StartIndex != 512 || offset.EndIndex != 516 || offset.Dimension != 5 {
			t.Errorf("doc quality offset = %+v, expected start=512, end=516, dim=5", offset)
		}
	})

	t.Run("OCR offset is correct", func(t *testing.T) {
		offset, found := GetFeatureOffset(FeatureGroupOCR)
		if !found {
			t.Fatal("OCR feature group not found")
		}
		if offset.StartIndex != 517 || offset.EndIndex != 526 || offset.Dimension != 10 {
			t.Errorf("OCR offset = %+v, expected start=517, end=526, dim=10", offset)
		}
	})

	t.Run("metadata offset is correct", func(t *testing.T) {
		offset, found := GetFeatureOffset(FeatureGroupMetadata)
		if !found {
			t.Fatal("metadata feature group not found")
		}
		if offset.StartIndex != 527 || offset.EndIndex != 542 || offset.Dimension != 16 {
			t.Errorf("metadata offset = %+v, expected start=527, end=542, dim=16", offset)
		}
	})

	t.Run("padding offset is correct", func(t *testing.T) {
		offset, found := GetFeatureOffset(FeatureGroupPadding)
		if !found {
			t.Fatal("padding feature group not found")
		}
		if offset.StartIndex != 543 || offset.EndIndex != 767 || offset.Dimension != 225 {
			t.Errorf("padding offset = %+v, expected start=543, end=767, dim=225", offset)
		}
	})
}

func TestOCRFieldNames(t *testing.T) {
	fields := OCRFieldNames()

	t.Run("has correct field count", func(t *testing.T) {
		if len(fields) != OCRFieldCount {
			t.Errorf("got %d OCR fields, expected %d", len(fields), OCRFieldCount)
		}
	})

	t.Run("has expected fields", func(t *testing.T) {
		expected := []OCRFieldName{
			OCRFieldName_Name,
			OCRFieldName_DateOfBirth,
			OCRFieldName_DocumentNumber,
			OCRFieldName_ExpiryDate,
			OCRFieldName_Nationality,
		}
		for i, f := range expected {
			if fields[i] != f {
				t.Errorf("field %d = %s, expected %s", i, fields[i], f)
			}
		}
	})

	t.Run("OCRFieldIndex returns correct indices", func(t *testing.T) {
		testCases := []struct {
			field OCRFieldName
			index int
		}{
			{OCRFieldName_Name, 0},
			{OCRFieldName_DateOfBirth, 1},
			{OCRFieldName_DocumentNumber, 2},
			{OCRFieldName_ExpiryDate, 3},
			{OCRFieldName_Nationality, 4},
		}

		for _, tc := range testCases {
			idx, found := OCRFieldIndex(tc.field)
			if !found {
				t.Errorf("field %s not found", tc.field)
				continue
			}
			if idx != tc.index {
				t.Errorf("field %s index = %d, expected %d", tc.field, idx, tc.index)
			}
		}
	})

	t.Run("OCRFieldIndex returns false for unknown field", func(t *testing.T) {
		_, found := OCRFieldIndex("unknown_field")
		if found {
			t.Error("expected false for unknown field")
		}
	})
}

func TestScopeTypeToConsentCategory(t *testing.T) {
	testCases := []struct {
		scopeType ScopeType
		expected  ConsentCategory
	}{
		{ScopeTypeIDDocument, ConsentCategoryBiometricPII},
		{ScopeTypeSelfie, ConsentCategoryBiometric},
		{ScopeTypeFaceVideo, ConsentCategoryBiometric},
		{ScopeTypeBiometric, ConsentCategoryBiometric},
		{ScopeTypeSSOMetadata, ConsentCategoryIdentityAttestation},
		{ScopeTypeEmailProof, ConsentCategoryContactVerification},
		{ScopeTypeSMSProof, ConsentCategoryContactVerification},
		{ScopeTypeDomainVerify, ConsentCategoryDomainOwnership},
		{ScopeTypeADSSO, ConsentCategoryEnterpriseIdentity},
	}

	for _, tc := range testCases {
		t.Run(string(tc.scopeType), func(t *testing.T) {
			result := ScopeTypeToConsentCategory(tc.scopeType)
			if result != tc.expected {
				t.Errorf("ScopeTypeToConsentCategory(%s) = %s, expected %s", tc.scopeType, result, tc.expected)
			}
		})
	}
}

func TestRequiresExplicitConsent(t *testing.T) {
	testCases := []struct {
		category ConsentCategory
		expected bool
	}{
		{ConsentCategoryBiometricPII, true},
		{ConsentCategoryBiometric, true},
		{ConsentCategoryIdentityAttestation, false},
		{ConsentCategoryContactVerification, false},
		{ConsentCategoryDomainOwnership, false},
		{ConsentCategoryEnterpriseIdentity, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.category), func(t *testing.T) {
			result := RequiresExplicitConsent(tc.category)
			if result != tc.expected {
				t.Errorf("RequiresExplicitConsent(%s) = %v, expected %v", tc.category, result, tc.expected)
			}
		})
	}
}

func TestMLFeatureVector(t *testing.T) {
	t.Run("NewMLFeatureVector creates correct dimensions", func(t *testing.T) {
		v := NewMLFeatureVector()
		if len(v.Features) != TotalFeatureDim {
			t.Errorf("feature vector length = %d, expected %d", len(v.Features), TotalFeatureDim)
		}
	})

	t.Run("NewMLFeatureVector sets schema version", func(t *testing.T) {
		v := NewMLFeatureVector()
		if v.SchemaVersion != MLFeatureSchemaVersion {
			t.Errorf("schema version = %s, expected %s", v.SchemaVersion, MLFeatureSchemaVersion)
		}
	})

	t.Run("NewMLFeatureVector sets doc quality defaults", func(t *testing.T) {
		v := NewMLFeatureVector()
		offset, _ := GetFeatureOffset(FeatureGroupDocQuality)
		for i, r := range DocQualityFeatureRanges() {
			val := v.Features[offset.StartIndex+i]
			if val != r.Default {
				t.Errorf("doc quality feature %s default = %.2f, expected %.2f", r.Name, val, r.Default)
			}
		}
	})

	t.Run("Validate accepts valid vector", func(t *testing.T) {
		v := NewMLFeatureVector()
		err := v.Validate()
		if err != nil {
			t.Errorf("Validate() returned error for valid vector: %v", err)
		}
	})

	t.Run("Validate rejects wrong dimension", func(t *testing.T) {
		v := &MLFeatureVector{
			SchemaVersion: MLFeatureSchemaVersion,
			Features:      make([]float32, 100),
		}
		err := v.Validate()
		if err == nil {
			t.Error("Validate() should reject wrong dimension")
		}
	})

	t.Run("Validate rejects empty schema version", func(t *testing.T) {
		v := &MLFeatureVector{
			SchemaVersion: "",
			Features:      make([]float32, TotalFeatureDim),
		}
		err := v.Validate()
		if err == nil {
			t.Error("Validate() should reject empty schema version")
		}
	})

	t.Run("Validate checks face embedding normalization", func(t *testing.T) {
		v := NewMLFeatureVector()
		// Set non-zero, non-normalized embedding
		for i := 0; i < FaceEmbeddingDim; i++ {
			v.Features[i] = 2.0 // L2 norm will be sqrt(512*4) = ~45.25
		}
		err := v.Validate()
		if err == nil {
			t.Error("Validate() should reject non-normalized face embedding")
		}
	})

	t.Run("Validate accepts normalized face embedding", func(t *testing.T) {
		v := NewMLFeatureVector()
		// Set normalized embedding (unit length)
		norm := float32(math.Sqrt(float64(FaceEmbeddingDim)))
		for i := 0; i < FaceEmbeddingDim; i++ {
			v.Features[i] = 1.0 / norm
		}
		err := v.Validate()
		if err != nil {
			t.Errorf("Validate() rejected normalized face embedding: %v", err)
		}
	})

	t.Run("Validate rejects out-of-range doc quality", func(t *testing.T) {
		v := NewMLFeatureVector()
		offset, _ := GetFeatureOffset(FeatureGroupDocQuality)
		v.Features[offset.StartIndex] = 1.5 // Out of range [0, 1]
		err := v.Validate()
		if err == nil {
			t.Error("Validate() should reject out-of-range doc quality")
		}
	})
}

func TestGetSetFeatureGroup(t *testing.T) {
	t.Run("GetFeatureGroup returns correct slice", func(t *testing.T) {
		v := NewMLFeatureVector()
		// Set some values in doc quality
		offset, _ := GetFeatureOffset(FeatureGroupDocQuality)
		for i := 0; i < DocQualityDim; i++ {
			v.Features[offset.StartIndex+i] = float32(i) * 0.1
		}

		group, err := v.GetFeatureGroup(FeatureGroupDocQuality)
		if err != nil {
			t.Fatalf("GetFeatureGroup failed: %v", err)
		}

		if len(group) != DocQualityDim {
			t.Errorf("group length = %d, expected %d", len(group), DocQualityDim)
		}

		for i := 0; i < DocQualityDim; i++ {
			expected := float32(i) * 0.1
			if group[i] != expected {
				t.Errorf("group[%d] = %.2f, expected %.2f", i, group[i], expected)
			}
		}
	})

	t.Run("SetFeatureGroup sets values correctly", func(t *testing.T) {
		v := NewMLFeatureVector()
		values := []float32{0.9, 0.8, 0.7, 0.6, 0.5}

		err := v.SetFeatureGroup(FeatureGroupDocQuality, values)
		if err != nil {
			t.Fatalf("SetFeatureGroup failed: %v", err)
		}

		group, _ := v.GetFeatureGroup(FeatureGroupDocQuality)
		for i, val := range values {
			if group[i] != val {
				t.Errorf("group[%d] = %.2f, expected %.2f", i, group[i], val)
			}
		}
	})

	t.Run("SetFeatureGroup rejects wrong dimension", func(t *testing.T) {
		v := NewMLFeatureVector()
		values := []float32{0.9, 0.8, 0.7} // Wrong length

		err := v.SetFeatureGroup(FeatureGroupDocQuality, values)
		if err == nil {
			t.Error("SetFeatureGroup should reject wrong dimension")
		}
	})

	t.Run("GetFeatureGroup rejects unknown group", func(t *testing.T) {
		v := NewMLFeatureVector()
		_, err := v.GetFeatureGroup("unknown")
		if err == nil {
			t.Error("GetFeatureGroup should reject unknown group")
		}
	})
}

func TestIsSchemaCompatible(t *testing.T) {
	testCases := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "1.1.0", true},
		{"1.0.0", "1.2.5", true},
		{"1.0.0", "2.0.0", false},
		{"2.0.0", "1.0.0", false},
		{"invalid", "1.0.0", false},
		{"1.0.0", "invalid", false},
		{"invalid", "invalid", true}, // Falls back to exact match
	}

	for _, tc := range testCases {
		t.Run(tc.v1+"_vs_"+tc.v2, func(t *testing.T) {
			result := IsSchemaCompatible(tc.v1, tc.v2)
			if result != tc.expected {
				t.Errorf("IsSchemaCompatible(%s, %s) = %v, expected %v", tc.v1, tc.v2, result, tc.expected)
			}
		})
	}
}

func TestGetSchemaInfo(t *testing.T) {
	info := GetSchemaInfo()

	t.Run("contains version", func(t *testing.T) {
		if info["version"] != MLFeatureSchemaVersion {
			t.Errorf("version = %v, expected %s", info["version"], MLFeatureSchemaVersion)
		}
	})

	t.Run("contains total_dimension", func(t *testing.T) {
		if info["total_dimension"] != TotalFeatureDim {
			t.Errorf("total_dimension = %v, expected %d", info["total_dimension"], TotalFeatureDim)
		}
	})

	t.Run("contains all dimension keys", func(t *testing.T) {
		requiredKeys := []string{
			"version", "major", "minor", "patch",
			"total_dimension", "face_embedding_dim", "doc_quality_dim",
			"ocr_features_dim", "metadata_dim", "padding_dim",
			"ocr_field_count", "scope_schema_version",
		}

		for _, key := range requiredKeys {
			if _, ok := info[key]; !ok {
				t.Errorf("missing key: %s", key)
			}
		}
	})
}

func TestFeatureRanges(t *testing.T) {
	t.Run("doc quality ranges are valid", func(t *testing.T) {
		ranges := DocQualityFeatureRanges()
		if len(ranges) != DocQualityDim {
			t.Errorf("got %d doc quality ranges, expected %d", len(ranges), DocQualityDim)
		}
		for _, r := range ranges {
			if r.Min > r.Max {
				t.Errorf("range %s has min > max: %.2f > %.2f", r.Name, r.Min, r.Max)
			}
			if r.Default < r.Min || r.Default > r.Max {
				t.Errorf("range %s has default outside range: %.2f not in [%.2f, %.2f]", r.Name, r.Default, r.Min, r.Max)
			}
		}
	})

	t.Run("OCR ranges are valid", func(t *testing.T) {
		ranges := OCRFeatureRanges()
		if len(ranges) != 2 { // confidence and validated
			t.Errorf("got %d OCR ranges, expected 2", len(ranges))
		}
		for _, r := range ranges {
			if r.Min > r.Max {
				t.Errorf("range %s has min > max: %.2f > %.2f", r.Name, r.Min, r.Max)
			}
		}
	})

	t.Run("metadata ranges are valid", func(t *testing.T) {
		ranges := MetadataFeatureRanges()
		if len(ranges) != MetadataFeaturesDim {
			t.Errorf("got %d metadata ranges, expected %d", len(ranges), MetadataFeaturesDim)
		}
		for _, r := range ranges {
			if r.Min > r.Max {
				t.Errorf("range %s has min > max: %.2f > %.2f", r.Name, r.Min, r.Max)
			}
		}
	})
}

// TestFeatureVectorDeterminism verifies that feature vector creation is deterministic
func TestFeatureVectorDeterminism(t *testing.T) {
	t.Run("NewMLFeatureVector is deterministic", func(t *testing.T) {
		v1 := NewMLFeatureVector()
		v2 := NewMLFeatureVector()

		if len(v1.Features) != len(v2.Features) {
			t.Fatal("feature vectors have different lengths")
		}

		for i := range v1.Features {
			if v1.Features[i] != v2.Features[i] {
				t.Errorf("feature[%d] differs: %.6f vs %.6f", i, v1.Features[i], v2.Features[i])
			}
		}
	})
}

// Example test showing how to construct a feature vector
func ExampleMLFeatureVector() {
	// Create a new feature vector with defaults
	v := NewMLFeatureVector()

	// Set document quality features
	docQuality := []float32{0.85, 0.72, 0.68, 0.88, 0.92}
	_ = v.SetFeatureGroup(FeatureGroupDocQuality, docQuality)

	// Validate the vector
	err := v.Validate()
	if err != nil {
		panic(err)
	}

	// Get a specific feature group
	quality, _ := v.GetFeatureGroup(FeatureGroupDocQuality)
	_ = quality // Use the quality features
}
