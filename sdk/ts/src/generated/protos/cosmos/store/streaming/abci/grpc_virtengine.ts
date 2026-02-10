import { ListenCommitRequest, ListenCommitResponse, ListenFinalizeBlockRequest, ListenFinalizeBlockResponse } from "./grpc.ts";

export const ABCIListenerService = {
  typeName: "cosmos.store.streaming.abci.ABCIListenerService",
  methods: {
    listenFinalizeBlock: {
      name: "ListenFinalizeBlock",
      input: ListenFinalizeBlockRequest,
      output: ListenFinalizeBlockResponse,
      get parent() { return ABCIListenerService; },
    },
    listenCommit: {
      name: "ListenCommit",
      input: ListenCommitRequest,
      output: ListenCommitResponse,
      get parent() { return ABCIListenerService; },
    },
  },
} as const;
