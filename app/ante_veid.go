package app

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	roleskeeper "github.com/virtengine/virtengine/x/roles/keeper"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	stakingtypes "github.com/virtengine/virtengine/x/staking/types"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// VEIDDecorator enforces VEID tier/score requirements per message.
type VEIDDecorator struct {
	veidKeeper   veidkeeper.Keeper
	mfaKeeper    mfakeeper.Keeper
	govKeeper    *govkeeper.Keeper
	govAuthority string
	rolesKeeper  roleskeeper.Keeper
}

// NewVEIDDecorator creates a new VEID gating decorator.
func NewVEIDDecorator(
	veidKeeper veidkeeper.Keeper,
	mfaKeeper mfakeeper.Keeper,
	govKeeper *govkeeper.Keeper,
	rolesKeeper roleskeeper.Keeper,
) VEIDDecorator {
	authority := ""
	if govKeeper != nil {
		authority = govKeeper.GetAuthority()
	}

	return VEIDDecorator{
		veidKeeper:   veidKeeper,
		mfaKeeper:    mfaKeeper,
		govKeeper:    govKeeper,
		govAuthority: authority,
		rolesKeeper:  rolesKeeper,
	}
}

// AnteHandle implements sdk.AnteDecorator.
func (d VEIDDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.BlockHeight() <= 0 {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return next(ctx, tx, simulate)
	}

	for _, msg := range msgs {
		if isGovernanceMsgTypeURL(sdk.MsgTypeURL(msg)) {
			continue
		}

		requirement := d.getMessageRequirement(ctx, msg)
		signers := msg.GetSigners()
		for _, signer := range signers {
			if d.isGovernanceAuthority(signer) {
				continue
			}

			if err := d.checkVEIDRequirements(ctx, signer, requirement, sdk.MsgTypeURL(msg)); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

type veidMessageRequirement struct {
	MinTier                   int
	MinScore                  uint32
	RequireAuthorization      bool
	RequireGovernanceApproval bool
	SensitiveTxType           mfatypes.SensitiveTransactionType
	Description               string
}

func defaultMessageRequirement() veidMessageRequirement {
	return veidMessageRequirement{
		MinTier:  veidtypes.TierUnverified,
		MinScore: 0,
	}
}

func (d VEIDDecorator) getMessageRequirement(ctx sdk.Context, msg sdk.Msg) veidMessageRequirement {
	requirement := defaultMessageRequirement()
	typeURL := sdk.MsgTypeURL(msg)
	if typeURL == "" {
		return requirement
	}

	txType, ok := mfatypes.GetSensitiveTransactionType(typeURL)
	if !ok {
		return requirement
	}

	config, found := d.mfaKeeper.GetSensitiveTxConfig(ctx, txType)
	if !found {
		config = defaultSensitiveTxConfig(txType)
		if config == nil {
			return requirement
		}
	}

	if !config.Enabled {
		return requirement
	}

	requirement.MinScore = config.MinVEIDScore
	requirement.MinTier = minTierForScore(config.MinVEIDScore)
	requirement.RequireAuthorization = true
	requirement.SensitiveTxType = txType
	requirement.Description = config.Description
	if txType == mfatypes.SensitiveTxValidatorRegistration {
		requirement.RequireGovernanceApproval = true
	}
	return requirement
}

func (d VEIDDecorator) checkVEIDRequirements(
	ctx sdk.Context,
	signer sdk.AccAddress,
	requirement veidMessageRequirement,
	msgTypeURL string,
) error {
	if requirement.MinTier <= veidtypes.TierUnverified && requirement.MinScore == 0 && !requirement.RequireGovernanceApproval {
		return nil
	}

	needsTierCheck := requirement.MinTier > veidtypes.TierUnverified
	needsScoreCheck := requirement.MinScore > 0
	needsAuditEvent := requirement.SensitiveTxType == mfatypes.SensitiveTxValidatorRegistration

	tier := veidtypes.TierUnverified
	if needsTierCheck {
		var tierErr error
		tier, tierErr = d.veidKeeper.GetAccountTier(ctx, signer.String())
		if tierErr != nil {
			tier = veidtypes.TierUnverified
		}
	}

	if needsTierCheck && tier < requirement.MinTier {
		return sdkerrors.ErrUnauthorized.Wrapf(
			"VEID tier %s below required %s for %s",
			veidtypes.TierToString(tier),
			veidtypes.TierToString(requirement.MinTier),
			requirement.actionLabel(msgTypeURL),
		)
	}

	score := uint32(0)
	status := veidtypes.AccountStatusUnknown
	if needsScoreCheck || needsAuditEvent {
		var found bool
		score, status, found = d.veidKeeper.GetScore(ctx, signer.String())
		if !found {
			score = 0
			status = veidtypes.AccountStatusUnknown
		}
	}

	if needsScoreCheck {
		if !d.veidKeeper.IsScoreAboveThreshold(ctx, signer.String(), requirement.MinScore) {
			return sdkerrors.ErrUnauthorized.Wrapf(
				"VEID score %d (status %s) below required %d for %s",
				score,
				status,
				requirement.MinScore,
				requirement.actionLabel(msgTypeURL),
			)
		}
	}

	governanceApproved := true
	if requirement.RequireGovernanceApproval {
		governanceApproved = d.rolesKeeper.HasRole(ctx, signer, rolestypes.RoleValidator)
		if !governanceApproved {
			return sdkerrors.ErrUnauthorized.Wrap("governance approval required for validator registration")
		}
	}

	if needsAuditEvent {
		return d.emitValidatorIdentityEvent(ctx, signer, score, status, requirement.MinScore, governanceApproved)
	}

	return nil
}

func (d VEIDDecorator) emitValidatorIdentityEvent(
	ctx sdk.Context,
	signer sdk.AccAddress,
	score uint32,
	status veidtypes.AccountStatus,
	requiredScore uint32,
	governanceApproved bool,
) error {
	return ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			stakingtypes.EventTypeValidatorIdentityVerified,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidatorAddress, signer.String()),
			sdk.NewAttribute(stakingtypes.AttributeKeyVEIDScore, fmt.Sprintf("%d", score)),
			sdk.NewAttribute(veidtypes.AttributeKeyStatus, status.String()),
			sdk.NewAttribute(stakingtypes.AttributeKeyRequiredScore, fmt.Sprintf("%d", requiredScore)),
			sdk.NewAttribute(stakingtypes.AttributeKeyGovernanceApproved, fmt.Sprintf("%t", governanceApproved)),
		),
	)
}

func (d VEIDDecorator) isGovernanceAuthority(signer sdk.AccAddress) bool {
	if d.govAuthority == "" {
		return false
	}
	return signer.String() == d.govAuthority
}

func (r veidMessageRequirement) actionLabel(typeURL string) string {
	if r.Description != "" {
		return r.Description
	}
	if r.SensitiveTxType != mfatypes.SensitiveTxUnspecified {
		return r.SensitiveTxType.String()
	}
	if typeURL != "" {
		return typeURL
	}
	return "message"
}

func minTierForScore(score uint32) int {
	switch {
	case score >= veidtypes.ThresholdPremium:
		return veidtypes.TierPremium
	case score >= veidtypes.ThresholdStandard:
		return veidtypes.TierStandard
	case score >= veidtypes.ThresholdBasic:
		return veidtypes.TierBasic
	default:
		return veidtypes.TierUnverified
	}
}

func defaultSensitiveTxConfig(txType mfatypes.SensitiveTransactionType) *mfatypes.SensitiveTxConfig {
	for _, cfg := range mfatypes.GetDefaultSensitiveTxConfigs() {
		if cfg.TransactionType == txType {
			candidate := cfg
			return &candidate
		}
	}
	return nil
}

func isGovernanceMsgTypeURL(typeURL string) bool {
	if typeURL == "" {
		return false
	}

	for _, prefix := range governanceTypeURLPrefixes {
		if strings.HasPrefix(typeURL, prefix) {
			return true
		}
	}

	return false
}

var governanceTypeURLPrefixes = []string{
	"/cosmos.gov.",
	"/virtengine.gov.",
}
