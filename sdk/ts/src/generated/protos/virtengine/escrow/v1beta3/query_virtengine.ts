import { QueryAccountsRequest, QueryAccountsResponse, QueryPaymentsRequest, QueryPaymentsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.escrow.v1beta3.Query",
  methods: {
    accounts: {
      name: "Accounts",
      httpPath: "/virtengine/escrow/v1beta3/types/accounts/list",
      input: QueryAccountsRequest,
      output: QueryAccountsResponse,
      get parent() { return Query; },
    },
    payments: {
      name: "Payments",
      httpPath: "/virtengine/escrow/v1beta3/types/payments/list",
      input: QueryPaymentsRequest,
      output: QueryPaymentsResponse,
      get parent() { return Query; },
    },
  },
} as const;
