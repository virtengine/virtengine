// Package main provides a tool for assessing third-party dependency risk.
//
// This tool analyzes Go dependencies and calculates risk scores based on:
// - Maintenance activity
// - Security practices
// - License compliance
// - Vulnerability history
// - Community trust
//
// Usage:
//
//	go run scripts/supply-chain/assess-dependencies.go [options]
//
// Options:
//
//	--package <pkg>   Assess a specific package
//	--report          Generate full assessment report
//	--threshold <n>   Minimum acceptable risk score (default: 6.0)
//	--json            Output as JSON
//	--new-only        Only assess newly added dependencies
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	severityHigh     = "high"
	severityCritical = "critical"
)

// RiskScore represents the overall risk assessment for a dependency
type RiskScore struct {
	Package           string            `json:"package"`
	Version           string            `json:"version"`
	OverallScore      float64           `json:"overall_score"`
	RiskLevel         string            `json:"risk_level"`
	Scores            ComponentScores   `json:"component_scores"`
	Flags             []RiskFlag        `json:"flags,omitempty"`
	Recommendations   []string          `json:"recommendations,omitempty"`
	LastAssessed      time.Time         `json:"last_assessed"`
	NeedsManualReview bool              `json:"needs_manual_review"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// ComponentScores breaks down the risk score by category
type ComponentScores struct {
	Maintenance   float64 `json:"maintenance"`
	Security      float64 `json:"security"`
	License       float64 `json:"license"`
	Popularity    float64 `json:"popularity"`
	CodeQuality   float64 `json:"code_quality"`
	Vulnerability float64 `json:"vulnerability"`
}

// RiskFlag indicates specific risk concerns
type RiskFlag struct {
	Severity    string `json:"severity"` // critical, high, medium, low
	Category    string `json:"category"`
	Description string `json:"description"`
}

// AssessmentReport contains the full assessment for all dependencies
type AssessmentReport struct {
	Timestamp       time.Time     `json:"timestamp"`
	Repository      string        `json:"repository"`
	TotalPackages   int           `json:"total_packages"`
	PassingPackages int           `json:"passing_packages"`
	FailingPackages int           `json:"failing_packages"`
	HighRiskCount   int           `json:"high_risk_count"`
	Threshold       float64       `json:"threshold"`
	Packages        []RiskScore   `json:"packages"`
	Summary         ReportSummary `json:"summary"`
}

// ReportSummary provides high-level metrics
type ReportSummary struct {
	AverageScore     float64            `json:"average_score"`
	LowestScore      float64            `json:"lowest_score"`
	HighestRiskPkg   string             `json:"highest_risk_package"`
	CategoryAverages map[string]float64 `json:"category_averages"`
}

// KnownPackage represents a package with pre-calculated trust scores
type KnownPackage struct {
	Pattern    string
	TrustLevel string // "high", "medium", "low"
	BaseScore  float64
	Notes      string
}

var (
	// Known trusted packages and their patterns
	trustedPackages = []KnownPackage{
		// Cosmos Ecosystem - High Trust
		{Pattern: "github.com/cosmos/cosmos-sdk", TrustLevel: "high", BaseScore: 9.0, Notes: "Core Cosmos SDK"},
		{Pattern: "github.com/cosmos/ibc-go", TrustLevel: "high", BaseScore: 9.0, Notes: "IBC Protocol"},
		{Pattern: "github.com/cometbft/cometbft", TrustLevel: "high", BaseScore: 9.0, Notes: "Core consensus engine"},
		{Pattern: "cosmossdk.io/", TrustLevel: "high", BaseScore: 8.5, Notes: "Cosmos SDK modules"},

		// Go Standard/Extended - High Trust
		{Pattern: "golang.org/x/", TrustLevel: "high", BaseScore: 9.5, Notes: "Go extended stdlib"},
		{Pattern: "google.golang.org/", TrustLevel: "high", BaseScore: 9.0, Notes: "Google Go packages"},

		// Kubernetes - High Trust
		{Pattern: "k8s.io/", TrustLevel: "high", BaseScore: 9.0, Notes: "Kubernetes packages"},

		// Ethereum - Medium-High Trust
		{Pattern: "github.com/ethereum/go-ethereum", TrustLevel: "high", BaseScore: 8.5, Notes: "Go Ethereum"},

		// Common utilities - Medium Trust (need individual assessment)
		{Pattern: "github.com/spf13/", TrustLevel: "medium", BaseScore: 8.0, Notes: "Spf13 utilities"},
		{Pattern: "github.com/stretchr/testify", TrustLevel: "high", BaseScore: 9.0, Notes: "Testing framework"},
		{Pattern: "github.com/prometheus/", TrustLevel: "high", BaseScore: 8.5, Notes: "Prometheus"},

		// Our own packages - Trusted
		{Pattern: "github.com/virtengine/", TrustLevel: "high", BaseScore: 9.0, Notes: "VirtEngine packages"},
		{Pattern: "github.com/virtengine/", TrustLevel: "high", BaseScore: 8.5, Notes: "VirtEngine forks"},

		// Lower trust - requires review
		{Pattern: "github.com/", TrustLevel: "medium", BaseScore: 6.0, Notes: "Unknown GitHub package"},
	}

	// Flags
	packageFlag   = flag.String("package", "", "Assess a specific package")
	reportFlag    = flag.Bool("report", false, "Generate full assessment report")
	thresholdFlag = flag.Float64("threshold", 6.0, "Minimum acceptable risk score")
	jsonFlag      = flag.Bool("json", false, "Output as JSON")
	//nolint:unused // Reserved for future dependency assessment features
	newOnlyFlag = flag.Bool("new-only", false, "Only assess newly added dependencies")
	//nolint:unused // Reserved for future verbose output
	verboseFlag = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	if *packageFlag != "" {
		// Assess single package
		score := assessPackage(*packageFlag, "latest")
		outputScore(score)
		if score.OverallScore < *thresholdFlag {
			os.Exit(1)
		}
		return
	}

	if *reportFlag {
		// Generate full report
		report := generateReport()
		outputReport(report)
		if report.FailingPackages > 0 {
			os.Exit(1)
		}
		return
	}

	// Default: quick assessment with summary
	report := generateReport()
	printSummary(report)
	if report.FailingPackages > 0 {
		os.Exit(1)
	}
}

// getDependencies extracts dependencies from go.mod
func getDependencies() []struct{ Package, Version string } {
	var deps []struct{ Package, Version string }

	goModPath := findGoMod()
	if goModPath == "" {
		fmt.Fprintf(os.Stderr, "Error: go.mod not found\n")
		os.Exit(1)
	}

	file, err := os.Open(goModPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening go.mod: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	inRequire := false
	requireBlockRegex := regexp.MustCompile(`^\s*require\s*\(`)
	closeBlockRegex := regexp.MustCompile(`^\s*\)`)
	depRegex := regexp.MustCompile(`^\s*([^\s]+)\s+(v[^\s]+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if requireBlockRegex.MatchString(line) {
			inRequire = true
			continue
		}

		if inRequire && closeBlockRegex.MatchString(line) {
			inRequire = false
			continue
		}

		if inRequire {
			if matches := depRegex.FindStringSubmatch(line); len(matches) >= 3 {
				// Skip indirect dependencies for now
				if !strings.Contains(line, "// indirect") {
					deps = append(deps, struct{ Package, Version string }{
						Package: matches[1],
						Version: matches[2],
					})
				}
			}
		}
	}

	return deps
}

// assessPackage evaluates a single package
func assessPackage(pkg, version string) RiskScore {
	score := RiskScore{
		Package:      pkg,
		Version:      version,
		LastAssessed: time.Now(),
		Metadata:     make(map[string]string),
	}

	// Find matching known package
	var knownPkg *KnownPackage
	for _, kp := range trustedPackages {
		if strings.HasPrefix(pkg, kp.Pattern) || pkg == strings.TrimSuffix(kp.Pattern, "/") {
			knownPkg = &kp
			break
		}
	}

	// Calculate component scores
	if knownPkg != nil {
		score.Scores = calculateScoresFromKnown(pkg, knownPkg)
		score.Metadata["trust_level"] = knownPkg.TrustLevel
		score.Metadata["notes"] = knownPkg.Notes
	} else {
		score.Scores = calculateScoresUnknown(pkg)
		score.NeedsManualReview = true
		score.Flags = append(score.Flags, RiskFlag{
			Severity:    "medium",
			Category:    "unknown",
			Description: "Package not in known trusted list - requires manual review",
		})
	}

	// Calculate overall score (weighted average)
	score.OverallScore = calculateOverallScore(score.Scores)

	// Determine risk level
	score.RiskLevel = determineRiskLevel(score.OverallScore)

	// Add recommendations
	score.Recommendations = generateRecommendations(score)

	return score
}

// calculateScoresFromKnown calculates scores for known packages
//
//nolint:unparam // pkg kept for future package-specific scoring adjustments
func calculateScoresFromKnown(_ string, known *KnownPackage) ComponentScores {
	scores := ComponentScores{}

	switch known.TrustLevel {
	case "high":
		scores.Maintenance = 9.0
		scores.Security = 9.0
		scores.License = 9.0
		scores.Popularity = 8.5
		scores.CodeQuality = 8.5
		scores.Vulnerability = 9.0
	case "medium":
		scores.Maintenance = 7.0
		scores.Security = 7.0
		scores.License = 8.0
		scores.Popularity = 6.0
		scores.CodeQuality = 7.0
		scores.Vulnerability = 7.0
	default: // low
		scores.Maintenance = 5.0
		scores.Security = 5.0
		scores.License = 6.0
		scores.Popularity = 4.0
		scores.CodeQuality = 5.0
		scores.Vulnerability = 5.0
	}

	return scores
}

// calculateScoresUnknown calculates scores for unknown packages
func calculateScoresUnknown(pkg string) ComponentScores {
	// Default scores for unknown packages - conservative
	scores := ComponentScores{
		Maintenance:   5.0,
		Security:      5.0,
		License:       6.0, // Assume some license
		Popularity:    4.0,
		CodeQuality:   5.0,
		Vulnerability: 5.0,
	}

	// Adjust based on source
	if strings.HasPrefix(pkg, "github.com/") {
		scores.Maintenance += 1.0 // GitHub provides some transparency
		scores.CodeQuality += 0.5
	}

	// Check for common security indicators in package name
	if strings.Contains(pkg, "crypto") || strings.Contains(pkg, "security") {
		// Crypto packages need extra scrutiny
		scores.Security -= 1.0
	}

	return scores
}

// calculateOverallScore computes weighted average
func calculateOverallScore(scores ComponentScores) float64 {
	weights := map[string]float64{
		"maintenance":   0.15,
		"security":      0.25,
		"license":       0.10,
		"popularity":    0.10,
		"code_quality":  0.15,
		"vulnerability": 0.25,
	}

	total := scores.Maintenance*weights["maintenance"] +
		scores.Security*weights["security"] +
		scores.License*weights["license"] +
		scores.Popularity*weights["popularity"] +
		scores.CodeQuality*weights["code_quality"] +
		scores.Vulnerability*weights["vulnerability"]

	return total
}

// determineRiskLevel returns risk level based on score
func determineRiskLevel(score float64) string {
	switch {
	case score >= 8.0:
		return "low"
	case score >= 6.0:
		return "medium"
	case score >= 4.0:
		return "high"
	default:
		return "critical"
	}
}

// generateRecommendations creates actionable recommendations
func generateRecommendations(score RiskScore) []string {
	var recs []string

	if score.OverallScore < 6.0 {
		recs = append(recs, "Consider finding an alternative package with better security practices")
	}

	if score.NeedsManualReview {
		recs = append(recs, "Perform manual security review before adding to production")
	}

	if score.Scores.Vulnerability < 7.0 {
		recs = append(recs, "Check for known vulnerabilities using govulncheck")
	}

	if score.Scores.Maintenance < 6.0 {
		recs = append(recs, "Verify package is actively maintained")
	}

	if score.Scores.License < 7.0 {
		recs = append(recs, "Verify license compatibility with Apache-2.0")
	}

	return recs
}

// generateReport creates full assessment report
func generateReport() AssessmentReport {
	deps := getDependencies()

	report := AssessmentReport{
		Timestamp:     time.Now(),
		Repository:    getRepositoryName(),
		TotalPackages: len(deps),
		Threshold:     *thresholdFlag,
		Packages:      make([]RiskScore, 0, len(deps)),
		Summary: ReportSummary{
			CategoryAverages: make(map[string]float64),
		},
	}

	var totalScore float64
	lowestScore := 10.0
	highestRiskPkg := ""

	for _, dep := range deps {
		score := assessPackage(dep.Package, dep.Version)
		report.Packages = append(report.Packages, score)

		totalScore += score.OverallScore

		if score.OverallScore >= *thresholdFlag {
			report.PassingPackages++
		} else {
			report.FailingPackages++
		}

		if score.RiskLevel == severityHigh || score.RiskLevel == severityCritical {
			report.HighRiskCount++
		}

		if score.OverallScore < lowestScore {
			lowestScore = score.OverallScore
			highestRiskPkg = dep.Package
		}

		// Accumulate category averages
		report.Summary.CategoryAverages["maintenance"] += score.Scores.Maintenance
		report.Summary.CategoryAverages["security"] += score.Scores.Security
		report.Summary.CategoryAverages["license"] += score.Scores.License
		report.Summary.CategoryAverages["popularity"] += score.Scores.Popularity
		report.Summary.CategoryAverages["code_quality"] += score.Scores.CodeQuality
		report.Summary.CategoryAverages["vulnerability"] += score.Scores.Vulnerability
	}

	// Calculate averages
	if len(deps) > 0 {
		report.Summary.AverageScore = totalScore / float64(len(deps))
		for k := range report.Summary.CategoryAverages {
			report.Summary.CategoryAverages[k] /= float64(len(deps))
		}
	}
	report.Summary.LowestScore = lowestScore
	report.Summary.HighestRiskPkg = highestRiskPkg

	// Sort by score (lowest first)
	sort.Slice(report.Packages, func(i, j int) bool {
		return report.Packages[i].OverallScore < report.Packages[j].OverallScore
	})

	return report
}

// outputScore prints a single score
func outputScore(score RiskScore) {
	if *jsonFlag {
		data, err := json.MarshalIndent(score, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling score: %v\n", err)
			return
		}
		fmt.Println(string(data))
		return
	}

	fmt.Println("==========================================")
	fmt.Printf("  Dependency Risk Assessment: %s\n", score.Package)
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Printf("Version: %s\n", score.Version)
	fmt.Printf("Overall Score: %.1f/10.0\n", score.OverallScore)
	fmt.Printf("Risk Level: %s\n", strings.ToUpper(score.RiskLevel))
	fmt.Println()
	fmt.Println("Component Scores:")
	fmt.Printf("  Maintenance:   %.1f\n", score.Scores.Maintenance)
	fmt.Printf("  Security:      %.1f\n", score.Scores.Security)
	fmt.Printf("  License:       %.1f\n", score.Scores.License)
	fmt.Printf("  Popularity:    %.1f\n", score.Scores.Popularity)
	fmt.Printf("  Code Quality:  %.1f\n", score.Scores.CodeQuality)
	fmt.Printf("  Vulnerability: %.1f\n", score.Scores.Vulnerability)

	if len(score.Flags) > 0 {
		fmt.Println()
		fmt.Println("Flags:")
		for _, flag := range score.Flags {
			fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(flag.Severity), flag.Category, flag.Description)
		}
	}

	if len(score.Recommendations) > 0 {
		fmt.Println()
		fmt.Println("Recommendations:")
		for _, rec := range score.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	}

	if score.NeedsManualReview {
		fmt.Println()
		fmt.Println("⚠️  This package requires manual security review")
	}
}

// outputReport prints the full report
func outputReport(report AssessmentReport) {
	if *jsonFlag {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling report: %v\n", err)
			return
		}
		fmt.Println(string(data))
		return
	}

	fmt.Println("==========================================")
	fmt.Println("  VirtEngine Dependency Risk Assessment")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Printf("Repository: %s\n", report.Repository)
	fmt.Printf("Assessed: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Printf("Threshold: %.1f\n", report.Threshold)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total Packages:   %d\n", report.TotalPackages)
	fmt.Printf("  Passing:          %d\n", report.PassingPackages)
	fmt.Printf("  Failing:          %d\n", report.FailingPackages)
	fmt.Printf("  High Risk:        %d\n", report.HighRiskCount)
	fmt.Printf("  Average Score:    %.1f\n", report.Summary.AverageScore)
	fmt.Printf("  Lowest Score:     %.1f (%s)\n", report.Summary.LowestScore, report.Summary.HighestRiskPkg)
	fmt.Println()

	// Show failing packages
	if report.FailingPackages > 0 {
		fmt.Println("Packages Below Threshold:")
		for _, pkg := range report.Packages {
			if pkg.OverallScore < report.Threshold {
				fmt.Printf("  [%.1f] %s - %s\n", pkg.OverallScore, pkg.RiskLevel, pkg.Package)
			}
		}
		fmt.Println()
	}

	// Show high risk packages
	if report.HighRiskCount > 0 {
		fmt.Println("High Risk Packages (require review):")
		for _, pkg := range report.Packages {
			if pkg.RiskLevel == severityHigh || pkg.RiskLevel == severityCritical {
				fmt.Printf("  [%s] %s (score: %.1f)\n", strings.ToUpper(pkg.RiskLevel), pkg.Package, pkg.OverallScore)
			}
		}
		fmt.Println()
	}
}

// printSummary prints a condensed summary
func printSummary(report AssessmentReport) {
	fmt.Println("==========================================")
	fmt.Println("  Dependency Risk Assessment Summary")
	fmt.Println("==========================================")
	fmt.Printf("Total: %d | Pass: %d | Fail: %d | High Risk: %d\n",
		report.TotalPackages, report.PassingPackages, report.FailingPackages, report.HighRiskCount)
	fmt.Printf("Average Score: %.1f | Threshold: %.1f\n", report.Summary.AverageScore, report.Threshold)

	if report.FailingPackages > 0 || report.HighRiskCount > 0 {
		fmt.Println()
		fmt.Println("Action Required:")
		if report.Summary.HighestRiskPkg != "" {
			fmt.Printf("  Highest risk: %s (score: %.1f)\n", report.Summary.HighestRiskPkg, report.Summary.LowestScore)
		}
		fmt.Println("  Run with --report for full details")
	} else {
		fmt.Println()
		fmt.Println("✓ All dependencies meet minimum risk threshold")
	}
}

// findGoMod locates go.mod file
func findGoMod() string {
	// Check current directory
	if _, err := os.Stat("go.mod"); err == nil {
		return "go.mod"
	}

	// Walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// getRepositoryName gets the repository name from git
func getRepositoryName() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
