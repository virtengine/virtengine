import { QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.consensus.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/cosmos/consensus/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
