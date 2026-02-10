package v1

func (s State) IsValid() bool {
	switch s {
	case StateOpen, StateClosed, StateOverdrawn:
		return true
	default:
	}

	return false
}
