package types

import (
	"github.com/cosmos/gogoproto/proto"
)

type XID interface {
	proto.Message

	GetOwner() string
}
