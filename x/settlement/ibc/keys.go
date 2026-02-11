// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/binary"
)

// Store key prefixes for settlement IBC.
var (
	PrefixPendingPacket    = []byte{0x70}
	PrefixAckPacket        = []byte{0x71}
	PrefixTimeoutPacket    = []byte{0x72}
	PrefixRateLimit        = []byte{0x73}
	PrefixRateLimitRelayer = []byte{0x74}
	PrefixRateLimitConfig  = []byte{0x75}
	PrefixHandshake        = []byte{0x76}
)

// PendingPacketKey returns the store key for pending packet data.
func PendingPacketKey(channelID string, sequence uint64) []byte {
	key := make([]byte, 0, len(PrefixPendingPacket)+len(channelID)+1+8)
	key = append(key, PrefixPendingPacket...)
	key = append(key, []byte(channelID)...)
	key = append(key, byte('/'))
	key = appendUint64(key, sequence)
	return key
}

// AckPacketKey returns the store key for acknowledged packets.
func AckPacketKey(channelID string, sequence uint64) []byte {
	key := make([]byte, 0, len(PrefixAckPacket)+len(channelID)+1+8)
	key = append(key, PrefixAckPacket...)
	key = append(key, []byte(channelID)...)
	key = append(key, byte('/'))
	key = appendUint64(key, sequence)
	return key
}

// TimeoutPacketKey returns the store key for timed-out packets.
func TimeoutPacketKey(channelID string, sequence uint64) []byte {
	key := make([]byte, 0, len(PrefixTimeoutPacket)+len(channelID)+1+8)
	key = append(key, PrefixTimeoutPacket...)
	key = append(key, []byte(channelID)...)
	key = append(key, byte('/'))
	key = appendUint64(key, sequence)
	return key
}

// RateLimitKey returns the store key for global rate limit counters.
func RateLimitKey(blockHeight uint64, packetType PacketType) []byte {
	key := make([]byte, 0, len(PrefixRateLimit)+8+1+len(packetType))
	key = append(key, PrefixRateLimit...)
	key = appendUint64(key, blockHeight)
	key = append(key, byte('/'))
	key = append(key, []byte(packetType)...)
	return key
}

// RateLimitRelayerKey returns the store key for per-relayer rate limit counters.
func RateLimitRelayerKey(blockHeight uint64, relayer string, packetType PacketType) []byte {
	key := make([]byte, 0, len(PrefixRateLimitRelayer)+8+1+len(relayer)+1+len(packetType))
	key = append(key, PrefixRateLimitRelayer...)
	key = appendUint64(key, blockHeight)
	key = append(key, byte('/'))
	key = append(key, []byte(relayer)...)
	key = append(key, byte('/'))
	key = append(key, []byte(packetType)...)
	return key
}

// RateLimitConfigKey returns the store key for rate limit configuration.
func RateLimitConfigKey() []byte {
	return PrefixRateLimitConfig
}

// HandshakeKey returns the store key for a channel handshake start record.
func HandshakeKey(channelID string) []byte {
	key := make([]byte, 0, len(PrefixHandshake)+len(channelID))
	key = append(key, PrefixHandshake...)
	key = append(key, []byte(channelID)...)
	return key
}

func appendUint64(bz []byte, n uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return append(bz, buf...)
}

func readUint64(bz []byte) uint64 {
	if len(bz) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}
