package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// GetTxVEIDCmd returns the transaction commands for the VEID module
func GetTxVEIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        veidv1.ModuleName,
		Short:                      "VEID transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxVEIDRequestVerificationCmd(),
		GetTxVEIDUploadScopeCmd(),
		GetTxVEIDRevokeScopeCmd(),
		GetTxVEIDCreateWalletCmd(),
		GetTxVEIDUpdateConsentCmd(),
		GetTxVEIDUpdateVerificationCmd(),
		GetTxVEIDUpdateScoreCmd(),
		GetTxVEIDRegisterModelCmd(),
		GetTxVEIDProposeModelUpdateCmd(),
	)
	return cmd
}

// GetTxVEIDRequestVerificationCmd returns the command to request identity verification
func GetTxVEIDRequestVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "request-verification [scope-id]",
		Short:             "Request identity verification for a scope",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := &veidv1.MsgRequestVerification{
				Sender:  cctx.FromAddress.String(),
				ScopeId: args[0],
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetTxVEIDUploadScopeCmd returns the command to upload an identity scope
func GetTxVEIDUploadScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-scope [scope-id] [scope-type]",
		Short: "Upload an encrypted identity scope",
		Long: `Upload an encrypted identity scope with client and user signatures.

Scope types:
  - id_document: Government-issued ID documents (passport, driver's license)
  - selfie: Selfie photo for face verification
  - face_video: Video for liveness detection
  - biometric: Biometric data (fingerprint, voice, etc.)
  - sso_metadata: SSO provider metadata pointers
  - email_proof: Email verification proof
  - sms_proof: SMS/phone verification proof
  - domain_verify: Domain ownership verification
  - ad_sso: Active Directory SSO verification

Example:
  virtengine tx veid upload-scope my-scope-123 selfie \
    --encrypted-payload-file payload.bin \
    --algorithm-id X25519-XSalsa20-Poly1305 \
    --recipient-key-id validator1 \
    --recipient-pubkey <hex> \
    --encrypted-key <hex> \
    --nonce <hex> \
    --sender-signature <hex> \
    --sender-pubkey <hex> \
    --salt <hex> \
    --device-fingerprint device123 \
    --client-id approved-client-1 \
    --client-signature <hex> \
    --user-signature <hex> \
    --payload-hash <hex> \
    --capture-timestamp 1234567890 \
    --geo-hint US-CA
`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			scopeID := args[0]
			scopeTypeStr := args[1]

			// Parse scope type
			scopeType := veidtypes.ScopeType(scopeTypeStr)
			if !veidtypes.IsValidScopeType(scopeType) {
				return fmt.Errorf("invalid scope type: %s", scopeTypeStr)
			}

			// Read encrypted payload from file
			payloadFile, err := cmd.Flags().GetString("encrypted-payload-file")
			if err != nil {
				return err
			}
			ciphertext, err := os.ReadFile(payloadFile)
			if err != nil {
				return fmt.Errorf("failed to read encrypted payload file: %w", err)
			}

			// Parse encryption envelope fields
			algorithmID, _ := cmd.Flags().GetString("algorithm-id")
			algorithmVersion, _ := cmd.Flags().GetString("algorithm-version")
			recipientKeyID, _ := cmd.Flags().GetString("recipient-key-id")
			recipientPubkeyHex, _ := cmd.Flags().GetString("recipient-pubkey")
			encryptedKeyHex, _ := cmd.Flags().GetString("encrypted-key")
			nonceHex, _ := cmd.Flags().GetString("nonce")
			senderSigHex, _ := cmd.Flags().GetString("sender-signature")
			senderPubkeyHex, _ := cmd.Flags().GetString("sender-pubkey")

			recipientPubkey, err := hex.DecodeString(recipientPubkeyHex)
			if err != nil {
				return fmt.Errorf("invalid recipient-pubkey hex: %w", err)
			}
			encryptedKey, err := hex.DecodeString(encryptedKeyHex)
			if err != nil {
				return fmt.Errorf("invalid encrypted-key hex: %w", err)
			}
			nonce, err := hex.DecodeString(nonceHex)
			if err != nil {
				return fmt.Errorf("invalid nonce hex: %w", err)
			}
			senderSig, err := hex.DecodeString(senderSigHex)
			if err != nil {
				return fmt.Errorf("invalid sender-signature hex: %w", err)
			}
			senderPubkey, err := hex.DecodeString(senderPubkeyHex)
			if err != nil {
				return fmt.Errorf("invalid sender-pubkey hex: %w", err)
			}

			// Parse signature fields
			saltHex, _ := cmd.Flags().GetString("salt")
			deviceFingerprint, _ := cmd.Flags().GetString("device-fingerprint")
			clientID, _ := cmd.Flags().GetString("client-id")
			clientSigHex, _ := cmd.Flags().GetString("client-signature")
			userSigHex, _ := cmd.Flags().GetString("user-signature")
			payloadHashHex, _ := cmd.Flags().GetString("payload-hash")

			salt, err := hex.DecodeString(saltHex)
			if err != nil {
				return fmt.Errorf("invalid salt hex: %w", err)
			}
			clientSig, err := hex.DecodeString(clientSigHex)
			if err != nil {
				return fmt.Errorf("invalid client-signature hex: %w", err)
			}
			userSig, err := hex.DecodeString(userSigHex)
			if err != nil {
				return fmt.Errorf("invalid user-signature hex: %w", err)
			}
			payloadHash, err := hex.DecodeString(payloadHashHex)
			if err != nil {
				return fmt.Errorf("invalid payload-hash hex: %w", err)
			}

			// Optional fields
			captureTimestamp, _ := cmd.Flags().GetInt64("capture-timestamp")
			if captureTimestamp == 0 {
				captureTimestamp = time.Now().Unix()
			}
			geoHint, _ := cmd.Flags().GetString("geo-hint")

			// Convert algorithm version string to uint32
			algorithmVer, err := strconv.ParseUint(algorithmVersion, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid algorithm version: %w", err)
			}

			// Build encryption envelope
			encryptedPayload := encryptiontypes.EncryptedPayloadEnvelope{
				Version:             1,
				AlgorithmID:         algorithmID,
				AlgorithmVersion:    uint32(algorithmVer),
				RecipientKeyIDs:     []string{recipientKeyID},
				RecipientPublicKeys: [][]byte{recipientPubkey},
				EncryptedKeys:       [][]byte{encryptedKey},
				Nonce:               nonce,
				Ciphertext:          ciphertext,
				SenderSignature:     senderSig,
				SenderPubKey:        senderPubkey,
				Metadata:            make(map[string]string),
			}

			// Create message using constructor
			msg := veidtypes.NewMsgUploadScope(
				cctx.FromAddress.String(),
				scopeID,
				scopeType,
				encryptedPayload,
				salt,
				deviceFingerprint,
				clientID,
				clientSig,
				userSig,
				payloadHash,
			)

			// Set optional fields directly on the message
			msg.CaptureTimestamp = captureTimestamp
			msg.GeoHint = geoHint

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	// Encryption envelope flags
	cmd.Flags().String("encrypted-payload-file", "", "Path to encrypted payload file (required)")
	cmd.Flags().String("algorithm-id", "X25519-XSalsa20-Poly1305", "Encryption algorithm ID")
	cmd.Flags().String("algorithm-version", "1", "Algorithm version")
	cmd.Flags().String("recipient-key-id", "", "Recipient validator key ID (required)")
	cmd.Flags().String("recipient-pubkey", "", "Recipient public key (hex) (required)")
	cmd.Flags().String("encrypted-key", "", "Encrypted symmetric key (hex) (required)")
	cmd.Flags().String("nonce", "", "Encryption nonce (hex) (required)")
	cmd.Flags().String("sender-signature", "", "Sender signature over payload (hex) (required)")
	cmd.Flags().String("sender-pubkey", "", "Sender public key (hex) (required)")

	// Signature fields
	cmd.Flags().String("salt", "", "Unique salt for cryptographic binding (hex) (required)")
	cmd.Flags().String("device-fingerprint", "", "Device fingerprint (required)")
	cmd.Flags().String("client-id", "", "Approved client ID (required)")
	cmd.Flags().String("client-signature", "", "Client signature (hex) (required)")
	cmd.Flags().String("user-signature", "", "User signature (hex) (required)")
	cmd.Flags().String("payload-hash", "", "Hash of encrypted payload (hex) (required)")

	// Optional fields
	cmd.Flags().Int64("capture-timestamp", 0, "Capture timestamp (Unix seconds, default: now)")
	cmd.Flags().String("geo-hint", "", "Geographic hint (e.g., US-CA)")

	_ = cmd.MarkFlagRequired("encrypted-payload-file")
	_ = cmd.MarkFlagRequired("recipient-key-id")
	_ = cmd.MarkFlagRequired("recipient-pubkey")
	_ = cmd.MarkFlagRequired("encrypted-key")
	_ = cmd.MarkFlagRequired("nonce")
	_ = cmd.MarkFlagRequired("sender-signature")
	_ = cmd.MarkFlagRequired("sender-pubkey")
	_ = cmd.MarkFlagRequired("salt")
	_ = cmd.MarkFlagRequired("device-fingerprint")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("client-signature")
	_ = cmd.MarkFlagRequired("user-signature")
	_ = cmd.MarkFlagRequired("payload-hash")

	return cmd
}

// GetTxVEIDRevokeScopeCmd returns the command to revoke an identity scope
func GetTxVEIDRevokeScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-scope [scope-id]",
		Short: "Revoke an identity scope",
		Long: `Revoke an identity scope, removing it from consideration for verification.

Example:
  virtengine tx veid revoke-scope my-scope-123 --reason "No longer valid"
`,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			scopeID := args[0]
			reason, _ := cmd.Flags().GetString("reason")

			msg := &veidv1.MsgRevokeScope{
				Sender:  cctx.FromAddress.String(),
				ScopeId: scopeID,
				Reason:  reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("reason", "", "Reason for revocation")

	return cmd
}

// GetTxVEIDCreateWalletCmd returns the command to create an identity wallet
func GetTxVEIDCreateWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-wallet",
		Short: "Create an identity wallet",
		Long: `Create an identity wallet to manage identity scopes and consent.

Example:
  virtengine tx veid create-wallet --binding-signature <hex> --binding-pubkey <hex>
`,
		Args:              cobra.NoArgs,
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			bindingSig, _ := cmd.Flags().GetString("binding-signature")
			bindingPubKey, _ := cmd.Flags().GetString("binding-pubkey")

			bindingSigBytes, err := hex.DecodeString(bindingSig)
			if err != nil {
				return fmt.Errorf("invalid binding signature hex: %w", err)
			}

			bindingPubKeyBytes, err := hex.DecodeString(bindingPubKey)
			if err != nil {
				return fmt.Errorf("invalid binding pubkey hex: %w", err)
			}

			msg := &veidv1.MsgCreateIdentityWallet{
				Sender:           cctx.FromAddress.String(),
				BindingSignature: bindingSigBytes,
				BindingPubKey:    bindingPubKeyBytes,
				InitialConsent:   nil,
				Metadata:         make(map[string]string),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("binding-signature", "", "Binding signature proving account ownership (hex)")
	cmd.Flags().String("binding-pubkey", "", "Binding public key (hex)")
	_ = cmd.MarkFlagRequired("binding-signature")
	_ = cmd.MarkFlagRequired("binding-pubkey")

	return cmd
}

// GetTxVEIDUpdateConsentCmd returns the command to update consent settings
func GetTxVEIDUpdateConsentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-consent",
		Short: "Update consent settings",
		Long: `Update consent settings for data processing and sharing.

Example:
  virtengine tx veid update-consent \
    --scope-id my-scope \
    --grant-consent=true \
    --purpose "identity verification" \
    --expires-at 1234567890 \
    --user-signature <hex>
`,
		Args:              cobra.NoArgs,
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			scopeID, _ := cmd.Flags().GetString("scope-id")
			grantConsent, _ := cmd.Flags().GetBool("grant-consent")
			purpose, _ := cmd.Flags().GetString("purpose")
			expiresAt, _ := cmd.Flags().GetInt64("expires-at")
			userSig, _ := cmd.Flags().GetString("user-signature")

			userSigBytes, err := hex.DecodeString(userSig)
			if err != nil {
				return fmt.Errorf("invalid user signature hex: %w", err)
			}

			msg := &veidv1.MsgUpdateConsentSettings{
				Sender:        cctx.FromAddress.String(),
				ScopeId:       scopeID,
				GrantConsent:  grantConsent,
				Purpose:       purpose,
				ExpiresAt:     expiresAt,
				UserSignature: userSigBytes,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("scope-id", "", "Scope ID (empty for global consent settings)")
	cmd.Flags().Bool("grant-consent", true, "Grant or revoke consent")
	cmd.Flags().String("purpose", "", "Purpose for granting consent")
	cmd.Flags().Int64("expires-at", 0, "Expiration timestamp (Unix seconds)")
	cmd.Flags().String("user-signature", "", "User signature authorizing consent update (hex)")
	_ = cmd.MarkFlagRequired("user-signature")

	return cmd
}

// GetTxVEIDUpdateVerificationCmd returns the command to update verification status (VALIDATOR ONLY)
func GetTxVEIDUpdateVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-verification [account-address] [scope-id] [status]",
		Short: "Update verification status (VALIDATOR ONLY)",
		Long: `Update the verification status of an identity scope.

This command is restricted to validators only. The sender must be a registered
validator with authority to update verification status.

Valid status values:
  - unknown: Status unknown
  - pending: Verification pending
  - verified: Successfully verified
  - rejected: Verification rejected
  - expired: Verification expired
  - revoked: Verification revoked

Example:
  virtengine tx veid update-verification virtengine1abc... scope-123 verified \
    --confidence-score 95 \
    --reason "Face match confirmed"
`,
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			accountAddress := args[0]
			scopeID := args[1]
			statusStr := args[2]

			// Parse verification status
			var status veidv1.VerificationStatus
			switch statusStr {
			case "unknown":
				status = veidv1.VerificationStatusUnknown
			case "pending":
				status = veidv1.VerificationStatusPending
			case "verified":
				status = veidv1.VerificationStatusVerified
			case "rejected":
				status = veidv1.VerificationStatusRejected
			case "expired":
				status = veidv1.VerificationStatusExpired
			case "in-progress":
				status = veidv1.VerificationStatusInProgress
			case "needs-additional-factor":
				status = veidv1.VerificationStatusNeedsAdditionalFactor
			case "additional-factor-pending":
				status = veidv1.VerificationStatusAdditionalFactorPending
			default:
				return fmt.Errorf("invalid verification status: %s", statusStr)
			}

			reason, _ := cmd.Flags().GetString("reason")

			msg := &veidv1.MsgUpdateVerificationStatus{
				Sender:         cctx.FromAddress.String(),
				AccountAddress: accountAddress,
				ScopeId:        scopeID,
				NewStatus:      status,
				Reason:         reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("reason", "", "Reason for status update")
	_ = cmd.MarkFlagRequired("reason")

	return cmd
}

// GetTxVEIDUpdateScoreCmd returns the command to update VEID score (VALIDATOR ONLY)
func GetTxVEIDUpdateScoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-score [account-address] [new-score]",
		Short: "Update VEID score (VALIDATOR ONLY)",
		Long: `Update the VEID identity score for an account.

This command is restricted to validators only. The sender must be a registered
validator with authority to update identity scores.

The score must be in the range 0-1000, where:
  - 0-299: Low trust
  - 300-599: Medium trust
  - 600-799: High trust
  - 800-1000: Very high trust

Example:
  virtengine tx veid update-score virtengine1abc... 750 --score-version v2.1.0
`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			accountAddress := args[0]
			newScoreStr := args[1]

			newScore, err := strconv.ParseUint(newScoreStr, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid score: %w", err)
			}
			if newScore > 1000 {
				return fmt.Errorf("score must be between 0 and 1000, got %d", newScore)
			}

			scoreVersion, _ := cmd.Flags().GetString("score-version")
			if scoreVersion == "" {
				scoreVersion = "v1.0.0"
			}

			msg := veidtypes.NewMsgUpdateScore(
				cctx.FromAddress.String(),
				accountAddress,
				uint32(newScore),
				scoreVersion,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("score-version", "v1.0.0", "Score model version")

	return cmd
}

// GetTxVEIDRegisterModelCmd returns the command to register a new ML model
func GetTxVEIDRegisterModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-model [model-type] [model-id] [sha256-hash] [version]",
		Short: "Register a new ML model on-chain",
		Long: `Register a new ML model for VEID identity verification.

The model type must be one of: face_verification, liveness, ocr, trust_score, gan_detection.
The SHA256 hash should match the output of scripts/compute_model_hash.sh.`,
		Args:              cobra.ExactArgs(4),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")

			msg := &veidv1.MsgRegisterModel{
				Authority: cctx.FromAddress.String(),
				ModelInfo: veidv1.MLModelInfo{
					ModelId:     args[1],
					ModelType:   args[0],
					Sha256Hash:  args[2],
					Version:     args[3],
					Name:        name,
					Description: description,
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("name", "", "Human-readable model name")
	cmd.Flags().String("description", "", "Model description")

	return cmd
}

// GetTxVEIDProposeModelUpdateCmd returns the command to propose an ML model update
func GetTxVEIDProposeModelUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-model-update [model-type] [new-model-id] [new-model-hash]",
		Short: "Propose an ML model update via governance",
		Long: `Propose updating an active ML model through governance voting.

The proposal must specify the model type, the new model's registered ID,
and the new model's SHA256 hash for verification.`,
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")
			activationDelay, _ := cmd.Flags().GetInt64("activation-delay")

			msg := &veidv1.MsgProposeModelUpdate{
				Proposer: cctx.FromAddress.String(),
				Proposal: veidv1.ModelUpdateProposal{
					Title:           title,
					Description:     description,
					ModelType:       args[0],
					NewModelId:      args[1],
					NewModelHash:    args[2],
					ActivationDelay: activationDelay,
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("title", "", "Proposal title (required)")
	cmd.Flags().String("description", "", "Proposal description")
	cmd.Flags().Int64("activation-delay", 100, "Blocks to wait after approval before activation")

	return cmd
}
