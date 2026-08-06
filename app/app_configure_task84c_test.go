package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderMigrationsRunsTask84CDependenciesFirstAndAuthLast(t *testing.T) {
	modules := []string{"zeta", "market", "auth", "resources", "mktplace", "hpc", "alpha"}
	require.Equal(t, []string{"resources", "market", "hpc", "mktplace", "alpha", "zeta", "auth"}, OrderMigrations(modules))
}
