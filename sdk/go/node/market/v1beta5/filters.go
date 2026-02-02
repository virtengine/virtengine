package v1beta5

// Accept returns whether order filters valid or not
func (filters *OrderFilters) Accept(obj Order, stateVal Order_State) bool {
	// Checking owner filter
	if filters.Owner != "" && filters.Owner != obj.ID.Owner {
		return false
	}

	// Checking dseq filter
	if filters.DSeq != 0 && filters.DSeq != obj.ID.DSeq {
		return false
	}

	// Checking gseq filter
	if filters.GSeq != 0 && filters.GSeq != obj.ID.GSeq {
		return false
	}

	// Checking oseq filter
	if filters.OSeq != 0 && filters.OSeq != obj.ID.OSeq {
		return false
	}

	// Checking state filter
	if stateVal != 0 && stateVal != obj.State {
		return false
	}

	return true
}

// Accept returns whether bid filters valid or not
func (filters *BidFilters) Accept(obj Bid, stateVal Bid_State) bool {
	// Checking owner filter
	if filters.Owner != "" && filters.Owner != obj.ID.Owner {
		return false
	}

	// Checking dseq filter
	if filters.DSeq != 0 && filters.DSeq != obj.ID.DSeq {
		return false
	}

	// Checking gseq filter
	if filters.GSeq != 0 && filters.GSeq != obj.ID.GSeq {
		return false
	}

	// Checking oseq filter
	if filters.OSeq != 0 && filters.OSeq != obj.ID.OSeq {
		return false
	}

	// Checking provider filter
	if filters.Provider != "" && filters.Provider != obj.ID.Provider {
		return false
	}

	// Checking bseq filter
	if filters.BSeq != 0 && filters.BSeq != obj.ID.BSeq {
		return false
	}

	// Checking state filter
	if stateVal != 0 && stateVal != obj.State {
		return false
	}

	return true
}
