package provider_daemon

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

// SnapshotIngestSubmitter ingests signed Waldur snapshots through the
// canonical MsgIngestWaldurOffering path.
type SnapshotIngestSubmitter interface {
	// IngestSnapshot signs and submits a snapshot for the import. It returns
	// the on-chain offering ID when known (empty when the chain assigns it
	// asynchronously) and the ingestion action performed.
	IngestSnapshot(ctx context.Context, imp *marketplace.WaldurOfferingImport) (offeringID string, action string, err error)
}

// BuildWaldurSnapshot converts a canonical import into a wire snapshot.
//
// Adapters MUST populate Created/Modified from Waldur's actual timestamps:
// the checksum that both sides sign covers those fields, so any skew between
// the adapter's view and the chain's reconstruction breaks verification.
func BuildWaldurSnapshot(imp *marketplace.WaldurOfferingImport) *marketplacev1.WaldurOfferingSnapshot {
	if imp == nil {
		return nil
	}
	snapshot := &marketplacev1.WaldurOfferingSnapshot{
		Uuid:           imp.UUID,
		InstanceId:     imp.InstanceID,
		Name:           imp.Name,
		Description:    imp.Description,
		Type:           imp.Type,
		State:          imp.State,
		CategoryUuid:   imp.CategoryUUID,
		CustomerUuid:   imp.CustomerUUID,
		Shared:         imp.Shared,
		Billable:       imp.Billable,
		Created:        imp.Created.Unix(),
		Modified:       imp.Modified.Unix(),
		SnapshotHeight: SnapshotHeightForImport(imp),
	}
	if len(imp.Attributes) > 0 {
		snapshot.Attributes = make(map[string]string, len(imp.Attributes))
		for key, value := range imp.Attributes {
			switch v := value.(type) {
			case string:
				snapshot.Attributes[key] = v
			default:
				snapshot.Attributes[key] = fmt.Sprintf("%v", v)
			}
		}
	}
	for _, component := range imp.Components {
		snapshot.Components = append(snapshot.Components, marketplacev1.WaldurPricingComponent{
			Type:         component.Type,
			Name:         component.Name,
			MeasuredUnit: component.MeasuredUnit,
			BillingType:  component.BillingType,
			Price:        component.Price,
		})
	}
	return snapshot
}

// SnapshotHeightForImport returns the monotonic snapshot height: the import's
// modification time, falling back to creation time, then to 1. Heights only
// ever increase for a given offering as long as Waldur timestamps do.
func SnapshotHeightForImport(imp *marketplace.WaldurOfferingImport) uint64 {
	if imp == nil {
		return 1
	}
	if unix := imp.Modified.Unix(); unix > 0 {
		return uint64(unix)
	}
	if unix := imp.Created.Unix(); unix > 0 {
		return uint64(unix)
	}
	return 1
}

// ChainSnapshotSubmitter signs snapshots with the configured Waldur key and
// submits them through the provider mutation pipeline.
type ChainSnapshotSubmitter struct {
	submitter *ProviderMutationSubmitter
	relayer   string
	signer    ed25519.PrivateKey
}

// NewChainSnapshotSubmitter creates a snapshot submitter. The signer must be
// the ed25519 private key whose public key is registered on-chain for the
// Waldur instance (see MsgRegisterWaldurSource).
func NewChainSnapshotSubmitter(submitter *ProviderMutationSubmitter, relayer string, signer ed25519.PrivateKey) (*ChainSnapshotSubmitter, error) {
	if submitter == nil {
		return nil, fmt.Errorf("mutation submitter is required")
	}
	if strings.TrimSpace(relayer) == "" {
		return nil, fmt.Errorf("relayer address is required")
	}
	if len(signer) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid snapshot signing key length: %d", len(signer))
	}
	return &ChainSnapshotSubmitter{submitter: submitter, relayer: relayer, signer: signer}, nil
}

// IngestSnapshot implements SnapshotIngestSubmitter.
func (s *ChainSnapshotSubmitter) IngestSnapshot(ctx context.Context, imp *marketplace.WaldurOfferingImport) (string, string, error) {
	if imp == nil {
		return "", "", fmt.Errorf("offering import is required")
	}
	snapshot := BuildWaldurSnapshot(imp)
	if snapshot == nil {
		return "", "", fmt.Errorf("snapshot build failed")
	}
	attestation := marketplace.SignWaldurOfferingAttestation(
		imp.InstanceID,
		imp.UUID,
		imp.IngestChecksum(),
		snapshot.SnapshotHeight,
		s.signer,
	)
	msg := &marketplacev1.MsgIngestWaldurOffering{
		Relayer:   s.relayer,
		Snapshot:  snapshot,
		Signature: attestation.Signature,
	}
	if _, err := sdk.AccAddressFromBech32(msg.Relayer); err != nil {
		return "", "", fmt.Errorf("invalid relayer address: %w", err)
	}
	if _, err := s.submitter.Submit(ctx, MutationMarketplaceIngestOffering, msg); err != nil {
		return "", "", err
	}
	// The chain assigns provider-scoped sequences on ingest, so the offering
	// ID is resolved by UUID on subsequent reconciliation, not here.
	return "", string(marketplace.IngestActionCreate), nil
}

// LoadSnapshotSignerKey loads a hex-encoded ed25519 seed (32 bytes) or private
// key (64 bytes), either inline or from a file path prefixed with "file:".
func LoadSnapshotSignerKey(value string) (ed25519.PrivateKey, error) {
	raw := strings.TrimSpace(value)
	if after, ok := strings.CutPrefix(raw, "file:"); ok {
		data, err := os.ReadFile(strings.TrimSpace(after))
		if err != nil {
			return nil, fmt.Errorf("read signing key file: %w", err)
		}
		raw = strings.TrimSpace(string(data))
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("signing key must be hex: %w", err)
	}
	switch len(key) {
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(key), nil
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(append([]byte(nil), key...)), nil
	default:
		return nil, fmt.Errorf("signing key must be %d or %d bytes, got %d", ed25519.SeedSize, ed25519.PrivateKeySize, len(key))
	}
}

var (
	_ SnapshotIngestSubmitter = (*ChainSnapshotSubmitter)(nil)
	_ sdk.Msg                 = (*marketplacev1.MsgIngestWaldurOffering)(nil)
)

// SnapshotIngestPublicKey returns the hex-encoded public key for registration.
func SnapshotIngestPublicKey(signer ed25519.PrivateKey) string {
	if len(signer) != ed25519.PrivateKeySize {
		return ""
	}
	pub, ok := signer.Public().(ed25519.PublicKey)
	if !ok {
		return ""
	}
	return hex.EncodeToString(pub)
}

// SnapshotIngestAuditNote renders an operator-facing note for audit logs.
func SnapshotIngestAuditNote(imp *marketplace.WaldurOfferingImport, height uint64) string {
	if imp == nil {
		return "nil import"
	}
	return fmt.Sprintf("instance=%s uuid=%s height=%d checksum=%s at=%s",
		imp.InstanceID, imp.UUID, height, imp.IngestChecksum(), time.Now().UTC().Format(time.RFC3339))
}
