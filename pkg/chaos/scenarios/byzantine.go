// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package scenarios provides chaos engineering experiment scenarios for VirtEngine.
//
// This file implements Byzantine behavior simulation scenarios for testing consensus
// safety and liveness properties. Byzantine scenarios simulate malicious or faulty
// validator behavior including:
//
//   - Double signing: Validators signing conflicting blocks at the same height
//   - Equivocation: Validators sending conflicting messages
//   - Invalid blocks: Validators proposing malformed or invalid blocks
//   - Message tampering: Corrupting consensus messages
//   - Selective forwarding: Validators selectively relaying messages
//
// These scenarios are critical for validating the BFT consensus implementation
// and ensuring the system can detect and handle Byzantine faults correctly.
package scenarios

import (
	"errors"
	"fmt"
	"time"
)

// Byzantine experiment types.
const (
	// ExperimentTypeByzantineDoubleSigning simulates double-signing attacks.
	ExperimentTypeByzantineDoubleSigning ExperimentType = "byzantine-double-signing"

	// ExperimentTypeByzantineEquivocation simulates equivocation attacks.
	ExperimentTypeByzantineEquivocation ExperimentType = "byzantine-equivocation"

	// ExperimentTypeByzantineInvalidBlock simulates invalid block proposals.
	ExperimentTypeByzantineInvalidBlock ExperimentType = "byzantine-invalid-block"

	// ExperimentTypeByzantineMessageTampering simulates message corruption.
	ExperimentTypeByzantineMessageTampering ExperimentType = "byzantine-message-tampering"

	// ExperimentTypeByzantineSelectiveForwarding simulates selective message forwarding.
	ExperimentTypeByzantineSelectiveForwarding ExperimentType = "byzantine-selective-forwarding"

	// ExperimentTypeByzantineGeneric is a generic Byzantine fault type.
	ExperimentTypeByzantineGeneric ExperimentType = "byzantine"
)

// ByzantineSpec contains the configuration for Byzantine behavior experiments.
type ByzantineSpec struct {
	// ByzantineType specifies the type of Byzantine behavior to simulate.
	ByzantineType string `json:"byzantine_type"`

	// Targets are the validators to exhibit Byzantine behavior.
	Targets []string `json:"targets"`

	// Probability is the chance (0-100) of exhibiting Byzantine behavior.
	Probability float64 `json:"probability"`

	// Height is the specific height to target (0 for any height).
	Height int64 `json:"height,omitempty"`

	// Round is the specific round to target (0 for any round).
	Round int64 `json:"round,omitempty"`

	// Parameters contains type-specific configuration.
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ByzantineScenario is the interface for Byzantine behavior scenarios.
type ByzantineScenario interface {
	// Name returns the scenario identifier.
	Name() string

	// Description returns a human-readable description.
	Description() string

	// Type returns the experiment type.
	Type() ExperimentType

	// Build constructs an Experiment from the scenario.
	Build() (*Experiment, error)

	// Validate checks the scenario configuration.
	Validate() error
}

// DoubleSigningScenario simulates validators double-signing blocks.
// Double-signing is when a validator signs two different blocks at the same height,
// which is a severe consensus violation that should be detected and slashed.
type DoubleSigningScenario struct {
	// name is the scenario identifier.
	name string

	// description provides scenario details.
	description string

	// Validators are the validators to simulate double-signing.
	Validators []string

	// TargetHeight is the height at which to double-sign (0 for next height).
	TargetHeight int64

	// Duration is how long to maintain the Byzantine behavior.
	Duration time.Duration

	// DetectionWindow is the expected time for the system to detect the fault.
	DetectionWindow time.Duration
}

// Name returns the scenario identifier.
func (d *DoubleSigningScenario) Name() string {
	if d.name != "" {
		return d.name
	}
	return "double-signing"
}

// Description returns the scenario description.
func (d *DoubleSigningScenario) Description() string {
	if d.description != "" {
		return d.description
	}
	return "Simulates validators double-signing blocks to test slashing detection"
}

// Type returns the experiment type.
func (d *DoubleSigningScenario) Type() ExperimentType {
	return ExperimentTypeByzantineDoubleSigning
}

// Validate checks the scenario configuration.
func (d *DoubleSigningScenario) Validate() error {
	if len(d.Validators) == 0 {
		return errors.New("at least one validator is required for double-signing scenario")
	}

	if d.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	if d.DetectionWindow <= 0 {
		d.DetectionWindow = 30 * time.Second // Default detection window
	}

	return nil
}

// Build constructs an Experiment from the scenario.
func (d *DoubleSigningScenario) Build() (*Experiment, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        d.Name(),
		Description: d.Description(),
		Type:        ExperimentTypeByzantineDoubleSigning,
		Duration:    d.Duration,
		Spec: ByzantineSpec{
			ByzantineType: "double-signing",
			Targets:       d.Validators,
			Height:        d.TargetHeight,
			Probability:   100, // Always double-sign when triggered
			Parameters: map[string]interface{}{
				"detection_window": d.DetectionWindow.Seconds(),
			},
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeByzantineDoubleSigning),
			"chaos.virtengine.io/category": "byzantine",
			"chaos.virtengine.io/severity": "critical",
		},
	}, nil
}

// NewDoubleSigningScenario creates a double-signing scenario for the specified validators.
func NewDoubleSigningScenario(validators []string, duration time.Duration) *DoubleSigningScenario {
	return &DoubleSigningScenario{
		name:            "double-signing-test",
		description:     fmt.Sprintf("Double-signing test for %d validators", len(validators)),
		Validators:      validators,
		Duration:        duration,
		DetectionWindow: 30 * time.Second,
	}
}

// EquivocationScenario simulates validators sending conflicting messages.
// Equivocation is when a validator sends different prevotes or precommits
// to different peers in the same round.
type EquivocationScenario struct {
	// name is the scenario identifier.
	name string

	// description provides scenario details.
	description string

	// Validators are the validators to simulate equivocation.
	Validators []string

	// MessageType is the type of message to equivocate (prevote, precommit).
	MessageType string

	// TargetRound is the round at which to equivocate (0 for any round).
	TargetRound int64

	// Duration is how long to maintain the Byzantine behavior.
	Duration time.Duration

	// Probability is the chance (0-100) of equivocating on each message.
	Probability float64
}

// Name returns the scenario identifier.
func (e *EquivocationScenario) Name() string {
	if e.name != "" {
		return e.name
	}
	return "equivocation"
}

// Description returns the scenario description.
func (e *EquivocationScenario) Description() string {
	if e.description != "" {
		return e.description
	}
	return "Simulates validators sending conflicting consensus messages"
}

// Type returns the experiment type.
func (e *EquivocationScenario) Type() ExperimentType {
	return ExperimentTypeByzantineEquivocation
}

// Validate checks the scenario configuration.
func (e *EquivocationScenario) Validate() error {
	if len(e.Validators) == 0 {
		return errors.New("at least one validator is required")
	}

	if e.MessageType != "" && e.MessageType != "prevote" && e.MessageType != "precommit" {
		return errors.New("message type must be 'prevote' or 'precommit'")
	}

	if e.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	if e.Probability <= 0 || e.Probability > 100 {
		e.Probability = 100
	}

	return nil
}

// Build constructs an Experiment from the scenario.
func (e *EquivocationScenario) Build() (*Experiment, error) {
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        e.Name(),
		Description: e.Description(),
		Type:        ExperimentTypeByzantineEquivocation,
		Duration:    e.Duration,
		Spec: ByzantineSpec{
			ByzantineType: "equivocation",
			Targets:       e.Validators,
			Round:         e.TargetRound,
			Probability:   e.Probability,
			Parameters: map[string]interface{}{
				"message_type": e.MessageType,
			},
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeByzantineEquivocation),
			"chaos.virtengine.io/category": "byzantine",
		},
	}, nil
}

// NewEquivocationScenario creates an equivocation scenario.
func NewEquivocationScenario(validators []string, messageType string, duration time.Duration) *EquivocationScenario {
	return &EquivocationScenario{
		name:        "equivocation-test",
		description: fmt.Sprintf("Equivocation test (%s) for %d validators", messageType, len(validators)),
		Validators:  validators,
		MessageType: messageType,
		Duration:    duration,
		Probability: 100,
	}
}

// InvalidBlockScenario simulates validators proposing invalid blocks.
// This tests the block validation logic and ensures invalid blocks are rejected.
type InvalidBlockScenario struct {
	// name is the scenario identifier.
	name string

	// description provides scenario details.
	description string

	// Validators are the validators to propose invalid blocks.
	Validators []string

	// InvalidationType specifies what makes the block invalid.
	// Options: "malformed", "wrong_app_hash", "future_timestamp", "invalid_signature"
	InvalidationType string

	// Duration is how long to maintain the Byzantine behavior.
	Duration time.Duration

	// Probability is the chance (0-100) of proposing an invalid block.
	Probability float64
}

// Name returns the scenario identifier.
func (i *InvalidBlockScenario) Name() string {
	if i.name != "" {
		return i.name
	}
	return "invalid-block"
}

// Description returns the scenario description.
func (i *InvalidBlockScenario) Description() string {
	if i.description != "" {
		return i.description
	}
	return "Simulates validators proposing invalid blocks"
}

// Type returns the experiment type.
func (i *InvalidBlockScenario) Type() ExperimentType {
	return ExperimentTypeByzantineInvalidBlock
}

// Validate checks the scenario configuration.
func (i *InvalidBlockScenario) Validate() error {
	if len(i.Validators) == 0 {
		return errors.New("at least one validator is required")
	}

	validTypes := map[string]bool{
		"malformed": true, "wrong_app_hash": true,
		"future_timestamp": true, "invalid_signature": true,
	}
	if i.InvalidationType != "" && !validTypes[i.InvalidationType] {
		return fmt.Errorf("invalid invalidation type: %s", i.InvalidationType)
	}

	if i.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	return nil
}

// Build constructs an Experiment from the scenario.
func (i *InvalidBlockScenario) Build() (*Experiment, error) {
	if err := i.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        i.Name(),
		Description: i.Description(),
		Type:        ExperimentTypeByzantineInvalidBlock,
		Duration:    i.Duration,
		Spec: ByzantineSpec{
			ByzantineType: "invalid-block",
			Targets:       i.Validators,
			Probability:   i.Probability,
			Parameters: map[string]interface{}{
				"invalidation_type": i.InvalidationType,
			},
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeByzantineInvalidBlock),
			"chaos.virtengine.io/category": "byzantine",
		},
	}, nil
}

// NewInvalidBlockScenario creates an invalid block proposal scenario.
func NewInvalidBlockScenario(validators []string, invalidationType string, duration time.Duration) *InvalidBlockScenario {
	return &InvalidBlockScenario{
		name:             "invalid-block-test",
		description:      fmt.Sprintf("Invalid block test (%s) for %d validators", invalidationType, len(validators)),
		Validators:       validators,
		InvalidationType: invalidationType,
		Duration:         duration,
		Probability:      100,
	}
}

// MessageTamperingScenario simulates corruption of consensus messages.
type MessageTamperingScenario struct {
	// name is the scenario identifier.
	name string

	// description provides scenario details.
	description string

	// Validators are the validators to tamper with messages.
	Validators []string

	// TamperType specifies how to corrupt messages.
	// Options: "corrupt_signature", "corrupt_hash", "modify_height", "drop_fields"
	TamperType string

	// Duration is how long to maintain the Byzantine behavior.
	Duration time.Duration

	// Probability is the chance (0-100) of tampering with each message.
	Probability float64
}

// Name returns the scenario identifier.
func (m *MessageTamperingScenario) Name() string {
	if m.name != "" {
		return m.name
	}
	return "message-tampering"
}

// Description returns the scenario description.
func (m *MessageTamperingScenario) Description() string {
	if m.description != "" {
		return m.description
	}
	return "Simulates corruption of consensus messages"
}

// Type returns the experiment type.
func (m *MessageTamperingScenario) Type() ExperimentType {
	return ExperimentTypeByzantineMessageTampering
}

// Validate checks the scenario configuration.
func (m *MessageTamperingScenario) Validate() error {
	if len(m.Validators) == 0 {
		return errors.New("at least one validator is required")
	}

	if m.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	return nil
}

// Build constructs an Experiment from the scenario.
func (m *MessageTamperingScenario) Build() (*Experiment, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        m.Name(),
		Description: m.Description(),
		Type:        ExperimentTypeByzantineMessageTampering,
		Duration:    m.Duration,
		Spec: ByzantineSpec{
			ByzantineType: "message-tampering",
			Targets:       m.Validators,
			Probability:   m.Probability,
			Parameters: map[string]interface{}{
				"tamper_type": m.TamperType,
			},
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeByzantineMessageTampering),
			"chaos.virtengine.io/category": "byzantine",
		},
	}, nil
}

// NewMessageTamperingScenario creates a message tampering scenario.
func NewMessageTamperingScenario(validators []string, tamperType string, duration time.Duration) *MessageTamperingScenario {
	return &MessageTamperingScenario{
		name:        "message-tampering-test",
		description: fmt.Sprintf("Message tampering test (%s) for %d validators", tamperType, len(validators)),
		Validators:  validators,
		TamperType:  tamperType,
		Duration:    duration,
		Probability: 50, // 50% of messages tampered by default
	}
}

// SelectiveForwardingScenario simulates validators selectively forwarding messages.
type SelectiveForwardingScenario struct {
	// name is the scenario identifier.
	name string

	// description provides scenario details.
	description string

	// Validators are the validators to exhibit selective forwarding.
	Validators []string

	// DropTargets are specific validators whose messages should be dropped.
	DropTargets []string

	// DropMessageTypes are the message types to selectively drop.
	DropMessageTypes []string

	// Duration is how long to maintain the Byzantine behavior.
	Duration time.Duration

	// DropProbability is the chance (0-100) of dropping matching messages.
	DropProbability float64
}

// Name returns the scenario identifier.
func (s *SelectiveForwardingScenario) Name() string {
	if s.name != "" {
		return s.name
	}
	return "selective-forwarding"
}

// Description returns the scenario description.
func (s *SelectiveForwardingScenario) Description() string {
	if s.description != "" {
		return s.description
	}
	return "Simulates validators selectively forwarding messages"
}

// Type returns the experiment type.
func (s *SelectiveForwardingScenario) Type() ExperimentType {
	return ExperimentTypeByzantineSelectiveForwarding
}

// Validate checks the scenario configuration.
func (s *SelectiveForwardingScenario) Validate() error {
	if len(s.Validators) == 0 {
		return errors.New("at least one validator is required")
	}

	if s.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	return nil
}

// Build constructs an Experiment from the scenario.
func (s *SelectiveForwardingScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        s.Name(),
		Description: s.Description(),
		Type:        ExperimentTypeByzantineSelectiveForwarding,
		Duration:    s.Duration,
		Spec: ByzantineSpec{
			ByzantineType: "selective-forwarding",
			Targets:       s.Validators,
			Probability:   s.DropProbability,
			Parameters: map[string]interface{}{
				"drop_targets":       s.DropTargets,
				"drop_message_types": s.DropMessageTypes,
			},
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeByzantineSelectiveForwarding),
			"chaos.virtengine.io/category": "byzantine",
		},
	}, nil
}

// NewSelectiveForwardingScenario creates a selective forwarding scenario.
func NewSelectiveForwardingScenario(validators, dropTargets []string, duration time.Duration) *SelectiveForwardingScenario {
	return &SelectiveForwardingScenario{
		name:             "selective-forwarding-test",
		description:      fmt.Sprintf("Selective forwarding test for %d validators", len(validators)),
		Validators:       validators,
		DropTargets:      dropTargets,
		DropMessageTypes: []string{"prevote", "precommit"},
		Duration:         duration,
		DropProbability:  100,
	}
}

// DefaultByzantineScenarios returns a collection of Byzantine behavior scenarios
// for testing consensus safety and liveness properties.
func DefaultByzantineScenarios() []ByzantineScenario {
	defaultValidators := []string{"validator-0"}

	return []ByzantineScenario{
		// Double-signing detection
		NewDoubleSigningScenario(defaultValidators, 5*time.Minute),

		// Equivocation scenarios
		NewEquivocationScenario(defaultValidators, "prevote", 5*time.Minute),
		NewEquivocationScenario(defaultValidators, "precommit", 5*time.Minute),

		// Invalid block proposals
		NewInvalidBlockScenario(defaultValidators, "malformed", 5*time.Minute),
		NewInvalidBlockScenario(defaultValidators, "wrong_app_hash", 5*time.Minute),
		NewInvalidBlockScenario(defaultValidators, "future_timestamp", 5*time.Minute),

		// Message tampering
		NewMessageTamperingScenario(defaultValidators, "corrupt_signature", 5*time.Minute),
		NewMessageTamperingScenario(defaultValidators, "corrupt_hash", 5*time.Minute),

		// Selective forwarding
		NewSelectiveForwardingScenario(defaultValidators, []string{"validator-1"}, 5*time.Minute),
	}
}

