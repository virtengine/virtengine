// Package types contains proto.Message stub implementations for the benchmark module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgSubmitBenchmarks
func (m *MsgSubmitBenchmarks) ProtoMessage()  {}
func (m *MsgSubmitBenchmarks) Reset()         { *m = MsgSubmitBenchmarks{} }
func (m *MsgSubmitBenchmarks) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRequestChallenge
func (m *MsgRequestChallenge) ProtoMessage()  {}
func (m *MsgRequestChallenge) Reset()         { *m = MsgRequestChallenge{} }
func (m *MsgRequestChallenge) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRespondChallenge
func (m *MsgRespondChallenge) ProtoMessage()  {}
func (m *MsgRespondChallenge) Reset()         { *m = MsgRespondChallenge{} }
func (m *MsgRespondChallenge) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgFlagProvider
func (m *MsgFlagProvider) ProtoMessage()  {}
func (m *MsgFlagProvider) Reset()         { *m = MsgFlagProvider{} }
func (m *MsgFlagProvider) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUnflagProvider
func (m *MsgUnflagProvider) ProtoMessage()  {}
func (m *MsgUnflagProvider) Reset()         { *m = MsgUnflagProvider{} }
func (m *MsgUnflagProvider) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgResolveAnomalyFlag
func (m *MsgResolveAnomalyFlag) ProtoMessage()  {}
func (m *MsgResolveAnomalyFlag) Reset()         { *m = MsgResolveAnomalyFlag{} }
func (m *MsgResolveAnomalyFlag) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// Proto.Message interface stubs for MsgSubmitBenchmarksResponse
type MsgSubmitBenchmarksResponse struct{}

func (m *MsgSubmitBenchmarksResponse) ProtoMessage()  {}
func (m *MsgSubmitBenchmarksResponse) Reset()         { *m = MsgSubmitBenchmarksResponse{} }
func (m *MsgSubmitBenchmarksResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRequestChallengeResponse
type MsgRequestChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
}

func (m *MsgRequestChallengeResponse) ProtoMessage()  {}
func (m *MsgRequestChallengeResponse) Reset()         { *m = MsgRequestChallengeResponse{} }
func (m *MsgRequestChallengeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRespondChallengeResponse
type MsgRespondChallengeResponse struct{}

func (m *MsgRespondChallengeResponse) ProtoMessage()  {}
func (m *MsgRespondChallengeResponse) Reset()         { *m = MsgRespondChallengeResponse{} }
func (m *MsgRespondChallengeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgFlagProviderResponse
type MsgFlagProviderResponse struct{}

func (m *MsgFlagProviderResponse) ProtoMessage()  {}
func (m *MsgFlagProviderResponse) Reset()         { *m = MsgFlagProviderResponse{} }
func (m *MsgFlagProviderResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUnflagProviderResponse
type MsgUnflagProviderResponse struct{}

func (m *MsgUnflagProviderResponse) ProtoMessage()  {}
func (m *MsgUnflagProviderResponse) Reset()         { *m = MsgUnflagProviderResponse{} }
func (m *MsgUnflagProviderResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgResolveAnomalyFlagResponse
type MsgResolveAnomalyFlagResponse struct{}

func (m *MsgResolveAnomalyFlagResponse) ProtoMessage()  {}
func (m *MsgResolveAnomalyFlagResponse) Reset()         { *m = MsgResolveAnomalyFlagResponse{} }
func (m *MsgResolveAnomalyFlagResponse) String() string { return fmt.Sprintf("%+v", *m) }
