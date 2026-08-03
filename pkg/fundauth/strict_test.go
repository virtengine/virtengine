package fundauth

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"
)

func TestVerifyRequiresCompleteTransactionBinding(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	omissions := map[string]func(*TransactionBinding){
		"domain":        func(value *TransactionBinding) { value.Domain = "" },
		"version":       func(value *TransactionBinding) { value.Version = 0 },
		"chain":         func(value *TransactionBinding) { value.ChainID = "" },
		"account":       func(value *TransactionBinding) { value.AccountID = "" },
		"key":           func(value *TransactionBinding) { value.SignerKeyID = "" },
		"key epoch":     func(value *TransactionBinding) { value.SignerKeyEpoch = 0 },
		"source":        func(value *TransactionBinding) { value.SourceID = "" },
		"message":       func(value *TransactionBinding) { value.MessageDigestHex = "" },
		"policy":        func(value *TransactionBinding) { value.PolicyDigestHex = "" },
		"nonce":         func(value *TransactionBinding) { value.NonceDigestHex = "" },
		"issued block":  func(value *TransactionBinding) { value.IssuedAtBlock = 0 },
		"issued time":   func(value *TransactionBinding) { value.IssuedAtUnix = 0 },
		"current block": func(value *TransactionBinding) { value.CurrentBlock = 0 },
		"current time":  func(value *TransactionBinding) { value.CurrentTime = time.Time{} },
		"clock skew":    func(value *TransactionBinding) { value.MaxClockSkew = 0 },
		"lifetime":      func(value *TransactionBinding) { value.MaxLifetime = 0 },
	}
	for name, omit := range omissions {
		t.Run(name, func(t *testing.T) {
			changed := binding
			omit(&changed)
			if _, err := Verify(context.Background(), signed, DefaultRegistry(), resolver, changed); err == nil {
				t.Fatal("incomplete transaction binding accepted")
			}
		})
	}
}

func TestVerifyComparesEveryAmountPartyAndPolicyClaim(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	mismatches := map[string]func(*TransactionBinding){
		"amount value":   func(value *TransactionBinding) { value.Amounts[0].MinorUnits = "43" },
		"amount denom":   func(value *TransactionBinding) { value.Amounts[0].Denom = "abc" },
		"missing amount": func(value *TransactionBinding) { value.Amounts = value.Amounts[1:] },
		"extra amount": func(value *TransactionBinding) {
			value.Amounts = append(value.Amounts, Amount{Denom: "zzz", MinorUnits: "1"})
		},
		"party account": func(value *TransactionBinding) { value.Parties[1].AccountID = "account:mallory" },
		"party role":    func(value *TransactionBinding) { value.Parties[1].Role = PartyRolePayee },
		"missing party": func(value *TransactionBinding) { value.Parties = value.Parties[:1] },
		"extra party": func(value *TransactionBinding) {
			value.Parties = append(value.Parties, PartyBinding{Role: PartyRoleOwner, AccountID: "account:alice"})
		},
		"MFA mode":         func(value *TransactionBinding) { value.MFAMode = MFAPossessionOnlyPolicyApproved },
		"eligibility mode": func(value *TransactionBinding) { value.EligibilityMode = EligibilityNotRequired },
		"case":             func(value *TransactionBinding) { value.CaseDigestHex = "" },
		"order":            func(value *TransactionBinding) { value.OrderDigestHex = "" },
		"reference":        func(value *TransactionBinding) { value.ReferenceDigestHex = "" },
	}
	for name, mismatch := range mismatches {
		t.Run(name, func(t *testing.T) {
			changed := binding
			changed.Amounts = append([]Amount(nil), binding.Amounts...)
			changed.Parties = append([]PartyBinding(nil), binding.Parties...)
			mismatch(&changed)
			if _, err := Verify(context.Background(), signed, DefaultRegistry(), resolver, changed); err == nil {
				t.Fatal("transaction binding mismatch accepted")
			}
		})
	}
}

func TestVerifyEnforcesDescriptorAndEvidenceModes(t *testing.T) {
	signed, resolver, privateKey := signedFixture(t)
	for name, mutate := range map[string]func(*FundAuthorization){
		"zero amount":           func(auth *FundAuthorization) { auth.Amounts[0].MinorUnits = "0" },
		"missing required role": func(auth *FundAuthorization) { auth.Parties = auth.Parties[:1] },
		"extra role": func(auth *FundAuthorization) {
			auth.Parties = append(auth.Parties, PartyBinding{Role: PartyRoleOwner, AccountID: auth.AccountID})
		},
		"account wrong role": func(auth *FundAuthorization) {
			auth.Parties[0].AccountID = "account:mallory"
			auth.Parties[1].AccountID = auth.AccountID
		},
		"MFA required empty":         func(auth *FundAuthorization) { auth.MFADigestHex = "" },
		"possession MFA evidence":    func(auth *FundAuthorization) { auth.MFAMode = MFAPossessionOnlyPolicyApproved },
		"eligibility required empty": func(auth *FundAuthorization) { auth.EligibilityDigestHex = "" },
		"unexpected eligibility":     func(auth *FundAuthorization) { auth.EligibilityMode = EligibilityNotRequired },
	} {
		t.Run(name, func(t *testing.T) {
			changed := signed
			changed.Authorization = cloneAuthorization(signed.Authorization)
			mutate(&changed.Authorization)
			if signBytes, _, err := CanonicalSignBytes(changed.Authorization); err == nil {
				changed.Signature = ed25519.Sign(privateKey, signBytes)
			}
			if _, err := Verify(context.Background(), changed, DefaultRegistry(), resolver, verifyOptions(changed.Authorization)); err == nil {
				t.Fatal("invalid descriptor or evidence policy accepted")
			}
		})
	}
}

func TestVerifyRejectsInactiveAliasedKeysAndCancellation(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	inactive := resolver
	inactive.active = false
	if _, err := Verify(context.Background(), signed, DefaultRegistry(), inactive, binding); err == nil {
		t.Fatal("inactive possession key accepted")
	}
	if _, err := Verify(context.Background(), signed, DefaultRegistry(), aliasResolver{key: ResolvedPossessionKey{AccountID: "account:alias", KeyID: signed.Authorization.SignerKeyID, Epoch: signed.Authorization.SignerKeyEpoch, PublicKey: resolver.publicKey, Active: true}}, binding); err == nil {
		t.Fatal("aliased possession key accepted")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Verify(canceled, signed, DefaultRegistry(), resolver, binding); !errors.Is(err, context.Canceled) {
		t.Fatalf("entry cancellation error = %v", err)
	}
	afterResolve, cancelAfterResolve := context.WithCancel(context.Background())
	if _, err := Verify(afterResolve, signed, DefaultRegistry(), cancelingResolver{fixedResolver: resolver, cancel: cancelAfterResolve}, binding); !errors.Is(err, context.Canceled) {
		t.Fatalf("post-resolver cancellation error = %v", err)
	}
}

func TestVerifyEnforcesTrustedClockPolicy(t *testing.T) {
	signed, resolver, privateKey := signedFixture(t)

	future := signed
	future.Authorization.IssuedAtUnix = 1_800_000_100
	bytesToSign, _, err := CanonicalSignBytes(future.Authorization)
	if err != nil {
		t.Fatal(err)
	}
	future.Signature = ed25519.Sign(privateKey, bytesToSign)
	if _, err = Verify(context.Background(), future, DefaultRegistry(), resolver, verifyOptions(future.Authorization)); err == nil {
		t.Fatal("authorization beyond maximum clock skew accepted")
	}

	staleBinding := verifyOptions(signed.Authorization)
	staleBinding.CurrentTime = time.Unix(1_800_010_000, 0)
	if _, err = Verify(context.Background(), signed, DefaultRegistry(), resolver, staleBinding); err == nil {
		t.Fatal("authorization beyond maximum lifetime accepted")
	}
}

type aliasResolver struct{ key ResolvedPossessionKey }

func (resolver aliasResolver) ResolveEd25519(context.Context, string, string, uint64) (ResolvedPossessionKey, error) {
	return resolver.key, nil
}

type cancelingResolver struct {
	fixedResolver
	cancel context.CancelFunc
}

func (resolver cancelingResolver) ResolveEd25519(ctx context.Context, accountID, keyID string, epoch uint64) (ResolvedPossessionKey, error) {
	resolved, err := resolver.fixedResolver.ResolveEd25519(ctx, accountID, keyID, epoch)
	resolver.cancel()
	return resolved, err
}
