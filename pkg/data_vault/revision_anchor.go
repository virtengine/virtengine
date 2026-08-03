package data_vault

import "github.com/virtengine/virtengine/pkg/data_vault/contracts"

type RevisionAnchor = contracts.RevisionAnchor
type ProcessRevisionAnchor = contracts.ProcessRevisionAnchor
type FixtureSecurityOptions = contracts.FixtureSecurityOptions

var ErrRevisionRollback = contracts.ErrRevisionRollback

func NewProcessRevisionAnchor() *ProcessRevisionAnchor {
	return contracts.NewProcessRevisionAnchor()
}
