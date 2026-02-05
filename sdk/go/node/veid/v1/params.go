package v1

import proto "github.com/cosmos/gogoproto/proto"

// String satisfies the proto.Message interface for Params when generated code omits it.
func (m *Params) String() string {
	return proto.CompactTextString(m)
}
