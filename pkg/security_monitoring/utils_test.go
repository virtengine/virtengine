// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"testing"
	"time"
)

func TestGenerateEventID(t *testing.T) {
	id1 := generateEventID()
	time.Sleep(time.Millisecond) // Ensure different timestamp
	id2 := generateEventID()

	if id1 == "" {
		t.Error("generateEventID should not return empty string")
	}

	if id2 == "" {
		t.Error("generateEventID should not return empty string")
	}

	// IDs should be unique (with time gap)
	if id1 == id2 {
		t.Error("generateEventID should generate unique IDs")
	}
}

func TestGenerateIncidentID(t *testing.T) {
	id1 := generateIncidentID()
	time.Sleep(time.Millisecond)
	id2 := generateIncidentID()

	if id1 == "" {
		t.Error("generateIncidentID should not return empty string")
	}

	if id2 == "" {
		t.Error("generateIncidentID should not return empty string")
	}

	// IDs should be unique (with time gap)
	if id1 == id2 {
		t.Error("generateIncidentID should generate unique IDs")
	}
}

func TestGenerateAlertID(t *testing.T) {
	id1 := generateAlertID()
	time.Sleep(time.Millisecond)
	id2 := generateAlertID()

	if id1 == "" {
		t.Error("generateAlertID should not return empty string")
	}

	if id2 == "" {
		t.Error("generateAlertID should not return empty string")
	}

	// IDs should be unique (with time gap)
	if id1 == id2 {
		t.Error("generateAlertID should generate unique IDs")
	}
}

func TestGenerateExecutionID(t *testing.T) {
	id1 := generateExecutionID()
	time.Sleep(time.Millisecond)
	id2 := generateExecutionID()

	if id1 == "" {
		t.Error("generateExecutionID should not return empty string")
	}

	if id2 == "" {
		t.Error("generateExecutionID should not return empty string")
	}

	// IDs should be unique (with time gap)
	if id1 == id2 {
		t.Error("generateExecutionID should generate unique IDs")
	}
}

func TestIDPrefixes(t *testing.T) {
	eventID := generateEventID()
	incidentID := generateIncidentID()
	alertID := generateAlertID()
	execID := generateExecutionID()

	// Check for expected prefixes
	if len(eventID) < 4 || eventID[:4] != "evt-" {
		t.Errorf("Event ID should start with 'evt-', got '%s'", eventID)
	}

	if len(incidentID) < 4 || incidentID[:4] != "inc-" {
		t.Errorf("Incident ID should start with 'inc-', got '%s'", incidentID)
	}

	if len(alertID) < 4 || alertID[:4] != "alt-" {
		t.Errorf("Alert ID should start with 'alt-', got '%s'", alertID)
	}

	if len(execID) < 4 || execID[:4] != "exe-" {
		t.Errorf("Execution ID should start with 'exe-', got '%s'", execID)
	}
}

func TestIDLength(t *testing.T) {
	eventID := generateEventID()
	incidentID := generateIncidentID()
	alertID := generateAlertID()

	// IDs should have reasonable length (prefix + timestamp/random)
	minLength := 10
	if len(eventID) < minLength {
		t.Errorf("Event ID too short: %s", eventID)
	}

	if len(incidentID) < minLength {
		t.Errorf("Incident ID too short: %s", incidentID)
	}

	if len(alertID) < minLength {
		t.Errorf("Alert ID too short: %s", alertID)
	}
}

func TestGenerateIDWithPrefix(t *testing.T) {
	id := generateID("test")
	if id == "" {
		t.Error("generateID should not return empty string")
	}

	if len(id) < 5 || id[:5] != "test-" {
		t.Errorf("ID should start with 'test-', got '%s'", id)
	}
}
