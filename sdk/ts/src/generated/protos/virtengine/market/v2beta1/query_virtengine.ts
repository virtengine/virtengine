import { QueryBidRequest, QueryBidResponse, QueryBidsRequest, QueryBidsResponse, QueryLeaseRequest, QueryLeaseResponse, QueryLeasesRequest, QueryLeasesResponse, QueryOrderRequest, QueryOrderResponse, QueryOrdersRequest, QueryOrdersResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.market.v2beta1.Query",
  methods: {
    orders: {
      name: "Orders",
      httpPath: "/virtengine/market/v2beta1/orders/list",
      input: QueryOrdersRequest,
      output: QueryOrdersResponse,
      get parent() { return Query; },
    },
    order: {
      name: "Order",
      httpPath: "/virtengine/market/v2beta1/orders/info",
      input: QueryOrderRequest,
      output: QueryOrderResponse,
      get parent() { return Query; },
    },
    bids: {
      name: "Bids",
      httpPath: "/virtengine/market/v2beta1/bids/list",
      input: QueryBidsRequest,
      output: QueryBidsResponse,
      get parent() { return Query; },
    },
    bid: {
      name: "Bid",
      httpPath: "/virtengine/market/v2beta1/bids/info",
      input: QueryBidRequest,
      output: QueryBidResponse,
      get parent() { return Query; },
    },
    leases: {
      name: "Leases",
      httpPath: "/virtengine/market/v2beta1/leases/list",
      input: QueryLeasesRequest,
      output: QueryLeasesResponse,
      get parent() { return Query; },
    },
    lease: {
      name: "Lease",
      httpPath: "/virtengine/market/v2beta1/leases/info",
      input: QueryLeaseRequest,
      output: QueryLeaseResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/market/v2beta1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
