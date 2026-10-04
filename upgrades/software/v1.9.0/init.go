// Package v1_9_0 registers deterministic market resolution activation.
package v1_9_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() { utypes.RegisterUpgrade(UpgradeName, initUpgrade) }
