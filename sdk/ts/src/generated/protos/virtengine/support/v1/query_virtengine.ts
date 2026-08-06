import { QueryExternalRefRequest, QueryExternalRefResponse, QueryExternalRefsByOwnerRequest, QueryExternalRefsByOwnerResponse, QueryParamsRequest, QueryParamsResponse, QuerySupportRequestRequest, QuerySupportRequestResponse, QuerySupportRequestsBySubmitterRequest, QuerySupportRequestsBySubmitterResponse, QuerySupportResponsesByRequestRequest, QuerySupportResponsesByRequestResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.support.v1.Query",
  methods: {
    supportRequest: {
      name: "SupportRequest",
      httpPath: "/virtengine/support/v1/requests/{ticket_id=**}",
      input: QuerySupportRequestRequest,
      output: QuerySupportRequestResponse,
      get parent() { return Query; },
    },
    supportRequestsBySubmitter: {
      name: "SupportRequestsBySubmitter",
      httpPath: "/virtengine/support/v1/submitters/{submitter_address}/requests",
      input: QuerySupportRequestsBySubmitterRequest,
      output: QuerySupportRequestsBySubmitterResponse,
      get parent() { return Query; },
    },
    supportResponsesByRequest: {
      name: "SupportResponsesByRequest",
      httpPath: "/virtengine/support/v1/requests/{ticket_id=**}/responses",
      input: QuerySupportResponsesByRequestRequest,
      output: QuerySupportResponsesByRequestResponse,
      get parent() { return Query; },
    },
    externalRef: {
      name: "ExternalRef",
      httpPath: "/virtengine/support/v1/external_refs/{resource_type}/{resource_id=**}",
      input: QueryExternalRefRequest,
      output: QueryExternalRefResponse,
      get parent() { return Query; },
    },
    externalRefsByOwner: {
      name: "ExternalRefsByOwner",
      httpPath: "/virtengine/support/v1/owners/{owner_address}/external_refs",
      input: QueryExternalRefsByOwnerRequest,
      output: QueryExternalRefsByOwnerResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/support/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
