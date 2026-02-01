package migrate

import (
	"github.com/cosmos/cosmos-sdk/codec"

	v1 "github.com/virtengine/virtengine/sdk/go/node/cert/v1"
	"github.com/virtengine/virtengine/sdk/go/node/cert/v1beta3"
)

func CertV1beta3Prefix() []byte {
	return v1beta3.PrefixCertificateID()
}

func CertFromV1beta3(cdc codec.BinaryCodec, fromBz []byte) v1.Certificate {
	var from v1beta3.Certificate
	cdc.MustUnmarshal(fromBz, &from)

	to := v1.Certificate{
		State:  v1.State(from.State),
		Cert:   make([]byte, len(from.Cert)),
		Pubkey: make([]byte, len(from.Pubkey)),
	}

	copy(to.Cert, from.Cert)
	copy(to.Pubkey, from.Pubkey)

	return to
}

