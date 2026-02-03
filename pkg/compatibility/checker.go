package compatibility

import (
	"fmt"
	"time"
)

// CompatibilityChecker provides methods to check compatibility between versions
// and enforce deprecation policies.
type CompatibilityChecker struct {
	// CurrentVersion is the current server version
	CurrentVersion Version

	// SupportPolicy defines how many minor versions to support (default: 2)
	SupportPolicy int

	// DeprecationWindow is the number of versions before removal (default: 2)
	DeprecationWindow int

	// ModuleVersions maps module names to their API versions
	ModuleVersions map[string]APIVersionSupport

	// ProtocolNegotiator handles protocol version negotiation
	ProtocolNegotiator *MultiProtocolNegotiator
}

// APIVersionSupport describes version support for an API.
type APIVersionSupport struct {
	// ModuleName is the module name (e.g., "veid", "market")
	ModuleName string

	// CurrentVersion is the current API version
	CurrentVersion APIVersion

	// SupportedVersions lists all supported API versions
	SupportedVersions []APIVersion

	// DeprecatedVersions lists deprecated but still supported versions
	DeprecatedVersions []DeprecatedAPIVersion

	// NextVersion is the planned next version (if any)
	NextVersion string
}

// DeprecatedAPIVersion describes a deprecated API version.
type DeprecatedAPIVersion struct {
	// Version is the deprecated version
	Version APIVersion

	// DeprecatedAt is when deprecation was announced
	DeprecatedAt time.Time

	// SunsetAt is when the version will be removed
	SunsetAt time.Time

	// Message is the deprecation message
	Message string

	// Successor is the recommended replacement version
	Successor string
}

// CompatibilityResult describes the result of a compatibility check.
type CompatibilityResult struct {
	// Compatible indicates if versions are compatible
	Compatible bool

	// Warnings contains any compatibility warnings
	Warnings []string

	// Errors contains any compatibility errors
	Errors []string

	// DeprecationWarnings contains deprecation notices
	DeprecationWarnings []DeprecationWarning
}

// DeprecationWarning represents a warning about deprecated functionality.
type DeprecationWarning struct {
	// Type is the type of deprecated item (api, protocol, feature)
	Type string

	// Item is the deprecated item name
	Item string

	// Message is the deprecation message
	Message string

	// SunsetDate is when the item will be removed
	SunsetDate *time.Time

	// Successor is the recommended replacement
	Successor string
}

// NewCompatibilityChecker creates a new compatibility checker.
func NewCompatibilityChecker(currentVersion Version) *CompatibilityChecker {
	return &CompatibilityChecker{
		CurrentVersion:     currentVersion,
		SupportPolicy:      2, // N-2 support policy
		DeprecationWindow:  2,
		ModuleVersions:     make(map[string]APIVersionSupport),
		ProtocolNegotiator: DefaultMultiProtocolNegotiator(),
	}
}

// RegisterModuleVersion registers a module's version support information.
func (c *CompatibilityChecker) RegisterModuleVersion(support APIVersionSupport) {
	c.ModuleVersions[support.ModuleName] = support
}

// CheckClientVersion checks if a client version is compatible with the server.
func (c *CompatibilityChecker) CheckClientVersion(clientVersion Version) CompatibilityResult {
	result := CompatibilityResult{
		Compatible: true,
	}

	supportLevel := GetSupportLevel(clientVersion, c.CurrentVersion)

	switch supportLevel {
	case SupportLevelActive:
		// Fully compatible
	case SupportLevelMaintenance:
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Client version %s is in maintenance mode. Consider upgrading to %s.",
				clientVersion.String(), c.CurrentVersion.String()))
	case SupportLevelDeprecated:
		result.Compatible = false
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Client version %s is deprecated. Upgrade required to %s.",
				clientVersion.String(), c.CurrentVersion.String()))
		result.DeprecationWarnings = append(result.DeprecationWarnings, DeprecationWarning{
			Type:      "client",
			Item:      clientVersion.String(),
			Message:   "Client version is deprecated",
			Successor: c.CurrentVersion.String(),
		})
	case SupportLevelEOL:
		result.Compatible = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Client version %s is no longer supported. Upgrade to %s.",
				clientVersion.String(), c.CurrentVersion.String()))
	}

	return result
}

// CheckAPIVersion checks if a requested API version is compatible.
func (c *CompatibilityChecker) CheckAPIVersion(moduleName string, requestedVersion APIVersion) CompatibilityResult {
	result := CompatibilityResult{
		Compatible: true,
	}

	support, ok := c.ModuleVersions[moduleName]
	if !ok {
		result.Errors = append(result.Errors, fmt.Sprintf("Unknown module: %s", moduleName))
		result.Compatible = false
		return result
	}

	// Check if version is supported
	found := false
	for _, v := range support.SupportedVersions {
		if v.Compare(requestedVersion) == 0 {
			found = true
			break
		}
	}

	if !found {
		result.Compatible = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("API version %s is not supported for module %s. Supported: %v",
				requestedVersion.String(), moduleName, apiVersionsToStrings(support.SupportedVersions)))
		return result
	}

	// Check if version is deprecated
	for _, dv := range support.DeprecatedVersions {
		if dv.Version.Compare(requestedVersion) == 0 {
			result.DeprecationWarnings = append(result.DeprecationWarnings, DeprecationWarning{
				Type:       "api",
				Item:       fmt.Sprintf("%s/%s", moduleName, requestedVersion.String()),
				Message:    dv.Message,
				SunsetDate: &dv.SunsetAt,
				Successor:  dv.Successor,
			})
			result.Warnings = append(result.Warnings, dv.Message)
			break
		}
	}

	return result
}

// CheckProtocolVersion checks if a protocol version is compatible.
func (c *CompatibilityChecker) CheckProtocolVersion(protocol string, version string) CompatibilityResult {
	result := CompatibilityResult{
		Compatible: true,
	}

	negotiator, ok := c.ProtocolNegotiator.Get(protocol)
	if !ok {
		result.Errors = append(result.Errors, fmt.Sprintf("Unknown protocol: %s", protocol))
		result.Compatible = false
		return result
	}

	if !negotiator.IsSupported(version) {
		result.Compatible = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Protocol version %s is not supported for %s. Supported: %v",
				version, protocol, negotiator.GetSupportedVersionStrings()))
		return result
	}

	if msg, successor, deprecated := negotiator.GetDeprecationInfo(version); deprecated {
		result.DeprecationWarnings = append(result.DeprecationWarnings, DeprecationWarning{
			Type:      "protocol",
			Item:      fmt.Sprintf("%s/%s", protocol, version),
			Message:   msg,
			Successor: successor,
		})
		result.Warnings = append(result.Warnings, msg)
	}

	return result
}

// CheckBreakingChange checks if a change is breaking based on the change description.
func (c *CompatibilityChecker) CheckBreakingChange(changeType BreakingChangeType) bool {
	return changeType.IsBreaking()
}

// BreakingChangeType represents types of changes that can be breaking or non-breaking.
type BreakingChangeType int

const (
	// ChangeAddField is adding a new optional field (non-breaking)
	ChangeAddField BreakingChangeType = iota

	// ChangeRemoveField is removing a field (breaking)
	ChangeRemoveField

	// ChangeRenameField is renaming a field (breaking)
	ChangeRenameField

	// ChangeTypeField is changing a field's type (breaking)
	ChangeTypeField

	// ChangeAddEndpoint is adding a new endpoint (non-breaking)
	ChangeAddEndpoint

	// ChangeRemoveEndpoint is removing an endpoint (breaking)
	ChangeRemoveEndpoint

	// ChangeModifyEndpoint is modifying an endpoint's behavior (breaking)
	ChangeModifyEndpoint

	// ChangeAddEnumValue is adding a new enum value (non-breaking)
	ChangeAddEnumValue

	// ChangeRemoveEnumValue is removing an enum value (breaking)
	ChangeRemoveEnumValue

	// ChangeModifyDefault is modifying default values with behavioral impact (breaking)
	ChangeModifyDefault

	// ChangeModifyErrorCodes is modifying error codes (breaking)
	ChangeModifyErrorCodes

	// ChangeModifyAuth is modifying authentication/authorization (breaking)
	ChangeModifyAuth
)

// IsBreaking returns true if this change type is breaking.
func (ct BreakingChangeType) IsBreaking() bool {
	switch ct {
	case ChangeAddField, ChangeAddEndpoint, ChangeAddEnumValue:
		return false
	default:
		return true
	}
}

// String returns the string representation of the change type.
func (ct BreakingChangeType) String() string {
	names := map[BreakingChangeType]string{
		ChangeAddField:         "add_field",
		ChangeRemoveField:      "remove_field",
		ChangeRenameField:      "rename_field",
		ChangeTypeField:        "type_field",
		ChangeAddEndpoint:      "add_endpoint",
		ChangeRemoveEndpoint:   "remove_endpoint",
		ChangeModifyEndpoint:   "modify_endpoint",
		ChangeAddEnumValue:     "add_enum_value",
		ChangeRemoveEnumValue:  "remove_enum_value",
		ChangeModifyDefault:    "modify_default",
		ChangeModifyErrorCodes: "modify_error_codes",
		ChangeModifyAuth:       "modify_auth",
	}
	if name, ok := names[ct]; ok {
		return name
	}
	return "unknown"
}

// ValidateCompatibility performs comprehensive compatibility validation.
func (c *CompatibilityChecker) ValidateCompatibility(
	clientVersion Version,
	apiVersions map[string]APIVersion,
	protocolVersions map[string]string,
) CompatibilityResult {
	result := CompatibilityResult{
		Compatible: true,
	}

	// Check client version
	clientResult := c.CheckClientVersion(clientVersion)
	result.merge(clientResult)

	// Check API versions
	for module, version := range apiVersions {
		apiResult := c.CheckAPIVersion(module, version)
		result.merge(apiResult)
	}

	// Check protocol versions
	for protocol, version := range protocolVersions {
		protoResult := c.CheckProtocolVersion(protocol, version)
		result.merge(protoResult)
	}

	return result
}

// merge merges another result into this one.
func (r *CompatibilityResult) merge(other CompatibilityResult) {
	if !other.Compatible {
		r.Compatible = false
	}
	r.Warnings = append(r.Warnings, other.Warnings...)
	r.Errors = append(r.Errors, other.Errors...)
	r.DeprecationWarnings = append(r.DeprecationWarnings, other.DeprecationWarnings...)
}

// IsValid returns true if there are no errors.
func (r *CompatibilityResult) IsValid() bool {
	return len(r.Errors) == 0
}

// HasWarnings returns true if there are warnings.
func (r *CompatibilityResult) HasWarnings() bool {
	return len(r.Warnings) > 0 || len(r.DeprecationWarnings) > 0
}

// apiVersionsToStrings converts a slice of APIVersions to strings.
func apiVersionsToStrings(versions []APIVersion) []string {
	strs := make([]string, len(versions))
	for i, v := range versions {
		strs[i] = v.String()
	}
	return strs
}

// DefaultModuleVersionSupport returns default module version support configuration.
func DefaultModuleVersionSupport() map[string]APIVersionSupport {
	return map[string]APIVersionSupport{
		"veid": {
			ModuleName:        "veid",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
		},
		"mfa": {
			ModuleName:        "mfa",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
		},
		"encryption": {
			ModuleName:        "encryption",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
			NextVersion:       "v2",
		},
		"market": {
			ModuleName:        "market",
			CurrentVersion:    MustParseAPIVersion("v1beta5"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1beta4"), MustParseAPIVersion("v1beta5")},
			DeprecatedVersions: []DeprecatedAPIVersion{
				{
					Version:   MustParseAPIVersion("v1beta4"),
					Message:   "v1beta4 is deprecated, use v1beta5",
					Successor: "v1beta5",
				},
			},
		},
		"deployment": {
			ModuleName:        "deployment",
			CurrentVersion:    MustParseAPIVersion("v1beta4"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1beta3"), MustParseAPIVersion("v1beta4")},
		},
		"provider": {
			ModuleName:        "provider",
			CurrentVersion:    MustParseAPIVersion("v1beta4"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1beta3"), MustParseAPIVersion("v1beta4")},
		},
		"escrow": {
			ModuleName:        "escrow",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
		},
		"audit": {
			ModuleName:        "audit",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
		},
		"cert": {
			ModuleName:        "cert",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
		},
		"hpc": {
			ModuleName:        "hpc",
			CurrentVersion:    MustParseAPIVersion("v1"),
			SupportedVersions: []APIVersion{MustParseAPIVersion("v1")},
		},
	}
}

// NewDefaultCompatibilityChecker creates a checker with default configuration.
func NewDefaultCompatibilityChecker(currentVersion Version) *CompatibilityChecker {
	checker := NewCompatibilityChecker(currentVersion)
	for _, support := range DefaultModuleVersionSupport() {
		checker.RegisterModuleVersion(support)
	}
	return checker
}
