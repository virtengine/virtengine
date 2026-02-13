package testutil

import (
	"math/rand" //nolint:gosec // G404: test helpers use weak random for non-security data
	"time"
)

// non-constant random seed for math/rand functions

func init() {
	rand.Seed(time.Now().Unix()) // nolint: staticcheck
}
