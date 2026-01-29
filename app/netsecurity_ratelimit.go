package app

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

// NetworkRateLimiter implements network-layer rate limiting.
type NetworkRateLimiter struct {
	config    NetworkRateLimitConfig
	logger    log.Logger
	
	// Global rate limiting
	globalConnBucket *TokenBucket
	globalMsgBucket  *TokenBucket
	
	// Per-IP rate limiting
	ipLimiters map[string]*IPLimiter
	
	// Bandwidth tracking
	bytesSent     int64
	bytesReceived int64
	
	// Adaptive rate limiting
	adaptiveMultiplier float64
	systemLoad         float64
	
	// Whitelist tracking
	whitelistedIPs map[string]bool
	
	mu sync.RWMutex
}

// IPLimiter tracks rate limits for a single IP address.
type IPLimiter struct {
	ConnectionBucket *TokenBucket
	MessageBucket    *TokenBucket
	BandwidthBucket  *TokenBucket
	LastSeen         time.Time
	ConnectionCount  int
	BlockedCount     int
}

// TokenBucket implements the token bucket algorithm for rate limiting.
type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	
	mu sync.Mutex
}

// NewTokenBucket creates a new token bucket.
func NewTokenBucket(maxTokens float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// TryConsume attempts to consume tokens from the bucket.
func (tb *TokenBucket) TryConsume(tokens float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}
	return false
}

// refill adds tokens based on elapsed time.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now
}

// Available returns the current number of available tokens.
func (tb *TokenBucket) Available() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

// NewNetworkRateLimiter creates a new network rate limiter.
func NewNetworkRateLimiter(config NetworkRateLimitConfig, logger log.Logger) *NetworkRateLimiter {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	// Create whitelist map
	whitelistedIPs := make(map[string]bool)
	for _, ip := range config.WhitelistedIPs {
		whitelistedIPs[ip] = true
	}

	rl := &NetworkRateLimiter{
		config:             config,
		logger:             logger.With("module", "network-ratelimit"),
		globalConnBucket:   NewTokenBucket(float64(config.BurstSize), float64(config.ConnectionsPerSecond)),
		globalMsgBucket:    NewTokenBucket(float64(config.BurstSize*10), float64(config.MessagesPerSecond*10)),
		ipLimiters:         make(map[string]*IPLimiter),
		adaptiveMultiplier: 1.0,
		whitelistedIPs:     whitelistedIPs,
	}

	return rl
}

// AllowConnection checks if a new connection is allowed.
func (rl *NetworkRateLimiter) AllowConnection(remoteAddr net.Addr) (bool, string) {
	if !rl.config.Enabled {
		return true, ""
	}

	ip := extractIP(remoteAddr)
	if ip == "" {
		return false, "invalid address"
	}

	// Check whitelist
	rl.mu.RLock()
	if rl.whitelistedIPs[ip] {
		rl.mu.RUnlock()
		return true, ""
	}
	rl.mu.RUnlock()

	// Check global connection rate
	if !rl.globalConnBucket.TryConsume(1) {
		rl.recordBlocked("connection", ip, "global_limit")
		return false, "global connection rate limit exceeded"
	}

	// Get or create IP limiter
	limiter := rl.getOrCreateIPLimiter(ip)

	// Check per-IP connection rate
	if !limiter.ConnectionBucket.TryConsume(1) {
		rl.recordBlocked("connection", ip, "ip_limit")
		return false, "per-IP connection rate limit exceeded"
	}

	limiter.LastSeen = time.Now()
	limiter.ConnectionCount++

	rl.logger.Debug("connection allowed",
		"ip", ip,
		"count", limiter.ConnectionCount)

	return true, ""
}

// AllowMessage checks if a message is allowed from a peer.
func (rl *NetworkRateLimiter) AllowMessage(remoteAddr net.Addr, size int) (bool, string) {
	if !rl.config.Enabled {
		return true, ""
	}

	ip := extractIP(remoteAddr)
	if ip == "" {
		return false, "invalid address"
	}

	// Check whitelist
	rl.mu.RLock()
	if rl.whitelistedIPs[ip] {
		rl.mu.RUnlock()
		return true, ""
	}
	rl.mu.RUnlock()

	// Check global message rate
	if !rl.globalMsgBucket.TryConsume(1) {
		rl.recordBlocked("message", ip, "global_limit")
		return false, "global message rate limit exceeded"
	}

	// Get or create IP limiter
	limiter := rl.getOrCreateIPLimiter(ip)

	// Check per-IP message rate
	if !limiter.MessageBucket.TryConsume(1) {
		rl.recordBlocked("message", ip, "ip_limit")
		return false, "per-IP message rate limit exceeded"
	}

	// Check bandwidth limit
	bytesPerToken := float64(rl.config.BytesPerSecond) / float64(rl.config.MessagesPerSecond)
	tokensNeeded := float64(size) / bytesPerToken
	if !limiter.BandwidthBucket.TryConsume(tokensNeeded) {
		rl.recordBlocked("bandwidth", ip, "bandwidth_limit")
		return false, "bandwidth limit exceeded"
	}

	limiter.LastSeen = time.Now()

	return true, ""
}

// RecordBytes records bytes sent/received for bandwidth tracking.
func (rl *NetworkRateLimiter) RecordBytes(sent, received int64) {
	atomic.AddInt64(&rl.bytesSent, sent)
	atomic.AddInt64(&rl.bytesReceived, received)
}

// GetBandwidthStats returns current bandwidth statistics.
func (rl *NetworkRateLimiter) GetBandwidthStats() (sent, received int64) {
	return atomic.LoadInt64(&rl.bytesSent), atomic.LoadInt64(&rl.bytesReceived)
}

// UpdateSystemLoad updates the system load for adaptive rate limiting.
func (rl *NetworkRateLimiter) UpdateSystemLoad(cpuLoad, memoryLoad float64) {
	if !rl.config.AdaptiveEnabled {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Use the higher of CPU or memory load
	rl.systemLoad = cpuLoad
	if memoryLoad > cpuLoad {
		rl.systemLoad = memoryLoad
	}

	// Calculate adaptive multiplier
	if rl.systemLoad > rl.config.AdaptiveThreshold {
		// Reduce rate limits as load increases
		excess := rl.systemLoad - rl.config.AdaptiveThreshold
		reduction := excess / (1 - rl.config.AdaptiveThreshold) // 0 to 1
		rl.adaptiveMultiplier = 1 - (reduction * 0.5)          // Reduce by up to 50%
		if rl.adaptiveMultiplier < 0.25 {
			rl.adaptiveMultiplier = 0.25 // Never reduce by more than 75%
		}

		rl.logger.Info("adaptive rate limiting active",
			"load", rl.systemLoad,
			"multiplier", rl.adaptiveMultiplier)
	} else {
		rl.adaptiveMultiplier = 1.0
	}
}

// getOrCreateIPLimiter gets or creates a rate limiter for an IP.
func (rl *NetworkRateLimiter) getOrCreateIPLimiter(ip string) *IPLimiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, ok := rl.ipLimiters[ip]
	if !ok {
		// Apply adaptive multiplier
		connRate := float64(rl.config.ConnectionsPerMinutePerIP) / 60.0 * rl.adaptiveMultiplier
		msgRate := float64(rl.config.MessagesPerSecond) * rl.adaptiveMultiplier
		bwRate := float64(rl.config.BytesPerSecond) * rl.adaptiveMultiplier

		limiter = &IPLimiter{
			ConnectionBucket: NewTokenBucket(float64(rl.config.BurstSize), connRate),
			MessageBucket:    NewTokenBucket(float64(rl.config.BurstSize), msgRate),
			BandwidthBucket:  NewTokenBucket(float64(rl.config.BytesPerSecond), bwRate),
			LastSeen:         time.Now(),
		}
		rl.ipLimiters[ip] = limiter
	}

	return limiter
}

// CleanupStaleEntries removes IP limiters that haven't been seen recently.
func (rl *NetworkRateLimiter) CleanupStaleEntries(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, limiter := range rl.ipLimiters {
		if now.Sub(limiter.LastSeen) > maxAge {
			delete(rl.ipLimiters, ip)
		}
	}
}

// recordBlocked records a blocked request for metrics.
func (rl *NetworkRateLimiter) recordBlocked(requestType, ip, reason string) {
	telemetry.IncrCounter(1, "network", "ratelimit", "blocked", requestType)

	rl.mu.Lock()
	if limiter, ok := rl.ipLimiters[ip]; ok {
		limiter.BlockedCount++
	}
	rl.mu.Unlock()

	rl.logger.Warn("request blocked by rate limiter",
		"type", requestType,
		"ip", ip,
		"reason", reason)
}

// AddWhitelistedIP adds an IP to the whitelist.
func (rl *NetworkRateLimiter) AddWhitelistedIP(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.whitelistedIPs[ip] = true
	rl.logger.Info("IP added to whitelist", "ip", ip)
}

// RemoveWhitelistedIP removes an IP from the whitelist.
func (rl *NetworkRateLimiter) RemoveWhitelistedIP(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.whitelistedIPs, ip)
	rl.logger.Info("IP removed from whitelist", "ip", ip)
}

// GetStats returns current rate limiter statistics.
func (rl *NetworkRateLimiter) GetStats() NetworkRateLimiterStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	totalBlocked := 0
	for _, limiter := range rl.ipLimiters {
		totalBlocked += limiter.BlockedCount
	}

	return NetworkRateLimiterStats{
		ActiveIPs:          len(rl.ipLimiters),
		WhitelistedIPs:     len(rl.whitelistedIPs),
		TotalBlocked:       totalBlocked,
		AdaptiveMultiplier: rl.adaptiveMultiplier,
		SystemLoad:         rl.systemLoad,
		BytesSent:          atomic.LoadInt64(&rl.bytesSent),
		BytesReceived:      atomic.LoadInt64(&rl.bytesReceived),
	}
}

// NetworkRateLimiterStats contains rate limiter statistics.
type NetworkRateLimiterStats struct {
	ActiveIPs          int
	WhitelistedIPs     int
	TotalBlocked       int
	AdaptiveMultiplier float64
	SystemLoad         float64
	BytesSent          int64
	BytesReceived      int64
}

// extractIP extracts the IP address from a net.Addr.
func extractIP(addr net.Addr) string {
	if addr == nil {
		return ""
	}

	switch v := addr.(type) {
	case *net.TCPAddr:
		return v.IP.String()
	case *net.UDPAddr:
		return v.IP.String()
	default:
		// Try parsing as host:port
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return addr.String()
		}
		return host
	}
}

// extractSubnet extracts the /24 subnet from an IP address.
func extractSubnet(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}

	// Handle IPv4
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		return net.IPv4(ipv4[0], ipv4[1], ipv4[2], 0).String() + "/24"
	}

	// Handle IPv6 - use /48 for IPv6
	if len(parsedIP) == 16 {
		subnet := make(net.IP, 16)
		copy(subnet[:6], parsedIP[:6])
		return subnet.String() + "/48"
	}

	return ""
}

// BandwidthLimiter implements per-connection bandwidth limiting.
type BandwidthLimiter struct {
	bucket        *TokenBucket
	bytesPerToken float64
}

// NewBandwidthLimiter creates a bandwidth limiter with the given bytes per second limit.
func NewBandwidthLimiter(bytesPerSecond int64, burstSize int) *BandwidthLimiter {
	return &BandwidthLimiter{
		bucket:        NewTokenBucket(float64(burstSize), float64(bytesPerSecond)/1024), // tokens per KB
		bytesPerToken: 1024,
	}
}

// Allow checks if the given number of bytes can be sent/received.
func (bl *BandwidthLimiter) Allow(bytes int) bool {
	tokens := float64(bytes) / bl.bytesPerToken
	return bl.bucket.TryConsume(tokens)
}

// WaitForBandwidth waits until the given number of bytes can be sent/received.
func (bl *BandwidthLimiter) WaitForBandwidth(bytes int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	tokens := float64(bytes) / bl.bytesPerToken

	for time.Now().Before(deadline) {
		if bl.bucket.TryConsume(tokens) {
			return true
		}
		// Wait for some tokens to refill
		time.Sleep(10 * time.Millisecond)
	}

	return false
}
