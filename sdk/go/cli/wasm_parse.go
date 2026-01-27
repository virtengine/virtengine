package cli

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	flag "github.com/spf13/pflag"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm/v3"
	"github.com/distribution/reference"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

func ParseWasmVerificationFlags(gzippedWasm []byte, flags *flag.FlagSet) (string, string, []byte, error) {
	source, err := flags.GetString(cflags.FlagSource)
	if err != nil {
		return "", "", nil, fmt.Errorf("source: %s", err)
	}
	builder, err := flags.GetString(cflags.FlagBuilder)
	if err != nil {
		return "", "", nil, fmt.Errorf("builder: %s", err)
	}
	codeHash, err := flags.GetBytesHex(cflags.FlagCodeHash)
	if err != nil {
		return "", "", nil, fmt.Errorf("codeHash: %s", err)
	}

	// if any set require others to be set
	if len(source) != 0 || len(builder) != 0 || len(codeHash) != 0 {
		if source == "" {
			return "", "", nil, errors.New("source is required")
		}
		if _, err = url.ParseRequestURI(source); err != nil {
			return "", "", nil, fmt.Errorf("source: %s", err)
		}
		if builder == "" {
			return "", "", nil, errors.New("builder is required")
		}
		if _, err := reference.ParseDockerRef(builder); err != nil {
			return "", "", nil, fmt.Errorf("builder: %s", err)
		}
		if len(codeHash) == 0 {
			return "", "", nil, errors.New("code hash is required")
		}
		// wasm is gzipped in parseStoreCodeArgs
		// checksum generation will be decoupled here
		// reference https://github.com/CosmWasm/wasmvm/issues/359
		raw, err := ioutils.Uncompress(gzippedWasm, int64(types.MaxWasmSize))
		if err != nil {
			return "", "", nil, fmt.Errorf("invalid zip: %w", err)
		}
		checksum, err := wasmvm.CreateChecksum(raw)
		if err != nil {
			return "", "", nil, fmt.Errorf("checksum: %s", err)
		}
		if !bytes.Equal(checksum[:], codeHash) {
			return "", "", nil, fmt.Errorf("code-hash mismatch: %X, checksum: %X", codeHash, checksum)
		}
	}
	return source, builder, codeHash, nil
}

// ParseWasmStoreCodeArgs prepares MsgStoreCode object from flags with gzipped wasm byte code field
func ParseWasmStoreCodeArgs(file, sender string, flags *flag.FlagSet) (types.MsgStoreCode, error) {
	wasm, err := os.ReadFile(file)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)
		if err != nil {
			return types.MsgStoreCode{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCode{}, errors.New("invalid input file. Use wasm binary or gzip")
	}

	perm, err := ParseWasmAccessConfigFlags(flags)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	msg := types.MsgStoreCode{
		Sender:                sender,
		WASMByteCode:          wasm,
		InstantiatePermission: perm,
	}
	return msg, msg.ValidateBasic()
}

func ParseWasmAccessConfigFlags(flags *flag.FlagSet) (*types.AccessConfig, error) {
	addrs, err := flags.GetStringSlice(cflags.FlagInstantiateByAnyOfAddress)
	if err != nil {
		return nil, fmt.Errorf("flag any of: %s", err)
	}
	if len(addrs) != 0 {
		acceptedAddrs := make([]sdk.AccAddress, len(addrs))
		for i, v := range addrs {
			acceptedAddrs[i], err = sdk.AccAddressFromBech32(v)
			if err != nil {
				return nil, fmt.Errorf("parse %q: %w", v, err)
			}
		}
		x := types.AccessTypeAnyOfAddresses.With(acceptedAddrs...)
		return &x, nil
	}

	onlyAddrStr, err := flags.GetString(cflags.FlagInstantiateByAddress)
	if err != nil {
		return nil, fmt.Errorf("instantiate by address: %s", err)
	}
	if onlyAddrStr != "" {
		return nil, fmt.Errorf("not supported anymore. Use: %s", cflags.FlagInstantiateByAnyOfAddress)
	}
	everybodyStr, err := flags.GetString(cflags.FlagInstantiateByEverybody)
	if err != nil {
		return nil, fmt.Errorf("instantiate by everybody: %s", err)
	}
	if everybodyStr != "" {
		ok, err := strconv.ParseBool(everybodyStr)
		if err != nil {
			return nil, fmt.Errorf("boolean value expected for instantiate by everybody: %s", err)
		}
		if ok {
			return &types.AllowEverybody, nil
		}
	}

	nobodyStr, err := flags.GetString(cflags.FlagInstantiateNobody)
	if err != nil {
		return nil, fmt.Errorf("instantiate by nobody: %s", err)
	}
	if nobodyStr != "" {
		ok, err := strconv.ParseBool(nobodyStr)
		if err != nil {
			return nil, fmt.Errorf("boolean value expected for instantiate by nobody: %s", err)
		}
		if ok {
			return &types.AllowNobody, nil
		}
	}
	return nil, nil
}

func ParseWasmInstantiateArgs(rawCodeID, initMsg string, kr keyring.Keyring, sender string, flags *flag.FlagSet) (*types.MsgInstantiateContract, error) {
	// get the id of the code to instantiate
	codeID, err := strconv.ParseUint(rawCodeID, 10, 64)
	if err != nil {
		return nil, err
	}

	amountStr, err := flags.GetString(cflags.FlagAmount)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	label, err := flags.GetString(cflags.FlagLabel)
	if err != nil {
		return nil, fmt.Errorf("label: %s", err)
	}
	if label == "" {
		return nil, errors.New("label is required on all contracts")
	}
	adminStr, err := flags.GetString(cflags.FlagAdmin)
	if err != nil {
		return nil, fmt.Errorf("admin: %s", err)
	}

	noAdmin, err := flags.GetBool(cflags.FlagNoAdmin)
	if err != nil {
		return nil, fmt.Errorf("no-admin: %s", err)
	}

	// ensure sensible admin is set (or explicitly immutable)
	if adminStr == "" && !noAdmin {
		return nil, errors.New("you must set an admin or explicitly pass --no-admin to make it immutable (wasmd issue #719)")
	}
	if adminStr != "" && noAdmin {
		return nil, errors.New("you set an admin and passed --no-admin, those cannot both be true")
	}

	if adminStr != "" {
		addr, err := sdk.AccAddressFromBech32(adminStr)
		if err != nil {
			info, err := kr.Key(adminStr)
			if err != nil {
				return nil, fmt.Errorf("admin %s", err)
			}
			admin, err := info.GetAddress()
			if err != nil {
				return nil, err
			}
			adminStr = admin.String()
		} else {
			adminStr = addr.String()
		}
	}

	// build and sign the transaction, then broadcast to Tendermint
	msg := types.MsgInstantiateContract{
		Sender: sender,
		CodeID: codeID,
		Label:  label,
		Funds:  amount,
		Msg:    []byte(initMsg),
		Admin:  adminStr,
	}
	return &msg, msg.ValidateBasic()
}

func ParseWasmExecuteArgs(contractAddr, execMsg string, sender sdk.AccAddress, flags *flag.FlagSet) (types.MsgExecuteContract, error) {
	amountStr, err := flags.GetString(cflags.FlagAmount)
	if err != nil {
		return types.MsgExecuteContract{}, fmt.Errorf("amount: %s", err)
	}

	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return types.MsgExecuteContract{}, err
	}

	return types.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: contractAddr,
		Funds:    amount,
		Msg:      []byte(execMsg),
	}, nil
}

func ParseWasmAccessConfig(raw string) (c types.AccessConfig, err error) {
	switch raw {
	case "nobody":
		return types.AllowNobody, nil
	case "everybody":
		return types.AllowEverybody, nil
	default:
		parts := strings.Split(raw, ",")
		addrs := make([]sdk.AccAddress, len(parts))
		for i, v := range parts {
			addr, err := sdk.AccAddressFromBech32(v)
			if err != nil {
				return types.AccessConfig{}, fmt.Errorf("unable to parse address %q: %s", v, err)
			}
			addrs[i] = addr
		}
		defer func() { // convert panic in ".With" to error for better output
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
		cfg := types.AccessTypeAnyOfAddresses.With(addrs...)
		return cfg, cfg.ValidateBasic()
	}
}

func ParseWasmAccessConfigUpdates(args []string) ([]types.AccessConfigUpdate, error) {
	updates := make([]types.AccessConfigUpdate, len(args))
	for i, c := range args {
		// format: code_id:access_config
		// access_config: nobody|everybody|address(es)
		parts := strings.Split(c, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format")
		}

		codeID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid code ID: %s", err)
		}

		accessConfig, err := ParseWasmAccessConfig(parts[1])
		if err != nil {
			return nil, err
		}
		updates[i] = types.AccessConfigUpdate{
			CodeID:                codeID,
			InstantiatePermission: accessConfig,
		}
	}
	return updates, nil
}

func ParseStoreCodeGrants(args []string) ([]types.CodeGrant, error) {
	grants := make([]types.CodeGrant, len(args))
	for i, c := range args {
		// format: code_hash:access_config
		// access_config: nobody|everybody|address(es)
		parts := strings.Split(c, ":")
		if len(parts) != 2 {
			return nil, errors.New("invalid format")
		}

		if parts[1] == "*" {
			grants[i] = types.CodeGrant{
				CodeHash: []byte(parts[0]),
			}
			continue
		}

		accessConfig, err := ParseWasmAccessConfig(parts[1])
		if err != nil {
			return nil, err
		}
		grants[i] = types.CodeGrant{
			CodeHash:              []byte(parts[0]),
			InstantiatePermission: &accessConfig,
		}
	}
	return grants, nil
}
