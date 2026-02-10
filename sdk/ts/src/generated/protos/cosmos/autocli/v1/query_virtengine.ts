import { AppOptionsRequest, AppOptionsResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.autocli.v1.Query",
  methods: {
    appOptions: {
      name: "AppOptions",
      input: AppOptionsRequest,
      output: AppOptionsResponse,
      get parent() { return Query; },
    },
  },
} as const;
