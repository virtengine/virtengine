package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Severity levels.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// Status tracks finding lifecycle.
type Status string

const (
	StatusNew        Status = "NEW"
	StatusTriaged    Status = "TRIAGED"
	StatusInProgress Status = "IN_PROGRESS"
	StatusFixed      Status = "FIXED"
	StatusVerified   Status = "VERIFIED"
	StatusWontFix    Status = "WONTFIX"
	StatusFalsePos   Status = "FALSE_POSITIVE"
)

// Finding represents a security audit finding.
type Finding struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Severity     Severity  `json:"severity"`
	Status       Status    `json:"status"`
	AuditFirm    string    `json:"audit_firm"`
	Category     string    `json:"category"`
	Description  string    `json:"description"`
	Impact       string    `json:"impact"`
	Location     []string  `json:"location"`
	Remediation  string    `json:"remediation"`
	AssignedTo   string    `json:"assigned_to"`
	DueDate      time.Time `json:"due_date"`
	FixedIn      string    `json:"fixed_in"`
	VerifiedBy   string    `json:"verified_by"`
	VerifiedDate time.Time `json:"verified_date"`
	Notes        []Note    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Note is a comment on a finding.
type Note struct {
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// Tracker manages audit findings.
type Tracker struct {
	Findings []*Finding `json:"findings"`
	path     string
}

// NewTracker loads or creates a tracker.
func NewTracker(path string) (*Tracker, error) {
	t := &Tracker{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return t, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return t, nil
	}

	if err := json.Unmarshal(data, t); err != nil {
		return nil, err
	}

	return t, nil
}

// Save persists the tracker.
func (t *Tracker) Save() error {
	if t.path == "" {
		return errors.New("tracker path not set")
	}

	if err := ensureDir(t.path); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(t.path, data, 0600)
}

// AddFinding adds a new finding.
func (t *Tracker) AddFinding(f *Finding) error {
	if f == nil {
		return errors.New("finding is nil")
	}
	if strings.TrimSpace(f.Title) == "" {
		return errors.New("finding title required")
	}
	if f.ID == "" {
		f.ID = generateFindingID()
	}

	now := time.Now().UTC()
	f.CreatedAt = now
	f.UpdatedAt = now
	if f.Status == "" {
		f.Status = StatusNew
	}

	t.Findings = append(t.Findings, f)
	return nil
}

// FindByID returns a finding by ID.
func (t *Tracker) FindByID(id string) (*Finding, error) {
	for _, f := range t.Findings {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, fmt.Errorf("finding %s not found", id)
}

// UpdateFinding updates a finding and refreshes timestamps.
func (t *Tracker) UpdateFinding(id string, updater func(*Finding) error) error {
	f, err := t.FindByID(id)
	if err != nil {
		return err
	}
	if err := updater(f); err != nil {
		return err
	}
	f.UpdatedAt = time.Now().UTC()
	return nil
}

// AddNote appends a note to a finding.
func (t *Tracker) AddNote(id, author, text string) error {
	return t.UpdateFinding(id, func(f *Finding) error {
		if strings.TrimSpace(text) == "" {
			return errors.New("note text required")
		}
		note := Note{
			Author:    strings.TrimSpace(author),
			Text:      strings.TrimSpace(text),
			Timestamp: time.Now().UTC(),
		}
		f.Notes = append(f.Notes, note)
		return nil
	})
}

// Summary generates a status summary.
func (t *Tracker) Summary() map[string]int {
	summary := make(map[string]int)

	for _, f := range t.Findings {
		key := fmt.Sprintf("%s_%s", f.Severity, f.Status)
		summary[key]++
		summary[string(f.Severity)]++
		summary[string(f.Status)]++
	}

	return summary
}

// UnresolvedCritical returns unresolved critical findings.
func (t *Tracker) UnresolvedCritical() []*Finding {
	var result []*Finding
	for _, f := range t.Findings {
		if f.Severity == SeverityCritical &&
			f.Status != StatusVerified &&
			f.Status != StatusWontFix &&
			f.Status != StatusFalsePos {
			result = append(result, f)
		}
	}
	return result
}

func generateFindingID() string {
	buf := make([]byte, 3)
	_, _ = rand.Read(buf)
	suffix := strings.ToUpper(hex.EncodeToString(buf))
	return fmt.Sprintf("AUD-%s-%s", time.Now().UTC().Format("20060102"), suffix)
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
