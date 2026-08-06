import { QueryExportJobRequest, QueryExportJobResponse, QueryExportJobsRequest, QueryExportJobsResponse, QueryLogEntriesRequest, QueryLogEntriesResponse, QueryLogEntryRequest, QueryLogEntryResponse, QueryParamsRequest, QueryParamsResponse } from "./query_audit_log.ts";

export const QueryAuditLog = {
  typeName: "virtengine.audit.v1.QueryAuditLog",
  methods: {
    queryLogEntries: {
      name: "QueryLogEntries",
      httpPath: "/virtengine/audit/v1/logs",
      input: QueryLogEntriesRequest,
      output: QueryLogEntriesResponse,
      get parent() { return QueryAuditLog; },
    },
    queryLogEntry: {
      name: "QueryLogEntry",
      httpPath: "/virtengine/audit/v1/logs/{id}",
      input: QueryLogEntryRequest,
      output: QueryLogEntryResponse,
      get parent() { return QueryAuditLog; },
    },
    queryExportJobs: {
      name: "QueryExportJobs",
      httpPath: "/virtengine/audit/v1/exports",
      input: QueryExportJobsRequest,
      output: QueryExportJobsResponse,
      get parent() { return QueryAuditLog; },
    },
    queryExportJob: {
      name: "QueryExportJob",
      httpPath: "/virtengine/audit/v1/exports/{id}",
      input: QueryExportJobRequest,
      output: QueryExportJobResponse,
      get parent() { return QueryAuditLog; },
    },
    queryParams: {
      name: "QueryParams",
      httpPath: "/virtengine/audit/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return QueryAuditLog; },
    },
  },
} as const;
