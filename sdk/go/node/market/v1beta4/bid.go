package v1beta4

import (
	"sort"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta3"
)

type ResourcesOffer []ResourceOffer

var _ sort.Interface = (*ResourcesOffer)(nil)

func (s ResourcesOffer) MatchGSpec(gspec dtypes.GroupSpec) bool {
	if len(s) == 0 {
		return true
	}

	ru := make(map[uint32]*dtypes.ResourceUnit)

	for idx := range gspec.Resources {
		ru[gspec.Resources[idx].ID] = &gspec.Resources[idx]
	}

	for _, ro := range s {
		res, exists := ru[ro.Resources.ID]
		if !exists {
			return false
		}

		ru[ro.Resources.ID] = nil

		if res.Count != ro.Count {
			return false
		}
	}

	return true
}

func (r *ResourceOffer) Dup() ResourceOffer {
	return ResourceOffer{
		Resources: r.Resources.Dup(),
		Count:     r.Count,
	}
}

func (s ResourcesOffer) Len() int {
	return len(s)
}

func (s ResourcesOffer) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ResourcesOffer) Less(i, j int) bool {
	return s[i].Resources.ID < s[j].Resources.ID
}

func (s ResourcesOffer) Dup() ResourcesOffer {
	res := make(ResourcesOffer, 0, len(s))

	for _, ru := range s {
		res = append(res, ru.Dup())
	}

	return res
}

func ResourceOfferFromRU(ru dtypes.ResourceUnits) ResourcesOffer {
	res := make(ResourcesOffer, 0, len(ru))

	for _, r := range ru {
		res = append(res, ResourceOffer{
			Resources: r.Resources,
			Count:     r.Count,
		})
	}

	return res
}

