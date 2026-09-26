// Package marketplace provides types for the marketplace on-chain module.
//
// This file implements the Waldur command/attestation protocol defined by
// _docs/adr/ADR-010-unified-market-resolution-and-waldur-supply.md. Commands are
// durable, signed instructions emitted by the chain for off-chain Waldur
// adapters; attestations are the signed inputs that enter consensus.
package marketplace

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// WaldurCommandKeyPrefix is the store prefix for durable Waldur commands.
var WaldurCommandKeyPrefix = []byte{0x12}

// WaldurSourceKeyPrefix is the store prefix for registered Waldur instances.
var WaldurSourceKeyPrefix = []byte{0x13}

// WaldurCommandKind identifies the kind of command emitted for Waldur.
type WaldurCommandKind string

const (
	// WaldurCommandCreateOffering instructs the adapter to create an offering.
	WaldurCommandCreateOffering WaldurCommandKind = "create_offering"

	// WaldurCommandUpdateOffering instructs the adapter to update an offering.
	WaldurCommandUpdateOffering WaldurCommandKind = "update_offering"

	// WaldurCommandCreateOrder instructs the adapter to create an order.
	WaldurCommandCreateOrder WaldurCommandKind = "create_order"

	// WaldurCommandApproveOrder instructs the adapter to approve an order.
	WaldurCommandApproveOrder WaldurCommandKind = "approve_order"

	// WaldurCommandSetBackendID instructs the adapter to set the backend id.
	WaldurCommandSetBackendID WaldurCommandKind = "set_backend_id"

	// WaldurCommandLifecycle instructs the adapter to run a lifecycle action.
	WaldurCommandLifecycle WaldurCommandKind = "lifecycle"

	// WaldurCommandUsage instructs the adapter to submit a usage report.
	WaldurCommandUsage WaldurCommandKind = "usage"
)

// IsValid returns true if the command kind is valid.
func (k WaldurCommandKind) IsValid() bool {
	switch k {
	case WaldurCommandCreateOffering, WaldurCommandUpdateOffering, WaldurCommandCreateOrder,
		WaldurCommandApproveOrder, WaldurCommandSetBackendID, WaldurCommandLifecycle, WaldurCommandUsage:
		return true
	default:
		return false
	}
}

// WaldurCommand is a durable instruction for an off-chain Waldur adapter.
type WaldurCommand struct {
	// ID is the unique command identifier (deterministic).
	ID string `json:"id"`

	// Kind is the command kind.
	Kind WaldurCommandKind `json:"kind"`

	// InstanceID identifies the target Waldur instance.
	InstanceID string `json:"instance_id"`

	// WaldurOfferingUUID is the target Waldur offering UUID.
	WaldurOfferingUUID string `json:"waldur_offering_uuid,omitempty"`

	// WaldurOrderUUID is the target Waldur order UUID.
	WaldurOrderUUID string `json:"waldur_order_uuid,omitempty"`

	// WaldurResourceUUID is the target Waldur resource UUID.
	WaldurResourceUUID string `json:"waldur_resource_uuid,omitempty"`

	// ChainEntityType is the on-chain entity type being synced.
	ChainEntityType WaldurSyncType `json:"chain_entity_type,omitempty"`

	// ChainEntityID is the on-chain entity ID.
	ChainEntityID string `json:"chain_entity_id,omitempty"`

	// BackendID is the canonical VirtEngine identifier for reconciliation.
	BackendID string `json:"backend_id,omitempty"`

	// Payload carries additional command-specific attributes.
	Payload map[string]string `json:"payload,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is the expiry timestamp (zero means no expiry).
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// Attempts counts dispatch attempts by the adapter.
	Attempts uint32 `json:"attempts"`

	// Acked indicates the adapter has acknowledged completion.
	Acked bool `json:"acked"`

	// AckedAt is when the command was acknowledged.
	AckedAt *time.Time `json:"acked_at,omitempty"`

	// LastError is the most recent adapter error.
	LastError string `json:"last_error,omitempty"`
}

// NewWaldurCommandAt creates a validated command with a deterministic ID.
func NewWaldurCommandAt(kind WaldurCommandKind, instanceID, entityID string, now time.Time) (*WaldurCommand, error) {
	if !kind.IsValid() {
		return nil, fmt.Errorf("invalid waldur command kind: %s", kind)
	}
	if strings.TrimSpace(instanceID) == "" {
		return nil, fmt.Errorf("waldur instance id is required")
	}
	createdAt := now.UTC()
	return &WaldurCommand{
		ID:         WaldurCommandID(string(kind), instanceID, entityID),
		Kind:       kind,
		InstanceID: instanceID,
		CreatedAt:  createdAt,
		Payload:    map[string]string{},
	}, nil
}

// WaldurCommandID derives a deterministic command identifier.
func WaldurCommandID(kind, instanceID, entityID string) string {
	h := sha256.New()
	h.Write([]byte(kind))
	h.Write([]byte("|"))
	h.Write([]byte(instanceID))
	h.Write([]byte("|"))
	h.Write([]byte(entityID))
	return hex.EncodeToString(h.Sum(nil))
}

// Validate validates the command.
func (c *WaldurCommand) Validate() error {
	if c == nil {
		return fmt.Errorf("waldur command is required")
	}
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("waldur command id is required")
	}
	if !c.Kind.IsValid() {
		return fmt.Errorf("invalid waldur command kind: %s", c.Kind)
	}
	if strings.TrimSpace(c.InstanceID) == "" {
		return fmt.Errorf("waldur instance id is required")
	}
	return nil
}

// MarkAcked records acknowledgement at the given time.
func (c *WaldurCommand) MarkAcked(now time.Time) {
	ackedAt := now.UTC()
	c.Acked = true
	c.AckedAt = &ackedAt
	c.LastError = ""
}

// MarkFailed records a failed dispatch attempt.
func (c *WaldurCommand) MarkFailed(err string) {
	c.Attempts++
	c.LastError = err
}

// IsExpired returns true if the command has expired.
func (c *WaldurCommand) IsExpired(now time.Time) bool {
	return !c.ExpiresAt.IsZero() && now.After(c.ExpiresAt)
}

// WaldurCommandKey returns the store key for a command.
func WaldurCommandKey(id string) []byte {
	return append(append([]byte{}, WaldurCommandKeyPrefix...), []byte(id)...)
}

// WaldurSource is a registered Waldur instance whose signatures are trusted for
// deterministic offering ingestion.
type WaldurSource struct {
	// InstanceID is the unique instance identifier.
	InstanceID string `json:"instance_id"`

	// BaseURL is the Waldur API base URL (informational).
	BaseURL string `json:"base_url,omitempty"`

	// PublicKey is the hex-encoded ed25519 public key used to sign snapshots.
	PublicKey string `json:"public_key"`

	// RelayerQuorum is the number of relayers required for shared listings.
	RelayerQuorum uint32 `json:"relayer_quorum"`

	// RegisteredAt is the registration timestamp.
	RegisteredAt time.Time `json:"registered_at"`

	// Active indicates whether the source may submit ingestions.
	Active bool `json:"active"`
}

// Validate validates the Waldur source.
func (s *WaldurSource) Validate() error {
	if s == nil {
		return fmt.Errorf("waldur source is required")
	}
	if strings.TrimSpace(s.InstanceID) == "" {
		return fmt.Errorf("waldur source instance_id is required")
	}
	key, err := hex.DecodeString(strings.TrimSpace(s.PublicKey))
	if err != nil {
		return fmt.Errorf("waldur source public_key must be hex: %w", err)
	}
	if len(key) != ed25519.PublicKeySize {
		return fmt.Errorf("waldur source public_key must be %d bytes, got %d", ed25519.PublicKeySize, len(key))
	}
	return nil
}

// WaldurSourceKey returns the store key for a source.
func WaldurSourceKey(instanceID string) []byte {
	return append(append([]byte{}, WaldurSourceKeyPrefix...), []byte(instanceID)...)
}

// WaldurOfferingAttestation is the signed snapshot header for an ingested
// Waldur offering.
type WaldurOfferingAttestation struct {
	// InstanceID identifies the signing Waldur instance.
	InstanceID string `json:"instance_id"`

	// OfferingUUID is the Waldur offering UUID.
	OfferingUUID string `json:"offering_uuid"`

	// SnapshotHash is the canonical checksum of the offering snapshot.
	SnapshotHash string `json:"snapshot_hash"`

	// SnapshotHeight is the monotonic Waldur revision.
	SnapshotHeight uint64 `json:"snapshot_height"`

	// PublicKey is the hex-encoded signer public key.
	PublicKey string `json:"public_key"`

	// Signature is the hex-encoded ed25519 signature over the attestation digest.
	Signature string `json:"signature"`
}

// Digest returns the canonical digest signed by a Waldur instance.
func (a *WaldurOfferingAttestation) Digest() []byte {
	h := sha256.New()
	h.Write([]byte(a.InstanceID))
	h.Write([]byte("|"))
	h.Write([]byte(a.OfferingUUID))
	h.Write([]byte("|"))
	h.Write([]byte(a.SnapshotHash))
	h.Write([]byte("|"))
	fmt.Fprintf(h, "%d", a.SnapshotHeight)
	return h.Sum(nil)
}

// Verify checks the attestation signature against the provided expected public
// key (hex). It returns an error when the signature is missing or invalid.
func (a *WaldurOfferingAttestation) Verify(expectedPublicKey string) error {
	if a == nil {
		return fmt.Errorf("attestation is required")
	}
	if strings.TrimSpace(a.InstanceID) == "" || strings.TrimSpace(a.OfferingUUID) == "" {
		return fmt.Errorf("attestation instance_id and offering_uuid are required")
	}
	if strings.TrimSpace(a.Signature) == "" {
		return fmt.Errorf("attestation signature is required")
	}

	pubHex := strings.TrimSpace(expectedPublicKey)
	if pubHex == "" {
		pubHex = strings.TrimSpace(a.PublicKey)
	}
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}
	if len(pubBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key length: %d", len(pubBytes))
	}

	sig, err := hex.DecodeString(strings.TrimSpace(a.Signature))
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature length: %d", len(sig))
	}

	if !ed25519.Verify(ed25519.PublicKey(pubBytes), a.Digest(), sig) {
		return fmt.Errorf("waldur attestation signature verification failed")
	}
	return nil
}

// SignWaldurOfferingAttestation signs an attestation with the given private key.
// It is intended for adapters and tests.
func SignWaldurOfferingAttestation(instanceID, offeringUUID, snapshotHash string, snapshotHeight uint64, priv ed25519.PrivateKey) *WaldurOfferingAttestation {
	att := &WaldurOfferingAttestation{
		InstanceID:     instanceID,
		OfferingUUID:   offeringUUID,
		SnapshotHash:   snapshotHash,
		SnapshotHeight: snapshotHeight,
		PublicKey:      hex.EncodeToString(priv.Public().(ed25519.PublicKey)),
		Signature:      hex.EncodeToString(ed25519.Sign(priv, (&WaldurOfferingAttestation{InstanceID: instanceID, OfferingUUID: offeringUUID, SnapshotHash: snapshotHash, SnapshotHeight: snapshotHeight}).Digest())),
	}
	return att
}
