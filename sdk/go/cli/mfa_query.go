package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
)

// GetQueryMFACmd returns the query commands for the MFA module
func GetQueryMFACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "MFA multi-factor authentication query commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetMFAPolicyCmd(),
		GetMFAFactorEnrollmentsCmd(),
		GetMFAFactorEnrollmentCmd(),
		GetMFAChallengeCmd(),
		GetMFAPendingChallengesCmd(),
		GetMFAAuthorizationSessionCmd(),
		GetMFATrustedDevicesCmd(),
		GetMFASensitiveTxConfigCmd(),
		GetMFARequiredCmd(),
		GetMFAParamsCmd(),
	)

	return cmd
}

func GetMFAPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "policy [address]",
		Short:             "Query MFA policy for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryMFAPolicyRequest{
				Address: args[0],
			}

			res, err := cl.Query().MFA().MFAPolicy(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetMFAFactorEnrollmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "enrollments [address]",
		Short:             "Query all MFA factor enrollments for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryFactorEnrollmentsRequest{
				Address:    args[0],
				Pagination: pageReq,
			}

			res, err := cl.Query().MFA().FactorEnrollments(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "enrollments")
	return cmd
}

func GetMFAFactorEnrollmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "enrollment [address] [factor-id]",
		Short:             "Query a specific MFA factor enrollment",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryFactorEnrollmentRequest{
				Address:  args[0],
				FactorId: args[1],
			}

			res, err := cl.Query().MFA().FactorEnrollment(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetMFAChallengeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "challenge [challenge-id]",
		Short:             "Query a specific MFA challenge",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryChallengeRequest{
				ChallengeId: args[0],
			}

			res, err := cl.Query().MFA().Challenge(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetMFAPendingChallengesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "pending-challenges [address]",
		Short:             "Query pending MFA challenges for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryPendingChallengesRequest{
				Address:    args[0],
				Pagination: pageReq,
			}

			res, err := cl.Query().MFA().PendingChallenges(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "pending-challenges")
	return cmd
}

func GetMFAAuthorizationSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "session [session-id]",
		Short:             "Query an MFA authorization session",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryAuthorizationSessionRequest{
				SessionId: args[0],
			}

			res, err := cl.Query().MFA().AuthorizationSession(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetMFATrustedDevicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "trusted-devices [address]",
		Short:             "Query trusted devices for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryTrustedDevicesRequest{
				Address:    args[0],
				Pagination: pageReq,
			}

			res, err := cl.Query().MFA().TrustedDevices(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "trusted-devices")
	return cmd
}

func GetMFASensitiveTxConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "sensitive-tx-config [tx-type]",
		Short:             "Query sensitive transaction configuration",
		Long:              "Query sensitive transaction configuration by type. Valid types: account_recovery, key_rotation, large_withdrawal, provider_registration, validator_registration, high_value_order, role_assignment, governance_proposal, etc.",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			txType, err := parseSensitiveTransactionType(args[0])
			if err != nil {
				return err
			}

			req := &types.QuerySensitiveTxConfigRequest{
				TransactionType: txType,
			}

			res, err := cl.Query().MFA().SensitiveTxConfig(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetMFARequiredCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "required [address] [tx-type]",
		Short:             "Check if MFA is required for a transaction",
		Long:              "Check if MFA is required for a transaction type. Valid types: account_recovery, key_rotation, large_withdrawal, provider_registration, validator_registration, high_value_order, role_assignment, governance_proposal, etc.",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			txType, err := parseSensitiveTransactionType(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryMFARequiredRequest{
				Address:         args[0],
				TransactionType: txType,
			}

			res, err := cl.Query().MFA().MFARequired(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetMFAParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Query MFA module parameters",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().MFA().Params(ctx, &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// parseSensitiveTransactionType parses a string to SensitiveTransactionType enum
// Accepts both numeric values and string names (with or without SENSITIVE_TX_ prefix)
func parseSensitiveTransactionType(s string) (types.SensitiveTransactionType, error) {
	// Try parsing as integer first
	if v, err := strconv.ParseInt(s, 10, 32); err == nil {
		return types.SensitiveTransactionType(v), nil
	}

	// Try exact match
	if v, ok := types.SensitiveTransactionType_value[s]; ok {
		return types.SensitiveTransactionType(v), nil
	}

	// Try with SENSITIVE_TX_ prefix
	prefixed := "SENSITIVE_TX_" + strings.ToUpper(s)
	if v, ok := types.SensitiveTransactionType_value[prefixed]; ok {
		return types.SensitiveTransactionType(v), nil
	}

	// Try uppercase
	if v, ok := types.SensitiveTransactionType_value[strings.ToUpper(s)]; ok {
		return types.SensitiveTransactionType(v), nil
	}

	return 0, fmt.Errorf("invalid transaction type: %s", s)
}
