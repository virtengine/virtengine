// Command consensusdeterminism enforces high-signal deterministic-consensus policy.
package main

import (
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
)

const (
	ruleWallClock        = "wall-clock"
	ruleRandomness       = "randomness"
	ruleFilesystem       = "filesystem"
	ruleExternalNetwork  = "external-network"
	ruleExternalCallback = "external-callback"
	ruleMapIteration     = "map-iteration"
	ruleFloatingDecision = "floating-decision"
)

type finding struct {
	Rule     string
	Path     string
	Function string
	Line     int
	Detail   string
	Allowed  bool
	Reason   string
}

type allowance struct {
	Rule     string
	Path     string
	Function string
	Reason   string
}

var defaultAllowlist = []allowance{
	{Rule: ruleMapIteration, Path: "x/audit/keeper/keeper.go", Function: "CreateOrUpdateProviderAttributes", Reason: "map entries are sorted by attribute key/value before protobuf persistence"},
	{Rule: ruleMapIteration, Path: "x/audit/keeper/keeper.go", Function: "DeleteProviderAttributes", Reason: "map entries are sorted by attribute key before protobuf persistence"},
	{Rule: ruleMapIteration, Path: "x/hpc/keeper/msg_server_templates.go", Function: "resolveTemplateWorkloadSpec", Reason: "map is copied into a JSON object; encoding/json canonicalizes string map keys"},
	{Rule: ruleMapIteration, Path: "x/hpc/keeper/settlement.go", Function: "copyStringMap", Reason: "pure map copy with no ordering-dependent side effect"},
	{Rule: ruleMapIteration, Path: "x/settlement/keeper/staking_reward_routing.go", Function: "buildStakeRecipientsForValidator", Reason: "addresses are collected then sorted before recipient construction"},
	{Rule: ruleMapIteration, Path: "x/veid/keeper/gdpr_erasure.go", Function: "getAccountKeyFingerprints", Reason: "set is converted to a slice used only in query/export responses, not persisted consensus state"},
	{Rule: ruleMapIteration, Path: "x/veid/keeper/geo_restrictions.go", Function: "GetBlockedCountries", Reason: "set is converted to a sorted slice before return"},
	{Rule: ruleMapIteration, Path: "x/veid/keeper/privacy_proofs.go", Function: "deterministicClaimsString", Reason: "keys are collected and sorted before canonical string construction"},
	{Rule: ruleMapIteration, Path: "x/veid/keeper/model_version.go", Function: "ReportValidatorModelVersions", Reason: "iteration only computes order-independent mismatch membership; persisted map JSON keys are canonicalized"},
	{Rule: ruleFilesystem, Path: "x/veid/keeper/model_hash_governance.go", Function: "ComputeLocalModelHash", Reason: "off-chain startup/operator compatibility helper; no production keeper call site"},
	{Rule: ruleFilesystem, Path: "x/veid/keeper/scoring.go", Function: "DefaultTensorFlowScoringConfig", Reason: "off-chain scorer construction compatibility; active vote-extension carrier emits no evidence"},
	{Rule: ruleFilesystem, Path: "x/veid/keeper/scoring.go", Function: "isTensorFlowEnabled", Reason: "off-chain scorer construction compatibility; active vote-extension carrier emits no evidence"},
	{Rule: ruleFilesystem, Path: "x/veid/keeper/scoring.go", Function: "isRealInferenceReady", Reason: "off-chain scorer readiness compatibility; active vote-extension carrier emits no evidence"},
	{Rule: ruleFilesystem, Path: "x/veid/keeper/scoring.go", Function: "getEnvOrDefault", Reason: "off-chain scorer configuration helper; active vote-extension carrier emits no evidence"},
	{Rule: ruleFilesystem, Path: "x/veid/keeper/zkproofs_circuits.go", Function: "NewZKProofSystem", Reason: "keeper construction reads optional proving-key location; no state transition branches on host availability"},
	{Rule: ruleRandomness, Path: "x/veid/keeper/biometric_hash.go", Function: "GenerateTemplateSalt", Reason: "off-chain cryptographic salt helper; no production consensus caller"},
	{Rule: ruleFloatingDecision, Path: "x/veid/keeper/biometric_hash.go", Function: "MatchTemplateHash", Reason: "source-compatible off-chain biometric comparison helper; no production consensus caller"},
	{Rule: ruleFloatingDecision, Path: "x/veid/keeper/data_lifecycle.go", Function: "ValidateNoRawBiometricsOnChain", Reason: "deterministic byte-content validation; IEEE operations do not depend on host or external input"},
	{Rule: ruleMapIteration, Path: "x/veid/keeper/evidence_pipeline.go", Function: "averageConfidenceBP", Reason: "carrier version 0 emits no evidence; future activation must replace float-map aggregation with canonical fixed point"},
	{Rule: ruleFloatingDecision, Path: "x/veid/keeper/evidence_pipeline.go", Function: "floatToBasisPoints", Reason: "carrier version 0 emits no evidence; future activation must replace this compatibility converter with fixed point"},
	{Rule: ruleFloatingDecision, Path: "x/veid/keeper/feature_extraction.go", Function: "extractLivenessFeatures", Reason: "inactive local ML compatibility pipeline; carrier version 0 never executes or commits its output"},
	{Rule: ruleFloatingDecision, Path: "x/veid/keeper/feature_extraction.go", Function: "generateDeterministicEmbedding", Reason: "inactive local ML compatibility pipeline; carrier version 0 never executes or commits its output"},
	{Rule: ruleFloatingDecision, Path: "x/veid/keeper/scoring.go", Function: "computeConfidence", Reason: "inactive stub ML compatibility scorer; carrier version 0 never executes or commits its output"},
}

var externalMethodNames = map[string]struct{}{
	"Cancel":               {},
	"ExecuteSwap":          {},
	"FindPayoutByMetadata": {},
	"GetQuote":             {},
	"GetStatus":            {},
	"InitiatePayout":       {},
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	findings, err := scanRepository(*root, defaultAllowlist)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	unapproved := 0
	approved := 0
	for _, item := range findings {
		if item.Allowed {
			approved++
			continue
		}
		unapproved++
		fmt.Fprintf(os.Stderr, "%s:%d: %s in %s: %s\n", item.Path, item.Line, item.Rule, item.Function, item.Detail)
	}
	fmt.Printf("consensus determinism: %d unapproved finding(s), %d narrowly allowlisted finding(s)\n", unapproved, approved)
	if unapproved != 0 {
		os.Exit(1)
	}
}

func scanRepository(root string, allowlist []allowance) ([]finding, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var findings []finding
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if !isConsensusPolicyFile(rel) {
			return nil
		}

		fileFindings, scanErr := scanFile(path, rel)
		if scanErr != nil {
			return scanErr
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i := range findings {
		for _, allowed := range allowlist {
			if findings[i].Rule == allowed.Rule && findings[i].Path == allowed.Path && findings[i].Function == allowed.Function {
				findings[i].Allowed = true
				findings[i].Reason = allowed.Reason
				break
			}
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Rule < findings[j].Rule
	})
	return findings, nil
}

func isConsensusPolicyFile(path string) bool {
	base := filepath.Base(path)
	if path == "app/proposal_handler.go" || strings.HasPrefix(path, "app/ante") {
		return true
	}
	if !strings.HasPrefix(path, "x/") {
		return false
	}
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return false
	}
	module := parts[1]
	if module != "settlement" && module != "veid" && module != "hpc" && module != "audit" {
		return strings.Contains(path, "/keeper/") && (strings.HasPrefix(base, "msg_server") || base == "begin_block.go" || base == "end_block.go")
	}
	if base == "module.go" || base == "genesis.go" {
		return true
	}
	return strings.Contains(path, "/keeper/")
}

func scanFile(path, rel string) ([]finding, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	aliases := make(map[string]string)
	for _, imported := range parsed.Imports {
		importPath := strings.Trim(imported.Path.Value, "\"")
		name := filepath.Base(importPath)
		if imported.Name != nil {
			name = imported.Name.Name
		}
		aliases[name] = importPath
	}

	mapTypes := packageMapTypes(filepath.Dir(path))

	var findings []finding
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		functionName := fn.Name.Name
		mapVars := declaredMapVariables(fn, mapTypes)
		sortedKeyRanges := sortedMapKeyRanges(fn.Body)
		orderIndependentRanges := orderIndependentMapRanges(fn.Body)
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.CallExpr:
				if item, ok := callFinding(fset, rel, functionName, aliases, value); ok {
					findings = append(findings, item)
				}
			case *ast.RangeStmt:
				if isMapExpression(value.X, mapVars, mapTypes) && !sortedKeyRanges[value] && !orderIndependentRanges[value] {
					findings = append(findings, newFinding(fset, rel, functionName, ruleMapIteration, value.Pos(), "range over a map in consensus policy code"))
				}
			case *ast.IfStmt:
				if containsFloatingPoint(value.Cond) {
					findings = append(findings, newFinding(fset, rel, functionName, ruleFloatingDecision, value.Cond.Pos(), "floating-point value controls a consensus branch"))
				}
			case *ast.SwitchStmt:
				if value.Tag != nil && containsFloatingPoint(value.Tag) {
					findings = append(findings, newFinding(fset, rel, functionName, ruleFloatingDecision, value.Tag.Pos(), "floating-point value controls a consensus switch"))
				}
			}
			return true
		})
	}
	return findings, nil
}

func orderIndependentMapRanges(body *ast.BlockStmt) map[*ast.RangeStmt]bool {
	result := make(map[*ast.RangeStmt]bool)
	ast.Inspect(body, func(node ast.Node) bool {
		rangeStmt, ok := node.(*ast.RangeStmt)
		if !ok || rangeStmt.Key == nil || rangeStmt.Value == nil {
			return true
		}
		key, keyOK := rangeStmt.Key.(*ast.Ident)
		value, valueOK := rangeStmt.Value.(*ast.Ident)
		if !keyOK || !valueOK {
			return true
		}
		safe := len(rangeStmt.Body.List) > 0
		for _, stmt := range rangeStmt.Body.List {
			assign, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
				safe = false
				break
			}
			index, ok := assign.Lhs[0].(*ast.IndexExpr)
			if !ok {
				safe = false
				break
			}
			indexKey, ok := index.Index.(*ast.Ident)
			rhs, rhsOK := assign.Rhs[0].(*ast.Ident)
			if !ok || !rhsOK || indexKey.Name != key.Name || rhs.Name != value.Name {
				safe = false
				break
			}
		}
		if safe {
			result[rangeStmt] = true
		}
		return true
	})
	return result
}

func sortedMapKeyRanges(body *ast.BlockStmt) map[*ast.RangeStmt]bool {
	result := make(map[*ast.RangeStmt]bool)
	ast.Inspect(body, func(node ast.Node) bool {
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range block.List {
			rangeStmt, ok := stmt.(*ast.RangeStmt)
			if !ok || rangeStmt.Value != nil || rangeStmt.Key == nil {
				continue
			}
			key, ok := rangeStmt.Key.(*ast.Ident)
			if !ok || key.Name == "_" {
				continue
			}
			collector := ""
			ast.Inspect(rangeStmt.Body, func(child ast.Node) bool {
				assign, ok := child.(*ast.AssignStmt)
				if !ok || len(assign.Rhs) != 1 || len(assign.Lhs) != 1 {
					return true
				}
				call, ok := assign.Rhs[0].(*ast.CallExpr)
				if !ok || len(call.Args) != 2 {
					return true
				}
				appendFn, ok := call.Fun.(*ast.Ident)
				if !ok || appendFn.Name != "append" {
					return true
				}
				dst, ok := call.Args[0].(*ast.Ident)
				if !ok {
					return true
				}
				appended, ok := call.Args[1].(*ast.Ident)
				if ok && appended.Name == key.Name {
					collector = dst.Name
				}
				return true
			})
			if collector == "" {
				continue
			}
			for _, following := range block.List[i+1:] {
				exprStmt, ok := following.(*ast.ExprStmt)
				if !ok {
					continue
				}
				call, ok := exprStmt.X.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					continue
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || (selector.Sel.Name != "Strings" && selector.Sel.Name != "Slice" && selector.Sel.Name != "Stable") {
					continue
				}
				owner, ok := selector.X.(*ast.Ident)
				arg, argOK := call.Args[0].(*ast.Ident)
				if ok && argOK && owner.Name == "sort" && arg.Name == collector {
					result[rangeStmt] = true
					break
				}
			}
		}
		return true
	})
	return result
}

func callFinding(fset *token.FileSet, rel, functionName string, aliases map[string]string, call *ast.CallExpr) (finding, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return finding{}, false
	}
	owner, ownerIsIdent := selector.X.(*ast.Ident)
	importPath := ""
	if ownerIsIdent {
		importPath = aliases[owner.Name]
	}
	if importPath == "" && isExternalReceiver(selector.X) {
		if _, external := externalMethodNames[selector.Sel.Name]; external {
			return newFinding(fset, rel, functionName, ruleExternalCallback, call.Pos(), "external-service method "+selector.Sel.Name), true
		}
	}
	if !ownerIsIdent {
		if selector.Sel.Name == "Seconds" {
			return newFinding(fset, rel, functionName, ruleFloatingDecision, call.Pos(), "time.Duration.Seconds introduces floating-point consensus arithmetic"), true
		}
		return finding{}, false
	}
	if importPath == "time" && isOneOf(selector.Sel.Name, "Now", "Since", "Until") {
		return newFinding(fset, rel, functionName, ruleWallClock, call.Pos(), "time."+selector.Sel.Name+" is forbidden"), true
	}
	if selector.Sel.Name == "Seconds" {
		return newFinding(fset, rel, functionName, ruleFloatingDecision, call.Pos(), "time.Duration.Seconds introduces floating-point consensus arithmetic"), true
	}
	switch {
	case (importPath == "crypto/rand" || importPath == "math/rand") && selector.Sel.Name != "New" && selector.Sel.Name != "NewSource":
		return newFinding(fset, rel, functionName, ruleRandomness, call.Pos(), importPath+"."+selector.Sel.Name+" is forbidden"), true
	case importPath == "os" && isOneOf(selector.Sel.Name, "Open", "OpenFile", "ReadFile", "Stat", "ReadDir", "Getenv"):
		return newFinding(fset, rel, functionName, ruleFilesystem, call.Pos(), "host filesystem/environment call os."+selector.Sel.Name), true
	case importPath == "path/filepath" && isOneOf(selector.Sel.Name, "Walk", "WalkDir", "Glob", "EvalSymlinks"):
		return newFinding(fset, rel, functionName, ruleFilesystem, call.Pos(), "host filesystem call filepath."+selector.Sel.Name), true
	case (importPath == "net" || importPath == "net/http" || strings.Contains(importPath, "google.golang.org/grpc")) && isOneOf(selector.Sel.Name, "Dial", "DialContext", "DialTimeout", "Get", "Head", "Post", "PostForm", "Do", "LookupAddr", "LookupCNAME", "LookupHost", "LookupIP", "LookupIPAddr", "LookupMX", "LookupNS", "LookupPort", "LookupSRV", "LookupTXT"):
		return newFinding(fset, rel, functionName, ruleExternalNetwork, call.Pos(), "external network call "+selector.Sel.Name), true
	default:
		return finding{}, false
	}
}

func isExternalReceiver(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	name := strings.ToLower(ident.Name)
	return strings.Contains(name, "bridge") || strings.Contains(name, "dex") || strings.Contains(name, "external") || strings.Contains(name, "provider")
}

func packageMapTypes(dir string) map[string]bool {
	result := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, decl := range parsed.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := typeSpec.Type.(*ast.MapType); ok {
					result[typeSpec.Name.Name] = true
				}
			}
		}
	}
	return result
}

func declaredMapVariables(fn *ast.FuncDecl, mapTypes map[string]bool) map[string]bool {
	result := make(map[string]bool)
	mark := func(name *ast.Ident, typ ast.Expr) {
		if _, ok := typ.(*ast.MapType); ok {
			result[name.Name] = true
			return
		}
		if ident, ok := typ.(*ast.Ident); ok && mapTypes[ident.Name] {
			result[name.Name] = true
		}
	}
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				mark(name, field.Type)
			}
		}
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.AssignStmt:
			for i, rhs := range value.Rhs {
				call, ok := rhs.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					continue
				}
				ident, ok := call.Fun.(*ast.Ident)
				if !ok || ident.Name != "make" {
					continue
				}
				if _, ok := call.Args[0].(*ast.MapType); !ok {
					continue
				}
				if i < len(value.Lhs) {
					if name, ok := value.Lhs[i].(*ast.Ident); ok {
						result[name.Name] = true
					}
				}
			}
		case *ast.DeclStmt:
			gen, ok := value.Decl.(*ast.GenDecl)
			if !ok {
				return true
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok || valueSpec.Type == nil {
					continue
				}
				for _, name := range valueSpec.Names {
					mark(name, valueSpec.Type)
				}
			}
		}
		return true
	})
	return result
}

func isMapExpression(expr ast.Expr, mapVars, mapTypes map[string]bool) bool {
	switch value := expr.(type) {
	case *ast.Ident:
		return mapVars[value.Name]
	case *ast.CompositeLit:
		if _, ok := value.Type.(*ast.MapType); ok {
			return true
		}
		if ident, ok := value.Type.(*ast.Ident); ok {
			return mapTypes[ident.Name]
		}
	}
	return false
}

func containsFloatingPoint(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.FLOAT {
			found = true
			return false
		}
		if ident, ok := node.(*ast.Ident); ok && (ident.Name == "float32" || ident.Name == "float64") {
			found = true
			return false
		}
		return true
	})
	return found
}

func newFinding(fset *token.FileSet, path, functionName, rule string, pos token.Pos, detail string) finding {
	return finding{Rule: rule, Path: path, Function: functionName, Line: fset.Position(pos).Line, Detail: detail}
}

func isOneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
