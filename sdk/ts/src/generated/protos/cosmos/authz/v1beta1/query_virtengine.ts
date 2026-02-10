import { QueryGranteeGrantsRequest, QueryGranteeGrantsResponse, QueryGranterGrantsRequest, QueryGranterGrantsResponse, QueryGrantsRequest, QueryGrantsResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.authz.v1beta1.Query",
  methods: {
    grants: {
      name: "Grants",
      httpPath: "/cosmos/authz/v1beta1/grants",
      input: QueryGrantsRequest,
      output: QueryGrantsResponse,
      get parent() { return Query; },
    },
    granterGrants: {
      name: "GranterGrants",
      httpPath: "/cosmos/authz/v1beta1/grants/granter/{granter}",
      input: QueryGranterGrantsRequest,
      output: QueryGranterGrantsResponse,
      get parent() { return Query; },
    },
    granteeGrants: {
      name: "GranteeGrants",
      httpPath: "/cosmos/authz/v1beta1/grants/grantee/{grantee}",
      input: QueryGranteeGrantsRequest,
      output: QueryGranteeGrantsResponse,
      get parent() { return Query; },
    },
  },
} as const;
