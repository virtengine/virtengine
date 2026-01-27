import { AccountResponse, AccountsResponse, DisabledListResponse, QueryAccountRequest, QueryAccountsRequest, QueryDisabledListRequest } from "./query.ts";

export const Query = {
  typeName: "cosmos.circuit.v1.Query",
  methods: {
    account: {
      name: "Account",
      httpPath: "/cosmos/circuit/v1/accounts/{address}",
      input: QueryAccountRequest,
      output: AccountResponse,
      get parent() { return Query; },
    },
    accounts: {
      name: "Accounts",
      httpPath: "/cosmos/circuit/v1/accounts",
      input: QueryAccountsRequest,
      output: AccountsResponse,
      get parent() { return Query; },
    },
    disabledList: {
      name: "DisabledList",
      httpPath: "/cosmos/circuit/v1/disable_list",
      input: QueryDisabledListRequest,
      output: DisabledListResponse,
      get parent() { return Query; },
    },
  },
} as const;
