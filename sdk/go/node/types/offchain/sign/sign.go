package sign

import (
	"encoding/json"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	MsgTypeSignData = ""
)

var _ sdk.Msg = (*MsgSignData)(nil)

func init() {
	MsgTypeSignData = reflect.TypeOf(&MsgSignData{}).Elem().Name()

	legacy.Cdc.RegisterConcrete(&MsgSignData{}, "sign/MsgSignData", nil)
}

// Type implements the sdk.Msg interface
func (m *MsgSignData) Type() string {
	return MsgTypeSignData
}

// GetSignBytes encodes the message for signing
func (m *MsgSignData) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(m))
}

func (m *MsgSignData) Route() string {
	return "signData"
}

// Deprecated: please delete this code eventually.
func mustSortJSON(bz []byte) []byte {
	var c any
	err := json.Unmarshal(bz, &c)
	if err != nil {
		panic(err)
	}
	js, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return js
}

// StdSignBytes returns the bytes to sign for a transaction.
// Deprecated: Please use x/tx/signing/aminojson instead.
func StdSignBytes(cdc *codec.LegacyAmino, chainID string, accnum, sequence, timeout uint64, fee legacytx.StdFee, msgs []sdk.Msg, memo string) []byte {
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		bz := cdc.MustMarshalJSON(msg)
		msgsBytes = append(msgsBytes, mustSortJSON(bz))
	}

	bz, err := cdc.MarshalJSON(legacytx.StdSignDoc{
		AccountNumber: accnum,
		ChainID:       chainID,
		Fee:           fee.Bytes(),
		Memo:          memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: timeout,
	})
	if err != nil {
		panic(err)
	}

	return mustSortJSON(bz)
}
