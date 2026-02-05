package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/log"
	"golang.org/x/crypto/curve25519"
	"google.golang.org/grpc"

	"github.com/virtengine/virtengine/pkg/servicedesk"
	"github.com/virtengine/virtengine/pkg/waldur"
	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

// SupportEncryptedPayload aliases the on-chain payload type.
type SupportEncryptedPayload = supporttypes.EncryptedSupportPayload

// Encryption query client aliases.
type EncryptionQueryClient = encryptionv1.QueryClient
type EncryptionKeyByFingerprintRequest = encryptionv1.QueryKeyByFingerprintRequest
type EncryptionKeyByFingerprintResponse = encryptionv1.QueryKeyByFingerprintResponse

// NewEncryptionQueryClient constructs a query client.
func NewEncryptionQueryClient(conn grpc.ClientConnInterface) EncryptionQueryClient {
	return encryptionv1.NewQueryClient(conn)
}

// SupportKeyPair wraps the encryption key pair.
type SupportKeyPair = encryptioncrypto.KeyPair

// NewSupportKeyPair builds a key pair from a private key.
func NewSupportKeyPair(privateKey []byte) (*SupportKeyPair, error) {
	if len(privateKey) != 32 {
		return nil, fmt.Errorf("support private key must be 32 bytes")
	}
	var pk [32]byte
	copy(pk[:], privateKey)
	var pub [32]byte
	curve25519.ScalarBaseMult(&pub, &pk)
	return &SupportKeyPair{
		PrivateKey: pk,
		PublicKey:  pub,
	}, nil
}

// BuildSupportResponsePayload marshals a response payload.
func BuildSupportResponsePayload(message string) ([]byte, error) {
	payload := supporttypes.SupportResponsePayload{Message: message}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(payload)
}

// EncryptSupportPayload encrypts payload for recipients.
func EncryptSupportPayload(plaintext []byte, recipientPublicKeys [][]byte, sender *SupportKeyPair) (*SupportEncryptedPayload, error) {
	if len(recipientPublicKeys) == 0 {
		return nil, fmt.Errorf("recipient keys required")
	}
	envelope, err := encryptioncrypto.CreateMultiRecipientEnvelope(plaintext, recipientPublicKeys, sender)
	if err != nil {
		return nil, err
	}
	payload := &supporttypes.EncryptedSupportPayload{Envelope: envelope}
	payload.EnsureEnvelopeHash()
	return payload, nil
}

// SupportInboundHandler processes inbound updates from external service desks.
type SupportInboundHandler struct {
	providerAddress string
	bridge          *servicedesk.Bridge
	encryptor       SupportResponseEncryptor
	chainWriter     SupportChainWriter
	logger          log.Logger
	keyManager      *KeyManager
}

// HandleInboundUpdate handles inbound updates.
func (h *SupportInboundHandler) HandleInboundUpdate(ctx context.Context, event *servicedesk.SyncEvent) error {
	if event == nil {
		return nil
	}
	if h.chainWriter == nil {
		return nil
	}

	status := extractStatus(event.Payload)
	if status != "" {
		status = normalizeStatus(event, status)
		update := &SupportUpdateRequest{
			TicketID:      event.TicketID,
			Status:        status,
			AssignedAgent: extractAssignedAgent(event.Payload),
			UpdatedBy:     h.providerAddress,
			Metadata:      extractMetadata(event.Payload),
		}
		if err := h.chainWriter.UpdateSupportRequest(ctx, update); err != nil {
			return err
		}
	}

	message := extractMessage(event.Payload)
	if message == "" {
		return nil
	}

	recipients := h.recipientsForTicket(event.TicketID)
	if len(recipients) == 0 {
		return fmt.Errorf("missing recipients for ticket %s", event.TicketID)
	}

	var encrypted *SupportEncryptedPayload
	if h.encryptor != nil {
		var err error
		encrypted, err = h.encryptor.EncryptResponse(ctx, recipients, message)
		if err != nil {
			return err
		}
	}

	response := &SupportAddResponse{
		TicketID: event.TicketID,
		Author:   h.providerAddress,
		IsAgent:  true,
		Message:  message,
		Payload:  encrypted,
	}
	return h.chainWriter.AddSupportResponse(ctx, response)
}

func (h *SupportInboundHandler) recipientsForTicket(ticketID string) []string {
	if h.bridge == nil {
		return nil
	}
	return h.bridge.GetTicketRecipients(ticketID)
}

func extractStatus(payload map[string]interface{}) string {
	if payload == nil {
		return ""
	}
	if status, ok := payload["status"].(string); ok && status != "" {
		return status
	}
	if state, ok := payload["state"].(string); ok && state != "" {
		return state
	}
	if status, ok := payload["to_status"].(string); ok && status != "" {
		return status
	}
	return ""
}

func extractAssignedAgent(payload map[string]interface{}) string {
	if payload == nil {
		return ""
	}
	if assignee, ok := payload["assignee"].(string); ok {
		return assignee
	}
	if assignee, ok := payload["assignee_uuid"].(string); ok {
		return assignee
	}
	if assignee, ok := payload["to_assignee"].(string); ok {
		return assignee
	}
	return ""
}

func extractMessage(payload map[string]interface{}) string {
	if payload == nil {
		return ""
	}
	if msg, ok := payload["message"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := payload["comment_body"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := payload["description"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := payload["comment"].(string); ok && msg != "" {
		return msg
	}
	return ""
}

func extractMetadata(payload map[string]interface{}) map[string]string {
	if payload == nil {
		return nil
	}
	meta := map[string]string{}
	for _, key := range []string{"external_id", "comment_id", "source"} {
		if value, ok := payload[key].(string); ok && value != "" {
			meta[key] = value
		}
	}
	if len(meta) == 0 {
		return nil
	}
	return meta
}

func normalizeStatus(event *servicedesk.SyncEvent, status string) string {
	if event == nil {
		return status
	}
	switch event.Direction {
	case servicedesk.SyncDirectionInbound:
		if event.Type == "status_changed" && strings.TrimSpace(status) != "" {
			return status
		}
	}
	if event.Payload != nil {
		if sd, ok := event.Payload["service_desk"].(string); ok && sd == string(servicedesk.ServiceDeskWaldur) {
			return waldur.MapWaldurStateToVirtEngine(waldur.IssueState(status))
		}
	}
	return status
}

// SupportExternalRefHandler registers external ticket references on-chain.
type SupportExternalRefHandler struct {
	chainWriter     SupportChainWriter
	providerAddress string
	logger          log.Logger
}

// NewSupportExternalRefHandler builds a handler for external refs.
func NewSupportExternalRefHandler(chainWriter SupportChainWriter, providerAddress string, logger log.Logger) *SupportExternalRefHandler {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return &SupportExternalRefHandler{
		chainWriter:     chainWriter,
		providerAddress: providerAddress,
		logger:          logger,
	}
}

// HandleExternalRef registers an external reference.
func (h *SupportExternalRefHandler) HandleExternalRef(ctx context.Context, ticketID string, ref servicedesk.ExternalTicketRef) error {
	if h == nil || h.chainWriter == nil {
		return nil
	}
	msg := &SupportRegisterExternal{
		ResourceID:       ticketID,
		ResourceType:     "support_request",
		ExternalSystem:   ref.Type.String(),
		ExternalTicketID: ref.ExternalID,
		ExternalURL:      ref.ExternalURL,
		CreatedBy:        h.providerAddress,
	}
	if err := h.chainWriter.RegisterExternalTicket(ctx, msg); err != nil {
		h.logger.Error("failed to register external ticket", "error", err, "ticket_id", ticketID)
		return err
	}
	return nil
}

// nolint:unused
var _ = time.Second
