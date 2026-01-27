import { QueryDenomHashRequest, QueryDenomHashResponse, QueryDenomRequest, QueryDenomResponse, QueryDenomsRequest, QueryDenomsResponse, QueryEscrowAddressRequest, QueryEscrowAddressResponse, QueryParamsRequest, QueryParamsResponse, QueryTotalEscrowForDenomRequest, QueryTotalEscrowForDenomResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.applications.transfer.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/ibc/apps/transfer/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    denoms: {
      name: "Denoms",
      httpPath: "/ibc/apps/transfer/v1/denoms",
      input: QueryDenomsRequest,
      output: QueryDenomsResponse,
      get parent() { return Query; },
    },
    denom: {
      name: "Denom",
      httpPath: "/ibc/apps/transfer/v1/denoms/{hash=**}",
      input: QueryDenomRequest,
      output: QueryDenomResponse,
      get parent() { return Query; },
    },
    denomHash: {
      name: "DenomHash",
      httpPath: "/ibc/apps/transfer/v1/denom_hashes/{trace=**}",
      input: QueryDenomHashRequest,
      output: QueryDenomHashResponse,
      get parent() { return Query; },
    },
    escrowAddress: {
      name: "EscrowAddress",
      httpPath: "/ibc/apps/transfer/v1/channels/{channel_id}/ports/{port_id}/escrow_address",
      input: QueryEscrowAddressRequest,
      output: QueryEscrowAddressResponse,
      get parent() { return Query; },
    },
    totalEscrowForDenom: {
      name: "TotalEscrowForDenom",
      httpPath: "/ibc/apps/transfer/v1/total_escrow/{denom=**}",
      input: QueryTotalEscrowForDenomRequest,
      output: QueryTotalEscrowForDenomResponse,
      get parent() { return Query; },
    },
  },
} as const;
