package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvidenceTrustInventoryExactCoverage(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		collect  func(EvidenceTrustDescriptor) []string
	}{
		{"attestation", attestationStrings(AllAttestationTypes()), func(d EvidenceTrustDescriptor) []string { return attestationStrings(d.AttestationTypes) }},
		{"evidence", evidenceStrings(AllEvidenceTypes()), func(d EvidenceTrustDescriptor) []string { return evidenceStrings(d.EvidenceTypes) }},
		{"scope", scopeStrings(AllScopeTypes()), func(d EvidenceTrustDescriptor) []string { return scopeStrings(d.ScopeTypes) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counts := make(map[string]int)
			for _, descriptor := range EvidenceTrustInventory() {
				for _, value := range test.collect(descriptor) {
					counts[value]++
				}
			}
			if len(counts) != len(test.expected) {
				t.Fatalf("coverage size = %d, want %d: %#v", len(counts), len(test.expected), counts)
			}
			for _, value := range test.expected {
				if counts[value] != 1 {
					t.Errorf("coverage count for %q = %d, want 1", value, counts[value])
				}
			}
		})
	}
}

func TestEvidenceTrustInventoryDeterministicAndDefensive(t *testing.T) {
	first := EvidenceTrustInventory()
	second := EvidenceTrustInventory()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("inventory is not deterministic")
	}
	first[0].ID = "mutated"
	first[0].AttestationTypes[0] = AttestationType("mutated")
	third := EvidenceTrustInventory()
	if reflect.DeepEqual(first, third) || !reflect.DeepEqual(second, third) {
		t.Fatal("inventory does not return defensive copies")
	}

	descriptor, found := EvidenceTrustByAttestationType(AttestationTypeEmailVerification)
	if !found {
		t.Fatal("email descriptor not found")
	}
	descriptor.ScopeTypes[0] = ScopeType("mutated")
	again, found := EvidenceTrustByAttestationType(AttestationTypeEmailVerification)
	if !found || again.ScopeTypes[0] != ScopeTypeEmailProof {
		t.Fatal("lookup does not return a defensive descriptor")
	}
}

func TestValidateEvidenceTrustInventory(t *testing.T) {
	if err := ValidateEvidenceTrustInventory(); err != nil {
		t.Fatalf("ValidateEvidenceTrustInventory() error = %v", err)
	}
}

func TestEvidenceTrustInventoryRequiredNonTaxonomyMechanisms(t *testing.T) {
	requiredIDs := []string{
		"veid.compliance_attestation.v1",
		"veid.consensus_validator_vote.v1",
		"veid.cross_chain_ibc_attestation.v1",
		"veid.encrypted_scope_upload.v1",
		"veid.appeal_supporting_evidence.v1",
	}
	descriptors := make(map[string]EvidenceTrustDescriptor)
	for _, descriptor := range EvidenceTrustInventory() {
		descriptors[descriptor.ID] = descriptor
	}
	for _, id := range requiredIDs {
		t.Run(id, func(t *testing.T) {
			descriptor, found := descriptors[id]
			if !found {
				t.Fatalf("required evidence trust descriptor %q not found", id)
			}
			if descriptor.ProductionEligible {
				t.Error("required non-taxonomy mechanism is production eligible")
			}
			if descriptor.FailClosedReason == "" {
				t.Error("required non-taxonomy mechanism has no fail-closed reason")
			}
		})
	}
}

func TestEvidenceTrustUnknownLookupFailsClosed(t *testing.T) {
	tests := []struct {
		name     string
		lookup   func() bool
		eligible func() bool
	}{
		{"attestation", func() bool { _, found := EvidenceTrustByAttestationType(AttestationType("unknown")); return found }, func() bool { return IsProductionEligibleAttestationType(AttestationType("unknown")) }},
		{"evidence", func() bool { _, found := EvidenceTrustByEvidenceType(EvidenceType("unknown")); return found }, func() bool { return IsProductionEligibleEvidenceType(EvidenceType("unknown")) }},
		{"scope", func() bool { _, found := EvidenceTrustByScopeType(ScopeType("unknown")); return found }, func() bool { return IsProductionEligibleScopeType(ScopeType("unknown")) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.lookup() {
				t.Error("unknown value unexpectedly resolved")
			}
			if test.eligible() {
				t.Error("unknown value unexpectedly production eligible")
			}
		})
	}
}

func TestEvidenceTrustProductionEligibilityIsExplicit(t *testing.T) {
	eligibleDescriptors := 0
	for _, descriptor := range EvidenceTrustInventory() {
		if descriptor.ProductionEligible {
			eligibleDescriptors++
			t.Errorf("%s is unexpectedly production eligible", descriptor.ID)
		}
	}
	if eligibleDescriptors != 0 {
		t.Fatalf("production-eligible descriptor count = %d, want 0", eligibleDescriptors)
	}
	for _, value := range AllAttestationTypes() {
		if IsProductionEligibleAttestationType(value) {
			t.Errorf("attestation type %q is unexpectedly production eligible", value)
		}
	}
	for _, value := range AllEvidenceTypes() {
		if IsProductionEligibleEvidenceType(value) {
			t.Errorf("evidence type %q is unexpectedly production eligible", value)
		}
	}
	for _, value := range AllScopeTypes() {
		if IsProductionEligibleScopeType(value) {
			t.Errorf("scope type %q is unexpectedly production eligible", value)
		}
	}

	descriptor, found := EvidenceTrustByAttestationType(AttestationTypeInferenceReceipt)
	if !found || descriptor.TrustClassification != EvidenceTrustCryptographicallyComplete {
		t.Fatal("inference receipt must be explicitly inventoried as cryptographically complete")
	}
	if descriptor.ProductionEligible || IsProductionEligibleAttestationType(AttestationTypeInferenceReceipt) {
		t.Fatal("complete classification implicitly enabled the unwired inference receipt producer")
	}
}

func TestEvidenceTrustWebPathsRemainProductionIneligible(t *testing.T) {
	webAttestationTypes := []AttestationType{
		AttestationTypeEmailVerification,
		AttestationTypeSMSVerification,
		AttestationTypeSSOVerification,
		AttestationTypeSocialMediaVerification,
	}
	for _, value := range webAttestationTypes {
		descriptor, found := EvidenceTrustByAttestationType(value)
		if !found {
			t.Fatalf("web descriptor for %q not found", value)
		}
		if descriptor.TrustClassification != EvidenceTrustCryptographicallyComplete {
			t.Errorf("web descriptor %q classification = %q, want cryptographically complete", descriptor.ID, descriptor.TrustClassification)
		}
		for _, requiredReason := range []string{"governed signer-key provisioning", "atomic score-lineage failure handling"} {
			if !strings.Contains(descriptor.FailClosedReason, requiredReason) {
				t.Errorf("web descriptor %q fail-closed reason does not mention %q", descriptor.ID, requiredReason)
			}
		}
	}
}

func TestEvidenceTrustInventoryValidatorRejectsMalformedTables(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		mutate  func([]EvidenceTrustDescriptor) []EvidenceTrustDescriptor
	}{
		{"duplicate id", "duplicate evidence trust descriptor id", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[1].ID = inventory[0].ID
			return inventory
		}},
		{"missing security field", "missing required field issuer_signer_policy", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[0].IssuerSignerPolicy = ""
			return inventory
		}},
		{"missing mutation field", "missing required field mutation_point_status_update", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[0].MutationPointStatusUpdate = ""
			return inventory
		}},
		{"invalid classification", "invalid trust classification", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[0].TrustClassification = EvidenceTrustClassification("unknown")
			return inventory
		}},
		{"ineligible classification", "production eligibility requires", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[0].ProductionEligible = true
			return inventory
		}},
		{"duplicate taxonomy ownership", "duplicate attestation type ownership", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[1].AttestationTypes = append(inventory[1].AttestationTypes, inventory[0].AttestationTypes[0])
			return inventory
		}},
		{"missing coverage", "missing attestation type coverage", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			inventory[0].AttestationTypes = nil
			return inventory
		}},
		{"unrecognized taxonomy-less descriptor", "not a recognized non-taxonomy mechanism", func(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
			descriptor := inventory[0]
			descriptor.ID = "veid.unrecognized_mechanism.v1"
			descriptor.AttestationTypes = nil
			descriptor.EvidenceTypes = nil
			descriptor.ScopeTypes = nil
			return append(inventory, descriptor)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inventory := test.mutate(EvidenceTrustInventory())
			err := validateEvidenceTrustInventory(inventory)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validateEvidenceTrustInventory() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

func attestationStrings(values []AttestationType) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func evidenceStrings(values []EvidenceType) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func scopeStrings(values []ScopeType) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}
