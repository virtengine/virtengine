import { QueryActiveValidatorEnclaveKeysRequest, QueryActiveValidatorEnclaveKeysResponse, QueryAttestedResultRequest, QueryAttestedResultResponse, QueryCommitteeEnclaveKeysRequest, QueryCommitteeEnclaveKeysResponse, QueryEnclaveIdentityRequest, QueryEnclaveIdentityResponse, QueryKeyRotationRequest, QueryKeyRotationResponse, QueryMeasurementAllowlistRequest, QueryMeasurementAllowlistResponse, QueryMeasurementRequest, QueryMeasurementResponse, QueryParamsRequest, QueryParamsResponse, QueryValidKeySetRequest, QueryValidKeySetResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.enclave.v1.Query",
  methods: {
    enclaveIdentity: {
      name: "EnclaveIdentity",
      httpPath: "/virtengine/enclave/v1/identity/{validator_address}",
      input: QueryEnclaveIdentityRequest,
      output: QueryEnclaveIdentityResponse,
      get parent() { return Query; },
    },
    activeValidatorEnclaveKeys: {
      name: "ActiveValidatorEnclaveKeys",
      httpPath: "/virtengine/enclave/v1/active_keys",
      input: QueryActiveValidatorEnclaveKeysRequest,
      output: QueryActiveValidatorEnclaveKeysResponse,
      get parent() { return Query; },
    },
    committeeEnclaveKeys: {
      name: "CommitteeEnclaveKeys",
      httpPath: "/virtengine/enclave/v1/committee_keys",
      input: QueryCommitteeEnclaveKeysRequest,
      output: QueryCommitteeEnclaveKeysResponse,
      get parent() { return Query; },
    },
    measurementAllowlist: {
      name: "MeasurementAllowlist",
      httpPath: "/virtengine/enclave/v1/measurements",
      input: QueryMeasurementAllowlistRequest,
      output: QueryMeasurementAllowlistResponse,
      get parent() { return Query; },
    },
    measurement: {
      name: "Measurement",
      httpPath: "/virtengine/enclave/v1/measurement/{measurement_hash}",
      input: QueryMeasurementRequest,
      output: QueryMeasurementResponse,
      get parent() { return Query; },
    },
    keyRotation: {
      name: "KeyRotation",
      httpPath: "/virtengine/enclave/v1/rotation/{validator_address}",
      input: QueryKeyRotationRequest,
      output: QueryKeyRotationResponse,
      get parent() { return Query; },
    },
    validKeySet: {
      name: "ValidKeySet",
      httpPath: "/virtengine/enclave/v1/valid_keys",
      input: QueryValidKeySetRequest,
      output: QueryValidKeySetResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/enclave/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    attestedResult: {
      name: "AttestedResult",
      httpPath: "/virtengine/enclave/v1/attested_result/{block_height}/{scope_id}",
      input: QueryAttestedResultRequest,
      output: QueryAttestedResultResponse,
      get parent() { return Query; },
    },
  },
} as const;
