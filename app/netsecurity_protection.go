package app

import (
	"container/ring"
	"net"
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

// DDoSProtector implements DDoS mitigation strategies.
type DDoSProtector struct {
	config  DDoSProtectionConfig
	logger  log.Logger
	
	// Detection windows (sliding windows for attack detection)
	connectionWindow *SlidingWindow
	messageWindow    *SlidingWindow
	
	// Banned IPs with expiry
	bannedIPs map[string]time.Time
	
	// Alert callback
	alertCallback func(AlertEvent)
	
	mu sync.RWMutex
}

// SlidingWindow implements a sliding time window for counting events.
type SlidingWindow struct {
	windowSize time.Duration
	bucketSize time.Duration
	buckets    *ring.Ring
	totalCount int64
	
	mu sync.Mutex
}

// WindowBucket holds count for a time bucket.
type WindowBucket struct {
	Count     int64
	Timestamp time.Time
}

// NewSlidingWindow creates a new sliding window.
func NewSlidingWindow(windowSize, bucketSize time.Duration) *SlidingWindow {
	numBuckets := int(windowSize / bucketSize)
	if numBuckets < 1 {
		numBuckets = 1
	}

	r := ring.New(numBuckets)
	for i := 0; i < numBuckets; i++ {
		r.Value = &WindowBucket{Timestamp: time.Now()}
		r = r.Next()
	}

	return &SlidingWindow{
		windowSize: windowSize,
		bucketSize: bucketSize,
		buckets:    r,
	}
}

// Add adds an event to the window.
func (sw *SlidingWindow) Add(count int64) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.rotate()
	bucket := sw.buckets.Value.(*WindowBucket)
	bucket.Count += count
	sw.totalCount += count
}

// Count returns the total count in the window.
func (sw *SlidingWindow) Count() int64 {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.rotate()
	return sw.totalCount
}

// rotate removes old buckets and adds new ones.
func (sw *SlidingWindow) rotate() {
	now := time.Now()
	
	// Rotate through buckets and remove expired ones
	sw.buckets.Do(func(val interface{}) {
		bucket := val.(*WindowBucket)
		if now.Sub(bucket.Timestamp) > sw.windowSize {
			sw.totalCount -= bucket.Count
			bucket.Count = 0
			bucket.Timestamp = now
		}
	})

	// Move to next bucket if current one is old
	currentBucket := sw.buckets.Value.(*WindowBucket)
	if now.Sub(currentBucket.Timestamp) > sw.bucketSize {
		sw.buckets = sw.buckets.Next()
		newBucket := sw.buckets.Value.(*WindowBucket)
		sw.totalCount -= newBucket.Count
		newBucket.Count = 0
		newBucket.Timestamp = now
	}
}

// AlertEvent represents a security alert.
type AlertEvent struct {
	Type      string
	Severity  string // low, medium, high, critical
	Source    string // IP or peer ID
	Message   string
	Timestamp time.Time
	Details   map[string]interface{}
}

// NewDDoSProtector creates a new DDoS protector.
func NewDDoSProtector(config DDoSProtectionConfig, logger log.Logger) *DDoSProtector {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &DDoSProtector{
		config:           config,
		logger:           logger.With("module", "ddos-protection"),
		connectionWindow: NewSlidingWindow(time.Minute, time.Second),
		messageWindow:    NewSlidingWindow(time.Minute, time.Second),
		bannedIPs:        make(map[string]time.Time),
	}
}

// SetAlertCallback sets the callback for security alerts.
func (d *DDoSProtector) SetAlertCallback(callback func(AlertEvent)) {
	d.alertCallback = callback
}

// RecordConnection records a new connection attempt.
func (d *DDoSProtector) RecordConnection(ip string) (blocked bool, reason string) {
	if !d.config.Enabled {
		return false, ""
	}

	// Check if banned
	if d.isBanned(ip) {
		return true, "IP is banned"
	}

	d.connectionWindow.Add(1)
	count := d.connectionWindow.Count()

	// Check for connection flood
	if int(count) > d.config.ConnectionFloodThreshold {
		d.handleConnectionFlood(ip, count)
		return true, "connection flood detected"
	}

	return false, ""
}

// RecordMessage records a new message.
func (d *DDoSProtector) RecordMessage(ip string, size int) (blocked bool, reason string) {
	if !d.config.Enabled {
		return false, ""
	}

	// Check if banned
	if d.isBanned(ip) {
		return true, "IP is banned"
	}

	d.messageWindow.Add(1)
	count := d.messageWindow.Count()

	// Check for message flood
	if int(count) > d.config.MessageFloodThreshold {
		d.handleMessageFlood(ip, count)
		return true, "message flood detected"
	}

	return false, ""
}

// handleConnectionFlood handles detected connection flood.
func (d *DDoSProtector) handleConnectionFlood(ip string, count int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Warn("connection flood detected",
		"ip", ip,
		"count", count,
		"threshold", d.config.ConnectionFloodThreshold)

	// Ban the IP
	d.bannedIPs[ip] = time.Now().Add(d.config.BanDuration)

	telemetry.IncrCounter(1, "network", "ddos", "connection_flood")

	// Send alert
	if d.alertCallback != nil && int(count) > d.config.AlertThreshold {
		d.alertCallback(AlertEvent{
			Type:      "connection_flood",
			Severity:  "high",
			Source:    ip,
			Message:   "Connection flood attack detected",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"count":     count,
				"threshold": d.config.ConnectionFloodThreshold,
			},
		})
	}
}

// handleMessageFlood handles detected message flood.
func (d *DDoSProtector) handleMessageFlood(ip string, count int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Warn("message flood detected",
		"ip", ip,
		"count", count,
		"threshold", d.config.MessageFloodThreshold)

	// Ban the IP
	d.bannedIPs[ip] = time.Now().Add(d.config.BanDuration)

	telemetry.IncrCounter(1, "network", "ddos", "message_flood")

	// Send alert
	if d.alertCallback != nil && int(count) > d.config.AlertThreshold {
		d.alertCallback(AlertEvent{
			Type:      "message_flood",
			Severity:  "high",
			Source:    ip,
			Message:   "Message flood attack detected",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"count":     count,
				"threshold": d.config.MessageFloodThreshold,
			},
		})
	}
}

// isBanned checks if an IP is currently banned.
func (d *DDoSProtector) isBanned(ip string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	expiry, banned := d.bannedIPs[ip]
	if banned && time.Now().Before(expiry) {
		return true
	}
	return false
}

// BanIP manually bans an IP.
func (d *DDoSProtector) BanIP(ip string, duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.bannedIPs[ip] = time.Now().Add(duration)
	d.logger.Info("IP banned", "ip", ip, "duration", duration)
}

// UnbanIP removes a ban.
func (d *DDoSProtector) UnbanIP(ip string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.bannedIPs, ip)
	d.logger.Info("IP unbanned", "ip", ip)
}

// CleanupExpiredBans removes expired bans.
func (d *DDoSProtector) CleanupExpiredBans() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for ip, expiry := range d.bannedIPs {
		if now.After(expiry) {
			delete(d.bannedIPs, ip)
		}
	}
}

// SybilProtector implements Sybil attack prevention.
type SybilProtector struct {
	config     SybilProtectionConfig
	peerConfig PeerConfig
	logger     log.Logger
	
	// Subnet tracking
	subnetPeers map[string][]PeerID // subnet -> list of peers
	
	// ASN tracking
	asnPeers map[uint32][]PeerID // ASN -> list of peers
	
	mu sync.RWMutex
}

// NewSybilProtector creates a new Sybil protector.
func NewSybilProtector(config SybilProtectionConfig, peerConfig PeerConfig, logger log.Logger) *SybilProtector {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &SybilProtector{
		config:      config,
		peerConfig:  peerConfig,
		logger:      logger.With("module", "sybil-protection"),
		subnetPeers: make(map[string][]PeerID),
		asnPeers:    make(map[uint32][]PeerID),
	}
}

// AllowPeer checks if a peer is allowed based on Sybil protection rules.
func (s *SybilProtector) AllowPeer(info PeerInfo) (bool, string) {
	if !s.config.Enabled {
		return true, ""
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check subnet diversity
	if info.Subnet != "" {
		subnetCount := len(s.subnetPeers[info.Subnet])
		if subnetCount >= s.config.MaxPeersPerSubnet {
			s.logger.Warn("subnet limit reached",
				"peer", info.ID,
				"subnet", info.Subnet,
				"count", subnetCount,
				"limit", s.config.MaxPeersPerSubnet)
			return false, "too many peers from same subnet"
		}
	}

	// Check ASN diversity
	if info.ASN != 0 {
		asnCount := len(s.asnPeers[info.ASN])
		if asnCount >= s.config.MaxPeersPerASN {
			s.logger.Warn("ASN limit reached",
				"peer", info.ID,
				"asn", info.ASN,
				"count", asnCount,
				"limit", s.config.MaxPeersPerASN)
			return false, "too many peers from same ASN"
		}
	}

	// Check stake requirement
	if s.config.RequireStakeForConnection && info.Stake < s.config.MinimumStake {
		s.logger.Debug("insufficient stake",
			"peer", info.ID,
			"stake", info.Stake,
			"required", s.config.MinimumStake)
		return false, "insufficient stake"
	}

	return true, ""
}

// RegisterPeer registers a peer for Sybil tracking.
func (s *SybilProtector) RegisterPeer(info PeerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if info.Subnet != "" {
		s.subnetPeers[info.Subnet] = append(s.subnetPeers[info.Subnet], info.ID)
	}
	if info.ASN != 0 {
		s.asnPeers[info.ASN] = append(s.asnPeers[info.ASN], info.ID)
	}
}

// UnregisterPeer removes a peer from Sybil tracking.
func (s *SybilProtector) UnregisterPeer(info PeerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if info.Subnet != "" {
		s.subnetPeers[info.Subnet] = removePeerID(s.subnetPeers[info.Subnet], info.ID)
		if len(s.subnetPeers[info.Subnet]) == 0 {
			delete(s.subnetPeers, info.Subnet)
		}
	}
	if info.ASN != 0 {
		s.asnPeers[info.ASN] = removePeerID(s.asnPeers[info.ASN], info.ID)
		if len(s.asnPeers[info.ASN]) == 0 {
			delete(s.asnPeers, info.ASN)
		}
	}
}

// GetDiversityScore returns a score representing network diversity.
func (s *SybilProtector) GetDiversityScore() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uniqueSubnets := len(s.subnetPeers)
	uniqueASNs := len(s.asnPeers)

	// Calculate diversity score (0-100)
	subnetScore := float64(uniqueSubnets) / float64(s.config.DiversityRequirement) * 50
	if subnetScore > 50 {
		subnetScore = 50
	}

	asnScore := float64(uniqueASNs) / float64(s.config.DiversityRequirement) * 50
	if asnScore > 50 {
		asnScore = 50
	}

	return subnetScore + asnScore
}

// EclipseProtector implements Eclipse attack prevention.
type EclipseProtector struct {
	config      EclipseProtectionConfig
	logger      log.Logger
	
	// Anchor connections (long-lived, trusted connections)
	anchors map[PeerID]*AnchorConnection
	
	// Outbound-only slots
	outboundOnlyPeers map[PeerID]bool
	
	// Seed node connections
	seedNodes       []string
	lastSeedRefresh time.Time
	
	// Peer rotation tracking
	peerConnectTimes map[PeerID]time.Time
	lastRotation     time.Time
	
	mu sync.RWMutex
}

// AnchorConnection represents a long-lived anchor connection.
type AnchorConnection struct {
	PeerID      PeerID
	ConnectedAt time.Time
	IsValidator bool
	Stake       int64
}

// NewEclipseProtector creates a new Eclipse protector.
func NewEclipseProtector(config EclipseProtectionConfig, seedNodes []string, logger log.Logger) *EclipseProtector {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &EclipseProtector{
		config:            config,
		logger:            logger.With("module", "eclipse-protection"),
		anchors:           make(map[PeerID]*AnchorConnection),
		outboundOnlyPeers: make(map[PeerID]bool),
		seedNodes:         seedNodes,
		peerConnectTimes:  make(map[PeerID]time.Time),
	}
}

// ShouldReserveOutboundSlot checks if a slot should be reserved for outbound.
func (e *EclipseProtector) ShouldReserveOutboundSlot(currentOutbound int) bool {
	if !e.config.Enabled {
		return false
	}
	return currentOutbound < e.config.OutboundOnlySlots
}

// IsOutboundOnlySlot checks if a peer is in an outbound-only slot.
func (e *EclipseProtector) IsOutboundOnlySlot(peerID PeerID) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.outboundOnlyPeers[peerID]
}

// RegisterOutboundPeer registers a peer in an outbound-only slot.
func (e *EclipseProtector) RegisterOutboundPeer(peerID PeerID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.outboundOnlyPeers[peerID] = true
	e.peerConnectTimes[peerID] = time.Now()
}

// UnregisterOutboundPeer removes a peer from outbound-only tracking.
func (e *EclipseProtector) UnregisterOutboundPeer(peerID PeerID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.outboundOnlyPeers, peerID)
	delete(e.peerConnectTimes, peerID)
}

// SetAnchor marks a peer as an anchor connection.
func (e *EclipseProtector) SetAnchor(info PeerInfo) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.anchors) >= e.config.AnchorConnections {
		return false
	}

	e.anchors[info.ID] = &AnchorConnection{
		PeerID:      info.ID,
		ConnectedAt: time.Now(),
		IsValidator: info.IsValidator,
		Stake:       info.Stake,
	}

	e.logger.Info("anchor connection established", "peer", info.ID)
	return true
}

// IsAnchor checks if a peer is an anchor connection.
func (e *EclipseProtector) IsAnchor(peerID PeerID) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.anchors[peerID]
	return ok
}

// RemoveAnchor removes an anchor connection.
func (e *EclipseProtector) RemoveAnchor(peerID PeerID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.anchors, peerID)
	e.logger.Info("anchor connection removed", "peer", peerID)
}

// GetPeersToRotate returns peers that should be rotated out.
func (e *EclipseProtector) GetPeersToRotate() []PeerID {
	if !e.config.Enabled {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now()
	var peersToRotate []PeerID

	for peerID, connTime := range e.peerConnectTimes {
		// Don't rotate anchors
		if _, isAnchor := e.anchors[peerID]; isAnchor {
			continue
		}

		// Don't rotate outbound-only slots
		if e.outboundOnlyPeers[peerID] {
			continue
		}

		// Rotate peers that have been connected longer than the interval
		if now.Sub(connTime) > e.config.PeerRotationInterval {
			peersToRotate = append(peersToRotate, peerID)
		}
	}

	return peersToRotate
}

// NeedsSeedRefresh checks if seed nodes should be refreshed.
func (e *EclipseProtector) NeedsSeedRefresh() bool {
	if !e.config.Enabled {
		return false
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return time.Since(e.lastSeedRefresh) > e.config.SeedNodeRefreshInterval
}

// MarkSeedRefreshed marks seed nodes as refreshed.
func (e *EclipseProtector) MarkSeedRefreshed() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastSeedRefresh = time.Now()
}

// SelectPeersRandomly returns a random selection of peers.
// Used to ensure some randomness in peer selection to prevent eclipse attacks.
func (e *EclipseProtector) SelectPeersRandomly(peers []PeerID, count int) []PeerID {
	if !e.config.Enabled || count >= len(peers) {
		return peers
	}

	// Calculate how many to select randomly vs by reputation
	randomCount := int(float64(count) * e.config.RandomSelectionRatio)
	if randomCount < 1 {
		randomCount = 1
	}

	// Simple selection (in production, use crypto/rand for shuffling)
	selected := make([]PeerID, 0, count)
	for i := 0; i < min(randomCount, len(peers)); i++ {
		selected = append(selected, peers[i])
	}

	return selected
}

// GetAnchorCount returns the number of anchor connections.
func (e *EclipseProtector) GetAnchorCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.anchors)
}

// Helper functions

func removePeerID(slice []PeerID, id PeerID) []PeerID {
	for i, pid := range slice {
		if pid == id {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetSubnetFromAddr extracts the /24 subnet from a network address.
func GetSubnetFromAddr(addr net.Addr) string {
	ip := extractIP(addr)
	return extractSubnet(ip)
}
