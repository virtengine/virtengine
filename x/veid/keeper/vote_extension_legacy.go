package keeper

import (
	"encoding/json"
	"time"

	"github.com/virtengine/virtengine/x/veid/types"
)

// VoteExtension is the legacy JSON helper retained for source compatibility.
// Consensus handlers use VEIDVoteExtension protobuf messages exclusively.
type VoteExtension struct {
	Version             uint8                 `json:"version"`
	Height              int64                 `json:"height"`
	ValidatorAddress    string                `json:"validator_address"`
	VerificationResults []VoteExtensionResult `json:"verification_results,omitempty"`
	Timestamp           time.Time             `json:"timestamp"`
	ModelVersion        string                `json:"model_version"`
}

// VoteExtensionResult is the legacy compact JSON result.
type VoteExtensionResult struct {
	RequestID  string                         `json:"request_id"`
	Score      uint32                         `json:"score"`
	Status     types.VerificationResultStatus `json:"status"`
	InputHash  []byte                         `json:"input_hash"`
	ResultHash []byte                         `json:"result_hash"`
}

// NewVoteExtension creates a legacy helper with deterministic time.
func NewVoteExtension(height int64, validatorAddress, modelVersion string) *VoteExtension {
	return NewVoteExtensionWithTime(height, validatorAddress, modelVersion, time.Unix(0, 0))
}

// NewVoteExtensionWithTime creates a legacy helper at an explicit time.
func NewVoteExtensionWithTime(height int64, validatorAddress, modelVersion string, timestamp time.Time) *VoteExtension {
	return &VoteExtension{
		Version:             uint8(VoteExtensionVersion), //nolint:gosec // protocol version is bounded to one byte
		Height:              height,
		ValidatorAddress:    validatorAddress,
		VerificationResults: make([]VoteExtensionResult, 0),
		Timestamp:           timestamp.UTC(),
		ModelVersion:        modelVersion,
	}
}

// AddResult adds a result to the legacy helper.
func (ve *VoteExtension) AddResult(result types.VerificationResult) {
	inputHash := result.InputHash
	if len(inputHash) > 8 {
		inputHash = inputHash[:8]
	}
	ve.VerificationResults = append(ve.VerificationResults, VoteExtensionResult{
		RequestID:  result.RequestID,
		Score:      result.Score,
		Status:     result.Status,
		InputHash:  inputHash,
		ResultHash: ComputeResultHash(result),
	})
}

// Marshal serializes the inactive legacy helper.
func (ve *VoteExtension) Marshal() ([]byte, error) { return json.Marshal(ve) }

// UnmarshalVoteExtension decodes the inactive legacy helper.
func UnmarshalVoteExtension(bz []byte) (*VoteExtension, error) {
	var extension VoteExtension
	if err := json.Unmarshal(bz, &extension); err != nil {
		return nil, err
	}
	return &extension, nil
}
