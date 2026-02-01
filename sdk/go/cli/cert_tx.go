package cli

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/cert/v1"
	utiltls "github.com/virtengine/virtengine/sdk/go/util/tls"
)

const (
	// flagOverwrite = "overwrite"
	flagSerial    = "serial"
	flagValidTime = "valid-duration"
	flagStart     = "start-time"
	flagToGenesis = "to-genesis"
)

var (
	errCertificateDoesNotExist    = fmt.Errorf("%w: does not exist", utiltls.ErrCertificate)
	errCannotOverwriteCertificate = fmt.Errorf("%w: cannot overwrite certificate", utiltls.ErrCertificate)
)

func GetTxCertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Certificates transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	// Commands
	// 1. Generate - create public / private key pair
	// 2. Publish - publish a key pair to the blockchain
	// 3. Revoke - revoke a key pair on the blockchain

	cmd.AddCommand(
		GetTxCertGenerateCmd(),
		GetTxCertPublishCmd(),
		GetTxCertRevokeCmd(),
	)

	return cmd
}

func doCertGenerateCmd(cmd *cobra.Command, domains []string) error {
	allowOverwrite := viper.GetBool(cflags.FlagOverwrite)

	cctx, err := sdkclient.GetClientTxContext(cmd)
	if err != nil {
		return err
	}
	fromAddress := cctx.GetFromAddress()

	kpm, err := utiltls.NewKeyPairManager(cctx, fromAddress)
	if err != nil {
		return err
	}

	exists, err := kpm.KeyExists()
	if err != nil {
		return err
	}
	if !allowOverwrite && exists {
		return errCannotOverwriteCertificate
	}

	var startTime time.Time
	startTimeStr := viper.GetString(flagStart)
	if len(startTimeStr) == 0 {
		startTime = time.Now().Truncate(time.Second)
	} else {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return err
		}
	}
	validDuration := viper.GetDuration(flagValidTime)

	return kpm.Generate(startTime, startTime.Add(validDuration), domains)
}

func doPublishCmd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cl := MustClientFromContext(ctx)
	cctx := cl.ClientContext()

	toGenesis := viper.GetBool(flagToGenesis)

	fromAddress := cctx.GetFromAddress()

	kpm, err := utiltls.NewKeyPairManager(cctx, fromAddress)
	if err != nil {
		return err
	}

	exists, err := kpm.KeyExists()
	if err != nil {
		return err
	}
	if !exists {
		return errCertificateDoesNotExist
	}

	cert, _, pubKey, err := kpm.Read()
	if err != nil {
		return err
	}

	msg := &types.MsgCreateCertificate{
		Owner: fromAddress.String(),
		Cert: pem.EncodeToMemory(&pem.Block{
			Type:  types.PemBlkTypeCertificate,
			Bytes: cert,
		}),
		Pubkey: pem.EncodeToMemory(&pem.Block{
			Type:  types.PemBlkTypeECPublicKey,
			Bytes: pubKey,
		}),
	}

	if err = msg.ValidateBasic(); err != nil {
		return err
	}

	if toGenesis {
		return addCertToGenesis(cmd, types.GenesisCertificate{
			Owner: msg.Owner,
			Certificate: types.Certificate{
				State:  types.CertificateValid,
				Cert:   msg.Cert,
				Pubkey: msg.Pubkey,
			},
		})

	}

	resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
	if err != nil {
		return err
	}

	return cl.PrintMessage(resp)
}

func doRevokeCmd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cl := MustClientFromContext(ctx)
	cctx := cl.ClientContext()

	serial := viper.GetString(flagSerial)

	fromAddress := cctx.GetFromAddress()

	if len(serial) != 0 {
		if _, valid := new(big.Int).SetString(serial, 10); !valid {
			return utiltls.ErrInvalidSerialFlag
		}
	} else {
		kpm, err := utiltls.NewKeyPairManager(cctx, fromAddress)
		if err != nil {
			return err
		}

		cert, _, _, err := kpm.Read()
		if err != nil {
			return err
		}

		parsedCert, err := x509.ParseCertificate(cert)
		if err != nil {
			return err
		}

		serial = parsedCert.SerialNumber.String()
	}

	req := &types.QueryCertificatesRequest{
		Filter: types.CertificateFilter{
			Owner:  fromAddress.String(),
			Serial: serial,
			State:  types.CertificateValid.String(),
		},
	}

	res, err := cl.Query().Certs().Certificates(cmd.Context(), req)
	if err != nil {
		return err
	}

	exists := len(res.Certificates) != 0
	if !exists {
		return fmt.Errorf("%w: certificate with serial %v does not exist on chain and cannot be revoked", utiltls.ErrCertificate, serial)
	}

	msg := &types.MsgRevokeCertificate{
		ID: types.ID{
			Owner:  cctx.FromAddress.String(),
			Serial: serial,
		},
	}

	resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
	if err != nil {
		return err
	}

	return cl.PrintMessage(resp)
}

func GetTxCertGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "generate",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxCertGenerateClientCmd(),
		GetTxCertGenerateServerCmd(),
	)

	return cmd
}

func addTxCertGenerateFlags(cmd *cobra.Command) error {
	cmd.Flags().String(flagStart, "", "certificate is not valid before this date. default current timestamp. RFC3339")
	if err := viper.BindPFlag(flagStart, cmd.Flags().Lookup(flagStart)); err != nil {
		return err
	}

	cmd.Flags().Duration(flagValidTime, time.Hour*24*365, "certificate is not valid after this date. RFC3339")
	if err := viper.BindPFlag(flagValidTime, cmd.Flags().Lookup(flagValidTime)); err != nil {
		return err
	}
	cmd.Flags().Bool(cflags.FlagOverwrite, false, "overwrite existing certificate if present")
	if err := viper.BindPFlag(cflags.FlagOverwrite, cmd.Flags().Lookup(cflags.FlagOverwrite)); err != nil {
		return err
	}

	cflags.AddTxFlagsToCmd(cmd) // TODO - add just the keyring flags? not all the TX ones

	return nil
}

func certGeneratePersistentPreRunE(cmd *cobra.Command, args []string) error {
	err := cmd.Flags().Set(cflags.FlagOffline, "true")
	if err != nil {
		return err
	}

	return TxPersistentPreRunE(cmd, args)
}

func certPublishPersistentPreRunE(cmd *cobra.Command, args []string) error {
	toGenesis := viper.GetBool(flagToGenesis)

	if toGenesis {
		err := cmd.Flags().Set(cflags.FlagOffline, "true")
		if err != nil {
			return err
		}
	}

	return TxPersistentPreRunE(cmd, args)
}
func GetTxCertGenerateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "client",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          certGeneratePersistentPreRunE,
		RunE:                       doCertGenerateCmd,
		SilenceUsage:               true,
		Args:                       cobra.ExactArgs(0),
	}
	err := addTxCertGenerateFlags(cmd)
	if err != nil {
		panic(err)
	}

	return cmd
}

func GetTxCertGenerateServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "server",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          certGeneratePersistentPreRunE,
		RunE:                       doCertGenerateCmd,
		SilenceUsage:               true,
		Args:                       cobra.MinimumNArgs(1),
	}
	err := addTxCertGenerateFlags(cmd)
	if err != nil {
		panic(err)
	}

	return cmd
}

func GetTxCertPublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "publish",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxCertPublishClientCmd(),
		GetTxCertPublishServerCmd())

	return cmd
}

func GetTxCertPublishClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "client",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          certPublishPersistentPreRunE,
		RunE:                       doPublishCmd,
		SilenceUsage:               true,
		Args:                       cobra.ExactArgs(0),
	}
	err := addTxCertPublishFlags(cmd)
	if err != nil {
		panic(err)
	}

	return cmd
}

func GetTxCertPublishServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "server",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          certPublishPersistentPreRunE,
		RunE:                       doPublishCmd,
		SilenceUsage:               true,
		Args:                       cobra.ExactArgs(0),
	}
	err := addTxCertPublishFlags(cmd)
	if err != nil {
		panic(err)
	}

	return cmd
}

func addTxCertPublishFlags(cmd *cobra.Command) error {
	cmd.Flags().Bool(flagToGenesis, false, "add to genesis")
	if err := viper.BindPFlag(flagToGenesis, cmd.Flags().Lookup(flagToGenesis)); err != nil {
		return err
	}

	cflags.AddTxFlagsToCmd(cmd)

	return nil
}

func addCertToGenesis(cmd *cobra.Command, cert types.GenesisCertificate) error {
	cctx, err := sdkclient.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	cdc := cctx.Codec

	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config

	config.SetRoot(cctx.HomeDir)

	if err := cert.Validate(); err != nil {
		return fmt.Errorf("%w: failed to validate new genesis certificate", err)
	}

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal genesis state", err)
	}

	certsGenState := types.GetGenesisStateFromAppState(cdc, appState)

	if certsGenState.Certificates.Contains(cert) {
		return fmt.Errorf("%w: cannot add already existing certificate", err)
	}
	certsGenState.Certificates = append(certsGenState.Certificates, cert)

	certsGenStateBz, err := cdc.MarshalJSON(certsGenState)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal auth genesis state", err)
	}

	appState[types.ModuleName] = certsGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal application genesis state", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

func GetTxCertRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "revoke",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxCertsRevokeClientCmd(),
		GetTxCertRevokeServerCmd())

	return cmd
}

func GetTxCertsRevokeClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "client",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          TxPersistentPreRunE,
		RunE:                       doRevokeCmd,
		SilenceUsage:               true,
		Args:                       cobra.ExactArgs(0),
	}

	err := addRevokeCmdFlags(cmd)

	if err != nil {
		panic(err)
	}

	return cmd
}

func GetTxCertRevokeServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "server",
		Short:                      "",
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE:          TxPersistentPreRunE,
		RunE:                       doRevokeCmd,
		SilenceUsage:               true,
		Args:                       cobra.ExactArgs(0),
	}
	err := addRevokeCmdFlags(cmd)
	if err != nil {
		panic(err)
	}

	return cmd
}

func addRevokeCmdFlags(cmd *cobra.Command) error {
	cmd.Flags().String(flagSerial, "", "revoke certificate by serial number")
	if err := viper.BindPFlag(flagSerial, cmd.Flags().Lookup(flagSerial)); err != nil {
		return err
	}

	cflags.AddTxFlagsToCmd(cmd)
	return nil
}

