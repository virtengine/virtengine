import { QueryCommunityPoolRequest, QueryCommunityPoolResponse, QueryContinuousFundRequest, QueryContinuousFundResponse, QueryContinuousFundsRequest, QueryContinuousFundsResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.protocolpool.v1.Query",
  methods: {
    communityPool: {
      name: "CommunityPool",
      httpPath: "/cosmos/protocolpool/v1/community_pool",
      input: QueryCommunityPoolRequest,
      output: QueryCommunityPoolResponse,
      get parent() { return Query; },
    },
    continuousFund: {
      name: "ContinuousFund",
      httpPath: "/cosmos/protocolpool/v1/continuous_funds/{recipient}",
      input: QueryContinuousFundRequest,
      output: QueryContinuousFundResponse,
      get parent() { return Query; },
    },
    continuousFunds: {
      name: "ContinuousFunds",
      httpPath: "/cosmos/protocolpool/v1/continuous_funds",
      input: QueryContinuousFundsRequest,
      output: QueryContinuousFundsResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/cosmos/protocolpool/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
