package app

import (
	"errors"
	"time"
)

// NetworkSecurityConfig holds all network security configuration parameters.
type NetworkSecurityConfig struct {
	// Noise Protocol configuration
	Noise NoiseConfig `json:"noise" yaml:"noise"`

	// Peer management configuration
	Peer PeerConfig `json:"peer" yaml:"peer"`

	// Rate limiting configuration
	RateLimit NetworkRateLimitConfig `json:"rate_limit" yaml:"rate_limit"`

	// Protection configuration (Sybil, Eclipse, DDoS)
	Protection ProtectionConfig `json:"protection" yaml:"protection"`

	// Firewall configuration
	Firewall FirewallConfig `json:"firewall" yaml:"firewall"`

	// IDS configuration
	IDS IDSConfig `json:"ids" yaml:"ids"`
}

// NoiseConfig configures Noise Protocol encryption.
type NoiseConfig struct {
	// Enabled determines if Noise Protocol encryption is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// HandshakeTimeout is the maximum time allowed for handshake completion
	HandshakeTimeout time.Duration `json:"handshake_timeout" yaml:"handshake_timeout"`

	// StaticKeyPath is the path to the static private key file
	StaticKeyPath string `json:"static_key_path" yaml:"static_key_path"`

	// RequirePeerAuthentication requires peers to authenticate during handshake
	RequirePeerAuthentication bool `json:"require_peer_authentication" yaml:"require_peer_authentication"`

	// AllowedCipherSuites specifies which cipher suites are allowed
	// Supported: "ChaChaPoly", "AESGCM"
	AllowedCipherSuites []string `json:"allowed_cipher_suites" yaml:"allowed_cipher_suites"`
}

// PeerConfig configures peer authentication and authorization.
type PeerConfig struct {
	// Enabled determines if peer authentication is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// MaxPeers is the maximum number of connected peers
	MaxPeers int `json:"max_peers" yaml:"max_peers"`

	// MaxInboundPeers is the maximum number of inbound peer connections
	MaxInboundPeers int `json:"max_inbound_peers" yaml:"max_inbound_peers"`

	// MaxOutboundPeers is the maximum number of outbound peer connections
	MaxOutboundPeers int `json:"max_outbound_peers" yaml:"max_outbound_peers"`

	// TrustedPeers is a list of peer IDs that are always allowed
	TrustedPeers []string `json:"trusted_peers" yaml:"trusted_peers"`

	// BannedPeers is a list of peer IDs that are always rejected
	BannedPeers []string `json:"banned_peers" yaml:"banned_peers"`

	// MinStakeForTrust is the minimum stake required for peer trust score boost
	MinStakeForTrust int64 `json:"min_stake_for_trust" yaml:"min_stake_for_trust"`

	// PeerScoreThreshold is the minimum score for a peer to remain connected
	PeerScoreThreshold float64 `json:"peer_score_threshold" yaml:"peer_score_threshold"`

	// ConnectionTimeout is the timeout for establishing connections
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"`

	// HandshakeTimeout is the timeout for completing peer handshake
	HandshakeTimeout time.Duration `json:"handshake_timeout" yaml:"handshake_timeout"`

	// PingInterval is the interval for peer health checks
	PingInterval time.Duration `json:"ping_interval" yaml:"ping_interval"`
}

// NetworkRateLimitConfig configures network-layer rate limiting.
type NetworkRateLimitConfig struct {
	// Enabled determines if rate limiting is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ConnectionsPerSecond is the maximum new connections per second
	ConnectionsPerSecond int `json:"connections_per_second" yaml:"connections_per_second"`

	// ConnectionsPerMinutePerIP is the maximum connections per minute from a single IP
	ConnectionsPerMinutePerIP int `json:"connections_per_minute_per_ip" yaml:"connections_per_minute_per_ip"`

	// MessagesPerSecond is the maximum messages per second per connection
	MessagesPerSecond int `json:"messages_per_second" yaml:"messages_per_second"`

	// BytesPerSecond is the maximum bytes per second per connection (bandwidth limit)
	BytesPerSecond int64 `json:"bytes_per_second" yaml:"bytes_per_second"`

	// BurstSize is the token bucket burst size for rate limiting
	BurstSize int `json:"burst_size" yaml:"burst_size"`

	// AdaptiveEnabled enables adaptive rate limiting based on system load
	AdaptiveEnabled bool `json:"adaptive_enabled" yaml:"adaptive_enabled"`

	// AdaptiveThreshold is the CPU/memory threshold for triggering adaptive limiting
	AdaptiveThreshold float64 `json:"adaptive_threshold" yaml:"adaptive_threshold"`

	// WhitelistedIPs are IPs exempt from rate limiting
	WhitelistedIPs []string `json:"whitelisted_ips" yaml:"whitelisted_ips"`
}

// ProtectionConfig configures attack prevention mechanisms.
type ProtectionConfig struct {
	// Sybil attack prevention
	Sybil SybilProtectionConfig `json:"sybil" yaml:"sybil"`

	// Eclipse attack prevention
	Eclipse EclipseProtectionConfig `json:"eclipse" yaml:"eclipse"`

	// DDoS mitigation
	DDoS DDoSProtectionConfig `json:"ddos" yaml:"ddos"`
}

// SybilProtectionConfig configures Sybil attack prevention.
type SybilProtectionConfig struct {
	// Enabled determines if Sybil protection is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// MaxPeersPerSubnet limits peers from the same /24 subnet
	MaxPeersPerSubnet int `json:"max_peers_per_subnet" yaml:"max_peers_per_subnet"`

	// MaxPeersPerASN limits peers from the same Autonomous System
	MaxPeersPerASN int `json:"max_peers_per_asn" yaml:"max_peers_per_asn"`

	// RequireStakeForConnection requires minimum stake for peer connection
	RequireStakeForConnection bool `json:"require_stake_for_connection" yaml:"require_stake_for_connection"`

	// MinimumStake is the minimum stake required if RequireStakeForConnection is true
	MinimumStake int64 `json:"minimum_stake" yaml:"minimum_stake"`

	// DiversityRequirement is the minimum number of unique subnets required
	DiversityRequirement int `json:"diversity_requirement" yaml:"diversity_requirement"`
}

// EclipseProtectionConfig configures Eclipse attack prevention.
type EclipseProtectionConfig struct {
	// Enabled determines if Eclipse protection is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// OutboundOnlySlots is the number of slots reserved for outbound-only connections
	OutboundOnlySlots int `json:"outbound_only_slots" yaml:"outbound_only_slots"`

	// PeerRotationInterval is how often to rotate peer connections
	PeerRotationInterval time.Duration `json:"peer_rotation_interval" yaml:"peer_rotation_interval"`

	// SeedNodeRefreshInterval is how often to refresh seed node connections
	SeedNodeRefreshInterval time.Duration `json:"seed_node_refresh_interval" yaml:"seed_node_refresh_interval"`

	// AnchorConnections is the number of long-lived anchor connections to maintain
	AnchorConnections int `json:"anchor_connections" yaml:"anchor_connections"`

	// RandomSelectionRatio is the ratio of peers selected randomly vs by reputation
	RandomSelectionRatio float64 `json:"random_selection_ratio" yaml:"random_selection_ratio"`
}

// DDoSProtectionConfig configures DDoS mitigation.
type DDoSProtectionConfig struct {
	// Enabled determines if DDoS protection is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// SYNFloodThreshold is the threshold for SYN flood detection
	SYNFloodThreshold int `json:"syn_flood_threshold" yaml:"syn_flood_threshold"`

	// ConnectionFloodThreshold is the threshold for connection flood detection
	ConnectionFloodThreshold int `json:"connection_flood_threshold" yaml:"connection_flood_threshold"`

	// MessageFloodThreshold is the threshold for message flood detection
	MessageFloodThreshold int `json:"message_flood_threshold" yaml:"message_flood_threshold"`

	// BanDuration is how long to ban detected attackers
	BanDuration time.Duration `json:"ban_duration" yaml:"ban_duration"`

	// EnableBlackholeRouting enables automatic blackhole routing for detected attackers
	EnableBlackholeRouting bool `json:"enable_blackhole_routing" yaml:"enable_blackhole_routing"`

	// AlertThreshold is the threshold for triggering DDoS alerts
	AlertThreshold int `json:"alert_threshold" yaml:"alert_threshold"`
}

// FirewallConfig configures firewall rules generation.
type FirewallConfig struct {
	// Enabled determines if firewall integration is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RulesPath is the path where firewall rules are written
	RulesPath string `json:"rules_path" yaml:"rules_path"`

	// FirewallType is the type of firewall (iptables, nftables, pf, windows)
	FirewallType string `json:"firewall_type" yaml:"firewall_type"`

	// AutoApply automatically applies generated rules
	AutoApply bool `json:"auto_apply" yaml:"auto_apply"`

	// AllowedPorts are the ports to allow inbound traffic
	AllowedPorts []int `json:"allowed_ports" yaml:"allowed_ports"`

	// DefaultDeny enables default deny policy for unspecified traffic
	DefaultDeny bool `json:"default_deny" yaml:"default_deny"`
}

// IDSConfig configures Intrusion Detection System integration.
type IDSConfig struct {
	// Enabled determines if IDS integration is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// IDSType is the type of IDS to integrate with (suricata, snort, ossec, custom)
	IDSType string `json:"ids_type" yaml:"ids_type"`

	// AlertEndpoint is the endpoint to send IDS alerts
	AlertEndpoint string `json:"alert_endpoint" yaml:"alert_endpoint"`

	// LogPath is the path for IDS-compatible logging
	LogPath string `json:"log_path" yaml:"log_path"`

	// AlertLevel is the minimum level for alerting (low, medium, high, critical)
	AlertLevel string `json:"alert_level" yaml:"alert_level"`

	// EnableMetrics enables IDS metrics collection
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
}

// DefaultNetworkSecurityConfig returns a Config with secure default values.
func DefaultNetworkSecurityConfig() NetworkSecurityConfig {
	return NetworkSecurityConfig{
		Noise: NoiseConfig{
			Enabled:                   true,
			HandshakeTimeout:          10 * time.Second,
			StaticKeyPath:             "",
			RequirePeerAuthentication: true,
			AllowedCipherSuites:       []string{"ChaChaPoly"},
		},
		Peer: PeerConfig{
			Enabled:            true,
			MaxPeers:           50,
			MaxInboundPeers:    25,
			MaxOutboundPeers:   25,
			TrustedPeers:       []string{},
			BannedPeers:        []string{},
			MinStakeForTrust:   1000000, // 1 VIRT
			PeerScoreThreshold: -100,
			ConnectionTimeout:  30 * time.Second,
			HandshakeTimeout:   20 * time.Second,
			PingInterval:       30 * time.Second,
		},
		RateLimit: NetworkRateLimitConfig{
			Enabled:                   true,
			ConnectionsPerSecond:      10,
			ConnectionsPerMinutePerIP: 30,
			MessagesPerSecond:         100,
			BytesPerSecond:            10 * 1024 * 1024, // 10 MB/s
			BurstSize:                 50,
			AdaptiveEnabled:           true,
			AdaptiveThreshold:         0.8,
			WhitelistedIPs:            []string{},
		},
		Protection: ProtectionConfig{
			Sybil: SybilProtectionConfig{
				Enabled:                   true,
				MaxPeersPerSubnet:         3,
				MaxPeersPerASN:            10,
				RequireStakeForConnection: false,
				MinimumStake:              0,
				DiversityRequirement:      5,
			},
			Eclipse: EclipseProtectionConfig{
				Enabled:                 true,
				OutboundOnlySlots:       8,
				PeerRotationInterval:    1 * time.Hour,
				SeedNodeRefreshInterval: 24 * time.Hour,
				AnchorConnections:       4,
				RandomSelectionRatio:    0.3,
			},
			DDoS: DDoSProtectionConfig{
				Enabled:                  true,
				SYNFloodThreshold:        1000,
				ConnectionFloodThreshold: 100,
				MessageFloodThreshold:    500,
				BanDuration:              1 * time.Hour,
				EnableBlackholeRouting:   false,
				AlertThreshold:           50,
			},
		},
		Firewall: FirewallConfig{
			Enabled:      false,
			RulesPath:    "/etc/virtengine/firewall.rules",
			FirewallType: "iptables",
			AutoApply:    false,
			AllowedPorts: []int{26656, 26657, 1317, 9090}, // P2P, RPC, REST, gRPC
			DefaultDeny:  true,
		},
		IDS: IDSConfig{
			Enabled:       false,
			IDSType:       "custom",
			AlertEndpoint: "",
			LogPath:       "/var/log/virtengine/ids.log",
			AlertLevel:    "medium",
			EnableMetrics: true,
		},
	}
}

// Validate checks the configuration for errors.
func (c *NetworkSecurityConfig) Validate() error {
	if c.Noise.Enabled {
		if c.Noise.HandshakeTimeout <= 0 {
			return errors.New("noise handshake timeout must be positive")
		}
		if len(c.Noise.AllowedCipherSuites) == 0 {
			return errors.New("at least one cipher suite must be allowed")
		}
		for _, suite := range c.Noise.AllowedCipherSuites {
			if suite != "ChaChaPoly" && suite != "AESGCM" {
				return errors.New("invalid cipher suite: " + suite)
			}
		}
	}

	if c.Peer.Enabled {
		if c.Peer.MaxPeers <= 0 {
			return errors.New("max peers must be positive")
		}
		if c.Peer.MaxInboundPeers+c.Peer.MaxOutboundPeers > c.Peer.MaxPeers {
			return errors.New("inbound + outbound peers cannot exceed max peers")
		}
		if c.Peer.ConnectionTimeout <= 0 {
			return errors.New("connection timeout must be positive")
		}
	}

	if c.RateLimit.Enabled {
		if c.RateLimit.ConnectionsPerSecond <= 0 {
			return errors.New("connections per second must be positive")
		}
		if c.RateLimit.MessagesPerSecond <= 0 {
			return errors.New("messages per second must be positive")
		}
		if c.RateLimit.BytesPerSecond <= 0 {
			return errors.New("bytes per second must be positive")
		}
	}

	if c.Protection.Sybil.Enabled {
		if c.Protection.Sybil.MaxPeersPerSubnet <= 0 {
			return errors.New("max peers per subnet must be positive")
		}
	}

	if c.Protection.Eclipse.Enabled {
		if c.Protection.Eclipse.OutboundOnlySlots < 0 {
			return errors.New("outbound only slots cannot be negative")
		}
		if c.Protection.Eclipse.RandomSelectionRatio < 0 || c.Protection.Eclipse.RandomSelectionRatio > 1 {
			return errors.New("random selection ratio must be between 0 and 1")
		}
	}

	if c.Firewall.Enabled {
		validTypes := map[string]bool{"iptables": true, "nftables": true, "pf": true, "windows": true}
		if !validTypes[c.Firewall.FirewallType] {
			return errors.New("invalid firewall type: " + c.Firewall.FirewallType)
		}
	}

	if c.IDS.Enabled {
		validLevels := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
		if !validLevels[c.IDS.AlertLevel] {
			return errors.New("invalid IDS alert level: " + c.IDS.AlertLevel)
		}
	}

	return nil
}
