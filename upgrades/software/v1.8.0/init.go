// Package v1_8_0 registers authenticated fiat conversion observations.
package v1_8_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() { utypes.RegisterUpgrade(UpgradeName, initUpgrade) }
