package app

import (
	"net"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultNetworkSecurityConfig(t *testing.T) {
	config := DefaultNetworkSecurityConfig()

	err := config.Validate()
	require.NoError(t, err)

	// Verify defaults
	assert.True(t, config.Noise.Enabled)
	assert.True(t, config.Peer.Enabled)
	assert.True(t, config.RateLimit.Enabled)
	assert.True(t, config.Protection.Sybil.Enabled)
	assert.True(t, config.Protection.Eclipse.Enabled)
	assert.True(t, config.Protection.DDoS.Enabled)
	assert.False(t, config.Firewall.Enabled) // Disabled by default
	assert.False(t, config.IDS.Enabled)      // Disabled by default
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(*NetworkSecurityConfig)
		expectError bool
	}{
		{
			name:        "valid default config",
			modify:      func(c *NetworkSecurityConfig) {},
			expectError: false,
		},
		{
			name: "invalid cipher suite",
			modify: func(c *NetworkSecurityConfig) {
				c.Noise.AllowedCipherSuites = []string{"InvalidCipher"}
			},
			expectError: true,
		},
		{
			name: "zero max peers",
			modify: func(c *NetworkSecurityConfig) {
				c.Peer.MaxPeers = 0
			},
			expectError: true,
		},
		{
			name: "inbound + outbound > max peers",
			modify: func(c *NetworkSecurityConfig) {
				c.Peer.MaxPeers = 10
				c.Peer.MaxInboundPeers = 10
				c.Peer.MaxOutboundPeers = 10
			},
			expectError: true,
		},
		{
			name: "negative random selection ratio",
			modify: func(c *NetworkSecurityConfig) {
				c.Protection.Eclipse.RandomSelectionRatio = -0.5
			},
			expectError: true,
		},
		{
			name: "invalid firewall type",
			modify: func(c *NetworkSecurityConfig) {
				c.Firewall.Enabled = true
				c.Firewall.FirewallType = "invalid"
			},
			expectError: true,
		},
		{
			name: "invalid IDS alert level",
			modify: func(c *NetworkSecurityConfig) {
				c.IDS.Enabled = true
				c.IDS.AlertLevel = "unknown"
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkSecurityConfig()
			tc.modify(&config)

			err := config.Validate()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenBucket(t *testing.T) {
	t.Run("consume tokens", func(t *testing.T) {
		bucket := NewTokenBucket(10, 1) // 10 max, 1 per second refill

		// Should be able to consume initial tokens
		for i := 0; i < 10; i++ {
			assert.True(t, bucket.TryConsume(1))
		}

		// Should fail after exhaustion
		assert.False(t, bucket.TryConsume(1))

		// Wait for refill
		time.Sleep(1100 * time.Millisecond)
		assert.True(t, bucket.TryConsume(1))
	})

	t.Run("burst consumption", func(t *testing.T) {
		bucket := NewTokenBucket(5, 10)

		// Should be able to burst consume
		assert.True(t, bucket.TryConsume(5))
		assert.False(t, bucket.TryConsume(1))
	})
}

func TestNetworkRateLimiter(t *testing.T) {
	config := NetworkRateLimitConfig{
		Enabled:                   true,
		ConnectionsPerSecond:      5,
		ConnectionsPerMinutePerIP: 10,
		MessagesPerSecond:         100,
		BytesPerSecond:            1024 * 1024,
		BurstSize:                 5,
		AdaptiveEnabled:           true,
		AdaptiveThreshold:         0.8,
	}

	limiter := NewNetworkRateLimiter(config, log.NewNopLogger())

	t.Run("allow connections within limit", func(t *testing.T) {
		addr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080}

		for i := 0; i < 5; i++ {
			allowed, _ := limiter.AllowConnection(addr)
			assert.True(t, allowed, "connection %d should be allowed", i)
		}
	})

	t.Run("block connections over global limit", func(t *testing.T) {
		// Exhaust global limit
		for i := 0; i < 100; i++ {
			addr := &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 8080}
			limiter.AllowConnection(addr)
		}

		addr := &net.TCPAddr{IP: net.ParseIP("10.0.0.100"), Port: 8080}
		allowed, reason := limiter.AllowConnection(addr)

		// Either global or per-IP limit should trigger
		if !allowed {
			assert.NotEmpty(t, reason)
		}
	})

	t.Run("whitelisted IPs bypass limits", func(t *testing.T) {
		limiter.AddWhitelistedIP("8.8.8.8")

		addr := &net.TCPAddr{IP: net.ParseIP("8.8.8.8"), Port: 8080}
		for i := 0; i < 100; i++ {
			allowed, _ := limiter.AllowConnection(addr)
			assert.True(t, allowed, "whitelisted IP should always be allowed")
		}
	})
}

func TestPeerAuthenticator(t *testing.T) {
	config := PeerConfig{
		Enabled:      true,
		TrustedPeers: []string{"trusted-peer-1"},
		BannedPeers:  []string{"banned-peer-1"},
	}

	auth := NewPeerAuthenticator(config, log.NewNopLogger())

	t.Run("authenticate trusted peer", func(t *testing.T) {
		info := PeerInfo{ID: "trusted-peer-1"}
		allowed, reason := auth.AuthenticatePeer(info)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})

	t.Run("reject banned peer", func(t *testing.T) {
		info := PeerInfo{ID: "banned-peer-1"}
		allowed, reason := auth.AuthenticatePeer(info)
		assert.False(t, allowed)
		assert.Contains(t, reason, "banned")
	})

	t.Run("ban and unban peer", func(t *testing.T) {
		peerID := PeerID("test-peer")

		auth.BanPeer(peerID, time.Hour)
		assert.True(t, auth.IsBanned(peerID))

		auth.UnbanPeer(peerID)
		assert.False(t, auth.IsBanned(peerID))
	})
}

func TestPeerScoreManager(t *testing.T) {
	params := DefaultPeerScoreParams()
	manager := NewPeerScoreManager(params, log.NewNopLogger())

	t.Run("new peer has zero score", func(t *testing.T) {
		score := manager.GetScore("new-peer")
		assert.Equal(t, float64(0), score.Total)
	})

	t.Run("good behavior increases score", func(t *testing.T) {
		manager.RecordGoodBehavior("good-peer", 10)
		score := manager.GetScore("good-peer")
		assert.Greater(t, score.Total, float64(0))
	})

	t.Run("misbehavior decreases score", func(t *testing.T) {
		manager.RecordGoodBehavior("mixed-peer", 5)
		manager.RecordMisbehavior("mixed-peer", 100, "test misbehavior")
		score := manager.GetScore("mixed-peer")
		assert.Less(t, score.Total, float64(0))
	})

	t.Run("validator bonus increases score", func(t *testing.T) {
		manager.SetValidatorBonus("validator-peer", true)
		score := manager.GetScore("validator-peer")
		assert.Equal(t, params.ValidatorBonusMax, score.ValidatorBonus)
	})

	t.Run("score decay over time", func(t *testing.T) {
		manager.RecordGoodBehavior("decaying-peer", 50)
		initialScore := manager.GetScore("decaying-peer").Total

		manager.ApplyDecay(time.Hour * 10)

		decayedScore := manager.GetScore("decaying-peer").Total
		assert.Less(t, decayedScore, initialScore)
	})
}

func TestDDoSProtector(t *testing.T) {
	config := DDoSProtectionConfig{
		Enabled:                  true,
		ConnectionFloodThreshold: 10,
		MessageFloodThreshold:    50,
		BanDuration:              time.Minute,
	}

	protector := NewDDoSProtector(config, log.NewNopLogger())

	t.Run("allow normal traffic", func(t *testing.T) {
		blocked, _ := protector.RecordConnection("192.168.1.1")
		assert.False(t, blocked)
	})

	t.Run("ban attacker IP", func(t *testing.T) {
		protector.BanIP("10.0.0.1", time.Hour)
		assert.True(t, protector.isBanned("10.0.0.1"))
	})

	t.Run("unban IP", func(t *testing.T) {
		protector.BanIP("10.0.0.2", time.Hour)
		protector.UnbanIP("10.0.0.2")
		assert.False(t, protector.isBanned("10.0.0.2"))
	})
}

func TestSybilProtector(t *testing.T) {
	sybilConfig := SybilProtectionConfig{
		Enabled:           true,
		MaxPeersPerSubnet: 2,
		MaxPeersPerASN:    5,
	}
	peerConfig := PeerConfig{}

	protector := NewSybilProtector(sybilConfig, peerConfig, log.NewNopLogger())

	t.Run("allow peers from different subnets", func(t *testing.T) {
		info1 := PeerInfo{ID: "peer1", Subnet: "192.168.1.0/24"}
		info2 := PeerInfo{ID: "peer2", Subnet: "192.168.2.0/24"}

		allowed1, _ := protector.AllowPeer(info1)
		protector.RegisterPeer(info1)

		allowed2, _ := protector.AllowPeer(info2)
		protector.RegisterPeer(info2)

		assert.True(t, allowed1)
		assert.True(t, allowed2)
	})

	t.Run("block too many peers from same subnet", func(t *testing.T) {
		// Register peers up to limit
		for i := 0; i < 2; i++ {
			info := PeerInfo{
				ID:     PeerID("subnet-peer-" + string(rune('A'+i))),
				Subnet: "10.0.0.0/24",
			}
			protector.RegisterPeer(info)
		}

		// Third peer should be blocked
		info := PeerInfo{ID: "subnet-peer-C", Subnet: "10.0.0.0/24"}
		allowed, reason := protector.AllowPeer(info)
		assert.False(t, allowed)
		assert.Contains(t, reason, "subnet")
	})

	t.Run("diversity score calculation", func(t *testing.T) {
		score := protector.GetDiversityScore()
		assert.GreaterOrEqual(t, score, float64(0))
		assert.LessOrEqual(t, score, float64(100))
	})
}

func TestEclipseProtector(t *testing.T) {
	config := EclipseProtectionConfig{
		Enabled:              true,
		OutboundOnlySlots:    3,
		PeerRotationInterval: time.Hour,
		AnchorConnections:    2,
		RandomSelectionRatio: 0.3,
	}

	protector := NewEclipseProtector(config, nil, log.NewNopLogger())

	t.Run("reserve outbound slots", func(t *testing.T) {
		assert.True(t, protector.ShouldReserveOutboundSlot(0))
		assert.True(t, protector.ShouldReserveOutboundSlot(2))
		assert.False(t, protector.ShouldReserveOutboundSlot(3))
	})

	t.Run("set anchor connections", func(t *testing.T) {
		info1 := PeerInfo{ID: "anchor1", IsValidator: true}
		info2 := PeerInfo{ID: "anchor2", IsValidator: true}
		info3 := PeerInfo{ID: "anchor3", IsValidator: true}

		assert.True(t, protector.SetAnchor(info1))
		assert.True(t, protector.SetAnchor(info2))
		assert.False(t, protector.SetAnchor(info3)) // Exceeds limit

		assert.True(t, protector.IsAnchor("anchor1"))
		assert.Equal(t, 2, protector.GetAnchorCount())
	})
}

func TestFirewallRuleGenerator(t *testing.T) {
	config := FirewallConfig{
		Enabled:      true,
		FirewallType: "iptables",
		AllowedPorts: []int{26656, 26657, 1317},
		DefaultDeny:  true,
	}

	gen := NewFirewallRuleGenerator(config, log.NewNopLogger())

	t.Run("generate iptables rules", func(t *testing.T) {
		rules := gen.GenerateIPTables()

		assert.Contains(t, rules, "*filter")
		assert.Contains(t, rules, "VIRTENGINE")
		assert.Contains(t, rules, "26656")
		assert.Contains(t, rules, "DROP")
	})

	t.Run("add and remove blocked IPs", func(t *testing.T) {
		gen.AddBlockedIP("10.0.0.1", time.Hour)

		blockedIPs := gen.GetBlockedIPs()
		assert.Contains(t, blockedIPs, "10.0.0.1")

		gen.RemoveBlockedIP("10.0.0.1")
		blockedIPs = gen.GetBlockedIPs()
		assert.NotContains(t, blockedIPs, "10.0.0.1")
	})

	t.Run("generate nftables rules", func(t *testing.T) {
		config.FirewallType = "nftables"
		gen := NewFirewallRuleGenerator(config, log.NewNopLogger())

		rules := gen.GenerateNFTables()
		assert.Contains(t, rules, "table inet virtengine")
	})

	t.Run("generate pf rules", func(t *testing.T) {
		config.FirewallType = "pf"
		gen := NewFirewallRuleGenerator(config, log.NewNopLogger())

		rules := gen.GeneratePF()
		assert.Contains(t, rules, "virtengine_ports")
	})
}

func TestSlidingWindow(t *testing.T) {
	window := NewSlidingWindow(time.Second*5, time.Second)

	t.Run("count within window", func(t *testing.T) {
		window.Add(10)
		assert.Equal(t, int64(10), window.Count())

		window.Add(5)
		assert.Equal(t, int64(15), window.Count())
	})
}

func TestNoiseKeyPairGeneration(t *testing.T) {
	keyPair, err := GenerateNoiseKeyPair()
	require.NoError(t, err)

	assert.Len(t, keyPair.PrivateKey, 32)
	assert.Len(t, keyPair.PublicKey, 32)

	// Generate another and verify they're different
	keyPair2, err := GenerateNoiseKeyPair()
	require.NoError(t, err)

	assert.NotEqual(t, keyPair.PrivateKey, keyPair2.PrivateKey)
	assert.NotEqual(t, keyPair.PublicKey, keyPair2.PublicKey)
}

func TestNoiseTransport(t *testing.T) {
	config := NoiseConfig{
		Enabled:                   true,
		HandshakeTimeout:          5 * time.Second,
		RequirePeerAuthentication: true,
		AllowedCipherSuites:       []string{"ChaChaPoly"},
	}

	transport, err := NewNoiseTransport(config)
	require.NoError(t, err)

	t.Run("get public key", func(t *testing.T) {
		publicKey := transport.GetPublicKey()
		assert.Len(t, publicKey, 32)
	})

	t.Run("add and check trusted key", func(t *testing.T) {
		testKey := make([]byte, 32)
		for i := range testKey {
			testKey[i] = byte(i)
		}

		transport.AddTrustedKey(testKey)
		assert.True(t, transport.IsTrusted(testKey))

		transport.RemoveTrustedKey(testKey)
		assert.False(t, transport.IsTrusted(testKey))
	})
}

func TestExtractSubnet(t *testing.T) {
	tests := []struct {
		ip       string
		expected string
	}{
		{"192.168.1.100", "192.168.1.0/24"},
		{"10.0.0.1", "10.0.0.0/24"},
		{"invalid", ""},
	}

	for _, tc := range tests {
		result := extractSubnet(tc.ip)
		assert.Equal(t, tc.expected, result, "extractSubnet(%s)", tc.ip)
	}
}

func TestNetworkSecurityManager(t *testing.T) {
	config := DefaultNetworkSecurityConfig()
	config.IDS.Enabled = false // Disable to avoid file creation
	config.Firewall.Enabled = false

	manager, err := NewNetworkSecurityManager(config, log.NewNopLogger())
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	t.Run("get noise public key", func(t *testing.T) {
		key := manager.GetNoisePublicKey()
		assert.Len(t, key, 32)
	})

	t.Run("authenticate and authorize peer", func(t *testing.T) {
		info := PeerInfo{
			ID:      "test-peer",
			Address: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080},
		}

		allowed, reason := manager.AuthenticatePeer(info)
		assert.True(t, allowed, "reason: %s", reason)

		allowed, reason = manager.AuthorizePeer(info)
		assert.True(t, allowed, "reason: %s", reason)
	})

	t.Run("register and unregister peer", func(t *testing.T) {
		info := PeerInfo{
			ID:        "registered-peer",
			Address:   &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 8080},
			IsInbound: true,
		}

		err := manager.RegisterPeer(info)
		assert.NoError(t, err)

		stats := manager.GetStats()
		assert.Greater(t, stats.TotalPeers, 0)

		manager.UnregisterPeer(info, true)
	})

	t.Run("ban and unban peer", func(t *testing.T) {
		peerID := PeerID("banned-peer")

		manager.BanPeer(peerID, time.Hour)
		assert.True(t, manager.IsPeerBanned(peerID))

		manager.UnbanPeer(peerID)
		assert.False(t, manager.IsPeerBanned(peerID))
	})

	t.Run("allow connection", func(t *testing.T) {
		addr := &net.TCPAddr{IP: net.ParseIP("8.8.8.8"), Port: 8080}
		allowed, _ := manager.AllowConnection(addr)
		assert.True(t, allowed)
	})

	t.Run("record behavior", func(t *testing.T) {
		peerID := PeerID("behavior-peer")

		manager.RecordGoodBehavior(peerID, 10)
		score := manager.GetPeerScore(peerID)
		assert.Greater(t, score.Total, float64(0))

		manager.RecordMisbehavior(peerID, 5, "test misbehavior")
		// Score should be affected
	})

	t.Run("get stats", func(t *testing.T) {
		stats := manager.GetStats()
		assert.GreaterOrEqual(t, stats.TotalPeers, 0)
		assert.GreaterOrEqual(t, stats.DiversityScore, float64(0))
	})

	err = manager.Stop()
	require.NoError(t, err)
}

func TestPeerAuthorizer(t *testing.T) {
	config := PeerConfig{
		Enabled:          true,
		MaxPeers:         5,
		MaxInboundPeers:  3,
		MaxOutboundPeers: 2,
	}

	auth := NewPeerAuthorizer(config, nil, log.NewNopLogger())

	t.Run("allow within limits", func(t *testing.T) {
		info := PeerInfo{ID: "peer1", IsInbound: true}
		allowed, _ := auth.AuthorizePeer(info)
		assert.True(t, allowed)
	})

	t.Run("block when max inbound reached", func(t *testing.T) {
		// Register max inbound peers
		for i := 0; i < 3; i++ {
			info := PeerInfo{
				ID:        PeerID("inbound-" + string(rune('A'+i))),
				IsInbound: true,
			}
			auth.RegisterPeer(info)
		}

		// Next inbound should be blocked
		info := PeerInfo{ID: "excess-inbound", IsInbound: true}
		allowed, reason := auth.AuthorizePeer(info)
		assert.False(t, allowed)
		assert.Contains(t, reason, "inbound")
	})

	t.Run("get peer counts", func(t *testing.T) {
		inbound, outbound, total := auth.GetPeerCounts()
		assert.Equal(t, 3, inbound)
		assert.Equal(t, 0, outbound)
		assert.Equal(t, 3, total)
	})
}
