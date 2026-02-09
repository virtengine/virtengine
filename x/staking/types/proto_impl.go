// Package types provides proto.Message implementations for staking query types.
package types

import "fmt"

// Proto.Message interface implementations - Query request types.
func (m *QueryParamsRequest) ProtoMessage()  {}
func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorPerformanceRequest) ProtoMessage()  {}
func (m *QueryValidatorPerformanceRequest) Reset()         { *m = QueryValidatorPerformanceRequest{} }
func (m *QueryValidatorPerformanceRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorPerformancesRequest) ProtoMessage()  {}
func (m *QueryValidatorPerformancesRequest) Reset()         { *m = QueryValidatorPerformancesRequest{} }
func (m *QueryValidatorPerformancesRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorRewardRequest) ProtoMessage()  {}
func (m *QueryValidatorRewardRequest) Reset()         { *m = QueryValidatorRewardRequest{} }
func (m *QueryValidatorRewardRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorRewardsRequest) ProtoMessage()  {}
func (m *QueryValidatorRewardsRequest) Reset()         { *m = QueryValidatorRewardsRequest{} }
func (m *QueryValidatorRewardsRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryRewardEpochRequest) ProtoMessage()  {}
func (m *QueryRewardEpochRequest) Reset()         { *m = QueryRewardEpochRequest{} }
func (m *QueryRewardEpochRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QuerySlashRecordsRequest) ProtoMessage()  {}
func (m *QuerySlashRecordsRequest) Reset()         { *m = QuerySlashRecordsRequest{} }
func (m *QuerySlashRecordsRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QuerySigningInfoRequest) ProtoMessage()  {}
func (m *QuerySigningInfoRequest) Reset()         { *m = QuerySigningInfoRequest{} }
func (m *QuerySigningInfoRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryCurrentEpochRequest) ProtoMessage()  {}
func (m *QueryCurrentEpochRequest) Reset()         { *m = QueryCurrentEpochRequest{} }
func (m *QueryCurrentEpochRequest) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface implementations - Query response types.
func (m *QueryParamsResponse) ProtoMessage()  {}
func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorPerformanceResponse) ProtoMessage()  {}
func (m *QueryValidatorPerformanceResponse) Reset()         { *m = QueryValidatorPerformanceResponse{} }
func (m *QueryValidatorPerformanceResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorPerformancesResponse) ProtoMessage()  {}
func (m *QueryValidatorPerformancesResponse) Reset()         { *m = QueryValidatorPerformancesResponse{} }
func (m *QueryValidatorPerformancesResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorRewardResponse) ProtoMessage()  {}
func (m *QueryValidatorRewardResponse) Reset()         { *m = QueryValidatorRewardResponse{} }
func (m *QueryValidatorRewardResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryValidatorRewardsResponse) ProtoMessage()  {}
func (m *QueryValidatorRewardsResponse) Reset()         { *m = QueryValidatorRewardsResponse{} }
func (m *QueryValidatorRewardsResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryRewardEpochResponse) ProtoMessage()  {}
func (m *QueryRewardEpochResponse) Reset()         { *m = QueryRewardEpochResponse{} }
func (m *QueryRewardEpochResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QuerySlashRecordsResponse) ProtoMessage()  {}
func (m *QuerySlashRecordsResponse) Reset()         { *m = QuerySlashRecordsResponse{} }
func (m *QuerySlashRecordsResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QuerySigningInfoResponse) ProtoMessage()  {}
func (m *QuerySigningInfoResponse) Reset()         { *m = QuerySigningInfoResponse{} }
func (m *QuerySigningInfoResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (m *QueryCurrentEpochResponse) ProtoMessage()  {}
func (m *QueryCurrentEpochResponse) Reset()         { *m = QueryCurrentEpochResponse{} }
func (m *QueryCurrentEpochResponse) String() string { return fmt.Sprintf("%+v", *m) }
