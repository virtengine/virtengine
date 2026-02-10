#!/bin/bash
# Copyright 2024 VirtEngine Contributors
# SPDX-License-Identifier: Apache-2.0

# Monitor IBC relayer health and channel status.
# Run periodically via cron or systemd timer.

set -euo pipefail

LOG_FILE="${LOG_FILE:-/var/log/virtengine/relayer-monitor.log}"

log() {
  echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] $*" | tee -a "$LOG_FILE"
}

check_channel_health() {
  local chain_a="$1"
  local chain_b="$2"
  local port="$3"
  local channel="$4"

  local state
  state=$(hermes query channel end --chain "$chain_a" --port "$port" --channel "$channel" 2>/dev/null | grep -o '"state":"[^"]*"' | head -1 || echo "unknown")

  if echo "$state" | grep -q "Open"; then
    log "OK: Channel $channel ($port) between $chain_a and $chain_b is OPEN"
    return 0
  else
    log "WARN: Channel $channel ($port) between $chain_a and $chain_b state: $state"
    return 1
  fi
}

check_pending_packets() {
  local chain="$1"
  local port="$2"
  local channel="$3"

  local pending
  pending=$(hermes query packet pending --chain "$chain" --port "$port" --channel "$channel" 2>/dev/null | grep -c "sequence" || echo "0")

  if [ "$pending" -gt 100 ]; then
    log "ALERT: $pending pending packets on $chain:$port/$channel"
    return 1
  elif [ "$pending" -gt 0 ]; then
    log "INFO: $pending pending packets on $chain:$port/$channel"
  fi
  return 0
}

clear_stuck_packets() {
  local chain="$1"
  local port="$2"
  local channel="$3"

  log "Clearing stuck packets on $chain:$port/$channel..."
  hermes clear packets --chain "$chain" --port "$port" --channel "$channel" 2>/dev/null || {
    log "ERROR: Failed to clear packets on $chain:$port/$channel"
    return 1
  }
  log "Packets cleared on $chain:$port/$channel"
  return 0
}

# Main monitoring loop
log "=== VirtEngine IBC Health Check ==="

# Check transfer channels
check_channel_health "virtengine-1" "cosmoshub-4" "transfer" "channel-0" || true
check_channel_health "virtengine-1" "osmosis-1" "transfer" "channel-1" || true

# Check VEID attestation channel
check_channel_health "virtengine-1" "cosmoshub-4" "veid" "channel-2" || true

# Check pending packets
check_pending_packets "virtengine-1" "transfer" "channel-0" || true
check_pending_packets "virtengine-1" "transfer" "channel-1" || true
check_pending_packets "virtengine-1" "veid" "channel-2" || true

# Auto-clear if there are stuck packets (optional, controlled by env var)
if [ "${AUTO_CLEAR_PACKETS:-false}" = "true" ]; then
  clear_stuck_packets "virtengine-1" "transfer" "channel-0" || true
  clear_stuck_packets "virtengine-1" "transfer" "channel-1" || true
  clear_stuck_packets "virtengine-1" "veid" "channel-2" || true
fi

log "=== Health check complete ==="
