import { QueryAuditLogRequest, QueryAuditLogResponse, QueryFraudReportRequest, QueryFraudReportResponse, QueryFraudReportsByReportedPartyRequest, QueryFraudReportsByReportedPartyResponse, QueryFraudReportsByReporterRequest, QueryFraudReportsByReporterResponse, QueryFraudReportsRequest, QueryFraudReportsResponse, QueryModeratorQueueRequest, QueryModeratorQueueResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.fraud.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/virtengine/fraud/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    fraudReport: {
      name: "FraudReport",
      httpPath: "/virtengine/fraud/v1/reports/{report_id}",
      input: QueryFraudReportRequest,
      output: QueryFraudReportResponse,
      get parent() { return Query; },
    },
    fraudReports: {
      name: "FraudReports",
      httpPath: "/virtengine/fraud/v1/reports",
      input: QueryFraudReportsRequest,
      output: QueryFraudReportsResponse,
      get parent() { return Query; },
    },
    fraudReportsByReporter: {
      name: "FraudReportsByReporter",
      httpPath: "/virtengine/fraud/v1/reports/reporter/{reporter}",
      input: QueryFraudReportsByReporterRequest,
      output: QueryFraudReportsByReporterResponse,
      get parent() { return Query; },
    },
    fraudReportsByReportedParty: {
      name: "FraudReportsByReportedParty",
      httpPath: "/virtengine/fraud/v1/reports/reported/{reported_party}",
      input: QueryFraudReportsByReportedPartyRequest,
      output: QueryFraudReportsByReportedPartyResponse,
      get parent() { return Query; },
    },
    auditLog: {
      name: "AuditLog",
      httpPath: "/virtengine/fraud/v1/reports/{report_id}/audit",
      input: QueryAuditLogRequest,
      output: QueryAuditLogResponse,
      get parent() { return Query; },
    },
    moderatorQueue: {
      name: "ModeratorQueue",
      httpPath: "/virtengine/fraud/v1/moderator/queue",
      input: QueryModeratorQueueRequest,
      output: QueryModeratorQueueResponse,
      get parent() { return Query; },
    },
  },
} as const;
