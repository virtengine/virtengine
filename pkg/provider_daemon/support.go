package provider_daemon

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"cosmossdk.io/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/virtengine/virtengine/pkg/servicedesk"
)

// SupportServiceConfig configures support ticket routing and sync.
type SupportServiceConfig struct {
	Enabled         bool
	ProviderAddress string
	ChainID         string
	CometRPC        string
	CometWS         string
	GRPCEndpoint    string
	SubscriberID    string

	ServiceDeskConfig *servicedesk.Config
	EventListener     *servicedesk.EventListenerConfig
	Encryption        SupportEncryptionConfig
}

// SupportEncryptionConfig provides sender key for encrypting responses.
type SupportEncryptionConfig struct {
	SenderPrivateKeyBase64 string
	SenderPrivateKeyPath   string
}

// LoadPrivateKey loads the sender private key bytes.
func (c SupportEncryptionConfig) LoadPrivateKey() ([]byte, error) {
	if c.SenderPrivateKeyBase64 != "" {
		key, err := base64.StdEncoding.DecodeString(c.SenderPrivateKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("decode support private key: %w", err)
		}
		return key, nil
	}
	if c.SenderPrivateKeyPath == "" {
		return nil, nil
	}
	key, err := os.ReadFile(c.SenderPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read support private key: %w", err)
	}
	return key, nil
}

// DefaultSupportServiceConfig returns default support service configuration.
func DefaultSupportServiceConfig() SupportServiceConfig {
	cfg := SupportServiceConfig{
		CometWS: "/websocket",
	}
	cfg.ServiceDeskConfig = servicedesk.DefaultConfig()
	return cfg
}

// SupportService wires chain events to service desk and inbound updates to chain.
type SupportService struct {
	cfg            SupportServiceConfig
	logger         log.Logger
	bridge         *servicedesk.Bridge
	listener       *servicedesk.ChainEventListener
	inboundHandler *SupportInboundHandler
	grpcConn       *grpc.ClientConn
}

// NewSupportService creates a support service instance.
func NewSupportService(cfg SupportServiceConfig, keyManager *KeyManager, logger log.Logger) (*SupportService, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.ServiceDeskConfig == nil {
		cfg.ServiceDeskConfig = servicedesk.DefaultConfig()
	}
	if cfg.ServiceDeskConfig.Decryption == nil {
		cfg.ServiceDeskConfig.Decryption = &servicedesk.DecryptionConfig{}
	}
	if cfg.EventListener == nil {
		cfg.EventListener = servicedesk.DefaultEventListenerConfig()
	}
	if cfg.CometWS == "" {
		cfg.CometWS = "/websocket"
	}
	if cfg.EventListener.CometWS == "" {
		cfg.EventListener.CometWS = cfg.CometWS
	}
	if cfg.SubscriberID != "" {
		cfg.EventListener.SubscriberID = cfg.SubscriberID
	}

	if logger == nil {
		logger = log.NewNopLogger()
	}

	bridge, err := servicedesk.NewBridge(cfg.ServiceDeskConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("create service desk bridge: %w", err)
	}

	var grpcConn *grpc.ClientConn
	if cfg.GRPCEndpoint != "" {
		conn, err := grpc.NewClient(cfg.GRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("connect support grpc: %w", err)
		}
		grpcConn = conn
	}

	recipientResolver := NewGRPCRecipientKeyResolver(grpcConn)
	encryptor, err := NewSupportResponseEncryptor(cfg, recipientResolver)
	if err != nil {
		if grpcConn != nil {
			_ = grpcConn.Close()
		}
		return nil, err
	}

	chainWriter := NewSupportChainLogger(logger.With("component", "support_chain_logger"))
	inboundHandler := NewSupportInboundHandler(SupportInboundHandlerConfig{
		ProviderAddress: cfg.ProviderAddress,
		Bridge:          bridge,
		Encryptor:       encryptor,
		ChainWriter:     chainWriter,
		Logger:          logger.With("component", "support_inbound"),
		KeyManager:      keyManager,
	})
	bridge.SetInboundHandler(inboundHandler)

	externalRefHandler := NewSupportExternalRefHandler(chainWriter, cfg.ProviderAddress, logger.With("component", "support_external_ref"))
	bridge.SetExternalRefHandler(externalRefHandler)

	listener := servicedesk.NewChainEventListener(bridge, cfg.CometRPC, logger, cfg.EventListener)

	return &SupportService{
		cfg:            cfg,
		logger:         logger.With("component", "support_service"),
		bridge:         bridge,
		listener:       listener,
		inboundHandler: inboundHandler,
		grpcConn:       grpcConn,
	}, nil
}

// Start starts the support service.
func (s *SupportService) Start(ctx context.Context) error {
	if s == nil || s.bridge == nil || s.listener == nil {
		return nil
	}

	if err := s.bridge.Start(ctx); err != nil {
		return fmt.Errorf("start support bridge: %w", err)
	}
	if err := s.listener.Start(ctx); err != nil {
		return fmt.Errorf("start support event listener: %w", err)
	}
	s.logger.Info("support service started")
	return nil
}

// Stop stops the support service.
func (s *SupportService) Stop(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.listener != nil {
		_ = s.listener.Stop(ctx)
	}
	if s.bridge != nil {
		_ = s.bridge.Stop(ctx)
	}
	if s.grpcConn != nil {
		_ = s.grpcConn.Close()
	}
	s.logger.Info("support service stopped")
	return nil
}

// SupportInboundHandlerConfig configures inbound update handling.
type SupportInboundHandlerConfig struct {
	ProviderAddress string
	Bridge          *servicedesk.Bridge
	Encryptor       SupportResponseEncryptor
	ChainWriter     SupportChainWriter
	Logger          log.Logger
	KeyManager      *KeyManager
}

// NewSupportInboundHandler constructs a handler for inbound updates.
func NewSupportInboundHandler(cfg SupportInboundHandlerConfig) *SupportInboundHandler {
	return &SupportInboundHandler{
		providerAddress: cfg.ProviderAddress,
		bridge:          cfg.Bridge,
		encryptor:       cfg.Encryptor,
		chainWriter:     cfg.ChainWriter,
		logger:          cfg.Logger,
		keyManager:      cfg.KeyManager,
	}
}

// SupportChainLogger is a placeholder chain writer for support events.
type SupportChainLogger struct {
	logger log.Logger
}

// NewSupportChainLogger returns a logging chain writer.
func NewSupportChainLogger(logger log.Logger) *SupportChainLogger {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return &SupportChainLogger{logger: logger}
}

// NewSupportLogger returns a default logger for support services.
func NewSupportLogger() log.Logger {
	return log.NewLogger(os.Stdout)
}

// UpdateSupportRequest logs support request updates.
func (w *SupportChainLogger) UpdateSupportRequest(_ context.Context, msg *SupportUpdateRequest) error {
	if msg == nil {
		return nil
	}
	w.logger.Info("support update request",
		"ticket_id", msg.TicketID,
		"status", msg.Status,
		"assigned_agent", msg.AssignedAgent,
	)
	return nil
}

// AddSupportResponse logs support response submissions.
func (w *SupportChainLogger) AddSupportResponse(_ context.Context, msg *SupportAddResponse) error {
	if msg == nil {
		return nil
	}
	w.logger.Info("support response",
		"ticket_id", msg.TicketID,
		"author", msg.Author,
		"is_agent", msg.IsAgent,
	)
	return nil
}

// RegisterExternalTicket logs external ticket registration.
func (w *SupportChainLogger) RegisterExternalTicket(_ context.Context, msg *SupportRegisterExternal) error {
	if msg == nil {
		return nil
	}
	w.logger.Info("support external ticket",
		"ticket_id", msg.ResourceID,
		"external_system", msg.ExternalSystem,
		"external_ticket_id", msg.ExternalTicketID,
	)
	return nil
}

// SupportUpdateRequest represents a support request update payload.
type SupportUpdateRequest struct {
	TicketID      string
	Status        string
	AssignedAgent string
	UpdatedBy     string
	Metadata      map[string]string
}

// SupportAddResponse represents a support response payload.
type SupportAddResponse struct {
	TicketID string
	Author   string
	IsAgent  bool
	Message  string
	Payload  *SupportEncryptedPayload
}

// SupportRegisterExternal represents external ticket linkage.
type SupportRegisterExternal struct {
	ResourceID       string
	ResourceType     string
	ExternalSystem   string
	ExternalTicketID string
	ExternalURL      string
	CreatedBy        string
}

// SupportChainWriter submits support updates to chain.
type SupportChainWriter interface {
	UpdateSupportRequest(ctx context.Context, msg *SupportUpdateRequest) error
	AddSupportResponse(ctx context.Context, msg *SupportAddResponse) error
	RegisterExternalTicket(ctx context.Context, msg *SupportRegisterExternal) error
}

// SupportResponseEncryptor encrypts inbound support responses for on-chain storage.
type SupportResponseEncryptor interface {
	EncryptResponse(ctx context.Context, recipients []string, message string) (*SupportEncryptedPayload, error)
}

// SupportResponseEncryptorImpl encrypts responses with recipient public keys.
type SupportResponseEncryptorImpl struct {
	keyPair  *SupportKeyPair
	resolver RecipientKeyResolver
}

// NewSupportResponseEncryptor creates a response encryptor.
func NewSupportResponseEncryptor(cfg SupportServiceConfig, resolver RecipientKeyResolver) (*SupportResponseEncryptorImpl, error) {
	privateKey, err := cfg.Encryption.LoadPrivateKey()
	if err != nil {
		return nil, err
	}
	if len(privateKey) == 0 {
		return &SupportResponseEncryptorImpl{resolver: resolver}, nil
	}

	keyPair, err := NewSupportKeyPair(privateKey)
	if err != nil {
		return nil, err
	}

	return &SupportResponseEncryptorImpl{
		keyPair:  keyPair,
		resolver: resolver,
	}, nil
}

// EncryptResponse encrypts the response message for recipients.
func (e *SupportResponseEncryptorImpl) EncryptResponse(ctx context.Context, recipients []string, message string) (*SupportEncryptedPayload, error) {
	if message == "" {
		return nil, fmt.Errorf("response message is required")
	}
	if e == nil || e.keyPair == nil {
		return nil, fmt.Errorf("support encryption key not configured")
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("recipients are required for encryption")
	}
	publicKeys, err := e.resolver.ResolveRecipientPublicKeys(ctx, recipients)
	if err != nil {
		return nil, err
	}
	plaintext, err := BuildSupportResponsePayload(message)
	if err != nil {
		return nil, err
	}
	return EncryptSupportPayload(plaintext, publicKeys, e.keyPair)
}

// RecipientKeyResolver resolves recipient key IDs to public keys.
type RecipientKeyResolver interface {
	ResolveRecipientPublicKeys(ctx context.Context, keyIDs []string) ([][]byte, error)
}

// GRPCRecipientKeyResolver resolves keys using encryption gRPC queries.
type GRPCRecipientKeyResolver struct {
	client EncryptionQueryClient
}

// NewGRPCRecipientKeyResolver creates a new resolver.
func NewGRPCRecipientKeyResolver(conn *grpc.ClientConn) *GRPCRecipientKeyResolver {
	if conn == nil {
		return &GRPCRecipientKeyResolver{}
	}
	return &GRPCRecipientKeyResolver{client: NewEncryptionQueryClient(conn)}
}

// ResolveRecipientPublicKeys resolves public keys for fingerprints.
func (r *GRPCRecipientKeyResolver) ResolveRecipientPublicKeys(ctx context.Context, keyIDs []string) ([][]byte, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("encryption query client not configured")
	}
	publicKeys := make([][]byte, 0, len(keyIDs))
	for _, keyID := range keyIDs {
		resp, err := r.client.KeyByFingerprint(ctx, &EncryptionKeyByFingerprintRequest{Fingerprint: keyID})
		if err != nil {
			return nil, fmt.Errorf("lookup recipient key %s: %w", keyID, err)
		}
		if resp == nil || resp.Key == nil || len(resp.Key.PublicKey) == 0 {
			return nil, fmt.Errorf("recipient key %s not found", keyID)
		}
		publicKeys = append(publicKeys, append([]byte(nil), resp.Key.PublicKey...))
	}
	return publicKeys, nil
}
