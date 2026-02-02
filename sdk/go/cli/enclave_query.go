package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	aclient "github.com/virtengine/virtengine/sdk/go/node/client"
	enclavetypes "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

const (
	FlagValidatorAddress = "validator"
	FlagIncludeRevoked   = "include-revoked"
	FlagForBlockHeight   = "for-height"
	FlagCommitteeEpoch   = "epoch"
	FlagScopeID          = "scope-id"
	FlagBlockHeight      = "block-height"
)

// GetQueryEnclaveCmd returns the query commands for enclave module
func GetQueryEnclaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        enclavetypes.ModuleName,
		Short:                      "Enclave query commands",
		Long:                       "Query commands for the enclave module, including identity lookup, key queries, and measurement allowlist.",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryEnclaveIdentityCmd(),
		GetQueryEnclaveKeysCmd(),
		GetQueryEnclaveMeasurementsCmd(),
		GetQueryEnclaveMeasurementCmd(),
		GetQueryEnclaveRotationCmd(),
		GetQueryEnclaveParamsCmd(),
		GetQueryEnclaveValidKeySetCmd(),
		GetQueryEnclaveAttestedResultCmd(),
	)

	return cmd
}

// GetQueryEnclaveIdentityCmd returns the command to query an enclave identity
func GetQueryEnclaveIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity [validator-address]",
		Short: "Query a validator's enclave identity",
		Long: `Query the enclave identity for a specific validator.

Returns the enclave identity including TEE type, measurement hash, encryption keys,
attestation data, and status.

Example:
  $ virtengine query enclave identity virtengine1...
  $ virtengine query enclave identity --validator=virtengine1...`,
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			var validatorAddr string
			if len(args) > 0 {
				validatorAddr = args[0]
			} else {
				var err error
				validatorAddr, err = cmd.Flags().GetString(FlagValidatorAddress)
				if err != nil {
					return err
				}
			}

			if validatorAddr == "" {
				return fmt.Errorf("validator address is required")
			}

			req := &enclavetypes.QueryEnclaveIdentityRequest{
				ValidatorAddress: validatorAddr,
			}

			res, err := queryEnclaveIdentity(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagValidatorAddress, "", "Validator address to query")

	return cmd
}

// GetQueryEnclaveKeysCmd returns the command to query active validator enclave keys
func GetQueryEnclaveKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Query active validator enclave keys",
		Long: `Query all active validator enclave keys.

Returns a list of all validators with active enclave identities, including
their encryption public keys and measurement hashes.

Example:
  $ virtengine query enclave keys`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &enclavetypes.QueryActiveValidatorEnclaveKeysRequest{}

			res, err := queryActiveValidatorEnclaveKeys(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryEnclaveMeasurementsCmd returns the command to query the measurement allowlist
func GetQueryEnclaveMeasurementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Query allowlisted enclave measurements",
		Long: `Query the enclave measurement allowlist.

Returns all measurements that are approved for enclave registration.
Use --tee-type to filter by TEE type, and --include-revoked to include
revoked measurements.

Example:
  $ virtengine query enclave measurements
  $ virtengine query enclave measurements --tee-type=SGX
  $ virtengine query enclave measurements --include-revoked`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			teeType, _ := cmd.Flags().GetString(FlagTEEType)
			includeRevoked, _ := cmd.Flags().GetBool(FlagIncludeRevoked)

			req := &enclavetypes.QueryMeasurementAllowlistRequest{
				TeeType:        teeType,
				IncludeRevoked: includeRevoked,
			}

			res, err := queryMeasurementAllowlist(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagTEEType, "", "Filter by TEE type (SGX, SEV-SNP, NITRO, TRUSTZONE)")
	cmd.Flags().Bool(FlagIncludeRevoked, false, "Include revoked measurements")

	return cmd
}

// GetQueryEnclaveMeasurementCmd returns the command to query a specific measurement
func GetQueryEnclaveMeasurementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "measurement [measurement-hash]",
		Short: "Query a specific enclave measurement",
		Long: `Query a specific enclave measurement from the allowlist.

Returns the measurement details including TEE type, description, minimum ISVSVN,
and whether it is currently allowed.

Example:
  $ virtengine query enclave measurement 0123456789abcdef...
  $ virtengine query enclave measurement --measurement-hash=0123456789abcdef...`,
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			var measurementHashHex string
			if len(args) > 0 {
				measurementHashHex = args[0]
			} else {
				var err error
				measurementHashHex, err = cmd.Flags().GetString(FlagMeasurementHash)
				if err != nil {
					return err
				}
			}

			if measurementHashHex == "" {
				return fmt.Errorf("measurement hash is required")
			}

			req := &enclavetypes.QueryMeasurementRequest{
				MeasurementHash: measurementHashHex,
			}

			res, err := queryMeasurement(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagMeasurementHash, "", "Measurement hash to query (hex-encoded)")

	return cmd
}

// GetQueryEnclaveRotationCmd returns the command to query key rotation status
func GetQueryEnclaveRotationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotation [validator-address]",
		Short: "Query key rotation status for a validator",
		Long: `Query the key rotation status for a validator.

Returns information about any active key rotation, including the overlap period
and new key fingerprint.

Example:
  $ virtengine query enclave rotation virtengine1...
  $ virtengine query enclave rotation --validator=virtengine1...`,
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			var validatorAddr string
			if len(args) > 0 {
				validatorAddr = args[0]
			} else {
				var err error
				validatorAddr, err = cmd.Flags().GetString(FlagValidatorAddress)
				if err != nil {
					return err
				}
			}

			if validatorAddr == "" {
				return fmt.Errorf("validator address is required")
			}

			req := &enclavetypes.QueryKeyRotationRequest{
				ValidatorAddress: validatorAddr,
			}

			res, err := queryKeyRotation(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagValidatorAddress, "", "Validator address to query")

	return cmd
}

// GetQueryEnclaveParamsCmd returns the command to query module parameters
func GetQueryEnclaveParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query enclave module parameters",
		Long: `Query the current enclave module parameters.

Returns parameters including allowed TEE types, expiry settings, rate limits,
and attestation requirements.

Example:
  $ virtengine query enclave params`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &enclavetypes.QueryParamsRequest{}

			res, err := queryEnclaveParams(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryEnclaveValidKeySetCmd returns the command to query valid key set
func GetQueryEnclaveValidKeySetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "valid-keys",
		Short: "Query the current valid key set",
		Long: `Query the set of valid enclave keys for a given block height.

Returns all validator keys that are valid at the specified block height,
including keys in rotation overlap periods.

Example:
  $ virtengine query enclave valid-keys
  $ virtengine query enclave valid-keys --for-height=12345`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			forHeight, _ := cmd.Flags().GetInt64(FlagForBlockHeight)

			req := &enclavetypes.QueryValidKeySetRequest{
				ForBlockHeight: forHeight,
			}

			res, err := queryValidKeySet(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int64(FlagForBlockHeight, 0, "Block height to check validity for (0 for current)")

	return cmd
}

// GetQueryEnclaveAttestedResultCmd returns the command to query an attested result
func GetQueryEnclaveAttestedResultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attested-result",
		Short: "Query an attested scoring result",
		Long: `Query an attested scoring result by block height and scope ID.

Returns the scoring result including the score, enclave measurement, and
validator attestation.

Example:
  $ virtengine query enclave attested-result --block-height=12345 --scope-id=abc123`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			blockHeight, err := cmd.Flags().GetInt64(FlagBlockHeight)
			if err != nil {
				return err
			}
			if blockHeight <= 0 {
				return fmt.Errorf("block height is required and must be positive")
			}

			scopeID, err := cmd.Flags().GetString(FlagScopeID)
			if err != nil {
				return err
			}
			if scopeID == "" {
				return fmt.Errorf("scope ID is required")
			}

			req := &enclavetypes.QueryAttestedResultRequest{
				BlockHeight: blockHeight,
				ScopeId:     scopeID,
			}

			res, err := queryAttestedResult(ctx, cl, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int64(FlagBlockHeight, 0, "Block height of the result")
	cmd.Flags().String(FlagScopeID, "", "Scope ID of the result")

	_ = cmd.MarkFlagRequired(FlagBlockHeight)
	_ = cmd.MarkFlagRequired(FlagScopeID)

	return cmd
}

// Query helper functions - these use the gRPC client to call the enclave module

// queryEnclaveIdentity queries an enclave identity
func queryEnclaveIdentity(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryEnclaveIdentityRequest) (*enclavetypes.QueryEnclaveIdentityResponse, error) {
	return cl.Query().Enclave().EnclaveIdentity(ctx, req)
}

// queryActiveValidatorEnclaveKeys queries all active validator enclave keys
func queryActiveValidatorEnclaveKeys(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryActiveValidatorEnclaveKeysRequest) (*enclavetypes.QueryActiveValidatorEnclaveKeysResponse, error) {
	return cl.Query().Enclave().ActiveValidatorEnclaveKeys(ctx, req)
}

// queryMeasurementAllowlist queries the measurement allowlist
func queryMeasurementAllowlist(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryMeasurementAllowlistRequest) (*enclavetypes.QueryMeasurementAllowlistResponse, error) {
	return cl.Query().Enclave().MeasurementAllowlist(ctx, req)
}

// queryMeasurement queries a specific measurement
func queryMeasurement(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryMeasurementRequest) (*enclavetypes.QueryMeasurementResponse, error) {
	return cl.Query().Enclave().Measurement(ctx, req)
}

// queryKeyRotation queries key rotation status
func queryKeyRotation(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryKeyRotationRequest) (*enclavetypes.QueryKeyRotationResponse, error) {
	return cl.Query().Enclave().KeyRotation(ctx, req)
}

// queryEnclaveParams queries module parameters
func queryEnclaveParams(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryParamsRequest) (*enclavetypes.QueryParamsResponse, error) {
	return cl.Query().Enclave().Params(ctx, req)
}

// queryValidKeySet queries the valid key set
func queryValidKeySet(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryValidKeySetRequest) (*enclavetypes.QueryValidKeySetResponse, error) {
	return cl.Query().Enclave().ValidKeySet(ctx, req)
}

// queryAttestedResult queries an attested result
func queryAttestedResult(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryAttestedResultRequest) (*enclavetypes.QueryAttestedResultResponse, error) {
	return cl.Query().Enclave().AttestedResult(ctx, req)
}
