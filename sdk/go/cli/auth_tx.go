package cli

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetSignBatchCommand returns the transaction sign-batch command.
func GetSignBatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-batch [file] ([file2]...)",
		Short: "Sign transaction batch files",
		Long: `Sign batch files of transactions generated with --generate-only.
The command processes list of transactions from a file (one StdTx each line), or multiple files.
Then generates signed transactions or signatures and print their JSON encoding, delimited by '\n'.
As the signatures are generated, the command updates the account and sequence number accordingly.

If the --signature-only flag is set, it will output the signature parts only.

The --offline flag makes sure that the client will not reach out to full node.
As a result, the account and the sequence number queries will not be performed and
it is required to set such parameters manually. Note, invalid values will cause
the transaction to fail. The sequence will be incremented automatically for each
transaction that is signed.

If --account-number or --sequence flag is used when offline=false, they are ignored and 
overwritten by the default flag values.

The --multisig=<multisig_key> flag generates a signature on behalf of a multisig
account key. It implies --signature-only.
`,
		PreRun: preSignCmd,
		RunE:   makeSignBatchCmd(),
		Args:   cobra.MinimumNArgs(1),
	}

	cmd.Flags().String(cflags.FlagMultisig, "", "Address or key name of the multisig account on behalf of which the transaction shall be signed")
	cmd.Flags().String(cflags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().Bool(cflags.FlagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().Bool(cflags.FlagAppend, false, "Combine all message and generate single signed transaction for broadcast.")

	cflags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(cflags.FlagFrom)

	return cmd
}

func makeSignBatchCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		txFactory, err := tx.NewFactoryCLI(cctx, cmd.Flags())
		if err != nil {
			return err
		}
		txCfg := cctx.TxConfig
		printSignatureOnly, _ := cmd.Flags().GetBool(cflags.FlagSigOnly)

		ms, err := cmd.Flags().GetString(cflags.FlagMultisig)
		if err != nil {
			return err
		}

		// prepare output document
		closeFunc, err := setOutputFile(cmd)
		if err != nil {
			return err
		}
		defer closeFunc()
		cctx.WithOutput(cmd.OutOrStdout())

		// reads tx from args
		scanner, err := ReadTxsFromInput(txCfg, args...)
		if err != nil {
			return err
		}

		if !cctx.Offline {
			if ms == "" {
				from, err := cmd.Flags().GetString(cflags.FlagFrom)
				if err != nil {
					return err
				}

				addr, _, _, err := client.GetFromFields(cctx, txFactory.Keybase(), from)
				if err != nil {
					return err
				}

				acc, err := txFactory.AccountRetriever().GetAccount(cctx, addr)
				if err != nil {
					return err
				}

				txFactory = txFactory.WithAccountNumber(acc.GetAccountNumber()).WithSequence(acc.GetSequence())
			} else {
				txFactory = txFactory.WithAccountNumber(0).WithSequence(0)
			}
		}

		appendMessagesToSingleTx, _ := cmd.Flags().GetBool(cflags.FlagAppend)
		// Combines all tx msgs and create single signed transaction
		if appendMessagesToSingleTx {
			txBuilder := cctx.TxConfig.NewTxBuilder()
			msgs := make([]sdk.Msg, 0)
			newGasLimit := uint64(0)

			for scanner.Scan() {
				unsignedStdTx := scanner.Tx()
				fe, err := cctx.TxConfig.WrapTxBuilder(unsignedStdTx)
				if err != nil {
					return err
				}
				// increment the gas
				newGasLimit += fe.GetTx().GetGas()
				// append messages
				msgs = append(msgs, unsignedStdTx.GetMsgs()...)
			}
			// set the new appened msgs into builder
			_ = txBuilder.SetMsgs(msgs...)

			// set the memo,fees,feeGranter,feePayer from cmd flags
			txBuilder.SetMemo(txFactory.Memo())
			txBuilder.SetFeeAmount(txFactory.Fees())
			txBuilder.SetFeeGranter(cctx.FeeGranter)
			txBuilder.SetFeePayer(cctx.FeePayer)

			// set the gasLimit
			txBuilder.SetGasLimit(newGasLimit)

			// sign the txs
			if ms == "" {
				from, _ := cmd.Flags().GetString(cflags.FlagFrom)
				if err := sign(cctx, txBuilder, txFactory, from); err != nil {
					return err
				}
			} else {
				if err := multisigSign(cctx, txBuilder, txFactory, ms); err != nil {
					return err
				}
			}

			json, err := marshalSignatureJSON(txCfg, txBuilder, printSignatureOnly)
			if err != nil {
				return err
			}

			cmd.Printf("%s\n", json)
		} else {
			// It will generate signed tx for each tx
			for sequence := txFactory.Sequence(); scanner.Scan(); sequence++ {
				unsignedStdTx := scanner.Tx()
				txFactory = txFactory.WithSequence(sequence)
				txBuilder, err := txCfg.WrapTxBuilder(unsignedStdTx)
				if err != nil {
					return err
				}

				// sign the txs
				if ms == "" {
					from, _ := cmd.Flags().GetString(cflags.FlagFrom)
					if err := sign(cctx, txBuilder, txFactory, from); err != nil {
						return err
					}
				} else {
					if err := multisigSign(cctx, txBuilder, txFactory, ms); err != nil {
						return err
					}
				}

				json, err := marshalSignatureJSON(txCfg, txBuilder, printSignatureOnly)
				if err != nil {
					return err
				}
				cmd.Printf("%s\n", json)
			}
		}

		if err := scanner.UnmarshalErr(); err != nil {
			return err
		}

		return scanner.UnmarshalErr()
	}
}

func sign(clientCtx client.Context, txBuilder client.TxBuilder, txFactory tx.Factory, from string) error {
	_, fromName, _, err := client.GetFromFields(clientCtx, txFactory.Keybase(), from)
	if err != nil {
		return fmt.Errorf("error getting account from keybase: %w", err)
	}

	if err = SignTx(txFactory, clientCtx, fromName, txBuilder, true, true); err != nil {
		return err
	}

	return nil
}

func multisigSign(clientCtx client.Context, txBuilder client.TxBuilder, txFactory tx.Factory, multisig string) error {
	multisigAddr, multisigName, _, err := client.GetFromFields(clientCtx, txFactory.Keybase(), multisig)
	if err != nil {
		return fmt.Errorf("error getting account from keybase: %w", err)
	}

	fromRecord, err := clientCtx.Keyring.Key(clientCtx.FromName)
	if err != nil {
		return fmt.Errorf("error getting account from keybase: %w", err)
	}

	fromPubKey, err := fromRecord.GetPubKey()
	if err != nil {
		return err
	}

	multisigkey, err := clientCtx.Keyring.Key(multisigName)
	if err != nil {
		return err
	}

	multisigPubKey, err := multisigkey.GetPubKey()
	if err != nil {
		return err
	}

	isSigner, err := isMultisigSigner(clientCtx, multisigPubKey, fromPubKey)
	if err != nil {
		return err
	}

	if !isSigner {
		return fmt.Errorf("signing key is not a part of multisig key")
	}

	if err = authclient.SignTxWithSignerAddress(
		txFactory,
		clientCtx,
		multisigAddr,
		clientCtx.FromName,
		txBuilder,
		clientCtx.Offline,
		true,
	); err != nil {
		return err
	}

	return nil
}

// isMultisigSigner checks if the given pubkey is a signer in the multisig or in
// any of the nested multisig signers.
func isMultisigSigner(cctx client.Context, multisigPubKey, fromPubKey cryptotypes.PubKey) (bool, error) { //nolint unparam
	multisigLegacyPub := multisigPubKey.(*kmultisig.LegacyAminoPubKey)

	var found bool
	for _, pubkey := range multisigLegacyPub.GetPubKeys() {
		if pubkey.Equals(fromPubKey) {
			found = true
			break
		}

		if nestedMultisig, ok := pubkey.(*kmultisig.LegacyAminoPubKey); ok {
			var err error
			found, err = isMultisigSigner(cctx, nestedMultisig, fromPubKey)
			if err != nil {
				return false, err
			}
			if found {
				break
			}
		}
	}

	return found, nil
}

func setOutputFile(cmd *cobra.Command) (func(), error) {
	outputDoc, _ := cmd.Flags().GetString(cflags.FlagOutputDocument)
	if outputDoc == "" {
		return func() {}, nil
	}

	fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644) //nolint: gosec
	if err != nil {
		return func() {}, err
	}

	cmd.SetOut(fp)

	return func() { _ = fp.Close() }, nil
}

// GetSignCommand returns the transaction sign command.
func GetSignCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [file]",
		Short: "Sign a transaction generated offline",
		Long: `Sign a transaction created with the --generate-only flag.
It will read a transaction from [file], sign it, and print its JSON encoding.

If the --signature-only flag is set, it will output the signature parts only.

The --offline flag makes sure that the client will not reach out to full node.
As a result, the account and sequence number queries will not be performed and
it is required to set such parameters manually. Note, invalid values will cause
the transaction to fail.

The --multisig=<multisig_key> flag generates a signature on behalf of a multisig account
key. It implies --signature-only. Full multisig signed transactions may eventually
be generated via the 'multisign' command.
`,
		PreRun: preSignCmd,
		RunE:   makeSignCmd(),
		Args:   cobra.ExactArgs(1),
	}

	cmd.Flags().String(cflags.FlagMultisig, "", "Address or key name of the multisig account on behalf of which the transaction shall be signed")
	cmd.Flags().Bool(cflags.FlagOverwrite, false, "Overwrite existing signatures with a new one. If disabled, new signature will be appended")
	cmd.Flags().Bool(cflags.FlagSigOnly, false, "Print only the signatures")
	cmd.Flags().String(cflags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cflags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(cflags.FlagFrom)

	return cmd
}

func preSignCmd(cmd *cobra.Command, _ []string) {
	// Conditionally mark the account and sequence numbers required as no RPC
	// query will be done.
	if offline, _ := cmd.Flags().GetBool(cflags.FlagOffline); offline {
		_ = cmd.MarkFlagRequired(cflags.FlagAccountNumber)
		_ = cmd.MarkFlagRequired(cflags.FlagSequence)
	}
}

func makeSignCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		cctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		cctx, txF, newTx, err := readTxAndInitContexts(cctx, cmd, args[0])
		if err != nil {
			return err
		}

		return signTx(cmd, cctx, txF, newTx)
	}
}

func signTx(cmd *cobra.Command, clientCtx client.Context, txF tx.Factory, newTx sdk.Tx) error {
	f := cmd.Flags()
	txCfg := clientCtx.TxConfig
	txBuilder, err := txCfg.WrapTxBuilder(newTx)
	if err != nil {
		return err
	}

	printSignatureOnly, err := cmd.Flags().GetBool(cflags.FlagSigOnly)
	if err != nil {
		return err
	}

	multisig, err := cmd.Flags().GetString(cflags.FlagMultisig)
	if err != nil {
		return err
	}

	from, err := cmd.Flags().GetString(cflags.FlagFrom)
	if err != nil {
		return err
	}

	_, fromName, _, err := client.GetFromFields(clientCtx, txF.Keybase(), from)
	if err != nil {
		return fmt.Errorf("error getting account from keybase: %w", err)
	}

	overwrite, err := f.GetBool(cflags.FlagOverwrite)
	if err != nil {
		return err
	}

	if multisig != "" {
		// Bech32 decode error, maybe it's a name, we try to fetch from keyring
		multisigAddr, multisigName, _, err := client.GetFromFields(clientCtx, txF.Keybase(), multisig)
		if err != nil {
			return fmt.Errorf("error getting account from keybase: %w", err)
		}
		multisigkey, err := getMultisigRecord(clientCtx, multisigName)
		if err != nil {
			return err
		}
		multisigPubKey, err := multisigkey.GetPubKey()
		if err != nil {
			return err
		}

		fromRecord, err := clientCtx.Keyring.Key(fromName)
		if err != nil {
			return fmt.Errorf("error getting account from keybase: %w", err)
		}
		fromPubKey, err := fromRecord.GetPubKey()
		if err != nil {
			return err
		}

		isSigner, err := isMultisigSigner(clientCtx, multisigPubKey, fromPubKey)
		if err != nil {
			return err
		}
		if !isSigner {
			return fmt.Errorf("signing key is not a part of multisig key")
		}

		err = authclient.SignTxWithSignerAddress(
			txF, clientCtx, multisigAddr, fromName, txBuilder, clientCtx.Offline, overwrite)
		if err != nil {
			return err
		}
		printSignatureOnly = true
	} else {
		err = authclient.SignTx(txF, clientCtx, clientCtx.FromName, txBuilder, clientCtx.Offline, overwrite)
	}
	if err != nil {
		return err
	}

	// set output
	closeFunc, err := setOutputFile(cmd)
	if err != nil {
		return err
	}

	defer closeFunc()
	clientCtx.WithOutput(cmd.OutOrStdout())

	var json []byte
	json, err = marshalSignatureJSON(txCfg, txBuilder, printSignatureOnly)
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", json)

	return err
}

func marshalSignatureJSON(txConfig client.TxConfig, txBldr client.TxBuilder, signatureOnly bool) ([]byte, error) {
	parsedTx := txBldr.GetTx()
	if signatureOnly {
		sigs, err := parsedTx.GetSignaturesV2()
		if err != nil {
			return nil, err
		}
		return txConfig.MarshalSignatureJSON(sigs)
	}

	return txConfig.TxJSONEncoder()(parsedTx)
}

func GetValidateSignaturesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-signatures [file]",
		Short: "validate transactions signatures",
		Long: `Print the addresses that must sign the transaction, those who have already
signed it, and make sure that signatures are in the correct order.

The command would check whether all required signers have signed the transactions, whether
the signatures were collected in the right order, and if the signature is valid over the
given transaction. If the --offline flag is also set, signature validation over the
transaction will be not be performed as that will require RPC communication with a full node.
`,
		PreRun: preSignCmd,
		RunE:   makeValidateSignaturesCmd(),
		Args:   cobra.ExactArgs(1),
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

func makeValidateSignaturesCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		cctx, txBldr, stdTx, err := readTxAndInitContexts(cctx, cmd, args[0])
		if err != nil {
			return err
		}

		if !printAndValidateSigs(cmd, cctx, txBldr.ChainID(), stdTx, cctx.Offline) {
			return fmt.Errorf("signatures validation failed")
		}

		return nil
	}
}

// printAndValidateSigs will validate the signatures of a given transaction over its
// expected signers. In addition, if offline has not been supplied, the signature is
// verified over the transaction sign bytes. Returns false if the validation fails.
func printAndValidateSigs(
	cmd *cobra.Command, clientCtx client.Context, chainID string, tx sdk.Tx, offline bool,
) bool {
	sigTx := tx.(authsigning.SigVerifiableTx)
	signModeHandler := clientCtx.TxConfig.SignModeHandler()
	addrCdc := clientCtx.TxConfig.SigningContext().AddressCodec()

	cmd.Println("Signers:")
	signers, err := sigTx.GetSigners()
	if err != nil {
		panic(err)
	}

	for i, signer := range signers {
		signerStr, err := addrCdc.BytesToString(signer)
		if err != nil {
			panic(err)
		}
		cmd.Printf("  %v: %v\n", i, signerStr)
	}

	success := true
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		panic(err)
	}
	cmd.Println("")
	cmd.Println("Signatures:")

	if len(sigs) != len(signers) {
		success = false
	}

	for i, sig := range sigs {
		var (
			pubKey         = sig.PubKey
			multiSigHeader string
			multiSigMsg    string
			sigAddr        = sdk.AccAddress(pubKey.Address())
			sigSanity      = "OK"
		)

		if i >= len(signers) || !bytes.Equal(sigAddr, signers[i]) {
			sigSanity = "ERROR: signature does not match its respective signer"
			success = false
		}

		// validate the actual signature over the transaction bytes since we can
		// reach out to a full node to query accounts.
		if !offline && success {
			accNum, accSeq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, sigAddr)
			if err != nil {
				cmd.PrintErrf("failed to get account: %s\n", sigAddr)
				return false
			}

			signingData := authsigning.SignerData{
				Address:       sigAddr.String(),
				ChainID:       chainID,
				AccountNumber: accNum,
				Sequence:      accSeq,
				PubKey:        pubKey,
			}
			anyPk, err := codectypes.NewAnyWithValue(pubKey)
			if err != nil {
				cmd.PrintErrf("failed to pack public key: %v", err)
				return false
			}
			txSignerData := txsigning.SignerData{
				ChainID:       signingData.ChainID,
				AccountNumber: signingData.AccountNumber,
				Sequence:      signingData.Sequence,
				Address:       signingData.Address,
				PubKey: &anypb.Any{
					TypeUrl: anyPk.TypeUrl,
					Value:   anyPk.Value,
				},
			}

			adaptableTx, ok := tx.(authsigning.V2AdaptableTx)
			if !ok {
				cmd.PrintErrf("expected V2AdaptableTx, got %T", tx)
				return false
			}
			txData := adaptableTx.GetSigningTxData()

			err = authsigning.VerifySignature(cmd.Context(), pubKey, txSignerData, sig.Data, signModeHandler, txData)
			if err != nil {
				cmd.PrintErrf("failed to verify signature: %v", err)
				return false
			}
		}

		cmd.Printf("  %d: %s\t\t\t[%s]%s%s\n", i, sigAddr.String(), sigSanity, multiSigHeader, multiSigMsg)
	}

	cmd.Println("")

	return success
}

func readTxAndInitContexts(clientCtx client.Context, cmd *cobra.Command, filename string) (client.Context, tx.Factory, sdk.Tx, error) {
	stdTx, err := authclient.ReadTxFromFile(clientCtx, filename)
	if err != nil {
		return clientCtx, tx.Factory{}, nil, err
	}

	txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
	if err != nil {
		return clientCtx, tx.Factory{}, nil, err
	}

	return clientCtx, txFactory, stdTx, nil
}
