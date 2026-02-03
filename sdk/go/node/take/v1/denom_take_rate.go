package v1

import (
	"sort"
)

type DenomTakeRates []DenomTakeRate

var _ sort.Interface = (*DenomTakeRates)(nil)

func (u DenomTakeRates) Len() int {
	return len(u)
}

func (u DenomTakeRates) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u DenomTakeRates) Less(i, j int) bool {
	return u[i].Denom < u[j].Denom
}
