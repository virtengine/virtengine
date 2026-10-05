// Behaviour tests for the chaos scenario builders.
//
// pkg/chaos/scenarios measured 2.0% overall: labels_test.go covers the label
// constants and one representative builder per prefix family, but every
// scenario's Name/Description/Type/Validate/Build quartet was untested. That is
// not cosmetic. These builders are the hand-off from an operator's YAML to the
// chaos engine, and three of them MUTATE the receiver inside Validate:
//
//	DoubleSigningScenario.Validate  fills a zero DetectionWindow with 30s
//	EquivocationScenario.Validate  clamps an out-of-range Probability to 100
//
// A regression in either would not change an exit code -- it would change what
// the experiment actually does to a live validator set, which is exactly the
// failure a test suite exists to catch.
package scenarios

import (
	"strings"
	"testing"
	"time"
)

// TestByzantineScenarioBuilders pins each Byzantine builder's emitted
// Experiment: the type, the category label, and the spec fields the chaos
// engine dispatches on.
func TestByzantineScenarioBuilders(t *testing.T) {
	cases := []struct {
		name     string
		scenario interface {
			Build() (*Experiment, error)
		}
		wantType     ExperimentType
		wantCategory string
		wantName     string
	}{
		{
			name:         "double-signing",
			scenario:     NewDoubleSigningScenario([]string{"val-0", "val-1"}, time.Minute),
			wantType:     ExperimentTypeByzantineDoubleSigning,
			wantCategory: ChaosCategoryByzantine,
			wantName:     "double-signing-test",
		},
		{
			name:         "equivocation",
			scenario:     NewEquivocationScenario([]string{"val-0"}, "prevote", time.Minute),
			wantType:     ExperimentTypeByzantineEquivocation,
			wantCategory: ChaosCategoryByzantine,
			wantName:     "equivocation-test",
		},
		{
			name:         "invalid-block",
			scenario:     NewInvalidBlockScenario([]string{"val-0"}, "malformed", time.Minute),
			wantType:     ExperimentTypeByzantineInvalidBlock,
			wantCategory: ChaosCategoryByzantine,
			wantName:     "invalid-block-test",
		},
		{
			name:         "message-tampering",
			scenario:     NewMessageTamperingScenario([]string{"val-0"}, "replay", time.Minute),
			wantType:     ExperimentTypeByzantineMessageTampering,
			wantCategory: ChaosCategoryByzantine,
			wantName:     "message-tampering-test",
		},
		{
			name: "selective-forwarding",
			scenario: NewSelectiveForwardingScenario(
				[]string{"val-0"}, []string{"val-1"}, time.Minute),
			wantType:     ExperimentTypeByzantineSelectiveForwarding,
			wantCategory: ChaosCategoryByzantine,
			wantName:     "selective-forwarding-test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exp, err := tc.scenario.Build()
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			if exp.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", exp.Type, tc.wantType)
			}
			if got := exp.Labels[ChaosLabelCategory]; got != tc.wantCategory {
				t.Errorf("category label = %q, want %q", got, tc.wantCategory)
			}
			// The type label must agree with the struct field: the chaos
			// engine selects on the label, not the field, so a disagreement
			// would dispatch the right experiment to the wrong selector.
			if got := exp.Labels[ChaosLabelType]; got != string(tc.wantType) {
				t.Errorf("type label = %q, want %q", got, string(tc.wantType))
			}
			if exp.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", exp.Name, tc.wantName)
			}
			if exp.Duration != time.Minute {
				t.Errorf("Duration = %v, want %v", exp.Duration, time.Minute)
			}
			if strings.TrimSpace(exp.Description) == "" {
				t.Error("Description is empty; operators select scenarios by it")
			}
		})
	}
}

// TestDoubleSigningBuildPinsSpec pins the spec fields the double-signing
// experiment runs with. Probability must be 100 (always double-sign when
// triggered) and the detection window must be the default 30s.
func TestDoubleSigningBuildPinsSpec(t *testing.T) {
	exp, err := NewDoubleSigningScenario([]string{"val-0", "val-1", "val-2"}, time.Minute).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	spec, ok := exp.Spec.(ByzantineSpec)
	if !ok {
		t.Fatalf("Spec is %T, want ByzantineSpec", exp.Spec)
	}
	if len(spec.Targets) != 3 {
		t.Errorf("Targets = %v, want 3 validators", spec.Targets)
	}
	if spec.Probability != 100 {
		t.Errorf("Probability = %v, want 100", spec.Probability)
	}
	if got := spec.Parameters["detection_window"]; got != 30.0 {
		t.Errorf("detection_window = %v, want 30", got)
	}
}

// TestEquivocationBuildPinsMessageType pins that the operator's chosen message
// type reaches the spec -- this is the difference between equivocating a
// prevote and a precommit.
func TestEquivocationBuildPinsMessageType(t *testing.T) {
	for _, msgType := range []string{"prevote", "precommit"} {
		exp, err := NewEquivocationScenario([]string{"val-0"}, msgType, time.Minute).Build()
		if err != nil {
			t.Fatalf("Build(%q): %v", msgType, err)
		}
		spec, ok := exp.Spec.(ByzantineSpec)
		if !ok {
			t.Fatalf("Spec is %T, want ByzantineSpec", exp.Spec)
		}
		if got := spec.Parameters["message_type"]; got != msgType {
			t.Errorf("message_type = %v, want %q", got, msgType)
		}
	}
}

// TestValidateRejectsEmptyValidators covers the shared precondition: a
// Byzantine scenario with no validators targets nothing, and Build must fail
// loudly rather than emit an experiment that cannot fault anything.
func TestValidateRejectsEmptyValidators(t *testing.T) {
	cases := map[string]interface {
		Validate() error
		Build() (*Experiment, error)
	}{
		"double-signing": NewDoubleSigningScenario(nil, time.Minute),
		"equivocation":   NewEquivocationScenario(nil, "prevote", time.Minute),
		"invalid-block":  NewInvalidBlockScenario(nil, "malformed", time.Minute),
		"tampering":      NewMessageTamperingScenario(nil, "replay", time.Minute),
		"selective-fwd":  NewSelectiveForwardingScenario(nil, []string{"val-1"}, time.Minute),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			if err := s.Validate(); err == nil {
				t.Error("Validate with no validators = nil error; want a failure")
			}
			if _, err := s.Build(); err == nil {
				t.Error("Build with no validators = nil error; want a failure")
			}
		})
	}
}

// TestValidateRejectsNonPositiveDuration covers the other shared precondition.
func TestValidateRejectsNonPositiveDuration(t *testing.T) {
	cases := map[string]interface {
		Validate() error
	}{
		"double-signing": NewDoubleSigningScenario([]string{"val-0"}, 0),
		"equivocation":   NewEquivocationScenario([]string{"val-0"}, "prevote", -time.Minute),
		"invalid-block":  NewInvalidBlockScenario([]string{"val-0"}, "malformed", 0),
		"tampering":      NewMessageTamperingScenario([]string{"val-0"}, "replay", 0),
		"selective-fwd":  NewSelectiveForwardingScenario([]string{"val-0"}, nil, 0),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			if err := s.Validate(); err == nil {
				t.Errorf("Validate(%s) with a non-positive duration = nil error; "+
					"want a failure", name)
			}
		})
	}
}

// TestEquivocationValidateClampsProbability pins the MUTATION inside
// Validate: an out-of-range probability is silently rewritten to 100. That is
// the code's actual contract, so a test must assert it -- if the clamp were
// dropped, every out-of-range experiment would start faulting at its literal
// probability (e.g. 150%, i.e. always) or at 0 (never) without any error.
func TestEquivocationValidateClampsProbability(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want float64
	}{
		{"zero is clamped up to 100", 0, 100},
		{"negative is clamped to 100", -5, 100},
		{"above 100 is clamped to 100", 150, 100},
		{"in-range 50 is preserved", 50, 50},
		{"boundary 100 is preserved", 100, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewEquivocationScenario([]string{"val-0"}, "prevote", time.Minute)
			s.Probability = tc.in
			if err := s.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if s.Probability != tc.want {
				t.Errorf("Probability = %v after Validate(%v), want %v",
					s.Probability, tc.in, tc.want)
			}
		})
	}
}

// TestEquivocationValidateRejectsBadMessageType pins the one validation
// Equivocation has that the others do not: only prevote and precommit are
// equivocatable. A typo must fail, not reach the consensus engine.
func TestEquivocationValidateRejectsBadMessageType(t *testing.T) {
	for _, bad := range []string{"prevote2", "PREVOTE", "proposal", "commit"} {
		s := NewEquivocationScenario([]string{"val-0"}, bad, time.Minute)
		if err := s.Validate(); err == nil {
			t.Errorf("Validate(%q) = nil error; want a failure", bad)
		}
	}

	// Empty is explicitly allowed (it means "either"), and must not fail.
	s := NewEquivocationScenario([]string{"val-0"}, "", time.Minute)
	if err := s.Validate(); err != nil {
		t.Errorf("Validate with an empty message type = %v; want nil "+
			"(empty is the documented \"either\" case)", err)
	}
}

// TestInvalidBlockValidateAcceptsKnownTypes walks the accepted set. These are
// the four block shapes the engine knows how to inject; a fifth would be
// rejected, which is the intended safety property.
func TestInvalidBlockValidateAcceptsKnownTypes(t *testing.T) {
	valid := []string{"malformed", "wrong_app_hash", "future_timestamp", "invalid_signature"}
	for _, v := range valid {
		s := NewInvalidBlockScenario([]string{"val-0"}, v, time.Minute)
		if err := s.Validate(); err != nil {
			t.Errorf("Validate(%q) = %v; want nil", v, err)
		}
	}

	for _, bad := range []string{"grafted", "MALFORMED", "wrong-app-hash"} {
		s := NewInvalidBlockScenario([]string{"val-0"}, bad, time.Minute)
		if err := s.Validate(); err == nil {
			t.Errorf("Validate(%q) = nil error; want a failure", bad)
		}
	}
}

// TestDoubleSigningValidateFillsDetectionWindow pins the other MUTATION
// inside Validate: a zero DetectionWindow becomes 30s. Without it the
// experiment would fault a validator with no detection window at all.
func TestDoubleSigningValidateFillsDetectionWindow(t *testing.T) {
	s := NewDoubleSigningScenario([]string{"val-0"}, time.Minute)
	s.DetectionWindow = 0

	if err := s.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if s.DetectionWindow != 30*time.Second {
		t.Errorf("DetectionWindow = %v after Validate on a zero value, want %v",
			s.DetectionWindow, 30*time.Second)
	}

	// A positive window must be left alone.
	s2 := NewDoubleSigningScenario([]string{"val-0"}, time.Minute)
	s2.DetectionWindow = 5 * time.Second
	if err := s2.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if s2.DetectionWindow != 5*time.Second {
		t.Errorf("DetectionWindow = %v, want %v (an explicit value was overwritten)",
			s2.DetectionWindow, 5*time.Second)
	}
}
