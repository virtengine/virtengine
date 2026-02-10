import { LegacyDec } from "../../encoding/customTypes/LegacyDec.ts";
import type * as _protos_cosmos_base_v1beta1_coin from "../protos/cosmos/base/v1beta1/coin.ts";
import type * as _protos_cosmos_distribution_v1beta1_distribution from "../protos/cosmos/distribution/v1beta1/distribution.ts";
import type * as _protos_cosmos_distribution_v1beta1_genesis from "../protos/cosmos/distribution/v1beta1/genesis.ts";
import type * as _protos_cosmos_distribution_v1beta1_query from "../protos/cosmos/distribution/v1beta1/query.ts";
import type * as _protos_cosmos_distribution_v1beta1_tx from "../protos/cosmos/distribution/v1beta1/tx.ts";
import type * as _protos_cosmos_gov_v1beta1_gov from "../protos/cosmos/gov/v1beta1/gov.ts";
import { encodeBinary, decodeBinary } from "../../encoding/binaryEncoding.ts";
import type * as _protos_cosmos_gov_v1beta1_genesis from "../protos/cosmos/gov/v1beta1/genesis.ts";
import type * as _protos_cosmos_gov_v1beta1_query from "../protos/cosmos/gov/v1beta1/query.ts";
import type * as _protos_cosmos_gov_v1beta1_tx from "../protos/cosmos/gov/v1beta1/tx.ts";
import type * as _protos_cosmos_mint_v1beta1_mint from "../protos/cosmos/mint/v1beta1/mint.ts";
import type * as _protos_cosmos_mint_v1beta1_genesis from "../protos/cosmos/mint/v1beta1/genesis.ts";
import type * as _protos_cosmos_mint_v1beta1_query from "../protos/cosmos/mint/v1beta1/query.ts";
import type * as _protos_cosmos_mint_v1beta1_tx from "../protos/cosmos/mint/v1beta1/tx.ts";
import type * as _protos_cosmos_protocolpool_v1_types from "../protos/cosmos/protocolpool/v1/types.ts";
import type * as _protos_cosmos_protocolpool_v1_genesis from "../protos/cosmos/protocolpool/v1/genesis.ts";
import type * as _protos_cosmos_protocolpool_v1_query from "../protos/cosmos/protocolpool/v1/query.ts";
import type * as _protos_cosmos_protocolpool_v1_tx from "../protos/cosmos/protocolpool/v1/tx.ts";
import type * as _protos_cosmos_slashing_v1beta1_slashing from "../protos/cosmos/slashing/v1beta1/slashing.ts";
import type * as _protos_cosmos_slashing_v1beta1_genesis from "../protos/cosmos/slashing/v1beta1/genesis.ts";
import type * as _protos_cosmos_slashing_v1beta1_query from "../protos/cosmos/slashing/v1beta1/query.ts";
import type * as _protos_cosmos_slashing_v1beta1_tx from "../protos/cosmos/slashing/v1beta1/tx.ts";
import type * as _protos_cosmos_staking_v1beta1_staking from "../protos/cosmos/staking/v1beta1/staking.ts";
import type * as _protos_cosmos_staking_v1beta1_genesis from "../protos/cosmos/staking/v1beta1/genesis.ts";
import type * as _protos_cosmos_staking_v1beta1_query from "../protos/cosmos/staking/v1beta1/query.ts";
import type * as _protos_cosmos_staking_v1beta1_tx from "../protos/cosmos/staking/v1beta1/tx.ts";

const p = {
  "cosmos.base.v1beta1.DecCoin"(value: _protos_cosmos_base_v1beta1_coin.DecCoin | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.amount != null) newValue.amount = LegacyDec[transformType](value.amount);
    return newValue;
  },
  "cosmos.base.v1beta1.DecProto"(value: _protos_cosmos_base_v1beta1_coin.DecProto | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.dec != null) newValue.dec = LegacyDec[transformType](value.dec);
    return newValue;
  },
  "cosmos.distribution.v1beta1.Params"(value: _protos_cosmos_distribution_v1beta1_distribution.Params | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.communityTax != null) newValue.communityTax = LegacyDec[transformType](value.communityTax);
    if (value.baseProposerReward != null) newValue.baseProposerReward = LegacyDec[transformType](value.baseProposerReward);
    if (value.bonusProposerReward != null) newValue.bonusProposerReward = LegacyDec[transformType](value.bonusProposerReward);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorHistoricalRewards"(value: _protos_cosmos_distribution_v1beta1_distribution.ValidatorHistoricalRewards | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.cumulativeRewardRatio) newValue.cumulativeRewardRatio = value.cumulativeRewardRatio.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorCurrentRewards"(value: _protos_cosmos_distribution_v1beta1_distribution.ValidatorCurrentRewards | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards) newValue.rewards = value.rewards.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorAccumulatedCommission"(value: _protos_cosmos_distribution_v1beta1_distribution.ValidatorAccumulatedCommission | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.commission) newValue.commission = value.commission.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorOutstandingRewards"(value: _protos_cosmos_distribution_v1beta1_distribution.ValidatorOutstandingRewards | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards) newValue.rewards = value.rewards.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorSlashEvent"(value: _protos_cosmos_distribution_v1beta1_distribution.ValidatorSlashEvent | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.fraction != null) newValue.fraction = LegacyDec[transformType](value.fraction);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorSlashEvents"(value: _protos_cosmos_distribution_v1beta1_distribution.ValidatorSlashEvents | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.validatorSlashEvents) newValue.validatorSlashEvents = value.validatorSlashEvents.map((item) => p["cosmos.distribution.v1beta1.ValidatorSlashEvent"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.FeePool"(value: _protos_cosmos_distribution_v1beta1_distribution.FeePool | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.communityPool) newValue.communityPool = value.communityPool.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.DelegatorStartingInfo"(value: _protos_cosmos_distribution_v1beta1_distribution.DelegatorStartingInfo | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.stake != null) newValue.stake = LegacyDec[transformType](value.stake);
    return newValue;
  },
  "cosmos.distribution.v1beta1.DelegationDelegatorReward"(value: _protos_cosmos_distribution_v1beta1_distribution.DelegationDelegatorReward | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.reward) newValue.reward = value.reward.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorOutstandingRewardsRecord"(value: _protos_cosmos_distribution_v1beta1_genesis.ValidatorOutstandingRewardsRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.outstandingRewards) newValue.outstandingRewards = value.outstandingRewards.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorAccumulatedCommissionRecord"(value: _protos_cosmos_distribution_v1beta1_genesis.ValidatorAccumulatedCommissionRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.accumulated != null) newValue.accumulated = p["cosmos.distribution.v1beta1.ValidatorAccumulatedCommission"](value.accumulated, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorHistoricalRewardsRecord"(value: _protos_cosmos_distribution_v1beta1_genesis.ValidatorHistoricalRewardsRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards != null) newValue.rewards = p["cosmos.distribution.v1beta1.ValidatorHistoricalRewards"](value.rewards, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorCurrentRewardsRecord"(value: _protos_cosmos_distribution_v1beta1_genesis.ValidatorCurrentRewardsRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards != null) newValue.rewards = p["cosmos.distribution.v1beta1.ValidatorCurrentRewards"](value.rewards, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.DelegatorStartingInfoRecord"(value: _protos_cosmos_distribution_v1beta1_genesis.DelegatorStartingInfoRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.startingInfo != null) newValue.startingInfo = p["cosmos.distribution.v1beta1.DelegatorStartingInfo"](value.startingInfo, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.ValidatorSlashEventRecord"(value: _protos_cosmos_distribution_v1beta1_genesis.ValidatorSlashEventRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.validatorSlashEvent != null) newValue.validatorSlashEvent = p["cosmos.distribution.v1beta1.ValidatorSlashEvent"](value.validatorSlashEvent, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.GenesisState"(value: _protos_cosmos_distribution_v1beta1_genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.distribution.v1beta1.Params"](value.params, transformType);
    if (value.feePool != null) newValue.feePool = p["cosmos.distribution.v1beta1.FeePool"](value.feePool, transformType);
    if (value.outstandingRewards) newValue.outstandingRewards = value.outstandingRewards.map((item) => p["cosmos.distribution.v1beta1.ValidatorOutstandingRewardsRecord"](item, transformType)!);
    if (value.validatorAccumulatedCommissions) newValue.validatorAccumulatedCommissions = value.validatorAccumulatedCommissions.map((item) => p["cosmos.distribution.v1beta1.ValidatorAccumulatedCommissionRecord"](item, transformType)!);
    if (value.validatorHistoricalRewards) newValue.validatorHistoricalRewards = value.validatorHistoricalRewards.map((item) => p["cosmos.distribution.v1beta1.ValidatorHistoricalRewardsRecord"](item, transformType)!);
    if (value.validatorCurrentRewards) newValue.validatorCurrentRewards = value.validatorCurrentRewards.map((item) => p["cosmos.distribution.v1beta1.ValidatorCurrentRewardsRecord"](item, transformType)!);
    if (value.delegatorStartingInfos) newValue.delegatorStartingInfos = value.delegatorStartingInfos.map((item) => p["cosmos.distribution.v1beta1.DelegatorStartingInfoRecord"](item, transformType)!);
    if (value.validatorSlashEvents) newValue.validatorSlashEvents = value.validatorSlashEvents.map((item) => p["cosmos.distribution.v1beta1.ValidatorSlashEventRecord"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryParamsResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryParamsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.distribution.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryValidatorDistributionInfoResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryValidatorDistributionInfoResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.selfBondRewards) newValue.selfBondRewards = value.selfBondRewards.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    if (value.commission) newValue.commission = value.commission.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryValidatorOutstandingRewardsResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryValidatorOutstandingRewardsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards != null) newValue.rewards = p["cosmos.distribution.v1beta1.ValidatorOutstandingRewards"](value.rewards, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryValidatorCommissionResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryValidatorCommissionResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.commission != null) newValue.commission = p["cosmos.distribution.v1beta1.ValidatorAccumulatedCommission"](value.commission, transformType);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryValidatorSlashesResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryValidatorSlashesResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.slashes) newValue.slashes = value.slashes.map((item) => p["cosmos.distribution.v1beta1.ValidatorSlashEvent"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryDelegationRewardsResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryDelegationRewardsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards) newValue.rewards = value.rewards.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryDelegationTotalRewardsResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryDelegationTotalRewardsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rewards) newValue.rewards = value.rewards.map((item) => p["cosmos.distribution.v1beta1.DelegationDelegatorReward"](item, transformType)!);
    if (value.total) newValue.total = value.total.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.QueryCommunityPoolResponse"(value: _protos_cosmos_distribution_v1beta1_query.QueryCommunityPoolResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.pool) newValue.pool = value.pool.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "cosmos.distribution.v1beta1.MsgUpdateParams"(value: _protos_cosmos_distribution_v1beta1_tx.MsgUpdateParams | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.distribution.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.gov.v1beta1.WeightedVoteOption"(value: _protos_cosmos_gov_v1beta1_gov.WeightedVoteOption | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.weight != null) newValue.weight = LegacyDec[transformType](value.weight);
    return newValue;
  },
  "cosmos.gov.v1beta1.Vote"(value: _protos_cosmos_gov_v1beta1_gov.Vote | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.options) newValue.options = value.options.map((item) => p["cosmos.gov.v1beta1.WeightedVoteOption"](item, transformType)!);
    return newValue;
  },
  "cosmos.gov.v1beta1.TallyParams"(value: _protos_cosmos_gov_v1beta1_gov.TallyParams | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.quorum != null) newValue.quorum = encodeBinary(LegacyDec[transformType](decodeBinary(value.quorum)));
    if (value.threshold != null) newValue.threshold = encodeBinary(LegacyDec[transformType](decodeBinary(value.threshold)));
    if (value.vetoThreshold != null) newValue.vetoThreshold = encodeBinary(LegacyDec[transformType](decodeBinary(value.vetoThreshold)));
    return newValue;
  },
  "cosmos.gov.v1beta1.GenesisState"(value: _protos_cosmos_gov_v1beta1_genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.votes) newValue.votes = value.votes.map((item) => p["cosmos.gov.v1beta1.Vote"](item, transformType)!);
    if (value.tallyParams != null) newValue.tallyParams = p["cosmos.gov.v1beta1.TallyParams"](value.tallyParams, transformType);
    return newValue;
  },
  "cosmos.gov.v1beta1.QueryVoteResponse"(value: _protos_cosmos_gov_v1beta1_query.QueryVoteResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.vote != null) newValue.vote = p["cosmos.gov.v1beta1.Vote"](value.vote, transformType);
    return newValue;
  },
  "cosmos.gov.v1beta1.QueryVotesResponse"(value: _protos_cosmos_gov_v1beta1_query.QueryVotesResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.votes) newValue.votes = value.votes.map((item) => p["cosmos.gov.v1beta1.Vote"](item, transformType)!);
    return newValue;
  },
  "cosmos.gov.v1beta1.QueryParamsResponse"(value: _protos_cosmos_gov_v1beta1_query.QueryParamsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.tallyParams != null) newValue.tallyParams = p["cosmos.gov.v1beta1.TallyParams"](value.tallyParams, transformType);
    return newValue;
  },
  "cosmos.gov.v1beta1.MsgVoteWeighted"(value: _protos_cosmos_gov_v1beta1_tx.MsgVoteWeighted | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.options) newValue.options = value.options.map((item) => p["cosmos.gov.v1beta1.WeightedVoteOption"](item, transformType)!);
    return newValue;
  },
  "cosmos.mint.v1beta1.Minter"(value: _protos_cosmos_mint_v1beta1_mint.Minter | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.inflation != null) newValue.inflation = LegacyDec[transformType](value.inflation);
    if (value.annualProvisions != null) newValue.annualProvisions = LegacyDec[transformType](value.annualProvisions);
    return newValue;
  },
  "cosmos.mint.v1beta1.Params"(value: _protos_cosmos_mint_v1beta1_mint.Params | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.inflationRateChange != null) newValue.inflationRateChange = LegacyDec[transformType](value.inflationRateChange);
    if (value.inflationMax != null) newValue.inflationMax = LegacyDec[transformType](value.inflationMax);
    if (value.inflationMin != null) newValue.inflationMin = LegacyDec[transformType](value.inflationMin);
    if (value.goalBonded != null) newValue.goalBonded = LegacyDec[transformType](value.goalBonded);
    return newValue;
  },
  "cosmos.mint.v1beta1.GenesisState"(value: _protos_cosmos_mint_v1beta1_genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.minter != null) newValue.minter = p["cosmos.mint.v1beta1.Minter"](value.minter, transformType);
    if (value.params != null) newValue.params = p["cosmos.mint.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.mint.v1beta1.QueryParamsResponse"(value: _protos_cosmos_mint_v1beta1_query.QueryParamsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.mint.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.mint.v1beta1.QueryInflationResponse"(value: _protos_cosmos_mint_v1beta1_query.QueryInflationResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.inflation != null) newValue.inflation = encodeBinary(LegacyDec[transformType](decodeBinary(value.inflation)));
    return newValue;
  },
  "cosmos.mint.v1beta1.QueryAnnualProvisionsResponse"(value: _protos_cosmos_mint_v1beta1_query.QueryAnnualProvisionsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.annualProvisions != null) newValue.annualProvisions = encodeBinary(LegacyDec[transformType](decodeBinary(value.annualProvisions)));
    return newValue;
  },
  "cosmos.mint.v1beta1.MsgUpdateParams"(value: _protos_cosmos_mint_v1beta1_tx.MsgUpdateParams | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.mint.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.protocolpool.v1.ContinuousFund"(value: _protos_cosmos_protocolpool_v1_types.ContinuousFund | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.percentage != null) newValue.percentage = LegacyDec[transformType](value.percentage);
    return newValue;
  },
  "cosmos.protocolpool.v1.GenesisState"(value: _protos_cosmos_protocolpool_v1_genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.continuousFunds) newValue.continuousFunds = value.continuousFunds.map((item) => p["cosmos.protocolpool.v1.ContinuousFund"](item, transformType)!);
    return newValue;
  },
  "cosmos.protocolpool.v1.QueryContinuousFundResponse"(value: _protos_cosmos_protocolpool_v1_query.QueryContinuousFundResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.continuousFund != null) newValue.continuousFund = p["cosmos.protocolpool.v1.ContinuousFund"](value.continuousFund, transformType);
    return newValue;
  },
  "cosmos.protocolpool.v1.QueryContinuousFundsResponse"(value: _protos_cosmos_protocolpool_v1_query.QueryContinuousFundsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.continuousFunds) newValue.continuousFunds = value.continuousFunds.map((item) => p["cosmos.protocolpool.v1.ContinuousFund"](item, transformType)!);
    return newValue;
  },
  "cosmos.protocolpool.v1.MsgCreateContinuousFund"(value: _protos_cosmos_protocolpool_v1_tx.MsgCreateContinuousFund | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.percentage != null) newValue.percentage = LegacyDec[transformType](value.percentage);
    return newValue;
  },
  "cosmos.slashing.v1beta1.Params"(value: _protos_cosmos_slashing_v1beta1_slashing.Params | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.minSignedPerWindow != null) newValue.minSignedPerWindow = encodeBinary(LegacyDec[transformType](decodeBinary(value.minSignedPerWindow)));
    if (value.slashFractionDoubleSign != null) newValue.slashFractionDoubleSign = encodeBinary(LegacyDec[transformType](decodeBinary(value.slashFractionDoubleSign)));
    if (value.slashFractionDowntime != null) newValue.slashFractionDowntime = encodeBinary(LegacyDec[transformType](decodeBinary(value.slashFractionDowntime)));
    return newValue;
  },
  "cosmos.slashing.v1beta1.GenesisState"(value: _protos_cosmos_slashing_v1beta1_genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.slashing.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.slashing.v1beta1.QueryParamsResponse"(value: _protos_cosmos_slashing_v1beta1_query.QueryParamsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.slashing.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.slashing.v1beta1.MsgUpdateParams"(value: _protos_cosmos_slashing_v1beta1_tx.MsgUpdateParams | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.slashing.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.HistoricalInfo"(value: _protos_cosmos_staking_v1beta1_staking.HistoricalInfo | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.valset) newValue.valset = value.valset.map((item) => p["cosmos.staking.v1beta1.Validator"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.Validator"(value: _protos_cosmos_staking_v1beta1_staking.Validator | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.delegatorShares != null) newValue.delegatorShares = LegacyDec[transformType](value.delegatorShares);
    if (value.commission != null) newValue.commission = p["cosmos.staking.v1beta1.Commission"](value.commission, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.Commission"(value: _protos_cosmos_staking_v1beta1_staking.Commission | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.commissionRates != null) newValue.commissionRates = p["cosmos.staking.v1beta1.CommissionRates"](value.commissionRates, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.CommissionRates"(value: _protos_cosmos_staking_v1beta1_staking.CommissionRates | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rate != null) newValue.rate = LegacyDec[transformType](value.rate);
    if (value.maxRate != null) newValue.maxRate = LegacyDec[transformType](value.maxRate);
    if (value.maxChangeRate != null) newValue.maxChangeRate = LegacyDec[transformType](value.maxChangeRate);
    return newValue;
  },
  "cosmos.staking.v1beta1.Delegation"(value: _protos_cosmos_staking_v1beta1_staking.Delegation | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.shares != null) newValue.shares = LegacyDec[transformType](value.shares);
    return newValue;
  },
  "cosmos.staking.v1beta1.RedelegationEntry"(value: _protos_cosmos_staking_v1beta1_staking.RedelegationEntry | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.sharesDst != null) newValue.sharesDst = LegacyDec[transformType](value.sharesDst);
    return newValue;
  },
  "cosmos.staking.v1beta1.Redelegation"(value: _protos_cosmos_staking_v1beta1_staking.Redelegation | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.entries) newValue.entries = value.entries.map((item) => p["cosmos.staking.v1beta1.RedelegationEntry"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.Params"(value: _protos_cosmos_staking_v1beta1_staking.Params | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.minCommissionRate != null) newValue.minCommissionRate = LegacyDec[transformType](value.minCommissionRate);
    return newValue;
  },
  "cosmos.staking.v1beta1.DelegationResponse"(value: _protos_cosmos_staking_v1beta1_staking.DelegationResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.delegation != null) newValue.delegation = p["cosmos.staking.v1beta1.Delegation"](value.delegation, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.RedelegationEntryResponse"(value: _protos_cosmos_staking_v1beta1_staking.RedelegationEntryResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.redelegationEntry != null) newValue.redelegationEntry = p["cosmos.staking.v1beta1.RedelegationEntry"](value.redelegationEntry, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.RedelegationResponse"(value: _protos_cosmos_staking_v1beta1_staking.RedelegationResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.redelegation != null) newValue.redelegation = p["cosmos.staking.v1beta1.Redelegation"](value.redelegation, transformType);
    if (value.entries) newValue.entries = value.entries.map((item) => p["cosmos.staking.v1beta1.RedelegationEntryResponse"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.GenesisState"(value: _protos_cosmos_staking_v1beta1_genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.staking.v1beta1.Params"](value.params, transformType);
    if (value.validators) newValue.validators = value.validators.map((item) => p["cosmos.staking.v1beta1.Validator"](item, transformType)!);
    if (value.delegations) newValue.delegations = value.delegations.map((item) => p["cosmos.staking.v1beta1.Delegation"](item, transformType)!);
    if (value.redelegations) newValue.redelegations = value.redelegations.map((item) => p["cosmos.staking.v1beta1.Redelegation"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryValidatorsResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryValidatorsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.validators) newValue.validators = value.validators.map((item) => p["cosmos.staking.v1beta1.Validator"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryValidatorResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryValidatorResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.validator != null) newValue.validator = p["cosmos.staking.v1beta1.Validator"](value.validator, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryValidatorDelegationsResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryValidatorDelegationsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.delegationResponses) newValue.delegationResponses = value.delegationResponses.map((item) => p["cosmos.staking.v1beta1.DelegationResponse"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryDelegationResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryDelegationResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.delegationResponse != null) newValue.delegationResponse = p["cosmos.staking.v1beta1.DelegationResponse"](value.delegationResponse, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryDelegatorDelegationsResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryDelegatorDelegationsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.delegationResponses) newValue.delegationResponses = value.delegationResponses.map((item) => p["cosmos.staking.v1beta1.DelegationResponse"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryRedelegationsResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryRedelegationsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.redelegationResponses) newValue.redelegationResponses = value.redelegationResponses.map((item) => p["cosmos.staking.v1beta1.RedelegationResponse"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryDelegatorValidatorsResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryDelegatorValidatorsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.validators) newValue.validators = value.validators.map((item) => p["cosmos.staking.v1beta1.Validator"](item, transformType)!);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryDelegatorValidatorResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryDelegatorValidatorResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.validator != null) newValue.validator = p["cosmos.staking.v1beta1.Validator"](value.validator, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryHistoricalInfoResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryHistoricalInfoResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.hist != null) newValue.hist = p["cosmos.staking.v1beta1.HistoricalInfo"](value.hist, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.QueryParamsResponse"(value: _protos_cosmos_staking_v1beta1_query.QueryParamsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.staking.v1beta1.Params"](value.params, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.MsgCreateValidator"(value: _protos_cosmos_staking_v1beta1_tx.MsgCreateValidator | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.commission != null) newValue.commission = p["cosmos.staking.v1beta1.CommissionRates"](value.commission, transformType);
    return newValue;
  },
  "cosmos.staking.v1beta1.MsgEditValidator"(value: _protos_cosmos_staking_v1beta1_tx.MsgEditValidator | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.commissionRate != null) newValue.commissionRate = LegacyDec[transformType](value.commissionRate);
    return newValue;
  },
  "cosmos.staking.v1beta1.MsgUpdateParams"(value: _protos_cosmos_staking_v1beta1_tx.MsgUpdateParams | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.params != null) newValue.params = p["cosmos.staking.v1beta1.Params"](value.params, transformType);
    return newValue;
  }
};

export const patches = p;
