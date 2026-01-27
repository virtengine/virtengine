import { QueryAppliedPlanRequest, QueryAppliedPlanResponse, QueryAuthorityRequest, QueryAuthorityResponse, QueryCurrentPlanRequest, QueryCurrentPlanResponse, QueryModuleVersionsRequest, QueryModuleVersionsResponse, QueryUpgradedConsensusStateRequest, QueryUpgradedConsensusStateResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.upgrade.v1beta1.Query",
  methods: {
    currentPlan: {
      name: "CurrentPlan",
      httpPath: "/cosmos/upgrade/v1beta1/current_plan",
      input: QueryCurrentPlanRequest,
      output: QueryCurrentPlanResponse,
      get parent() { return Query; },
    },
    appliedPlan: {
      name: "AppliedPlan",
      httpPath: "/cosmos/upgrade/v1beta1/applied_plan/{name}",
      input: QueryAppliedPlanRequest,
      output: QueryAppliedPlanResponse,
      get parent() { return Query; },
    },
    upgradedConsensusState: {
      name: "UpgradedConsensusState",
      httpPath: "/cosmos/upgrade/v1beta1/upgraded_consensus_state/{last_height}",
      input: QueryUpgradedConsensusStateRequest,
      output: QueryUpgradedConsensusStateResponse,
      get parent() { return Query; },
    },
    moduleVersions: {
      name: "ModuleVersions",
      httpPath: "/cosmos/upgrade/v1beta1/module_versions",
      input: QueryModuleVersionsRequest,
      output: QueryModuleVersionsResponse,
      get parent() { return Query; },
    },
    authority: {
      name: "Authority",
      httpPath: "/cosmos/upgrade/v1beta1/authority",
      input: QueryAuthorityRequest,
      output: QueryAuthorityResponse,
      get parent() { return Query; },
    },
  },
} as const;
