// Package types contains types for the Benchmark module.
//
// VE-601: Genesis state and parameters
package types

import (
	"fmt"
)

// DefaultRetentionCount is the default number of reports to retain per cluster
const DefaultRetentionCount = 100

// DefaultChallengeDeadlineSeconds is the default challenge deadline
const DefaultChallengeDeadlineSeconds = 86400 // 24 hours

// Params contains the benchmark module parameters
type Params struct {
	// RetentionCount is the number of reports to retain per cluster
	RetentionCount int64 `json:"retention_count"`

	// DefaultChallengeDeadlineSeconds is the default challenge deadline
	DefaultChallengeDeadlineSeconds int64 `json:"default_challenge_deadline_seconds"`

	// MinBenchmarkInterval is the minimum interval between benchmarks in seconds
	MinBenchmarkInterval int64 `json:"min_benchmark_interval"`

	// MaxReportsPerSubmission is the maximum reports per submission
	MaxReportsPerSubmission int64 `json:"max_reports_per_submission"`

	// AnomalyThresholdJumpPercent is the threshold for sudden jump anomalies (fixed-point * 100)
	AnomalyThresholdJumpPercent int64 `json:"anomaly_threshold_jump_percent"`

	// AnomalyThresholdRepeatCount is the count threshold for repeated outputs
	AnomalyThresholdRepeatCount int64 `json:"anomaly_threshold_repeat_count"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		RetentionCount:                  DefaultRetentionCount,
		DefaultChallengeDeadlineSeconds: DefaultChallengeDeadlineSeconds,
		MinBenchmarkInterval:            300, // 5 minutes
		MaxReportsPerSubmission:         10,
		AnomalyThresholdJumpPercent:     50, // 50%
		AnomalyThresholdRepeatCount:     3,  // 3 identical outputs
	}
}

// Validate validates the parameters
func (p *Params) Validate() error {
	if p.RetentionCount <= 0 {
		return fmt.Errorf("retention_count must be positive")
	}

	if p.DefaultChallengeDeadlineSeconds <= 0 {
		return fmt.Errorf("default_challenge_deadline_seconds must be positive")
	}

	if p.MinBenchmarkInterval < 0 {
		return fmt.Errorf("min_benchmark_interval cannot be negative")
	}

	if p.MaxReportsPerSubmission <= 0 || p.MaxReportsPerSubmission > 100 {
		return fmt.Errorf("max_reports_per_submission must be between 1 and 100")
	}

	return nil
}

// GenesisState is the genesis state for the benchmark module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// Reports are the benchmark reports
	Reports []BenchmarkReport `json:"reports,omitempty"`

	// Scores are the reliability scores
	Scores []ReliabilityScore `json:"scores,omitempty"`

	// Challenges are pending challenges
	Challenges []BenchmarkChallenge `json:"challenges,omitempty"`

	// AnomalyFlags are anomaly flags
	AnomalyFlags []AnomalyFlag `json:"anomaly_flags,omitempty"`

	// ProviderFlags are provider moderation flags
	ProviderFlags []ProviderFlag `json:"provider_flags,omitempty"`

	// NextReportSequence is the next report sequence number
	NextReportSequence uint64 `json:"next_report_sequence"`

	// NextChallengeSequence is the next challenge sequence number
	NextChallengeSequence uint64 `json:"next_challenge_sequence"`

	// NextAnomalySequence is the next anomaly flag sequence number
	NextAnomalySequence uint64 `json:"next_anomaly_sequence"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                DefaultParams(),
		Reports:               []BenchmarkReport{},
		Scores:                []ReliabilityScore{},
		Challenges:            []BenchmarkChallenge{},
		AnomalyFlags:          []AnomalyFlag{},
		ProviderFlags:         []ProviderFlag{},
		NextReportSequence:    1,
		NextChallengeSequence: 1,
		NextAnomalySequence:   1,
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	reportIDs := make(map[string]bool)
	for i, report := range gs.Reports {
		if err := report.Validate(); err != nil {
			return fmt.Errorf("invalid report at index %d: %w", i, err)
		}
		if reportIDs[report.ReportID] {
			return fmt.Errorf("duplicate report_id: %s", report.ReportID)
		}
		reportIDs[report.ReportID] = true
	}

	scoreAddrs := make(map[string]bool)
	for i, score := range gs.Scores {
		if err := score.Validate(); err != nil {
			return fmt.Errorf("invalid score at index %d: %w", i, err)
		}
		if scoreAddrs[score.ProviderAddress] {
			return fmt.Errorf("duplicate provider_address in scores: %s", score.ProviderAddress)
		}
		scoreAddrs[score.ProviderAddress] = true
	}

	challengeIDs := make(map[string]bool)
	for i, challenge := range gs.Challenges {
		if err := challenge.Validate(); err != nil {
			return fmt.Errorf("invalid challenge at index %d: %w", i, err)
		}
		if challengeIDs[challenge.ChallengeID] {
			return fmt.Errorf("duplicate challenge_id: %s", challenge.ChallengeID)
		}
		challengeIDs[challenge.ChallengeID] = true
	}

	flagIDs := make(map[string]bool)
	for i, flag := range gs.AnomalyFlags {
		if err := flag.Validate(); err != nil {
			return fmt.Errorf("invalid anomaly flag at index %d: %w", i, err)
		}
		if flagIDs[flag.FlagID] {
			return fmt.Errorf("duplicate flag_id: %s", flag.FlagID)
		}
		flagIDs[flag.FlagID] = true
	}

	for i, flag := range gs.ProviderFlags {
		if err := flag.Validate(); err != nil {
			return fmt.Errorf("invalid provider flag at index %d: %w", i, err)
		}
	}

	return nil
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string { return fmt.Sprintf("%+v", *gs) }
