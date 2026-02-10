import { QueryInterchainAccountRequest, QueryInterchainAccountResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.applications.interchain_accounts.controller.v1.Query",
  methods: {
    interchainAccount: {
      name: "InterchainAccount",
      httpPath: "/ibc/apps/interchain_accounts/controller/v1/owners/{owner}/connections/{connection_id}",
      input: QueryInterchainAccountRequest,
      output: QueryInterchainAccountResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/ibc/apps/interchain_accounts/controller/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
