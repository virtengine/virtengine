import { QueryAllProvidersAttributesRequest, QueryAuditorAttributesRequest, QueryProviderAttributesRequest, QueryProviderAuditorRequest, QueryProvidersResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.audit.v1.Query",
  methods: {
    allProvidersAttributes: {
      name: "AllProvidersAttributes",
      httpPath: "/virtengine/audit/v1/audit/attributes/list",
      input: QueryAllProvidersAttributesRequest,
      output: QueryProvidersResponse,
      get parent() { return Query; },
    },
    providerAttributes: {
      name: "ProviderAttributes",
      httpPath: "/virtengine/audit/v1/audit/attributes/{owner}/list",
      input: QueryProviderAttributesRequest,
      output: QueryProvidersResponse,
      get parent() { return Query; },
    },
    providerAuditorAttributes: {
      name: "ProviderAuditorAttributes",
      httpPath: "/virtengine/audit/v1/audit/attributes/{auditor}/{owner}",
      input: QueryProviderAuditorRequest,
      output: QueryProvidersResponse,
      get parent() { return Query; },
    },
    auditorAttributes: {
      name: "AuditorAttributes",
      httpPath: "/virtengine/provider/v1/auditor/{auditor}/list",
      input: QueryAuditorAttributesRequest,
      output: QueryProvidersResponse,
      get parent() { return Query; },
    },
  },
} as const;
