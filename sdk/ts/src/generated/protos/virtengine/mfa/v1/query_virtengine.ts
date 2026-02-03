import { QueryAllSensitiveTxConfigsRequest, QueryAllSensitiveTxConfigsResponse, QueryAuthorizationSessionRequest, QueryAuthorizationSessionResponse, QueryChallengeRequest, QueryChallengeResponse, QueryFactorEnrollmentRequest, QueryFactorEnrollmentResponse, QueryFactorEnrollmentsRequest, QueryFactorEnrollmentsResponse, QueryMFAPolicyRequest, QueryMFAPolicyResponse, QueryMFARequiredRequest, QueryMFARequiredResponse, QueryParamsRequest, QueryParamsResponse, QueryPendingChallengesRequest, QueryPendingChallengesResponse, QuerySensitiveTxConfigRequest, QuerySensitiveTxConfigResponse, QueryTrustedDevicesRequest, QueryTrustedDevicesResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.mfa.v1.Query",
  methods: {
    mFAPolicy: {
      name: "MFAPolicy",
      httpPath: "/virtengine/mfa/v1/policy/{address}",
      input: QueryMFAPolicyRequest,
      output: QueryMFAPolicyResponse,
      get parent() { return Query; },
    },
    factorEnrollments: {
      name: "FactorEnrollments",
      httpPath: "/virtengine/mfa/v1/enrollments/{address}",
      input: QueryFactorEnrollmentsRequest,
      output: QueryFactorEnrollmentsResponse,
      get parent() { return Query; },
    },
    factorEnrollment: {
      name: "FactorEnrollment",
      httpPath: "/virtengine/mfa/v1/enrollment/{address}/{factor_id}",
      input: QueryFactorEnrollmentRequest,
      output: QueryFactorEnrollmentResponse,
      get parent() { return Query; },
    },
    challenge: {
      name: "Challenge",
      httpPath: "/virtengine/mfa/v1/challenge/{challenge_id}",
      input: QueryChallengeRequest,
      output: QueryChallengeResponse,
      get parent() { return Query; },
    },
    pendingChallenges: {
      name: "PendingChallenges",
      httpPath: "/virtengine/mfa/v1/challenges/{address}",
      input: QueryPendingChallengesRequest,
      output: QueryPendingChallengesResponse,
      get parent() { return Query; },
    },
    authorizationSession: {
      name: "AuthorizationSession",
      httpPath: "/virtengine/mfa/v1/session/{session_id}",
      input: QueryAuthorizationSessionRequest,
      output: QueryAuthorizationSessionResponse,
      get parent() { return Query; },
    },
    trustedDevices: {
      name: "TrustedDevices",
      httpPath: "/virtengine/mfa/v1/devices/{address}",
      input: QueryTrustedDevicesRequest,
      output: QueryTrustedDevicesResponse,
      get parent() { return Query; },
    },
    sensitiveTxConfig: {
      name: "SensitiveTxConfig",
      httpPath: "/virtengine/mfa/v1/sensitive_tx/{transaction_type}",
      input: QuerySensitiveTxConfigRequest,
      output: QuerySensitiveTxConfigResponse,
      get parent() { return Query; },
    },
    allSensitiveTxConfigs: {
      name: "AllSensitiveTxConfigs",
      httpPath: "/virtengine/mfa/v1/sensitive_tx",
      input: QueryAllSensitiveTxConfigsRequest,
      output: QueryAllSensitiveTxConfigsResponse,
      get parent() { return Query; },
    },
    mFARequired: {
      name: "MFARequired",
      httpPath: "/virtengine/mfa/v1/required/{address}/{transaction_type}",
      input: QueryMFARequiredRequest,
      output: QueryMFARequiredResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/mfa/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
