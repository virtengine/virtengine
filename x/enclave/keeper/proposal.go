package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// NewEnclaveProposalHandler returns the governance proposal handler for the enclave module.
func NewEnclaveProposalHandler(k Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch proposal := content.(type) {
		case *types.AddMeasurementProposal:
			return handleAddMeasurementProposal(ctx, k, proposal)
		case *types.RevokeMeasurementProposal:
			return handleRevokeMeasurementProposal(ctx, k, proposal)
		default:
			return fmt.Errorf("unrecognized enclave proposal content type: %T", content)
		}
	}
}

func handleAddMeasurementProposal(ctx sdk.Context, k Keeper, proposal *types.AddMeasurementProposal) error {
	if proposal == nil {
		return types.ErrInvalidMeasurement.Wrap("proposal cannot be nil")
	}

	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	if existing, found := k.GetMeasurement(ctx, proposal.MeasurementHash); found && !existing.Revoked {
		return types.ErrInvalidMeasurement.Wrap("measurement already allowlisted")
	}

	var expiryHeight int64
	if proposal.ExpiryBlocks > 0 {
		expiryHeight = ctx.BlockHeight() + proposal.ExpiryBlocks
	}

	measurement := &types.MeasurementRecord{
		MeasurementHash: proposal.MeasurementHash,
		TEEType:         proposal.TEEType,
		Description:     proposal.Description,
		MinISVSVN:       proposal.MinISVSVN,
		ExpiryHeight:    expiryHeight,
	}

	if err := k.AddMeasurement(ctx, measurement); err != nil {
		return err
	}

	k.RecordMeasurementProposal("add")
	return nil
}

func handleRevokeMeasurementProposal(ctx sdk.Context, k Keeper, proposal *types.RevokeMeasurementProposal) error {
	if proposal == nil {
		return types.ErrInvalidMeasurement.Wrap("proposal cannot be nil")
	}

	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	measurement, found := k.GetMeasurement(ctx, proposal.MeasurementHash)
	if !found {
		return types.ErrMeasurementNotAllowlisted
	}
	if measurement.Revoked {
		return types.ErrMeasurementRevoked
	}

	if err := k.RevokeMeasurement(ctx, proposal.MeasurementHash, proposal.Reason, 0); err != nil {
		return err
	}

	k.RecordMeasurementProposal("revoke")
	return nil
}
