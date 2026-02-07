package mrz

import "testing"

func TestCountryName(t *testing.T) {
	name, ok := CountryName("USA")
	if !ok || name != "United States" {
		t.Fatalf("unexpected country name: %v %v", name, ok)
	}

	name, ok = CountryName("gb")
	if !ok || name != "United Kingdom" {
		t.Fatalf("expected GB -> United Kingdom, got %v %v", name, ok)
	}

	if _, ok := CountryName("ZZZ"); ok {
		t.Fatalf("expected unknown country to return false")
	}
}
