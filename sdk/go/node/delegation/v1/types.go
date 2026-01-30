// Package v1 provides Equal methods for generated delegation types.
package v1

// Equal returns true if two Params are equal
func (m *Params) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}
	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return m == nil
	} else if m == nil {
		return false
	}
	if m.UnbondingPeriod != that1.UnbondingPeriod {
		return false
	}
	if m.MaxValidators != that1.MaxValidators {
		return false
	}
	if m.MinDelegation != that1.MinDelegation {
		return false
	}
	if m.RedelegationCooldown != that1.RedelegationCooldown {
		return false
	}
	return true
}

// Equal returns true if two Delegation are equal
func (m *Delegation) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}
	that1, ok := that.(*Delegation)
	if !ok {
		that2, ok := that.(Delegation)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return m == nil
	} else if m == nil {
		return false
	}
	if m.Delegator != that1.Delegator {
		return false
	}
	if m.Validator != that1.Validator {
		return false
	}
	if m.Shares != that1.Shares {
		return false
	}
	return true
}

// Equal returns true if two UnbondingDelegation are equal
func (m *UnbondingDelegation) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}
	that1, ok := that.(*UnbondingDelegation)
	if !ok {
		that2, ok := that.(UnbondingDelegation)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return m == nil
	} else if m == nil {
		return false
	}
	if m.Delegator != that1.Delegator {
		return false
	}
	if m.Validator != that1.Validator {
		return false
	}
	if m.Amount != that1.Amount {
		return false
	}
	if m.CompletionTime != that1.CompletionTime {
		return false
	}
	return true
}

// Equal returns true if two Redelegation are equal
func (m *Redelegation) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}
	that1, ok := that.(*Redelegation)
	if !ok {
		that2, ok := that.(Redelegation)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return m == nil
	} else if m == nil {
		return false
	}
	if m.Delegator != that1.Delegator {
		return false
	}
	if m.SrcValidator != that1.SrcValidator {
		return false
	}
	if m.DstValidator != that1.DstValidator {
		return false
	}
	if m.Amount != that1.Amount {
		return false
	}
	if m.CompletionTime != that1.CompletionTime {
		return false
	}
	return true
}
