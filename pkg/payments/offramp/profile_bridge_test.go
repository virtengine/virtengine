package offramp

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

type profiledTestAdapter struct {
	*contractProvider
	profile         PayoutProfile
	profileTestOnly bool
}

type acceptingProfileAuthorizer struct{}

func (acceptingProfileAuthorizer) AuthorizePayoutProfile(PayoutProfile) error { return nil }

type durableTestPayoutRepository struct {
	mu      sync.Mutex
	payouts map[string]PayoutResult
}

func newDurableTestPayoutRepository() *durableTestPayoutRepository {
	return &durableTestPayoutRepository{payouts: make(map[string]PayoutResult)}
}

func (r *durableTestPayoutRepository) GetPayout(_ context.Context, payoutID string) (PayoutResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result, ok := r.payouts[payoutID]
	if !ok {
		return PayoutResult{}, ErrPayoutNotFound
	}
	return clonePayoutResult(result), nil
}

func (r *durableTestPayoutRepository) FindPayout(_ context.Context, provider string, metadata map[string]string) (PayoutResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, result := range r.payouts {
		if (provider == "" || result.Provider == provider) && metadataMatches(result.Metadata, metadata) {
			return clonePayoutResult(result), nil
		}
	}
	return PayoutResult{}, ErrPayoutNotFound
}

func (r *durableTestPayoutRepository) PutPayout(_ context.Context, result PayoutResult) error {
	if result.ID == "" {
		return errors.New("missing payout ID")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.payouts[result.ID] = clonePayoutResult(result)
	return nil
}

func (*durableTestPayoutRepository) Durable() bool { return true }

type restartRecoveryAdapter struct {
	*contractProvider
	recovered  PayoutResult
	restoreErr error
	restores   int
}

func (a *restartRecoveryAdapter) RestorePayoutBinding(_ context.Context, expected PayoutResult) (PayoutResult, error) {
	a.restores++
	if a.restoreErr != nil {
		return PayoutResult{}, a.restoreErr
	}
	a.contractProvider.mu.Lock()
	a.contractProvider.storeResultLocked(a.recovered)
	a.contractProvider.mu.Unlock()
	return clonePayoutResult(a.recovered), nil
}

func (a *profiledTestAdapter) Profile() PayoutProfile { return a.profile }

func (a *profiledTestAdapter) IsTestOnly() bool { return a.profileTestOnly }

func TestProductionBridgeRejectsMockAndExternalBlockedProfiles(t *testing.T) {
	t.Parallel()
	bridge := NewProductionBridge()
	require.ErrorIs(t, bridge.RegisterAdapter(NewMockProvider("mock", []string{"USD"}, []string{"ach"})), ErrTestAdapter)

	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	blocked := &profiledTestAdapter{
		contractProvider: newContractProvider("contract-partner", clock.Now, Quote{}),
		profile:          blockedSandboxProfile(),
	}
	require.ErrorIs(t, bridge.RegisterAdapter(blocked), ErrProfileNotExecutable)

	uncertified := blockedSandboxProfile()
	uncertified.Environment = EnvironmentProduction
	uncertified.CredentialSecretRefs[0].Scope = string(EnvironmentProduction)
	uncertified.CredentialSecretRefs[0].Ref = "vault://offramp/production/api"
	blocked.profile = uncertified
	require.ErrorIs(t, bridge.RegisterAdapter(blocked), ErrProfileNotExecutable)
}

func TestProductionBridgeAcceptsCertifiedProfileButRejectsTestMarker(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter := &profiledTestAdapter{
		contractProvider: newContractProvider("contract-partner", clock.Now, Quote{}),
		profile:          certifiedProductionProfile(), profileTestOnly: true,
	}
	bridge := NewProductionBridge(ProductionBridgeConfig{Repository: newDurableTestPayoutRepository(), Authorizer: acceptingProfileAuthorizer{}})
	require.ErrorIs(t, bridge.RegisterAdapter(adapter), ErrTestAdapter)
	adapter.profileTestOnly = false
	require.ErrorIs(t, bridge.RegisterAdapter(adapter), ErrProfileNotExecutable, "only the package-sealed real HTTP adapter is production eligible")
}

func TestBridgeRestartRestoresAdapterBindingBeforeStatusPoll(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	repository := newDurableTestPayoutRepository()
	metadata := map[string]string{"conversion_id": "conversion-restart", "idempotency_key": "idempotency-restart", "correlation_id": "correlation-restart"}
	expected := PayoutResult{ID: "payout-restart", QuoteID: "quote-restart", Provider: "partner", Status: StatusProcessing,
		FiatAmount: sdkmath.LegacyNewDec(100), CryptoAmount: sdkmath.NewInt(99), Fee: sdkmath.NewInt(1), Reference: "reference-restart",
		Metadata: metadata, InitiatedAt: clock.Now(), StatusUpdatedAt: clock.Now(), DailyReservationKey: "2026-07-23|US-USD-ach", DailyReservationOperationID: metadata["idempotency_key"]}
	require.NoError(t, repository.PutPayout(ctx, expected))
	completedAt := clock.Now().Add(time.Minute)
	recovered := clonePayoutResult(expected)
	recovered.Status, recovered.StatusUpdatedAt, recovered.CompletedAt = StatusCompleted, completedAt, &completedAt
	provider := &restartRecoveryAdapter{contractProvider: newContractProvider("partner", clock.Now, Quote{}), recovered: recovered}
	bridge := newBridgeWithDependencies(ExecutionModeLegacy, false, repository, nil, clock.Now)
	require.NoError(t, bridge.RegisterAdapter(provider))
	result, err := bridge.GetStatus(ctx, expected.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, result.Status)
	require.Equal(t, 1, provider.restores)
	require.Equal(t, "durable_binding_restore", result.AuditFields["bridge_recovery_reason"])
}

func TestBridgeRestartRejectsRecoveredMismatchAndAmbiguity(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		mutate     func(*PayoutResult)
		restoreErr error
	}{
		{name: "payout id mismatch", mutate: func(result *PayoutResult) { result.ID = "different-payout" }},
		{name: "quote mismatch", mutate: func(result *PayoutResult) { result.QuoteID = "different-quote" }},
		{name: "economic mismatch", mutate: func(result *PayoutResult) { result.FiatAmount = sdkmath.LegacyNewDec(101) }},
		{name: "ambiguous lookup", restoreErr: &ProviderError{Provider: "partner", Operation: operationMetadataLookup, Kind: ErrorKindAmbiguous, Retryable: true, Ambiguous: true, err: ErrProviderAmbiguous}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
			repository := newDurableTestPayoutRepository()
			expected := PayoutResult{ID: "payout-restart", QuoteID: "quote-restart", Provider: "partner", Status: StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(100), CryptoAmount: sdkmath.NewInt(99), Fee: sdkmath.NewInt(1), Reference: "reference-restart",
				Metadata: map[string]string{"conversion_id": "conversion-restart", "idempotency_key": "idempotency-restart", "correlation_id": "correlation-restart"}, InitiatedAt: clock.Now(), StatusUpdatedAt: clock.Now()}
			require.NoError(t, repository.PutPayout(ctx, expected))
			recovered := clonePayoutResult(expected)
			if test.mutate != nil {
				test.mutate(&recovered)
			}
			provider := &restartRecoveryAdapter{contractProvider: newContractProvider("partner", clock.Now, Quote{}), recovered: recovered, restoreErr: test.restoreErr}
			bridge := newBridgeWithDependencies(ExecutionModeLegacy, false, repository, nil, clock.Now)
			require.NoError(t, bridge.RegisterAdapter(provider))
			_, err := bridge.GetStatus(ctx, expected.ID)
			require.Error(t, err)
			if test.restoreErr != nil {
				require.True(t, IsAmbiguous(err))
			} else {
				require.ErrorIs(t, err, ErrProviderRejected)
			}
		})
	}
}

func TestEngineeringSandboxRequiresExplicitExternalBlockedOptIn(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter := &profiledTestAdapter{
		contractProvider: newContractProvider("contract-partner", clock.Now, Quote{
			ID: "sandbox-quote", FiatAmount: sdkmath.LegacyNewDec(100), ExchangeRate: sdkmath.LegacyOneDec(), Fee: sdkmath.NewInt(1),
		}),
		profile: blockedSandboxProfile(),
	}
	adapter.methods = map[string]bool{"ach": true}
	require.ErrorIs(t, NewEngineeringSandboxBridge(false).RegisterAdapter(adapter), ErrProfileNotExecutable)
	bridge := newBridgeWithOptions(ExecutionModeSandbox, true, clock.Now)
	require.NoError(t, bridge.RegisterAdapter(adapter))
	req := httpQuoteRequest(clock.Now())
	quote, err := bridge.GetQuote(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "contract-partner", quote.Provider)
}

func TestCertifiedProfileFailsClosedWithoutEveryEvidenceClass(t *testing.T) {
	t.Parallel()
	profile := certifiedProductionProfile()
	require.NoError(t, profile.ValidateForExecution(ExecutionModeProduction, false))
	profile.Evidence.DPA = ApprovalEvidence{}
	require.ErrorIs(t, profile.ValidateForExecution(ExecutionModeProduction, false), ErrProfileNotExecutable)
	profile = certifiedProductionProfile()
	profile.CredentialSecretRefs = nil
	require.ErrorIs(t, profile.ValidateForExecution(ExecutionModeProduction, false), ErrProfileNotExecutable)
}

func TestNonExecutableProfileStatesFailClosed(t *testing.T) {
	t.Parallel()
	for _, state := range []ProfileState{ProfileUnsupported, ProfileEngineeringIncomplete, ProfilePaused} {
		profile := blockedSandboxProfile()
		profile.State = state
		require.ErrorIs(t, profile.ValidateForExecution(ExecutionModeSandbox, true), ErrProfileNotExecutable, string(state))
	}
}
