package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

const webScopeScoringModelVersion = "web-scope-v1"

type webScopeContribution struct {
	FeatureName  string
	ScoreBasisPt uint32
	EvidenceHash string
}

func (k Keeper) applyWebScopeScore(
	ctx sdk.Context,
	accountAddr string,
	contributions []webScopeContribution,
) error {
	if len(contributions) == 0 {
		return types.ErrInvalidScore.Wrap("no web-scope contributions provided")
	}

	// Deterministic ordering
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].FeatureName < contributions[j].FeatureName
	})

	var totalBasisPoints uint32
	for _, c := range contributions {
		totalBasisPoints += c.ScoreBasisPt
	}

	delta := totalBasisPoints / 100 // convert basis points to 0-100 scale

	currentScore, status, found := k.GetScore(ctx, accountAddr)
	newScore := currentScore + delta
	if newScore > types.MaxScore {
		newScore = types.MaxScore
	}

	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	modelVersion := webScopeScoringModelVersion
	if record, ok := k.GetIdentityRecord(ctx, address); ok {
		if record.ScoreVersion != "" {
			modelVersion = record.ScoreVersion
		}
	}

	// Build evidence summary for scoring history
	summary := types.NewEvidenceSummary(modelVersion, ctx.BlockHeight(), ctx.BlockTime())
	maxWeight := uint32(types.MaxBasisPoints)
	for _, c := range contributions {
		summary.AddContribution(types.FeatureContribution{
			FeatureName:     c.FeatureName,
			RawScore:        c.ScoreBasisPt,
			Weight:          maxWeight,
			WeightedScore:   c.ScoreBasisPt,
			PassedThreshold: c.ScoreBasisPt > 0,
		})
	}

	inputHash := computeWebScopeInputHash(accountAddr, contributions)
	passed := newScore >= types.ThresholdBasic
	summary.SetResult(newScore, passed, inputHash)

	if err := k.RecordScoringResult(ctx, accountAddr, summary); err != nil {
		k.Logger(ctx).Error("failed to record web-scope scoring result", "error", err)
	}

	// Update score store
	if found {
		if err := k.SetScoreWithDetails(ctx, accountAddr, newScore, ScoreDetails{
			Status:           status,
			ModelVersion:     modelVersion,
			VerificationHash: inputHash,
			Reason:           "web-scope evidence update",
		}); err != nil {
			return err
		}
	} else {
		newStatus := types.AccountStatusPending
		if newScore >= types.ThresholdBasic {
			newStatus = types.AccountStatusVerified
		}
		status = newStatus
		if err := k.SetScoreWithDetails(ctx, accountAddr, newScore, ScoreDetails{
			Status:           newStatus,
			ModelVersion:     modelVersion,
			VerificationHash: inputHash,
			Reason:           "web-scope evidence update",
		}); err != nil {
			return err
		}
	}

	// Update wallet history if available
	if _, ok := k.GetWallet(ctx, address); ok {
		scopeIDs := make([]string, 0, len(contributions))
		for _, c := range contributions {
			scopeIDs = append(scopeIDs, c.FeatureName)
		}
		if err := k.UpdateWalletScore(ctx, address, newScore, status, modelVersion, "", scopeIDs, "web-scope evidence update"); err != nil {
			k.Logger(ctx).Error("failed to update wallet score for web-scope evidence", "error", err)
		}
	}

	return nil
}

func computeWebScopeInputHash(accountAddr string, contributions []webScopeContribution) []byte {
	h := sha256.New()
	h.Write([]byte(accountAddr))

	for _, c := range contributions {
		h.Write([]byte(c.FeatureName))
		scoreBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(scoreBytes, c.ScoreBasisPt)
		h.Write(scoreBytes)
		if c.EvidenceHash != "" {
			h.Write([]byte(c.EvidenceHash))
		}
	}

	return h.Sum(nil)
}
