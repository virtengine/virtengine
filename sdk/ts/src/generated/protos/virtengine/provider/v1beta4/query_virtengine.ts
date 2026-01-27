import { QueryProviderRequest, QueryProviderResponse, QueryProvidersRequest, QueryProvidersResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.provider.v1beta4.Query",
  methods: {
    providers: {
      name: "Providers",
      httpPath: "/virtengine/provider/v1beta4/providers",
      input: QueryProvidersRequest,
      output: QueryProvidersResponse,
      get parent() { return Query; },
    },
    provider: {
      name: "Provider",
      httpPath: "/virtengine/provider/v1beta4/providers/{owner}",
      input: QueryProviderRequest,
      output: QueryProviderResponse,
      get parent() { return Query; },
    },
  },
} as const;
