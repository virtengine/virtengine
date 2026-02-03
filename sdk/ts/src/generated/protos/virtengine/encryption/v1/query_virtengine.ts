import { QueryAlgorithmsRequest, QueryAlgorithmsResponse, QueryKeyByFingerprintRequest, QueryKeyByFingerprintResponse, QueryParamsRequest, QueryParamsResponse, QueryRecipientKeyRequest, QueryRecipientKeyResponse, QueryValidateEnvelopeRequest, QueryValidateEnvelopeResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.encryption.v1.Query",
  methods: {
    recipientKey: {
      name: "RecipientKey",
      httpPath: "/virtengine/encryption/v1/key/{address}",
      input: QueryRecipientKeyRequest,
      output: QueryRecipientKeyResponse,
      get parent() { return Query; },
    },
    keyByFingerprint: {
      name: "KeyByFingerprint",
      httpPath: "/virtengine/encryption/v1/fingerprint/{fingerprint}",
      input: QueryKeyByFingerprintRequest,
      output: QueryKeyByFingerprintResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/encryption/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    algorithms: {
      name: "Algorithms",
      httpPath: "/virtengine/encryption/v1/algorithms",
      input: QueryAlgorithmsRequest,
      output: QueryAlgorithmsResponse,
      get parent() { return Query; },
    },
    validateEnvelope: {
      name: "ValidateEnvelope",
      httpMethod: "post",
      httpPath: "/virtengine/encryption/v1/validate",
      input: QueryValidateEnvelopeRequest,
      output: QueryValidateEnvelopeResponse,
      get parent() { return Query; },
    },
  },
} as const;
