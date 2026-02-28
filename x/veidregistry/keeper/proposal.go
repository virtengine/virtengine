package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/virtengine/virtengine/x/veidregistry/types"
)

func NewProposalHandler(k Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch proposal := content.(type) {
		case *types.AddVerifierVersionProposal:
			return handleAddVerifierVersionProposal(ctx, k, proposal)
		case *types.ActivateVerifierProposal:
			return handleActivateVerifierProposal(ctx, k, proposal)
		case *types.UpdateParamsProposal:
			return handleUpdateParamsProposal(ctx, k, proposal)
		default:
			return fmt.Errorf("unrecognized veidregistry proposal content type: %T", content)
		}
	}
}

func handleAddVerifierVersionProposal(ctx sdk.Context, k Keeper, proposal *types.AddVerifierVersionProposal) error {
	if proposal == nil {
		return fmt.Errorf("proposal cannot be nil")
	}
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	ms := NewMsgServerImpl(k)
	proposalID := proposal.Verifier.GovernanceProposalID
	if proposalID == 0 {
		proposalID = uint64(ctx.BlockHeight())
	}
	_, err := ms.UpsertVerifierVersion(sdk.WrapSDKContext(ctx), &types.MsgUpsertVerifierVersion{
		Authority: k.GetAuthority(),
		Verifier: types.VerifierVersion{
			VerifierID:          proposal.Verifier.VerifierID,
			SpecVersion:         proposal.Verifier.SpecVersion,
			SpecCID:             proposal.Verifier.SpecCID,
			SpecSHA256:          proposal.Verifier.SpecSHA256,
			WeightsCID:          proposal.Verifier.WeightsCID,
			WeightsSHA256:       proposal.Verifier.WeightsSHA256,
			TestVectorsCID:      proposal.Verifier.TestVectorsCID,
			TestVectorsSHA256:   proposal.Verifier.TestVectorsSHA256,
			BuildMetadataCID:    proposal.Verifier.BuildMetadataCID,
			BuildMetadataSHA256: proposal.Verifier.BuildMetadataSHA256,
			ImageHash:           proposal.Verifier.ImageHash,
			ModelManifestHash:   proposal.Verifier.ModelManifestHash,
			Status:              string(types.VerifierStatusProposed),
		},
	})
	if err != nil {
		return err
	}
	_, err = ms.ApproveVerifierVersion(sdk.WrapSDKContext(ctx), &types.MsgApproveVerifierVersion{
		Authority:            k.GetAuthority(),
		VerifierID:           proposal.Verifier.VerifierID,
		GovernanceProposalID: proposalID,
		ActivationHeight:     proposal.Verifier.ActivationHeight,
		SecurityFix:          proposal.Verifier.SecurityFix,
	})
	return err
}

func handleActivateVerifierProposal(ctx sdk.Context, k Keeper, proposal *types.ActivateVerifierProposal) error {
	if proposal == nil {
		return fmt.Errorf("proposal cannot be nil")
	}
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	return k.SetActiveVerifier(ctx, proposal.Active)
}

func handleUpdateParamsProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateParamsProposal) error {
	if proposal == nil {
		return fmt.Errorf("proposal cannot be nil")
	}
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	_, err := NewMsgServerImpl(k).UpdateParams(sdk.WrapSDKContext(ctx), &types.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    proposal.Params,
	})
	return err
}
