// Package observability provides structured logging, metrics, and tracing
// for the VirtEngine platform.
//
// MONITOR-001: Comprehensive Prometheus metrics integration
package observability

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Global registry for all VirtEngine metrics
	globalRegistry     *prometheus.Registry
	globalRegistryOnce sync.Once
)

// GetRegistry returns the global Prometheus registry
func GetRegistry() *prometheus.Registry {
	globalRegistryOnce.Do(func() {
		globalRegistry = prometheus.NewRegistry()
		// Register standard Go collectors
		globalRegistry.MustRegister(collectors.NewGoCollector())
		globalRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	})
	return globalRegistry
}

// MetricsHandler returns an HTTP handler for Prometheus metrics
func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(GetRegistry(), promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// ============================================================================
// Chain Health Metrics
// ============================================================================

// ChainMetrics contains metrics for blockchain health monitoring
type ChainMetrics struct {
	// Block metrics
	BlockHeight     prometheus.Gauge
	BlockTime       prometheus.Gauge
	BlockTxCount    prometheus.Histogram
	BlockSize       prometheus.Histogram
	BlockLatency    prometheus.Histogram
	MissedBlocks    prometheus.Counter
	BlocksProduced  prometheus.Counter

	// Consensus metrics
	ConsensusRounds         prometheus.Histogram
	ConsensusLatency        prometheus.Histogram
	ValidatorVotingPower    *prometheus.GaugeVec
	ValidatorMissedBlocks   *prometheus.CounterVec
	ValidatorUptime         *prometheus.GaugeVec
	ActiveValidators        prometheus.Gauge
	TotalValidators         prometheus.Gauge

	// Transaction metrics
	TxPoolSize              prometheus.Gauge
	TxProcessed             *prometheus.CounterVec
	TxLatency               *prometheus.HistogramVec
	TxGasUsed               prometheus.Histogram
	TxFees                  prometheus.Histogram
	TxPerSecond             prometheus.Gauge

	// State metrics
	StateSize               prometheus.Gauge
	StateSyncHeight         prometheus.Gauge
	StateSyncPeers          prometheus.Gauge

	// P2P metrics
	PeerCount               prometheus.Gauge
	InboundPeers            prometheus.Gauge
	OutboundPeers           prometheus.Gauge
	PeerLatency             *prometheus.HistogramVec
	MessagesReceived        *prometheus.CounterVec
	MessagesSent            *prometheus.CounterVec
	BytesReceived           prometheus.Counter
	BytesSent               prometheus.Counter
}

// NewChainMetrics creates and registers chain health metrics
func NewChainMetrics(namespace string) *ChainMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	m := &ChainMetrics{
		// Block metrics
		BlockHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "block_height",
			Help:      "Current block height",
		}),
		BlockTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "block_time_seconds",
			Help:      "Time of the last block in unix seconds",
		}),
		BlockTxCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "block_tx_count",
			Help:      "Number of transactions per block",
			Buckets:   []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000},
		}),
		BlockSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "block_size_bytes",
			Help:      "Size of blocks in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 15), // 1KB to 16MB
		}),
		BlockLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "block_latency_seconds",
			Help:      "Time between blocks in seconds",
			Buckets:   []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 30},
		}),
		MissedBlocks: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "missed_blocks_total",
			Help:      "Total number of missed blocks",
		}),
		BlocksProduced: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "chain",
			Name:      "blocks_produced_total",
			Help:      "Total number of blocks produced",
		}),

		// Consensus metrics
		ConsensusRounds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "consensus",
			Name:      "rounds",
			Help:      "Number of consensus rounds per block",
			Buckets:   []float64{1, 2, 3, 4, 5, 10, 15, 20},
		}),
		ConsensusLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "consensus",
			Name:      "latency_seconds",
			Help:      "Time to reach consensus in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 3, 4, 5, 7, 10, 15, 20},
		}),
		ValidatorVotingPower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "validator",
			Name:      "voting_power",
			Help:      "Voting power of validators",
		}, []string{"validator"}),
		ValidatorMissedBlocks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "validator",
			Name:      "missed_blocks_total",
			Help:      "Total missed blocks per validator",
		}, []string{"validator"}),
		ValidatorUptime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "validator",
			Name:      "uptime_percentage",
			Help:      "Validator uptime percentage",
		}, []string{"validator"}),
		ActiveValidators: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "validator",
			Name:      "active_count",
			Help:      "Number of active validators",
		}),
		TotalValidators: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "validator",
			Name:      "total_count",
			Help:      "Total number of validators",
		}),

		// Transaction metrics
		TxPoolSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "tx",
			Name:      "pool_size",
			Help:      "Number of transactions in the mempool",
		}),
		TxProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "tx",
			Name:      "processed_total",
			Help:      "Total processed transactions",
		}, []string{"status", "type"}),
		TxLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "tx",
			Name:      "latency_seconds",
			Help:      "Transaction processing latency",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		}, []string{"type"}),
		TxGasUsed: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "tx",
			Name:      "gas_used",
			Help:      "Gas used per transaction",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 15),
		}),
		TxFees: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "tx",
			Name:      "fees",
			Help:      "Transaction fees",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 20),
		}),
		TxPerSecond: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "tx",
			Name:      "per_second",
			Help:      "Transactions per second",
		}),

		// State metrics
		StateSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "state",
			Name:      "size_bytes",
			Help:      "Size of the state database in bytes",
		}),
		StateSyncHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "state",
			Name:      "sync_height",
			Help:      "State sync target height",
		}),
		StateSyncPeers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "state",
			Name:      "sync_peers",
			Help:      "Number of state sync peers",
		}),

		// P2P metrics
		PeerCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "peer_count",
			Help:      "Total number of connected peers",
		}),
		InboundPeers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "inbound_peers",
			Help:      "Number of inbound peer connections",
		}),
		OutboundPeers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "outbound_peers",
			Help:      "Number of outbound peer connections",
		}),
		PeerLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "peer_latency_seconds",
			Help:      "Latency to peers in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		}, []string{"peer_id"}),
		MessagesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "messages_received_total",
			Help:      "Total messages received",
		}, []string{"type"}),
		MessagesSent: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "messages_sent_total",
			Help:      "Total messages sent",
		}, []string{"type"}),
		BytesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "bytes_received_total",
			Help:      "Total bytes received",
		}),
		BytesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "p2p",
			Name:      "bytes_sent_total",
			Help:      "Total bytes sent",
		}),
	}

	// Register all metrics
	reg := GetRegistry()
	reg.MustRegister(
		m.BlockHeight, m.BlockTime, m.BlockTxCount, m.BlockSize,
		m.BlockLatency, m.MissedBlocks, m.BlocksProduced,
		m.ConsensusRounds, m.ConsensusLatency,
		m.ValidatorVotingPower, m.ValidatorMissedBlocks, m.ValidatorUptime,
		m.ActiveValidators, m.TotalValidators,
		m.TxPoolSize, m.TxProcessed, m.TxLatency, m.TxGasUsed, m.TxFees, m.TxPerSecond,
		m.StateSize, m.StateSyncHeight, m.StateSyncPeers,
		m.PeerCount, m.InboundPeers, m.OutboundPeers,
		m.PeerLatency, m.MessagesReceived, m.MessagesSent,
		m.BytesReceived, m.BytesSent,
	)

	return m
}

// ============================================================================
// VEID Module Metrics
// ============================================================================

// VEIDMetrics contains metrics for VEID identity verification
type VEIDMetrics struct {
	// Verification metrics
	VerificationsTotal      *prometheus.CounterVec
	VerificationLatency     *prometheus.HistogramVec
	VerificationScores      prometheus.Histogram
	VerificationMismatches  prometheus.Counter
	VerificationTimeouts    prometheus.Counter

	// Scope metrics
	ScopesUploaded          *prometheus.CounterVec
	ScopeValidationErrors   *prometheus.CounterVec
	ScopeSize               prometheus.Histogram
	ActiveScopes            prometheus.Gauge

	// ML inference metrics
	InferenceLatency        *prometheus.HistogramVec
	InferenceErrors         *prometheus.CounterVec
	InferenceDeterminism    prometheus.Counter
	ModelVersion            *prometheus.GaugeVec

	// Identity metrics
	IdentitiesVerified      prometheus.Counter
	IdentitiesActive        prometheus.Gauge
	IdentityScoreDecay      prometheus.Histogram
	BorderlineCases         prometheus.Counter

	// Rate limiting metrics
	RateLimitHits           *prometheus.CounterVec
	SMSVerifications        *prometheus.CounterVec
	EmailVerifications      *prometheus.CounterVec
}

// NewVEIDMetrics creates and registers VEID metrics
func NewVEIDMetrics(namespace string) *VEIDMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	m := &VEIDMetrics{
		VerificationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "verifications_total",
			Help:      "Total VEID verifications",
		}, []string{"status", "type"}),
		VerificationLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "verification_latency_seconds",
			Help:      "VEID verification latency",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
		}, []string{"type"}),
		VerificationScores: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "verification_scores",
			Help:      "Distribution of verification scores",
			Buckets:   []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		}),
		VerificationMismatches: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "verification_mismatches_total",
			Help:      "Total verification result mismatches between validators",
		}),
		VerificationTimeouts: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "verification_timeouts_total",
			Help:      "Total verification timeouts",
		}),

		ScopesUploaded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "scopes_uploaded_total",
			Help:      "Total scopes uploaded",
		}, []string{"type"}),
		ScopeValidationErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "scope_validation_errors_total",
			Help:      "Total scope validation errors",
		}, []string{"error_type"}),
		ScopeSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "scope_size_bytes",
			Help:      "Size of uploaded scopes in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 15),
		}),
		ActiveScopes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "active_scopes",
			Help:      "Number of active verification scopes",
		}),

		InferenceLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "inference_latency_seconds",
			Help:      "ML inference latency",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		}, []string{"model"}),
		InferenceErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "inference_errors_total",
			Help:      "Total ML inference errors",
		}, []string{"error_type", "model"}),
		InferenceDeterminism: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "inference_non_deterministic_total",
			Help:      "Total non-deterministic inference results (CRITICAL)",
		}),
		ModelVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "model_version",
			Help:      "Current ML model version (1 = active)",
		}, []string{"model", "version"}),

		IdentitiesVerified: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "identities_verified_total",
			Help:      "Total identities verified",
		}),
		IdentitiesActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "identities_active",
			Help:      "Number of active verified identities",
		}),
		IdentityScoreDecay: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "identity_score_decay",
			Help:      "Score decay per identity",
			Buckets:   []float64{1, 2, 5, 10, 15, 20, 25, 30},
		}),
		BorderlineCases: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "borderline_cases_total",
			Help:      "Total borderline verification cases requiring review",
		}),

		RateLimitHits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "rate_limit_hits_total",
			Help:      "Total rate limit hits",
		}, []string{"type"}),
		SMSVerifications: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "sms_verifications_total",
			Help:      "Total SMS verifications",
		}, []string{"status"}),
		EmailVerifications: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "veid",
			Name:      "email_verifications_total",
			Help:      "Total email verifications",
		}, []string{"status"}),
	}

	reg := GetRegistry()
	reg.MustRegister(
		m.VerificationsTotal, m.VerificationLatency, m.VerificationScores,
		m.VerificationMismatches, m.VerificationTimeouts,
		m.ScopesUploaded, m.ScopeValidationErrors, m.ScopeSize, m.ActiveScopes,
		m.InferenceLatency, m.InferenceErrors, m.InferenceDeterminism, m.ModelVersion,
		m.IdentitiesVerified, m.IdentitiesActive, m.IdentityScoreDecay, m.BorderlineCases,
		m.RateLimitHits, m.SMSVerifications, m.EmailVerifications,
	)

	return m
}

// ============================================================================
// Marketplace Metrics
// ============================================================================

// MarketplaceMetrics contains metrics for marketplace operations
type MarketplaceMetrics struct {
	// Order metrics
	OrdersTotal           *prometheus.CounterVec
	OrdersActive          prometheus.Gauge
	OrderLatency          *prometheus.HistogramVec
	OrderValue            prometheus.Histogram

	// Bid metrics
	BidsTotal             *prometheus.CounterVec
	BidsActive            prometheus.Gauge
	BidLatency            prometheus.Histogram
	BidValue              prometheus.Histogram
	BidSpread             prometheus.Histogram

	// Lease metrics
	LeasesTotal           *prometheus.CounterVec
	LeasesActive          prometheus.Gauge
	LeaseValue            prometheus.Histogram
	LeaseDuration         prometheus.Histogram

	// Provider metrics
	ProvidersActive       prometheus.Gauge
	ProvidersTotal        prometheus.Gauge
	ProviderCapacity      *prometheus.GaugeVec
	ProviderUtilization   *prometheus.GaugeVec
	ProviderReliability   *prometheus.GaugeVec

	// Market health metrics
	MarketEfficiency      prometheus.Gauge
	MarketVolume24h       prometheus.Gauge
	MarketFillRate        prometheus.Gauge
	MarketSpreadAvg       prometheus.Gauge

	// Escrow metrics
	EscrowTotal           prometheus.Gauge
	EscrowLocked          prometheus.Gauge
	EscrowReleased        prometheus.Counter
	EscrowDisputed        prometheus.Counter
}

// NewMarketplaceMetrics creates and registers marketplace metrics
func NewMarketplaceMetrics(namespace string) *MarketplaceMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	m := &MarketplaceMetrics{
		OrdersTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "orders_total",
			Help:      "Total orders",
		}, []string{"status", "type"}),
		OrdersActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "orders_active",
			Help:      "Number of active orders",
		}),
		OrderLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "order_latency_seconds",
			Help:      "Order processing latency",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 300},
		}, []string{"type"}),
		OrderValue: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "order_value",
			Help:      "Order values",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 20),
		}),

		BidsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "bids_total",
			Help:      "Total bids",
		}, []string{"status"}),
		BidsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "bids_active",
			Help:      "Number of active bids",
		}),
		BidLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "bid_latency_seconds",
			Help:      "Bid placement latency",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
		}),
		BidValue: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "bid_value",
			Help:      "Bid values",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 20),
		}),
		BidSpread: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "bid_spread_bps",
			Help:      "Bid-ask spread in basis points",
			Buckets:   []float64{10, 25, 50, 100, 200, 500, 1000},
		}),

		LeasesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "leases_total",
			Help:      "Total leases",
		}, []string{"status"}),
		LeasesActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "leases_active",
			Help:      "Number of active leases",
		}),
		LeaseValue: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "lease_value",
			Help:      "Lease values",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 20),
		}),
		LeaseDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "lease_duration_hours",
			Help:      "Lease durations in hours",
			Buckets:   []float64{1, 6, 12, 24, 48, 72, 168, 336, 720},
		}),

		ProvidersActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "providers_active",
			Help:      "Number of active providers",
		}),
		ProvidersTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "providers_total",
			Help:      "Total number of providers",
		}),
		ProviderCapacity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "provider_capacity",
			Help:      "Provider capacity by resource type",
		}, []string{"provider", "resource_type"}),
		ProviderUtilization: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "provider_utilization_percent",
			Help:      "Provider utilization percentage",
		}, []string{"provider", "resource_type"}),
		ProviderReliability: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "provider_reliability_score",
			Help:      "Provider reliability score (0-10000)",
		}, []string{"provider"}),

		MarketEfficiency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "efficiency_score",
			Help:      "Overall market efficiency score (0-100)",
		}),
		MarketVolume24h: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "volume_24h",
			Help:      "24-hour trading volume",
		}),
		MarketFillRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "fill_rate_percent",
			Help:      "Order fill rate percentage",
		}),
		MarketSpreadAvg: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "market",
			Name:      "spread_avg_bps",
			Help:      "Average bid-ask spread in basis points",
		}),

		EscrowTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "escrow",
			Name:      "total",
			Help:      "Total escrow value",
		}),
		EscrowLocked: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "escrow",
			Name:      "locked",
			Help:      "Locked escrow value",
		}),
		EscrowReleased: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "escrow",
			Name:      "released_total",
			Help:      "Total released escrow value",
		}),
		EscrowDisputed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "escrow",
			Name:      "disputed_total",
			Help:      "Total disputed escrow events",
		}),
	}

	reg := GetRegistry()
	reg.MustRegister(
		m.OrdersTotal, m.OrdersActive, m.OrderLatency, m.OrderValue,
		m.BidsTotal, m.BidsActive, m.BidLatency, m.BidValue, m.BidSpread,
		m.LeasesTotal, m.LeasesActive, m.LeaseValue, m.LeaseDuration,
		m.ProvidersActive, m.ProvidersTotal,
		m.ProviderCapacity, m.ProviderUtilization, m.ProviderReliability,
		m.MarketEfficiency, m.MarketVolume24h, m.MarketFillRate, m.MarketSpreadAvg,
		m.EscrowTotal, m.EscrowLocked, m.EscrowReleased, m.EscrowDisputed,
	)

	return m
}

// ============================================================================
// Provider Daemon Metrics
// ============================================================================

// ProviderDaemonMetrics contains metrics for provider daemon operations
type ProviderDaemonMetrics struct {
	// Bid engine metrics
	BidEngineOrders         prometheus.Gauge
	BidEngineDecisions      *prometheus.CounterVec
	BidEngineLatency        prometheus.Histogram

	// Deployment metrics
	DeploymentsTotal        *prometheus.CounterVec
	DeploymentsActive       prometheus.Gauge
	DeploymentLatency       *prometheus.HistogramVec
	DeploymentState         *prometheus.GaugeVec

	// Usage metering metrics
	UsageSubmissions        *prometheus.CounterVec
	UsageLatency            prometheus.Histogram
	UsageCPU                *prometheus.GaugeVec
	UsageMemory             *prometheus.GaugeVec
	UsageStorage            *prometheus.GaugeVec
	UsageGPU                *prometheus.GaugeVec
	UsageNetwork            *prometheus.GaugeVec

	// Infrastructure adapter metrics
	AdapterOperations       *prometheus.CounterVec
	AdapterLatency          *prometheus.HistogramVec
	AdapterErrors           *prometheus.CounterVec

	// Key management metrics
	KeyOperations           *prometheus.CounterVec
	KeyRotations            prometheus.Counter
}

// NewProviderDaemonMetrics creates and registers provider daemon metrics
func NewProviderDaemonMetrics(namespace string) *ProviderDaemonMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	m := &ProviderDaemonMetrics{
		BidEngineOrders: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "bid_engine_orders",
			Help:      "Number of orders in bid engine queue",
		}),
		BidEngineDecisions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "bid_engine_decisions_total",
			Help:      "Bid engine decisions",
		}, []string{"decision"}),
		BidEngineLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "bid_engine_latency_seconds",
			Help:      "Bid engine decision latency",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		}),

		DeploymentsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "deployments_total",
			Help:      "Total deployments",
		}, []string{"status", "adapter"}),
		DeploymentsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "deployments_active",
			Help:      "Number of active deployments",
		}),
		DeploymentLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "deployment_latency_seconds",
			Help:      "Deployment latency",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
		}, []string{"adapter", "operation"}),
		DeploymentState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "deployment_state",
			Help:      "Deployment state (1=active for that state)",
		}, []string{"deployment_id", "state"}),

		UsageSubmissions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_submissions_total",
			Help:      "Total usage record submissions",
		}, []string{"status"}),
		UsageLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_submission_latency_seconds",
			Help:      "Usage submission latency",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		}),
		UsageCPU: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_cpu_millicores",
			Help:      "CPU usage in millicores",
		}, []string{"deployment_id"}),
		UsageMemory: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_memory_bytes",
			Help:      "Memory usage in bytes",
		}, []string{"deployment_id"}),
		UsageStorage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_storage_bytes",
			Help:      "Storage usage in bytes",
		}, []string{"deployment_id"}),
		UsageGPU: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_gpu_count",
			Help:      "GPU usage count",
		}, []string{"deployment_id", "gpu_type"}),
		UsageNetwork: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "usage_network_bytes",
			Help:      "Network usage in bytes",
		}, []string{"deployment_id", "direction"}),

		AdapterOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "adapter_operations_total",
			Help:      "Infrastructure adapter operations",
		}, []string{"adapter", "operation", "status"}),
		AdapterLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "adapter_latency_seconds",
			Help:      "Infrastructure adapter latency",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
		}, []string{"adapter", "operation"}),
		AdapterErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "adapter_errors_total",
			Help:      "Infrastructure adapter errors",
		}, []string{"adapter", "error_type"}),

		KeyOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "key_operations_total",
			Help:      "Key management operations",
		}, []string{"operation", "status"}),
		KeyRotations: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "provider",
			Name:      "key_rotations_total",
			Help:      "Total key rotations",
		}),
	}

	reg := GetRegistry()
	reg.MustRegister(
		m.BidEngineOrders, m.BidEngineDecisions, m.BidEngineLatency,
		m.DeploymentsTotal, m.DeploymentsActive, m.DeploymentLatency, m.DeploymentState,
		m.UsageSubmissions, m.UsageLatency,
		m.UsageCPU, m.UsageMemory, m.UsageStorage, m.UsageGPU, m.UsageNetwork,
		m.AdapterOperations, m.AdapterLatency, m.AdapterErrors,
		m.KeyOperations, m.KeyRotations,
	)

	return m
}

// ============================================================================
// API Metrics
// ============================================================================

// APIMetrics contains metrics for gRPC and HTTP API endpoints
type APIMetrics struct {
	// Request metrics
	RequestsTotal         *prometheus.CounterVec
	RequestDuration       *prometheus.HistogramVec
	RequestSize           *prometheus.HistogramVec
	ResponseSize          *prometheus.HistogramVec

	// Error metrics
	ErrorsTotal           *prometheus.CounterVec

	// Connection metrics
	ConnectionsTotal      *prometheus.CounterVec
	ConnectionsActive     *prometheus.GaugeVec

	// Rate limiting metrics
	RateLimitHits         *prometheus.CounterVec
	RateLimitRemaining    *prometheus.GaugeVec
}

// NewAPIMetrics creates and registers API metrics
func NewAPIMetrics(namespace string) *APIMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	m := &APIMetrics{
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "requests_total",
			Help:      "Total API requests",
		}, []string{"method", "endpoint", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "request_duration_seconds",
			Help:      "API request duration",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "endpoint"}),
		RequestSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "request_size_bytes",
			Help:      "API request size",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 15),
		}, []string{"method", "endpoint"}),
		ResponseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "response_size_bytes",
			Help:      "API response size",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 15),
		}, []string{"method", "endpoint"}),

		ErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "errors_total",
			Help:      "Total API errors",
		}, []string{"method", "endpoint", "error_type"}),

		ConnectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "connections_total",
			Help:      "Total API connections",
		}, []string{"protocol"}),
		ConnectionsActive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "connections_active",
			Help:      "Active API connections",
		}, []string{"protocol"}),

		RateLimitHits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "rate_limit_hits_total",
			Help:      "Rate limit hits",
		}, []string{"endpoint", "tier"}),
		RateLimitRemaining: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "rate_limit_remaining",
			Help:      "Remaining rate limit quota",
		}, []string{"endpoint", "tier"}),
	}

	reg := GetRegistry()
	reg.MustRegister(
		m.RequestsTotal, m.RequestDuration, m.RequestSize, m.ResponseSize,
		m.ErrorsTotal,
		m.ConnectionsTotal, m.ConnectionsActive,
		m.RateLimitHits, m.RateLimitRemaining,
	)

	return m
}

// ============================================================================
// SLO Metrics
// ============================================================================

// SLOMetrics contains metrics for SLO tracking
type SLOMetrics struct {
	// Availability SLOs
	AvailabilitySLO       *prometheus.GaugeVec
	AvailabilityBudget    *prometheus.GaugeVec
	AvailabilityBurnRate  *prometheus.GaugeVec

	// Latency SLOs
	LatencySLO            *prometheus.GaugeVec
	LatencyBudget         *prometheus.GaugeVec
	LatencyBurnRate       *prometheus.GaugeVec

	// Error rate SLOs
	ErrorRateSLO          *prometheus.GaugeVec
	ErrorRateBudget       *prometheus.GaugeVec
	ErrorRateBurnRate     *prometheus.GaugeVec

	// SLO violations
	SLOViolations         *prometheus.CounterVec
}

// NewSLOMetrics creates and registers SLO metrics
func NewSLOMetrics(namespace string) *SLOMetrics {
	if namespace == "" {
		namespace = "virtengine"
	}

	m := &SLOMetrics{
		AvailabilitySLO: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "availability_current",
			Help:      "Current availability percentage",
		}, []string{"service", "slo_id"}),
		AvailabilityBudget: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "availability_budget_remaining",
			Help:      "Remaining availability error budget (0-1)",
		}, []string{"service", "slo_id"}),
		AvailabilityBurnRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "availability_burn_rate",
			Help:      "Availability error budget burn rate",
		}, []string{"service", "slo_id"}),

		LatencySLO: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "latency_current_seconds",
			Help:      "Current latency at SLO percentile",
		}, []string{"service", "slo_id", "percentile"}),
		LatencyBudget: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "latency_budget_remaining",
			Help:      "Remaining latency error budget (0-1)",
		}, []string{"service", "slo_id"}),
		LatencyBurnRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "latency_burn_rate",
			Help:      "Latency error budget burn rate",
		}, []string{"service", "slo_id"}),

		ErrorRateSLO: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "error_rate_current",
			Help:      "Current error rate",
		}, []string{"service", "slo_id"}),
		ErrorRateBudget: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "error_rate_budget_remaining",
			Help:      "Remaining error rate budget (0-1)",
		}, []string{"service", "slo_id"}),
		ErrorRateBurnRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "error_rate_burn_rate",
			Help:      "Error rate budget burn rate",
		}, []string{"service", "slo_id"}),

		SLOViolations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "slo",
			Name:      "violations_total",
			Help:      "Total SLO violations",
		}, []string{"service", "slo_id", "slo_type"}),
	}

	reg := GetRegistry()
	reg.MustRegister(
		m.AvailabilitySLO, m.AvailabilityBudget, m.AvailabilityBurnRate,
		m.LatencySLO, m.LatencyBudget, m.LatencyBurnRate,
		m.ErrorRateSLO, m.ErrorRateBudget, m.ErrorRateBurnRate,
		m.SLOViolations,
	)

	return m
}

// ============================================================================
// Timer Helpers
// ============================================================================

// Timer is a helper for measuring operation duration
type Timer struct {
	start    time.Time
	observer prometheus.Observer
}

// NewTimer creates a new timer
func NewTimer(observer prometheus.Observer) *Timer {
	return &Timer{
		start:    time.Now(),
		observer: observer,
	}
}

// ObserveDuration records the duration since timer creation
func (t *Timer) ObserveDuration() {
	t.observer.Observe(time.Since(t.start).Seconds())
}

// ObserveDurationVec records duration to a histogram vec
func (t *Timer) ObserveDurationVec(vec *prometheus.HistogramVec, labels ...string) {
	vec.WithLabelValues(labels...).Observe(time.Since(t.start).Seconds())
}
