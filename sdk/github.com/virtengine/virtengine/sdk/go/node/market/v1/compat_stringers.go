package v1

import proto "github.com/cosmos/gogoproto/proto"

// String methods for gogo proto compatibility.
func (m *BidID) String() string   { return proto.CompactTextString(m) }
func (m *LeaseID) String() string { return proto.CompactTextString(m) }
func (m *Lease) String() string   { return proto.CompactTextString(m) }
func (m *OrderID) String() string { return proto.CompactTextString(m) }
