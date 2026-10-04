// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func quietLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).Level(zerolog.Disabled)
}

func drainEvents(ch chan *SecurityEvent) []*SecurityEvent {
	var out []*SecurityEvent
	for {
		select {
		case e := <-ch:
			out = append(out, e)
		default:
			return out
		}
	}
}

func hasEvent(events []*SecurityEvent, eventType string) bool {
	for _, e := range events {
		if e.Type == eventType {
			return true
		}
	}
	return false
}

// --- audit_log.go ---

func TestAuditLogSeverityHelpers_AllBranches(t *testing.T) {
	cases := []struct {
		sev  SecurityEventSeverity
		want string
	}{
		{SeverityCritical, "error"},
		{SeverityHigh, "warn"},
		{SeverityMedium, "warn"},
		{SeverityLow, "info"},
		{SeverityInfo, "info"},
		{SecurityEventSeverity("bogus"), "info"},
	}
	for _, tc := range cases {
		if got := severityToLogLevel(tc.sev); got != tc.want {
			t.Errorf("severityToLogLevel(%q) = %q, want %q", tc.sev, got, tc.want)
		}
	}

	if got := severityToZerologLevel(SeverityCritical); got != zerolog.ErrorLevel {
		t.Errorf("critical zerolog level = %v, want error", got)
	}
	if got := severityToZerologLevel(SeverityHigh); got != zerolog.WarnLevel {
		t.Errorf("high zerolog level = %v, want warn", got)
	}
	if got := severityToZerologLevel(SeverityMedium); got != zerolog.WarnLevel {
		t.Errorf("medium zerolog level = %v, want warn", got)
	}
	if got := severityToZerologLevel(SeverityLow); got != zerolog.InfoLevel {
		t.Errorf("low zerolog level = %v, want info", got)
	}
	if got := severityToZerologLevel(SecurityEventSeverity("weird")); got != zerolog.InfoLevel {
		t.Errorf("unknown zerolog level = %v, want info", got)
	}
}

func TestAuditLogWriteToFile_AllEntryKinds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	audit, err := NewAuditLog(path, quietLogger())
	if err != nil {
		t.Fatalf("NewAuditLog: %v", err)
	}

	audit.LogEvent(&SecurityEvent{
		ID:          "evt-1",
		Type:        "test_event",
		Severity:    SeverityHigh,
		Timestamp:   time.Unix(1700000000, 0).UTC(),
		Source:      "10.0.0.1",
		Description: "something happened",
		Metadata:    map[string]interface{}{"k": "v"},
	})
	audit.LogAlert(&SecurityAlert{
		ID:        "alt-1",
		EventID:   "evt-1",
		Type:      "test_alert",
		Severity:  SeverityCritical,
		Title:     "Alert",
		Timestamp: time.Unix(1700000001, 0).UTC(),
	})
	audit.LogIncidentAction("inc-1", "block_ip", "responder", "ok")
	audit.LogPlaybookExecution("pb-1", "inc-1", "completed", []string{"step1", "step2"}, 1500*time.Millisecond)
	if err := audit.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 audit lines, got %d: %q", len(lines), string(data))
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("event line not valid JSON: %v", err)
	}
	if entry["event_id"] != "evt-1" {
		t.Errorf("event_id = %v, want evt-1", entry["event_id"])
	}
	if entry["level"] != "warn" {
		t.Errorf("level = %v, want warn", entry["level"])
	}
	if entry["log_type"] != "security_audit" {
		t.Errorf("log_type = %v, want security_audit", entry["log_type"])
	}

	seen := map[string]bool{}
	for _, ln := range lines {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(ln), &m); err != nil {
			t.Fatalf("line not valid JSON: %v (%q)", err, ln)
		}
		seen[m["log_type"].(string)] = true
	}
	for _, want := range []string{"security_audit", "security_alert", "incident_action", "playbook_execution"} {
		if !seen[want] {
			t.Errorf("missing log_type %q in audit file", want)
		}
	}
}

func TestAuditLogClose_NoFile(t *testing.T) {
	audit, err := NewAuditLog("", quietLogger())
	if err != nil {
		t.Fatalf("NewAuditLog: %v", err)
	}
	if err := audit.Close(); err != nil {
		t.Errorf("Close with no file should be nil, got %v", err)
	}
}

func TestNewAuditLog_UnwritablePath(t *testing.T) {
	// A path whose parent directory does not exist must fail to open.
	bad := filepath.Join(t.TempDir(), "missing-dir", "audit.log")
	if _, err := NewAuditLog(bad, quietLogger()); err == nil {
		t.Fatal("expected error opening audit log in a non-existent directory")
	}
}

// --- security_metrics.go ---

func TestThreatLevelValue_AllSeverities(t *testing.T) {
	cases := []struct {
		sev  SecurityEventSeverity
		want float64
	}{
		{SeverityLow, 0},
		{SeverityMedium, 1},
		{SeverityHigh, 2},
		{SeverityCritical, 3},
		{SeverityInfo, 0},
		{SecurityEventSeverity("nope"), 0},
	}
	for _, tc := range cases {
		if got := ThreatLevelValue(tc.sev); got != tc.want {
			t.Errorf("ThreatLevelValue(%q) = %v, want %v", tc.sev, got, tc.want)
		}
	}
}

// --- incident_response.go ---

func TestIncidentResponderFindMatchingPlaybooks(t *testing.T) {
	ir, err := NewIncidentResponder("", quietLogger())
	if err != nil {
		t.Fatalf("NewIncidentResponder: %v", err)
	}

	// A non-matching trigger type yields no playbooks.
	unrelated := &SecurityIncident{ID: "i2", Type: "some_unrelated_type", Severity: SeverityCritical}
	if got := ir.findMatchingPlaybooks(unrelated); len(got) != 0 {
		t.Errorf("unrelated type matched %d playbooks, want 0", len(got))
	}

	// A playbook whose MinSeverity is more severe than the incident is skipped.
	// Severity orders by Rank, so "critical" outranks "high" here.
	ir.AddPlaybook(&Playbook{
		ID: "critical-only", Enabled: true, MinSeverity: SeverityCritical,
		TriggerTypes: []string{"rate_limit_breach"},
	})
	hit := &SecurityIncident{ID: "i3", Type: "rate_limit_breach", Severity: SeverityHigh}
	for _, pb := range ir.findMatchingPlaybooks(hit) {
		if pb.ID == "critical-only" {
			t.Error("playbook with a higher MinSeverity should be skipped")
		}
	}

	// ddos-response triggers on rate_limit_breach at SeverityHigh.
	matches := ir.findMatchingPlaybooks(hit)
	if len(matches) != 1 || matches[0].ID != "ddos-response" {
		t.Fatalf("expected exactly ddos-response, got %+v", matches)
	}

	// The same CRITICAL incident must also match, since critical >= high.
	crit := &SecurityIncident{ID: "i4", Type: "rate_limit_breach", Severity: SeverityCritical}
	var sawDDOS, sawCriticalOnly bool
	for _, pb := range ir.findMatchingPlaybooks(crit) {
		switch pb.ID {
		case "ddos-response":
			sawDDOS = true
		case "critical-only":
			sawCriticalOnly = true
		}
	}
	if !sawDDOS || !sawCriticalOnly {
		t.Errorf("critical incident matched ddos=%v critical-only=%v, want both",
			sawDDOS, sawCriticalOnly)
	}

	// disabled playbooks are skipped
	ir.DisablePlaybook("ddos-response")
	if pb, ok := ir.GetPlaybook("ddos-response"); !ok || pb.Enabled {
		t.Fatalf("DisablePlaybook did not clear Enabled: %+v", pb)
	}
	if got := ir.findMatchingPlaybooks(hit); len(got) != 0 {
		t.Errorf("disabled playbook still matched %d playbooks", len(got))
	}
	ir.EnablePlaybook("ddos-response")
	if pb, _ := ir.GetPlaybook("ddos-response"); !pb.Enabled {
		t.Fatal("EnablePlaybook did not set Enabled")
	}
	// enabling a non-existent playbook is a no-op, not a panic
	ir.EnablePlaybook("does-not-exist")
	ir.DisablePlaybook("does-not-exist")
}

// TestSeverityOrderingIsRankOrdered pins the severity ORDER. It used to be
// TestSeverityOrderingIsLexicographic_KnownDefect and asserted the opposite:
// SecurityEventSeverity is a `string`, so `<`/`>` sort alphabetically
// ("critical" < "high", "high" < "medium"), which meant findMatchingPlaybooks
// dropped CRITICAL incidents from every HIGH playbook and updateIncidents
// never escalated an incident. Severity now orders through Rank/AtLeast/Above;
// this test fails if any comparison site reverts to a raw string operator.
func TestSeverityOrderingIsRankOrdered(t *testing.T) {
	ir, _ := NewIncidentResponder("", quietLogger())

	// 1. A CRITICAL incident must match the playbooks that require HIGH --
	//    lexicographically "critical" < "high" skipped all of them.
	critical := &SecurityIncident{ID: "c", Type: "rate_limit_breach", Severity: SeverityCritical}
	got := ir.findMatchingPlaybooks(critical)
	if len(got) == 0 {
		t.Error("a critical rate_limit_breach incident matches no playbooks; " +
			"severity ordering has regressed to lexicographic")
	}
	sawDDOS := false
	for _, pb := range got {
		if pb.ID == "ddos-response" {
			sawDDOS = true
		}
	}
	if !sawDDOS {
		t.Errorf("critical incident did not match ddos-response (MinSeverity HIGH); got %+v", got)
	}

	// 2. An incident must escalate when a more severe event arrives. The map
	//    key is type+":"+source, which is how updateIncidents finds it.
	inc := &SecurityIncident{ID: "i", Type: "t", Severity: SeverityMedium}
	sm := &SecurityMonitor{
		config:          DefaultSecurityMonitorConfig(),
		metrics:         GetSecurityMetrics(),
		activeIncidents: map[string]*SecurityIncident{"t:s": inc},
	}
	sm.updateIncidents(&SecurityEvent{ID: "e", Type: "t", Severity: SeverityCritical, Source: "s"})
	if inc.Severity != SeverityCritical {
		t.Errorf("incident severity after a CRITICAL event = %q, want %q", inc.Severity, SeverityCritical)
	}

	// Escalation must not walk an incident back down either.
	sm.updateIncidents(&SecurityEvent{ID: "e2", Type: "t", Severity: SeverityLow, Source: "s"})
	if inc.Severity != SeverityCritical {
		t.Errorf("incident severity after a LOW event = %q, want it held at %q",
			inc.Severity, SeverityCritical)
	}

	// 3. A HIGH event must open an incident and raise an alert. Lexicographic
	//    ordering made "high" < "medium", so HIGH and CRITICAL events were
	//    silently dropped by both guards -- the most severe events in the
	//    system raised nothing at all.
	for _, sev := range []SecurityEventSeverity{SeverityMedium, SeverityHigh, SeverityCritical} {
		t.Run("opens-incident/"+string(sev), func(t *testing.T) {
			sm := newTestMonitor(t, nil)
			sm.handleEvent(&SecurityEvent{
				ID: "x", Type: "t", Severity: sev, Source: "s", Timestamp: time.Unix(1700000000, 0),
			})
			if got := len(sm.GetActiveIncidents()); got != 1 {
				t.Errorf("severity %q opened %d incidents, want 1", sev, got)
			}
		})
		// A separate monitor for the alert probe: handleEvent above already
		// consumed the cooldown/rate-limit budget for this type+source.
		sm := newTestMonitor(t, nil)
		if !sm.shouldAlert(&SecurityEvent{ID: "y", Type: "t", Severity: sev, Source: "s"}) {
			t.Errorf("severity %q raised no alert, want an alert", sev)
		}
	}

	// Info and low must stay below the alert/incident floor.
	for _, sev := range []SecurityEventSeverity{SeverityInfo, SeverityLow} {
		t.Run("below-floor/"+string(sev), func(t *testing.T) {
			sm := newTestMonitor(t, nil)
			sm.handleEvent(&SecurityEvent{
				ID: "x", Type: "t", Severity: sev, Source: "s", Timestamp: time.Unix(1700000000, 0),
			})
			if got := len(sm.GetActiveIncidents()); got != 0 {
				t.Errorf("severity %q opened %d incidents, want 0", sev, got)
			}
			if sm.shouldAlert(&SecurityEvent{ID: "y", Type: "t", Severity: sev, Source: "s"}) {
				t.Errorf("severity %q raised an alert, want none", sev)
			}
		})
	}
}

// TestSeverityRankHelpers pins the ordering helpers themselves, including the
// unknown-value case: an unrecognised severity must not outrank a known one.
func TestSeverityRankHelpers(t *testing.T) {
	order := []SecurityEventSeverity{
		SeverityInfo, SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical,
	}
	for i, lo := range order {
		if got := lo.Rank(); got != i {
			t.Errorf("Rank(%q) = %d, want %d", lo, got, i)
		}
		for j, hi := range order {
			if got := hi.AtLeast(lo); got != (j >= i) {
				t.Errorf("%q.AtLeast(%q) = %v, want %v", hi, lo, got, j >= i)
			}
			if got := hi.Above(lo); got != (j > i) {
				t.Errorf("%q.Above(%q) = %v, want %v", hi, lo, got, j > i)
			}
		}
	}

	// An unknown severity ranks below every known one, so it can never open an
	// incident, trigger a playbook or escalate one -- a corrupt value fails
	// safe. The converse holds too: a known severity outranks an unknown one.
	unknown := SecurityEventSeverity("not-a-severity")
	if got := unknown.Rank(); got != -1 {
		t.Errorf("Rank(unknown) = %d, want -1", got)
	}
	if unknown.AtLeast(unknown) != true || unknown.Above(unknown) != false {
		t.Error("unknown compared against itself must be equal, not strictly above")
	}
	for _, known := range order {
		if unknown.AtLeast(known) {
			t.Errorf("unknown severity AtLeast(%q) = true, want false -- "+
				"a corrupt severity must not escalate or trigger a playbook", known)
		}
		if unknown.Above(known) {
			t.Errorf("unknown severity Above(%q) = true, want false", known)
		}
		if !known.Above(unknown) {
			t.Errorf("%q.Above(unknown) = false, want true -- known must outrank unknown", known)
		}
	}

	// The wire format is unchanged: these are still the lowercase strings.
	for _, tc := range []struct {
		sev  SecurityEventSeverity
		want string
	}{
		{SeverityInfo, "info"}, {SeverityLow, "low"}, {SeverityMedium, "medium"},
		{SeverityHigh, "high"}, {SeverityCritical, "critical"},
	} {
		if string(tc.sev) != tc.want {
			t.Errorf("string(%v) = %q, want %q", tc.sev, string(tc.sev), tc.want)
		}
	}
}

func TestIncidentResponderExecutePlaybook_AllActions(t *testing.T) {
	ir, err := NewIncidentResponder("", quietLogger())
	if err != nil {
		t.Fatalf("NewIncidentResponder: %v", err)
	}

	incident := &SecurityIncident{
		ID:             "inc-actions",
		Type:           "rate_limit_breach",
		Severity:       SeverityCritical,
		AffectedAssets: []string{"1.2.3.4"},
	}

	actions := []PlaybookAction{
		ActionLogEvent,
		ActionSendAlert,
		ActionBlockIP,
		ActionRevokeKey,
		ActionSuspendAccount,
		ActionSuspendProvider,
		ActionNotifyTeam,
		ActionCollectEvidence,
		ActionEscalate,
	}
	for _, action := range actions {
		if err := ir.executeAction(context.Background(), action, map[string]string{"duration": "60", "channel": "c"}, incident); err != nil {
			t.Errorf("action %q returned error: %v", action, err)
		}
	}

	if err := ir.executeAction(context.Background(), PlaybookAction("nope"), nil, incident); err == nil {
		t.Error("expected error for unknown playbook action")
	}
}

func TestIncidentResponderActionIncreaseSeverity(t *testing.T) {
	ir, _ := NewIncidentResponder("", quietLogger())

	// actionIncreaseSeverity escalates one step below critical. It used to gate
	// on the lexicographic `incident.Severity < SeverityCritical`, which is true
	// for every value except "critical" itself, so the outer block never ran and
	// the inner switch was unreachable -- the action was a total no-op.
	cases := []struct {
		start SecurityEventSeverity
		want  SecurityEventSeverity
	}{
		{SeverityInfo, SeverityInfo}, // below the switch floor: unchanged
		{SeverityLow, SeverityMedium},
		{SeverityMedium, SeverityHigh},
		{SeverityHigh, SeverityCritical},
		{SeverityCritical, SeverityCritical}, // already maxed
	}
	for _, tc := range cases {
		inc := &SecurityIncident{ID: "x", Severity: tc.start}
		if err := ir.actionIncreaseSeverity(inc); err != nil {
			t.Fatalf("actionIncreaseSeverity: %v", err)
		}
		if inc.Severity != tc.want {
			t.Errorf("severity %q after actionIncreaseSeverity = %q, want %q",
				tc.start, inc.Severity, tc.want)
		}
	}
}

func TestIncidentResponderExecutePlaybook_StepOutcomes(t *testing.T) {
	ir, _ := NewIncidentResponder("", quietLogger())
	incident := &SecurityIncident{ID: "inc-1", Type: "test", Severity: SeverityHigh}

	// A step with a failing (unknown) action and ContinueOnFailure=false
	// must abort the playbook with status "failed".
	failing := &Playbook{
		ID:    "pb-fail",
		Steps: []PlaybookStep{{Name: "boom", Action: "not_a_real_action", Timeout: 1}},
	}
	ir.executePlaybook(context.Background(), failing, incident)
	// NOTE: executePlaybook deliberately does NOT write incident.PlaybookID.
	// The incident is caller-owned and this runs on a background goroutine;
	// writing there raced with the caller's own reads. HandleIncident sets
	// PlaybookID synchronously instead. See TestIncidentResponderHandleIncident.
	if incident.PlaybookID != "" {
		t.Errorf("executePlaybook must not mutate caller-owned incident; PlaybookID = %q", incident.PlaybookID)
	}

	// A playbook whose first step fails but sets ContinueOnFailure keeps going.
	resilient := &Playbook{
		ID: "pb-continue",
		Steps: []PlaybookStep{
			{Name: "boom", Action: "not_a_real_action", Timeout: 1, ContinueOnFailure: true},
			{Name: "ok", Action: string(ActionLogEvent), Timeout: 1},
		},
	}
	stepExec := ir.executeStep(context.Background(), resilient.Steps[0], incident)
	if stepExec.Success {
		t.Error("expected unknown action step to fail")
	}
	if stepExec.Error == "" {
		t.Error("failed step should carry an error string")
	}
	if stepExec.CompletedAt == nil {
		t.Error("step execution should record CompletedAt")
	}

	okExec := ir.executeStep(context.Background(), resilient.Steps[1], incident)
	if !okExec.Success || okExec.Output != "success" {
		t.Errorf("expected successful step, got %+v", okExec)
	}
}

func TestIncidentResponderHandleIncident(t *testing.T) {
	ir, _ := NewIncidentResponder("", quietLogger())

	// No matching playbook -> returns without doing anything observable.
	ir.HandleIncident(context.Background(), &SecurityIncident{
		ID: "inc-none", Type: "unrelated", Severity: SeverityCritical,
	})

	// Matching playbook: a SeverityHigh rate_limit_breach selects
	// ddos-response. (SeverityCritical would NOT, because of the
	// lexicographic severity ordering documented in
	// TestSeverityOrderingIsLexicographic_KnownDefect.)
	incident := &SecurityIncident{
		ID: "inc-hit", Type: "rate_limit_breach", Severity: SeverityHigh,
		AffectedAssets: []string{"9.9.9.9"},
	}
	ir.HandleIncident(context.Background(), incident)

	// PlaybookID is set synchronously by HandleIncident (before the
	// background playbook goroutines are spawned), so it is readable as soon
	// as HandleIncident returns -- no polling and no race.
	if incident.PlaybookID != "ddos-response" {
		t.Errorf("incident.PlaybookID = %q, want ddos-response", incident.PlaybookID)
	}
}

func TestIncidentResponderAddRemovePlaybook(t *testing.T) {
	ir, _ := NewIncidentResponder("", quietLogger())

	before := len(ir.GetAllPlaybooks())
	ir.AddPlaybook(&Playbook{ID: "custom", Name: "Custom", Enabled: true})
	if after := len(ir.GetAllPlaybooks()); after != before+1 {
		t.Errorf("AddPlaybook: got %d playbooks, want %d", after, before+1)
	}
	if pb, ok := ir.GetPlaybook("custom"); !ok || pb.Name != "Custom" {
		t.Fatalf("GetPlaybook(custom) = %+v, %v", pb, ok)
	}
	ir.RemovePlaybook("custom")
	if _, ok := ir.GetPlaybook("custom"); ok {
		t.Error("RemovePlaybook did not remove the playbook")
	}
}

// --- security_monitor.go ---

func newTestMonitor(t *testing.T, tweak func(*SecurityMonitorConfig)) *SecurityMonitor {
	t.Helper()
	cfg := DefaultSecurityMonitorConfig()
	cfg.EnableAutoResponse = false
	if tweak != nil {
		tweak(cfg)
	}
	sm, err := NewSecurityMonitor(cfg, quietLogger())
	if err != nil {
		t.Fatalf("NewSecurityMonitor: %v", err)
	}
	t.Cleanup(sm.Stop)
	return sm
}

func TestNewSecurityMonitor_NilConfig(t *testing.T) {
	sm, err := NewSecurityMonitor(nil, quietLogger())
	if err != nil {
		t.Fatalf("NewSecurityMonitor(nil): %v", err)
	}
	defer sm.Stop()

	if sm.config == nil {
		t.Fatal("nil config should fall back to defaults")
	}
	if !sm.config.EnableTransactionDetector || !sm.config.EnableFraudDetector {
		t.Error("default config should enable detectors")
	}
	if sm.txDetector == nil || sm.fraudDetector == nil ||
		sm.cryptoDetector == nil || sm.providerSec == nil {
		t.Error("all detectors should be constructed from the default config")
	}
	if sm.incidentResponder == nil {
		t.Error("auto response should be enabled by default")
	}
}

func TestNewSecurityMonitor_DisabledDetectorsAreNil(t *testing.T) {
	cfg := DefaultSecurityMonitorConfig()
	cfg.EnableTransactionDetector = false
	cfg.EnableFraudDetector = false
	cfg.EnableCryptoAnomalyDetector = false
	cfg.EnableProviderSecurity = false
	cfg.EnableAutoResponse = false

	sm, err := NewSecurityMonitor(cfg, quietLogger())
	if err != nil {
		t.Fatalf("NewSecurityMonitor: %v", err)
	}
	defer sm.Stop()

	if sm.txDetector != nil || sm.fraudDetector != nil ||
		sm.cryptoDetector != nil || sm.providerSec != nil {
		t.Error("disabled detectors must not be constructed")
	}

	// Record* must be safe no-ops when the detector is absent.
	sm.RecordTransactionEvent(&TransactionData{Sender: "a"})
	sm.RecordVEIDVerification(&VEIDVerificationData{AccountAddress: "a"})
	sm.RecordCryptoOperation(&CryptoOperationData{AccountAddress: "a"})
	sm.RecordProviderActivity(&ProviderActivityData{ProviderID: "p"})
}

func TestNewSecurityMonitor_BadAuditLogPath(t *testing.T) {
	cfg := DefaultSecurityMonitorConfig()
	cfg.AuditLogPath = filepath.Join(t.TempDir(), "nope", "audit.log")
	cfg.EnableAutoResponse = false

	if _, err := NewSecurityMonitor(cfg, quietLogger()); err == nil {
		t.Fatal("expected error when the audit log path is not creatable")
	}
}

func TestSecurityMonitorHandleEvent_IncidentLifecycle(t *testing.T) {
	sm := newTestMonitor(t, nil)

	base := time.Unix(1700000000, 0)
	low := &SecurityEvent{ID: "e1", Type: "noise", Severity: SeverityLow, Timestamp: base, Source: "s1"}

	// Low severity must not open an incident.
	sm.handleEvent(low)
	if got := sm.GetActiveIncidents(); len(got) != 0 {
		t.Errorf("low severity event created %d incidents, want 0", len(got))
	}
	// Low severity must not alert.
	if sm.shouldAlert(low) {
		t.Error("low severity event must not raise an alert")
	}

	med := &SecurityEvent{ID: "e2", Type: "attack", Severity: SeverityMedium, Timestamp: base, Source: "s1", Description: "d"}
	sm.handleEvent(med)
	incidents := sm.GetActiveIncidents()
	if len(incidents) != 1 {
		t.Fatalf("expected 1 incident after medium event, got %d", len(incidents))
	}
	inc := incidents[0]
	if inc.Type != "attack" {
		t.Fatalf("unexpected incident %+v", inc)
	}
	if inc.EventCount != 1 || inc.Status != "open" {
		t.Errorf("incident count/status = %d/%q, want 1/open", inc.EventCount, inc.Status)
	}

	// Second event of the same type+source updates the existing incident
	// rather than opening a second one, and escalates it because high is more
	// severe than medium (see TestSeverityOrderingIsRankOrdered).
	med2 := &SecurityEvent{ID: "e3", Type: "attack", Severity: SeverityHigh, Timestamp: base.Add(time.Second), Source: "s1"}
	sm.handleEvent(med2)
	incidents = sm.GetActiveIncidents()
	if len(incidents) != 1 {
		t.Fatalf("expected incident to be reused, got %d", len(incidents))
	}
	inc = incidents[0]
	if inc.EventCount != 2 {
		t.Errorf("EventCount = %d, want 2", inc.EventCount)
	}
	if !inc.LastEventAt.Equal(base.Add(time.Second)) {
		t.Errorf("LastEventAt = %v, want the second event's timestamp", inc.LastEventAt)
	}

	if err := sm.ResolveIncident(inc.ID); err != nil {
		t.Fatalf("ResolveIncident: %v", err)
	}
	if sm.GetActiveIncidents()[0].Status != "resolved" {
		t.Error("ResolveIncident should mark the incident resolved")
	}
	// Resolving an unknown ID is a silent no-op.
	if err := sm.ResolveIncident("no-such-incident"); err != nil {
		t.Errorf("ResolveIncident(unknown) = %v, want nil", err)
	}
}

func TestSecurityMonitorUpdateThreatLevel_Thresholds(t *testing.T) {
	t.Run("no incidents -> low", func(t *testing.T) {
		sm := newTestMonitor(t, nil)
		sm.updateThreatLevel()
	})

	t.Run("only low/medium incidents -> medium", func(t *testing.T) {
		sm := newTestMonitor(t, nil)
		sm.activeIncidents["k"] = &SecurityIncident{ID: "a", Severity: SeverityMedium}
		sm.updateThreatLevel()
	})

	t.Run("one high incident -> high", func(t *testing.T) {
		sm := newTestMonitor(t, nil)
		sm.activeIncidents["k"] = &SecurityIncident{ID: "a", Severity: SeverityHigh}
		sm.updateThreatLevel()
	})

	t.Run("critical incident -> critical", func(t *testing.T) {
		sm := newTestMonitor(t, nil)
		sm.activeIncidents["k"] = &SecurityIncident{ID: "a", Severity: SeverityCritical}
		sm.updateThreatLevel()
	})

	t.Run("score floors at zero", func(t *testing.T) {
		sm := newTestMonitor(t, nil)
		for i := 0; i < 40; i++ {
			sm.activeIncidents[string(rune('a'+i%26))+string(rune(i))] = &SecurityIncident{ID: "x", Severity: SeverityCritical}
		}
		sm.updateThreatLevel() // must not panic and must clamp
	})
}

func TestSecurityMonitorShouldAlert_CooldownAndRateLimit(t *testing.T) {
	sm := newTestMonitor(t, func(c *SecurityMonitorConfig) {
		c.AlertCooldownSecs = 60
		c.MaxAlertsPerMinute = 2
	})
	ev := &SecurityEvent{ID: "e", Type: "t", Severity: SeverityMedium, Source: "s"}

	if sm.shouldAlert(ev) != true {
		t.Fatal("first alert should be allowed")
	}
	if sm.shouldAlert(ev) != false {
		t.Fatal("second alert inside the cooldown window should be suppressed")
	}

	// Backdate the last alert so the cooldown has elapsed, but keep the
	// counter over the rate limit so the rate-limit branch is what rejects.
	sm.mu.Lock()
	sm.lastAlertTimes["t:s"] = time.Now().Add(-2 * time.Minute)
	sm.recentAlertCounts["t:s"] = 5
	sm.mu.Unlock()
	if sm.shouldAlert(ev) != false {
		t.Fatal("alert should be suppressed by the per-minute rate limit")
	}

	// Below medium severity never alerts, even with no cooldown state.
	sm2 := newTestMonitor(t, nil)
	low := &SecurityEvent{ID: "e", Type: "t", Severity: SeverityInfo, Source: "s"}
	if sm2.shouldAlert(low) {
		t.Error("info severity must not alert")
	}
}

func TestSecurityMonitorSendAlert_DropsWhenChannelFull(t *testing.T) {
	sm := newTestMonitor(t, nil)
	sm.alertChan = make(chan *SecurityAlert, 1)
	ev := &SecurityEvent{ID: "e1", Type: "t", Severity: SeverityHigh, Source: "s"}

	sm.sendAlert(ev)
	sm.sendAlert(ev) // channel is full: must be dropped, not block

	if len(sm.alertChan) != 1 {
		t.Errorf("alert channel length = %d, want 1 (second alert dropped)", len(sm.alertChan))
	}
	got := <-sm.alertChan
	if got.EventID != "e1" || got.Type != "t" {
		t.Errorf("unexpected alert %+v", got)
	}
	if got.Title != "Security Alert: t" {
		t.Errorf("alert title = %q", got.Title)
	}
}

func TestSecurityMonitorGetSeverityLogLevel(t *testing.T) {
	sm := newTestMonitor(t, nil)
	cases := []struct {
		sev  SecurityEventSeverity
		want zerolog.Level
	}{
		{SeverityCritical, zerolog.ErrorLevel},
		{SeverityHigh, zerolog.WarnLevel},
		{SeverityMedium, zerolog.WarnLevel},
		{SeverityLow, zerolog.InfoLevel},
		{SeverityInfo, zerolog.InfoLevel},
	}
	for _, tc := range cases {
		if got := sm.getSeverityLogLevel(tc.sev); got != tc.want {
			t.Errorf("getSeverityLogLevel(%q) = %v, want %v", tc.sev, got, tc.want)
		}
	}
}

func TestSecurityMonitorHandleAlert_WithAndWithoutWebhook(t *testing.T) {
	sm := newTestMonitor(t, func(c *SecurityMonitorConfig) { c.AlertWebhookURL = "" })
	alert := &SecurityAlert{ID: "a1", EventID: "e1", Type: "t", Severity: SeverityCritical, Source: "s", Title: "T"}
	sm.handleAlert(alert) // no webhook configured

	sm.config.AlertWebhookURL = "https://example.invalid/hook"
	sm.handleAlert(alert) // webhook path spawns a goroutine
	sm.sendAlertWebhook(alert)
}

func TestSecurityMonitorUpdateMetricsAndCleanup(t *testing.T) {
	sm := newTestMonitor(t, nil)
	old := time.Now().Add(-2 * time.Hour)

	sm.activeIncidents["stale"] = &SecurityIncident{
		ID: "stale", Status: "resolved", LastEventAt: old,
	}
	sm.activeIncidents["fresh"] = &SecurityIncident{
		ID: "fresh", Status: "resolved", LastEventAt: time.Now(),
	}
	sm.activeIncidents["open"] = &SecurityIncident{
		ID: "open", Status: "open", LastEventAt: old,
	}
	sm.recentAlertCounts["k"] = 42

	sm.updateMetrics()
	sm.cleanup()

	if _, ok := sm.activeIncidents["stale"]; ok {
		t.Error("cleanup should drop resolved incidents older than 1 hour")
	}
	if _, ok := sm.activeIncidents["fresh"]; !ok {
		t.Error("cleanup should keep recently resolved incidents")
	}
	if _, ok := sm.activeIncidents["open"]; !ok {
		t.Error("cleanup must never drop an open incident")
	}
	if len(sm.recentAlertCounts) != 0 {
		t.Error("cleanup should reset the alert rate counters")
	}
}

func TestSecurityMonitorRecordRateLimitBreach(t *testing.T) {
	sm := newTestMonitor(t, nil)

	sm.RecordRateLimitBreach(&RateLimitBreachData{
		LimitType:     "per_minute",
		SourceIP:      "1.2.3.4",
		CurrentCount:  120,
		Limit:         100,
		BypassAttempt: true,
		Description:   "over limit",
		Severity:      SeverityHigh,
	})

	ev := <-sm.eventChan
	if ev.Type != "rate_limit_breach" {
		t.Errorf("event type = %q, want rate_limit_breach", ev.Type)
	}
	if ev.Source != "1.2.3.4" {
		t.Errorf("event source = %q, want 1.2.3.4", ev.Source)
	}
	if ev.Metadata["limit_type"] != "per_minute" {
		t.Errorf("metadata limit_type = %v", ev.Metadata["limit_type"])
	}
	if ev.Metadata["bypass_attempt"] != true {
		t.Errorf("metadata bypass_attempt = %v, want true", ev.Metadata["bypass_attempt"])
	}
}

func TestSecurityMonitorRecordAuthFailure(t *testing.T) {
	sm := newTestMonitor(t, nil)
	sm.RecordAuthFailure("bad_signature", "10.0.0.1")

	ev := <-sm.eventChan
	if ev.Type != "auth_failure" {
		t.Errorf("event type = %q, want auth_failure", ev.Type)
	}
	if ev.Severity != SeverityMedium {
		t.Errorf("auth failure severity = %q, want medium", ev.Severity)
	}
	if ev.Description != "Authentication failure: bad_signature" {
		t.Errorf("description = %q", ev.Description)
	}
	if ev.Metadata["reason"] != "bad_signature" {
		t.Errorf("metadata reason = %v", ev.Metadata["reason"])
	}
}

func TestSecurityMonitorRecordBridges_Detectors(t *testing.T) {
	sm := newTestMonitor(t, nil)

	// Wire the detectors' event channel to the monitor's. The monitor's own
	// eventProcessor is deliberately not started, so nothing drains the
	// channel and the assertions below are race-free.
	sm.txDetector.Start(sm.ctx, sm.eventChan)
	sm.fraudDetector.Start(sm.ctx, sm.eventChan)
	sm.cryptoDetector.Start(sm.ctx, sm.eventChan)
	sm.providerSec.Start(sm.ctx, sm.eventChan)

	sm.RecordTransactionEvent(&TransactionData{
		TxHash: "hash-1", Sender: "acct", Amount: 2_000_000, Timestamp: time.Now(), Success: true,
	})
	sm.RecordVEIDVerification(&VEIDVerificationData{
		AccountAddress: "acct", Timestamp: time.Now(), ComputedScore: 10,
		OCRConfidence: 0.1, ReasonCodes: []string{"DOCUMENT_TAMPERED"},
	})
	sm.RecordCryptoOperation(&CryptoOperationData{
		OperationID: "op1", OperationType: "keygen", Algorithm: "MD5",
		AccountAddress: "acct", Timestamp: time.Now(), Success: true, EntropyScore: 0.5,
	})
	sm.RecordProviderActivity(&ProviderActivityData{
		ProviderID: "prov", ActivityType: "bid", Timestamp: time.Now(), BidAmount: 0.001,
	})

	events := drainEvents(sm.eventChan)
	if len(events) == 0 {
		t.Fatal("expected the detector bridges to emit security events")
	}
	for _, want := range []string{"tx_value_anomaly", "document_tampering", "deprecated_algorithm", "bid_manipulation"} {
		if !hasEvent(events, want) {
			t.Errorf("missing expected event type %q; got %d events", want, len(events))
		}
	}
}

func TestSecurityMonitorEventChannelFull_DropsEvent(t *testing.T) {
	sm := newTestMonitor(t, nil)
	sm.eventChan = make(chan *SecurityEvent, 1)
	sm.eventChan <- &SecurityEvent{ID: "filler"}

	sm.RecordRateLimitBreach(&RateLimitBreachData{LimitType: "x", Severity: SeverityLow})
	sm.RecordAuthFailure("reason", "src")

	if len(sm.eventChan) != 1 {
		t.Errorf("event channel length = %d, want 1 (overflow dropped)", len(sm.eventChan))
	}
}

func TestSecurityMonitorStartStop_WithAutoResponse(t *testing.T) {
	cfg := DefaultSecurityMonitorConfig()
	cfg.EnableAutoResponse = true
	cfg.MetricsIntervalSecs = 3600
	cfg.CleanupIntervalSecs = 3600
	sm, err := NewSecurityMonitor(cfg, quietLogger())
	if err != nil {
		t.Fatalf("NewSecurityMonitor: %v", err)
	}
	if err := sm.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	sm.RecordAuthFailure("probe", "10.0.0.9")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sm.mu.RLock()
		n := len(sm.activeIncidents)
		sm.mu.RUnlock()
		if n > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	sm.mu.RLock()
	n := len(sm.activeIncidents)
	sm.mu.RUnlock()
	if n == 0 {
		t.Error("running monitor should have processed the auth failure into an incident")
	}

	sm.Stop()
	// Stop is idempotent for the audit log; a second Stop must not panic.
	sm.Stop()
}

// --- transaction_detector.go ---

func TestCalculateStats(t *testing.T) {
	if m, sd := calculateStats(nil); m != 0 || sd != 0 {
		t.Errorf("calculateStats(nil) = %v, %v; want 0, 0", m, sd)
	}
	if m, sd := calculateStats([]float64{5}); m != 5 || sd != 0 {
		t.Errorf("calculateStats([5]) = %v, %v; want 5, 0", m, sd)
	}
	m, sd := calculateStats([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if m != 5 {
		t.Errorf("mean = %v, want 5", m)
	}
	if sd <= 0 {
		t.Errorf("stdDev = %v, want > 0", sd)
	}
}

func TestTransactionDetectorValueAnomaly_Statistical(t *testing.T) {
	cfg := DefaultTransactionDetectorConfig()
	d := NewTransactionDetector(cfg, quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	// Build a stable history of ten identical amounts, then a huge outlier.
	for i := 0; i < 10; i++ {
		d.Analyze(&TransactionData{
			TxHash:    "stable-" + string(rune('a'+i)),
			Sender:    "acct",
			Recipient: "r",
			Amount:    100,
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}
	drainEvents(events)

	d.Analyze(&TransactionData{
		TxHash: "outlier", Sender: "acct", Recipient: "r",
		Amount: 100000, Timestamp: now.Add(11 * time.Second),
	})

	got := drainEvents(events)
	if !hasEvent(got, "tx_statistical_anomaly") {
		t.Errorf("expected a statistical anomaly event, got %d events", len(got))
	}
}

func TestTransactionDetectorHighRiskAccount(t *testing.T) {
	cfg := DefaultTransactionDetectorConfig()
	cfg.HighRiskAccountTxLimit = 3
	d := NewTransactionDetector(cfg, quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	for i := 0; i < 6; i++ {
		d.Analyze(&TransactionData{
			TxHash:            "hr-" + string(rune('a'+i)),
			Sender:            "risky",
			Recipient:         "r",
			Amount:            1,
			Timestamp:         now.Add(time.Duration(i) * time.Second),
			IsHighRiskAccount: true,
		})
	}
	if !hasEvent(drainEvents(events), "tx_high_risk_account") {
		t.Error("expected high-risk account limit event")
	}
}

func TestTransactionDetectorNewAccountAbuseAndSplit(t *testing.T) {
	cfg := DefaultTransactionDetectorConfig()
	cfg.SplitTransactionThreshold = 4
	// New accounts get MaxTxPerHourPerAccount/4, so lower the base limit to
	// make that derived threshold reachable in a short test.
	cfg.MaxTxPerHourPerAccount = 8
	d := NewTransactionDetector(cfg, quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	for i := 0; i < 6; i++ {
		d.Analyze(&TransactionData{
			TxHash:      "new-" + string(rune('a'+i)),
			Sender:      "fresh",
			Recipient:   "same",
			Amount:      10,
			Timestamp:   now.Add(time.Duration(i) * time.Second),
			AccountAge:  5 * time.Minute, // inside the new-account cooldown
			MsgType:     "send",
			BlockHeight: int64(100 + i),
		})
	}
	got := drainEvents(events)
	if !hasEvent(got, "tx_split_pattern") {
		t.Error("expected a split-transaction pattern event")
	}
	if !hasEvent(got, "tx_new_account_abuse") {
		t.Error("expected a new-account abuse event")
	}
	if !hasEvent(got, "tx_rapid_fire") {
		t.Error("expected a rapid-fire pattern event")
	}
}

// --- fraud_detector.go ---

func TestAbs32(t *testing.T) {
	if got := abs32(5); got != 5 {
		t.Errorf("abs32(5) = %d", got)
	}
	if got := abs32(-5); got != 5 {
		t.Errorf("abs32(-5) = %d", got)
	}
	if got := abs32(0); got != 0 {
		t.Errorf("abs32(0) = %d", got)
	}
}

func TestHashData(t *testing.T) {
	// Reserved helper: verify it is a stable SHA-256 hex digest.
	a := hashData([]byte("payload"))
	b := hashData([]byte("payload"))
	if a != b {
		t.Error("hashData must be deterministic")
	}
	if a == hashData([]byte("other")) {
		t.Error("hashData must distinguish inputs")
	}
	if len(a) != 64 {
		t.Errorf("hashData length = %d, want 64 hex chars", len(a))
	}
}

func TestFraudDetectorReplayAndBiometricReuse(t *testing.T) {
	d := NewFraudDetector(DefaultFraudDetectorConfig(), quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	scope := strings.Repeat("a", 32)
	biometric := strings.Repeat("b", 32)

	// First account registers the scope hash and biometric.
	d.Analyze(&VEIDVerificationData{
		RequestID: "r1", AccountAddress: "acct1", Timestamp: now,
		ComputedScore: 90, Success: true, ScopeHashes: []string{scope}, BiometricHash: biometric,
	})
	if got := drainEvents(events); len(got) != 0 {
		t.Errorf("first use should be clean, got %d events", len(got))
	}

	// A different account reusing both is a replay attack AND a
	// multi-identity biometric hit.
	d.Analyze(&VEIDVerificationData{
		RequestID: "r2", AccountAddress: "acct2", Timestamp: now.Add(time.Second),
		ComputedScore: 90, Success: true, ScopeHashes: []string{scope}, BiometricHash: biometric,
	})
	got := drainEvents(events)
	if !hasEvent(got, string(FraudIndicatorReplayAttack)) {
		t.Error("expected a replay attack event")
	}
	if !hasEvent(got, string(FraudIndicatorMultipleIdentities)) {
		t.Error("expected a multiple-identities event from the shared biometric")
	}
}

func TestFraudDetectorRepeatedScopeSubmission(t *testing.T) {
	d := NewFraudDetector(DefaultFraudDetectorConfig(), quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	scope := strings.Repeat("c", 32)
	// checkReplayAttack flags a repeat submission only once the stored
	// seenCount exceeds 3, i.e. on the 5th submission of the same scope by
	// the same account (1,2,3,4 are absorbed silently).
	for i := 0; i < 5; i++ {
		d.Analyze(&VEIDVerificationData{
			RequestID: "r", AccountAddress: "acct1",
			Timestamp:     now.Add(time.Duration(i) * time.Second),
			ComputedScore: 90, Success: true, ScopeHashes: []string{scope},
		})
	}
	if !hasEvent(drainEvents(events), string(FraudIndicatorReplayAttack)) {
		t.Error("expected a repeated-scope-submission replay event")
	}

	// Fewer than five submissions must stay silent.
	d2 := NewFraudDetector(DefaultFraudDetectorConfig(), quietLogger())
	events2 := make(chan *SecurityEvent, 100)
	d2.Start(context.Background(), events2)
	for i := 0; i < 4; i++ {
		d2.Analyze(&VEIDVerificationData{
			RequestID: "r", AccountAddress: "acct1",
			Timestamp:     now.Add(time.Duration(i) * time.Second),
			ComputedScore: 90, Success: true, ScopeHashes: []string{scope},
		})
	}
	if hasEvent(drainEvents(events2), string(FraudIndicatorReplayAttack)) {
		t.Error("four submissions must not trip the repeated-scope branch")
	}
}

func TestFraudDetectorScoreAnomalies(t *testing.T) {
	d := NewFraudDetector(DefaultFraudDetectorConfig(), quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	// Three consecutive low scores trip the consecutive-low branch.
	for i := 0; i < 3; i++ {
		d.Analyze(&VEIDVerificationData{
			RequestID: "r", AccountAddress: "acct1",
			Timestamp:     now.Add(time.Duration(i) * time.Second),
			ComputedScore: 5, Success: false, FailureReason: "low_score",
		})
	}
	got := drainEvents(events)
	if !hasEvent(got, string(FraudIndicatorScoreAnomaly)) {
		t.Error("expected a consecutive-low-score anomaly event")
	}
	if !hasEvent(got, string(FraudIndicatorSuspiciousBehavior)) {
		t.Error("expected a multiple-failures suspicious-behavior event")
	}

	// A large proposer/computed variance on a mismatched score.
	d2 := NewFraudDetector(DefaultFraudDetectorConfig(), quietLogger())
	events2 := make(chan *SecurityEvent, 100)
	d2.Start(context.Background(), events2)
	d2.Analyze(&VEIDVerificationData{
		RequestID: "r2", AccountAddress: "acct2", Timestamp: now,
		ProposerScore: 10, ComputedScore: 90, ScoreDifference: -80, Match: false,
	})
	if !hasEvent(drainEvents(events2), string(FraudIndicatorScoreAnomaly)) {
		t.Error("expected a score-variance anomaly event")
	}
}

func TestFraudDetectorDocumentAndLiveness(t *testing.T) {
	d := NewFraudDetector(DefaultFraudDetectorConfig(), quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	docHash := strings.Repeat("d", 32)
	d.Analyze(&VEIDVerificationData{
		RequestID: "r1", AccountAddress: "acct1", Timestamp: now,
		ComputedScore: 90, Success: true, DocumentHash: docHash,
		DocumentType: "passport", OCRConfidence: 0.4, LivenessScore: 0.2,
	})
	got := drainEvents(events)
	if !hasEvent(got, string(FraudIndicatorDocumentForgery)) {
		t.Error("expected a low-OCR-confidence document forgery event")
	}
	if !hasEvent(got, string(FraudIndicatorLivenessFailure)) {
		t.Error("expected a liveness failure event")
	}

	// Same document, different account -> synthetic identity.
	d.Analyze(&VEIDVerificationData{
		RequestID: "r2", AccountAddress: "acct2", Timestamp: now.Add(time.Second),
		ComputedScore: 90, Success: true, DocumentHash: docHash, DocumentType: "passport",
	})
	if !hasEvent(drainEvents(events), string(FraudIndicatorSyntheticIdentity)) {
		t.Error("expected a synthetic-identity event")
	}
}

func TestFraudDetectorNilConfigAndNilEventChan(t *testing.T) {
	// nil config falls back to defaults; no eventChan set means emitEvent
	// must drop rather than block or panic.
	d := NewFraudDetector(nil, quietLogger())
	if d.config == nil || d.config.MaxVerificationAttemptsPerAccount != 5 {
		t.Fatalf("nil config did not fall back to defaults: %+v", d.config)
	}
	d.Analyze(&VEIDVerificationData{
		RequestID: "r", AccountAddress: "a", Timestamp: time.Now(),
		ComputedScore: 1, Success: false, FaceSimilarityScore: 0.1,
		LivenessScore: 0.1, OCRConfidence: 0.1,
		BiometricHash: strings.Repeat("e", 32),
	})
}

// --- crypto_anomaly.go ---

func TestCryptoDetectorKeyReuseAcrossAccounts(t *testing.T) {
	d := NewCryptoAnomalyDetector(DefaultCryptoAnomalyConfig(), quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	now := time.Now()
	fp := strings.Repeat("f", 32)
	d.Analyze(&CryptoOperationData{
		OperationID: "o1", OperationType: "sign", Algorithm: "Ed25519",
		KeyFingerprint: fp, AccountAddress: "acct1", Timestamp: time.Now(), Success: true,
	})
	if got := drainEvents(events); len(got) != 0 {
		t.Errorf("first key use should be clean, got %d events", len(got))
	}

	d.Analyze(&CryptoOperationData{
		OperationID: "o2", OperationType: "sign", Algorithm: "Ed25519",
		KeyFingerprint: fp, AccountAddress: "acct2", Timestamp: time.Now(), Success: true,
	})
	if !hasEvent(drainEvents(events), string(CryptoAnomalyKeyReuse)) {
		t.Error("expected a cross-account key reuse event")
	}

	// Same account reusing its own key twice must NOT be reported: only the
	// cross-account case is an anomaly (the useCount is still incremented).
	d2 := NewCryptoAnomalyDetector(DefaultCryptoAnomalyConfig(), quietLogger())
	events2 := make(chan *SecurityEvent, 100)
	d2.Start(context.Background(), events2)
	own := strings.Repeat("9", 32)
	for i := 0; i < 2; i++ {
		d2.Analyze(&CryptoOperationData{
			OperationID: "own" + string(rune('a'+i)), OperationType: "sign", Algorithm: "Ed25519",
			KeyFingerprint: own, AccountAddress: "acct1", Timestamp: now, Success: true,
		})
	}
	if got := drainEvents(events2); hasEvent(got, string(CryptoAnomalyKeyReuse)) {
		t.Error("same-account key reuse must not be reported as cross-account reuse")
	}

	// A different account using it is the anomaly.
	d2.Analyze(&CryptoOperationData{
		OperationID: "other", OperationType: "sign", Algorithm: "Ed25519",
		KeyFingerprint: own, AccountAddress: "acct2", Timestamp: now, Success: true,
	})
	if !hasEvent(drainEvents(events2), string(CryptoAnomalyKeyReuse)) {
		t.Error("expected a cross-account key reuse event")
	}

	// Key-reuse detection disabled: a cross-account reuse is silent.
	cfg := DefaultCryptoAnomalyConfig()
	cfg.EnableKeyReuseDetection = false
	d3 := NewCryptoAnomalyDetector(cfg, quietLogger())
	events3 := make(chan *SecurityEvent, 100)
	d3.Start(context.Background(), events3)
	d3.Analyze(&CryptoOperationData{
		OperationID: "a", OperationType: "sign", Algorithm: "Ed25519",
		KeyFingerprint: own, AccountAddress: "acct1", Timestamp: now, Success: true,
	})
	d3.Analyze(&CryptoOperationData{
		OperationID: "b", OperationType: "sign", Algorithm: "Ed25519",
		KeyFingerprint: own, AccountAddress: "acct2", Timestamp: now, Success: true,
	})
	if hasEvent(drainEvents(events3), string(CryptoAnomalyKeyReuse)) {
		t.Error("key reuse detection is disabled; no event expected")
	}
}

func TestCryptoDetectorRecordHelpers(t *testing.T) {
	d := NewCryptoAnomalyDetector(DefaultCryptoAnomalyConfig(), quietLogger())
	events := make(chan *SecurityEvent, 500)
	d.Start(context.Background(), events)

	d.RecordSignatureVerification("acct1", strings.Repeat("1", 32), "MD5", false, "bad_signature")
	d.RecordEncryptionOperation("acct1", strings.Repeat("2", 32), "AES-128", "encrypt", true, "", 0.4)

	got := drainEvents(events)
	if !hasEvent(got, string(CryptoAnomalyDeprecatedAlgorithm)) {
		t.Error("expected a deprecated-algorithm event from the signature failure")
	}
	if !hasEvent(got, string(CryptoAnomalyWeakEntropy)) {
		t.Error("expected a weak-entropy event from the encryption operation")
	}
}

func TestCryptoDetectorVelocityAndGlobalFailureStorm(t *testing.T) {
	cfg := DefaultCryptoAnomalyConfig()
	cfg.MaxKeyOperationsPerMinute = 5
	cfg.MaxSignatureFailuresPerAccount = 3
	cfg.MaxSignatureFailuresPerHour = 8
	d := NewCryptoAnomalyDetector(cfg, quietLogger())
	events := make(chan *SecurityEvent, 1000)
	d.Start(context.Background(), events)

	now := time.Now()
	for i := 0; i < 12; i++ {
		d.Analyze(&CryptoOperationData{
			OperationID: "op" + string(rune('a'+i)), OperationType: "sign",
			Algorithm: "Ed25519", AccountAddress: "acct",
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Success:   false, FailureReason: "verify_failed",
		})
	}
	got := drainEvents(events)
	if !hasEvent(got, string(CryptoAnomalyRapidOperations)) {
		t.Error("expected a rapid-operations event")
	}
	if !hasEvent(got, string(CryptoAnomalySignatureFailure)) {
		t.Error("expected a signature-failure event")
	}
}

func TestCryptoDetectorNilConfig(t *testing.T) {
	d := NewCryptoAnomalyDetector(nil, quietLogger())
	if d.config == nil || d.config.MaxSignatureFailuresPerHour != 50 {
		t.Fatalf("nil config did not fall back to defaults")
	}
	// Entropy analysis disabled -> no weak-entropy event even for keygen.
	cfg := DefaultCryptoAnomalyConfig()
	cfg.EnableEntropyAnalysis = false
	d2 := NewCryptoAnomalyDetector(cfg, quietLogger())
	events := make(chan *SecurityEvent, 10)
	d2.Start(context.Background(), events)
	d2.Analyze(&CryptoOperationData{
		OperationID: "o", OperationType: "keygen", Algorithm: "Ed25519",
		AccountAddress: "a", Timestamp: time.Now(), Success: true, EntropyScore: 0.1,
	})
	if hasEvent(drainEvents(events), string(CryptoAnomalyWeakEntropy)) {
		t.Error("entropy analysis is disabled; no weak-entropy event expected")
	}
}

// --- provider_security.go ---

func TestMatchesIPRange(t *testing.T) {
	if !matchesIPRange("10.0.0.1", "10.0.0.1") {
		t.Error("exact match should succeed")
	}
	if matchesIPRange("10.0.0.2", "10.0.0.1") {
		t.Error("different addresses must not match")
	}
}

func TestProviderSecurityKeyFailures(t *testing.T) {
	cfg := DefaultProviderSecurityConfig()
	cfg.MaxFailedSignaturesPerHour = 3
	m := NewProviderSecurityMonitor(cfg, quietLogger())
	events := make(chan *SecurityEvent, 500)
	m.Start(context.Background(), events)

	now := time.Now()
	for i := 0; i < 5; i++ {
		m.Analyze(&ProviderActivityData{
			ProviderID: "p1", ActivityType: "key_usage", KeyID: "k1",
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Success:   false, FailureReason: "bad_signature",
		})
	}
	if !hasEvent(drainEvents(events), string(ProviderIndicatorKeyCompromise)) {
		t.Error("expected a key compromise event from repeated key failures")
	}
}

func TestProviderSecurityLocationAnomaly(t *testing.T) {
	cfg := DefaultProviderSecurityConfig()
	cfg.AllowedIPRanges = []string{"10.0.0.1"}
	m := NewProviderSecurityMonitor(cfg, quietLogger())
	events := make(chan *SecurityEvent, 500)
	m.Start(context.Background(), events)

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) // inside normal hours

	// Unauthorized IP -> high anomaly.
	m.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "bid", Timestamp: base, SourceIP: "203.0.113.9",
	})
	got := drainEvents(events)
	if !hasEvent(got, string(ProviderIndicatorAnomalousLocation)) {
		t.Error("expected an anomalous-location event for an unauthorized IP")
	}

	// Rapid change of location from an allowed IP to a different one.
	m.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "bid", Timestamp: base.Add(time.Second), SourceIP: "10.0.0.1",
	})
	m.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "bid", Timestamp: base.Add(2 * time.Second), SourceIP: "198.51.100.4",
	})
	got = drainEvents(events)
	if !hasEvent(got, string(ProviderIndicatorAnomalousLocation)) {
		t.Error("expected a rapid-location-change event")
	}

	// Tracking disabled -> no location events at all.
	cfg2 := DefaultProviderSecurityConfig()
	cfg2.EnableLocationTracking = false
	cfg2.AllowedIPRanges = []string{"10.0.0.1"}
	m2 := NewProviderSecurityMonitor(cfg2, quietLogger())
	events2 := make(chan *SecurityEvent, 50)
	m2.Start(context.Background(), events2)
	m2.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "bid", Timestamp: base, SourceIP: "203.0.113.9",
	})
	if hasEvent(drainEvents(events2), string(ProviderIndicatorAnomalousLocation)) {
		t.Error("location tracking is disabled; no location event expected")
	}
}

func TestProviderSecurityTimeAnomaly(t *testing.T) {
	m := NewProviderSecurityMonitor(DefaultProviderSecurityConfig(), quietLogger())
	events := make(chan *SecurityEvent, 100)
	m.Start(context.Background(), events)

	// 03:00 is outside the default 06:00-22:00 window.
	m.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "lease",
		Timestamp: time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC),
	})
	if !hasEvent(drainEvents(events), string(ProviderIndicatorAnomalousTime)) {
		t.Error("expected an anomalous-time event at 03:00")
	}

	// 12:00 is inside the window -> no event.
	m.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "lease",
		Timestamp: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	})
	if hasEvent(drainEvents(events), string(ProviderIndicatorAnomalousTime)) {
		t.Error("no anomalous-time event expected at 12:00")
	}
}

func TestProviderSecurityResourceAnomaly(t *testing.T) {
	cfg := DefaultProviderSecurityConfig()
	cfg.ResourceUsageVarianceThreshold = 0.2
	m := NewProviderSecurityMonitor(cfg, quietLogger())
	events := make(chan *SecurityEvent, 200)
	m.Start(context.Background(), events)

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	// Nine steady samples establish the baseline (sampleCount reaches 10).
	for i := 0; i < 9; i++ {
		m.Analyze(&ProviderActivityData{
			ProviderID: "p1", ActivityType: "lease", Timestamp: base,
			CPUUsage: 50, MemoryUsage: 40, StorageUsage: 10,
		})
	}
	drainEvents(events)

	// A large spike must be flagged.
	m.Analyze(&ProviderActivityData{
		ProviderID: "p1", ActivityType: "lease", Timestamp: base,
		CPUUsage: 500, MemoryUsage: 400, StorageUsage: 10,
	})
	got := drainEvents(events)
	if !hasEvent(got, string(ProviderIndicatorResourceAnomaly)) {
		t.Error("expected a resource anomaly event after the baseline was established")
	}

	// Detection disabled -> no event.
	cfg2 := DefaultProviderSecurityConfig()
	cfg2.EnableResourceAnomalyDetection = false
	m2 := NewProviderSecurityMonitor(cfg2, quietLogger())
	events2 := make(chan *SecurityEvent, 50)
	m2.Start(context.Background(), events2)
	m2.Analyze(&ProviderActivityData{ProviderID: "p", ActivityType: "lease", Timestamp: base, CPUUsage: 1})
	if hasEvent(drainEvents(events2), string(ProviderIndicatorResourceAnomaly)) {
		t.Error("resource anomaly detection is disabled; no event expected")
	}
}

func TestProviderSecurityVelocityLimits(t *testing.T) {
	cfg := DefaultProviderSecurityConfig()
	cfg.MaxBidsPerMinute = 5
	cfg.MaxLeasesPerHour = 8
	cfg.MaxDeploymentsPerHour = 4
	m := NewProviderSecurityMonitor(cfg, quietLogger())
	events := make(chan *SecurityEvent, 1000)
	m.Start(context.Background(), events)

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		ts := base.Add(time.Duration(i) * time.Second)
		m.Analyze(&ProviderActivityData{ProviderID: "p1", ActivityType: "bid", Timestamp: ts})
		m.Analyze(&ProviderActivityData{ProviderID: "p1", ActivityType: "lease", Timestamp: ts})
		m.Analyze(&ProviderActivityData{ProviderID: "p1", ActivityType: "deploy", Timestamp: ts})
	}

	got := drainEvents(events)
	for _, want := range []string{
		string(ProviderIndicatorRapidActivity), // bids + deploys
		string(ProviderIndicatorLeaseAbuse),    // leases
		string(ProviderIndicatorAnomalousTime), // 12:00 is in-window, so this must NOT appear
	} {
		if want == string(ProviderIndicatorAnomalousTime) {
			if hasEvent(got, want) {
				t.Error("12:00 is inside the normal-hours window; no anomalous-time event expected")
			}
			continue
		}
		if !hasEvent(got, want) {
			t.Errorf("expected %q event from velocity limits", want)
		}
	}
}

func TestProviderSecurityNilConfig(t *testing.T) {
	m := NewProviderSecurityMonitor(nil, quietLogger())
	if m.config == nil || m.config.MaxKeyUsagePerMinute != 20 {
		t.Fatalf("nil config did not fall back to defaults")
	}
	// No eventChan set: emitEvent must drop silently, not block.
	m.Analyze(&ProviderActivityData{
		ProviderID: "p", ActivityType: "bid", Timestamp: time.Now(), BidAmount: 0.001,
	})
}
