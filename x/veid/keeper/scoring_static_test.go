package keeper

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestConsensusKeeperPathsDoNotSelectScorersFromEnvironment(t *testing.T) {
	for _, file := range []string{
		"consensus_system_tx.go",
		"inference_receipt.go",
		"verification_pipeline.go",
		"vote_extension.go",
		"vote_extension_consensus.go",
	} {
		bz, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		source := string(bz)
		for _, forbidden := range []string{
			"os.Getenv",
			"DefaultMLScoringConfig",
			"DefaultDevelopmentMLScoringConfig",
			"DefaultTensorFlowScoringConfig",
			"DefaultDevelopmentTensorFlowScoringConfig",
			"NewStubMLScorer",
			"createTensorFlowScorer",
			"inference.NewScorer",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s must not contain production scorer selector %q", file, forbidden)
			}
		}
	}
}

func TestConsensusScoringEntryPointsRemainInjectedOnly(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "scoring.go", nil, 0)
	if err != nil {
		t.Fatalf("parse scoring.go: %v", err)
	}
	for _, funcName := range []string{"ComputeIdentityScore", "getMLScorer"} {
		fn := findFuncDecl(file, funcName)
		if fn == nil {
			t.Fatalf("function %s not found", funcName)
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if forbiddenConsensusScoringCall(call) {
				t.Fatalf("%s contains forbidden scorer/environment call at %s", funcName, fset.Position(call.Pos()))
			}
			return true
		})
	}
}

func findFuncDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func forbiddenConsensusScoringCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		switch fun.Name {
		case "DefaultMLScoringConfig",
			"DefaultDevelopmentMLScoringConfig",
			"DefaultTensorFlowScoringConfig",
			"DefaultDevelopmentTensorFlowScoringConfig",
			"NewStubMLScorer",
			"createTensorFlowScorer",
			"isTensorFlowEnabled",
			"getEnvOrDefault":
			return true
		}
	case *ast.SelectorExpr:
		if ident, ok := fun.X.(*ast.Ident); ok {
			return (ident.Name == "os" && fun.Sel.Name == "Getenv") ||
				(ident.Name == "inference" && fun.Sel.Name == "NewScorer")
		}
	}
	return false
}
