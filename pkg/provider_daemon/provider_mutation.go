// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	tmtypes "github.com/cometbft/cometbft/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/gogoproto/proto"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	hpcmodule "github.com/virtengine/virtengine/x/hpc"
	marketmodule "github.com/virtengine/virtengine/x/market"
	marketplacemodule "github.com/virtengine/virtengine/x/marketplace"
	providermodule "github.com/virtengine/virtengine/x/provider"
	resourcesmodule "github.com/virtengine/virtengine/x/resources"
	settlementmodule "github.com/virtengine/virtengine/x/settlement"
	supportmodule "github.com/virtengine/virtengine/x/support"
)

const providerMutationSchemaVersion uint32 = 1

// ProviderMutationKind is a stable persisted identifier for a provider-originated write.
type ProviderMutationKind string

const (
	MutationMarketCreateBid           ProviderMutationKind = "market.create_bid"
	MutationMarketCloseBid            ProviderMutationKind = "market.close_bid"
	MutationMarketWithdrawLease       ProviderMutationKind = "market.withdraw_lease"
	MutationMarketCreateLease         ProviderMutationKind = "market.create_lease"
	MutationMarketCloseLease          ProviderMutationKind = "market.close_lease"
	MutationHPCRegisterCluster        ProviderMutationKind = "hpc.register_cluster"
	MutationHPCUpdateCluster          ProviderMutationKind = "hpc.update_cluster"
	MutationHPCDeregisterCluster      ProviderMutationKind = "hpc.deregister_cluster"
	MutationHPCCreateOffering         ProviderMutationKind = "hpc.create_offering"
	MutationHPCUpdateOffering         ProviderMutationKind = "hpc.update_offering"
	MutationHPCReportJobStatus        ProviderMutationKind = "hpc.report_job_status"
	MutationHPCUpdateNodeMetadata     ProviderMutationKind = "hpc.update_node_metadata"
	MutationResourcesHeartbeat        ProviderMutationKind = "resources.provider_heartbeat"
	MutationResourcesActivate         ProviderMutationKind = "resources.activate_allocation"
	MutationResourcesRelease          ProviderMutationKind = "resources.release_allocation"
	MutationSettlementRecordUsage     ProviderMutationKind = "settlement.record_usage"
	MutationSettlementSettleOrder     ProviderMutationKind = "settlement.settle_order"
	MutationSettlementFiatObservation ProviderMutationKind = "settlement.record_fiat_conversion_observation"
	MutationProviderCreate            ProviderMutationKind = "provider.create"
	MutationProviderUpdate            ProviderMutationKind = "provider.update"
	MutationProviderDelete            ProviderMutationKind = "provider.delete"
	MutationProviderRequestDomain     ProviderMutationKind = "provider.request_domain_verification"
	MutationProviderConfirmDomain     ProviderMutationKind = "provider.confirm_domain_verification"
	MutationProviderRevokeDomain      ProviderMutationKind = "provider.revoke_domain_verification"
	MutationProviderGenerateDomain    ProviderMutationKind = "provider.generate_domain_token"
	MutationProviderVerifyDomain      ProviderMutationKind = "provider.verify_domain"
	MutationProviderSetSigningKey     ProviderMutationKind = "provider.set_signing_key"
	MutationProviderRotateKey         ProviderMutationKind = "provider.rotate_signing_key"
	MutationProviderRevokeKey         ProviderMutationKind = "provider.revoke_signing_key"
	MutationMarketplaceCallback       ProviderMutationKind = "marketplace.waldur_callback"
	MutationSupportUpdateRequest      ProviderMutationKind = "support.update_request"
	MutationSupportAddResponse        ProviderMutationKind = "support.add_response"
	MutationSupportRegisterExternal   ProviderMutationKind = "support.register_external"
	MutationSupportUpdateExternal     ProviderMutationKind = "support.update_external"
)

// ProviderMutationState records a durable stage. In-progress stages are recovered
// as ambiguous after restart and reconciled before any replacement transaction.
type ProviderMutationState string

const (
	MutationStatePending      ProviderMutationState = "pending"
	MutationStateBuilding     ProviderMutationState = "building"
	MutationStateBuilt        ProviderMutationState = "built"
	MutationStateBroadcasting ProviderMutationState = "broadcasting"
	MutationStateAmbiguous    ProviderMutationState = "ambiguous"
	MutationStateIncluded     ProviderMutationState = "included"
	MutationStateConfirmed    ProviderMutationState = "confirmed"
	MutationStateRetry        ProviderMutationState = "retry"
	MutationStateDeadLetter   ProviderMutationState = "dead_letter"
)

// ProviderMutationClassification is persisted and intentionally bounded.
type ProviderMutationClassification string

const (
	MutationClassNone             ProviderMutationClassification = "none"
	MutationClassUnavailable      ProviderMutationClassification = "unavailable"
	MutationClassSequenceMismatch ProviderMutationClassification = "sequence_mismatch"
	MutationClassOutOfGas         ProviderMutationClassification = "out_of_gas"
	MutationClassMempoolReject    ProviderMutationClassification = "mempool_reject"
	MutationClassTimeout          ProviderMutationClassification = "timeout"
	MutationClassReplacement      ProviderMutationClassification = "replacement"
	MutationClassInvalid          ProviderMutationClassification = "invalid"
	MutationClassUnauthorized     ProviderMutationClassification = "unauthorized"
	MutationClassUnknown          ProviderMutationClassification = "unknown"
)

var (
	ErrProviderMutationUnavailable = errors.New("provider mutation pipeline unavailable")
	ErrProviderMutationNotReady    = errors.New("provider mutation pipeline not ready")
	ErrUnknownProviderMutation     = errors.New("unknown provider mutation type")
	ErrProviderMutationConflict    = errors.New("provider mutation idempotency conflict")
	ErrProviderMutationDeadLetter  = errors.New("provider mutation dead letter")
	ErrProviderMutationReorg       = errors.New("provider mutation confirmation reorged")
	ErrProviderMutationEvidence    = errors.New("provider mutation confirmation evidence incomplete")
	ErrProviderMutationStaleState  = errors.New("provider mutation stale state")
)

// ProviderMutationError is safe for callers and does not include message payloads.
type ProviderMutationError struct {
	Op             string
	MutationID     string
	Classification ProviderMutationClassification
	Retryable      bool
	Ambiguous      bool
	Err            error
}

func (e *ProviderMutationError) Error() string {
	if e == nil {
		return "provider mutation failed"
	}
	if e.MutationID == "" {
		return fmt.Sprintf("provider mutation %s failed (%s): %v", e.Op, e.Classification, e.Err)
	}
	return fmt.Sprintf("provider mutation %s failed id=%s (%s): %v", e.Op, e.MutationID, e.Classification, e.Err)
}

func (e *ProviderMutationError) Unwrap() error { return e.Err }

// ProviderMutationAttempt contains bounded, non-sensitive attempt evidence.
type ProviderMutationAttempt struct {
	Number         int                            `json:"number"`
	StartedAt      time.Time                      `json:"started_at"`
	FinishedAt     time.Time                      `json:"finished_at,omitempty"`
	AccountNumber  uint64                         `json:"account_number"`
	Sequence       uint64                         `json:"sequence"`
	GasWanted      uint64                         `json:"gas_wanted"`
	TxHash         string                         `json:"tx_hash,omitempty"`
	Classification ProviderMutationClassification `json:"classification"`
	Outcome        string                         `json:"outcome"`
}

// ProviderMutationEnvelope is the versioned durable queue record. Canonical
// protobuf payloads are stored because they are required to reconstruct and
// independently verify the signed SDK transaction after a crash.
type ProviderMutationEnvelope struct {
	SchemaVersion  uint32                `json:"schema_version"`
	ID             string                `json:"id"`
	Kind           ProviderMutationKind  `json:"kind"`
	TypeURL        string                `json:"type_url"`
	MessageBytes   []byte                `json:"message_bytes"`
	MessageDigest  string                `json:"message_digest"`
	Signer         string                `json:"signer"`
	IdempotencyKey string                `json:"idempotency_key"`
	State          ProviderMutationState `json:"state"`

	AccountNumber uint64 `json:"account_number,omitempty"`
	Sequence      uint64 `json:"sequence,omitempty"`
	GasLimit      uint64 `json:"gas_limit,omitempty"`
	TxBytes       []byte `json:"tx_bytes,omitempty"`
	TxDigest      string `json:"tx_digest,omitempty"`
	TxHash        string `json:"tx_hash,omitempty"`

	Attempts       []ProviderMutationAttempt      `json:"attempts,omitempty"`
	AttemptCount   int                            `json:"attempt_count"`
	Classification ProviderMutationClassification `json:"classification"`
	NextAttemptAt  time.Time                      `json:"next_attempt_at,omitempty"`
	CreatedAt      time.Time                      `json:"created_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
	LastAttemptAt  time.Time                      `json:"last_attempt_at,omitempty"`

	ConfirmationHeight    int64     `json:"confirmation_height,omitempty"`
	ConfirmationBlockHash string    `json:"confirmation_block_hash,omitempty"`
	IncludedAt            time.Time `json:"included_at,omitempty"`
	FinalityHeight        int64     `json:"finality_height,omitempty"`
	ConfirmedAt           time.Time `json:"confirmed_at,omitempty"`
	ConfirmationLatencyNS int64     `json:"confirmation_latency_ns,omitempty"`

	ReconciliationState string `json:"reconciliation_state,omitempty"`
	ReconciliationCount int    `json:"reconciliation_count"`
	TerminalResult      string `json:"terminal_result,omitempty"`
	DeadLetterReason    string `json:"dead_letter_reason,omitempty"`
	LeaseToken          uint64 `json:"lease_token,omitempty"`
}

// ProviderMutationResult is returned by enqueue and status operations.
type ProviderMutationResult struct {
	ID                 string
	IdempotencyKey     string
	State              ProviderMutationState
	TxHash             string
	ConfirmationHeight int64
	Final              bool
	Existed            bool
}

// ProviderMutationMetrics is a bounded snapshot with no payload-derived labels.
type ProviderMutationMetrics struct {
	QueueDepth          int
	OldestPendingAge    time.Duration
	MaxAttempts         int
	Ambiguous           int
	DeadLetters         int
	ConfirmationLatency time.Duration
	LastSuccess         time.Time
	LastFailure         time.Time
}

// ProviderMutationReadiness reports all mandatory production dependencies.
type ProviderMutationReadiness struct {
	Ready             bool
	Started           bool
	StoreReady        bool
	LeaseHeld         bool
	KeyReady          bool
	ChainReady        bool
	ConfirmationReady bool
	QueueDepth        int
	OldestPendingAge  time.Duration
	Reason            string
}

// ProviderMutationStore is pluggable for Task 85C. Implementations must make
// Put-if-absent and Update atomic and must return independent envelope copies.
type ProviderMutationStore interface {
	Open(context.Context) error
	Close() error
	PutIfAbsent(context.Context, *ProviderMutationEnvelope) (*ProviderMutationEnvelope, bool, error)
	Get(context.Context, string) (*ProviderMutationEnvelope, error)
	Update(context.Context, string, func(*ProviderMutationEnvelope) error) (*ProviderMutationEnvelope, error)
	List(context.Context) ([]*ProviderMutationEnvelope, error)
}

// QueueStore is the public 85A queue-store seam. It is an alias to the
// provider-specific name kept for backward compatibility with existing callers.
type QueueStore = ProviderMutationStore

// SubmitterLease provides the Task 85C fencing seam. Token must increase when
// ownership changes; 85A's local implementation enforces one owner per process.
type SubmitterLease interface {
	Acquire(context.Context, string, time.Duration) (uint64, error)
	Renew(context.Context, string, uint64, time.Duration) error
	Release(context.Context, string, uint64) error
	Held(context.Context, string, uint64) bool
}

// LocalSubmitterLease is the single-process implementation for Task 85A.
type LocalSubmitterLease struct {
	mu   sync.Mutex
	held map[string]localLeaseRecord
	next uint64
}

type localLeaseRecord struct {
	token   uint64
	expires time.Time
}

func NewLocalSubmitterLease() *LocalSubmitterLease {
	return &LocalSubmitterLease{held: make(map[string]localLeaseRecord)}
}

func (l *LocalSubmitterLease) Acquire(_ context.Context, name string, ttl time.Duration) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if record, ok := l.held[name]; ok && time.Now().Before(record.expires) {
		return 0, fmt.Errorf("submitter lease %s already held", name)
	}
	l.next++
	if l.next == 0 {
		l.next++
	}
	l.held[name] = localLeaseRecord{token: l.next, expires: time.Now().Add(ttl)}
	return l.next, nil
}

func (l *LocalSubmitterLease) Renew(_ context.Context, name string, token uint64, ttl time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	record, ok := l.held[name]
	if !ok || record.token != token || time.Now().After(record.expires) {
		return fmt.Errorf("submitter lease %s is not held", name)
	}
	record.expires = time.Now().Add(ttl)
	l.held[name] = record
	return nil
}

func (l *LocalSubmitterLease) Release(_ context.Context, name string, token uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if record, ok := l.held[name]; ok && record.token == token {
		delete(l.held, name)
	}
	return nil
}

func (l *LocalSubmitterLease) Held(_ context.Context, name string, token uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	record, ok := l.held[name]
	return ok && record.token == token && time.Now().Before(record.expires)
}

// ProviderMutationChain is the transport and reconciliation boundary.
type ProviderMutationChain interface {
	AccountSequenceResolver
	EstimateGas(context.Context, []byte) (uint64, error)
	BroadcastTx(context.Context, []byte) (string, error)
	ConfirmTx(context.Context, string) (ProviderTxConfirmation, error)
	LatestHeight(context.Context) (int64, error)
	BlockHash(context.Context, int64) (string, error)
	ReconcileMutation(context.Context, *ProviderMutationEnvelope, sdk.Msg) (ProviderMutationReconciliation, error)
}

type ProviderTxConfirmation struct {
	Found     bool
	TxHash    string
	Height    int64
	BlockHash string
	Code      uint32
	Log       string
}

type ProviderMutationReconciliation struct {
	Committed  bool
	Conflicted bool
	TxHash     string
	Height     int64
	BlockHash  string
	Reason     string
}

type mutationRegistration struct {
	Kind       ProviderMutationKind
	TypeURL    string
	New        func() sdk.Msg
	Validate   func(sdk.Msg) error
	Signer     func(sdk.Msg) (string, error)
	LogicalKey func(sdk.Msg) (string, error)
}

// ProviderMutationRegistry rejects unregistered protobuf types and owns all
// per-message validation, codec, signer and idempotency behavior.
type ProviderMutationRegistry struct {
	byKind    map[ProviderMutationKind]mutationRegistration
	byTypeURL map[string]mutationRegistration
}

func NewProviderMutationRegistry() *ProviderMutationRegistry {
	r := &ProviderMutationRegistry{
		byKind:    make(map[ProviderMutationKind]mutationRegistration),
		byTypeURL: make(map[string]mutationRegistration),
	}
	r.registerDefaults()
	return r
}

func (r *ProviderMutationRegistry) register(reg mutationRegistration) {
	r.byKind[reg.Kind] = reg
	r.byTypeURL[reg.TypeURL] = reg
}

func registration[T sdk.Msg](kind ProviderMutationKind, sample T, newMsg func() T, signer func(T) string, logical func(T) string) mutationRegistration {
	return mutationRegistration{
		Kind: kind, TypeURL: sdk.MsgTypeURL(sample),
		New: func() sdk.Msg { return newMsg() },
		Validate: func(msg sdk.Msg) error {
			typed, ok := msg.(T)
			if !ok {
				return fmt.Errorf("unexpected message type %T", msg)
			}
			if validator, ok := any(typed).(interface{ ValidateBasic() error }); ok {
				return validator.ValidateBasic()
			}
			return nil
		},
		Signer: func(msg sdk.Msg) (string, error) {
			typed, ok := msg.(T)
			if !ok {
				return "", fmt.Errorf("unexpected message type %T", msg)
			}
			value := strings.TrimSpace(signer(typed))
			if value == "" {
				return "", fmt.Errorf("message signer is required")
			}
			return value, nil
		},
		LogicalKey: func(msg sdk.Msg) (string, error) {
			typed, ok := msg.(T)
			if !ok {
				return "", fmt.Errorf("unexpected message type %T", msg)
			}
			value := strings.TrimSpace(logical(typed))
			if value == "" {
				return "", fmt.Errorf("message logical idempotency key is required")
			}
			return value, nil
		},
	}
}

func (r *ProviderMutationRegistry) registerDefaults() {
	r.register(registration(MutationMarketCreateBid, &marketv1beta5.MsgCreateBid{}, func() *marketv1beta5.MsgCreateBid { return &marketv1beta5.MsgCreateBid{} }, func(m *marketv1beta5.MsgCreateBid) string { return m.ID.Provider }, func(m *marketv1beta5.MsgCreateBid) string { return m.ID.String() }))
	r.register(registration(MutationMarketCloseBid, &marketv1beta5.MsgCloseBid{}, func() *marketv1beta5.MsgCloseBid { return &marketv1beta5.MsgCloseBid{} }, func(m *marketv1beta5.MsgCloseBid) string { return m.ID.Provider }, func(m *marketv1beta5.MsgCloseBid) string { return m.ID.String() }))
	r.register(registration(MutationMarketWithdrawLease, &marketv1beta5.MsgWithdrawLease{}, func() *marketv1beta5.MsgWithdrawLease { return &marketv1beta5.MsgWithdrawLease{} }, func(m *marketv1beta5.MsgWithdrawLease) string { return m.ID.Provider }, func(m *marketv1beta5.MsgWithdrawLease) string { return m.ID.String() }))
	// MsgCreateLease and MsgCloseLease are customer/owner-signed market
	// mutations. They are intentionally excluded from this provider signer
	// registry and covered by a negative contract test.
	r.register(registration(MutationHPCRegisterCluster, &hpcv1.MsgRegisterCluster{}, func() *hpcv1.MsgRegisterCluster { return &hpcv1.MsgRegisterCluster{} }, func(m *hpcv1.MsgRegisterCluster) string { return m.ProviderAddress }, func(m *hpcv1.MsgRegisterCluster) string {
		return strings.Join([]string{m.ProviderAddress, m.Region, m.Name}, "|")
	}))
	r.register(registration(MutationHPCUpdateCluster, &hpcv1.MsgUpdateCluster{}, func() *hpcv1.MsgUpdateCluster { return &hpcv1.MsgUpdateCluster{} }, func(m *hpcv1.MsgUpdateCluster) string { return m.ProviderAddress }, func(m *hpcv1.MsgUpdateCluster) string { return m.ClusterId }))
	r.register(registration(MutationHPCDeregisterCluster, &hpcv1.MsgDeregisterCluster{}, func() *hpcv1.MsgDeregisterCluster { return &hpcv1.MsgDeregisterCluster{} }, func(m *hpcv1.MsgDeregisterCluster) string { return m.ProviderAddress }, func(m *hpcv1.MsgDeregisterCluster) string { return m.ClusterId }))
	r.register(registration(MutationHPCCreateOffering, &hpcv1.MsgCreateOffering{}, func() *hpcv1.MsgCreateOffering { return &hpcv1.MsgCreateOffering{} }, func(m *hpcv1.MsgCreateOffering) string { return m.ProviderAddress }, func(m *hpcv1.MsgCreateOffering) string {
		return strings.Join([]string{m.ProviderAddress, m.ClusterId, m.Name}, "|")
	}))
	r.register(registration(MutationHPCUpdateOffering, &hpcv1.MsgUpdateOffering{}, func() *hpcv1.MsgUpdateOffering { return &hpcv1.MsgUpdateOffering{} }, func(m *hpcv1.MsgUpdateOffering) string { return m.ProviderAddress }, func(m *hpcv1.MsgUpdateOffering) string { return m.OfferingId }))
	r.register(registration(MutationHPCReportJobStatus, &hpcv1.MsgReportJobStatus{}, func() *hpcv1.MsgReportJobStatus { return &hpcv1.MsgReportJobStatus{} }, func(m *hpcv1.MsgReportJobStatus) string { return m.ProviderAddress }, func(m *hpcv1.MsgReportJobStatus) string {
		return fmt.Sprintf("%s|%d|%d", m.JobId, m.State, m.SignedTimestamp)
	}))
	r.register(registration(MutationHPCUpdateNodeMetadata, &hpcv1.MsgUpdateNodeMetadata{}, func() *hpcv1.MsgUpdateNodeMetadata { return &hpcv1.MsgUpdateNodeMetadata{} }, func(m *hpcv1.MsgUpdateNodeMetadata) string { return m.ProviderAddress }, func(m *hpcv1.MsgUpdateNodeMetadata) string {
		return fmt.Sprintf("%s|%d", m.NodeId, m.LastSequenceNumber)
	}))
	r.register(registration(MutationResourcesHeartbeat, &resourcesv1.MsgProviderHeartbeat{}, func() *resourcesv1.MsgProviderHeartbeat { return &resourcesv1.MsgProviderHeartbeat{} }, func(m *resourcesv1.MsgProviderHeartbeat) string { return m.ProviderAddress }, func(m *resourcesv1.MsgProviderHeartbeat) string {
		return fmt.Sprintf("%s|%d|%d", m.InventoryId, m.ResourceClass, m.Sequence)
	}))
	r.register(registration(MutationResourcesActivate, &resourcesv1.MsgActivateAllocation{}, func() *resourcesv1.MsgActivateAllocation { return &resourcesv1.MsgActivateAllocation{} }, func(m *resourcesv1.MsgActivateAllocation) string { return m.ProviderAddress }, func(m *resourcesv1.MsgActivateAllocation) string { return m.AllocationId }))
	r.register(registration(MutationResourcesRelease, &resourcesv1.MsgReleaseAllocation{}, func() *resourcesv1.MsgReleaseAllocation { return &resourcesv1.MsgReleaseAllocation{} }, func(m *resourcesv1.MsgReleaseAllocation) string { return m.RequesterAddress }, func(m *resourcesv1.MsgReleaseAllocation) string { return m.AllocationId }))
	r.register(registration(MutationSettlementRecordUsage, &settlementv1.MsgRecordUsage{}, func() *settlementv1.MsgRecordUsage { return &settlementv1.MsgRecordUsage{} }, func(m *settlementv1.MsgRecordUsage) string { return m.Sender }, func(m *settlementv1.MsgRecordUsage) string { return hex.EncodeToString(m.IdempotencyKey) }))
	r.register(registration(MutationSettlementSettleOrder, &settlementv1.MsgSettleOrder{}, func() *settlementv1.MsgSettleOrder { return &settlementv1.MsgSettleOrder{} }, func(m *settlementv1.MsgSettleOrder) string { return m.Sender }, func(m *settlementv1.MsgSettleOrder) string {
		return fmt.Sprintf("%s|%t|%s", m.OrderId, m.IsFinal, strings.Join(m.UsageRecordIds, ","))
	}))
	r.register(registration(MutationSettlementFiatObservation, &settlementv1.MsgRecordFiatConversionObservation{}, func() *settlementv1.MsgRecordFiatConversionObservation {
		return &settlementv1.MsgRecordFiatConversionObservation{}
	}, func(m *settlementv1.MsgRecordFiatConversionObservation) string { return m.Sender }, func(m *settlementv1.MsgRecordFiatConversionObservation) string {
		return fmt.Sprintf("%s|%d|%s", m.ConversionId, m.ObservationSequence, hex.EncodeToString(m.IdempotencyKey))
	}))
	r.register(registration(MutationProviderCreate, &providerv1beta4.MsgCreateProvider{}, func() *providerv1beta4.MsgCreateProvider { return &providerv1beta4.MsgCreateProvider{} }, func(m *providerv1beta4.MsgCreateProvider) string { return m.Owner }, func(m *providerv1beta4.MsgCreateProvider) string { return m.Owner }))
	r.register(registration(MutationProviderUpdate, &providerv1beta4.MsgUpdateProvider{}, func() *providerv1beta4.MsgUpdateProvider { return &providerv1beta4.MsgUpdateProvider{} }, func(m *providerv1beta4.MsgUpdateProvider) string { return m.Owner }, func(m *providerv1beta4.MsgUpdateProvider) string { return m.Owner }))
	r.register(registration(MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{}, func() *providerv1beta4.MsgDeleteProvider { return &providerv1beta4.MsgDeleteProvider{} }, func(m *providerv1beta4.MsgDeleteProvider) string { return m.Owner }, func(m *providerv1beta4.MsgDeleteProvider) string { return m.Owner }))
	r.register(registration(MutationProviderRequestDomain, &providerv1beta4.MsgRequestDomainVerification{}, func() *providerv1beta4.MsgRequestDomainVerification {
		return &providerv1beta4.MsgRequestDomainVerification{}
	}, func(m *providerv1beta4.MsgRequestDomainVerification) string { return m.Owner }, func(m *providerv1beta4.MsgRequestDomainVerification) string { return m.Owner + "|" + m.Domain }))
	r.register(registration(MutationProviderConfirmDomain, &providerv1beta4.MsgConfirmDomainVerification{}, func() *providerv1beta4.MsgConfirmDomainVerification {
		return &providerv1beta4.MsgConfirmDomainVerification{}
	}, func(m *providerv1beta4.MsgConfirmDomainVerification) string { return m.Owner }, func(m *providerv1beta4.MsgConfirmDomainVerification) string { return m.Owner }))
	r.register(registration(MutationProviderRevokeDomain, &providerv1beta4.MsgRevokeDomainVerification{}, func() *providerv1beta4.MsgRevokeDomainVerification {
		return &providerv1beta4.MsgRevokeDomainVerification{}
	}, func(m *providerv1beta4.MsgRevokeDomainVerification) string { return m.Owner }, func(m *providerv1beta4.MsgRevokeDomainVerification) string { return m.Owner }))
	r.register(registration(MutationProviderGenerateDomain, &providerv1beta4.MsgGenerateDomainVerificationToken{}, func() *providerv1beta4.MsgGenerateDomainVerificationToken {
		return &providerv1beta4.MsgGenerateDomainVerificationToken{}
	}, func(m *providerv1beta4.MsgGenerateDomainVerificationToken) string { return m.Owner }, func(m *providerv1beta4.MsgGenerateDomainVerificationToken) string { return m.Owner + "|" + m.Domain }))
	r.register(registration(MutationProviderVerifyDomain, &providerv1beta4.MsgVerifyProviderDomain{}, func() *providerv1beta4.MsgVerifyProviderDomain { return &providerv1beta4.MsgVerifyProviderDomain{} }, func(m *providerv1beta4.MsgVerifyProviderDomain) string { return m.Owner }, func(m *providerv1beta4.MsgVerifyProviderDomain) string { return m.Owner }))
	r.register(registration(MutationProviderSetSigningKey, &providerv1beta4.MsgSetProviderSigningKey{}, func() *providerv1beta4.MsgSetProviderSigningKey { return &providerv1beta4.MsgSetProviderSigningKey{} }, func(m *providerv1beta4.MsgSetProviderSigningKey) string { return m.Owner }, func(m *providerv1beta4.MsgSetProviderSigningKey) string {
		return m.Owner + "|" + providerv1beta4.ComputeProviderKeyID(m.KeyType, m.PublicKey)
	}))
	r.register(registration(MutationProviderRotateKey, &providerv1beta4.MsgRotateProviderSigningKey{}, func() *providerv1beta4.MsgRotateProviderSigningKey {
		return &providerv1beta4.MsgRotateProviderSigningKey{}
	}, func(m *providerv1beta4.MsgRotateProviderSigningKey) string { return m.Owner }, func(m *providerv1beta4.MsgRotateProviderSigningKey) string {
		return m.Owner + "|" + providerv1beta4.ComputeProviderKeyID(m.NewKeyType, m.NewPublicKey)
	}))
	r.register(registration(MutationProviderRevokeKey, &providerv1beta4.MsgRevokeProviderSigningKey{}, func() *providerv1beta4.MsgRevokeProviderSigningKey {
		return &providerv1beta4.MsgRevokeProviderSigningKey{}
	}, func(m *providerv1beta4.MsgRevokeProviderSigningKey) string { return m.Owner }, func(m *providerv1beta4.MsgRevokeProviderSigningKey) string { return m.Owner + "|" + m.KeyId }))
	r.register(registration(MutationMarketplaceCallback, &marketplacev1.MsgWaldurCallback{}, func() *marketplacev1.MsgWaldurCallback { return &marketplacev1.MsgWaldurCallback{} }, func(m *marketplacev1.MsgWaldurCallback) string { return m.Sender }, func(m *marketplacev1.MsgWaldurCallback) string {
		return strings.Join([]string{m.ResourceId, m.CallbackType, m.Status}, "|")
	}))
	r.register(registration(MutationSupportUpdateRequest, &supportv1.MsgUpdateSupportRequest{}, func() *supportv1.MsgUpdateSupportRequest { return &supportv1.MsgUpdateSupportRequest{} }, func(m *supportv1.MsgUpdateSupportRequest) string { return m.Sender }, func(m *supportv1.MsgUpdateSupportRequest) string {
		return m.TicketId + "|" + m.Status + "|" + m.AssignedAgent
	}))
	r.register(registration(MutationSupportAddResponse, &supportv1.MsgAddSupportResponse{}, func() *supportv1.MsgAddSupportResponse { return &supportv1.MsgAddSupportResponse{} }, func(m *supportv1.MsgAddSupportResponse) string { return m.Sender }, func(m *supportv1.MsgAddSupportResponse) string {
		return m.TicketId + "|" + hex.EncodeToString(m.Payload.EnvelopeHash)
	}))
	r.register(registration(MutationSupportRegisterExternal, &supportv1.MsgRegisterExternalTicket{}, func() *supportv1.MsgRegisterExternalTicket { return &supportv1.MsgRegisterExternalTicket{} }, func(m *supportv1.MsgRegisterExternalTicket) string { return m.Sender }, func(m *supportv1.MsgRegisterExternalTicket) string {
		return strings.Join([]string{m.ResourceType, m.ResourceId, m.ExternalSystem, m.ExternalTicketId}, "|")
	}))
	r.register(registration(MutationSupportUpdateExternal, &supportv1.MsgUpdateExternalTicket{}, func() *supportv1.MsgUpdateExternalTicket { return &supportv1.MsgUpdateExternalTicket{} }, func(m *supportv1.MsgUpdateExternalTicket) string { return m.Sender }, func(m *supportv1.MsgUpdateExternalTicket) string {
		return strings.Join([]string{m.ResourceType, m.ResourceId, m.ExternalTicketId, m.ExternalUrl}, "|")
	}))
}

func (r *ProviderMutationRegistry) Kinds() []ProviderMutationKind {
	kinds := make([]ProviderMutationKind, 0, len(r.byKind))
	for kind := range r.byKind {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
	return kinds
}

func deterministicProtoBytes(msg sdk.Msg) ([]byte, error) {
	return proto.Marshal(msg)
}

func providerMutationDigestBytes(msg sdk.Msg, messageBytes []byte) ([]byte, error) {
	switch typed := msg.(type) {
	case *supportv1.MsgUpdateSupportRequest:
		type supportUpdateDigest struct {
			Sender         string                             `json:"sender"`
			TicketID       string                             `json:"ticket_id"`
			Category       string                             `json:"category"`
			Priority       string                             `json:"priority"`
			Status         string                             `json:"status"`
			AssignedAgent  string                             `json:"assigned_agent"`
			Payload        *supportv1.EncryptedSupportPayload `json:"payload,omitempty"`
			PublicMetadata map[string]string                  `json:"public_metadata,omitempty"`
		}
		return json.Marshal(supportUpdateDigest{
			Sender:         typed.Sender,
			TicketID:       typed.TicketId,
			Category:       typed.Category,
			Priority:       typed.Priority,
			Status:         typed.Status,
			AssignedAgent:  typed.AssignedAgent,
			Payload:        typed.Payload,
			PublicMetadata: typed.PublicMetadata,
		})
	default:
		return messageBytes, nil
	}
}

func (r *ProviderMutationRegistry) Encode(chainID string, kind ProviderMutationKind, msg sdk.Msg) (*ProviderMutationEnvelope, error) {
	reg, ok := r.byKind[kind]
	if !ok || msg == nil {
		return nil, ErrUnknownProviderMutation
	}
	if typeURL := sdk.MsgTypeURL(msg); typeURL != reg.TypeURL {
		return nil, fmt.Errorf("%w: kind %s expects %s, got %s", ErrUnknownProviderMutation, kind, reg.TypeURL, typeURL)
	}
	if err := reg.Validate(msg); err != nil {
		return nil, fmt.Errorf("validate %s: %w", kind, err)
	}
	signer, err := reg.Signer(msg)
	if err != nil {
		return nil, err
	}
	logical, err := reg.LogicalKey(msg)
	if err != nil {
		return nil, err
	}
	messageBytes, err := deterministicProtoBytes(msg)
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", kind, err)
	}
	digestInput, err := providerMutationDigestBytes(msg, messageBytes)
	if err != nil {
		return nil, fmt.Errorf("digest %s: %w", kind, err)
	}
	digest := sha256.Sum256(digestInput)
	logicalDigest := sha256.Sum256([]byte(strings.Join([]string{chainID, string(kind), signer, logical}, "\x00")))
	key := hex.EncodeToString(logicalDigest[:])
	idDigest := sha256.Sum256([]byte(key + "\x00" + hex.EncodeToString(digest[:])))
	now := time.Now().UTC()
	return &ProviderMutationEnvelope{
		SchemaVersion: providerMutationSchemaVersion,
		ID:            hex.EncodeToString(idDigest[:16]), Kind: kind, TypeURL: reg.TypeURL,
		MessageBytes: messageBytes, MessageDigest: hex.EncodeToString(digest[:]), Signer: signer,
		IdempotencyKey: key, State: MutationStatePending, Classification: MutationClassNone,
		CreatedAt: now, UpdatedAt: now, NextAttemptAt: now,
	}, nil
}

func (r *ProviderMutationRegistry) Decode(envelope *ProviderMutationEnvelope) (sdk.Msg, error) {
	if envelope == nil || envelope.SchemaVersion != providerMutationSchemaVersion {
		return nil, fmt.Errorf("unsupported provider mutation envelope version")
	}
	reg, ok := r.byKind[envelope.Kind]
	if !ok || reg.TypeURL != envelope.TypeURL {
		return nil, ErrUnknownProviderMutation
	}
	msg := reg.New()
	if err := proto.Unmarshal(envelope.MessageBytes, msg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", envelope.Kind, err)
	}
	if err := reg.Validate(msg); err != nil {
		return nil, err
	}
	signer, err := reg.Signer(msg)
	if err != nil || signer != envelope.Signer {
		return nil, fmt.Errorf("provider mutation signer mismatch")
	}
	digestInput, err := providerMutationDigestBytes(msg, envelope.MessageBytes)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(digestInput)
	if !strings.EqualFold(envelope.MessageDigest, hex.EncodeToString(digest[:])) {
		return nil, fmt.Errorf("provider mutation message digest mismatch")
	}
	return msg, nil
}

// ProviderMutationSubmitterConfig controls the durable fenced pipeline.
type ProviderMutationSubmitterConfig struct {
	ChainID             string
	ProviderAddress     string
	QueueStatePath      string
	GasLimit            uint64
	GasAdjustment       float64
	MaxAttempts         int
	RetryBackoff        time.Duration
	MaxRetryBackoff     time.Duration
	PollInterval        time.Duration
	ConfirmationTimeout time.Duration
	FinalityBlocks      int64
	LeaseTTL            time.Duration
	Store               ProviderMutationStore
	Lease               SubmitterLease
	Registry            *ProviderMutationRegistry
	Chain               ProviderMutationChain
	Production          bool
	DevelopmentNoop     bool
}

func DefaultProviderMutationSubmitterConfig() ProviderMutationSubmitterConfig {
	return ProviderMutationSubmitterConfig{
		QueueStatePath: filepathJoinDefaultMutationQueue(), GasLimit: 200000, GasAdjustment: 1.25,
		MaxAttempts: 6, RetryBackoff: 2 * time.Second, MaxRetryBackoff: time.Minute,
		PollInterval: 2 * time.Second, ConfirmationTimeout: 45 * time.Second,
		FinalityBlocks: 2, LeaseTTL: 30 * time.Second, Production: true,
	}
}

func filepathJoinDefaultMutationQueue() string {
	return ".cache/provider_daemon/provider_mutation_queue.json"
}

// ProviderMutationSubmitter is one type-safe mutation owner per signer.
type ProviderMutationSubmitter struct {
	cfg        ProviderMutationSubmitterConfig
	registry   *ProviderMutationRegistry
	store      ProviderMutationStore
	lease      SubmitterLease
	chain      ProviderMutationChain
	keyManager *KeyManager
	encCfg     sdkutil.EncodingConfig

	mu          sync.RWMutex
	processMu   sync.Mutex
	running     bool
	leaseName   string
	leaseToken  uint64
	stopCh      chan struct{}
	wg          sync.WaitGroup
	lastSuccess time.Time
	lastFailure time.Time
	storeReady  bool
}

func NewProviderMutationSubmitter(cfg ProviderMutationSubmitterConfig, keyManager *KeyManager) (*ProviderMutationSubmitter, error) {
	defaults := DefaultProviderMutationSubmitterConfig()
	if cfg.ChainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}
	if cfg.ProviderAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}
	if cfg.Production && cfg.DevelopmentNoop {
		return nil, fmt.Errorf("development no-op is forbidden in production")
	}
	if !cfg.DevelopmentNoop && cfg.Chain == nil {
		return nil, fmt.Errorf("%w: chain transport is required", ErrProviderMutationUnavailable)
	}
	if keyManager == nil || keyManager.IsLocked() {
		return nil, fmt.Errorf("%w: unlocked key manager is required", ErrProviderMutationUnavailable)
	}
	if cfg.QueueStatePath == "" {
		cfg.QueueStatePath = defaults.QueueStatePath
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = defaults.GasLimit
	}
	if cfg.GasAdjustment < 1 {
		cfg.GasAdjustment = defaults.GasAdjustment
	}
	if math.IsNaN(cfg.GasAdjustment) || math.IsInf(cfg.GasAdjustment, 0) {
		return nil, fmt.Errorf("gas adjustment must be finite")
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaults.MaxAttempts
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = defaults.RetryBackoff
	}
	if cfg.MaxRetryBackoff <= 0 {
		cfg.MaxRetryBackoff = defaults.MaxRetryBackoff
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaults.PollInterval
	}
	if cfg.ConfirmationTimeout <= 0 {
		cfg.ConfirmationTimeout = defaults.ConfirmationTimeout
	}
	if cfg.FinalityBlocks < 0 {
		return nil, fmt.Errorf("finality blocks cannot be negative")
	}
	if cfg.LeaseTTL <= 0 {
		cfg.LeaseTTL = defaults.LeaseTTL
	}
	if cfg.LeaseTTL < 3*time.Nanosecond {
		return nil, fmt.Errorf("lease TTL must be at least 3ns")
	}
	registry := cfg.Registry
	if registry == nil {
		registry = NewProviderMutationRegistry()
	}
	store := cfg.Store
	if store == nil {
		fileStore, err := NewFileProviderMutationStore(cfg.QueueStatePath)
		if err != nil {
			return nil, err
		}
		store = fileStore
	}
	lease := cfg.Lease
	if lease == nil {
		if cfg.Production {
			return nil, fmt.Errorf("%w: production requires an explicit durable submitter lease", ErrProviderMutationUnavailable)
		}
		lease = NewLocalSubmitterLease()
	}
	if cfg.Production {
		if _, unsafe := lease.(*LocalSubmitterLease); unsafe {
			return nil, fmt.Errorf("%w: process-local submitter lease is forbidden in production", ErrProviderMutationUnavailable)
		}
	}
	encCfg := sdkutil.MakeEncodingConfig(
		marketmodule.AppModuleBasic{}, hpcmodule.AppModuleBasic{}, resourcesmodule.AppModuleBasic{},
		providermodule.AppModuleBasic{}, settlementmodule.AppModuleBasic{}, marketplacemodule.AppModuleBasic{}, supportmodule.AppModuleBasic{},
	)
	return &ProviderMutationSubmitter{
		cfg: cfg, registry: registry, store: store, lease: lease, chain: cfg.Chain,
		keyManager: keyManager, encCfg: encCfg, leaseName: "provider-mutation:" + cfg.ProviderAddress,
		stopCh: make(chan struct{}),
	}, nil
}

func (s *ProviderMutationSubmitter) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	if err := s.store.Open(ctx); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("open mutation store: %w", err)
	}
	s.storeReady = true
	token, err := s.lease.Acquire(ctx, s.leaseName, s.cfg.LeaseTTL)
	if err != nil {
		if !s.cfg.Production {
			_ = s.store.Close()
			s.storeReady = false
			s.mu.Unlock()
			return err
		}
		token = 0
	}
	s.leaseToken = token
	if token != 0 {
		if err := s.recoverInProgress(ctx, token); err != nil {
			_ = s.lease.Release(ctx, s.leaseName, token)
			_ = s.store.Close()
			s.storeReady = false
			s.mu.Unlock()
			return err
		}
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.wg.Add(1)
	s.mu.Unlock()
	go s.worker(ctx)
	return nil
}

func (s *ProviderMutationSubmitter) activateLease(ctx context.Context, token uint64) error {
	if token == 0 {
		return fmt.Errorf("%w: zero fencing token", ErrProviderMutationNotReady)
	}
	if err := s.recoverInProgress(ctx, token); err != nil {
		_ = s.lease.Release(ctx, s.leaseName, token)
		return err
	}
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		_ = s.lease.Release(ctx, s.leaseName, token)
		return ErrProviderMutationNotReady
	}
	s.leaseToken = token
	s.mu.Unlock()
	return nil
}

func (s *ProviderMutationSubmitter) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.stopCh)
	token := s.leaseToken
	s.mu.Unlock()
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	_ = s.lease.Release(context.Background(), s.leaseName, token)
	err := s.store.Close()
	s.mu.Lock()
	s.storeReady = false
	s.mu.Unlock()
	return err
}

func (s *ProviderMutationSubmitter) recoverInProgress(ctx context.Context, token uint64) error {
	items, err := s.store.List(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.State == MutationStateBuilding || item.State == MutationStateBuilt || item.State == MutationStateBroadcasting || item.State == MutationStateIncluded {
			_, err = s.store.Update(ctx, item.ID, func(stored *ProviderMutationEnvelope) error {
				stored.State = MutationStateAmbiguous
				stored.LeaseToken = token
				stored.ReconciliationState = "restart_reconciliation_required"
				stored.NextAttemptAt = time.Now().UTC()
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ProviderMutationSubmitter) worker(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.cfg.PollInterval)
	acquireTicker := time.NewTicker(s.cfg.PollInterval)
	leaseInterval := s.cfg.LeaseTTL / 3
	if leaseInterval <= 0 {
		leaseInterval = time.Nanosecond
	}
	leaseTicker := time.NewTicker(leaseInterval)
	defer ticker.Stop()
	defer acquireTicker.Stop()
	defer leaseTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-acquireTicker.C:
			s.mu.RLock()
			token := s.leaseToken
			s.mu.RUnlock()
			if token != 0 {
				continue
			}
			acquired, err := s.lease.Acquire(ctx, s.leaseName, s.cfg.LeaseTTL)
			if err != nil {
				continue
			}
			if err := s.activateLease(ctx, acquired); err != nil {
				s.mu.Lock()
				s.lastFailure = time.Now().UTC()
				s.mu.Unlock()
			}
		case <-leaseTicker.C:
			s.mu.RLock()
			token := s.leaseToken
			s.mu.RUnlock()
			if token == 0 {
				continue
			}
			if err := s.lease.Renew(ctx, s.leaseName, token, s.cfg.LeaseTTL); err != nil {
				s.mu.Lock()
				s.leaseToken = 0
				s.lastFailure = time.Now().UTC()
				s.mu.Unlock()
			}
		case <-ticker.C:
			if s.Readiness(ctx).Ready {
				_ = s.ProcessDue(ctx, 32)
			}
		}
	}
}

func (s *ProviderMutationSubmitter) Submit(ctx context.Context, kind ProviderMutationKind, msg sdk.Msg) (ProviderMutationResult, error) {
	if msg == nil {
		return ProviderMutationResult{}, fmt.Errorf("mutation message is nil")
	}
	readiness := s.Readiness(ctx)
	if !readiness.Ready {
		return ProviderMutationResult{}, fmt.Errorf("%w: %s", ErrProviderMutationNotReady, readiness.Reason)
	}
	envelope, err := s.registry.Encode(s.cfg.ChainID, kind, msg)
	if err != nil {
		return ProviderMutationResult{}, err
	}
	if envelope.Signer != s.cfg.ProviderAddress {
		return ProviderMutationResult{}, fmt.Errorf("mutation signer %s does not match configured provider", envelope.Signer)
	}
	s.mu.RLock()
	leaseToken := s.leaseToken
	s.mu.RUnlock()
	envelope.LeaseToken = leaseToken
	stored, existed, err := s.store.PutIfAbsent(ctx, envelope)
	if err != nil {
		return ProviderMutationResult{}, err
	}
	if existed && !strings.EqualFold(stored.MessageDigest, envelope.MessageDigest) {
		return ProviderMutationResult{}, ErrProviderMutationConflict
	}
	result := resultFromEnvelope(stored, existed)
	if stored.State == MutationStateConfirmed || stored.State == MutationStateIncluded {
		return result, nil
	}
	if stored.State == MutationStateDeadLetter {
		return result, &ProviderMutationError{Op: "submit", MutationID: stored.ID, Classification: stored.Classification, Err: ErrProviderMutationDeadLetter}
	}
	err = s.process(ctx, stored.ID)
	latest, getErr := s.store.Get(ctx, stored.ID)
	if getErr == nil {
		result = resultFromEnvelope(latest, existed)
	}
	return result, err
}

// SubmitFiatConversionObservation durably submits one provider-signed
// settlement observation through the same crash-safe mutation owner.
func (s *ProviderMutationSubmitter) SubmitFiatConversionObservation(ctx context.Context, msg *settlementv1.MsgRecordFiatConversionObservation) (ProviderMutationResult, error) {
	return s.Submit(ctx, MutationSettlementFiatObservation, msg)
}

func (s *ProviderMutationSubmitter) SubmitAndWait(ctx context.Context, kind ProviderMutationKind, msg sdk.Msg) (ProviderMutationResult, error) {
	result, err := s.Submit(ctx, kind, msg)
	if err != nil && !isProviderMutationRetryable(err) {
		return result, err
	}
	return s.Wait(ctx, result.ID)
}

func (s *ProviderMutationSubmitter) Wait(ctx context.Context, id string) (ProviderMutationResult, error) {
	ticker := time.NewTicker(minDuration(s.cfg.PollInterval, time.Second))
	defer ticker.Stop()
	for {
		envelope, err := s.store.Get(ctx, id)
		if err != nil {
			return ProviderMutationResult{}, err
		}
		result := resultFromEnvelope(envelope, false)
		switch envelope.State {
		case MutationStateConfirmed:
			return result, nil
		case MutationStateDeadLetter:
			return result, &ProviderMutationError{Op: "wait", MutationID: id, Classification: envelope.Classification, Err: ErrProviderMutationDeadLetter}
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-ticker.C:
			_ = s.process(ctx, id)
		}
	}
}

func (s *ProviderMutationSubmitter) Status(ctx context.Context, id string) (ProviderMutationResult, error) {
	envelope, err := s.store.Get(ctx, id)
	if err != nil {
		return ProviderMutationResult{}, err
	}
	return resultFromEnvelope(envelope, false), nil
}

func resultFromEnvelope(envelope *ProviderMutationEnvelope, existed bool) ProviderMutationResult {
	if envelope == nil {
		return ProviderMutationResult{}
	}
	return ProviderMutationResult{
		ID: envelope.ID, IdempotencyKey: envelope.IdempotencyKey, State: envelope.State,
		TxHash: envelope.TxHash, ConfirmationHeight: envelope.ConfirmationHeight,
		Final: envelope.State == MutationStateConfirmed, Existed: existed,
	}
}

func providerMutationTerminalState(state ProviderMutationState) bool {
	return state == MutationStateConfirmed || state == MutationStateDeadLetter
}

func (s *ProviderMutationSubmitter) ensureLeaseHeld(ctx context.Context) error {
	if s == nil || s.lease == nil {
		return fmt.Errorf("%w: submitter lease lost", ErrProviderMutationNotReady)
	}
	s.mu.RLock()
	token := s.leaseToken
	s.mu.RUnlock()
	if !s.lease.Held(ctx, s.leaseName, token) {
		return fmt.Errorf("%w: submitter lease lost", ErrProviderMutationNotReady)
	}
	return nil
}

func (s *ProviderMutationSubmitter) updateOwned(ctx context.Context, id string, fn func(*ProviderMutationEnvelope) error) (*ProviderMutationEnvelope, error) {
	if err := s.ensureLeaseHeld(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	token := s.leaseToken
	s.mu.RUnlock()
	return s.store.Update(ctx, id, func(item *ProviderMutationEnvelope) error {
		if item.LeaseToken != 0 && item.LeaseToken != token {
			return fmt.Errorf("%w: lease token changed", ErrProviderMutationNotReady)
		}
		item.LeaseToken = token
		if providerMutationTerminalState(item.State) {
			return ErrProviderMutationStaleState
		}
		return fn(item)
	})
}

func (s *ProviderMutationSubmitter) ProcessDue(ctx context.Context, limit int) error {
	items, err := s.store.List(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	processed := 0
	for _, item := range items {
		if limit > 0 && processed >= limit {
			break
		}
		if item.State == MutationStateConfirmed || item.State == MutationStateDeadLetter || item.NextAttemptAt.After(now) {
			continue
		}
		processed++
		if err := s.process(ctx, item.ID); err != nil && !isProviderMutationRetryable(err) {
			return err
		}
	}
	return nil
}

func (s *ProviderMutationSubmitter) process(ctx context.Context, id string) error {
	s.processMu.Lock()
	defer s.processMu.Unlock()
	s.mu.RLock()
	token := s.leaseToken
	s.mu.RUnlock()
	if !s.lease.Held(ctx, s.leaseName, token) {
		return fmt.Errorf("%w: submitter lease lost", ErrProviderMutationNotReady)
	}
	envelope, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}
	msg, err := s.registry.Decode(envelope)
	if err != nil {
		return s.deadLetter(ctx, envelope.ID, MutationClassInvalid, "registry_validation_failed", err)
	}
	if envelope.State == MutationStateConfirmed || envelope.State == MutationStateDeadLetter {
		return nil
	}
	if envelope.State == MutationStateAmbiguous || envelope.State == MutationStateIncluded || envelope.TxHash != "" {
		resolved, reconcileErr := s.reconcile(ctx, envelope, msg)
		if reconcileErr != nil {
			return s.schedule(ctx, envelope.ID, reconcileErr)
		}
		if resolved {
			return nil
		}
		envelope, err = s.store.Get(ctx, id)
		if err != nil {
			return err
		}
	}
	if envelope.AttemptCount >= s.cfg.MaxAttempts {
		return s.deadLetter(ctx, envelope.ID, envelope.Classification, "attempts_exhausted", ErrProviderMutationDeadLetter)
	}
	return s.buildBroadcastConfirm(ctx, envelope, msg)
}

func (s *ProviderMutationSubmitter) buildBroadcastConfirm(ctx context.Context, envelope *ProviderMutationEnvelope, msg sdk.Msg) error {
	accountNumber, sequence, err := s.chain.ResolveAccountSequence(ctx, envelope.Signer)
	if err != nil {
		return s.schedule(ctx, envelope.ID, err)
	}
	updated, err := s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
		item.State = MutationStateBuilding
		item.AccountNumber = accountNumber
		item.Sequence = sequence
		item.LastAttemptAt = time.Now().UTC()
		item.AttemptCount++
		item.Attempts = appendBoundedAttempt(item.Attempts, ProviderMutationAttempt{Number: item.AttemptCount, StartedAt: item.LastAttemptAt, AccountNumber: accountNumber, Sequence: sequence, Outcome: "building", Classification: MutationClassNone})
		return nil
	})
	if err != nil {
		return err
	}
	attemptNumber := updated.AttemptCount
	baseGasLimit := maxUint64(s.cfg.GasLimit, updated.GasLimit)
	txBytes, err := s.buildSignedTx(msg, accountNumber, sequence, baseGasLimit)
	if err != nil {
		return s.deadLetter(ctx, envelope.ID, MutationClassInvalid, "tx_build_failed", err)
	}
	estimated, err := s.chain.EstimateGas(ctx, txBytes)
	if err != nil {
		return s.schedule(ctx, envelope.ID, fmt.Errorf("simulate transaction: %w", err))
	}
	gasLimit := estimated
	if gasLimit == 0 {
		gasLimit = baseGasLimit
	}
	if gasLimit < baseGasLimit {
		gasLimit = baseGasLimit
	}
	gasLimit, err = checkedAdjustedGasLimit(gasLimit, s.cfg.GasAdjustment)
	if err != nil {
		return s.deadLetter(ctx, envelope.ID, MutationClassInvalid, "gas_adjustment_invalid", err)
	}
	txBytes, err = s.buildSignedTx(msg, accountNumber, sequence, gasLimit)
	if err != nil {
		return s.deadLetter(ctx, envelope.ID, MutationClassInvalid, "tx_rebuild_failed", err)
	}
	txDigest := sha256.Sum256(txBytes)
	localHash := strings.ToUpper(hex.EncodeToString(tmtypes.Tx(txBytes).Hash()))
	_, err = s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
		if item.State != MutationStateBuilding || item.AttemptCount != attemptNumber {
			return ErrProviderMutationStaleState
		}
		item.State = MutationStateBuilt
		item.GasLimit = gasLimit
		item.TxBytes = append([]byte(nil), txBytes...)
		item.TxDigest = hex.EncodeToString(txDigest[:])
		item.TxHash = localHash
		updateLatestAttempt(item, "built", MutationClassNone, gasLimit, localHash)
		return nil
	})
	if err != nil {
		return err
	}
	_, err = s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
		if item.State != MutationStateBuilt || item.AttemptCount != attemptNumber || item.TxHash != localHash {
			return ErrProviderMutationStaleState
		}
		item.State = MutationStateBroadcasting
		updateLatestAttempt(item, "broadcasting", MutationClassNone, gasLimit, localHash)
		return nil
	})
	if err != nil {
		return err
	}
	if err := s.ensureLeaseHeld(ctx); err != nil {
		return err
	}
	txHash, broadcastErr := s.chain.BroadcastTx(ctx, txBytes)
	if txHash == "" {
		txHash = localHash
	}
	if broadcastErr != nil {
		classification, retryable, ambiguous := classifyProviderMutationError(broadcastErr)
		if classification == MutationClassSequenceMismatch {
			ambiguous = true
		}
		if ambiguous {
			_, updateErr := s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
				if item.State != MutationStateBroadcasting || item.AttemptCount != attemptNumber {
					return ErrProviderMutationStaleState
				}
				item.State = MutationStateAmbiguous
				item.Classification = classification
				item.TxHash = txHash
				item.ReconciliationState = "broadcast_outcome_ambiguous"
				updateLatestAttempt(item, "ambiguous", classification, gasLimit, txHash)
				return nil
			})
			if updateErr != nil {
				return updateErr
			}
			latest, _ := s.store.Get(ctx, envelope.ID)
			if resolved, reconcileErr := s.reconcile(ctx, latest, msg); reconcileErr == nil && resolved {
				return nil
			}
			latest, _ = s.store.Get(ctx, envelope.ID)
			if latest != nil && latest.State == MutationStateDeadLetter {
				return &ProviderMutationError{Op: "broadcast", MutationID: envelope.ID, Classification: latest.Classification, Err: ErrProviderMutationDeadLetter}
			}
		}
		if retryable {
			if classification == MutationClassOutOfGas {
				if _, bumpErr := s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
					bumped, bumpErr := checkedAdjustedGasLimit(maxUint64(item.GasLimit, gasLimit), 1.5)
					if bumpErr != nil {
						return bumpErr
					}
					item.TxBytes = nil
					item.TxDigest = ""
					item.TxHash = ""
					item.GasLimit = bumped
					return nil
				}); bumpErr != nil {
					return bumpErr
				}
			}
			return s.schedule(ctx, envelope.ID, broadcastErr)
		}
		return s.deadLetter(ctx, envelope.ID, classification, "broadcast_rejected", broadcastErr)
	}
	_, err = s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
		if item.State != MutationStateBroadcasting || item.AttemptCount != attemptNumber {
			return ErrProviderMutationStaleState
		}
		item.State = MutationStateAmbiguous
		item.TxHash = txHash
		item.ReconciliationState = "broadcast_accepted_awaiting_inclusion"
		updateLatestAttempt(item, "broadcast_accepted", MutationClassNone, gasLimit, txHash)
		return nil
	})
	if err != nil {
		return err
	}
	return s.confirm(ctx, envelope.ID, txHash)
}

func (s *ProviderMutationSubmitter) buildSignedTx(msg sdk.Msg, accountNumber, sequence, gasLimit uint64) ([]byte, error) {
	builder := s.encCfg.TxConfig.NewTxBuilder()
	if err := builder.SetMsgs(msg); err != nil {
		return nil, err
	}
	builder.SetGasLimit(gasLimit)
	key, err := s.keyManager.GetActiveKey()
	if err != nil {
		return nil, err
	}
	if key.ProviderAddress != s.cfg.ProviderAddress {
		return nil, ErrProviderKeyMismatch
	}
	if key.Algorithm != string(HSMKeyTypeEd25519) {
		return nil, fmt.Errorf("unsupported SDK signer algorithm %s", key.Algorithm)
	}
	priv := &cosmosed25519.PrivKey{Key: append([]byte(nil), key.privateKey...)}
	placeholder := signing.SignatureV2{PubKey: priv.PubKey(), Data: &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT}, Sequence: sequence}
	if err := builder.SetSignatures(placeholder); err != nil {
		return nil, err
	}
	sig, err := clienttx.SignWithPrivKey(context.Background(), signing.SignMode_SIGN_MODE_DIRECT, authsigning.SignerData{ChainID: s.cfg.ChainID, AccountNumber: accountNumber, Sequence: sequence}, builder, priv, s.encCfg.TxConfig, sequence)
	if err != nil {
		return nil, err
	}
	if err := builder.SetSignatures(sig); err != nil {
		return nil, err
	}
	return s.encCfg.TxConfig.TxEncoder()(builder.GetTx())
}

func (s *ProviderMutationSubmitter) confirm(ctx context.Context, id, txHash string) error {
	confirmCtx, cancel := context.WithTimeout(ctx, s.cfg.ConfirmationTimeout)
	defer cancel()
	ticker := time.NewTicker(minDuration(s.cfg.PollInterval, time.Second))
	defer ticker.Stop()
	for {
		confirmation, err := s.chain.ConfirmTx(confirmCtx, txHash)
		if err == nil && confirmation.Found {
			if confirmation.Code != 0 {
				classified := classifyBroadcastError(confirmation.Log)
				classification, retryable, _ := classifyProviderMutationError(classified)
				if retryable {
					return s.schedule(ctx, id, classified)
				}
				return s.deadLetter(ctx, id, classification, "deliver_tx_failed", classified)
			}
			if evidenceErr := validateProviderTxConfirmation(confirmation, txHash); evidenceErr != nil {
				return s.scheduleAmbiguous(ctx, id, evidenceErr)
			}
			includedAt := time.Now().UTC()
			_, updateErr := s.updateOwned(ctx, id, func(item *ProviderMutationEnvelope) error {
				if item.TxHash != "" && !strings.EqualFold(item.TxHash, txHash) {
					return ErrProviderMutationStaleState
				}
				item.State = MutationStateIncluded
				item.TxHash = confirmation.TxHash
				item.ConfirmationHeight = confirmation.Height
				item.ConfirmationBlockHash = confirmation.BlockHash
				item.IncludedAt = includedAt
				item.ReconciliationState = "included_awaiting_finality"
				updateLatestAttempt(item, "included", MutationClassNone, item.GasLimit, confirmation.TxHash)
				return nil
			})
			if updateErr != nil {
				return updateErr
			}
			return s.awaitFinality(ctx, id)
		}
		select {
		case <-confirmCtx.Done():
			return s.scheduleAmbiguous(ctx, id, confirmCtx.Err())
		case <-ticker.C:
		}
	}
}

func (s *ProviderMutationSubmitter) awaitFinality(ctx context.Context, id string) error {
	envelope, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if envelope.TxHash == "" || envelope.ConfirmationHeight <= 0 || envelope.ConfirmationBlockHash == "" {
		return s.scheduleAmbiguous(ctx, id, ErrProviderMutationEvidence)
	}
	target := envelope.ConfirmationHeight + s.cfg.FinalityBlocks
	for {
		actualHash, hashErr := s.chain.BlockHash(ctx, envelope.ConfirmationHeight)
		if hashErr != nil {
			return s.scheduleAmbiguous(ctx, id, hashErr)
		}
		if !strings.EqualFold(actualHash, envelope.ConfirmationBlockHash) {
			_, _ = s.updateOwned(ctx, id, func(item *ProviderMutationEnvelope) error {
				item.State = MutationStateAmbiguous
				item.Classification = MutationClassReplacement
				item.ReconciliationState = "reorg_detected"
				item.ConfirmationHeight = 0
				item.ConfirmationBlockHash = ""
				item.FinalityHeight = 0
				return nil
			})
			return &ProviderMutationError{Op: "finality", MutationID: id, Classification: MutationClassReplacement, Retryable: true, Ambiguous: true, Err: ErrProviderMutationReorg}
		}
		height, heightErr := s.chain.LatestHeight(ctx)
		if heightErr != nil {
			return s.scheduleAmbiguous(ctx, id, heightErr)
		}
		if height >= target {
			now := time.Now().UTC()
			_, updateErr := s.updateOwned(ctx, id, func(item *ProviderMutationEnvelope) error {
				if item.State != MutationStateIncluded || item.ConfirmationHeight != envelope.ConfirmationHeight || !strings.EqualFold(item.ConfirmationBlockHash, envelope.ConfirmationBlockHash) {
					return ErrProviderMutationStaleState
				}
				item.State = MutationStateConfirmed
				item.ConfirmedAt = now
				item.FinalityHeight = height
				item.ReconciliationState = "finalized"
				item.TerminalResult = "confirmed"
				item.ConfirmationLatencyNS = now.Sub(item.CreatedAt).Nanoseconds()
				item.Classification = MutationClassNone
				updateLatestAttempt(item, "confirmed", MutationClassNone, item.GasLimit, item.TxHash)
				return nil
			})
			if updateErr != nil {
				return updateErr
			}
			s.mu.Lock()
			s.lastSuccess = now
			s.mu.Unlock()
			return nil
		}
		timer := time.NewTimer(minDuration(s.cfg.PollInterval, time.Second))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (s *ProviderMutationSubmitter) reconcile(ctx context.Context, envelope *ProviderMutationEnvelope, msg sdk.Msg) (bool, error) {
	if envelope == nil {
		return false, fmt.Errorf("mutation envelope unavailable")
	}
	_, err := s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
		item.ReconciliationCount++
		item.ReconciliationState = "querying_tx_and_logical_state"
		return nil
	})
	if err != nil {
		return false, err
	}
	if envelope.TxHash != "" {
		confirmation, confirmErr := s.chain.ConfirmTx(ctx, envelope.TxHash)
		if confirmErr == nil && confirmation.Found {
			if confirmation.Code != 0 {
				return false, classifyBroadcastError(confirmation.Log)
			}
			if evidenceErr := validateProviderTxConfirmation(confirmation, envelope.TxHash); evidenceErr != nil {
				return false, evidenceErr
			}
			_, err = s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
				item.State = MutationStateIncluded
				item.TxHash = confirmation.TxHash
				item.ConfirmationHeight = confirmation.Height
				item.ConfirmationBlockHash = confirmation.BlockHash
				item.IncludedAt = time.Now().UTC()
				item.ReconciliationState = "tx_hash_committed"
				return nil
			})
			if err != nil {
				return false, err
			}
			return true, s.awaitFinality(ctx, envelope.ID)
		}
	}
	reconciled, err := s.chain.ReconcileMutation(ctx, envelope, msg)
	if err != nil {
		return false, err
	}
	if reconciled.Conflicted {
		return false, s.deadLetter(ctx, envelope.ID, MutationClassReplacement, "logical_idempotency_conflict", ErrProviderMutationConflict)
	}
	if reconciled.Committed {
		if reconciled.TxHash == "" || reconciled.Height <= 0 || reconciled.BlockHash == "" {
			return false, ErrProviderMutationEvidence
		}
		_, err = s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
			item.State = MutationStateIncluded
			item.TxHash = reconciled.TxHash
			item.ConfirmationHeight = reconciled.Height
			item.ConfirmationBlockHash = reconciled.BlockHash
			item.IncludedAt = time.Now().UTC()
			item.ReconciliationState = reconciled.Reason
			return nil
		})
		if err != nil {
			return false, err
		}
		return true, s.awaitFinality(ctx, envelope.ID)
	}
	accountNumber, sequence, err := s.chain.ResolveAccountSequence(ctx, envelope.Signer)
	if err != nil {
		return false, err
	}
	if envelope.Sequence != 0 && sequence > envelope.Sequence {
		return false, &ProviderMutationError{Op: "reconcile", MutationID: envelope.ID, Classification: MutationClassSequenceMismatch, Retryable: true, Ambiguous: true, Err: ErrSequenceMismatch}
	}
	_, err = s.updateOwned(ctx, envelope.ID, func(item *ProviderMutationEnvelope) error {
		item.State = MutationStateRetry
		item.AccountNumber = accountNumber
		item.Sequence = sequence
		item.TxBytes = nil
		item.TxDigest = ""
		item.TxHash = ""
		item.ConfirmationHeight = 0
		item.ConfirmationBlockHash = ""
		item.NextAttemptAt = time.Now().UTC()
		item.ReconciliationState = "not_committed_safe_to_rebuild"
		return nil
	})
	return false, err
}

func (s *ProviderMutationSubmitter) scheduleAmbiguous(ctx context.Context, id string, cause error) error {
	_, err := s.updateOwned(ctx, id, func(item *ProviderMutationEnvelope) error {
		item.State = MutationStateAmbiguous
		item.Classification = MutationClassTimeout
		item.ReconciliationState = "confirmation_ambiguous"
		item.NextAttemptAt = time.Now().UTC().Add(s.retryDelay(item.AttemptCount))
		updateLatestAttempt(item, "ambiguous", MutationClassTimeout, item.GasLimit, item.TxHash)
		return nil
	})
	if err != nil {
		return err
	}
	return &ProviderMutationError{Op: "confirm", MutationID: id, Classification: MutationClassTimeout, Retryable: true, Ambiguous: true, Err: cause}
}

func (s *ProviderMutationSubmitter) schedule(ctx context.Context, id string, cause error) error {
	classification, retryable, ambiguous := classifyProviderMutationError(cause)
	if !retryable {
		return s.deadLetter(ctx, id, classification, "terminal_error", cause)
	}
	_, err := s.updateOwned(ctx, id, func(item *ProviderMutationEnvelope) error {
		if item.AttemptCount >= s.cfg.MaxAttempts {
			item.State = MutationStateDeadLetter
			item.DeadLetterReason = "attempts_exhausted"
			item.TerminalResult = string(queueItemStatusFailed)
		} else if ambiguous {
			item.State = MutationStateAmbiguous
		} else {
			item.State = MutationStateRetry
		}
		item.Classification = classification
		item.NextAttemptAt = time.Now().UTC().Add(s.retryDelay(item.AttemptCount))
		updateLatestAttempt(item, string(item.State), classification, item.GasLimit, item.TxHash)
		return nil
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.lastFailure = time.Now().UTC()
	s.mu.Unlock()
	return &ProviderMutationError{Op: "process", MutationID: id, Classification: classification, Retryable: true, Ambiguous: ambiguous, Err: cause}
}

func (s *ProviderMutationSubmitter) deadLetter(ctx context.Context, id string, classification ProviderMutationClassification, reason string, cause error) error {
	_, err := s.updateOwned(ctx, id, func(item *ProviderMutationEnvelope) error {
		item.State = MutationStateDeadLetter
		item.Classification = classification
		item.DeadLetterReason = reason
		item.TerminalResult = string(queueItemStatusFailed)
		updateLatestAttempt(item, "dead_letter", classification, item.GasLimit, item.TxHash)
		return nil
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.lastFailure = time.Now().UTC()
	s.mu.Unlock()
	return &ProviderMutationError{Op: "dead_letter", MutationID: id, Classification: classification, Err: cause}
}

func (s *ProviderMutationSubmitter) retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := s.cfg.RetryBackoff
	for i := 1; i < attempt && delay < s.cfg.MaxRetryBackoff; i++ {
		delay *= 2
	}
	if delay > s.cfg.MaxRetryBackoff {
		delay = s.cfg.MaxRetryBackoff
	}
	return delay
}

func (s *ProviderMutationSubmitter) Readiness(ctx context.Context) ProviderMutationReadiness {
	s.mu.RLock()
	started := s.running
	token := s.leaseToken
	storeOpen := s.storeReady
	s.mu.RUnlock()
	result := ProviderMutationReadiness{Started: started, KeyReady: s.keyManager != nil && s.keyManager.Ready(), ChainReady: s.cfg.DevelopmentNoop || s.chain != nil, ConfirmationReady: s.cfg.DevelopmentNoop || s.chain != nil}
	result.LeaseHeld = started && s.lease.Held(ctx, s.leaseName, token)
	result.StoreReady = storeOpen
	if storeOpen {
		items, err := s.store.List(ctx)
		result.StoreReady = err == nil
		if err == nil {
			metrics := metricsFromItems(items, time.Now().UTC())
			result.QueueDepth = metrics.QueueDepth
			result.OldestPendingAge = metrics.OldestPendingAge
		}
	}
	result.Ready = result.Started && result.StoreReady && result.LeaseHeld && result.KeyReady && result.ChainReady && result.ConfirmationReady
	switch {
	case !result.Started:
		result.Reason = "submitter not started"
	case !result.StoreReady:
		result.Reason = "queue store unavailable"
	case !result.LeaseHeld:
		result.Reason = "submitter lease not held"
	case !result.KeyReady:
		result.Reason = "key manager unavailable or locked"
	case !result.ChainReady:
		result.Reason = "chain transport unavailable"
	case !result.ConfirmationReady:
		result.Reason = "confirmation transport unavailable"
	}
	return result
}

func (s *ProviderMutationSubmitter) Metrics(ctx context.Context) ProviderMutationMetrics {
	items, err := s.store.List(ctx)
	if err != nil {
		return ProviderMutationMetrics{}
	}
	metrics := metricsFromItems(items, time.Now().UTC())
	s.mu.RLock()
	metrics.LastSuccess = s.lastSuccess
	metrics.LastFailure = s.lastFailure
	s.mu.RUnlock()
	return metrics
}

func metricsFromItems(items []*ProviderMutationEnvelope, now time.Time) ProviderMutationMetrics {
	metrics := ProviderMutationMetrics{}
	for _, item := range items {
		if item == nil {
			continue
		}
		if item.AttemptCount > metrics.MaxAttempts {
			metrics.MaxAttempts = item.AttemptCount
		}
		switch item.State {
		case MutationStateConfirmed:
			if item.ConfirmationLatencyNS > int64(metrics.ConfirmationLatency) {
				metrics.ConfirmationLatency = time.Duration(item.ConfirmationLatencyNS)
			}
		case MutationStateDeadLetter:
			metrics.DeadLetters++
		default:
			metrics.QueueDepth++
			age := now.Sub(item.CreatedAt)
			if age > metrics.OldestPendingAge {
				metrics.OldestPendingAge = age
			}
			if item.State == MutationStateAmbiguous || item.State == MutationStateBroadcasting || item.State == MutationStateIncluded {
				metrics.Ambiguous++
			}
		}
	}
	return metrics
}

func appendBoundedAttempt(attempts []ProviderMutationAttempt, attempt ProviderMutationAttempt) []ProviderMutationAttempt {
	const maxAttemptHistory = 32
	attempts = append(attempts, attempt)
	if len(attempts) > maxAttemptHistory {
		attempts = append([]ProviderMutationAttempt(nil), attempts[len(attempts)-maxAttemptHistory:]...)
	}
	return attempts
}

func updateLatestAttempt(item *ProviderMutationEnvelope, outcome string, classification ProviderMutationClassification, gas uint64, txHash string) {
	if item == nil || len(item.Attempts) == 0 {
		return
	}
	idx := len(item.Attempts) - 1
	item.Attempts[idx].Outcome = outcome
	item.Attempts[idx].Classification = classification
	item.Attempts[idx].GasWanted = gas
	item.Attempts[idx].TxHash = txHash
	if outcome == "confirmed" || outcome == "dead_letter" {
		item.Attempts[idx].FinishedAt = time.Now().UTC()
	}
}

func classifyProviderMutationError(err error) (ProviderMutationClassification, bool, bool) {
	if err == nil {
		return MutationClassNone, false, false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return MutationClassTimeout, true, true
	}
	if errors.Is(err, context.Canceled) {
		return MutationClassTimeout, true, true
	}
	if errors.Is(err, ErrSequenceMismatch) {
		return MutationClassSequenceMismatch, true, false
	}
	if errors.Is(err, ErrProviderMutationNotReady) || errors.Is(err, ErrProviderMutationUnavailable) {
		return MutationClassUnavailable, true, true
	}
	if errors.Is(err, ErrProviderMutationEvidence) || errors.Is(err, ErrProviderMutationStaleState) {
		return MutationClassReplacement, true, true
	}
	var classified *classifiedBroadcastError
	if errors.As(err, &classified) {
		classification, _, ambiguous := classifyProviderMutationError(errors.New(classified.Message))
		return classification, classified.Retryable, ambiguous
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "sequence mismatch") || strings.Contains(message, "account sequence"):
		return MutationClassSequenceMismatch, true, false
	case strings.Contains(message, "out of gas"):
		return MutationClassOutOfGas, true, false
	case strings.Contains(message, "mempool") || strings.Contains(message, "rate limit"):
		return MutationClassMempoolReject, true, false
	case strings.Contains(message, "timeout") || strings.Contains(message, "deadline"):
		return MutationClassTimeout, true, true
	case strings.Contains(message, "connection") || strings.Contains(message, "unavailable") || strings.Contains(message, "eof"):
		return MutationClassUnavailable, true, true
	case strings.Contains(message, "insufficient funds") || strings.Contains(message, "unauthorized") || strings.Contains(message, "signature"):
		return MutationClassUnauthorized, false, false
	case strings.Contains(message, "invalid"):
		return MutationClassInvalid, false, false
	default:
		return MutationClassUnknown, true, true
	}
}

func isProviderMutationRetryable(err error) bool {
	var mutationErr *ProviderMutationError
	return errors.As(err, &mutationErr) && mutationErr.Retryable
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func checkedAdjustedGasLimit(base uint64, adjustment float64) (uint64, error) {
	if base == 0 {
		return 0, fmt.Errorf("gas limit must be positive")
	}
	if math.IsNaN(adjustment) || math.IsInf(adjustment, 0) || adjustment < 1 {
		return 0, fmt.Errorf("gas adjustment must be finite and >= 1")
	}
	rat, ok := new(big.Rat).SetString(strconv.FormatFloat(adjustment, 'f', -1, 64))
	if !ok || rat.Sign() <= 0 {
		return 0, fmt.Errorf("gas adjustment is not representable")
	}
	numerator := new(big.Int).SetUint64(base)
	numerator.Mul(numerator, rat.Num())
	quotient, remainder := new(big.Int).QuoRem(numerator, rat.Denom(), new(big.Int))
	if remainder.Sign() != 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if !quotient.IsUint64() {
		return 0, fmt.Errorf("gas adjustment overflows uint64")
	}
	adjusted := quotient.Uint64()
	if adjusted < base {
		return 0, fmt.Errorf("gas adjustment overflowed")
	}
	return adjusted, nil
}

func validateProviderTxConfirmation(confirmation ProviderTxConfirmation, expectedTxHash string) error {
	switch {
	case strings.TrimSpace(confirmation.TxHash) == "":
		return fmt.Errorf("%w: missing tx hash", ErrProviderMutationEvidence)
	case expectedTxHash != "" && !strings.EqualFold(confirmation.TxHash, expectedTxHash):
		return fmt.Errorf("%w: tx hash mismatch", ErrProviderMutationEvidence)
	case confirmation.Height <= 0:
		return fmt.Errorf("%w: missing block height", ErrProviderMutationEvidence)
	case strings.TrimSpace(confirmation.BlockHash) == "":
		return fmt.Errorf("%w: missing block hash", ErrProviderMutationEvidence)
	default:
		return nil
	}
}

// MarshalJSON is deliberately standard JSON; the persisted message bytes and
// digests, rather than JSON map ordering, are the canonical transaction input.
var _ = json.Marshaler(nil)
