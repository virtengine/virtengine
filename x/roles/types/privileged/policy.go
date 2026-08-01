package privileged

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
)

// ScopeAtom names one exact operation on one exact resource.
type ScopeAtom struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Operation    string `json:"operation"`
}

func (s ScopeAtom) key() string { return s.ResourceType + "\x00" + s.ResourceID + "\x00" + s.Operation }

func (s ScopeAtom) Validate() error {
	if invalidExactValue(s.ResourceType) || invalidExactValue(s.ResourceID) || invalidExactValue(s.Operation) {
		return fmt.Errorf("scope atom requires exact resource type, resource ID, and operation")
	}
	return nil
}

func validateScope(scope []ScopeAtom) error {
	if len(scope) == 0 {
		return fmt.Errorf("at least one exact scope atom is required")
	}
	seen := make(map[string]struct{}, len(scope))
	for _, atom := range scope {
		if err := atom.Validate(); err != nil {
			return err
		}
		if _, exists := seen[atom.key()]; exists {
			return fmt.Errorf("duplicate scope atom %q", atom.key())
		}
		seen[atom.key()] = struct{}{}
	}
	return nil
}

func ScopeDigest(action string, scope []ScopeAtom) ([32]byte, error) {
	if invalidExactValue(action) {
		return [32]byte{}, fmt.Errorf("exact action is required")
	}
	if err := validateScope(scope); err != nil {
		return [32]byte{}, err
	}
	var output bytes.Buffer
	_ = writeString(&output, "virtengine.roles.privileged/action/v1")
	_ = writeString(&output, action)
	for _, atom := range sortedScope(scope) {
		_ = writeString(&output, atom.ResourceType)
		_ = writeString(&output, atom.ResourceID)
		_ = writeString(&output, atom.Operation)
	}
	return sha256.Sum256(output.Bytes()), nil
}

func scopeSubset(candidate, allowed []ScopeAtom) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, atom := range allowed {
		set[atom.key()] = struct{}{}
	}
	for _, atom := range candidate {
		if _, exists := set[atom.key()]; !exists {
			return false
		}
	}
	return true
}

// ActionPolicy binds one action to exact scope, approval, and MFA policies.
type ActionPolicy struct {
	PolicyID                    string      `json:"policy_id"`
	Version                     uint64      `json:"version"`
	Revision                    uint64      `json:"revision"`
	Action                      string      `json:"action"`
	Scope                       []ScopeAtom `json:"scope"`
	ApprovalPolicyID            string      `json:"approval_policy_id"`
	MFARequirementID            string      `json:"mfa_requirement_id"`
	MaximumEmergencyDurationSec int64       `json:"maximum_emergency_duration_seconds"`
}

func (p ActionPolicy) Validate() error {
	if invalidExactValue(p.PolicyID) || p.Version == 0 || p.Revision == 0 || invalidExactValue(p.Action) {
		return fmt.Errorf("action policy identity, version, revision, and exact action are required")
	}
	if invalidExactValue(p.ApprovalPolicyID) || invalidExactValue(p.MFARequirementID) {
		return fmt.Errorf("approval and MFA policy bindings are required")
	}
	if p.MaximumEmergencyDurationSec <= 0 {
		return fmt.Errorf("positive emergency duration limit is required")
	}
	return validateScope(p.Scope)
}

// PolicyRegistry is immutable at a version and revision.
type PolicyRegistry struct {
	ContractVersion string         `json:"contract_version"`
	RegistryID      string         `json:"registry_id"`
	Version         uint64         `json:"version"`
	Revision        uint64         `json:"revision"`
	Policies        []ActionPolicy `json:"policies"`
}

func (r PolicyRegistry) Validate() error {
	if r.ContractVersion != ContractVersion || invalidExactValue(r.RegistryID) || r.Version == 0 || r.Revision == 0 {
		return fmt.Errorf("registry contract, identity, version, and revision are required")
	}
	if len(r.Policies) == 0 {
		return fmt.Errorf("registry requires action policies")
	}
	seenID := map[string]struct{}{}
	seenAction := map[string]struct{}{}
	for _, policy := range r.Policies {
		if err := policy.Validate(); err != nil {
			return err
		}
		if _, exists := seenID[policy.PolicyID]; exists {
			return fmt.Errorf("duplicate policy ID %q", policy.PolicyID)
		}
		if _, exists := seenAction[policy.Action]; exists {
			return fmt.Errorf("duplicate action %q", policy.Action)
		}
		seenID[policy.PolicyID] = struct{}{}
		seenAction[policy.Action] = struct{}{}
	}
	return nil
}

func (r PolicyRegistry) Resolve(action string, scope []ScopeAtom) (ActionPolicy, error) {
	if err := r.Validate(); err != nil {
		return ActionPolicy{}, err
	}
	for _, policy := range r.Policies {
		if policy.Action == action {
			if len(scope) != len(policy.Scope) || !scopeSubset(scope, policy.Scope) {
				return ActionPolicy{}, fmt.Errorf("requested scope does not exactly match policy")
			}
			return policy, nil
		}
	}
	return ActionPolicy{}, fmt.Errorf("action %q is not registered", action)
}

func (r PolicyRegistry) Digest() ([32]byte, error) {
	if err := r.Validate(); err != nil {
		return [32]byte{}, err
	}
	policies := append([]ActionPolicy(nil), r.Policies...)
	sort.Slice(policies, func(i, j int) bool { return policies[i].PolicyID < policies[j].PolicyID })
	var output bytes.Buffer
	_ = writeString(&output, r.ContractVersion)
	_ = writeString(&output, r.RegistryID)
	writeUint64(&output, r.Version)
	writeUint64(&output, r.Revision)
	for _, policy := range policies {
		_ = writeString(&output, policy.PolicyID)
		writeUint64(&output, policy.Version)
		writeUint64(&output, policy.Revision)
		_ = writeString(&output, policy.Action)
		for _, atom := range sortedScope(policy.Scope) {
			_ = writeString(&output, atom.ResourceType)
			_ = writeString(&output, atom.ResourceID)
			_ = writeString(&output, atom.Operation)
		}
		_ = writeString(&output, policy.ApprovalPolicyID)
		_ = writeString(&output, policy.MFARequirementID)
		writeInt64(&output, policy.MaximumEmergencyDurationSec)
	}
	return sha256.Sum256(output.Bytes()), nil
}
