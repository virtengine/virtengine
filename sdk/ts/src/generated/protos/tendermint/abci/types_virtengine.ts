import { RequestApplySnapshotChunk, RequestCheckTx, RequestCommit, RequestEcho, RequestExtendVote, RequestFinalizeBlock, RequestFlush, RequestInfo, RequestInitChain, RequestListSnapshots, RequestLoadSnapshotChunk, RequestOfferSnapshot, RequestPrepareProposal, RequestProcessProposal, RequestQuery, RequestVerifyVoteExtension, ResponseApplySnapshotChunk, ResponseCheckTx, ResponseCommit, ResponseEcho, ResponseExtendVote, ResponseFinalizeBlock, ResponseFlush, ResponseInfo, ResponseInitChain, ResponseListSnapshots, ResponseLoadSnapshotChunk, ResponseOfferSnapshot, ResponsePrepareProposal, ResponseProcessProposal, ResponseQuery, ResponseVerifyVoteExtension } from "./types.ts";

export const ABCI = {
  typeName: "tendermint.abci.ABCI",
  methods: {
    echo: {
      name: "Echo",
      input: RequestEcho,
      output: ResponseEcho,
      get parent() { return ABCI; },
    },
    flush: {
      name: "Flush",
      input: RequestFlush,
      output: ResponseFlush,
      get parent() { return ABCI; },
    },
    info: {
      name: "Info",
      input: RequestInfo,
      output: ResponseInfo,
      get parent() { return ABCI; },
    },
    checkTx: {
      name: "CheckTx",
      input: RequestCheckTx,
      output: ResponseCheckTx,
      get parent() { return ABCI; },
    },
    query: {
      name: "Query",
      input: RequestQuery,
      output: ResponseQuery,
      get parent() { return ABCI; },
    },
    commit: {
      name: "Commit",
      input: RequestCommit,
      output: ResponseCommit,
      get parent() { return ABCI; },
    },
    initChain: {
      name: "InitChain",
      input: RequestInitChain,
      output: ResponseInitChain,
      get parent() { return ABCI; },
    },
    listSnapshots: {
      name: "ListSnapshots",
      input: RequestListSnapshots,
      output: ResponseListSnapshots,
      get parent() { return ABCI; },
    },
    offerSnapshot: {
      name: "OfferSnapshot",
      input: RequestOfferSnapshot,
      output: ResponseOfferSnapshot,
      get parent() { return ABCI; },
    },
    loadSnapshotChunk: {
      name: "LoadSnapshotChunk",
      input: RequestLoadSnapshotChunk,
      output: ResponseLoadSnapshotChunk,
      get parent() { return ABCI; },
    },
    applySnapshotChunk: {
      name: "ApplySnapshotChunk",
      input: RequestApplySnapshotChunk,
      output: ResponseApplySnapshotChunk,
      get parent() { return ABCI; },
    },
    prepareProposal: {
      name: "PrepareProposal",
      input: RequestPrepareProposal,
      output: ResponsePrepareProposal,
      get parent() { return ABCI; },
    },
    processProposal: {
      name: "ProcessProposal",
      input: RequestProcessProposal,
      output: ResponseProcessProposal,
      get parent() { return ABCI; },
    },
    extendVote: {
      name: "ExtendVote",
      input: RequestExtendVote,
      output: ResponseExtendVote,
      get parent() { return ABCI; },
    },
    verifyVoteExtension: {
      name: "VerifyVoteExtension",
      input: RequestVerifyVoteExtension,
      output: ResponseVerifyVoteExtension,
      get parent() { return ABCI; },
    },
    finalizeBlock: {
      name: "FinalizeBlock",
      input: RequestFinalizeBlock,
      output: ResponseFinalizeBlock,
      get parent() { return ABCI; },
    },
  },
} as const;
