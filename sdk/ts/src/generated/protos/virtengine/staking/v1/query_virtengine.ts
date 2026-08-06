import { QueryCurrentEpochRequest, QueryCurrentEpochResponse, QueryParamsRequest, QueryParamsResponse, QueryRewardEpochRequest, QueryRewardEpochResponse, QuerySigningInfoRequest, QuerySigningInfoResponse, QuerySlashRecordsRequest, QuerySlashRecordsResponse, QueryValidatorPerformanceRequest, QueryValidatorPerformanceResponse, QueryValidatorPerformancesRequest, QueryValidatorPerformancesResponse, QueryValidatorRewardRequest, QueryValidatorRewardResponse, QueryValidatorRewardsRequest, QueryValidatorRewardsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.staking.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/virtengine/staking/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    validatorPerformance: {
      name: "ValidatorPerformance",
      httpPath: "/virtengine/staking/v1/validators/{validator_address}/performance/{epoch}",
      input: QueryValidatorPerformanceRequest,
      output: QueryValidatorPerformanceResponse,
      get parent() { return Query; },
    },
    validatorPerformances: {
      name: "ValidatorPerformances",
      httpPath: "/virtengine/staking/v1/validators/performance/{epoch}",
      input: QueryValidatorPerformancesRequest,
      output: QueryValidatorPerformancesResponse,
      get parent() { return Query; },
    },
    validatorReward: {
      name: "ValidatorReward",
      httpPath: "/virtengine/staking/v1/validators/{validator_address}/reward/{epoch}",
      input: QueryValidatorRewardRequest,
      output: QueryValidatorRewardResponse,
      get parent() { return Query; },
    },
    validatorRewards: {
      name: "ValidatorRewards",
      httpPath: "/virtengine/staking/v1/validators/{validator_address}/rewards",
      input: QueryValidatorRewardsRequest,
      output: QueryValidatorRewardsResponse,
      get parent() { return Query; },
    },
    rewardEpoch: {
      name: "RewardEpoch",
      httpPath: "/virtengine/staking/v1/reward-epoch/{epoch}",
      input: QueryRewardEpochRequest,
      output: QueryRewardEpochResponse,
      get parent() { return Query; },
    },
    slashRecords: {
      name: "SlashRecords",
      httpPath: "/virtengine/staking/v1/validators/{validator_address}/slashes",
      input: QuerySlashRecordsRequest,
      output: QuerySlashRecordsResponse,
      get parent() { return Query; },
    },
    signingInfo: {
      name: "SigningInfo",
      httpPath: "/virtengine/staking/v1/validators/{validator_address}/signing-info",
      input: QuerySigningInfoRequest,
      output: QuerySigningInfoResponse,
      get parent() { return Query; },
    },
    currentEpoch: {
      name: "CurrentEpoch",
      httpPath: "/virtengine/staking/v1/current-epoch",
      input: QueryCurrentEpochRequest,
      output: QueryCurrentEpochResponse,
      get parent() { return Query; },
    },
  },
} as const;
