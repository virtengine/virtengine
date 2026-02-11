#!/bin/bash
# scripts/ci/backup-restore-smoke-test.sh
# CI smoke test for backup/restore and snapshot integrity verification.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="${ROOT_DIR}/scripts/dr"

if [ ! -x "${SCRIPT_DIR}/backup-chain-state.sh" ]; then
    echo "backup-chain-state.sh not found or not executable" >&2
    exit 1
fi

if ! command -v openssl > /dev/null 2>&1; then
    echo "openssl not available" >&2
    exit 1
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

NODE_HOME="${TMP_DIR}/node"
SNAPSHOT_DIR="${TMP_DIR}/snapshots"
MOCK_BIN="${TMP_DIR}/bin"
KEY_DIR="${TMP_DIR}/keys"

mkdir -p "$NODE_HOME/data" "$SNAPSHOT_DIR" "$MOCK_BIN" "$KEY_DIR"

# Create mock virtengine binary
cat > "${MOCK_BIN}/virtengine" << 'EOF'
#!/bin/bash
set -euo pipefail
if [ "$1" = "status" ]; then
  cat << JSON
{
  "sync_info": {
    "latest_block_height": "1000",
    "latest_app_hash": "ABCDEF1234",
    "catching_up": false
  },
  "node_info": {
    "network": "virtengine-test"
  }
}
JSON
  exit 0
fi

if [ "$1" = "export" ]; then
  echo '{"app_state":"ok"}'
  exit 0
fi

echo "unsupported command" >&2
exit 1
EOF
chmod +x "${MOCK_BIN}/virtengine"

# Create mock systemctl
cat > "${MOCK_BIN}/systemctl" << 'EOF'
#!/bin/bash
exit 0
EOF
chmod +x "${MOCK_BIN}/systemctl"

# Generate signing keys
openssl genpkey -algorithm RSA -out "${KEY_DIR}/snapshot_signing.pem" -pkeyopt rsa_keygen_bits:2048 > /dev/null 2>&1
openssl pkey -in "${KEY_DIR}/snapshot_signing.pem" -pubout -out "${KEY_DIR}/snapshot_signing.pub" > /dev/null 2>&1

export PATH="${MOCK_BIN}:$PATH"
export NODE_HOME
export SNAPSHOT_DIR
export SNAPSHOT_SIGNING_KEY="${KEY_DIR}/snapshot_signing.pem"
export SNAPSHOT_VERIFY_PUBKEY="${KEY_DIR}/snapshot_signing.pub"
export SNAPSHOT_SIGNING_REQUIRED=1
export RESTORE_SKIP_SERVICE=1
export RESTORE_AUTO_APPROVE=1
export RESTORE_FALLBACK_ENABLED=1
export SYSTEMCTL_CMD="${MOCK_BIN}/systemctl"
export VIRTENGINE_CMD="${MOCK_BIN}/virtengine"

# Snapshot 1
echo "snapshot-one" > "${NODE_HOME}/data/state.txt"
"${SCRIPT_DIR}/backup-chain-state.sh" --snapshot-only > /dev/null

first_snapshot=$(ls -t "${SNAPSHOT_DIR}"/state_*_data.tar.gz | tail -1 | xargs -r basename | sed 's/_data.tar.gz//')

# Snapshot 2 with different data
sleep 1
echo "snapshot-two" > "${NODE_HOME}/data/state.txt"
"${SCRIPT_DIR}/backup-chain-state.sh" --snapshot-only > /dev/null

latest_snapshot=$(ls -t "${SNAPSHOT_DIR}"/state_*_data.tar.gz | head -1 | xargs -r basename | sed 's/_data.tar.gz//')

# Corrupt latest snapshot
printf "corrupt" >> "${SNAPSHOT_DIR}/${latest_snapshot}_data.tar.gz"

# Restore should fall back to the valid snapshot
"${SCRIPT_DIR}/backup-chain-state.sh" --restore 1000 > /dev/null 2>&1

# Verify restored data matches first snapshot
if ! grep -q "snapshot-one" "${NODE_HOME}/data/state.txt"; then
    echo "restore did not fallback to valid snapshot" >&2
    exit 1
fi

echo "backup/restore smoke test passed"
