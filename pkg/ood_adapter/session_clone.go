// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
package ood_adapter

func cloneSession(session *OODSession) *OODSession {
	if session == nil {
		return nil
	}

	clone := *session
	if session.Resources != nil {
		resources := *session.Resources
		clone.Resources = &resources
	}
	if session.StartedAt != nil {
		started := *session.StartedAt
		clone.StartedAt = &started
	}
	if session.EndedAt != nil {
		ended := *session.EndedAt
		clone.EndedAt = &ended
	}
	return &clone
}
