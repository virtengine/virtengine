// Package types provides VEID module types.
//
// This file defines model versioning for ML model consensus tracking.
// All validators must use the same model version for deterministic scoring.
//
// Task Reference: VE-3007 - Model Versioning and Governance
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// Model Types (Valid model types for registration)
// ============================================================================

// ModelType represents the type of ML model
type ModelType string

const (
	// ModelTypeTrustScore is the trust score calculation model
	ModelTypeTrustScore ModelType = "trust_score"

	// ModelTypeFaceVerification is the facial verification model
	ModelTypeFaceVerification ModelType = "face_verification"

	// ModelTypeLiveness is the liveness detection model
	ModelTypeLiveness ModelType = "liveness"

	// ModelTypeGANDetection is the GAN-generated image detection model
	ModelTypeGANDetection ModelType = "gan_detection"

	// ModelTypeOCR is the OCR extraction model
	ModelTypeOCR ModelType = "ocr"
)

// ValidModelTypes returns all valid model types
func ValidModelTypes() []ModelType {
	return []ModelType{
		ModelTypeTrustScore,
		ModelTypeFaceVerification,
		ModelTypeLiveness,
		ModelTypeGANDetection,
		ModelTypeOCR,
	}
}

// IsValidModelType checks if a model type is valid
func IsValidModelType(mt string) bool {
	for _, valid := range ValidModelTypes() {
		if string(valid) == mt {
			return true
		}
	}
	return false
}

// ============================================================================
// MLModelInfo - Registered ML Model Information
// ============================================================================

// MLModelInfo describes a registered ML model for version tracking
// This is separate from the pipeline ModelInfo which describes model weights/config
type MLModelInfo struct {
	// ModelID is the unique identifier for this model
	ModelID string `json:"model_id"`

	// Name is the human-readable name of the model
	Name string `json:"name"`

	// Version is the semantic version of the model (e.g., "1.0.0")
	Version string `json:"version"`

	// ModelType is the type of model (e.g., "face_verification")
	ModelType string `json:"model_type"`

	// SHA256Hash is the SHA256 hash of the model binary
	SHA256Hash string `json:"sha256_hash"`

	// Description is a human-readable description of the model
	Description string `json:"description"`

	// ActivatedAt is the block height when this model was activated
	ActivatedAt int64 `json:"activated_at"`

	// RegisteredAt is the block height when this model was registered
	RegisteredAt int64 `json:"registered_at"`

	// RegisteredBy is the address that registered this model
	RegisteredBy string `json:"registered_by"`

	// GovernanceID is the governance proposal ID that approved this model (if any)
	GovernanceID uint64 `json:"governance_id,omitempty"`

	// Status is the current status of the model
	Status ModelStatus `json:"status"`
}

// ModelStatus represents the lifecycle status of a model
type ModelStatus string

const (
	// ModelStatusPending indicates the model is pending activation
	ModelStatusPending ModelStatus = "pending"

	// ModelStatusActive indicates the model is currently active
	ModelStatusActive ModelStatus = "active"

	// ModelStatusDeprecated indicates the model has been deprecated
	ModelStatusDeprecated ModelStatus = "deprecated"

	// ModelStatusRevoked indicates the model has been revoked
	ModelStatusRevoked ModelStatus = "revoked"
)

// Validate validates the MLModelInfo fields
func (m MLModelInfo) Validate() error {
	if m.ModelID == "" {
		return fmt.Errorf("model_id cannot be empty")
	}
	if m.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if m.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	if !IsValidModelType(m.ModelType) {
		return fmt.Errorf("invalid model type: %s", m.ModelType)
	}
	if m.SHA256Hash == "" {
		return fmt.Errorf("sha256_hash cannot be empty")
	}
	// Validate hash is valid hex
	if _, err := hex.DecodeString(m.SHA256Hash); err != nil {
		return fmt.Errorf("sha256_hash must be valid hex: %w", err)
	}
	if len(m.SHA256Hash) != 64 {
		return fmt.Errorf("sha256_hash must be 64 hex characters (32 bytes)")
	}
	if m.RegisteredBy == "" {
		return fmt.Errorf("registered_by cannot be empty")
	}
	return nil
}

// ComputeMLModelID generates a deterministic model ID from name, version, and hash
func ComputeMLModelID(name, version, hash string) string {
	data := fmt.Sprintf("%s:%s:%s", name, version, hash)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:16]) // Use first 16 bytes for shorter ID
}

// ============================================================================
// ModelVersionState - Current Active Model Versions
// ============================================================================

// ModelVersionState tracks the current active model versions for consensus
type ModelVersionState struct {
	// TrustScoreModel is the active model ID for trust scoring
	TrustScoreModel string `json:"trust_score_model"`

	// FaceVerificationModel is the active model ID for face verification
	FaceVerificationModel string `json:"face_verification_model"`

	// LivenessModel is the active model ID for liveness detection
	LivenessModel string `json:"liveness_model"`

	// GANDetectionModel is the active model ID for GAN detection
	GANDetectionModel string `json:"gan_detection_model"`

	// OCRModel is the active model ID for OCR extraction
	OCRModel string `json:"ocr_model"`

	// LastUpdated is the block height when state was last updated
	LastUpdated int64 `json:"last_updated"`
}

// GetModelID returns the active model ID for a given model type
func (s ModelVersionState) GetModelID(modelType string) string {
	switch ModelType(modelType) {
	case ModelTypeTrustScore:
		return s.TrustScoreModel
	case ModelTypeFaceVerification:
		return s.FaceVerificationModel
	case ModelTypeLiveness:
		return s.LivenessModel
	case ModelTypeGANDetection:
		return s.GANDetectionModel
	case ModelTypeOCR:
		return s.OCRModel
	default:
		return ""
	}
}

// SetModelID sets the active model ID for a given model type
func (s *ModelVersionState) SetModelID(modelType string, modelID string) error {
	switch ModelType(modelType) {
	case ModelTypeTrustScore:
		s.TrustScoreModel = modelID
	case ModelTypeFaceVerification:
		s.FaceVerificationModel = modelID
	case ModelTypeLiveness:
		s.LivenessModel = modelID
	case ModelTypeGANDetection:
		s.GANDetectionModel = modelID
	case ModelTypeOCR:
		s.OCRModel = modelID
	default:
		return fmt.Errorf("invalid model type: %s", modelType)
	}
	return nil
}

// Validate validates the ModelVersionState
func (s ModelVersionState) Validate() error {
	// All model IDs are optional (can be empty initially)
	return nil
}

// DefaultModelVersionState returns the default (empty) model version state
func DefaultModelVersionState() ModelVersionState {
	return ModelVersionState{}
}

// ============================================================================
// ModelUpdateProposal - Governance Proposal for Model Updates
// ============================================================================

// ModelUpdateProposal for governance-controlled model updates
type ModelUpdateProposal struct {
	// Title is the proposal title
	Title string `json:"title"`

	// Description is the detailed proposal description
	Description string `json:"description"`

	// ModelType is the type of model being updated
	ModelType string `json:"model_type"`

	// NewModelID is the ID of the new model to activate
	NewModelID string `json:"new_model_id"`

	// NewModelHash is the SHA256 hash of the new model
	NewModelHash string `json:"new_model_hash"`

	// ActivationDelay is the number of blocks to wait after approval
	ActivationDelay int64 `json:"activation_delay"`

	// ProposedAt is the block height when proposal was submitted
	ProposedAt int64 `json:"proposed_at"`

	// ProposerAddress is the address that submitted the proposal
	ProposerAddress string `json:"proposer_address"`

	// Status is the current status of the proposal
	Status ModelProposalStatus `json:"status"`

	// GovernanceID is the governance proposal ID (set when submitted to gov)
	GovernanceID uint64 `json:"governance_id,omitempty"`

	// ActivationHeight is the block height when activation should occur
	ActivationHeight int64 `json:"activation_height,omitempty"`
}

// ModelProposalStatus represents the status of a model update proposal
type ModelProposalStatus string

const (
	// ModelProposalStatusPending indicates the proposal is pending
	ModelProposalStatusPending ModelProposalStatus = "pending"

	// ModelProposalStatusApproved indicates the proposal was approved
	ModelProposalStatusApproved ModelProposalStatus = "approved"

	// ModelProposalStatusRejected indicates the proposal was rejected
	ModelProposalStatusRejected ModelProposalStatus = "rejected"

	// ModelProposalStatusActivated indicates the model has been activated
	ModelProposalStatusActivated ModelProposalStatus = "activated"

	// ModelProposalStatusExpired indicates the proposal has expired
	ModelProposalStatusExpired ModelProposalStatus = "expired"
)

// Validate validates the ModelUpdateProposal
func (p ModelUpdateProposal) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if len(p.Title) > 256 {
		return fmt.Errorf("title too long (max 256 characters)")
	}
	if p.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if len(p.Description) > 10000 {
		return fmt.Errorf("description too long (max 10000 characters)")
	}
	if !IsValidModelType(p.ModelType) {
		return fmt.Errorf("invalid model type: %s", p.ModelType)
	}
	if p.NewModelID == "" {
		return fmt.Errorf("new_model_id cannot be empty")
	}
	if p.NewModelHash == "" {
		return fmt.Errorf("new_model_hash cannot be empty")
	}
	if len(p.NewModelHash) != 64 {
		return fmt.Errorf("new_model_hash must be 64 hex characters")
	}
	if p.ActivationDelay < 0 {
		return fmt.Errorf("activation_delay cannot be negative")
	}
	if p.ProposerAddress == "" {
		return fmt.Errorf("proposer_address cannot be empty")
	}
	return nil
}

// ============================================================================
// ModelVersionHistory - Version Change History
// ============================================================================

// ModelVersionHistory tracks model version changes
type ModelVersionHistory struct {
	// HistoryID is the unique identifier for this history entry
	HistoryID string `json:"history_id"`

	// ModelType is the type of model that was changed
	ModelType string `json:"model_type"`

	// OldModelID is the previous model ID (empty if first model)
	OldModelID string `json:"old_model_id"`

	// NewModelID is the new model ID
	NewModelID string `json:"new_model_id"`

	// OldModelHash is the previous model hash
	OldModelHash string `json:"old_model_hash,omitempty"`

	// NewModelHash is the new model hash
	NewModelHash string `json:"new_model_hash"`

	// ChangedAt is the block height when the change occurred
	ChangedAt int64 `json:"changed_at"`

	// GovernanceID is the governance proposal ID that approved this change
	GovernanceID uint64 `json:"governance_id"`

	// ProposerAddress is the address that proposed this change
	ProposerAddress string `json:"proposer_address"`

	// Reason is the reason for the change
	Reason string `json:"reason,omitempty"`
}

// Validate validates the ModelVersionHistory
func (h ModelVersionHistory) Validate() error {
	if h.HistoryID == "" {
		return fmt.Errorf("history_id cannot be empty")
	}
	if !IsValidModelType(h.ModelType) {
		return fmt.Errorf("invalid model type: %s", h.ModelType)
	}
	if h.NewModelID == "" {
		return fmt.Errorf("new_model_id cannot be empty")
	}
	if h.ChangedAt <= 0 {
		return fmt.Errorf("changed_at must be positive")
	}
	return nil
}

// GenerateHistoryID generates a deterministic history ID
func GenerateHistoryID(modelType string, changedAt int64) string {
	data := fmt.Sprintf("%s:%d", modelType, changedAt)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:16])
}

// ============================================================================
// ValidatorModelReport - Validator's Reported Model Versions
// ============================================================================

// ValidatorModelReport represents a validator's reported model versions
type ValidatorModelReport struct {
	// ValidatorAddress is the validator's operator address
	ValidatorAddress string `json:"validator_address"`

	// ModelVersions maps model type to SHA256 hash
	ModelVersions map[string]string `json:"model_versions"`

	// ReportedAt is the block height when this report was submitted
	ReportedAt int64 `json:"reported_at"`

	// LastVerified is the block height when versions were last verified
	LastVerified int64 `json:"last_verified"`

	// IsSynced indicates if all model versions match consensus
	IsSynced bool `json:"is_synced"`

	// MismatchedModels lists model types with version mismatches
	MismatchedModels []string `json:"mismatched_models,omitempty"`
}

// Validate validates the ValidatorModelReport
func (r ValidatorModelReport) Validate() error {
	if r.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}
	if r.ModelVersions == nil {
		return fmt.Errorf("model_versions cannot be nil")
	}
	for modelType, hash := range r.ModelVersions {
		if !IsValidModelType(modelType) {
			return fmt.Errorf("invalid model type: %s", modelType)
		}
		if len(hash) != 64 {
			return fmt.Errorf("invalid hash length for %s", modelType)
		}
	}
	return nil
}

// ============================================================================
// ModelParams - Module Parameters for Model Management
// ============================================================================

// ModelParams contains parameters for model management
type ModelParams struct {
	// RequiredModelTypes lists model types that must be registered
	RequiredModelTypes []string `json:"required_model_types"`

	// ActivationDelayBlocks is the default delay for model activation
	ActivationDelayBlocks int64 `json:"activation_delay_blocks"`

	// MaxModelAgeDays is the maximum age of a model before requiring update
	MaxModelAgeDays int32 `json:"max_model_age_days"`

	// AllowedRegistrars lists addresses allowed to register models
	AllowedRegistrars []string `json:"allowed_registrars"`

	// ValidatorSyncGracePeriod is blocks allowed for validators to sync
	ValidatorSyncGracePeriod int64 `json:"validator_sync_grace_period"`

	// ModelUpdateQuorum is the minimum voting power for model updates
	ModelUpdateQuorum uint32 `json:"model_update_quorum"`

	// EnableGovernanceUpdates enables governance-controlled updates
	EnableGovernanceUpdates bool `json:"enable_governance_updates"`
}

// Validate validates the ModelParams
func (p ModelParams) Validate() error {
	for _, mt := range p.RequiredModelTypes {
		if !IsValidModelType(mt) {
			return fmt.Errorf("invalid required model type: %s", mt)
		}
	}
	if p.ActivationDelayBlocks < 0 {
		return fmt.Errorf("activation_delay_blocks cannot be negative")
	}
	if p.MaxModelAgeDays < 0 {
		return fmt.Errorf("max_model_age_days cannot be negative")
	}
	if p.ValidatorSyncGracePeriod < 0 {
		return fmt.Errorf("validator_sync_grace_period cannot be negative")
	}
	if p.ModelUpdateQuorum > 100 {
		return fmt.Errorf("model_update_quorum cannot exceed 100")
	}
	return nil
}

// DefaultModelParams returns the default model parameters
func DefaultModelParams() ModelParams {
	return ModelParams{
		RequiredModelTypes: []string{
			string(ModelTypeTrustScore),
			string(ModelTypeFaceVerification),
			string(ModelTypeLiveness),
			string(ModelTypeGANDetection),
			string(ModelTypeOCR),
		},
		ActivationDelayBlocks:    1000, // ~1.5 hours at 5s blocks
		MaxModelAgeDays:          365,  // 1 year max age
		AllowedRegistrars:        []string{},
		ValidatorSyncGracePeriod: 500, // ~42 minutes at 5s blocks
		ModelUpdateQuorum:        67,  // 67% quorum
		EnableGovernanceUpdates:  true,
	}
}

// ============================================================================
// Query Types
// ============================================================================

// QueryModelVersionRequest is the request for querying a model version
type QueryModelVersionRequest struct {
	ModelType string `json:"model_type"`
}

// QueryModelVersionResponse is the response for querying a model version
type QueryModelVersionResponse struct {
	ModelInfo *MLModelInfo `json:"model_info"`
}

// QueryActiveModelsRequest is the request for querying all active models
type QueryActiveModelsRequest struct{}

// QueryActiveModelsResponse is the response for querying all active models
type QueryActiveModelsResponse struct {
	State  ModelVersionState `json:"state"`
	Models []*MLModelInfo    `json:"models"`
}

// QueryModelHistoryRequest is the request for querying model history
type QueryModelHistoryRequest struct {
	ModelType  string  `json:"model_type"`
	Pagination *uint32 `json:"pagination,omitempty"`
}

// QueryModelHistoryResponse is the response for querying model history
type QueryModelHistoryResponse struct {
	History []*ModelVersionHistory `json:"history"`
}

// QueryValidatorModelSyncRequest is the request for querying validator model sync
type QueryValidatorModelSyncRequest struct {
	ValidatorAddress string `json:"validator_address"`
}

// QueryValidatorModelSyncResponse is the response for querying validator model sync
type QueryValidatorModelSyncResponse struct {
	Report   *ValidatorModelReport `json:"report"`
	IsSynced bool                  `json:"is_synced"`
}

// ============================================================================
// Genesis Types
// ============================================================================

// ModelGenesisState represents the genesis state for model versioning
type ModelGenesisState struct {
	// Models contains all registered models
	Models []MLModelInfo `json:"models"`

	// State contains the current active model versions
	State ModelVersionState `json:"state"`

	// History contains model version change history
	History []ModelVersionHistory `json:"history"`

	// Params contains model management parameters
	Params ModelParams `json:"params"`

	// PendingProposals contains pending model update proposals
	PendingProposals []ModelUpdateProposal `json:"pending_proposals"`

	// ValidatorReports contains validator model reports
	ValidatorReports []ValidatorModelReport `json:"validator_reports"`
}

// DefaultModelGenesisState returns the default model genesis state
func DefaultModelGenesisState() ModelGenesisState {
	return ModelGenesisState{
		Models:           []MLModelInfo{},
		State:            DefaultModelVersionState(),
		History:          []ModelVersionHistory{},
		Params:           DefaultModelParams(),
		PendingProposals: []ModelUpdateProposal{},
		ValidatorReports: []ValidatorModelReport{},
	}
}

// Validate validates the model genesis state
func (gs ModelGenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate models
	modelIDs := make(map[string]bool)
	for _, model := range gs.Models {
		if err := model.Validate(); err != nil {
			return fmt.Errorf("invalid model %s: %w", model.ModelID, err)
		}
		if modelIDs[model.ModelID] {
			return fmt.Errorf("duplicate model ID: %s", model.ModelID)
		}
		modelIDs[model.ModelID] = true
	}

	// Validate state references existing models
	if gs.State.TrustScoreModel != "" && !modelIDs[gs.State.TrustScoreModel] {
		return fmt.Errorf("state references unknown trust_score model: %s", gs.State.TrustScoreModel)
	}
	if gs.State.FaceVerificationModel != "" && !modelIDs[gs.State.FaceVerificationModel] {
		return fmt.Errorf("state references unknown face_verification model: %s", gs.State.FaceVerificationModel)
	}
	if gs.State.LivenessModel != "" && !modelIDs[gs.State.LivenessModel] {
		return fmt.Errorf("state references unknown liveness model: %s", gs.State.LivenessModel)
	}
	if gs.State.GANDetectionModel != "" && !modelIDs[gs.State.GANDetectionModel] {
		return fmt.Errorf("state references unknown gan_detection model: %s", gs.State.GANDetectionModel)
	}
	if gs.State.OCRModel != "" && !modelIDs[gs.State.OCRModel] {
		return fmt.Errorf("state references unknown ocr model: %s", gs.State.OCRModel)
	}

	// Validate history
	for _, h := range gs.History {
		if err := h.Validate(); err != nil {
			return fmt.Errorf("invalid history entry: %w", err)
		}
	}

	// Validate pending proposals
	for _, p := range gs.PendingProposals {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("invalid pending proposal: %w", err)
		}
	}

	return nil
}

// ============================================================================
// Time Helpers
// ============================================================================

// BlocksToTime estimates the time for a given number of blocks
// Assumes 5 second block time (Cosmos SDK default)
func BlocksToTime(blocks int64) time.Duration {
	return time.Duration(blocks) * 5 * time.Second
}

// TimeToBlocks estimates the number of blocks for a given duration
// Assumes 5 second block time (Cosmos SDK default)
func TimeToBlocks(d time.Duration) int64 {
	return int64(d / (5 * time.Second))
}
