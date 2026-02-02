package v1

type LeaseClosedReasonRange int

const (
	LeaseClosedReasonRangeOwner LeaseClosedReasonRange = iota
	LeaseClosedReasonRangeProvider
	LeaseClosedReasonRangeNetwork
)

func (m LeaseClosedReason) IsRange(r LeaseClosedReasonRange) bool {
	switch r {
	case LeaseClosedReasonRangeOwner:
		return m >= 0 && m <= 9999
	case LeaseClosedReasonRangeProvider:
		return m >= 10000 && m <= 19999
	case LeaseClosedReasonRangeNetwork:
		return m >= 20000 && m <= 29999
	}

	return false
}
