// Package v1 provides additional methods for generated benchmark types.
package v1

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Equal methods for genesis state types

// Equal returns true if Params are equal
func (m *Params) Equal(that *Params) bool {
	if m == nil && that == nil {
		return true
	}
	if m == nil || that == nil {
		return false
	}
	if m.ChallengeTimeout != that.ChallengeTimeout {
		return false
	}
	if m.BenchmarkValidityPeriod != that.BenchmarkValidityPeriod {
		return false
	}
	if m.MinBenchmarkScore != that.MinBenchmarkScore {
		return false
	}
	if m.MaxAnomalyFlags != that.MaxAnomalyFlags {
		return false
	}
	return true
}

// Equal returns true if ProviderBenchmarks are equal
func (m *ProviderBenchmark) Equal(that *ProviderBenchmark) bool {
	if m == nil && that == nil {
		return true
	}
	if m == nil || that == nil {
		return false
	}
	if m.Provider != that.Provider {
		return false
	}
	if m.ReliabilityScore != that.ReliabilityScore {
		return false
	}
	if m.LastUpdated != that.LastUpdated {
		return false
	}
	if len(m.Results) != len(that.Results) {
		return false
	}
	for i := range m.Results {
		if !m.Results[i].Equal(&that.Results[i]) {
			return false
		}
	}
	return true
}

// Equal returns true if BenchmarkResults are equal
func (m *BenchmarkResult) Equal(that *BenchmarkResult) bool {
	if m == nil && that == nil {
		return true
	}
	if m == nil || that == nil {
		return false
	}
	if m.BenchmarkType != that.BenchmarkType {
		return false
	}
	if m.Score != that.Score {
		return false
	}
	if m.Timestamp != that.Timestamp {
		return false
	}
	if m.HardwareInfo != that.HardwareInfo {
		return false
	}
	if !bytes.Equal(m.RawData, that.RawData) {
		return false
	}
	return true
}

// Equal returns true if Challenges are equal
func (m *Challenge) Equal(that *Challenge) bool {
	if m == nil && that == nil {
		return true
	}
	if m == nil || that == nil {
		return false
	}
	if m.ChallengeId != that.ChallengeId {
		return false
	}
	if m.Requester != that.Requester {
		return false
	}
	if m.Provider != that.Provider {
		return false
	}
	if m.BenchmarkType != that.BenchmarkType {
		return false
	}
	if m.Status != that.Status {
		return false
	}
	if m.RequestedAt != that.RequestedAt {
		return false
	}
	if m.ExpiresAt != that.ExpiresAt {
		return false
	}
	// Compare Response pointers
	if m.Response == nil && that.Response != nil {
		return false
	}
	if m.Response != nil && that.Response == nil {
		return false
	}
	if m.Response != nil && that.Response != nil && !m.Response.Equal(that.Response) {
		return false
	}
	return true
}

// Equal returns true if AnomalyFlags are equal
func (m *AnomalyFlag) Equal(that *AnomalyFlag) bool {
	if m == nil && that == nil {
		return true
	}
	if m == nil || that == nil {
		return false
	}
	if m.Provider != that.Provider {
		return false
	}
	if m.Reporter != that.Reporter {
		return false
	}
	if m.Reason != that.Reason {
		return false
	}
	if m.Evidence != that.Evidence {
		return false
	}
	if m.Status != that.Status {
		return false
	}
	if m.FlaggedAt != that.FlaggedAt {
		return false
	}
	if m.Resolution != that.Resolution {
		return false
	}
	if m.ResolvedAt != that.ResolvedAt {
		return false
	}
	return true
}

// sdk.Msg interface methods for MsgSubmitBenchmarks

func (msg *MsgSubmitBenchmarks) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidProvider.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidProvider.Wrapf("invalid provider address: %v", err)
	}

	if len(msg.Results) == 0 {
		return ErrMissingResults.Wrap("at least one benchmark result is required")
	}

	return nil
}

func (msg *MsgSubmitBenchmarks) GetSigners() []sdk.AccAddress {
	provider, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{provider}
}

// sdk.Msg interface methods for MsgRequestChallenge

func (msg *MsgRequestChallenge) ValidateBasic() error {
	if msg.Requester == "" {
		return ErrInvalidRequester.Wrap("requester address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Requester); err != nil {
		return ErrInvalidRequester.Wrapf("invalid requester address: %v", err)
	}

	if msg.Provider == "" {
		return ErrInvalidProvider.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidProvider.Wrapf("invalid provider address: %v", err)
	}

	if msg.BenchmarkType == "" {
		return ErrInvalidBenchmarkType.Wrap("benchmark type is required")
	}

	return nil
}

func (msg *MsgRequestChallenge) GetSigners() []sdk.AccAddress {
	requester, _ := sdk.AccAddressFromBech32(msg.Requester)
	return []sdk.AccAddress{requester}
}

// sdk.Msg interface methods for MsgRespondChallenge

func (msg *MsgRespondChallenge) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidProvider.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidProvider.Wrapf("invalid provider address: %v", err)
	}

	if msg.ChallengeId == "" {
		return ErrInvalidChallengeID.Wrap("challenge ID is required")
	}

	return nil
}

func (msg *MsgRespondChallenge) GetSigners() []sdk.AccAddress {
	provider, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{provider}
}

// sdk.Msg interface methods for MsgFlagProvider

func (msg *MsgFlagProvider) ValidateBasic() error {
	if msg.Reporter == "" {
		return ErrInvalidReporter.Wrap("reporter address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Reporter); err != nil {
		return ErrInvalidReporter.Wrapf("invalid reporter address: %v", err)
	}

	if msg.Provider == "" {
		return ErrInvalidProvider.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidProvider.Wrapf("invalid provider address: %v", err)
	}

	if msg.Reason == "" {
		return ErrInvalidReason.Wrap("reason is required")
	}

	if len(msg.Reason) > MaxReasonLength {
		return ErrReasonTooLong.Wrapf("maximum length is %d characters", MaxReasonLength)
	}

	return nil
}

func (msg *MsgFlagProvider) GetSigners() []sdk.AccAddress {
	reporter, _ := sdk.AccAddressFromBech32(msg.Reporter)
	return []sdk.AccAddress{reporter}
}

// sdk.Msg interface methods for MsgUnflagProvider

func (msg *MsgUnflagProvider) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAuthority.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAuthority.Wrapf("invalid authority address: %v", err)
	}

	if msg.Provider == "" {
		return ErrInvalidProvider.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidProvider.Wrapf("invalid provider address: %v", err)
	}

	return nil
}

func (msg *MsgUnflagProvider) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgResolveAnomalyFlag

func (msg *MsgResolveAnomalyFlag) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAuthority.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAuthority.Wrapf("invalid authority address: %v", err)
	}

	if msg.Provider == "" {
		return ErrInvalidProvider.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidProvider.Wrapf("invalid provider address: %v", err)
	}

	if msg.Resolution == "" {
		return ErrInvalidResolution.Wrap("resolution is required")
	}

	if len(msg.Resolution) > MaxResolutionLength {
		return ErrResolutionTooLong.Wrapf("maximum length is %d characters", MaxResolutionLength)
	}

	return nil
}

func (msg *MsgResolveAnomalyFlag) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

