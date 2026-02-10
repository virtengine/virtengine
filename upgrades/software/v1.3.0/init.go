// Package v1_3_0
// nolint revive
package v1_3_0

import (
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

func init() {
	utypes.RegisterUpgrade(UpgradeName, initUpgrade)
}
