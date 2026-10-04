// Package v1_10_0 registers the per-account assurance vector state introduction.
package v1_10_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() { utypes.RegisterUpgrade(UpgradeName, initUpgrade) }
