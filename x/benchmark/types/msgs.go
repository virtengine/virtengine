// Package types contains types for the Benchmark module.
//
// VE-601: Benchmark module messages
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for the benchmark module
const (
	TypeMsgSubmitBenchmarks     = "submit_benchmarks"
	TypeMsgRequestChallenge     = "request_challenge"
	TypeMsgRespondChallenge     = "respond_challenge"
	TypeMsgFlagProvider         = "flag_provider"
	TypeMsgUnflagProvider       = "unflag_provider"
	TypeMsgResolveAnomalyFlag   = "resolve_anomaly_flag"
	TypeMsgUpdateReliability    = "update_reliability"
)

// MsgSubmitBenchmarks is a message to submit benchmark reports
type MsgSubmitBenchmarks struct {
	// ProviderAddress is the provider submitting the benchmarks
	ProviderAddress string `json:"provider_address"`

	// Reports contains one or more benchmark reports
	Reports []BenchmarkReport `json:"reports"`
}

// Route implements sdk.Msg
func (m *MsgSubmitBenchmarks) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgSubmitBenchmarks) Type() string { return TypeMsgSubmitBenchmarks }

// ValidateBasic implements sdk.Msg
func (m *MsgSubmitBenchmarks) ValidateBasic() error {
	if m.ProviderAddress == "" {
		return ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ProviderAddress); err != nil {
		return ErrInvalidBenchmark.Wrapf("invalid provider_address: %v", err)
	}

	if len(m.Reports) == 0 {
		return ErrInvalidBenchmark.Wrap("at least one report is required")
	}

	if len(m.Reports) > 10 {
		return ErrInvalidBenchmark.Wrap("too many reports (max 10)")
	}

	for i, report := range m.Reports {
		if err := report.Validate(); err != nil {
			return ErrInvalidBenchmark.Wrapf("report %d: %v", i, err)
		}

		if report.ProviderAddress != m.ProviderAddress {
			return ErrInvalidBenchmark.Wrapf("report %d: provider_address mismatch", i)
		}
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgSubmitBenchmarks) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m *MsgSubmitBenchmarks) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// MsgRequestChallenge is a message to request a benchmark challenge
type MsgRequestChallenge struct {
	// Requester is the address requesting the challenge
	Requester string `json:"requester"`

	// ProviderAddress is the provider to challenge
	ProviderAddress string `json:"provider_address"`

	// ClusterID is the cluster to benchmark
	ClusterID string `json:"cluster_id"`

	// OfferingID is the optional offering to benchmark
	OfferingID string `json:"offering_id,omitempty"`

	// SuiteVersion is the required benchmark suite version
	SuiteVersion string `json:"suite_version"`

	// DeadlineSeconds is the deadline in seconds from now
	DeadlineSeconds int64 `json:"deadline_seconds"`
}

// Route implements sdk.Msg
func (m *MsgRequestChallenge) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgRequestChallenge) Type() string { return TypeMsgRequestChallenge }

// ValidateBasic implements sdk.Msg
func (m *MsgRequestChallenge) ValidateBasic() error {
	if m.Requester == "" {
		return ErrInvalidBenchmark.Wrap("requester cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Requester); err != nil {
		return ErrInvalidBenchmark.Wrapf("invalid requester: %v", err)
	}

	if m.ProviderAddress == "" {
		return ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
	}

	if m.ClusterID == "" {
		return ErrInvalidBenchmark.Wrap("cluster_id cannot be empty")
	}

	if m.SuiteVersion == "" {
		return ErrInvalidBenchmark.Wrap("suite_version cannot be empty")
	}

	if m.DeadlineSeconds <= 0 || m.DeadlineSeconds > 86400*7 {
		return ErrInvalidBenchmark.Wrap("deadline_seconds must be between 1 and 604800 (7 days)")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgRequestChallenge) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Requester)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m *MsgRequestChallenge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// MsgRespondChallenge is a message to respond to a benchmark challenge
type MsgRespondChallenge struct {
	// ProviderAddress is the provider responding
	ProviderAddress string `json:"provider_address"`

	// ChallengeID is the challenge being responded to
	ChallengeID string `json:"challenge_id"`

	// Report is the benchmark report
	Report BenchmarkReport `json:"report"`

	// ExplanationRef is an optional encrypted explanation reference
	ExplanationRef string `json:"explanation_ref,omitempty"`
}

// Route implements sdk.Msg
func (m *MsgRespondChallenge) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgRespondChallenge) Type() string { return TypeMsgRespondChallenge }

// ValidateBasic implements sdk.Msg
func (m *MsgRespondChallenge) ValidateBasic() error {
	if m.ProviderAddress == "" {
		return ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ProviderAddress); err != nil {
		return ErrInvalidBenchmark.Wrapf("invalid provider_address: %v", err)
	}

	if m.ChallengeID == "" {
		return ErrInvalidBenchmark.Wrap("challenge_id cannot be empty")
	}

	if err := m.Report.Validate(); err != nil {
		return ErrInvalidBenchmark.Wrapf("report: %v", err)
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgRespondChallenge) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m *MsgRespondChallenge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// MsgFlagProvider is a message to flag a provider
type MsgFlagProvider struct {
	// Moderator is the moderator flagging the provider
	Moderator string `json:"moderator"`

	// ProviderAddress is the provider being flagged
	ProviderAddress string `json:"provider_address"`

	// Reason is the reason for flagging
	Reason string `json:"reason"`

	// ExpiresInSeconds is when the flag expires (0 = permanent)
	ExpiresInSeconds int64 `json:"expires_in_seconds,omitempty"`
}

// Route implements sdk.Msg
func (m *MsgFlagProvider) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgFlagProvider) Type() string { return TypeMsgFlagProvider }

// ValidateBasic implements sdk.Msg
func (m *MsgFlagProvider) ValidateBasic() error {
	if m.Moderator == "" {
		return ErrUnauthorized.Wrap("moderator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorized.Wrapf("invalid moderator: %v", err)
	}

	if m.ProviderAddress == "" {
		return ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
	}

	if m.Reason == "" {
		return ErrInvalidBenchmark.Wrap("reason cannot be empty")
	}

	if len(m.Reason) > 500 {
		return ErrInvalidBenchmark.Wrap("reason exceeds maximum length (500)")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgFlagProvider) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m *MsgFlagProvider) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// MsgUnflagProvider is a message to remove a provider flag
type MsgUnflagProvider struct {
	// Moderator is the moderator removing the flag
	Moderator string `json:"moderator"`

	// ProviderAddress is the provider being unflagged
	ProviderAddress string `json:"provider_address"`
}

// Route implements sdk.Msg
func (m *MsgUnflagProvider) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgUnflagProvider) Type() string { return TypeMsgUnflagProvider }

// ValidateBasic implements sdk.Msg
func (m *MsgUnflagProvider) ValidateBasic() error {
	if m.Moderator == "" {
		return ErrUnauthorized.Wrap("moderator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorized.Wrapf("invalid moderator: %v", err)
	}

	if m.ProviderAddress == "" {
		return ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgUnflagProvider) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m *MsgUnflagProvider) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// MsgResolveAnomalyFlag is a message to resolve an anomaly flag
type MsgResolveAnomalyFlag struct {
	// Moderator is the moderator resolving the flag
	Moderator string `json:"moderator"`

	// FlagID is the anomaly flag being resolved
	FlagID string `json:"flag_id"`

	// Resolution is the resolution details
	Resolution string `json:"resolution"`
}

// Route implements sdk.Msg
func (m *MsgResolveAnomalyFlag) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgResolveAnomalyFlag) Type() string { return TypeMsgResolveAnomalyFlag }

// ValidateBasic implements sdk.Msg
func (m *MsgResolveAnomalyFlag) ValidateBasic() error {
	if m.Moderator == "" {
		return ErrUnauthorized.Wrap("moderator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorized.Wrapf("invalid moderator: %v", err)
	}

	if m.FlagID == "" {
		return ErrInvalidBenchmark.Wrap("flag_id cannot be empty")
	}

	if m.Resolution == "" {
		return ErrInvalidBenchmark.Wrap("resolution cannot be empty")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgResolveAnomalyFlag) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m *MsgResolveAnomalyFlag) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}
