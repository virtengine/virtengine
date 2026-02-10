import { QueryAnnualProvisionsRequest, QueryAnnualProvisionsResponse, QueryInflationRequest, QueryInflationResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.mint.v1beta1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/cosmos/mint/v1beta1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    inflation: {
      name: "Inflation",
      httpPath: "/cosmos/mint/v1beta1/inflation",
      input: QueryInflationRequest,
      output: QueryInflationResponse,
      get parent() { return Query; },
    },
    annualProvisions: {
      name: "AnnualProvisions",
      httpPath: "/cosmos/mint/v1beta1/annual_provisions",
      input: QueryAnnualProvisionsRequest,
      output: QueryAnnualProvisionsResponse,
      get parent() { return Query; },
    },
  },
} as const;
