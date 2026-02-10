import { QueryCurrentEpochRequest, QueryCurrentEpochResponse, QueryEpochInfosRequest, QueryEpochInfosResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.epochs.v1beta1.Query",
  methods: {
    epochInfos: {
      name: "EpochInfos",
      httpPath: "/cosmos/epochs/v1beta1/epochs",
      input: QueryEpochInfosRequest,
      output: QueryEpochInfosResponse,
      get parent() { return Query; },
    },
    currentEpoch: {
      name: "CurrentEpoch",
      httpPath: "/cosmos/epochs/v1beta1/current_epoch",
      input: QueryCurrentEpochRequest,
      output: QueryCurrentEpochResponse,
      get parent() { return Query; },
    },
  },
} as const;
