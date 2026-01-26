package types

import (
	"strconv"
	"strings"
)

// VersionConstraint represents a version constraint for approved clients
type VersionConstraint struct {
	// MinVersion is the minimum required version (inclusive)
	MinVersion string `json:"min_version"`

	// MaxVersion is the maximum allowed version (inclusive, optional)
	MaxVersion string `json:"max_version,omitempty"`
}

// NewVersionConstraint creates a new VersionConstraint
func NewVersionConstraint(minVersion, maxVersion string) *VersionConstraint {
	return &VersionConstraint{
		MinVersion: minVersion,
		MaxVersion: maxVersion,
	}
}

// Validate validates the version constraint
func (v *VersionConstraint) Validate() error {
	if v.MinVersion == "" {
		return ErrInvalidVersionConstraint.Wrap("min_version cannot be empty")
	}

	if !isValidSemver(v.MinVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid min_version: %s", v.MinVersion)
	}

	if v.MaxVersion != "" {
		if !isValidSemver(v.MaxVersion) {
			return ErrInvalidVersionConstraint.Wrapf("invalid max_version: %s", v.MaxVersion)
		}

		// Ensure min <= max
		cmp, err := compareSemver(v.MinVersion, v.MaxVersion)
		if err != nil {
			return err
		}
		if cmp > 0 {
			return ErrInvalidVersionConstraint.Wrap("min_version cannot be greater than max_version")
		}
	}

	return nil
}

// Satisfies checks if a version satisfies this constraint
func (v *VersionConstraint) Satisfies(version string) bool {
	if !isValidSemver(version) {
		return false
	}

	// Check minimum version
	cmp, err := compareSemver(version, v.MinVersion)
	if err != nil || cmp < 0 {
		return false
	}

	// Check maximum version if set
	if v.MaxVersion != "" {
		cmp, err = compareSemver(version, v.MaxVersion)
		if err != nil || cmp > 0 {
			return false
		}
	}

	return true
}

// semverParts represents parsed semver components
type semverParts struct {
	major      int
	minor      int
	patch      int
	prerelease string
	build      string
}

// parseSemver parses a semantic version string
func parseSemver(version string) (*semverParts, error) {
	parts := &semverParts{}

	// Split off build metadata
	buildSplit := strings.SplitN(version, "+", 2)
	if len(buildSplit) == 2 {
		parts.build = buildSplit[1]
	}
	version = buildSplit[0]

	// Split off prerelease
	preSplit := strings.SplitN(version, "-", 2)
	if len(preSplit) == 2 {
		parts.prerelease = preSplit[1]
	}
	version = preSplit[0]

	// Parse version numbers
	numParts := strings.Split(version, ".")
	if len(numParts) != 3 {
		return nil, ErrInvalidVersionConstraint.Wrapf("invalid version format: %s", version)
	}

	var err error
	parts.major, err = strconv.Atoi(numParts[0])
	if err != nil {
		return nil, ErrInvalidVersionConstraint.Wrapf("invalid major version: %s", numParts[0])
	}

	parts.minor, err = strconv.Atoi(numParts[1])
	if err != nil {
		return nil, ErrInvalidVersionConstraint.Wrapf("invalid minor version: %s", numParts[1])
	}

	parts.patch, err = strconv.Atoi(numParts[2])
	if err != nil {
		return nil, ErrInvalidVersionConstraint.Wrapf("invalid patch version: %s", numParts[2])
	}

	return parts, nil
}

// compareSemver compares two semantic versions
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareSemver(a, b string) (int, error) {
	aParts, err := parseSemver(a)
	if err != nil {
		return 0, err
	}

	bParts, err := parseSemver(b)
	if err != nil {
		return 0, err
	}

	// Compare major
	if aParts.major < bParts.major {
		return -1, nil
	}
	if aParts.major > bParts.major {
		return 1, nil
	}

	// Compare minor
	if aParts.minor < bParts.minor {
		return -1, nil
	}
	if aParts.minor > bParts.minor {
		return 1, nil
	}

	// Compare patch
	if aParts.patch < bParts.patch {
		return -1, nil
	}
	if aParts.patch > bParts.patch {
		return 1, nil
	}

	// Compare prerelease (version without prerelease is higher than with)
	if aParts.prerelease == "" && bParts.prerelease != "" {
		return 1, nil
	}
	if aParts.prerelease != "" && bParts.prerelease == "" {
		return -1, nil
	}
	if aParts.prerelease < bParts.prerelease {
		return -1, nil
	}
	if aParts.prerelease > bParts.prerelease {
		return 1, nil
	}

	return 0, nil
}

// CompareVersions is a public wrapper for compareSemver
func CompareVersions(a, b string) (int, error) {
	return compareSemver(a, b)
}

// IsVersionInRange checks if a version is within the given range
func IsVersionInRange(version, minVersion, maxVersion string) bool {
	constraint := NewVersionConstraint(minVersion, maxVersion)
	return constraint.Satisfies(version)
}
