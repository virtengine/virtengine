import { QueryConfigRequest, QueryConfigResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.app.v1alpha1.Query",
  methods: {
    config: {
      name: "Config",
      input: QueryConfigRequest,
      output: QueryConfigResponse,
      get parent() { return Query; },
    },
  },
} as const;
