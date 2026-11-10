package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// BASELINE_COUNT is the number of unreadable annotations in this tree at the
// time of writing, measured with:
//
//	GOWORK=off go run ./scripts/gosecsuppressions -root . -write-baseline
//
// 370 is not a target anyone is trying to hit. It is a measurement of an
// existing condition: these annotations predate the guard and, on the evidence
// measured by panel review, sit on lines gosec does not currently report, so
// they are not 370 live findings. What the baseline buys is a stopping point --
// from here, one more is a failure.
//
// The count is deliberately NOT the enforcement mechanism. See the "ratchet" note
// in main.go: a total is defeated by deleting one annotation while shipping one
// annotation, so the enforced state is the keyed baseline file.

// Ratchet review window. A recorded budget with no expiry is a waiver nobody has
// to renew, which is the failure mode ESTATE.md's allowlist policy already
// guards against (policy.max_allowlist_age_days, capped at 30). Past this date
// the gate fails until the baseline is deliberately re-reviewed against a fresh
// measurement, so the budget cannot quietly outlive the evidence.
const (
	ratchetReviewed = "2026-10-06"
	ratchetExpires  = "2026-11-05"
)

// BUDGET_ISSUE records the decision to accept the existing 370 rather than
// rewrite them in one commit, and carries the work of paying the debt down.
// Required so the number is never unowned.
const budgetIssue = "virtengine/virtengine#1258"

// defaultAllowance is how many ADDITIONAL unreadable annotations above the
// recorded baseline are tolerated. Deliberately 0: the class this card exists to
// stop is a NEW file shipping //nolint-only suppressions, and any allowance > 0
// reopens it immediately. MAX_ALLOWANCE exists only so a deliberate, expiring,
// reviewable decision can be made in a diff -- not as headroom.
const defaultAllowance = 0

// maxAllowance is the hard ceiling on the MAX_ALLOWANCE override. An override
// that is unbounded makes the ratchet a no-op; this keeps one bad export from
// silently disabling the gate.
const maxAllowance = 5

// resolveAllowance is the allowance after the one-sided override is applied.
//
// The override may only TIGHTEN (lower) the allowance. It can never raise it:
// an env var that can raise the ceiling is a laundering vector -- anyone able to
// set a repo variable or export a var in the step could wave new suppressions
// through as "allowed drift" without touching the recorded budget, and the
// review would show no diff. ESTATE.md records that exact move being rejected on
// the pages perf gate. Raising the bar means editing defaultAllowance here, which
// puts it in review next to ratchetExpires and budgetIssue.
func resolveAllowance() int {
	raw := os.Getenv("GOSEC_SUPPRESSION_MAX")
	if raw == "" {
		return defaultAllowance
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 0 {
		fmt.Fprintf(os.Stderr,
			"::warning::GOSEC_SUPPRESSION_MAX=%q is not a non-negative integer; using the recorded allowance %d\n",
			raw, defaultAllowance)
		return defaultAllowance
	}
	if n > maxAllowance {
		// It is REFUSED, not clamped DOWN to the ceiling -- clamping would buy
		// the exporter 5 free annotations, which is exactly what they asked for,
		// and would make the cap itself a laundering vector. The override either
		// yields a value it is allowed to yield or none at all. An earlier
		// version returned maxAllowance here and the probe caught it: a request
		// for 99 allowances bought 5.
		fmt.Fprintf(os.Stderr,
			"::warning::GOSEC_SUPPRESSION_MAX=%d exceeds the hard ceiling %d and was REFUSED; using the recorded allowance %d. "+
				"To allow more unreadable annotations, edit defaultAllowance in "+
				"scripts/gosecsuppressions/baseline.go with a re-measurement and a review date (tracking %s).\n",
			n, maxAllowance, defaultAllowance, budgetIssue)
		return defaultAllowance
	}
	if n > defaultAllowance {
		fmt.Fprintf(os.Stderr,
			"::warning::GOSEC_SUPPRESSION_MAX=%d is ABOVE the recorded allowance %d and was REFUSED. "+
				"The override may only tighten the gate (tracking %s).\n",
			n, defaultAllowance, budgetIssue)
		return defaultAllowance
	}
	return n
}

// reviewExpired reports whether the recorded budget has outlived its window.
func reviewExpired(today string) bool {
	return today > ratchetExpires
}

// baselineEntry is one recorded (path, annotation) -> count pair.
type baselineEntry struct {
	path  string
	count int
	// text is kept for rewriting the file, not for comparison.
	text string
}

// baseline is the recorded per-file state.
type baseline struct {
	entries map[string]baselineEntry
	// total is the recorded grand total, kept so a human diff sees the number.
	total int
}

func baselineKey(path, text string) string {
	return path + "\t" + text
}

// loadBaseline reads the recorded baseline.
//
// A MISSING baseline is an error (exit 2), not a pass and not a re-baseline:
// silently treating "no baseline" as "no annotations" would let the guard be
// introduced against an empty record and pass everything on its first run.
//
// found is the number of unreadable annotations in the tree being graded. It is
// needed for one cross-check that cannot be made from the file alone: a record
// holding ZERO entries is a coherent record for a tree with zero annotations and
// a contradiction for a tree that has some. A contradiction is reported as a
// missing record (exit 2) rather than allowed to fall through to the ordinary
// grading path, because that path would call every one of those annotations NEW
// debt and emit hundreds of violations against files nobody touched -- accurate,
// but a cascade that reads as a broken guard and gets "fixed" by disabling it.
// The operational accident being caught is the emptied one: a truncated file, a
// merge conflict, or a checkout that dropped the record.
func loadBaseline(path string, found int) (*baseline, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read baseline %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	b := &baseline{entries: map[string]baselineEntry{}}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		raw := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(raw), "#") || strings.TrimSpace(raw) == "" {
			continue
		}
		// path<TAB>count<TAB>text
		parts := strings.SplitN(raw, "\t", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("baseline %s:%d: want path<TAB>count<TAB>text, got %q", path, line, raw)
		}
		n, cerr := strconv.Atoi(parts[1])
		if cerr != nil || n <= 0 {
			return nil, fmt.Errorf("baseline %s:%d: count %q is not a positive integer", path, line, parts[1])
		}
		e := baselineEntry{path: parts[0], count: n, text: parts[2]}
		b.entries[baselineKey(e.path, e.text)] = e
		b.total += n
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read baseline %s: %w", path, err)
	}
	if len(b.entries) == 0 && b.total != 0 {
		return nil, fmt.Errorf("baseline %s contains no entries but records %d annotation(s); the record is malformed", path, b.total)
	}
	if len(b.entries) == 0 && found > 0 {
		return nil, fmt.Errorf("baseline %s contains no entries, but the tree has %d unreadable annotation(s): "+
			"the record is empty, not a record of zero", path, found)
	}
	return b, nil
}

// writeBaseline records the current findings as the new baseline. Never run in
// CI -- it is the deliberate, reviewable act of re-baselining, and it must show
// up in a diff.
func writeBaseline(path string, findings []finding) error {
	type agg struct {
		count int
		order []int
	}
	byKey := map[string]*agg{}
	var keys []string
	for _, f := range findings {
		k := baselineKey(f.Path, f.Annotation)
		if _, ok := byKey[k]; !ok {
			byKey[k] = &agg{}
			keys = append(keys, k)
		}
		byKey[k].count++
	}
	sort.Strings(keys)

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Unreadable gosec suppression annotations, recorded %s (reviewed %s, expires %s).\n",
		ratchetReviewed, ratchetReviewed, ratchetExpires)
	fmt.Fprintf(&sb, "# %d annotation(s). Tracking: %s\n", len(findings), budgetIssue)
	sb.WriteString("# Regenerate with: go run ./scripts/gosecsuppressions -root . -write-baseline\n")
	sb.WriteString("#\n")
	sb.WriteString("# Format: path<TAB>count<TAB>comment-text-as-written. The key is path+text and\n")
	sb.WriteString("# deliberately NOT path+line: inserting a line above shifts every line number in\n")
	sb.WriteString("# the file, so a line-keyed baseline would fail on ordinary edits.\n")
	sb.WriteString("# Falling entries are improvements -- drop them from this file when you re-record.\n")
	total := 0
	for _, k := range keys {
		path, text, _ := strings.Cut(k, "\t")
		fmt.Fprintf(&sb, "%s\t%d\t%s\n", path, byKey[k].count, text)
		total += byKey[k].count
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create baseline dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o600); err != nil {
		return fmt.Errorf("write baseline %s: %w", path, err)
	}
	fmt.Fprintf(os.Stderr, "recorded %d annotation(s) across %d key(s) in %s (tracking %s)\n",
		total, len(keys), path, budgetIssue)
	return nil
}

// violation is one way the current tree exceeds the recorded baseline.
type violation struct {
	path string
	line int
	text string
	// reason names the rule, so the failure message is not just a count.
	reason string
}

// compare grades the current findings against the baseline.
//
// Three rules, no fourth (see main.go):
//   - a path+text key absent from the baseline FAILS -- this is the new-file
//     case and the only rule that catches it;
//   - a count above the recorded count for a known key FAILS;
//   - a key or count below the recorded value is an IMPROVEMENT and is returned
//     separately so it can be reported without failing.
func compare(findings []finding, b *baseline) (violations []violation, improvements []string) {
	cur := map[string][]finding{}
	for _, f := range findings {
		k := baselineKey(f.Path, f.Annotation)
		cur[k] = append(cur[k], f)
	}

	keys := make([]string, 0, len(cur))
	for k := range cur {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		got := cur[k]
		path, text, _ := strings.Cut(k, "\t")
		rec, known := b.entries[k]
		switch {
		case !known:
			for _, f := range got {
				violations = append(violations, violation{f.Path, f.Line, f.Annotation,
					"this annotation is not in the recorded baseline -- it is NEW debt"})
			}
		case len(got) > rec.count:
			extra := len(got) - rec.count
			first := got[extra]
			violations = append(violations, violation{path, first.Line, text,
				fmt.Sprintf("count rose to %d from the recorded %d (+%d)", len(got), rec.count, extra)})
		case len(got) < rec.count:
			improvements = append(improvements,
				fmt.Sprintf("%s: %q %d -> %d", path, text, rec.count, len(got)))
		}
	}

	for _, k := range sortedKeys(b.entries) {
		if _, still := cur[k]; !still {
			improvements = append(improvements,
				fmt.Sprintf("%s: %q %d -> 0 (annotation removed)", b.entries[k].path, b.entries[k].text, b.entries[k].count))
		}
	}
	return violations, improvements
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
