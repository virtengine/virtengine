package fundauth

import (
	"context"
	"crypto/ed25519"
	"testing"
	"time"
)

func TestVerifyBindsPossessionContextAndAllFields(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	opts := verifyOptions(signed.Authorization)
	if _, err := Verify(context.Background(), signed, DefaultRegistry(), resolver, opts); err != nil {
		t.Fatalf("valid authorization rejected: %v", err)
	}

	mutations := map[string]func(*FundAuthorization){
		"chain":       func(auth *FundAuthorization) { auth.ChainID += "-wrong" },
		"account":     func(auth *FundAuthorization) { auth.AccountID += "-wrong" },
		"source":      func(auth *FundAuthorization) { auth.SourceID = "/cosmos.bank.v1beta1.MsgMultiSend" },
		"type":        func(auth *FundAuthorization) { auth.TypeURL = "/cosmos.bank.v1beta1.MsgMultiSend" },
		"message":     func(auth *FundAuthorization) { auth.MessageDigestHex = testDigest("wrong-message") },
		"amount":      func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "99" },
		"party":       func(auth *FundAuthorization) { auth.Parties[1].AccountID = "account:mallory" },
		"case":        func(auth *FundAuthorization) { auth.CaseDigestHex = testDigest("wrong-case") },
		"order":       func(auth *FundAuthorization) { auth.OrderDigestHex = testDigest("wrong-order") },
		"MFA":         func(auth *FundAuthorization) { auth.MFADigestHex = testDigest("wrong-mfa") },
		"eligibility": func(auth *FundAuthorization) { auth.EligibilityDigestHex = testDigest("wrong-eligibility") },
		"policy":      func(auth *FundAuthorization) { auth.PolicyDigestHex = testDigest("wrong-policy") },
		"nonce":       func(auth *FundAuthorization) { auth.NonceDigestHex = testDigest("wrong-nonce") },
		"lower bound": func(auth *FundAuthorization) { auth.LowerBlock = 151 },
		"upper bound": func(auth *FundAuthorization) { auth.UpperBlock = 149 },
		"expiry":      func(auth *FundAuthorization) { auth.Expiry.Value = 149 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			changed := signed
			changed.Authorization = cloneAuthorization(signed.Authorization)
			mutate(&changed.Authorization)
			if _, err := Verify(context.Background(), changed, DefaultRegistry(), resolver, opts); err == nil {
				t.Fatal("mutated authorization accepted")
			}
		})
	}

	t.Run("signature", func(t *testing.T) {
		changed := signed
		changed.Signature = append([]byte(nil), signed.Signature...)
		changed.Signature[0] ^= 1
		if _, err := Verify(context.Background(), changed, DefaultRegistry(), resolver, opts); err == nil {
			t.Fatal("wrong signature accepted")
		}
	})
	t.Run("biometric only", func(t *testing.T) {
		changed := signed
		changed.Signature = nil
		if changed.Authorization.MFADigestHex == "" || changed.Authorization.EligibilityDigestHex == "" {
			t.Fatal("fixture lacks supplemental evidence")
		}
		if _, err := Verify(context.Background(), changed, DefaultRegistry(), resolver, opts); err == nil {
			t.Fatal("supplemental evidence replaced possession signature")
		}
	})
}

func TestVerifyRejectsSubstitutionEpochBoundsAndExpiry(t *testing.T) {
	signed, resolver, privateKey := signedFixture(t)
	opts := verifyOptions(signed.Authorization)

	confused := signed
	confused.Authorization.SourceID = "/cosmos.bank.v1beta1.MsgMultiSend"
	if _, err := Verify(context.Background(), confused, DefaultRegistry(), resolver, opts); err == nil {
		t.Fatal("same-message source/type confusion accepted")
	}
	unknown := signed
	unknown.Authorization.SourceID, unknown.Authorization.TypeURL = "/unknown", "/unknown"
	if _, err := Verify(context.Background(), unknown, DefaultRegistry(), resolver, opts); err == nil {
		t.Fatal("unknown route accepted")
	}

	epoch := signed
	epoch.Authorization.SignerKeyEpoch++
	bytesToSign, _, err := CanonicalSignBytes(epoch.Authorization)
	if err != nil {
		t.Fatal(err)
	}
	epoch.Signature = ed25519.Sign(privateKey, bytesToSign)
	if _, err := Verify(context.Background(), epoch, DefaultRegistry(), resolver, opts); err == nil {
		t.Fatal("wrong key epoch accepted")
	}

	for name, currentBlock := range map[string]uint64{"below": 99, "above": 201} {
		t.Run(name+" bounds", func(t *testing.T) {
			changedOpts := opts
			changedOpts.CurrentBlock = currentBlock
			if _, err := Verify(context.Background(), signed, DefaultRegistry(), resolver, changedOpts); err == nil {
				t.Fatal("out-of-bounds authorization accepted")
			}
		})
	}
	expired := signed
	expired.Authorization.Expiry = ExpiryCoordinate{Kind: ExpiryAtUnixTime, Value: 1_700_000_000}
	bytesToSign, _, err = CanonicalSignBytes(expired.Authorization)
	if err != nil {
		t.Fatal(err)
	}
	expired.Signature = ed25519.Sign(privateKey, bytesToSign)
	expiredOpts := opts
	expiredOpts.CurrentTime = time.Unix(1_800_000_000, 0)
	if _, err := Verify(context.Background(), expired, DefaultRegistry(), resolver, expiredOpts); err == nil {
		t.Fatal("expired authorization accepted")
	}

	wrongExpected := opts
	wrongExpected.MessageDigestHex = testDigest("different expected message")
	if _, err := Verify(context.Background(), signed, DefaultRegistry(), resolver, wrongExpected); err == nil {
		t.Fatal("wrong expected message accepted")
	}
}
