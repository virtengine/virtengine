import { MsgDeposit, MsgDepositResponse, MsgSubmitProposal, MsgSubmitProposalResponse, MsgVote, MsgVoteResponse, MsgVoteWeighted, MsgVoteWeightedResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.gov.v1beta1.Msg",
  methods: {
    submitProposal: {
      name: "SubmitProposal",
      input: MsgSubmitProposal,
      output: MsgSubmitProposalResponse,
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
  },
} as const;
