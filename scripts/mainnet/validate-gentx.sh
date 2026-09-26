#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/validate-gentx.sh \
  --gentx-dir <dir> \
  [--constraints <path>] \
  [--binary <virtengine-path>]

Validates gentx files against mainnet constraints.
USAGE
}

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

resolve_python() {
  if command -v python3 >/dev/null 2>&1 && python3 -c "import sys" >/dev/null 2>&1; then
    PYTHON_CMD=(python3)
    return 0
  fi
  if command -v python >/dev/null 2>&1 && python -c "import sys" >/dev/null 2>&1; then
    PYTHON_CMD=(python)
    return 0
  fi
  if command -v py >/dev/null 2>&1 && py -3 -c "import sys" >/dev/null 2>&1; then
    PYTHON_CMD=(py -3)
    return 0
  fi
  die "a working Python 3 interpreter is required"
}

trim() {
  printf '%s' "$1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
}

is_placeholder_string() {
  local value=$1
  [[ -z "$value" ]] && return 1

  local lower
  lower=$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]')

  [[ "$value" =~ ^\<.*\>$ ]] && return 0
  [[ "$lower" == *"placeholder"* ]] && return 0
  [[ "$lower" == *"example.com"* || "$lower" == *"example.org"* || "$lower" == *"example.net"* ]] && return 0
  [[ "$lower" == *"changeme"* || "$lower" == *"todo"* || "$lower" == *"tbd"* ]] && return 0
  [[ "$value" =~ ^(ve1|vevaloper1|virtengine1).*[xX]{4,}.*$ ]] && return 0
  return 1
}

dec_compare() {
  "${PYTHON_CMD[@]}" - "$1" "$2" "$3" <<'PY'
from decimal import Decimal
import sys

lhs = Decimal(sys.argv[1])
op = sys.argv[2]
rhs = Decimal(sys.argv[3])

if op == "ge":
    ok = lhs >= rhs
elif op == "le":
    ok = lhs <= rhs
else:
    raise ValueError(f"unsupported op: {op}")

print("true" if ok else "false")
PY
}

validate_memo() {
  "${PYTHON_CMD[@]}" - "$1" <<'PY'
import ipaddress
import re
import sys

memo = sys.argv[1].strip()
if not memo:
    raise SystemExit("memo is empty")

match = re.match(r"^([0-9a-f]{40,})@(.+):([0-9]{2,5})$", memo, re.IGNORECASE)
if not match:
    raise SystemExit("memo must look like <node-id>@<host>:<port>")

host = match.group(2).strip()
port = int(match.group(3))
if port < 1 or port > 65535:
    raise SystemExit("memo port must be between 1 and 65535")

host_lower = host.lower()
if host_lower in {"localhost", "0.0.0.0"} or host_lower.endswith(".local") or host_lower.endswith(".localdomain"):
    raise SystemExit("memo host must be routable and not localhost/private")
if any(domain in host_lower for domain in ("example.com", "example.org", "example.net")):
    raise SystemExit("memo host must not use example.* placeholder domains")
if "<" in host or ">" in host:
    raise SystemExit("memo host must not contain placeholder markers")

if host.startswith("[") and host.endswith("]"):
    host = host[1:-1]

try:
    ip = ipaddress.ip_address(host)
except ValueError:
    pass
else:
    if any([
        ip.is_private,
        ip.is_loopback,
        ip.is_link_local,
        ip.is_multicast,
        ip.is_reserved,
        ip.is_unspecified,
    ]):
        raise SystemExit("memo host must not use private, loopback, reserved, or unspecified IP space")
PY
}

GENTX_DIR=""
CONSTRAINTS="config/mainnet/gentx-constraints.json"
VIRTENGINE_BIN="${VIRTENGINE_BIN:-virtengine}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --gentx-dir)
      GENTX_DIR="$2"
      shift 2
      ;;
    --constraints)
      CONSTRAINTS="$2"
      shift 2
      ;;
    --binary)
      VIRTENGINE_BIN="$2"
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

[[ -n "$GENTX_DIR" ]] || {
  usage
  die "--gentx-dir is required"
}
[[ -d "$GENTX_DIR" ]] || die "gentx dir not found: $GENTX_DIR"
[[ -f "$CONSTRAINTS" ]] || die "constraints file not found: $CONSTRAINTS"

require_cmd jq
require_cmd "$VIRTENGINE_BIN"
resolve_python

min_self=$(jq -r '.min_self_delegation' "$CONSTRAINTS")
bond_denom=$(jq -r '.bond_denom' "$CONSTRAINTS")
min_rate=$(jq -r '.commission.min_rate' "$CONSTRAINTS")
max_rate=$(jq -r '.commission.max_rate' "$CONSTRAINTS")
max_change_rate=$(jq -r '.commission.max_change_rate' "$CONSTRAINTS")
description_fields_json=$(jq -c '.required_description_fields // .min_description_fields // ["moniker","identity","website"]' "$CONSTRAINTS")
require_security_contact=$(jq -r '.require_security_contact // true' "$CONSTRAINTS")

failures=0
declare -A seen_validator_addresses=()
declare -A seen_validator_pubkeys=()
declare -A seen_validator_accounts=()

shopt -s nullglob
files=("$GENTX_DIR"/*.json)
if [[ ${#files[@]} -eq 0 ]]; then
  die "no gentx files found in $GENTX_DIR"
fi

for file in "${files[@]}"; do
  msg='.body.messages[0]'
  message_count=$(jq -r '.body.messages | length' "$file")
  message_type=$(jq -r '.body.messages[0]["@type"] // (.body.messages[0].type // empty)' "$file")
  amount=$(jq -r "$msg.value.amount // $msg.amount.amount // $msg.amount[0].amount // $msg.value[0].amount // empty" "$file")
  denom=$(jq -r "$msg.value.denom // $msg.amount.denom // $msg.amount[0].denom // $msg.value[0].denom // empty" "$file")
  rate=$(jq -r "$msg.commission.rate // empty" "$file")
  max_rate_val=$(jq -r "$msg.commission.max_rate // empty" "$file")
  max_change_val=$(jq -r "$msg.commission.max_change_rate // empty" "$file")
  min_self_val=$(jq -r "$msg.min_self_delegation // empty" "$file")
  website=$(jq -r "$msg.description.website // empty" "$file")
  security_contact=$(jq -r "$msg.description.security_contact // empty" "$file")
  details=$(jq -r "$msg.description.details // empty" "$file")
  memo=$(jq -r '.body.memo // empty' "$file")
  validator_address=$(jq -r "$msg.validator_address // empty" "$file")
  pubkey_type=$(jq -r "$msg.pubkey[\"@type\"] // empty" "$file")
  pubkey_key=$(jq -r "$msg.pubkey.key // empty" "$file")
  signature_count=$(jq -r '.signatures | length' "$file")

  if [[ "$message_count" != "1" ]]; then
    echo "FAIL: $file must contain exactly one genesis message" >&2
    failures=$((failures + 1))
    continue
  fi

  if [[ "$message_type" != "/cosmos.staking.v1beta1.MsgCreateValidator" ]]; then
    echo "FAIL: $file message type $message_type is not MsgCreateValidator" >&2
    failures=$((failures + 1))
  fi

  if [[ -z "$amount" || -z "$denom" || ! "$amount" =~ ^[0-9]+$ ]]; then
    echo "FAIL: $file missing or invalid self-delegation amount/denom" >&2
    failures=$((failures + 1))
  fi

  if [[ "$denom" != "$bond_denom" ]]; then
    echo "FAIL: $file denom $denom != $bond_denom" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$min_self_val" && "$min_self_val" =~ ^[0-9]+$ ]]; then
    if [[ "$min_self_val" -lt "$min_self" ]]; then
      echo "FAIL: $file min_self_delegation $min_self_val < $min_self" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing min_self_delegation" >&2
    failures=$((failures + 1))
  fi

  if [[ "$amount" =~ ^[0-9]+$ && "$amount" -lt "$min_self" ]]; then
    echo "FAIL: $file stake amount $amount < $min_self" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$rate" ]]; then
    if [[ "$(dec_compare "$rate" ge "$min_rate")" != "true" ]]; then
      echo "FAIL: $file commission rate $rate < $min_rate" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing commission rate" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$max_rate_val" ]]; then
    if [[ "$(dec_compare "$max_rate_val" le "$max_rate")" != "true" ]]; then
      echo "FAIL: $file commission max_rate $max_rate_val > $max_rate" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing commission max_rate" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$max_change_val" ]]; then
    if [[ "$(dec_compare "$max_change_val" le "$max_change_rate")" != "true" ]]; then
      echo "FAIL: $file commission max_change_rate $max_change_val > $max_change_rate" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing commission max_change_rate" >&2
    failures=$((failures + 1))
  fi

  while IFS= read -r required_field; do
    required_field=$(trim "$required_field")
    [[ -n "$required_field" ]] || continue
    field_value=$(jq -r "$msg.description[\"$required_field\"] // empty" "$file")
    if [[ -z "$field_value" ]]; then
      echo "FAIL: $file missing required validator description field '$required_field'" >&2
      failures=$((failures + 1))
      continue
    fi
    if is_placeholder_string "$field_value"; then
      echo "FAIL: $file uses placeholder content for '$required_field': $field_value" >&2
      failures=$((failures + 1))
    fi
  done < <(echo "$description_fields_json" | jq -r '.[]')

  if [[ -z "$website" || ! "$website" =~ ^https:// ]]; then
    echo "FAIL: $file website must use https://" >&2
    failures=$((failures + 1))
  fi

  if [[ "$require_security_contact" == "true" ]]; then
    if [[ -z "$security_contact" || ! "$security_contact" =~ ^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$ ]]; then
      echo "FAIL: $file requires a valid security_contact email" >&2
      failures=$((failures + 1))
    fi
  fi

  if [[ -n "$details" ]] && is_placeholder_string "$details"; then
    echo "FAIL: $file uses placeholder validator details" >&2
    failures=$((failures + 1))
  fi

  if [[ -z "$memo" ]]; then
    echo "FAIL: $file missing P2P memo" >&2
    failures=$((failures + 1))
  elif ! validate_memo "$memo" >/dev/null 2>&1; then
    echo "FAIL: $file has invalid or unroutable memo: $memo" >&2
    failures=$((failures + 1))
  fi

  if [[ -z "$validator_address" || ! "$validator_address" =~ ^vevaloper1[0-9a-z]+$ ]]; then
    echo "FAIL: $file missing or invalid validator_address" >&2
    failures=$((failures + 1))
  fi

  if [[ -z "$pubkey_type" || -z "$pubkey_key" ]]; then
    echo "FAIL: $file missing validator pubkey" >&2
    failures=$((failures + 1))
  fi

  if [[ "$signature_count" == "0" ]]; then
    echo "FAIL: $file has no signatures" >&2
    failures=$((failures + 1))
  fi

  validator_account=""
  if [[ -n "$validator_address" ]]; then
    if ! validator_account=$("$VIRTENGINE_BIN" debug bech32-convert --prefix ve "$validator_address" 2>/dev/null); then
      echo "FAIL: $file validator_address could not be converted to account prefix" >&2
      failures=$((failures + 1))
    fi
  fi

  if [[ -n "$validator_address" ]]; then
    if [[ -n "${seen_validator_addresses[$validator_address]+set}" ]]; then
      echo "FAIL: duplicate validator_address across gentxs: $validator_address" >&2
      failures=$((failures + 1))
    fi
    seen_validator_addresses["$validator_address"]=1
  fi

  if [[ -n "$pubkey_key" ]]; then
    if [[ -n "${seen_validator_pubkeys[$pubkey_key]+set}" ]]; then
      echo "FAIL: duplicate validator consensus pubkey across gentxs" >&2
      failures=$((failures + 1))
    fi
    seen_validator_pubkeys["$pubkey_key"]=1
  fi

  if [[ -n "$validator_account" ]]; then
    if [[ -n "${seen_validator_accounts[$validator_account]+set}" ]]; then
      echo "FAIL: duplicate validator funding account across gentxs: $validator_account" >&2
      failures=$((failures + 1))
    fi
    seen_validator_accounts["$validator_account"]=1
  fi
done

if [[ $failures -gt 0 ]]; then
  echo "Gentx validation failed: $failures issue(s)" >&2
  exit 1
fi

echo "Gentx validation passed for ${#files[@]} file(s)"
