package flags

import (
	"fmt"
	"strconv"

	"github.com/spf13/pflag"

	cltypes "github.com/virtengine/virtengine/sdk/go/node/client/types"
)

// ClientOptionsFromFlags reads client options from cli flag set.
func ClientOptionsFromFlags(flagSet *pflag.FlagSet) ([]cltypes.ClientOption, error) {
	opts := make([]cltypes.ClientOption, 0)

	if flagSet.Changed(FlagAccountNumber) {
		accNum, _ := flagSet.GetUint64(FlagAccountNumber)
		opts = append(opts, cltypes.WithAccountNumber(accNum))
	}

	if flagSet.Changed(FlagSequence) {
		accSeq, _ := flagSet.GetUint64(FlagSequence)
		opts = append(opts, cltypes.WithAccountSequence(accSeq))
	}

	gasAdj, _ := flagSet.GetFloat64(FlagGasAdjustment)
	opts = append(opts, cltypes.WithGasAdjustment(gasAdj))

	if flagSet.Changed(FlagNote) {
		memo, _ := flagSet.GetString(FlagNote)
		opts = append(opts, cltypes.WithNote(memo))
	}

	if flagSet.Changed(FlagTimeoutHeight) {
		timeoutHeight, _ := flagSet.GetUint64(FlagTimeoutHeight)
		opts = append(opts, cltypes.WithTimeoutHeight(timeoutHeight))
	}

	if flagSet.Changed(FlagSkipConfirmation) {
		skip, _ := flagSet.GetBool(FlagSkipConfirmation)
		opts = append(opts, cltypes.WithSkipConfirm(skip))
	}

	if flagSet.Changed(FlagGas) {
		gasStr, _ := flagSet.GetString(FlagGas)
		gasSetting, _ := ParseGasSetting(gasStr)
		opts = append(opts, cltypes.WithGas(gasSetting))
	}

	if flagSet.Changed(FlagFees) {
		feesStr, _ := flagSet.GetString(FlagFees)
		opts = append(opts, cltypes.WithFees(feesStr))
	}

	if flagSet.Changed(FlagGasPrices) {
		gasPrices, _ := flagSet.GetString(FlagGasPrices)
		opts = append(opts, cltypes.WithGasPrices(gasPrices))
	}

	signMode := SignModeDirect
	if flagSet.Changed(FlagSignMode) {
		signMode, _ = flagSet.GetString(FlagSignMode)
	}

	opts = append(opts, cltypes.WithSignMode(signMode))

	return opts, nil
}

// ParseGasSetting parses a string gas value. The value may either be 'auto',
// which indicates a transaction should be executed in simulate mode to
// automatically find a sufficient gas value, or a string integer. It returns an
// error if a string integer is provided which cannot be parsed.
func ParseGasSetting(gasStr string) (cltypes.GasSetting, error) {
	switch gasStr {
	case "":
		return cltypes.GasSetting{Simulate: false, Gas: DefaultGasLimit}, nil

	case GasFlagAuto:
		return cltypes.GasSetting{Simulate: true, Gas: 0}, nil

	default:
		gas, err := strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			return cltypes.GasSetting{}, fmt.Errorf("gas must be either integer or %s", GasFlagAuto)
		}

		return cltypes.GasSetting{Simulate: false, Gas: gas}, nil
	}
}

