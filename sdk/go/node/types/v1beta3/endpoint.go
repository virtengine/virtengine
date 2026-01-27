package v1beta3

import (
	"sort"
)

type Endpoints []Endpoint

var _ sort.Interface = (*Endpoints)(nil)

func (u Endpoints) Dup() Endpoints {
	res := make(Endpoints, len(u))

	copy(res, u)

	return res
}

func (u Endpoints) Len() int {
	return len(u)
}

func (u Endpoints) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u Endpoints) Less(i, j int) bool {
	return u[i].SequenceNumber < u[j].SequenceNumber
}
