import { Empty } from "../../../google/protobuf/empty.ts";
import { Status } from "./status.ts";

export const ProviderRPC = {
  typeName: "virtengine.provider.v1.ProviderRPC",
  methods: {
    getStatus: {
      name: "GetStatus",
      httpPath: "/v1/status",
      input: Empty,
      output: Status,
      get parent() { return ProviderRPC; },
    },
    streamStatus: {
      name: "StreamStatus",
      kind: "server_streaming",
      input: Empty,
      output: Status,
      get parent() { return ProviderRPC; },
    },
  },
} as const;
