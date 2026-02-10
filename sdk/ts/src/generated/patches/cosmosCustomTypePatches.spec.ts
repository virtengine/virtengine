import { Coin, DecCoin, DecProto } from "../protos/cosmos/base/v1beta1/coin.ts";
import { DelegationDelegatorReward, DelegatorStartingInfo, FeePool, Params, ValidatorAccumulatedCommission, ValidatorCurrentRewards, ValidatorHistoricalRewards, ValidatorOutstandingRewards, ValidatorSlashEvent, ValidatorSlashEvents } from "../protos/cosmos/distribution/v1beta1/distribution.ts";
import { DelegatorStartingInfoRecord, GenesisState, ValidatorAccumulatedCommissionRecord, ValidatorCurrentRewardsRecord, ValidatorHistoricalRewardsRecord, ValidatorOutstandingRewardsRecord, ValidatorSlashEventRecord } from "../protos/cosmos/distribution/v1beta1/genesis.ts";
import { QueryCommunityPoolResponse, QueryDelegationRewardsResponse, QueryDelegationTotalRewardsResponse, QueryParamsResponse, QueryValidatorCommissionResponse, QueryValidatorDistributionInfoResponse, QueryValidatorOutstandingRewardsResponse, QueryValidatorSlashesResponse } from "../protos/cosmos/distribution/v1beta1/query.ts";
import { MsgUpdateParams } from "../protos/cosmos/distribution/v1beta1/tx.ts";
import { TallyParams, Vote, WeightedVoteOption } from "../protos/cosmos/gov/v1beta1/gov.ts";
import { GenesisState as GenesisState$1 } from "../protos/cosmos/gov/v1beta1/genesis.ts";
import { QueryParamsResponse as QueryParamsResponse$1, QueryVoteResponse, QueryVotesResponse } from "../protos/cosmos/gov/v1beta1/query.ts";
import { MsgVoteWeighted } from "../protos/cosmos/gov/v1beta1/tx.ts";
import { Minter, Params as Params$1 } from "../protos/cosmos/mint/v1beta1/mint.ts";
import { GenesisState as GenesisState$2 } from "../protos/cosmos/mint/v1beta1/genesis.ts";
import { QueryAnnualProvisionsResponse, QueryInflationResponse, QueryParamsResponse as QueryParamsResponse$2 } from "../protos/cosmos/mint/v1beta1/query.ts";
import { MsgUpdateParams as MsgUpdateParams$1 } from "../protos/cosmos/mint/v1beta1/tx.ts";
import { ContinuousFund } from "../protos/cosmos/protocolpool/v1/types.ts";
import { GenesisState as GenesisState$3 } from "../protos/cosmos/protocolpool/v1/genesis.ts";
import { Timestamp } from "../protos/google/protobuf/timestamp.ts";
import { QueryContinuousFundResponse, QueryContinuousFundsResponse } from "../protos/cosmos/protocolpool/v1/query.ts";
import { MsgCreateContinuousFund } from "../protos/cosmos/protocolpool/v1/tx.ts";
import { Params as Params$2 } from "../protos/cosmos/slashing/v1beta1/slashing.ts";
import { GenesisState as GenesisState$4 } from "../protos/cosmos/slashing/v1beta1/genesis.ts";
import { Duration } from "../protos/google/protobuf/duration.ts";
import { QueryParamsResponse as QueryParamsResponse$3 } from "../protos/cosmos/slashing/v1beta1/query.ts";
import { MsgUpdateParams as MsgUpdateParams$2 } from "../protos/cosmos/slashing/v1beta1/tx.ts";
import { Commission, CommissionRates, Delegation, DelegationResponse, Description, HistoricalInfo, Params as Params$3, Redelegation, RedelegationEntry, RedelegationEntryResponse, RedelegationResponse, Validator } from "../protos/cosmos/staking/v1beta1/staking.ts";
import { Any } from "../protos/google/protobuf/any.ts";
import { GenesisState as GenesisState$5 } from "../protos/cosmos/staking/v1beta1/genesis.ts";
import { QueryDelegationResponse, QueryDelegatorDelegationsResponse, QueryDelegatorValidatorResponse, QueryDelegatorValidatorsResponse, QueryHistoricalInfoResponse, QueryParamsResponse as QueryParamsResponse$4, QueryRedelegationsResponse, QueryValidatorDelegationsResponse, QueryValidatorResponse, QueryValidatorsResponse } from "../protos/cosmos/staking/v1beta1/query.ts";
import { Consensus } from "../protos/tendermint/version/types.ts";
import { BlockID, Header, PartSetHeader } from "../protos/tendermint/types/types.ts";
import { MsgCreateValidator, MsgEditValidator, MsgUpdateParams as MsgUpdateParams$3 } from "../protos/cosmos/staking/v1beta1/tx.ts";

import { expect, describe, it } from "@jest/globals";
import { patches } from "./cosmosCustomTypePatches.ts";
import { generateMessage, type MessageSchema } from "@test/helpers/generateMessage";
import type { TypePatches } from "../../sdk/client/applyPatches.ts";

const messageTypes: Record<string, MessageSchema> = {
  "cosmos.base.v1beta1.DecCoin": {
    type: DecCoin,
    fields: [{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.base.v1beta1.DecProto": {
    type: DecProto,
    fields: [{name: "dec",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.distribution.v1beta1.Params": {
    type: Params,
    fields: [{name: "communityTax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "baseProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "bonusProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.distribution.v1beta1.ValidatorHistoricalRewards": {
    type: ValidatorHistoricalRewards,
    fields: [{name: "cumulativeRewardRatio",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.ValidatorCurrentRewards": {
    type: ValidatorCurrentRewards,
    fields: [{name: "rewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.ValidatorAccumulatedCommission": {
    type: ValidatorAccumulatedCommission,
    fields: [{name: "commission",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.ValidatorOutstandingRewards": {
    type: ValidatorOutstandingRewards,
    fields: [{name: "rewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.ValidatorSlashEvent": {
    type: ValidatorSlashEvent,
    fields: [{name: "fraction",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.distribution.v1beta1.ValidatorSlashEvents": {
    type: ValidatorSlashEvents,
    fields: [{name: "validatorSlashEvents",kind: "list",message: {fields: [{name: "validatorPeriod",kind: "scalar",scalarType: 4,},{name: "fraction",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: ValidatorSlashEvent},},],
  },
  "cosmos.distribution.v1beta1.FeePool": {
    type: FeePool,
    fields: [{name: "communityPool",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.DelegatorStartingInfo": {
    type: DelegatorStartingInfo,
    fields: [{name: "stake",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.distribution.v1beta1.DelegationDelegatorReward": {
    type: DelegationDelegatorReward,
    fields: [{name: "reward",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.ValidatorOutstandingRewardsRecord": {
    type: ValidatorOutstandingRewardsRecord,
    fields: [{name: "outstandingRewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.ValidatorAccumulatedCommissionRecord": {
    type: ValidatorAccumulatedCommissionRecord,
    fields: [{name: "accumulated",kind: "message",message: {fields: [{name: "commission",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: ValidatorAccumulatedCommission},},],
  },
  "cosmos.distribution.v1beta1.ValidatorHistoricalRewardsRecord": {
    type: ValidatorHistoricalRewardsRecord,
    fields: [{name: "rewards",kind: "message",message: {fields: [{name: "cumulativeRewardRatio",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},{name: "referenceCount",kind: "scalar",scalarType: 13,},],type: ValidatorHistoricalRewards},},],
  },
  "cosmos.distribution.v1beta1.ValidatorCurrentRewardsRecord": {
    type: ValidatorCurrentRewardsRecord,
    fields: [{name: "rewards",kind: "message",message: {fields: [{name: "rewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},{name: "period",kind: "scalar",scalarType: 4,},],type: ValidatorCurrentRewards},},],
  },
  "cosmos.distribution.v1beta1.DelegatorStartingInfoRecord": {
    type: DelegatorStartingInfoRecord,
    fields: [{name: "startingInfo",kind: "message",message: {fields: [{name: "previousPeriod",kind: "scalar",scalarType: 4,},{name: "stake",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "height",kind: "scalar",scalarType: 4,},],type: DelegatorStartingInfo},},],
  },
  "cosmos.distribution.v1beta1.ValidatorSlashEventRecord": {
    type: ValidatorSlashEventRecord,
    fields: [{name: "validatorSlashEvent",kind: "message",message: {fields: [{name: "validatorPeriod",kind: "scalar",scalarType: 4,},{name: "fraction",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: ValidatorSlashEvent},},],
  },
  "cosmos.distribution.v1beta1.GenesisState": {
    type: GenesisState,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "communityTax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "baseProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "bonusProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "withdrawAddrEnabled",kind: "scalar",scalarType: 8,},],type: Params},},{name: "feePool",kind: "message",message: {fields: [{name: "communityPool",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: FeePool},},{name: "outstandingRewards",kind: "list",message: {fields: [{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "outstandingRewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: ValidatorOutstandingRewardsRecord},},{name: "validatorAccumulatedCommissions",kind: "list",message: {fields: [{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "accumulated",kind: "message",message: {fields: [{name: "commission",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: ValidatorAccumulatedCommission},},],type: ValidatorAccumulatedCommissionRecord},},{name: "validatorHistoricalRewards",kind: "list",message: {fields: [{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "period",kind: "scalar",scalarType: 4,},{name: "rewards",kind: "message",message: {fields: [{name: "cumulativeRewardRatio",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},{name: "referenceCount",kind: "scalar",scalarType: 13,},],type: ValidatorHistoricalRewards},},],type: ValidatorHistoricalRewardsRecord},},{name: "validatorCurrentRewards",kind: "list",message: {fields: [{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "rewards",kind: "message",message: {fields: [{name: "rewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},{name: "period",kind: "scalar",scalarType: 4,},],type: ValidatorCurrentRewards},},],type: ValidatorCurrentRewardsRecord},},{name: "delegatorStartingInfos",kind: "list",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "startingInfo",kind: "message",message: {fields: [{name: "previousPeriod",kind: "scalar",scalarType: 4,},{name: "stake",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "height",kind: "scalar",scalarType: 4,},],type: DelegatorStartingInfo},},],type: DelegatorStartingInfoRecord},},{name: "validatorSlashEvents",kind: "list",message: {fields: [{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "height",kind: "scalar",scalarType: 4,},{name: "period",kind: "scalar",scalarType: 4,},{name: "validatorSlashEvent",kind: "message",message: {fields: [{name: "validatorPeriod",kind: "scalar",scalarType: 4,},{name: "fraction",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: ValidatorSlashEvent},},],type: ValidatorSlashEventRecord},},],
  },
  "cosmos.distribution.v1beta1.QueryParamsResponse": {
    type: QueryParamsResponse,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "communityTax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "baseProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "bonusProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "withdrawAddrEnabled",kind: "scalar",scalarType: 8,},],type: Params},},],
  },
  "cosmos.distribution.v1beta1.QueryValidatorDistributionInfoResponse": {
    type: QueryValidatorDistributionInfoResponse,
    fields: [{name: "selfBondRewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},{name: "commission",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.QueryValidatorOutstandingRewardsResponse": {
    type: QueryValidatorOutstandingRewardsResponse,
    fields: [{name: "rewards",kind: "message",message: {fields: [{name: "rewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: ValidatorOutstandingRewards},},],
  },
  "cosmos.distribution.v1beta1.QueryValidatorCommissionResponse": {
    type: QueryValidatorCommissionResponse,
    fields: [{name: "commission",kind: "message",message: {fields: [{name: "commission",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: ValidatorAccumulatedCommission},},],
  },
  "cosmos.distribution.v1beta1.QueryValidatorSlashesResponse": {
    type: QueryValidatorSlashesResponse,
    fields: [{name: "slashes",kind: "list",message: {fields: [{name: "validatorPeriod",kind: "scalar",scalarType: 4,},{name: "fraction",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: ValidatorSlashEvent},},],
  },
  "cosmos.distribution.v1beta1.QueryDelegationRewardsResponse": {
    type: QueryDelegationRewardsResponse,
    fields: [{name: "rewards",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.QueryDelegationTotalRewardsResponse": {
    type: QueryDelegationTotalRewardsResponse,
    fields: [{name: "rewards",kind: "list",message: {fields: [{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "reward",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],type: DelegationDelegatorReward},},{name: "total",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.QueryCommunityPoolResponse": {
    type: QueryCommunityPoolResponse,
    fields: [{name: "pool",kind: "list",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: DecCoin},},],
  },
  "cosmos.distribution.v1beta1.MsgUpdateParams": {
    type: MsgUpdateParams,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "communityTax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "baseProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "bonusProposerReward",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "withdrawAddrEnabled",kind: "scalar",scalarType: 8,},],type: Params},},],
  },
  "cosmos.gov.v1beta1.WeightedVoteOption": {
    type: WeightedVoteOption,
    fields: [{name: "weight",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.gov.v1beta1.Vote": {
    type: Vote,
    fields: [{name: "options",kind: "list",message: {fields: [{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "weight",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: WeightedVoteOption},},],
  },
  "cosmos.gov.v1beta1.TallyParams": {
    type: TallyParams,
    fields: [{name: "quorum",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "threshold",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "vetoThreshold",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],
  },
  "cosmos.gov.v1beta1.GenesisState": {
    type: GenesisState$1,
    fields: [{name: "votes",kind: "list",message: {fields: [{name: "proposalId",kind: "scalar",scalarType: 4,},{name: "voter",kind: "scalar",scalarType: 9,},{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "options",kind: "list",message: {fields: [{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "weight",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: WeightedVoteOption},},],type: Vote},},{name: "tallyParams",kind: "message",message: {fields: [{name: "quorum",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "threshold",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "vetoThreshold",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],type: TallyParams},},],
  },
  "cosmos.gov.v1beta1.QueryVoteResponse": {
    type: QueryVoteResponse,
    fields: [{name: "vote",kind: "message",message: {fields: [{name: "proposalId",kind: "scalar",scalarType: 4,},{name: "voter",kind: "scalar",scalarType: 9,},{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "options",kind: "list",message: {fields: [{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "weight",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: WeightedVoteOption},},],type: Vote},},],
  },
  "cosmos.gov.v1beta1.QueryVotesResponse": {
    type: QueryVotesResponse,
    fields: [{name: "votes",kind: "list",message: {fields: [{name: "proposalId",kind: "scalar",scalarType: 4,},{name: "voter",kind: "scalar",scalarType: 9,},{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "options",kind: "list",message: {fields: [{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "weight",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: WeightedVoteOption},},],type: Vote},},],
  },
  "cosmos.gov.v1beta1.QueryParamsResponse": {
    type: QueryParamsResponse$1,
    fields: [{name: "tallyParams",kind: "message",message: {fields: [{name: "quorum",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "threshold",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "vetoThreshold",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],type: TallyParams},},],
  },
  "cosmos.gov.v1beta1.MsgVoteWeighted": {
    type: MsgVoteWeighted,
    fields: [{name: "options",kind: "list",message: {fields: [{name: "option",kind: "enum",enum: ["UNSPECIFIED","YES","ABSTAIN","NO","NO_WITH_VETO"],},{name: "weight",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: WeightedVoteOption},},],
  },
  "cosmos.mint.v1beta1.Minter": {
    type: Minter,
    fields: [{name: "inflation",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "annualProvisions",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.mint.v1beta1.Params": {
    type: Params$1,
    fields: [{name: "inflationRateChange",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMin",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "goalBonded",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.mint.v1beta1.GenesisState": {
    type: GenesisState$2,
    fields: [{name: "minter",kind: "message",message: {fields: [{name: "inflation",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "annualProvisions",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Minter},},{name: "params",kind: "message",message: {fields: [{name: "mintDenom",kind: "scalar",scalarType: 9,},{name: "inflationRateChange",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMin",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "goalBonded",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "blocksPerYear",kind: "scalar",scalarType: 4,},],type: Params$1},},],
  },
  "cosmos.mint.v1beta1.QueryParamsResponse": {
    type: QueryParamsResponse$2,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "mintDenom",kind: "scalar",scalarType: 9,},{name: "inflationRateChange",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMin",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "goalBonded",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "blocksPerYear",kind: "scalar",scalarType: 4,},],type: Params$1},},],
  },
  "cosmos.mint.v1beta1.QueryInflationResponse": {
    type: QueryInflationResponse,
    fields: [{name: "inflation",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],
  },
  "cosmos.mint.v1beta1.QueryAnnualProvisionsResponse": {
    type: QueryAnnualProvisionsResponse,
    fields: [{name: "annualProvisions",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],
  },
  "cosmos.mint.v1beta1.MsgUpdateParams": {
    type: MsgUpdateParams$1,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "mintDenom",kind: "scalar",scalarType: 9,},{name: "inflationRateChange",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMax",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "inflationMin",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "goalBonded",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "blocksPerYear",kind: "scalar",scalarType: 4,},],type: Params$1},},],
  },
  "cosmos.protocolpool.v1.ContinuousFund": {
    type: ContinuousFund,
    fields: [{name: "percentage",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.protocolpool.v1.GenesisState": {
    type: GenesisState$3,
    fields: [{name: "continuousFunds",kind: "list",message: {fields: [{name: "recipient",kind: "scalar",scalarType: 9,},{name: "percentage",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "expiry",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: ContinuousFund},},],
  },
  "cosmos.protocolpool.v1.QueryContinuousFundResponse": {
    type: QueryContinuousFundResponse,
    fields: [{name: "continuousFund",kind: "message",message: {fields: [{name: "recipient",kind: "scalar",scalarType: 9,},{name: "percentage",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "expiry",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: ContinuousFund},},],
  },
  "cosmos.protocolpool.v1.QueryContinuousFundsResponse": {
    type: QueryContinuousFundsResponse,
    fields: [{name: "continuousFunds",kind: "list",message: {fields: [{name: "recipient",kind: "scalar",scalarType: 9,},{name: "percentage",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "expiry",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: ContinuousFund},},],
  },
  "cosmos.protocolpool.v1.MsgCreateContinuousFund": {
    type: MsgCreateContinuousFund,
    fields: [{name: "percentage",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.slashing.v1beta1.Params": {
    type: Params$2,
    fields: [{name: "minSignedPerWindow",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "slashFractionDoubleSign",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "slashFractionDowntime",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],
  },
  "cosmos.slashing.v1beta1.GenesisState": {
    type: GenesisState$4,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "signedBlocksWindow",kind: "scalar",scalarType: 3,},{name: "minSignedPerWindow",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "downtimeJailDuration",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Duration},},{name: "slashFractionDoubleSign",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "slashFractionDowntime",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],type: Params$2},},],
  },
  "cosmos.slashing.v1beta1.QueryParamsResponse": {
    type: QueryParamsResponse$3,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "signedBlocksWindow",kind: "scalar",scalarType: 3,},{name: "minSignedPerWindow",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "downtimeJailDuration",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Duration},},{name: "slashFractionDoubleSign",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "slashFractionDowntime",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],type: Params$2},},],
  },
  "cosmos.slashing.v1beta1.MsgUpdateParams": {
    type: MsgUpdateParams$2,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "signedBlocksWindow",kind: "scalar",scalarType: 3,},{name: "minSignedPerWindow",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "downtimeJailDuration",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Duration},},{name: "slashFractionDoubleSign",kind: "scalar",scalarType: 12,customType: "LegacyDec",},{name: "slashFractionDowntime",kind: "scalar",scalarType: 12,customType: "LegacyDec",},],type: Params$2},},],
  },
  "cosmos.staking.v1beta1.HistoricalInfo": {
    type: HistoricalInfo,
    fields: [{name: "valset",kind: "list",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},],
  },
  "cosmos.staking.v1beta1.Validator": {
    type: Validator,
    fields: [{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},],
  },
  "cosmos.staking.v1beta1.Commission": {
    type: Commission,
    fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},],
  },
  "cosmos.staking.v1beta1.CommissionRates": {
    type: CommissionRates,
    fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.staking.v1beta1.Delegation": {
    type: Delegation,
    fields: [{name: "shares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.staking.v1beta1.RedelegationEntry": {
    type: RedelegationEntry,
    fields: [{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.staking.v1beta1.Redelegation": {
    type: Redelegation,
    fields: [{name: "entries",kind: "list",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},],
  },
  "cosmos.staking.v1beta1.Params": {
    type: Params$3,
    fields: [{name: "minCommissionRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.staking.v1beta1.DelegationResponse": {
    type: DelegationResponse,
    fields: [{name: "delegation",kind: "message",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "shares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Delegation},},],
  },
  "cosmos.staking.v1beta1.RedelegationEntryResponse": {
    type: RedelegationEntryResponse,
    fields: [{name: "redelegationEntry",kind: "message",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},],
  },
  "cosmos.staking.v1beta1.RedelegationResponse": {
    type: RedelegationResponse,
    fields: [{name: "redelegation",kind: "message",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorSrcAddress",kind: "scalar",scalarType: 9,},{name: "validatorDstAddress",kind: "scalar",scalarType: 9,},{name: "entries",kind: "list",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},],type: Redelegation},},{name: "entries",kind: "list",message: {fields: [{name: "redelegationEntry",kind: "message",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},{name: "balance",kind: "scalar",scalarType: 9,},],type: RedelegationEntryResponse},},],
  },
  "cosmos.staking.v1beta1.GenesisState": {
    type: GenesisState$5,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Duration},},{name: "maxValidators",kind: "scalar",scalarType: 13,},{name: "maxEntries",kind: "scalar",scalarType: 13,},{name: "historicalEntries",kind: "scalar",scalarType: 13,},{name: "bondDenom",kind: "scalar",scalarType: 9,},{name: "minCommissionRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Params$3},},{name: "validators",kind: "list",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},{name: "delegations",kind: "list",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "shares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Delegation},},{name: "redelegations",kind: "list",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorSrcAddress",kind: "scalar",scalarType: 9,},{name: "validatorDstAddress",kind: "scalar",scalarType: 9,},{name: "entries",kind: "list",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},],type: Redelegation},},],
  },
  "cosmos.staking.v1beta1.QueryValidatorsResponse": {
    type: QueryValidatorsResponse,
    fields: [{name: "validators",kind: "list",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},],
  },
  "cosmos.staking.v1beta1.QueryValidatorResponse": {
    type: QueryValidatorResponse,
    fields: [{name: "validator",kind: "message",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},],
  },
  "cosmos.staking.v1beta1.QueryValidatorDelegationsResponse": {
    type: QueryValidatorDelegationsResponse,
    fields: [{name: "delegationResponses",kind: "list",message: {fields: [{name: "delegation",kind: "message",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "shares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Delegation},},{name: "balance",kind: "message",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,},],type: Coin},},],type: DelegationResponse},},],
  },
  "cosmos.staking.v1beta1.QueryDelegationResponse": {
    type: QueryDelegationResponse,
    fields: [{name: "delegationResponse",kind: "message",message: {fields: [{name: "delegation",kind: "message",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "shares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Delegation},},{name: "balance",kind: "message",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,},],type: Coin},},],type: DelegationResponse},},],
  },
  "cosmos.staking.v1beta1.QueryDelegatorDelegationsResponse": {
    type: QueryDelegatorDelegationsResponse,
    fields: [{name: "delegationResponses",kind: "list",message: {fields: [{name: "delegation",kind: "message",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorAddress",kind: "scalar",scalarType: 9,},{name: "shares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Delegation},},{name: "balance",kind: "message",message: {fields: [{name: "denom",kind: "scalar",scalarType: 9,},{name: "amount",kind: "scalar",scalarType: 9,},],type: Coin},},],type: DelegationResponse},},],
  },
  "cosmos.staking.v1beta1.QueryRedelegationsResponse": {
    type: QueryRedelegationsResponse,
    fields: [{name: "redelegationResponses",kind: "list",message: {fields: [{name: "redelegation",kind: "message",message: {fields: [{name: "delegatorAddress",kind: "scalar",scalarType: 9,},{name: "validatorSrcAddress",kind: "scalar",scalarType: 9,},{name: "validatorDstAddress",kind: "scalar",scalarType: 9,},{name: "entries",kind: "list",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},],type: Redelegation},},{name: "entries",kind: "list",message: {fields: [{name: "redelegationEntry",kind: "message",message: {fields: [{name: "creationHeight",kind: "scalar",scalarType: 3,},{name: "completionTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "initialBalance",kind: "scalar",scalarType: 9,},{name: "sharesDst",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "unbondingId",kind: "scalar",scalarType: 4,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},],type: RedelegationEntry},},{name: "balance",kind: "scalar",scalarType: 9,},],type: RedelegationEntryResponse},},],type: RedelegationResponse},},],
  },
  "cosmos.staking.v1beta1.QueryDelegatorValidatorsResponse": {
    type: QueryDelegatorValidatorsResponse,
    fields: [{name: "validators",kind: "list",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},],
  },
  "cosmos.staking.v1beta1.QueryDelegatorValidatorResponse": {
    type: QueryDelegatorValidatorResponse,
    fields: [{name: "validator",kind: "message",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},],
  },
  "cosmos.staking.v1beta1.QueryHistoricalInfoResponse": {
    type: QueryHistoricalInfoResponse,
    fields: [{name: "hist",kind: "message",message: {fields: [{name: "header",kind: "message",message: {fields: [{name: "version",kind: "message",message: {fields: [{name: "block",kind: "scalar",scalarType: 4,},{name: "app",kind: "scalar",scalarType: 4,},],type: Consensus},},{name: "chainId",kind: "scalar",scalarType: 9,},{name: "height",kind: "scalar",scalarType: 3,},{name: "time",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "lastBlockId",kind: "message",message: {fields: [{name: "hash",kind: "scalar",scalarType: 12,},{name: "partSetHeader",kind: "message",message: {fields: [{name: "total",kind: "scalar",scalarType: 13,},{name: "hash",kind: "scalar",scalarType: 12,},],type: PartSetHeader},},],type: BlockID},},{name: "lastCommitHash",kind: "scalar",scalarType: 12,},{name: "dataHash",kind: "scalar",scalarType: 12,},{name: "validatorsHash",kind: "scalar",scalarType: 12,},{name: "nextValidatorsHash",kind: "scalar",scalarType: 12,},{name: "consensusHash",kind: "scalar",scalarType: 12,},{name: "appHash",kind: "scalar",scalarType: 12,},{name: "lastResultsHash",kind: "scalar",scalarType: 12,},{name: "evidenceHash",kind: "scalar",scalarType: 12,},{name: "proposerAddress",kind: "scalar",scalarType: 12,},],type: Header},},{name: "valset",kind: "list",message: {fields: [{name: "operatorAddress",kind: "scalar",scalarType: 9,},{name: "consensusPubkey",kind: "message",message: {fields: [{name: "typeUrl",kind: "scalar",scalarType: 9,},{name: "value",kind: "scalar",scalarType: 12,},],type: Any},},{name: "jailed",kind: "scalar",scalarType: 8,},{name: "status",kind: "enum",enum: ["UNSPECIFIED","UNBONDED","UNBONDING","BONDED"],},{name: "tokens",kind: "scalar",scalarType: 9,},{name: "delegatorShares",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "description",kind: "message",message: {fields: [{name: "moniker",kind: "scalar",scalarType: 9,},{name: "identity",kind: "scalar",scalarType: 9,},{name: "website",kind: "scalar",scalarType: 9,},{name: "securityContact",kind: "scalar",scalarType: 9,},{name: "details",kind: "scalar",scalarType: 9,},],type: Description},},{name: "unbondingHeight",kind: "scalar",scalarType: 3,},{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},{name: "commission",kind: "message",message: {fields: [{name: "commissionRates",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},{name: "updateTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Timestamp},},],type: Commission},},{name: "minSelfDelegation",kind: "scalar",scalarType: 9,},{name: "unbondingOnHoldRefCount",kind: "scalar",scalarType: 3,},{name: "unbondingIds",kind: "list",},],type: Validator},},],type: HistoricalInfo},},],
  },
  "cosmos.staking.v1beta1.QueryParamsResponse": {
    type: QueryParamsResponse$4,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Duration},},{name: "maxValidators",kind: "scalar",scalarType: 13,},{name: "maxEntries",kind: "scalar",scalarType: 13,},{name: "historicalEntries",kind: "scalar",scalarType: 13,},{name: "bondDenom",kind: "scalar",scalarType: 9,},{name: "minCommissionRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Params$3},},],
  },
  "cosmos.staking.v1beta1.MsgCreateValidator": {
    type: MsgCreateValidator,
    fields: [{name: "commission",kind: "message",message: {fields: [{name: "rate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},{name: "maxChangeRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: CommissionRates},},],
  },
  "cosmos.staking.v1beta1.MsgEditValidator": {
    type: MsgEditValidator,
    fields: [{name: "commissionRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],
  },
  "cosmos.staking.v1beta1.MsgUpdateParams": {
    type: MsgUpdateParams$3,
    fields: [{name: "params",kind: "message",message: {fields: [{name: "unbondingTime",kind: "message",message: {fields: [{name: "seconds",kind: "scalar",scalarType: 3,},{name: "nanos",kind: "scalar",scalarType: 5,},],type: Duration},},{name: "maxValidators",kind: "scalar",scalarType: 13,},{name: "maxEntries",kind: "scalar",scalarType: 13,},{name: "historicalEntries",kind: "scalar",scalarType: 13,},{name: "bondDenom",kind: "scalar",scalarType: 9,},{name: "minCommissionRate",kind: "scalar",scalarType: 9,customType: "LegacyDec",},],type: Params$3},},],
  },
};
describe("cosmosCustomTypePatches.ts", () => {
  describe.each(Object.entries(patches))('patch %s', (typeName, patch: TypePatches[keyof TypePatches]) => {
    it('returns undefined if receives null or undefined', () => {
      expect(patch(null, 'encode')).toBe(undefined);
      expect(patch(null, 'decode')).toBe(undefined);
      expect(patch(undefined, 'encode')).toBe(undefined);
      expect(patch(undefined, 'decode')).toBe(undefined);
    });

    it.each(generateTestCases(typeName, messageTypes))('patches and returns cloned value: %s', (name, value) => {
      const transformedValue = patch(patch(value, 'encode'), 'decode');
      expect(value).toEqual(transformedValue);
      expect(value).not.toBe(transformedValue);
    });
  });

  function generateTestCases(typeName: string, messageTypes: Record<string, MessageSchema>) {
    const type = messageTypes[typeName];
    const cases = type.fields.map((field) => ["single " + field.name + " field", generateMessage(typeName, {
      ...messageTypes,
      [typeName]: {
        ...type,
        fields: [field],
      }
    })]);
    cases.push(["all fields", generateMessage(typeName, messageTypes)]);
    return cases;
  }
});
