import { QueryProviderRequest, QueryProviderResponse, QueryProviderSigningKeyEpochsRequest, QueryProviderSigningKeyEpochsResponse, QueryProviderSigningKeyRequest, QueryProviderSigningKeyResponse, QueryProvidersRequest, QueryProvidersResponse } from "./query.ts";

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
    providerSigningKey: {
      name: "ProviderSigningKey",
      httpPath: "/virtengine/provider/v1beta4/providers/{owner}/signing-keys/{epoch}",
      input: QueryProviderSigningKeyRequest,
      output: QueryProviderSigningKeyResponse,
      get parent() { return Query; },
    },
    providerSigningKeyEpochs: {
      name: "ProviderSigningKeyEpochs",
      httpPath: "/virtengine/provider/v1beta4/providers/{owner}/signing-keys",
      input: QueryProviderSigningKeyEpochsRequest,
      output: QueryProviderSigningKeyEpochsResponse,
      get parent() { return Query; },
    },
  },
} as const;
