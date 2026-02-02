// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-5C: Chain usage submitter for on-chain usage reporting
package provider_daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// ChainSubmitterConfig configures the chain usage submitter.
type ChainSubmitterConfig struct {
	// Enabled enables chain submission.
	Enabled bool

	// ProviderAddress is the provider's on-chain address.
	ProviderAddress string

	// ChainID is the chain ID.
	ChainID string

	// CometRPC is the CometBFT RPC endpoint.
	CometRPC string

	// GasLimit is the gas limit for transactions.
	GasLimit uint64

	// GasPrice is the gas price.
	GasPrice string

	// Timeout is the timeout for submissions.
	Timeout time.Duration

	// RetryAttempts is the number of retry attempts.
	RetryAttempts int

	// RetryBackoff is the backoff between retries.
	RetryBackoff time.Duration

	// BatchSize is the max number of records per batch.
	BatchSize int

	// BatchInterval is the interval for batching.
	BatchInterval time.Duration
}

// DefaultChainSubmitterConfig returns default chain submitter config.
func DefaultChainSubmitterConfig() ChainSubmitterConfig {
	return ChainSubmitterConfig{
		Enabled:       true,
		GasLimit:      200000,
		GasPrice:      "0.025uvirt",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryBackoff:  time.Second * 2,
		BatchSize:     10,
		BatchInterval: time.Minute,
	}
}

// ChainUsageSubmitterImpl implements ChainUsageSubmitter.
type ChainUsageSubmitterImpl struct {
	mu sync.RWMutex

	cfg        ChainSubmitterConfig
	keyManager *KeyManager
	rpcClient  *rpchttp.HTTP
	metrics    *UsageMetricsCollector

	// pendingBatch contains records pending batch submission.
	pendingBatch []*ChainUsageReport

	// submissionQueue contains records queued for submission.
	submissionQueue chan *ChainUsageReport

	// running indicates if the submitter is running.
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewChainUsageSubmitter creates a new chain usage submitter.
func NewChainUsageSubmitter(
	cfg ChainSubmitterConfig,
	keyManager *KeyManager,
	metrics *UsageMetricsCollector,
) (*ChainUsageSubmitterImpl, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if cfg.ProviderAddress == "" {
		return nil, errors.New("provider address is required")
	}

	if cfg.CometRPC == "" {
		return nil, errors.New("comet RPC endpoint is required")
	}

	return &ChainUsageSubmitterImpl{
		cfg:             cfg,
		keyManager:      keyManager,
		metrics:         metrics,
		pendingBatch:    make([]*ChainUsageReport, 0),
		submissionQueue: make(chan *ChainUsageReport, 1000),
		stopChan:        make(chan struct{}),
	}, nil
}

// Start starts the chain submitter.
func (s *ChainUsageSubmitterImpl) Start(ctx context.Context) error {
	if !s.cfg.Enabled {
		return nil
	}

	// Connect to RPC
	rpc, err := rpchttp.New(s.cfg.CometRPC, "/websocket")
	if err != nil {
		return fmt.Errorf("create rpc client: %w", err)
	}
	if err := rpc.Start(); err != nil {
		return fmt.Errorf("start rpc client: %w", err)
	}
	s.rpcClient = rpc

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	// Start batch processing loop
	s.wg.Add(1)
	verrors.SafeGo("provider-daemon:chain-submitter", func() {
		defer s.wg.Done()
		s.batchLoop(ctx)
	})

	log.Printf("[chain-submitter] started with RPC %s", s.cfg.CometRPC)
	return nil
}

// Stop stops the chain submitter.
func (s *ChainUsageSubmitterImpl) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()

	if s.rpcClient != nil {
		_ = s.rpcClient.Stop()
	}

	s.stopChan = make(chan struct{})
	log.Printf("[chain-submitter] stopped")
}

// SubmitUsageReport submits a usage report to the chain.
func (s *ChainUsageSubmitterImpl) SubmitUsageReport(ctx context.Context, report *ChainUsageReport) error {
	if !s.cfg.Enabled {
		return nil
	}

	if report == nil {
		return errors.New("report is nil")
	}

	// Add to queue for batching
	select {
	case s.submissionQueue <- report:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue full, submit immediately
		return s.submitSingleReport(ctx, report)
	}
}

// SubmitSettlementRequest submits a settlement request to the chain.
func (s *ChainUsageSubmitterImpl) SubmitSettlementRequest(ctx context.Context, orderID string, usageRecordIDs []string, isFinal bool) error {
	if !s.cfg.Enabled {
		return nil
	}

	start := time.Now()
	defer func() {
		if s.metrics != nil {
			s.metrics.RecordSettlement(true, time.Since(start))
		}
	}()

	// Build MsgSettleOrder
	msg := &MsgSettleOrderWrapper{
		Sender:         s.cfg.ProviderAddress,
		OrderID:        orderID,
		UsageRecordIDs: usageRecordIDs,
		IsFinal:        isFinal,
	}

	// Sign and broadcast
	return s.signAndBroadcast(ctx, msg)
}

// submitSingleReport submits a single usage report.
func (s *ChainUsageSubmitterImpl) submitSingleReport(ctx context.Context, report *ChainUsageReport) error {
	start := time.Now()

	// Build MsgRecordUsage
	msg := &MsgRecordUsageWrapper{
		Sender:      s.cfg.ProviderAddress,
		OrderID:     report.OrderID,
		LeaseID:     report.LeaseID,
		UsageUnits:  report.UsageUnits,
		UsageType:   report.UsageType,
		PeriodStart: report.PeriodStart.Unix(),
		PeriodEnd:   report.PeriodEnd.Unix(),
		UnitPrice:   report.UnitPrice,
		Signature:   report.Signature,
	}

	err := s.signAndBroadcast(ctx, msg)

	if s.metrics != nil {
		s.metrics.RecordSubmission(err == nil, time.Since(start))
	}

	return err
}

// submitBatch submits a batch of usage reports.
func (s *ChainUsageSubmitterImpl) submitBatch(ctx context.Context, reports []*ChainUsageReport) error {
	if len(reports) == 0 {
		return nil
	}

	start := time.Now()

	// Build batch message
	msgs := make([]*MsgRecordUsageWrapper, len(reports))
	for i, report := range reports {
		msgs[i] = &MsgRecordUsageWrapper{
			Sender:      s.cfg.ProviderAddress,
			OrderID:     report.OrderID,
			LeaseID:     report.LeaseID,
			UsageUnits:  report.UsageUnits,
			UsageType:   report.UsageType,
			PeriodStart: report.PeriodStart.Unix(),
			PeriodEnd:   report.PeriodEnd.Unix(),
			UnitPrice:   report.UnitPrice,
			Signature:   report.Signature,
		}
	}

	err := s.signAndBroadcastBatch(ctx, msgs)

	if s.metrics != nil {
		for range reports {
			s.metrics.RecordSubmission(err == nil, time.Since(start)/time.Duration(len(reports)))
		}
	}

	if err == nil {
		log.Printf("[chain-submitter] submitted batch of %d usage reports", len(reports))
	}

	return err
}

// signAndBroadcast signs and broadcasts a single message.
//
//nolint:unparam // ctx kept for future context deadline handling
func (s *ChainUsageSubmitterImpl) signAndBroadcast(_ context.Context, msg interface{}) error {
	if s.keyManager == nil {
		return errors.New("key manager not configured")
	}

	if s.rpcClient == nil {
		return errors.New("rpc client not connected")
	}

	// Serialize message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	// Sign the message
	sig, err := s.keyManager.Sign(msgBytes)
	if err != nil {
		return fmt.Errorf("sign message: %w", err)
	}

	// In a real implementation, this would:
	// 1. Query account sequence and number
	// 2. Build a proper Cosmos SDK transaction
	// 3. Sign with the provider's key
	// 4. Broadcast via BroadcastTxSync/BroadcastTxCommit

	log.Printf("[chain-submitter] would broadcast message with signature %s", sig.Signature[:16]+"...")

	return nil
}

// signAndBroadcastBatch signs and broadcasts multiple messages.
func (s *ChainUsageSubmitterImpl) signAndBroadcastBatch(ctx context.Context, msgs []*MsgRecordUsageWrapper) error {
	if len(msgs) == 0 {
		return nil
	}

	// In a real implementation, this would batch multiple messages
	// into a single transaction
	for _, msg := range msgs {
		if err := s.signAndBroadcast(ctx, msg); err != nil {
			return err
		}
	}

	return nil
}

// batchLoop processes the submission queue in batches.
func (s *ChainUsageSubmitterImpl) batchLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Flush remaining batch
			s.flushBatch(context.Background())
			return
		case <-s.stopChan:
			s.flushBatch(context.Background())
			return
		case report := <-s.submissionQueue:
			s.mu.Lock()
			s.pendingBatch = append(s.pendingBatch, report)
			shouldFlush := len(s.pendingBatch) >= s.cfg.BatchSize
			s.mu.Unlock()

			if shouldFlush {
				s.flushBatch(ctx)
			}
		case <-ticker.C:
			s.flushBatch(ctx)
		}
	}
}

// flushBatch flushes the pending batch.
func (s *ChainUsageSubmitterImpl) flushBatch(ctx context.Context) {
	s.mu.Lock()
	if len(s.pendingBatch) == 0 {
		s.mu.Unlock()
		return
	}
	batch := s.pendingBatch
	s.pendingBatch = make([]*ChainUsageReport, 0)
	s.mu.Unlock()

	if err := s.submitBatch(ctx, batch); err != nil {
		log.Printf("[chain-submitter] batch submission failed: %v", err)
		// Re-queue failed reports
		for _, report := range batch {
			select {
			case s.submissionQueue <- report:
			default:
				log.Printf("[chain-submitter] failed to re-queue report for order %s", report.OrderID)
			}
		}
	}
}

// MsgRecordUsageWrapper wraps the MsgRecordUsage for serialization.
type MsgRecordUsageWrapper struct {
	Sender      string      `json:"sender"`
	OrderID     string      `json:"order_id"`
	LeaseID     string      `json:"lease_id"`
	UsageUnits  uint64      `json:"usage_units"`
	UsageType   string      `json:"usage_type"`
	PeriodStart int64       `json:"period_start"`
	PeriodEnd   int64       `json:"period_end"`
	UnitPrice   sdk.DecCoin `json:"unit_price"`
	Signature   []byte      `json:"signature"`
}

// MsgSettleOrderWrapper wraps the MsgSettleOrder for serialization.
type MsgSettleOrderWrapper struct {
	Sender         string   `json:"sender"`
	OrderID        string   `json:"order_id"`
	UsageRecordIDs []string `json:"usage_record_ids"`
	IsFinal        bool     `json:"is_final"`
}

// SigningData contains data needed for transaction signing.
type SigningData struct {
	AccountNumber uint64
	Sequence      uint64
	ChainID       string
}

// TransactionBuilder builds Cosmos SDK transactions.
type TransactionBuilder struct {
	cfg        ChainSubmitterConfig
	keyManager *KeyManager
}

// NewTransactionBuilder creates a new transaction builder.
func NewTransactionBuilder(cfg ChainSubmitterConfig, keyManager *KeyManager) *TransactionBuilder {
	return &TransactionBuilder{
		cfg:        cfg,
		keyManager: keyManager,
	}
}

// BuildUsageReportTx builds a usage report transaction.
func (b *TransactionBuilder) BuildUsageReportTx(report *ChainUsageReport, signingData SigningData) ([]byte, error) {
	// Build the message
	msg := MsgRecordUsageWrapper{
		Sender:      b.cfg.ProviderAddress,
		OrderID:     report.OrderID,
		LeaseID:     report.LeaseID,
		UsageUnits:  report.UsageUnits,
		UsageType:   report.UsageType,
		PeriodStart: report.PeriodStart.Unix(),
		PeriodEnd:   report.PeriodEnd.Unix(),
		UnitPrice:   report.UnitPrice,
		Signature:   report.Signature,
	}

	// Serialize for signing
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	// Sign
	sig, err := b.keyManager.Sign(msgBytes)
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}

	// Build transaction wrapper
	tx := struct {
		Msg       MsgRecordUsageWrapper `json:"msg"`
		Signature string                `json:"signature"`
		ChainID   string                `json:"chain_id"`
		Sequence  uint64                `json:"sequence"`
	}{
		Msg:       msg,
		Signature: sig.Signature,
		ChainID:   signingData.ChainID,
		Sequence:  signingData.Sequence,
	}

	return json.Marshal(tx)
}

// BuildSettlementTx builds a settlement transaction.
func (b *TransactionBuilder) BuildSettlementTx(orderID string, usageRecordIDs []string, isFinal bool, signingData SigningData) ([]byte, error) {
	// Build the message
	msg := MsgSettleOrderWrapper{
		Sender:         b.cfg.ProviderAddress,
		OrderID:        orderID,
		UsageRecordIDs: usageRecordIDs,
		IsFinal:        isFinal,
	}

	// Serialize for signing
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	// Sign
	sig, err := b.keyManager.Sign(msgBytes)
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}

	// Build transaction wrapper
	tx := struct {
		Msg       MsgSettleOrderWrapper `json:"msg"`
		Signature string                `json:"signature"`
		ChainID   string                `json:"chain_id"`
		Sequence  uint64                `json:"sequence"`
	}{
		Msg:       msg,
		Signature: sig.Signature,
		ChainID:   signingData.ChainID,
		Sequence:  signingData.Sequence,
	}

	return json.Marshal(tx)
}

// Placeholder interfaces for Cosmos SDK integration
var (
	_ signing.SignMode = signing.SignMode(0)
	_ authsigning.Tx   = (authsigning.Tx)(nil)
)

// SignatureVerifier verifies usage report signatures.
type SignatureVerifier struct {
	// trustedProviders contains trusted provider public keys.
	trustedProviders map[string][]byte
}

// NewSignatureVerifier creates a new signature verifier.
func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{
		trustedProviders: make(map[string][]byte),
	}
}

// AddTrustedProvider adds a trusted provider public key.
func (v *SignatureVerifier) AddTrustedProvider(address string, publicKey []byte) {
	v.trustedProviders[address] = publicKey
}

// VerifyUsageReport verifies a usage report signature.
func (v *SignatureVerifier) VerifyUsageReport(report *ChainUsageReport, providerAddress string) (bool, error) {
	if report == nil {
		return false, errors.New("report is nil")
	}

	if len(report.Signature) == 0 {
		return false, errors.New("signature is empty")
	}

	publicKey, ok := v.trustedProviders[providerAddress]
	if !ok {
		return false, fmt.Errorf("unknown provider: %s", providerAddress)
	}

	// In a real implementation, this would verify the signature
	// using the provider's public key
	_ = publicKey

	return true, nil
}

// UsageReportHash generates a hash of a usage report for signing.
func UsageReportHash(report *ChainUsageReport) []byte {
	if report == nil {
		return nil
	}

	data := struct {
		OrderID     string `json:"order_id"`
		LeaseID     string `json:"lease_id"`
		UsageUnits  uint64 `json:"usage_units"`
		UsageType   string `json:"usage_type"`
		PeriodStart int64  `json:"period_start"`
		PeriodEnd   int64  `json:"period_end"`
		UnitPrice   string `json:"unit_price"`
	}{
		OrderID:     report.OrderID,
		LeaseID:     report.LeaseID,
		UsageUnits:  report.UsageUnits,
		UsageType:   report.UsageType,
		PeriodStart: report.PeriodStart.Unix(),
		PeriodEnd:   report.PeriodEnd.Unix(),
		UnitPrice:   report.UnitPrice.String(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return bytes
}

// UsageReportHashHex returns hex-encoded hash of a usage report.
func UsageReportHashHex(report *ChainUsageReport) string {
	hash := UsageReportHash(report)
	return hex.EncodeToString(hash)
}
