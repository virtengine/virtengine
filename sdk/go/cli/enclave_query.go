package cli

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
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
			cl := MustQueryClientFromContext(ctx)

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

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

			req := &enclavetypes.QueryActiveValidatorEnclaveKeysRequest{}

			res, err := queryActiveValidatorEnclaveKeys(ctx, cl, req)
			if err != nil {
				return err
			}

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

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

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

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

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

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

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

			req := &enclavetypes.QueryParamsRequest{}

			res, err := queryEnclaveParams(ctx, cl, req)
			if err != nil {
				return err
			}

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

			forHeight, _ := cmd.Flags().GetInt64(FlagForBlockHeight)

			req := &enclavetypes.QueryValidKeySetRequest{
				ForBlockHeight: forHeight,
			}

			res, err := queryValidKeySet(ctx, cl, req)
			if err != nil {
				return err
			}

			return printJSON(cmd, res)
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
			cl := MustQueryClientFromContext(ctx)

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

			return printJSON(cmd, res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int64(FlagBlockHeight, 0, "Block height of the result")
	cmd.Flags().String(FlagScopeID, "", "Scope ID of the result")

	_ = cmd.MarkFlagRequired(FlagBlockHeight)
	_ = cmd.MarkFlagRequired(FlagScopeID)

	return cmd
}

// Query helper functions - these make gRPC calls to the enclave module
// In production, these would use the actual gRPC client

// MustQueryClientFromContext is an alias for MustLightClientFromContext
// that provides enclave query functionality
func MustQueryClientFromContext(ctx context.Context) aclient.LightClient {
	return MustLightClientFromContext(ctx)
}

// queryEnclaveIdentity queries an enclave identity
func queryEnclaveIdentity(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryEnclaveIdentityRequest) (*enclavetypes.QueryEnclaveIdentityResponse, error) {
	// This uses the standard REST/gRPC query path
	// The actual implementation would use cl.Query().Enclave().EnclaveIdentity(ctx, req)
	// For now, we simulate via the ABCIQuery interface

	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/EnclaveIdentity")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryEnclaveIdentityResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryActiveValidatorEnclaveKeys queries all active validator enclave keys
func queryActiveValidatorEnclaveKeys(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryActiveValidatorEnclaveKeysRequest) (*enclavetypes.QueryActiveValidatorEnclaveKeysResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/ActiveValidatorEnclaveKeys")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryActiveValidatorEnclaveKeysResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryMeasurementAllowlist queries the measurement allowlist
func queryMeasurementAllowlist(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryMeasurementAllowlistRequest) (*enclavetypes.QueryMeasurementAllowlistResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/MeasurementAllowlist")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryMeasurementAllowlistResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryMeasurement queries a specific measurement
func queryMeasurement(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryMeasurementRequest) (*enclavetypes.QueryMeasurementResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/Measurement")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryMeasurementResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryKeyRotation queries key rotation status
func queryKeyRotation(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryKeyRotationRequest) (*enclavetypes.QueryKeyRotationResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/KeyRotation")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryKeyRotationResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryEnclaveParams queries module parameters
func queryEnclaveParams(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryParamsRequest) (*enclavetypes.QueryParamsResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/Params")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryParamsResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryValidKeySet queries the valid key set
func queryValidKeySet(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryValidKeySetRequest) (*enclavetypes.QueryValidKeySetResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/ValidKeySet")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryValidKeySetResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// queryAttestedResult queries an attested result
func queryAttestedResult(ctx context.Context, cl aclient.LightClient, req *enclavetypes.QueryAttestedResultRequest) (*enclavetypes.QueryAttestedResultResponse, error) {
	queryPath := fmt.Sprintf("/virtengine.enclave.v1.Query/AttestedResult")
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := abciQuery(ctx, cl, queryPath, reqBytes)
	if err != nil {
		return nil, err
	}

	var res enclavetypes.QueryAttestedResultResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// abciQuery performs an ABCI query using the client context
func abciQuery(_ context.Context, cl aclient.LightClient, path string, data []byte) ([]byte, error) {
	cctx := cl.ClientContext()
	resp, err := cctx.QueryABCI(abci.RequestQuery{
		Path: path,
		Data: data,
	})
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("query failed with code %d: %s", resp.Code, resp.Log)
	}

	return resp.Value, nil
}

// printJSON prints a response as JSON
func printJSON(cmd *cobra.Command, v interface{}) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(out))
	return nil
}


