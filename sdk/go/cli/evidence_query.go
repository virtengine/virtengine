package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/version"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryEvidenceCmd returns the CLI command with all evidence module query commands
// mounted.
func GetQueryEvidenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Query for evidence by hash or for all (paginated) submitted evidence",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for specific submitted evidence by hash or query for all (paginated) evidence:

Example:
$ %s query %s DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ %s query %s --page=2 --limit=50
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args:                       cobra.MaximumNArgs(1),
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			if len(args) > 0 {
				params := &types.QueryEvidenceRequest{Hash: args[0]}
				res, err := cl.Query().Evidence().Evidence(ctx, params)
				if err != nil {
					return err
				}

				return cl.PrintMessage(res.Evidence)
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryAllEvidenceRequest{
				Pagination: pageReq,
			}

			res, err := cl.Query().Evidence().AllEvidence(ctx, params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "evidence")

	return cmd
}
