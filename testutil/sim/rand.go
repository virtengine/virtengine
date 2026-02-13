package sim

import (
	"math/rand" //nolint:gosec // G404: simulation uses weak random for reproducibility
)

func RandIdx(r *rand.Rand, val int) int {
	if val == 0 {
		return 0
	}

	return r.Intn(val) //nolint:gosec // G404: simulation randomness is non-security-critical
}
