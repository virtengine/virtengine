// Package types contains types for the Benchmark module.
//
// VE-601: Benchmark module messages
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
)

// Type aliases to generated protobuf types
type (
	MsgSubmitBenchmarks           = benchmarkv1.MsgSubmitBenchmarks
	MsgSubmitBenchmarksResponse   = benchmarkv1.MsgSubmitBenchmarksResponse
	MsgRequestChallenge           = benchmarkv1.MsgRequestChallenge
	MsgRequestChallengeResponse   = benchmarkv1.MsgRequestChallengeResponse
	MsgRespondChallenge           = benchmarkv1.MsgRespondChallenge
	MsgRespondChallengeResponse   = benchmarkv1.MsgRespondChallengeResponse
	MsgFlagProvider               = benchmarkv1.MsgFlagProvider
	MsgFlagProviderResponse       = benchmarkv1.MsgFlagProviderResponse
	MsgUnflagProvider             = benchmarkv1.MsgUnflagProvider
	MsgUnflagProviderResponse     = benchmarkv1.MsgUnflagProviderResponse
	MsgResolveAnomalyFlag         = benchmarkv1.MsgResolveAnomalyFlag
	MsgResolveAnomalyFlagResponse = benchmarkv1.MsgResolveAnomalyFlagResponse
)

// Message type constants
const (
	TypeMsgSubmitBenchmarks   = "submit_benchmarks"
	TypeMsgRequestChallenge   = "request_challenge"
	TypeMsgRespondChallenge   = "respond_challenge"
	TypeMsgFlagProvider       = "flag_provider"
	TypeMsgUnflagProvider     = "unflag_provider"
	TypeMsgResolveAnomalyFlag = "resolve_anomaly_flag"
)

var (
	_ sdk.Msg = &MsgSubmitBenchmarks{}
	_ sdk.Msg = &MsgRequestChallenge{}
	_ sdk.Msg = &MsgRespondChallenge{}
	_ sdk.Msg = &MsgFlagProvider{}
	_ sdk.Msg = &MsgUnflagProvider{}
	_ sdk.Msg = &MsgResolveAnomalyFlag{}
)

// NewMsgSubmitBenchmarks creates a new MsgSubmitBenchmarks
func NewMsgSubmitBenchmarks(provider, clusterID string, results []benchmarkv1.BenchmarkResult, signature []byte) *MsgSubmitBenchmarks {
	return &MsgSubmitBenchmarks{
		Provider:  provider,
		ClusterId: clusterID,
		Results:   results,
		Signature: signature,
	}
}

// NewMsgRequestChallenge creates a new MsgRequestChallenge
func NewMsgRequestChallenge(requester, provider, benchmarkType string) *MsgRequestChallenge {
	return &MsgRequestChallenge{
		Requester:     requester,
		Provider:      provider,
		BenchmarkType: benchmarkType,
	}
}

// NewMsgRespondChallenge creates a new MsgRespondChallenge
func NewMsgRespondChallenge(provider, challengeID string, result benchmarkv1.BenchmarkResult, signature []byte) *MsgRespondChallenge {
	return &MsgRespondChallenge{
		Provider:    provider,
		ChallengeId: challengeID,
		Result:      result,
		Signature:   signature,
	}
}

// NewMsgFlagProvider creates a new MsgFlagProvider
func NewMsgFlagProvider(reporter, provider, reason, evidence string) *MsgFlagProvider {
	return &MsgFlagProvider{
		Reporter: reporter,
		Provider: provider,
		Reason:   reason,
		Evidence: evidence,
	}
}

// NewMsgUnflagProvider creates a new MsgUnflagProvider
func NewMsgUnflagProvider(authority, provider string) *MsgUnflagProvider {
	return &MsgUnflagProvider{
		Authority: authority,
		Provider:  provider,
	}
}

// NewMsgResolveAnomalyFlag creates a new MsgResolveAnomalyFlag
func NewMsgResolveAnomalyFlag(authority, provider, resolution string, isValid bool) *MsgResolveAnomalyFlag {
	return &MsgResolveAnomalyFlag{
		Authority:  authority,
		Provider:   provider,
		Resolution: resolution,
		IsValid:    isValid,
	}
}
