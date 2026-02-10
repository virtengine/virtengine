import { QueryAccountsRequest, QueryAccountsResponse, QueryPaymentsRequest, QueryPaymentsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.escrow.v1.Query",
  methods: {
    accounts: {
      name: "Accounts",
      httpPath: "/virtengine/escrow/v1/types/accounts",
      input: QueryAccountsRequest,
      output: QueryAccountsResponse,
      get parent() { return Query; },
    },
    payments: {
      name: "Payments",
      httpPath: "/virtengine/escrow/v1/types/payments",
      input: QueryPaymentsRequest,
      output: QueryPaymentsResponse,
      get parent() { return Query; },
    },
  },
} as const;
