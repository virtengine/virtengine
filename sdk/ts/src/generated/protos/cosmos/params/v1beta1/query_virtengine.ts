import { QueryParamsRequest, QueryParamsResponse, QuerySubspacesRequest, QuerySubspacesResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.params.v1beta1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/cosmos/params/v1beta1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    subspaces: {
      name: "Subspaces",
      httpPath: "/cosmos/params/v1beta1/subspaces",
      input: QuerySubspacesRequest,
      output: QuerySubspacesResponse,
      get parent() { return Query; },
    },
  },
} as const;
