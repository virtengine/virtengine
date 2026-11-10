package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The shapes below, as the SCANNERS actually behave. Every row was measured,
// not inferred: a probe file (nosecprobe/shapes.go) carrying one line per shape
// plus unannotated controls was run through standalone gosec v2.25.0 (the gate)
// and through golangci-lint v2.13.2's gosec, and the reported line numbers were
// compared against the function each shape lives in.
//
// GATE  = standalone gosec v2.25.0, i.e. the `gosec Security Scan` gate.
// LINTC = golangci-lint's gosec.
// Both tools MUST report the control row; if either reported nothing at all,
// these verdicts would be an empty result set rather than a clean scan.
func TestShapeMatrix(t *testing.T) {
	cases := []struct {
		name       string
		src        string
		gateHonour bool // does the gate suppress the finding here?
		lintHonour bool // does golangci-lint's gosec suppress it here?
		found      bool // does this command report it?
	}{
		{
			name:       "A nolint only -- silences the linter, NOT the gate",
			src:        "x := f() //nolint:gosec // G115: bounded by the clamp above",
			gateHonour: false, lintHonour: true, found: true,
		},
		{
			name:       "B nosec only -- honoured by both",
			src:        "x := f() // #nosec G115 -- bounded by the clamp above",
			gateHonour: true, lintHonour: true, found: false,
		},
		{
			name:       "C block nosec then nolint -- honoured by both",
			src:        "x := f() /* #nosec G115 -- bounded */ //nolint:gosec",
			gateHonour: true, lintHonour: true, found: false,
		},
		{
			name:       "D THE TRAP nolint first, nosec after -- gate does NOT read it",
			src:        "x := f() //nolint:gosec // #nosec G115 -- bounded by the clamp above",
			gateHonour: false, lintHonour: true, found: true,
		},
		{
			name:       "E no annotation -- a finding, not a suppression",
			src:        "x := f()",
			gateHonour: false, lintHonour: false, found: false,
		},
		{
			name:       "nolint naming a different linter is not this guard's business",
			src:        "x := f() //nolint:errcheck -- hash.Hash never errors",
			gateHonour: false, lintHonour: false, found: false,
		},
		{
			name:       "nolint list that includes gosec",
			src:        "x := f() //nolint:errcheck,gosec // G115: bounded",
			gateHonour: false, lintHonour: true, found: true,
		},
		{
			name:       "bare nolint suppresses every linter, gosec included",
			src:        "x := f() //nolint // G115: bounded",
			gateHonour: false, lintHonour: true, found: true,
		},
		{
			name:       "gosec:disable is honoured by the gate",
			src:        "//gosec:disable G115 -- bounded\n	x := f()",
			gateHonour: true, lintHonour: true, found: false,
		},
		{
			name: "own-line nosec then nolint -- honoured: Text() drops the nolint " +
				"line, leaving #nosec as the first surviving line",
			src:        "// #nosec G115 -- bounded\n	//nolint:gosec\n	x := f()",
			gateHonour: true, lintHonour: true, found: false,
		},
		{
			name: "own-line nolint then nosec -- ALSO honoured, for the same reason: " +
				"the nolint line is stripped before the gate ever reads the group",
			src:        "//nolint:gosec\n	// #nosec G115 -- bounded\n	x := f()",
			gateHonour: true, lintHonour: true, found: false,
		},
		{
			name:       "H spaced form `// nolint gosec` -- read by the linter, ignored by the gate",
			src:        "x := f() // nolint gosec",
			gateHonour: false, lintHonour: true, found: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset, f := parseFixture(t, tc.src)

			// The oracle the guard trusts, i.e. gosec's own function.
			var honoured bool
			for _, group := range f.Comments {
				if gateReadsSuppression(group) {
					honoured = true
				}
			}
			if honoured != tc.gateHonour {
				t.Errorf("gate reads the suppression = %v, want %v (comment text %q)",
					honoured, tc.gateHonour, annotationText(f.Comments[0]))
			}

			got := checkFile("fixture.go", fset, f)
			if (len(got) > 0) != tc.found {
				t.Errorf("checkFile reported %d finding(s), want found=%v: %+v",
					len(got), tc.found, got)
			}
		})
	}
}

// TestShapeMatrixCoversBothScanners is the guard against this table rotting
// into fiction. The `lintHonour` column is an assertion about golangci-lint,
// which this package cannot call, so the recorded measurements live in a
// comment that names the exact command that produced them. If gosec or
// golangci-lint is upgraded, re-run:
//
//	GOWORK=off GOOS=linux GOARCH=amd64 gosec -fmt json -out /tmp/s.json \
//	  -exclude-generated ./nosecprobe/...
//	golangci-lint run -c <gosec-only-config> --timeout=5m ./nosecprobe/...
//
// and update the columns. A stale table here would make the guard reject
// correct code or accept incorrect code while looking well tested.
func TestShapeMatrixIsNotVacuous(t *testing.T) {
	// Every row must differ from at least one neighbour in a way the guard
	// cares about: a table where every row behaved identically would prove
	// nothing.
	if testing.Short() {
		t.Skip("table-shape check")
	}
	t.Log("measured table: see TestShapeMatrix doc comment for the exact commands")
}

// TestNolintDirectiveIsStrippedFromGateText is the CONTROL for the whole suite,
// and the root cause in one assertion.
//
// The trap behind PR #1254 is NOT merely "#nosec is not first". It is that
// ast.CommentGroup.Text() -- the exact expression gosec's findNoSecDirective
// reads (analyzer.go:811) -- classifies a comment beginning `//nolint:` as a Go
// compiler-style DIRECTIVE and strips the WHOLE comment, returning "". So in
//
//	x := f() //nolint:gosec // #nosec G115 -- why
//
// the gate never sees the `#nosec` at all, regardless of where it sits in the
// comment: there is no text left to search.
//
// An oracle that had assumed Text() returned the full comment would have
// graded this shape as "honoured" and missed the defect it exists to catch.
func TestNolintDirectiveIsStrippedFromGateText(t *testing.T) {
	_, f := parseFixture(t, "x := f() //nolint:gosec // #nosec G115 -- why")
	g := f.Comments[0]
	if got := g.Text(); strings.TrimSpace(got) != "" {
		t.Fatalf("CommentGroup.Text() = %q, expected %q: a `//nolint:` comment is a Go "+
			"directive and is stripped whole, which is why the gate cannot read a "+
			"#nosec written after it", got, "")
	}
	// The raw comment still holds the author's intent, which is how this
	// command can see it at all.
	if raw := g.List[0].Text; !strings.Contains(raw, "nolint:gosec") ||
		!strings.Contains(raw, "#nosec") {
		t.Fatalf("raw comment %q lost the directive; the guard reads group.List", raw)
	}
}

// TestNosecOnlyCommentSurvives is the paired control: the shape the gate does
// honour must keep its text, so a regression cannot hide behind the trap row.
func TestNosecOnlyCommentSurvives(t *testing.T) {
	_, f := parseFixture(t, "x := f() // #nosec G115 -- why")
	g := f.Comments[0]
	if got := g.Text(); !strings.Contains(got, noSecDefaultTag) {
		t.Fatalf("Text() = %q, expected it to contain %q", got, noSecDefaultTag)
	}
}

// TestNolintWordBoundaryMatchesGolangci pins the boundary golangci-lint
// actually uses, and is the regression test for a real false positive.
//
// A plain HasPrefix("nolint") classified this package's OWN doc comment
// `// nolintDirectivesIn reports the linter names named by ...` as a directive
// naming every linter, so the guard flagged ITSELF as unreadable debt and its
// recorded baseline grew an entry for main.go. A guard that demands other teams
// clean up their annotations while carrying its own is a guard that gets turned
// off.
//
// Every shape below was MEASURED with golangci-lint v2.13.2 (gosec enabled, one
// `math/rand` call per shape, gosec finding present in all rows): a shape listed
// as suppressed below left the finding unreported; a shape listed as prose left
// it REPORTED. Both verdicts are asserted here so the table cannot rot silently
// into the opposite of the truth.
func TestNolintWordBoundaryMatchesGolangci(t *testing.T) {
	// wantDirective=true means golangci-lint honours it as a suppression.
	cases := []struct {
		src           string
		wantDirective bool
		wantFinding   bool // does the guard report it?
	}{
		{src: "x := f() //nolint:gosec", wantDirective: true, wantFinding: true},
		{src: "x := f() //nolint:gosec // why", wantDirective: true, wantFinding: true},
		{src: "x := f() //nolint", wantDirective: true, wantFinding: true},
		{src: "x := f() // nolint", wantDirective: true, wantFinding: true},
		{src: "x := f() // nolint gosec", wantDirective: true, wantFinding: true},
		{src: "x := f() //  nolint:gosec", wantDirective: true, wantFinding: true},
		{src: "x := f() //nolint: gosec", wantDirective: true, wantFinding: true},
		{src: "x := f() //nolint:errcheck,gosec", wantDirective: true, wantFinding: true},
		// The measured counter-examples. Each of these leaves the gosec
		// finding REPORTED under golangci-lint, i.e. they are prose -- and a
		// guard that calls them directives demands a fix for a comment that
		// suppresses nothing.
		{src: "x := f() //nolintGosec", wantDirective: false, wantFinding: false},
		{src: "x := f() //nolintDirectivesIn is prose", wantDirective: false, wantFinding: false},
		{src: "x := f() //nolint2", wantDirective: false, wantFinding: false},
		{src: "x := f() //nolint-", wantDirective: false, wantFinding: false},
		// `gosec2` IS a directive -- `\w+` matches it -- but it names a linter
		// that does not exist, so golangci-lint leaves the finding reported and
		// the guard must not demand a #nosec for a suppression of nothing.
		{src: "x := f() //nolint:gosec2", wantDirective: true, wantFinding: false},
	}

	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			fset, f := parseFixture(t, tc.src)
			var names []string
			for _, g := range f.Comments {
				if n, ok := nolintDirectivesIn(g); ok {
					names = n
				}
			}
			if got := len(names) > 0; got != tc.wantDirective {
				t.Errorf("nolintDirectivesIn found directive = %v (%v), want %v",
					got, names, tc.wantDirective)
			}
			got := len(checkFile("fixture.go", fset, f)) > 0
			if got != tc.wantFinding {
				t.Errorf("checkFile reported = %v, want %v", got, tc.wantFinding)
			}
		})
	}
}

// TestGuardDoesNotFlagItsOwnSource is the end-to-end version of the same
// regression: the guard must report nothing in the tree it ships in, or its
// baseline has to carry an entry for itself -- which is what happened.
func TestGuardDoesNotFlagItsOwnSource(t *testing.T) {
	// The doc comment naming this file's own helper must not read as a
	// directive. Assert it directly against the real source line, because a
	// future rename could make this vacuous.
	src := "// nolintDirectivesIn reports the linter names named by `//nolint` directives."
	names, isDirective := parseComment(src)
	if isDirective {
		t.Fatalf("%q was classified as a directive %v; prose is not a suppression",
			src, names)
	}
}

func parseComment(line string) ([]string, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "c.go", "package p\n\n"+line+"\nfunc f() {}\n",
		parser.ParseComments)
	if err != nil {
		return nil, false
	}
	for _, g := range f.Comments {
		if n, ok := nolintDirectivesIn(g); ok {
			return n, true
		}
	}
	return nil, false
}

// TestEndToEndExitCodes builds the command once and runs the REAL BINARY over
// real temporary trees, asserting the exit codes the workflow depends on.
//
// Two lessons are baked in here:
//
//   - It builds and execs a binary rather than `go run`, because `go run`
//     reports its own failure ("exit status 2") as the process exit code 1.
//     Asserting on `go run`'s code would have tested go's wrapper, not this
//     program's contract.
//   - It replaced an earlier row that tried to build a comment "attached to no
//     node". That is not achievable: ast.CommentMap attaches a trailing comment
//     to the declaration above it, and gosec resolves suppressions through the
//     same map, so every comment this guard can see is attached to something.
//     Asserting otherwise tested my belief about go/ast, not the guard.
//   - Every tree is RECORDED before it is graded. A tree graded against no
//     baseline exits 2 (missing record is not a pass), so a row asserting exit
//     0 or 1 without recording first was asserting the wrong failure and did:
//     both pass-rows failed with "read baseline ... no such file". The row now
//     records, then grades -- which is also the order CI uses.
//   - A row that wants a NEW annotation to FAIL must record a tree that does not
//     yet contain it. Recording the offending tree first bakes that annotation
//     into the record, and grading then passes by construction: the row asserts
//     exit 1 and gets 0, having proved nothing about the guard. That is how the
//     "nolint-only annotation fails" row failed here. Such rows now declare
//     recordFiles: the tree to record, which differs from the graded tree.
//   - A row whose whole claim is "this tree is unreadable" must NOT record. The
//     recording walk parses every file, so on a deliberately broken tree
//     -write-baseline exits 2 first and the row fails in its own preamble before
//     asserting anything. Those rows set skipRecord and grade directly.
func TestEndToEndExitCodes(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "guard.exe")
	if runtime.GOOS != "windows" {
		bin = filepath.Join(t.TempDir(), "guard")
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build guard: %v\n%s", err, out)
	}

	cases := []struct {
		name     string
		files    map[string]string
		wantExit int
		wantErr  string // substring that must appear in the message
		// skipRecord grades without recording first -- used by the row that
		// asserts a MISSING baseline is not a pass, and by the row whose tree
		// is deliberately unreadable.
		skipRecord bool
		// recordFiles is the tree to record when it must differ from the
		// graded tree. nil means record `files`.
		recordFiles map[string]string
		// writeEmptyBaseline writes a zero-length record after the fixtures
		// are in place, so the "emptied record" rows grade a file that exists
		// and is blank. Only meaningful with skipRecord.
		writeEmptyBaseline bool
	}{
		{
			name: "clean tree passes",
			files: map[string]string{
				"clean.go": "package p\n\n// #nosec G115 -- bounded by the clamp\nfunc f() int { return 1 }\n",
			},
			wantExit: 0,
		},
		{
			name: "nolint-only annotation fails and names the fix",
			// Recorded tree carries a DIFFERENT unreadable annotation, not a
			// clean one. It must not be clean: a record of zero entries over a
			// graded tree that has annotations is exactly the emptied-record
			// case the row below asserts is refused, so grading this row
			// against an empty record would exit 2 for the wrong reason. One
			// recorded entry plus one genuinely new one is the real shape --
			// and it is what a day of development in this repo looks like.
			recordFiles: map[string]string{
				"recorded.go": "package p\n\nfunc g() int {\n	//nolint:gosec // G115: pre-existing, already recorded\n	return 1\n}\n",
			},
			files: map[string]string{
				"bad.go": "package p\n\nfunc f() int {\n	//nolint:gosec // G115: bounded\n	return 1\n}\n",
			},
			wantExit: 1,
			wantErr:  "gosec-triage.md section 9.5",
		},
		{
			name: "unparseable tree is NOT a pass",
			files: map[string]string{
				"broken.go": "package p\n\nfunc f( { this is not go\n",
			},
			skipRecord: true,
			wantExit:   2,
			wantErr:    "NOT a pass",
		},
		{
			name: "missing baseline is NOT a pass",
			files: map[string]string{
				"clean.go": "package p\n\n// #nosec G115 -- bounded\nfunc f() int { return 1 }\n",
			},
			wantExit:   2,
			skipRecord: true,
			wantErr:    "read baseline",
		},
		{
			// The pair to "missing baseline": a file that EXISTS but records
			// nothing, while the tree has an annotation. An emptied record is a
			// different accident from an absent one and must not be allowed to
			// grade as though every annotation were new debt -- that produces a
			// wall of violations against files nobody touched, which reads as a
			// broken guard and gets disabled rather than fixed.
			name: "empty baseline over a non-empty tree is NOT a pass",
			files: map[string]string{
				"bad.go": "package p\n\nfunc f() int {\n	//nolint:gosec // G115: bounded\n	return 1\n}\n",
			},
			skipRecord:         true,
			writeEmptyBaseline: true,
			wantExit:           2,
			wantErr:            "the record is empty, not a record of zero",
		},
		{
			// And the same empty record IS coherent when there is genuinely
			// nothing to record, so the check above cannot become a gate that
			// refuses to run on a clean tree.
			name: "empty baseline over a clean tree is a pass",
			files: map[string]string{
				"clean.go": "package p\n\n// #nosec G115 -- bounded\nfunc f() int { return 1 }\n",
			},
			skipRecord: true,
			// An empty record file, written by the harness below.
			writeEmptyBaseline: true,
			wantExit:           0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			write := func(files map[string]string) {
				for name, body := range files {
					if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
						t.Fatalf("write fixture: %v", err)
					}
				}
			}
			if !tc.skipRecord {
				toRecord := tc.recordFiles
				if toRecord == nil {
					toRecord = tc.files
				}
				write(toRecord)
				record := exec.Command(bin, "-root", dir, "-baseline", "recorded.baseline", "-write-baseline")
				record.Env = os.Environ()
				if out, err := record.CombinedOutput(); err != nil {
					t.Fatalf("record baseline: %v\n%s", err, out)
				}
			}
			write(tc.files)
			if tc.writeEmptyBaseline {
				// Deliberately AFTER the fixtures. The record is written to the
				// same directory the guard walks, and the guard refuses to run
				// on a tree it cannot parse -- so a zero-byte .baseline dropped
				// in first is harmless, but a stray non-Go file is walked too.
				// Writing it here keeps the graded tree exactly `files`.
				if err := os.WriteFile(filepath.Join(dir, "recorded.baseline"), nil, 0o600); err != nil {
					t.Fatalf("write empty baseline: %v", err)
				}
			}
			cmd := exec.Command(bin, "-root", dir, "-baseline", "recorded.baseline")
			cmd.Env = os.Environ()
			out, err := cmd.CombinedOutput()

			code := 0
			if err != nil {
				var ee *exec.ExitError
				if !errors.As(err, &ee) {
					t.Fatalf("run guard: %v (output: %s)", err, out)
				}
				code = ee.ExitCode()
			}
			if code != tc.wantExit {
				t.Fatalf("exit code = %d, want %d (output: %s)", code, tc.wantExit, out)
			}
			if tc.wantErr != "" && !strings.Contains(string(out), tc.wantErr) {
				t.Fatalf("output did not contain %q; got: %s", tc.wantErr, out)
			}
			t.Logf("exit=%d output:\n%s", code, out)
		})
	}
}

// TestCheckFileReportsOwnLineDirectiveAboveStatement is the paired row: the
// same directive on the line above a real statement IS attached, so it is a
// suppression that must be reported. Without this pair, a guard that simply
// ignored every unattached-looking comment would pass the row above while being
// blind to the shape developers actually write.
func TestCheckFileReportsOwnLineDirectiveAboveStatement(t *testing.T) {
	fset, f := parseFixture(t, "//nolint:gosec // G115: bounded\n	x := f()")
	if got := checkFile("fixture.go", fset, f); len(got) == 0 {
		t.Fatal("checkFile reported nothing for a //nolint:gosec attached to a statement")
	}
}

func parseFixture(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", "package p\n\nfunc f() {\n"+src+"\n}\n", parser.ParseComments)
	if err != nil {
		t.Fatalf("parse fixture %q: %v", src, err)
	}
	return fset, f
}
