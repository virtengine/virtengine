// Package v1_4_0 registers the Task 84A consensus-admission upgrade.
//
//nolint:revive
package v1_4_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() {
	utypes.RegisterUpgrade(UpgradeName, initUpgrade)
}
