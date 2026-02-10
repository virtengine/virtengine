import { QueryChecksumsRequest, QueryChecksumsResponse, QueryCodeRequest, QueryCodeResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.lightclients.wasm.v1.Query",
  methods: {
    checksums: {
      name: "Checksums",
      httpPath: "/ibc/lightclients/wasm/v1/checksums",
      input: QueryChecksumsRequest,
      output: QueryChecksumsResponse,
      get parent() { return Query; },
    },
    code: {
      name: "Code",
      httpPath: "/ibc/lightclients/wasm/v1/checksums/{checksum}/code",
      input: QueryCodeRequest,
      output: QueryCodeResponse,
      get parent() { return Query; },
    },
  },
} as const;
