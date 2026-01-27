import { ConfigRequest, ConfigResponse, StatusRequest, StatusResponse } from "./query.ts";

export const Service = {
  typeName: "cosmos.base.node.v1beta1.Service",
  methods: {
    config: {
      name: "Config",
      httpPath: "/cosmos/base/node/v1beta1/config",
      input: ConfigRequest,
      output: ConfigResponse,
      get parent() { return Service; },
    },
    status: {
      name: "Status",
      httpPath: "/cosmos/base/node/v1beta1/status",
      input: StatusRequest,
      output: StatusResponse,
      get parent() { return Service; },
    },
  },
} as const;
