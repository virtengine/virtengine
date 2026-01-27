package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/pkg/errors"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"

	wtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// Proposal flags
const (
	flagVoter     = "voter"
	flagDepositor = "depositor"
	flagStatus    = "status"
)

// ProposalFlags defines the core required fields of a legacy proposal. It is used to
// verify that these values are not provided in conjunction with a JSON proposal
// file.
var ProposalFlags = []string{
	cflags.FlagTitle,
	cflags.FlagDescription,  // nolint:staticcheck
	cflags.FlagProposalType, // nolint:staticcheck
	cflags.FlagDeposit,
}

const (
	proposalText          = "text"
	proposalOther         = "other"
	draftProposalFileName = "draft_proposal.json"
	draftMetadataFileName = "draft_metadata.json"
)

// DefaultGovAuthority is set to the gov module address.
// Extension point for chains to overwrite the default
var DefaultGovAuthority = sdk.AccAddress(address.Module("gov"))

var suggestedProposalTypes = []proposalType{
	{
		Name:    proposalText,
		MsgType: "", // no message for text proposal
	},
	{
		Name:    "community-pool-spend",
		MsgType: "/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",
	},
	{
		Name:    "software-upgrade",
		MsgType: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
	},
	{
		Name:    "cancel-software-upgrade",
		MsgType: "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
	},
	{
		Name:    proposalOther,
		MsgType: "", // user will input the message type
	},
}

type proposalType struct {
	Name    string
	MsgType string
	Msg     sdk.Msg
}

// ProposalMsg defines the new Msg-based proposal.
type ProposalMsg struct {
	// Msgs defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Messages  []json.RawMessage `json:"messages,omitempty"`
	Metadata  string            `json:"metadata"`
	Deposit   string            `json:"deposit"`
	Title     string            `json:"title"`
	Summary   string            `json:"summary"`
	Expedited bool              `json:"expedited"`
}

type legacyProposal struct {
	Title       string
	Description string
	Type        string
	Deposit     string
}

// GetTxGovCmd returns the transaction commands for this module
// governance ModuleClient is slightly different from other ModuleClients in that
// it contains a slice of legacy "proposal" child commands. These commands are respective
// to the proposal type handlers that are implemented in other modules but are mounted
// under the governance CLI (eg. parameter change proposals).
func GetTxGovCmd(legacyPropCmds []*cobra.Command) *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdSubmitLegacyProp := GetTxGovSubmitLegacyProposalCmd()
	for _, propCmd := range legacyPropCmds {
		cflags.AddTxFlagsToCmd(propCmd)
		cmdSubmitLegacyProp.AddCommand(propCmd)
	}

	govTxCmd.AddCommand(
		GetTxGovDepositCmd(),
		GetTxGovVoteCmd(),
		GetTxGovWeightedVoteCmd(),
		GetTxGovSubmitProposalCmd(),
		GetTxGovDraftProposalCmd(),
		GetTxGovCancelProposalCmd(),

		// Deprecated
		cmdSubmitLegacyProp,
	)

	return govTxCmd
}

// GetTxGovSubmitProposalCmd implements submitting a proposal transaction command.
func GetTxGovSubmitProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [path/to/proposal.json]",
		Short: "Submit a proposal along with some messages, metadata and deposit",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a proposal along with some messages, metadata and deposit.
They should be defined in a JSON file.

Example:
$ %s tx gov submit-proposal path/to/proposal.json

Where proposal.json contains:

{
  // array of proto-JSON-encoded sdk.Msgs
  "messages": [
    {
      "@type": "/cosmos.bank.v1beta1.MsgSend",
      "from_address": "cosmos1...",
      "to_address": "cosmos1...",
      "amount":[{"denom": "stake","amount": "10"}]
    }
  ],
  // metadata can be any of base64 encoded, raw text, stringified json, IPFS link to json
  // see below for example metadata
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "10stake",
  "title": "My proposal",
  "summary": "A short summary of my proposal",
  "expedited": false
}

metadata example:
{
	"title": "",
	"authors": [""],
	"summary": "",
	"details": "",
	"proposal_forum_url": "",
	"vote_option_context": "",
}
`,
				version.AppName,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			proposal, msgs, deposit, err := parseSubmitProposal(cctx.Codec, args[0])
			if err != nil {
				return err
			}

			msg, err := v1.NewMsgSubmitProposal(
				msgs,
				deposit,
				cctx.GetFromAddress().String(),
				proposal.Metadata,
				proposal.Title,
				proposal.Summary,
				proposal.Expedited,
			)
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	cmd.AddCommand(
		GetTxGovWasmProposalStoreCodeCmd(),
		GetTxGovWasmProposalInstantiateContractCmd(),
		GetTxGovWasmProposalInstantiateContract2Cmd(),
		GetTxGovWasmProposalStoreAndInstantiateContractCmd(),
		GetTxGovWasmProposalMigrateContractCmd(),
		GetTxGovWasmProposalExecuteContractCmd(),
		GetTxGovWasmProposalSudoContractCmd(),
		GetTxGovWasmProposalUpdateContractAdminCmd(),
		GetTxGovWasmProposalClearContractAdminCmd(),
		GetTxGovWasmProposalPinCodesCmd(),
		GetTxGovWasmProposalUnpinCodesCmd(),
		GetTxGovWasmProposalUpdateInstantiateConfigCmd(),
		GetTxGovWasmProposalAddCodeUploadParamsAddresses(),
		GetTxGovWasmProposalRemoveCodeUploadParamsAddresses(),
		GetTxGovWasmProposalStoreAndMigrateContractCmd(),
	)

	return cmd
}

// GetTxGovSubmitLegacyProposalCmd implements submitting a proposal transaction command.
// Deprecated: please use GetTxGovSubmitProposalCmd instead.
func GetTxGovSubmitLegacyProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-legacy-proposal",
		Short: "Submit a legacy proposal along with an initial deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a legacy proposal along with an initial deposit.
Proposal title, description, type and deposit can be given directly or through a proposal JSON file.

Example:
$ %s tx gov submit-legacy-proposal --proposal="path/to/proposal.json" --from mykey

Where proposal.json contains:

{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "type": "Text",
  "deposit": "10test"
}

Which is equivalent to:

$ %s tx gov submit-legacy-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10test" --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			proposal, err := parseSubmitLegacyProposal(cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to parse proposal: %w", err)
			}

			amount, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			content, ok := v1beta1.ContentFromProposalType(proposal.Title, proposal.Description, proposal.Type)
			if !ok {
				return fmt.Errorf("failed to create proposal content: unknown proposal type %s", proposal.Type)
			}

			msg, err := v1beta1.NewMsgSubmitProposal(content, amount, cctx.GetFromAddress())
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String(cflags.FlagTitle, "", "The proposal title")
	cmd.Flags().String(cflags.FlagDescription, "", "The proposal description") // nolint:staticcheck
	cmd.Flags().String(cflags.FlagProposalType, "", "The proposal Type")       // nolint:staticcheck
	cmd.Flags().String(cflags.FlagDeposit, "", "The proposal deposit")
	cmd.Flags().String(cflags.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)") // nolint:staticcheck

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxGovDepositCmd implements depositing tokens for an active proposal.
func GetTxGovDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit tokens for an active proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a deposit for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov deposit 1 10stake --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// Get depositor address
			from := cctx.GetFromAddress()

			// Get amount of coins
			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := v1.NewMsgDeposit(from, proposalID, amount)

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

// GetTxGovVoteCmd implements creating a new vote command.
func GetTxGovVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [option]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov vote 1 yes --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// Get voting address
			from := cctx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Find out which vote option user chose
			byteVoteOption, err := v1.VoteOptionFromString(govutils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			metadata, err := cmd.Flags().GetString(cflags.FlagMetadata)
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := v1.NewMsgVote(from, proposalID, byteVoteOption, metadata)

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String(cflags.FlagMetadata, "", "Specify metadata of the vote")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxGovWeightedVoteCmd implements creating a new weighted vote command.
func GetTxGovWeightedVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weighted-vote [proposal-id] [weighted-options]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov weighted-vote 1 yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05 --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// Get voter address
			from := cctx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Figure out which vote options user chose
			options, err := v1.WeightedVoteOptionsFromString(govutils.NormalizeWeightedVoteOptions(args[1]))
			if err != nil {
				return err
			}

			metadata, err := cmd.Flags().GetString(cflags.FlagMetadata)
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := v1.NewMsgVoteWeighted(from, proposalID, options, metadata)

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String(cflags.FlagMetadata, "", "Specify metadata of the weighted vote")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxGovDraftProposalCmd let a user generate a draft proposal.
func GetTxGovDraftProposalCmd() *cobra.Command {
	flagSkipMetadata := "skip-metadata"

	cmd := &cobra.Command{
		Use:          "draft-proposal",
		Short:        "Generate a draft proposal json file. The generated proposal json contains only one message (skeleton).",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// prompt proposal type
			proposalTypesPrompt := promptui.Select{
				Label: "Select proposal type",
				Items: getProposalSuggestions(),
			}

			_, selectedProposalType, err := proposalTypesPrompt.Run()
			if err != nil {
				return fmt.Errorf("failed to prompt proposal types: %w", err)
			}

			var proposal proposalType
			for _, p := range suggestedProposalTypes {
				if strings.EqualFold(p.Name, selectedProposalType) {
					proposal = p
					break
				}
			}

			// create any proposal type
			if proposal.Name == proposalOther {
				// prompt proposal type
				msgPrompt := promptui.Select{
					Label: "Select proposal message type:",
					Items: func() []string {
						msgs := clientCtx.InterfaceRegistry.ListImplementations(sdk.MsgInterfaceProtoName)
						sort.Strings(msgs)
						return msgs
					}(),
				}

				_, result, err := msgPrompt.Run()
				if err != nil {
					return fmt.Errorf("failed to prompt proposal types: %w", err)
				}

				proposal.MsgType = result
			}

			if proposal.MsgType != "" {
				proposal.Msg, err = sdk.GetMsgFromTypeURL(clientCtx.Codec, proposal.MsgType)
				if err != nil {
					// should never happen
					panic(err)
				}
			}

			skipMetadataPrompt, _ := cmd.Flags().GetBool(flagSkipMetadata)

			result, metadata, err := proposal.Prompt(clientCtx.Codec, skipMetadataPrompt)
			if err != nil {
				return err
			}

			if err := writeFile(draftProposalFileName, result); err != nil {
				return err
			}

			if !skipMetadataPrompt {
				if err := writeFile(draftMetadataFileName, metadata); err != nil {
					return err
				}
			}

			cmd.Println("The draft proposal has successfully been generated.\nProposals should contain off-chain metadata, please upload the metadata JSON to IPFS.\nThen, replace the generated metadata field with the IPFS CID.")

			return nil
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(flagSkipMetadata, false, "skip metadata prompt")

	return cmd
}

// GetTxGovCancelProposalCmd implements submitting a cancel proposal transaction command.
func GetTxGovCancelProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-proposal [proposal-id]",
		Short:   "Cancel governance proposal before the voting period ends. Must be signed by the proposal creator.",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`$ %s tx gov cancel-proposal 1 --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// Get proposer address
			from := cctx.GetFromAddress()
			msg := v1.NewMsgCancelProposal(proposalID, from.String())

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

func GetTxGovWasmProposalStoreCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm-store [wasm file] --title [text] --summary [text] --authority [address]",
		Short: "Submit a wasm binary proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			storeCodeMsg, err := ParseWasmStoreCodeArgs(args[0], authority, cmd.Flags())
			if err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&storeCodeMsg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	addInstantiatePermissionFlags(cmd)

	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalInstantiateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instantiate-contract [code_id_int64] [json_encoded_init_args] --authority [address] --label [text] --title [text] --summary [text] --admin [address,optional] --amount [coins,optional]",
		Short: "Submit an instantiate wasm contract proposal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			instantiateMsg, err := ParseWasmInstantiateArgs(args[0], args[1], clientCtx.Keyring, authority, cmd.Flags())
			if err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{instantiateMsg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(cflags.FlagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(cflags.FlagAdmin, "", "Address or key name of an admin")
	cmd.Flags().Bool(cflags.FlagNoAdmin, false, "You must set this explicitly if you don't want an admin")

	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalInstantiateContract2Cmd() *cobra.Command {
	decoder := newArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use: "instantiate-contract-2 [code_id_int64] [json_encoded_init_args] [salt] --authority [address] --label [text] --title [text] " +
			"--summary [text] --admin [address,optional] --amount [coins,optional] --fix-msg [bool,optional]",
		Short: "Submit an instantiate wasm contract proposal with predictable address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}
			salt, err := decoder.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("salt: %w", err)
			}
			fixMsg, err := cmd.Flags().GetBool(cflags.FlagFixMsg)
			if err != nil {
				return fmt.Errorf("fix msg: %w", err)
			}
			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			data, err := ParseWasmInstantiateArgs(args[0], args[1], clientCtx.Keyring, authority, cmd.Flags())
			if err != nil {
				return err
			}
			instantiateMsg := &wtypes.MsgInstantiateContract2{
				Sender: data.Sender,
				Admin:  data.Admin,
				CodeID: data.CodeID,
				Label:  data.Label,
				Msg:    data.Msg,
				Funds:  data.Funds,
				Salt:   salt,
				FixMsg: fixMsg,
			}

			if err = instantiateMsg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{instantiateMsg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(cflags.FlagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(cflags.FlagAdmin, "", "Address of an admin")
	cmd.Flags().Bool(cflags.FlagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	cmd.Flags().Bool(cflags.FlagFixMsg, false, "An optional flag to include the json_encoded_init_args for the predictable address generation mode")
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")

	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalStoreAndInstantiateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "store-instantiate [wasm file] [json_encoded_init_args] --authority [address] --label [text] --title [text] --summary [text]" +
			"--unpin-code [unpin_code,optional] --source [source,optional] --builder [builder,optional] --code-hash [code_hash,optional] --admin [address,optional] --amount [coins,optional]",
		Short: "Submit a store and instantiate wasm contract proposal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			// Variable storeCodeMsg is not really used. But this allows us to reuse parseStoreCodeArgs.
			storeCodeMsg, err := ParseWasmStoreCodeArgs(args[0], authority, cmd.Flags())
			if err != nil {
				return err
			}

			unpinCode, err := cmd.Flags().GetBool(cflags.FlagUnpinCode)
			if err != nil {
				return err
			}

			source, builder, codeHash, err := ParseWasmVerificationFlags(storeCodeMsg.WASMByteCode, cmd.Flags())
			if err != nil {
				return err
			}

			amountStr, err := cmd.Flags().GetString(cflags.FlagAmount)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}
			amount, err := sdk.ParseCoinsNormalized(amountStr)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}
			label, err := cmd.Flags().GetString(cflags.FlagLabel)
			if err != nil {
				return fmt.Errorf("label: %s", err)
			}
			if label == "" {
				return errors.New("label is required on all contracts")
			}
			adminStr, err := cmd.Flags().GetString(cflags.FlagAdmin)
			if err != nil {
				return fmt.Errorf("admin: %s", err)
			}
			noAdmin, err := cmd.Flags().GetBool(cflags.FlagNoAdmin)
			if err != nil {
				return fmt.Errorf("no-admin: %s", err)
			}

			// ensure sensible admin is set (or explicitly immutable)
			if adminStr == "" && !noAdmin {
				return errors.New("you must set an admin or explicitly pass --no-admin to make it immutable (wasmd issue #719)")
			}
			if adminStr != "" && noAdmin {
				return errors.New("you set an admin and passed --no-admin, those cannot both be true")
			}

			if adminStr != "" {
				addr, err := sdk.AccAddressFromBech32(adminStr)
				if err != nil {
					info, err := clientCtx.Keyring.Key(adminStr)
					if err != nil {
						return fmt.Errorf("admin %s", err)
					}
					admin, err := info.GetAddress()
					if err != nil {
						return err
					}
					adminStr = admin.String()
				} else {
					adminStr = addr.String()
				}
			}

			storeAndInstantiateMsg := wtypes.MsgStoreAndInstantiateContract{
				Authority:             authority,
				WASMByteCode:          storeCodeMsg.WASMByteCode,
				InstantiatePermission: storeCodeMsg.InstantiatePermission,
				UnpinCode:             unpinCode,
				Source:                source,
				Builder:               builder,
				CodeHash:              codeHash,
				Admin:                 adminStr,
				Label:                 label,
				Msg:                   []byte(args[1]),
				Funds:                 amount,
			}
			if err = storeAndInstantiateMsg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&storeAndInstantiateMsg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().Bool(cflags.FlagUnpinCode, false, "Unpin code on upload, optional")
	cmd.Flags().String(cflags.FlagSource, "", "Code Source URL is a valid absolute HTTPS URI to the contract's source code,")
	cmd.Flags().String(cflags.FlagBuilder, "", "Builder is a valid docker image name with tag, such as \"cosmwasm/workspace-optimizer:0.12.9\"")
	cmd.Flags().BytesHex(cflags.FlagCodeHash, nil, "CodeHash is the sha256 hash of the wasm code")
	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(cflags.FlagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(cflags.FlagAdmin, "", "Address or key name of an admin")
	cmd.Flags().Bool(cflags.FlagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	addInstantiatePermissionFlags(cmd)
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalMigrateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-contract [contract_addr_bech32] [new_code_id_int64] [json_encoded_migration_args] --title [text] --summary [text] --authority [address]",
		Short: "Submit a migrate wasm contract to a new code version proposal",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			migrateMsg, err := parseMigrateContractArgs(args, authority)
			if err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&migrateMsg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalExecuteContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-contract [contract_addr_bech32] [json_encoded_execution_args] --title [text] --summary [text] --authority [address]",
		Short: "Submit a execute wasm contract proposal (run by any address)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			contract := args[0]
			execMsg := []byte(args[1])
			amountStr, err := cmd.Flags().GetString(cflags.FlagAmount)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}
			funds, err := sdk.ParseCoinsNormalized(amountStr)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}

			msg := wtypes.MsgExecuteContract{
				Sender:   authority,
				Contract: contract,
				Msg:      execMsg,
				Funds:    funds,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract during instantiation")

	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalSudoContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sudo-contract [contract_addr_bech32] [json_encoded_migration_args] --title [text] --summary [text] --authority [address]",
		Short: "Submit a sudo wasm contract proposal (to call privileged commands)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			msg := wtypes.MsgSudoContract{
				Authority: authority,
				Contract:  args[0],
				Msg:       []byte(args[1]),
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalUpdateContractAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-contract-admin [contract_addr_bech32] [new_admin_addr_bech32] --title [text] --summary [text] --authority [address]",
		Short: "Submit a new admin for a contract proposal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			upgradeAdminMsg, err := parseUpdateContractAdminArgs(args, authority)
			if err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&upgradeAdminMsg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalClearContractAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-contract-admin [contract_addr_bech32] --title [text] --summary [text] --authority [address]",
		Short: "Submit a clear admin for a contract to prevent further migrations proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			msg := wtypes.MsgClearAdmin{
				Sender:   authority,
				Contract: args[0],
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalPinCodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pin-codes [code-ids] --title [text] --summary [text] --authority [address]",
		Short: "Submit a pin code proposal for pinning a code to cache",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			codeIds, err := parsePinCodesArgs(args)
			if err != nil {
				return err
			}

			msg := wtypes.MsgPinCodes{
				Authority: authority,
				CodeIDs:   codeIds,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalUnpinCodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpin-codes [code-ids] --title [text] --summary [text] --authority [address]",
		Short: "Submit an unpin code proposal for unpinning a code to cache",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			codeIds, err := parsePinCodesArgs(args)
			if err != nil {
				return err
			}

			msg := wtypes.MsgUnpinCodes{
				Authority: authority,
				CodeIDs:   codeIds,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalUpdateInstantiateConfigCmd() *cobra.Command {
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	cmd := &cobra.Command{
		Use:   "update-instantiate-config [code-id:permission] --title [text] --summary [text] --authority [address]",
		Short: "Submit an update instantiate config proposal.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an update instantiate config proposal for multiple code ids.

Example:
$ %s tx gov submit-proposal update-instantiate-config 1:nobody 2:everybody 3:%s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm,%s1vx8knpllrj7n963p9ttd80w47kpacrhuts497x
`, version.AppName, bech32Prefix, bech32Prefix)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			updates, err := ParseWasmAccessConfigUpdates(args)
			if err != nil {
				return err
			}

			msgs := make([]sdk.Msg, len(updates))
			for i, update := range updates {
				permission := update.InstantiatePermission
				msg := &wtypes.MsgUpdateInstantiateConfig{
					Sender:                   authority,
					CodeID:                   update.CodeID,
					NewInstantiatePermission: &permission,
				}
				if err = msg.ValidateBasic(); err != nil {
					return err
				}
				msgs[i] = msg
			}

			proposalMsg, err := v1.NewMsgSubmitProposal(msgs, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalAddCodeUploadParamsAddresses() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-code-upload-params-addresses [addresses] --title [text] --summary [text] --authority [address]",
		Short: "Submit an add code upload params addresses proposal to add addresses to code upload config params",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			msg := wtypes.MsgAddCodeUploadParamsAddresses{
				Authority: authority,
				Addresses: args,
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalRemoveCodeUploadParamsAddresses() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-code-upload-params-addresses [addresses] --title [text] --summary [text] --authority [address]",
		Short: "Submit a remove code upload params addresses proposal to remove addresses from code upload config params",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			msg := wtypes.MsgRemoveCodeUploadParamsAddresses{
				Authority: authority,
				Addresses: args,
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func GetTxGovWasmProposalStoreAndMigrateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-migrate [wasm file] [contract_addr_bech32] [json_encoded_migration_args] --title [text] --summary [text] --authority [address]",
		Short: "Submit a store and migrate wasm contract proposal",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, expedite, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(cflags.FlagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			// Variable storeCodeMsg is not really used. But this allows us to reuse parseStoreCodeArgs.
			storeCodeMsg, err := ParseWasmStoreCodeArgs(args[0], authority, cmd.Flags())
			if err != nil {
				return err
			}

			msg := wtypes.MsgStoreAndMigrateContract{
				Authority:             authority,
				WASMByteCode:          storeCodeMsg.WASMByteCode,
				InstantiatePermission: storeCodeMsg.InstantiatePermission,
				Msg:                   []byte(args[2]),
				Contract:              args[1],
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, expedite)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}

	addInstantiatePermissionFlags(cmd)
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func addCommonProposalFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(cli.FlagTitle, "", "Title of proposal")
	cmd.Flags().String(cli.FlagSummary, "", "Summary of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "Deposit of proposal")
	cmd.Flags().String(cflags.FlagAuthority, DefaultGovAuthority.String(), "The address of the governance account. Default is the sdk gov module account")
	cmd.Flags().Bool(cflags.FlagExpedite, false, "Expedite proposals have shorter voting period but require higher voting threshold")
}

func getProposalInfo(cmd *cobra.Command) (client.Context, string, string, sdk.Coins, bool, error) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return client.Context{}, "", "", nil, false, err
	}

	proposalTitle, err := cmd.Flags().GetString(cflags.FlagTitle)
	if err != nil {
		return clientCtx, proposalTitle, "", nil, false, err
	}

	summary, err := cmd.Flags().GetString(cflags.FlagSummary)
	if err != nil {
		return client.Context{}, proposalTitle, summary, nil, false, err
	}

	depositArg, err := cmd.Flags().GetString(cflags.FlagDeposit)
	if err != nil {
		return client.Context{}, proposalTitle, summary, nil, false, err
	}

	deposit, err := sdk.ParseCoinsNormalized(depositArg)
	if err != nil {
		return client.Context{}, proposalTitle, summary, deposit, false, err
	}

	expedite, err := cmd.Flags().GetBool(cflags.FlagExpedite)
	if err != nil {
		return client.Context{}, proposalTitle, summary, deposit, false, err
	}

	return clientCtx, proposalTitle, summary, deposit, expedite, nil
}

func parsePinCodesArgs(args []string) ([]uint64, error) {
	codeIDs := make([]uint64, len(args))
	for i, c := range args {
		codeID, err := strconv.ParseUint(c, 10, 64)
		if err != nil {
			return codeIDs, fmt.Errorf("code IDs: %s", err)
		}
		codeIDs[i] = codeID
	}
	return codeIDs, nil
}

// getProposalSuggestions suggests a list of proposal types
func getProposalSuggestions() []string {
	types := make([]string, len(suggestedProposalTypes))
	for i, p := range suggestedProposalTypes {
		types[i] = p.Name
	}
	return types
}

// Prompt the proposal type values and return the proposal and its metadata
func (p *proposalType) Prompt(cdc codec.Codec, skipMetadata bool) (*ProposalMsg, types.ProposalMetadata, error) {
	metadata, err := PromptMetadata(skipMetadata)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal metadata: %w", err)
	}

	proposal := &ProposalMsg{
		Metadata: "ipfs://CID", // the metadata must be saved on IPFS, set placeholder
		Title:    metadata.Title,
		Summary:  metadata.Summary,
	}

	// set deposit
	depositPrompt := promptui.Prompt{
		Label:    "Enter proposal deposit",
		Validate: client.ValidatePromptCoins,
	}
	proposal.Deposit, err = depositPrompt.Run()
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal deposit: %w", err)
	}

	if p.Msg == nil {
		return proposal, metadata, nil
	}

	// set messages field
	result, err := Prompt(p.Msg, "msg")
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal message: %w", err)
	}

	message, err := cdc.MarshalInterfaceJSON(result)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to marshal proposal message: %w", err)
	}
	proposal.Messages = append(proposal.Messages, message)

	return proposal, metadata, nil
}

// Prompt prompts the user for all values of the given type.
// data is the struct to be filled
// namePrefix is the name to be displayed as "Enter <namePrefix> <field>"
func Prompt[T any](data T, namePrefix string) (T, error) {
	v := reflect.ValueOf(&data).Elem()
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		// if the field is a struct skip or not slice of string or int then skip
		switch v.Field(i).Kind() {
		case reflect.Struct:
			// TODO(@julienrbrt) in the future we can add a recursive call to Prompt
			continue
		case reflect.Slice:
			if v.Field(i).Type().Elem().Kind() != reflect.String && v.Field(i).Type().Elem().Kind() != reflect.Int {
				continue
			}
		}

		// create prompts
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter %s %s", namePrefix, strings.ToLower(client.CamelCaseToString(v.Type().Field(i).Name))),
			Validate: client.ValidatePromptNotEmpty,
		}

		fieldName := strings.ToLower(v.Type().Field(i).Name)

		if strings.EqualFold(fieldName, "authority") {
			// pre-fill with gov address
			prompt.Default = authtypes.NewModuleAddress(types.ModuleName).String()
			prompt.Validate = client.ValidatePromptAddress
		}

		// TODO(@julienrbrt) use scalar annotation instead of dumb string name matching
		if strings.Contains(fieldName, "addr") ||
			strings.Contains(fieldName, "sender") ||
			strings.Contains(fieldName, "voter") ||
			strings.Contains(fieldName, "depositor") ||
			strings.Contains(fieldName, "granter") ||
			strings.Contains(fieldName, "grantee") ||
			strings.Contains(fieldName, "recipient") {
			prompt.Validate = client.ValidatePromptAddress
		}

		result, err := prompt.Run()
		if err != nil {
			return data, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(result)
		case reflect.Int:
			resultInt, err := strconv.ParseInt(result, 10, 0)
			if err != nil {
				return data, fmt.Errorf("invalid value for int: %w", err)
			}
			// If a value was successfully parsed the ranges of:
			//      [minInt,     maxInt]
			// are within the ranges of:
			//      [minInt64, maxInt64]
			// of which on 64-bit machines, which are most common,
			// int==int64
			v.Field(i).SetInt(resultInt)
		case reflect.Slice:
			switch v.Field(i).Type().Elem().Kind() {
			case reflect.String:
				v.Field(i).Set(reflect.ValueOf([]string{result}))
			case reflect.Int:
				resultInt, err := strconv.ParseInt(result, 10, 0)
				if err != nil {
					return data, fmt.Errorf("invalid value for int: %w", err)
				}

				v.Field(i).Set(reflect.ValueOf([]int{int(resultInt)}))
			}
		default:
			// skip any other types
			continue
		}
	}

	return data, nil
}

// PromptMetadata prompts for proposal metadata or only title and summary if skip is true
func PromptMetadata(skip bool) (types.ProposalMetadata, error) {
	if !skip {
		metadata, err := Prompt(types.ProposalMetadata{}, "proposal")
		if err != nil {
			return metadata, fmt.Errorf("failed to set proposal metadata: %w", err)
		}

		return metadata, nil
	}

	// prompt for title and summary
	titlePrompt := promptui.Prompt{
		Label:    "Enter proposal title",
		Validate: client.ValidatePromptNotEmpty,
	}

	title, err := titlePrompt.Run()
	if err != nil {
		return types.ProposalMetadata{}, fmt.Errorf("failed to set proposal title: %w", err)
	}

	summaryPrompt := promptui.Prompt{
		Label:    "Enter proposal summary",
		Validate: client.ValidatePromptNotEmpty,
	}

	summary, err := summaryPrompt.Run()
	if err != nil {
		return types.ProposalMetadata{}, fmt.Errorf("failed to set proposal summary: %w", err)
	}

	return types.ProposalMetadata{Title: title, Summary: summary}, nil
}

// writeFile writes the input to the file
func writeFile(fileName string, input any) error {
	raw, err := json.MarshalIndent(input, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	if err := os.WriteFile(fileName, raw, 0o600); err != nil {
		return err
	}

	return nil
}

// parseSubmitProposal reads and parses the proposal.
func parseSubmitProposal(cdc codec.Codec, path string) (ProposalMsg, []sdk.Msg, sdk.Coins, error) {
	var proposal ProposalMsg

	var fl *os.File
	var err error
	if path == "-" || (path == "" && !term.IsTerminal(0)) {
		fl = os.Stdin
	} else {
		fl, err = os.Open(path)
		if err != nil {
			return proposal, nil, nil, err
		}
		defer fl.Close()
	}
	contents, err := io.ReadAll(fl)
	if err != nil {
		return proposal, nil, nil, err
	}

	err = json.Unmarshal(contents, &proposal)
	if err != nil {
		return proposal, nil, nil, err
	}

	msgs := make([]sdk.Msg, len(proposal.Messages))
	for i, anyJSON := range proposal.Messages {
		var msg sdk.Msg
		err := cdc.UnmarshalInterfaceJSON(anyJSON, &msg)
		if err != nil {
			return proposal, nil, nil, err
		}

		msgs[i] = msg
	}

	deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
	if err != nil {
		return proposal, nil, nil, err
	}

	return proposal, msgs, deposit, nil
}

// parseSubmitLegacyProposal reads and parses the legacy proposal.
func parseSubmitLegacyProposal(fs *pflag.FlagSet) (*legacyProposal, error) {
	proposal := &legacyProposal{}
	proposalFile, _ := fs.GetString(cflags.FlagProposal) // nolint:staticcheck

	if proposalFile == "" {
		proposalType, _ := fs.GetString(cflags.FlagProposalType) // nolint:staticcheck
		proposal.Title, _ = fs.GetString(cflags.FlagTitle)
		proposal.Description, _ = fs.GetString(cflags.FlagDescription) // nolint:staticcheck
		proposal.Type = govutils.NormalizeProposalType(proposalType)
		proposal.Deposit, _ = fs.GetString(cflags.FlagDeposit)
		if err := proposal.validate(); err != nil {
			return nil, err
		}

		return proposal, nil
	}

	for _, flag := range ProposalFlags {
		if v, _ := fs.GetString(flag); v != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --proposal, which is a noop", flag)
		}
	}

	contents, err := os.ReadFile(proposalFile) //nolint: gosec
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, proposal)
	if err != nil {
		return nil, err
	}

	if err := proposal.validate(); err != nil {
		return nil, err
	}

	return proposal, nil
}

// validate the legacyProposal
func (p legacyProposal) validate() error {
	if p.Type == "" {
		return fmt.Errorf("proposal type is required")
	}

	if p.Title == "" {
		return fmt.Errorf("proposal title is required")
	}

	if p.Description == "" {
		return fmt.Errorf("proposal description is required")
	}
	return nil
}
