import { QueryAllocationHistoryRequest, QueryAllocationHistoryResponse, QueryAllocationRequest, QueryAllocationResponse, QueryAllocationsByProviderRequest, QueryAllocationsByProviderResponse, QueryAvailableResourcesRequest, QueryAvailableResourcesResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.resources.v1.Query",
  methods: {
    availableResources: {
      name: "AvailableResources",
      httpPath: "/virtengine/resources/v1/availability",
      input: QueryAvailableResourcesRequest,
      output: QueryAvailableResourcesResponse,
      get parent() { return Query; },
    },
    allocation: {
      name: "Allocation",
      httpPath: "/virtengine/resources/v1/allocation/{allocation_id}",
      input: QueryAllocationRequest,
      output: QueryAllocationResponse,
      get parent() { return Query; },
    },
    allocationHistory: {
      name: "AllocationHistory",
      httpPath: "/virtengine/resources/v1/allocation/{allocation_id}/history",
      input: QueryAllocationHistoryRequest,
      output: QueryAllocationHistoryResponse,
      get parent() { return Query; },
    },
    allocationsByProvider: {
      name: "AllocationsByProvider",
      httpPath: "/virtengine/resources/v1/allocations/provider/{provider_address}",
      input: QueryAllocationsByProviderRequest,
      output: QueryAllocationsByProviderResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/resources/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
