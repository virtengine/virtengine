import { AddressBytesToStringRequest, AddressBytesToStringResponse, AddressStringToBytesRequest, AddressStringToBytesResponse, Bech32PrefixRequest, Bech32PrefixResponse, QueryAccountAddressByIDRequest, QueryAccountAddressByIDResponse, QueryAccountInfoRequest, QueryAccountInfoResponse, QueryAccountRequest, QueryAccountResponse, QueryAccountsRequest, QueryAccountsResponse, QueryModuleAccountByNameRequest, QueryModuleAccountByNameResponse, QueryModuleAccountsRequest, QueryModuleAccountsResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.auth.v1beta1.Query",
  methods: {
    accounts: {
      name: "Accounts",
      httpPath: "/cosmos/auth/v1beta1/accounts",
      input: QueryAccountsRequest,
      output: QueryAccountsResponse,
      get parent() { return Query; },
    },
    account: {
      name: "Account",
      httpPath: "/cosmos/auth/v1beta1/accounts/{address}",
      input: QueryAccountRequest,
      output: QueryAccountResponse,
      get parent() { return Query; },
    },
    accountAddressByID: {
      name: "AccountAddressByID",
      httpPath: "/cosmos/auth/v1beta1/address_by_id/{id}",
      input: QueryAccountAddressByIDRequest,
      output: QueryAccountAddressByIDResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/cosmos/auth/v1beta1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    moduleAccounts: {
      name: "ModuleAccounts",
      httpPath: "/cosmos/auth/v1beta1/module_accounts",
      input: QueryModuleAccountsRequest,
      output: QueryModuleAccountsResponse,
      get parent() { return Query; },
    },
    moduleAccountByName: {
      name: "ModuleAccountByName",
      httpPath: "/cosmos/auth/v1beta1/module_accounts/{name}",
      input: QueryModuleAccountByNameRequest,
      output: QueryModuleAccountByNameResponse,
      get parent() { return Query; },
    },
    bech32Prefix: {
      name: "Bech32Prefix",
      httpPath: "/cosmos/auth/v1beta1/bech32",
      input: Bech32PrefixRequest,
      output: Bech32PrefixResponse,
      get parent() { return Query; },
    },
    addressBytesToString: {
      name: "AddressBytesToString",
      httpPath: "/cosmos/auth/v1beta1/bech32/{address_bytes}",
      input: AddressBytesToStringRequest,
      output: AddressBytesToStringResponse,
      get parent() { return Query; },
    },
    addressStringToBytes: {
      name: "AddressStringToBytes",
      httpPath: "/cosmos/auth/v1beta1/bech32/{address_string}",
      input: AddressStringToBytesRequest,
      output: AddressStringToBytesResponse,
      get parent() { return Query; },
    },
    accountInfo: {
      name: "AccountInfo",
      httpPath: "/cosmos/auth/v1beta1/account_info/{address}",
      input: QueryAccountInfoRequest,
      output: QueryAccountInfoResponse,
      get parent() { return Query; },
    },
  },
} as const;
