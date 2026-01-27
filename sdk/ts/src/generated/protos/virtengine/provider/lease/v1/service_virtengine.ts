import { SendManifestRequest, SendManifestResponse, ServiceLogsRequest, ServiceLogsResponse, ServiceStatusRequest, ServiceStatusResponse } from "./service.ts";

export const LeaseRPC = {
  typeName: "virtengine.provider.lease.v1.LeaseRPC",
  methods: {
    sendManifest: {
      name: "SendManifest",
      input: SendManifestRequest,
      output: SendManifestResponse,
      get parent() { return LeaseRPC; },
    },
    serviceStatus: {
      name: "ServiceStatus",
      input: ServiceStatusRequest,
      output: ServiceStatusResponse,
      get parent() { return LeaseRPC; },
    },
    streamServiceStatus: {
      name: "StreamServiceStatus",
      kind: "server_streaming",
      input: ServiceStatusRequest,
      output: ServiceStatusResponse,
      get parent() { return LeaseRPC; },
    },
    serviceLogs: {
      name: "ServiceLogs",
      input: ServiceLogsRequest,
      output: ServiceLogsResponse,
      get parent() { return LeaseRPC; },
    },
    streamServiceLogs: {
      name: "StreamServiceLogs",
      kind: "server_streaming",
      input: ServiceLogsRequest,
      output: ServiceLogsResponse,
      get parent() { return LeaseRPC; },
    },
  },
} as const;
