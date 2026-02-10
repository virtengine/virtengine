import { QueryClientConnectionsRequest, QueryClientConnectionsResponse, QueryConnectionClientStateRequest, QueryConnectionClientStateResponse, QueryConnectionConsensusStateRequest, QueryConnectionConsensusStateResponse, QueryConnectionParamsRequest, QueryConnectionParamsResponse, QueryConnectionRequest, QueryConnectionResponse, QueryConnectionsRequest, QueryConnectionsResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.core.connection.v1.Query",
  methods: {
    connection: {
      name: "Connection",
      httpPath: "/ibc/core/connection/v1/connections/{connection_id}",
      input: QueryConnectionRequest,
      output: QueryConnectionResponse,
      get parent() { return Query; },
    },
    connections: {
      name: "Connections",
      httpPath: "/ibc/core/connection/v1/connections",
      input: QueryConnectionsRequest,
      output: QueryConnectionsResponse,
      get parent() { return Query; },
    },
    clientConnections: {
      name: "ClientConnections",
      httpPath: "/ibc/core/connection/v1/client_connections/{client_id}",
      input: QueryClientConnectionsRequest,
      output: QueryClientConnectionsResponse,
      get parent() { return Query; },
    },
    connectionClientState: {
      name: "ConnectionClientState",
      httpPath: "/ibc/core/connection/v1/connections/{connection_id}/client_state",
      input: QueryConnectionClientStateRequest,
      output: QueryConnectionClientStateResponse,
      get parent() { return Query; },
    },
    connectionConsensusState: {
      name: "ConnectionConsensusState",
      httpPath: "/ibc/core/connection/v1/connections/{connection_id}/consensus_state/revision/{revision_number}/height/{revision_height}",
      input: QueryConnectionConsensusStateRequest,
      output: QueryConnectionConsensusStateResponse,
      get parent() { return Query; },
    },
    connectionParams: {
      name: "ConnectionParams",
      httpPath: "/ibc/core/connection/v1/params",
      input: QueryConnectionParamsRequest,
      output: QueryConnectionParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;
