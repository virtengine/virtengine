import { QueryDelegationRequest, QueryDelegationResponse, QueryDelegatorAllRewardsRequest, QueryDelegatorAllRewardsResponse, QueryDelegatorDelegationsRequest, QueryDelegatorDelegationsResponse, QueryDelegatorRedelegationsRequest, QueryDelegatorRedelegationsResponse, QueryDelegatorRewardsRequest, QueryDelegatorRewardsResponse, QueryDelegatorUnbondingDelegationsRequest, QueryDelegatorUnbondingDelegationsResponse, QueryParamsRequest, QueryParamsResponse, QueryRedelegationRequest, QueryRedelegationResponse, QueryUnbondingDelegationRequest, QueryUnbondingDelegationResponse, QueryValidatorDelegationsRequest, QueryValidatorDelegationsResponse, QueryValidatorSharesRequest, QueryValidatorSharesResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.delegation.v1.Query",
  methods: {
    params: {
      name: "Params",
      httpPath: "/virtengine/delegation/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
    delegation: {
      name: "Delegation",
      httpPath: "/virtengine/delegation/v1/delegator/{delegator_address}/validator/{validator_address}",
      input: QueryDelegationRequest,
      output: QueryDelegationResponse,
      get parent() { return Query; },
    },
    delegatorDelegations: {
      name: "DelegatorDelegations",
      httpPath: "/virtengine/delegation/v1/delegator/{delegator_address}/delegations",
      input: QueryDelegatorDelegationsRequest,
      output: QueryDelegatorDelegationsResponse,
      get parent() { return Query; },
    },
    validatorDelegations: {
      name: "ValidatorDelegations",
      httpPath: "/virtengine/delegation/v1/validator/{validator_address}/delegations",
      input: QueryValidatorDelegationsRequest,
      output: QueryValidatorDelegationsResponse,
      get parent() { return Query; },
    },
    unbondingDelegation: {
      name: "UnbondingDelegation",
      httpPath: "/virtengine/delegation/v1/unbonding/{unbonding_id}",
      input: QueryUnbondingDelegationRequest,
      output: QueryUnbondingDelegationResponse,
      get parent() { return Query; },
    },
    delegatorUnbondingDelegations: {
      name: "DelegatorUnbondingDelegations",
      httpPath: "/virtengine/delegation/v1/delegator/{delegator_address}/unbonding",
      input: QueryDelegatorUnbondingDelegationsRequest,
      output: QueryDelegatorUnbondingDelegationsResponse,
      get parent() { return Query; },
    },
    redelegation: {
      name: "Redelegation",
      httpPath: "/virtengine/delegation/v1/redelegation/{redelegation_id}",
      input: QueryRedelegationRequest,
      output: QueryRedelegationResponse,
      get parent() { return Query; },
    },
    delegatorRedelegations: {
      name: "DelegatorRedelegations",
      httpPath: "/virtengine/delegation/v1/delegator/{delegator_address}/redelegations",
      input: QueryDelegatorRedelegationsRequest,
      output: QueryDelegatorRedelegationsResponse,
      get parent() { return Query; },
    },
    delegatorRewards: {
      name: "DelegatorRewards",
      httpPath: "/virtengine/delegation/v1/delegator/{delegator_address}/rewards/{validator_address}",
      input: QueryDelegatorRewardsRequest,
      output: QueryDelegatorRewardsResponse,
      get parent() { return Query; },
    },
    delegatorAllRewards: {
      name: "DelegatorAllRewards",
      httpPath: "/virtengine/delegation/v1/delegator/{delegator_address}/rewards",
      input: QueryDelegatorAllRewardsRequest,
      output: QueryDelegatorAllRewardsResponse,
      get parent() { return Query; },
    },
    validatorShares: {
      name: "ValidatorShares",
      httpPath: "/virtengine/delegation/v1/validator/{validator_address}/shares",
      input: QueryValidatorSharesRequest,
      output: QueryValidatorSharesResponse,
      get parent() { return Query; },
    },
  },
} as const;
