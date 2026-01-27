import { MsgAuthorizeCircuitBreaker, MsgAuthorizeCircuitBreakerResponse, MsgResetCircuitBreaker, MsgResetCircuitBreakerResponse, MsgTripCircuitBreaker, MsgTripCircuitBreakerResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.circuit.v1.Msg",
  methods: {
    authorizeCircuitBreaker: {
      name: "AuthorizeCircuitBreaker",
      input: MsgAuthorizeCircuitBreaker,
      output: MsgAuthorizeCircuitBreakerResponse,
      get parent() { return Msg; },
    },
    tripCircuitBreaker: {
      name: "TripCircuitBreaker",
      input: MsgTripCircuitBreaker,
      output: MsgTripCircuitBreakerResponse,
      get parent() { return Msg; },
    },
    resetCircuitBreaker: {
      name: "ResetCircuitBreaker",
      input: MsgResetCircuitBreaker,
      output: MsgResetCircuitBreakerResponse,
      get parent() { return Msg; },
    },
  },
} as const;
