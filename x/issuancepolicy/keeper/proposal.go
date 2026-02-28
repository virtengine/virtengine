package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

func NewProposalHandler(k Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch proposal := content.(type) {
		case *types.UpsertPolicyProposal:
			return handleUpsertPolicyProposal(ctx, k, proposal)
		case *types.SetActivePolicyProposal:
			return handleSetActivePolicyProposal(ctx, k, proposal)
		case *types.UpdateParamsProposal:
			return handleUpdateParamsProposal(ctx, k, proposal)
		default:
			return fmt.Errorf("unrecognized issuancepolicy proposal content type: %T", content)
		}
	}
}

func handleUpsertPolicyProposal(ctx sdk.Context, k Keeper, proposal *types.UpsertPolicyProposal) error {
	if proposal == nil {
		return fmt.Errorf("proposal cannot be nil")
	}
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	_, err := NewMsgServerImpl(k).UpsertPolicy(sdk.WrapSDKContext(ctx), &types.MsgUpsertPolicy{
		Authority: k.GetAuthority(),
		Policy:    proposal.Policy,
	})
	return err
}

func handleSetActivePolicyProposal(ctx sdk.Context, k Keeper, proposal *types.SetActivePolicyProposal) error {
	if proposal == nil {
		return fmt.Errorf("proposal cannot be nil")
	}
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	_, err := NewMsgServerImpl(k).SetActivePolicy(sdk.WrapSDKContext(ctx), &types.MsgSetActivePolicy{
		Authority: k.GetAuthority(),
		PolicyID:  proposal.PolicyID,
	})
	return err
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
