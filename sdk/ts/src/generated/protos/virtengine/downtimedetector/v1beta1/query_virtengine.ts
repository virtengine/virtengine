import { RecoveredSinceDowntimeOfLengthRequest, RecoveredSinceDowntimeOfLengthResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.downtimedetector.v1beta1.Query",
  methods: {
    recoveredSinceDowntimeOfLength: {
      name: "RecoveredSinceDowntimeOfLength",
      httpPath: "/virtengine/downtime-detector/v1beta1/RecoveredSinceDowntimeOfLength",
      input: RecoveredSinceDowntimeOfLengthRequest,
      output: RecoveredSinceDowntimeOfLengthResponse,
      get parent() { return Query; },
    },
  },
} as const;
