// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	errorsmod "cosmossdk.io/errors"
)

// IBC settlement module errors.
var (
	ErrInvalidVersion     = errorsmod.Register(ModuleName, 3100, "invalid IBC version")
	ErrInvalidPort        = errorsmod.Register(ModuleName, 3101, "invalid IBC port")
	ErrInvalidPacket      = errorsmod.Register(ModuleName, 3102, "invalid IBC packet")
	ErrRateLimited        = errorsmod.Register(ModuleName, 3103, "IBC rate limit exceeded")
	ErrChannelNotFound    = errorsmod.Register(ModuleName, 3104, "IBC channel not found")
	ErrHandshakeTimedOut  = errorsmod.Register(ModuleName, 3105, "IBC channel handshake timed out")
	ErrSettlementConflict = errorsmod.Register(ModuleName, 3106, "settlement record conflict")
)
