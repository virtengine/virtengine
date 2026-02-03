package app

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

// NetworkSecurityManager coordinates all network security components.
type NetworkSecurityManager struct {
	config NetworkSecurityConfig
	logger log.Logger

	// Security components
	noiseTransport     *NoiseTransport
	peerAuthenticator  *PeerAuthenticator
	peerAuthorizer     *PeerAuthorizer
	peerScoreManager   *PeerScoreManager
	networkRateLimiter *NetworkRateLimiter
	ddosProtector      *DDoSProtector
	sybilProtector     *SybilProtector
	eclipseProtector   *EclipseProtector
	firewallGenerator  *FirewallRuleGenerator
	idsIntegration     *IDSIntegration

	// State
	started bool

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// NewNetworkSecurityManager creates a new network security manager.
func NewNetworkSecurityManager(config NetworkSecurityConfig, logger log.Logger) (*NetworkSecurityManager, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid network security config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	nsm := &NetworkSecurityManager{
		config: config,
		logger: logger.With("module", "network-security"),
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize components based on configuration
	var err error

	// Noise Protocol transport
	if config.Noise.Enabled {
		nsm.noiseTransport, err = NewNoiseTransport(config.Noise)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize Noise transport: %w", err)
		}
		nsm.logger.Info("Noise Protocol encryption enabled")
	}

	// Peer scoring
	nsm.peerScoreManager = NewPeerScoreManager(DefaultPeerScoreParams(), logger)

	// Peer authentication
	if config.Peer.Enabled {
		nsm.peerAuthenticator = NewPeerAuthenticator(config.Peer, logger)
		nsm.peerAuthorizer = NewPeerAuthorizer(config.Peer, nsm.peerScoreManager, logger)
		nsm.logger.Info("peer authentication enabled")
	}

	// Network rate limiting
	if config.RateLimit.Enabled {
		nsm.networkRateLimiter = NewNetworkRateLimiter(config.RateLimit, logger)
		nsm.logger.Info("network rate limiting enabled")
	}

	// DDoS protection
	if config.Protection.DDoS.Enabled {
		nsm.ddosProtector = NewDDoSProtector(config.Protection.DDoS, logger)
		nsm.logger.Info("DDoS protection enabled")
	}

	// Sybil protection
	if config.Protection.Sybil.Enabled {
		nsm.sybilProtector = NewSybilProtector(config.Protection.Sybil, config.Peer, logger)
		nsm.logger.Info("Sybil protection enabled")
	}

	// Eclipse protection
	if config.Protection.Eclipse.Enabled {
		nsm.eclipseProtector = NewEclipseProtector(config.Protection.Eclipse, nil, logger)
		nsm.logger.Info("Eclipse protection enabled")
	}

	// Firewall integration
	if config.Firewall.Enabled {
		nsm.firewallGenerator = NewFirewallRuleGenerator(config.Firewall, logger)
		nsm.logger.Info("firewall integration enabled")
	}

	// IDS integration
	if config.IDS.Enabled {
		nsm.idsIntegration, err = NewIDSIntegration(config.IDS, logger)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize IDS integration: %w", err)
		}

		// Connect DDoS alerts to IDS
		if nsm.ddosProtector != nil && nsm.idsIntegration != nil {
			nsm.ddosProtector.SetAlertCallback(func(event AlertEvent) {
				nsm.idsIntegration.AlertFromEvent(event)
			})
		}

		nsm.logger.Info("IDS integration enabled")
	}

	return nsm, nil
}

// Start starts all background security workers.
func (nsm *NetworkSecurityManager) Start() error {
	nsm.mu.Lock()
	defer nsm.mu.Unlock()

	if nsm.started {
		return nil
	}

	// Start IDS processing
	if nsm.idsIntegration != nil {
		nsm.idsIntegration.Start()
	}

	// Start background maintenance
	nsm.wg.Add(1)
	go nsm.maintenanceLoop()

	nsm.started = true
	nsm.logger.Info("network security manager started")

	return nil
}

// Stop stops all security components.
func (nsm *NetworkSecurityManager) Stop() error {
	nsm.mu.Lock()
	if !nsm.started {
		nsm.mu.Unlock()
		return nil
	}
	nsm.mu.Unlock()

	nsm.cancel()
	nsm.wg.Wait()

	if nsm.idsIntegration != nil {
		nsm.idsIntegration.Stop()
	}

	nsm.logger.Info("network security manager stopped")
	return nil
}

// maintenanceLoop performs periodic maintenance tasks.
func (nsm *NetworkSecurityManager) maintenanceLoop() {
	defer nsm.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	scoreTicker := time.NewTicker(time.Hour)
	defer scoreTicker.Stop()

	for {
		select {
		case <-nsm.ctx.Done():
			return
		case <-ticker.C:
			nsm.performMaintenance()
		case <-scoreTicker.C:
			nsm.decayScores()
		}
	}
}

// performMaintenance runs periodic maintenance tasks.
func (nsm *NetworkSecurityManager) performMaintenance() {
	// Cleanup expired rate limit entries
	if nsm.networkRateLimiter != nil {
		nsm.networkRateLimiter.CleanupStaleEntries(30 * time.Minute)
	}

	// Cleanup expired bans
	if nsm.ddosProtector != nil {
		nsm.ddosProtector.CleanupExpiredBans()
	}

	// Cleanup expired firewall entries
	if nsm.firewallGenerator != nil {
		nsm.firewallGenerator.CleanupExpired()
	}

	nsm.logger.Debug("maintenance completed")
}

// decayScores applies score decay to all peers.
func (nsm *NetworkSecurityManager) decayScores() {
	if nsm.peerScoreManager != nil {
		nsm.peerScoreManager.ApplyDecay(time.Hour)
		nsm.logger.Debug("peer scores decayed")
	}
}

// SecureOutbound wraps a connection with Noise Protocol encryption for outbound connections.
func (nsm *NetworkSecurityManager) SecureOutbound(conn net.Conn, remotePublicKey []byte) (net.Conn, error) {
	if nsm.noiseTransport == nil || !nsm.config.Noise.Enabled {
		return conn, nil // Return unwrapped connection if Noise is disabled
	}

	session, err := nsm.noiseTransport.SecureOutbound(conn, remotePublicKey)
	if err != nil {
		telemetry.IncrCounter(1, "network", "noise", "handshake_failed", "outbound")
		return nil, fmt.Errorf("noise handshake failed: %w", err)
	}

	telemetry.IncrCounter(1, "network", "noise", "handshake_success", "outbound")
	return session, nil
}

// SecureInbound wraps a connection with Noise Protocol encryption for inbound connections.
func (nsm *NetworkSecurityManager) SecureInbound(conn net.Conn) (net.Conn, error) {
	if nsm.noiseTransport == nil || !nsm.config.Noise.Enabled {
		return conn, nil // Return unwrapped connection if Noise is disabled
	}

	session, err := nsm.noiseTransport.SecureInbound(conn)
	if err != nil {
		telemetry.IncrCounter(1, "network", "noise", "handshake_failed", "inbound")
		return nil, fmt.Errorf("noise handshake failed: %w", err)
	}

	telemetry.IncrCounter(1, "network", "noise", "handshake_success", "inbound")
	return session, nil
}

// GetNoisePublicKey returns the node's Noise Protocol static public key.
func (nsm *NetworkSecurityManager) GetNoisePublicKey() []byte {
	if nsm.noiseTransport == nil {
		return nil
	}
	return nsm.noiseTransport.GetPublicKey()
}

// AuthenticatePeer checks if a peer is allowed to connect.
func (nsm *NetworkSecurityManager) AuthenticatePeer(info PeerInfo) (bool, string) {
	if nsm.peerAuthenticator == nil {
		return true, ""
	}
	return nsm.peerAuthenticator.AuthenticatePeer(info)
}

// AuthorizePeer checks if a peer should be accepted based on current limits.
func (nsm *NetworkSecurityManager) AuthorizePeer(info PeerInfo) (bool, string) {
	// Check Sybil protection first
	if nsm.sybilProtector != nil {
		if allowed, reason := nsm.sybilProtector.AllowPeer(info); !allowed {
			if nsm.idsIntegration != nil {
				nsm.idsIntegration.AlertSybil(reason, extractIP(info.Address), info.Subnet, 0)
			}
			return false, reason
		}
	}

	// Check peer authorization
	if nsm.peerAuthorizer != nil {
		return nsm.peerAuthorizer.AuthorizePeer(info)
	}

	return true, ""
}

// RegisterPeer registers a new peer connection.
func (nsm *NetworkSecurityManager) RegisterPeer(info PeerInfo) error {
	// Compute subnet if not provided
	if info.Subnet == "" && info.Address != nil {
		info.Subnet = GetSubnetFromAddr(info.Address)
	}

	if nsm.peerAuthorizer != nil {
		if err := nsm.peerAuthorizer.RegisterPeer(info); err != nil {
			return err
		}
	}

	if nsm.sybilProtector != nil {
		nsm.sybilProtector.RegisterPeer(info)
	}

	if nsm.peerScoreManager != nil {
		nsm.peerScoreManager.SetStakeScore(info.ID, info.Stake, nsm.config.Peer.MinStakeForTrust)
		nsm.peerScoreManager.SetValidatorBonus(info.ID, info.IsValidator)
	}

	// Try to set as anchor if it's a good peer
	if nsm.eclipseProtector != nil && info.IsValidator {
		nsm.eclipseProtector.SetAnchor(info)
	}

	telemetry.IncrCounter(1, "network", "peers", "registered")
	return nil
}

// UnregisterPeer removes a peer from all tracking.
func (nsm *NetworkSecurityManager) UnregisterPeer(info PeerInfo, wasClean bool) {
	if nsm.peerAuthorizer != nil {
		nsm.peerAuthorizer.UnregisterPeer(info)
	}

	if nsm.sybilProtector != nil {
		nsm.sybilProtector.UnregisterPeer(info)
	}

	if nsm.peerScoreManager != nil {
		nsm.peerScoreManager.RecordDisconnection(info.ID, wasClean)
	}

	if nsm.eclipseProtector != nil {
		nsm.eclipseProtector.UnregisterOutboundPeer(info.ID)
		nsm.eclipseProtector.RemoveAnchor(info.ID)
	}

	telemetry.IncrCounter(1, "network", "peers", "unregistered")
}

// AllowConnection checks if a new connection is allowed.
func (nsm *NetworkSecurityManager) AllowConnection(remoteAddr net.Addr) (bool, string) {
	// Check DDoS protection
	if nsm.ddosProtector != nil {
		ip := extractIP(remoteAddr)
		if blocked, reason := nsm.ddosProtector.RecordConnection(ip); blocked {
			return false, reason
		}
	}

	// Check rate limiting
	if nsm.networkRateLimiter != nil {
		return nsm.networkRateLimiter.AllowConnection(remoteAddr)
	}

	return true, ""
}

// AllowMessage checks if a message is allowed.
func (nsm *NetworkSecurityManager) AllowMessage(remoteAddr net.Addr, size int) (bool, string) {
	// Check DDoS protection
	if nsm.ddosProtector != nil {
		ip := extractIP(remoteAddr)
		if blocked, reason := nsm.ddosProtector.RecordMessage(ip, size); blocked {
			return false, reason
		}
	}

	// Check rate limiting
	if nsm.networkRateLimiter != nil {
		return nsm.networkRateLimiter.AllowMessage(remoteAddr, size)
	}

	return true, ""
}

// RecordGoodBehavior records good peer behavior.
func (nsm *NetworkSecurityManager) RecordGoodBehavior(peerID PeerID, amount float64) {
	if nsm.peerScoreManager != nil {
		nsm.peerScoreManager.RecordGoodBehavior(peerID, amount)
	}
}

// RecordMisbehavior records peer misbehavior.
func (nsm *NetworkSecurityManager) RecordMisbehavior(peerID PeerID, severity float64, reason string) {
	if nsm.peerScoreManager != nil {
		nsm.peerScoreManager.RecordMisbehavior(peerID, severity, reason)
	}

	// Check if peer should be banned
	score := nsm.GetPeerScore(peerID)
	if score.Total < nsm.config.Peer.PeerScoreThreshold && nsm.peerAuthenticator != nil {
		nsm.peerAuthenticator.BanPeer(peerID, time.Hour)
	}
}

// GetPeerScore returns the score for a peer.
func (nsm *NetworkSecurityManager) GetPeerScore(peerID PeerID) PeerScore {
	if nsm.peerScoreManager != nil {
		return nsm.peerScoreManager.GetScore(peerID)
	}
	return PeerScore{}
}

// BanPeer bans a peer for the specified duration.
func (nsm *NetworkSecurityManager) BanPeer(peerID PeerID, duration time.Duration) {
	if nsm.peerAuthenticator != nil {
		nsm.peerAuthenticator.BanPeer(peerID, duration)
	}
}

// UnbanPeer removes a ban from a peer.
func (nsm *NetworkSecurityManager) UnbanPeer(peerID PeerID) {
	if nsm.peerAuthenticator != nil {
		nsm.peerAuthenticator.UnbanPeer(peerID)
	}
}

// BanIP bans an IP address for the specified duration.
func (nsm *NetworkSecurityManager) BanIP(ip string, duration time.Duration) {
	if nsm.ddosProtector != nil {
		nsm.ddosProtector.BanIP(ip, duration)
	}
	if nsm.firewallGenerator != nil {
		nsm.firewallGenerator.AddBlockedIP(ip, duration)
	}
}

// UpdateSystemLoad updates system load for adaptive rate limiting.
func (nsm *NetworkSecurityManager) UpdateSystemLoad(cpuLoad, memoryLoad float64) {
	if nsm.networkRateLimiter != nil {
		nsm.networkRateLimiter.UpdateSystemLoad(cpuLoad, memoryLoad)
	}
}

// GetPeersToRotate returns peers that should be disconnected for rotation.
func (nsm *NetworkSecurityManager) GetPeersToRotate() []PeerID {
	if nsm.eclipseProtector != nil {
		return nsm.eclipseProtector.GetPeersToRotate()
	}
	return nil
}

// NeedsSeedRefresh checks if seed nodes should be refreshed.
func (nsm *NetworkSecurityManager) NeedsSeedRefresh() bool {
	if nsm.eclipseProtector != nil {
		return nsm.eclipseProtector.NeedsSeedRefresh()
	}
	return false
}

// MarkSeedRefreshed marks seed nodes as refreshed.
func (nsm *NetworkSecurityManager) MarkSeedRefreshed() {
	if nsm.eclipseProtector != nil {
		nsm.eclipseProtector.MarkSeedRefreshed()
	}
}

// GenerateFirewallRules generates firewall rules for the configured firewall type.
func (nsm *NetworkSecurityManager) GenerateFirewallRules() (string, error) {
	if nsm.firewallGenerator == nil {
		return "", fmt.Errorf("firewall integration not enabled")
	}
	return nsm.firewallGenerator.Generate()
}

// GetStats returns comprehensive security statistics.
func (nsm *NetworkSecurityManager) GetStats() NetworkSecurityStats {
	stats := NetworkSecurityStats{}

	if nsm.peerAuthorizer != nil {
		stats.InboundPeers, stats.OutboundPeers, stats.TotalPeers = nsm.peerAuthorizer.GetPeerCounts()
	}

	if nsm.networkRateLimiter != nil {
		rlStats := nsm.networkRateLimiter.GetStats()
		stats.RateLimitActiveIPs = rlStats.ActiveIPs
		stats.RateLimitBlocked = rlStats.TotalBlocked
		stats.AdaptiveMultiplier = rlStats.AdaptiveMultiplier
	}

	if nsm.sybilProtector != nil {
		stats.DiversityScore = nsm.sybilProtector.GetDiversityScore()
	}

	if nsm.eclipseProtector != nil {
		stats.AnchorConnections = nsm.eclipseProtector.GetAnchorCount()
	}

	if nsm.idsIntegration != nil {
		idsStats := nsm.idsIntegration.GetStats()
		stats.IDSAlertsSent = idsStats.AlertsSent
		stats.IDSAlertsDropped = idsStats.AlertsDropped
	}

	return stats
}

// NetworkSecurityStats contains comprehensive security statistics.
type NetworkSecurityStats struct {
	// Peer statistics
	TotalPeers    int
	InboundPeers  int
	OutboundPeers int

	// Rate limiting
	RateLimitActiveIPs int
	RateLimitBlocked   int
	AdaptiveMultiplier float64

	// Attack prevention
	DiversityScore    float64
	AnchorConnections int

	// IDS
	IDSAlertsSent    int64
	IDSAlertsDropped int64
}

// GetConfig returns the current configuration.
func (nsm *NetworkSecurityManager) GetConfig() NetworkSecurityConfig {
	nsm.mu.RLock()
	defer nsm.mu.RUnlock()
	return nsm.config
}

// AddTrustedPeer adds a peer to the trusted list.
func (nsm *NetworkSecurityManager) AddTrustedPeer(peerID PeerID) {
	if nsm.peerAuthenticator != nil {
		nsm.peerAuthenticator.AddTrustedPeer(peerID)
	}
}

// AddTrustedKey adds a Noise Protocol trusted public key.
func (nsm *NetworkSecurityManager) AddTrustedKey(publicKey []byte) {
	if nsm.noiseTransport != nil {
		nsm.noiseTransport.AddTrustedKey(publicKey)
	}
}

// IsPeerBanned checks if a peer is banned.
func (nsm *NetworkSecurityManager) IsPeerBanned(peerID PeerID) bool {
	if nsm.peerAuthenticator != nil {
		return nsm.peerAuthenticator.IsBanned(peerID)
	}
	return false
}
