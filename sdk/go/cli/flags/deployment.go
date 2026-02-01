package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dv1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	dv1beta "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
)

type DeploymentIDOptions struct {
	NoOwner bool
}

type DeploymentIDOption func(*DeploymentIDOptions)

// DeploymentIDOptionNoOwner do not add mark as required owner flag
func DeploymentIDOptionNoOwner(val bool) DeploymentIDOption {
	return func(opt *DeploymentIDOptions) {
		opt.NoOwner = val
	}
}

type MarketOptions struct {
	Owner    sdk.AccAddress
	Provider sdk.AccAddress
}

type MarketOption func(*MarketOptions)

func WithOwner(val sdk.AccAddress) MarketOption {
	return func(opt *MarketOptions) {
		opt.Owner = val
	}
}

func WithProvider(val sdk.AccAddress) MarketOption {
	return func(opt *MarketOptions) {
		opt.Provider = val
	}
}

// AddDeploymentIDFlags add flags for deployment except for Owner when NoOwner is set
func AddDeploymentIDFlags(flags *pflag.FlagSet, opts ...DeploymentIDOption) {
	opt := &DeploymentIDOptions{}

	for _, o := range opts {
		o(opt)
	}

	if !opt.NoOwner {
		flags.String(FlagOwner, "", "Deployment Owner")
	}

	flags.Uint64(FlagDSeq, 0, "Deployment Sequence")
}

// MarkReqDeploymentIDFlags marks flags required except for Owner when NoOwner is set
func MarkReqDeploymentIDFlags(cmd *cobra.Command, opts ...DeploymentIDOption) {
	opt := &DeploymentIDOptions{}

	for _, o := range opts {
		o(opt)
	}

	if !opt.NoOwner {
		_ = cmd.MarkFlagRequired(FlagOwner)
	}

	_ = cmd.MarkFlagRequired(FlagDSeq)
}

// DeploymentIDFromFlags returns DeploymentID with given flags, owner and error if occurred
func DeploymentIDFromFlags(flags *pflag.FlagSet, opts ...MarketOption) (dv1.DeploymentID, error) {
	var id dv1.DeploymentID
	opt := &MarketOptions{}

	for _, o := range opts {
		o(opt)
	}

	var owner string
	if flag := flags.Lookup(FlagOwner); flag != nil {
		owner = flag.Value.String()
	}

	// if --owner flag was explicitly provided, use that.
	var err error
	if owner != "" {
		opt.Owner, err = sdk.AccAddressFromBech32(owner)
		if err != nil {
			return id, err
		}
	}

	id.Owner = opt.Owner.String()

	if id.DSeq, err = flags.GetUint64(FlagDSeq); err != nil {
		return id, err
	}

	return id, nil
}

// DeploymentIDFromFlagsForOwner returns DeploymentID with given flags, owner and error if occurred
func DeploymentIDFromFlagsForOwner(flags *pflag.FlagSet, owner sdk.Address) (dv1.DeploymentID, error) {
	id := dv1.DeploymentID{
		Owner: owner.String(),
	}

	var err error
	if id.DSeq, err = flags.GetUint64(FlagDSeq); err != nil {
		return id, err
	}

	return id, nil
}

// AddGroupIDFlags add flags for Group
func AddGroupIDFlags(flags *pflag.FlagSet, opts ...DeploymentIDOption) {
	AddDeploymentIDFlags(flags, opts...)
	flags.Uint32(FlagGSeq, 1, "Group Sequence")
}

// MarkReqGroupIDFlags marks flags required for group
func MarkReqGroupIDFlags(cmd *cobra.Command, opts ...DeploymentIDOption) {
	MarkReqDeploymentIDFlags(cmd, opts...)
}

// GroupIDFromFlags returns GroupID with given flags and error if occurred
func GroupIDFromFlags(flags *pflag.FlagSet, opts ...MarketOption) (dv1.GroupID, error) {
	var id dv1.GroupID
	prev, err := DeploymentIDFromFlags(flags, opts...)
	if err != nil {
		return id, err
	}

	gseq, err := flags.GetUint32(FlagGSeq)
	if err != nil {
		return id, err
	}
	return dv1.MakeGroupID(prev, gseq), nil
}

// AddDeploymentFilterFlags add flags to filter for deployment list
func AddDeploymentFilterFlags(flags *pflag.FlagSet) {
	flags.String(FlagOwner, "", "deployment owner address to filter")
	flags.String(FlagState, "", "deployment state to filter (active,closed)")
	flags.Uint64(FlagDSeq, 0, "deployment sequence to filter")
}

// DepFiltersFromFlags returns DeploymentFilters with given flags and error if occurred
func DepFiltersFromFlags(flags *pflag.FlagSet) (dv1beta.DeploymentFilters, error) {
	var dfilters dv1beta.DeploymentFilters
	owner, err := flags.GetString(FlagOwner)
	if err != nil {
		return dfilters, err
	}

	if owner != "" {
		_, err = sdk.AccAddressFromBech32(owner)
		if err != nil {
			return dfilters, err
		}
	}

	dfilters.Owner = owner

	if dfilters.State, err = flags.GetString(FlagState); err != nil {
		return dfilters, err
	}

	if dfilters.DSeq, err = flags.GetUint64(FlagDSeq); err != nil {
		return dfilters, err
	}

	return dfilters, nil
}

func AddDepositSourcesFlags(flags *pflag.FlagSet) {
	flags.StringSlice(FlagDepositSources, []string{"grant", "balance"}, "Comma separated list of deposit sources. allowed values grant|balance")
}

func AddDepositFlags(flags *pflag.FlagSet) {
	flags.String(FlagDeposit, "", "Deposit amount")
	flags.StringSlice(FlagDepositSources, []string{"grant", "balance"}, "Comma separated list of deposit sources. allowed values grant|balance")
}

