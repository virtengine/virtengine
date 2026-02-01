package cli

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	dv1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

var (
	DefaultPowerReduction   = sdkmath.NewIntFromUint64(sdk.DefaultPowerReduction.Uint64())
	DefaultMinDepositTokens = sdkmath.NewIntFromUint64(govv1.DefaultMinDepositTokens.Uint64())
)

type FlagsSet []string

type Stringer interface {
	String() string
}

func TestFlags() FlagsSet {
	return FlagsSet{}
}

func (df FlagsSet) With(flags ...string) FlagsSet {
	res := make([]string, len(df), len(df)+len(flags))

	copy(res, df)
	res = append(res, flags...)

	return res
}

func (df FlagsSet) WithFlag(key string, val any) FlagsSet {
	var str string
	switch v := val.(type) {
	case string:
		str = v
	case int, uint, int64, uint64:
		str = fmt.Sprintf("%d", v)
	case bool:
		str = strconv.FormatBool(v)
	case Stringer:
		str = v.String()
	default:
		panic(fmt.Sprintf("val %s is not a type of real|bool|string|Stringer", reflect.TypeOf(val).String()))
	}
	res := make([]string, len(df), len(df)+1)

	copy(res, df)
	res = append(res, fmt.Sprintf("--%s=%s", key, str))

	return res
}

func (df FlagsSet) WithLimit(val int64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagLimit, val))

	return res
}

func (df FlagsSet) Append(rhs FlagsSet) FlagsSet {
	res := make([]string, len(df), len(df)+len(rhs))

	copy(res, df)
	res = append(res, rhs...)

	return res
}

func (df FlagsSet) WithAllowedMsgs(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagAllowedMsgs, val))

	return res
}

func (df FlagsSet) WithGas(val int) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagGas, val))

	return res
}

func (df FlagsSet) WithGasAdjustment(val float64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%f", cflags.FlagGasAdjustment, val))

	return res
}

func (df FlagsSet) WithGasPrices(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagGasPrices, val))

	return res
}

func (df FlagsSet) WithGasAuto() FlagsSet {
	res := make([]string, len(df), len(df)+3)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagGas, cflags.GasFlagAuto))
	res = append(res, fmt.Sprintf("--%s=0.0025uve", cflags.FlagGasPrices))
	res = append(res, fmt.Sprintf("--%s=1.5", cflags.FlagGasAdjustment))

	return res
}

// WithGasAutoFlags is an alias for WithGasAuto for backward compatibility.
func (df FlagsSet) WithGasAutoFlags() FlagsSet {
	return df.WithGasAuto()
}

func (df FlagsSet) WithGenerateOnly() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagGenerateOnly))

	return res
}

func (df FlagsSet) WithOverwrite() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagOverwrite))

	return res
}

func (df FlagsSet) WithOffline() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagOffline))

	return res
}

func (df FlagsSet) WithAccountNumber(val uint64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagAccountNumber, val))

	return res
}

func (df FlagsSet) WithSequence(val uint64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagSequence, val))

	return res
}

func (df FlagsSet) WithSkipConfirm() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagSkipConfirmation))

	return res
}

func (df FlagsSet) WithSignatureOnly() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagSigOnly))

	return res
}

func (df FlagsSet) WithNote(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagNote, val))

	return res
}

func (df FlagsSet) WithSpendLimit(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagSpendLimit, val))

	return res
}

func (df FlagsSet) WithScope(val []string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagScope, strings.Join(val, ",")))

	return res
}

func (df FlagsSet) WithPeriodLimit(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagPeriodLimit, val))

	return res
}

func (df FlagsSet) WithPeriod(val int64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagPeriod, val))

	return res
}

func (df FlagsSet) WithEvents(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagEvents, val))

	return res
}

func (df FlagsSet) WithDenom(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagDenom, val))

	return res
}

func (df FlagsSet) WithMsgType(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagMsgType, val))

	return res
}

func (df FlagsSet) WithBroadcastModeSync() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagBroadcastMode, cflags.BroadcastSync))

	return res
}

func (df FlagsSet) WithExpiration(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagExpiration, val))

	return res
}

func (df FlagsSet) WithCommission() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagCommission))

	return res
}

func (df FlagsSet) WithAllowList(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagAllowList, val))

	return res
}

func (df FlagsSet) WithAllowedValidators(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagAllowedValidators, val))

	return res
}

func (df FlagsSet) WithDenyValidators(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagDenyValidators, val))

	return res
}

func (df FlagsSet) WithSignMode(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagSignMode, val))

	return res
}

func (df FlagsSet) WithTip(val sdk.Coin) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagTip, val.String()))

	return res
}

func (df FlagsSet) WithAux() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=true", cflags.FlagAux))

	return res
}

func (df FlagsSet) WithMultisig(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagMultisig, val))

	return res
}

func (df FlagsSet) WithMetadata(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagMetadata, val))

	return res
}

func (df FlagsSet) WithProposal(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagProposal, val)) // nolint:staticcheck

	return res
}

func (df FlagsSet) WithTitle(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagTitle, val))

	return res
}

func (df FlagsSet) WithType(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagType, val))

	return res
}

func (df FlagsSet) WithProposalType(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagProposalType, val)) // nolint:staticcheck

	return res
}

func (df FlagsSet) WithDescription(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagDescription, val)) // nolint:staticcheck

	return res
}

func (df FlagsSet) WithBroadcastModeBlock() FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagBroadcastMode, cflags.BroadcastBlock))

	return res
}

func (df FlagsSet) WithHome(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagHome, val))

	return res
}

func (df FlagsSet) WithChainID(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagChainID, val))

	return res
}

func (df FlagsSet) WithFees(coins sdk.Coins) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagFees, coins.String()))

	return res
}

func (df FlagsSet) WithDeposit(coin sdk.Coin) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagDeposit, coin))

	return res
}

func (df FlagsSet) WithPrice(coin sdk.DecCoin) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagPrice, coin))

	return res
}

func (df FlagsSet) WithFrom(acc string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagFrom, acc))

	return res
}

func (df FlagsSet) WithFeeGranter(val sdk.AccAddress) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagFeeGranter, val.String()))

	return res
}

func (df FlagsSet) WithFeePayer(val sdk.AccAddress) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagFeePayer, val.String()))

	return res
}

func (df FlagsSet) WithDepositSources(sources []string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagDepositSources, strings.Join(sources, ",")))

	return res
}

func (df FlagsSet) WithOutput(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOutput, val))

	return res
}

func (df FlagsSet) WithSerial(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagSerial, val))

	return res
}

func (df FlagsSet) WithOwner(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOwner, val))

	return res
}

func (df FlagsSet) WithProvider(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagProvider, val))

	return res
}

func (df FlagsSet) WithDSeq(val uint64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagDSeq, val))

	return res
}

func (df FlagsSet) WithGSeq(val uint32) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagGSeq, val))

	return res
}

func (df FlagsSet) WithOSeq(val uint32) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagOSeq, val))

	return res
}

func (df FlagsSet) WithDeploymentID(val dv1.DeploymentID) FlagsSet {
	res := make([]string, len(df), len(df)+2)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOwner, val.Owner))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagDSeq, val.DSeq))

	return res
}

func (df FlagsSet) WithGroupID(val dv1.GroupID) FlagsSet {
	res := make([]string, len(df), len(df)+3)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOwner, val.Owner))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagDSeq, val.DSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagGSeq, val.GSeq))

	return res
}

func (df FlagsSet) WithOrderID(val mtypes.OrderID) FlagsSet {
	res := make([]string, len(df), len(df)+4)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOwner, val.Owner))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagDSeq, val.DSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagGSeq, val.GSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagOSeq, val.OSeq))

	return res
}

func (df FlagsSet) WithBidID(val mtypes.BidID) FlagsSet {
	res := make([]string, len(df), len(df)+5)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOwner, val.Owner))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagDSeq, val.DSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagGSeq, val.GSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagOSeq, val.OSeq))
	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagProvider, val.Provider))

	return res
}

func (df FlagsSet) WithLeaseID(val mtypes.LeaseID) FlagsSet {
	res := make([]string, len(df), len(df)+5)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagOwner, val.Owner))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagDSeq, val.DSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagGSeq, val.GSeq))
	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagOSeq, val.OSeq))
	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagProvider, val.Provider))

	return res
}

func (df FlagsSet) WithState(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagState, val))

	return res
}

func (df FlagsSet) WithStatus(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagStatus, val))

	return res
}

func (df FlagsSet) WithHeight(val uint64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagHeight, val))

	return res
}

func (df FlagsSet) WithOutputJSON() FlagsSet {
	return df.WithOutput("json")
}

func (df FlagsSet) WithOutputYAML() FlagsSet {
	return df.WithOutput("yaml")
}

func (df FlagsSet) WithOutputText() FlagsSet {
	return df.WithOutput("text")
}

func (df FlagsSet) WithIdentity(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagIdentity, val))

	return res
}

func (df FlagsSet) WithWebsite(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagWebsite, val))

	return res
}

func (df FlagsSet) WithSecurityContact(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagSecurityContact, val))

	return res
}

func (df FlagsSet) WithDetails(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagDetails, val))

	return res
}

func (df FlagsSet) WithCommissionRate(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagCommissionRate, val))

	return res
}

func (df FlagsSet) WithCommissionMaxRate(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagCommissionMaxRate, val))

	return res
}

func (df FlagsSet) WithCommissionMaxChangeRate(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagCommissionMaxChangeRate, val))

	return res
}

func (df FlagsSet) WithAmount(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagAmount, val))

	return res
}

func (df FlagsSet) WithPubkey(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagPubKey, val))

	return res
}

func (df FlagsSet) WithMoniker(val string) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%s", cflags.FlagMoniker, val))

	return res
}

func (df FlagsSet) WithTimeoutHeight(val uint64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagTimeoutHeight, val))

	return res
}

func (df FlagsSet) WithTimeoutDuration(val interface{}) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%v", cflags.TimeoutDuration, val))

	return res
}

func (df FlagsSet) WithUnordered(val bool) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%t", cflags.FlagUnordered, val))

	return res
}

func (df FlagsSet) WithOffset(val uint64) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%d", cflags.FlagOffset, val))

	return res
}

func (df FlagsSet) WithCountTotal(val bool) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%t", cflags.FlagCountTotal, val))

	return res
}

func (df FlagsSet) WithReverse(val bool) FlagsSet {
	res := make([]string, len(df), len(df)+1)

	copy(res, df)

	res = append(res, fmt.Sprintf("--%s=%t", cflags.FlagReverse, val))

	return res
}

// // ExecTestCLICmd builds the client context, mocks the output and executes the command.
// func ExecTestCLICmd(ctx context.Context, cctx client.Context, cmd *cobra.Command, extraArgs ...string) (sdktest.BufferWriter, error) {
// 	_, out := sdktest.ApplyMockIO(cmd)
//
// 	{
// 		dupFlags := make(map[string]bool)
// 		for _, arg := range extraArgs {
// 			if !strings.HasPrefix(arg, "--") {
// 				continue
// 			}
//
// 			arg = strings.TrimPrefix(arg, "--")
// 			tokens := strings.Split(arg, "=")
//
// 			if _, exists := dupFlags[tokens[0]]; exists {
// 				return out, fmt.Errorf("test: duplicated flag \"%s\"", tokens[0])
// 			}
//
// 			dupFlags[tokens[0]] = true
// 		}
// 	}
//
// 	cmd.SetArgs(extraArgs)
// 	err := cmd.ParseFlags(extraArgs)
//
// 	if err != nil {
// 		return out, err
// 	}
//
// 	cctx = cctx.WithOutput(out)
//
// 	if ctx == nil {
// 		ctx = context.Background()
// 	}
//
// 	opts, err := cflags.ClientOptionsFromFlags(cmd.Flags())
// 	if err != nil {
// 		return out, err
// 	}
//
// 	ctx = context.WithValue(ctx, cli.ClientContextKey, &client.Context{})
// 	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())
// 	cmd.SetContext(ctx)
//
// 	if err = client.SetCmdClientContextHandler(cctx, cmd); err != nil {
// 		return out, err
// 	}
//
// 	tctx, err := client.GetClientTxContext(cmd)
// 	if err != nil {
// 		return out, err
// 	}
//
// 	cl, err := DiscoverClient(ctx, tctx, opts...)
// 	if err != nil {
// 		return out, err
// 	}
//
// 	ctx = context.WithValue(ctx, ContextTypeClient, cl)
//
// 	cmd.SetContext(ctx)
//
// 	if err := cmd.Execute(); err != nil {
// 		return out, err
// 	}
//
// 	return out, nil
// }
//
// // ExecQueryTestCLICmd builds the client context, mocks the output and executes the command.
// func ExecQueryTestCLICmd(ctx context.Context, cctx client.Context, cmd *cobra.Command, extraArgs ...string) (sdktest.BufferWriter, error) {
// 	_, out := sdktest.ApplyMockIO(cmd)
//
// 	{
// 		dupFlags := make(map[string]bool)
// 		for _, arg := range extraArgs {
// 			if !strings.HasPrefix(arg, "--") {
// 				continue
// 			}
//
// 			arg = strings.TrimPrefix(arg, "--")
// 			tokens := strings.Split(arg, "=")
//
// 			if _, exists := dupFlags[tokens[0]]; exists {
// 				return out, fmt.Errorf("test: duplicated flag \"%s\"", tokens[0])
// 			}
//
// 			dupFlags[tokens[0]] = true
// 		}
// 	}
//
// 	cmd.SetArgs(extraArgs)
// 	err := cmd.ParseFlags(extraArgs)
//
// 	if err != nil {
// 		return out, err
// 	}
//
// 	cctx = cctx.WithOutput(out)
//
// 	if ctx == nil {
// 		ctx = context.Background()
// 	}
//
// 	ctx = context.WithValue(ctx, cli.ClientContextKey, &client.Context{})
// 	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())
// 	cmd.SetContext(ctx)
//
// 	if err = client.SetCmdClientContextHandler(cctx, cmd); err != nil {
// 		return out, err
// 	}
//
// 	qctx, err := GetClientQueryContext(cmd)
// 	if err != nil {
// 		return out, err
// 	}
//
// 	qcl, err := DiscoverQueryClient(ctx, qctx)
// 	if err != nil {
// 		return out, err
// 	}
//
// 	ctx = context.WithValue(ctx, ContextTypeQueryClient, qcl)
// 	cmd.SetContext(ctx)
//
// 	if err := cmd.Execute(); err != nil {
// 		return out, err
// 	}
//
// 	return out, nil
// }
//
// func MsgSendExec(ctx context.Context, cctx client.Context, from, to, amount fmt.Stringer, extraArgs ...string) (sdktest.BufferWriter, error) {
// 	args := []string{from.String(), to.String(), amount.String()}
// 	args = append(args, extraArgs...)
//
// 	return ExecTestCLICmd(ctx, cctx, GetTxBankSendTxCmd(), args...)
// }
//
// func QueryBalancesExec(ctx context.Context, cctx client.Context, address fmt.Stringer, extraArgs ...string) (sdktest.BufferWriter, error) {
// 	args := []string{address.String(), fmt.Sprintf("--%s=json", cflags.FlagOutput)}
// 	args = append(args, extraArgs...)
//
// 	return ExecQueryTestCLICmd(ctx, cctx, cli.GetBalancesCmd(), args...)
// }

