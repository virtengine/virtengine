import { QueryAllocationsByCustomerRequest, QueryAllocationsByProviderRequest, QueryAllocationsResponse, QueryCatalogRequest, QueryCatalogResponse, QueryOfferingPriceRequest, QueryOfferingPriceResponse, QueryWaldurCommandsRequest, QueryWaldurCommandsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.marketplace.v1.Query",
  methods: {
    offeringPrice: {
      name: "OfferingPrice",
      httpPath: "/virtengine/marketplace/v1/offerings/{offering_id}/price",
      input: QueryOfferingPriceRequest,
      output: QueryOfferingPriceResponse,
      get parent() { return Query; },
    },
    allocationsByCustomer: {
      name: "AllocationsByCustomer",
      httpPath: "/virtengine/marketplace/v1/allocations/customer/{customer_address}",
      input: QueryAllocationsByCustomerRequest,
      output: QueryAllocationsResponse,
      get parent() { return Query; },
    },
    allocationsByProvider: {
      name: "AllocationsByProvider",
      httpPath: "/virtengine/marketplace/v1/allocations/provider/{provider_address}",
      input: QueryAllocationsByProviderRequest,
      output: QueryAllocationsResponse,
      get parent() { return Query; },
    },
    catalog: {
      name: "Catalog",
      httpPath: "/virtengine/marketplace/v1/catalog",
      input: QueryCatalogRequest,
      output: QueryCatalogResponse,
      get parent() { return Query; },
    },
    waldurCommands: {
      name: "WaldurCommands",
      httpPath: "/virtengine/marketplace/v1/waldur-commands",
      input: QueryWaldurCommandsRequest,
      output: QueryWaldurCommandsResponse,
      get parent() { return Query; },
    },
  },
} as const;
