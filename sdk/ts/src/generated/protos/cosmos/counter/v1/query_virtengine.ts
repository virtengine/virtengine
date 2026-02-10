import { QueryGetCountRequest, QueryGetCountResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.counter.v1.Query",
  methods: {
    getCount: {
      name: "GetCount",
      input: QueryGetCountRequest,
      output: QueryGetCountResponse,
      get parent() { return Query; },
    },
  },
} as const;
