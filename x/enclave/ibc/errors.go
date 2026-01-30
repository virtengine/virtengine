package ibc

import (
	"cosmossdk.io/errors"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// IBC enclave module sentinel errors
var (
	// ErrInvalidVersion is returned when the IBC version is invalid
	ErrInvalidVersion = errors.Register(types.ModuleName, 2000, "invalid IBC module version")

	// ErrInvalidPacket is returned when a packet is invalid
	ErrInvalidPacket = errors.Register(types.ModuleName, 2001, "invalid IBC packet")

	// ErrUnknownPacketType is returned when the packet type is unknown
	ErrUnknownPacketType = errors.Register(types.ModuleName, 2002, "unknown packet type")

	// ErrChannelNotFound is returned when a channel is not found
	ErrChannelNotFound = errors.Register(types.ModuleName, 2003, "IBC channel not found")

	// ErrInvalidChannelState is returned when channel state is invalid
	ErrInvalidChannelState = errors.Register(types.ModuleName, 2004, "invalid channel state")

	// ErrInvalidPort is returned when the port ID is invalid
	ErrInvalidPort = errors.Register(types.ModuleName, 2005, "invalid port ID")

	// ErrPortAlreadyBound is returned when the port is already bound
	ErrPortAlreadyBound = errors.Register(types.ModuleName, 2006, "port already bound")

	// ErrUntrustedChannel is returned when trying to sync via an untrusted channel
	ErrUntrustedChannel = errors.Register(types.ModuleName, 2007, "channel is not trusted for sync operations")

	// ErrMeasurementSyncFailed is returned when measurement sync fails
	ErrMeasurementSyncFailed = errors.Register(types.ModuleName, 2008, "measurement sync failed")

	// ErrIdentityQueryFailed is returned when identity query fails
	ErrIdentityQueryFailed = errors.Register(types.ModuleName, 2009, "identity query failed")

	// ErrAcknowledgementFailed is returned when acknowledgement processing fails
	ErrAcknowledgementFailed = errors.Register(types.ModuleName, 2010, "acknowledgement processing failed")

	// ErrPacketTimeout is returned when a packet times out
	ErrPacketTimeout = errors.Register(types.ModuleName, 2011, "packet timeout")

	// ErrChannelCapabilityNotFound is returned when channel capability is not found
	ErrChannelCapabilityNotFound = errors.Register(types.ModuleName, 2012, "channel capability not found")

	// ErrInvalidCounterparty is returned when the counterparty is invalid
	ErrInvalidCounterparty = errors.Register(types.ModuleName, 2013, "invalid counterparty")

	// ErrFederatedIdentityNotTrusted is returned when a federated identity is not trusted
	ErrFederatedIdentityNotTrusted = errors.Register(types.ModuleName, 2014, "federated identity not trusted")

	// ErrMeasurementAlreadyExists is returned when trying to sync a measurement that already exists
	ErrMeasurementAlreadyExists = errors.Register(types.ModuleName, 2015, "measurement already exists")
)
