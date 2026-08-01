package fundauth

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"
)

func validPolicyContext(auth FundAuthorization, binding TransactionBinding) AuthorizationPolicyContext {
	return AuthorizationPolicyContext{
		AccountID: auth.AccountID, AccountState: AuthorizationAccountActive,
		AccountRevision: 11, AuthorizedAccountRevision: 11, KeyEpoch: auth.SignerKeyEpoch,
		PolicyDigestHex: auth.PolicyDigestHex, PolicyRevision: 13, AuthorizedPolicyRevision: 13,
		FactorMode: FactorPossessionPlusMFA, MFADigestHex: auth.MFADigestHex,
		EligibilityMode: EligibilityEvidenceRequired, EligibilityState: EligibilityEligible,
		EligibilityDigestHex: auth.EligibilityDigestHex, EligibilityFreshAtBlock: binding.CurrentBlock,
		EligibilityFreshAtTime: binding.CurrentTime, CurrentBlock: binding.CurrentBlock, CurrentTime: binding.CurrentTime,
	}
}

func TestAuthorizationPolicyAccountStates(t *testing.T) {
	signed, _, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	for _, state := range []AuthorizationAccountState{AuthorizationAccountActive, AuthorizationAccountSuspended, AuthorizationAccountHeld, AuthorizationAccountRecoveryPending, AuthorizationAccountTerminated, AuthorizationAccountClosed} {
		policy := validPolicyContext(signed.Authorization, binding)
		policy.AccountState = state
		err := ValidateAuthorizationPolicy(context.Background(), signed.Authorization, binding, policy)
		if state == AuthorizationAccountActive && err != nil {
			t.Fatalf("active: %v", err)
		}
		if state != AuthorizationAccountActive && !errors.Is(err, ErrPolicyDenied) {
			t.Fatalf("state %d accepted: %v", state, err)
		}
	}
}

func TestVerifyPolicyAndConsumeOrdering(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	policy := validPolicyContext(signed.Authorization, binding)

	consumer := newMemoryConsumer()
	badPolicy := policy
	badPolicy.PolicyRevision++
	if err := VerifyPolicyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, badPolicy, consumer, func(context.Context) error { return nil }); !errors.Is(err, ErrPolicyDenied) {
		t.Fatalf("stale policy error = %v", err)
	}
	if consumer.calls.Load() != 0 {
		t.Fatal("consumer called before policy gate completed")
	}

	badSignature := signed
	badSignature.Signature = nil
	if err := VerifyPolicyAndConsume(context.Background(), badSignature, DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error { return nil }); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("missing possession signature error = %v", err)
	}
	if consumer.calls.Load() != 0 {
		t.Fatal("consumer called before possession proof completed")
	}
	badSignature.Signature = append([]byte(nil), signed.Signature...)
	badSignature.Signature[0] ^= 0xff
	if err := VerifyPolicyAndConsume(context.Background(), badSignature, DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error { return nil }); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("bad possession signature error = %v", err)
	}

	if err := VerifyPolicyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPolicyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error { return nil }); !errors.Is(err, errFixtureReplay) {
		t.Fatalf("replay error = %v", err)
	}

	rollbackConsumer := newMemoryConsumer()
	callbackFailure := errors.New("policy-protected callback failed")
	if err := VerifyPolicyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, policy, rollbackConsumer, func(context.Context) error { return callbackFailure }); !errors.Is(err, callbackFailure) {
		t.Fatalf("callback error = %v", err)
	}
	if err := VerifyPolicyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, policy, rollbackConsumer, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("retry after callback rollback: %v", err)
	}
}

func TestAuthorizationPolicyStalenessAndEvidence(t *testing.T) {
	signed, _, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	base := validPolicyContext(signed.Authorization, binding)
	tests := map[string]func(*AuthorizationPolicyContext){
		"account revision":        func(value *AuthorizationPolicyContext) { value.AccountRevision++ },
		"key epoch":               func(value *AuthorizationPolicyContext) { value.KeyEpoch++ },
		"policy digest":           func(value *AuthorizationPolicyContext) { value.PolicyDigestHex = testDigest("changed policy") },
		"policy revision":         func(value *AuthorizationPolicyContext) { value.PolicyRevision++ },
		"MFA missing":             func(value *AuthorizationPolicyContext) { value.MFADigestHex = "" },
		"MFA mismatch":            func(value *AuthorizationPolicyContext) { value.MFADigestHex = testDigest("changed mfa") },
		"eligibility ineligible":  func(value *AuthorizationPolicyContext) { value.EligibilityState = EligibilityIneligible },
		"eligibility stale state": func(value *AuthorizationPolicyContext) { value.EligibilityState = EligibilityStale },
		"eligibility unavailable": func(value *AuthorizationPolicyContext) { value.EligibilityState = EligibilityUnavailable },
		"eligibility mismatch": func(value *AuthorizationPolicyContext) {
			value.EligibilityDigestHex = testDigest("changed eligibility")
		},
		"eligibility stale block": func(value *AuthorizationPolicyContext) { value.EligibilityFreshAtBlock-- },
		"eligibility stale time": func(value *AuthorizationPolicyContext) {
			value.EligibilityFreshAtTime = value.CurrentTime.Add(-time.Second)
		},
		"hold recovery digest": func(value *AuthorizationPolicyContext) { value.HoldRecoveryDigestHex = testDigest("hold") },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			policy := base
			mutate(&policy)
			if err := ValidateAuthorizationPolicy(context.Background(), signed.Authorization, binding, policy); !errors.Is(err, ErrPolicyDenied) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestPossessionOnlyPolicyAndCancellation(t *testing.T) {
	signed, resolver, privateKey := signedFixture(t)
	signed.Authorization.MFAMode = MFAPossessionOnlyPolicyApproved
	signed.Authorization.MFADigestHex = ""
	signed.Authorization.EligibilityMode = EligibilityNotRequired
	signed.Authorization.EligibilityDigestHex = ""
	signBytes, _, err := CanonicalSignBytes(signed.Authorization)
	if err != nil {
		t.Fatal(err)
	}
	signed.Signature = ed25519.Sign(privateKey, signBytes)
	binding := verifyOptions(signed.Authorization)
	policy := validPolicyContext(signed.Authorization, binding)
	policy.FactorMode = FactorPossessionOnlyPolicyApproved
	policy.MFADigestHex = ""
	policy.EligibilityMode = EligibilityNotRequired
	policy.EligibilityState = 0
	policy.EligibilityDigestHex = ""
	policy.EligibilityFreshAtBlock = 0
	policy.EligibilityFreshAtTime = time.Time{}
	if err := VerifyPolicyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, policy, newMemoryConsumer(), func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	consumer := newMemoryConsumer()
	if err := VerifyPolicyAndConsume(canceled, signed, DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	if consumer.calls.Load() != 0 {
		t.Fatal("consumer called for canceled context")
	}
}
