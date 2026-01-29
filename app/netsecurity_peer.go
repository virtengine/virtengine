package app

import (
	"fmt"
	"net"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// PeerID represents a unique peer identifier (typically a hex-encoded public key).
type PeerID string

// PeerInfo contains information about a connected peer.
type PeerInfo struct {
	ID           PeerID
	Address      net.Addr
	PublicKey    []byte
	IsInbound    bool
	ConnectedAt  time.Time
	LastSeen     time.Time
	Stake        int64
	IsValidator  bool
	ASN          uint32
	Subnet       string // /24 subnet
}

// PeerScore represents the reputation score for a peer.
type PeerScore struct {
	// Base score components
	UptimeScore       float64 // Score based on connection uptime
	ResponseScore     float64 // Score based on message response times
	BehaviorScore     float64 // Score based on protocol compliance
	StakeScore        float64 // Score based on staked amount
	ValidatorBonus    float64 // Bonus for being a validator
	
	// Penalty components
	DisconnectionPenalty float64 // Penalty for frequent disconnections
	MisbehaviorPenalty   float64 // Penalty for protocol violations
	RateLimitPenalty     float64 // Penalty for hitting rate limits
	
	// Computed total
	Total         float64
	LastUpdated   time.Time
}

// PeerScoreParams defines the parameters for peer scoring.
type PeerScoreParams struct {
	// Score weights
	UptimeWeight      float64
	ResponseWeight    float64
	BehaviorWeight    float64
	StakeWeight       float64
	ValidatorBonusMax float64

	// Thresholds
	MinScore         float64 // Minimum score before disconnection
	WarningScore     float64 // Score threshold for warnings
	
	// Decay rates
	ScoreDecayRate   float64 // How fast scores decay over time
	PenaltyDecayRate float64 // How fast penalties decay
}

// DefaultPeerScoreParams returns default scoring parameters.
func DefaultPeerScoreParams() PeerScoreParams {
	return PeerScoreParams{
		UptimeWeight:      1.0,
		ResponseWeight:    2.0,
		BehaviorWeight:    3.0,
		StakeWeight:       1.5,
		ValidatorBonusMax: 20.0,
		MinScore:          -100.0,
		WarningScore:      -50.0,
		ScoreDecayRate:    0.01,
		PenaltyDecayRate:  0.005,
	}
}

// PeerAuthenticator handles peer authentication.
type PeerAuthenticator struct {
	config     PeerConfig
	logger     log.Logger
	trustedSet map[PeerID]bool
	bannedSet  map[PeerID]time.Time // PeerID -> ban expiry time
	
	mu sync.RWMutex
}

// NewPeerAuthenticator creates a new peer authenticator.
func NewPeerAuthenticator(config PeerConfig, logger log.Logger) *PeerAuthenticator {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	trustedSet := make(map[PeerID]bool)
	for _, p := range config.TrustedPeers {
		trustedSet[PeerID(p)] = true
	}

	bannedSet := make(map[PeerID]time.Time)
	for _, p := range config.BannedPeers {
		// Permanently banned peers get far-future expiry
		bannedSet[PeerID(p)] = time.Now().Add(100 * 365 * 24 * time.Hour)
	}

	return &PeerAuthenticator{
		config:     config,
		logger:     logger.With("module", "peer-auth"),
		trustedSet: trustedSet,
		bannedSet:  bannedSet,
	}
}

// AuthenticatePeer verifies if a peer is allowed to connect.
func (pa *PeerAuthenticator) AuthenticatePeer(info PeerInfo) (bool, string) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	// Check if banned
	if expiry, banned := pa.bannedSet[info.ID]; banned {
		if time.Now().Before(expiry) {
			return false, "peer is banned"
		}
	}

	// Check if trusted (always allowed)
	if pa.trustedSet[info.ID] {
		pa.logger.Debug("trusted peer authenticated", "peer", info.ID)
		return true, ""
	}

	// Validate public key if provided
	if len(info.PublicKey) > 0 && len(info.PublicKey) != 32 {
		return false, "invalid public key length"
	}

	return true, ""
}

// BanPeer adds a peer to the ban list.
func (pa *PeerAuthenticator) BanPeer(id PeerID, duration time.Duration) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.bannedSet[id] = time.Now().Add(duration)
	pa.logger.Warn("peer banned", "peer", id, "duration", duration)
}

// UnbanPeer removes a peer from the ban list.
func (pa *PeerAuthenticator) UnbanPeer(id PeerID) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	delete(pa.bannedSet, id)
	pa.logger.Info("peer unbanned", "peer", id)
}

// AddTrustedPeer adds a peer to the trusted list.
func (pa *PeerAuthenticator) AddTrustedPeer(id PeerID) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.trustedSet[id] = true
	pa.logger.Info("peer trusted", "peer", id)
}

// RemoveTrustedPeer removes a peer from the trusted list.
func (pa *PeerAuthenticator) RemoveTrustedPeer(id PeerID) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	delete(pa.trustedSet, id)
	pa.logger.Info("peer trust revoked", "peer", id)
}

// IsBanned checks if a peer is currently banned.
func (pa *PeerAuthenticator) IsBanned(id PeerID) bool {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if expiry, banned := pa.bannedSet[id]; banned {
		return time.Now().Before(expiry)
	}
	return false
}

// PeerAuthorizer handles peer authorization decisions.
type PeerAuthorizer struct {
	config       PeerConfig
	logger       log.Logger
	scoreManager *PeerScoreManager
	
	// Connection tracking
	inboundCount  int
	outboundCount int
	subnetCounts  map[string]int // /24 subnet -> peer count
	asnCounts     map[uint32]int // ASN -> peer count
	
	mu sync.RWMutex
}

// NewPeerAuthorizer creates a new peer authorizer.
func NewPeerAuthorizer(config PeerConfig, scoreManager *PeerScoreManager, logger log.Logger) *PeerAuthorizer {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &PeerAuthorizer{
		config:       config,
		logger:       logger.With("module", "peer-authz"),
		scoreManager: scoreManager,
		subnetCounts: make(map[string]int),
		asnCounts:    make(map[uint32]int),
	}
}

// AuthorizePeer checks if a peer is authorized to connect based on current limits.
func (a *PeerAuthorizer) AuthorizePeer(info PeerInfo) (bool, string) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	totalPeers := a.inboundCount + a.outboundCount

	// Check total peer limit
	if totalPeers >= a.config.MaxPeers {
		return false, "max peers reached"
	}

	// Check direction-specific limits
	if info.IsInbound && a.inboundCount >= a.config.MaxInboundPeers {
		return false, "max inbound peers reached"
	}
	if !info.IsInbound && a.outboundCount >= a.config.MaxOutboundPeers {
		return false, "max outbound peers reached"
	}

	// Check peer score threshold
	if a.scoreManager != nil {
		score := a.scoreManager.GetScore(info.ID)
		if score.Total < a.config.PeerScoreThreshold {
			return false, fmt.Sprintf("peer score too low: %.2f", score.Total)
		}
	}

	return true, ""
}

// RegisterPeer registers a new peer connection.
func (a *PeerAuthorizer) RegisterPeer(info PeerInfo) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if info.IsInbound {
		a.inboundCount++
	} else {
		a.outboundCount++
	}

	if info.Subnet != "" {
		a.subnetCounts[info.Subnet]++
	}
	if info.ASN != 0 {
		a.asnCounts[info.ASN]++
	}

	a.logger.Debug("peer registered",
		"peer", info.ID,
		"inbound", info.IsInbound,
		"total", a.inboundCount+a.outboundCount)

	return nil
}

// UnregisterPeer removes a peer connection.
func (a *PeerAuthorizer) UnregisterPeer(info PeerInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if info.IsInbound {
		a.inboundCount--
		if a.inboundCount < 0 {
			a.inboundCount = 0
		}
	} else {
		a.outboundCount--
		if a.outboundCount < 0 {
			a.outboundCount = 0
		}
	}

	if info.Subnet != "" {
		a.subnetCounts[info.Subnet]--
		if a.subnetCounts[info.Subnet] <= 0 {
			delete(a.subnetCounts, info.Subnet)
		}
	}
	if info.ASN != 0 {
		a.asnCounts[info.ASN]--
		if a.asnCounts[info.ASN] <= 0 {
			delete(a.asnCounts, info.ASN)
		}
	}

	a.logger.Debug("peer unregistered",
		"peer", info.ID,
		"inbound", info.IsInbound,
		"total", a.inboundCount+a.outboundCount)
}

// GetPeerCounts returns current peer counts.
func (a *PeerAuthorizer) GetPeerCounts() (inbound, outbound, total int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.inboundCount, a.outboundCount, a.inboundCount + a.outboundCount
}

// PeerScoreManager manages peer reputation scores.
type PeerScoreManager struct {
	params    PeerScoreParams
	scores    map[PeerID]*PeerScore
	logger    log.Logger
	
	mu sync.RWMutex
}

// NewPeerScoreManager creates a new peer score manager.
func NewPeerScoreManager(params PeerScoreParams, logger log.Logger) *PeerScoreManager {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &PeerScoreManager{
		params: params,
		scores: make(map[PeerID]*PeerScore),
		logger: logger.With("module", "peer-score"),
	}
}

// GetScore returns the current score for a peer.
func (m *PeerScoreManager) GetScore(id PeerID) PeerScore {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if score, ok := m.scores[id]; ok {
		return *score
	}

	// Return default score for unknown peers
	return PeerScore{
		Total:       0,
		LastUpdated: time.Now(),
	}
}

// UpdateScore updates a peer's score based on observed behavior.
func (m *PeerScoreManager) UpdateScore(id PeerID, update func(*PeerScore)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	score, ok := m.scores[id]
	if !ok {
		score = &PeerScore{LastUpdated: time.Now()}
		m.scores[id] = score
	}

	update(score)
	m.recalculateTotal(score)
	score.LastUpdated = time.Now()

	m.logger.Debug("peer score updated",
		"peer", id,
		"score", score.Total)
}

// recalculateTotal computes the total score from components.
func (m *PeerScoreManager) recalculateTotal(score *PeerScore) {
	positive := score.UptimeScore*m.params.UptimeWeight +
		score.ResponseScore*m.params.ResponseWeight +
		score.BehaviorScore*m.params.BehaviorWeight +
		score.StakeScore*m.params.StakeWeight +
		score.ValidatorBonus

	negative := score.DisconnectionPenalty +
		score.MisbehaviorPenalty +
		score.RateLimitPenalty

	score.Total = positive - negative
}

// ApplyDecay applies time-based decay to all scores.
func (m *PeerScoreManager) ApplyDecay(elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hours := elapsed.Hours()
	if hours <= 0 {
		return
	}

	for _, score := range m.scores {
		// Decay positive scores toward zero
		score.UptimeScore *= (1 - m.params.ScoreDecayRate*hours)
		score.ResponseScore *= (1 - m.params.ScoreDecayRate*hours)
		score.BehaviorScore *= (1 - m.params.ScoreDecayRate*hours)
		
		// Decay penalties toward zero
		score.DisconnectionPenalty *= (1 - m.params.PenaltyDecayRate*hours)
		score.MisbehaviorPenalty *= (1 - m.params.PenaltyDecayRate*hours)
		score.RateLimitPenalty *= (1 - m.params.PenaltyDecayRate*hours)

		m.recalculateTotal(score)
		score.LastUpdated = time.Now()
	}
}

// RecordGoodBehavior rewards a peer for good behavior.
func (m *PeerScoreManager) RecordGoodBehavior(id PeerID, amount float64) {
	m.UpdateScore(id, func(s *PeerScore) {
		s.BehaviorScore += amount
		if s.BehaviorScore > 100 {
			s.BehaviorScore = 100
		}
	})
}

// RecordMisbehavior penalizes a peer for bad behavior.
func (m *PeerScoreManager) RecordMisbehavior(id PeerID, severity float64, reason string) {
	m.UpdateScore(id, func(s *PeerScore) {
		s.MisbehaviorPenalty += severity
	})
	m.logger.Warn("peer misbehavior recorded",
		"peer", id,
		"severity", severity,
		"reason", reason)
}

// RecordDisconnection records a peer disconnection.
func (m *PeerScoreManager) RecordDisconnection(id PeerID, wasClean bool) {
	penalty := 1.0
	if !wasClean {
		penalty = 5.0 // Higher penalty for unclean disconnections
	}
	
	m.UpdateScore(id, func(s *PeerScore) {
		s.DisconnectionPenalty += penalty
	})
}

// RecordRateLimitHit records when a peer hits rate limits.
func (m *PeerScoreManager) RecordRateLimitHit(id PeerID) {
	m.UpdateScore(id, func(s *PeerScore) {
		s.RateLimitPenalty += 2.0
	})
}

// SetStakeScore updates a peer's stake-based score.
func (m *PeerScoreManager) SetStakeScore(id PeerID, stake int64, minStakeForTrust int64) {
	m.UpdateScore(id, func(s *PeerScore) {
		if minStakeForTrust > 0 {
			s.StakeScore = float64(stake) / float64(minStakeForTrust) * 10
			if s.StakeScore > 50 {
				s.StakeScore = 50
			}
		}
	})
}

// SetValidatorBonus sets the validator bonus for a peer.
func (m *PeerScoreManager) SetValidatorBonus(id PeerID, isValidator bool) {
	m.UpdateScore(id, func(s *PeerScore) {
		if isValidator {
			s.ValidatorBonus = m.params.ValidatorBonusMax
		} else {
			s.ValidatorBonus = 0
		}
	})
}

// GetPeersAboveThreshold returns peers with scores above the given threshold.
func (m *PeerScoreManager) GetPeersAboveThreshold(threshold float64) []PeerID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []PeerID
	for id, score := range m.scores {
		if score.Total >= threshold {
			result = append(result, id)
		}
	}
	return result
}

// GetPeersBelowThreshold returns peers with scores below the given threshold.
func (m *PeerScoreManager) GetPeersBelowThreshold(threshold float64) []PeerID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []PeerID
	for id, score := range m.scores {
		if score.Total < threshold {
			result = append(result, id)
		}
	}
	return result
}

// RemovePeer removes a peer's score record.
func (m *PeerScoreManager) RemovePeer(id PeerID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.scores, id)
}
