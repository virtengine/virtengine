package app

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	veid "github.com/virtengine/virtengine/x/veid"
)

func TestSystemTransactionUsesCanonicalSDKEncodingAndNoOrdinarySigner(t *testing.T) {
	encoding := sdkutil.MakeEncodingConfig(veid.AppModuleBasic{})
	message := &veidv1.MsgSubmitConsensusVerification{
		Version:        1,
		ChainId:        "chain-A",
		Height:         10,
		ExtendedCommit: []byte{1},
	}
	builder := encoding.TxConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(message))
	builder.SetGasLimit(systemTxGasLimit)
	builder.SetFeeAmount(sdk.NewCoins())
	encoded, err := encoding.TxConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	decoded, err := encoding.TxConfig.TxDecoder()(encoded)
	require.NoError(t, err)
	require.Len(t, decoded.GetMsgs(), 1)
	signable := decoded.(authsigning.SigVerifiableTx)
	signers, err := signable.GetSigners()
	require.NoError(t, err)
	require.Empty(t, signers)
	canonical, err := encoding.TxConfig.TxEncoder()(decoded)
	require.NoError(t, err)
	require.Equal(t, encoded, canonical)
	require.Equal(t, "0a4a0a480a322f76697274656e67696e652e766569642e76312e4d73675375626d6974436f6e73656e737573566572696669636174696f6e121208011207636861696e2d41180a2201012a001206120410c0843d", hex.EncodeToString(encoded))
}
