package migrate

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/virtengine/virtengine/sdk/go/node/provider/v1beta3"
	"github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

func ProviderInfoFromV1beta3(from v1beta3.ProviderInfo) v1beta4.Info {
	return v1beta4.Info{
		EMail:   from.EMail,
		Website: from.Website,
	}
}

func ProviderFromV1beta3(cdc codec.BinaryCodec, fromBz []byte) v1beta4.Provider {
	var from v1beta3.Provider
	cdc.MustUnmarshal(fromBz, &from)

	return v1beta4.Provider{
		Owner:      from.Owner,
		HostURI:    from.HostURI,
		Attributes: AttributesFromV1Beta3(from.Attributes),
		Info:       ProviderInfoFromV1beta3(from.Info),
	}
}

