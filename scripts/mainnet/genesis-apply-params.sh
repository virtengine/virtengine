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

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

apply_filter() {
  local next_file="${TMP_GENESIS}.next"
  jq "$@" "$TMP_GENESIS" > "$next_file"
  mv "$next_file" "$TMP_GENESIS"
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
      die "unknown argument: $1"
      ;;
  esac
done

[[ -n "$GENESIS" ]] || {
  usage
  die "--genesis is required"
}
[[ -f "$GENESIS" ]] || die "genesis file not found: $GENESIS"
[[ -f "$PARAMS" ]] || die "params file not found: $PARAMS"

require_cmd jq

TMP_PARAMS=$(mktemp)
TMP_GENESIS=$(mktemp)
trap 'rm -f "$TMP_PARAMS" "$TMP_GENESIS" "${TMP_PARAMS}.next" "${TMP_GENESIS}.next"' EXIT

cp "$PARAMS" "$TMP_PARAMS"
cp "$GENESIS" "$TMP_GENESIS"

if [[ -n "$CHAIN_ID" || -n "$GENESIS_TIME" ]]; then
  jq \
    --arg chain_id "$CHAIN_ID" \
    --arg genesis_time "$GENESIS_TIME" \
    '(.chain_id = (if $chain_id != "" then $chain_id else .chain_id end)) |
     (.genesis_time = (if $genesis_time != "" then $genesis_time else .genesis_time end))' \
    "$TMP_PARAMS" > "${TMP_PARAMS}.next"
  mv "${TMP_PARAMS}.next" "$TMP_PARAMS"
fi

apply_filter --slurpfile params "$TMP_PARAMS" '
  .app_name =
    (if ($params[0].app_name? // "") != "" then
       $params[0].app_name
     elif (.app_name // "") == "" or .app_name == "<appd>" then
       "virtengine"
     else
       .app_name
     end)'

while IFS= read -r key; do
  key=${key%$'\r'}
  [[ -n "$key" ]] || continue
  case "$key" in
    app_name)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_name = ($params[0].app_name // "virtengine")'
      ;;
    chain_id)
      apply_filter --slurpfile params "$TMP_PARAMS" '.chain_id = $params[0].chain_id'
      ;;
    genesis_time)
      apply_filter --slurpfile params "$TMP_PARAMS" '.genesis_time = $params[0].genesis_time'
      ;;
    consensus_params)
      if jq -e '.consensus_params? != null' "$TMP_PARAMS" >/dev/null 2>&1; then
        apply_filter --slurpfile params "$TMP_PARAMS" '.consensus_params = ((.consensus_params // {}) * $params[0].consensus_params)'
      fi
      ;;
    denom_metadata)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_state.bank.denom_metadata = $params[0].denom_metadata'
      ;;
    gov_params)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_state.gov.params = ((.app_state.gov.params // {}) * $params[0].gov_params)'
      ;;
    gov_deposit_params)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_state.gov.deposit_params = ((.app_state.gov.deposit_params // {}) * $params[0].gov_deposit_params)'
      ;;
    gov_voting_params)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_state.gov.voting_params = ((.app_state.gov.voting_params // {}) * $params[0].gov_voting_params)'
      ;;
    gov_tally_params)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_state.gov.tally_params = ((.app_state.gov.tally_params // {}) * $params[0].gov_tally_params)'
      ;;
    crisis_constant_fee)
      apply_filter --slurpfile params "$TMP_PARAMS" '.app_state.crisis.constant_fee = $params[0].crisis_constant_fee'
      ;;
    *_params)
      module_name=${key%_params}
      if jq -e --arg module "$module_name" '.app_state[$module].params? != null' "$TMP_GENESIS" >/dev/null 2>&1; then
        apply_filter --arg module "$module_name" --arg key "$key" --slurpfile params "$TMP_PARAMS" \
          '.app_state[$module].params = ((.app_state[$module].params // {}) * $params[0][$key])'
      elif jq -e --arg module "$module_name" '.app_state[$module].Params? != null' "$TMP_GENESIS" >/dev/null 2>&1; then
        apply_filter --arg module "$module_name" --arg key "$key" --slurpfile params "$TMP_PARAMS" \
          '.app_state[$module].Params = ((.app_state[$module].Params // {}) * $params[0][$key])'
      else
        die "params key '$key' does not map to a known module params path in $GENESIS"
      fi
      ;;
    *)
      die "unsupported params key '$key' in $PARAMS"
      ;;
  esac
done < <(jq -r 'keys[]' "$TMP_PARAMS")

mv "$TMP_GENESIS" "$GENESIS"
echo "Applied mainnet params to $GENESIS"
