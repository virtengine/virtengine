package v1

import proto "github.com/cosmos/gogoproto/proto"

// String methods are required for legacy gogo proto message interfaces.
func (m *DeploymentID) String() string { return proto.CompactTextString(m) }
func (m *GroupID) String() string      { return proto.CompactTextString(m) }
