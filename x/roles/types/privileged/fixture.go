package privileged

// Fixture bundles deterministic contracts while keeping every activation path denied.
type Fixture struct {
	Gate             FeatureGate      `json:"gate"`
	Registry         PolicyRegistry   `json:"registry"`
	ApprovalPolicies []ApprovalPolicy `json:"approval_policies"`
	MFARequirements  []MFARequirement `json:"mfa_requirements"`
}

func DeterministicFixture() Fixture {
	scope := []ScopeAtom{{ResourceType: "account", ResourceID: "fixture-target-001", Operation: "suspend"}}
	actionDigest, _ := ScopeDigest("account.suspend", scope)
	return Fixture{
		Gate: FeatureGate{
			State: FeatureFixtureOnly,
			Blockers: []string{
				"T4 exact-SHA integration gate is absent",
				"durable MFA replay and custody dependencies are absent",
				"privileged routes are not registered",
			},
		},
		Registry: PolicyRegistry{
			ContractVersion: ContractVersion,
			RegistryID:      "fixture-privileged-registry",
			Version:         1,
			Revision:        1,
			Policies: []ActionPolicy{{
				PolicyID:                    "fixture-account-suspend",
				Version:                     1,
				Revision:                    1,
				Action:                      "account.suspend",
				Scope:                       scope,
				ApprovalPolicyID:            "fixture-approval-policy",
				MFARequirementID:            "fixture-mfa-requirement",
				MaximumEmergencyDurationSec: 900,
			}},
		},
		ApprovalPolicies: []ApprovalPolicy{{
			PolicyID:  "fixture-approval-policy",
			Version:   1,
			Revision:  1,
			Threshold: 2,
			Members: []ApprovalMember{
				{AccountID: "fixture-approver-a", RoleID: "security-reviewer", Weight: 1},
				{AccountID: "fixture-approver-b", RoleID: "security-reviewer", Weight: 1},
				{AccountID: "fixture-approver-c", RoleID: "auditor", Weight: 1},
			},
		}},
		MFARequirements: []MFARequirement{{
			RequirementID:       "fixture-mfa-requirement",
			ContractVersion:     "virtengine.mfa.factor-profile/v1",
			ProfileID:           "fixture-strong-factor",
			ChallengeVersion:    "virtengine.mfa.challenge/v1",
			VerifierKeyEpoch:    1,
			RootEpoch:           1,
			ProofEpoch:          1,
			MinimumProofs:       1,
			RequireStrongFactor: true,
			ActionDigest:        actionDigest,
		}},
	}
}

func (f Fixture) Validate() error {
	if err := f.Gate.Validate(); err != nil {
		return err
	}
	if err := f.Registry.Validate(); err != nil {
		return err
	}
	for _, policy := range f.ApprovalPolicies {
		if err := policy.Validate(); err != nil {
			return err
		}
	}
	for _, requirement := range f.MFARequirements {
		if err := requirement.Validate(); err != nil {
			return err
		}
	}
	return nil
}
