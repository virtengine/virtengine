import { MsgCancelProposal, MsgCancelProposalResponse, MsgDeposit, MsgDepositResponse, MsgExecLegacyContent, MsgExecLegacyContentResponse, MsgSubmitProposal, MsgSubmitProposalResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgVote, MsgVoteResponse, MsgVoteWeighted, MsgVoteWeightedResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.gov.v1.Msg",
  methods: {
    submitProposal: {
      name: "SubmitProposal",
      input: MsgSubmitProposal,
      output: MsgSubmitProposalResponse,
      get parent() { return Msg; },
    },
    execLegacyContent: {
      name: "ExecLegacyContent",
      input: MsgExecLegacyContent,
      output: MsgExecLegacyContentResponse,
      get parent() { return Msg; },
    },
    vote: {
      name: "Vote",
      input: MsgVote,
      output: MsgVoteResponse,
      get parent() { return Msg; },
    },
    voteWeighted: {
      name: "VoteWeighted",
      input: MsgVoteWeighted,
      output: MsgVoteWeightedResponse,
      get parent() { return Msg; },
    },
    deposit: {
      name: "Deposit",
      input: MsgDeposit,
      output: MsgDepositResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    cancelProposal: {
      name: "CancelProposal",
      input: MsgCancelProposal,
      output: MsgCancelProposalResponse,
      get parent() { return Msg; },
    },
  },
} as const;
