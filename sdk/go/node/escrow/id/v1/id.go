package v1

import (
	"bytes"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/node/escrow/module"
)

type ID interface {
	Key() string
}

var (
	_ ID = (*Account)(nil)
	_ ID = (*Payment)(nil)
)

func (obj *Account) Key() string {
	buf := &bytes.Buffer{}

	buf.WriteString(Scope_name[int32(obj.Scope)])
	buf.WriteRune('/')
	buf.WriteString(obj.XID)

	return buf.String()
}

func (obj *Payment) Key() string {
	buf := &bytes.Buffer{}

	buf.WriteString(obj.AID.Key())
	buf.WriteRune('/')
	buf.WriteString(obj.XID)

	return buf.String()
}

func (obj *Account) ValidateBasic() error {
	parts := strings.Split(obj.XID, "/")

	switch obj.Scope {
	case ScopeDeployment:
		if len(parts) != 2 {
			return module.ErrInvalidID.Wrap("invalid xid format")
		}
	case ScopeBid:
		if len(parts) != 5 {
			return module.ErrInvalidID.Wrap("invalid xid format")
		}
	default:
		return module.ErrInvalidID.Wrap("invalid scope")
	}

	_, err := sdk.AccAddressFromBech32(parts[0])
	if err != nil {
		return module.ErrInvalidID.Wrapf("invalid xid/owner: %s", err.Error())
	}

	_, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return module.ErrInvalidID.Wrapf("invalid xid/dseq: %s", err.Error())
	}

	if obj.Scope == ScopeBid {
		parts = parts[2:]
		err = validateBidScope(parts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (obj *Payment) ValidateBasic() error {
	err := obj.AID.ValidateBasic()
	if err != nil {
		return err
	}

	parts := strings.Split(obj.XID, "/")
	if len(parts) != 3 {
		return module.ErrInvalidID.Wrap("invalid xid format")
	}

	err = validateBidScope(parts)
	if err != nil {
		return err
	}

	return nil
}

func validateBidScope(parts []string) error {
	_, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return module.ErrInvalidID.Wrapf("invalid xid/gseq: %s", err.Error())
	}

	_, err = strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return module.ErrInvalidID.Wrapf("invalid xid/oseq: %s", err.Error())
	}

	_, err = sdk.AccAddressFromBech32(parts[2])
	if err != nil {
		return module.ErrInvalidID.Wrapf("invalid xid/provider: %s", err.Error())
	}

	return nil
}

func (obj *Payment) Account() Account {
	return obj.AID
}

