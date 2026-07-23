// Package v1_7_0 registers canonical financial cases.
package v1_7_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() { utypes.RegisterUpgrade(UpgradeName, initUpgrade) }
