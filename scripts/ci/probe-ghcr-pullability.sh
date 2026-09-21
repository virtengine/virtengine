#!/usr/bin/env bash
#
# Probe whether a GHCR package is anonymously pullable.
#
# Why this exists
# ---------------
# The DR CronJob pods (infra/kubernetes/dr/backup-cronjobs.yaml) run under
# dr-backup-sa, which mounts NO imagePullSecrets. They can therefore only ever
# pull an anonymously-readable package. A digest pin that resolves for an
# authenticated CI runner can still be unreachable from the cluster, so the
# digest check alone is not sufficient evidence — this probe covers that gap.
#
# Why it is self-validating
# -------------------------
# A pullability probe has one dangerous failure mode: reporting "not pullable"
# because the probe itself is broken (bad token exchange, wrong header, network
# policy), which is indistinguishable from a genuinely private package and
# silently certifies the wrong thing forever. So every run probes a KNOWN
# anonymous-public control first and only interprets the subject's result when
# the control returned 200:
#
#   control == 200, subject == 200  -> anonymously pullable (success branch)
#   control == 200, subject != 200  -> real finding: the pod will ImagePullBackOff
#   control != 200                  -> INCONCLUSIVE, the probe itself is suspect
#
# The `--self-test` mode points the probe at the control as its subject and
# requires the 200 success branch to actually print, which is what proves the
# success branch is reachable rather than dead code.
#
# Usage
#   scripts/ci/probe-ghcr-pullability.sh [options]
#     --subject <repo> <ref>   default: parsed from $CRONJOB_MANIFEST
#     --control <repo> <ref>   default: a known anonymous-public manifest
#     --self-test              subject := control; fail unless the 200 branch fires
#     -h, --help
#
# Exit codes: 0 advisory result reported (including warnings),
#             1 self-test failed or usage error.
#
# Advisory by design: the package visibility decision is an owner call
# (t_46de26f7), so a private package must not fail unrelated PRs.

set -uo pipefail

MANIFEST="${CRONJOB_MANIFEST:-infra/kubernetes/dr/backup-cronjobs.yaml}"

# A manifest that is anonymous-public today and content-addressed, so it cannot
# silently change. Used purely as a control: if THIS does not come back 200, the
# probe is misbehaving and the subject result must not be trusted.
CONTROL_REPO="${CONTROL_REPO:-virtengine/virtengine}"
CONTROL_REF="${CONTROL_REF:-sha256:da0ac87c431e4774ba327604769f7609f2e56275ece6e398c4d5a3f744f8219b}"

# Composed at runtime rather than written inline. Output from agent/review
# tooling that masks secret-like text rewrites the adjacent form
# "Bearer <identifier>" to "Bearer ***" on display, which already produced one
# false review finding against this file (t_46de26f7 round 3: "the probe sends a
# literal ***", with the requested fix rendering identically to the code it
# called buggy). Keeping the scheme in a variable means a reviewer reading this
# through such a mask still sees a variable reference, not a placeholder.
AUTH_SCHEME='Bearer'

subject_repo=''
subject_ref=''
self_test=0

usage() { sed -n '2,40p' "$0" | sed 's/^# \{0,1\}//'; }

die() { echo "error: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --subject)
      [[ $# -ge 3 ]] || die "--subject needs <repo> <ref>"
      subject_repo="$2"; subject_ref="$3"; shift 3 ;;
    --control)
      [[ $# -ge 3 ]] || die "--control needs <repo> <ref>"
      CONTROL_REPO="$2"; CONTROL_REF="$3"; shift 3 ;;
    --self-test) self_test=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown argument: $1" ;;
  esac
done

# The pinned reference the CronJobs actually use.
if [[ -z "$subject_repo" || -z "$subject_ref" ]]; then
  [[ -f "$MANIFEST" ]] || die "CronJob manifest not found: $MANIFEST"
  pinned="$(grep -oE 'ghcr\.io/virtengine/dr-tools@sha256:[a-f0-9]{64}' "$MANIFEST" | head -n1)"
  [[ -n "$pinned" ]] || die "no dr-tools digest pin found in $MANIFEST"
  subject_repo="${pinned%@*}"
  subject_repo="${subject_repo#ghcr.io/}"
  subject_ref="${pinned##*@}"
fi

if [[ "$self_test" == "1" ]]; then
  subject_repo="$CONTROL_REPO"
  subject_ref="$CONTROL_REF"
fi

# Anonymous pull token for one repository. Empty on denial or any malformed body
# (a 403 page, an empty response) — a traceback here would bury the warning.
token_for() { # repo
  curl -sS --max-time 30 \
    "https://ghcr.io/token?service=ghcr.io&scope=repository:$1:pull" 2>/dev/null \
    | python3 -c 'import json,sys
d = sys.stdin.read().strip()
print(json.loads(d).get("token", "") if d.startswith("{") else "")' 2>/dev/null || true
}

# Manifest HEAD-ish check. GET with -o /dev/null; body is not needed.
manifest_code() { # repo ref token
  curl -s --max-time 30 -o /dev/null -w '%{http_code}' \
    -H "Authorization: ${AUTH_SCHEME} $3" \
    "https://ghcr.io/v2/$1/manifests/$2" 2>/dev/null || echo 000
}

# probe <repo> <ref> -> prints "<code> <token_len>", or "denied 0" when the
# anonymous token exchange itself was refused. The two are different findings:
# a denied token means the repository is not anonymously readable at all, which
# is what a never-published or private package returns, whereas a non-200 code
# with a valid token means the repo exists but that manifest does not.
probe() { # repo ref
  local tok code
  tok="$(token_for "$1")"
  [[ -n "$tok" ]] || { echo "denied 0"; return 0; }
  code="$(manifest_code "$1" "$2" "$tok")"
  echo "${code:-000} ${#tok}"
}

warn_private() { # repo ref code
  echo "::warning::anonymous manifest fetch returned $3 for $1@$2."
  echo "The DR CronJob pods mount no imagePullSecrets, so they will stay in"
  echo "ImagePullBackOff. Either make the package public or add a pull secret"
  echo "to dr-backup-sa. Decision tracked by t_46de26f7."
}

echo "probing anonymous pullability (control: ${CONTROL_REPO}@${CONTROL_REF:0:19}...)"

# The control is retried before it is declared broken. It is a network call to a
# registry that this job does not own, and the failure mode is fail-closed (the
# self-test below exits 1), so a single transient blip would red the gate for a
# reason that has nothing to do with the DR image. Three attempts absorbs that
# without weakening the conclusion: if the control genuinely cannot resolve, the
# result is still INCONCLUSIVE and nothing is certified.
control_code=''
control_len=''
for attempt in 1 2 3; do
  read -r control_code control_len <<<"$(probe "$CONTROL_REPO" "$CONTROL_REF")"
  [[ "$control_code" == "200" ]] && break
  echo "control attempt ${attempt}/3 -> HTTP ${control_code}"
  [[ "$attempt" -lt 3 ]] && sleep 5
done
echo "control ${CONTROL_REPO} -> HTTP ${control_code} (token_len=${control_len})"

if [[ "$control_code" != "200" ]]; then
  echo "::error::probe INCONCLUSIVE: the known-public control returned ${control_code} on 3/3 attempts."
  echo "The probe itself is not working (token exchange, egress policy or the"
  echo "control manifest moved), so the subject result below cannot be trusted."
  echo "No conclusion is drawn about ${subject_repo}@${subject_ref}."
  echo "Failing closed: an unverifiable pullability check must not report success."
  exit 1
fi
echo "ok    control resolved anonymously -> the probe's 200 path is live"

read -r subject_code subject_len <<<"$(probe "$subject_repo" "$subject_ref")"
echo "subject ${subject_repo}@${subject_ref} -> HTTP ${subject_code} (token_len=${subject_len})"

if [[ "$subject_code" == "200" ]]; then
  # The reachable success branch. Exercised by --self-test.
  echo "ANONYMOUS-PULL-OK: ${subject_repo}@${subject_ref} is anonymously pullable"
  echo "package is anonymously pullable — the CronJob image reference is complete."
  exit 0
fi

warn_private "$subject_repo" "$subject_ref" "$subject_code"
if [[ "$self_test" == "1" ]]; then
  echo "SELF-TEST FAILED: expected the 200 success branch, got ${subject_code}"
  exit 1
fi
exit 0
