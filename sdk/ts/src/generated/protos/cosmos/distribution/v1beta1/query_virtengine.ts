import { QueryCommunityPoolRequest, QueryCommunityPoolResponse, QueryDelegationRewardsRequest, QueryDelegationRewardsResponse, QueryDelegationTotalRewardsRequest, QueryDelegationTotalRewardsResponse, QueryDelegatorValidatorsRequest, QueryDelegatorValidatorsResponse, QueryDelegatorWithdrawAddressRequest, QueryDelegatorWithdrawAddressResponse, QueryParamsRequest, QueryParamsResponse, QueryValidatorCommissionRequest, QueryValidatorCommissionResponse, QueryValidatorDistributionInfoRequest, QueryValidatorDistributionInfoResponse, QueryValidatorOutstandingRewardsRequest, QueryValidatorOutstandingRewardsResponse, QueryValidatorSlashesRequest, QueryValidatorSlashesResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.distribution.v1beta1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/cosmos/distribution/v1beta1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    validatorDistributionInfo: {
      name: "ValidatorDistributionInfo",
      httpPath: "/cosmos/distribution/v1beta1/validators/{validator_address}",
      input: QueryValidatorDistributionInfoRequest,
      output: QueryValidatorDistributionInfoResponse,
      get parent() { return Query; },
    },
    validatorOutstandingRewards: {
      name: "ValidatorOutstandingRewards",
      httpPath: "/cosmos/distribution/v1beta1/validators/{validator_address}/outstanding_rewards",
      input: QueryValidatorOutstandingRewardsRequest,
      output: QueryValidatorOutstandingRewardsResponse,
      get parent() { return Query; },
    },
    validatorCommission: {
      name: "ValidatorCommission",
      httpPath: "/cosmos/distribution/v1beta1/validators/{validator_address}/commission",
      input: QueryValidatorCommissionRequest,
      output: QueryValidatorCommissionResponse,
      get parent() { return Query; },
    },
    validatorSlashes: {
      name: "ValidatorSlashes",
      httpPath: "/cosmos/distribution/v1beta1/validators/{validator_address}/slashes",
      input: QueryValidatorSlashesRequest,
      output: QueryValidatorSlashesResponse,
      get parent() { return Query; },
    },
    delegationRewards: {
      name: "DelegationRewards",
      httpPath: "/cosmos/distribution/v1beta1/delegators/{delegator_address}/rewards/{validator_address}",
      input: QueryDelegationRewardsRequest,
      output: QueryDelegationRewardsResponse,
      get parent() { return Query; },
    },
    delegationTotalRewards: {
      name: "DelegationTotalRewards",
      httpPath: "/cosmos/distribution/v1beta1/delegators/{delegator_address}/rewards",
      input: QueryDelegationTotalRewardsRequest,
      output: QueryDelegationTotalRewardsResponse,
      get parent() { return Query; },
    },
    delegatorValidators: {
      name: "DelegatorValidators",
      httpPath: "/cosmos/distribution/v1beta1/delegators/{delegator_address}/validators",
      input: QueryDelegatorValidatorsRequest,
      output: QueryDelegatorValidatorsResponse,
      get parent() { return Query; },
    },
    delegatorWithdrawAddress: {
      name: "DelegatorWithdrawAddress",
      httpPath: "/cosmos/distribution/v1beta1/delegators/{delegator_address}/withdraw_address",
      input: QueryDelegatorWithdrawAddressRequest,
      output: QueryDelegatorWithdrawAddressResponse,
      get parent() { return Query; },
    },
    communityPool: {
      name: "CommunityPool",
      httpPath: "/cosmos/distribution/v1beta1/community_pool",
      input: QueryCommunityPoolRequest,
      output: QueryCommunityPoolResponse,
      get parent() { return Query; },
    },
  },
} as const;
