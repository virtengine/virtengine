// Package compatibility provides version negotiation, compatibility checking,
// and deprecation management utilities for the VirtEngine API.
//
// This package implements the backwards compatibility infrastructure as
// specified in docs/COMPATIBILITY.md, supporting:
//   - Semantic version parsing and comparison
//   - Version range validation
//   - Protocol version negotiation
//   - Deprecation tracking and enforcement
package compatibility

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version with optional pre-release and build metadata.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

// APIVersion represents a versioned API identifier (e.g., "v1", "v1beta2").
type APIVersion struct {
	Version     int
	Stability   Stability
	Revision    int // For beta/alpha versions (e.g., beta2 has Revision=2)
	RawVersion  string
}

// Stability represents the stability level of an API version.
type Stability string

const (
	// StabilityStable represents a stable, production-ready API.
	StabilityStable Stability = "stable"

	// StabilityBeta represents a beta API that may have breaking changes.
	StabilityBeta Stability = "beta"

	// StabilityAlpha represents an experimental API.
	StabilityAlpha Stability = "alpha"
)

// VersionRange represents a range of supported versions.
type VersionRange struct {
	Min     Version
	Max     Version
	MinStr  string
	MaxStr  string
}

// SupportLevel indicates the support status of a version.
type SupportLevel string

const (
	// SupportLevelActive indicates full support with new features and bug fixes.
	SupportLevelActive SupportLevel = "active"

	// SupportLevelMaintenance indicates security fixes and critical bugs only.
	SupportLevelMaintenance SupportLevel = "maintenance"

	// SupportLevelDeprecated indicates no support, migration strongly encouraged.
	SupportLevelDeprecated SupportLevel = "deprecated"

	// SupportLevelEOL indicates end of life, not supported.
	SupportLevelEOL SupportLevel = "eol"
)

var (
	// semverRegex matches semantic version strings.
	semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*))?(?:\+([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*))?$`)

	// apiVersionRegex matches API version strings like "v1", "v1beta2", "v2alpha1".
	apiVersionRegex = regexp.MustCompile(`^v(\d+)((?:alpha|beta)(\d*))?$`)
)

// ParseVersion parses a semantic version string into a Version struct.
func ParseVersion(s string) (Version, error) {
	matches := semverRegex.FindStringSubmatch(s)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid semantic version: %s", s)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: matches[4],
		Build:      matches[5],
	}, nil
}

// MustParseVersion parses a version string and panics on error.
func MustParseVersion(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns the string representation of a version.
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return compareInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return compareInt(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return compareInt(v.Patch, other.Patch)
	}

	// Pre-release versions have lower precedence
	if v.PreRelease == "" && other.PreRelease != "" {
		return 1
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		return -1
	}
	if v.PreRelease != other.PreRelease {
		return strings.Compare(v.PreRelease, other.PreRelease)
	}

	return 0
}

// LessThan returns true if v < other.
func (v Version) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

// LessThanOrEqual returns true if v <= other.
func (v Version) LessThanOrEqual(other Version) bool {
	return v.Compare(other) <= 0
}

// GreaterThan returns true if v > other.
func (v Version) GreaterThan(other Version) bool {
	return v.Compare(other) > 0
}

// GreaterThanOrEqual returns true if v >= other.
func (v Version) GreaterThanOrEqual(other Version) bool {
	return v.Compare(other) >= 0
}

// Equal returns true if v == other.
func (v Version) Equal(other Version) bool {
	return v.Compare(other) == 0
}

// IsPreRelease returns true if this is a pre-release version.
func (v Version) IsPreRelease() bool {
	return v.PreRelease != ""
}

// IsMainnet returns true if this is a mainnet version (even minor number).
func (v Version) IsMainnet() bool {
	return v.Minor%2 == 0
}

// IsTestnet returns true if this is a testnet version (odd minor number).
func (v Version) IsTestnet() bool {
	return v.Minor%2 == 1
}

// ParseAPIVersion parses an API version string like "v1", "v1beta2".
func ParseAPIVersion(s string) (APIVersion, error) {
	matches := apiVersionRegex.FindStringSubmatch(strings.ToLower(s))
	if matches == nil {
		return APIVersion{}, fmt.Errorf("invalid API version: %s", s)
	}

	version, _ := strconv.Atoi(matches[1])
	stability := StabilityStable
	revision := 0

	if matches[2] != "" {
		if strings.HasPrefix(matches[2], "beta") {
			stability = StabilityBeta
		} else if strings.HasPrefix(matches[2], "alpha") {
			stability = StabilityAlpha
		}
		if matches[3] != "" {
			revision, _ = strconv.Atoi(matches[3])
		} else {
			revision = 1 // v1beta == v1beta1
		}
	}

	return APIVersion{
		Version:    version,
		Stability:  stability,
		Revision:   revision,
		RawVersion: s,
	}, nil
}

// MustParseAPIVersion parses an API version string and panics on error.
func MustParseAPIVersion(s string) APIVersion {
	v, err := ParseAPIVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns the string representation of an API version.
func (v APIVersion) String() string {
	if v.RawVersion != "" {
		return v.RawVersion
	}
	s := fmt.Sprintf("v%d", v.Version)
	if v.Stability == StabilityBeta {
		s += fmt.Sprintf("beta%d", v.Revision)
	} else if v.Stability == StabilityAlpha {
		s += fmt.Sprintf("alpha%d", v.Revision)
	}
	return s
}

// Compare compares two API versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v APIVersion) Compare(other APIVersion) int {
	if v.Version != other.Version {
		return compareInt(v.Version, other.Version)
	}

	// Stable > Beta > Alpha
	stabilityOrder := map[Stability]int{
		StabilityAlpha:  0,
		StabilityBeta:   1,
		StabilityStable: 2,
	}

	vOrder := stabilityOrder[v.Stability]
	otherOrder := stabilityOrder[other.Stability]

	if vOrder != otherOrder {
		return compareInt(vOrder, otherOrder)
	}

	return compareInt(v.Revision, other.Revision)
}

// IsStable returns true if this is a stable API version.
func (v APIVersion) IsStable() bool {
	return v.Stability == StabilityStable
}

// NewVersionRange creates a version range from min and max version strings.
func NewVersionRange(min, max string) (VersionRange, error) {
	minVer, err := ParseVersion(min)
	if err != nil {
		return VersionRange{}, fmt.Errorf("invalid min version: %w", err)
	}

	maxVer, err := ParseVersion(max)
	if err != nil {
		return VersionRange{}, fmt.Errorf("invalid max version: %w", err)
	}

	if minVer.GreaterThan(maxVer) {
		return VersionRange{}, fmt.Errorf("min version %s is greater than max version %s", min, max)
	}

	return VersionRange{
		Min:    minVer,
		Max:    maxVer,
		MinStr: min,
		MaxStr: max,
	}, nil
}

// Contains returns true if the version is within the range.
func (r VersionRange) Contains(v Version) bool {
	return v.GreaterThanOrEqual(r.Min) && v.LessThanOrEqual(r.Max)
}

// Overlaps returns true if the ranges overlap.
func (r VersionRange) Overlaps(other VersionRange) bool {
	return r.Min.LessThanOrEqual(other.Max) && other.Min.LessThanOrEqual(r.Max)
}

// String returns the string representation of a version range.
func (r VersionRange) String() string {
	return fmt.Sprintf("%s-%s", r.MinStr, r.MaxStr)
}

// Helper function to compare integers.
func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// GetSupportedVersions returns the list of supported versions based on
// the current version and the N-2 support policy.
func GetSupportedVersions(current Version) []Version {
	supported := make([]Version, 0, 3)

	// Current version (Active)
	supported = append(supported, current)

	// N-1 (Maintenance)
	if current.Minor > 0 {
		n1 := Version{
			Major: current.Major,
			Minor: current.Minor - 1,
			Patch: 0, // Assume latest patch
		}
		supported = append(supported, n1)
	}

	// N-2 (Maintenance)
	if current.Minor > 1 {
		n2 := Version{
			Major: current.Major,
			Minor: current.Minor - 2,
			Patch: 0, // Assume latest patch
		}
		supported = append(supported, n2)
	}

	return supported
}

// GetSupportLevel returns the support level for a version given the current version.
func GetSupportLevel(v, current Version) SupportLevel {
	if v.Major != current.Major {
		// Different major version
		if v.Major < current.Major {
			return SupportLevelEOL
		}
		return SupportLevelActive // Future major version
	}

	minorDiff := current.Minor - v.Minor

	switch {
	case minorDiff < 0:
		return SupportLevelActive // Future version
	case minorDiff == 0:
		return SupportLevelActive
	case minorDiff <= 2:
		return SupportLevelMaintenance
	case minorDiff <= 4:
		return SupportLevelDeprecated
	default:
		return SupportLevelEOL
	}
}
