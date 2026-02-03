import { QueryAccountRolesRequest, QueryAccountRolesResponse, QueryAccountStateRequest, QueryAccountStateResponse, QueryGenesisAccountsRequest, QueryGenesisAccountsResponse, QueryHasRoleRequest, QueryHasRoleResponse, QueryParamsRequest, QueryParamsResponse, QueryRoleMembersRequest, QueryRoleMembersResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.roles.v1.Query",
  methods: {
    accountRoles: {
      name: "AccountRoles",
      httpPath: "/virtengine/roles/v1/account/{address}/roles",
      input: QueryAccountRolesRequest,
      output: QueryAccountRolesResponse,
      get parent() { return Query; },
    },
    roleMembers: {
      name: "RoleMembers",
      httpPath: "/virtengine/roles/v1/role/{role}/members",
      input: QueryRoleMembersRequest,
      output: QueryRoleMembersResponse,
      get parent() { return Query; },
    },
    accountState: {
      name: "AccountState",
      httpPath: "/virtengine/roles/v1/account/{address}/state",
      input: QueryAccountStateRequest,
      output: QueryAccountStateResponse,
      get parent() { return Query; },
    },
    genesisAccounts: {
      name: "GenesisAccounts",
      httpPath: "/virtengine/roles/v1/genesis_accounts",
      input: QueryGenesisAccountsRequest,
      output: QueryGenesisAccountsResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/roles/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    hasRole: {
      name: "HasRole",
      httpPath: "/virtengine/roles/v1/account/{address}/has_role/{role}",
      input: QueryHasRoleRequest,
      output: QueryHasRoleResponse,
      get parent() { return Query; },
    },
  },
} as const;
