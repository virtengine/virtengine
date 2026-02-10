import { QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.applications.interchain_accounts.host.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/ibc/apps/interchain_accounts/host/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
