import { QueryCertificatesRequest, QueryCertificatesResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.cert.v1.Query",
  methods: {
    certificates: {
      name: "Certificates",
      httpPath: "/virtengine/cert/v1/certificates/list",
      input: QueryCertificatesRequest,
      output: QueryCertificatesResponse,
      get parent() { return Query; },
    },
  },
} as const;
