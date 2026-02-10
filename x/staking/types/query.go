// Package types contains query request/response types for staking gRPC queries.
package types

// QueryParamsRequest is the request for module params.
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for module params.
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

// QueryValidatorPerformanceRequest fetches a validator performance record.
type QueryValidatorPerformanceRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address"`
	Epoch            uint64 `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch"`
}

// QueryValidatorPerformanceResponse returns a validator performance record.
type QueryValidatorPerformanceResponse struct {
	Performance ValidatorPerformance `protobuf:"bytes,1,opt,name=performance,proto3" json:"performance"`
}

// QueryValidatorPerformancesRequest fetches performance records for an epoch.
type QueryValidatorPerformancesRequest struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch"`
}

// QueryValidatorPerformancesResponse returns performance records.
type QueryValidatorPerformancesResponse struct {
	Performances []ValidatorPerformance `protobuf:"bytes,1,rep,name=performances,proto3" json:"performances"`
}

// QueryValidatorRewardRequest fetches a validator reward for an epoch.
type QueryValidatorRewardRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address"`
	Epoch            uint64 `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch"`
}

// QueryValidatorRewardResponse returns a validator reward.
type QueryValidatorRewardResponse struct {
	Reward ValidatorReward `protobuf:"bytes,1,opt,name=reward,proto3" json:"reward"`
}

// QueryValidatorRewardsRequest fetches all rewards for a validator.
type QueryValidatorRewardsRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address"`
}

// QueryValidatorRewardsResponse returns validator rewards.
type QueryValidatorRewardsResponse struct {
	Rewards []ValidatorReward `protobuf:"bytes,1,rep,name=rewards,proto3" json:"rewards"`
}

// QueryRewardEpochRequest fetches a reward epoch by number.
type QueryRewardEpochRequest struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch"`
}

// QueryRewardEpochResponse returns a reward epoch.
type QueryRewardEpochResponse struct {
	RewardEpoch RewardEpoch `protobuf:"bytes,1,opt,name=reward_epoch,json=rewardEpoch,proto3" json:"reward_epoch"`
}

// QuerySlashRecordsRequest fetches slash records for a validator.
type QuerySlashRecordsRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address"`
}

// QuerySlashRecordsResponse returns slash records for a validator.
type QuerySlashRecordsResponse struct {
	Records []SlashRecord `protobuf:"bytes,1,rep,name=records,proto3" json:"records"`
}

// QuerySigningInfoRequest fetches validator signing info.
type QuerySigningInfoRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address"`
}

// QuerySigningInfoResponse returns validator signing info.
type QuerySigningInfoResponse struct {
	Info ValidatorSigningInfo `protobuf:"bytes,1,opt,name=info,proto3" json:"info"`
}

// QueryCurrentEpochRequest fetches the current epoch.
type QueryCurrentEpochRequest struct{}

// QueryCurrentEpochResponse returns the current epoch.
type QueryCurrentEpochResponse struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch"`
}
