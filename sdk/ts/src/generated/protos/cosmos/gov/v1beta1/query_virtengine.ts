import { QueryDepositRequest, QueryDepositResponse, QueryDepositsRequest, QueryDepositsResponse, QueryParamsRequest, QueryParamsResponse, QueryProposalRequest, QueryProposalResponse, QueryProposalsRequest, QueryProposalsResponse, QueryTallyResultRequest, QueryTallyResultResponse, QueryVoteRequest, QueryVoteResponse, QueryVotesRequest, QueryVotesResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.gov.v1beta1.Query",
  methods: {
    proposal: {
      name: "Proposal",
      httpPath: "/cosmos/gov/v1beta1/proposals/{proposal_id}",
      input: QueryProposalRequest,
      output: QueryProposalResponse,
      get parent() { return Query; },
    },
    proposals: {
      name: "Proposals",
      httpPath: "/cosmos/gov/v1beta1/proposals",
      input: QueryProposalsRequest,
      output: QueryProposalsResponse,
      get parent() { return Query; },
    },
    vote: {
      name: "Vote",
      httpPath: "/cosmos/gov/v1beta1/proposals/{proposal_id}/votes/{voter}",
      input: QueryVoteRequest,
      output: QueryVoteResponse,
      get parent() { return Query; },
    },
    votes: {
      name: "Votes",
      httpPath: "/cosmos/gov/v1beta1/proposals/{proposal_id}/votes",
      input: QueryVotesRequest,
      output: QueryVotesResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/cosmos/gov/v1beta1/params/{params_type}",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    deposit: {
      name: "Deposit",
      httpPath: "/cosmos/gov/v1beta1/proposals/{proposal_id}/deposits/{depositor}",
      input: QueryDepositRequest,
      output: QueryDepositResponse,
      get parent() { return Query; },
    },
    deposits: {
      name: "Deposits",
      httpPath: "/cosmos/gov/v1beta1/proposals/{proposal_id}/deposits",
      input: QueryDepositsRequest,
      output: QueryDepositsResponse,
      get parent() { return Query; },
    },
    tallyResult: {
      name: "TallyResult",
      httpPath: "/cosmos/gov/v1beta1/proposals/{proposal_id}/tally",
      input: QueryTallyResultRequest,
      output: QueryTallyResultResponse,
      get parent() { return Query; },
    },
  },
} as const;
