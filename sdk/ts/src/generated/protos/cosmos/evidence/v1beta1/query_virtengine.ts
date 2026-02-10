import { QueryAllEvidenceRequest, QueryAllEvidenceResponse, QueryEvidenceRequest, QueryEvidenceResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.evidence.v1beta1.Query",
  methods: {
    evidence: {
      name: "Evidence",
      httpPath: "/cosmos/evidence/v1beta1/evidence/{hash}",
      input: QueryEvidenceRequest,
      output: QueryEvidenceResponse,
      get parent() { return Query; },
    },
    allEvidence: {
      name: "AllEvidence",
      httpPath: "/cosmos/evidence/v1beta1/evidence",
      input: QueryAllEvidenceRequest,
      output: QueryAllEvidenceResponse,
      get parent() { return Query; },
    },
  },
} as const;
