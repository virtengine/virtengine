import { QueryConfigRequest, QueryConfigResponse, QueryCounterpartyInfoRequest, QueryCounterpartyInfoResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.core.client.v2.Query",
  methods: {
    counterpartyInfo: {
      name: "CounterpartyInfo",
      httpPath: "/ibc/core/client/v2/counterparty_info/{client_id}",
      input: QueryCounterpartyInfoRequest,
      output: QueryCounterpartyInfoResponse,
      get parent() { return Query; },
    },
    config: {
      name: "Config",
      httpPath: "/ibc/core/client/v2/config/{client_id}",
      input: QueryConfigRequest,
      output: QueryConfigResponse,
      get parent() { return Query; },
    },
  },
} as const;
