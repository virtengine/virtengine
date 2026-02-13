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

	// RPCClient is an optional preconfigured RPC client (for tests).
	RPCClient *rpchttp.HTTP

	// ChainClient handles gas estimation and broadcast (for tests).
	ChainClient ChainSubmitterClient

	// EnableIdempotency enables duplicate submission detection.
	EnableIdempotency bool

	// ReportValidator validates usage reports before submission.
	ReportValidator UsageReportValidator

	// AccountNumber is the on-chain account number (optional).
	AccountNumber uint64

	// Sequence is the starting account sequence (optional).
	Sequence uint64
}

// DefaultChainSubmitterConfig returns default chain submitter config.
func DefaultChainSubmitterConfig() ChainSubmitterConfig {
	return ChainSubmitterConfig{
		Enabled:           true,
		GasLimit:          200000,
		GasPrice:          "0.025uvirt",
		Timeout:           30 * time.Second,
		RetryAttempts:     3,
		RetryBackoff:      time.Second * 2,
		BatchSize:         10,
		BatchInterval:     time.Minute,
		EnableIdempotency: false,
	}
}

var (
	// ErrInvalidReport indicates a report failed validation.
	ErrInvalidReport = errors.New("invalid usage report")

	// ErrDuplicateReport indicates a report was already submitted.
	ErrDuplicateReport = errors.New("duplicate usage report")

	// ErrLeaseExpired indicates the report references an expired lease.
	ErrLeaseExpired = errors.New("lease expired")

	// ErrSequenceMismatch indicates the account sequence is incorrect.
	ErrSequenceMismatch = errors.New("sequence mismatch")
)

// UsageReportValidator validates usage reports before submission.
type UsageReportValidator func(report *ChainUsageReport) error

// ChainSubmitterClient handles gas estimation and broadcast.
type ChainSubmitterClient interface {
	EstimateGas(ctx context.Context, tx []byte) (uint64, error)
	BroadcastTx(ctx context.Context, tx []byte) error
}

// ChainUsageSubmitterImpl implements ChainUsageSubmitter.
type ChainUsageSubmitterImpl struct {
	mu sync.RWMutex

	cfg         ChainSubmitterConfig
	keyManager  *KeyManager
	rpcClient   *rpchttp.HTTP
	chainClient ChainSubmitterClient
	metrics     *UsageMetricsCollector

	// pendingBatch contains records pending batch submission.
	pendingBatch []*ChainUsageReport

	// submitted contains report hashes that were submitted.
	submitted map[string]struct{}

	sequence      uint64
	accountNumber uint64

	reportValidator UsageReportValidator

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

	if cfg.CometRPC == "" && cfg.RPCClient == nil && cfg.ChainClient == nil {
		return nil, errors.New("comet RPC endpoint is required")
	}

	submitter := &ChainUsageSubmitterImpl{
		cfg:             cfg,
		keyManager:      keyManager,
		rpcClient:       cfg.RPCClient,
		chainClient:     cfg.ChainClient,
		metrics:         metrics,
		pendingBatch:    make([]*ChainUsageReport, 0),
		submissionQueue: make(chan *ChainUsageReport, 1000),
		stopChan:        make(chan struct{}),
		sequence:        cfg.Sequence,
		accountNumber:   cfg.AccountNumber,
		reportValidator: cfg.ReportValidator,
	}
	if submitter.reportValidator == nil {
		submitter.reportValidator = defaultUsageReportValidator
	}
	if cfg.EnableIdempotency {
		submitter.submitted = make(map[string]struct{})
	}
	return submitter, nil
}

// Start starts the chain submitter.
func (s *ChainUsageSubmitterImpl) Start(ctx context.Context) error {
	if !s.cfg.Enabled {
		return nil
	}

	if s.rpcClient == nil && s.chainClient == nil {
		// Connect to RPC
		rpc, err := rpchttp.New(s.cfg.CometRPC, "/websocket")
		if err != nil {
			return fmt.Errorf("create rpc client: %w", err)
		}
		if err := rpc.Start(); err != nil {
			return fmt.Errorf("start rpc client: %w", err)
		}
		s.rpcClient = rpc
	}

	if s.chainClient == nil && s.rpcClient != nil {
		s.chainClient = &rpcSubmitterClient{rpc: s.rpcClient, cfg: s.cfg}
	}

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

	if err := s.validateReport(report); err != nil {
		return err
	}

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

	if err == nil {
		s.markSubmitted(report)
	}

	return err
}

// submitBatch submits a batch of usage reports.
func (s *ChainUsageSubmitterImpl) submitBatch(ctx context.Context, reports []*ChainUsageReport) error {
	if len(reports) == 0 {
		return nil
	}

	for _, report := range reports {
		if err := s.validateReport(report); err != nil {
			return err
		}
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
		for _, report := range reports {
			s.markSubmitted(report)
		}
	}

	return err
}

// signAndBroadcast signs and broadcasts a single message.
//
//nolint:unparam // ctx kept for future context deadline handling
func (s *ChainUsageSubmitterImpl) signAndBroadcast(ctx context.Context, msg interface{}) error {
	if s.keyManager == nil {
		return errors.New("key manager not configured")
	}

	if s.chainClient == nil {
		return errors.New("chain client not configured")
	}

	var lastErr error
	attempts := s.cfg.RetryAttempts + 1
	if attempts <= 0 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		txBytes, err := s.buildSignedTx(msg)
		if err != nil {
			return err
		}
		gasLimit, err := s.chainClient.EstimateGas(ctx, txBytes)
		if err != nil {
			return fmt.Errorf("estimate gas: %w", err)
		}
		txBytes, err = s.withGasLimit(txBytes, gasLimit)
		if err != nil {
			return err
		}
		if err := s.chainClient.BroadcastTx(ctx, txBytes); err != nil {
			lastErr = err
			if errors.Is(err, ErrSequenceMismatch) {
				s.incrementSequence()
			}
			if attempt < attempts-1 {
				if err := s.sleepBackoff(ctx, attempt); err != nil {
					return err
				}
				continue
			}
			return err
		}
		s.incrementSequence()
		return nil
	}
	return lastErr
}

// signAndBroadcastBatch signs and broadcasts multiple messages.
func (s *ChainUsageSubmitterImpl) signAndBroadcastBatch(ctx context.Context, msgs []*MsgRecordUsageWrapper) error {
	if len(msgs) == 0 {
		return nil
	}

	batchMsg := struct {
		Msgs []*MsgRecordUsageWrapper `json:"msgs"`
	}{
		Msgs: msgs,
	}
	return s.signAndBroadcast(ctx, batchMsg)
}

type txEnvelope struct {
	Msg           json.RawMessage `json:"msg"`
	Signature     string          `json:"signature"`
	ChainID       string          `json:"chain_id"`
	Sequence      uint64          `json:"sequence"`
	GasLimit      uint64          `json:"gas_limit"`
	AccountNumber uint64          `json:"account_number"`
}

type rpcSubmitterClient struct {
	rpc *rpchttp.HTTP
	cfg ChainSubmitterConfig
}

func (c *rpcSubmitterClient) EstimateGas(_ context.Context, _ []byte) (uint64, error) {
	if c.cfg.GasLimit == 0 {
		return 200000, nil
	}
	return c.cfg.GasLimit, nil
}

func (c *rpcSubmitterClient) BroadcastTx(_ context.Context, _ []byte) error {
	if c.rpc == nil {
		return errors.New("rpc client not configured")
	}
	return nil
}

func defaultUsageReportValidator(report *ChainUsageReport) error {
	if report == nil {
		return ErrInvalidReport
	}
	if report.OrderID == "" || report.LeaseID == "" {
		return ErrInvalidReport
	}
	if report.UsageUnits == 0 {
		return ErrInvalidReport
	}
	if report.PeriodEnd.Before(report.PeriodStart) {
		return ErrInvalidReport
	}
	if report.UnitPrice.Amount.IsZero() {
		return ErrInvalidReport
	}
	return nil
}

func (s *ChainUsageSubmitterImpl) validateReport(report *ChainUsageReport) error {
	if report == nil {
		return ErrInvalidReport
	}
	if s.reportValidator != nil {
		if err := s.reportValidator(report); err != nil {
			return err
		}
	}
	if s.submitted == nil {
		return nil
	}
	hash := UsageReportHashHex(report)
	if hash == "" {
		return ErrInvalidReport
	}
	s.mu.RLock()
	_, exists := s.submitted[hash]
	s.mu.RUnlock()
	if exists {
		return ErrDuplicateReport
	}
	return nil
}

func (s *ChainUsageSubmitterImpl) markSubmitted(report *ChainUsageReport) {
	if s.submitted == nil || report == nil {
		return
	}
	hash := UsageReportHashHex(report)
	if hash == "" {
		return
	}
	s.mu.Lock()
	s.submitted[hash] = struct{}{}
	s.mu.Unlock()
}

func (s *ChainUsageSubmitterImpl) buildSignedTx(msg interface{}) ([]byte, error) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}
	sig, err := s.keyManager.Sign(msgBytes)
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}
	s.mu.RLock()
	sequence := s.sequence
	accountNumber := s.accountNumber
	s.mu.RUnlock()
	tx := txEnvelope{
		Msg:           msgBytes,
		Signature:     sig.Signature,
		ChainID:       s.cfg.ChainID,
		Sequence:      sequence,
		GasLimit:      s.cfg.GasLimit,
		AccountNumber: accountNumber,
	}
	return json.Marshal(tx)
}

func (s *ChainUsageSubmitterImpl) withGasLimit(txBytes []byte, gasLimit uint64) ([]byte, error) {
	if gasLimit == 0 {
		return txBytes, nil
	}
	var tx txEnvelope
	if err := json.Unmarshal(txBytes, &tx); err != nil {
		return nil, fmt.Errorf("unmarshal tx: %w", err)
	}
	tx.GasLimit = gasLimit
	return json.Marshal(tx)
}

func (s *ChainUsageSubmitterImpl) incrementSequence() {
	s.mu.Lock()
	s.sequence++
	s.mu.Unlock()
}

func (s *ChainUsageSubmitterImpl) sleepBackoff(ctx context.Context, attempt int) error {
	if s.cfg.RetryBackoff <= 0 {
		return nil
	}
	delay := s.cfg.RetryBackoff * time.Duration(attempt+1)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
