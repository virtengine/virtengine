package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestTextDropsNolintDirective pins the actual mechanism of the PR #1254 trap.
//
// The reason a single-line `//nolint:gosec // #nosec G115 -- why` is invisible
// to the gate is NOT only that "#nosec is not first". It is that
// ast.CommentGroup.Text() -- which is the very thing gosec's
// findNoSecDirective reads -- classifies `//nolint:gosec` as a compiler-style
// DIRECTIVE (isDirective in go/ast) and DROPS the entire comment, so Text()
// returns "" and no #nosec anywhere in that comment is ever examined.
func TestTextDropsNolintDirective(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "bare nolint directive alone -> dropped entirely",
			src:  "x := f() //nolint:gosec",
			want: "",
		},
		{
			name: "nolint directive with trailing prose -> whole comment dropped",
			src:  "x := f() //nolint:gosec // #nosec G115 -- why",
			want: "",
		},
		{
			name: "block nosec then nolint -> the block comment survives",
			src:  "x := f() /* #nosec G115 -- why */ //nolint:gosec",
			want: " #nosec G115 -- why\n",
		},
		{
			name: "nosec first in its own comment survives",
			src:  "x := f() // #nosec G115 -- why",
			want: "#nosec G115 -- why\n",
		},
		{
			name: "own-line nolint is dropped, own-line nosec after it survives",
			src:  "//nolint:gosec\n\t// #nosec G115 -- why\n\tx := f()",
			want: "#nosec G115 -- why\n",
		},
		{
			name: "nolint without a colon is NOT a directive, it survives as prose",
			src:  "x := f() // nolint gosec",
			want: "nolint gosec\n",
		},
		{
			name: "errcheck nolint is also dropped",
			src:  "x := f() //nolint:errcheck // hash writes never fail",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "f.go", "package p\n\nfunc f() {\n"+tc.src+"\n}\n", parser.ParseComments)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			// Every comment in the group is joined, so walk them all.
			var joined []string
			for _, g := range f.Comments {
				joined = append(joined, g.Text())
			}
			got := strings.Join(joined, "")
			if got != tc.want {
				t.Fatalf("CommentGroup.Text() = %q, want %q", got, tc.want)
			}
			// A directive that starts with a colon-pattern word is dropped, which
			// is the premise the guard's whole failure mode rests on.
			if tc.name == "nosec first in its own comment survives" && !strings.Contains(got, "#nosec") {
				t.Fatalf("premise broken: honoured shape lost its #nosec")
			}
		})
	}
}
