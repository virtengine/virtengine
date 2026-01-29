package types

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	providerkeeper "github.com/virtengine/virtengine/x/provider/keeper"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type marketplaceVEIDAdapter struct {
	keeper veidkeeper.Keeper
}

func newMarketplaceVEIDAdapter(keeper veidkeeper.Keeper) marketplacekeeper.VEIDKeeper {
	return marketplaceVEIDAdapter{keeper: keeper}
}

func (a marketplaceVEIDAdapter) GetIdentityScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool) {
	return a.keeper.GetVEIDScore(ctx, address)
}

func (a marketplaceVEIDAdapter) GetIdentityStatus(ctx sdk.Context, address sdk.AccAddress) (string, bool) {
	_, status, found := a.keeper.GetScore(ctx, address.String())
	if !found {
		return "", false
	}
	return string(status), true
}

func (a marketplaceVEIDAdapter) IsEmailVerified(ctx sdk.Context, address sdk.AccAddress) bool {
	scopes := a.keeper.GetScopesByType(ctx, address, veidtypes.ScopeTypeEmailProof)
	return hasVerifiedActiveScope(scopes)
}

func (a marketplaceVEIDAdapter) IsDomainVerified(ctx sdk.Context, address sdk.AccAddress) bool {
	scopes := a.keeper.GetScopesByType(ctx, address, veidtypes.ScopeTypeDomainVerify)
	return hasVerifiedActiveScope(scopes)
}

func hasVerifiedActiveScope(scopes []veidtypes.IdentityScope) bool {
	for _, scope := range scopes {
		if scope.IsVerified() && scope.IsActive() {
			return true
		}
	}
	return false
}

type marketplaceMFAAdapter struct {
	keeper mfakeeper.Keeper
}

func newMarketplaceMFAAdapter(keeper mfakeeper.Keeper) marketplacekeeper.MFAKeeper {
	return marketplaceMFAAdapter{keeper: keeper}
}

func (a marketplaceMFAAdapter) HasActiveFactors(ctx sdk.Context, address sdk.AccAddress) bool {
	enrollments := a.keeper.GetFactorEnrollments(ctx, address)
	for _, enrollment := range enrollments {
		if enrollment.Status == mfatypes.EnrollmentStatusActive {
			return true
		}
	}
	return false
}

func (a marketplaceMFAAdapter) GetLastMFAVerification(ctx sdk.Context, address sdk.AccAddress) (*time.Time, bool) {
	sessions := a.keeper.GetAccountSessions(ctx, address)
	var latest int64
	for _, session := range sessions {
		if session.CreatedAt > latest {
			latest = session.CreatedAt
		}
	}
	if latest == 0 {
		return nil, false
	}
	timestamp := time.Unix(latest, 0).UTC()
	return &timestamp, true
}

func (a marketplaceMFAAdapter) IsTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) bool {
	return a.keeper.IsTrustedDevice(ctx, address, fingerprint)
}

func (a marketplaceMFAAdapter) CreateChallenge(ctx sdk.Context, address sdk.AccAddress, actionType string) (string, error) {
	enrollments := a.keeper.GetFactorEnrollments(ctx, address)
	factor, ok := selectBestEnrollment(enrollments)
	if !ok {
		return "", mfatypes.ErrNoActiveFactors.Wrapf("no active MFA factors for %s", address.String())
	}

	params := a.keeper.GetParams(ctx)
	ttl := params.ChallengeTTL
	if ttl <= 0 {
		ttl = mfatypes.DefaultParams().ChallengeTTL
	}
	maxAttempts := params.MaxChallengeAttempts
	if maxAttempts == 0 {
		maxAttempts = mfatypes.DefaultParams().MaxChallengeAttempts
	}

	txType := resolveSensitiveTxType(actionType)
	challenge, err := mfatypes.NewChallenge(
		address.String(),
		factor.FactorType,
		factor.FactorID,
		txType,
		ttl,
		maxAttempts,
	)
	if err != nil {
		return "", err
	}

	if err := a.keeper.CreateChallenge(ctx, challenge); err != nil {
		return "", err
	}
	return challenge.ChallengeID, nil
}

func (a marketplaceMFAAdapter) VerifyChallenge(ctx sdk.Context, challengeID string, response interface{}) (bool, error) {
	switch r := response.(type) {
	case *mfatypes.ChallengeResponse:
		return a.keeper.VerifyMFAChallenge(ctx, challengeID, r)
	case mfatypes.ChallengeResponse:
		resp := r
		return a.keeper.VerifyMFAChallenge(ctx, challengeID, &resp)
	case []byte:
		return a.verifyChallengeBytes(ctx, challengeID, r)
	case string:
		return a.verifyChallengeBytes(ctx, challengeID, []byte(r))
	default:
		return false, mfatypes.ErrInvalidChallengeResponse.Wrap("unsupported response type")
	}
}

func (a marketplaceMFAAdapter) verifyChallengeBytes(ctx sdk.Context, challengeID string, data []byte) (bool, error) {
	challenge, found := a.keeper.GetChallenge(ctx, challengeID)
	if !found {
		return false, mfatypes.ErrChallengeNotFound.Wrapf("challenge %s not found", challengeID)
	}

	response := &mfatypes.ChallengeResponse{
		ChallengeID:  challengeID,
		FactorType:   challenge.FactorType,
		ResponseData: data,
	}
	return a.keeper.VerifyMFAChallenge(ctx, challengeID, response)
}

func selectBestEnrollment(enrollments []mfatypes.FactorEnrollment) (mfatypes.FactorEnrollment, bool) {
	var active []mfatypes.FactorEnrollment
	for _, enrollment := range enrollments {
		if enrollment.Status == mfatypes.EnrollmentStatusActive {
			active = append(active, enrollment)
		}
	}
	if len(active) == 0 {
		return mfatypes.FactorEnrollment{}, false
	}

	sort.SliceStable(active, func(i, j int) bool {
		a := active[i]
		b := active[j]
		securityA := a.FactorType.GetSecurityLevel()
		securityB := b.FactorType.GetSecurityLevel()
		if securityA != securityB {
			return securityA > securityB
		}
		if a.VerifiedAt != b.VerifiedAt {
			return a.VerifiedAt > b.VerifiedAt
		}
		return a.EnrolledAt > b.EnrolledAt
	})

	return active[0], true
}

func resolveSensitiveTxType(actionType string) mfatypes.SensitiveTransactionType {
	if actionType == "" {
		return mfatypes.SensitiveTxUnspecified
	}
	if txType, err := mfatypes.SensitiveTransactionTypeFromString(actionType); err == nil {
		return txType
	}

	switch actionType {
	case "place_order", "modify_order", "place_bid", "accept_bid", "settlement":
		return mfatypes.SensitiveTxHighValueOrder
	case "create_offering":
		return mfatypes.SensitiveTxFirstOfferingCreate
	case "key_rotation":
		return mfatypes.SensitiveTxKeyRotation
	default:
		return mfatypes.SensitiveTxUnspecified
	}
}

type marketplaceProviderAdapter struct {
	keeper providerkeeper.IKeeper
}

func newMarketplaceProviderAdapter(keeper providerkeeper.IKeeper) marketplacekeeper.ProviderKeeper {
	return marketplaceProviderAdapter{keeper: keeper}
}

func (a marketplaceProviderAdapter) IsProvider(ctx sdk.Context, address sdk.AccAddress) bool {
	return a.keeper.IsProvider(ctx, address)
}

func (a marketplaceProviderAdapter) GetProvider(ctx sdk.Context, address sdk.AccAddress) (interface{}, bool) {
	provider, found := a.keeper.Get(ctx, address)
	if !found {
		return nil, false
	}
	return provider, true
}

var _ marketplacekeeper.VEIDKeeper = marketplaceVEIDAdapter{}
var _ marketplacekeeper.MFAKeeper = marketplaceMFAAdapter{}
var _ marketplacekeeper.ProviderKeeper = marketplaceProviderAdapter{}
