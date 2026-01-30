// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package security_monitoring

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewIncidentResponder(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	ir, err := NewIncidentResponder("", logger)
	if err != nil {
		t.Fatalf("NewIncidentResponder failed: %v", err)
	}
	if ir == nil {
		t.Fatal("NewIncidentResponder returned nil")
	}
}

func TestIncidentResponderDefaultPlaybooks(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	ir, err := NewIncidentResponder("", logger)
	if err != nil {
		t.Fatalf("NewIncidentResponder failed: %v", err)
	}

	playbooks := ir.GetAllPlaybooks()
	if len(playbooks) == 0 {
		t.Error("Expected default playbooks to be loaded")
	}
}

func TestPlaybookStructure(t *testing.T) {
	playbook := &Playbook{
		ID:              "pb_test",
		Name:            "Test Playbook",
		Description:     "A test playbook",
		TriggerTypes:    []string{"fraud_detected"},
		MinSeverity:     SeverityMedium,
		Enabled:         true,
		CooldownMinutes: 5,
		Steps: []PlaybookStep{
			{
				Name:   "log-event",
				Action: string(ActionLogEvent),
				Parameters: map[string]string{
					"message": "Test event",
				},
				Timeout: 30,
			},
		},
	}

	if playbook.ID == "" {
		t.Error("Playbook ID should not be empty")
	}

	if playbook.Name == "" {
		t.Error("Playbook Name should not be empty")
	}

	if len(playbook.Steps) == 0 {
		t.Error("Playbook should have at least one step")
	}
}

func TestPlaybookStepStructure(t *testing.T) {
	step := PlaybookStep{
		Name:              "test-step",
		Action:            string(ActionSendAlert),
		Parameters:        map[string]string{"channel": "security"},
		Timeout:           60,
		ContinueOnFailure: true,
		Condition:         "severity == 'critical'",
	}

	if step.Name == "" {
		t.Error("Step Name should not be empty")
	}

	if step.Action == "" {
		t.Error("Step Action should not be empty")
	}

	if step.Timeout == 0 {
		t.Error("Step Timeout should not be zero")
	}
}

func TestPlaybookActions(t *testing.T) {
	actions := []PlaybookAction{
		ActionLogEvent,
		ActionSendAlert,
		ActionBlockIP,
		ActionRevokeKey,
		ActionSuspendAccount,
		ActionSuspendProvider,
		ActionIncreaseSeverity,
		ActionTriggerBackup,
		ActionNotifyTeam,
		ActionRunScript,
		ActionUpdateFirewall,
		ActionCollectEvidence,
		ActionEscalate,
	}

	for _, action := range actions {
		if action == "" {
			t.Error("Playbook action constant should not be empty")
		}
	}
}

func TestIncidentResponderAddPlaybook(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	ir, err := NewIncidentResponder("", logger)
	if err != nil {
		t.Fatalf("NewIncidentResponder failed: %v", err)
	}

	playbook := &Playbook{
		ID:           "pb_custom",
		Name:         "Custom Playbook",
		Description:  "A custom test playbook",
		TriggerTypes: []string{"test_event"},
		MinSeverity:  SeverityLow,
		Enabled:      true,
		Steps: []PlaybookStep{
			{
				Name:   "log-step",
				Action: string(ActionLogEvent),
			},
		},
	}

	ir.AddPlaybook(playbook)

	// Verify registration
	pb, exists := ir.GetPlaybook("pb_custom")
	if !exists {
		t.Error("Registered playbook not found")
	}
	if pb == nil {
		t.Error("Playbook should not be nil")
	}
}

func TestPlaybookExecutionStructure(t *testing.T) {
	now := time.Now()
	execution := &PlaybookExecution{
		ID:          "exec_123",
		PlaybookID:  "pb_test",
		IncidentID:  "inc_456",
		StartedAt:   now,
		Status:      "running",
		StepsExecuted: []StepExecution{
			{
				StepName:    "step1",
				Action:      string(ActionLogEvent),
				StartedAt:   now,
				Success:     true,
			},
		},
	}

	if execution.ID == "" {
		t.Error("Execution ID should not be empty")
	}

	if execution.PlaybookID == "" {
		t.Error("Execution PlaybookID should not be empty")
	}

	if execution.Status == "" {
		t.Error("Execution Status should not be empty")
	}
}

func TestStepExecutionStructure(t *testing.T) {
	now := time.Now()
	step := StepExecution{
		StepName:    "test-step",
		Action:      string(ActionSendAlert),
		StartedAt:   now,
		CompletedAt: &now,
		Success:     true,
		Output:      "Alert sent successfully",
	}

	if step.StepName == "" {
		t.Error("StepName should not be empty")
	}

	if step.Action == "" {
		t.Error("Action should not be empty")
	}

	if step.CompletedAt == nil {
		t.Error("CompletedAt should not be nil for completed step")
	}
}

func TestIncidentResponderGetAllPlaybooks(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	ir, err := NewIncidentResponder("", logger)
	if err != nil {
		t.Fatalf("NewIncidentResponder failed: %v", err)
	}

	playbooks := ir.GetAllPlaybooks()
	if playbooks == nil {
		t.Error("GetAllPlaybooks should not return nil")
	}
}

func TestIncidentResponderGetPlaybook(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	ir, err := NewIncidentResponder("", logger)
	if err != nil {
		t.Fatalf("NewIncidentResponder failed: %v", err)
	}

	// Try to get non-existent playbook
	pb, exists := ir.GetPlaybook("nonexistent")
	if exists {
		t.Error("GetPlaybook should return false for nonexistent playbook")
	}
	if pb != nil {
		t.Error("Playbook should be nil for nonexistent playbook")
	}
}

func TestIncidentResponderRemovePlaybook(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	ir, err := NewIncidentResponder("", logger)
	if err != nil {
		t.Fatalf("NewIncidentResponder failed: %v", err)
	}

	// Add a playbook first
	playbook := &Playbook{
		ID:           "pb_remove",
		Name:         "Remove Test",
		TriggerTypes: []string{"test"},
		Enabled:      true,
	}
	ir.AddPlaybook(playbook)

	// Verify it exists
	_, exists := ir.GetPlaybook("pb_remove")
	if !exists {
		t.Error("Playbook should exist after add")
	}

	// Remove it
	ir.RemovePlaybook("pb_remove")

	// Verify it's gone
	_, exists = ir.GetPlaybook("pb_remove")
	if exists {
		t.Error("Playbook should not exist after remove")
	}
}
