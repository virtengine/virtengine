package v2beta3

import (
	"sort"
)

type Services []Service

var _ sort.Interface = (*ServiceExposes)(nil)

func (s Services) Len() int {
	return len(s)
}

func (s Services) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Services) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
