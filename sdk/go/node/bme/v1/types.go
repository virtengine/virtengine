package v1

func (m Status) Equal(other Status) bool {
	return (m.Status == other.Status) && (m.PreviousStatus == other.PreviousStatus) && (m.EpochHeightDiff == other.EpochHeightDiff)
}
