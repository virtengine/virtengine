package migrate

import (
	"github.com/cosmos/cosmos-sdk/codec"

	v1 "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	"github.com/virtengine/virtengine/sdk/go/node/audit/v1beta3"
)

func AuditedAttributesV1beta3Prefix() []byte {
	return v1beta3.PrefixProviderID()
}

func AuditedProviderFromV1beta3(cdc codec.BinaryCodec, fromBz []byte) v1.AuditedProvider {
	var from v1beta3.Provider
	cdc.MustUnmarshal(fromBz, &from)

	to := v1.AuditedProvider{
		Owner:      from.Owner,
		Auditor:    from.Auditor,
		Attributes: AttributesFromV1Beta3(from.Attributes),
	}

	return to
}
