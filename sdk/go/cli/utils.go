package cli

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/pflag"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	aclient "github.com/virtengine/virtengine/sdk/go/node/client"
	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

func DetectDeploymentDeposit(ctx context.Context, flags *pflag.FlagSet, cl aclient.QueryClient) (sdk.Coin, error) {
	var deposit sdk.Coin
	var depositStr string
	var err error

	if !flags.Changed(cflags.FlagDeposit) {
		resp, err := cl.Deployment().Params(ctx, &dtypes.QueryParamsRequest{})
		if err != nil {
			return sdk.Coin{}, err
		}

		// always default to AKT
		for _, sCoin := range resp.Params.MinDeposits {
			if sCoin.Denom == "uve" {
				depositStr = fmt.Sprintf("%s%s", sCoin.Amount, sCoin.Denom)
				break
			}
		}

		if depositStr == "" {
			return sdk.Coin{}, fmt.Errorf("couldn't query default deposit amount for uve")
		}
	} else {
		depositStr, err = flags.GetString(cflags.FlagDeposit)
		if err != nil {
			return sdk.Coin{}, err
		}
	}

	deposit, err = sdk.ParseCoinNormalized(depositStr)
	if err != nil {
		return sdk.Coin{}, err
	}

	return deposit, nil
}

func DetectBidDeposit(ctx context.Context, flags *pflag.FlagSet, cl aclient.QueryClient) (sdk.Coin, error) {
	var deposit sdk.Coin
	var depositStr string
	var err error

	if !flags.Changed(cflags.FlagDeposit) {
		resp, err := cl.Market().Params(ctx, &mtypes.QueryParamsRequest{})
		if err != nil {
			return sdk.Coin{}, err
		}

		depositStr = resp.Params.BidMinDeposit.String()
	} else {
		depositStr, err = flags.GetString(cflags.FlagDeposit)
		if err != nil {
			return sdk.Coin{}, err
		}
	}

	deposit, err = sdk.ParseCoinNormalized(depositStr)
	if err != nil {
		return sdk.Coin{}, err
	}

	return deposit, nil
}

func DepositSources(flags *pflag.FlagSet) (deposit.Sources, error) {
	sourcesStr, err := flags.GetStringSlice(cflags.FlagDepositSources)
	if err != nil {
		return nil, err
	}

	sources := make(deposit.Sources, 0, len(sourcesStr))

	dupMap := make(map[string]int)
	for _, source := range sourcesStr {
		val, valid := deposit.Source_value[source]
		if !valid {
			return nil, fmt.Errorf("invalid deposit-source \"%s\"", source)
		}

		if _, exists := dupMap[source]; exists {
			return nil, fmt.Errorf("duplicated deposit source \"%s\"", source)
		}

		dupMap[source] = 0

		sources = append(sources, deposit.Source(val))
	}

	return sources, nil
}

// DetectDeposit returns the deposit sources
func DetectDeposit(ctx context.Context, flags *pflag.FlagSet, cl aclient.QueryClient, querier func(ctx context.Context, flags *pflag.FlagSet, cl aclient.QueryClient) (sdk.Coin, error)) (deposit.Deposit, error) {
	amount, err := querier(ctx, flags, cl)
	if err != nil {
		return deposit.Deposit{}, err
	}

	sources, err := DepositSources(flags)
	if err != nil {
		return deposit.Deposit{}, err
	}

	return deposit.Deposit{
		Amount:  amount,
		Sources: sources,
	}, nil
}

func watchSignals(ctx context.Context, cancel context.CancelFunc) <-chan struct{} {
	donech := make(chan struct{})
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		defer close(donech)
		defer signal.Stop(sigch)
		select {
		case <-ctx.Done():
		case <-sigch:
			cancel()
		}
	}()
	return donech
}

// RunForever runs a function in the background, forever. Returns error in case of failure.
func RunForever(fn func(ctx context.Context) error) error {
	return RunForeverWithContext(context.Background(), fn)
}

func RunForeverWithContext(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)

	donech := watchSignals(ctx, cancel)

	err := fn(ctx)

	cancel()
	<-donech

	return err
}

func PrintJSON(ctx sdkclient.Context, v interface{}) error {
	marshaled, err := json.Marshal(v)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = json.Indent(buf, marshaled, "", "  ")
	if err != nil {
		return err
	}

	// Add a newline, for printing in the terminal
	_, err = buf.WriteRune('\n')
	if err != nil {
		return err
	}

	return ctx.PrintString(buf.String())
}

// ReadPageRequest reads and builds the necessary page request flags for pagination.
func ReadPageRequest(flagSet *pflag.FlagSet) (*query.PageRequest, error) {
	pageKeyStr, _ := flagSet.GetString(cflags.FlagPageKey)
	offset, _ := flagSet.GetUint64(cflags.FlagOffset)
	limit, _ := flagSet.GetUint64(cflags.FlagLimit)
	countTotal, _ := flagSet.GetBool(cflags.FlagCountTotal)
	page, _ := flagSet.GetUint64(cflags.FlagPage)
	reverse, _ := flagSet.GetBool(cflags.FlagReverse)

	if page > 1 && offset > 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "page and offset cannot be used together")
	}

	if page > 1 {
		offset = (page - 1) * limit
	}

	var pageKey []byte
	if pageKeyStr != "" {
		var err error
		pageKey, err = b64.StdEncoding.DecodeString(pageKeyStr)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid pagination key")
		}
	}

	return &query.PageRequest{
		Key:        pageKey,
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
		Reverse:    reverse,
	}, nil
}

