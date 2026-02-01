package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// HealthStatus represents the overall health state of an enclave
type HealthStatus int32

const (
	HealthStatusUnknown   HealthStatus = 0
	HealthStatusHealthy   HealthStatus = 1
	HealthStatusDegraded  HealthStatus = 2
	HealthStatusUnhealthy HealthStatus = 3
)

// String returns the string representation of HealthStatus
func (s HealthStatus) String() string {
	switch s {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// MarshalJSON implements json.Marshaler
func (s HealthStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (s *HealthStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "healthy":
		*s = HealthStatusHealthy
	case "degraded":
		*s = HealthStatusDegraded
	case "unhealthy":
		*s = HealthStatusUnhealthy
	case "unknown":
		*s = HealthStatusUnknown
	default:
		return fmt.Errorf("unknown health status: %s", str)
	}
	return nil
}

// EnclaveHealthStatus tracks the health metrics of a validator's enclave
type EnclaveHealthStatus struct {
	// ValidatorAddress is the validator this health status belongs to
	ValidatorAddress string `json:"validator_address"`

	// LastHeartbeat is the timestamp of the last successful heartbeat
	LastHeartbeat time.Time `json:"last_heartbeat"`

	// LastAttestation is the timestamp of the last successful attestation
	LastAttestation time.Time `json:"last_attestation"`

	// AttestationFailures is the consecutive attestation failure count
	AttestationFailures uint32 `json:"attestation_failures"`

	// SignatureFailures is the consecutive signature verification failure count
	SignatureFailures uint32 `json:"signature_failures"`

	// Status is the current health status of the enclave
	Status HealthStatus `json:"status"`

	// LastStatusChange is when the status last changed
	LastStatusChange time.Time `json:"last_status_change"`

	// TotalHeartbeats is the total number of heartbeats received
	TotalHeartbeats uint64 `json:"total_heartbeats"`

	// MissedHeartbeats is the consecutive missed heartbeat count
	MissedHeartbeats uint32 `json:"missed_heartbeats"`
}

// NewEnclaveHealthStatus creates a new EnclaveHealthStatus with default values
func NewEnclaveHealthStatus(validatorAddress string) EnclaveHealthStatus {
	now := time.Now()
	return EnclaveHealthStatus{
		ValidatorAddress:    validatorAddress,
		LastHeartbeat:       now,
		LastAttestation:     now,
		AttestationFailures: 0,
		SignatureFailures:   0,
		Status:              HealthStatusHealthy,
		LastStatusChange:    now,
		TotalHeartbeats:     0,
		MissedHeartbeats:    0,
	}
}

// Validate performs basic validation of the EnclaveHealthStatus
func (h *EnclaveHealthStatus) Validate() error {
	if h.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	if h.Status < HealthStatusUnknown || h.Status > HealthStatusUnhealthy {
		return fmt.Errorf("invalid health status: %d", h.Status)
	}
	return nil
}

// IsHealthy returns true if the enclave is in healthy status
func (h *EnclaveHealthStatus) IsHealthy() bool {
	return h.Status == HealthStatusHealthy
}

// IsDegraded returns true if the enclave is in degraded status
func (h *EnclaveHealthStatus) IsDegraded() bool {
	return h.Status == HealthStatusDegraded
}

// IsUnhealthy returns true if the enclave is in unhealthy status
func (h *EnclaveHealthStatus) IsUnhealthy() bool {
	return h.Status == HealthStatusUnhealthy
}

// UpdateStatus updates the health status and tracks when it changed
func (h *EnclaveHealthStatus) UpdateStatus(newStatus HealthStatus) {
	if h.Status != newStatus {
		h.Status = newStatus
		h.LastStatusChange = time.Now()
	}
}

// RecordHeartbeat records a successful heartbeat
func (h *EnclaveHealthStatus) RecordHeartbeat(timestamp time.Time) {
	h.LastHeartbeat = timestamp
	h.TotalHeartbeats++
	h.MissedHeartbeats = 0
}

// RecordMissedHeartbeat increments the missed heartbeat counter
func (h *EnclaveHealthStatus) RecordMissedHeartbeat() {
	h.MissedHeartbeats++
}

// RecordAttestation records a successful attestation
func (h *EnclaveHealthStatus) RecordAttestation(timestamp time.Time) {
	h.LastAttestation = timestamp
	h.AttestationFailures = 0
}

// RecordAttestationFailure increments the attestation failure counter
func (h *EnclaveHealthStatus) RecordAttestationFailure() {
	h.AttestationFailures++
}

// RecordSignatureFailure increments the signature failure counter
func (h *EnclaveHealthStatus) RecordSignatureFailure() {
	h.SignatureFailures++
}

// ResetSignatureFailures resets the signature failure counter
func (h *EnclaveHealthStatus) ResetSignatureFailures() {
	h.SignatureFailures = 0
}

// HealthCheckParams contains thresholds for health status determination
type HealthCheckParams struct {
	// MaxMissedHeartbeats is the threshold for degraded status
	MaxMissedHeartbeats uint32 `json:"max_missed_heartbeats"`

	// MaxAttestationFailures is the threshold for unhealthy status
	MaxAttestationFailures uint32 `json:"max_attestation_failures"`

	// MaxSignatureFailures is the threshold for unhealthy status
	MaxSignatureFailures uint32 `json:"max_signature_failures"`

	// HeartbeatTimeoutBlocks is blocks before marking heartbeat as missed
	HeartbeatTimeoutBlocks int64 `json:"heartbeat_timeout_blocks"`

	// AttestationTimeoutBlocks is blocks before requiring fresh attestation
	AttestationTimeoutBlocks int64 `json:"attestation_timeout_blocks"`

	// DegradedThreshold is the threshold for entering degraded state
	DegradedThreshold uint32 `json:"degraded_threshold"`

	// UnhealthyThreshold is the threshold for entering unhealthy state
	UnhealthyThreshold uint32 `json:"unhealthy_threshold"`
}

// DefaultHealthCheckParams returns default health check parameters
func DefaultHealthCheckParams() HealthCheckParams {
	return HealthCheckParams{
		MaxMissedHeartbeats:      3,   // Allow 3 missed heartbeats before degraded
		MaxAttestationFailures:   5,   // Allow 5 attestation failures before unhealthy
		MaxSignatureFailures:     10,  // Allow 10 signature failures before unhealthy
		HeartbeatTimeoutBlocks:   100, // 100 blocks (~10 minutes at 6s/block)
		AttestationTimeoutBlocks: 1000, // 1000 blocks (~100 minutes)
		DegradedThreshold:        3,   // General threshold for degraded state
		UnhealthyThreshold:       10,  // General threshold for unhealthy state
	}
}

// Validate performs basic validation of HealthCheckParams
func (p *HealthCheckParams) Validate() error {
	if p.MaxMissedHeartbeats == 0 {
		return fmt.Errorf("max missed heartbeats must be greater than 0")
	}
	if p.MaxAttestationFailures == 0 {
		return fmt.Errorf("max attestation failures must be greater than 0")
	}
	if p.MaxSignatureFailures == 0 {
		return fmt.Errorf("max signature failures must be greater than 0")
	}
	if p.HeartbeatTimeoutBlocks <= 0 {
		return fmt.Errorf("heartbeat timeout blocks must be positive")
	}
	if p.AttestationTimeoutBlocks <= 0 {
		return fmt.Errorf("attestation timeout blocks must be positive")
	}
	if p.DegradedThreshold >= p.UnhealthyThreshold {
		return fmt.Errorf("degraded threshold must be less than unhealthy threshold")
	}
	return nil
}

// EvaluateHealth determines the health status based on current metrics
func (p *HealthCheckParams) EvaluateHealth(health *EnclaveHealthStatus, currentTime time.Time, currentHeight int64) HealthStatus {
	// Check for unhealthy conditions first
	if health.AttestationFailures >= p.MaxAttestationFailures {
		return HealthStatusUnhealthy
	}
	if health.SignatureFailures >= p.MaxSignatureFailures {
		return HealthStatusUnhealthy
	}
	if health.MissedHeartbeats >= p.UnhealthyThreshold {
		return HealthStatusUnhealthy
	}

	// Check for degraded conditions
	if health.MissedHeartbeats >= p.MaxMissedHeartbeats {
		return HealthStatusDegraded
	}
	if health.AttestationFailures >= p.DegradedThreshold {
		return HealthStatusDegraded
	}
	if health.SignatureFailures >= p.DegradedThreshold {
		return HealthStatusDegraded
	}

	// Otherwise, healthy
	return HealthStatusHealthy
}

// MsgEnclaveHeartbeat represents a heartbeat message from an enclave
type MsgEnclaveHeartbeat struct {
	// ValidatorAddress is the validator sending the heartbeat
	ValidatorAddress string `json:"validator_address"`

	// Timestamp is when the heartbeat was generated
	Timestamp time.Time `json:"timestamp"`

	// AttestationProof is an optional fresh attestation quote
	AttestationProof []byte `json:"attestation_proof,omitempty"`

	// Signature is the enclave's signature over the heartbeat data
	Signature []byte `json:"signature"`

	// Nonce prevents replay attacks
	Nonce uint64 `json:"nonce"`
}

// Route implements sdk.Msg
func (msg MsgEnclaveHeartbeat) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgEnclaveHeartbeat) Type() string {
	return TypeMsgEnclaveHeartbeat
}

// ValidateBasic performs basic validation of the message
func (msg MsgEnclaveHeartbeat) ValidateBasic() error {
	if msg.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	if msg.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	if len(msg.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}
	// Nonce can be any value, including 0
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg MsgEnclaveHeartbeat) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// GetSigners returns the addresses that must sign the transaction
func (msg MsgEnclaveHeartbeat) GetSigners() []string {
	return []string{msg.ValidatorAddress}
}

// MsgEnclaveHeartbeatResponse is the response for MsgEnclaveHeartbeat
type MsgEnclaveHeartbeatResponse struct {
	// Success indicates if the heartbeat was processed successfully
	Success bool `json:"success"`

	// CurrentStatus is the health status after processing the heartbeat
	CurrentStatus HealthStatus `json:"current_status"`

	// Message provides additional information
	Message string `json:"message,omitempty"`
}

// Proto message interface implementations
func (MsgEnclaveHeartbeat) XXX_MessageName() string {
	return "virtengine.enclave.v1.MsgEnclaveHeartbeat"
}

func (MsgEnclaveHeartbeatResponse) XXX_MessageName() string {
	return "virtengine.enclave.v1.MsgEnclaveHeartbeatResponse"
}

func (EnclaveHealthStatus) XXX_MessageName() string {
	return "virtengine.enclave.v1.EnclaveHealthStatus"
}

func (HealthCheckParams) XXX_MessageName() string {
	return "virtengine.enclave.v1.HealthCheckParams"
}

// Proto.Message interface implementation
func (m *MsgEnclaveHeartbeat) Reset()         { *m = MsgEnclaveHeartbeat{} }
func (m *MsgEnclaveHeartbeat) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgEnclaveHeartbeat) ProtoMessage()  {}

func (m *MsgEnclaveHeartbeatResponse) Reset()         { *m = MsgEnclaveHeartbeatResponse{} }
func (m *MsgEnclaveHeartbeatResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgEnclaveHeartbeatResponse) ProtoMessage()  {}

func (m *EnclaveHealthStatus) Reset()         { *m = EnclaveHealthStatus{} }
func (m *EnclaveHealthStatus) String() string { return fmt.Sprintf("%+v", *m) }
func (m *EnclaveHealthStatus) ProtoMessage()  {}

func (m *HealthCheckParams) Reset()         { *m = HealthCheckParams{} }
func (m *HealthCheckParams) String() string { return fmt.Sprintf("%+v", *m) }
func (m *HealthCheckParams) ProtoMessage()  {}
