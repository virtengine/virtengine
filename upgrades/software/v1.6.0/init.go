// Package v1_6_0 registers canonical market reservations.
package v1_6_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() { utypes.RegisterUpgrade(UpgradeName, initUpgrade) }
