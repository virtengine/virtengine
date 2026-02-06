package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const defaultTrackerPath = "_docs/audit/findings.json"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "add":
		runAdd(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "summary":
		runSummary(os.Args[2:])
	case "update":
		runUpdate(os.Args[2:])
	case "note":
		runNote(os.Args[2:])
	case "export":
		runExport(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: audit-tracker <command> [options]")
	fmt.Println("Commands: init, add, list, summary, update, note, export")
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	_ = fs.Parse(args)

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	if err := tracker.Save(); err != nil {
		exitOnErr(err)
	}
	fmt.Printf("Initialized tracker at %s\n", *path)
}

func runAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	title := fs.String("title", "", "finding title")
	severity := fs.String("severity", "", "severity (critical/high/medium/low/info)")
	auditFirm := fs.String("audit-firm", "", "audit firm")
	category := fs.String("category", "", "category")
	description := fs.String("description", "", "description")
	impact := fs.String("impact", "", "impact")
	location := fs.String("location", "", "comma-separated file paths")
	remediation := fs.String("remediation", "", "remediation guidance")
	assignedTo := fs.String("assigned-to", "", "assignee")
	due := fs.String("due", "", "due date (YYYY-MM-DD)")
	_ = fs.Parse(args)

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	dueDate, err := parseDate(*due)
	exitOnErr(err)

	sev, err := parseSeverity(*severity)
	exitOnErr(err)

	finding := &Finding{
		Title:       strings.TrimSpace(*title),
		Severity:    sev,
		AuditFirm:   strings.TrimSpace(*auditFirm),
		Category:    strings.TrimSpace(*category),
		Description: strings.TrimSpace(*description),
		Impact:      strings.TrimSpace(*impact),
		Location:    splitList(*location),
		Remediation: strings.TrimSpace(*remediation),
		AssignedTo:  strings.TrimSpace(*assignedTo),
		DueDate:     dueDate,
	}

	exitOnErr(tracker.AddFinding(finding))
	exitOnErr(tracker.Save())
	fmt.Printf("Added finding %s\n", finding.ID)
}

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	status := fs.String("status", "", "status filter (comma-separated)")
	severity := fs.String("severity", "", "severity filter (comma-separated)")
	_ = fs.Parse(args)

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	statusFilter := normalizeList(*status)
	severityFilter := normalizeList(*severity)

	for _, finding := range tracker.Findings {
		if len(statusFilter) > 0 && !contains(statusFilter, string(finding.Status)) {
			continue
		}
		if len(severityFilter) > 0 && !contains(severityFilter, string(finding.Severity)) {
			continue
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", finding.ID, finding.Severity, finding.Status, finding.Title)
	}
}

func runSummary(args []string) {
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	_ = fs.Parse(args)

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	summary := tracker.Summary()
	keys := make([]string, 0, len(summary))
	for key := range summary {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, summary[key])
	}
}

func runUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	id := fs.String("id", "", "finding ID")
	status := fs.String("status", "", "status")
	severity := fs.String("severity", "", "severity")
	assignedTo := fs.String("assigned-to", "", "assignee")
	remediation := fs.String("remediation", "", "remediation guidance")
	due := fs.String("due", "", "due date (YYYY-MM-DD)")
	fixedIn := fs.String("fixed-in", "", "fixed in commit/PR")
	verifiedBy := fs.String("verified-by", "", "verified by")
	verifiedDate := fs.String("verified-date", "", "verified date (YYYY-MM-DD)")
	_ = fs.Parse(args)

	if strings.TrimSpace(*id) == "" {
		exitOnErr(fmt.Errorf("id is required"))
	}

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	exitOnErr(tracker.UpdateFinding(*id, func(f *Finding) error {
		if *status != "" {
			parsed, err := parseStatus(*status)
			if err != nil {
				return err
			}
			f.Status = parsed
		}
		if *severity != "" {
			parsed, err := parseSeverity(*severity)
			if err != nil {
				return err
			}
			f.Severity = parsed
		}
		if *assignedTo != "" {
			f.AssignedTo = strings.TrimSpace(*assignedTo)
		}
		if *remediation != "" {
			f.Remediation = strings.TrimSpace(*remediation)
		}
		if *due != "" {
			date, err := parseDate(*due)
			if err != nil {
				return err
			}
			f.DueDate = date
		}
		if *fixedIn != "" {
			f.FixedIn = strings.TrimSpace(*fixedIn)
		}
		if *verifiedBy != "" {
			f.VerifiedBy = strings.TrimSpace(*verifiedBy)
		}
		if *verifiedDate != "" {
			date, err := parseDate(*verifiedDate)
			if err != nil {
				return err
			}
			f.VerifiedDate = date
		}
		if f.Status == StatusVerified && f.VerifiedDate.IsZero() {
			f.VerifiedDate = time.Now().UTC()
		}
		return nil
	}))

	exitOnErr(tracker.Save())
	fmt.Printf("Updated finding %s\n", *id)
}

func runNote(args []string) {
	fs := flag.NewFlagSet("note", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	id := fs.String("id", "", "finding ID")
	author := fs.String("author", "", "note author")
	text := fs.String("text", "", "note text")
	_ = fs.Parse(args)

	if strings.TrimSpace(*id) == "" {
		exitOnErr(fmt.Errorf("id is required"))
	}

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	exitOnErr(tracker.AddNote(*id, *author, *text))
	exitOnErr(tracker.Save())
	fmt.Printf("Added note to %s\n", *id)
}

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	path := fs.String("path", defaultTrackerPath, "tracker file path")
	_ = fs.Parse(args)

	tracker, err := NewTracker(*path)
	exitOnErr(err)

	data, err := json.MarshalIndent(tracker, "", "  ")
	exitOnErr(err)
	fmt.Println(string(data))
}

func parseDate(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}

func parseSeverity(value string) (Severity, error) {
	if strings.TrimSpace(value) == "" {
		return "", errorsNew("severity is required")
	}
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(SeverityCritical), "CRIT":
		return SeverityCritical, nil
	case string(SeverityHigh):
		return SeverityHigh, nil
	case string(SeverityMedium), "MED":
		return SeverityMedium, nil
	case string(SeverityLow):
		return SeverityLow, nil
	case string(SeverityInfo):
		return SeverityInfo, nil
	default:
		return "", errorsNew("invalid severity: %s", value)
	}
}

func parseStatus(value string) (Status, error) {
	if strings.TrimSpace(value) == "" {
		return "", errorsNew("status is required")
	}
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(StatusNew):
		return StatusNew, nil
	case string(StatusTriaged):
		return StatusTriaged, nil
	case string(StatusInProgress), "INPROGRESS":
		return StatusInProgress, nil
	case string(StatusFixed):
		return StatusFixed, nil
	case string(StatusVerified):
		return StatusVerified, nil
	case string(StatusWontFix):
		return StatusWontFix, nil
	case string(StatusFalsePos), "FALSEPOSITIVE":
		return StatusFalsePos, nil
	default:
		return "", errorsNew("invalid status: %s", value)
	}
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func normalizeList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.ToUpper(strings.TrimSpace(part))
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func contains(haystack []string, value string) bool {
	for _, item := range haystack {
		if item == value {
			return true
		}
	}
	return false
}

func errorsNew(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
