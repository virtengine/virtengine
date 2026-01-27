import { QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.take.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/virtengine/take/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
