#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/genesis-apply-params.sh \
  --genesis <path> \
  [--params <path>] \
  [--chain-id <id>] \
  [--genesis-time <rfc3339>]

Applies mainnet parameter overrides to a genesis.json file.
USAGE
}

GENESIS=""
PARAMS="config/mainnet/genesis-params.json"
CHAIN_ID=""
GENESIS_TIME=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --genesis)
      GENESIS="$2"
      shift 2
      ;;
    --params)
      PARAMS="$2"
      shift 2
      ;;
    --chain-id)
      CHAIN_ID="$2"
      shift 2
      ;;
    --genesis-time)
      GENESIS_TIME="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$GENESIS" ]]; then
  echo "ERROR: --genesis is required" >&2
  usage
  exit 1
fi

if [[ ! -f "$GENESIS" ]]; then
  echo "ERROR: genesis file not found: $GENESIS" >&2
  exit 1
fi

if [[ ! -f "$PARAMS" ]]; then
  echo "ERROR: params file not found: $PARAMS" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

tmp_params=$(mktemp)
cp "$PARAMS" "$tmp_params"

if [[ -n "$CHAIN_ID" || -n "$GENESIS_TIME" ]]; then
  jq \
    --arg chain_id "$CHAIN_ID" \
    --arg genesis_time "$GENESIS_TIME" \
    '(.chain_id = (if $chain_id != "" then $chain_id else .chain_id end)) |
     (.genesis_time = (if $genesis_time != "" then $genesis_time else .genesis_time end))' \
    "$tmp_params" > "${tmp_params}.tmp"
  mv "${tmp_params}.tmp" "$tmp_params"
fi

tmp_genesis=$(mktemp)

jq --argfile params "$tmp_params" '
  .chain_id = $params.chain_id |
  .genesis_time = $params.genesis_time |
  (if $params.consensus_params? and $params.consensus_params != null then .consensus_params = $params.consensus_params else . end) |
  .app_state.bank.denom_metadata = $params.denom_metadata |
  .app_state.staking.params = $params.staking_params |
  .app_state.mint.params = $params.mint_params |
  .app_state.distribution.params = $params.distribution_params |
  (if .app_state.gov.params? then .app_state.gov.params = $params.gov_params else . end) |
  (if .app_state.gov.deposit_params? then .app_state.gov.deposit_params = $params.gov_deposit_params else . end) |
  (if .app_state.gov.voting_params? then .app_state.gov.voting_params = $params.gov_voting_params else . end) |
  (if .app_state.gov.tally_params? then .app_state.gov.tally_params = $params.gov_tally_params else . end) |
  .app_state.crisis.constant_fee = $params.crisis_constant_fee |
  .app_state.slashing.params = $params.slashing_params |
  .app_state.veid.params = $params.veid_params |
  .app_state.mfa.params = $params.mfa_params |
  .app_state.encryption.params = $params.encryption_params |
  .app_state.hpc.params = $params.hpc_params
' "$GENESIS" > "$tmp_genesis"

mv "$tmp_genesis" "$GENESIS"
rm -f "$tmp_params"

echo "Applied mainnet params to $GENESIS"