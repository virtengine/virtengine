import { QueryParamsRequest, QueryParamsResponse, QuerySigningInfoRequest, QuerySigningInfoResponse, QuerySigningInfosRequest, QuerySigningInfosResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.slashing.v1beta1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/cosmos/slashing/v1beta1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    signingInfo: {
      name: "SigningInfo",
      httpPath: "/cosmos/slashing/v1beta1/signing_infos/{cons_address}",
      input: QuerySigningInfoRequest,
      output: QuerySigningInfoResponse,
      get parent() { return Query; },
    },
    signingInfos: {
      name: "SigningInfos",
      httpPath: "/cosmos/slashing/v1beta1/signing_infos",
      input: QuerySigningInfosRequest,
      output: QuerySigningInfosResponse,
      get parent() { return Query; },
    },
  },
} as const;
