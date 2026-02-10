import { QueryAllowanceRequest, QueryAllowanceResponse, QueryAllowancesByGranterRequest, QueryAllowancesByGranterResponse, QueryAllowancesRequest, QueryAllowancesResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.feegrant.v1beta1.Query",
  methods: {
    allowance: {
      name: "Allowance",
      httpPath: "/cosmos/feegrant/v1beta1/allowance/{granter}/{grantee}",
      input: QueryAllowanceRequest,
      output: QueryAllowanceResponse,
      get parent() { return Query; },
    },
    allowances: {
      name: "Allowances",
      httpPath: "/cosmos/feegrant/v1beta1/allowances/{grantee}",
      input: QueryAllowancesRequest,
      output: QueryAllowancesResponse,
      get parent() { return Query; },
    },
    allowancesByGranter: {
      name: "AllowancesByGranter",
      httpPath: "/cosmos/feegrant/v1beta1/issued/{granter}",
      input: QueryAllowancesByGranterRequest,
      output: QueryAllowancesByGranterResponse,
      get parent() { return Query; },
    },
  },
} as const;
