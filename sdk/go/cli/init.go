package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	cobra.EnableTraverseRunHooks = true
}
