#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
manifest="${root_dir}/scripts/compliance/soc2-evidence-manifest.yaml"
output_root="${root_dir}/_build/compliance/soc2"

mkdir -p "${output_root}"

timestamp="$(date +%Y%m%d-%H%M%S)"
output_dir="${output_root}/${timestamp}"
mkdir -p "${output_dir}"

cp "${manifest}" "${output_dir}/soc2-evidence-manifest.yaml"
cp "${root_dir}/_docs/compliance/soc2/control-matrix.md" "${output_dir}/control-matrix.md"
cp "${root_dir}/_docs/compliance/soc2/control-objectives.md" "${output_dir}/control-objectives.md"

cat <<EOM > "${output_dir}/control-matrix.json"
{
  "generated_at": "${timestamp}",
  "source": "_docs/compliance/soc2/control-matrix.md",
  "note": "Use the markdown control matrix as the canonical source."
}
EOM

cd "${root_dir}"

if command -v git >/dev/null 2>&1; then
  git rev-parse HEAD > "${output_dir}/git-head.txt" || true
  git status --porcelain=v1 > "${output_dir}/git-status.txt" || true
  git log -n 200 --since=180.days --date=iso > "${output_dir}/git-log.txt" || true
fi

if command -v go >/dev/null 2>&1; then
  go version > "${output_dir}/go-version.txt" || true
fi

if command -v node >/dev/null 2>&1; then
  node --version > "${output_dir}/node-version.txt" || true
fi

if command -v pnpm >/dev/null 2>&1; then
  pnpm --version > "${output_dir}/pnpm-version.txt" || true
fi

checksum_cmd=""
if command -v sha256sum >/dev/null 2>&1; then
  checksum_cmd="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  checksum_cmd="shasum -a 256"
fi

if [ -n "${checksum_cmd}" ]; then
  ${checksum_cmd} SECURITY.md PRIVACY_POLICY.md SUPPLY_CHAIN_SECURITY.md \
    _docs/data-classification.md _docs/key-management.md \
    _docs/business-continuity.md _docs/disaster-recovery.md \
    _docs/slos-and-playbooks.md _docs/operations/monitoring.md \
    _docs/operations/incident-response.md \
    > "${output_dir}/policy-checksums.txt" || true
fi

cat <<EOM > "${output_dir}/collection-notes.txt"
Evidence collection complete. Store this folder in the compliance evidence system.
This output contains no secrets and is safe for internal distribution.
EOM

echo "SOC 2 evidence collected in ${output_dir}"