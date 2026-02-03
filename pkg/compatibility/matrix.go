package compatibility

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// CompatibilityMatrix defines version compatibility between different components.
type CompatibilityMatrix struct {
	mu sync.RWMutex

	// ServerVersions maps server versions to their compatibility info
	ServerVersions map[string]ServerCompatibility

	// ClientVersions maps client versions to their compatibility info
	ClientVersions map[string]ClientCompatibility

	// ModuleVersions maps module names to their version matrices
	ModuleVersions map[string]ModuleCompatibilityMatrix

	// UpdatedAt is when the matrix was last updated
	UpdatedAt time.Time
}

// ServerCompatibility describes what a server version is compatible with.
type ServerCompatibility struct {
	// Version is the server version
	Version string `json:"version"`

	// MinClientVersion is the minimum compatible client version
	MinClientVersion string `json:"min_client_version"`

	// MaxClientVersion is the maximum compatible client version (optional)
	MaxClientVersion string `json:"max_client_version,omitempty"`

	// SupportedProtocols maps protocol names to supported versions
	SupportedProtocols map[string][]string `json:"supported_protocols"`

	// SupportedModuleVersions maps module names to supported API versions
	SupportedModuleVersions map[string][]string `json:"supported_module_versions"`

	// ReleaseDate is when this version was released
	ReleaseDate time.Time `json:"release_date"`

	// Status is the support status
	Status SupportLevel `json:"status"`

	// Notes contains any compatibility notes
	Notes string `json:"notes,omitempty"`
}

// ClientCompatibility describes what a client version is compatible with.
type ClientCompatibility struct {
	// Version is the client version
	Version string `json:"version"`

	// MinServerVersion is the minimum compatible server version
	MinServerVersion string `json:"min_server_version"`

	// MaxServerVersion is the maximum compatible server version (optional)
	MaxServerVersion string `json:"max_server_version,omitempty"`

	// SupportsProtocols maps protocol names to supported versions
	SupportsProtocols map[string][]string `json:"supports_protocols"`

	// SupportsModuleVersions maps module names to supported API versions
	SupportsModuleVersions map[string][]string `json:"supports_module_versions"`

	// ReleaseDate is when this version was released
	ReleaseDate time.Time `json:"release_date"`

	// Status is the support status
	Status SupportLevel `json:"status"`
}

// ModuleCompatibilityMatrix describes compatibility between module API versions.
type ModuleCompatibilityMatrix struct {
	// ModuleName is the module name
	ModuleName string `json:"module_name"`

	// CurrentVersion is the current/recommended version
	CurrentVersion string `json:"current_version"`

	// Versions lists all versions with their compatibility info
	Versions []ModuleVersionCompatibility `json:"versions"`
}

// ModuleVersionCompatibility describes a single module version's compatibility.
type ModuleVersionCompatibility struct {
	// Version is the API version
	Version string `json:"version"`

	// Status is the support status
	Status SupportLevel `json:"status"`

	// IntroducedIn is the server version this was introduced
	IntroducedIn string `json:"introduced_in"`

	// DeprecatedIn is the server version this was deprecated (if applicable)
	DeprecatedIn string `json:"deprecated_in,omitempty"`

	// RemovedIn is the server version this was removed (if applicable)
	RemovedIn string `json:"removed_in,omitempty"`

	// CompatibleWith lists other API versions it's compatible with
	CompatibleWith []string `json:"compatible_with,omitempty"`

	// MigrationGuide is a link to the migration guide
	MigrationGuide string `json:"migration_guide,omitempty"`
}

// NewCompatibilityMatrix creates a new compatibility matrix.
func NewCompatibilityMatrix() *CompatibilityMatrix {
	return &CompatibilityMatrix{
		ServerVersions: make(map[string]ServerCompatibility),
		ClientVersions: make(map[string]ClientCompatibility),
		ModuleVersions: make(map[string]ModuleCompatibilityMatrix),
		UpdatedAt:      time.Now(),
	}
}

// AddServerVersion adds a server version to the matrix.
func (m *CompatibilityMatrix) AddServerVersion(compat ServerCompatibility) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ServerVersions[compat.Version] = compat
	m.UpdatedAt = time.Now()
}

// AddClientVersion adds a client version to the matrix.
func (m *CompatibilityMatrix) AddClientVersion(compat ClientCompatibility) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ClientVersions[compat.Version] = compat
	m.UpdatedAt = time.Now()
}

// AddModuleMatrix adds a module's compatibility matrix.
func (m *CompatibilityMatrix) AddModuleMatrix(matrix ModuleCompatibilityMatrix) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ModuleVersions[matrix.ModuleName] = matrix
	m.UpdatedAt = time.Now()
}

// CheckServerClientCompatibility checks if a server and client version are compatible.
func (m *CompatibilityMatrix) CheckServerClientCompatibility(serverVersion, clientVersion string) (bool, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var warnings []string

	serverCompat, ok := m.ServerVersions[serverVersion]
	if !ok {
		return false, []string{fmt.Sprintf("Unknown server version: %s", serverVersion)}
	}

	clientCompat, ok := m.ClientVersions[clientVersion]
	if !ok {
		return false, []string{fmt.Sprintf("Unknown client version: %s", clientVersion)}
	}

	// Check server's client requirements
	clientV := MustParseVersion(clientVersion)
	minClientV := MustParseVersion(serverCompat.MinClientVersion)

	if clientV.LessThan(minClientV) {
		return false, []string{fmt.Sprintf("Client %s is below minimum required %s for server %s",
			clientVersion, serverCompat.MinClientVersion, serverVersion)}
	}

	if serverCompat.MaxClientVersion != "" {
		maxClientV := MustParseVersion(serverCompat.MaxClientVersion)
		if clientV.GreaterThan(maxClientV) {
			warnings = append(warnings, fmt.Sprintf("Client %s may have features not supported by server %s",
				clientVersion, serverVersion))
		}
	}

	// Check client's server requirements
	serverV := MustParseVersion(serverVersion)
	minServerV := MustParseVersion(clientCompat.MinServerVersion)

	if serverV.LessThan(minServerV) {
		return false, []string{fmt.Sprintf("Server %s is below minimum required %s for client %s",
			serverVersion, clientCompat.MinServerVersion, clientVersion)}
	}

	// Check support status
	if serverCompat.Status == SupportLevelEOL || clientCompat.Status == SupportLevelEOL {
		return false, []string{"One or both versions are end-of-life"}
	}

	if serverCompat.Status == SupportLevelDeprecated {
		warnings = append(warnings, fmt.Sprintf("Server version %s is deprecated", serverVersion))
	}
	if clientCompat.Status == SupportLevelDeprecated {
		warnings = append(warnings, fmt.Sprintf("Client version %s is deprecated", clientVersion))
	}

	return true, warnings
}

// GetCompatibleClientVersions returns client versions compatible with a server version.
func (m *CompatibilityMatrix) GetCompatibleClientVersions(serverVersion string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serverCompat, ok := m.ServerVersions[serverVersion]
	if !ok {
		return nil
	}

	var compatible []string
	minClientV := MustParseVersion(serverCompat.MinClientVersion)

	for version, clientCompat := range m.ClientVersions {
		clientV := MustParseVersion(version)

		if clientV.GreaterThanOrEqual(minClientV) && clientCompat.Status != SupportLevelEOL {
			// Check if client supports this server version
			minServerV := MustParseVersion(clientCompat.MinServerVersion)
			serverV := MustParseVersion(serverVersion)
			if serverV.GreaterThanOrEqual(minServerV) {
				compatible = append(compatible, version)
			}
		}
	}

	sort.Slice(compatible, func(i, j int) bool {
		vi := MustParseVersion(compatible[i])
		vj := MustParseVersion(compatible[j])
		return vi.GreaterThan(vj) // Descending order
	})

	return compatible
}

// GetCompatibleServerVersions returns server versions compatible with a client version.
func (m *CompatibilityMatrix) GetCompatibleServerVersions(clientVersion string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clientCompat, ok := m.ClientVersions[clientVersion]
	if !ok {
		return nil
	}

	var compatible []string
	minServerV := MustParseVersion(clientCompat.MinServerVersion)

	for version, serverCompat := range m.ServerVersions {
		serverV := MustParseVersion(version)

		if serverV.GreaterThanOrEqual(minServerV) && serverCompat.Status != SupportLevelEOL {
			// Check if server supports this client version
			minClientV := MustParseVersion(serverCompat.MinClientVersion)
			clientV := MustParseVersion(clientVersion)
			if clientV.GreaterThanOrEqual(minClientV) {
				compatible = append(compatible, version)
			}
		}
	}

	sort.Slice(compatible, func(i, j int) bool {
		vi := MustParseVersion(compatible[i])
		vj := MustParseVersion(compatible[j])
		return vi.GreaterThan(vj) // Descending order
	})

	return compatible
}

// GetModuleVersionStatus returns the status of a module API version.
func (m *CompatibilityMatrix) GetModuleVersionStatus(moduleName, apiVersion string) (SupportLevel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	moduleMatrix, ok := m.ModuleVersions[moduleName]
	if !ok {
		return "", fmt.Errorf("unknown module: %s", moduleName)
	}

	for _, v := range moduleMatrix.Versions {
		if v.Version == apiVersion {
			return v.Status, nil
		}
	}

	return "", fmt.Errorf("unknown API version %s for module %s", apiVersion, moduleName)
}

// DefaultCompatibilityMatrix returns a matrix with default VirtEngine versions.
func DefaultCompatibilityMatrix() *CompatibilityMatrix {
	matrix := NewCompatibilityMatrix()

	// Server versions
	matrix.AddServerVersion(ServerCompatibility{
		Version:          "v0.10.0",
		MinClientVersion: "v0.8.0",
		SupportedProtocols: map[string][]string{
			"capture":  {"1"},
			"provider": {"1", "2"},
			"manifest": {"2.0", "2.1"},
		},
		SupportedModuleVersions: map[string][]string{
			"veid":       {"v1"},
			"market":     {"v1beta4", "v1beta5"},
			"deployment": {"v1beta3", "v1beta4"},
			"provider":   {"v1beta3", "v1beta4"},
		},
		Status: SupportLevelActive,
	})

	matrix.AddServerVersion(ServerCompatibility{
		Version:          "v0.9.0",
		MinClientVersion: "v0.8.0",
		SupportedProtocols: map[string][]string{
			"capture":  {"1"},
			"provider": {"1", "2"},
			"manifest": {"2.0", "2.1"},
		},
		SupportedModuleVersions: map[string][]string{
			"veid":       {"v1"},
			"market":     {"v1beta4", "v1beta5"},
			"deployment": {"v1beta3", "v1beta4"},
			"provider":   {"v1beta3", "v1beta4"},
		},
		Status: SupportLevelMaintenance,
	})

	matrix.AddServerVersion(ServerCompatibility{
		Version:          "v0.8.0",
		MinClientVersion: "v0.8.0",
		SupportedProtocols: map[string][]string{
			"capture":  {"1"},
			"provider": {"1"},
			"manifest": {"2.0"},
		},
		SupportedModuleVersions: map[string][]string{
			"veid":       {"v1"},
			"market":     {"v1beta4"},
			"deployment": {"v1beta3"},
			"provider":   {"v1beta3"},
		},
		Status: SupportLevelMaintenance,
	})

	// Client versions
	matrix.AddClientVersion(ClientCompatibility{
		Version:          "v0.10.0",
		MinServerVersion: "v0.8.0",
		SupportsProtocols: map[string][]string{
			"capture":  {"1"},
			"provider": {"1", "2"},
			"manifest": {"2.0", "2.1"},
		},
		SupportsModuleVersions: map[string][]string{
			"veid":       {"v1"},
			"market":     {"v1beta4", "v1beta5"},
			"deployment": {"v1beta3", "v1beta4"},
			"provider":   {"v1beta3", "v1beta4"},
		},
		Status: SupportLevelActive,
	})

	matrix.AddClientVersion(ClientCompatibility{
		Version:          "v0.9.0",
		MinServerVersion: "v0.8.0",
		SupportsProtocols: map[string][]string{
			"capture":  {"1"},
			"provider": {"1", "2"},
			"manifest": {"2.0", "2.1"},
		},
		SupportsModuleVersions: map[string][]string{
			"veid":       {"v1"},
			"market":     {"v1beta4", "v1beta5"},
			"deployment": {"v1beta3", "v1beta4"},
			"provider":   {"v1beta3", "v1beta4"},
		},
		Status: SupportLevelMaintenance,
	})

	matrix.AddClientVersion(ClientCompatibility{
		Version:          "v0.8.0",
		MinServerVersion: "v0.8.0",
		SupportsProtocols: map[string][]string{
			"capture":  {"1"},
			"provider": {"1"},
			"manifest": {"2.0"},
		},
		SupportsModuleVersions: map[string][]string{
			"veid":       {"v1"},
			"market":     {"v1beta4"},
			"deployment": {"v1beta3"},
			"provider":   {"v1beta3"},
		},
		Status: SupportLevelMaintenance,
	})

	// Module matrices
	matrix.AddModuleMatrix(ModuleCompatibilityMatrix{
		ModuleName:     "market",
		CurrentVersion: "v1beta5",
		Versions: []ModuleVersionCompatibility{
			{
				Version:      "v1beta5",
				Status:       SupportLevelActive,
				IntroducedIn: "v0.9.0",
			},
			{
				Version:        "v1beta4",
				Status:         SupportLevelDeprecated,
				IntroducedIn:   "v0.8.0",
				DeprecatedIn:   "v0.10.0",
				CompatibleWith: []string{"v1beta5"},
				MigrationGuide: "docs/migrations/market-v1beta4-to-v1beta5.md",
			},
		},
	})

	matrix.AddModuleMatrix(ModuleCompatibilityMatrix{
		ModuleName:     "deployment",
		CurrentVersion: "v1beta4",
		Versions: []ModuleVersionCompatibility{
			{
				Version:      "v1beta4",
				Status:       SupportLevelActive,
				IntroducedIn: "v0.9.0",
			},
			{
				Version:        "v1beta3",
				Status:         SupportLevelMaintenance,
				IntroducedIn:   "v0.8.0",
				CompatibleWith: []string{"v1beta4"},
			},
		},
	})

	return matrix
}
