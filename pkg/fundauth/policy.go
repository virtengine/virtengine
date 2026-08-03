package fundauth

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrPolicyDenied = errors.New("fund authorization policy denied")

type AuthorizationAccountState uint8

const (
	AuthorizationAccountActive AuthorizationAccountState = iota + 1
	AuthorizationAccountSuspended
	AuthorizationAccountHeld
	AuthorizationAccountRecoveryPending
	AuthorizationAccountTerminated
	AuthorizationAccountClosed
)

type FactorMode uint8

const (
	FactorPossessionOnlyPolicyApproved FactorMode = iota + 1
	FactorPossessionPlusMFA
)

type EligibilityState uint8

const (
	EligibilityEligible EligibilityState = iota + 1
	EligibilityIneligible
	EligibilityStale
	EligibilityUnavailable
)

// AuthorizationPolicyContext is a fail-closed snapshot supplied by the
// account and policy keepers. Authorized revisions are the revisions captured
// when the authorization policy was approved; current revisions are read at
// execution time.
type AuthorizationPolicyContext struct {
	AccountID                 string
	AccountState              AuthorizationAccountState
	AccountRevision           uint64
	AuthorizedAccountRevision uint64
	KeyEpoch                  uint64
	PolicyDigestHex           string
	PolicyRevision            uint64
	AuthorizedPolicyRevision  uint64
	FactorMode                FactorMode
	MFADigestHex              string
	EligibilityMode           EligibilityMode
	EligibilityState          EligibilityState
	EligibilityDigestHex      string
	EligibilityFreshAtBlock   uint64
	EligibilityFreshAtTime    time.Time
	CurrentBlock              uint64
	CurrentTime               time.Time
	HoldRecoveryDigestHex     string
}

func ValidateAuthorizationPolicy(ctx context.Context, auth FundAuthorization, binding TransactionBinding, policy AuthorizationPolicyContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if policy.AccountID == "" || policy.AccountRevision == 0 || policy.AuthorizedAccountRevision == 0 || policy.KeyEpoch == 0 || policy.PolicyRevision == 0 || policy.AuthorizedPolicyRevision == 0 || policy.CurrentBlock == 0 || policy.CurrentTime.IsZero() {
		return fmt.Errorf("%w: incomplete policy context", ErrPolicyDenied)
	}
	if policy.AccountState != AuthorizationAccountActive {
		return fmt.Errorf("%w: account is not active", ErrPolicyDenied)
	}
	if policy.HoldRecoveryDigestHex != "" {
		return fmt.Errorf("%w: active account has hold or recovery state", ErrPolicyDenied)
	}
	if policy.AccountID != auth.AccountID || policy.AccountID != binding.AccountID || policy.AccountRevision != policy.AuthorizedAccountRevision {
		return fmt.Errorf("%w: stale account revision", ErrPolicyDenied)
	}
	if policy.KeyEpoch != auth.SignerKeyEpoch || policy.KeyEpoch != binding.SignerKeyEpoch {
		return fmt.Errorf("%w: stale key epoch", ErrPolicyDenied)
	}
	if policy.PolicyDigestHex != auth.PolicyDigestHex || policy.PolicyDigestHex != binding.PolicyDigestHex || policy.PolicyRevision != policy.AuthorizedPolicyRevision {
		return fmt.Errorf("%w: stale policy", ErrPolicyDenied)
	}
	if policy.CurrentBlock != binding.CurrentBlock || !policy.CurrentTime.Equal(binding.CurrentTime) {
		return fmt.Errorf("%w: execution coordinate mismatch", ErrPolicyDenied)
	}

	switch policy.FactorMode {
	case FactorPossessionOnlyPolicyApproved:
		if auth.MFAMode != MFAPossessionOnlyPolicyApproved || binding.MFAMode != MFAPossessionOnlyPolicyApproved || auth.MFADigestHex != "" || binding.MFADigestHex != "" || policy.MFADigestHex != "" {
			return fmt.Errorf("%w: possession-only policy mismatch", ErrPolicyDenied)
		}
	case FactorPossessionPlusMFA:
		if auth.MFAMode != MFAEvidenceRequired || binding.MFAMode != MFAEvidenceRequired || policy.MFADigestHex == "" || policy.MFADigestHex != auth.MFADigestHex || policy.MFADigestHex != binding.MFADigestHex {
			return fmt.Errorf("%w: MFA evidence missing or mismatched", ErrPolicyDenied)
		}
	default:
		return fmt.Errorf("%w: unknown factor mode", ErrPolicyDenied)
	}

	switch policy.EligibilityMode {
	case EligibilityNotRequired:
		if auth.EligibilityMode != EligibilityNotRequired || binding.EligibilityMode != EligibilityNotRequired || auth.EligibilityDigestHex != "" || binding.EligibilityDigestHex != "" || policy.EligibilityDigestHex != "" {
			return fmt.Errorf("%w: unexpected eligibility evidence", ErrPolicyDenied)
		}
	case EligibilityEvidenceRequired:
		if auth.EligibilityMode != EligibilityEvidenceRequired || binding.EligibilityMode != EligibilityEvidenceRequired || policy.EligibilityState != EligibilityEligible || policy.EligibilityDigestHex == "" || policy.EligibilityDigestHex != auth.EligibilityDigestHex || policy.EligibilityDigestHex != binding.EligibilityDigestHex {
			return fmt.Errorf("%w: eligibility evidence missing, mismatched, or ineligible", ErrPolicyDenied)
		}
		if policy.EligibilityFreshAtBlock < policy.CurrentBlock || policy.EligibilityFreshAtTime.Before(policy.CurrentTime) {
			return fmt.Errorf("%w: stale eligibility evidence", ErrPolicyDenied)
		}
	default:
		return fmt.Errorf("%w: unknown eligibility mode", ErrPolicyDenied)
	}
	return ctx.Err()
}

func VerifyPolicyAndConsume(ctx context.Context, signed SignedAuthorization, registry *Registry, resolver Ed25519KeyResolver, binding TransactionBinding, policy AuthorizationPolicyContext, consumer AtomicAuthorizationConsumer, protected func(context.Context) error) error {
	if consumer == nil || !consumer.KeeperRequired() || protected == nil {
		return fmt.Errorf("%w: nil consumer or callback", ErrInvalidAuthorization)
	}
	authDigest, err := Verify(ctx, signed, registry, resolver, binding)
	if err != nil {
		return err
	}
	if err := ValidateAuthorizationPolicy(ctx, signed.Authorization, binding, policy); err != nil {
		return err
	}
	nonceDigest, err := parseRequiredDigest(signed.Authorization.NonceDigestHex, "nonce")
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return consumer.WithAuthorization(ctx, signed.Authorization.AccountID, nonceDigest, authDigest, protected)
}
