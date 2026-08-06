// Package v1_5_0 registers authenticated metering and provider key epochs.
//
//nolint:revive
package v1_5_0

import utypes "github.com/virtengine/virtengine/upgrades/types"

func init() {
	utypes.RegisterUpgrade(UpgradeName, initUpgrade)
}
