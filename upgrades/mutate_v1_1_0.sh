#!/usr/bin/env bash
# Mutation falsification for upgrades/upgrade_v1_1_0_test.go.
#
# A test suite that passes proves nothing unless it fails when the code under
# test is broken. Each mutation below edits upgrades/software/v1.1.0/upgrade.go,
# runs the v1.1.0 suite, and records CAUGHT (suite went red) or ESCAPED (suite
# stayed green - a real gap in the tests).
#
# A no-op CONTROL row runs first and MUST PASS. Without it, a harness that always
# reports CAUGHT looks identical to a harness that measures.
set -uo pipefail

TARGET="upgrades/software/v1.1.0/upgrade.go"
BACKUP="$TMPDIR/v110_orig.go"
LOGDIR="$TMPDIR/mutlog"
mkdir -p "$LOGDIR"

restore() { cp "$BACKUP" "$TARGET"; }

# go vet precheck: a mutation that does not compile is a HARNESS ERROR, not a
# CAUGHT result. Scoring a broken build as "the test caught it" would be a lie.
score() {
  local name="$1"
  local out="$LOGDIR/$name.log"

  if ! go vet ./upgrades/ >"$LOGDIR/$name.vet" 2>&1; then
    printf '%-34s HARNESS ERROR (does not compile)\n' "$name"
    restore
    return
  fi

  if go test -count=1 -run 'TestUpgradeV110' ./upgrades/ >"$out" 2>&1; then
    printf '%-34s PASS (suite green)\n' "$name"
  else
    local failing
    failing=$(grep -cE '^--- FAIL' "$out")
    printf '%-34s CAUGHT (%d failing)\n' "$name" "$failing"
  fi
  restore
}

mutate() {
  local name="$1" from="$2" to="$3"
  python - "$TARGET" "$from" "$to" <<'PY'
import sys
path, frm, to = sys.argv[1], sys.argv[2], sys.argv[3]
src = open(path, encoding='utf-8').read()
if src.count(frm) != 1:
    print("ANCHOR-MISS count=%d for %r" % (src.count(frm), frm))
    sys.exit(3)
open(path, 'w', encoding='utf-8').write(src.replace(frm, to))
PY
  if [ $? -eq 3 ]; then
    printf '%-34s HARNESS ERROR (anchor not found)\n' "$name"
    return
  fi
  score "$name"
}

# mutate_line targets ONE specific line number, for anchors that appear more than
# once in the file. Scoring by line rather than by first occurrence matters: an
# ambiguous anchor that silently patches the wrong loop would report a mutation
# as CAUGHT when a different line was actually broken.
mutate_line() {
  local name="$1" lineno="$2" from="$3" to="$4"
  python - "$TARGET" "$lineno" "$from" "$to" <<'PY'
import sys
path, lineno, frm, to = sys.argv[1], int(sys.argv[2]), sys.argv[3], sys.argv[4]
lines = open(path, encoding='utf-8').read().split('\n')
idx = lineno - 1
if idx >= len(lines) or frm not in lines[idx]:
    print("LINE-MISS %d: %r not on that line (actual: %r)"
          % (lineno, frm, lines[idx] if idx < len(lines) else None))
    sys.exit(3)
lines[idx] = lines[idx].replace(frm, to)
open(path, 'w', encoding='utf-8').write('\n'.join(lines))
PY
  if [ $? -eq 3 ]; then
    printf '%-34s HARNESS ERROR (anchor not found)\n' "$name"
    return
  fi
  score "$name"
}

echo "=== v1.1.0 handler mutation matrix ==="

# CONTROL: identical source, unmutated. MUST PASS (green). A control that went
# red would mean the harness, not the mutation, decides the verdict.
score "CONTROL_no_change"

# 1. Owed computation: drop the MulInt64Mut so owed never scales by height.
mutate "owed_drop_height_multiplier" \
  'owed.MulInt64Mut(heightDelta)' \
  '_ = heightDelta'

# 2. Owed computation: sum the wrong thing (drop the accumulate).
mutate "owed_no_rate_accumulation" \
  'totalBlockRate.AddMut(pmnt.State.Rate.Amount)' \
  '_ = pmnt'

# 3. Funds[0] debit: subtract nothing.
mutate "funds0_no_debit" \
  'val.State.Funds[0].Amount = val.State.Funds[0].Amount.Sub(owed)' \
  '_ = owed'

# 4. Funds[0] indexing: guard it, which is the card's requested remediation.
#    The panic test MUST fail: it asserts the CURRENT unguarded behaviour.
mutate "funds0_add_length_guard" \
  'if val.State.Funds[0].Denom != "ibc/170C677610AC31DF0904FFE09CD3B5C657492170E7E52372E48756B71E56F2F1" {' \
  'if len(val.State.Funds) == 0 || val.State.Funds[0].Denom != "ibc/170C677610AC31DF0904FFE09CD3B5C657492170E7E52372E48756B71E56F2F1" {'

# 5. Denom guard: remove it, so every account is swept regardless of denom.
#    The scoping test MUST fail.
mutate "denom_guard_removed" \
  'if val.State.Funds[0].Denom != "ibc/170C677610AC31DF0904FFE09CD3B5C657492170E7E52372E48756B71E56F2F1" {
			continue
		}' \
  ''

# 6. Idempotence: delete the account's old key so a second run re-processes it.
#    The digest test MUST fail.
mutate "account_key_not_deleted" \
  'store.Delete(key)

		if !overdraft {' \
  'if !overdraft {'

# 7. Payment rewrite: drop the balance zeroing.
#    The zeroing line appears TWICE (first loop line 167, second loop line 229),
#    so an unqualified string anchor is ambiguous. Patch by line number instead:
#    only the FIRST-loop occurrence, which is the one the payment tests exercise.
mutate_line "payment_balance_not_zeroed" 167 \
  'payments[i].State.Balance.Amount.Set(sdkmath.LegacyZeroDec())' \
  '_ = payments[i]'

# 8. Payment rewrite: drop the Unsettled backfill.
mutate "payment_unsettled_not_backfilled" \
  'payments[i].State.Unsettled.Amount.Set(payments[i].State.Rate.Amount.MulInt64Mut(heightDelta))' \
  '_ = heightDelta'

# 9. Overdraft branch: always close, ignoring the overdraft decision.
mutate "overdraft_always_closed" \
  'if !overdraft {
			val.State.State = etypes.StateClosed
		}' \
  'val.State.State = etypes.StateClosed'

# 10. Overdraft latch: stop promoting the account when a payment is overdrawn.
#     MUST fail the OVERDRAWN branch test.
mutate "overdrawn_payment_no_latch" \
  'if pmnt.State.State == etypes.StateOverdrawn {
				val.State.State = etypes.StateOverdrawn
			}' \
  ''

restore
echo "=== restored pristine source ==="
git diff --stat -- "$TARGET"