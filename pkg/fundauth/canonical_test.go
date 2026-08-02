package fundauth

import (
	"strings"
	"testing"
)

func TestEverySecurityFieldMutationChangesDigest(t *testing.T) {
	base := validAuthorization()
	baseDigest, err := AuthorizationDigest(base)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*FundAuthorization){
		"chain":         func(auth *FundAuthorization) { auth.ChainID = "virtengine-test-2" },
		"account":       func(auth *FundAuthorization) { auth.AccountID = "account:carol" },
		"key id":        func(auth *FundAuthorization) { auth.SignerKeyID = "device-key:backup" },
		"key epoch":     func(auth *FundAuthorization) { auth.SignerKeyEpoch++ },
		"source":        func(auth *FundAuthorization) { auth.SourceID = "/cosmos.bank.v1beta1.MsgMultiSend" },
		"type URL":      func(auth *FundAuthorization) { auth.TypeURL = "/cosmos.bank.v1beta1.MsgMultiSend" },
		"phase":         func(auth *FundAuthorization) { auth.Phase = PhaseDeferred },
		"effect":        func(auth *FundAuthorization) { auth.Effect = EffectReward },
		"message":       func(auth *FundAuthorization) { auth.MessageDigestHex = testDigest("message-2") },
		"amount denom":  func(auth *FundAuthorization) { auth.Amounts[0].Denom = "abc" },
		"amount value":  func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "43" },
		"party role":    func(auth *FundAuthorization) { auth.Parties[1].Role = PartyRolePayer },
		"party account": func(auth *FundAuthorization) { auth.Parties[1].AccountID = "account:carol" },
		"case":          func(auth *FundAuthorization) { auth.CaseDigestHex = testDigest("case-2") },
		"order":         func(auth *FundAuthorization) { auth.OrderDigestHex = testDigest("order-2") },
		"reference":     func(auth *FundAuthorization) { auth.ReferenceDigestHex = testDigest("reference-2") },
		"MFA mode":      func(auth *FundAuthorization) { auth.MFAMode = MFAPossessionOnlyPolicyApproved; auth.MFADigestHex = "" },
		"MFA":           func(auth *FundAuthorization) { auth.MFADigestHex = testDigest("mfa-2") },
		"eligibility mode": func(auth *FundAuthorization) {
			auth.EligibilityMode = EligibilityNotRequired
			auth.EligibilityDigestHex = ""
		},
		"eligibility":  func(auth *FundAuthorization) { auth.EligibilityDigestHex = testDigest("eligibility-2") },
		"policy":       func(auth *FundAuthorization) { auth.PolicyDigestHex = testDigest("policy-2") },
		"nonce":        func(auth *FundAuthorization) { auth.NonceDigestHex = testDigest("nonce-2") },
		"issued block": func(auth *FundAuthorization) { auth.IssuedAtBlock++; auth.LowerBlock++ },
		"issued time":  func(auth *FundAuthorization) { auth.IssuedAtUnix++ },
		"lower bound":  func(auth *FundAuthorization) { auth.LowerBlock++; auth.IssuedAtBlock++ },
		"upper bound":  func(auth *FundAuthorization) { auth.UpperBlock++ },
		"expiry kind":  func(auth *FundAuthorization) { auth.Expiry.Kind = ExpiryAtUnixTime },
		"expiry value": func(auth *FundAuthorization) { auth.Expiry.Value++ },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			auth := cloneAuthorization(base)
			mutate(&auth)
			digest, err := AuthorizationDigest(auth)
			if err != nil {
				t.Fatalf("valid mutation rejected: %v", err)
			}
			if digest == baseDigest {
				t.Fatal("mutation did not change digest")
			}
		})
	}
}

func TestCanonicalValidationRejectsMalformedValues(t *testing.T) {
	tests := map[string]func(*FundAuthorization){
		"domain":                   func(auth *FundAuthorization) { auth.Domain = "other" },
		"version":                  func(auth *FundAuthorization) { auth.Version++ },
		"zero key epoch":           func(auth *FundAuthorization) { auth.SignerKeyEpoch = 0 },
		"zero amount":              func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "0" },
		"bad denom":                func(auth *FundAuthorization) { auth.Amounts[0].Denom = "ACT" },
		"amount duplicate":         func(auth *FundAuthorization) { auth.Amounts[1].Denom = auth.Amounts[0].Denom },
		"amount order":             func(auth *FundAuthorization) { auth.Amounts[0], auth.Amounts[1] = auth.Amounts[1], auth.Amounts[0] },
		"amount leading zero":      func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "042" },
		"amount signed":            func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "+42" },
		"amount overflow":          func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "18446744073709551616" },
		"empty parties":            func(auth *FundAuthorization) { auth.Parties = nil },
		"party role":               func(auth *FundAuthorization) { auth.Parties[0].Role = PartyRole(255) },
		"party duplicate":          func(auth *FundAuthorization) { auth.Parties[1] = auth.Parties[0] },
		"party order":              func(auth *FundAuthorization) { auth.Parties[0], auth.Parties[1] = auth.Parties[1], auth.Parties[0] },
		"uppercase digest":         func(auth *FundAuthorization) { auth.MessageDigestHex = strings.ToUpper(auth.MessageDigestHex) },
		"short digest":             func(auth *FundAuthorization) { auth.MessageDigestHex = "00" },
		"zero required digest":     func(auth *FundAuthorization) { auth.PolicyDigestHex = strings.Repeat("0", 64) },
		"zero optional digest":     func(auth *FundAuthorization) { auth.MFADigestHex = strings.Repeat("0", 64) },
		"unknown MFA mode":         func(auth *FundAuthorization) { auth.MFAMode = MFAMode(255) },
		"possession MFA evidence":  func(auth *FundAuthorization) { auth.MFAMode = MFAPossessionOnlyPolicyApproved },
		"unknown eligibility mode": func(auth *FundAuthorization) { auth.EligibilityMode = EligibilityMode(255) },
		"unexpected eligibility":   func(auth *FundAuthorization) { auth.EligibilityMode = EligibilityNotRequired },
		"unknown phase":            func(auth *FundAuthorization) { auth.Phase = Phase(255) },
		"unknown effect":           func(auth *FundAuthorization) { auth.Effect = Effect(255) },
		"empty opaque ID":          func(auth *FundAuthorization) { auth.AccountID = "" },
		"noncanonical opaque ID":   func(auth *FundAuthorization) { auth.AccountID = " account:alice" },
		"oversized field":          func(auth *FundAuthorization) { auth.AccountID = strings.Repeat("a", maxCanonicalTextLength+1) },
		"zero lower bound":         func(auth *FundAuthorization) { auth.LowerBlock = 0 },
		"issuance lower mismatch":  func(auth *FundAuthorization) { auth.IssuedAtBlock++ },
		"reversed bounds":          func(auth *FundAuthorization) { auth.UpperBlock = auth.LowerBlock - 1 },
		"unknown expiry":           func(auth *FundAuthorization) { auth.Expiry.Kind = ExpiryKind(255) },
		"zero expiry":              func(auth *FundAuthorization) { auth.Expiry.Value = 0 },
		"expiry before lower":      func(auth *FundAuthorization) { auth.Expiry.Value = auth.LowerBlock - 1 },
		"expiry above upper":       func(auth *FundAuthorization) { auth.Expiry.Value = auth.UpperBlock + 1 },
		"bad account ID":           func(auth *FundAuthorization) { auth.AccountID = "Account:Alice" },
		"bad key ID":               func(auth *FundAuthorization) { auth.SignerKeyID = "device key" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			auth := cloneAuthorization(validAuthorization())
			mutate(&auth)
			if _, err := AuthorizationDigest(auth); err == nil {
				t.Fatal("malformed authorization accepted")
			}
		})
	}
}

func TestCanonicalRecoveryControlAllowsNoAmount(t *testing.T) {
	auth := validAuthorization()
	auth.SourceID = "/virtengine.veid.v1.MsgRebindWallet"
	auth.TypeURL = auth.SourceID
	auth.Phase = PhaseControl
	auth.Effect = EffectRecoveryControl
	auth.Amounts = nil
	auth.Parties = []PartyBinding{{Role: PartyRoleOwner, AccountID: auth.AccountID}}
	if _, err := AuthorizationDigest(auth); err != nil {
		t.Fatalf("zero-amount recovery control rejected: %v", err)
	}
}
