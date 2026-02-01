package compatibility

import (
	"fmt"
	"sort"
)

// NegotiationResult represents the result of protocol version negotiation.
type NegotiationResult struct {
	// SelectedVersion is the negotiated version
	SelectedVersion string

	// ClientVersions are the versions supported by the client
	ClientVersions []string

	// ServerVersions are the versions supported by the server
	ServerVersions []string

	// Success indicates if negotiation was successful
	Success bool

	// Error contains the error message if negotiation failed
	Error string
}

// ProtocolVersionInfo describes a protocol version's support status.
type ProtocolVersionInfo struct {
	// Version is the protocol version identifier
	Version string `json:"version"`

	// Current indicates this is the current/recommended version
	Current bool `json:"current"`

	// Deprecated: indicates this version is deprecated
	Deprecated bool `json:"deprecated"`

	// DeprecationMessage contains the deprecation notice
	DeprecationMessage string `json:"deprecation_message,omitempty"`

	// Successor is the version to migrate to if deprecated
	Successor string `json:"successor,omitempty"`

	// MinClientVersion is the minimum client version supporting this protocol
	MinClientVersion string `json:"min_client_version,omitempty"`
}

// ProtocolNegotiator handles version negotiation for a specific protocol.
type ProtocolNegotiator struct {
	// Name is the protocol name
	Name string

	// CurrentVersion is the current/recommended version
	CurrentVersion string

	// SupportedVersions lists all supported versions in order of preference
	SupportedVersions []ProtocolVersionInfo

	// versionMap for quick lookup
	versionMap map[string]ProtocolVersionInfo
}

// NewProtocolNegotiator creates a new protocol negotiator.
func NewProtocolNegotiator(name, currentVersion string, supported []ProtocolVersionInfo) *ProtocolNegotiator {
	versionMap := make(map[string]ProtocolVersionInfo)
	for _, v := range supported {
		versionMap[v.Version] = v
	}

	return &ProtocolNegotiator{
		Name:              name,
		CurrentVersion:    currentVersion,
		SupportedVersions: supported,
		versionMap:        versionMap,
	}
}

// Negotiate performs version negotiation between client and server versions.
// It selects the highest mutually supported version.
func (n *ProtocolNegotiator) Negotiate(clientVersions []string) NegotiationResult {
	result := NegotiationResult{
		ClientVersions: clientVersions,
		ServerVersions: n.GetSupportedVersionStrings(),
		Success:        false,
	}

	if len(clientVersions) == 0 {
		result.Error = "client provided no versions"
		return result
	}

	// Find intersection of client and server versions
	common := n.findCommonVersions(clientVersions)
	if len(common) == 0 {
		result.Error = fmt.Sprintf("no common versions found between client %v and server %v",
			clientVersions, n.GetSupportedVersionStrings())
		return result
	}

	// Select highest common version (versions are sorted by preference)
	result.SelectedVersion = common[0]
	result.Success = true

	return result
}

// NegotiateRange performs version negotiation using a version range.
func (n *ProtocolNegotiator) NegotiateRange(minVersion, maxVersion string) NegotiationResult {
	result := NegotiationResult{
		ClientVersions: []string{fmt.Sprintf("%s-%s", minVersion, maxVersion)},
		ServerVersions: n.GetSupportedVersionStrings(),
		Success:        false,
	}

	// Find server versions within the range
	var compatible []string
	for _, v := range n.SupportedVersions {
		if n.versionInRange(v.Version, minVersion, maxVersion) {
			compatible = append(compatible, v.Version)
		}
	}

	if len(compatible) == 0 {
		result.Error = fmt.Sprintf("no server versions in range %s-%s", minVersion, maxVersion)
		return result
	}

	result.SelectedVersion = compatible[0]
	result.Success = true
	return result
}

// GetVersionInfo returns information about a specific version.
func (n *ProtocolNegotiator) GetVersionInfo(version string) (ProtocolVersionInfo, bool) {
	info, ok := n.versionMap[version]
	return info, ok
}

// IsSupported checks if a version is supported.
func (n *ProtocolNegotiator) IsSupported(version string) bool {
	_, ok := n.versionMap[version]
	return ok
}

// IsDeprecated checks if a version is deprecated.
func (n *ProtocolNegotiator) IsDeprecated(version string) bool {
	info, ok := n.versionMap[version]
	return ok && info.Deprecated
}

// GetDeprecationInfo returns deprecation information for a version.
func (n *ProtocolNegotiator) GetDeprecationInfo(version string) (message, successor string, isDeprecated bool) {
	info, ok := n.versionMap[version]
	if !ok || !info.Deprecated {
		return "", "", false
	}
	return info.DeprecationMessage, info.Successor, true
}

// GetSupportedVersionStrings returns all supported version strings.
func (n *ProtocolNegotiator) GetSupportedVersionStrings() []string {
	versions := make([]string, len(n.SupportedVersions))
	for i, v := range n.SupportedVersions {
		versions[i] = v.Version
	}
	return versions
}

// findCommonVersions finds common versions between client and server,
// sorted by server preference order.
func (n *ProtocolNegotiator) findCommonVersions(clientVersions []string) []string {
	clientSet := make(map[string]bool)
	for _, v := range clientVersions {
		clientSet[v] = true
	}

	var common []string
	for _, v := range n.SupportedVersions {
		if clientSet[v.Version] {
			common = append(common, v.Version)
		}
	}

	return common
}

// versionInRange checks if a version is within a min-max range.
// This is a simple string comparison for protocol versions.
func (n *ProtocolNegotiator) versionInRange(version, min, max string) bool {
	return version >= min && version <= max
}

// DefaultCaptureProtocolNegotiator returns the negotiator for the capture protocol.
func DefaultCaptureProtocolNegotiator() *ProtocolNegotiator {
	return NewProtocolNegotiator(
		"capture",
		"1",
		[]ProtocolVersionInfo{
			{
				Version: "1",
				Current: true,
			},
		},
	)
}

// DefaultProviderProtocolNegotiator returns the negotiator for the provider protocol.
func DefaultProviderProtocolNegotiator() *ProtocolNegotiator {
	return NewProtocolNegotiator(
		"provider",
		"2",
		[]ProtocolVersionInfo{
			{
				Version: "2",
				Current: true,
			},
			{
				Version:            "1",
				Deprecated:         true,
				DeprecationMessage: "Provider protocol v1 is deprecated. Upgrade to v2.",
				Successor:          "2",
			},
		},
	)
}

// DefaultManifestProtocolNegotiator returns the negotiator for the manifest protocol.
func DefaultManifestProtocolNegotiator() *ProtocolNegotiator {
	return NewProtocolNegotiator(
		"manifest",
		"2.1",
		[]ProtocolVersionInfo{
			{
				Version: "2.1",
				Current: true,
			},
			{
				Version:            "2.0",
				Deprecated:         true,
				DeprecationMessage: "Manifest v2.0 is deprecated. Use v2.1.",
				Successor:          "2.1",
			},
		},
	)
}

// NegotiationHeaders contains the standard headers for version negotiation.
type NegotiationHeaders struct {
	// RequestedVersions is sent by client indicating supported versions
	RequestedVersions string

	// SelectedVersion is sent by server indicating the selected version
	SelectedVersion string

	// SupportedVersions is sent by server indicating all supported versions
	SupportedVersions string

	// DeprecationWarning is sent by server if using deprecated version
	DeprecationWarning string
}

// StandardNegotiationHeaders returns the standard header names for a protocol.
func StandardNegotiationHeaders(protocol string) NegotiationHeaders {
	prefix := fmt.Sprintf("X-%s-", capitalize(protocol))
	return NegotiationHeaders{
		RequestedVersions:  prefix + "Protocol-Version",
		SelectedVersion:    prefix + "Protocol-Version",
		SupportedVersions:  prefix + "Supported-Versions",
		DeprecationWarning: prefix + "Deprecation-Warning",
	}
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return string(s[0]-32) + s[1:]
}

// MultiProtocolNegotiator manages multiple protocol negotiators.
type MultiProtocolNegotiator struct {
	negotiators map[string]*ProtocolNegotiator
}

// NewMultiProtocolNegotiator creates a new multi-protocol negotiator.
func NewMultiProtocolNegotiator() *MultiProtocolNegotiator {
	return &MultiProtocolNegotiator{
		negotiators: make(map[string]*ProtocolNegotiator),
	}
}

// Register registers a protocol negotiator.
func (m *MultiProtocolNegotiator) Register(negotiator *ProtocolNegotiator) {
	m.negotiators[negotiator.Name] = negotiator
}

// Get returns a negotiator by protocol name.
func (m *MultiProtocolNegotiator) Get(protocol string) (*ProtocolNegotiator, bool) {
	n, ok := m.negotiators[protocol]
	return n, ok
}

// Negotiate performs negotiation for a specific protocol.
func (m *MultiProtocolNegotiator) Negotiate(protocol string, clientVersions []string) NegotiationResult {
	n, ok := m.negotiators[protocol]
	if !ok {
		return NegotiationResult{
			Success: false,
			Error:   fmt.Sprintf("unknown protocol: %s", protocol),
		}
	}
	return n.Negotiate(clientVersions)
}

// GetProtocols returns all registered protocol names.
func (m *MultiProtocolNegotiator) GetProtocols() []string {
	protocols := make([]string, 0, len(m.negotiators))
	for name := range m.negotiators {
		protocols = append(protocols, name)
	}
	sort.Strings(protocols)
	return protocols
}

// DefaultMultiProtocolNegotiator returns a negotiator with all default protocols registered.
func DefaultMultiProtocolNegotiator() *MultiProtocolNegotiator {
	m := NewMultiProtocolNegotiator()
	m.Register(DefaultCaptureProtocolNegotiator())
	m.Register(DefaultProviderProtocolNegotiator())
	m.Register(DefaultManifestProtocolNegotiator())
	return m
}

