// Command gosecsuppressions rejects gosec suppression directives that the
// security gate cannot read.
//
// Why this exists
// ---------------
// PR #1254 added x/veid/types/assurance_vector.go with its gosec suppressions
// written only as `//nolint:gosec`. Standalone gosec v2.25.0 -- the scanner the
// `gosec Security Scan` gate runs -- does not read `//nolint` at all. It reads
// `#nosec`. So the annotation silenced the linter the author was running and
// did nothing to the gate, and all 7 findings landed in the gate. Nothing in
// the lint path or in the PR checks caught it; it was caught by reading a CI
// log after the fact.
//
// The trap is ASYMMETRIC, which is why "just run gosec" is not the fix. Measured
// on the annotation shapes (see main_test.go, TestShapeMatrix -- re-measured
// against gosec and golangci-lint 2.13.2, not inferred):
//
//	shape                                    standalone gosec (GATE)   golangci-lint gosec
//	//nolint:gosec // why                    REPORTS                   suppressed
//	// #nosec G404 -- why                     suppressed                suppressed
//	/* #nosec G404 -- why */ //nolint:gosec   suppressed                suppressed
//	//nolint:gosec // #nosec G404 -- why      REPORTS                   suppressed
//	// nolint gosec                           REPORTS                   suppressed
//	(no annotation)                           REPORTS                   REPORTS
//
// The last row is a CONTROL. Both tools must report it; if either reported
// nothing at all, every "suppressed" above would be an empty result set rather
// than a verdict.
//
// So a `//nolint:gosec`-only annotation is not merely a no-op, it is a
// suppression that reports itself as suppressed to the developer's editor and
// to golangci-lint while the security gate still sees the finding. That is
// strictly worse than writing nothing, and it is what this command refuses.
//
// What this does NOT do
// ---------------------
//   - It does not run gosec, and it does not decide whether a finding is real.
//     A justified `#nosec` stays justified; this only rejects the *form* that
//     cannot work.
//   - It does not reject `#nosec`. That is the annotation the gate honours, and
//     suppressing a reviewed finding with a written justification is the policy
//     in _docs/security/gosec-triage.md section 5.
//   - It does not reject a `//nolint:gosec` that is ALSO accompanied by a
//     `#nosec` the gate can read, because that combination genuinely suppresses
//     both tools.
//
// The ratchet, and why a bare count is not enough
// ------------------------------------------------
// This tree already holds 370 unreadable annotations (BASELINE_COUNT), so a
// hard fail would red every gate on the day it lands and teach everyone to
// ignore it. But a plain "total must not rise" count does NOT satisfy the card
// either: shipping one new //nolint:gosec-only file while deleting one old
// unreadable annotation leaves the total at 370 and the guard silent, which is
// exactly the trade this card exists to stop.
//
// So the baseline is KEYED, not just counted. Each entry is
//
//	path<TAB>count<TAB>annotation
//
// and is stable under edits ABOVE it: inserting a line shifts every line number
// in the file, so a line-numbered baseline would detonate into hundreds of false
// failures on ordinary edits and be turned off within a week. The comment text
// plus its path is the stable identity; the count disambiguates the same comment
// repeated in one file.
//
// Monotone by construction -- three rules, no fourth:
//   - an annotation path+text NOT in the baseline FAILS (this is the new-file
//     case, and the only rule that catches it);
//   - a count rising for a known key FAILS;
//   - keys or counts falling is an IMPROVEMENT: reported, never failed, and the
//     baseline file is expected to be re-recorded with -write-baseline.
//
// Raising the ceiling therefore requires editing this file, which puts the new
// number in a reviewable diff next to RATCHET_EXPIRES and BUDGET_ISSUE.
// MAX_ALLOWANCE, like every override in this repo's gates, is one-sided: it can
// only tighten.
//
// Usage:
//
//	go run ./scripts/gosecsuppressions -root .
//	go run ./scripts/gosecsuppressions -root . -json
//	go run ./scripts/gosecsuppressions -root . -write-baseline   # record, never in CI
//
// Exit codes:
//
//	0  every NEW gosec-suppressing annotation is one the gate can read
//	1  a new/raised annotation, a missing baseline, or an expired review window
//	2  could not walk the tree / a file did not parse / baseline unreadable
//	   -- deliberately NOT a pass, so an unreadable tree can never report clean
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The three constants below and findNoSecDirective/findNoSecTag are copied
// VERBATIM from github.com/securego/gosec/v2@v2.25.0/analyzer.go (lines 66-70
// and 803-855). This command exists to grade annotations by the rule the gate
// actually applies, so a second, approximate copy of that rule would be the
// defect it is meant to prevent. TestNolintDirectiveIsStrippedFromGateText and
// TestNosecOnlyCommentSurvives pin the behaviour of these functions; if a gosec
// bump changes the rule, those tests fail here instead of the guard quietly
// grading something else.
const directivePrefix = "//gosec:disable"

// analyzer.go:953-964 -- with no .gosec config these are the defaults.
const (
	noSecDefaultTag     = "#nosec"
	noSecAlternativeTag = "#nosec"
)

// findNoSecDirective, verbatim from gosec v2.25.0 analyzer.go:805-830.
func findNoSecDirective(group *ast.CommentGroup, noSecDefaultTag, noSecAlternativeTag string) (bool, string) {
	if group == nil {
		return false, ""
	}

	// Join all comments in the group once to support multi-line nosec tags
	text := group.Text()

	// Check for nosec tags
	for _, tag := range []string{noSecDefaultTag, noSecAlternativeTag} {
		if found, args := findNoSecTag(text, tag); found {
			return true, args
		}
	}

	// Check for directive comments individually
	for _, c := range group.List {
		if after, ok := strings.CutPrefix(c.Text, directivePrefix); ok {
			if len(after) == 0 || after[0] == ' ' {
				return true, strings.TrimSpace(after)
			}
		}
	}

	return false, ""
}

// findNoSecTag, verbatim from gosec v2.25.0 analyzer.go:832-855.
func findNoSecTag(text, tag string) (bool, string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return false, ""
	}

	if strings.HasPrefix(text, tag) {
		return true, text[len(tag):]
	}

	if idx := strings.Index(text, tag); idx > 0 {
		// Check if it's at the beginning of a line (possibly with space)
		for i := idx - 1; i >= 0; i-- {
			if text[i] == '\n' {
				return true, text[idx+len(tag):]
			}
			if text[i] != ' ' && text[i] != '\t' {
				break
			}
		}
	}

	return false, ""
}

// finding is one annotation that the gate cannot read.
type finding struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	// Annotation is the comment text as written.
	Annotation string `json:"annotation"`
	// NolintLinters is the linter list the nolint directive named, for the
	// message. Empty means a bare //nolint, which golangci-lint reads as
	// "every linter".
	NolintLinters string `json:"nolint_linters,omitempty"`
}

// nolintDirectivesIn reports the linter names named by `//nolint` directives in
// this comment group, per golangci-lint's parser: the directive is the text
// after `//`, left-trimmed, beginning `nolint`, optionally followed by
// `:<list>` or `:<all>` or ` all`. A bare `//nolint` suppresses every linter,
// so it is reported as ["*"].
//
// It reads group.List rather than group.Text() on purpose, and it must keep
// doing so. See directiveDrop: a comment beginning `//nolint:` is classified as
// a Go *directive* and stripped from group.Text() entirely, so gosec never sees
// it -- and neither would this function. It is the one thing that must be read
// from the raw comment.
//
// It returns false when the group carries no nolint directive at all.
func nolintDirectivesIn(group *ast.CommentGroup) ([]string, bool) {
	if group == nil {
		return nil, false
	}
	var names []string
	for _, c := range group.List {
		body := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(c.Text), "//"))
		body = strings.TrimPrefix(body, "/*")
		body = strings.TrimSuffix(body, "*/")
		body = strings.TrimSpace(body)
		// The word boundary is REQUIRED, not cosmetic. golangci-lint matches
		//
		//	^//\s*nolint(?::\s*(all|\w+(?:,\s*\w+)*))?\b
		//
		// and a plain HasPrefix("nolint") matched this package's OWN doc comment
		// `// nolintDirectivesIn reports the linter names named by ...` -- so the
		// guard flagged itself as unreadable debt and its baseline grew an entry
		// for main.go. Measured against golangci-lint v2.13.2 (see
		// TestNolintWordBoundaryMatchesGolangci): `//nolintGosec`,
		// `//nolintDirectivesIn`, `//nolint2` and `//nolint-` are NOT directives
		// and leave the finding reported, so treating them as directives makes
		// the guard demand a fix for prose.
		list, hasList, ok := cutNolint(body)
		if !ok {
			continue
		}
		if !hasList {
			// golangci-lint matches `//\s*nolint` optionally followed by
			// `:list`. With no colon the directive covers EVERY linter and the
			// remainder is the author's justification -- which is why
			// `//nolint // G115: bounded` and the spaced form
			// `// nolint gosec` both suppress gosec too. golangci-lint's own
			// expression for this is
			//   ^//\s*nolint(?::\s*(all|\w+(?:,\s*\w+)*))?\b
			names = append(names, "*")
			continue
		}
		if list == "" || list == "all" {
			names = append(names, "*")
			continue
		}
		for _, n := range strings.Split(list, ",") {
			// The directive ENDS at the first whitespace: in
			// `//nolint:gosec // G115: bounded`, "gosec" is the linter and the
			// rest is the author's justification. Splitting on the comma alone
			// yields the single token "gosec // G115: bounded", which matches
			// nothing and made the guard miss shapes A and D entirely.
			n, _, _ = strings.Cut(strings.TrimSpace(n), " ")
			if n = strings.ToLower(strings.TrimSpace(n)); n != "" {
				names = append(names, n)
			}
		}
	}
	return names, len(names) > 0
}

func namesGosec(names []string) bool {
	for _, n := range names {
		if n == "*" || n == "gosec" {
			return true
		}
	}
	return false
}

// cutNolint matches golangci-lint's directive grammar at the START of body.
//
// The rule, matching the expression golangci-lint documents:
//
//	^//\s*nolint(?::\s*(all|\w+(?:,\s*\w+)*))?\b
//
// so after `nolint` the character must be a WORD BOUNDARY: end of text, a
// colon, or whitespace. `nolintGosec`, `nolintDirectivesIn`, `nolint2` and
// `nolint-` do not qualify, and each was measured to leave a gosec finding
// REPORTED by golangci-lint v2.13.2 -- i.e. they are prose, not suppressions.
//
// `list` is the linter list and `hasList` says whether a colon introduced it.
// The colon is load-bearing: `//nolint // G115: bounded` has NO colon, so it
// covers every linter, while `//nolint:gosec // G115: bounded` has one and the
// text after it is the author's justification, not a second linter. Collapsing
// the two made the bare form miss entirely.
//
// ok == false means "this comment is not a directive at all"; the caller then
// ignores it, because a prose comment naming a function after "nolint" is not a
// suppression the gate is being tricked by.
func cutNolint(body string) (list string, hasList bool, ok bool) {
	if !strings.HasPrefix(body, "nolint") {
		return "", false, false
	}
	after := body[len("nolint"):]
	if after != "" && after[0] != ':' && !isSpaceByte(after[0]) {
		return "", false, false
	}
	if !strings.HasPrefix(after, ":") {
		// No colon: the directive covers every linter and `after` is the
		// author's justification.
		return "", false, true
	}
	return strings.TrimSpace(after[1:]), true, true
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '	'
}

// annotationText renders a comment group the way the error message shows it.
func annotationText(group *ast.CommentGroup) string {
	raw := make([]string, 0, len(group.List))
	for _, c := range group.List {
		raw = append(raw, c.Text)
	}
	s := strings.Join(raw, " ")
	if len(s) > 160 {
		s = s[:157] + "..."
	}
	return s
}

// gateReadsSuppression answers the only question that matters: given the
// comment group attached to a node, does gosec's own findNoSecDirective see a
// suppression it will honour?
//
// It deliberately calls gosec's function on the group's Text(), the same
// expression gosec uses (analyzer.go:811). Text() strips any comment beginning
// `//nolint:`, so a `//nolint:gosec // #nosec G115 -- why` arrives here as an
// empty string and is correctly graded "not honoured". That asymmetry is the
// bug this whole command exists to catch, so grading must not paper over it.
func gateReadsSuppression(group *ast.CommentGroup) bool {
	honoured, _ := findNoSecDirective(group, noSecDefaultTag, noSecAlternativeTag)
	return honoured
}

// checkFile returns every finding in one parsed file.
func checkFile(path string, fset *token.FileSet, f *ast.File) []finding {
	var out []finding

	// gosec resolves a node's suppression from the comment map, so this must
	// too: a `#nosec` on the line above a statement and a `//nolint:gosec`
	// elsewhere in the same function are attached to DIFFERENT nodes, and only
	// the node-attached one suppresses.
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	// gosec keys on node identity, so walk nodes rather than comment groups.
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		groups, ok := cmap[n]
		if !ok {
			return true
		}
		for _, group := range groups {
			names, hasNolint := nolintDirectivesIn(group)
			if !hasNolint || !namesGosec(names) {
				continue
			}
			// The annotation names gosec. Does the gate read it?
			if gateReadsSuppression(group) {
				continue
			}
			out = append(out, finding{
				Path:          path,
				Line:          fset.Position(group.Pos()).Line,
				Annotation:    annotationText(group),
				NolintLinters: strings.Join(names, ","),
			})
		}
		return true
	})
	return out
}

// skipDirs are never scanned. The set MIRRORS the gate's own gosec invocation
// (.github/workflows/security.yaml, `gosec-scan`):
//
//	gosec -fmt json -exclude-generated -exclude-dir=vendor -exclude-dir=testutil ./...
//
// so that this guard grades exactly the population the gate scores. `vendor` is
// gosec's -exclude-dir=vendor, `testutil` is its -exclude-dir=testutil, and
// `.git`/`.worktrees`/`node_modules` are not Go source the gate can reach.
//
// The `testutil` entry is not cosmetic. Before it was added this guard reported 407
// findings against a tree where only 370 are in gate scope: 37 of them are under
// testutil/, which the gate never opens. A baseline measured over a different
// population than the gate enforces is not a tighter gate, it is a wrong one -- and
// it would make the honest work of paying the debt look larger than it is.
//
// `node_modules` and `.worktrees` are here because this guard walks the
// FILESYSTEM. gosec does too (measured: a probe committed under `testdata/`,
// `_probe/` and `.probe/` were all scanned by `gosec ./...` while `go list ./...`
// reported a single package), so a Go file anywhere under the repo root is a file
// the gate may analyse.
func skipDirs() map[string]bool {
	return map[string]bool{
		"vendor":       true,
		"testutil":     true,
		"node_modules": true,
		".git":         true,
		".worktrees":   true,
	}
}

func main() {
	root := flag.String("root", ".", "repository root to scan")
	asJSON := flag.Bool("json", false, "emit JSON")
	baselinePath := flag.String("baseline", "scripts/gosecsuppressions/gosec-suppressions.baseline",
		"recorded baseline to grade against")
	writeBaselineFlag := flag.Bool("write-baseline", false,
		"record the current findings as the new baseline (deliberate act, never in CI)")
	flag.Parse()

	var findings []finding
	var walkErr error

	err := filepath.WalkDir(*root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs()[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			// A file gosec cannot parse cannot be suppressed either, but a
			// parse error here means THIS COMMAND is looking at a tree it does
			// not understand. Report it as an error, never as a clean estate.
			return fmt.Errorf("parse %s: %w", path, perr)
		}
		rel, rerr := filepath.Rel(*root, path)
		if rerr != nil {
			rel = path
		}
		findings = append(findings, checkFile(filepath.ToSlash(rel), fset, f)...)
		return nil
	})
	if err != nil {
		walkErr = err
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Line < findings[j].Line
	})

	if walkErr != nil {
		fmt.Fprintf(os.Stderr, "::error::could not scan the tree: %v\n", walkErr)
		fmt.Fprintf(os.Stderr, "::error::this is NOT a pass; refusing to report a clean estate from an unreadable tree\n")
		os.Exit(2)
	}

	// Resolve the baseline path ONCE, before either mode uses it. It is relative
	// by default, and it must resolve against -root rather than the caller's cwd
	// or a CI step that cds per module would write or read a file in the wrong
	// place. Doing this after the -write-baseline branch made that mode write to
	// ./recorded.baseline in whatever directory the caller happened to be in,
	// while the grade step then looked for <root>/recorded.baseline and failed --
	// so the two modes disagreed about where the record lives.
	basePath := *baselinePath
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(*root, basePath)
	}

	if *writeBaselineFlag {
		// Re-baselining is a deliberate act that must show up in a reviewable
		// diff, so it is a separate mode rather than a fallback when grading
		// fails. A CI run that could silently re-record the baseline would make
		// every gate above it meaningless.
		if err := writeBaseline(basePath, findings); err != nil {
			fmt.Fprintf(os.Stderr, "::error::%v\n", err)
			os.Exit(2)
		}
		os.Exit(0)
	}

	b, err := loadBaseline(basePath, len(findings))
	if err != nil {
		fmt.Fprintf(os.Stderr, "::error::%v\n", err)
		fmt.Fprintf(os.Stderr, "::error::this is NOT a pass. A missing or unreadable baseline means the guard "+
			"cannot tell new debt from recorded debt; grading an estate against no record would pass everything.\n")
		fmt.Fprintf(os.Stderr, "::error::Record one deliberately with: go run ./scripts/gosecsuppressions "+
			"-root %s -write-baseline\n", *root)
		os.Exit(2)
	}

	violations, improvements := compare(findings, b)
	allowance := resolveAllowance()

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{
			"total":          len(findings),
			"baseline":       b.total,
			"violations":     violations,
			"improvements":   improvements,
			"allowance":      resolveAllowance(),
			"ratchet_expiry": ratchetExpires,
			"tracking":       budgetIssue,
		})
	} else {
		fmt.Printf("unreadable gosec annotations: %d now, %d recorded (tracking %s)\n",
			len(findings), b.total, budgetIssue)
		for _, imp := range improvements {
			fmt.Printf("improved: %s\n", imp)
		}
		if len(improvements) > 0 {
			fmt.Printf("note: re-record the baseline with -write-baseline so improvements stick\n")
		}
		for _, v := range violations {
			fmt.Printf("%s:%d: %s -- %s\n", v.path, v.line, v.text, v.reason)
		}
	}

	if len(violations) > allowance {
		fmt.Fprintf(os.Stderr, "\n::error::%d unreadable gosec suppression annotation(s) beyond the recorded baseline (allowance %d)\n",
			len(violations)-allowance, allowance)
		fmt.Fprintf(os.Stderr, "::error::Standalone gosec v2.25.0 (the `gosec Security Scan` gate) reads only #nosec. A //nolint:gosec comment is read by golangci-lint and ignored by the gate, so it hides the finding from your editor while the gate still reports it.\n")
		fmt.Fprintf(os.Stderr, "::error::Fix: put #nosec FIRST in the comment group, e.g. `// #nosec G115 -- <why this is safe>`; if you also want golangci-lint quiet, add //nolint:gosec AFTER it. See _docs/security/gosec-triage.md section 9.5.\n")
		if reviewExpired(time.Now().Format("2006-01-02")) {
			fmt.Fprintf(os.Stderr, "::error::NOTE: the ratchet review window closed on %s (reviewed %s). Re-measure and re-record regardless of the findings above. Tracking %s.\n",
				ratchetExpires, ratchetReviewed, budgetIssue)
		}
		os.Exit(1)
	}

	if reviewExpired(time.Now().Format("2006-01-02")) {
		fmt.Fprintf(os.Stderr, "::error::The gosec-suppression budget review window closed on %s (reviewed %s). "+
			"Re-measure with -write-baseline, then update ratchetReviewed/ratchetExpires in "+
			"scripts/gosecsuppressions/baseline.go and commit the result (tracking %s).\n",
			ratchetExpires, ratchetReviewed, budgetIssue)
		os.Exit(1)
	}

	if !*asJSON {
		fmt.Printf("ok: no gosec suppression annotation beyond the recorded baseline (ratchet expires %s, tracking %s)\n",
			ratchetExpires, budgetIssue)
	}
}
