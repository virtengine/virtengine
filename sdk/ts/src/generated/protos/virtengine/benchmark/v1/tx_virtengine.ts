import { MsgFlagProvider, MsgFlagProviderResponse, MsgRequestChallenge, MsgRequestChallengeResponse, MsgResolveAnomalyFlag, MsgResolveAnomalyFlagResponse, MsgRespondChallenge, MsgRespondChallengeResponse, MsgSubmitBenchmarks, MsgSubmitBenchmarksResponse, MsgUnflagProvider, MsgUnflagProviderResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.benchmark.v1.Msg",
  methods: {
    submitBenchmarks: {
      name: "SubmitBenchmarks",
      input: MsgSubmitBenchmarks,
      output: MsgSubmitBenchmarksResponse,
      get parent() { return Msg; },
    },
    requestChallenge: {
      name: "RequestChallenge",
      input: MsgRequestChallenge,
      output: MsgRequestChallengeResponse,
      get parent() { return Msg; },
    },
    respondChallenge: {
      name: "RespondChallenge",
      input: MsgRespondChallenge,
      output: MsgRespondChallengeResponse,
      get parent() { return Msg; },
    },
    flagProvider: {
      name: "FlagProvider",
      input: MsgFlagProvider,
      output: MsgFlagProviderResponse,
      get parent() { return Msg; },
    },
    unflagProvider: {
      name: "UnflagProvider",
      input: MsgUnflagProvider,
      output: MsgUnflagProviderResponse,
      get parent() { return Msg; },
    },
    resolveAnomalyFlag: {
      name: "ResolveAnomalyFlag",
      input: MsgResolveAnomalyFlag,
      output: MsgResolveAnomalyFlagResponse,
      get parent() { return Msg; },
    },
  },
} as const;
