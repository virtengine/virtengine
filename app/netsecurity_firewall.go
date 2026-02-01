package app

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"cosmossdk.io/log"
)

// FirewallRuleGenerator generates firewall rules for various firewall types.
type FirewallRuleGenerator struct {
	config     FirewallConfig
	logger     log.Logger
	
	// Dynamic rules
	allowedIPs  map[string]time.Time // IP -> expiry (zero time means permanent)
	blockedIPs  map[string]time.Time // IP -> expiry
	allowedNets []*net.IPNet
	
	mu sync.RWMutex
}

// FirewallRule represents a single firewall rule.
type FirewallRule struct {
	Action    string // "allow", "deny", "drop"
	Direction string // "inbound", "outbound"
	Protocol  string // "tcp", "udp", "icmp", "all"
	Port      int    // 0 means all ports
	PortRange string // e.g., "8000:9000"
	SourceIP  string
	DestIP    string
	Comment   string
	Priority  int
}

// NewFirewallRuleGenerator creates a new firewall rule generator.
func NewFirewallRuleGenerator(config FirewallConfig, logger log.Logger) *FirewallRuleGenerator {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &FirewallRuleGenerator{
		config:     config,
		logger:     logger.With("module", "firewall"),
		allowedIPs: make(map[string]time.Time),
		blockedIPs: make(map[string]time.Time),
	}
}

// AddAllowedIP adds an IP to the allow list.
func (f *FirewallRuleGenerator) AddAllowedIP(ip string, duration time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var expiry time.Time
	if duration > 0 {
		expiry = time.Now().Add(duration)
	}
	f.allowedIPs[ip] = expiry
	f.logger.Info("IP added to firewall allow list", "ip", ip, "duration", duration)
}

// RemoveAllowedIP removes an IP from the allow list.
func (f *FirewallRuleGenerator) RemoveAllowedIP(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.allowedIPs, ip)
	f.logger.Info("IP removed from firewall allow list", "ip", ip)
}

// AddBlockedIP adds an IP to the block list.
func (f *FirewallRuleGenerator) AddBlockedIP(ip string, duration time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var expiry time.Time
	if duration > 0 {
		expiry = time.Now().Add(duration)
	}
	f.blockedIPs[ip] = expiry
	f.logger.Info("IP added to firewall block list", "ip", ip, "duration", duration)
}

// RemoveBlockedIP removes an IP from the block list.
func (f *FirewallRuleGenerator) RemoveBlockedIP(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.blockedIPs, ip)
	f.logger.Info("IP removed from firewall block list", "ip", ip)
}

// AddAllowedNetwork adds a network CIDR to the allow list.
func (f *FirewallRuleGenerator) AddAllowedNetwork(cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.allowedNets = append(f.allowedNets, ipNet)
	f.logger.Info("network added to firewall allow list", "cidr", cidr)
	return nil
}

// CleanupExpired removes expired entries.
func (f *FirewallRuleGenerator) CleanupExpired() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	
	for ip, expiry := range f.allowedIPs {
		if !expiry.IsZero() && now.After(expiry) {
			delete(f.allowedIPs, ip)
		}
	}
	
	for ip, expiry := range f.blockedIPs {
		if !expiry.IsZero() && now.After(expiry) {
			delete(f.blockedIPs, ip)
		}
	}
}

// GenerateRules generates firewall rules based on current configuration.
func (f *FirewallRuleGenerator) GenerateRules() []FirewallRule {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rules := make([]FirewallRule, 0, len(f.blockedIPs)+len(f.allowedIPs)*len(f.config.AllowedPorts))
	priority := 100

	// Add blocked IPs first (highest priority)
	now := time.Now()
	for ip, expiry := range f.blockedIPs {
		if expiry.IsZero() || now.Before(expiry) {
			rules = append(rules, FirewallRule{
				Action:    "drop",
				Direction: "inbound",
				Protocol:  "all",
				SourceIP:  ip,
				Comment:   "Blocked IP",
				Priority:  priority,
			})
			priority++
		}
	}

	// Add allowed IPs
	for ip, expiry := range f.allowedIPs {
		if expiry.IsZero() || now.Before(expiry) {
			for _, port := range f.config.AllowedPorts {
				rules = append(rules, FirewallRule{
					Action:    "allow",
					Direction: "inbound",
					Protocol:  "tcp",
					Port:      port,
					SourceIP:  ip,
					Comment:   "Allowed IP",
					Priority:  priority,
				})
			}
			priority++
		}
	}

	// Add allowed networks
	for _, ipNet := range f.allowedNets {
		for _, port := range f.config.AllowedPorts {
			rules = append(rules, FirewallRule{
				Action:    "allow",
				Direction: "inbound",
				Protocol:  "tcp",
				Port:      port,
				SourceIP:  ipNet.String(),
				Comment:   "Allowed network",
				Priority:  priority,
			})
		}
		priority++
	}

	// Add port-based allow rules for everyone
	for _, port := range f.config.AllowedPorts {
		rules = append(rules, FirewallRule{
			Action:    "allow",
			Direction: "inbound",
			Protocol:  "tcp",
			Port:      port,
			Comment:   fmt.Sprintf("VirtEngine port %d", port),
			Priority:  priority,
		})
		priority++
	}

	// Add default deny if enabled
	if f.config.DefaultDeny {
		rules = append(rules, FirewallRule{
			Action:    "drop",
			Direction: "inbound",
			Protocol:  "all",
			Comment:   "Default deny",
			Priority:  9999,
		})
	}

	// Sort by priority
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	return rules
}

// GenerateIPTables generates iptables rules.
func (f *FirewallRuleGenerator) GenerateIPTables() string {
	rules := f.GenerateRules()
	
	var buf bytes.Buffer
	buf.WriteString("# VirtEngine Firewall Rules - Generated\n")
	buf.WriteString("# Apply with: iptables-restore < this_file\n\n")
	buf.WriteString("*filter\n")
	buf.WriteString(":INPUT ACCEPT [0:0]\n")
	buf.WriteString(":FORWARD ACCEPT [0:0]\n")
	buf.WriteString(":OUTPUT ACCEPT [0:0]\n")
	buf.WriteString(":VIRTENGINE - [0:0]\n\n")
	
	// Jump to VIRTENGINE chain
	buf.WriteString("-A INPUT -j VIRTENGINE\n\n")
	
	// Flush existing rules
	buf.WriteString("-F VIRTENGINE\n")
	
	// Allow established connections
	buf.WriteString("-A VIRTENGINE -m state --state ESTABLISHED,RELATED -j ACCEPT\n")
	buf.WriteString("-A VIRTENGINE -i lo -j ACCEPT\n\n")
	
	for _, rule := range rules {
		buf.WriteString(f.ruleToIPTables(rule))
		buf.WriteString("\n")
	}
	
	buf.WriteString("\nCOMMIT\n")
	
	return buf.String()
}

func (f *FirewallRuleGenerator) ruleToIPTables(rule FirewallRule) string {
	var parts []string
	parts = append(parts, "-A VIRTENGINE")
	
	if rule.Protocol != "all" && rule.Protocol != "" {
		parts = append(parts, "-p", rule.Protocol)
	}
	
	if rule.SourceIP != "" {
		parts = append(parts, "-s", rule.SourceIP)
	}
	
	if rule.DestIP != "" {
		parts = append(parts, "-d", rule.DestIP)
	}
	
	if rule.Port > 0 {
		parts = append(parts, "--dport", fmt.Sprintf("%d", rule.Port))
	} else if rule.PortRange != "" {
		parts = append(parts, "--dport", rule.PortRange)
	}
	
	switch rule.Action {
	case "allow":
		parts = append(parts, "-j ACCEPT")
	case "deny":
		parts = append(parts, "-j REJECT")
	case "drop":
		parts = append(parts, "-j DROP")
	}
	
	if rule.Comment != "" {
		parts = append(parts, "-m comment --comment", fmt.Sprintf("%q", rule.Comment))
	}
	
	return strings.Join(parts, " ")
}

// GenerateNFTables generates nftables rules.
func (f *FirewallRuleGenerator) GenerateNFTables() string {
	rules := f.GenerateRules()
	
	tmpl := `#!/usr/sbin/nft -f
# VirtEngine Firewall Rules - Generated
# Apply with: nft -f this_file

flush ruleset

table inet virtengine {
    chain input {
        type filter hook input priority 0; policy drop;
        
        # Allow established connections
        ct state established,related accept
        
        # Allow loopback
        iifname "lo" accept
        
        # VirtEngine rules
{{range .}}        {{.}}
{{end}}
    }
    
    chain forward {
        type filter hook forward priority 0; policy drop;
    }
    
    chain output {
        type filter hook output priority 0; policy accept;
    }
}
`
	
	ruleStrings := make([]string, 0, len(rules))
	for _, rule := range rules {
		ruleStrings = append(ruleStrings, f.ruleToNFTables(rule))
	}
	
	t := template.Must(template.New("nftables").Parse(tmpl))
	var buf bytes.Buffer
	t.Execute(&buf, ruleStrings)
	
	return buf.String()
}

func (f *FirewallRuleGenerator) ruleToNFTables(rule FirewallRule) string {
	var parts []string
	
	if rule.SourceIP != "" {
		parts = append(parts, fmt.Sprintf("ip saddr %s", rule.SourceIP))
	}
	
	if rule.Protocol != "all" && rule.Protocol != "" {
		parts = append(parts, rule.Protocol)
	}
	
	if rule.Port > 0 {
		parts = append(parts, fmt.Sprintf("dport %d", rule.Port))
	} else if rule.PortRange != "" {
		parts = append(parts, fmt.Sprintf("dport %s", rule.PortRange))
	}
	
	switch rule.Action {
	case "allow":
		parts = append(parts, "accept")
	case "deny":
		parts = append(parts, "reject")
	case "drop":
		parts = append(parts, "drop")
	}
	
	if rule.Comment != "" {
		parts = append(parts, fmt.Sprintf("comment %q", rule.Comment))
	}
	
	return strings.Join(parts, " ")
}

// GeneratePF generates OpenBSD PF rules.
func (f *FirewallRuleGenerator) GeneratePF() string {
	rules := f.GenerateRules()
	
	var buf bytes.Buffer
	buf.WriteString("# VirtEngine Firewall Rules - Generated\n")
	buf.WriteString("# Apply with: pfctl -f this_file\n\n")
	
	// Macros
	buf.WriteString("# Macros\n")
	buf.WriteString(fmt.Sprintf("virtengine_ports = \"{ %s }\"\n\n", 
		strings.Trim(strings.ReplaceAll(fmt.Sprint(f.config.AllowedPorts), " ", ", "), "[]")))
	
	// Default policies
	buf.WriteString("# Default policies\n")
	buf.WriteString("set skip on lo0\n")
	buf.WriteString("block in all\n")
	buf.WriteString("pass out all keep state\n\n")
	
	// Rules
	buf.WriteString("# VirtEngine rules\n")
	for _, rule := range rules {
		buf.WriteString(f.ruleToPF(rule))
		buf.WriteString("\n")
	}
	
	return buf.String()
}

func (f *FirewallRuleGenerator) ruleToPF(rule FirewallRule) string {
	var action string
	switch rule.Action {
	case "allow":
		action = "pass"
	case "deny", "drop":
		action = "block"
	}
	
	var parts []string
	parts = append(parts, action, "in")
	
	if rule.Protocol != "all" && rule.Protocol != "" {
		parts = append(parts, "proto", rule.Protocol)
	}
	
	if rule.SourceIP != "" {
		parts = append(parts, "from", rule.SourceIP)
	} else {
		parts = append(parts, "from any")
	}
	
	if rule.Port > 0 {
		parts = append(parts, fmt.Sprintf("to any port %d", rule.Port))
	} else if rule.PortRange != "" {
		parts = append(parts, fmt.Sprintf("to any port %s", strings.Replace(rule.PortRange, "-", ":", 1)))
	}
	
	if rule.Action == "allow" {
		parts = append(parts, "keep state")
	}
	
	return strings.Join(parts, " ")
}

// GenerateWindowsFirewall generates Windows Firewall PowerShell commands.
func (f *FirewallRuleGenerator) GenerateWindowsFirewall() string {
	rules := f.GenerateRules()
	
	var buf bytes.Buffer
	buf.WriteString("# VirtEngine Firewall Rules - Generated\n")
	buf.WriteString("# Run as Administrator\n\n")
	
	// Remove existing VirtEngine rules
	buf.WriteString("# Remove existing VirtEngine rules\n")
	buf.WriteString("Get-NetFirewallRule -DisplayName \"VirtEngine*\" -ErrorAction SilentlyContinue | Remove-NetFirewallRule\n\n")
	
	ruleNum := 1
	for _, rule := range rules {
		if rule.Action == "allow" && rule.Port > 0 {
			buf.WriteString(fmt.Sprintf(
				"New-NetFirewallRule -DisplayName \"VirtEngine Port %d\" -Direction Inbound -Protocol TCP -LocalPort %d -Action Allow\n",
				rule.Port, rule.Port))
			ruleNum++
		}
	}
	
	// Block rules
	for _, rule := range rules {
		if rule.Action != "allow" && rule.SourceIP != "" {
			buf.WriteString(fmt.Sprintf(
				"New-NetFirewallRule -DisplayName \"VirtEngine Block %s\" -Direction Inbound -RemoteAddress %s -Action Block\n",
				rule.SourceIP, rule.SourceIP))
			ruleNum++
		}
	}
	
	return buf.String()
}

// Generate generates rules for the configured firewall type.
func (f *FirewallRuleGenerator) Generate() (string, error) {
	switch f.config.FirewallType {
	case "iptables":
		return f.GenerateIPTables(), nil
	case "nftables":
		return f.GenerateNFTables(), nil
	case "pf":
		return f.GeneratePF(), nil
	case "windows":
		return f.GenerateWindowsFirewall(), nil
	default:
		return "", fmt.Errorf("unsupported firewall type: %s", f.config.FirewallType)
	}
}

// GetBlockedIPs returns currently blocked IPs.
func (f *FirewallRuleGenerator) GetBlockedIPs() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	var ips []string
	for ip, expiry := range f.blockedIPs {
		if expiry.IsZero() || now.Before(expiry) {
			ips = append(ips, ip)
		}
	}
	return ips
}

// GetAllowedIPs returns currently allowed IPs.
func (f *FirewallRuleGenerator) GetAllowedIPs() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	var ips []string
	for ip, expiry := range f.allowedIPs {
		if expiry.IsZero() || now.Before(expiry) {
			ips = append(ips, ip)
		}
	}
	return ips
}
