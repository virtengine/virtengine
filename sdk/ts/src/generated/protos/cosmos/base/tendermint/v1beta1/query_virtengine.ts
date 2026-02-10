import { ABCIQueryRequest, ABCIQueryResponse, GetBlockByHeightRequest, GetBlockByHeightResponse, GetLatestBlockRequest, GetLatestBlockResponse, GetLatestValidatorSetRequest, GetLatestValidatorSetResponse, GetNodeInfoRequest, GetNodeInfoResponse, GetSyncingRequest, GetSyncingResponse, GetValidatorSetByHeightRequest, GetValidatorSetByHeightResponse } from "./query.ts";

export const Service = {
  typeName: "cosmos.base.tendermint.v1beta1.Service",
  methods: {
    getNodeInfo: {
      name: "GetNodeInfo",
      httpPath: "/cosmos/base/tendermint/v1beta1/node_info",
      input: GetNodeInfoRequest,
      output: GetNodeInfoResponse,
      get parent() { return Service; },
    },
    getSyncing: {
      name: "GetSyncing",
      httpPath: "/cosmos/base/tendermint/v1beta1/syncing",
      input: GetSyncingRequest,
      output: GetSyncingResponse,
      get parent() { return Service; },
    },
    getLatestBlock: {
      name: "GetLatestBlock",
      httpPath: "/cosmos/base/tendermint/v1beta1/blocks/latest",
      input: GetLatestBlockRequest,
      output: GetLatestBlockResponse,
      get parent() { return Service; },
    },
    getBlockByHeight: {
      name: "GetBlockByHeight",
      httpPath: "/cosmos/base/tendermint/v1beta1/blocks/{height}",
      input: GetBlockByHeightRequest,
      output: GetBlockByHeightResponse,
      get parent() { return Service; },
    },
    getLatestValidatorSet: {
      name: "GetLatestValidatorSet",
      httpPath: "/cosmos/base/tendermint/v1beta1/validatorsets/latest",
      input: GetLatestValidatorSetRequest,
      output: GetLatestValidatorSetResponse,
      get parent() { return Service; },
    },
    getValidatorSetByHeight: {
      name: "GetValidatorSetByHeight",
      httpPath: "/cosmos/base/tendermint/v1beta1/validatorsets/{height}",
      input: GetValidatorSetByHeightRequest,
      output: GetValidatorSetByHeightResponse,
      get parent() { return Service; },
    },
    aBCIQuery: {
      name: "ABCIQuery",
      httpPath: "/cosmos/base/tendermint/v1beta1/abci_query",
      input: ABCIQueryRequest,
      output: ABCIQueryResponse,
      get parent() { return Service; },
    },
  },
} as const;
