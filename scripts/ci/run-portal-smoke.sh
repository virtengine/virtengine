#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASE_URL="${PLAYWRIGHT_BASE_URL:-${VE_PORTAL_URL:-http://127.0.0.1:3000}}"
PROJECT="${PLAYWRIGHT_PROJECT:-chromium}"
GREP="${PLAYWRIGHT_GREP:-@smoke}"

export PLAYWRIGHT_BASE_URL="${BASE_URL%/}"
export PLAYWRIGHT_PROJECT="${PROJECT}"
export VE_PORTAL_URL="${PLAYWRIGHT_BASE_URL}"

mkdir -p "${ROOT_DIR}/output/playwright/portal-smoke"
printf '%s' "${PLAYWRIGHT_BASE_URL}" > "${ROOT_DIR}/output/playwright/portal-smoke/base-url.txt"

cd "${ROOT_DIR}/portal"

PLAYWRIGHT_BASE_URL="${PLAYWRIGHT_BASE_URL}" VE_PORTAL_URL="${VE_PORTAL_URL}" npx playwright test \
  --config ../tests/smoke/portal/playwright.config.ts \
  --project "${PROJECT}" \
  --grep "${GREP}"
