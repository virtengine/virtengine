<!-- This file is auto-generated. Please do not modify it yourself. -->
 # Protobuf Documentation
 <a name="top"></a>

 ## Table of Contents
 
 - [virtengine/base/attributes/v1/attribute.proto](#virtengine/base/attributes/v1/attribute.proto)
     - [Attribute](#virtengine.base.attributes.v1.Attribute)
     - [PlacementRequirements](#virtengine.base.attributes.v1.PlacementRequirements)
     - [SignedBy](#virtengine.base.attributes.v1.SignedBy)
   
 - [virtengine/audit/v1/audit.proto](#virtengine/audit/v1/audit.proto)
     - [AttributesFilters](#virtengine.audit.v1.AttributesFilters)
     - [AuditedAttributesStore](#virtengine.audit.v1.AuditedAttributesStore)
     - [AuditedProvider](#virtengine.audit.v1.AuditedProvider)
   
 - [virtengine/audit/v1/event.proto](#virtengine/audit/v1/event.proto)
     - [EventTrustedAuditorCreated](#virtengine.audit.v1.EventTrustedAuditorCreated)
     - [EventTrustedAuditorDeleted](#virtengine.audit.v1.EventTrustedAuditorDeleted)
   
 - [virtengine/audit/v1/genesis.proto](#virtengine/audit/v1/genesis.proto)
     - [GenesisState](#virtengine.audit.v1.GenesisState)
   
 - [virtengine/audit/v1/msg.proto](#virtengine/audit/v1/msg.proto)
     - [MsgDeleteProviderAttributes](#virtengine.audit.v1.MsgDeleteProviderAttributes)
     - [MsgDeleteProviderAttributesResponse](#virtengine.audit.v1.MsgDeleteProviderAttributesResponse)
     - [MsgSignProviderAttributes](#virtengine.audit.v1.MsgSignProviderAttributes)
     - [MsgSignProviderAttributesResponse](#virtengine.audit.v1.MsgSignProviderAttributesResponse)
   
 - [virtengine/audit/v1/query.proto](#virtengine/audit/v1/query.proto)
     - [QueryAllProvidersAttributesRequest](#virtengine.audit.v1.QueryAllProvidersAttributesRequest)
     - [QueryAuditorAttributesRequest](#virtengine.audit.v1.QueryAuditorAttributesRequest)
     - [QueryProviderAttributesRequest](#virtengine.audit.v1.QueryProviderAttributesRequest)
     - [QueryProviderAuditorRequest](#virtengine.audit.v1.QueryProviderAuditorRequest)
     - [QueryProviderRequest](#virtengine.audit.v1.QueryProviderRequest)
     - [QueryProvidersResponse](#virtengine.audit.v1.QueryProvidersResponse)
   
     - [Query](#virtengine.audit.v1.Query)
   
 - [virtengine/audit/v1/service.proto](#virtengine/audit/v1/service.proto)
     - [Msg](#virtengine.audit.v1.Msg)
   
 - [virtengine/base/deposit/v1/deposit.proto](#virtengine/base/deposit/v1/deposit.proto)
     - [Deposit](#virtengine.base.deposit.v1.Deposit)
   
     - [Source](#virtengine.base.deposit.v1.Source)
   
 - [virtengine/base/offchain/sign/v1/sign.proto](#virtengine/base/offchain/sign/v1/sign.proto)
     - [MsgSignData](#virtengine.base.offchain.sign.v1.MsgSignData)
   
 - [virtengine/base/resources/v1beta4/resourcevalue.proto](#virtengine/base/resources/v1beta4/resourcevalue.proto)
     - [ResourceValue](#virtengine.base.resources.v1beta4.ResourceValue)
   
 - [virtengine/base/resources/v1beta4/cpu.proto](#virtengine/base/resources/v1beta4/cpu.proto)
     - [CPU](#virtengine.base.resources.v1beta4.CPU)
   
 - [virtengine/base/resources/v1beta4/endpoint.proto](#virtengine/base/resources/v1beta4/endpoint.proto)
     - [Endpoint](#virtengine.base.resources.v1beta4.Endpoint)
   
     - [Endpoint.Kind](#virtengine.base.resources.v1beta4.Endpoint.Kind)
   
 - [virtengine/base/resources/v1beta4/gpu.proto](#virtengine/base/resources/v1beta4/gpu.proto)
     - [GPU](#virtengine.base.resources.v1beta4.GPU)
   
 - [virtengine/base/resources/v1beta4/memory.proto](#virtengine/base/resources/v1beta4/memory.proto)
     - [Memory](#virtengine.base.resources.v1beta4.Memory)
   
 - [virtengine/base/resources/v1beta4/storage.proto](#virtengine/base/resources/v1beta4/storage.proto)
     - [Storage](#virtengine.base.resources.v1beta4.Storage)
   
 - [virtengine/base/resources/v1beta4/resources.proto](#virtengine/base/resources/v1beta4/resources.proto)
     - [Resources](#virtengine.base.resources.v1beta4.Resources)
   
 - [virtengine/benchmark/v1/tx.proto](#virtengine/benchmark/v1/tx.proto)
     - [AnomalyDetectedEvent](#virtengine.benchmark.v1.AnomalyDetectedEvent)
     - [AnomalyFlag](#virtengine.benchmark.v1.AnomalyFlag)
     - [AnomalyResolvedEvent](#virtengine.benchmark.v1.AnomalyResolvedEvent)
     - [BenchmarkResult](#virtengine.benchmark.v1.BenchmarkResult)
     - [BenchmarksPrunedEvent](#virtengine.benchmark.v1.BenchmarksPrunedEvent)
     - [BenchmarksSubmittedEvent](#virtengine.benchmark.v1.BenchmarksSubmittedEvent)
     - [Challenge](#virtengine.benchmark.v1.Challenge)
     - [ChallengeCompletedEvent](#virtengine.benchmark.v1.ChallengeCompletedEvent)
     - [ChallengeExpiredEvent](#virtengine.benchmark.v1.ChallengeExpiredEvent)
     - [ChallengeRequestedEvent](#virtengine.benchmark.v1.ChallengeRequestedEvent)
     - [GenesisState](#virtengine.benchmark.v1.GenesisState)
     - [MsgFlagProvider](#virtengine.benchmark.v1.MsgFlagProvider)
     - [MsgFlagProviderResponse](#virtengine.benchmark.v1.MsgFlagProviderResponse)
     - [MsgRequestChallenge](#virtengine.benchmark.v1.MsgRequestChallenge)
     - [MsgRequestChallengeResponse](#virtengine.benchmark.v1.MsgRequestChallengeResponse)
     - [MsgResolveAnomalyFlag](#virtengine.benchmark.v1.MsgResolveAnomalyFlag)
     - [MsgResolveAnomalyFlagResponse](#virtengine.benchmark.v1.MsgResolveAnomalyFlagResponse)
     - [MsgRespondChallenge](#virtengine.benchmark.v1.MsgRespondChallenge)
     - [MsgRespondChallengeResponse](#virtengine.benchmark.v1.MsgRespondChallengeResponse)
     - [MsgSubmitBenchmarks](#virtengine.benchmark.v1.MsgSubmitBenchmarks)
     - [MsgSubmitBenchmarksResponse](#virtengine.benchmark.v1.MsgSubmitBenchmarksResponse)
     - [MsgUnflagProvider](#virtengine.benchmark.v1.MsgUnflagProvider)
     - [MsgUnflagProviderResponse](#virtengine.benchmark.v1.MsgUnflagProviderResponse)
     - [Params](#virtengine.benchmark.v1.Params)
     - [ProviderBenchmark](#virtengine.benchmark.v1.ProviderBenchmark)
     - [ProviderFlaggedEvent](#virtengine.benchmark.v1.ProviderFlaggedEvent)
     - [ProviderUnflaggedEvent](#virtengine.benchmark.v1.ProviderUnflaggedEvent)
     - [ReliabilityScoreUpdatedEvent](#virtengine.benchmark.v1.ReliabilityScoreUpdatedEvent)
   
     - [Msg](#virtengine.benchmark.v1.Msg)
   
 - [virtengine/bme/v1/types.proto](#virtengine/bme/v1/types.proto)
     - [BurnMintPair](#virtengine.bme.v1.BurnMintPair)
     - [CoinPrice](#virtengine.bme.v1.CoinPrice)
     - [CollateralRatio](#virtengine.bme.v1.CollateralRatio)
     - [LedgerID](#virtengine.bme.v1.LedgerID)
     - [LedgerPendingRecord](#virtengine.bme.v1.LedgerPendingRecord)
     - [LedgerRecord](#virtengine.bme.v1.LedgerRecord)
     - [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID)
     - [MintEpoch](#virtengine.bme.v1.MintEpoch)
     - [State](#virtengine.bme.v1.State)
     - [Status](#virtengine.bme.v1.Status)
   
     - [LedgerRecordStatus](#virtengine.bme.v1.LedgerRecordStatus)
     - [MintStatus](#virtengine.bme.v1.MintStatus)
   
 - [virtengine/bme/v1/events.proto](#virtengine/bme/v1/events.proto)
     - [EventLedgerRecordExecuted](#virtengine.bme.v1.EventLedgerRecordExecuted)
     - [EventMintStatusChange](#virtengine.bme.v1.EventMintStatusChange)
     - [EventVaultSeeded](#virtengine.bme.v1.EventVaultSeeded)
   
 - [virtengine/bme/v1/params.proto](#virtengine/bme/v1/params.proto)
     - [Params](#virtengine.bme.v1.Params)
   
 - [virtengine/bme/v1/genesis.proto](#virtengine/bme/v1/genesis.proto)
     - [GenesisLedgerPendingRecord](#virtengine.bme.v1.GenesisLedgerPendingRecord)
     - [GenesisLedgerRecord](#virtengine.bme.v1.GenesisLedgerRecord)
     - [GenesisLedgerState](#virtengine.bme.v1.GenesisLedgerState)
     - [GenesisState](#virtengine.bme.v1.GenesisState)
     - [GenesisVaultState](#virtengine.bme.v1.GenesisVaultState)
   
 - [virtengine/bme/v1/msgs.proto](#virtengine/bme/v1/msgs.proto)
     - [MsgBurnACT](#virtengine.bme.v1.MsgBurnACT)
     - [MsgBurnACTResponse](#virtengine.bme.v1.MsgBurnACTResponse)
     - [MsgBurnMint](#virtengine.bme.v1.MsgBurnMint)
     - [MsgBurnMintResponse](#virtengine.bme.v1.MsgBurnMintResponse)
     - [MsgMintACT](#virtengine.bme.v1.MsgMintACT)
     - [MsgMintACTResponse](#virtengine.bme.v1.MsgMintACTResponse)
     - [MsgSeedVault](#virtengine.bme.v1.MsgSeedVault)
     - [MsgSeedVaultResponse](#virtengine.bme.v1.MsgSeedVaultResponse)
     - [MsgUpdateParams](#virtengine.bme.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.bme.v1.MsgUpdateParamsResponse)
   
 - [virtengine/bme/v1/query.proto](#virtengine/bme/v1/query.proto)
     - [QueryParamsRequest](#virtengine.bme.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.bme.v1.QueryParamsResponse)
     - [QueryStatusRequest](#virtengine.bme.v1.QueryStatusRequest)
     - [QueryStatusResponse](#virtengine.bme.v1.QueryStatusResponse)
     - [QueryVaultStateRequest](#virtengine.bme.v1.QueryVaultStateRequest)
     - [QueryVaultStateResponse](#virtengine.bme.v1.QueryVaultStateResponse)
   
     - [Query](#virtengine.bme.v1.Query)
   
 - [virtengine/bme/v1/service.proto](#virtengine/bme/v1/service.proto)
     - [Msg](#virtengine.bme.v1.Msg)
   
 - [virtengine/cert/v1/cert.proto](#virtengine/cert/v1/cert.proto)
     - [Certificate](#virtengine.cert.v1.Certificate)
     - [ID](#virtengine.cert.v1.ID)
   
     - [State](#virtengine.cert.v1.State)
   
 - [virtengine/cert/v1/filters.proto](#virtengine/cert/v1/filters.proto)
     - [CertificateFilter](#virtengine.cert.v1.CertificateFilter)
   
 - [virtengine/cert/v1/genesis.proto](#virtengine/cert/v1/genesis.proto)
     - [GenesisCertificate](#virtengine.cert.v1.GenesisCertificate)
     - [GenesisState](#virtengine.cert.v1.GenesisState)
   
 - [virtengine/cert/v1/msg.proto](#virtengine/cert/v1/msg.proto)
     - [MsgCreateCertificate](#virtengine.cert.v1.MsgCreateCertificate)
     - [MsgCreateCertificateResponse](#virtengine.cert.v1.MsgCreateCertificateResponse)
     - [MsgRevokeCertificate](#virtengine.cert.v1.MsgRevokeCertificate)
     - [MsgRevokeCertificateResponse](#virtengine.cert.v1.MsgRevokeCertificateResponse)
   
 - [virtengine/cert/v1/query.proto](#virtengine/cert/v1/query.proto)
     - [CertificateResponse](#virtengine.cert.v1.CertificateResponse)
     - [QueryCertificatesRequest](#virtengine.cert.v1.QueryCertificatesRequest)
     - [QueryCertificatesResponse](#virtengine.cert.v1.QueryCertificatesResponse)
   
     - [Query](#virtengine.cert.v1.Query)
   
 - [virtengine/cert/v1/service.proto](#virtengine/cert/v1/service.proto)
     - [Msg](#virtengine.cert.v1.Msg)
   
 - [virtengine/config/v1/tx.proto](#virtengine/config/v1/tx.proto)
     - [ApprovedClient](#virtengine.config.v1.ApprovedClient)
     - [EventClientReactivated](#virtengine.config.v1.EventClientReactivated)
     - [EventClientRegistered](#virtengine.config.v1.EventClientRegistered)
     - [EventClientRevoked](#virtengine.config.v1.EventClientRevoked)
     - [EventClientSuspended](#virtengine.config.v1.EventClientSuspended)
     - [EventClientUpdated](#virtengine.config.v1.EventClientUpdated)
     - [EventSignatureVerified](#virtengine.config.v1.EventSignatureVerified)
     - [GenesisState](#virtengine.config.v1.GenesisState)
     - [MsgReactivateApprovedClient](#virtengine.config.v1.MsgReactivateApprovedClient)
     - [MsgReactivateApprovedClientResponse](#virtengine.config.v1.MsgReactivateApprovedClientResponse)
     - [MsgRegisterApprovedClient](#virtengine.config.v1.MsgRegisterApprovedClient)
     - [MsgRegisterApprovedClientResponse](#virtengine.config.v1.MsgRegisterApprovedClientResponse)
     - [MsgRevokeApprovedClient](#virtengine.config.v1.MsgRevokeApprovedClient)
     - [MsgRevokeApprovedClientResponse](#virtengine.config.v1.MsgRevokeApprovedClientResponse)
     - [MsgSuspendApprovedClient](#virtengine.config.v1.MsgSuspendApprovedClient)
     - [MsgSuspendApprovedClientResponse](#virtengine.config.v1.MsgSuspendApprovedClientResponse)
     - [MsgUpdateApprovedClient](#virtengine.config.v1.MsgUpdateApprovedClient)
     - [MsgUpdateApprovedClientResponse](#virtengine.config.v1.MsgUpdateApprovedClientResponse)
     - [MsgUpdateParams](#virtengine.config.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.config.v1.MsgUpdateParamsResponse)
     - [Params](#virtengine.config.v1.Params)
   
     - [Msg](#virtengine.config.v1.Msg)
   
 - [virtengine/delegation/v1/types.proto](#virtengine/delegation/v1/types.proto)
     - [Delegation](#virtengine.delegation.v1.Delegation)
     - [DelegatorReward](#virtengine.delegation.v1.DelegatorReward)
     - [Params](#virtengine.delegation.v1.Params)
     - [Redelegation](#virtengine.delegation.v1.Redelegation)
     - [RedelegationEntry](#virtengine.delegation.v1.RedelegationEntry)
     - [UnbondingDelegation](#virtengine.delegation.v1.UnbondingDelegation)
     - [UnbondingDelegationEntry](#virtengine.delegation.v1.UnbondingDelegationEntry)
     - [ValidatorShares](#virtengine.delegation.v1.ValidatorShares)
   
     - [DelegationStatus](#virtengine.delegation.v1.DelegationStatus)
   
 - [virtengine/delegation/v1/genesis.proto](#virtengine/delegation/v1/genesis.proto)
     - [GenesisState](#virtengine.delegation.v1.GenesisState)
   
 - [virtengine/delegation/v1/query.proto](#virtengine/delegation/v1/query.proto)
     - [QueryDelegationRequest](#virtengine.delegation.v1.QueryDelegationRequest)
     - [QueryDelegationResponse](#virtengine.delegation.v1.QueryDelegationResponse)
     - [QueryDelegatorAllRewardsRequest](#virtengine.delegation.v1.QueryDelegatorAllRewardsRequest)
     - [QueryDelegatorAllRewardsResponse](#virtengine.delegation.v1.QueryDelegatorAllRewardsResponse)
     - [QueryDelegatorDelegationsRequest](#virtengine.delegation.v1.QueryDelegatorDelegationsRequest)
     - [QueryDelegatorDelegationsResponse](#virtengine.delegation.v1.QueryDelegatorDelegationsResponse)
     - [QueryDelegatorRedelegationsRequest](#virtengine.delegation.v1.QueryDelegatorRedelegationsRequest)
     - [QueryDelegatorRedelegationsResponse](#virtengine.delegation.v1.QueryDelegatorRedelegationsResponse)
     - [QueryDelegatorRewardsRequest](#virtengine.delegation.v1.QueryDelegatorRewardsRequest)
     - [QueryDelegatorRewardsResponse](#virtengine.delegation.v1.QueryDelegatorRewardsResponse)
     - [QueryDelegatorUnbondingDelegationsRequest](#virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest)
     - [QueryDelegatorUnbondingDelegationsResponse](#virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse)
     - [QueryParamsRequest](#virtengine.delegation.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.delegation.v1.QueryParamsResponse)
     - [QueryRedelegationRequest](#virtengine.delegation.v1.QueryRedelegationRequest)
     - [QueryRedelegationResponse](#virtengine.delegation.v1.QueryRedelegationResponse)
     - [QueryUnbondingDelegationRequest](#virtengine.delegation.v1.QueryUnbondingDelegationRequest)
     - [QueryUnbondingDelegationResponse](#virtengine.delegation.v1.QueryUnbondingDelegationResponse)
     - [QueryValidatorDelegationsRequest](#virtengine.delegation.v1.QueryValidatorDelegationsRequest)
     - [QueryValidatorDelegationsResponse](#virtengine.delegation.v1.QueryValidatorDelegationsResponse)
     - [QueryValidatorSharesRequest](#virtengine.delegation.v1.QueryValidatorSharesRequest)
     - [QueryValidatorSharesResponse](#virtengine.delegation.v1.QueryValidatorSharesResponse)
   
     - [Query](#virtengine.delegation.v1.Query)
   
 - [virtengine/delegation/v1/tx.proto](#virtengine/delegation/v1/tx.proto)
     - [MsgClaimAllRewards](#virtengine.delegation.v1.MsgClaimAllRewards)
     - [MsgClaimAllRewardsResponse](#virtengine.delegation.v1.MsgClaimAllRewardsResponse)
     - [MsgClaimRewards](#virtengine.delegation.v1.MsgClaimRewards)
     - [MsgClaimRewardsResponse](#virtengine.delegation.v1.MsgClaimRewardsResponse)
     - [MsgDelegate](#virtengine.delegation.v1.MsgDelegate)
     - [MsgDelegateResponse](#virtengine.delegation.v1.MsgDelegateResponse)
     - [MsgRedelegate](#virtengine.delegation.v1.MsgRedelegate)
     - [MsgRedelegateResponse](#virtengine.delegation.v1.MsgRedelegateResponse)
     - [MsgUndelegate](#virtengine.delegation.v1.MsgUndelegate)
     - [MsgUndelegateResponse](#virtengine.delegation.v1.MsgUndelegateResponse)
     - [MsgUpdateParams](#virtengine.delegation.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.delegation.v1.MsgUpdateParamsResponse)
   
     - [Msg](#virtengine.delegation.v1.Msg)
   
 - [virtengine/deployment/v1/deployment.proto](#virtengine/deployment/v1/deployment.proto)
     - [Deployment](#virtengine.deployment.v1.Deployment)
     - [DeploymentID](#virtengine.deployment.v1.DeploymentID)
   
     - [Deployment.State](#virtengine.deployment.v1.Deployment.State)
   
 - [virtengine/deployment/v1/group.proto](#virtengine/deployment/v1/group.proto)
     - [GroupID](#virtengine.deployment.v1.GroupID)
   
 - [virtengine/deployment/v1/event.proto](#virtengine/deployment/v1/event.proto)
     - [EventDeploymentClosed](#virtengine.deployment.v1.EventDeploymentClosed)
     - [EventDeploymentCreated](#virtengine.deployment.v1.EventDeploymentCreated)
     - [EventDeploymentUpdated](#virtengine.deployment.v1.EventDeploymentUpdated)
     - [EventGroupClosed](#virtengine.deployment.v1.EventGroupClosed)
     - [EventGroupPaused](#virtengine.deployment.v1.EventGroupPaused)
     - [EventGroupStarted](#virtengine.deployment.v1.EventGroupStarted)
   
 - [virtengine/deployment/v1beta4/resourceunit.proto](#virtengine/deployment/v1beta4/resourceunit.proto)
     - [ResourceUnit](#virtengine.deployment.v1beta4.ResourceUnit)
   
 - [virtengine/deployment/v1beta4/groupspec.proto](#virtengine/deployment/v1beta4/groupspec.proto)
     - [GroupSpec](#virtengine.deployment.v1beta4.GroupSpec)
   
 - [virtengine/deployment/v1beta4/deploymentmsg.proto](#virtengine/deployment/v1beta4/deploymentmsg.proto)
     - [MsgCloseDeployment](#virtengine.deployment.v1beta4.MsgCloseDeployment)
     - [MsgCloseDeploymentResponse](#virtengine.deployment.v1beta4.MsgCloseDeploymentResponse)
     - [MsgCreateDeployment](#virtengine.deployment.v1beta4.MsgCreateDeployment)
     - [MsgCreateDeploymentResponse](#virtengine.deployment.v1beta4.MsgCreateDeploymentResponse)
     - [MsgUpdateDeployment](#virtengine.deployment.v1beta4.MsgUpdateDeployment)
     - [MsgUpdateDeploymentResponse](#virtengine.deployment.v1beta4.MsgUpdateDeploymentResponse)
   
 - [virtengine/deployment/v1beta4/filters.proto](#virtengine/deployment/v1beta4/filters.proto)
     - [DeploymentFilters](#virtengine.deployment.v1beta4.DeploymentFilters)
     - [GroupFilters](#virtengine.deployment.v1beta4.GroupFilters)
   
 - [virtengine/deployment/v1beta4/group.proto](#virtengine/deployment/v1beta4/group.proto)
     - [Group](#virtengine.deployment.v1beta4.Group)
   
     - [Group.State](#virtengine.deployment.v1beta4.Group.State)
   
 - [virtengine/deployment/v1beta4/params.proto](#virtengine/deployment/v1beta4/params.proto)
     - [Params](#virtengine.deployment.v1beta4.Params)
   
 - [virtengine/deployment/v1beta4/genesis.proto](#virtengine/deployment/v1beta4/genesis.proto)
     - [GenesisDeployment](#virtengine.deployment.v1beta4.GenesisDeployment)
     - [GenesisState](#virtengine.deployment.v1beta4.GenesisState)
   
 - [virtengine/deployment/v1beta4/groupmsg.proto](#virtengine/deployment/v1beta4/groupmsg.proto)
     - [MsgCloseGroup](#virtengine.deployment.v1beta4.MsgCloseGroup)
     - [MsgCloseGroupResponse](#virtengine.deployment.v1beta4.MsgCloseGroupResponse)
     - [MsgPauseGroup](#virtengine.deployment.v1beta4.MsgPauseGroup)
     - [MsgPauseGroupResponse](#virtengine.deployment.v1beta4.MsgPauseGroupResponse)
     - [MsgStartGroup](#virtengine.deployment.v1beta4.MsgStartGroup)
     - [MsgStartGroupResponse](#virtengine.deployment.v1beta4.MsgStartGroupResponse)
   
 - [virtengine/deployment/v1beta4/paramsmsg.proto](#virtengine/deployment/v1beta4/paramsmsg.proto)
     - [MsgUpdateParams](#virtengine.deployment.v1beta4.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.deployment.v1beta4.MsgUpdateParamsResponse)
   
 - [virtengine/escrow/id/v1/id.proto](#virtengine/escrow/id/v1/id.proto)
     - [Account](#virtengine.escrow.id.v1.Account)
     - [Payment](#virtengine.escrow.id.v1.Payment)
   
     - [Scope](#virtengine.escrow.id.v1.Scope)
   
 - [virtengine/escrow/types/v1/balance.proto](#virtengine/escrow/types/v1/balance.proto)
     - [Balance](#virtengine.escrow.types.v1.Balance)
   
 - [virtengine/escrow/types/v1/deposit.proto](#virtengine/escrow/types/v1/deposit.proto)
     - [Depositor](#virtengine.escrow.types.v1.Depositor)
   
 - [virtengine/escrow/types/v1/state.proto](#virtengine/escrow/types/v1/state.proto)
     - [State](#virtengine.escrow.types.v1.State)
   
 - [virtengine/escrow/types/v1/account.proto](#virtengine/escrow/types/v1/account.proto)
     - [Account](#virtengine.escrow.types.v1.Account)
     - [AccountState](#virtengine.escrow.types.v1.AccountState)
   
 - [virtengine/deployment/v1beta4/query.proto](#virtengine/deployment/v1beta4/query.proto)
     - [QueryDeploymentRequest](#virtengine.deployment.v1beta4.QueryDeploymentRequest)
     - [QueryDeploymentResponse](#virtengine.deployment.v1beta4.QueryDeploymentResponse)
     - [QueryDeploymentsRequest](#virtengine.deployment.v1beta4.QueryDeploymentsRequest)
     - [QueryDeploymentsResponse](#virtengine.deployment.v1beta4.QueryDeploymentsResponse)
     - [QueryGroupRequest](#virtengine.deployment.v1beta4.QueryGroupRequest)
     - [QueryGroupResponse](#virtengine.deployment.v1beta4.QueryGroupResponse)
     - [QueryParamsRequest](#virtengine.deployment.v1beta4.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.deployment.v1beta4.QueryParamsResponse)
   
     - [Query](#virtengine.deployment.v1beta4.Query)
   
 - [virtengine/deployment/v1beta4/service.proto](#virtengine/deployment/v1beta4/service.proto)
     - [Msg](#virtengine.deployment.v1beta4.Msg)
   
 - [virtengine/discovery/v1/client_info.proto](#virtengine/discovery/v1/client_info.proto)
     - [ClientInfo](#virtengine.discovery.v1.ClientInfo)
   
 - [virtengine/discovery/v1/virtengine.proto](#virtengine/discovery/v1/virtengine.proto)
     - [VirtEngine](#virtengine.discovery.v1.VirtEngine)
   
 - [virtengine/downtimedetector/v1beta1/downtime_duration.proto](#virtengine/downtimedetector/v1beta1/downtime_duration.proto)
     - [Downtime](#virtengine.downtimedetector.v1beta1.Downtime)
   
 - [virtengine/downtimedetector/v1beta1/genesis.proto](#virtengine/downtimedetector/v1beta1/genesis.proto)
     - [GenesisDowntimeEntry](#virtengine.downtimedetector.v1beta1.GenesisDowntimeEntry)
     - [GenesisState](#virtengine.downtimedetector.v1beta1.GenesisState)
   
 - [virtengine/downtimedetector/v1beta1/query.proto](#virtengine/downtimedetector/v1beta1/query.proto)
     - [RecoveredSinceDowntimeOfLengthRequest](#virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest)
     - [RecoveredSinceDowntimeOfLengthResponse](#virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse)
   
     - [Query](#virtengine.downtimedetector.v1beta1.Query)
   
 - [virtengine/enclave/v1/events.proto](#virtengine/enclave/v1/events.proto)
     - [EventConsensusVerificationFailed](#virtengine.enclave.v1.EventConsensusVerificationFailed)
     - [EventEnclaveIdentityExpired](#virtengine.enclave.v1.EventEnclaveIdentityExpired)
     - [EventEnclaveIdentityRegistered](#virtengine.enclave.v1.EventEnclaveIdentityRegistered)
     - [EventEnclaveIdentityRevoked](#virtengine.enclave.v1.EventEnclaveIdentityRevoked)
     - [EventEnclaveIdentityUpdated](#virtengine.enclave.v1.EventEnclaveIdentityUpdated)
     - [EventEnclaveKeyRotated](#virtengine.enclave.v1.EventEnclaveKeyRotated)
     - [EventKeyRotationCompleted](#virtengine.enclave.v1.EventKeyRotationCompleted)
     - [EventMeasurementAdded](#virtengine.enclave.v1.EventMeasurementAdded)
     - [EventMeasurementRevoked](#virtengine.enclave.v1.EventMeasurementRevoked)
     - [EventVEIDScoreComputedAttested](#virtengine.enclave.v1.EventVEIDScoreComputedAttested)
     - [EventVEIDScoreRejectedAttestation](#virtengine.enclave.v1.EventVEIDScoreRejectedAttestation)
   
 - [virtengine/enclave/v1/types.proto](#virtengine/enclave/v1/types.proto)
     - [AttestedScoringResult](#virtengine.enclave.v1.AttestedScoringResult)
     - [EnclaveIdentity](#virtengine.enclave.v1.EnclaveIdentity)
     - [KeyRotationRecord](#virtengine.enclave.v1.KeyRotationRecord)
     - [MeasurementRecord](#virtengine.enclave.v1.MeasurementRecord)
     - [Params](#virtengine.enclave.v1.Params)
     - [ValidatorKeyInfo](#virtengine.enclave.v1.ValidatorKeyInfo)
   
     - [EnclaveIdentityStatus](#virtengine.enclave.v1.EnclaveIdentityStatus)
     - [KeyRotationStatus](#virtengine.enclave.v1.KeyRotationStatus)
     - [TEEType](#virtengine.enclave.v1.TEEType)
   
 - [virtengine/enclave/v1/genesis.proto](#virtengine/enclave/v1/genesis.proto)
     - [GenesisState](#virtengine.enclave.v1.GenesisState)
   
 - [virtengine/enclave/v1/query.proto](#virtengine/enclave/v1/query.proto)
     - [QueryActiveValidatorEnclaveKeysRequest](#virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysRequest)
     - [QueryActiveValidatorEnclaveKeysResponse](#virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysResponse)
     - [QueryAttestedResultRequest](#virtengine.enclave.v1.QueryAttestedResultRequest)
     - [QueryAttestedResultResponse](#virtengine.enclave.v1.QueryAttestedResultResponse)
     - [QueryCommitteeEnclaveKeysRequest](#virtengine.enclave.v1.QueryCommitteeEnclaveKeysRequest)
     - [QueryCommitteeEnclaveKeysResponse](#virtengine.enclave.v1.QueryCommitteeEnclaveKeysResponse)
     - [QueryEnclaveIdentityRequest](#virtengine.enclave.v1.QueryEnclaveIdentityRequest)
     - [QueryEnclaveIdentityResponse](#virtengine.enclave.v1.QueryEnclaveIdentityResponse)
     - [QueryKeyRotationRequest](#virtengine.enclave.v1.QueryKeyRotationRequest)
     - [QueryKeyRotationResponse](#virtengine.enclave.v1.QueryKeyRotationResponse)
     - [QueryMeasurementAllowlistRequest](#virtengine.enclave.v1.QueryMeasurementAllowlistRequest)
     - [QueryMeasurementAllowlistResponse](#virtengine.enclave.v1.QueryMeasurementAllowlistResponse)
     - [QueryMeasurementRequest](#virtengine.enclave.v1.QueryMeasurementRequest)
     - [QueryMeasurementResponse](#virtengine.enclave.v1.QueryMeasurementResponse)
     - [QueryParamsRequest](#virtengine.enclave.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.enclave.v1.QueryParamsResponse)
     - [QueryValidKeySetRequest](#virtengine.enclave.v1.QueryValidKeySetRequest)
     - [QueryValidKeySetResponse](#virtengine.enclave.v1.QueryValidKeySetResponse)
   
     - [Query](#virtengine.enclave.v1.Query)
   
 - [virtengine/enclave/v1/tx.proto](#virtengine/enclave/v1/tx.proto)
     - [MsgProposeMeasurement](#virtengine.enclave.v1.MsgProposeMeasurement)
     - [MsgProposeMeasurementResponse](#virtengine.enclave.v1.MsgProposeMeasurementResponse)
     - [MsgRegisterEnclaveIdentity](#virtengine.enclave.v1.MsgRegisterEnclaveIdentity)
     - [MsgRegisterEnclaveIdentityResponse](#virtengine.enclave.v1.MsgRegisterEnclaveIdentityResponse)
     - [MsgRevokeMeasurement](#virtengine.enclave.v1.MsgRevokeMeasurement)
     - [MsgRevokeMeasurementResponse](#virtengine.enclave.v1.MsgRevokeMeasurementResponse)
     - [MsgRotateEnclaveIdentity](#virtengine.enclave.v1.MsgRotateEnclaveIdentity)
     - [MsgRotateEnclaveIdentityResponse](#virtengine.enclave.v1.MsgRotateEnclaveIdentityResponse)
     - [MsgUpdateParams](#virtengine.enclave.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.enclave.v1.MsgUpdateParamsResponse)
   
     - [Msg](#virtengine.enclave.v1.Msg)
   
 - [virtengine/encryption/v1/types.proto](#virtengine/encryption/v1/types.proto)
     - [AlgorithmInfo](#virtengine.encryption.v1.AlgorithmInfo)
     - [EncryptedPayloadEnvelope](#virtengine.encryption.v1.EncryptedPayloadEnvelope)
     - [EncryptedPayloadEnvelope.MetadataEntry](#virtengine.encryption.v1.EncryptedPayloadEnvelope.MetadataEntry)
     - [EventKeyRegistered](#virtengine.encryption.v1.EventKeyRegistered)
     - [EventKeyRevoked](#virtengine.encryption.v1.EventKeyRevoked)
     - [EventKeyUpdated](#virtengine.encryption.v1.EventKeyUpdated)
     - [MultiRecipientEnvelope](#virtengine.encryption.v1.MultiRecipientEnvelope)
     - [MultiRecipientEnvelope.MetadataEntry](#virtengine.encryption.v1.MultiRecipientEnvelope.MetadataEntry)
     - [Params](#virtengine.encryption.v1.Params)
     - [RecipientKeyRecord](#virtengine.encryption.v1.RecipientKeyRecord)
     - [WrappedKeyEntry](#virtengine.encryption.v1.WrappedKeyEntry)
   
     - [RecipientMode](#virtengine.encryption.v1.RecipientMode)
   
 - [virtengine/encryption/v1/genesis.proto](#virtengine/encryption/v1/genesis.proto)
     - [GenesisState](#virtengine.encryption.v1.GenesisState)
   
 - [virtengine/encryption/v1/query.proto](#virtengine/encryption/v1/query.proto)
     - [QueryAlgorithmsRequest](#virtengine.encryption.v1.QueryAlgorithmsRequest)
     - [QueryAlgorithmsResponse](#virtengine.encryption.v1.QueryAlgorithmsResponse)
     - [QueryKeyByFingerprintRequest](#virtengine.encryption.v1.QueryKeyByFingerprintRequest)
     - [QueryKeyByFingerprintResponse](#virtengine.encryption.v1.QueryKeyByFingerprintResponse)
     - [QueryParamsRequest](#virtengine.encryption.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.encryption.v1.QueryParamsResponse)
     - [QueryRecipientKeyRequest](#virtengine.encryption.v1.QueryRecipientKeyRequest)
     - [QueryRecipientKeyResponse](#virtengine.encryption.v1.QueryRecipientKeyResponse)
     - [QueryValidateEnvelopeRequest](#virtengine.encryption.v1.QueryValidateEnvelopeRequest)
     - [QueryValidateEnvelopeResponse](#virtengine.encryption.v1.QueryValidateEnvelopeResponse)
   
     - [Query](#virtengine.encryption.v1.Query)
   
 - [virtengine/encryption/v1/tx.proto](#virtengine/encryption/v1/tx.proto)
     - [MsgRegisterRecipientKey](#virtengine.encryption.v1.MsgRegisterRecipientKey)
     - [MsgRegisterRecipientKeyResponse](#virtengine.encryption.v1.MsgRegisterRecipientKeyResponse)
     - [MsgRevokeRecipientKey](#virtengine.encryption.v1.MsgRevokeRecipientKey)
     - [MsgRevokeRecipientKeyResponse](#virtengine.encryption.v1.MsgRevokeRecipientKeyResponse)
     - [MsgUpdateKeyLabel](#virtengine.encryption.v1.MsgUpdateKeyLabel)
     - [MsgUpdateKeyLabelResponse](#virtengine.encryption.v1.MsgUpdateKeyLabelResponse)
   
     - [Msg](#virtengine.encryption.v1.Msg)
   
 - [virtengine/epochs/v1beta1/events.proto](#virtengine/epochs/v1beta1/events.proto)
     - [EventEpochEnd](#virtengine.epochs.v1beta1.EventEpochEnd)
     - [EventEpochStart](#virtengine.epochs.v1beta1.EventEpochStart)
   
 - [virtengine/epochs/v1beta1/genesis.proto](#virtengine/epochs/v1beta1/genesis.proto)
     - [EpochInfo](#virtengine.epochs.v1beta1.EpochInfo)
     - [GenesisState](#virtengine.epochs.v1beta1.GenesisState)
   
 - [virtengine/epochs/v1beta1/query.proto](#virtengine/epochs/v1beta1/query.proto)
     - [QueryCurrentEpochRequest](#virtengine.epochs.v1beta1.QueryCurrentEpochRequest)
     - [QueryCurrentEpochResponse](#virtengine.epochs.v1beta1.QueryCurrentEpochResponse)
     - [QueryEpochInfosRequest](#virtengine.epochs.v1beta1.QueryEpochInfosRequest)
     - [QueryEpochInfosResponse](#virtengine.epochs.v1beta1.QueryEpochInfosResponse)
   
     - [Query](#virtengine.epochs.v1beta1.Query)
   
 - [virtengine/escrow/types/v1/payment.proto](#virtengine/escrow/types/v1/payment.proto)
     - [Payment](#virtengine.escrow.types.v1.Payment)
     - [PaymentState](#virtengine.escrow.types.v1.PaymentState)
   
 - [virtengine/escrow/v1/authz.proto](#virtengine/escrow/v1/authz.proto)
     - [DepositAuthorization](#virtengine.escrow.v1.DepositAuthorization)
   
     - [DepositAuthorization.Scope](#virtengine.escrow.v1.DepositAuthorization.Scope)
   
 - [virtengine/escrow/v1/genesis.proto](#virtengine/escrow/v1/genesis.proto)
     - [GenesisState](#virtengine.escrow.v1.GenesisState)
   
 - [virtengine/escrow/v1/msg.proto](#virtengine/escrow/v1/msg.proto)
     - [MsgAccountDeposit](#virtengine.escrow.v1.MsgAccountDeposit)
     - [MsgAccountDepositResponse](#virtengine.escrow.v1.MsgAccountDepositResponse)
   
 - [virtengine/escrow/v1/query.proto](#virtengine/escrow/v1/query.proto)
     - [QueryAccountsRequest](#virtengine.escrow.v1.QueryAccountsRequest)
     - [QueryAccountsResponse](#virtengine.escrow.v1.QueryAccountsResponse)
     - [QueryPaymentsRequest](#virtengine.escrow.v1.QueryPaymentsRequest)
     - [QueryPaymentsResponse](#virtengine.escrow.v1.QueryPaymentsResponse)
   
     - [Query](#virtengine.escrow.v1.Query)
   
 - [virtengine/escrow/v1/service.proto](#virtengine/escrow/v1/service.proto)
     - [Msg](#virtengine.escrow.v1.Msg)
   
 - [virtengine/fraud/v1/types.proto](#virtengine/fraud/v1/types.proto)
     - [EncryptedEvidence](#virtengine.fraud.v1.EncryptedEvidence)
     - [FraudAuditLog](#virtengine.fraud.v1.FraudAuditLog)
     - [FraudReport](#virtengine.fraud.v1.FraudReport)
     - [ModeratorQueueEntry](#virtengine.fraud.v1.ModeratorQueueEntry)
   
     - [AuditAction](#virtengine.fraud.v1.AuditAction)
     - [FraudCategory](#virtengine.fraud.v1.FraudCategory)
     - [FraudReportStatus](#virtengine.fraud.v1.FraudReportStatus)
     - [ResolutionType](#virtengine.fraud.v1.ResolutionType)
   
 - [virtengine/fraud/v1/params.proto](#virtengine/fraud/v1/params.proto)
     - [Params](#virtengine.fraud.v1.Params)
   
 - [virtengine/fraud/v1/genesis.proto](#virtengine/fraud/v1/genesis.proto)
     - [GenesisState](#virtengine.fraud.v1.GenesisState)
   
 - [virtengine/fraud/v1/query.proto](#virtengine/fraud/v1/query.proto)
     - [QueryAuditLogRequest](#virtengine.fraud.v1.QueryAuditLogRequest)
     - [QueryAuditLogResponse](#virtengine.fraud.v1.QueryAuditLogResponse)
     - [QueryFraudReportRequest](#virtengine.fraud.v1.QueryFraudReportRequest)
     - [QueryFraudReportResponse](#virtengine.fraud.v1.QueryFraudReportResponse)
     - [QueryFraudReportsByReportedPartyRequest](#virtengine.fraud.v1.QueryFraudReportsByReportedPartyRequest)
     - [QueryFraudReportsByReportedPartyResponse](#virtengine.fraud.v1.QueryFraudReportsByReportedPartyResponse)
     - [QueryFraudReportsByReporterRequest](#virtengine.fraud.v1.QueryFraudReportsByReporterRequest)
     - [QueryFraudReportsByReporterResponse](#virtengine.fraud.v1.QueryFraudReportsByReporterResponse)
     - [QueryFraudReportsRequest](#virtengine.fraud.v1.QueryFraudReportsRequest)
     - [QueryFraudReportsResponse](#virtengine.fraud.v1.QueryFraudReportsResponse)
     - [QueryModeratorQueueRequest](#virtengine.fraud.v1.QueryModeratorQueueRequest)
     - [QueryModeratorQueueResponse](#virtengine.fraud.v1.QueryModeratorQueueResponse)
     - [QueryParamsRequest](#virtengine.fraud.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.fraud.v1.QueryParamsResponse)
   
     - [Query](#virtengine.fraud.v1.Query)
   
 - [virtengine/fraud/v1/tx.proto](#virtengine/fraud/v1/tx.proto)
     - [MsgAssignModerator](#virtengine.fraud.v1.MsgAssignModerator)
     - [MsgAssignModeratorResponse](#virtengine.fraud.v1.MsgAssignModeratorResponse)
     - [MsgEscalateFraudReport](#virtengine.fraud.v1.MsgEscalateFraudReport)
     - [MsgEscalateFraudReportResponse](#virtengine.fraud.v1.MsgEscalateFraudReportResponse)
     - [MsgRejectFraudReport](#virtengine.fraud.v1.MsgRejectFraudReport)
     - [MsgRejectFraudReportResponse](#virtengine.fraud.v1.MsgRejectFraudReportResponse)
     - [MsgResolveFraudReport](#virtengine.fraud.v1.MsgResolveFraudReport)
     - [MsgResolveFraudReportResponse](#virtengine.fraud.v1.MsgResolveFraudReportResponse)
     - [MsgSubmitFraudReport](#virtengine.fraud.v1.MsgSubmitFraudReport)
     - [MsgSubmitFraudReportResponse](#virtengine.fraud.v1.MsgSubmitFraudReportResponse)
     - [MsgUpdateParams](#virtengine.fraud.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.fraud.v1.MsgUpdateParamsResponse)
     - [MsgUpdateReportStatus](#virtengine.fraud.v1.MsgUpdateReportStatus)
     - [MsgUpdateReportStatusResponse](#virtengine.fraud.v1.MsgUpdateReportStatusResponse)
   
     - [Msg](#virtengine.fraud.v1.Msg)
   
 - [virtengine/hpc/v1/types.proto](#virtengine/hpc/v1/types.proto)
     - [ClusterCandidate](#virtengine.hpc.v1.ClusterCandidate)
     - [ClusterMetadata](#virtengine.hpc.v1.ClusterMetadata)
     - [DataReference](#virtengine.hpc.v1.DataReference)
     - [HPCCluster](#virtengine.hpc.v1.HPCCluster)
     - [HPCDispute](#virtengine.hpc.v1.HPCDispute)
     - [HPCJob](#virtengine.hpc.v1.HPCJob)
     - [HPCOffering](#virtengine.hpc.v1.HPCOffering)
     - [HPCPricing](#virtengine.hpc.v1.HPCPricing)
     - [HPCRewardRecipient](#virtengine.hpc.v1.HPCRewardRecipient)
     - [HPCRewardRecord](#virtengine.hpc.v1.HPCRewardRecord)
     - [HPCUsageMetrics](#virtengine.hpc.v1.HPCUsageMetrics)
     - [JobAccounting](#virtengine.hpc.v1.JobAccounting)
     - [JobResources](#virtengine.hpc.v1.JobResources)
     - [JobWorkloadSpec](#virtengine.hpc.v1.JobWorkloadSpec)
     - [JobWorkloadSpec.EnvironmentEntry](#virtengine.hpc.v1.JobWorkloadSpec.EnvironmentEntry)
     - [LatencyMeasurement](#virtengine.hpc.v1.LatencyMeasurement)
     - [NodeMetadata](#virtengine.hpc.v1.NodeMetadata)
     - [NodeResources](#virtengine.hpc.v1.NodeResources)
     - [NodeReward](#virtengine.hpc.v1.NodeReward)
     - [Params](#virtengine.hpc.v1.Params)
     - [Partition](#virtengine.hpc.v1.Partition)
     - [PreconfiguredWorkload](#virtengine.hpc.v1.PreconfiguredWorkload)
     - [QueueOption](#virtengine.hpc.v1.QueueOption)
     - [RewardCalculationDetails](#virtengine.hpc.v1.RewardCalculationDetails)
     - [RewardCalculationDetails.InputMetricsEntry](#virtengine.hpc.v1.RewardCalculationDetails.InputMetricsEntry)
     - [SchedulingDecision](#virtengine.hpc.v1.SchedulingDecision)
   
     - [ClusterState](#virtengine.hpc.v1.ClusterState)
     - [DisputeStatus](#virtengine.hpc.v1.DisputeStatus)
     - [HPCRewardSource](#virtengine.hpc.v1.HPCRewardSource)
     - [JobState](#virtengine.hpc.v1.JobState)
   
 - [virtengine/hpc/v1/genesis.proto](#virtengine/hpc/v1/genesis.proto)
     - [GenesisState](#virtengine.hpc.v1.GenesisState)
   
 - [virtengine/hpc/v1/query.proto](#virtengine/hpc/v1/query.proto)
     - [QueryClusterRequest](#virtengine.hpc.v1.QueryClusterRequest)
     - [QueryClusterResponse](#virtengine.hpc.v1.QueryClusterResponse)
     - [QueryClustersByProviderRequest](#virtengine.hpc.v1.QueryClustersByProviderRequest)
     - [QueryClustersByProviderResponse](#virtengine.hpc.v1.QueryClustersByProviderResponse)
     - [QueryClustersRequest](#virtengine.hpc.v1.QueryClustersRequest)
     - [QueryClustersResponse](#virtengine.hpc.v1.QueryClustersResponse)
     - [QueryDisputeRequest](#virtengine.hpc.v1.QueryDisputeRequest)
     - [QueryDisputeResponse](#virtengine.hpc.v1.QueryDisputeResponse)
     - [QueryDisputesRequest](#virtengine.hpc.v1.QueryDisputesRequest)
     - [QueryDisputesResponse](#virtengine.hpc.v1.QueryDisputesResponse)
     - [QueryJobAccountingRequest](#virtengine.hpc.v1.QueryJobAccountingRequest)
     - [QueryJobAccountingResponse](#virtengine.hpc.v1.QueryJobAccountingResponse)
     - [QueryJobRequest](#virtengine.hpc.v1.QueryJobRequest)
     - [QueryJobResponse](#virtengine.hpc.v1.QueryJobResponse)
     - [QueryJobsByCustomerRequest](#virtengine.hpc.v1.QueryJobsByCustomerRequest)
     - [QueryJobsByCustomerResponse](#virtengine.hpc.v1.QueryJobsByCustomerResponse)
     - [QueryJobsByProviderRequest](#virtengine.hpc.v1.QueryJobsByProviderRequest)
     - [QueryJobsByProviderResponse](#virtengine.hpc.v1.QueryJobsByProviderResponse)
     - [QueryJobsRequest](#virtengine.hpc.v1.QueryJobsRequest)
     - [QueryJobsResponse](#virtengine.hpc.v1.QueryJobsResponse)
     - [QueryNodeMetadataRequest](#virtengine.hpc.v1.QueryNodeMetadataRequest)
     - [QueryNodeMetadataResponse](#virtengine.hpc.v1.QueryNodeMetadataResponse)
     - [QueryNodesByClusterRequest](#virtengine.hpc.v1.QueryNodesByClusterRequest)
     - [QueryNodesByClusterResponse](#virtengine.hpc.v1.QueryNodesByClusterResponse)
     - [QueryOfferingRequest](#virtengine.hpc.v1.QueryOfferingRequest)
     - [QueryOfferingResponse](#virtengine.hpc.v1.QueryOfferingResponse)
     - [QueryOfferingsByClusterRequest](#virtengine.hpc.v1.QueryOfferingsByClusterRequest)
     - [QueryOfferingsByClusterResponse](#virtengine.hpc.v1.QueryOfferingsByClusterResponse)
     - [QueryOfferingsRequest](#virtengine.hpc.v1.QueryOfferingsRequest)
     - [QueryOfferingsResponse](#virtengine.hpc.v1.QueryOfferingsResponse)
     - [QueryParamsRequest](#virtengine.hpc.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.hpc.v1.QueryParamsResponse)
     - [QueryRewardRequest](#virtengine.hpc.v1.QueryRewardRequest)
     - [QueryRewardResponse](#virtengine.hpc.v1.QueryRewardResponse)
     - [QueryRewardsByJobRequest](#virtengine.hpc.v1.QueryRewardsByJobRequest)
     - [QueryRewardsByJobResponse](#virtengine.hpc.v1.QueryRewardsByJobResponse)
     - [QuerySchedulingDecisionByJobRequest](#virtengine.hpc.v1.QuerySchedulingDecisionByJobRequest)
     - [QuerySchedulingDecisionByJobResponse](#virtengine.hpc.v1.QuerySchedulingDecisionByJobResponse)
     - [QuerySchedulingDecisionRequest](#virtengine.hpc.v1.QuerySchedulingDecisionRequest)
     - [QuerySchedulingDecisionResponse](#virtengine.hpc.v1.QuerySchedulingDecisionResponse)
   
     - [Query](#virtengine.hpc.v1.Query)
   
 - [virtengine/hpc/v1/tx.proto](#virtengine/hpc/v1/tx.proto)
     - [MsgCancelJob](#virtengine.hpc.v1.MsgCancelJob)
     - [MsgCancelJobResponse](#virtengine.hpc.v1.MsgCancelJobResponse)
     - [MsgCreateOffering](#virtengine.hpc.v1.MsgCreateOffering)
     - [MsgCreateOfferingResponse](#virtengine.hpc.v1.MsgCreateOfferingResponse)
     - [MsgDeregisterCluster](#virtengine.hpc.v1.MsgDeregisterCluster)
     - [MsgDeregisterClusterResponse](#virtengine.hpc.v1.MsgDeregisterClusterResponse)
     - [MsgFlagDispute](#virtengine.hpc.v1.MsgFlagDispute)
     - [MsgFlagDisputeResponse](#virtengine.hpc.v1.MsgFlagDisputeResponse)
     - [MsgRegisterCluster](#virtengine.hpc.v1.MsgRegisterCluster)
     - [MsgRegisterClusterResponse](#virtengine.hpc.v1.MsgRegisterClusterResponse)
     - [MsgReportJobStatus](#virtengine.hpc.v1.MsgReportJobStatus)
     - [MsgReportJobStatusResponse](#virtengine.hpc.v1.MsgReportJobStatusResponse)
     - [MsgResolveDispute](#virtengine.hpc.v1.MsgResolveDispute)
     - [MsgResolveDisputeResponse](#virtengine.hpc.v1.MsgResolveDisputeResponse)
     - [MsgSubmitJob](#virtengine.hpc.v1.MsgSubmitJob)
     - [MsgSubmitJobResponse](#virtengine.hpc.v1.MsgSubmitJobResponse)
     - [MsgUpdateCluster](#virtengine.hpc.v1.MsgUpdateCluster)
     - [MsgUpdateClusterResponse](#virtengine.hpc.v1.MsgUpdateClusterResponse)
     - [MsgUpdateNodeMetadata](#virtengine.hpc.v1.MsgUpdateNodeMetadata)
     - [MsgUpdateNodeMetadataResponse](#virtengine.hpc.v1.MsgUpdateNodeMetadataResponse)
     - [MsgUpdateOffering](#virtengine.hpc.v1.MsgUpdateOffering)
     - [MsgUpdateOfferingResponse](#virtengine.hpc.v1.MsgUpdateOfferingResponse)
     - [MsgUpdateParams](#virtengine.hpc.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.hpc.v1.MsgUpdateParamsResponse)
   
     - [Msg](#virtengine.hpc.v1.Msg)
   
 - [virtengine/market/v1/bid.proto](#virtengine/market/v1/bid.proto)
     - [BidID](#virtengine.market.v1.BidID)
   
 - [virtengine/market/v1/order.proto](#virtengine/market/v1/order.proto)
     - [OrderID](#virtengine.market.v1.OrderID)
   
 - [virtengine/market/v1/types.proto](#virtengine/market/v1/types.proto)
     - [LeaseClosedReason](#virtengine.market.v1.LeaseClosedReason)
   
 - [virtengine/market/v1/lease.proto](#virtengine/market/v1/lease.proto)
     - [Lease](#virtengine.market.v1.Lease)
     - [LeaseID](#virtengine.market.v1.LeaseID)
   
     - [Lease.State](#virtengine.market.v1.Lease.State)
   
 - [virtengine/market/v1/event.proto](#virtengine/market/v1/event.proto)
     - [EventBidClosed](#virtengine.market.v1.EventBidClosed)
     - [EventBidCreated](#virtengine.market.v1.EventBidCreated)
     - [EventLeaseClosed](#virtengine.market.v1.EventLeaseClosed)
     - [EventLeaseCreated](#virtengine.market.v1.EventLeaseCreated)
     - [EventOrderClosed](#virtengine.market.v1.EventOrderClosed)
     - [EventOrderCreated](#virtengine.market.v1.EventOrderCreated)
   
 - [virtengine/market/v1/filters.proto](#virtengine/market/v1/filters.proto)
     - [LeaseFilters](#virtengine.market.v1.LeaseFilters)
   
 - [virtengine/market/v1beta5/resourcesoffer.proto](#virtengine/market/v1beta5/resourcesoffer.proto)
     - [ResourceOffer](#virtengine.market.v1beta5.ResourceOffer)
   
 - [virtengine/market/v1beta5/bid.proto](#virtengine/market/v1beta5/bid.proto)
     - [Bid](#virtengine.market.v1beta5.Bid)
   
     - [Bid.State](#virtengine.market.v1beta5.Bid.State)
   
 - [virtengine/market/v1beta5/bidmsg.proto](#virtengine/market/v1beta5/bidmsg.proto)
     - [MsgCloseBid](#virtengine.market.v1beta5.MsgCloseBid)
     - [MsgCloseBidResponse](#virtengine.market.v1beta5.MsgCloseBidResponse)
     - [MsgCreateBid](#virtengine.market.v1beta5.MsgCreateBid)
     - [MsgCreateBidResponse](#virtengine.market.v1beta5.MsgCreateBidResponse)
   
 - [virtengine/market/v1beta5/filters.proto](#virtengine/market/v1beta5/filters.proto)
     - [BidFilters](#virtengine.market.v1beta5.BidFilters)
     - [OrderFilters](#virtengine.market.v1beta5.OrderFilters)
   
 - [virtengine/market/v1beta5/params.proto](#virtengine/market/v1beta5/params.proto)
     - [Params](#virtengine.market.v1beta5.Params)
   
 - [virtengine/market/v1beta5/order.proto](#virtengine/market/v1beta5/order.proto)
     - [Order](#virtengine.market.v1beta5.Order)
   
     - [Order.State](#virtengine.market.v1beta5.Order.State)
   
 - [virtengine/market/v1beta5/genesis.proto](#virtengine/market/v1beta5/genesis.proto)
     - [GenesisState](#virtengine.market.v1beta5.GenesisState)
   
 - [virtengine/market/v1beta5/leasemsg.proto](#virtengine/market/v1beta5/leasemsg.proto)
     - [MsgCloseLease](#virtengine.market.v1beta5.MsgCloseLease)
     - [MsgCloseLeaseResponse](#virtengine.market.v1beta5.MsgCloseLeaseResponse)
     - [MsgCreateLease](#virtengine.market.v1beta5.MsgCreateLease)
     - [MsgCreateLeaseResponse](#virtengine.market.v1beta5.MsgCreateLeaseResponse)
     - [MsgWithdrawLease](#virtengine.market.v1beta5.MsgWithdrawLease)
     - [MsgWithdrawLeaseResponse](#virtengine.market.v1beta5.MsgWithdrawLeaseResponse)
   
 - [virtengine/market/v1beta5/paramsmsg.proto](#virtengine/market/v1beta5/paramsmsg.proto)
     - [MsgUpdateParams](#virtengine.market.v1beta5.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.market.v1beta5.MsgUpdateParamsResponse)
   
 - [virtengine/market/v1beta5/query.proto](#virtengine/market/v1beta5/query.proto)
     - [QueryBidRequest](#virtengine.market.v1beta5.QueryBidRequest)
     - [QueryBidResponse](#virtengine.market.v1beta5.QueryBidResponse)
     - [QueryBidsRequest](#virtengine.market.v1beta5.QueryBidsRequest)
     - [QueryBidsResponse](#virtengine.market.v1beta5.QueryBidsResponse)
     - [QueryLeaseRequest](#virtengine.market.v1beta5.QueryLeaseRequest)
     - [QueryLeaseResponse](#virtengine.market.v1beta5.QueryLeaseResponse)
     - [QueryLeasesRequest](#virtengine.market.v1beta5.QueryLeasesRequest)
     - [QueryLeasesResponse](#virtengine.market.v1beta5.QueryLeasesResponse)
     - [QueryOrderRequest](#virtengine.market.v1beta5.QueryOrderRequest)
     - [QueryOrderResponse](#virtengine.market.v1beta5.QueryOrderResponse)
     - [QueryOrdersRequest](#virtengine.market.v1beta5.QueryOrdersRequest)
     - [QueryOrdersResponse](#virtengine.market.v1beta5.QueryOrdersResponse)
     - [QueryParamsRequest](#virtengine.market.v1beta5.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.market.v1beta5.QueryParamsResponse)
   
     - [Query](#virtengine.market.v1beta5.Query)
   
 - [virtengine/market/v1beta5/service.proto](#virtengine/market/v1beta5/service.proto)
     - [Msg](#virtengine.market.v1beta5.Msg)
   
 - [virtengine/marketplace/v1/types.proto](#virtengine/marketplace/v1/types.proto)
     - [EncryptedProviderSecrets](#virtengine.marketplace.v1.EncryptedProviderSecrets)
     - [IdentityRequirement](#virtengine.marketplace.v1.IdentityRequirement)
     - [Offering](#virtengine.marketplace.v1.Offering)
     - [Offering.PublicMetadataEntry](#virtengine.marketplace.v1.Offering.PublicMetadataEntry)
     - [Offering.SpecificationsEntry](#virtengine.marketplace.v1.Offering.SpecificationsEntry)
     - [OfferingID](#virtengine.marketplace.v1.OfferingID)
     - [PricingInfo](#virtengine.marketplace.v1.PricingInfo)
     - [PricingInfo.UsageRatesEntry](#virtengine.marketplace.v1.PricingInfo.UsageRatesEntry)
   
     - [OfferingCategory](#virtengine.marketplace.v1.OfferingCategory)
     - [OfferingState](#virtengine.marketplace.v1.OfferingState)
     - [PricingModel](#virtengine.marketplace.v1.PricingModel)
   
 - [virtengine/marketplace/v1/tx.proto](#virtengine/marketplace/v1/tx.proto)
     - [MsgAcceptBid](#virtengine.marketplace.v1.MsgAcceptBid)
     - [MsgAcceptBidResponse](#virtengine.marketplace.v1.MsgAcceptBidResponse)
     - [MsgCreateOffering](#virtengine.marketplace.v1.MsgCreateOffering)
     - [MsgCreateOfferingResponse](#virtengine.marketplace.v1.MsgCreateOfferingResponse)
     - [MsgDeactivateOffering](#virtengine.marketplace.v1.MsgDeactivateOffering)
     - [MsgDeactivateOfferingResponse](#virtengine.marketplace.v1.MsgDeactivateOfferingResponse)
     - [MsgTerminateAllocation](#virtengine.marketplace.v1.MsgTerminateAllocation)
     - [MsgTerminateAllocationResponse](#virtengine.marketplace.v1.MsgTerminateAllocationResponse)
     - [MsgUpdateOffering](#virtengine.marketplace.v1.MsgUpdateOffering)
     - [MsgUpdateOfferingResponse](#virtengine.marketplace.v1.MsgUpdateOfferingResponse)
     - [MsgWaldurCallback](#virtengine.marketplace.v1.MsgWaldurCallback)
     - [MsgWaldurCallbackResponse](#virtengine.marketplace.v1.MsgWaldurCallbackResponse)
   
     - [Msg](#virtengine.marketplace.v1.Msg)
   
 - [virtengine/mfa/v1/types.proto](#virtengine/mfa/v1/types.proto)
     - [AuthorizationSession](#virtengine.mfa.v1.AuthorizationSession)
     - [Challenge](#virtengine.mfa.v1.Challenge)
     - [ChallengeMetadata](#virtengine.mfa.v1.ChallengeMetadata)
     - [ChallengeResponse](#virtengine.mfa.v1.ChallengeResponse)
     - [ClientInfo](#virtengine.mfa.v1.ClientInfo)
     - [DeviceInfo](#virtengine.mfa.v1.DeviceInfo)
     - [EventChallengeVerified](#virtengine.mfa.v1.EventChallengeVerified)
     - [EventFactorEnrolled](#virtengine.mfa.v1.EventFactorEnrolled)
     - [EventFactorRevoked](#virtengine.mfa.v1.EventFactorRevoked)
     - [EventMFAPolicyUpdated](#virtengine.mfa.v1.EventMFAPolicyUpdated)
     - [FIDO2ChallengeData](#virtengine.mfa.v1.FIDO2ChallengeData)
     - [FIDO2CredentialInfo](#virtengine.mfa.v1.FIDO2CredentialInfo)
     - [FactorCombination](#virtengine.mfa.v1.FactorCombination)
     - [FactorEnrollment](#virtengine.mfa.v1.FactorEnrollment)
     - [FactorMetadata](#virtengine.mfa.v1.FactorMetadata)
     - [HardwareKeyChallenge](#virtengine.mfa.v1.HardwareKeyChallenge)
     - [HardwareKeyEnrollment](#virtengine.mfa.v1.HardwareKeyEnrollment)
     - [MFAPolicy](#virtengine.mfa.v1.MFAPolicy)
     - [MFAProof](#virtengine.mfa.v1.MFAProof)
     - [OTPChallengeInfo](#virtengine.mfa.v1.OTPChallengeInfo)
     - [Params](#virtengine.mfa.v1.Params)
     - [SensitiveTxConfig](#virtengine.mfa.v1.SensitiveTxConfig)
     - [SmartCardInfo](#virtengine.mfa.v1.SmartCardInfo)
     - [TrustedDevice](#virtengine.mfa.v1.TrustedDevice)
     - [TrustedDevicePolicy](#virtengine.mfa.v1.TrustedDevicePolicy)
   
     - [ChallengeStatus](#virtengine.mfa.v1.ChallengeStatus)
     - [FactorEnrollmentStatus](#virtengine.mfa.v1.FactorEnrollmentStatus)
     - [FactorSecurityLevel](#virtengine.mfa.v1.FactorSecurityLevel)
     - [FactorType](#virtengine.mfa.v1.FactorType)
     - [HardwareKeyType](#virtengine.mfa.v1.HardwareKeyType)
     - [RevocationStatus](#virtengine.mfa.v1.RevocationStatus)
     - [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType)
   
 - [virtengine/mfa/v1/genesis.proto](#virtengine/mfa/v1/genesis.proto)
     - [GenesisState](#virtengine.mfa.v1.GenesisState)
   
 - [virtengine/mfa/v1/query.proto](#virtengine/mfa/v1/query.proto)
     - [QueryAllSensitiveTxConfigsRequest](#virtengine.mfa.v1.QueryAllSensitiveTxConfigsRequest)
     - [QueryAllSensitiveTxConfigsResponse](#virtengine.mfa.v1.QueryAllSensitiveTxConfigsResponse)
     - [QueryAuthorizationSessionRequest](#virtengine.mfa.v1.QueryAuthorizationSessionRequest)
     - [QueryAuthorizationSessionResponse](#virtengine.mfa.v1.QueryAuthorizationSessionResponse)
     - [QueryChallengeRequest](#virtengine.mfa.v1.QueryChallengeRequest)
     - [QueryChallengeResponse](#virtengine.mfa.v1.QueryChallengeResponse)
     - [QueryFactorEnrollmentRequest](#virtengine.mfa.v1.QueryFactorEnrollmentRequest)
     - [QueryFactorEnrollmentResponse](#virtengine.mfa.v1.QueryFactorEnrollmentResponse)
     - [QueryFactorEnrollmentsRequest](#virtengine.mfa.v1.QueryFactorEnrollmentsRequest)
     - [QueryFactorEnrollmentsResponse](#virtengine.mfa.v1.QueryFactorEnrollmentsResponse)
     - [QueryMFAPolicyRequest](#virtengine.mfa.v1.QueryMFAPolicyRequest)
     - [QueryMFAPolicyResponse](#virtengine.mfa.v1.QueryMFAPolicyResponse)
     - [QueryMFARequiredRequest](#virtengine.mfa.v1.QueryMFARequiredRequest)
     - [QueryMFARequiredResponse](#virtengine.mfa.v1.QueryMFARequiredResponse)
     - [QueryParamsRequest](#virtengine.mfa.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.mfa.v1.QueryParamsResponse)
     - [QueryPendingChallengesRequest](#virtengine.mfa.v1.QueryPendingChallengesRequest)
     - [QueryPendingChallengesResponse](#virtengine.mfa.v1.QueryPendingChallengesResponse)
     - [QuerySensitiveTxConfigRequest](#virtengine.mfa.v1.QuerySensitiveTxConfigRequest)
     - [QuerySensitiveTxConfigResponse](#virtengine.mfa.v1.QuerySensitiveTxConfigResponse)
     - [QueryTrustedDevicesRequest](#virtengine.mfa.v1.QueryTrustedDevicesRequest)
     - [QueryTrustedDevicesResponse](#virtengine.mfa.v1.QueryTrustedDevicesResponse)
   
     - [Query](#virtengine.mfa.v1.Query)
   
 - [virtengine/mfa/v1/tx.proto](#virtengine/mfa/v1/tx.proto)
     - [MsgAddTrustedDevice](#virtengine.mfa.v1.MsgAddTrustedDevice)
     - [MsgAddTrustedDeviceResponse](#virtengine.mfa.v1.MsgAddTrustedDeviceResponse)
     - [MsgCreateChallenge](#virtengine.mfa.v1.MsgCreateChallenge)
     - [MsgCreateChallengeResponse](#virtengine.mfa.v1.MsgCreateChallengeResponse)
     - [MsgEnrollFactor](#virtengine.mfa.v1.MsgEnrollFactor)
     - [MsgEnrollFactorResponse](#virtengine.mfa.v1.MsgEnrollFactorResponse)
     - [MsgRemoveTrustedDevice](#virtengine.mfa.v1.MsgRemoveTrustedDevice)
     - [MsgRemoveTrustedDeviceResponse](#virtengine.mfa.v1.MsgRemoveTrustedDeviceResponse)
     - [MsgRevokeFactor](#virtengine.mfa.v1.MsgRevokeFactor)
     - [MsgRevokeFactorResponse](#virtengine.mfa.v1.MsgRevokeFactorResponse)
     - [MsgSetMFAPolicy](#virtengine.mfa.v1.MsgSetMFAPolicy)
     - [MsgSetMFAPolicyResponse](#virtengine.mfa.v1.MsgSetMFAPolicyResponse)
     - [MsgUpdateParams](#virtengine.mfa.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.mfa.v1.MsgUpdateParamsResponse)
     - [MsgUpdateSensitiveTxConfig](#virtengine.mfa.v1.MsgUpdateSensitiveTxConfig)
     - [MsgUpdateSensitiveTxConfigResponse](#virtengine.mfa.v1.MsgUpdateSensitiveTxConfigResponse)
     - [MsgVerifyChallenge](#virtengine.mfa.v1.MsgVerifyChallenge)
     - [MsgVerifyChallengeResponse](#virtengine.mfa.v1.MsgVerifyChallengeResponse)
   
     - [Msg](#virtengine.mfa.v1.Msg)
   
 - [virtengine/oracle/v1/prices.proto](#virtengine/oracle/v1/prices.proto)
     - [AggregatedPrice](#virtengine.oracle.v1.AggregatedPrice)
     - [DataID](#virtengine.oracle.v1.DataID)
     - [PriceData](#virtengine.oracle.v1.PriceData)
     - [PriceDataID](#virtengine.oracle.v1.PriceDataID)
     - [PriceDataRecordID](#virtengine.oracle.v1.PriceDataRecordID)
     - [PriceDataState](#virtengine.oracle.v1.PriceDataState)
     - [PriceHealth](#virtengine.oracle.v1.PriceHealth)
     - [PricesFilter](#virtengine.oracle.v1.PricesFilter)
     - [QueryPricesRequest](#virtengine.oracle.v1.QueryPricesRequest)
     - [QueryPricesResponse](#virtengine.oracle.v1.QueryPricesResponse)
   
 - [virtengine/oracle/v1/events.proto](#virtengine/oracle/v1/events.proto)
     - [EventPriceData](#virtengine.oracle.v1.EventPriceData)
     - [EventPriceRecovered](#virtengine.oracle.v1.EventPriceRecovered)
     - [EventPriceStaleWarning](#virtengine.oracle.v1.EventPriceStaleWarning)
     - [EventPriceStaled](#virtengine.oracle.v1.EventPriceStaled)
   
 - [virtengine/oracle/v1/params.proto](#virtengine/oracle/v1/params.proto)
     - [Params](#virtengine.oracle.v1.Params)
     - [PythContractParams](#virtengine.oracle.v1.PythContractParams)
   
 - [virtengine/oracle/v1/genesis.proto](#virtengine/oracle/v1/genesis.proto)
     - [GenesisState](#virtengine.oracle.v1.GenesisState)
   
 - [virtengine/oracle/v1/msgs.proto](#virtengine/oracle/v1/msgs.proto)
     - [MsgAddPriceEntry](#virtengine.oracle.v1.MsgAddPriceEntry)
     - [MsgAddPriceEntryResponse](#virtengine.oracle.v1.MsgAddPriceEntryResponse)
     - [MsgUpdateParams](#virtengine.oracle.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.oracle.v1.MsgUpdateParamsResponse)
   
 - [virtengine/oracle/v1/query.proto](#virtengine/oracle/v1/query.proto)
     - [QueryAggregatedPriceRequest](#virtengine.oracle.v1.QueryAggregatedPriceRequest)
     - [QueryAggregatedPriceResponse](#virtengine.oracle.v1.QueryAggregatedPriceResponse)
     - [QueryParamsRequest](#virtengine.oracle.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.oracle.v1.QueryParamsResponse)
     - [QueryPriceFeedConfigRequest](#virtengine.oracle.v1.QueryPriceFeedConfigRequest)
     - [QueryPriceFeedConfigResponse](#virtengine.oracle.v1.QueryPriceFeedConfigResponse)
   
     - [Query](#virtengine.oracle.v1.Query)
   
 - [virtengine/oracle/v1/service.proto](#virtengine/oracle/v1/service.proto)
     - [Msg](#virtengine.oracle.v1.Msg)
   
 - [virtengine/provider/v1beta4/event.proto](#virtengine/provider/v1beta4/event.proto)
     - [EventProviderCreated](#virtengine.provider.v1beta4.EventProviderCreated)
     - [EventProviderDeleted](#virtengine.provider.v1beta4.EventProviderDeleted)
     - [EventProviderDomainVerificationStarted](#virtengine.provider.v1beta4.EventProviderDomainVerificationStarted)
     - [EventProviderDomainVerified](#virtengine.provider.v1beta4.EventProviderDomainVerified)
     - [EventProviderUpdated](#virtengine.provider.v1beta4.EventProviderUpdated)
   
 - [virtengine/provider/v1beta4/provider.proto](#virtengine/provider/v1beta4/provider.proto)
     - [Info](#virtengine.provider.v1beta4.Info)
     - [Provider](#virtengine.provider.v1beta4.Provider)
   
 - [virtengine/provider/v1beta4/genesis.proto](#virtengine/provider/v1beta4/genesis.proto)
     - [GenesisState](#virtengine.provider.v1beta4.GenesisState)
   
 - [virtengine/provider/v1beta4/msg.proto](#virtengine/provider/v1beta4/msg.proto)
     - [MsgCreateProvider](#virtengine.provider.v1beta4.MsgCreateProvider)
     - [MsgCreateProviderResponse](#virtengine.provider.v1beta4.MsgCreateProviderResponse)
     - [MsgDeleteProvider](#virtengine.provider.v1beta4.MsgDeleteProvider)
     - [MsgDeleteProviderResponse](#virtengine.provider.v1beta4.MsgDeleteProviderResponse)
     - [MsgGenerateDomainVerificationToken](#virtengine.provider.v1beta4.MsgGenerateDomainVerificationToken)
     - [MsgGenerateDomainVerificationTokenResponse](#virtengine.provider.v1beta4.MsgGenerateDomainVerificationTokenResponse)
     - [MsgUpdateProvider](#virtengine.provider.v1beta4.MsgUpdateProvider)
     - [MsgUpdateProviderResponse](#virtengine.provider.v1beta4.MsgUpdateProviderResponse)
     - [MsgVerifyProviderDomain](#virtengine.provider.v1beta4.MsgVerifyProviderDomain)
     - [MsgVerifyProviderDomainResponse](#virtengine.provider.v1beta4.MsgVerifyProviderDomainResponse)
   
 - [virtengine/provider/v1beta4/query.proto](#virtengine/provider/v1beta4/query.proto)
     - [QueryProviderRequest](#virtengine.provider.v1beta4.QueryProviderRequest)
     - [QueryProviderResponse](#virtengine.provider.v1beta4.QueryProviderResponse)
     - [QueryProvidersRequest](#virtengine.provider.v1beta4.QueryProvidersRequest)
     - [QueryProvidersResponse](#virtengine.provider.v1beta4.QueryProvidersResponse)
   
     - [Query](#virtengine.provider.v1beta4.Query)
   
 - [virtengine/provider/v1beta4/service.proto](#virtengine/provider/v1beta4/service.proto)
     - [Msg](#virtengine.provider.v1beta4.Msg)
   
 - [virtengine/review/v1/tx.proto](#virtengine/review/v1/tx.proto)
     - [GenesisState](#virtengine.review.v1.GenesisState)
     - [MsgDeleteReview](#virtengine.review.v1.MsgDeleteReview)
     - [MsgDeleteReviewResponse](#virtengine.review.v1.MsgDeleteReviewResponse)
     - [MsgSubmitReview](#virtengine.review.v1.MsgSubmitReview)
     - [MsgSubmitReviewResponse](#virtengine.review.v1.MsgSubmitReviewResponse)
     - [MsgUpdateParams](#virtengine.review.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.review.v1.MsgUpdateParamsResponse)
     - [Params](#virtengine.review.v1.Params)
     - [Review](#virtengine.review.v1.Review)
   
     - [Msg](#virtengine.review.v1.Msg)
   
 - [virtengine/roles/v1/types.proto](#virtengine/roles/v1/types.proto)
     - [AccountStateRecord](#virtengine.roles.v1.AccountStateRecord)
     - [EventAccountStateChanged](#virtengine.roles.v1.EventAccountStateChanged)
     - [EventAdminNominated](#virtengine.roles.v1.EventAdminNominated)
     - [EventRoleAssigned](#virtengine.roles.v1.EventRoleAssigned)
     - [EventRoleRevoked](#virtengine.roles.v1.EventRoleRevoked)
     - [Params](#virtengine.roles.v1.Params)
     - [RoleAssignment](#virtengine.roles.v1.RoleAssignment)
   
     - [AccountState](#virtengine.roles.v1.AccountState)
     - [Role](#virtengine.roles.v1.Role)
   
 - [virtengine/roles/v1/genesis.proto](#virtengine/roles/v1/genesis.proto)
     - [GenesisState](#virtengine.roles.v1.GenesisState)
   
 - [virtengine/roles/v1/query.proto](#virtengine/roles/v1/query.proto)
     - [QueryAccountRolesRequest](#virtengine.roles.v1.QueryAccountRolesRequest)
     - [QueryAccountRolesResponse](#virtengine.roles.v1.QueryAccountRolesResponse)
     - [QueryAccountStateRequest](#virtengine.roles.v1.QueryAccountStateRequest)
     - [QueryAccountStateResponse](#virtengine.roles.v1.QueryAccountStateResponse)
     - [QueryGenesisAccountsRequest](#virtengine.roles.v1.QueryGenesisAccountsRequest)
     - [QueryGenesisAccountsResponse](#virtengine.roles.v1.QueryGenesisAccountsResponse)
     - [QueryHasRoleRequest](#virtengine.roles.v1.QueryHasRoleRequest)
     - [QueryHasRoleResponse](#virtengine.roles.v1.QueryHasRoleResponse)
     - [QueryParamsRequest](#virtengine.roles.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.roles.v1.QueryParamsResponse)
     - [QueryRoleMembersRequest](#virtengine.roles.v1.QueryRoleMembersRequest)
     - [QueryRoleMembersResponse](#virtengine.roles.v1.QueryRoleMembersResponse)
   
     - [Query](#virtengine.roles.v1.Query)
   
 - [virtengine/roles/v1/store.proto](#virtengine/roles/v1/store.proto)
     - [AccountStateStore](#virtengine.roles.v1.AccountStateStore)
     - [ParamsStore](#virtengine.roles.v1.ParamsStore)
     - [RoleAssignmentStore](#virtengine.roles.v1.RoleAssignmentStore)
   
 - [virtengine/roles/v1/tx.proto](#virtengine/roles/v1/tx.proto)
     - [MsgAssignRole](#virtengine.roles.v1.MsgAssignRole)
     - [MsgAssignRoleResponse](#virtengine.roles.v1.MsgAssignRoleResponse)
     - [MsgNominateAdmin](#virtengine.roles.v1.MsgNominateAdmin)
     - [MsgNominateAdminResponse](#virtengine.roles.v1.MsgNominateAdminResponse)
     - [MsgRevokeRole](#virtengine.roles.v1.MsgRevokeRole)
     - [MsgRevokeRoleResponse](#virtengine.roles.v1.MsgRevokeRoleResponse)
     - [MsgSetAccountState](#virtengine.roles.v1.MsgSetAccountState)
     - [MsgSetAccountStateResponse](#virtengine.roles.v1.MsgSetAccountStateResponse)
     - [MsgUpdateParams](#virtengine.roles.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.roles.v1.MsgUpdateParamsResponse)
   
     - [Msg](#virtengine.roles.v1.Msg)
   
 - [virtengine/settlement/v1/tx.proto](#virtengine/settlement/v1/tx.proto)
     - [MsgAcknowledgeUsage](#virtengine.settlement.v1.MsgAcknowledgeUsage)
     - [MsgAcknowledgeUsageResponse](#virtengine.settlement.v1.MsgAcknowledgeUsageResponse)
     - [MsgActivateEscrow](#virtengine.settlement.v1.MsgActivateEscrow)
     - [MsgActivateEscrowResponse](#virtengine.settlement.v1.MsgActivateEscrowResponse)
     - [MsgClaimRewards](#virtengine.settlement.v1.MsgClaimRewards)
     - [MsgClaimRewardsResponse](#virtengine.settlement.v1.MsgClaimRewardsResponse)
     - [MsgCreateEscrow](#virtengine.settlement.v1.MsgCreateEscrow)
     - [MsgCreateEscrowResponse](#virtengine.settlement.v1.MsgCreateEscrowResponse)
     - [MsgDisputeEscrow](#virtengine.settlement.v1.MsgDisputeEscrow)
     - [MsgDisputeEscrowResponse](#virtengine.settlement.v1.MsgDisputeEscrowResponse)
     - [MsgRecordUsage](#virtengine.settlement.v1.MsgRecordUsage)
     - [MsgRecordUsageResponse](#virtengine.settlement.v1.MsgRecordUsageResponse)
     - [MsgRefundEscrow](#virtengine.settlement.v1.MsgRefundEscrow)
     - [MsgRefundEscrowResponse](#virtengine.settlement.v1.MsgRefundEscrowResponse)
     - [MsgReleaseEscrow](#virtengine.settlement.v1.MsgReleaseEscrow)
     - [MsgReleaseEscrowResponse](#virtengine.settlement.v1.MsgReleaseEscrowResponse)
     - [MsgSettleOrder](#virtengine.settlement.v1.MsgSettleOrder)
     - [MsgSettleOrderResponse](#virtengine.settlement.v1.MsgSettleOrderResponse)
   
     - [Msg](#virtengine.settlement.v1.Msg)
   
 - [virtengine/staking/v1/params.proto](#virtengine/staking/v1/params.proto)
     - [Params](#virtengine.staking.v1.Params)
   
 - [virtengine/staking/v1/types.proto](#virtengine/staking/v1/types.proto)
     - [DoubleSignEvidence](#virtengine.staking.v1.DoubleSignEvidence)
     - [InvalidVEIDAttestation](#virtengine.staking.v1.InvalidVEIDAttestation)
     - [RewardEpoch](#virtengine.staking.v1.RewardEpoch)
     - [SlashConfig](#virtengine.staking.v1.SlashConfig)
     - [SlashRecord](#virtengine.staking.v1.SlashRecord)
     - [ValidatorPerformance](#virtengine.staking.v1.ValidatorPerformance)
     - [ValidatorReward](#virtengine.staking.v1.ValidatorReward)
     - [ValidatorSigningInfo](#virtengine.staking.v1.ValidatorSigningInfo)
   
     - [RewardType](#virtengine.staking.v1.RewardType)
     - [SlashReason](#virtengine.staking.v1.SlashReason)
   
 - [virtengine/staking/v1/genesis.proto](#virtengine/staking/v1/genesis.proto)
     - [GenesisState](#virtengine.staking.v1.GenesisState)
   
 - [virtengine/staking/v1/tx.proto](#virtengine/staking/v1/tx.proto)
     - [MsgRecordPerformance](#virtengine.staking.v1.MsgRecordPerformance)
     - [MsgRecordPerformanceResponse](#virtengine.staking.v1.MsgRecordPerformanceResponse)
     - [MsgSlashValidator](#virtengine.staking.v1.MsgSlashValidator)
     - [MsgSlashValidatorResponse](#virtengine.staking.v1.MsgSlashValidatorResponse)
     - [MsgUnjailValidator](#virtengine.staking.v1.MsgUnjailValidator)
     - [MsgUnjailValidatorResponse](#virtengine.staking.v1.MsgUnjailValidatorResponse)
     - [MsgUpdateParams](#virtengine.staking.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.staking.v1.MsgUpdateParamsResponse)
   
     - [Msg](#virtengine.staking.v1.Msg)
   
 - [virtengine/take/v1/params.proto](#virtengine/take/v1/params.proto)
     - [DenomTakeRate](#virtengine.take.v1.DenomTakeRate)
     - [Params](#virtengine.take.v1.Params)
   
 - [virtengine/take/v1/genesis.proto](#virtengine/take/v1/genesis.proto)
     - [GenesisState](#virtengine.take.v1.GenesisState)
   
 - [virtengine/take/v1/paramsmsg.proto](#virtengine/take/v1/paramsmsg.proto)
     - [MsgUpdateParams](#virtengine.take.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.take.v1.MsgUpdateParamsResponse)
   
 - [virtengine/take/v1/query.proto](#virtengine/take/v1/query.proto)
     - [QueryParamsRequest](#virtengine.take.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.take.v1.QueryParamsResponse)
   
     - [Query](#virtengine.take.v1.Query)
   
 - [virtengine/take/v1/service.proto](#virtengine/take/v1/service.proto)
     - [Msg](#virtengine.take.v1.Msg)
   
 - [virtengine/veid/v1/appeal.proto](#virtengine/veid/v1/appeal.proto)
     - [AppealParams](#virtengine.veid.v1.AppealParams)
     - [AppealRecord](#virtengine.veid.v1.AppealRecord)
     - [AppealSummary](#virtengine.veid.v1.AppealSummary)
     - [MsgClaimAppeal](#virtengine.veid.v1.MsgClaimAppeal)
     - [MsgClaimAppealResponse](#virtengine.veid.v1.MsgClaimAppealResponse)
     - [MsgResolveAppeal](#virtengine.veid.v1.MsgResolveAppeal)
     - [MsgResolveAppealResponse](#virtengine.veid.v1.MsgResolveAppealResponse)
     - [MsgSubmitAppeal](#virtengine.veid.v1.MsgSubmitAppeal)
     - [MsgSubmitAppealResponse](#virtengine.veid.v1.MsgSubmitAppealResponse)
     - [MsgWithdrawAppeal](#virtengine.veid.v1.MsgWithdrawAppeal)
     - [MsgWithdrawAppealResponse](#virtengine.veid.v1.MsgWithdrawAppealResponse)
   
     - [AppealStatus](#virtengine.veid.v1.AppealStatus)
   
 - [virtengine/veid/v1/compliance.proto](#virtengine/veid/v1/compliance.proto)
     - [ComplianceAttestation](#virtengine.veid.v1.ComplianceAttestation)
     - [ComplianceCheckResult](#virtengine.veid.v1.ComplianceCheckResult)
     - [ComplianceParams](#virtengine.veid.v1.ComplianceParams)
     - [ComplianceProvider](#virtengine.veid.v1.ComplianceProvider)
     - [ComplianceRecord](#virtengine.veid.v1.ComplianceRecord)
     - [MsgAttestCompliance](#virtengine.veid.v1.MsgAttestCompliance)
     - [MsgAttestComplianceResponse](#virtengine.veid.v1.MsgAttestComplianceResponse)
     - [MsgDeactivateComplianceProvider](#virtengine.veid.v1.MsgDeactivateComplianceProvider)
     - [MsgDeactivateComplianceProviderResponse](#virtengine.veid.v1.MsgDeactivateComplianceProviderResponse)
     - [MsgRegisterComplianceProvider](#virtengine.veid.v1.MsgRegisterComplianceProvider)
     - [MsgRegisterComplianceProviderResponse](#virtengine.veid.v1.MsgRegisterComplianceProviderResponse)
     - [MsgSubmitComplianceCheck](#virtengine.veid.v1.MsgSubmitComplianceCheck)
     - [MsgSubmitComplianceCheckResponse](#virtengine.veid.v1.MsgSubmitComplianceCheckResponse)
     - [MsgUpdateComplianceParams](#virtengine.veid.v1.MsgUpdateComplianceParams)
     - [MsgUpdateComplianceParamsResponse](#virtengine.veid.v1.MsgUpdateComplianceParamsResponse)
   
     - [ComplianceCheckType](#virtengine.veid.v1.ComplianceCheckType)
     - [ComplianceStatus](#virtengine.veid.v1.ComplianceStatus)
   
 - [virtengine/veid/v1/types.proto](#virtengine/veid/v1/types.proto)
     - [ApprovedClient](#virtengine.veid.v1.ApprovedClient)
     - [BorderlineParams](#virtengine.veid.v1.BorderlineParams)
     - [ConsentSettings](#virtengine.veid.v1.ConsentSettings)
     - [DerivedFeatures](#virtengine.veid.v1.DerivedFeatures)
     - [DerivedFeatures.DocFieldHashesEntry](#virtengine.veid.v1.DerivedFeatures.DocFieldHashesEntry)
     - [EncryptedPayloadEnvelope](#virtengine.veid.v1.EncryptedPayloadEnvelope)
     - [EncryptedPayloadEnvelope.MetadataEntry](#virtengine.veid.v1.EncryptedPayloadEnvelope.MetadataEntry)
     - [GlobalConsentUpdate](#virtengine.veid.v1.GlobalConsentUpdate)
     - [IdentityRecord](#virtengine.veid.v1.IdentityRecord)
     - [IdentityScope](#virtengine.veid.v1.IdentityScope)
     - [IdentityScore](#virtengine.veid.v1.IdentityScore)
     - [IdentityWallet](#virtengine.veid.v1.IdentityWallet)
     - [IdentityWallet.MetadataEntry](#virtengine.veid.v1.IdentityWallet.MetadataEntry)
     - [IdentityWallet.ScopeConsentsEntry](#virtengine.veid.v1.IdentityWallet.ScopeConsentsEntry)
     - [Params](#virtengine.veid.v1.Params)
     - [ScopeConsent](#virtengine.veid.v1.ScopeConsent)
     - [ScopeRef](#virtengine.veid.v1.ScopeRef)
     - [ScopeReference](#virtengine.veid.v1.ScopeReference)
     - [UploadMetadata](#virtengine.veid.v1.UploadMetadata)
     - [VerificationHistoryEntry](#virtengine.veid.v1.VerificationHistoryEntry)
   
     - [AccountStatus](#virtengine.veid.v1.AccountStatus)
     - [IdentityTier](#virtengine.veid.v1.IdentityTier)
     - [ScopeRefStatus](#virtengine.veid.v1.ScopeRefStatus)
     - [ScopeType](#virtengine.veid.v1.ScopeType)
     - [VerificationStatus](#virtengine.veid.v1.VerificationStatus)
     - [WalletStatus](#virtengine.veid.v1.WalletStatus)
   
 - [virtengine/veid/v1/model.proto](#virtengine/veid/v1/model.proto)
     - [MLModelInfo](#virtengine.veid.v1.MLModelInfo)
     - [ModelParams](#virtengine.veid.v1.ModelParams)
     - [ModelUpdateProposal](#virtengine.veid.v1.ModelUpdateProposal)
     - [ModelVersionHistory](#virtengine.veid.v1.ModelVersionHistory)
     - [ModelVersionState](#virtengine.veid.v1.ModelVersionState)
     - [MsgActivateModel](#virtengine.veid.v1.MsgActivateModel)
     - [MsgActivateModelResponse](#virtengine.veid.v1.MsgActivateModelResponse)
     - [MsgDeprecateModel](#virtengine.veid.v1.MsgDeprecateModel)
     - [MsgDeprecateModelResponse](#virtengine.veid.v1.MsgDeprecateModelResponse)
     - [MsgProposeModelUpdate](#virtengine.veid.v1.MsgProposeModelUpdate)
     - [MsgProposeModelUpdateResponse](#virtengine.veid.v1.MsgProposeModelUpdateResponse)
     - [MsgRegisterModel](#virtengine.veid.v1.MsgRegisterModel)
     - [MsgRegisterModelResponse](#virtengine.veid.v1.MsgRegisterModelResponse)
     - [MsgReportModelVersion](#virtengine.veid.v1.MsgReportModelVersion)
     - [MsgReportModelVersion.ModelVersionsEntry](#virtengine.veid.v1.MsgReportModelVersion.ModelVersionsEntry)
     - [MsgReportModelVersionResponse](#virtengine.veid.v1.MsgReportModelVersionResponse)
     - [MsgRevokeModel](#virtengine.veid.v1.MsgRevokeModel)
     - [MsgRevokeModelResponse](#virtengine.veid.v1.MsgRevokeModelResponse)
     - [ValidatorModelReport](#virtengine.veid.v1.ValidatorModelReport)
     - [ValidatorModelReport.ModelVersionsEntry](#virtengine.veid.v1.ValidatorModelReport.ModelVersionsEntry)
   
     - [ModelProposalStatus](#virtengine.veid.v1.ModelProposalStatus)
     - [ModelStatus](#virtengine.veid.v1.ModelStatus)
     - [ModelType](#virtengine.veid.v1.ModelType)
   
 - [virtengine/veid/v1/genesis.proto](#virtengine/veid/v1/genesis.proto)
     - [GenesisState](#virtengine.veid.v1.GenesisState)
   
 - [virtengine/veid/v1/query.proto](#virtengine/veid/v1/query.proto)
     - [PublicConsentInfo](#virtengine.veid.v1.PublicConsentInfo)
     - [PublicDerivedFeaturesInfo](#virtengine.veid.v1.PublicDerivedFeaturesInfo)
     - [PublicVerificationHistoryEntry](#virtengine.veid.v1.PublicVerificationHistoryEntry)
     - [PublicWalletInfo](#virtengine.veid.v1.PublicWalletInfo)
     - [QueryActiveModelsRequest](#virtengine.veid.v1.QueryActiveModelsRequest)
     - [QueryActiveModelsResponse](#virtengine.veid.v1.QueryActiveModelsResponse)
     - [QueryAppealParamsRequest](#virtengine.veid.v1.QueryAppealParamsRequest)
     - [QueryAppealParamsResponse](#virtengine.veid.v1.QueryAppealParamsResponse)
     - [QueryAppealRequest](#virtengine.veid.v1.QueryAppealRequest)
     - [QueryAppealResponse](#virtengine.veid.v1.QueryAppealResponse)
     - [QueryAppealsByScopeRequest](#virtengine.veid.v1.QueryAppealsByScopeRequest)
     - [QueryAppealsByScopeResponse](#virtengine.veid.v1.QueryAppealsByScopeResponse)
     - [QueryAppealsRequest](#virtengine.veid.v1.QueryAppealsRequest)
     - [QueryAppealsResponse](#virtengine.veid.v1.QueryAppealsResponse)
     - [QueryApprovedClientsRequest](#virtengine.veid.v1.QueryApprovedClientsRequest)
     - [QueryApprovedClientsResponse](#virtengine.veid.v1.QueryApprovedClientsResponse)
     - [QueryBorderlineParamsRequest](#virtengine.veid.v1.QueryBorderlineParamsRequest)
     - [QueryBorderlineParamsResponse](#virtengine.veid.v1.QueryBorderlineParamsResponse)
     - [QueryComplianceParamsRequest](#virtengine.veid.v1.QueryComplianceParamsRequest)
     - [QueryComplianceParamsResponse](#virtengine.veid.v1.QueryComplianceParamsResponse)
     - [QueryComplianceProviderRequest](#virtengine.veid.v1.QueryComplianceProviderRequest)
     - [QueryComplianceProviderResponse](#virtengine.veid.v1.QueryComplianceProviderResponse)
     - [QueryComplianceProvidersRequest](#virtengine.veid.v1.QueryComplianceProvidersRequest)
     - [QueryComplianceProvidersResponse](#virtengine.veid.v1.QueryComplianceProvidersResponse)
     - [QueryComplianceStatusRequest](#virtengine.veid.v1.QueryComplianceStatusRequest)
     - [QueryComplianceStatusResponse](#virtengine.veid.v1.QueryComplianceStatusResponse)
     - [QueryConsentSettingsRequest](#virtengine.veid.v1.QueryConsentSettingsRequest)
     - [QueryConsentSettingsResponse](#virtengine.veid.v1.QueryConsentSettingsResponse)
     - [QueryDerivedFeatureHashesRequest](#virtengine.veid.v1.QueryDerivedFeatureHashesRequest)
     - [QueryDerivedFeatureHashesResponse](#virtengine.veid.v1.QueryDerivedFeatureHashesResponse)
     - [QueryDerivedFeatureHashesResponse.DocFieldHashesEntry](#virtengine.veid.v1.QueryDerivedFeatureHashesResponse.DocFieldHashesEntry)
     - [QueryDerivedFeaturesRequest](#virtengine.veid.v1.QueryDerivedFeaturesRequest)
     - [QueryDerivedFeaturesResponse](#virtengine.veid.v1.QueryDerivedFeaturesResponse)
     - [QueryIdentityRecordRequest](#virtengine.veid.v1.QueryIdentityRecordRequest)
     - [QueryIdentityRecordResponse](#virtengine.veid.v1.QueryIdentityRecordResponse)
     - [QueryIdentityRequest](#virtengine.veid.v1.QueryIdentityRequest)
     - [QueryIdentityResponse](#virtengine.veid.v1.QueryIdentityResponse)
     - [QueryIdentityScoreRequest](#virtengine.veid.v1.QueryIdentityScoreRequest)
     - [QueryIdentityScoreResponse](#virtengine.veid.v1.QueryIdentityScoreResponse)
     - [QueryIdentityStatusRequest](#virtengine.veid.v1.QueryIdentityStatusRequest)
     - [QueryIdentityStatusResponse](#virtengine.veid.v1.QueryIdentityStatusResponse)
     - [QueryIdentityWalletRequest](#virtengine.veid.v1.QueryIdentityWalletRequest)
     - [QueryIdentityWalletResponse](#virtengine.veid.v1.QueryIdentityWalletResponse)
     - [QueryModelHistoryRequest](#virtengine.veid.v1.QueryModelHistoryRequest)
     - [QueryModelHistoryResponse](#virtengine.veid.v1.QueryModelHistoryResponse)
     - [QueryModelParamsRequest](#virtengine.veid.v1.QueryModelParamsRequest)
     - [QueryModelParamsResponse](#virtengine.veid.v1.QueryModelParamsResponse)
     - [QueryModelVersionRequest](#virtengine.veid.v1.QueryModelVersionRequest)
     - [QueryModelVersionResponse](#virtengine.veid.v1.QueryModelVersionResponse)
     - [QueryParamsRequest](#virtengine.veid.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.veid.v1.QueryParamsResponse)
     - [QueryScopeRequest](#virtengine.veid.v1.QueryScopeRequest)
     - [QueryScopeResponse](#virtengine.veid.v1.QueryScopeResponse)
     - [QueryScopesByTypeRequest](#virtengine.veid.v1.QueryScopesByTypeRequest)
     - [QueryScopesByTypeResponse](#virtengine.veid.v1.QueryScopesByTypeResponse)
     - [QueryScopesRequest](#virtengine.veid.v1.QueryScopesRequest)
     - [QueryScopesResponse](#virtengine.veid.v1.QueryScopesResponse)
     - [QueryValidatorModelSyncRequest](#virtengine.veid.v1.QueryValidatorModelSyncRequest)
     - [QueryValidatorModelSyncResponse](#virtengine.veid.v1.QueryValidatorModelSyncResponse)
     - [QueryVerificationHistoryRequest](#virtengine.veid.v1.QueryVerificationHistoryRequest)
     - [QueryVerificationHistoryResponse](#virtengine.veid.v1.QueryVerificationHistoryResponse)
     - [QueryWalletScopesRequest](#virtengine.veid.v1.QueryWalletScopesRequest)
     - [QueryWalletScopesResponse](#virtengine.veid.v1.QueryWalletScopesResponse)
     - [WalletScopeInfo](#virtengine.veid.v1.WalletScopeInfo)
   
     - [Query](#virtengine.veid.v1.Query)
   
 - [virtengine/veid/v1/tx.proto](#virtengine/veid/v1/tx.proto)
     - [MsgAddScopeToWallet](#virtengine.veid.v1.MsgAddScopeToWallet)
     - [MsgAddScopeToWalletResponse](#virtengine.veid.v1.MsgAddScopeToWalletResponse)
     - [MsgCompleteBorderlineFallback](#virtengine.veid.v1.MsgCompleteBorderlineFallback)
     - [MsgCompleteBorderlineFallbackResponse](#virtengine.veid.v1.MsgCompleteBorderlineFallbackResponse)
     - [MsgCreateIdentityWallet](#virtengine.veid.v1.MsgCreateIdentityWallet)
     - [MsgCreateIdentityWallet.MetadataEntry](#virtengine.veid.v1.MsgCreateIdentityWallet.MetadataEntry)
     - [MsgCreateIdentityWalletResponse](#virtengine.veid.v1.MsgCreateIdentityWalletResponse)
     - [MsgRebindWallet](#virtengine.veid.v1.MsgRebindWallet)
     - [MsgRebindWalletResponse](#virtengine.veid.v1.MsgRebindWalletResponse)
     - [MsgRequestVerification](#virtengine.veid.v1.MsgRequestVerification)
     - [MsgRequestVerificationResponse](#virtengine.veid.v1.MsgRequestVerificationResponse)
     - [MsgRevokeScope](#virtengine.veid.v1.MsgRevokeScope)
     - [MsgRevokeScopeFromWallet](#virtengine.veid.v1.MsgRevokeScopeFromWallet)
     - [MsgRevokeScopeFromWalletResponse](#virtengine.veid.v1.MsgRevokeScopeFromWalletResponse)
     - [MsgRevokeScopeResponse](#virtengine.veid.v1.MsgRevokeScopeResponse)
     - [MsgUpdateBorderlineParams](#virtengine.veid.v1.MsgUpdateBorderlineParams)
     - [MsgUpdateBorderlineParamsResponse](#virtengine.veid.v1.MsgUpdateBorderlineParamsResponse)
     - [MsgUpdateConsentSettings](#virtengine.veid.v1.MsgUpdateConsentSettings)
     - [MsgUpdateConsentSettingsResponse](#virtengine.veid.v1.MsgUpdateConsentSettingsResponse)
     - [MsgUpdateDerivedFeatures](#virtengine.veid.v1.MsgUpdateDerivedFeatures)
     - [MsgUpdateDerivedFeatures.DocFieldHashesEntry](#virtengine.veid.v1.MsgUpdateDerivedFeatures.DocFieldHashesEntry)
     - [MsgUpdateDerivedFeaturesResponse](#virtengine.veid.v1.MsgUpdateDerivedFeaturesResponse)
     - [MsgUpdateParams](#virtengine.veid.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.veid.v1.MsgUpdateParamsResponse)
     - [MsgUpdateScore](#virtengine.veid.v1.MsgUpdateScore)
     - [MsgUpdateScoreResponse](#virtengine.veid.v1.MsgUpdateScoreResponse)
     - [MsgUpdateVerificationStatus](#virtengine.veid.v1.MsgUpdateVerificationStatus)
     - [MsgUpdateVerificationStatusResponse](#virtengine.veid.v1.MsgUpdateVerificationStatusResponse)
     - [MsgUploadScope](#virtengine.veid.v1.MsgUploadScope)
     - [MsgUploadScopeResponse](#virtengine.veid.v1.MsgUploadScopeResponse)
   
     - [Msg](#virtengine.veid.v1.Msg)
   
 - [virtengine/wasm/v1/event.proto](#virtengine/wasm/v1/event.proto)
     - [EventMsgBlocked](#virtengine.wasm.v1.EventMsgBlocked)
   
 - [virtengine/wasm/v1/params.proto](#virtengine/wasm/v1/params.proto)
     - [Params](#virtengine.wasm.v1.Params)
   
 - [virtengine/wasm/v1/genesis.proto](#virtengine/wasm/v1/genesis.proto)
     - [GenesisState](#virtengine.wasm.v1.GenesisState)
   
 - [virtengine/wasm/v1/paramsmsg.proto](#virtengine/wasm/v1/paramsmsg.proto)
     - [MsgUpdateParams](#virtengine.wasm.v1.MsgUpdateParams)
     - [MsgUpdateParamsResponse](#virtengine.wasm.v1.MsgUpdateParamsResponse)
   
 - [virtengine/wasm/v1/query.proto](#virtengine/wasm/v1/query.proto)
     - [QueryParamsRequest](#virtengine.wasm.v1.QueryParamsRequest)
     - [QueryParamsResponse](#virtengine.wasm.v1.QueryParamsResponse)
   
     - [Query](#virtengine.wasm.v1.Query)
   
 - [virtengine/wasm/v1/service.proto](#virtengine/wasm/v1/service.proto)
     - [Msg](#virtengine.wasm.v1.Msg)
   
 - [Scalar Value Types](#scalar-value-types)

 
 
 <a name="virtengine/base/attributes/v1/attribute.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/attributes/v1/attribute.proto
 

 
 <a name="virtengine.base.attributes.v1.Attribute"></a>

 ### Attribute
 Attribute represents an arbitrary attribute key-value pair.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  | Key of the attribute (e.g., "region", "cpu_architecture", etc.). |
 | `value` | [string](#string) |  | Value of the attribute (e.g., "us-west", "x86_64", etc.). |
 
 

 

 
 <a name="virtengine.base.attributes.v1.PlacementRequirements"></a>

 ### PlacementRequirements
 PlacementRequirements represents the requirements for a provider placement on the network.
It is used to specify the characteristics and constraints of a provider that can be used to satisfy a deployment request.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `signed_by` | [SignedBy](#virtengine.base.attributes.v1.SignedBy) |  | SignedBy holds the list of keys that tenants expect to have signatures from. |
 | `attributes` | [Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attribute holds the list of attributes tenant expects from the provider. |
 
 

 

 
 <a name="virtengine.base.attributes.v1.SignedBy"></a>

 ### SignedBy
 SignedBy represents validation accounts that tenant expects signatures for provider attributes.
AllOf has precedence i.e. if there is at least one entry AnyOf is ignored regardless to how many
entries there.

TODO: this behaviour to be discussed

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `all_of` | [string](#string) | repeated | AllOf indicates all keys in this list must have signed attributes. |
 | `any_of` | [string](#string) | repeated | AnyOf means that at least one of the keys from the list must have signed attributes. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/audit/v1/audit.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/audit/v1/audit.proto
 

 
 <a name="virtengine.audit.v1.AttributesFilters"></a>

 ### AttributesFilters
 AttributesFilters defines attribute filters that can be used to filter deployments.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `auditors` | [string](#string) | repeated | Auditors contains a list of auditor account bech32 addresses. |
 | `owners` | [string](#string) | repeated | Owners contains a list of owner account bech32 addresses. |
 
 

 

 
 <a name="virtengine.audit.v1.AuditedAttributesStore"></a>

 ### AuditedAttributesStore
 AuditedAttributesStore stores the audited attributes of the provider.
Attributes that have been audited are those that have been verified by an auditor.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes holds a list of key-value pairs of provider attributes. Attributes are arbitrary values that a provider exposes. |
 
 

 

 
 <a name="virtengine.audit.v1.AuditedProvider"></a>

 ### AuditedProvider
 AuditedProvider stores owner, auditor and attributes details.
An AuditedProvider is a provider that has undergone a verification or auditing process to ensure that it meets certain standards or requirements by an auditor.
An auditor can be any valid account on-chain.
NOTE: There are certain teams providing auditing services, which should be accounted for when deploying.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `auditor` | [string](#string) |  | Auditor is the account bech32 address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes holds a list of key-value pairs of provider attributes. Attributes are arbitrary values that a provider exposes. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/audit/v1/event.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/audit/v1/event.proto
 

 
 <a name="virtengine.audit.v1.EventTrustedAuditorCreated"></a>

 ### EventTrustedAuditorCreated
 EventTrustedAuditorCreated defines an SDK message for when a trusted auditor is created.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

 
 <a name="virtengine.audit.v1.EventTrustedAuditorDeleted"></a>

 ### EventTrustedAuditorDeleted
 EventTrustedAuditorDeleted defines an event for when a trusted auditor is deleted.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/audit/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/audit/v1/genesis.proto
 

 
 <a name="virtengine.audit.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the basic genesis state used by audit module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `providers` | [AuditedProvider](#virtengine.audit.v1.AuditedProvider) | repeated | Providers contains a list of audited providers account addresses. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/audit/v1/msg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/audit/v1/msg.proto
 

 
 <a name="virtengine.audit.v1.MsgDeleteProviderAttributes"></a>

 ### MsgDeleteProviderAttributes
 MsgDeleteProviderAttributes defined the Msg/DeleteProviderAttributes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 | `keys` | [string](#string) | repeated | Keys holds a list of keys of audited provider attributes to delete from the audit. |
 
 

 

 
 <a name="virtengine.audit.v1.MsgDeleteProviderAttributesResponse"></a>

 ### MsgDeleteProviderAttributesResponse
 MsgDeleteProviderAttributesResponse defines the Msg/ProviderAttributes response type.

 

 

 
 <a name="virtengine.audit.v1.MsgSignProviderAttributes"></a>

 ### MsgSignProviderAttributes
 MsgSignProviderAttributes defines an SDK message for signing a provider attributes.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes holds a list of key-value pairs of provider attributes to be audited. Attributes are arbitrary values that a provider exposes. |
 
 

 

 
 <a name="virtengine.audit.v1.MsgSignProviderAttributesResponse"></a>

 ### MsgSignProviderAttributesResponse
 MsgSignProviderAttributesResponse defines the Msg/CreateProvider response type.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/audit/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/audit/v1/query.proto
 

 
 <a name="virtengine.audit.v1.QueryAllProvidersAttributesRequest"></a>

 ### QueryAllProvidersAttributesRequest
 QueryAllProvidersAttributesRequest is request type for the Query/All Providers RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.audit.v1.QueryAuditorAttributesRequest"></a>

 ### QueryAuditorAttributesRequest
 QueryAuditorAttributesRequest is request type for the Query/Providers RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProviderAttributesRequest"></a>

 ### QueryProviderAttributesRequest
 QueryProviderAttributesRequest is request type for the Query/Provider RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProviderAuditorRequest"></a>

 ### QueryProviderAuditorRequest
 QueryProviderAuditorRequest is request type for the Query/Providers RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProviderRequest"></a>

 ### QueryProviderRequest
 QueryProviderRequest is request type for the Query/Provider RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "ve1..." |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProvidersResponse"></a>

 ### QueryProvidersResponse
 QueryProvidersResponse is response type for the Query/Providers RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `providers` | [AuditedProvider](#virtengine.audit.v1.AuditedProvider) | repeated | Providers contains a list of audited providers account addresses. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is used to paginate results. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.audit.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the audit package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `AllProvidersAttributes` | [QueryAllProvidersAttributesRequest](#virtengine.audit.v1.QueryAllProvidersAttributesRequest) | [QueryProvidersResponse](#virtengine.audit.v1.QueryProvidersResponse) | AllProvidersAttributes queries all providers. buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/audit/v1/audit/attributes/list|
 | `ProviderAttributes` | [QueryProviderAttributesRequest](#virtengine.audit.v1.QueryProviderAttributesRequest) | [QueryProvidersResponse](#virtengine.audit.v1.QueryProvidersResponse) | ProviderAttributes queries all provider signed attributes. buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/audit/v1/audit/attributes/{owner}/list|
 | `ProviderAuditorAttributes` | [QueryProviderAuditorRequest](#virtengine.audit.v1.QueryProviderAuditorRequest) | [QueryProvidersResponse](#virtengine.audit.v1.QueryProvidersResponse) | ProviderAuditorAttributes queries provider signed attributes by specific auditor. buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/audit/v1/audit/attributes/{auditor}/{owner}|
 | `AuditorAttributes` | [QueryAuditorAttributesRequest](#virtengine.audit.v1.QueryAuditorAttributesRequest) | [QueryProvidersResponse](#virtengine.audit.v1.QueryProvidersResponse) | AuditorAttributes queries all providers signed by this auditor. buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/provider/v1/auditor/{auditor}/list|
 
  <!-- end services -->

 
 
 <a name="virtengine/audit/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/audit/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.audit.v1.Msg"></a>

 ### Msg
 Msg defines the audit Msg service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `SignProviderAttributes` | [MsgSignProviderAttributes](#virtengine.audit.v1.MsgSignProviderAttributes) | [MsgSignProviderAttributesResponse](#virtengine.audit.v1.MsgSignProviderAttributesResponse) | SignProviderAttributes defines a method that signs provider attributes. | |
 | `DeleteProviderAttributes` | [MsgDeleteProviderAttributes](#virtengine.audit.v1.MsgDeleteProviderAttributes) | [MsgDeleteProviderAttributesResponse](#virtengine.audit.v1.MsgDeleteProviderAttributesResponse) | DeleteProviderAttributes defines a method that deletes provider attributes. | |
 
  <!-- end services -->

 
 
 <a name="virtengine/base/deposit/v1/deposit.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/deposit/v1/deposit.proto
 

 
 <a name="virtengine.base.deposit.v1.Deposit"></a>

 ### Deposit
 Deposit is a data type used by MsgCreateDeployment, MsgDepositDeployment and MsgCreateBid to indicate source of the deposit.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | amount specifies the amount of coins to include in the deployment's first deposit. |
 | `sources` | [Source](#virtengine.base.deposit.v1.Source) | repeated | Sources is the set of deposit sources, each entry must be unique. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.base.deposit.v1.Source"></a>

 ### Source
 Source is an enum which lists source of funds for deployment deposit.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | balance | 1 | SourceBalance denotes account balance as source of funds |
 | grant | 2 | SourceGrant denotes authz grants as source of funds |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/offchain/sign/v1/sign.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/offchain/sign/v1/sign.proto
 

 
 <a name="virtengine.base.offchain.sign.v1.MsgSignData"></a>

 ### MsgSignData
 MsgSignData defines an arbitrary, general-purpose, off-chain message

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `signer` | [string](#string) |  | Signer is the sdk.AccAddress of the message signer |
 | `data` | [bytes](#bytes) |  | Data represents the raw bytes of the content that is signed (text, json, etc) |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/resourcevalue.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/resourcevalue.proto
 

 
 <a name="virtengine.base.resources.v1beta4.ResourceValue"></a>

 ### ResourceValue
 Unit stores cpu, memory and storage metrics.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `val` | [bytes](#bytes) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/cpu.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/cpu.proto
 

 
 <a name="virtengine.base.resources.v1beta4.CPU"></a>

 ### CPU
 CPU stores resource units and cpu config attributes.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `units` | [ResourceValue](#virtengine.base.resources.v1beta4.ResourceValue) |  | Units of the CPU, which represents the number of CPUs available. This field is required and must be a non-negative integer. |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes holds a list of key-value attributes that describe the GPU, such as its model, memory and interface. This field is required and must be a list of `Attribute` messages. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/endpoint.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/endpoint.proto
 

 
 <a name="virtengine.base.resources.v1beta4.Endpoint"></a>

 ### Endpoint
 Endpoint describes a publicly accessible IP service.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `kind` | [Endpoint.Kind](#virtengine.base.resources.v1beta4.Endpoint.Kind) |  | Kind describes how the endpoint is implemented when the lease is deployed. |
 | `sequence_number` | [uint32](#uint32) |  | SequenceNumber represents a sequence number for the Endpoint. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.base.resources.v1beta4.Endpoint.Kind"></a>

 ### Endpoint.Kind
 Kind describes how the endpoint is implemented when the lease is deployed.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | SHARED_HTTP | 0 | Describes an endpoint that becomes a Kubernetes Ingress. |
 | RANDOM_PORT | 1 | Describes an endpoint that becomes a Kubernetes NodePort. |
 | LEASED_IP | 2 | Describes an endpoint that becomes a leased IP. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/gpu.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/gpu.proto
 

 
 <a name="virtengine.base.resources.v1beta4.GPU"></a>

 ### GPU
 GPU stores resource units and gpu configuration attributes.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `units` | [ResourceValue](#virtengine.base.resources.v1beta4.ResourceValue) |  | The resource value of the GPU, which represents the number of GPUs available. This field is required and must be a non-negative integer. |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/memory.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/memory.proto
 

 
 <a name="virtengine.base.resources.v1beta4.Memory"></a>

 ### Memory
 Memory stores resource quantity and memory attributes.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `quantity` | [ResourceValue](#virtengine.base.resources.v1beta4.ResourceValue) |  | Quantity of memory available, which represents the amount of memory in bytes. This field is required and must be a non-negative integer. |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes that describe the memory, such as its type and speed. This field is required and must be a list of Attribute key-values. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/storage.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/storage.proto
 

 
 <a name="virtengine.base.resources.v1beta4.Storage"></a>

 ### Storage
 Storage stores resource quantity and storage attributes.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  | Name holds an arbitrary name for the storage resource. |
 | `quantity` | [ResourceValue](#virtengine.base.resources.v1beta4.ResourceValue) |  | Quantity of storage available, which represents the amount of memory in bytes. This field is required and must be a non-negative integer. |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes that describe the storage. This field is required and must be a list of Attribute key-values. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/base/resources/v1beta4/resources.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/base/resources/v1beta4/resources.proto
 

 
 <a name="virtengine.base.resources.v1beta4.Resources"></a>

 ### Resources
 Resources describes all available resources types for deployment/node etc
if field is nil resource is not present in the given data-structure

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [uint32](#uint32) |  | Id is a unique identifier for the resources. |
 | `cpu` | [CPU](#virtengine.base.resources.v1beta4.CPU) |  | CPU resources available, including the architecture, number of cores and other details. This field is optional and can be empty if no CPU resources are available. |
 | `memory` | [Memory](#virtengine.base.resources.v1beta4.Memory) |  | Memory resources available, including the quantity and attributes. This field is optional and can be empty if no memory resources are available. |
 | `storage` | [Storage](#virtengine.base.resources.v1beta4.Storage) | repeated | Storage resources available, including the quantity and attributes. This field is optional and can be empty if no storage resources are available. |
 | `gpu` | [GPU](#virtengine.base.resources.v1beta4.GPU) |  | GPU resources available, including the type, architecture and other details. This field is optional and can be empty if no GPU resources are available. |
 | `endpoints` | [Endpoint](#virtengine.base.resources.v1beta4.Endpoint) | repeated | Endpoint resources available |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/benchmark/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/benchmark/v1/tx.proto
 

 
 <a name="virtengine.benchmark.v1.AnomalyDetectedEvent"></a>

 ### AnomalyDetectedEvent
 AnomalyDetectedEvent is emitted when an anomaly is detected

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `anomaly_type` | [string](#string) |  |  |
 | `severity` | [string](#string) |  |  |
 | `detected_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.AnomalyFlag"></a>

 ### AnomalyFlag
 AnomalyFlag represents a provider anomaly flag

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `reporter` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `evidence` | [string](#string) |  |  |
 | `status` | [string](#string) |  |  |
 | `flagged_at` | [int64](#int64) |  |  |
 | `resolution` | [string](#string) |  |  |
 | `resolved_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.AnomalyResolvedEvent"></a>

 ### AnomalyResolvedEvent
 AnomalyResolvedEvent is emitted when an anomaly is resolved

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `resolution` | [string](#string) |  |  |
 | `resolved_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.BenchmarkResult"></a>

 ### BenchmarkResult
 BenchmarkResult represents a benchmark result

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `benchmark_type` | [string](#string) |  |  |
 | `score` | [string](#string) |  |  |
 | `timestamp` | [int64](#int64) |  |  |
 | `hardware_info` | [string](#string) |  |  |
 | `raw_data` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.BenchmarksPrunedEvent"></a>

 ### BenchmarksPrunedEvent
 BenchmarksPrunedEvent is emitted when benchmarks are pruned

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `pruned_count` | [uint32](#uint32) |  |  |
 | `pruned_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.BenchmarksSubmittedEvent"></a>

 ### BenchmarksSubmittedEvent
 BenchmarksSubmittedEvent is emitted when benchmarks are submitted

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `result_count` | [uint32](#uint32) |  |  |
 | `submitted_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.Challenge"></a>

 ### Challenge
 Challenge represents a benchmark challenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  |  |
 | `requester` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `benchmark_type` | [string](#string) |  |  |
 | `status` | [string](#string) |  |  |
 | `requested_at` | [int64](#int64) |  |  |
 | `expires_at` | [int64](#int64) |  |  |
 | `response` | [BenchmarkResult](#virtengine.benchmark.v1.BenchmarkResult) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ChallengeCompletedEvent"></a>

 ### ChallengeCompletedEvent
 ChallengeCompletedEvent is emitted when a challenge is completed

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `passed` | [bool](#bool) |  |  |
 | `completed_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ChallengeExpiredEvent"></a>

 ### ChallengeExpiredEvent
 ChallengeExpiredEvent is emitted when a challenge expires

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `expired_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ChallengeRequestedEvent"></a>

 ### ChallengeRequestedEvent
 ChallengeRequestedEvent is emitted when a challenge is requested

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  |  |
 | `requester` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `benchmark_type` | [string](#string) |  |  |
 | `requested_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.GenesisState"></a>

 ### GenesisState
 GenesisState is the genesis state for the benchmark module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.benchmark.v1.Params) |  |  |
 | `provider_benchmarks` | [ProviderBenchmark](#virtengine.benchmark.v1.ProviderBenchmark) | repeated |  |
 | `challenges` | [Challenge](#virtengine.benchmark.v1.Challenge) | repeated |  |
 | `anomaly_flags` | [AnomalyFlag](#virtengine.benchmark.v1.AnomalyFlag) | repeated |  |
 | `challenge_sequence` | [uint64](#uint64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgFlagProvider"></a>

 ### MsgFlagProvider
 MsgFlagProvider flags a provider for anomaly

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reporter` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `evidence` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgFlagProviderResponse"></a>

 ### MsgFlagProviderResponse
 MsgFlagProviderResponse is the response for MsgFlagProvider

 

 

 
 <a name="virtengine.benchmark.v1.MsgRequestChallenge"></a>

 ### MsgRequestChallenge
 MsgRequestChallenge requests a benchmark challenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `requester` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `benchmark_type` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgRequestChallengeResponse"></a>

 ### MsgRequestChallengeResponse
 MsgRequestChallengeResponse is the response for MsgRequestChallenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgResolveAnomalyFlag"></a>

 ### MsgResolveAnomalyFlag
 MsgResolveAnomalyFlag resolves an anomaly flag

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 | `resolution` | [string](#string) |  |  |
 | `is_valid` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgResolveAnomalyFlagResponse"></a>

 ### MsgResolveAnomalyFlagResponse
 MsgResolveAnomalyFlagResponse is the response for MsgResolveAnomalyFlag

 

 

 
 <a name="virtengine.benchmark.v1.MsgRespondChallenge"></a>

 ### MsgRespondChallenge
 MsgRespondChallenge responds to a benchmark challenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `challenge_id` | [string](#string) |  |  |
 | `result` | [BenchmarkResult](#virtengine.benchmark.v1.BenchmarkResult) |  |  |
 | `signature` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgRespondChallengeResponse"></a>

 ### MsgRespondChallengeResponse
 MsgRespondChallengeResponse is the response for MsgRespondChallenge

 

 

 
 <a name="virtengine.benchmark.v1.MsgSubmitBenchmarks"></a>

 ### MsgSubmitBenchmarks
 MsgSubmitBenchmarks submits benchmark results

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `results` | [BenchmarkResult](#virtengine.benchmark.v1.BenchmarkResult) | repeated |  |
 | `signature` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgSubmitBenchmarksResponse"></a>

 ### MsgSubmitBenchmarksResponse
 MsgSubmitBenchmarksResponse is the response for MsgSubmitBenchmarks

 

 

 
 <a name="virtengine.benchmark.v1.MsgUnflagProvider"></a>

 ### MsgUnflagProvider
 MsgUnflagProvider removes a provider flag

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `provider` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.MsgUnflagProviderResponse"></a>

 ### MsgUnflagProviderResponse
 MsgUnflagProviderResponse is the response for MsgUnflagProvider

 

 

 
 <a name="virtengine.benchmark.v1.Params"></a>

 ### Params
 Params defines the parameters for the benchmark module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_timeout` | [uint64](#uint64) |  |  |
 | `benchmark_validity_period` | [uint64](#uint64) |  |  |
 | `min_benchmark_score` | [string](#string) |  |  |
 | `max_anomaly_flags` | [uint32](#uint32) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ProviderBenchmark"></a>

 ### ProviderBenchmark
 ProviderBenchmark represents a provider's benchmark data

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `results` | [BenchmarkResult](#virtengine.benchmark.v1.BenchmarkResult) | repeated |  |
 | `reliability_score` | [string](#string) |  |  |
 | `last_updated` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ProviderFlaggedEvent"></a>

 ### ProviderFlaggedEvent
 ProviderFlaggedEvent is emitted when a provider is flagged

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `reporter` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `flagged_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ProviderUnflaggedEvent"></a>

 ### ProviderUnflaggedEvent
 ProviderUnflaggedEvent is emitted when a provider is unflagged

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `unflagged_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.benchmark.v1.ReliabilityScoreUpdatedEvent"></a>

 ### ReliabilityScoreUpdatedEvent
 ReliabilityScoreUpdatedEvent is emitted when reliability score is updated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `old_score` | [string](#string) |  |  |
 | `new_score` | [string](#string) |  |  |
 | `updated_at` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.benchmark.v1.Msg"></a>

 ### Msg
 Msg defines the benchmark Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `SubmitBenchmarks` | [MsgSubmitBenchmarks](#virtengine.benchmark.v1.MsgSubmitBenchmarks) | [MsgSubmitBenchmarksResponse](#virtengine.benchmark.v1.MsgSubmitBenchmarksResponse) | SubmitBenchmarks submits benchmark results | |
 | `RequestChallenge` | [MsgRequestChallenge](#virtengine.benchmark.v1.MsgRequestChallenge) | [MsgRequestChallengeResponse](#virtengine.benchmark.v1.MsgRequestChallengeResponse) | RequestChallenge requests a benchmark challenge | |
 | `RespondChallenge` | [MsgRespondChallenge](#virtengine.benchmark.v1.MsgRespondChallenge) | [MsgRespondChallengeResponse](#virtengine.benchmark.v1.MsgRespondChallengeResponse) | RespondChallenge responds to a benchmark challenge | |
 | `FlagProvider` | [MsgFlagProvider](#virtengine.benchmark.v1.MsgFlagProvider) | [MsgFlagProviderResponse](#virtengine.benchmark.v1.MsgFlagProviderResponse) | FlagProvider flags a provider for anomaly | |
 | `UnflagProvider` | [MsgUnflagProvider](#virtengine.benchmark.v1.MsgUnflagProvider) | [MsgUnflagProviderResponse](#virtengine.benchmark.v1.MsgUnflagProviderResponse) | UnflagProvider removes a provider flag | |
 | `ResolveAnomalyFlag` | [MsgResolveAnomalyFlag](#virtengine.benchmark.v1.MsgResolveAnomalyFlag) | [MsgResolveAnomalyFlagResponse](#virtengine.benchmark.v1.MsgResolveAnomalyFlagResponse) | ResolveAnomalyFlag resolves an anomaly flag | |
 
  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/types.proto
 

 
 <a name="virtengine.bme.v1.BurnMintPair"></a>

 ### BurnMintPair
 BurnMintPair represents a pair of burn and mint operations with their respective prices

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `burned` | [CoinPrice](#virtengine.bme.v1.CoinPrice) |  | burned is the coin burned |
 | `minted` | [CoinPrice](#virtengine.bme.v1.CoinPrice) |  | minted is coin minted |
 
 

 

 
 <a name="virtengine.bme.v1.CoinPrice"></a>

 ### CoinPrice
 CoinPrice represents a coin amount with its associated oracle price at a specific point in time

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | coin is the token amount |
 | `price` | [string](#string) |  | price (at oracle) of the coin at burn/mint event |
 
 

 

 
 <a name="virtengine.bme.v1.CollateralRatio"></a>

 ### CollateralRatio
 CollateralRatio represents the current collateral ratio

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `ratio` | [string](#string) |  | ratio is CR = (VaultAKT * Price) / OutstandingACT |
 | `status` | [MintStatus](#virtengine.bme.v1.MintStatus) |  | status indicates the current circuit breaker status |
 | `reference_price` | [string](#string) |  | reference_price is the price used to calculate CR |
 
 

 

 
 <a name="virtengine.bme.v1.LedgerID"></a>

 ### LedgerID
 LedgerID uniquely identifies a ledger entry by block height and sequence number

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `height` | [int64](#int64) |  | height is the block height when the ledger entry was created |
 | `sequence` | [int64](#int64) |  | sequence is the sequence number within the block (for ordering) |
 
 

 

 
 <a name="virtengine.bme.v1.LedgerPendingRecord"></a>

 ### LedgerPendingRecord
 LedgerPendingRecord

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | owner source of the coins to be burned |
 | `to` | [string](#string) |  | to destination of the minted coins. if minted coin is ACT, "to" must be same as signer |
 | `coins_to_burn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | coins_to_burn |
 | `denom_to_mint` | [string](#string) |  | denom_to_mint |
 
 

 

 
 <a name="virtengine.bme.v1.LedgerRecord"></a>

 ### LedgerRecord
 LedgerRecord stores information of burn/mint event of token A burn to mint token B

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `burned_from` | [string](#string) |  | burned_from source address of the tokens burned |
 | `minted_to` | [string](#string) |  | minted_to destination address of the tokens minted |
 | `burner` | [string](#string) |  | module is module account performing burn |
 | `minter` | [string](#string) |  | module is module account performing mint |
 | `burned` | [CoinPrice](#virtengine.bme.v1.CoinPrice) |  | burned is the coin burned at price |
 | `minted` | [CoinPrice](#virtengine.bme.v1.CoinPrice) |  | minted is coin minted at price |
 | `remint_credit_issued` | [CoinPrice](#virtengine.bme.v1.CoinPrice) |  |  |
 | `remint_credit_accrued` | [CoinPrice](#virtengine.bme.v1.CoinPrice) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.LedgerRecordID"></a>

 ### LedgerRecordID
 LedgerRecordID

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | denom is the asset denomination |
 | `to_denom` | [string](#string) |  | to_denom is what denom swap to |
 | `source` | [string](#string) |  |  |
 | `height` | [int64](#int64) |  |  |
 | `sequence` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.MintEpoch"></a>

 ### MintEpoch
 MintEpoch stores information about mint epoch

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `next_epoch` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.State"></a>

 ### State
 State tracks net burn metrics since BME start

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `balances` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | burned is the cumulative burn for tracked tokens |
 | `total_burned` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | burned is the cumulative burn for tracked tokens |
 | `total_minted` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | minted is the cumulative mint back for tracked tokens |
 | `remint_credits` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | remint_credits tracks available credits for reminting tokens (e.g., from previous burns that can be reminted without additional collateral) |
 
 

 

 
 <a name="virtengine.bme.v1.Status"></a>

 ### Status
 Status stores status of mint operations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `status` | [MintStatus](#virtengine.bme.v1.MintStatus) |  |  |
 | `previous_status` | [MintStatus](#virtengine.bme.v1.MintStatus) |  |  |
 | `epoch_height_diff` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.bme.v1.LedgerRecordStatus"></a>

 ### LedgerRecordStatus
 LedgerRecordStatus indicates the current state of a burn/mint ledger record

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | ledger_record_status_invalid | 0 | LEDGER_RECORD_STATUS_INVALID is the default/uninitialized value This status should never appear in a valid ledger record |
 | ledger_record_status_pending | 1 | LEDGER_RECORD_STATUS_PENDING indicates a burn/mint operation has been initiated but not yet executed (e.g., waiting for oracle price or circuit breaker clearance) |
 | ledger_record_status_executed | 2 | LEDGER_RECORD_STATUS_EXECUTED indicates the burn/mint operation has been successfully completed and tokens have been burned and minted |
 

 
 <a name="virtengine.bme.v1.MintStatus"></a>

 ### MintStatus
 MintStatus indicates the current state of mint

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | mint_status_unspecified | 0 | MINT_STATUS_UNSPECIFIED is the default value |
 | mint_status_healthy | 1 | MINT_STATUS_HEALTHY indicates normal operation (CR > warn threshold) |
 | mint_status_warning | 2 | MINT_STATUS_WARNING indicates CR is below warning threshold |
 | mint_status_halt_cr | 3 | MINT_STATUS_HALT_CR indicates CR is below halt threshold, mints paused |
 | mint_status_halt_oracle | 4 | MINT_STATUS_HALT_ORACLE indicates circuit breaker tripped due to unhealthy oracle price |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/events.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/events.proto
 

 
 <a name="virtengine.bme.v1.EventLedgerRecordExecuted"></a>

 ### EventLedgerRecordExecuted
 EventLedgerRecordExecuted emitted information of burn/mint event of token A burn to mint token B

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID) |  | burned_from source address of the tokens burned |
 
 

 

 
 <a name="virtengine.bme.v1.EventMintStatusChange"></a>

 ### EventMintStatusChange
 EventCircuitBreakerStatusChange is emitted when circuit breaker status changes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `previous_status` | [MintStatus](#virtengine.bme.v1.MintStatus) |  | previous_status is the previous status |
 | `new_status` | [MintStatus](#virtengine.bme.v1.MintStatus) |  | new_status is the new status |
 | `collateral_ratio` | [string](#string) |  | collateral_ratio is the CR that triggered the change |
 
 

 

 
 <a name="virtengine.bme.v1.EventVaultSeeded"></a>

 ### EventVaultSeeded
 EventVaultSeeded is emitted when the vault is seeded with AKT

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | amount is the AKT amount added to vault |
 | `source` | [string](#string) |  | source is where the funds came from |
 | `new_vault_balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | new_vault_balance is the new vault balance |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/params.proto
 

 
 <a name="virtengine.bme.v1.Params"></a>

 ### Params
 Params defines the parameters for the BME module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `circuit_breaker_warn_threshold` | [uint32](#uint32) |  | circuit_breaker_warn_threshold is the CR below which warning is triggered Stored as basis points * 100 (e.g., 9500 = 0.95) |
 | `circuit_breaker_halt_threshold` | [uint32](#uint32) |  | circuit_breaker_halt_threshold is the CR below which mints are halted Stored as basis points * 100 (e.g., 9000 = 0.90) |
 | `min_epoch_blocks` | [int64](#int64) |  | min_epoch_blocks is the minimum amount of blocks required for ACT mints |
 | `epoch_blocks_backoff` | [uint32](#uint32) |  | epoch_blocks_backoff increase of runway_blocks in % during warn threshold for drop in 1 basis point of circuit_breaker_warn_threshold Stored as basis points * 100 (e.g., 9500 = 0.95) e.g: runway_blocks = 100 min_runway_blocks_backoff = 1000 circuit_breaker_warn_threshold drops from 0.95 to 0.94 then runway_blocks = (100*0.1 + 100) = 110

 circuit_breaker_warn_threshold drops from 0.94 to 0.92 then runway_blocks = (110*(0.1*2) + 110) = 132 |
 | `mint_spread_bps` | [uint32](#uint32) |  | mint_spread_bps is the spread in basis points applied during ACT mint (default: 25 bps = 0.25%) |
 | `settle_spread_bps` | [uint32](#uint32) |  | settle_spread_bps is the spread in basis points applied during settlement (default: 0 for no provider tax) |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/genesis.proto
 

 
 <a name="virtengine.bme.v1.GenesisLedgerPendingRecord"></a>

 ### GenesisLedgerPendingRecord
 GenesisLedgerPendingRecord

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID) |  |  |
 | `record` | [LedgerPendingRecord](#virtengine.bme.v1.LedgerPendingRecord) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.GenesisLedgerRecord"></a>

 ### GenesisLedgerRecord
 GenesisLedgerRecord

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID) |  |  |
 | `record` | [LedgerRecord](#virtengine.bme.v1.LedgerRecord) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.GenesisLedgerState"></a>

 ### GenesisLedgerState
 GenesisLedgerState

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `records` | [GenesisLedgerRecord](#virtengine.bme.v1.GenesisLedgerRecord) | repeated |  |
 | `pending_records` | [GenesisLedgerPendingRecord](#virtengine.bme.v1.GenesisLedgerPendingRecord) | repeated |  |
 
 

 

 
 <a name="virtengine.bme.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the BME module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.bme.v1.Params) |  | params defines the module parameters |
 | `state` | [GenesisVaultState](#virtengine.bme.v1.GenesisVaultState) |  | state is the initial vault state |
 | `ledger` | [GenesisLedgerState](#virtengine.bme.v1.GenesisLedgerState) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.GenesisVaultState"></a>

 ### GenesisVaultState
 GenesisVaultState

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `total_burned` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | burned is the cumulative burn for tracked tokens |
 | `total_minted` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | minted is the cumulative mint back for tracked tokens |
 | `remint_credits` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | remint_credits tracks available credits for reminting tokens (e.g., from previous burns that can be reminted without additional collateral) |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/msgs.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/msgs.proto
 

 
 <a name="virtengine.bme.v1.MsgBurnACT"></a>

 ### MsgBurnACT
 MsgMintACT defines the message for burning one token to mint another
Allows burning AKT to mint ACT, or burning unused ACT back to AKT

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | owner source of the coins to be burned |
 | `to` | [string](#string) |  | to destination of the minted coins. if minted coin is ACT, "to" must be same as signer |
 | `coins_to_burn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | coins_to_burn |
 
 

 

 
 <a name="virtengine.bme.v1.MsgBurnACTResponse"></a>

 ### MsgBurnACTResponse
 MsgBurnMintResponse is the response type for MsgBurnMint

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID) |  |  |
 | `status` | [LedgerRecordStatus](#virtengine.bme.v1.LedgerRecordStatus) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.MsgBurnMint"></a>

 ### MsgBurnMint
 MsgBurnMint defines the message for burning one token to mint another
Allows burning AKT to mint ACT, or burning unused ACT back to AKT

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | owner source of the coins to be burned |
 | `to` | [string](#string) |  | to destination of the minted coins. if minted coin is ACT, "to" must be same as signer |
 | `coins_to_burn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | coins_to_burn |
 | `denom_to_mint` | [string](#string) |  | denom_to_mint |
 
 

 

 
 <a name="virtengine.bme.v1.MsgBurnMintResponse"></a>

 ### MsgBurnMintResponse
 MsgBurnMintResponse is the response type for MsgBurnMint

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID) |  |  |
 | `status` | [LedgerRecordStatus](#virtengine.bme.v1.LedgerRecordStatus) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.MsgMintACT"></a>

 ### MsgMintACT
 MsgMintACT defines the message for burning one token to mint another
Allows burning AKT to mint ACT, or burning unused ACT back to AKT

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | owner source of the coins to be burned |
 | `to` | [string](#string) |  | to destination of the minted coins. if minted coin is ACT, "to" must be same as signer |
 | `coins_to_burn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | coins_to_burn |
 
 

 

 
 <a name="virtengine.bme.v1.MsgMintACTResponse"></a>

 ### MsgMintACTResponse
 MsgBurnMintResponse is the response type for MsgBurnMint

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LedgerRecordID](#virtengine.bme.v1.LedgerRecordID) |  |  |
 | `status` | [LedgerRecordStatus](#virtengine.bme.v1.LedgerRecordStatus) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.MsgSeedVault"></a>

 ### MsgSeedVault
 MsgSeedVault defines the message for seeding the BME vault with AKT
This is used to provide an initial volatility buffer

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address that controls the module (governance) |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | amount is the AKT amount to seed the vault with |
 | `source` | [string](#string) |  | source is the source of funds (e.g., community pool) |
 
 

 

 
 <a name="virtengine.bme.v1.MsgSeedVaultResponse"></a>

 ### MsgSeedVaultResponse
 MsgSeedVaultResponse is the response type for MsgSeedVault

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `vault_akt` | [string](#string) |  | vault_akt is the new vault AKT balance |
 
 

 

 
 <a name="virtengine.bme.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams defines the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address that controls the module (governance) |
 | `params` | [Params](#virtengine.bme.v1.Params) |  | params defines the updated parameters |
 
 

 

 
 <a name="virtengine.bme.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response type for MsgUpdateParams

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/query.proto
 

 
 <a name="virtengine.bme.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request type for the Query/Params RPC method

 

 

 
 <a name="virtengine.bme.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response type for the Query/Params RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.bme.v1.Params) |  |  |
 
 

 

 
 <a name="virtengine.bme.v1.QueryStatusRequest"></a>

 ### QueryStatusRequest
 QueryStatusRequest is the request type for the circuit breaker status

 

 

 
 <a name="virtengine.bme.v1.QueryStatusResponse"></a>

 ### QueryStatusResponse
 QueryMintStatusResponse is the response type for the circuit breaker status

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `status` | [MintStatus](#virtengine.bme.v1.MintStatus) |  | status is the current circuit breaker status |
 | `collateral_ratio` | [string](#string) |  | collateral_ratio is the current CR |
 | `warn_threshold` | [string](#string) |  | warn_threshold is the warning threshold |
 | `halt_threshold` | [string](#string) |  | halt_threshold is the halt threshold |
 | `mints_allowed` | [bool](#bool) |  | mints_allowed indicates if new ACT mints are allowed |
 | `refunds_allowed` | [bool](#bool) |  | refunds_allowed indicates if ACT refunds are allowed |
 
 

 

 
 <a name="virtengine.bme.v1.QueryVaultStateRequest"></a>

 ### QueryVaultStateRequest
 QueryVaultStateRequest is the request type for the Query/VaultState RPC method

 

 

 
 <a name="virtengine.bme.v1.QueryVaultStateResponse"></a>

 ### QueryVaultStateResponse
 QueryVaultStateResponse is the response type for the Query/VaultState RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `vault_state` | [State](#virtengine.bme.v1.State) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.bme.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the BME module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Params` | [QueryParamsRequest](#virtengine.bme.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.bme.v1.QueryParamsResponse) | Params returns the module parameters | GET|/virtengine/bme/v1/params|
 | `VaultState` | [QueryVaultStateRequest](#virtengine.bme.v1.QueryVaultStateRequest) | [QueryVaultStateResponse](#virtengine.bme.v1.QueryVaultStateResponse) | VaultState returns the current vault state | GET|/virtengine/bme/v1/vault|
 | `Status` | [QueryStatusRequest](#virtengine.bme.v1.QueryStatusRequest) | [QueryStatusResponse](#virtengine.bme.v1.QueryStatusResponse) | Status returns the current circuit breaker status | GET|/virtengine/bme/v1/status|
 
  <!-- end services -->

 
 
 <a name="virtengine/bme/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/bme/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.bme.v1.Msg"></a>

 ### Msg
 Msg defines the BME (Burn/Mint Engine) transaction service.
The BME module manages the burn and mint operations for ACT tokens,
maintaining collateral ratios and enforcing circuit breaker rules.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.bme.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.bme.v1.MsgUpdateParamsResponse) | UpdateParams updates the module parameters. This operation can only be performed through governance proposals. | |
 | `BurnMint` | [MsgBurnMint](#virtengine.bme.v1.MsgBurnMint) | [MsgBurnMintResponse](#virtengine.bme.v1.MsgBurnMintResponse) | BurnMint allows users to burn one token and mint another at current oracle prices. Typically used to burn unused ACT tokens back to AKT. The operation may be delayed or rejected based on circuit breaker status. | |
 | `MintACT` | [MsgMintACT](#virtengine.bme.v1.MsgMintACT) | [MsgMintACTResponse](#virtengine.bme.v1.MsgMintACTResponse) | MintACT mints ACT tokens by burning the specified source token. The mint amount is calculated based on current oracle prices and the collateral ratio. May be halted if circuit breaker is triggered. | |
 | `BurnACT` | [MsgBurnACT](#virtengine.bme.v1.MsgBurnACT) | [MsgBurnACTResponse](#virtengine.bme.v1.MsgBurnACTResponse) | BurnACT burns ACT tokens and mints the specified destination token. The burn operation uses remint credits when available, otherwise requires adequate collateral backing based on oracle prices. | |
 
  <!-- end services -->

 
 
 <a name="virtengine/cert/v1/cert.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/cert/v1/cert.proto
 

 
 <a name="virtengine.cert.v1.Certificate"></a>

 ### Certificate
 Certificate stores state, certificate and it's public key.
The certificate is required for several transactions including deployment of a workload to verify the identity of the tenant and secure the deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [State](#virtengine.cert.v1.State) |  | State is the state of the certificate. CertificateValid denotes state for deployment active. CertificateRevoked denotes state for deployment closed. |
 | `cert` | [bytes](#bytes) |  | Cert holds the bytes of the certificate. |
 | `pubkey` | [bytes](#bytes) |  | PubKey holds the public key of the certificate. |
 
 

 

 
 <a name="virtengine.cert.v1.ID"></a>

 ### ID
 ID stores owner and sequence number.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the certificate. It is a string representing a valid account address.

Example: "ve1..." |
 | `serial` | [string](#string) |  | Serial is a sequence number for the certificate. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.cert.v1.State"></a>

 ### State
 State is an enum which refers to state of the certificate.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | valid | 1 | CertificateValid denotes state for deployment active. |
 | revoked | 2 | CertificateRevoked denotes state for deployment closed. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/cert/v1/filters.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/cert/v1/filters.proto
 

 
 <a name="virtengine.cert.v1.CertificateFilter"></a>

 ### CertificateFilter
 CertificateFilter defines filters used to filter certificates.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the certificate. It is a string representing a valid account address.

Example: "ve1..." |
 | `serial` | [string](#string) |  | Serial is a sequence number for the certificate. |
 | `state` | [string](#string) |  | State is the state of the certificate. CertificateValid denotes state for deployment active. CertificateRevoked denotes state for deployment closed. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/cert/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/cert/v1/genesis.proto
 

 
 <a name="virtengine.cert.v1.GenesisCertificate"></a>

 ### GenesisCertificate
 GenesisCertificate defines certificate entry at genesis.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the certificate. It is a string representing a valid account address.

Example: "ve1..." |
 | `certificate` | [Certificate](#virtengine.cert.v1.Certificate) |  | Certificate holds the certificate. |
 
 

 

 
 <a name="virtengine.cert.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the basic genesis state used by cert module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `certificates` | [GenesisCertificate](#virtengine.cert.v1.GenesisCertificate) | repeated | Certificates is a list of genesis certificates. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/cert/v1/msg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/cert/v1/msg.proto
 

 
 <a name="virtengine.cert.v1.MsgCreateCertificate"></a>

 ### MsgCreateCertificate
 MsgCreateCertificate defines an SDK message for creating certificate.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the certificate. It is a string representing a valid account address.

Example: "ve1..." |
 | `cert` | [bytes](#bytes) |  | Cert holds the bytes representing the certificate. |
 | `pubkey` | [bytes](#bytes) |  | PubKey holds the public key. |
 
 

 

 
 <a name="virtengine.cert.v1.MsgCreateCertificateResponse"></a>

 ### MsgCreateCertificateResponse
 MsgCreateCertificateResponse defines the Msg/CreateCertificate response type.

 

 

 
 <a name="virtengine.cert.v1.MsgRevokeCertificate"></a>

 ### MsgRevokeCertificate
 MsgRevokeCertificate defines an SDK message for revoking certificate.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [ID](#virtengine.cert.v1.ID) |  | Id corresponds to the certificate ID which includes owner and sequence number. |
 
 

 

 
 <a name="virtengine.cert.v1.MsgRevokeCertificateResponse"></a>

 ### MsgRevokeCertificateResponse
 MsgRevokeCertificateResponse defines the Msg/RevokeCertificate response type.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/cert/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/cert/v1/query.proto
 

 
 <a name="virtengine.cert.v1.CertificateResponse"></a>

 ### CertificateResponse
 CertificateResponse contains a single X509 certificate and its serial number.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `certificate` | [Certificate](#virtengine.cert.v1.Certificate) |  | Certificate holds the certificate. |
 | `serial` | [string](#string) |  | Serial is a sequence number for the certificate. |
 
 

 

 
 <a name="virtengine.cert.v1.QueryCertificatesRequest"></a>

 ### QueryCertificatesRequest
 QueryDeploymentsRequest is request type for the Query/Deployments RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `filter` | [CertificateFilter](#virtengine.cert.v1.CertificateFilter) |  | Filter allows for filtering of results. |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.cert.v1.QueryCertificatesResponse"></a>

 ### QueryCertificatesResponse
 QueryCertificatesResponse is response type for the Query/Certificates RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `certificates` | [CertificateResponse](#virtengine.cert.v1.CertificateResponse) | repeated | Certificates is a list of certificate. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.cert.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for certificates.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Certificates` | [QueryCertificatesRequest](#virtengine.cert.v1.QueryCertificatesRequest) | [QueryCertificatesResponse](#virtengine.cert.v1.QueryCertificatesResponse) | Certificates queries certificates on-chain. | GET|/virtengine/cert/v1/certificates/list|
 
  <!-- end services -->

 
 
 <a name="virtengine/cert/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/cert/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.cert.v1.Msg"></a>

 ### Msg
 Msg defines the provider Msg service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `CreateCertificate` | [MsgCreateCertificate](#virtengine.cert.v1.MsgCreateCertificate) | [MsgCreateCertificateResponse](#virtengine.cert.v1.MsgCreateCertificateResponse) | CreateCertificate defines a method to create new certificate given proper inputs. | |
 | `RevokeCertificate` | [MsgRevokeCertificate](#virtengine.cert.v1.MsgRevokeCertificate) | [MsgRevokeCertificateResponse](#virtengine.cert.v1.MsgRevokeCertificateResponse) | RevokeCertificate defines a method to revoke the certificate. | |
 
  <!-- end services -->

 
 
 <a name="virtengine/config/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/config/v1/tx.proto
 

 
 <a name="virtengine.config.v1.ApprovedClient"></a>

 ### ApprovedClient
 ApprovedClient represents an approved client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `public_key` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `version_constraint` | [string](#string) |  |  |
 | `allowed_scopes` | [string](#string) | repeated |  |
 | `status` | [string](#string) |  |  |
 | `registered_at` | [int64](#int64) |  |  |
 | `updated_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.EventClientReactivated"></a>

 ### EventClientReactivated
 EventClientReactivated is emitted when a client is reactivated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `reactivated_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.EventClientRegistered"></a>

 ### EventClientRegistered
 EventClientRegistered is emitted when a client is registered

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `registered_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.EventClientRevoked"></a>

 ### EventClientRevoked
 EventClientRevoked is emitted when a client is revoked

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `revoked_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.EventClientSuspended"></a>

 ### EventClientSuspended
 EventClientSuspended is emitted when a client is suspended

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `suspended_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.EventClientUpdated"></a>

 ### EventClientUpdated
 EventClientUpdated is emitted when a client is updated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `updated_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.EventSignatureVerified"></a>

 ### EventSignatureVerified
 EventSignatureVerified is emitted when a signature is verified

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  |  |
 | `signer` | [string](#string) |  |  |
 | `verified_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.GenesisState"></a>

 ### GenesisState
 GenesisState is the genesis state for the config module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.config.v1.Params) |  |  |
 | `approved_clients` | [ApprovedClient](#virtengine.config.v1.ApprovedClient) | repeated |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgReactivateApprovedClient"></a>

 ### MsgReactivateApprovedClient
 MsgReactivateApprovedClient reactivates a suspended client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `client_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgReactivateApprovedClientResponse"></a>

 ### MsgReactivateApprovedClientResponse
 MsgReactivateApprovedClientResponse is the response for MsgReactivateApprovedClient

 

 

 
 <a name="virtengine.config.v1.MsgRegisterApprovedClient"></a>

 ### MsgRegisterApprovedClient
 MsgRegisterApprovedClient registers a new approved client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `client_id` | [string](#string) |  |  |
 | `public_key` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `version_constraint` | [string](#string) |  |  |
 | `allowed_scopes` | [string](#string) | repeated |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgRegisterApprovedClientResponse"></a>

 ### MsgRegisterApprovedClientResponse
 MsgRegisterApprovedClientResponse is the response for MsgRegisterApprovedClient

 

 

 
 <a name="virtengine.config.v1.MsgRevokeApprovedClient"></a>

 ### MsgRevokeApprovedClient
 MsgRevokeApprovedClient revokes an approved client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `client_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgRevokeApprovedClientResponse"></a>

 ### MsgRevokeApprovedClientResponse
 MsgRevokeApprovedClientResponse is the response for MsgRevokeApprovedClient

 

 

 
 <a name="virtengine.config.v1.MsgSuspendApprovedClient"></a>

 ### MsgSuspendApprovedClient
 MsgSuspendApprovedClient suspends an approved client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `client_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgSuspendApprovedClientResponse"></a>

 ### MsgSuspendApprovedClientResponse
 MsgSuspendApprovedClientResponse is the response for MsgSuspendApprovedClient

 

 

 
 <a name="virtengine.config.v1.MsgUpdateApprovedClient"></a>

 ### MsgUpdateApprovedClient
 MsgUpdateApprovedClient updates an approved client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `client_id` | [string](#string) |  |  |
 | `public_key` | [string](#string) |  |  |
 | `version_constraint` | [string](#string) |  |  |
 | `allowed_scopes` | [string](#string) | repeated |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgUpdateApprovedClientResponse"></a>

 ### MsgUpdateApprovedClientResponse
 MsgUpdateApprovedClientResponse is the response for MsgUpdateApprovedClient

 

 

 
 <a name="virtengine.config.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams updates module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `params` | [Params](#virtengine.config.v1.Params) |  |  |
 
 

 

 
 <a name="virtengine.config.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

 
 <a name="virtengine.config.v1.Params"></a>

 ### Params
 Params defines the parameters for the config module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_clients` | [uint64](#uint64) |  |  |
 | `signature_validity_period` | [uint64](#uint64) |  |  |
 | `require_client_signature` | [bool](#bool) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.config.v1.Msg"></a>

 ### Msg
 Msg defines the config Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `RegisterApprovedClient` | [MsgRegisterApprovedClient](#virtengine.config.v1.MsgRegisterApprovedClient) | [MsgRegisterApprovedClientResponse](#virtengine.config.v1.MsgRegisterApprovedClientResponse) | RegisterApprovedClient registers a new approved client | |
 | `UpdateApprovedClient` | [MsgUpdateApprovedClient](#virtengine.config.v1.MsgUpdateApprovedClient) | [MsgUpdateApprovedClientResponse](#virtengine.config.v1.MsgUpdateApprovedClientResponse) | UpdateApprovedClient updates an approved client | |
 | `SuspendApprovedClient` | [MsgSuspendApprovedClient](#virtengine.config.v1.MsgSuspendApprovedClient) | [MsgSuspendApprovedClientResponse](#virtengine.config.v1.MsgSuspendApprovedClientResponse) | SuspendApprovedClient suspends an approved client | |
 | `RevokeApprovedClient` | [MsgRevokeApprovedClient](#virtengine.config.v1.MsgRevokeApprovedClient) | [MsgRevokeApprovedClientResponse](#virtengine.config.v1.MsgRevokeApprovedClientResponse) | RevokeApprovedClient revokes an approved client | |
 | `ReactivateApprovedClient` | [MsgReactivateApprovedClient](#virtengine.config.v1.MsgReactivateApprovedClient) | [MsgReactivateApprovedClientResponse](#virtengine.config.v1.MsgReactivateApprovedClientResponse) | ReactivateApprovedClient reactivates a suspended client | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.config.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.config.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/delegation/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/delegation/v1/types.proto
 

 
 <a name="virtengine.delegation.v1.Delegation"></a>

 ### Delegation
 Delegation represents a delegation from a delegator to a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 | `shares` | [string](#string) |  | shares is the delegation shares (fixed-point, 18 decimals) |
 | `initial_amount` | [string](#string) |  | initial_amount is the initial delegation amount in base units |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | created_at is when the delegation was created |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | updated_at is when the delegation was last updated |
 | `height` | [int64](#int64) |  | height is the block height when delegation was created |
 
 

 

 
 <a name="virtengine.delegation.v1.DelegatorReward"></a>

 ### DelegatorReward
 DelegatorReward represents rewards for a delegator from a specific validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 | `epoch_number` | [uint64](#uint64) |  | epoch_number is the epoch this reward belongs to |
 | `reward` | [string](#string) |  | reward is the reward amount in base units |
 | `shares_at_epoch` | [string](#string) |  | shares_at_epoch is the delegator's shares at the epoch |
 | `validator_total_shares_at_epoch` | [string](#string) |  | validator_total_shares_at_epoch is the validator's total shares at the epoch |
 | `calculated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | calculated_at is when the reward was calculated |
 | `claimed` | [bool](#bool) |  | claimed indicates if the reward has been claimed |
 | `claimed_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | claimed_at is when the reward was claimed (optional) |
 
 

 

 
 <a name="virtengine.delegation.v1.Params"></a>

 ### Params
 Params defines the parameters for the delegation module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `unbonding_period` | [int64](#int64) |  | unbonding_period is the duration for unbonding in seconds |
 | `max_validators_per_delegator` | [int64](#int64) |  | max_validators_per_delegator is the maximum number of validators a delegator can delegate to |
 | `min_delegation_amount` | [int64](#int64) |  | min_delegation_amount is the minimum delegation amount in base units |
 | `max_redelegations` | [int64](#int64) |  | max_redelegations is the maximum number of simultaneous redelegations |
 | `validator_commission_rate` | [int64](#int64) |  | validator_commission_rate is the validator commission rate in basis points (e.g., 1000 = 10%) |
 | `reward_denom` | [string](#string) |  | reward_denom is the denomination for rewards |
 | `stake_denom` | [string](#string) |  | stake_denom is the denomination for staking |
 
 

 

 
 <a name="virtengine.delegation.v1.Redelegation"></a>

 ### Redelegation
 Redelegation represents a redelegation from one validator to another

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  | id is the unique redelegation ID |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `validator_src_address` | [string](#string) |  | validator_src_address is the source validator's address |
 | `validator_dst_address` | [string](#string) |  | validator_dst_address is the destination validator's address |
 | `entries` | [RedelegationEntry](#virtengine.delegation.v1.RedelegationEntry) | repeated | entries are the redelegation entries |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | created_at is when the redelegation started |
 | `height` | [int64](#int64) |  | height is the block height when redelegation was initiated |
 
 

 

 
 <a name="virtengine.delegation.v1.RedelegationEntry"></a>

 ### RedelegationEntry
 RedelegationEntry represents a single redelegation entry

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `creation_height` | [int64](#int64) |  | creation_height is the height at which the redelegation was created |
 | `completion_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | completion_time is when the redelegation matures |
 | `initial_balance` | [string](#string) |  | initial_balance is the initial balance being redelegated |
 | `shares_dst` | [string](#string) |  | shares_dst is the shares received at the destination validator |
 
 

 

 
 <a name="virtengine.delegation.v1.UnbondingDelegation"></a>

 ### UnbondingDelegation
 UnbondingDelegation represents a delegation that is unbonding

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  | id is the unique unbonding delegation ID |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 | `entries` | [UnbondingDelegationEntry](#virtengine.delegation.v1.UnbondingDelegationEntry) | repeated | entries are the unbonding entries |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | created_at is when the unbonding started |
 | `height` | [int64](#int64) |  | height is the block height when unbonding was initiated |
 
 

 

 
 <a name="virtengine.delegation.v1.UnbondingDelegationEntry"></a>

 ### UnbondingDelegationEntry
 UnbondingDelegationEntry represents a single unbonding entry

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `creation_height` | [int64](#int64) |  | creation_height is the height at which the unbonding was created |
 | `completion_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | completion_time is when the unbonding will complete |
 | `initial_balance` | [string](#string) |  | initial_balance is the initial balance to undelegate |
 | `balance` | [string](#string) |  | balance is the remaining balance to return |
 | `unbonding_shares` | [string](#string) |  | unbonding_shares is the shares being unbonded |
 
 

 

 
 <a name="virtengine.delegation.v1.ValidatorShares"></a>

 ### ValidatorShares
 ValidatorShares represents the total shares for a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 | `total_shares` | [string](#string) |  | total_shares is the total delegation shares (fixed-point, 18 decimals) |
 | `total_stake` | [string](#string) |  | total_stake is the total stake amount in base units |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | updated_at is when the record was last updated |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.delegation.v1.DelegationStatus"></a>

 ### DelegationStatus
 DelegationStatus represents the status of a delegation

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | DELEGATION_STATUS_UNSPECIFIED | 0 | DELEGATION_STATUS_UNSPECIFIED is the default/invalid status |
 | DELEGATION_STATUS_ACTIVE | 1 | DELEGATION_STATUS_ACTIVE means the delegation is active |
 | DELEGATION_STATUS_UNBONDING | 2 | DELEGATION_STATUS_UNBONDING means the delegation is unbonding |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/delegation/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/delegation/v1/genesis.proto
 

 
 <a name="virtengine.delegation.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the delegation module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.delegation.v1.Params) |  | params are the module parameters |
 | `delegations` | [Delegation](#virtengine.delegation.v1.Delegation) | repeated | delegations are the initial delegations |
 | `unbonding_delegations` | [UnbondingDelegation](#virtengine.delegation.v1.UnbondingDelegation) | repeated | unbonding_delegations are the initial unbonding delegations |
 | `redelegations` | [Redelegation](#virtengine.delegation.v1.Redelegation) | repeated | redelegations are the initial redelegations |
 | `validator_shares` | [ValidatorShares](#virtengine.delegation.v1.ValidatorShares) | repeated | validator_shares are the initial validator shares |
 | `delegator_rewards` | [DelegatorReward](#virtengine.delegation.v1.DelegatorReward) | repeated | delegator_rewards are the initial delegator rewards |
 | `delegation_sequence` | [uint64](#uint64) |  | delegation_sequence is the next delegation sequence number |
 | `unbonding_sequence` | [uint64](#uint64) |  | unbonding_sequence is the next unbonding sequence number |
 | `redelegation_sequence` | [uint64](#uint64) |  | redelegation_sequence is the next redelegation sequence number |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/delegation/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/delegation/v1/query.proto
 

 
 <a name="virtengine.delegation.v1.QueryDelegationRequest"></a>

 ### QueryDelegationRequest
 QueryDelegationRequest is the request for QueryDelegation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegationResponse"></a>

 ### QueryDelegationResponse
 QueryDelegationResponse is the response for QueryDelegation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegation` | [Delegation](#virtengine.delegation.v1.Delegation) |  | delegation is the delegation record |
 | `found` | [bool](#bool) |  | found indicates if the delegation was found |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorAllRewardsRequest"></a>

 ### QueryDelegatorAllRewardsRequest
 QueryDelegatorAllRewardsRequest is the request for QueryDelegatorAllRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorAllRewardsResponse"></a>

 ### QueryDelegatorAllRewardsResponse
 QueryDelegatorAllRewardsResponse is the response for QueryDelegatorAllRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `rewards` | [DelegatorReward](#virtengine.delegation.v1.DelegatorReward) | repeated | rewards are the unclaimed reward records |
 | `total_reward` | [string](#string) |  | total_reward is the sum of all unclaimed rewards |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination is the pagination response |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorDelegationsRequest"></a>

 ### QueryDelegatorDelegationsRequest
 QueryDelegatorDelegationsRequest is the request for QueryDelegatorDelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorDelegationsResponse"></a>

 ### QueryDelegatorDelegationsResponse
 QueryDelegatorDelegationsResponse is the response for QueryDelegatorDelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegations` | [Delegation](#virtengine.delegation.v1.Delegation) | repeated | delegations are the delegation records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination is the pagination response |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorRedelegationsRequest"></a>

 ### QueryDelegatorRedelegationsRequest
 QueryDelegatorRedelegationsRequest is the request for QueryDelegatorRedelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorRedelegationsResponse"></a>

 ### QueryDelegatorRedelegationsResponse
 QueryDelegatorRedelegationsResponse is the response for QueryDelegatorRedelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `redelegations` | [Redelegation](#virtengine.delegation.v1.Redelegation) | repeated | redelegations are the redelegation records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination is the pagination response |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorRewardsRequest"></a>

 ### QueryDelegatorRewardsRequest
 QueryDelegatorRewardsRequest is the request for QueryDelegatorRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorRewardsResponse"></a>

 ### QueryDelegatorRewardsResponse
 QueryDelegatorRewardsResponse is the response for QueryDelegatorRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `rewards` | [DelegatorReward](#virtengine.delegation.v1.DelegatorReward) | repeated | rewards are the unclaimed reward records |
 | `total_reward` | [string](#string) |  | total_reward is the sum of all unclaimed rewards |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest"></a>

 ### QueryDelegatorUnbondingDelegationsRequest
 QueryDelegatorUnbondingDelegationsRequest is the request for QueryDelegatorUnbondingDelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator_address` | [string](#string) |  | delegator_address is the delegator's address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse"></a>

 ### QueryDelegatorUnbondingDelegationsResponse
 QueryDelegatorUnbondingDelegationsResponse is the response for QueryDelegatorUnbondingDelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `unbonding_delegations` | [UnbondingDelegation](#virtengine.delegation.v1.UnbondingDelegation) | repeated | unbonding_delegations are the unbonding delegation records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination is the pagination response |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for QueryParams

 

 

 
 <a name="virtengine.delegation.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for QueryParams

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.delegation.v1.Params) |  | params are the module parameters |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryRedelegationRequest"></a>

 ### QueryRedelegationRequest
 QueryRedelegationRequest is the request for QueryRedelegation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `redelegation_id` | [string](#string) |  | redelegation_id is the unique redelegation ID |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryRedelegationResponse"></a>

 ### QueryRedelegationResponse
 QueryRedelegationResponse is the response for QueryRedelegation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `redelegation` | [Redelegation](#virtengine.delegation.v1.Redelegation) |  | redelegation is the redelegation record |
 | `found` | [bool](#bool) |  | found indicates if the redelegation was found |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryUnbondingDelegationRequest"></a>

 ### QueryUnbondingDelegationRequest
 QueryUnbondingDelegationRequest is the request for QueryUnbondingDelegation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `unbonding_id` | [string](#string) |  | unbonding_id is the unique unbonding delegation ID |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryUnbondingDelegationResponse"></a>

 ### QueryUnbondingDelegationResponse
 QueryUnbondingDelegationResponse is the response for QueryUnbondingDelegation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `unbonding_delegation` | [UnbondingDelegation](#virtengine.delegation.v1.UnbondingDelegation) |  | unbonding_delegation is the unbonding delegation record |
 | `found` | [bool](#bool) |  | found indicates if the unbonding delegation was found |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryValidatorDelegationsRequest"></a>

 ### QueryValidatorDelegationsRequest
 QueryValidatorDelegationsRequest is the request for QueryValidatorDelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryValidatorDelegationsResponse"></a>

 ### QueryValidatorDelegationsResponse
 QueryValidatorDelegationsResponse is the response for QueryValidatorDelegations

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegations` | [Delegation](#virtengine.delegation.v1.Delegation) | repeated | delegations are the delegation records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination is the pagination response |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryValidatorSharesRequest"></a>

 ### QueryValidatorSharesRequest
 QueryValidatorSharesRequest is the request for QueryValidatorShares

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | validator_address is the validator's address |
 
 

 

 
 <a name="virtengine.delegation.v1.QueryValidatorSharesResponse"></a>

 ### QueryValidatorSharesResponse
 QueryValidatorSharesResponse is the response for QueryValidatorShares

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_shares` | [ValidatorShares](#virtengine.delegation.v1.ValidatorShares) |  | validator_shares is the validator shares record |
 | `found` | [bool](#bool) |  | found indicates if the validator shares were found |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.delegation.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the delegation module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Params` | [QueryParamsRequest](#virtengine.delegation.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.delegation.v1.QueryParamsResponse) | Params queries the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/params|
 | `Delegation` | [QueryDelegationRequest](#virtengine.delegation.v1.QueryDelegationRequest) | [QueryDelegationResponse](#virtengine.delegation.v1.QueryDelegationResponse) | Delegation queries a specific delegation buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/delegator/{delegator_address}/validator/{validator_address}|
 | `DelegatorDelegations` | [QueryDelegatorDelegationsRequest](#virtengine.delegation.v1.QueryDelegatorDelegationsRequest) | [QueryDelegatorDelegationsResponse](#virtengine.delegation.v1.QueryDelegatorDelegationsResponse) | DelegatorDelegations queries all delegations for a delegator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/delegator/{delegator_address}/delegations|
 | `ValidatorDelegations` | [QueryValidatorDelegationsRequest](#virtengine.delegation.v1.QueryValidatorDelegationsRequest) | [QueryValidatorDelegationsResponse](#virtengine.delegation.v1.QueryValidatorDelegationsResponse) | ValidatorDelegations queries all delegations for a validator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/validator/{validator_address}/delegations|
 | `UnbondingDelegation` | [QueryUnbondingDelegationRequest](#virtengine.delegation.v1.QueryUnbondingDelegationRequest) | [QueryUnbondingDelegationResponse](#virtengine.delegation.v1.QueryUnbondingDelegationResponse) | UnbondingDelegation queries a specific unbonding delegation buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/unbonding/{unbonding_id}|
 | `DelegatorUnbondingDelegations` | [QueryDelegatorUnbondingDelegationsRequest](#virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest) | [QueryDelegatorUnbondingDelegationsResponse](#virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse) | DelegatorUnbondingDelegations queries all unbonding delegations for a delegator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/delegator/{delegator_address}/unbonding|
 | `Redelegation` | [QueryRedelegationRequest](#virtengine.delegation.v1.QueryRedelegationRequest) | [QueryRedelegationResponse](#virtengine.delegation.v1.QueryRedelegationResponse) | Redelegation queries a specific redelegation buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/redelegation/{redelegation_id}|
 | `DelegatorRedelegations` | [QueryDelegatorRedelegationsRequest](#virtengine.delegation.v1.QueryDelegatorRedelegationsRequest) | [QueryDelegatorRedelegationsResponse](#virtengine.delegation.v1.QueryDelegatorRedelegationsResponse) | DelegatorRedelegations queries all redelegations for a delegator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/delegator/{delegator_address}/redelegations|
 | `DelegatorRewards` | [QueryDelegatorRewardsRequest](#virtengine.delegation.v1.QueryDelegatorRewardsRequest) | [QueryDelegatorRewardsResponse](#virtengine.delegation.v1.QueryDelegatorRewardsResponse) | DelegatorRewards queries unclaimed rewards for a delegator from a specific validator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/delegator/{delegator_address}/rewards/{validator_address}|
 | `DelegatorAllRewards` | [QueryDelegatorAllRewardsRequest](#virtengine.delegation.v1.QueryDelegatorAllRewardsRequest) | [QueryDelegatorAllRewardsResponse](#virtengine.delegation.v1.QueryDelegatorAllRewardsResponse) | DelegatorAllRewards queries all unclaimed rewards for a delegator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/delegator/{delegator_address}/rewards|
 | `ValidatorShares` | [QueryValidatorSharesRequest](#virtengine.delegation.v1.QueryValidatorSharesRequest) | [QueryValidatorSharesResponse](#virtengine.delegation.v1.QueryValidatorSharesResponse) | ValidatorShares queries the total shares for a validator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/delegation/v1/validator/{validator_address}/shares|
 
  <!-- end services -->

 
 
 <a name="virtengine/delegation/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/delegation/v1/tx.proto
 

 
 <a name="virtengine.delegation.v1.MsgClaimAllRewards"></a>

 ### MsgClaimAllRewards
 MsgClaimAllRewards claims rewards from all validators

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgClaimAllRewardsResponse"></a>

 ### MsgClaimAllRewardsResponse
 MsgClaimAllRewardsResponse is the response for MsgClaimAllRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `amount` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgClaimRewards"></a>

 ### MsgClaimRewards
 MsgClaimRewards claims rewards from a specific validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator` | [string](#string) |  |  |
 | `validator` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgClaimRewardsResponse"></a>

 ### MsgClaimRewardsResponse
 MsgClaimRewardsResponse is the response for MsgClaimRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `amount` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgDelegate"></a>

 ### MsgDelegate
 MsgDelegate delegates tokens to a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator` | [string](#string) |  |  |
 | `validator` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgDelegateResponse"></a>

 ### MsgDelegateResponse
 MsgDelegateResponse is the response for MsgDelegate

 

 

 
 <a name="virtengine.delegation.v1.MsgRedelegate"></a>

 ### MsgRedelegate
 MsgRedelegate redelegates tokens between validators

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator` | [string](#string) |  |  |
 | `src_validator` | [string](#string) |  |  |
 | `dst_validator` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgRedelegateResponse"></a>

 ### MsgRedelegateResponse
 MsgRedelegateResponse is the response for MsgRedelegate

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `completion_time` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgUndelegate"></a>

 ### MsgUndelegate
 MsgUndelegate undelegates tokens from a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delegator` | [string](#string) |  |  |
 | `validator` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgUndelegateResponse"></a>

 ### MsgUndelegateResponse
 MsgUndelegateResponse is the response for MsgUndelegate

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `completion_time` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams updates module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `params` | [Params](#virtengine.delegation.v1.Params) |  |  |
 
 

 

 
 <a name="virtengine.delegation.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.delegation.v1.Msg"></a>

 ### Msg
 Msg defines the delegation Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Delegate` | [MsgDelegate](#virtengine.delegation.v1.MsgDelegate) | [MsgDelegateResponse](#virtengine.delegation.v1.MsgDelegateResponse) | Delegate delegates tokens to a validator | |
 | `Undelegate` | [MsgUndelegate](#virtengine.delegation.v1.MsgUndelegate) | [MsgUndelegateResponse](#virtengine.delegation.v1.MsgUndelegateResponse) | Undelegate undelegates tokens from a validator | |
 | `Redelegate` | [MsgRedelegate](#virtengine.delegation.v1.MsgRedelegate) | [MsgRedelegateResponse](#virtengine.delegation.v1.MsgRedelegateResponse) | Redelegate redelegates tokens between validators | |
 | `ClaimRewards` | [MsgClaimRewards](#virtengine.delegation.v1.MsgClaimRewards) | [MsgClaimRewardsResponse](#virtengine.delegation.v1.MsgClaimRewardsResponse) | ClaimRewards claims rewards from a specific validator | |
 | `ClaimAllRewards` | [MsgClaimAllRewards](#virtengine.delegation.v1.MsgClaimAllRewards) | [MsgClaimAllRewardsResponse](#virtengine.delegation.v1.MsgClaimAllRewardsResponse) | ClaimAllRewards claims rewards from all validators | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.delegation.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.delegation.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1/deployment.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1/deployment.proto
 

 
 <a name="virtengine.deployment.v1.Deployment"></a>

 ### Deployment
 Deployment stores deploymentID, state and checksum details.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 | `state` | [Deployment.State](#virtengine.deployment.v1.Deployment.State) |  | State defines the sate of the deployment. A deployment can be either active or inactive. |
 | `hash` | [bytes](#bytes) |  | Hash is an hashed representation of the deployment. |
 | `created_at` | [int64](#int64) |  | CreatedAt indicates when the deployment was created as a block height value. |
 
 

 

 
 <a name="virtengine.deployment.v1.DeploymentID"></a>

 ### DeploymentID
 DeploymentID represents a unique identifier for a specific deployment on the network.
It is composed of two fields: an owner address and a sequence number (dseq).

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.deployment.v1.Deployment.State"></a>

 ### Deployment.State
 State is an enum which refers to state of deployment.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | active | 1 | DeploymentActive denotes state for deployment active. |
 | closed | 2 | DeploymentClosed denotes state for deployment closed. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1/group.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1/group.proto
 

 
 <a name="virtengine.deployment.v1.GroupID"></a>

 ### GroupID
 GroupID uniquely identifies a group within a deployment on the network.
A group represents a specific collection of resources or configurations
within a deployment.
It stores owner, deployment sequence number (dseq) and group sequence number (gseq).

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the group. It is a string representing a valid account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1/event.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1/event.proto
 

 
 <a name="virtengine.deployment.v1.EventDeploymentClosed"></a>

 ### EventDeploymentClosed
 EventDeploymentClosed is triggered when deployment is closed on chain.
It contains all the information required to identify a deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1.EventDeploymentCreated"></a>

 ### EventDeploymentCreated
 EventDeploymentCreated event is triggered when deployment is created on chain.
It contains all the information required to identify a deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 | `hash` | [bytes](#bytes) |  | Hash is an hashed representation of the deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1.EventDeploymentUpdated"></a>

 ### EventDeploymentUpdated
 EventDeploymentUpdated is triggered when deployment is updated on chain.
It contains all the information required to identify a deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 | `hash` | [bytes](#bytes) |  | Hash is an hashed representation of the deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1.EventGroupClosed"></a>

 ### EventGroupClosed
 EventGroupClosed is triggered when deployment group is closed.
It contains all the information required to identify a group.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [GroupID](#virtengine.deployment.v1.GroupID) |  | ID is the unique identifier of the group. |
 
 

 

 
 <a name="virtengine.deployment.v1.EventGroupPaused"></a>

 ### EventGroupPaused
 EventGroupPaused is triggered when deployment group is paused.
It contains all the information required to identify a group.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [GroupID](#virtengine.deployment.v1.GroupID) |  | ID is the unique identifier of the group. |
 
 

 

 
 <a name="virtengine.deployment.v1.EventGroupStarted"></a>

 ### EventGroupStarted
 EventGroupStarted is triggered when deployment group is started.
It contains all the information required to identify a group.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [GroupID](#virtengine.deployment.v1.GroupID) |  | ID is the unique identifier of the group. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/resourceunit.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/resourceunit.proto
 

 
 <a name="virtengine.deployment.v1beta4.ResourceUnit"></a>

 ### ResourceUnit
 ResourceUnit extends Resources and adds Count along with the Price.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `resource` | [virtengine.base.resources.v1beta4.Resources](#virtengine.base.resources.v1beta4.Resources) |  | Resource holds the amount of resources. |
 | `count` | [uint32](#uint32) |  | Count corresponds to the amount of replicas to run of the resources. |
 | `price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Price holds the pricing for the resource units. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/groupspec.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/groupspec.proto
 

 
 <a name="virtengine.deployment.v1beta4.GroupSpec"></a>

 ### GroupSpec
 GroupSpec defines a specification for a group in a deployment on the network.
This includes attributes like the group name, placement requirements, and resource units.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  | Name is the name of the group. |
 | `requirements` | [virtengine.base.attributes.v1.PlacementRequirements](#virtengine.base.attributes.v1.PlacementRequirements) |  | Requirements specifies the placement requirements for the group. This determines where the resources in the group can be deployed. |
 | `resources` | [ResourceUnit](#virtengine.deployment.v1beta4.ResourceUnit) | repeated | Resources is a list containing the resource units allocated to the group. Each ResourceUnit defines the specific resources (e.g., CPU, memory) assigned. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/deploymentmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/deploymentmsg.proto
 

 
 <a name="virtengine.deployment.v1beta4.MsgCloseDeployment"></a>

 ### MsgCloseDeployment
 MsgCloseDeployment defines an SDK message for closing deployment

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgCloseDeploymentResponse"></a>

 ### MsgCloseDeploymentResponse
 MsgCloseDeploymentResponse defines the Msg/CloseDeployment response type.

 

 

 
 <a name="virtengine.deployment.v1beta4.MsgCreateDeployment"></a>

 ### MsgCreateDeployment
 MsgCreateDeployment defines an SDK message for creating deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 | `groups` | [GroupSpec](#virtengine.deployment.v1beta4.GroupSpec) | repeated | GroupSpec is a list of group specifications for the deployment. This field is required and must be a list of GroupSpec. |
 | `hash` | [bytes](#bytes) |  | Hash of the deployment. |
 | `deposit` | [virtengine.base.deposit.v1.Deposit](#virtengine.base.deposit.v1.Deposit) |  | Deposit specifies the amount of coins to include in the deployment's first deposit. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgCreateDeploymentResponse"></a>

 ### MsgCreateDeploymentResponse
 MsgCreateDeploymentResponse defines the Msg/CreateDeployment response type.

 

 

 
 <a name="virtengine.deployment.v1beta4.MsgUpdateDeployment"></a>

 ### MsgUpdateDeployment
 MsgUpdateDeployment defines an SDK message for updating deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | ID is the unique identifier of the deployment. |
 | `hash` | [bytes](#bytes) |  | Hash of the deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgUpdateDeploymentResponse"></a>

 ### MsgUpdateDeploymentResponse
 MsgUpdateDeploymentResponse defines the Msg/UpdateDeployment response type.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/filters.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/filters.proto
 

 
 <a name="virtengine.deployment.v1beta4.DeploymentFilters"></a>

 ### DeploymentFilters
 DeploymentFilters defines filters used to filter deployments.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `state` | [string](#string) |  | State defines the sate of the deployment. A deployment can be either active or inactive. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.GroupFilters"></a>

 ### GroupFilters
 GroupFilters defines filters used to filter groups

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the group. It is a string representing a valid account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint64](#uint64) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `state` | [string](#string) |  | State defines the sate of the deployment. A deployment can be either active or inactive. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/group.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/group.proto
 

 
 <a name="virtengine.deployment.v1beta4.Group"></a>

 ### Group
 Group stores group id, state and specifications of a group.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.GroupID](#virtengine.deployment.v1.GroupID) |  | Id is the unique identifier for the group. |
 | `state` | [Group.State](#virtengine.deployment.v1beta4.Group.State) |  | State represents the state of the group. |
 | `group_spec` | [GroupSpec](#virtengine.deployment.v1beta4.GroupSpec) |  | GroupSpec holds the specification of a the Group. |
 | `created_at` | [int64](#int64) |  | CreatedAt is the block height at which the deployment was created. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.deployment.v1beta4.Group.State"></a>

 ### Group.State
 State is an enum which refers to state of group.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | open | 1 | GroupOpen denotes state for group open. |
 | paused | 2 | GroupOrdered denotes state for group ordered. |
 | insufficient_funds | 3 | GroupInsufficientFunds denotes state for group insufficient_funds. |
 | closed | 4 | GroupClosed denotes state for group closed. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/params.proto
 

 
 <a name="virtengine.deployment.v1beta4.Params"></a>

 ### Params
 Params defines the parameters for the x/deployment module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `min_deposits` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | MinDeposits holds a list of the minimum amount of deposits for each a coin. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/genesis.proto
 

 
 <a name="virtengine.deployment.v1beta4.GenesisDeployment"></a>

 ### GenesisDeployment
 GenesisDeployment defines the basic genesis state used by deployment module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `deployment` | [virtengine.deployment.v1.Deployment](#virtengine.deployment.v1.Deployment) |  | Deployments represents a deployment on the network. |
 | `groups` | [Group](#virtengine.deployment.v1beta4.Group) | repeated | Groups is a list of groups within a Deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.GenesisState"></a>

 ### GenesisState
 GenesisState stores slice of genesis deployment instance.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `deployments` | [GenesisDeployment](#virtengine.deployment.v1beta4.GenesisDeployment) | repeated | Deployments is a list of deployments on the network. |
 | `params` | [Params](#virtengine.deployment.v1beta4.Params) |  | Params defines the parameters for the x/deployment module. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/groupmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/groupmsg.proto
 

 
 <a name="virtengine.deployment.v1beta4.MsgCloseGroup"></a>

 ### MsgCloseGroup
 MsgCloseGroup defines SDK message to close a single Group within a Deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.GroupID](#virtengine.deployment.v1.GroupID) |  | Id is the unique identifier of the Group. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgCloseGroupResponse"></a>

 ### MsgCloseGroupResponse
 MsgCloseGroupResponse defines the Msg/CloseGroup response type.

 

 

 
 <a name="virtengine.deployment.v1beta4.MsgPauseGroup"></a>

 ### MsgPauseGroup
 MsgPauseGroup defines SDK message to pause a single Group within a Deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.GroupID](#virtengine.deployment.v1.GroupID) |  | Id is the unique identifier of the Group. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgPauseGroupResponse"></a>

 ### MsgPauseGroupResponse
 MsgPauseGroupResponse defines the Msg/PauseGroup response type.

 

 

 
 <a name="virtengine.deployment.v1beta4.MsgStartGroup"></a>

 ### MsgStartGroup
 MsgStartGroup defines SDK message to start a single Group within a Deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.GroupID](#virtengine.deployment.v1.GroupID) |  | Id is the unique identifier of the Group. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgStartGroupResponse"></a>

 ### MsgStartGroupResponse
 MsgStartGroupResponse defines the Msg/StartGroup response type.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/paramsmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/paramsmsg.proto
 

 
 <a name="virtengine.deployment.v1beta4.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the Msg/UpdateParams request type.

Since: akash v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address of the governance account. |
 | `params` | [Params](#virtengine.deployment.v1beta4.Params) |  | Params defines the x/deployment parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: akash v1.0.0

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/id/v1/id.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/id/v1/id.proto
 

 
 <a name="virtengine.escrow.id.v1.Account"></a>

 ### Account
 Account is the account identifier.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope` | [Scope](#virtengine.escrow.id.v1.Scope) |  |  |
 | `xid` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.escrow.id.v1.Payment"></a>

 ### Payment
 Payment is the payment identifier.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `aid` | [Account](#virtengine.escrow.id.v1.Account) |  |  |
 | `xid` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.escrow.id.v1.Scope"></a>

 ### Scope
 Scope is an enum which refers to the account scope

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | deployment | 1 | DeploymentActive denotes state for deployment active. |
 | bid | 2 | DeploymentClosed denotes state for deployment closed. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/types/v1/balance.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/types/v1/balance.proto
 

 
 <a name="virtengine.escrow.types.v1.Balance"></a>

 ### Balance
 Balance holds the unspent coin received from all deposits with same denom
DecCoin is not being used here as it does not support negative values,
and balance may go negative if account is overdrawn.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  |  |
 | `amount` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/types/v1/deposit.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/types/v1/deposit.proto
 

 
 <a name="virtengine.escrow.types.v1.Depositor"></a>

 ### Depositor
 Depositor stores state of a deposit.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the depositor. It is a string representing a valid account address.

Example: "ve1..." If depositor is same as the owner, then any incoming coins are added to the Balance. If depositor isn't same as the owner, then any incoming coins are added to the Funds. |
 | `height` | [int64](#int64) |  | Height blockchain height at which deposit was created |
 | `source` | [virtengine.base.deposit.v1.Source](#virtengine.base.deposit.v1.Source) |  | Source indicated origination of the funds |
 | `balance` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Balance amount of funds available to spend in this deposit. |
 | `direct` | [bool](#bool) |  | direct indicates if deposited currency should be swapped to ACT (false) at time of the deposit |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/types/v1/state.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/types/v1/state.proto
 

  <!-- end messages -->

 
 <a name="virtengine.escrow.types.v1.State"></a>

 ### State
 State stores state for an escrow account.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | AccountStateInvalid is an invalid state. |
 | open | 1 | StateOpen is the state when an object is open. |
 | closed | 2 | StateClosed is the state when an object is closed. |
 | overdrawn | 3 | StateOverdrawn is the state when an object are overdrawn. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/types/v1/account.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/types/v1/account.proto
 

 
 <a name="virtengine.escrow.types.v1.Account"></a>

 ### Account
 Account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.escrow.id.v1.Account](#virtengine.escrow.id.v1.Account) |  |  |
 | `state` | [AccountState](#virtengine.escrow.types.v1.AccountState) |  |  |
 
 

 

 
 <a name="virtengine.escrow.types.v1.AccountState"></a>

 ### AccountState
 Account stores state for an escrow account.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `state` | [State](#virtengine.escrow.types.v1.State) |  | State represents the current state of an Account. |
 | `transferred` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated | Transferred total coins spent by this account. |
 | `settled_at` | [int64](#int64) |  | SettledAt represents the block height at which this account was last settled. |
 | `funds` | [Balance](#virtengine.escrow.types.v1.Balance) | repeated | Funds holds the unspent coins received from all deposits |
 | `deposits` | [Depositor](#virtengine.escrow.types.v1.Depositor) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/query.proto
 

 
 <a name="virtengine.deployment.v1beta4.QueryDeploymentRequest"></a>

 ### QueryDeploymentRequest
 QueryDeploymentRequest is request type for the Query/Deployment RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.DeploymentID](#virtengine.deployment.v1.DeploymentID) |  | Id is the unique identifier of the deployment. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.QueryDeploymentResponse"></a>

 ### QueryDeploymentResponse
 QueryDeploymentResponse is response type for the Query/Deployment RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `deployment` | [virtengine.deployment.v1.Deployment](#virtengine.deployment.v1.Deployment) |  | Deployment represents a deployment on the network. |
 | `groups` | [Group](#virtengine.deployment.v1beta4.Group) | repeated | Groups is a list of deployment groups. |
 | `escrow_account` | [virtengine.escrow.types.v1.Account](#virtengine.escrow.types.v1.Account) |  | EscrowAccount represents an escrow mechanism where funds are held. This ensures that obligations of both tenants and providers involved in the transaction are met without direct access to each other's accounts. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.QueryDeploymentsRequest"></a>

 ### QueryDeploymentsRequest
 QueryDeploymentsRequest is request type for the Query/Deployments RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `filters` | [DeploymentFilters](#virtengine.deployment.v1beta4.DeploymentFilters) |  | Filters holds the deployment fields to filter the request. |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.QueryDeploymentsResponse"></a>

 ### QueryDeploymentsResponse
 QueryDeploymentsResponse is response type for the Query/Deployments RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `deployments` | [QueryDeploymentResponse](#virtengine.deployment.v1beta4.QueryDeploymentResponse) | repeated | Deployments is a list of Deployments. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.QueryGroupRequest"></a>

 ### QueryGroupRequest
 QueryGroupRequest is request type for the Query/Group RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.deployment.v1.GroupID](#virtengine.deployment.v1.GroupID) |  | Id is the unique identifier of the Group. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.QueryGroupResponse"></a>

 ### QueryGroupResponse
 QueryGroupResponse is response type for the Query/Group RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `group` | [Group](#virtengine.deployment.v1beta4.Group) |  | Group holds a deployment Group. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request type for the Query/Params RPC method.

 

 

 
 <a name="virtengine.deployment.v1beta4.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response type for the Query/Params RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.deployment.v1beta4.Params) |  | params defines the parameters of the module. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.deployment.v1beta4.Query"></a>

 ### Query
 Query defines the gRPC querier service for the Deployments package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Deployments` | [QueryDeploymentsRequest](#virtengine.deployment.v1beta4.QueryDeploymentsRequest) | [QueryDeploymentsResponse](#virtengine.deployment.v1beta4.QueryDeploymentsResponse) | Deployments queries deployments. | GET|/virtengine/deployment/v1beta4/deployments/list|
 | `Deployment` | [QueryDeploymentRequest](#virtengine.deployment.v1beta4.QueryDeploymentRequest) | [QueryDeploymentResponse](#virtengine.deployment.v1beta4.QueryDeploymentResponse) | Deployment queries deployment details. | GET|/virtengine/deployment/v1beta4/deployments/info|
 | `Group` | [QueryGroupRequest](#virtengine.deployment.v1beta4.QueryGroupRequest) | [QueryGroupResponse](#virtengine.deployment.v1beta4.QueryGroupResponse) | Group queries group details. | GET|/virtengine/deployment/v1beta4/groups/info|
 | `Params` | [QueryParamsRequest](#virtengine.deployment.v1beta4.QueryParamsRequest) | [QueryParamsResponse](#virtengine.deployment.v1beta4.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/virtengine/deployment/v1beta4/params|
 
  <!-- end services -->

 
 
 <a name="virtengine/deployment/v1beta4/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/deployment/v1beta4/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.deployment.v1beta4.Msg"></a>

 ### Msg
 Msg defines the x/deployment Msg service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `CreateDeployment` | [MsgCreateDeployment](#virtengine.deployment.v1beta4.MsgCreateDeployment) | [MsgCreateDeploymentResponse](#virtengine.deployment.v1beta4.MsgCreateDeploymentResponse) | CreateDeployment defines a method to create new deployment given proper inputs. | |
 | `UpdateDeployment` | [MsgUpdateDeployment](#virtengine.deployment.v1beta4.MsgUpdateDeployment) | [MsgUpdateDeploymentResponse](#virtengine.deployment.v1beta4.MsgUpdateDeploymentResponse) | UpdateDeployment defines a method to update a deployment given proper inputs. | |
 | `CloseDeployment` | [MsgCloseDeployment](#virtengine.deployment.v1beta4.MsgCloseDeployment) | [MsgCloseDeploymentResponse](#virtengine.deployment.v1beta4.MsgCloseDeploymentResponse) | CloseDeployment defines a method to close a deployment given proper inputs. | |
 | `CloseGroup` | [MsgCloseGroup](#virtengine.deployment.v1beta4.MsgCloseGroup) | [MsgCloseGroupResponse](#virtengine.deployment.v1beta4.MsgCloseGroupResponse) | CloseGroup defines a method to close a group of a deployment given proper inputs. | |
 | `PauseGroup` | [MsgPauseGroup](#virtengine.deployment.v1beta4.MsgPauseGroup) | [MsgPauseGroupResponse](#virtengine.deployment.v1beta4.MsgPauseGroupResponse) | PauseGroup defines a method to pause a group of a deployment given proper inputs. | |
 | `StartGroup` | [MsgStartGroup](#virtengine.deployment.v1beta4.MsgStartGroup) | [MsgStartGroupResponse](#virtengine.deployment.v1beta4.MsgStartGroupResponse) | StartGroup defines a method to start a group of a deployment given proper inputs. | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.deployment.v1beta4.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.deployment.v1beta4.MsgUpdateParamsResponse) | UpdateParams defines a governance operation for updating the x/deployment module parameters. The authority is hard-coded to the x/gov module account.

Since: akash v1.0.0 | |
 
  <!-- end services -->

 
 
 <a name="virtengine/discovery/v1/client_info.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/discovery/v1/client_info.proto
 

 
 <a name="virtengine.discovery.v1.ClientInfo"></a>

 ### ClientInfo
 ClientInfo is the virtengine specific client info.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `api_version` | [string](#string) |  | ApiVersion is the version of the API running on the client. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/discovery/v1/virtengine.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/discovery/v1/virtengine.proto
 

 
 <a name="virtengine.discovery.v1.VirtEngine"></a>

 ### VirtEngine
 VirtEngine virtengine specific RPC parameters.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_info` | [ClientInfo](#virtengine.discovery.v1.ClientInfo) |  | ClientInfo holds information about the client. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/downtimedetector/v1beta1/downtime_duration.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/downtimedetector/v1beta1/downtime_duration.proto
 

  <!-- end messages -->

 
 <a name="virtengine.downtimedetector.v1beta1.Downtime"></a>

 ### Downtime
 Downtime defines the predefined downtime durations that can be tracked
by the downtime detector module to monitor chain availability

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | DURATION_30S | 0 | DURATION_30S represents a 30 second downtime period |
 | DURATION_1M | 1 | DURATION_1M represents a 1 minute downtime period |
 | DURATION_2M | 2 | DURATION_2M represents a 2 minute downtime period |
 | DURATION_3M | 3 | DURATION_3M represents a 3 minute downtime period |
 | DURATION_4M | 4 | DURATION_4M represents a 4 minute downtime period |
 | DURATION_5M | 5 | DURATION_5M represents a 5 minute downtime period |
 | DURATION_10M | 6 | DURATION_10M represents a 10 minute downtime period |
 | DURATION_20M | 7 | DURATION_20M represents a 20 minute downtime period |
 | DURATION_30M | 8 | DURATION_30M represents a 30 minute downtime period |
 | DURATION_40M | 9 | DURATION_40M represents a 40 minute downtime period |
 | DURATION_50M | 10 | DURATION_50M represents a 50 minute downtime period |
 | DURATION_1H | 11 | DURATION_1H represents a 1 hour downtime period |
 | DURATION_1_5H | 12 | DURATION_1_5H represents a 1.5 hour downtime period |
 | DURATION_2H | 13 | DURATION_2H represents a 2 hour downtime period |
 | DURATION_2_5H | 14 | DURATION_2_5H represents a 2.5 hour downtime period |
 | DURATION_3H | 15 | DURATION_3H represents a 3 hour downtime period |
 | DURATION_4H | 16 | DURATION_4H represents a 4 hour downtime period |
 | DURATION_5H | 17 | DURATION_5H represents a 5 hour downtime period |
 | DURATION_6H | 18 | DURATION_6H represents a 6 hour downtime period |
 | DURATION_9H | 19 | DURATION_9H represents a 9 hour downtime period |
 | DURATION_12H | 20 | DURATION_12H represents a 12 hour downtime period |
 | DURATION_18H | 21 | DURATION_18H represents a 18 hour downtime period |
 | DURATION_24H | 22 | DURATION_24H represents a 24 hour downtime period |
 | DURATION_36H | 23 | DURATION_36H represents a 36 hour downtime period |
 | DURATION_48H | 24 | DURATION_48H represents a 48 hour downtime period |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/downtimedetector/v1beta1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/downtimedetector/v1beta1/genesis.proto
 

 
 <a name="virtengine.downtimedetector.v1beta1.GenesisDowntimeEntry"></a>

 ### GenesisDowntimeEntry
 GenesisDowntimeEntry tracks the last occurrence of a specific downtime duration

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `duration` | [Downtime](#virtengine.downtimedetector.v1beta1.Downtime) |  | duration is the downtime period being tracked |
 | `last_downtime` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | last_downtime is the timestamp when this downtime duration was last observed |
 
 

 

 
 <a name="virtengine.downtimedetector.v1beta1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the downtime detector module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `downtimes` | [GenesisDowntimeEntry](#virtengine.downtimedetector.v1beta1.GenesisDowntimeEntry) | repeated | downtimes is the list of tracked downtime entries |
 | `last_block_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | last_block_time is the timestamp of the last processed block |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/downtimedetector/v1beta1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/downtimedetector/v1beta1/query.proto
 

 
 <a name="virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest"></a>

 ### RecoveredSinceDowntimeOfLengthRequest
 RecoveredSinceDowntimeOfLengthRequest is the request type for querying if the chain
has been operational for at least the specified recovery duration since experiencing
downtime of the specified length

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `downtime` | [Downtime](#virtengine.downtimedetector.v1beta1.Downtime) |  | downtime is the downtime duration to check against |
 | `recovery` | [google.protobuf.Duration](#google.protobuf.Duration) |  | recovery is the minimum recovery duration required since the downtime |
 
 

 

 
 <a name="virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse"></a>

 ### RecoveredSinceDowntimeOfLengthResponse
 RecoveredSinceDowntimeOfLengthResponse is the response type for the recovery query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `succesfully_recovered` | [bool](#bool) |  | succesfully_recovered indicates if the chain has been up for at least the recovery duration since the last downtime of the specified length |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.downtimedetector.v1beta1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the downtime detector module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `RecoveredSinceDowntimeOfLength` | [RecoveredSinceDowntimeOfLengthRequest](#virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest) | [RecoveredSinceDowntimeOfLengthResponse](#virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse) | RecoveredSinceDowntimeOfLength queries if the chain has recovered for a specified duration since experiencing downtime of a given length | GET|/virtengine/downtime-detector/v1beta1/RecoveredSinceDowntimeOfLength|
 
  <!-- end services -->

 
 
 <a name="virtengine/enclave/v1/events.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/enclave/v1/events.proto
 

 
 <a name="virtengine.enclave.v1.EventConsensusVerificationFailed"></a>

 ### EventConsensusVerificationFailed
 EventConsensusVerificationFailed is emitted when consensus verification fails

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the identity scope |
 | `proposed_score` | [uint32](#uint32) |  | ProposedScore is the score proposed |
 | `computed_score` | [uint32](#uint32) |  | ComputedScore is the locally recomputed score |
 | `score_difference` | [int32](#int32) |  | ScoreDifference is the absolute difference between scores |
 | `reason` | [string](#string) |  | Reason is the failure reason |
 
 

 

 
 <a name="virtengine.enclave.v1.EventEnclaveIdentityExpired"></a>

 ### EventEnclaveIdentityExpired
 EventEnclaveIdentityExpired is emitted when an enclave identity expires

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator` | [string](#string) |  | Validator is the validator address |
 | `expiry_height` | [int64](#int64) |  | ExpiryHeight is the height at which the identity expired |
 
 

 

 
 <a name="virtengine.enclave.v1.EventEnclaveIdentityRegistered"></a>

 ### EventEnclaveIdentityRegistered
 EventEnclaveIdentityRegistered is emitted when an enclave identity is registered

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator` | [string](#string) |  | Validator is the validator address |
 | `tee_type` | [string](#string) |  | TEEType is the type of TEE |
 | `measurement_hash` | [string](#string) |  | MeasurementHash is the enclave measurement hash (hex-encoded) |
 | `encryption_key_id` | [string](#string) |  | EncryptionKeyID is the encryption key identifier |
 | `signing_key_id` | [string](#string) |  | SigningKeyID is the signing key identifier |
 | `epoch` | [uint64](#uint64) |  | Epoch is the registration epoch |
 | `expiry_height` | [int64](#int64) |  | ExpiryHeight is the block height when this identity expires |
 
 

 

 
 <a name="virtengine.enclave.v1.EventEnclaveIdentityRevoked"></a>

 ### EventEnclaveIdentityRevoked
 EventEnclaveIdentityRevoked is emitted when an enclave identity is revoked

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator` | [string](#string) |  | Validator is the validator address |
 | `reason` | [string](#string) |  | Reason is the reason for revocation |
 
 

 

 
 <a name="virtengine.enclave.v1.EventEnclaveIdentityUpdated"></a>

 ### EventEnclaveIdentityUpdated
 EventEnclaveIdentityUpdated is emitted when an enclave identity is updated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator` | [string](#string) |  | Validator is the validator address |
 | `status` | [string](#string) |  | Status is the new status |
 
 

 

 
 <a name="virtengine.enclave.v1.EventEnclaveKeyRotated"></a>

 ### EventEnclaveKeyRotated
 EventEnclaveKeyRotated is emitted when a key rotation is initiated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator` | [string](#string) |  | Validator is the validator address |
 | `old_key_fingerprint` | [string](#string) |  | OldKeyFingerprint is the fingerprint of the old key |
 | `new_key_fingerprint` | [string](#string) |  | NewKeyFingerprint is the fingerprint of the new key |
 | `overlap_start_height` | [int64](#int64) |  | OverlapStartHeight is when both keys become valid |
 | `overlap_end_height` | [int64](#int64) |  | OverlapEndHeight is when the old key becomes invalid |
 
 

 

 
 <a name="virtengine.enclave.v1.EventKeyRotationCompleted"></a>

 ### EventKeyRotationCompleted
 EventKeyRotationCompleted is emitted when a key rotation completes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator` | [string](#string) |  | Validator is the validator address |
 | `new_key_fingerprint` | [string](#string) |  | NewKeyFingerprint is the fingerprint of the new active key |
 
 

 

 
 <a name="virtengine.enclave.v1.EventMeasurementAdded"></a>

 ### EventMeasurementAdded
 EventMeasurementAdded is emitted when a measurement is added to the allowlist

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurement_hash` | [string](#string) |  | MeasurementHash is the measurement hash (hex-encoded) |
 | `tee_type` | [string](#string) |  | TEEType is the TEE type |
 | `description` | [string](#string) |  | Description is the measurement description |
 | `min_isv_svn` | [uint32](#uint32) |  | MinISVSVN is the minimum security version |
 
 

 

 
 <a name="virtengine.enclave.v1.EventMeasurementRevoked"></a>

 ### EventMeasurementRevoked
 EventMeasurementRevoked is emitted when a measurement is revoked

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurement_hash` | [string](#string) |  | MeasurementHash is the measurement hash (hex-encoded) |
 | `reason` | [string](#string) |  | Reason is the reason for revocation |
 
 

 

 
 <a name="virtengine.enclave.v1.EventVEIDScoreComputedAttested"></a>

 ### EventVEIDScoreComputedAttested
 EventVEIDScoreComputedAttested is emitted when a VEID score is computed with attestation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the identity scope |
 | `account_address` | [string](#string) |  | AccountAddress is the account that owns the identity |
 | `score` | [uint32](#uint32) |  | Score is the computed score |
 | `status` | [string](#string) |  | Status is the verification status |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height |
 
 

 

 
 <a name="virtengine.enclave.v1.EventVEIDScoreRejectedAttestation"></a>

 ### EventVEIDScoreRejectedAttestation
 EventVEIDScoreRejectedAttestation is emitted when a VEID score fails attestation verification

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the identity scope |
 | `account_address` | [string](#string) |  | AccountAddress is the account that owns the identity |
 | `proposed_score` | [uint32](#uint32) |  | ProposedScore is the score proposed by the proposer |
 | `computed_score` | [uint32](#uint32) |  | ComputedScore is the locally computed score |
 | `reason` | [string](#string) |  | Reason is the reason for rejection |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/enclave/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/enclave/v1/types.proto
 

 
 <a name="virtengine.enclave.v1.AttestedScoringResult"></a>

 ### AttestedScoringResult
 AttestedScoringResult represents an enclave-attested scoring output

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the identity scope that was scored |
 | `account_address` | [string](#string) |  | AccountAddress is the account that owns the identity |
 | `score` | [uint32](#uint32) |  | Score is the computed identity score (0-100) |
 | `status` | [string](#string) |  | Status is the verification status |
 | `reason_codes` | [string](#string) | repeated | ReasonCodes are structured reason codes for the score |
 | `model_version_hash` | [bytes](#bytes) |  | ModelVersionHash is the hash of the ML model used |
 | `input_hash` | [bytes](#bytes) |  | InputHash is the hash of the input data (for determinism verification) |
 | `evidence_hashes` | [bytes](#bytes) | repeated | EvidenceHashes are hashes of evidence artifacts (face embeddings, OCR, etc.) |
 | `enclave_measurement_hash` | [bytes](#bytes) |  | EnclaveMeasurementHash is the measurement of the enclave that computed this |
 | `enclave_signature` | [bytes](#bytes) |  | EnclaveSignature is the signature from the enclave signing key |
 | `attestation_reference` | [bytes](#bytes) |  | AttestationReference is a reference to the attestation quote (hash or ID) |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator that produced this result |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height where this result was produced |
 | `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp is when this result was computed |
 
 

 

 
 <a name="virtengine.enclave.v1.EnclaveIdentity"></a>

 ### EnclaveIdentity
 EnclaveIdentity represents a validator's enclave identity record

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator operator address |
 | `tee_type` | [TEEType](#virtengine.enclave.v1.TEEType) |  | TEEType is the type of TEE (SGX, SEV-SNP, NITRO) |
 | `measurement_hash` | [bytes](#bytes) |  | MeasurementHash is the enclave measurement (MRENCLAVE for SGX) |
 | `signer_hash` | [bytes](#bytes) |  | SignerHash is the signer measurement (MRSIGNER for SGX) |
 | `encryption_pub_key` | [bytes](#bytes) |  | EncryptionPubKey is the enclave's public key for encryption |
 | `signing_pub_key` | [bytes](#bytes) |  | SigningPubKey is the enclave's public key for signing attestations |
 | `attestation_quote` | [bytes](#bytes) |  | AttestationQuote is the raw attestation quote from the TEE |
 | `attestation_chain` | [bytes](#bytes) | repeated | AttestationChain is the certificate chain for attestation verification |
 | `isv_prod_id` | [uint32](#uint32) |  | ISVProdID is the Independent Software Vendor Product ID |
 | `isv_svn` | [uint32](#uint32) |  | ISVSVN is the Independent Software Vendor Security Version Number |
 | `quote_version` | [uint32](#uint32) |  | QuoteVersion is the attestation quote format version |
 | `debug_mode` | [bool](#bool) |  | DebugMode indicates if the enclave is in debug mode (must be false for production) |
 | `epoch` | [uint64](#uint64) |  | Epoch is the registration epoch |
 | `expiry_height` | [int64](#int64) |  | ExpiryHeight is the block height when this identity expires |
 | `registered_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | RegisteredAt is the timestamp when this identity was registered |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | UpdatedAt is the timestamp when this identity was last updated |
 | `status` | [EnclaveIdentityStatus](#virtengine.enclave.v1.EnclaveIdentityStatus) |  | Status is the current status of the enclave identity |
 
 

 

 
 <a name="virtengine.enclave.v1.KeyRotationRecord"></a>

 ### KeyRotationRecord
 KeyRotationRecord represents a key rotation event

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator operator address |
 | `epoch` | [uint64](#uint64) |  | Epoch is the epoch when rotation was initiated |
 | `old_key_fingerprint` | [string](#string) |  | OldKeyFingerprint is the fingerprint of the old key |
 | `new_key_fingerprint` | [string](#string) |  | NewKeyFingerprint is the fingerprint of the new key |
 | `overlap_start_height` | [int64](#int64) |  | OverlapStartHeight is when both keys become valid |
 | `overlap_end_height` | [int64](#int64) |  | OverlapEndHeight is when the old key becomes invalid |
 | `initiated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | InitiatedAt is when the rotation was initiated |
 | `completed_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | CompletedAt is when the rotation was completed (old key invalidated) |
 | `status` | [KeyRotationStatus](#virtengine.enclave.v1.KeyRotationStatus) |  | Status is the current status of the rotation |
 
 

 

 
 <a name="virtengine.enclave.v1.MeasurementRecord"></a>

 ### MeasurementRecord
 MeasurementRecord represents an approved enclave measurement in the allowlist

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurement_hash` | [bytes](#bytes) |  | MeasurementHash is the enclave measurement hash |
 | `tee_type` | [TEEType](#virtengine.enclave.v1.TEEType) |  | TEEType is the TEE type this measurement is for |
 | `description` | [string](#string) |  | Description is a human-readable description |
 | `min_isv_svn` | [uint32](#uint32) |  | MinISVSVN is the minimum required security version |
 | `added_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | AddedAt is when this measurement was added |
 | `added_by_proposal` | [uint64](#uint64) |  | AddedByProposal is the governance proposal ID that added this measurement |
 | `expiry_height` | [int64](#int64) |  | ExpiryHeight is when this measurement expires (0 for no expiry) |
 | `revoked` | [bool](#bool) |  | Revoked indicates if this measurement has been revoked |
 | `revoked_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | RevokedAt is when this measurement was revoked (if applicable) |
 | `revoked_by_proposal` | [uint64](#uint64) |  | RevokedByProposal is the governance proposal that revoked this (if applicable) |
 
 

 

 
 <a name="virtengine.enclave.v1.Params"></a>

 ### Params
 Params defines the parameters for the enclave module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_enclave_keys_per_validator` | [uint32](#uint32) |  | MaxEnclaveKeysPerValidator is the maximum number of enclave keys a validator can have |
 | `default_expiry_blocks` | [int64](#int64) |  | DefaultExpiryBlocks is the default number of blocks until enclave identity expires |
 | `key_rotation_overlap_blocks` | [int64](#int64) |  | KeyRotationOverlapBlocks is the default overlap period for key rotations |
 | `min_quote_version` | [uint32](#uint32) |  | MinQuoteVersion is the minimum attestation quote version required |
 | `allowed_tee_types` | [TEEType](#virtengine.enclave.v1.TEEType) | repeated | AllowedTEETypes is the list of allowed TEE types |
 | `score_tolerance` | [uint32](#uint32) |  | ScoreTolerance is the maximum allowed score difference for consensus |
 | `require_attestation_chain` | [bool](#bool) |  | RequireAttestationChain indicates if attestation chain verification is required |
 | `max_attestation_age` | [int64](#int64) |  | MaxAttestationAge is the maximum age of attestation in blocks |
 | `enable_committee_mode` | [bool](#bool) |  | EnableCommitteeMode enables committee-based identity processing |
 | `committee_size` | [uint32](#uint32) |  | CommitteeSize is the size of the identity committee (if committee mode enabled) |
 | `committee_epoch_blocks` | [int64](#int64) |  | CommitteeEpochBlocks is the number of blocks per committee epoch |
 | `enable_measurement_cleanup` | [bool](#bool) |  | EnableMeasurementCleanup enables automatic cleanup of expired measurements |
 | `max_registrations_per_block` | [uint32](#uint32) |  | MaxRegistrationsPerBlock limits registrations per block (0 = unlimited) |
 | `registration_cooldown_blocks` | [int64](#int64) |  | RegistrationCooldownBlocks enforces cooldown between re-registrations |
 
 

 

 
 <a name="virtengine.enclave.v1.ValidatorKeyInfo"></a>

 ### ValidatorKeyInfo
 ValidatorKeyInfo contains key information for a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator operator address |
 | `encryption_key_id` | [string](#string) |  | EncryptionKeyID is the key identifier |
 | `encryption_pub_key` | [bytes](#bytes) |  | EncryptionPubKey is the public key for encryption |
 | `measurement_hash` | [bytes](#bytes) |  | MeasurementHash is the enclave measurement hash |
 | `expiry_height` | [int64](#int64) |  | ExpiryHeight is when the identity expires |
 | `is_in_rotation` | [bool](#bool) |  | IsInRotation indicates if key rotation is in progress |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.enclave.v1.EnclaveIdentityStatus"></a>

 ### EnclaveIdentityStatus
 EnclaveIdentityStatus represents the status of an enclave identity

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | ENCLAVE_IDENTITY_STATUS_UNSPECIFIED | 0 | ENCLAVE_IDENTITY_STATUS_UNSPECIFIED is the default/invalid status |
 | ENCLAVE_IDENTITY_STATUS_ACTIVE | 1 | ENCLAVE_IDENTITY_STATUS_ACTIVE indicates the enclave identity is active |
 | ENCLAVE_IDENTITY_STATUS_PENDING | 2 | ENCLAVE_IDENTITY_STATUS_PENDING indicates the enclave identity is pending verification |
 | ENCLAVE_IDENTITY_STATUS_EXPIRED | 3 | ENCLAVE_IDENTITY_STATUS_EXPIRED indicates the enclave identity has expired |
 | ENCLAVE_IDENTITY_STATUS_REVOKED | 4 | ENCLAVE_IDENTITY_STATUS_REVOKED indicates the enclave identity has been revoked |
 | ENCLAVE_IDENTITY_STATUS_ROTATING | 5 | ENCLAVE_IDENTITY_STATUS_ROTATING indicates key rotation is in progress |
 

 
 <a name="virtengine.enclave.v1.KeyRotationStatus"></a>

 ### KeyRotationStatus
 KeyRotationStatus represents the status of a key rotation

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | KEY_ROTATION_STATUS_UNSPECIFIED | 0 | KEY_ROTATION_STATUS_UNSPECIFIED is the default/invalid status |
 | KEY_ROTATION_STATUS_PENDING | 1 | KEY_ROTATION_STATUS_PENDING indicates rotation is pending |
 | KEY_ROTATION_STATUS_ACTIVE | 2 | KEY_ROTATION_STATUS_ACTIVE indicates rotation is active (overlap period) |
 | KEY_ROTATION_STATUS_COMPLETED | 3 | KEY_ROTATION_STATUS_COMPLETED indicates rotation is completed |
 | KEY_ROTATION_STATUS_CANCELLED | 4 | KEY_ROTATION_STATUS_CANCELLED indicates rotation was cancelled |
 

 
 <a name="virtengine.enclave.v1.TEEType"></a>

 ### TEEType
 TEEType represents the type of Trusted Execution Environment

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | TEE_TYPE_UNSPECIFIED | 0 | TEE_TYPE_UNSPECIFIED is the default/invalid TEE type |
 | TEE_TYPE_SGX | 1 | TEE_TYPE_SGX is Intel SGX |
 | TEE_TYPE_SEV_SNP | 2 | TEE_TYPE_SEV_SNP is AMD SEV-SNP |
 | TEE_TYPE_NITRO | 3 | TEE_TYPE_NITRO is AWS Nitro Enclaves |
 | TEE_TYPE_TRUSTZONE | 4 | TEE_TYPE_TRUSTZONE is ARM TrustZone (future) |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/enclave/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/enclave/v1/genesis.proto
 

 
 <a name="virtengine.enclave.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the enclave module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `enclave_identities` | [EnclaveIdentity](#virtengine.enclave.v1.EnclaveIdentity) | repeated | EnclaveIdentities are the registered enclave identities |
 | `measurement_allowlist` | [MeasurementRecord](#virtengine.enclave.v1.MeasurementRecord) | repeated | MeasurementAllowlist is the list of approved enclave measurements |
 | `key_rotations` | [KeyRotationRecord](#virtengine.enclave.v1.KeyRotationRecord) | repeated | KeyRotations is the list of active key rotations |
 | `params` | [Params](#virtengine.enclave.v1.Params) |  | Params are the module parameters |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/enclave/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/enclave/v1/query.proto
 

 
 <a name="virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysRequest"></a>

 ### QueryActiveValidatorEnclaveKeysRequest
 QueryActiveValidatorEnclaveKeysRequest is the request for QueryActiveValidatorEnclaveKeys

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysResponse"></a>

 ### QueryActiveValidatorEnclaveKeysResponse
 QueryActiveValidatorEnclaveKeysResponse is the response for QueryActiveValidatorEnclaveKeys

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `identities` | [EnclaveIdentity](#virtengine.enclave.v1.EnclaveIdentity) | repeated | Identities are the active enclave identities |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryAttestedResultRequest"></a>

 ### QueryAttestedResultRequest
 QueryAttestedResultRequest is the request for QueryAttestedResult

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height of the result |
 | `scope_id` | [string](#string) |  | ScopeID is the scope ID of the result |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryAttestedResultResponse"></a>

 ### QueryAttestedResultResponse
 QueryAttestedResultResponse is the response for QueryAttestedResult

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `result` | [AttestedScoringResult](#virtengine.enclave.v1.AttestedScoringResult) |  | Result is the attested scoring result |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryCommitteeEnclaveKeysRequest"></a>

 ### QueryCommitteeEnclaveKeysRequest
 QueryCommitteeEnclaveKeysRequest is the request for QueryCommitteeEnclaveKeys

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `committee_epoch` | [uint64](#uint64) |  | CommitteeEpoch is the epoch for which to get committee keys |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryCommitteeEnclaveKeysResponse"></a>

 ### QueryCommitteeEnclaveKeysResponse
 QueryCommitteeEnclaveKeysResponse is the response for QueryCommitteeEnclaveKeys

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `identities` | [EnclaveIdentity](#virtengine.enclave.v1.EnclaveIdentity) | repeated | Identities are the committee enclave identities |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryEnclaveIdentityRequest"></a>

 ### QueryEnclaveIdentityRequest
 QueryEnclaveIdentityRequest is the request for QueryEnclaveIdentity

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator address to query |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryEnclaveIdentityResponse"></a>

 ### QueryEnclaveIdentityResponse
 QueryEnclaveIdentityResponse is the response for QueryEnclaveIdentity

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `identity` | [EnclaveIdentity](#virtengine.enclave.v1.EnclaveIdentity) |  | Identity is the enclave identity for the validator |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryKeyRotationRequest"></a>

 ### QueryKeyRotationRequest
 QueryKeyRotationRequest is the request for QueryKeyRotation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator address to query |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryKeyRotationResponse"></a>

 ### QueryKeyRotationResponse
 QueryKeyRotationResponse is the response for QueryKeyRotation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `rotation` | [KeyRotationRecord](#virtengine.enclave.v1.KeyRotationRecord) |  | Rotation is the key rotation record |
 | `has_active_rotation` | [bool](#bool) |  | HasActiveRotation indicates if there is an active rotation |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryMeasurementAllowlistRequest"></a>

 ### QueryMeasurementAllowlistRequest
 QueryMeasurementAllowlistRequest is the request for QueryMeasurementAllowlist

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `tee_type` | [string](#string) |  | TEEType optionally filters by TEE type |
 | `include_revoked` | [bool](#bool) |  | IncludeRevoked optionally includes revoked measurements |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryMeasurementAllowlistResponse"></a>

 ### QueryMeasurementAllowlistResponse
 QueryMeasurementAllowlistResponse is the response for QueryMeasurementAllowlist

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurements` | [MeasurementRecord](#virtengine.enclave.v1.MeasurementRecord) | repeated | Measurements are the measurement records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryMeasurementRequest"></a>

 ### QueryMeasurementRequest
 QueryMeasurementRequest is the request for QueryMeasurement

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurement_hash` | [string](#string) |  | MeasurementHash is the hex-encoded measurement hash to query |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryMeasurementResponse"></a>

 ### QueryMeasurementResponse
 QueryMeasurementResponse is the response for QueryMeasurement

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurement` | [MeasurementRecord](#virtengine.enclave.v1.MeasurementRecord) |  | Measurement is the measurement record |
 | `is_allowed` | [bool](#bool) |  | IsAllowed indicates if the measurement is currently allowed |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for QueryParams

 

 

 
 <a name="virtengine.enclave.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for QueryParams

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.enclave.v1.Params) |  | Params are the module parameters |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryValidKeySetRequest"></a>

 ### QueryValidKeySetRequest
 QueryValidKeySetRequest is the request for QueryValidKeySet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `for_block_height` | [int64](#int64) |  | ForBlockHeight is the block height to check validity for |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.enclave.v1.QueryValidKeySetResponse"></a>

 ### QueryValidKeySetResponse
 QueryValidKeySetResponse is the response for QueryValidKeySet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_keys` | [ValidatorKeyInfo](#virtengine.enclave.v1.ValidatorKeyInfo) | repeated | ValidatorKeys are the valid validator key infos |
 | `total_count` | [int32](#int32) |  | TotalCount is the total number of valid keys |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.enclave.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the enclave module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `EnclaveIdentity` | [QueryEnclaveIdentityRequest](#virtengine.enclave.v1.QueryEnclaveIdentityRequest) | [QueryEnclaveIdentityResponse](#virtengine.enclave.v1.QueryEnclaveIdentityResponse) | EnclaveIdentity queries an enclave identity for a validator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/identity/{validator_address}|
 | `ActiveValidatorEnclaveKeys` | [QueryActiveValidatorEnclaveKeysRequest](#virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysRequest) | [QueryActiveValidatorEnclaveKeysResponse](#virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysResponse) | ActiveValidatorEnclaveKeys queries all active validator enclave keys buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/active_keys|
 | `CommitteeEnclaveKeys` | [QueryCommitteeEnclaveKeysRequest](#virtengine.enclave.v1.QueryCommitteeEnclaveKeysRequest) | [QueryCommitteeEnclaveKeysResponse](#virtengine.enclave.v1.QueryCommitteeEnclaveKeysResponse) | CommitteeEnclaveKeys queries committee enclave keys for an epoch buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/committee_keys|
 | `MeasurementAllowlist` | [QueryMeasurementAllowlistRequest](#virtengine.enclave.v1.QueryMeasurementAllowlistRequest) | [QueryMeasurementAllowlistResponse](#virtengine.enclave.v1.QueryMeasurementAllowlistResponse) | MeasurementAllowlist queries the measurement allowlist buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/measurements|
 | `Measurement` | [QueryMeasurementRequest](#virtengine.enclave.v1.QueryMeasurementRequest) | [QueryMeasurementResponse](#virtengine.enclave.v1.QueryMeasurementResponse) | Measurement queries a specific measurement buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/measurement/{measurement_hash}|
 | `KeyRotation` | [QueryKeyRotationRequest](#virtengine.enclave.v1.QueryKeyRotationRequest) | [QueryKeyRotationResponse](#virtengine.enclave.v1.QueryKeyRotationResponse) | KeyRotation queries key rotation status for a validator buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/rotation/{validator_address}|
 | `ValidKeySet` | [QueryValidKeySetRequest](#virtengine.enclave.v1.QueryValidKeySetRequest) | [QueryValidKeySetResponse](#virtengine.enclave.v1.QueryValidKeySetResponse) | ValidKeySet queries the current valid key set buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/valid_keys|
 | `Params` | [QueryParamsRequest](#virtengine.enclave.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.enclave.v1.QueryParamsResponse) | Params queries the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/params|
 | `AttestedResult` | [QueryAttestedResultRequest](#virtengine.enclave.v1.QueryAttestedResultRequest) | [QueryAttestedResultResponse](#virtengine.enclave.v1.QueryAttestedResultResponse) | AttestedResult queries an attested scoring result buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/enclave/v1/attested_result/{block_height}/{scope_id}|
 
  <!-- end services -->

 
 
 <a name="virtengine/enclave/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/enclave/v1/tx.proto
 

 
 <a name="virtengine.enclave.v1.MsgProposeMeasurement"></a>

 ### MsgProposeMeasurement
 MsgProposeMeasurement proposes a new enclave measurement for the allowlist

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the governance authority address |
 | `measurement_hash` | [bytes](#bytes) |  | MeasurementHash is the enclave measurement hash to add |
 | `tee_type` | [TEEType](#virtengine.enclave.v1.TEEType) |  | TEEType is the TEE type this measurement is for |
 | `description` | [string](#string) |  | Description is a human-readable description |
 | `min_isv_svn` | [uint32](#uint32) |  | MinISVSVN is the minimum required security version |
 | `expiry_blocks` | [int64](#int64) |  | ExpiryBlocks is the number of blocks until expiry (0 for no expiry) |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgProposeMeasurementResponse"></a>

 ### MsgProposeMeasurementResponse
 MsgProposeMeasurementResponse is the response for MsgProposeMeasurement

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `measurement_hash` | [string](#string) |  | MeasurementHash is the hash of the added measurement (hex-encoded) |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgRegisterEnclaveIdentity"></a>

 ### MsgRegisterEnclaveIdentity
 MsgRegisterEnclaveIdentity registers a new enclave identity for a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator operator address (sender must be the validator operator) |
 | `tee_type` | [TEEType](#virtengine.enclave.v1.TEEType) |  | TEEType is the type of TEE |
 | `measurement_hash` | [bytes](#bytes) |  | MeasurementHash is the enclave measurement hash |
 | `signer_hash` | [bytes](#bytes) |  | SignerHash is the signer measurement (MRSIGNER) |
 | `encryption_pub_key` | [bytes](#bytes) |  | EncryptionPubKey is the enclave's public key for encryption |
 | `signing_pub_key` | [bytes](#bytes) |  | SigningPubKey is the enclave's public key for signing |
 | `attestation_quote` | [bytes](#bytes) |  | AttestationQuote is the raw attestation quote |
 | `attestation_chain` | [bytes](#bytes) | repeated | AttestationChain is the certificate chain |
 | `isv_prod_id` | [uint32](#uint32) |  | ISVProdID is the product ID |
 | `isv_svn` | [uint32](#uint32) |  | ISVSVN is the security version number |
 | `quote_version` | [uint32](#uint32) |  | QuoteVersion is the quote format version |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgRegisterEnclaveIdentityResponse"></a>

 ### MsgRegisterEnclaveIdentityResponse
 MsgRegisterEnclaveIdentityResponse is the response for MsgRegisterEnclaveIdentity

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key_fingerprint` | [string](#string) |  | KeyFingerprint is the fingerprint of the registered key |
 | `expiry_height` | [int64](#int64) |  | ExpiryHeight is the block height when the identity expires |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgRevokeMeasurement"></a>

 ### MsgRevokeMeasurement
 MsgRevokeMeasurement revokes an enclave measurement from the allowlist

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the governance authority address |
 | `measurement_hash` | [bytes](#bytes) |  | MeasurementHash is the enclave measurement hash to revoke |
 | `reason` | [string](#string) |  | Reason is the reason for revocation |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgRevokeMeasurementResponse"></a>

 ### MsgRevokeMeasurementResponse
 MsgRevokeMeasurementResponse is the response for MsgRevokeMeasurement

 

 

 
 <a name="virtengine.enclave.v1.MsgRotateEnclaveIdentity"></a>

 ### MsgRotateEnclaveIdentity
 MsgRotateEnclaveIdentity initiates a key rotation for a validator's enclave

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator operator address |
 | `new_encryption_pub_key` | [bytes](#bytes) |  | NewEncryptionPubKey is the new enclave encryption public key |
 | `new_signing_pub_key` | [bytes](#bytes) |  | NewSigningPubKey is the new enclave signing public key |
 | `new_attestation_quote` | [bytes](#bytes) |  | NewAttestationQuote is the new attestation quote |
 | `new_attestation_chain` | [bytes](#bytes) | repeated | NewAttestationChain is the new certificate chain |
 | `new_measurement_hash` | [bytes](#bytes) |  | NewMeasurementHash is the new measurement hash (if enclave was upgraded) |
 | `new_isv_svn` | [uint32](#uint32) |  | NewISVSVN is the new security version |
 | `overlap_blocks` | [int64](#int64) |  | OverlapBlocks is the number of blocks for which both keys are valid |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgRotateEnclaveIdentityResponse"></a>

 ### MsgRotateEnclaveIdentityResponse
 MsgRotateEnclaveIdentityResponse is the response for MsgRotateEnclaveIdentity

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `new_key_fingerprint` | [string](#string) |  | NewKeyFingerprint is the fingerprint of the new key |
 | `overlap_start_height` | [int64](#int64) |  | OverlapStartHeight is when both keys become valid |
 | `overlap_end_height` | [int64](#int64) |  | OverlapEndHeight is when the old key becomes invalid |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module (x/gov module account) |
 | `params` | [Params](#virtengine.enclave.v1.Params) |  | Params are the new module parameters |
 
 

 

 
 <a name="virtengine.enclave.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.enclave.v1.Msg"></a>

 ### Msg
 Msg defines the enclave Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `RegisterEnclaveIdentity` | [MsgRegisterEnclaveIdentity](#virtengine.enclave.v1.MsgRegisterEnclaveIdentity) | [MsgRegisterEnclaveIdentityResponse](#virtengine.enclave.v1.MsgRegisterEnclaveIdentityResponse) | RegisterEnclaveIdentity registers a new enclave identity for a validator | |
 | `RotateEnclaveIdentity` | [MsgRotateEnclaveIdentity](#virtengine.enclave.v1.MsgRotateEnclaveIdentity) | [MsgRotateEnclaveIdentityResponse](#virtengine.enclave.v1.MsgRotateEnclaveIdentityResponse) | RotateEnclaveIdentity initiates a key rotation for a validator's enclave | |
 | `ProposeMeasurement` | [MsgProposeMeasurement](#virtengine.enclave.v1.MsgProposeMeasurement) | [MsgProposeMeasurementResponse](#virtengine.enclave.v1.MsgProposeMeasurementResponse) | ProposeMeasurement proposes a new enclave measurement for the allowlist | |
 | `RevokeMeasurement` | [MsgRevokeMeasurement](#virtengine.enclave.v1.MsgRevokeMeasurement) | [MsgRevokeMeasurementResponse](#virtengine.enclave.v1.MsgRevokeMeasurementResponse) | RevokeMeasurement revokes an enclave measurement from the allowlist | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.enclave.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.enclave.v1.MsgUpdateParamsResponse) | UpdateParams updates the module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/encryption/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/encryption/v1/types.proto
 

 
 <a name="virtengine.encryption.v1.AlgorithmInfo"></a>

 ### AlgorithmInfo
 AlgorithmInfo contains metadata about an encryption algorithm

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  | ID is the algorithm identifier |
 | `version` | [uint32](#uint32) |  | Version is the algorithm version |
 | `description` | [string](#string) |  | Description is a human-readable description |
 | `key_size` | [int32](#int32) |  | KeySize is the public key size in bytes |
 | `nonce_size` | [int32](#int32) |  | NonceSize is the nonce/IV size in bytes |
 | `deprecated` | [bool](#bool) |  | Deprecated indicates if this algorithm should no longer be used for new encryptions |
 
 

 

 
 <a name="virtengine.encryption.v1.EncryptedPayloadEnvelope"></a>

 ### EncryptedPayloadEnvelope
 EncryptedPayloadEnvelope is the canonical encrypted payload structure
for all sensitive fields stored on-chain.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `version` | [uint32](#uint32) |  | Version is the envelope format version for future compatibility |
 | `algorithm_id` | [string](#string) |  | AlgorithmID identifies the encryption algorithm used |
 | `algorithm_version` | [uint32](#uint32) |  | AlgorithmVersion is the version of the algorithm used |
 | `recipient_key_ids` | [string](#string) | repeated | RecipientKeyIDs are the fingerprints of intended recipients' public keys |
 | `recipient_public_keys` | [bytes](#bytes) | repeated | RecipientPublicKeys are the public keys for intended recipients |
 | `encrypted_keys` | [bytes](#bytes) | repeated | EncryptedKeys contains the data encryption key encrypted for each recipient |
 | `wrapped_keys` | [WrappedKeyEntry](#virtengine.encryption.v1.WrappedKeyEntry) | repeated | WrappedKeys contains per-recipient wrapped DEKs keyed by recipient ID |
 | `nonce` | [bytes](#bytes) |  | Nonce is the initialization vector / nonce for encryption |
 | `ciphertext` | [bytes](#bytes) |  | Ciphertext is the encrypted payload data |
 | `sender_signature` | [bytes](#bytes) |  | SenderSignature is the signature over hash(version || algorithm || ciphertext || nonce || recipients) |
 | `sender_pub_key` | [bytes](#bytes) |  | SenderPubKey is the sender's public key for signature verification |
 | `metadata` | [EncryptedPayloadEnvelope.MetadataEntry](#virtengine.encryption.v1.EncryptedPayloadEnvelope.MetadataEntry) | repeated | Metadata contains optional public or encrypted metadata |
 
 

 

 
 <a name="virtengine.encryption.v1.EncryptedPayloadEnvelope.MetadataEntry"></a>

 ### EncryptedPayloadEnvelope.MetadataEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.encryption.v1.EventKeyRegistered"></a>

 ### EventKeyRegistered
 EventKeyRegistered is emitted when a recipient key is registered

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  |  |
 | `fingerprint` | [string](#string) |  |  |
 | `algorithm` | [string](#string) |  |  |
 | `label` | [string](#string) |  |  |
 | `registered_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.encryption.v1.EventKeyRevoked"></a>

 ### EventKeyRevoked
 EventKeyRevoked is emitted when a recipient key is revoked

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  |  |
 | `fingerprint` | [string](#string) |  |  |
 | `revoked_by` | [string](#string) |  |  |
 | `revoked_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.encryption.v1.EventKeyUpdated"></a>

 ### EventKeyUpdated
 EventKeyUpdated is emitted when a recipient key is updated (e.g., label changed)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  |  |
 | `fingerprint` | [string](#string) |  |  |
 | `field` | [string](#string) |  |  |
 | `old_value` | [string](#string) |  |  |
 | `new_value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.encryption.v1.MultiRecipientEnvelope"></a>

 ### MultiRecipientEnvelope
 MultiRecipientEnvelope extends the encrypted envelope to support
encrypting to multiple validator enclaves for consensus recomputation.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `version` | [uint32](#uint32) |  | Version is the envelope format version |
 | `algorithm_id` | [string](#string) |  | AlgorithmID identifies the payload encryption algorithm |
 | `algorithm_version` | [uint32](#uint32) |  | AlgorithmVersion is the version of the payload encryption algorithm |
 | `recipient_mode` | [RecipientMode](#virtengine.encryption.v1.RecipientMode) |  | RecipientMode specifies how recipients were selected |
 | `payload_ciphertext` | [bytes](#bytes) |  | PayloadCiphertext is the encrypted payload (symmetric encryption) |
 | `payload_nonce` | [bytes](#bytes) |  | PayloadNonce is the nonce used for payload encryption |
 | `wrapped_keys` | [WrappedKeyEntry](#virtengine.encryption.v1.WrappedKeyEntry) | repeated | WrappedKeys contains the data encryption key wrapped for each recipient |
 | `client_signature` | [bytes](#bytes) |  | ClientSignature is the approved client's signature over the payload |
 | `client_id` | [string](#string) |  | ClientID is the approved client identifier |
 | `user_signature` | [bytes](#bytes) |  | UserSignature is the user's signature over the payload |
 | `user_pub_key` | [bytes](#bytes) |  | UserPubKey is the user's public key for signature verification |
 | `metadata` | [MultiRecipientEnvelope.MetadataEntry](#virtengine.encryption.v1.MultiRecipientEnvelope.MetadataEntry) | repeated | Metadata contains additional public metadata |
 | `committee_epoch` | [uint64](#uint64) |  | CommitteeEpoch is the committee epoch if RecipientMode is committee |
 
 

 

 
 <a name="virtengine.encryption.v1.MultiRecipientEnvelope.MetadataEntry"></a>

 ### MultiRecipientEnvelope.MetadataEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.encryption.v1.Params"></a>

 ### Params
 Params defines the parameters for the encryption module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_recipients_per_envelope` | [uint32](#uint32) |  | MaxRecipientsPerEnvelope is the maximum number of recipients per envelope |
 | `max_keys_per_account` | [uint32](#uint32) |  | MaxKeysPerAccount is the maximum number of keys an account can register |
 | `allowed_algorithms` | [string](#string) | repeated | AllowedAlgorithms is the list of allowed encryption algorithms Empty means all supported algorithms are allowed |
 | `require_signature` | [bool](#bool) |  | RequireSignature determines if envelope signatures are mandatory |
 
 

 

 
 <a name="virtengine.encryption.v1.RecipientKeyRecord"></a>

 ### RecipientKeyRecord
 RecipientKeyRecord represents a registered public key for receiving encrypted data

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address that owns this key |
 | `public_key` | [bytes](#bytes) |  | PublicKey is the X25519 public key bytes |
 | `key_fingerprint` | [string](#string) |  | KeyFingerprint is a unique identifier derived from the public key |
 | `algorithm_id` | [string](#string) |  | AlgorithmID specifies which algorithm this key is for |
 | `registered_at` | [int64](#int64) |  | RegisteredAt is the block time when the key was registered |
 | `revoked_at` | [int64](#int64) |  | RevokedAt is the block time when the key was revoked (0 if active) |
 | `label` | [string](#string) |  | Label is an optional human-readable label for the key |
 
 

 

 
 <a name="virtengine.encryption.v1.WrappedKeyEntry"></a>

 ### WrappedKeyEntry
 WrappedKeyEntry represents a per-recipient wrapped key

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `recipient_id` | [string](#string) |  | RecipientID is the unique identifier for the recipient (key fingerprint or validator address) |
 | `wrapped_key` | [bytes](#bytes) |  | WrappedKey is the data encryption key wrapped for this recipient |
 | `algorithm` | [string](#string) |  | Algorithm is the key wrapping algorithm used |
 | `ephemeral_pub_key` | [bytes](#bytes) |  | EphemeralPubKey is the ephemeral public key used for this recipient (if applicable) |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.encryption.v1.RecipientMode"></a>

 ### RecipientMode
 RecipientMode defines how recipients are selected for encryption

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | RECIPIENT_MODE_UNSPECIFIED | 0 | RECIPIENT_MODE_UNSPECIFIED is the unspecified recipient mode |
 | RECIPIENT_MODE_FULL_VALIDATOR_SET | 1 | RECIPIENT_MODE_FULL_VALIDATOR_SET encrypts to all active validators |
 | RECIPIENT_MODE_COMMITTEE | 2 | RECIPIENT_MODE_COMMITTEE encrypts to a designated identity committee subset |
 | RECIPIENT_MODE_SPECIFIC | 3 | RECIPIENT_MODE_SPECIFIC encrypts to specific recipients |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/encryption/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/encryption/v1/genesis.proto
 

 
 <a name="virtengine.encryption.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the encryption module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.encryption.v1.Params) |  | Params are the module parameters |
 | `recipient_keys` | [RecipientKeyRecord](#virtengine.encryption.v1.RecipientKeyRecord) | repeated | RecipientKeys are the initial registered recipient keys |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/encryption/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/encryption/v1/query.proto
 

 
 <a name="virtengine.encryption.v1.QueryAlgorithmsRequest"></a>

 ### QueryAlgorithmsRequest
 QueryAlgorithmsRequest is the request for querying supported algorithms

 

 

 
 <a name="virtengine.encryption.v1.QueryAlgorithmsResponse"></a>

 ### QueryAlgorithmsResponse
 QueryAlgorithmsResponse is the response for QueryAlgorithmsRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `algorithms` | [AlgorithmInfo](#virtengine.encryption.v1.AlgorithmInfo) | repeated | Algorithms is the list of supported algorithm info |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryKeyByFingerprintRequest"></a>

 ### QueryKeyByFingerprintRequest
 QueryKeyByFingerprintRequest is the request for querying a key by fingerprint

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `fingerprint` | [string](#string) |  | Fingerprint is the key fingerprint to look up |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryKeyByFingerprintResponse"></a>

 ### QueryKeyByFingerprintResponse
 QueryKeyByFingerprintResponse is the response for QueryKeyByFingerprintRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [RecipientKeyRecord](#virtengine.encryption.v1.RecipientKeyRecord) |  | Key is the recipient key record |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for querying module parameters

 

 

 
 <a name="virtengine.encryption.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for QueryParamsRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.encryption.v1.Params) |  | Params are the module parameters |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryRecipientKeyRequest"></a>

 ### QueryRecipientKeyRequest
 QueryRecipientKeyRequest is the request for querying a recipient's public key

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryRecipientKeyResponse"></a>

 ### QueryRecipientKeyResponse
 QueryRecipientKeyResponse is the response for QueryRecipientKeyRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `keys` | [RecipientKeyRecord](#virtengine.encryption.v1.RecipientKeyRecord) | repeated | Keys is the list of recipient key records |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryValidateEnvelopeRequest"></a>

 ### QueryValidateEnvelopeRequest
 QueryValidateEnvelopeRequest is the request for validating an envelope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `envelope` | [EncryptedPayloadEnvelope](#virtengine.encryption.v1.EncryptedPayloadEnvelope) |  | Envelope is the encrypted payload envelope to validate |
 
 

 

 
 <a name="virtengine.encryption.v1.QueryValidateEnvelopeResponse"></a>

 ### QueryValidateEnvelopeResponse
 QueryValidateEnvelopeResponse is the response for QueryValidateEnvelopeRequest

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `valid` | [bool](#bool) |  | Valid indicates if the envelope is valid |
 | `error` | [string](#string) |  | Error contains the validation error message if not valid |
 | `recipient_count` | [int32](#int32) |  | RecipientCount is the number of recipients in the envelope |
 | `algorithm` | [string](#string) |  | Algorithm is the encryption algorithm used |
 | `signature_valid` | [bool](#bool) |  | SignatureValid indicates if the sender signature is valid |
 | `all_keys_registered` | [bool](#bool) |  | AllKeysRegistered indicates if all recipient keys are registered |
 | `missing_keys` | [string](#string) | repeated | MissingKeys is the list of missing key fingerprints |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.encryption.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the encryption module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `RecipientKey` | [QueryRecipientKeyRequest](#virtengine.encryption.v1.QueryRecipientKeyRequest) | [QueryRecipientKeyResponse](#virtengine.encryption.v1.QueryRecipientKeyResponse) | RecipientKey returns the recipient keys for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/encryption/v1/key/{address}|
 | `KeyByFingerprint` | [QueryKeyByFingerprintRequest](#virtengine.encryption.v1.QueryKeyByFingerprintRequest) | [QueryKeyByFingerprintResponse](#virtengine.encryption.v1.QueryKeyByFingerprintResponse) | KeyByFingerprint returns a key by its fingerprint buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/encryption/v1/fingerprint/{fingerprint}|
 | `Params` | [QueryParamsRequest](#virtengine.encryption.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.encryption.v1.QueryParamsResponse) | Params returns the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/encryption/v1/params|
 | `Algorithms` | [QueryAlgorithmsRequest](#virtengine.encryption.v1.QueryAlgorithmsRequest) | [QueryAlgorithmsResponse](#virtengine.encryption.v1.QueryAlgorithmsResponse) | Algorithms returns the supported encryption algorithms buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/encryption/v1/algorithms|
 | `ValidateEnvelope` | [QueryValidateEnvelopeRequest](#virtengine.encryption.v1.QueryValidateEnvelopeRequest) | [QueryValidateEnvelopeResponse](#virtengine.encryption.v1.QueryValidateEnvelopeResponse) | ValidateEnvelope validates an encrypted payload envelope buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | POST|/virtengine/encryption/v1/validate|
 
  <!-- end services -->

 
 
 <a name="virtengine/encryption/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/encryption/v1/tx.proto
 

 
 <a name="virtengine.encryption.v1.MsgRegisterRecipientKey"></a>

 ### MsgRegisterRecipientKey
 MsgRegisterRecipientKey is the message for registering a recipient public key

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account registering the key (must match the key owner) |
 | `public_key` | [bytes](#bytes) |  | PublicKey is the X25519 public key bytes (32 bytes) |
 | `algorithm_id` | [string](#string) |  | AlgorithmID specifies which algorithm this key is for |
 | `label` | [string](#string) |  | Label is an optional human-readable label for the key |
 
 

 

 
 <a name="virtengine.encryption.v1.MsgRegisterRecipientKeyResponse"></a>

 ### MsgRegisterRecipientKeyResponse
 MsgRegisterRecipientKeyResponse is the response for MsgRegisterRecipientKey

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key_fingerprint` | [string](#string) |  | KeyFingerprint is the fingerprint of the registered key |
 
 

 

 
 <a name="virtengine.encryption.v1.MsgRevokeRecipientKey"></a>

 ### MsgRevokeRecipientKey
 MsgRevokeRecipientKey is the message for revoking a recipient public key

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account revoking the key (must own the key) |
 | `key_fingerprint` | [string](#string) |  | KeyFingerprint is the fingerprint of the key to revoke |
 
 

 

 
 <a name="virtengine.encryption.v1.MsgRevokeRecipientKeyResponse"></a>

 ### MsgRevokeRecipientKeyResponse
 MsgRevokeRecipientKeyResponse is the response for MsgRevokeRecipientKey

 

 

 
 <a name="virtengine.encryption.v1.MsgUpdateKeyLabel"></a>

 ### MsgUpdateKeyLabel
 MsgUpdateKeyLabel is the message for updating a key's label

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account updating the key (must own the key) |
 | `key_fingerprint` | [string](#string) |  | KeyFingerprint is the fingerprint of the key to update |
 | `label` | [string](#string) |  | Label is the new label for the key |
 
 

 

 
 <a name="virtengine.encryption.v1.MsgUpdateKeyLabelResponse"></a>

 ### MsgUpdateKeyLabelResponse
 MsgUpdateKeyLabelResponse is the response for MsgUpdateKeyLabel

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.encryption.v1.Msg"></a>

 ### Msg
 Msg defines the encryption Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `RegisterRecipientKey` | [MsgRegisterRecipientKey](#virtengine.encryption.v1.MsgRegisterRecipientKey) | [MsgRegisterRecipientKeyResponse](#virtengine.encryption.v1.MsgRegisterRecipientKeyResponse) | RegisterRecipientKey registers a new recipient public key | |
 | `RevokeRecipientKey` | [MsgRevokeRecipientKey](#virtengine.encryption.v1.MsgRevokeRecipientKey) | [MsgRevokeRecipientKeyResponse](#virtengine.encryption.v1.MsgRevokeRecipientKeyResponse) | RevokeRecipientKey revokes an existing recipient public key | |
 | `UpdateKeyLabel` | [MsgUpdateKeyLabel](#virtengine.encryption.v1.MsgUpdateKeyLabel) | [MsgUpdateKeyLabelResponse](#virtengine.encryption.v1.MsgUpdateKeyLabelResponse) | UpdateKeyLabel updates the label of an existing recipient key | |
 
  <!-- end services -->

 
 
 <a name="virtengine/epochs/v1beta1/events.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/epochs/v1beta1/events.proto
 

 
 <a name="virtengine.epochs.v1beta1.EventEpochEnd"></a>

 ### EventEpochEnd
 EventEpochEnd is an event emitted when an epoch end.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `epoch_number` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.epochs.v1beta1.EventEpochStart"></a>

 ### EventEpochStart
 EventEpochStart is an event emitted when an epoch start.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `epoch_number` | [int64](#int64) |  |  |
 | `epoch_start_time` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/epochs/v1beta1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/epochs/v1beta1/genesis.proto
 

 
 <a name="virtengine.epochs.v1beta1.EpochInfo"></a>

 ### EpochInfo
 EpochInfo is a struct that describes the data going into
a timer defined by the x/epochs module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  | id is a unique reference to this particular timer. |
 | `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | start_time is the time at which the timer first ever ticks. If start_time is in the future, the epoch will not begin until the start time. |
 | `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  | duration is the time in between epoch ticks. In order for intended behavior to be met, duration should be greater than the chains expected block time. Duration must be non-zero. |
 | `current_epoch` | [int64](#int64) |  | current_epoch is the current epoch number, or in other words, how many times has the timer 'ticked'. The first tick (current_epoch=1) is defined as the first block whose blocktime is greater than the EpochInfo start_time. |
 | `current_epoch_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | current_epoch_start_time describes the start time of the current timer interval. The interval is (current_epoch_start_time, current_epoch_start_time + duration] When the timer ticks, this is set to current_epoch_start_time = last_epoch_start_time + duration only one timer tick for a given identifier can occur per block.

NOTE! The current_epoch_start_time may diverge significantly from the wall-clock time the epoch began at. Wall-clock time of epoch start may be >> current_epoch_start_time. Suppose current_epoch_start_time = 10, duration = 5. Suppose the chain goes offline at t=14, and comes back online at t=30, and produces blocks at every successive time. (t=31, 32, etc.) * The t=30 block will start the epoch for (10, 15] * The t=31 block will start the epoch for (15, 20] * The t=32 block will start the epoch for (20, 25] * The t=33 block will start the epoch for (25, 30] * The t=34 block will start the epoch for (30, 35] * The **t=36** block will start the epoch for (35, 40] |
 | `epoch_counting_started` | [bool](#bool) |  | epoch_counting_started is a boolean, that indicates whether this epoch timer has began yet. |
 | `current_epoch_start_height` | [int64](#int64) |  | current_epoch_start_height is the block height at which the current epoch started. (The block height at which the timer last ticked) |
 
 

 

 
 <a name="virtengine.epochs.v1beta1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the epochs module's genesis state.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `epochs` | [EpochInfo](#virtengine.epochs.v1beta1.EpochInfo) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/epochs/v1beta1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/epochs/v1beta1/query.proto
 

 
 <a name="virtengine.epochs.v1beta1.QueryCurrentEpochRequest"></a>

 ### QueryCurrentEpochRequest
 QueryCurrentEpochRequest defines the gRPC request structure for
querying an epoch by its identifier.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `identifier` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.epochs.v1beta1.QueryCurrentEpochResponse"></a>

 ### QueryCurrentEpochResponse
 QueryCurrentEpochResponse defines the gRPC response structure for
querying an epoch by its identifier.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `current_epoch` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.epochs.v1beta1.QueryEpochInfosRequest"></a>

 ### QueryEpochInfosRequest
 QueryEpochInfosRequest defines the gRPC request structure for
querying all epoch info.

 

 

 
 <a name="virtengine.epochs.v1beta1.QueryEpochInfosResponse"></a>

 ### QueryEpochInfosResponse
 QueryEpochInfosRequest defines the gRPC response structure for
querying all epoch info.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `epochs` | [EpochInfo](#virtengine.epochs.v1beta1.EpochInfo) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.epochs.v1beta1.Query"></a>

 ### Query
 Query defines the gRPC querier service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `EpochInfos` | [QueryEpochInfosRequest](#virtengine.epochs.v1beta1.QueryEpochInfosRequest) | [QueryEpochInfosResponse](#virtengine.epochs.v1beta1.QueryEpochInfosResponse) | EpochInfos provide running epochInfos | GET|/cosmos/epochs/v1beta1/epochs|
 | `CurrentEpoch` | [QueryCurrentEpochRequest](#virtengine.epochs.v1beta1.QueryCurrentEpochRequest) | [QueryCurrentEpochResponse](#virtengine.epochs.v1beta1.QueryCurrentEpochResponse) | CurrentEpoch provide current epoch of specified identifier | GET|/cosmos/epochs/v1beta1/current_epoch|
 
  <!-- end services -->

 
 
 <a name="virtengine/escrow/types/v1/payment.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/types/v1/payment.proto
 

 
 <a name="virtengine.escrow.types.v1.Payment"></a>

 ### Payment
 Payment

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.escrow.id.v1.Payment](#virtengine.escrow.id.v1.Payment) |  |  |
 | `state` | [PaymentState](#virtengine.escrow.types.v1.PaymentState) |  |  |
 
 

 

 
 <a name="virtengine.escrow.types.v1.PaymentState"></a>

 ### PaymentState
 Payment stores state for a payment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `state` | [State](#virtengine.escrow.types.v1.State) |  | State represents the state of the Payment. |
 | `rate` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Rate holds the rate of the Payment. |
 | `balance` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Balance is the current available coins. |
 | `unsettled` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Unsettled is the amount needed to settle payment if account is overdrawn |
 | `withdrawn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Withdrawn corresponds to the amount of coins withdrawn by the Payment. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/v1/authz.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/v1/authz.proto
 

 
 <a name="virtengine.escrow.v1.DepositAuthorization"></a>

 ### DepositAuthorization
 DepositAuthorization allows the grantee to deposit up to spend_limit coins from
the granter's account for VirtEngine deployments and bids. This authorization is used
within the Cosmos SDK authz module to grant scoped permissions for deposit operations.
The authorization can be restricted to specific scopes (deployment or bid) to limit
what types of deposits the grantee is authorized to make on behalf of the granter.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `spend_limit` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | SpendLimit is the maximum amount the grantee is authorized to spend from the granter's account. This limit applies cumulatively across all deposit operations within the authorized scopes. Once this limit is reached, the authorization becomes invalid and no further deposits can be made. |
 | `scopes` | [DepositAuthorization.Scope](#virtengine.escrow.v1.DepositAuthorization.Scope) | repeated | Scopes defines the specific types of deposit operations this authorization permits. This provides fine-grained control over what operations the grantee can perform using the granter's funds. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.escrow.v1.DepositAuthorization.Scope"></a>

 ### DepositAuthorization.Scope
 Scope defines the types of deposit operations that can be authorized.
This enum is used to restrict the authorization to specific deposit contexts,
allowing fine-grained permission control within the authz system.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | deployment | 1 | DepositScopeDeployment allows deposits for deployment-related operations. |
 | bid | 2 | DepositScopeBid allows deposits for bid-related operations. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/v1/genesis.proto
 

 
 <a name="virtengine.escrow.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the basic genesis state used by the escrow module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `accounts` | [virtengine.escrow.types.v1.Account](#virtengine.escrow.types.v1.Account) | repeated | Accounts is a list of accounts on the genesis state. |
 | `payments` | [virtengine.escrow.types.v1.Payment](#virtengine.escrow.types.v1.Payment) | repeated | Payments is a list of fractional payments on the genesis state.. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/v1/msg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/v1/msg.proto
 

 
 <a name="virtengine.escrow.v1.MsgAccountDeposit"></a>

 ### MsgAccountDeposit
 MsgAccountDeposit represents a message to deposit funds into an existing escrow account
on the blockchain. This is part of the interaction mechanism for managing
deployment-related resources.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `signer` | [string](#string) |  | Signer is the account bech32 address of the user who wants to deposit into an escrow account. Does not necessarily needs to be an owner of the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `id` | [virtengine.escrow.id.v1.Account](#virtengine.escrow.id.v1.Account) |  | ID is the unique identifier of the account. |
 | `deposit` | [virtengine.base.deposit.v1.Deposit](#virtengine.base.deposit.v1.Deposit) |  | Deposit contains information about the deposit amount and the source of the deposit to the escrow account. |
 
 

 

 
 <a name="virtengine.escrow.v1.MsgAccountDepositResponse"></a>

 ### MsgAccountDepositResponse
 MsgAccountDepositResponse defines response type for the MsgDeposit.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/escrow/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/v1/query.proto
 

 
 <a name="virtengine.escrow.v1.QueryAccountsRequest"></a>

 ### QueryAccountsRequest
 QueryAccountRequest is request type for the Query/Account RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [string](#string) |  | State represents the current state of an Account. |
 | `xid` | [string](#string) |  | Scope holds the scope of the account. |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.escrow.v1.QueryAccountsResponse"></a>

 ### QueryAccountsResponse
 QueryProvidersResponse is response type for the Query/Providers RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `accounts` | [virtengine.escrow.types.v1.Account](#virtengine.escrow.types.v1.Account) | repeated | Accounts is a list of Account. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

 
 <a name="virtengine.escrow.v1.QueryPaymentsRequest"></a>

 ### QueryPaymentsRequest
 QueryPaymentRequest is request type for the Query/Payment RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [string](#string) |  | State represents the current state of a Payment. |
 | `xid` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.escrow.v1.QueryPaymentsResponse"></a>

 ### QueryPaymentsResponse
 QueryProvidersResponse is response type for the Query/Providers RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `payments` | [virtengine.escrow.types.v1.Payment](#virtengine.escrow.types.v1.Payment) | repeated | Payments is a list of payments. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.escrow.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the escrow package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Accounts` | [QueryAccountsRequest](#virtengine.escrow.v1.QueryAccountsRequest) | [QueryAccountsResponse](#virtengine.escrow.v1.QueryAccountsResponse) | buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME Accounts queries all accounts. | GET|/virtengine/escrow/v1/types/accounts|
 | `Payments` | [QueryPaymentsRequest](#virtengine.escrow.v1.QueryPaymentsRequest) | [QueryPaymentsResponse](#virtengine.escrow.v1.QueryPaymentsResponse) | buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME Payments queries all payments. | GET|/virtengine/escrow/v1/types/payments|
 
  <!-- end services -->

 
 
 <a name="virtengine/escrow/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/escrow/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.escrow.v1.Msg"></a>

 ### Msg
 Msg defines the x/deployment Msg service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `AccountDeposit` | [MsgAccountDeposit](#virtengine.escrow.v1.MsgAccountDeposit) | [MsgAccountDepositResponse](#virtengine.escrow.v1.MsgAccountDepositResponse) | AccountDeposit deposits more funds into the escrow account. | |
 
  <!-- end services -->

 
 
 <a name="virtengine/fraud/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/fraud/v1/types.proto
 

 
 <a name="virtengine.fraud.v1.EncryptedEvidence"></a>

 ### EncryptedEvidence
 EncryptedEvidence holds encrypted evidence for a fraud report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `algorithm_id` | [string](#string) |  | AlgorithmID identifies the encryption algorithm used |
 | `recipient_key_ids` | [string](#string) | repeated | RecipientKeyIDs are the fingerprints of moderator public keys |
 | `encrypted_keys` | [bytes](#bytes) | repeated | EncryptedKeys contains the data encryption key encrypted for each recipient |
 | `nonce` | [bytes](#bytes) |  | Nonce is the initialization vector for encryption |
 | `ciphertext` | [bytes](#bytes) |  | Ciphertext is the encrypted evidence data |
 | `sender_signature` | [bytes](#bytes) |  | SenderSignature is the signature for authenticity verification |
 | `sender_pub_key` | [bytes](#bytes) |  | SenderPubKey is the sender's public key |
 | `content_type` | [string](#string) |  | ContentType indicates the type of evidence (e.g., "image/png", "application/json") |
 | `evidence_hash` | [string](#string) |  | EvidenceHash is SHA256 hash of the original evidence for integrity verification |
 
 

 

 
 <a name="virtengine.fraud.v1.FraudAuditLog"></a>

 ### FraudAuditLog
 FraudAuditLog represents an audit log entry for a fraud report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  | ID is the unique identifier for this log entry |
 | `report_id` | [string](#string) |  | ReportID is the associated fraud report ID |
 | `action` | [AuditAction](#virtengine.fraud.v1.AuditAction) |  | Action is the type of action performed |
 | `actor` | [string](#string) |  | Actor is the address that performed the action |
 | `previous_status` | [FraudReportStatus](#virtengine.fraud.v1.FraudReportStatus) |  | PreviousStatus is the status before the action (if applicable) |
 | `new_status` | [FraudReportStatus](#virtengine.fraud.v1.FraudReportStatus) |  | NewStatus is the status after the action (if applicable) |
 | `details` | [string](#string) |  | Details contains additional action-specific details |
 | `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp is when the action was performed |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height when the action occurred |
 | `tx_hash` | [string](#string) |  | TxHash is the transaction hash (if applicable) |
 
 

 

 
 <a name="virtengine.fraud.v1.FraudReport"></a>

 ### FraudReport
 FraudReport represents a fraud report submitted by a provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [string](#string) |  | ID is the unique identifier for this report |
 | `reporter` | [string](#string) |  | Reporter is the provider address who submitted the report |
 | `reported_party` | [string](#string) |  | ReportedParty is the address of the party being reported |
 | `category` | [FraudCategory](#virtengine.fraud.v1.FraudCategory) |  | Category is the type of fraud being reported |
 | `description` | [string](#string) |  | Description is the detailed description of the fraud |
 | `evidence` | [EncryptedEvidence](#virtengine.fraud.v1.EncryptedEvidence) | repeated | Evidence contains the encrypted evidence attachments |
 | `status` | [FraudReportStatus](#virtengine.fraud.v1.FraudReportStatus) |  | Status is the current status of the report |
 | `assigned_moderator` | [string](#string) |  | AssignedModerator is the moderator assigned to review this report |
 | `resolution` | [ResolutionType](#virtengine.fraud.v1.ResolutionType) |  | Resolution is the resolution type if resolved |
 | `resolution_notes` | [string](#string) |  | ResolutionNotes are notes provided by the moderator upon resolution |
 | `submitted_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | SubmittedAt is when the report was submitted |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | UpdatedAt is when the report was last updated |
 | `resolved_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | ResolvedAt is when the report was resolved (if applicable) |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height when submitted |
 | `content_hash` | [string](#string) |  | ContentHash is SHA256 hash of the report content for integrity |
 | `related_order_ids` | [string](#string) | repeated | RelatedOrderIDs are order IDs related to this fraud report |
 
 

 

 
 <a name="virtengine.fraud.v1.ModeratorQueueEntry"></a>

 ### ModeratorQueueEntry
 ModeratorQueueEntry represents an entry in the moderator queue

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID |
 | `priority` | [uint32](#uint32) |  | Priority is the queue priority (higher = more urgent) |
 | `queued_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | QueuedAt is when the report was added to the queue |
 | `category` | [FraudCategory](#virtengine.fraud.v1.FraudCategory) |  | Category is the fraud category for routing |
 | `assigned_to` | [string](#string) |  | AssignedTo is the moderator assigned (empty if unassigned) |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.fraud.v1.AuditAction"></a>

 ### AuditAction
 AuditAction represents the type of action recorded in the audit log

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | AUDIT_ACTION_UNSPECIFIED | 0 | AUDIT_ACTION_UNSPECIFIED represents an unspecified action |
 | AUDIT_ACTION_SUBMITTED | 1 | AUDIT_ACTION_SUBMITTED indicates report was submitted |
 | AUDIT_ACTION_ASSIGNED | 2 | AUDIT_ACTION_ASSIGNED indicates report was assigned to moderator |
 | AUDIT_ACTION_STATUS_CHANGED | 3 | AUDIT_ACTION_STATUS_CHANGED indicates status was changed |
 | AUDIT_ACTION_EVIDENCE_VIEWED | 4 | AUDIT_ACTION_EVIDENCE_VIEWED indicates evidence was viewed by moderator |
 | AUDIT_ACTION_RESOLVED | 5 | AUDIT_ACTION_RESOLVED indicates report was resolved |
 | AUDIT_ACTION_REJECTED | 6 | AUDIT_ACTION_REJECTED indicates report was rejected |
 | AUDIT_ACTION_ESCALATED | 7 | AUDIT_ACTION_ESCALATED indicates report was escalated |
 | AUDIT_ACTION_COMMENT_ADDED | 8 | AUDIT_ACTION_COMMENT_ADDED indicates a comment was added |
 

 
 <a name="virtengine.fraud.v1.FraudCategory"></a>

 ### FraudCategory
 FraudCategory represents the category of fraud being reported

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | FRAUD_CATEGORY_UNSPECIFIED | 0 | FRAUD_CATEGORY_UNSPECIFIED represents an unspecified category |
 | FRAUD_CATEGORY_FAKE_IDENTITY | 1 | FRAUD_CATEGORY_FAKE_IDENTITY indicates fake or stolen identity |
 | FRAUD_CATEGORY_PAYMENT_FRAUD | 2 | FRAUD_CATEGORY_PAYMENT_FRAUD indicates payment-related fraud |
 | FRAUD_CATEGORY_SERVICE_MISREPRESENTATION | 3 | FRAUD_CATEGORY_SERVICE_MISREPRESENTATION indicates misrepresented services |
 | FRAUD_CATEGORY_RESOURCE_ABUSE | 4 | FRAUD_CATEGORY_RESOURCE_ABUSE indicates abuse of allocated resources |
 | FRAUD_CATEGORY_SYBIL_ATTACK | 5 | FRAUD_CATEGORY_SYBIL_ATTACK indicates suspected sybil attack |
 | FRAUD_CATEGORY_MALICIOUS_CONTENT | 6 | FRAUD_CATEGORY_MALICIOUS_CONTENT indicates malicious content or software |
 | FRAUD_CATEGORY_TERMS_VIOLATION | 7 | FRAUD_CATEGORY_TERMS_VIOLATION indicates terms of service violation |
 | FRAUD_CATEGORY_OTHER | 8 | FRAUD_CATEGORY_OTHER indicates other fraud types |
 

 
 <a name="virtengine.fraud.v1.FraudReportStatus"></a>

 ### FraudReportStatus
 FraudReportStatus represents the state of a fraud report

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | FRAUD_REPORT_STATUS_UNSPECIFIED | 0 | FRAUD_REPORT_STATUS_UNSPECIFIED represents an unspecified status |
 | FRAUD_REPORT_STATUS_SUBMITTED | 1 | FRAUD_REPORT_STATUS_SUBMITTED indicates the report has been submitted and is awaiting review |
 | FRAUD_REPORT_STATUS_REVIEWING | 2 | FRAUD_REPORT_STATUS_REVIEWING indicates a moderator is actively reviewing the report |
 | FRAUD_REPORT_STATUS_RESOLVED | 3 | FRAUD_REPORT_STATUS_RESOLVED indicates the report has been resolved (action taken) |
 | FRAUD_REPORT_STATUS_REJECTED | 4 | FRAUD_REPORT_STATUS_REJECTED indicates the report was rejected (no fraud found) |
 | FRAUD_REPORT_STATUS_ESCALATED | 5 | FRAUD_REPORT_STATUS_ESCALATED indicates the report has been escalated to admin |
 

 
 <a name="virtengine.fraud.v1.ResolutionType"></a>

 ### ResolutionType
 ResolutionType represents the type of resolution for a fraud report

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | RESOLUTION_TYPE_UNSPECIFIED | 0 | RESOLUTION_TYPE_UNSPECIFIED represents no resolution yet |
 | RESOLUTION_TYPE_WARNING | 1 | RESOLUTION_TYPE_WARNING indicates a warning was issued |
 | RESOLUTION_TYPE_SUSPENSION | 2 | RESOLUTION_TYPE_SUSPENSION indicates account suspension |
 | RESOLUTION_TYPE_TERMINATION | 3 | RESOLUTION_TYPE_TERMINATION indicates account termination |
 | RESOLUTION_TYPE_REFUND | 4 | RESOLUTION_TYPE_REFUND indicates a refund was processed |
 | RESOLUTION_TYPE_NO_ACTION | 5 | RESOLUTION_TYPE_NO_ACTION indicates no action taken (rejected) |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/fraud/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/fraud/v1/params.proto
 

 
 <a name="virtengine.fraud.v1.Params"></a>

 ### Params
 Params defines the parameters for the fraud module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `min_description_length` | [int32](#int32) |  | MinDescriptionLength is the minimum description length |
 | `max_description_length` | [int32](#int32) |  | MaxDescriptionLength is the maximum description length |
 | `max_evidence_count` | [int32](#int32) |  | MaxEvidenceCount is the maximum number of evidence items per report |
 | `max_evidence_size_bytes` | [int64](#int64) |  | MaxEvidenceSizeBytes is the maximum size per evidence item |
 | `auto_assign_enabled` | [bool](#bool) |  | AutoAssignEnabled enables automatic moderator assignment |
 | `escalation_threshold_days` | [int32](#int32) |  | EscalationThresholdDays is days before auto-escalation |
 | `report_retention_days` | [int32](#int32) |  | ReportRetentionDays is how long to retain resolved reports |
 | `audit_log_retention_days` | [int32](#int32) |  | AuditLogRetentionDays is how long to retain audit logs |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/fraud/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/fraud/v1/genesis.proto
 

 
 <a name="virtengine.fraud.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the fraud module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.fraud.v1.Params) |  | Params defines the module parameters |
 | `reports` | [FraudReport](#virtengine.fraud.v1.FraudReport) | repeated | Reports contains all fraud reports |
 | `audit_logs` | [FraudAuditLog](#virtengine.fraud.v1.FraudAuditLog) | repeated | AuditLogs contains all audit log entries |
 | `moderator_queue` | [ModeratorQueueEntry](#virtengine.fraud.v1.ModeratorQueueEntry) | repeated | ModeratorQueue contains all queue entries |
 | `next_fraud_report_sequence` | [uint64](#uint64) |  | NextFraudReportSequence is the next fraud report sequence number |
 | `next_audit_log_sequence` | [uint64](#uint64) |  | NextAuditLogSequence is the next audit log sequence number |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/fraud/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/fraud/v1/query.proto
 

 
 <a name="virtengine.fraud.v1.QueryAuditLogRequest"></a>

 ### QueryAuditLogRequest
 QueryAuditLogRequest is the request for AuditLog query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination defines pagination options |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryAuditLogResponse"></a>

 ### QueryAuditLogResponse
 QueryAuditLogResponse is the response for AuditLog query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `audit_logs` | [FraudAuditLog](#virtengine.fraud.v1.FraudAuditLog) | repeated | AuditLogs are the audit log entries |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination defines the pagination response |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportRequest"></a>

 ### QueryFraudReportRequest
 QueryFraudReportRequest is the request for FraudReport query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID to query |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportResponse"></a>

 ### QueryFraudReportResponse
 QueryFraudReportResponse is the response for FraudReport query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `report` | [FraudReport](#virtengine.fraud.v1.FraudReport) |  | Report is the fraud report |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportsByReportedPartyRequest"></a>

 ### QueryFraudReportsByReportedPartyRequest
 QueryFraudReportsByReportedPartyRequest is the request for FraudReportsByReportedParty query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reported_party` | [string](#string) |  | ReportedParty is the reported party address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination defines pagination options |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportsByReportedPartyResponse"></a>

 ### QueryFraudReportsByReportedPartyResponse
 QueryFraudReportsByReportedPartyResponse is the response for FraudReportsByReportedParty query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reports` | [FraudReport](#virtengine.fraud.v1.FraudReport) | repeated | Reports are the fraud reports |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination defines the pagination response |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportsByReporterRequest"></a>

 ### QueryFraudReportsByReporterRequest
 QueryFraudReportsByReporterRequest is the request for FraudReportsByReporter query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reporter` | [string](#string) |  | Reporter is the reporter address |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination defines pagination options |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportsByReporterResponse"></a>

 ### QueryFraudReportsByReporterResponse
 QueryFraudReportsByReporterResponse is the response for FraudReportsByReporter query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reports` | [FraudReport](#virtengine.fraud.v1.FraudReport) | repeated | Reports are the fraud reports |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination defines the pagination response |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportsRequest"></a>

 ### QueryFraudReportsRequest
 QueryFraudReportsRequest is the request for FraudReports query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination defines pagination options |
 | `status` | [FraudReportStatus](#virtengine.fraud.v1.FraudReportStatus) |  | Status filters reports by status (optional) |
 | `category` | [FraudCategory](#virtengine.fraud.v1.FraudCategory) |  | Category filters reports by category (optional) |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryFraudReportsResponse"></a>

 ### QueryFraudReportsResponse
 QueryFraudReportsResponse is the response for FraudReports query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reports` | [FraudReport](#virtengine.fraud.v1.FraudReport) | repeated | Reports are the fraud reports |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination defines the pagination response |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryModeratorQueueRequest"></a>

 ### QueryModeratorQueueRequest
 QueryModeratorQueueRequest is the request for ModeratorQueue query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination defines pagination options |
 | `category` | [FraudCategory](#virtengine.fraud.v1.FraudCategory) |  | Category filters queue entries by category (optional) |
 | `assigned_to` | [string](#string) |  | AssignedTo filters queue entries by assigned moderator (optional) |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryModeratorQueueResponse"></a>

 ### QueryModeratorQueueResponse
 QueryModeratorQueueResponse is the response for ModeratorQueue query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `queue_entries` | [ModeratorQueueEntry](#virtengine.fraud.v1.ModeratorQueueEntry) | repeated | QueueEntries are the moderator queue entries |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination defines the pagination response |
 
 

 

 
 <a name="virtengine.fraud.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for Params query

 

 

 
 <a name="virtengine.fraud.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for Params query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.fraud.v1.Params) |  | Params holds all the parameters of this module |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.fraud.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for fraud module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Params` | [QueryParamsRequest](#virtengine.fraud.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.fraud.v1.QueryParamsResponse) | Params returns the module parameters | GET|/virtengine/fraud/v1/params|
 | `FraudReport` | [QueryFraudReportRequest](#virtengine.fraud.v1.QueryFraudReportRequest) | [QueryFraudReportResponse](#virtengine.fraud.v1.QueryFraudReportResponse) | FraudReport returns a fraud report by ID | GET|/virtengine/fraud/v1/reports/{report_id}|
 | `FraudReports` | [QueryFraudReportsRequest](#virtengine.fraud.v1.QueryFraudReportsRequest) | [QueryFraudReportsResponse](#virtengine.fraud.v1.QueryFraudReportsResponse) | FraudReports returns all fraud reports with optional filters | GET|/virtengine/fraud/v1/reports|
 | `FraudReportsByReporter` | [QueryFraudReportsByReporterRequest](#virtengine.fraud.v1.QueryFraudReportsByReporterRequest) | [QueryFraudReportsByReporterResponse](#virtengine.fraud.v1.QueryFraudReportsByReporterResponse) | FraudReportsByReporter returns fraud reports submitted by a reporter | GET|/virtengine/fraud/v1/reports/reporter/{reporter}|
 | `FraudReportsByReportedParty` | [QueryFraudReportsByReportedPartyRequest](#virtengine.fraud.v1.QueryFraudReportsByReportedPartyRequest) | [QueryFraudReportsByReportedPartyResponse](#virtengine.fraud.v1.QueryFraudReportsByReportedPartyResponse) | FraudReportsByReportedParty returns fraud reports against a reported party | GET|/virtengine/fraud/v1/reports/reported/{reported_party}|
 | `AuditLog` | [QueryAuditLogRequest](#virtengine.fraud.v1.QueryAuditLogRequest) | [QueryAuditLogResponse](#virtengine.fraud.v1.QueryAuditLogResponse) | AuditLog returns the audit log for a report | GET|/virtengine/fraud/v1/reports/{report_id}/audit|
 | `ModeratorQueue` | [QueryModeratorQueueRequest](#virtengine.fraud.v1.QueryModeratorQueueRequest) | [QueryModeratorQueueResponse](#virtengine.fraud.v1.QueryModeratorQueueResponse) | ModeratorQueue returns the moderator queue entries | GET|/virtengine/fraud/v1/moderator/queue|
 
  <!-- end services -->

 
 
 <a name="virtengine/fraud/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/fraud/v1/tx.proto
 

 
 <a name="virtengine.fraud.v1.MsgAssignModerator"></a>

 ### MsgAssignModerator
 MsgAssignModerator defines the message for assigning a moderator to a report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `moderator` | [string](#string) |  | Moderator is the moderator performing the assignment |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID to assign |
 | `assign_to` | [string](#string) |  | AssignTo is the moderator to assign (can be self) |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgAssignModeratorResponse"></a>

 ### MsgAssignModeratorResponse
 MsgAssignModeratorResponse is the response for MsgAssignModerator

 

 

 
 <a name="virtengine.fraud.v1.MsgEscalateFraudReport"></a>

 ### MsgEscalateFraudReport
 MsgEscalateFraudReport defines the message for escalating a fraud report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `moderator` | [string](#string) |  | Moderator is the moderator escalating the report |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID |
 | `reason` | [string](#string) |  | Reason is the escalation reason |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgEscalateFraudReportResponse"></a>

 ### MsgEscalateFraudReportResponse
 MsgEscalateFraudReportResponse is the response for MsgEscalateFraudReport

 

 

 
 <a name="virtengine.fraud.v1.MsgRejectFraudReport"></a>

 ### MsgRejectFraudReport
 MsgRejectFraudReport defines the message for rejecting a fraud report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `moderator` | [string](#string) |  | Moderator is the moderator rejecting the report |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID |
 | `notes` | [string](#string) |  | Notes are rejection notes |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgRejectFraudReportResponse"></a>

 ### MsgRejectFraudReportResponse
 MsgRejectFraudReportResponse is the response for MsgRejectFraudReport

 

 

 
 <a name="virtengine.fraud.v1.MsgResolveFraudReport"></a>

 ### MsgResolveFraudReport
 MsgResolveFraudReport defines the message for resolving a fraud report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `moderator` | [string](#string) |  | Moderator is the moderator resolving the report |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID |
 | `resolution` | [ResolutionType](#virtengine.fraud.v1.ResolutionType) |  | Resolution is the resolution type |
 | `notes` | [string](#string) |  | Notes are resolution notes |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgResolveFraudReportResponse"></a>

 ### MsgResolveFraudReportResponse
 MsgResolveFraudReportResponse is the response for MsgResolveFraudReport

 

 

 
 <a name="virtengine.fraud.v1.MsgSubmitFraudReport"></a>

 ### MsgSubmitFraudReport
 MsgSubmitFraudReport defines the message for submitting a fraud report

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reporter` | [string](#string) |  | Reporter is the provider address submitting the report |
 | `reported_party` | [string](#string) |  | ReportedParty is the address of the party being reported |
 | `category` | [FraudCategory](#virtengine.fraud.v1.FraudCategory) |  | Category is the type of fraud being reported |
 | `description` | [string](#string) |  | Description is the detailed description of the fraud |
 | `evidence` | [EncryptedEvidence](#virtengine.fraud.v1.EncryptedEvidence) | repeated | Evidence contains encrypted evidence attachments |
 | `related_order_ids` | [string](#string) | repeated | RelatedOrderIDs are optional order IDs related to this fraud |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgSubmitFraudReportResponse"></a>

 ### MsgSubmitFraudReportResponse
 MsgSubmitFraudReportResponse is the response for MsgSubmitFraudReport

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `report_id` | [string](#string) |  | ReportID is the unique identifier for the submitted report |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams defines the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module (defaults to x/gov) |
 | `params` | [Params](#virtengine.fraud.v1.Params) |  | Params are the new parameters |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

 
 <a name="virtengine.fraud.v1.MsgUpdateReportStatus"></a>

 ### MsgUpdateReportStatus
 MsgUpdateReportStatus defines the message for updating a report's status

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `moderator` | [string](#string) |  | Moderator is the moderator performing the update |
 | `report_id` | [string](#string) |  | ReportID is the fraud report ID |
 | `new_status` | [FraudReportStatus](#virtengine.fraud.v1.FraudReportStatus) |  | NewStatus is the new status for the report |
 | `notes` | [string](#string) |  | Notes are optional notes about the status change |
 
 

 

 
 <a name="virtengine.fraud.v1.MsgUpdateReportStatusResponse"></a>

 ### MsgUpdateReportStatusResponse
 MsgUpdateReportStatusResponse is the response for MsgUpdateReportStatus

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.fraud.v1.Msg"></a>

 ### Msg
 Msg defines the fraud Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `SubmitFraudReport` | [MsgSubmitFraudReport](#virtengine.fraud.v1.MsgSubmitFraudReport) | [MsgSubmitFraudReportResponse](#virtengine.fraud.v1.MsgSubmitFraudReportResponse) | SubmitFraudReport submits a new fraud report | |
 | `AssignModerator` | [MsgAssignModerator](#virtengine.fraud.v1.MsgAssignModerator) | [MsgAssignModeratorResponse](#virtengine.fraud.v1.MsgAssignModeratorResponse) | AssignModerator assigns a moderator to a fraud report | |
 | `UpdateReportStatus` | [MsgUpdateReportStatus](#virtengine.fraud.v1.MsgUpdateReportStatus) | [MsgUpdateReportStatusResponse](#virtengine.fraud.v1.MsgUpdateReportStatusResponse) | UpdateReportStatus updates the status of a fraud report | |
 | `ResolveFraudReport` | [MsgResolveFraudReport](#virtengine.fraud.v1.MsgResolveFraudReport) | [MsgResolveFraudReportResponse](#virtengine.fraud.v1.MsgResolveFraudReportResponse) | ResolveFraudReport resolves a fraud report with action | |
 | `RejectFraudReport` | [MsgRejectFraudReport](#virtengine.fraud.v1.MsgRejectFraudReport) | [MsgRejectFraudReportResponse](#virtengine.fraud.v1.MsgRejectFraudReportResponse) | RejectFraudReport rejects a fraud report | |
 | `EscalateFraudReport` | [MsgEscalateFraudReport](#virtengine.fraud.v1.MsgEscalateFraudReport) | [MsgEscalateFraudReportResponse](#virtengine.fraud.v1.MsgEscalateFraudReportResponse) | EscalateFraudReport escalates a fraud report to admin | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.fraud.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.fraud.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/hpc/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/hpc/v1/types.proto
 

 
 <a name="virtengine.hpc.v1.ClusterCandidate"></a>

 ### ClusterCandidate
 ClusterCandidate represents a candidate cluster for scheduling

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster_id` | [string](#string) |  |  |
 | `region` | [string](#string) |  |  |
 | `avg_latency_ms` | [int64](#int64) |  |  |
 | `available_nodes` | [int32](#int32) |  |  |
 | `latency_score` | [string](#string) |  |  |
 | `capacity_score` | [string](#string) |  |  |
 | `combined_score` | [string](#string) |  |  |
 | `eligible` | [bool](#bool) |  |  |
 | `ineligibility_reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.ClusterMetadata"></a>

 ### ClusterMetadata
 ClusterMetadata contains additional cluster metadata

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `total_cpu_cores` | [int64](#int64) |  |  |
 | `total_memory_gb` | [int64](#int64) |  |  |
 | `total_gpus` | [int64](#int64) |  |  |
 | `gpu_types` | [string](#string) | repeated |  |
 | `interconnect_type` | [string](#string) |  |  |
 | `storage_type` | [string](#string) |  |  |
 | `total_storage_gb` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.DataReference"></a>

 ### DataReference
 DataReference references external data

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reference_id` | [string](#string) |  |  |
 | `type` | [string](#string) |  |  |
 | `uri` | [string](#string) |  |  |
 | `encrypted` | [bool](#bool) |  |  |
 | `checksum` | [string](#string) |  |  |
 | `size_bytes` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCCluster"></a>

 ### HPCCluster
 HPCCluster represents a SLURM cluster registered on-chain

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster_id` | [string](#string) |  |  |
 | `provider_address` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `state` | [ClusterState](#virtengine.hpc.v1.ClusterState) |  |  |
 | `partitions` | [Partition](#virtengine.hpc.v1.Partition) | repeated |  |
 | `total_nodes` | [int32](#int32) |  |  |
 | `available_nodes` | [int32](#int32) |  |  |
 | `region` | [string](#string) |  |  |
 | `cluster_metadata` | [ClusterMetadata](#virtengine.hpc.v1.ClusterMetadata) |  |  |
 | `slurm_version` | [string](#string) |  |  |
 | `kubernetes_cluster_id` | [string](#string) |  |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCDispute"></a>

 ### HPCDispute
 HPCDispute represents a dispute for HPC rewards/usage

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `dispute_id` | [string](#string) |  |  |
 | `job_id` | [string](#string) |  |  |
 | `reward_id` | [string](#string) |  |  |
 | `disputer_address` | [string](#string) |  |  |
 | `dispute_type` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `evidence` | [string](#string) |  |  |
 | `status` | [DisputeStatus](#virtengine.hpc.v1.DisputeStatus) |  |  |
 | `resolution` | [string](#string) |  |  |
 | `resolver_address` | [string](#string) |  |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `resolved_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCJob"></a>

 ### HPCJob
 HPCJob represents an HPC job request

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 | `offering_id` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `provider_address` | [string](#string) |  |  |
 | `customer_address` | [string](#string) |  |  |
 | `slurm_job_id` | [string](#string) |  |  |
 | `state` | [JobState](#virtengine.hpc.v1.JobState) |  |  |
 | `queue_name` | [string](#string) |  |  |
 | `workload_spec` | [JobWorkloadSpec](#virtengine.hpc.v1.JobWorkloadSpec) |  |  |
 | `resources` | [JobResources](#virtengine.hpc.v1.JobResources) |  |  |
 | `data_references` | [DataReference](#virtengine.hpc.v1.DataReference) | repeated |  |
 | `encrypted_inputs_pointer` | [string](#string) |  |  |
 | `encrypted_outputs_pointer` | [string](#string) |  |  |
 | `max_runtime_seconds` | [int64](#int64) |  |  |
 | `agreed_price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `escrow_id` | [string](#string) |  |  |
 | `scheduling_decision_id` | [string](#string) |  |  |
 | `status_message` | [string](#string) |  |  |
 | `exit_code` | [int32](#int32) |  |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `queued_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `started_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `completed_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCOffering"></a>

 ### HPCOffering
 HPCOffering represents an HPC service offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offering_id` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `provider_address` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `queue_options` | [QueueOption](#virtengine.hpc.v1.QueueOption) | repeated |  |
 | `pricing` | [HPCPricing](#virtengine.hpc.v1.HPCPricing) |  |  |
 | `required_identity_threshold` | [int32](#int32) |  |  |
 | `max_runtime_seconds` | [int64](#int64) |  |  |
 | `preconfigured_workloads` | [PreconfiguredWorkload](#virtengine.hpc.v1.PreconfiguredWorkload) | repeated |  |
 | `supports_custom_workloads` | [bool](#bool) |  |  |
 | `active` | [bool](#bool) |  |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCPricing"></a>

 ### HPCPricing
 HPCPricing contains HPC pricing information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `base_node_hour_price` | [string](#string) |  |  |
 | `cpu_core_hour_price` | [string](#string) |  |  |
 | `gpu_hour_price` | [string](#string) |  |  |
 | `memory_gb_hour_price` | [string](#string) |  |  |
 | `storage_gb_price` | [string](#string) |  |  |
 | `network_gb_price` | [string](#string) |  |  |
 | `currency` | [string](#string) |  |  |
 | `minimum_charge` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCRewardRecipient"></a>

 ### HPCRewardRecipient
 HPCRewardRecipient represents a recipient of HPC rewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `recipient_type` | [string](#string) |  |  |
 | `node_id` | [string](#string) |  |  |
 | `contribution_weight` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCRewardRecord"></a>

 ### HPCRewardRecord
 HPCRewardRecord represents a reward distribution for HPC contribution

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reward_id` | [string](#string) |  |  |
 | `job_id` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `source` | [HPCRewardSource](#virtengine.hpc.v1.HPCRewardSource) |  |  |
 | `total_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `recipients` | [HPCRewardRecipient](#virtengine.hpc.v1.HPCRewardRecipient) | repeated |  |
 | `referenced_usage_records` | [string](#string) | repeated |  |
 | `job_completion_status` | [JobState](#virtengine.hpc.v1.JobState) |  |  |
 | `formula_version` | [string](#string) |  |  |
 | `calculation_details` | [RewardCalculationDetails](#virtengine.hpc.v1.RewardCalculationDetails) |  |  |
 | `disputed` | [bool](#bool) |  |  |
 | `dispute_id` | [string](#string) |  |  |
 | `issued_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.HPCUsageMetrics"></a>

 ### HPCUsageMetrics
 HPCUsageMetrics contains usage metrics for an HPC job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `wall_clock_seconds` | [int64](#int64) |  |  |
 | `cpu_core_seconds` | [int64](#int64) |  |  |
 | `memory_gb_seconds` | [int64](#int64) |  |  |
 | `gpu_seconds` | [int64](#int64) |  |  |
 | `storage_gb_hours` | [int64](#int64) |  |  |
 | `network_bytes_in` | [int64](#int64) |  |  |
 | `network_bytes_out` | [int64](#int64) |  |  |
 | `node_hours` | [int64](#int64) |  |  |
 | `nodes_used` | [int32](#int32) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.JobAccounting"></a>

 ### JobAccounting
 JobAccounting represents accounting data for an HPC job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `provider_address` | [string](#string) |  |  |
 | `customer_address` | [string](#string) |  |  |
 | `usage_metrics` | [HPCUsageMetrics](#virtengine.hpc.v1.HPCUsageMetrics) |  |  |
 | `total_cost` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `provider_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `node_rewards` | [NodeReward](#virtengine.hpc.v1.NodeReward) | repeated |  |
 | `platform_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `settlement_status` | [string](#string) |  |  |
 | `settlement_id` | [string](#string) |  |  |
 | `signed_usage_record_ids` | [string](#string) | repeated |  |
 | `job_completion_status` | [JobState](#virtengine.hpc.v1.JobState) |  |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `finalized_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.JobResources"></a>

 ### JobResources
 JobResources defines resource requirements for an HPC job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `nodes` | [int32](#int32) |  |  |
 | `cpu_cores_per_node` | [int32](#int32) |  |  |
 | `memory_gb_per_node` | [int32](#int32) |  |  |
 | `gpus_per_node` | [int32](#int32) |  |  |
 | `storage_gb` | [int32](#int32) |  |  |
 | `gpu_type` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.JobWorkloadSpec"></a>

 ### JobWorkloadSpec
 JobWorkloadSpec defines the workload for an HPC job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `container_image` | [string](#string) |  |  |
 | `command` | [string](#string) |  |  |
 | `arguments` | [string](#string) | repeated |  |
 | `environment` | [JobWorkloadSpec.EnvironmentEntry](#virtengine.hpc.v1.JobWorkloadSpec.EnvironmentEntry) | repeated |  |
 | `working_directory` | [string](#string) |  |  |
 | `preconfigured_workload_id` | [string](#string) |  |  |
 | `is_preconfigured` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.JobWorkloadSpec.EnvironmentEntry"></a>

 ### JobWorkloadSpec.EnvironmentEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.LatencyMeasurement"></a>

 ### LatencyMeasurement
 LatencyMeasurement contains latency measurement to another node

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `target_node_id` | [string](#string) |  |  |
 | `latency_ms` | [int64](#int64) |  |  |
 | `measured_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.NodeMetadata"></a>

 ### NodeMetadata
 NodeMetadata contains metadata about a compute node

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `node_id` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `provider_address` | [string](#string) |  |  |
 | `region` | [string](#string) |  |  |
 | `datacenter` | [string](#string) |  |  |
 | `latency_measurements` | [LatencyMeasurement](#virtengine.hpc.v1.LatencyMeasurement) | repeated |  |
 | `avg_latency_ms` | [int64](#int64) |  |  |
 | `network_bandwidth_mbps` | [int64](#int64) |  |  |
 | `resources` | [NodeResources](#virtengine.hpc.v1.NodeResources) |  |  |
 | `active` | [bool](#bool) |  |  |
 | `last_heartbeat` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `joined_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.NodeResources"></a>

 ### NodeResources
 NodeResources contains node resource capacity

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cpu_cores` | [int32](#int32) |  |  |
 | `memory_gb` | [int32](#int32) |  |  |
 | `gpus` | [int32](#int32) |  |  |
 | `gpu_type` | [string](#string) |  |  |
 | `storage_gb` | [int32](#int32) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.NodeReward"></a>

 ### NodeReward
 NodeReward represents a reward for a specific node

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `node_id` | [string](#string) |  |  |
 | `provider_address` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `contribution_weight` | [string](#string) |  |  |
 | `usage_seconds` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.Params"></a>

 ### Params
 Params defines the parameters for the HPC module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `platform_fee_rate` | [string](#string) |  |  |
 | `provider_reward_rate` | [string](#string) |  |  |
 | `node_reward_rate` | [string](#string) |  |  |
 | `min_job_duration_seconds` | [int64](#int64) |  |  |
 | `max_job_duration_seconds` | [int64](#int64) |  |  |
 | `default_identity_threshold` | [int32](#int32) |  |  |
 | `cluster_heartbeat_timeout` | [int64](#int64) |  |  |
 | `node_heartbeat_timeout` | [int64](#int64) |  |  |
 | `latency_weight_factor` | [string](#string) |  |  |
 | `capacity_weight_factor` | [string](#string) |  |  |
 | `max_latency_ms` | [int64](#int64) |  |  |
 | `dispute_resolution_period` | [int64](#int64) |  |  |
 | `reward_formula_version` | [string](#string) |  |  |
 | `enable_proximity_clustering` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.Partition"></a>

 ### Partition
 Partition represents a SLURM partition/queue

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `name` | [string](#string) |  |  |
 | `nodes` | [int32](#int32) |  |  |
 | `max_runtime` | [int64](#int64) |  |  |
 | `default_runtime` | [int64](#int64) |  |  |
 | `max_nodes` | [int32](#int32) |  |  |
 | `features` | [string](#string) | repeated |  |
 | `priority` | [int32](#int32) |  |  |
 | `state` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.PreconfiguredWorkload"></a>

 ### PreconfiguredWorkload
 PreconfiguredWorkload represents a pre-approved workload

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `workload_id` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `container_image` | [string](#string) |  |  |
 | `default_command` | [string](#string) |  |  |
 | `required_resources` | [JobResources](#virtengine.hpc.v1.JobResources) |  |  |
 | `category` | [string](#string) |  |  |
 | `version` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueueOption"></a>

 ### QueueOption
 QueueOption represents a queue/partition option in an offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `partition_name` | [string](#string) |  |  |
 | `display_name` | [string](#string) |  |  |
 | `max_nodes` | [int32](#int32) |  |  |
 | `max_runtime` | [int64](#int64) |  |  |
 | `features` | [string](#string) | repeated |  |
 | `price_multiplier` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.RewardCalculationDetails"></a>

 ### RewardCalculationDetails
 RewardCalculationDetails contains transparency data for reward calculation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `total_usage_value` | [string](#string) |  |  |
 | `reward_pool_contribution` | [string](#string) |  |  |
 | `platform_fee_rate` | [string](#string) |  |  |
 | `node_contribution_formula` | [string](#string) |  |  |
 | `input_metrics` | [RewardCalculationDetails.InputMetricsEntry](#virtengine.hpc.v1.RewardCalculationDetails.InputMetricsEntry) | repeated |  |
 
 

 

 
 <a name="virtengine.hpc.v1.RewardCalculationDetails.InputMetricsEntry"></a>

 ### RewardCalculationDetails.InputMetricsEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.SchedulingDecision"></a>

 ### SchedulingDecision
 SchedulingDecision records the decision trail for job scheduling

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `decision_id` | [string](#string) |  |  |
 | `job_id` | [string](#string) |  |  |
 | `selected_cluster_id` | [string](#string) |  |  |
 | `candidate_clusters` | [ClusterCandidate](#virtengine.hpc.v1.ClusterCandidate) | repeated |  |
 | `decision_reason` | [string](#string) |  |  |
 | `is_fallback` | [bool](#bool) |  |  |
 | `fallback_reason` | [string](#string) |  |  |
 | `latency_score` | [string](#string) |  |  |
 | `capacity_score` | [string](#string) |  |  |
 | `combined_score` | [string](#string) |  |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.hpc.v1.ClusterState"></a>

 ### ClusterState
 ClusterState represents the state of an HPC cluster

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | CLUSTER_STATE_UNSPECIFIED | 0 | CLUSTER_STATE_UNSPECIFIED represents an unspecified cluster state |
 | CLUSTER_STATE_PENDING | 1 | CLUSTER_STATE_PENDING indicates the cluster is pending registration |
 | CLUSTER_STATE_ACTIVE | 2 | CLUSTER_STATE_ACTIVE indicates the cluster is active and accepting jobs |
 | CLUSTER_STATE_DRAINING | 3 | CLUSTER_STATE_DRAINING indicates the cluster is draining (not accepting new jobs) |
 | CLUSTER_STATE_OFFLINE | 4 | CLUSTER_STATE_OFFLINE indicates the cluster is offline |
 | CLUSTER_STATE_DEREGISTERED | 5 | CLUSTER_STATE_DEREGISTERED indicates the cluster has been deregistered |
 

 
 <a name="virtengine.hpc.v1.DisputeStatus"></a>

 ### DisputeStatus
 DisputeStatus indicates the status of a dispute

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | DISPUTE_STATUS_UNSPECIFIED | 0 | DISPUTE_STATUS_UNSPECIFIED represents an unspecified dispute status |
 | DISPUTE_STATUS_PENDING | 1 | DISPUTE_STATUS_PENDING indicates the dispute is pending |
 | DISPUTE_STATUS_UNDER_REVIEW | 2 | DISPUTE_STATUS_UNDER_REVIEW indicates the dispute is under review |
 | DISPUTE_STATUS_RESOLVED | 3 | DISPUTE_STATUS_RESOLVED indicates the dispute is resolved |
 | DISPUTE_STATUS_REJECTED | 4 | DISPUTE_STATUS_REJECTED indicates the dispute was rejected |
 

 
 <a name="virtengine.hpc.v1.HPCRewardSource"></a>

 ### HPCRewardSource
 HPCRewardSource indicates the source of HPC rewards

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | HPC_REWARD_SOURCE_UNSPECIFIED | 0 | HPC_REWARD_SOURCE_UNSPECIFIED represents an unspecified reward source |
 | HPC_REWARD_SOURCE_JOB_COMPLETION | 1 | HPC_REWARD_SOURCE_JOB_COMPLETION is for completed job rewards |
 | HPC_REWARD_SOURCE_USAGE | 2 | HPC_REWARD_SOURCE_USAGE is for usage-based rewards |
 | HPC_REWARD_SOURCE_BONUS | 3 | HPC_REWARD_SOURCE_BONUS is for bonus rewards |
 

 
 <a name="virtengine.hpc.v1.JobState"></a>

 ### JobState
 JobState represents the state of an HPC job

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | JOB_STATE_UNSPECIFIED | 0 | JOB_STATE_UNSPECIFIED represents an unspecified job state |
 | JOB_STATE_PENDING | 1 | JOB_STATE_PENDING indicates the job is pending |
 | JOB_STATE_QUEUED | 2 | JOB_STATE_QUEUED indicates the job is queued in SLURM |
 | JOB_STATE_RUNNING | 3 | JOB_STATE_RUNNING indicates the job is running |
 | JOB_STATE_COMPLETED | 4 | JOB_STATE_COMPLETED indicates the job completed successfully |
 | JOB_STATE_FAILED | 5 | JOB_STATE_FAILED indicates the job failed |
 | JOB_STATE_CANCELLED | 6 | JOB_STATE_CANCELLED indicates the job was cancelled |
 | JOB_STATE_TIMEOUT | 7 | JOB_STATE_TIMEOUT indicates the job timed out |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/hpc/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/hpc/v1/genesis.proto
 

 
 <a name="virtengine.hpc.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the hpc module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.hpc.v1.Params) |  | Params are the module parameters |
 | `clusters` | [HPCCluster](#virtengine.hpc.v1.HPCCluster) | repeated | Clusters are the registered HPC clusters |
 | `offerings` | [HPCOffering](#virtengine.hpc.v1.HPCOffering) | repeated | Offerings are the HPC offerings |
 | `jobs` | [HPCJob](#virtengine.hpc.v1.HPCJob) | repeated | Jobs are the HPC jobs |
 | `job_accountings` | [JobAccounting](#virtengine.hpc.v1.JobAccounting) | repeated | JobAccountings are the job accounting records |
 | `node_metadatas` | [NodeMetadata](#virtengine.hpc.v1.NodeMetadata) | repeated | NodeMetadatas are the node metadata records |
 | `scheduling_decisions` | [SchedulingDecision](#virtengine.hpc.v1.SchedulingDecision) | repeated | SchedulingDecisions are the scheduling decision records |
 | `hpc_rewards` | [HPCRewardRecord](#virtengine.hpc.v1.HPCRewardRecord) | repeated | HPCRewards are the HPC reward records |
 | `disputes` | [HPCDispute](#virtengine.hpc.v1.HPCDispute) | repeated | Disputes are the dispute records |
 | `cluster_sequence` | [uint64](#uint64) |  | ClusterSequence is the next cluster sequence |
 | `offering_sequence` | [uint64](#uint64) |  | OfferingSequence is the next offering sequence |
 | `job_sequence` | [uint64](#uint64) |  | JobSequence is the next job sequence |
 | `decision_sequence` | [uint64](#uint64) |  | DecisionSequence is the next decision sequence |
 | `dispute_sequence` | [uint64](#uint64) |  | DisputeSequence is the next dispute sequence |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/hpc/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/hpc/v1/query.proto
 

 
 <a name="virtengine.hpc.v1.QueryClusterRequest"></a>

 ### QueryClusterRequest
 QueryClusterRequest is the request for Query/Cluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryClusterResponse"></a>

 ### QueryClusterResponse
 QueryClusterResponse is the response for Query/Cluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster` | [HPCCluster](#virtengine.hpc.v1.HPCCluster) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryClustersByProviderRequest"></a>

 ### QueryClustersByProviderRequest
 QueryClustersByProviderRequest is the request for Query/ClustersByProvider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryClustersByProviderResponse"></a>

 ### QueryClustersByProviderResponse
 QueryClustersByProviderResponse is the response for Query/ClustersByProvider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `clusters` | [HPCCluster](#virtengine.hpc.v1.HPCCluster) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryClustersRequest"></a>

 ### QueryClustersRequest
 QueryClustersRequest is the request for Query/Clusters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [ClusterState](#virtengine.hpc.v1.ClusterState) |  |  |
 | `region` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryClustersResponse"></a>

 ### QueryClustersResponse
 QueryClustersResponse is the response for Query/Clusters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `clusters` | [HPCCluster](#virtengine.hpc.v1.HPCCluster) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryDisputeRequest"></a>

 ### QueryDisputeRequest
 QueryDisputeRequest is the request for Query/Dispute

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `dispute_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryDisputeResponse"></a>

 ### QueryDisputeResponse
 QueryDisputeResponse is the response for Query/Dispute

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `dispute` | [HPCDispute](#virtengine.hpc.v1.HPCDispute) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryDisputesRequest"></a>

 ### QueryDisputesRequest
 QueryDisputesRequest is the request for Query/Disputes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `status` | [DisputeStatus](#virtengine.hpc.v1.DisputeStatus) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryDisputesResponse"></a>

 ### QueryDisputesResponse
 QueryDisputesResponse is the response for Query/Disputes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `disputes` | [HPCDispute](#virtengine.hpc.v1.HPCDispute) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobAccountingRequest"></a>

 ### QueryJobAccountingRequest
 QueryJobAccountingRequest is the request for Query/JobAccounting

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobAccountingResponse"></a>

 ### QueryJobAccountingResponse
 QueryJobAccountingResponse is the response for Query/JobAccounting

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `accounting` | [JobAccounting](#virtengine.hpc.v1.JobAccounting) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobRequest"></a>

 ### QueryJobRequest
 QueryJobRequest is the request for Query/Job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobResponse"></a>

 ### QueryJobResponse
 QueryJobResponse is the response for Query/Job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job` | [HPCJob](#virtengine.hpc.v1.HPCJob) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobsByCustomerRequest"></a>

 ### QueryJobsByCustomerRequest
 QueryJobsByCustomerRequest is the request for Query/JobsByCustomer

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `customer_address` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobsByCustomerResponse"></a>

 ### QueryJobsByCustomerResponse
 QueryJobsByCustomerResponse is the response for Query/JobsByCustomer

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `jobs` | [HPCJob](#virtengine.hpc.v1.HPCJob) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobsByProviderRequest"></a>

 ### QueryJobsByProviderRequest
 QueryJobsByProviderRequest is the request for Query/JobsByProvider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobsByProviderResponse"></a>

 ### QueryJobsByProviderResponse
 QueryJobsByProviderResponse is the response for Query/JobsByProvider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `jobs` | [HPCJob](#virtengine.hpc.v1.HPCJob) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobsRequest"></a>

 ### QueryJobsRequest
 QueryJobsRequest is the request for Query/Jobs

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [JobState](#virtengine.hpc.v1.JobState) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryJobsResponse"></a>

 ### QueryJobsResponse
 QueryJobsResponse is the response for Query/Jobs

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `jobs` | [HPCJob](#virtengine.hpc.v1.HPCJob) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryNodeMetadataRequest"></a>

 ### QueryNodeMetadataRequest
 QueryNodeMetadataRequest is the request for Query/NodeMetadata

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `node_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryNodeMetadataResponse"></a>

 ### QueryNodeMetadataResponse
 QueryNodeMetadataResponse is the response for Query/NodeMetadata

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `node` | [NodeMetadata](#virtengine.hpc.v1.NodeMetadata) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryNodesByClusterRequest"></a>

 ### QueryNodesByClusterRequest
 QueryNodesByClusterRequest is the request for Query/NodesByCluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster_id` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryNodesByClusterResponse"></a>

 ### QueryNodesByClusterResponse
 QueryNodesByClusterResponse is the response for Query/NodesByCluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `nodes` | [NodeMetadata](#virtengine.hpc.v1.NodeMetadata) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryOfferingRequest"></a>

 ### QueryOfferingRequest
 QueryOfferingRequest is the request for Query/Offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offering_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryOfferingResponse"></a>

 ### QueryOfferingResponse
 QueryOfferingResponse is the response for Query/Offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offering` | [HPCOffering](#virtengine.hpc.v1.HPCOffering) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryOfferingsByClusterRequest"></a>

 ### QueryOfferingsByClusterRequest
 QueryOfferingsByClusterRequest is the request for Query/OfferingsByCluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster_id` | [string](#string) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryOfferingsByClusterResponse"></a>

 ### QueryOfferingsByClusterResponse
 QueryOfferingsByClusterResponse is the response for Query/OfferingsByCluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offerings` | [HPCOffering](#virtengine.hpc.v1.HPCOffering) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryOfferingsRequest"></a>

 ### QueryOfferingsRequest
 QueryOfferingsRequest is the request for Query/Offerings

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `active_only` | [bool](#bool) |  |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryOfferingsResponse"></a>

 ### QueryOfferingsResponse
 QueryOfferingsResponse is the response for Query/Offerings

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offerings` | [HPCOffering](#virtengine.hpc.v1.HPCOffering) | repeated |  |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for Query/Params

 

 

 
 <a name="virtengine.hpc.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for Query/Params

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.hpc.v1.Params) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryRewardRequest"></a>

 ### QueryRewardRequest
 QueryRewardRequest is the request for Query/Reward

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reward_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryRewardResponse"></a>

 ### QueryRewardResponse
 QueryRewardResponse is the response for Query/Reward

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reward` | [HPCRewardRecord](#virtengine.hpc.v1.HPCRewardRecord) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryRewardsByJobRequest"></a>

 ### QueryRewardsByJobRequest
 QueryRewardsByJobRequest is the request for Query/RewardsByJob

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QueryRewardsByJobResponse"></a>

 ### QueryRewardsByJobResponse
 QueryRewardsByJobResponse is the response for Query/RewardsByJob

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `rewards` | [HPCRewardRecord](#virtengine.hpc.v1.HPCRewardRecord) | repeated |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QuerySchedulingDecisionByJobRequest"></a>

 ### QuerySchedulingDecisionByJobRequest
 QuerySchedulingDecisionByJobRequest is the request for Query/SchedulingDecisionByJob

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QuerySchedulingDecisionByJobResponse"></a>

 ### QuerySchedulingDecisionByJobResponse
 QuerySchedulingDecisionByJobResponse is the response for Query/SchedulingDecisionByJob

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `decision` | [SchedulingDecision](#virtengine.hpc.v1.SchedulingDecision) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QuerySchedulingDecisionRequest"></a>

 ### QuerySchedulingDecisionRequest
 QuerySchedulingDecisionRequest is the request for Query/SchedulingDecision

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `decision_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.QuerySchedulingDecisionResponse"></a>

 ### QuerySchedulingDecisionResponse
 QuerySchedulingDecisionResponse is the response for Query/SchedulingDecision

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `decision` | [SchedulingDecision](#virtengine.hpc.v1.SchedulingDecision) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.hpc.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the hpc module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Cluster` | [QueryClusterRequest](#virtengine.hpc.v1.QueryClusterRequest) | [QueryClusterResponse](#virtengine.hpc.v1.QueryClusterResponse) | Cluster returns a cluster by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/cluster/{cluster_id}|
 | `Clusters` | [QueryClustersRequest](#virtengine.hpc.v1.QueryClustersRequest) | [QueryClustersResponse](#virtengine.hpc.v1.QueryClustersResponse) | Clusters returns all clusters with optional filters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/clusters|
 | `ClustersByProvider` | [QueryClustersByProviderRequest](#virtengine.hpc.v1.QueryClustersByProviderRequest) | [QueryClustersByProviderResponse](#virtengine.hpc.v1.QueryClustersByProviderResponse) | ClustersByProvider returns clusters owned by a provider buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/clusters/provider/{provider_address}|
 | `Offering` | [QueryOfferingRequest](#virtengine.hpc.v1.QueryOfferingRequest) | [QueryOfferingResponse](#virtengine.hpc.v1.QueryOfferingResponse) | Offering returns an offering by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/offering/{offering_id}|
 | `Offerings` | [QueryOfferingsRequest](#virtengine.hpc.v1.QueryOfferingsRequest) | [QueryOfferingsResponse](#virtengine.hpc.v1.QueryOfferingsResponse) | Offerings returns all offerings with optional filters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/offerings|
 | `OfferingsByCluster` | [QueryOfferingsByClusterRequest](#virtengine.hpc.v1.QueryOfferingsByClusterRequest) | [QueryOfferingsByClusterResponse](#virtengine.hpc.v1.QueryOfferingsByClusterResponse) | OfferingsByCluster returns offerings for a cluster buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/offerings/cluster/{cluster_id}|
 | `Job` | [QueryJobRequest](#virtengine.hpc.v1.QueryJobRequest) | [QueryJobResponse](#virtengine.hpc.v1.QueryJobResponse) | Job returns a job by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/job/{job_id}|
 | `Jobs` | [QueryJobsRequest](#virtengine.hpc.v1.QueryJobsRequest) | [QueryJobsResponse](#virtengine.hpc.v1.QueryJobsResponse) | Jobs returns all jobs with optional filters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/jobs|
 | `JobsByCustomer` | [QueryJobsByCustomerRequest](#virtengine.hpc.v1.QueryJobsByCustomerRequest) | [QueryJobsByCustomerResponse](#virtengine.hpc.v1.QueryJobsByCustomerResponse) | JobsByCustomer returns jobs submitted by a customer buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/jobs/customer/{customer_address}|
 | `JobsByProvider` | [QueryJobsByProviderRequest](#virtengine.hpc.v1.QueryJobsByProviderRequest) | [QueryJobsByProviderResponse](#virtengine.hpc.v1.QueryJobsByProviderResponse) | JobsByProvider returns jobs handled by a provider buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/jobs/provider/{provider_address}|
 | `JobAccounting` | [QueryJobAccountingRequest](#virtengine.hpc.v1.QueryJobAccountingRequest) | [QueryJobAccountingResponse](#virtengine.hpc.v1.QueryJobAccountingResponse) | JobAccounting returns accounting data for a job buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/job/{job_id}/accounting|
 | `NodeMetadata` | [QueryNodeMetadataRequest](#virtengine.hpc.v1.QueryNodeMetadataRequest) | [QueryNodeMetadataResponse](#virtengine.hpc.v1.QueryNodeMetadataResponse) | NodeMetadata returns metadata for a node buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/node/{node_id}|
 | `NodesByCluster` | [QueryNodesByClusterRequest](#virtengine.hpc.v1.QueryNodesByClusterRequest) | [QueryNodesByClusterResponse](#virtengine.hpc.v1.QueryNodesByClusterResponse) | NodesByCluster returns nodes in a cluster buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/nodes/cluster/{cluster_id}|
 | `SchedulingDecision` | [QuerySchedulingDecisionRequest](#virtengine.hpc.v1.QuerySchedulingDecisionRequest) | [QuerySchedulingDecisionResponse](#virtengine.hpc.v1.QuerySchedulingDecisionResponse) | SchedulingDecision returns a scheduling decision by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/scheduling/{decision_id}|
 | `SchedulingDecisionByJob` | [QuerySchedulingDecisionByJobRequest](#virtengine.hpc.v1.QuerySchedulingDecisionByJobRequest) | [QuerySchedulingDecisionByJobResponse](#virtengine.hpc.v1.QuerySchedulingDecisionByJobResponse) | SchedulingDecisionByJob returns the scheduling decision for a job buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/job/{job_id}/scheduling|
 | `Reward` | [QueryRewardRequest](#virtengine.hpc.v1.QueryRewardRequest) | [QueryRewardResponse](#virtengine.hpc.v1.QueryRewardResponse) | Reward returns a reward record by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/reward/{reward_id}|
 | `RewardsByJob` | [QueryRewardsByJobRequest](#virtengine.hpc.v1.QueryRewardsByJobRequest) | [QueryRewardsByJobResponse](#virtengine.hpc.v1.QueryRewardsByJobResponse) | RewardsByJob returns rewards for a job buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/job/{job_id}/rewards|
 | `Dispute` | [QueryDisputeRequest](#virtengine.hpc.v1.QueryDisputeRequest) | [QueryDisputeResponse](#virtengine.hpc.v1.QueryDisputeResponse) | Dispute returns a dispute by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/dispute/{dispute_id}|
 | `Disputes` | [QueryDisputesRequest](#virtengine.hpc.v1.QueryDisputesRequest) | [QueryDisputesResponse](#virtengine.hpc.v1.QueryDisputesResponse) | Disputes returns all disputes with optional filters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/disputes|
 | `Params` | [QueryParamsRequest](#virtengine.hpc.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.hpc.v1.QueryParamsResponse) | Params returns the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/hpc/v1/params|
 
  <!-- end services -->

 
 
 <a name="virtengine/hpc/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/hpc/v1/tx.proto
 

 
 <a name="virtengine.hpc.v1.MsgCancelJob"></a>

 ### MsgCancelJob
 MsgCancelJob cancels an HPC job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `requester_address` | [string](#string) |  |  |
 | `job_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgCancelJobResponse"></a>

 ### MsgCancelJobResponse
 MsgCancelJobResponse is the response for MsgCancelJob

 

 

 
 <a name="virtengine.hpc.v1.MsgCreateOffering"></a>

 ### MsgCreateOffering
 MsgCreateOffering creates a new HPC offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `queue_options` | [QueueOption](#virtengine.hpc.v1.QueueOption) | repeated |  |
 | `pricing` | [HPCPricing](#virtengine.hpc.v1.HPCPricing) |  |  |
 | `required_identity_threshold` | [int32](#int32) |  |  |
 | `max_runtime_seconds` | [int64](#int64) |  |  |
 | `preconfigured_workloads` | [PreconfiguredWorkload](#virtengine.hpc.v1.PreconfiguredWorkload) | repeated |  |
 | `supports_custom_workloads` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgCreateOfferingResponse"></a>

 ### MsgCreateOfferingResponse
 MsgCreateOfferingResponse is the response for MsgCreateOffering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offering_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgDeregisterCluster"></a>

 ### MsgDeregisterCluster
 MsgDeregisterCluster deregisters an HPC cluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgDeregisterClusterResponse"></a>

 ### MsgDeregisterClusterResponse
 MsgDeregisterClusterResponse is the response for MsgDeregisterCluster

 

 

 
 <a name="virtengine.hpc.v1.MsgFlagDispute"></a>

 ### MsgFlagDispute
 MsgFlagDispute flags a dispute for moderation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `disputer_address` | [string](#string) |  |  |
 | `job_id` | [string](#string) |  |  |
 | `reward_id` | [string](#string) |  |  |
 | `dispute_type` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `evidence` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgFlagDisputeResponse"></a>

 ### MsgFlagDisputeResponse
 MsgFlagDisputeResponse is the response for MsgFlagDispute

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `dispute_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgRegisterCluster"></a>

 ### MsgRegisterCluster
 MsgRegisterCluster registers a new HPC cluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `region` | [string](#string) |  |  |
 | `partitions` | [Partition](#virtengine.hpc.v1.Partition) | repeated |  |
 | `total_nodes` | [int32](#int32) |  |  |
 | `cluster_metadata` | [ClusterMetadata](#virtengine.hpc.v1.ClusterMetadata) |  |  |
 | `slurm_version` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgRegisterClusterResponse"></a>

 ### MsgRegisterClusterResponse
 MsgRegisterClusterResponse is the response for MsgRegisterCluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `cluster_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgReportJobStatus"></a>

 ### MsgReportJobStatus
 MsgReportJobStatus reports job status from the provider daemon

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `job_id` | [string](#string) |  |  |
 | `slurm_job_id` | [string](#string) |  |  |
 | `state` | [JobState](#virtengine.hpc.v1.JobState) |  |  |
 | `status_message` | [string](#string) |  |  |
 | `exit_code` | [int32](#int32) |  |  |
 | `usage_metrics` | [HPCUsageMetrics](#virtengine.hpc.v1.HPCUsageMetrics) |  |  |
 | `signature` | [string](#string) |  |  |
 | `signed_timestamp` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgReportJobStatusResponse"></a>

 ### MsgReportJobStatusResponse
 MsgReportJobStatusResponse is the response for MsgReportJobStatus

 

 

 
 <a name="virtengine.hpc.v1.MsgResolveDispute"></a>

 ### MsgResolveDispute
 MsgResolveDispute resolves a dispute (moderator only)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `resolver_address` | [string](#string) |  |  |
 | `dispute_id` | [string](#string) |  |  |
 | `status` | [DisputeStatus](#virtengine.hpc.v1.DisputeStatus) |  |  |
 | `resolution` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgResolveDisputeResponse"></a>

 ### MsgResolveDisputeResponse
 MsgResolveDisputeResponse is the response for MsgResolveDispute

 

 

 
 <a name="virtengine.hpc.v1.MsgSubmitJob"></a>

 ### MsgSubmitJob
 MsgSubmitJob submits a new HPC job

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `customer_address` | [string](#string) |  |  |
 | `offering_id` | [string](#string) |  |  |
 | `queue_name` | [string](#string) |  |  |
 | `workload_spec` | [JobWorkloadSpec](#virtengine.hpc.v1.JobWorkloadSpec) |  |  |
 | `resources` | [JobResources](#virtengine.hpc.v1.JobResources) |  |  |
 | `data_references` | [DataReference](#virtengine.hpc.v1.DataReference) | repeated |  |
 | `encrypted_inputs_pointer` | [string](#string) |  |  |
 | `encrypted_outputs_pointer` | [string](#string) |  |  |
 | `max_runtime_seconds` | [int64](#int64) |  |  |
 | `max_price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgSubmitJobResponse"></a>

 ### MsgSubmitJobResponse
 MsgSubmitJobResponse is the response for MsgSubmitJob

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `job_id` | [string](#string) |  |  |
 | `escrow_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateCluster"></a>

 ### MsgUpdateCluster
 MsgUpdateCluster updates an existing HPC cluster

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `state` | [ClusterState](#virtengine.hpc.v1.ClusterState) |  |  |
 | `partitions` | [Partition](#virtengine.hpc.v1.Partition) | repeated |  |
 | `total_nodes` | [int32](#int32) |  |  |
 | `available_nodes` | [int32](#int32) |  |  |
 | `cluster_metadata` | [ClusterMetadata](#virtengine.hpc.v1.ClusterMetadata) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateClusterResponse"></a>

 ### MsgUpdateClusterResponse
 MsgUpdateClusterResponse is the response for MsgUpdateCluster

 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateNodeMetadata"></a>

 ### MsgUpdateNodeMetadata
 MsgUpdateNodeMetadata updates node metadata

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `node_id` | [string](#string) |  |  |
 | `cluster_id` | [string](#string) |  |  |
 | `region` | [string](#string) |  |  |
 | `datacenter` | [string](#string) |  |  |
 | `latency_measurements` | [LatencyMeasurement](#virtengine.hpc.v1.LatencyMeasurement) | repeated |  |
 | `network_bandwidth_mbps` | [int64](#int64) |  |  |
 | `resources` | [NodeResources](#virtengine.hpc.v1.NodeResources) |  |  |
 | `active` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateNodeMetadataResponse"></a>

 ### MsgUpdateNodeMetadataResponse
 MsgUpdateNodeMetadataResponse is the response for MsgUpdateNodeMetadata

 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateOffering"></a>

 ### MsgUpdateOffering
 MsgUpdateOffering updates an existing HPC offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `offering_id` | [string](#string) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `queue_options` | [QueueOption](#virtengine.hpc.v1.QueueOption) | repeated |  |
 | `pricing` | [HPCPricing](#virtengine.hpc.v1.HPCPricing) |  |  |
 | `required_identity_threshold` | [int32](#int32) |  |  |
 | `max_runtime_seconds` | [int64](#int64) |  |  |
 | `preconfigured_workloads` | [PreconfiguredWorkload](#virtengine.hpc.v1.PreconfiguredWorkload) | repeated |  |
 | `supports_custom_workloads` | [bool](#bool) |  |  |
 | `active` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateOfferingResponse"></a>

 ### MsgUpdateOfferingResponse
 MsgUpdateOfferingResponse is the response for MsgUpdateOffering

 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams updates module parameters (governance only)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `params` | [Params](#virtengine.hpc.v1.Params) |  |  |
 
 

 

 
 <a name="virtengine.hpc.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.hpc.v1.Msg"></a>

 ### Msg
 Msg defines the hpc Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `RegisterCluster` | [MsgRegisterCluster](#virtengine.hpc.v1.MsgRegisterCluster) | [MsgRegisterClusterResponse](#virtengine.hpc.v1.MsgRegisterClusterResponse) | RegisterCluster registers a new HPC cluster | |
 | `UpdateCluster` | [MsgUpdateCluster](#virtengine.hpc.v1.MsgUpdateCluster) | [MsgUpdateClusterResponse](#virtengine.hpc.v1.MsgUpdateClusterResponse) | UpdateCluster updates an existing HPC cluster | |
 | `DeregisterCluster` | [MsgDeregisterCluster](#virtengine.hpc.v1.MsgDeregisterCluster) | [MsgDeregisterClusterResponse](#virtengine.hpc.v1.MsgDeregisterClusterResponse) | DeregisterCluster deregisters an HPC cluster | |
 | `CreateOffering` | [MsgCreateOffering](#virtengine.hpc.v1.MsgCreateOffering) | [MsgCreateOfferingResponse](#virtengine.hpc.v1.MsgCreateOfferingResponse) | CreateOffering creates a new HPC offering | |
 | `UpdateOffering` | [MsgUpdateOffering](#virtengine.hpc.v1.MsgUpdateOffering) | [MsgUpdateOfferingResponse](#virtengine.hpc.v1.MsgUpdateOfferingResponse) | UpdateOffering updates an existing HPC offering | |
 | `SubmitJob` | [MsgSubmitJob](#virtengine.hpc.v1.MsgSubmitJob) | [MsgSubmitJobResponse](#virtengine.hpc.v1.MsgSubmitJobResponse) | SubmitJob submits a new HPC job | |
 | `CancelJob` | [MsgCancelJob](#virtengine.hpc.v1.MsgCancelJob) | [MsgCancelJobResponse](#virtengine.hpc.v1.MsgCancelJobResponse) | CancelJob cancels an HPC job | |
 | `ReportJobStatus` | [MsgReportJobStatus](#virtengine.hpc.v1.MsgReportJobStatus) | [MsgReportJobStatusResponse](#virtengine.hpc.v1.MsgReportJobStatusResponse) | ReportJobStatus reports job status from the provider daemon | |
 | `UpdateNodeMetadata` | [MsgUpdateNodeMetadata](#virtengine.hpc.v1.MsgUpdateNodeMetadata) | [MsgUpdateNodeMetadataResponse](#virtengine.hpc.v1.MsgUpdateNodeMetadataResponse) | UpdateNodeMetadata updates node metadata | |
 | `FlagDispute` | [MsgFlagDispute](#virtengine.hpc.v1.MsgFlagDispute) | [MsgFlagDisputeResponse](#virtengine.hpc.v1.MsgFlagDisputeResponse) | FlagDispute flags a dispute for moderation | |
 | `ResolveDispute` | [MsgResolveDispute](#virtengine.hpc.v1.MsgResolveDispute) | [MsgResolveDisputeResponse](#virtengine.hpc.v1.MsgResolveDisputeResponse) | ResolveDispute resolves a dispute (moderator only) | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.hpc.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.hpc.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/market/v1/bid.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1/bid.proto
 

 
 <a name="virtengine.market.v1.BidID"></a>

 ### BidID
 BidID stores owner and all other seq numbers.
A successful bid becomes a Lease(ID).

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "ve1..." |
 | `bseq` | [uint32](#uint32) |  | BSeq (bid sequence) distinguishes multiple bids associated with a single deployment from same provider. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1/order.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1/order.proto
 

 
 <a name="virtengine.market.v1.OrderID"></a>

 ### OrderID
 OrderId stores owner and all other seq numbers.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1/types.proto
 

  <!-- end messages -->

 
 <a name="virtengine.market.v1.LeaseClosedReason"></a>

 ### LeaseClosedReason
 LeaseClosedReason indicates reason bid was closed

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | lease_closed_invalid | 0 | LeaseClosedReasonInvalid represents the default zero value for LeaseClosedReason. This value indicates an uninitialized or invalid lease closure reason and should not be used |
 | lease_closed_owner | 1 | values between 1..9999 indicate owner‑initiated close |
 | lease_closed_reason_unstable | 10000 | values between 10000..19999 are indicating provider initiated close lease_closed_reason_unstable lease workloads have been unstable |
 | lease_closed_reason_decommission | 10001 | lease_closed_reason_decommission provider is being decommissioned |
 | lease_closed_reason_unspecified | 10002 | lease_closed_reason_unspecified provider did not specify reason |
 | lease_closed_reason_manifest_timeout | 10003 | lease_closed_reason_manifest_timeout provider closed leases due to manifest not received |
 | lease_closed_reason_insufficient_funds | 20000 | values between 20000..29999 indicate network‑initiated close |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1/lease.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1/lease.proto
 

 
 <a name="virtengine.market.v1.Lease"></a>

 ### Lease
 Lease stores LeaseID, state of lease and price.
The Lease defines the terms under which the provider allocates resources to fulfill
the tenant's deployment requirements.
Leases are paid from the tenant to the provider through a deposit and withdraw mechanism and are priced in blocks.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LeaseID](#virtengine.market.v1.LeaseID) |  | Id is the unique identifier of the Lease. |
 | `state` | [Lease.State](#virtengine.market.v1.Lease.State) |  | State represents the state of the Lease. |
 | `price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Price holds the settled price for the Lease. |
 | `created_at` | [int64](#int64) |  | CreatedAt is the block height at which the Lease was created. |
 | `closed_on` | [int64](#int64) |  | ClosedOn is the block height at which the Lease was closed. |
 | `reason` | [LeaseClosedReason](#virtengine.market.v1.LeaseClosedReason) |  |  |
 
 

 

 
 <a name="virtengine.market.v1.LeaseID"></a>

 ### LeaseID
 LeaseID stores bid details of lease.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "ve1..." |
 | `bseq` | [uint32](#uint32) |  | BSeq (bid sequence) distinguishes multiple bids associated with a single deployment from same provider. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.market.v1.Lease.State"></a>

 ### Lease.State
 State is an enum which refers to state of lease.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | active | 1 | LeaseActive denotes state for lease active. |
 | insufficient_funds | 2 | LeaseInsufficientFunds denotes state for lease insufficient_funds. |
 | closed | 3 | LeaseClosed denotes state for lease closed. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1/event.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1/event.proto
 

 
 <a name="virtengine.market.v1.EventBidClosed"></a>

 ### EventBidClosed
 EventBidClosed is triggered when a bid is closed.
It contains all the information required to identify a bid.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [BidID](#virtengine.market.v1.BidID) |  | Id is the unique identifier of the Bid. |
 
 

 

 
 <a name="virtengine.market.v1.EventBidCreated"></a>

 ### EventBidCreated
 EventBidCreated is triggered when a bid is created.
It contains all the information required to identify a bid.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [BidID](#virtengine.market.v1.BidID) |  | Id is the unique identifier of the Bid. |
 | `price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Price stated on the Bid. |
 
 

 

 
 <a name="virtengine.market.v1.EventLeaseClosed"></a>

 ### EventLeaseClosed
 EventLeaseClosed is triggered when a lease is closed.
It contains all the information required to identify a lease.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LeaseID](#virtengine.market.v1.LeaseID) |  | Id is the unique identifier of the Lease. |
 | `reason` | [LeaseClosedReason](#virtengine.market.v1.LeaseClosedReason) |  |  |
 
 

 

 
 <a name="virtengine.market.v1.EventLeaseCreated"></a>

 ### EventLeaseCreated
 EventLeaseCreated is triggered when a lease is created.
It contains all the information required to identify a lease.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [LeaseID](#virtengine.market.v1.LeaseID) |  | Id is the unique identifier of the Lease. |
 | `price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Price settled for the lease. |
 
 

 

 
 <a name="virtengine.market.v1.EventOrderClosed"></a>

 ### EventOrderClosed
 EventOrderClosed is triggered when an order is closed.
It contains all the information required to identify an order.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [OrderID](#virtengine.market.v1.OrderID) |  | Id is the unique identifier of the Order. |
 
 

 

 
 <a name="virtengine.market.v1.EventOrderCreated"></a>

 ### EventOrderCreated
 EventOrderCreated is triggered when an order is created.
It contains all the information required to identify an order.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [OrderID](#virtengine.market.v1.OrderID) |  | Id is the unique identifier of the Order. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1/filters.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1/filters.proto
 

 
 <a name="virtengine.market.v1.LeaseFilters"></a>

 ### LeaseFilters
 LeaseFilters defines flags for lease list filtering.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "ve1..." |
 | `state` | [string](#string) |  | State represents the state of the lease. |
 | `bseq` | [uint32](#uint32) |  | BSeq (bid sequence) distinguishes multiple bids associated with a single deployment from same provider. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/resourcesoffer.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/resourcesoffer.proto
 

 
 <a name="virtengine.market.v1beta5.ResourceOffer"></a>

 ### ResourceOffer
 ResourceOffer describes resources that provider is offering
for deployment.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `resources` | [virtengine.base.resources.v1beta4.Resources](#virtengine.base.resources.v1beta4.Resources) |  | Resources holds information about bid resources. |
 | `count` | [uint32](#uint32) |  | Count is the number of resources. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/bid.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/bid.proto
 

 
 <a name="virtengine.market.v1beta5.Bid"></a>

 ### Bid
 Bid stores BidID, state of bid and price.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.BidID](#virtengine.market.v1.BidID) |  | BidID stores owner and all other seq numbers. A successful bid becomes a Lease(ID). |
 | `state` | [Bid.State](#virtengine.market.v1beta5.Bid.State) |  | State represents the state of the Bid. |
 | `price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Price holds the pricing stated on the Bid. |
 | `created_at` | [int64](#int64) |  | CreatedAt is the block height at which the Bid was created. |
 | `resources_offer` | [ResourceOffer](#virtengine.market.v1beta5.ResourceOffer) | repeated | ResourceOffer is a list of offers. |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.market.v1beta5.Bid.State"></a>

 ### Bid.State
 BidState is an enum which refers to state of bid.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | open | 1 | BidOpen denotes state for bid open. |
 | active | 2 | BidMatched denotes state for bid open. |
 | lost | 3 | BidLost denotes state for bid lost. |
 | closed | 4 | BidClosed denotes state for bid closed. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/bidmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/bidmsg.proto
 

 
 <a name="virtengine.market.v1beta5.MsgCloseBid"></a>

 ### MsgCloseBid
 MsgCloseBid defines an SDK message for closing bid.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.BidID](#virtengine.market.v1.BidID) |  | Id is the unique identifier of the Bid. |
 | `reason` | [virtengine.market.v1.LeaseClosedReason](#virtengine.market.v1.LeaseClosedReason) |  |  |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgCloseBidResponse"></a>

 ### MsgCloseBidResponse
 MsgCloseBidResponse defines the Msg/CloseBid response type.

 

 

 
 <a name="virtengine.market.v1beta5.MsgCreateBid"></a>

 ### MsgCreateBid
 MsgCreateBid defines an SDK message for creating Bid.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.BidID](#virtengine.market.v1.BidID) |  |  |
 | `price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | Price holds the pricing stated on the Bid. |
 | `deposit` | [virtengine.base.deposit.v1.Deposit](#virtengine.base.deposit.v1.Deposit) |  | Deposit holds the amount of coins to deposit. |
 | `resources_offer` | [ResourceOffer](#virtengine.market.v1beta5.ResourceOffer) | repeated | ResourceOffer is a list of resource offers. |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgCreateBidResponse"></a>

 ### MsgCreateBidResponse
 MsgCreateBidResponse defines the Msg/CreateBid response type.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/filters.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/filters.proto
 

 
 <a name="virtengine.market.v1beta5.BidFilters"></a>

 ### BidFilters
 BidFilters defines flags for bid list filter.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "ve1..." |
 | `state` | [string](#string) |  | State represents the state of the lease. |
 | `bseq` | [uint32](#uint32) |  | BSeq (bid sequence) distinguishes multiple bids associated with a single deployment from same provider. |
 
 

 

 
 <a name="virtengine.market.v1beta5.OrderFilters"></a>

 ### OrderFilters
 OrderFilters defines flags for order list filter

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "ve1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `state` | [string](#string) |  | State represents the state of the lease. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/params.proto
 

 
 <a name="virtengine.market.v1beta5.Params"></a>

 ### Params
 Params is the params for the x/market module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `bid_min_deposit` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | BidMinDeposit is a parameter for the minimum deposit on a Bid. |
 | `order_max_bids` | [uint32](#uint32) |  | OrderMaxBids is a parameter for the maximum number of bids in an order. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/order.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/order.proto
 

 
 <a name="virtengine.market.v1beta5.Order"></a>

 ### Order
 Order stores orderID, state of order and other details.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.OrderID](#virtengine.market.v1.OrderID) |  | Id is the unique identifier of the order. |
 | `state` | [Order.State](#virtengine.market.v1beta5.Order.State) |  |  |
 | `spec` | [virtengine.deployment.v1beta4.GroupSpec](#virtengine.deployment.v1beta4.GroupSpec) |  |  |
 | `created_at` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.market.v1beta5.Order.State"></a>

 ### Order.State
 State is an enum which refers to state of order.

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | invalid | 0 | Prefix should start with 0 in enum. So declaring dummy state. |
 | open | 1 | OrderOpen denotes state for order open. |
 | active | 2 | OrderMatched denotes state for order matched. |
 | closed | 3 | OrderClosed denotes state for order lost. |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/genesis.proto
 

 
 <a name="virtengine.market.v1beta5.GenesisState"></a>

 ### GenesisState
 GenesisState defines the basic genesis state used by market module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.market.v1beta5.Params) |  | Params holds parameters of the genesis of market. |
 | `orders` | [Order](#virtengine.market.v1beta5.Order) | repeated | Orders is a list of orders in the genesis state. |
 | `leases` | [virtengine.market.v1.Lease](#virtengine.market.v1.Lease) | repeated | Leases is a list of leases in the genesis state. |
 | `bids` | [Bid](#virtengine.market.v1beta5.Bid) | repeated | Bids is a list of bids in the genesis state. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/leasemsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/leasemsg.proto
 

 
 <a name="virtengine.market.v1beta5.MsgCloseLease"></a>

 ### MsgCloseLease
 MsgCloseLease defines an SDK message for closing order.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  | LeaseID is the unique identifier of the Lease. |
 | `reason` | [virtengine.market.v1.LeaseClosedReason](#virtengine.market.v1.LeaseClosedReason) |  |  |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgCloseLeaseResponse"></a>

 ### MsgCloseLeaseResponse
 MsgCloseLeaseResponse defines the Msg/CloseLease response type.

 

 

 
 <a name="virtengine.market.v1beta5.MsgCreateLease"></a>

 ### MsgCreateLease
 MsgCreateLease is sent to create a lease.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `bid_id` | [virtengine.market.v1.BidID](#virtengine.market.v1.BidID) |  | BidId is the unique identifier of the Bid. |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgCreateLeaseResponse"></a>

 ### MsgCreateLeaseResponse
 MsgCreateLeaseResponse is the response from creating a lease.

 

 

 
 <a name="virtengine.market.v1beta5.MsgWithdrawLease"></a>

 ### MsgWithdrawLease
 MsgWithdrawLease defines an SDK message for withdrawing lease funds.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  | BidId is the unique identifier of the Bid. |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgWithdrawLeaseResponse"></a>

 ### MsgWithdrawLeaseResponse
 MsgWithdrawLeaseResponse defines the Msg/WithdrawLease response type.

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/paramsmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/paramsmsg.proto
 

 
 <a name="virtengine.market.v1beta5.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the Msg/UpdateParams request type.

Since: virtengine v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.market.v1beta5.Params) |  | params defines the x/deployment parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: virtengine v1.0.0

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/query.proto
 

 
 <a name="virtengine.market.v1beta5.QueryBidRequest"></a>

 ### QueryBidRequest
 QueryBidRequest is request type for the Query/Bid RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.BidID](#virtengine.market.v1.BidID) |  | Id is the unique identifier for the Bid. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryBidResponse"></a>

 ### QueryBidResponse
 QueryBidResponse is response type for the Query/Bid RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `bid` | [Bid](#virtengine.market.v1beta5.Bid) |  | Bid represents a deployment bid. |
 | `escrow_account` | [virtengine.escrow.types.v1.Account](#virtengine.escrow.types.v1.Account) |  | EscrowAccount represents the escrow account created for the Bid. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryBidsRequest"></a>

 ### QueryBidsRequest
 QueryBidsRequest is request type for the Query/Bids RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `filters` | [BidFilters](#virtengine.market.v1beta5.BidFilters) |  | Filters holds the fields to filter bids. |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryBidsResponse"></a>

 ### QueryBidsResponse
 QueryBidsResponse is response type for the Query/Bids RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `bids` | [QueryBidResponse](#virtengine.market.v1beta5.QueryBidResponse) | repeated | Bids is a list of deployment bids. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryLeaseRequest"></a>

 ### QueryLeaseRequest
 QueryLeaseRequest is request type for the Query/Lease RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.LeaseID](#virtengine.market.v1.LeaseID) |  | Id is the unique identifier of the Lease. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryLeaseResponse"></a>

 ### QueryLeaseResponse
 QueryLeaseResponse is response type for the Query/Lease RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `lease` | [virtengine.market.v1.Lease](#virtengine.market.v1.Lease) |  | Lease holds the lease for a deployment. |
 | `escrow_payment` | [virtengine.escrow.types.v1.Payment](#virtengine.escrow.types.v1.Payment) |  | EscrowPayment holds information about the Lease's fractional payment. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryLeasesRequest"></a>

 ### QueryLeasesRequest
 QueryLeasesRequest is request type for the Query/Leases RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `filters` | [virtengine.market.v1.LeaseFilters](#virtengine.market.v1.LeaseFilters) |  | Filters holds the fields to filter leases. |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryLeasesResponse"></a>

 ### QueryLeasesResponse
 QueryLeasesResponse is response type for the Query/Leases RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `leases` | [QueryLeaseResponse](#virtengine.market.v1beta5.QueryLeaseResponse) | repeated | Leases is a list of Lease. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryOrderRequest"></a>

 ### QueryOrderRequest
 QueryOrderRequest is request type for the Query/Order RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [virtengine.market.v1.OrderID](#virtengine.market.v1.OrderID) |  | Id is the unique identifier of the Order. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryOrderResponse"></a>

 ### QueryOrderResponse
 QueryOrderResponse is response type for the Query/Order RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `order` | [Order](#virtengine.market.v1beta5.Order) |  | Order represents a market order. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryOrdersRequest"></a>

 ### QueryOrdersRequest
 QueryOrdersRequest is request type for the Query/Orders RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `filters` | [OrderFilters](#virtengine.market.v1beta5.OrderFilters) |  | Filters holds the fields to filter orders. |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryOrdersResponse"></a>

 ### QueryOrdersResponse
 QueryOrdersResponse is response type for the Query/Orders RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `orders` | [Order](#virtengine.market.v1beta5.Order) | repeated | Orders is a list of market orders. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

 
 <a name="virtengine.market.v1beta5.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request type for the Query/Params RPC method.

 

 

 
 <a name="virtengine.market.v1beta5.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response type for the Query/Params RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.market.v1beta5.Params) |  | params defines the parameters of the module. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.market.v1beta5.Query"></a>

 ### Query
 Query defines the gRPC querier service for the market package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Orders` | [QueryOrdersRequest](#virtengine.market.v1beta5.QueryOrdersRequest) | [QueryOrdersResponse](#virtengine.market.v1beta5.QueryOrdersResponse) | Orders queries orders with filters. | GET|/virtengine/market/v1beta5/orders/list|
 | `Order` | [QueryOrderRequest](#virtengine.market.v1beta5.QueryOrderRequest) | [QueryOrderResponse](#virtengine.market.v1beta5.QueryOrderResponse) | Order queries order details. | GET|/virtengine/market/v1beta5/orders/info|
 | `Bids` | [QueryBidsRequest](#virtengine.market.v1beta5.QueryBidsRequest) | [QueryBidsResponse](#virtengine.market.v1beta5.QueryBidsResponse) | Bids queries bids with filters. | GET|/virtengine/market/v1beta5/bids/list|
 | `Bid` | [QueryBidRequest](#virtengine.market.v1beta5.QueryBidRequest) | [QueryBidResponse](#virtengine.market.v1beta5.QueryBidResponse) | Bid queries bid details. | GET|/virtengine/market/v1beta5/bids/info|
 | `Leases` | [QueryLeasesRequest](#virtengine.market.v1beta5.QueryLeasesRequest) | [QueryLeasesResponse](#virtengine.market.v1beta5.QueryLeasesResponse) | Leases queries leases with filters. | GET|/virtengine/market/v1beta5/leases/list|
 | `Lease` | [QueryLeaseRequest](#virtengine.market.v1beta5.QueryLeaseRequest) | [QueryLeaseResponse](#virtengine.market.v1beta5.QueryLeaseResponse) | Lease queries lease details. | GET|/virtengine/market/v1beta5/leases/info|
 | `Params` | [QueryParamsRequest](#virtengine.market.v1beta5.QueryParamsRequest) | [QueryParamsResponse](#virtengine.market.v1beta5.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/virtengine/market/v1beta5/params|
 
  <!-- end services -->

 
 
 <a name="virtengine/market/v1beta5/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/market/v1beta5/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.market.v1beta5.Msg"></a>

 ### Msg
 Msg defines the market Msg service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `CreateBid` | [MsgCreateBid](#virtengine.market.v1beta5.MsgCreateBid) | [MsgCreateBidResponse](#virtengine.market.v1beta5.MsgCreateBidResponse) | CreateBid defines a method to create a bid given proper inputs. | |
 | `CloseBid` | [MsgCloseBid](#virtengine.market.v1beta5.MsgCloseBid) | [MsgCloseBidResponse](#virtengine.market.v1beta5.MsgCloseBidResponse) | CloseBid defines a method to close a bid given proper inputs. | |
 | `WithdrawLease` | [MsgWithdrawLease](#virtengine.market.v1beta5.MsgWithdrawLease) | [MsgWithdrawLeaseResponse](#virtengine.market.v1beta5.MsgWithdrawLeaseResponse) | WithdrawLease withdraws accrued funds from the lease payment | |
 | `CreateLease` | [MsgCreateLease](#virtengine.market.v1beta5.MsgCreateLease) | [MsgCreateLeaseResponse](#virtengine.market.v1beta5.MsgCreateLeaseResponse) | CreateLease creates a new lease | |
 | `CloseLease` | [MsgCloseLease](#virtengine.market.v1beta5.MsgCloseLease) | [MsgCloseLeaseResponse](#virtengine.market.v1beta5.MsgCloseLeaseResponse) | CloseLease defines a method to close an order given proper inputs. | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.market.v1beta5.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.market.v1beta5.MsgUpdateParamsResponse) | UpdateParams defines a governance operation for updating the x/market module parameters. The authority is hard-coded to the x/gov module account.

Since: virtengine v1.0.0 | |
 
  <!-- end services -->

 
 
 <a name="virtengine/marketplace/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/marketplace/v1/types.proto
 

 
 <a name="virtengine.marketplace.v1.EncryptedProviderSecrets"></a>

 ### EncryptedProviderSecrets
 EncryptedProviderSecrets holds encrypted provider secrets

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `envelope` | [virtengine.encryption.v1.EncryptedPayloadEnvelope](#virtengine.encryption.v1.EncryptedPayloadEnvelope) |  |  |
 | `envelope_ref` | [string](#string) |  |  |
 | `recipient_key_ids` | [string](#string) | repeated |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.IdentityRequirement"></a>

 ### IdentityRequirement
 IdentityRequirement defines the identity verification requirements for an offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `min_score` | [uint32](#uint32) |  |  |
 | `required_status` | [string](#string) |  |  |
 | `require_verified_email` | [bool](#bool) |  |  |
 | `require_verified_domain` | [bool](#bool) |  |  |
 | `require_mfa` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.Offering"></a>

 ### Offering
 Offering represents a marketplace offering from a provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [OfferingID](#virtengine.marketplace.v1.OfferingID) |  |  |
 | `state` | [OfferingState](#virtengine.marketplace.v1.OfferingState) |  |  |
 | `category` | [OfferingCategory](#virtengine.marketplace.v1.OfferingCategory) |  |  |
 | `name` | [string](#string) |  |  |
 | `description` | [string](#string) |  |  |
 | `version` | [string](#string) |  |  |
 | `pricing` | [PricingInfo](#virtengine.marketplace.v1.PricingInfo) |  |  |
 | `identity_requirement` | [IdentityRequirement](#virtengine.marketplace.v1.IdentityRequirement) |  |  |
 | `require_mfa_for_orders` | [bool](#bool) |  |  |
 | `public_metadata` | [Offering.PublicMetadataEntry](#virtengine.marketplace.v1.Offering.PublicMetadataEntry) | repeated |  |
 | `encrypted_secrets` | [EncryptedProviderSecrets](#virtengine.marketplace.v1.EncryptedProviderSecrets) |  |  |
 | `specifications` | [Offering.SpecificationsEntry](#virtengine.marketplace.v1.Offering.SpecificationsEntry) | repeated |  |
 | `tags` | [string](#string) | repeated |  |
 | `regions` | [string](#string) | repeated |  |
 | `created_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `activated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `terminated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
 | `max_concurrent_orders` | [uint32](#uint32) |  |  |
 | `total_order_count` | [uint64](#uint64) |  |  |
 | `active_order_count` | [uint64](#uint64) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.Offering.PublicMetadataEntry"></a>

 ### Offering.PublicMetadataEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.Offering.SpecificationsEntry"></a>

 ### Offering.SpecificationsEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.OfferingID"></a>

 ### OfferingID
 OfferingID is the unique identifier for an offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  |  |
 | `sequence` | [uint64](#uint64) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.PricingInfo"></a>

 ### PricingInfo
 PricingInfo defines the pricing structure for an offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `model` | [PricingModel](#virtengine.marketplace.v1.PricingModel) |  |  |
 | `base_price` | [uint64](#uint64) |  |  |
 | `currency` | [string](#string) |  |  |
 | `usage_rates` | [PricingInfo.UsageRatesEntry](#virtengine.marketplace.v1.PricingInfo.UsageRatesEntry) | repeated |  |
 | `minimum_commitment` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.PricingInfo.UsageRatesEntry"></a>

 ### PricingInfo.UsageRatesEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [uint64](#uint64) |  |  |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.marketplace.v1.OfferingCategory"></a>

 ### OfferingCategory
 OfferingCategory represents the category of an offering

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | OFFERING_CATEGORY_UNSPECIFIED | 0 | OFFERING_CATEGORY_UNSPECIFIED represents an unspecified category |
 | OFFERING_CATEGORY_COMPUTE | 1 | OFFERING_CATEGORY_COMPUTE represents compute/VM offerings |
 | OFFERING_CATEGORY_STORAGE | 2 | OFFERING_CATEGORY_STORAGE represents storage offerings |
 | OFFERING_CATEGORY_NETWORK | 3 | OFFERING_CATEGORY_NETWORK represents network offerings |
 | OFFERING_CATEGORY_HPC | 4 | OFFERING_CATEGORY_HPC represents high-performance computing offerings |
 | OFFERING_CATEGORY_GPU | 5 | OFFERING_CATEGORY_GPU represents GPU compute offerings |
 | OFFERING_CATEGORY_ML | 6 | OFFERING_CATEGORY_ML represents machine learning offerings |
 | OFFERING_CATEGORY_OTHER | 7 | OFFERING_CATEGORY_OTHER represents other/custom offerings |
 

 
 <a name="virtengine.marketplace.v1.OfferingState"></a>

 ### OfferingState
 OfferingState represents the lifecycle state of an offering

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | OFFERING_STATE_UNSPECIFIED | 0 | OFFERING_STATE_UNSPECIFIED represents an unspecified offering state |
 | OFFERING_STATE_ACTIVE | 1 | OFFERING_STATE_ACTIVE indicates the offering is active and available for orders |
 | OFFERING_STATE_PAUSED | 2 | OFFERING_STATE_PAUSED indicates the offering is temporarily paused |
 | OFFERING_STATE_SUSPENDED | 3 | OFFERING_STATE_SUSPENDED indicates the offering is suspended by admin/moderator |
 | OFFERING_STATE_DEPRECATED | 4 | OFFERING_STATE_DEPRECATED indicates the offering is deprecated (no new orders) |
 | OFFERING_STATE_TERMINATED | 5 | OFFERING_STATE_TERMINATED indicates the offering is permanently terminated |
 

 
 <a name="virtengine.marketplace.v1.PricingModel"></a>

 ### PricingModel
 PricingModel represents how an offering is priced

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | PRICING_MODEL_UNSPECIFIED | 0 | PRICING_MODEL_UNSPECIFIED represents an unspecified pricing model |
 | PRICING_MODEL_HOURLY | 1 | PRICING_MODEL_HOURLY represents hourly pricing |
 | PRICING_MODEL_DAILY | 2 | PRICING_MODEL_DAILY represents daily pricing |
 | PRICING_MODEL_MONTHLY | 3 | PRICING_MODEL_MONTHLY represents monthly pricing |
 | PRICING_MODEL_USAGE_BASED | 4 | PRICING_MODEL_USAGE_BASED represents usage-based pricing |
 | PRICING_MODEL_FIXED | 5 | PRICING_MODEL_FIXED represents fixed/one-time pricing |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/marketplace/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/marketplace/v1/tx.proto
 

 
 <a name="virtengine.marketplace.v1.MsgAcceptBid"></a>

 ### MsgAcceptBid
 MsgAcceptBid accepts a bid for an order

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `customer` | [string](#string) |  |  |
 | `order_id` | [string](#string) |  |  |
 | `bid_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgAcceptBidResponse"></a>

 ### MsgAcceptBidResponse
 MsgAcceptBidResponse is the response for MsgAcceptBid

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `allocation_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgCreateOffering"></a>

 ### MsgCreateOffering
 MsgCreateOffering creates a new marketplace offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `offering` | [Offering](#virtengine.marketplace.v1.Offering) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgCreateOfferingResponse"></a>

 ### MsgCreateOfferingResponse
 MsgCreateOfferingResponse is the response for MsgCreateOffering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `offering_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgDeactivateOffering"></a>

 ### MsgDeactivateOffering
 MsgDeactivateOffering deactivates an existing marketplace offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `offering_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgDeactivateOfferingResponse"></a>

 ### MsgDeactivateOfferingResponse
 MsgDeactivateOfferingResponse is the response for MsgDeactivateOffering

 

 

 
 <a name="virtengine.marketplace.v1.MsgTerminateAllocation"></a>

 ### MsgTerminateAllocation
 MsgTerminateAllocation terminates an allocation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `customer` | [string](#string) |  |  |
 | `allocation_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgTerminateAllocationResponse"></a>

 ### MsgTerminateAllocationResponse
 MsgTerminateAllocationResponse is the response for MsgTerminateAllocation

 

 

 
 <a name="virtengine.marketplace.v1.MsgUpdateOffering"></a>

 ### MsgUpdateOffering
 MsgUpdateOffering updates an existing marketplace offering

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [string](#string) |  |  |
 | `offering_id` | [string](#string) |  |  |
 | `updates` | [Offering](#virtengine.marketplace.v1.Offering) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgUpdateOfferingResponse"></a>

 ### MsgUpdateOfferingResponse
 MsgUpdateOfferingResponse is the response for MsgUpdateOffering

 

 

 
 <a name="virtengine.marketplace.v1.MsgWaldurCallback"></a>

 ### MsgWaldurCallback
 MsgWaldurCallback handles callbacks from Waldur integration

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `callback_type` | [string](#string) |  |  |
 | `resource_id` | [string](#string) |  |  |
 | `status` | [string](#string) |  |  |
 | `payload` | [string](#string) |  |  |
 | `signature` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.marketplace.v1.MsgWaldurCallbackResponse"></a>

 ### MsgWaldurCallbackResponse
 MsgWaldurCallbackResponse is the response for MsgWaldurCallback

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.marketplace.v1.Msg"></a>

 ### Msg
 Msg defines the marketplace Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `CreateOffering` | [MsgCreateOffering](#virtengine.marketplace.v1.MsgCreateOffering) | [MsgCreateOfferingResponse](#virtengine.marketplace.v1.MsgCreateOfferingResponse) | CreateOffering creates a new marketplace offering | |
 | `UpdateOffering` | [MsgUpdateOffering](#virtengine.marketplace.v1.MsgUpdateOffering) | [MsgUpdateOfferingResponse](#virtengine.marketplace.v1.MsgUpdateOfferingResponse) | UpdateOffering updates an existing marketplace offering | |
 | `DeactivateOffering` | [MsgDeactivateOffering](#virtengine.marketplace.v1.MsgDeactivateOffering) | [MsgDeactivateOfferingResponse](#virtengine.marketplace.v1.MsgDeactivateOfferingResponse) | DeactivateOffering deactivates an existing marketplace offering | |
 | `AcceptBid` | [MsgAcceptBid](#virtengine.marketplace.v1.MsgAcceptBid) | [MsgAcceptBidResponse](#virtengine.marketplace.v1.MsgAcceptBidResponse) | AcceptBid accepts a provider bid on an order | |
 | `TerminateAllocation` | [MsgTerminateAllocation](#virtengine.marketplace.v1.MsgTerminateAllocation) | [MsgTerminateAllocationResponse](#virtengine.marketplace.v1.MsgTerminateAllocationResponse) | TerminateAllocation terminates an allocation | |
 | `WaldurCallback` | [MsgWaldurCallback](#virtengine.marketplace.v1.MsgWaldurCallback) | [MsgWaldurCallbackResponse](#virtengine.marketplace.v1.MsgWaldurCallbackResponse) | WaldurCallback handles callbacks from Waldur integration | |
 
  <!-- end services -->

 
 
 <a name="virtengine/mfa/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/mfa/v1/types.proto
 

 
 <a name="virtengine.mfa.v1.AuthorizationSession"></a>

 ### AuthorizationSession
 AuthorizationSession represents a temporary elevated session after MFA verification

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `session_id` | [string](#string) |  | SessionID is the unique identifier for this session |
 | `account_address` | [string](#string) |  | AccountAddress is the account this session belongs to |
 | `transaction_type` | [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType) |  | TransactionType is the type of transaction authorized |
 | `verified_factors` | [FactorType](#virtengine.mfa.v1.FactorType) | repeated | VerifiedFactors are the factors that were verified for this session |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the session was created |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the session expires |
 | `used_at` | [int64](#int64) |  | UsedAt is when the session was used (if single-use) |
 | `is_single_use` | [bool](#bool) |  | IsSingleUse indicates if the session can only be used once |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is the device this session is bound to |
 
 

 

 
 <a name="virtengine.mfa.v1.Challenge"></a>

 ### Challenge
 Challenge represents an MFA challenge issued to a user

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  | ChallengeID is the unique identifier for this challenge |
 | `account_address` | [string](#string) |  | AccountAddress is the account this challenge is for |
 | `factor_type` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorType is the type of factor this challenge is for |
 | `factor_id` | [string](#string) |  | FactorID is the specific factor enrollment being challenged |
 | `transaction_type` | [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType) |  | TransactionType is the sensitive transaction this challenge is for |
 | `status` | [ChallengeStatus](#virtengine.mfa.v1.ChallengeStatus) |  | Status is the current status of the challenge |
 | `challenge_data` | [bytes](#bytes) |  | ChallengeData contains factor-specific challenge data |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the challenge was created |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the challenge expires |
 | `verified_at` | [int64](#int64) |  | VerifiedAt is when the challenge was verified (if successful) |
 | `attempt_count` | [uint32](#uint32) |  | AttemptCount tracks the number of verification attempts |
 | `max_attempts` | [uint32](#uint32) |  | MaxAttempts is the maximum number of verification attempts allowed |
 | `nonce` | [string](#string) |  | Nonce is a random value for replay protection |
 | `session_id` | [string](#string) |  | SessionID links this challenge to an authorization session |
 | `metadata` | [ChallengeMetadata](#virtengine.mfa.v1.ChallengeMetadata) |  | Metadata contains additional challenge-specific data |
 
 

 

 
 <a name="virtengine.mfa.v1.ChallengeMetadata"></a>

 ### ChallengeMetadata
 ChallengeMetadata contains additional challenge information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `fido2_challenge` | [FIDO2ChallengeData](#virtengine.mfa.v1.FIDO2ChallengeData) |  | FIDO2Challenge contains FIDO2-specific challenge data |
 | `otp_info` | [OTPChallengeInfo](#virtengine.mfa.v1.OTPChallengeInfo) |  | OTPInfo contains OTP-specific tracking data |
 | `client_info` | [ClientInfo](#virtengine.mfa.v1.ClientInfo) |  | ClientInfo contains information about the requesting client |
 | `hardware_key_challenge` | [HardwareKeyChallenge](#virtengine.mfa.v1.HardwareKeyChallenge) |  | HardwareKeyChallenge contains hardware key challenge data |
 
 

 

 
 <a name="virtengine.mfa.v1.ChallengeResponse"></a>

 ### ChallengeResponse
 ChallengeResponse represents a response to an MFA challenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  | ChallengeID is the challenge being responded to |
 | `factor_type` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorType is the type of factor used |
 | `response_data` | [bytes](#bytes) |  | ResponseData contains the verification data |
 | `client_info` | [ClientInfo](#virtengine.mfa.v1.ClientInfo) |  | ClientInfo contains information about the responding client |
 | `timestamp` | [int64](#int64) |  | Timestamp is when the response was created |
 
 

 

 
 <a name="virtengine.mfa.v1.ClientInfo"></a>

 ### ClientInfo
 ClientInfo contains information about the requesting client

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is a hash identifying the device |
 | `ip_hash` | [string](#string) |  | IPHash is a hash of the client IP |
 | `user_agent` | [string](#string) |  | UserAgent is the sanitized user agent string |
 | `requested_at` | [int64](#int64) |  | RequestedAt is when the request was made |
 
 

 

 
 <a name="virtengine.mfa.v1.DeviceInfo"></a>

 ### DeviceInfo
 DeviceInfo contains information about a trusted device

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `fingerprint` | [string](#string) |  | Fingerprint is a unique hash identifying the device |
 | `user_agent` | [string](#string) |  | UserAgent is the user agent string (sanitized) |
 | `first_seen_at` | [int64](#int64) |  | FirstSeenAt is when this device was first seen |
 | `last_seen_at` | [int64](#int64) |  | LastSeenAt is when this device was last seen |
 | `ip_hash` | [string](#string) |  | IPHash is a hash of the last known IP (for change detection, not tracking) |
 | `trust_expires_at` | [int64](#int64) |  | TrustExpiresAt is when the device trust expires |
 
 

 

 
 <a name="virtengine.mfa.v1.EventChallengeVerified"></a>

 ### EventChallengeVerified
 EventChallengeVerified is emitted when a challenge is verified

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account that verified the challenge |
 | `challenge_id` | [string](#string) |  | ChallengeID is the ID of the verified challenge |
 | `factor_type` | [string](#string) |  | FactorType is the type of factor used |
 | `transaction_type` | [string](#string) |  | TransactionType is the transaction type that was authorized |
 
 

 

 
 <a name="virtengine.mfa.v1.EventFactorEnrolled"></a>

 ### EventFactorEnrolled
 EventFactorEnrolled is emitted when a factor is enrolled

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account that enrolled the factor |
 | `factor_type` | [string](#string) |  | FactorType is the type of factor enrolled |
 | `factor_id` | [string](#string) |  | FactorID is the ID of the enrolled factor |
 
 

 

 
 <a name="virtengine.mfa.v1.EventFactorRevoked"></a>

 ### EventFactorRevoked
 EventFactorRevoked is emitted when a factor is revoked

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account that revoked the factor |
 | `factor_type` | [string](#string) |  | FactorType is the type of factor revoked |
 | `factor_id` | [string](#string) |  | FactorID is the ID of the revoked factor |
 
 

 

 
 <a name="virtengine.mfa.v1.EventMFAPolicyUpdated"></a>

 ### EventMFAPolicyUpdated
 EventMFAPolicyUpdated is emitted when an MFA policy is updated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account whose policy was updated |
 | `enabled` | [bool](#bool) |  | Enabled indicates if MFA is now enabled |
 
 

 

 
 <a name="virtengine.mfa.v1.FIDO2ChallengeData"></a>

 ### FIDO2ChallengeData
 FIDO2ChallengeData contains FIDO2-specific challenge information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge` | [bytes](#bytes) |  | Challenge is the random challenge bytes |
 | `relying_party_id` | [string](#string) |  | RelyingPartyID is the relying party identifier |
 | `allowed_credentials` | [bytes](#bytes) | repeated | AllowedCredentials is the list of allowed credential IDs |
 | `user_verification` | [string](#string) |  | UserVerificationRequirement specifies if user verification is required |
 
 

 

 
 <a name="virtengine.mfa.v1.FIDO2CredentialInfo"></a>

 ### FIDO2CredentialInfo
 FIDO2CredentialInfo contains FIDO2-specific credential information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `credential_id` | [bytes](#bytes) |  | CredentialID is the FIDO2 credential identifier |
 | `public_key` | [bytes](#bytes) |  | PublicKey is the COSE-encoded public key |
 | `aaguid` | [bytes](#bytes) |  | AAGUID is the Authenticator Attestation GUID |
 | `sign_count` | [uint32](#uint32) |  | SignCount is the signature counter for clone detection |
 | `attestation_type` | [string](#string) |  | AttestationType indicates the attestation type used |
 
 

 

 
 <a name="virtengine.mfa.v1.FactorCombination"></a>

 ### FactorCombination
 FactorCombination represents a set of factors that must ALL be satisfied (AND logic)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `factors` | [FactorType](#virtengine.mfa.v1.FactorType) | repeated | Factors is the list of factor types that must all be verified |
 | `min_security_level` | [FactorSecurityLevel](#virtengine.mfa.v1.FactorSecurityLevel) |  | MinSecurityLevel is the minimum security level required for the combination |
 
 

 

 
 <a name="virtengine.mfa.v1.FactorEnrollment"></a>

 ### FactorEnrollment
 FactorEnrollment represents a factor enrollment record

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account that owns this enrollment |
 | `factor_type` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorType is the type of factor |
 | `factor_id` | [string](#string) |  | FactorID is a unique identifier for this factor enrollment |
 | `public_identifier` | [bytes](#bytes) |  | PublicIdentifier is the public component of the factor |
 | `label` | [string](#string) |  | Label is a user-friendly label for the factor |
 | `status` | [FactorEnrollmentStatus](#virtengine.mfa.v1.FactorEnrollmentStatus) |  | Status is the current enrollment status |
 | `enrolled_at` | [int64](#int64) |  | EnrolledAt is the timestamp when the factor was enrolled |
 | `verified_at` | [int64](#int64) |  | VerifiedAt is the timestamp when the enrollment was verified |
 | `revoked_at` | [int64](#int64) |  | RevokedAt is the timestamp when the factor was revoked |
 | `last_used_at` | [int64](#int64) |  | LastUsedAt is the timestamp of the last successful use |
 | `use_count` | [uint64](#uint64) |  | UseCount tracks how many times this factor has been used |
 | `metadata` | [FactorMetadata](#virtengine.mfa.v1.FactorMetadata) |  | Metadata contains additional factor-specific metadata |
 
 

 

 
 <a name="virtengine.mfa.v1.FactorMetadata"></a>

 ### FactorMetadata
 FactorMetadata contains type-specific metadata for enrollments

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `veid_threshold` | [uint32](#uint32) |  | VEIDThreshold is the minimum VEID score required (for VEID factor) |
 | `device_info` | [DeviceInfo](#virtengine.mfa.v1.DeviceInfo) |  | DeviceInfo contains device information (for TrustedDevice factor) |
 | `fido2_info` | [FIDO2CredentialInfo](#virtengine.mfa.v1.FIDO2CredentialInfo) |  | FIDO2Info contains FIDO2-specific metadata |
 | `contact_hash` | [string](#string) |  | ContactHash contains a hash of the contact info (for SMS/Email verification tracking) |
 | `hardware_key_info` | [HardwareKeyEnrollment](#virtengine.mfa.v1.HardwareKeyEnrollment) |  | HardwareKeyInfo contains hardware key/X.509/smart card metadata |
 
 

 

 
 <a name="virtengine.mfa.v1.HardwareKeyChallenge"></a>

 ### HardwareKeyChallenge
 HardwareKeyChallenge represents a challenge for hardware key authentication

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge` | [bytes](#bytes) |  | Challenge is the random challenge bytes to be signed |
 | `key_id` | [string](#string) |  | KeyID is the expected key ID to use for signing |
 | `nonce` | [string](#string) |  | Nonce is a random nonce for replay protection |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the challenge was created |
 
 

 

 
 <a name="virtengine.mfa.v1.HardwareKeyEnrollment"></a>

 ### HardwareKeyEnrollment
 HardwareKeyEnrollment represents a hardware key enrollment for MFA

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key_type` | [HardwareKeyType](#virtengine.mfa.v1.HardwareKeyType) |  | KeyType is the type of hardware key |
 | `key_id` | [string](#string) |  | KeyID is a unique identifier for this key (certificate fingerprint or card serial) |
 | `subject_dn` | [string](#string) |  | SubjectDN is the distinguished name from the certificate subject |
 | `issuer_dn` | [string](#string) |  | IssuerDN is the distinguished name of the certificate issuer |
 | `serial_number` | [string](#string) |  | SerialNumber is the certificate serial number (hex encoded) |
 | `public_key_fingerprint` | [string](#string) |  | PublicKeyFingerprint is SHA-256 hash of the public key (hex encoded) |
 | `not_before` | [int64](#int64) |  | NotBefore is the certificate validity start time |
 | `not_after` | [int64](#int64) |  | NotAfter is the certificate validity end time |
 | `key_usage` | [string](#string) | repeated | KeyUsage indicates the allowed key usage flags |
 | `extended_key_usage` | [string](#string) | repeated | ExtendedKeyUsage indicates extended key usage purposes |
 | `smart_card_info` | [SmartCardInfo](#virtengine.mfa.v1.SmartCardInfo) |  | SmartCardInfo contains smart card specific metadata |
 | `revocation_check_enabled` | [bool](#bool) |  | RevocationCheckEnabled indicates if revocation checking is enabled |
 | `last_revocation_check` | [int64](#int64) |  | LastRevocationCheck is the timestamp of the last revocation check |
 | `revocation_status` | [RevocationStatus](#virtengine.mfa.v1.RevocationStatus) |  | RevocationStatus is the current revocation status |
 | `trusted_ca_cert_fingerprints` | [string](#string) | repeated | TrustedCACertFingerprints are the fingerprints of trusted CA certificates |
 
 

 

 
 <a name="virtengine.mfa.v1.MFAPolicy"></a>

 ### MFAPolicy
 MFAPolicy defines the MFA requirements for an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account this policy applies to |
 | `required_factors` | [FactorCombination](#virtengine.mfa.v1.FactorCombination) | repeated | RequiredFactors is a list of factor combinations (OR logic between combinations) |
 | `trusted_device_rule` | [TrustedDevicePolicy](#virtengine.mfa.v1.TrustedDevicePolicy) |  | TrustedDeviceRule defines how trusted devices affect MFA requirements |
 | `recovery_factors` | [FactorCombination](#virtengine.mfa.v1.FactorCombination) | repeated | RecoveryFactors defines factor combinations for account recovery |
 | `key_rotation_factors` | [FactorCombination](#virtengine.mfa.v1.FactorCombination) | repeated | KeyRotationFactors defines factor combinations for key rotation |
 | `session_duration` | [int64](#int64) |  | SessionDuration is how long an MFA session remains valid (in seconds) |
 | `veid_threshold` | [uint32](#uint32) |  | VEIDThreshold is the minimum VEID score required for VEID-based factors |
 | `enabled` | [bool](#bool) |  | Enabled indicates if MFA is enabled for this account |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the policy was created |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when the policy was last updated |
 
 

 

 
 <a name="virtengine.mfa.v1.MFAProof"></a>

 ### MFAProof
 MFAProof represents proof of MFA verification

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `session_id` | [string](#string) |  | SessionID is the authorization session ID |
 | `verified_factors` | [FactorType](#virtengine.mfa.v1.FactorType) | repeated | VerifiedFactors are the factors that were verified |
 | `timestamp` | [int64](#int64) |  | Timestamp is when the proof was generated (Unix timestamp) |
 | `signature` | [bytes](#bytes) |  | Signature is the signature over the proof data |
 
 

 

 
 <a name="virtengine.mfa.v1.OTPChallengeInfo"></a>

 ### OTPChallengeInfo
 OTPChallengeInfo contains OTP tracking information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `delivery_method` | [string](#string) |  | DeliveryMethod indicates how the OTP was delivered |
 | `delivery_destination_masked` | [string](#string) |  | DeliveryDestinationMasked is the masked delivery destination |
 | `sent_at` | [int64](#int64) |  | SentAt is when the OTP was sent |
 | `resend_count` | [uint32](#uint32) |  | ResendCount tracks how many times the OTP was resent |
 | `last_resend_at` | [int64](#int64) |  | LastResendAt is when the OTP was last resent |
 
 

 

 
 <a name="virtengine.mfa.v1.Params"></a>

 ### Params
 Params defines the parameters for the mfa module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `default_session_duration` | [int64](#int64) |  | DefaultSessionDuration is the default MFA session duration in seconds |
 | `max_factors_per_account` | [uint32](#uint32) |  | MaxFactorsPerAccount is the maximum number of factors per account |
 | `max_challenge_attempts` | [uint32](#uint32) |  | MaxChallengeAttempts is the maximum verification attempts per challenge |
 | `challenge_ttl` | [int64](#int64) |  | ChallengeTTL is the challenge time-to-live in seconds |
 | `max_trusted_devices` | [uint32](#uint32) |  | MaxTrustedDevices is the maximum trusted devices per account |
 | `trusted_device_ttl` | [int64](#int64) |  | TrustedDeviceTTL is the trusted device time-to-live in seconds |
 | `min_veid_score_for_mfa` | [uint32](#uint32) |  | MinVEIDScoreForMFA is the minimum VEID score to enable MFA |
 | `require_at_least_one_factor` | [bool](#bool) |  | RequireAtLeastOneFactor requires at least one factor when MFA is enabled |
 | `allowed_factor_types` | [FactorType](#virtengine.mfa.v1.FactorType) | repeated | AllowedFactorTypes lists the factor types allowed on this chain |
 
 

 

 
 <a name="virtengine.mfa.v1.SensitiveTxConfig"></a>

 ### SensitiveTxConfig
 SensitiveTxConfig represents the configuration for a sensitive transaction type

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `transaction_type` | [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType) |  | TransactionType is the type of transaction |
 | `enabled` | [bool](#bool) |  | Enabled indicates if MFA is required for this transaction type |
 | `min_veid_score` | [uint32](#uint32) |  | MinVEIDScore is the minimum VEID score required |
 | `required_factor_combinations` | [FactorCombination](#virtengine.mfa.v1.FactorCombination) | repeated | RequiredFactorCombinations are the default factor combinations required |
 | `session_duration` | [int64](#int64) |  | SessionDuration is the authorization session duration in seconds |
 | `is_single_use` | [bool](#bool) |  | IsSingleUse indicates if the authorization is single-use |
 | `allow_trusted_device_reduction` | [bool](#bool) |  | AllowTrustedDeviceReduction indicates if trusted devices can reduce MFA |
 | `value_threshold` | [string](#string) |  | ValueThreshold is the value threshold for amount-based transactions |
 | `cooldown_period` | [int64](#int64) |  | CooldownPeriod is the cooldown period in seconds for rate-limited operations |
 | `description` | [string](#string) |  | Description provides a human-readable description |
 
 

 

 
 <a name="virtengine.mfa.v1.SmartCardInfo"></a>

 ### SmartCardInfo
 SmartCardInfo contains smart card/PIV specific information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `card_serial_number` | [string](#string) |  | CardSerialNumber is the smart card serial number |
 | `card_type` | [string](#string) |  | CardType indicates the type of smart card (PIV, CAC, etc.) |
 | `slot_id` | [string](#string) |  | SlotID indicates which slot the certificate was read from |
 | `chuid` | [string](#string) |  | CHUID is the Card Holder Unique Identifier (for PIV cards) |
 | `fascn` | [string](#string) |  | FASC_N is the Federal Agency Smart Credential Number (if available) |
 | `card_holder_name` | [string](#string) |  | CardHolderName is the name on the card (from CHUID or certificate) |
 | `expiration_date` | [int64](#int64) |  | ExpirationDate is when the card expires |
 | `last_pin_verification` | [int64](#int64) |  | LastPINVerification is when PIN was last verified |
 
 

 

 
 <a name="virtengine.mfa.v1.TrustedDevice"></a>

 ### TrustedDevice
 TrustedDevice represents a stored trusted device record

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account that trusts this device |
 | `device_info` | [DeviceInfo](#virtengine.mfa.v1.DeviceInfo) |  | DeviceInfo contains the device information |
 | `added_at` | [int64](#int64) |  | AddedAt is when the device was added |
 | `last_used_at` | [int64](#int64) |  | LastUsedAt is when the device was last used |
 
 

 

 
 <a name="virtengine.mfa.v1.TrustedDevicePolicy"></a>

 ### TrustedDevicePolicy
 TrustedDevicePolicy defines how trusted devices can reduce MFA requirements

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `enabled` | [bool](#bool) |  | Enabled indicates if trusted device reduction is enabled |
 | `trust_duration` | [int64](#int64) |  | TrustDuration is how long a device remains trusted (in seconds) |
 | `reduced_factors` | [FactorCombination](#virtengine.mfa.v1.FactorCombination) |  | ReducedFactors is the factor combination to use for trusted devices |
 | `max_trusted_devices` | [uint32](#uint32) |  | MaxTrustedDevices is the maximum number of trusted devices per account |
 | `require_reauth_for_sensitive` | [bool](#bool) |  | RequireReauthForSensitive if true, still requires full MFA for critical actions |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.mfa.v1.ChallengeStatus"></a>

 ### ChallengeStatus
 ChallengeStatus represents the status of an MFA challenge

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | CHALLENGE_STATUS_UNSPECIFIED | 0 | CHALLENGE_STATUS_UNSPECIFIED represents an unspecified status |
 | CHALLENGE_STATUS_PENDING | 1 | CHALLENGE_STATUS_PENDING represents a challenge awaiting response |
 | CHALLENGE_STATUS_VERIFIED | 2 | CHALLENGE_STATUS_VERIFIED represents a successfully verified challenge |
 | CHALLENGE_STATUS_FAILED | 3 | CHALLENGE_STATUS_FAILED represents a failed challenge |
 | CHALLENGE_STATUS_EXPIRED | 4 | CHALLENGE_STATUS_EXPIRED represents an expired challenge |
 | CHALLENGE_STATUS_CANCELLED | 5 | CHALLENGE_STATUS_CANCELLED represents a cancelled challenge |
 

 
 <a name="virtengine.mfa.v1.FactorEnrollmentStatus"></a>

 ### FactorEnrollmentStatus
 FactorEnrollmentStatus represents the status of a factor enrollment

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | ENROLLMENT_STATUS_UNSPECIFIED | 0 | ENROLLMENT_STATUS_UNSPECIFIED represents an unspecified status |
 | ENROLLMENT_STATUS_PENDING | 1 | ENROLLMENT_STATUS_PENDING represents a pending enrollment awaiting verification |
 | ENROLLMENT_STATUS_ACTIVE | 2 | ENROLLMENT_STATUS_ACTIVE represents an active, verified enrollment |
 | ENROLLMENT_STATUS_REVOKED | 3 | ENROLLMENT_STATUS_REVOKED represents a revoked enrollment |
 | ENROLLMENT_STATUS_EXPIRED | 4 | ENROLLMENT_STATUS_EXPIRED represents an expired enrollment |
 

 
 <a name="virtengine.mfa.v1.FactorSecurityLevel"></a>

 ### FactorSecurityLevel
 FactorSecurityLevel represents the security level of a factor

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | FACTOR_SECURITY_LEVEL_UNSPECIFIED | 0 | FACTOR_SECURITY_LEVEL_UNSPECIFIED represents an unspecified security level |
 | FACTOR_SECURITY_LEVEL_LOW | 1 | FACTOR_SECURITY_LEVEL_LOW represents low security factors (email, SMS) |
 | FACTOR_SECURITY_LEVEL_MEDIUM | 2 | FACTOR_SECURITY_LEVEL_MEDIUM represents medium security factors (TOTP) |
 | FACTOR_SECURITY_LEVEL_HIGH | 3 | FACTOR_SECURITY_LEVEL_HIGH represents high security factors (FIDO2, VEID) |
 

 
 <a name="virtengine.mfa.v1.FactorType"></a>

 ### FactorType
 FactorType represents the type of authentication factor

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | FACTOR_TYPE_UNSPECIFIED | 0 | FACTOR_TYPE_UNSPECIFIED represents an unspecified factor type |
 | FACTOR_TYPE_TOTP | 1 | FACTOR_TYPE_TOTP represents Time-based One-Time Password |
 | FACTOR_TYPE_FIDO2 | 2 | FACTOR_TYPE_FIDO2 represents FIDO2/WebAuthn authentication |
 | FACTOR_TYPE_SMS | 3 | FACTOR_TYPE_SMS represents SMS OTP authentication |
 | FACTOR_TYPE_EMAIL | 4 | FACTOR_TYPE_EMAIL represents Email OTP authentication |
 | FACTOR_TYPE_VEID | 5 | FACTOR_TYPE_VEID represents VEID identity score threshold |
 | FACTOR_TYPE_TRUSTED_DEVICE | 6 | FACTOR_TYPE_TRUSTED_DEVICE represents trusted browser/device binding |
 | FACTOR_TYPE_HARDWARE_KEY | 7 | FACTOR_TYPE_HARDWARE_KEY represents X.509/smart card/PIV hardware key authentication |
 

 
 <a name="virtengine.mfa.v1.HardwareKeyType"></a>

 ### HardwareKeyType
 HardwareKeyType represents the type of hardware key

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | HARDWARE_KEY_TYPE_UNSPECIFIED | 0 | HARDWARE_KEY_TYPE_UNSPECIFIED represents an unspecified hardware key type |
 | HARDWARE_KEY_TYPE_X509 | 1 | HARDWARE_KEY_TYPE_X509 represents X.509 certificate-based authentication |
 | HARDWARE_KEY_TYPE_SMART_CARD | 2 | HARDWARE_KEY_TYPE_SMART_CARD represents smart card/PIV authentication |
 | HARDWARE_KEY_TYPE_PIV | 3 | HARDWARE_KEY_TYPE_PIV represents PIV (Personal Identity Verification) card |
 

 
 <a name="virtengine.mfa.v1.RevocationStatus"></a>

 ### RevocationStatus
 RevocationStatus represents the revocation status of a certificate

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | REVOCATION_STATUS_UNKNOWN | 0 | REVOCATION_STATUS_UNKNOWN indicates the revocation status is unknown |
 | REVOCATION_STATUS_GOOD | 1 | REVOCATION_STATUS_GOOD indicates the certificate is not revoked |
 | REVOCATION_STATUS_REVOKED | 2 | REVOCATION_STATUS_REVOKED indicates the certificate has been revoked |
 | REVOCATION_STATUS_CHECK_FAILED | 3 | REVOCATION_STATUS_CHECK_FAILED indicates the revocation check failed |
 

 
 <a name="virtengine.mfa.v1.SensitiveTransactionType"></a>

 ### SensitiveTransactionType
 SensitiveTransactionType represents types of transactions that require MFA

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | SENSITIVE_TX_UNSPECIFIED | 0 | SENSITIVE_TX_UNSPECIFIED represents an unspecified transaction type |
 | SENSITIVE_TX_ACCOUNT_RECOVERY | 1 | SENSITIVE_TX_ACCOUNT_RECOVERY represents account recovery operations |
 | SENSITIVE_TX_KEY_ROTATION | 2 | SENSITIVE_TX_KEY_ROTATION represents key rotation operations |
 | SENSITIVE_TX_LARGE_WITHDRAWAL | 3 | SENSITIVE_TX_LARGE_WITHDRAWAL represents large token withdrawals |
 | SENSITIVE_TX_PROVIDER_REGISTRATION | 4 | SENSITIVE_TX_PROVIDER_REGISTRATION represents provider registration |
 | SENSITIVE_TX_VALIDATOR_REGISTRATION | 5 | SENSITIVE_TX_VALIDATOR_REGISTRATION represents validator registration |
 | SENSITIVE_TX_HIGH_VALUE_ORDER | 6 | SENSITIVE_TX_HIGH_VALUE_ORDER represents high-value marketplace orders |
 | SENSITIVE_TX_ROLE_ASSIGNMENT | 7 | SENSITIVE_TX_ROLE_ASSIGNMENT represents role assignment operations |
 | SENSITIVE_TX_GOVERNANCE_PROPOSAL | 8 | SENSITIVE_TX_GOVERNANCE_PROPOSAL represents governance proposal creation |
 | SENSITIVE_TX_PRIMARY_EMAIL_CHANGE | 9 | SENSITIVE_TX_PRIMARY_EMAIL_CHANGE represents primary email changes |
 | SENSITIVE_TX_PHONE_NUMBER_CHANGE | 10 | SENSITIVE_TX_PHONE_NUMBER_CHANGE represents phone number changes |
 | SENSITIVE_TX_TWO_FACTOR_DISABLE | 11 | SENSITIVE_TX_TWO_FACTOR_DISABLE represents disabling 2FA |
 | SENSITIVE_TX_ACCOUNT_DELETION | 12 | SENSITIVE_TX_ACCOUNT_DELETION represents account deletion |
 | SENSITIVE_TX_GOVERNANCE_VOTE | 13 | SENSITIVE_TX_GOVERNANCE_VOTE represents high-stake governance votes |
 | SENSITIVE_TX_FIRST_OFFERING_CREATE | 14 | SENSITIVE_TX_FIRST_OFFERING_CREATE represents first offering creation by provider |
 | SENSITIVE_TX_TRANSFER_TO_NEW_ADDRESS | 15 | SENSITIVE_TX_TRANSFER_TO_NEW_ADDRESS represents transfers to new addresses |
 | SENSITIVE_TX_MEDIUM_WITHDRAWAL | 16 | SENSITIVE_TX_MEDIUM_WITHDRAWAL represents medium-value withdrawals |
 | SENSITIVE_TX_API_KEY_GENERATION | 17 | SENSITIVE_TX_API_KEY_GENERATION represents API key generation |
 | SENSITIVE_TX_WEBHOOK_CONFIGURATION | 18 | SENSITIVE_TX_WEBHOOK_CONFIGURATION represents webhook configuration |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/mfa/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/mfa/v1/genesis.proto
 

 
 <a name="virtengine.mfa.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the mfa module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.mfa.v1.Params) |  | Params are the module parameters |
 | `mfa_policies` | [MFAPolicy](#virtengine.mfa.v1.MFAPolicy) | repeated | MFAPolicies are the initial MFA policies |
 | `factor_enrollments` | [FactorEnrollment](#virtengine.mfa.v1.FactorEnrollment) | repeated | FactorEnrollments are the initial factor enrollments |
 | `sensitive_tx_configs` | [SensitiveTxConfig](#virtengine.mfa.v1.SensitiveTxConfig) | repeated | SensitiveTxConfigs are the sensitive transaction configurations |
 | `trusted_devices` | [TrustedDevice](#virtengine.mfa.v1.TrustedDevice) | repeated | TrustedDevices are the initial trusted devices |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/mfa/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/mfa/v1/query.proto
 

 
 <a name="virtengine.mfa.v1.QueryAllSensitiveTxConfigsRequest"></a>

 ### QueryAllSensitiveTxConfigsRequest
 QueryAllSensitiveTxConfigsRequest is the request for QueryAllSensitiveTxConfigs

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryAllSensitiveTxConfigsResponse"></a>

 ### QueryAllSensitiveTxConfigsResponse
 QueryAllSensitiveTxConfigsResponse is the response for QueryAllSensitiveTxConfigs

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `configs` | [SensitiveTxConfig](#virtengine.mfa.v1.SensitiveTxConfig) | repeated | Configs is the list of sensitive transaction configurations |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryAuthorizationSessionRequest"></a>

 ### QueryAuthorizationSessionRequest
 QueryAuthorizationSessionRequest is the request for QueryAuthorizationSession

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `session_id` | [string](#string) |  | SessionID is the ID of the session to query |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryAuthorizationSessionResponse"></a>

 ### QueryAuthorizationSessionResponse
 QueryAuthorizationSessionResponse is the response for QueryAuthorizationSession

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `session` | [AuthorizationSession](#virtengine.mfa.v1.AuthorizationSession) |  | Session is the authorization session |
 | `found` | [bool](#bool) |  | Found indicates if the session was found |
 | `is_valid` | [bool](#bool) |  | IsValid indicates if the session is still valid |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryChallengeRequest"></a>

 ### QueryChallengeRequest
 QueryChallengeRequest is the request for QueryChallenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  | ChallengeID is the ID of the challenge to query |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryChallengeResponse"></a>

 ### QueryChallengeResponse
 QueryChallengeResponse is the response for QueryChallenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge` | [Challenge](#virtengine.mfa.v1.Challenge) |  | Challenge is the MFA challenge |
 | `found` | [bool](#bool) |  | Found indicates if the challenge was found |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryFactorEnrollmentRequest"></a>

 ### QueryFactorEnrollmentRequest
 QueryFactorEnrollmentRequest is the request for QueryFactorEnrollment

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query |
 | `factor_id` | [string](#string) |  | FactorID is the ID of the factor to query |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryFactorEnrollmentResponse"></a>

 ### QueryFactorEnrollmentResponse
 QueryFactorEnrollmentResponse is the response for QueryFactorEnrollment

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `enrollment` | [FactorEnrollment](#virtengine.mfa.v1.FactorEnrollment) |  | Enrollment is the factor enrollment |
 | `found` | [bool](#bool) |  | Found indicates if the enrollment was found |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryFactorEnrollmentsRequest"></a>

 ### QueryFactorEnrollmentsRequest
 QueryFactorEnrollmentsRequest is the request for QueryFactorEnrollments

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query |
 | `factor_type_filter` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorTypeFilter optionally filters by factor type |
 | `status_filter` | [FactorEnrollmentStatus](#virtengine.mfa.v1.FactorEnrollmentStatus) |  | StatusFilter optionally filters by enrollment status |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryFactorEnrollmentsResponse"></a>

 ### QueryFactorEnrollmentsResponse
 QueryFactorEnrollmentsResponse is the response for QueryFactorEnrollments

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `enrollments` | [FactorEnrollment](#virtengine.mfa.v1.FactorEnrollment) | repeated | Enrollments is the list of factor enrollments |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryMFAPolicyRequest"></a>

 ### QueryMFAPolicyRequest
 QueryMFAPolicyRequest is the request for QueryMFAPolicy

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryMFAPolicyResponse"></a>

 ### QueryMFAPolicyResponse
 QueryMFAPolicyResponse is the response for QueryMFAPolicy

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `policy` | [MFAPolicy](#virtengine.mfa.v1.MFAPolicy) |  | Policy is the MFA policy for the account |
 | `found` | [bool](#bool) |  | Found indicates if the policy was found |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryMFARequiredRequest"></a>

 ### QueryMFARequiredRequest
 QueryMFARequiredRequest is the request for QueryMFARequired

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to check |
 | `transaction_type` | [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType) |  | TransactionType is the transaction type to check |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryMFARequiredResponse"></a>

 ### QueryMFARequiredResponse
 QueryMFARequiredResponse is the response for QueryMFARequired

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `required` | [bool](#bool) |  | Required indicates if MFA is required |
 | `factor_combinations` | [FactorCombination](#virtengine.mfa.v1.FactorCombination) | repeated | FactorCombinations lists the acceptable factor combinations |
 | `min_veid_score` | [uint32](#uint32) |  | MinVEIDScore is the minimum VEID score required |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for QueryParams

 

 

 
 <a name="virtengine.mfa.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for QueryParams

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.mfa.v1.Params) |  | Params are the module parameters |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryPendingChallengesRequest"></a>

 ### QueryPendingChallengesRequest
 QueryPendingChallengesRequest is the request for QueryPendingChallenges

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryPendingChallengesResponse"></a>

 ### QueryPendingChallengesResponse
 QueryPendingChallengesResponse is the response for QueryPendingChallenges

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenges` | [Challenge](#virtengine.mfa.v1.Challenge) | repeated | Challenges is the list of pending challenges |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.mfa.v1.QuerySensitiveTxConfigRequest"></a>

 ### QuerySensitiveTxConfigRequest
 QuerySensitiveTxConfigRequest is the request for QuerySensitiveTxConfig

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `transaction_type` | [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType) |  | TransactionType is the transaction type to query |
 
 

 

 
 <a name="virtengine.mfa.v1.QuerySensitiveTxConfigResponse"></a>

 ### QuerySensitiveTxConfigResponse
 QuerySensitiveTxConfigResponse is the response for QuerySensitiveTxConfig

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `config` | [SensitiveTxConfig](#virtengine.mfa.v1.SensitiveTxConfig) |  | Config is the sensitive transaction configuration |
 | `found` | [bool](#bool) |  | Found indicates if the config was found |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryTrustedDevicesRequest"></a>

 ### QueryTrustedDevicesRequest
 QueryTrustedDevicesRequest is the request for QueryTrustedDevices

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.mfa.v1.QueryTrustedDevicesResponse"></a>

 ### QueryTrustedDevicesResponse
 QueryTrustedDevicesResponse is the response for QueryTrustedDevices

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `devices` | [TrustedDevice](#virtengine.mfa.v1.TrustedDevice) | repeated | Devices is the list of trusted devices |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.mfa.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the mfa module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `MFAPolicy` | [QueryMFAPolicyRequest](#virtengine.mfa.v1.QueryMFAPolicyRequest) | [QueryMFAPolicyResponse](#virtengine.mfa.v1.QueryMFAPolicyResponse) | MFAPolicy returns the MFA policy for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/policy/{address}|
 | `FactorEnrollments` | [QueryFactorEnrollmentsRequest](#virtengine.mfa.v1.QueryFactorEnrollmentsRequest) | [QueryFactorEnrollmentsResponse](#virtengine.mfa.v1.QueryFactorEnrollmentsResponse) | FactorEnrollments returns all factor enrollments for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/enrollments/{address}|
 | `FactorEnrollment` | [QueryFactorEnrollmentRequest](#virtengine.mfa.v1.QueryFactorEnrollmentRequest) | [QueryFactorEnrollmentResponse](#virtengine.mfa.v1.QueryFactorEnrollmentResponse) | FactorEnrollment returns a specific factor enrollment buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/enrollment/{address}/{factor_id}|
 | `Challenge` | [QueryChallengeRequest](#virtengine.mfa.v1.QueryChallengeRequest) | [QueryChallengeResponse](#virtengine.mfa.v1.QueryChallengeResponse) | Challenge returns a challenge by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/challenge/{challenge_id}|
 | `PendingChallenges` | [QueryPendingChallengesRequest](#virtengine.mfa.v1.QueryPendingChallengesRequest) | [QueryPendingChallengesResponse](#virtengine.mfa.v1.QueryPendingChallengesResponse) | PendingChallenges returns pending challenges for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/challenges/{address}|
 | `AuthorizationSession` | [QueryAuthorizationSessionRequest](#virtengine.mfa.v1.QueryAuthorizationSessionRequest) | [QueryAuthorizationSessionResponse](#virtengine.mfa.v1.QueryAuthorizationSessionResponse) | AuthorizationSession returns an authorization session by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/session/{session_id}|
 | `TrustedDevices` | [QueryTrustedDevicesRequest](#virtengine.mfa.v1.QueryTrustedDevicesRequest) | [QueryTrustedDevicesResponse](#virtengine.mfa.v1.QueryTrustedDevicesResponse) | TrustedDevices returns trusted devices for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/devices/{address}|
 | `SensitiveTxConfig` | [QuerySensitiveTxConfigRequest](#virtengine.mfa.v1.QuerySensitiveTxConfigRequest) | [QuerySensitiveTxConfigResponse](#virtengine.mfa.v1.QuerySensitiveTxConfigResponse) | SensitiveTxConfig returns the configuration for a sensitive tx type buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/sensitive_tx/{transaction_type}|
 | `AllSensitiveTxConfigs` | [QueryAllSensitiveTxConfigsRequest](#virtengine.mfa.v1.QueryAllSensitiveTxConfigsRequest) | [QueryAllSensitiveTxConfigsResponse](#virtengine.mfa.v1.QueryAllSensitiveTxConfigsResponse) | AllSensitiveTxConfigs returns all sensitive tx configurations buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/sensitive_tx|
 | `MFARequired` | [QueryMFARequiredRequest](#virtengine.mfa.v1.QueryMFARequiredRequest) | [QueryMFARequiredResponse](#virtengine.mfa.v1.QueryMFARequiredResponse) | MFARequired checks if MFA is required for a transaction buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/required/{address}/{transaction_type}|
 | `Params` | [QueryParamsRequest](#virtengine.mfa.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.mfa.v1.QueryParamsResponse) | Params returns the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/mfa/v1/params|
 
  <!-- end services -->

 
 
 <a name="virtengine/mfa/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/mfa/v1/tx.proto
 

 
 <a name="virtengine.mfa.v1.MsgAddTrustedDevice"></a>

 ### MsgAddTrustedDevice
 MsgAddTrustedDevice is the message for adding a trusted device

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account adding the device |
 | `device_info` | [DeviceInfo](#virtengine.mfa.v1.DeviceInfo) |  | DeviceInfo contains the device information |
 | `mfa_proof` | [MFAProof](#virtengine.mfa.v1.MFAProof) |  | MFAProof is proof of MFA for this operation |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgAddTrustedDeviceResponse"></a>

 ### MsgAddTrustedDeviceResponse
 MsgAddTrustedDeviceResponse is the response for MsgAddTrustedDevice

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `success` | [bool](#bool) |  | Success indicates if the device was added |
 | `trust_expires_at` | [int64](#int64) |  | TrustExpiresAt is when the device trust expires |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgCreateChallenge"></a>

 ### MsgCreateChallenge
 MsgCreateChallenge is the message for creating an MFA challenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account requesting the challenge |
 | `factor_type` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorType is the type of factor to challenge |
 | `factor_id` | [string](#string) |  | FactorID is the specific factor to challenge (optional) |
 | `transaction_type` | [SensitiveTransactionType](#virtengine.mfa.v1.SensitiveTransactionType) |  | TransactionType is the sensitive transaction this challenge is for |
 | `client_info` | [ClientInfo](#virtengine.mfa.v1.ClientInfo) |  | ClientInfo contains client information |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgCreateChallengeResponse"></a>

 ### MsgCreateChallengeResponse
 MsgCreateChallengeResponse is the response for MsgCreateChallenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `challenge_id` | [string](#string) |  | ChallengeID is the unique identifier for the challenge |
 | `challenge_data` | [bytes](#bytes) |  | ChallengeData contains factor-specific challenge data |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the challenge expires |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgEnrollFactor"></a>

 ### MsgEnrollFactor
 MsgEnrollFactor is the message for enrolling a new MFA factor

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account enrolling the factor |
 | `factor_type` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorType is the type of factor to enroll |
 | `label` | [string](#string) |  | Label is a user-friendly label for the factor |
 | `public_identifier` | [bytes](#bytes) |  | PublicIdentifier is the public component (for FIDO2, hardware keys) |
 | `metadata` | [FactorMetadata](#virtengine.mfa.v1.FactorMetadata) |  | Metadata contains factor-specific metadata |
 | `initial_verification_proof` | [bytes](#bytes) |  | InitialVerificationProof is proof of factor possession during enrollment |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgEnrollFactorResponse"></a>

 ### MsgEnrollFactorResponse
 MsgEnrollFactorResponse is the response for MsgEnrollFactor

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `factor_id` | [string](#string) |  | FactorID is the unique identifier for the enrolled factor |
 | `status` | [FactorEnrollmentStatus](#virtengine.mfa.v1.FactorEnrollmentStatus) |  | Status is the enrollment status |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgRemoveTrustedDevice"></a>

 ### MsgRemoveTrustedDevice
 MsgRemoveTrustedDevice is the message for removing a trusted device

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account removing the device |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is the fingerprint of the device to remove |
 | `mfa_proof` | [MFAProof](#virtengine.mfa.v1.MFAProof) |  | MFAProof is proof of MFA for this operation (optional if removing from trusted device) |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgRemoveTrustedDeviceResponse"></a>

 ### MsgRemoveTrustedDeviceResponse
 MsgRemoveTrustedDeviceResponse is the response for MsgRemoveTrustedDevice

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `success` | [bool](#bool) |  | Success indicates if the device was removed |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgRevokeFactor"></a>

 ### MsgRevokeFactor
 MsgRevokeFactor is the message for revoking an enrolled factor

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account revoking the factor |
 | `factor_type` | [FactorType](#virtengine.mfa.v1.FactorType) |  | FactorType is the type of factor to revoke |
 | `factor_id` | [string](#string) |  | FactorID is the ID of the factor to revoke |
 | `mfa_proof` | [MFAProof](#virtengine.mfa.v1.MFAProof) |  | MFAProof is proof of MFA for this operation (required if MFA is enabled) |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgRevokeFactorResponse"></a>

 ### MsgRevokeFactorResponse
 MsgRevokeFactorResponse is the response for MsgRevokeFactor

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `success` | [bool](#bool) |  | Success indicates if the revocation was successful |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgSetMFAPolicy"></a>

 ### MsgSetMFAPolicy
 MsgSetMFAPolicy is the message for setting MFA policy

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account setting the policy |
 | `policy` | [MFAPolicy](#virtengine.mfa.v1.MFAPolicy) |  | Policy is the new MFA policy |
 | `mfa_proof` | [MFAProof](#virtengine.mfa.v1.MFAProof) |  | MFAProof is proof of MFA for this operation (required if MFA is already enabled) |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgSetMFAPolicyResponse"></a>

 ### MsgSetMFAPolicyResponse
 MsgSetMFAPolicyResponse is the response for MsgSetMFAPolicy

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `success` | [bool](#bool) |  | Success indicates if the policy update was successful |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module (x/gov module account) |
 | `params` | [Params](#virtengine.mfa.v1.Params) |  | Params are the new module parameters |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

 
 <a name="virtengine.mfa.v1.MsgUpdateSensitiveTxConfig"></a>

 ### MsgUpdateSensitiveTxConfig
 MsgUpdateSensitiveTxConfig is the message for updating sensitive tx config (governance)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the governance authority |
 | `config` | [SensitiveTxConfig](#virtengine.mfa.v1.SensitiveTxConfig) |  | Config is the new configuration |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgUpdateSensitiveTxConfigResponse"></a>

 ### MsgUpdateSensitiveTxConfigResponse
 MsgUpdateSensitiveTxConfigResponse is the response for MsgUpdateSensitiveTxConfig

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `success` | [bool](#bool) |  | Success indicates if the update was successful |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgVerifyChallenge"></a>

 ### MsgVerifyChallenge
 MsgVerifyChallenge is the message for verifying a challenge response

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account verifying the challenge |
 | `challenge_id` | [string](#string) |  | ChallengeID is the challenge being verified |
 | `response` | [ChallengeResponse](#virtengine.mfa.v1.ChallengeResponse) |  | Response is the challenge response |
 
 

 

 
 <a name="virtengine.mfa.v1.MsgVerifyChallengeResponse"></a>

 ### MsgVerifyChallengeResponse
 MsgVerifyChallengeResponse is the response for MsgVerifyChallenge

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `verified` | [bool](#bool) |  | Verified indicates if verification was successful |
 | `session_id` | [string](#string) |  | SessionID is the authorization session ID if verified |
 | `session_expires_at` | [int64](#int64) |  | SessionExpiresAt is when the session expires |
 | `remaining_factors` | [FactorType](#virtengine.mfa.v1.FactorType) | repeated | RemainingFactors lists factors still needed to satisfy policy |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.mfa.v1.Msg"></a>

 ### Msg
 Msg defines the mfa Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `EnrollFactor` | [MsgEnrollFactor](#virtengine.mfa.v1.MsgEnrollFactor) | [MsgEnrollFactorResponse](#virtengine.mfa.v1.MsgEnrollFactorResponse) | EnrollFactor enrolls a new MFA factor | |
 | `RevokeFactor` | [MsgRevokeFactor](#virtengine.mfa.v1.MsgRevokeFactor) | [MsgRevokeFactorResponse](#virtengine.mfa.v1.MsgRevokeFactorResponse) | RevokeFactor revokes an enrolled factor | |
 | `SetMFAPolicy` | [MsgSetMFAPolicy](#virtengine.mfa.v1.MsgSetMFAPolicy) | [MsgSetMFAPolicyResponse](#virtengine.mfa.v1.MsgSetMFAPolicyResponse) | SetMFAPolicy sets the MFA policy for an account | |
 | `CreateChallenge` | [MsgCreateChallenge](#virtengine.mfa.v1.MsgCreateChallenge) | [MsgCreateChallengeResponse](#virtengine.mfa.v1.MsgCreateChallengeResponse) | CreateChallenge creates an MFA challenge | |
 | `VerifyChallenge` | [MsgVerifyChallenge](#virtengine.mfa.v1.MsgVerifyChallenge) | [MsgVerifyChallengeResponse](#virtengine.mfa.v1.MsgVerifyChallengeResponse) | VerifyChallenge verifies an MFA challenge response | |
 | `AddTrustedDevice` | [MsgAddTrustedDevice](#virtengine.mfa.v1.MsgAddTrustedDevice) | [MsgAddTrustedDeviceResponse](#virtengine.mfa.v1.MsgAddTrustedDeviceResponse) | AddTrustedDevice adds a trusted device | |
 | `RemoveTrustedDevice` | [MsgRemoveTrustedDevice](#virtengine.mfa.v1.MsgRemoveTrustedDevice) | [MsgRemoveTrustedDeviceResponse](#virtengine.mfa.v1.MsgRemoveTrustedDeviceResponse) | RemoveTrustedDevice removes a trusted device | |
 | `UpdateSensitiveTxConfig` | [MsgUpdateSensitiveTxConfig](#virtengine.mfa.v1.MsgUpdateSensitiveTxConfig) | [MsgUpdateSensitiveTxConfigResponse](#virtengine.mfa.v1.MsgUpdateSensitiveTxConfigResponse) | UpdateSensitiveTxConfig updates sensitive transaction configuration (governance only) | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.mfa.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.mfa.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/prices.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/prices.proto
 

 
 <a name="virtengine.oracle.v1.AggregatedPrice"></a>

 ### AggregatedPrice
 AggregatedPrice represents the final aggregated price from all sources

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | denom is the asset denomination |
 | `twap` | [string](#string) |  | twap is the time-weighted average price over the configured window |
 | `median_price` | [string](#string) |  | median_price is the median of all source prices |
 | `min_price` | [string](#string) |  | min_price is the minimum price from all sources |
 | `max_price` | [string](#string) |  | max_price is the maximum price from all sources |
 | `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | timestamp is when the aggregated price was computed |
 | `num_sources` | [uint32](#uint32) |  | num_sources is the number of price sources contributing to this aggregation |
 | `deviation_bps` | [uint64](#uint64) |  | deviation_bps is the price deviation in basis points between min and max prices |
 
 

 

 
 <a name="virtengine.oracle.v1.DataID"></a>

 ### DataID
 DataID uniquely identifies a price pair by asset and base denomination

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | denom is the asset denomination (e.g., "uve") |
 | `base_denom` | [string](#string) |  | base_denom is the base denomination for the price pair (e.g., "usd") |
 
 

 

 
 <a name="virtengine.oracle.v1.PriceData"></a>

 ### PriceData
 PriceData combines a price record identifier with its state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `id` | [PriceDataRecordID](#virtengine.oracle.v1.PriceDataRecordID) |  | id uniquely identifies this price record |
 | `state` | [PriceDataState](#virtengine.oracle.v1.PriceDataState) |  | state contains the price value and timestamp |
 
 

 

 
 <a name="virtengine.oracle.v1.PriceDataID"></a>

 ### PriceDataID
 PriceDataID identifies price data from a specific source for a specific pair

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `source` | [uint32](#uint32) |  | source is the index of the price source (oracle provider) |
 | `denom` | [string](#string) |  | denom is the asset denomination |
 | `base_denom` | [string](#string) |  | base_denom is the base denomination for the price pair |
 
 

 

 
 <a name="virtengine.oracle.v1.PriceDataRecordID"></a>

 ### PriceDataRecordID
 PriceDataRecordID represents a price from a specific source at a specific time.
It also represents a single data point in TWAP history

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `source` | [uint32](#uint32) |  | source is the index of the price source (oracle provider) |
 | `denom` | [string](#string) |  | denom is the asset denomination |
 | `base_denom` | [string](#string) |  | base_denom is the base denomination for the price pair |
 | `height` | [int64](#int64) |  | height is the block height when this price was recorded |
 
 

 

 
 <a name="virtengine.oracle.v1.PriceDataState"></a>

 ### PriceDataState
 PriceDataState represents the price value and timestamp for a price entry

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `price` | [string](#string) |  | price is the decimal price value |
 | `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | timestamp is when the price was recorded |
 
 

 

 
 <a name="virtengine.oracle.v1.PriceHealth"></a>

 ### PriceHealth
 PriceHealth represents the health status of a price feed

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | denom is the asset denomination |
 | `is_healthy` | [bool](#bool) |  | is_healthy indicates if the price feed meets all health requirements |
 | `has_min_sources` | [bool](#bool) |  | has_min_sources indicates if minimum number of sources are reporting |
 | `deviation_ok` | [bool](#bool) |  | deviation_ok indicates if price deviation is within acceptable limits |
 | `total_sources` | [uint32](#uint32) |  | total_sources indicates total amount of sources registered for price calculations |
 | `total_healthy_sources` | [uint32](#uint32) |  | total_healthy_sources indicates total usable sources for price calculations |
 | `failure_reason` | [string](#string) | repeated | failure_reason lists reasons for unhealthy status, if any |
 
 

 

 
 <a name="virtengine.oracle.v1.PricesFilter"></a>

 ### PricesFilter
 PricesFilter defines filters used to query price data

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `asset_denom` | [string](#string) |  | asset_denom is the asset denomination to filter by |
 | `base_denom` | [string](#string) |  | base_denom is the base denomination to filter by |
 | `height` | [int64](#int64) |  | height is the block height to filter by |
 
 

 

 
 <a name="virtengine.oracle.v1.QueryPricesRequest"></a>

 ### QueryPricesRequest
 QueryPricesRequest is the request type for querying price history

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `filters` | [PricesFilter](#virtengine.oracle.v1.PricesFilter) |  | filters holds the price fields to filter the request |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination is used to paginate the request |
 
 

 

 
 <a name="virtengine.oracle.v1.QueryPricesResponse"></a>

 ### QueryPricesResponse
 QueryPricesResponse is the response type for querying price history

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `prices` | [PriceData](#virtengine.oracle.v1.PriceData) | repeated | prices is the list of historical price data matching the filters |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination contains the information about response pagination |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/events.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/events.proto
 

 
 <a name="virtengine.oracle.v1.EventPriceData"></a>

 ### EventPriceData
 EventPriceData is emitted when new price data is added to the oracle

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `source` | [string](#string) |  | source is the address of the price source (oracle provider) |
 | `id` | [DataID](#virtengine.oracle.v1.DataID) |  | id identifies the price pair (denom and base_denom) |
 | `data` | [PriceDataState](#virtengine.oracle.v1.PriceDataState) |  | data contains the price value and timestamp |
 
 

 

 
 <a name="virtengine.oracle.v1.EventPriceRecovered"></a>

 ### EventPriceRecovered
 EventPriceRecovered is emitted when a stale price has started receiving updates again

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `source` | [string](#string) |  | source is the address of the price source |
 | `id` | [DataID](#virtengine.oracle.v1.DataID) |  | id identifies the price pair |
 | `height` | [int64](#int64) |  | height is the block height when the price recovery was detected |
 
 

 

 
 <a name="virtengine.oracle.v1.EventPriceStaleWarning"></a>

 ### EventPriceStaleWarning
 EventPriceStaleWarning is emitted when price has not been updated and is about to become stale

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `source` | [string](#string) |  | source is the address of the price source |
 | `id` | [DataID](#virtengine.oracle.v1.DataID) |  | id identifies the price pair |
 | `last_height` | [int64](#int64) |  | last_height is the block height when the price was last updated |
 | `blocks_to_stall` | [int64](#int64) |  | blocks_to_stall is the number of blocks until the price becomes stale |
 
 

 

 
 <a name="virtengine.oracle.v1.EventPriceStaled"></a>

 ### EventPriceStaled
 EventPriceStaled is emitted when a price has become stale

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `source` | [string](#string) |  | source is the address of the price source |
 | `id` | [DataID](#virtengine.oracle.v1.DataID) |  | id identifies the price pair |
 | `last_height` | [int64](#int64) |  | last_height is the block height when the price was last updated before becoming stale |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/params.proto
 

 
 <a name="virtengine.oracle.v1.Params"></a>

 ### Params
 Params defines the parameters for the oracle module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sources` | [string](#string) | repeated | sources addresses allowed to write prices into oracle module those are to be smartcontract addresses |
 | `min_price_sources` | [uint32](#uint32) |  | Minimum number of price sources required (default: 2) |
 | `max_price_staleness_blocks` | [int64](#int64) |  | Maximum price staleness in blocks (default: 50 = ~ 5 minutes) |
 | `twap_window` | [int64](#int64) |  | TWAP window in blocks (default: 50 = ~ 5 minutes) |
 | `max_price_deviation_bps` | [uint64](#uint64) |  | Maximum price deviation in basis points (default: 150 = 1.5%) |
 | `feed_contracts_params` | [google.protobuf.Any](#google.protobuf.Any) | repeated | feed_contracts_params contains the configuration for the price feed contracts |
 
 

 

 
 <a name="virtengine.oracle.v1.PythContractParams"></a>

 ### PythContractParams
 PythContractParams contains configuration for Pyth price feeds

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `akt_price_feed_id` | [string](#string) |  | akt_price_feed_id is the Pyth price feed identifier for AKT/USD |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/genesis.proto
 

 
 <a name="virtengine.oracle.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the oracle module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.oracle.v1.Params) |  | params holds the oracle module parameters |
 | `prices` | [PriceData](#virtengine.oracle.v1.PriceData) | repeated | prices is the list of all historical price data entries |
 | `latest_height` | [PriceDataID](#virtengine.oracle.v1.PriceDataID) | repeated | latest_height tracks the most recent block height for each price feed source |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/msgs.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/msgs.proto
 

 
 <a name="virtengine.oracle.v1.MsgAddPriceEntry"></a>

 ### MsgAddPriceEntry
 MsgAddPriceEntry defines an SDK message to add oracle price entry.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `signer` | [string](#string) |  | Signer is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `id` | [DataID](#virtengine.oracle.v1.DataID) |  | id uniquely identifies the price data by denomination and base denomination |
 | `price` | [PriceDataState](#virtengine.oracle.v1.PriceDataState) |  | price contains the price value and timestamp for this entry |
 
 

 

 
 <a name="virtengine.oracle.v1.MsgAddPriceEntryResponse"></a>

 ### MsgAddPriceEntryResponse
 MsgAddPriceEntryResponse defines the Msg/MsgAddDPriceEntry response type.

 

 

 
 <a name="virtengine.oracle.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the Msg/UpdateParams request type.

Since: akash v2.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.oracle.v1.Params) |  | params defines the x/oracle parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.oracle.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: akash v2.0.0

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/query.proto
 

 
 <a name="virtengine.oracle.v1.QueryAggregatedPriceRequest"></a>

 ### QueryAggregatedPriceRequest
 QueryAggregatedPriceRequest is the request type for aggregated price.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | denom is the asset denomination |
 
 

 

 
 <a name="virtengine.oracle.v1.QueryAggregatedPriceResponse"></a>

 ### QueryAggregatedPriceResponse
 QueryAggregatedPriceResponse is the response type for aggregated price.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `aggregated_price` | [AggregatedPrice](#virtengine.oracle.v1.AggregatedPrice) |  | aggregated_price is the aggregated price data |
 | `price_health` | [PriceHealth](#virtengine.oracle.v1.PriceHealth) |  | price_health is the health status for the price feed |
 
 

 

 
 <a name="virtengine.oracle.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request type for the Query/Params RPC method.

 

 

 
 <a name="virtengine.oracle.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response type for the Query/Params RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.oracle.v1.Params) |  | params defines the parameters of the module. |
 
 

 

 
 <a name="virtengine.oracle.v1.QueryPriceFeedConfigRequest"></a>

 ### QueryPriceFeedConfigRequest
 QueryPriceFeedConfigRequest is the request type for price feed config.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | denom is the denomination to query the price feed configuration for |
 
 

 

 
 <a name="virtengine.oracle.v1.QueryPriceFeedConfigResponse"></a>

 ### QueryPriceFeedConfigResponse
 QueryPriceFeedConfigResponse is the response type for price feed config.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `price_feed_id` | [string](#string) |  | price_feed_id is the Pyth price feed identifier for this denomination |
 | `pyth_contract_address` | [string](#string) |  | pyth_contract_address is the address of the Pyth smart contract |
 | `enabled` | [bool](#bool) |  | enabled indicates if the price feed is enabled for this denomination |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.oracle.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service of the oracle package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Prices` | [QueryPricesRequest](#virtengine.oracle.v1.QueryPricesRequest) | [QueryPricesResponse](#virtengine.oracle.v1.QueryPricesResponse) | Prices query prices for specific denom | GET|/virtengine/oracle/v1/prices|
 | `Params` | [QueryParamsRequest](#virtengine.oracle.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.oracle.v1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/virtengine/oracle/v1/params|
 | `PriceFeedConfig` | [QueryPriceFeedConfigRequest](#virtengine.oracle.v1.QueryPriceFeedConfigRequest) | [QueryPriceFeedConfigResponse](#virtengine.oracle.v1.QueryPriceFeedConfigResponse) | PriceFeedConfig queries the price feed configuration for a given denom. | GET|/virtengine/oracle/v1/price_feed_config/{denom}|
 | `AggregatedPrice` | [QueryAggregatedPriceRequest](#virtengine.oracle.v1.QueryAggregatedPriceRequest) | [QueryAggregatedPriceResponse](#virtengine.oracle.v1.QueryAggregatedPriceResponse) | AggregatedPrice queries the aggregated price for a given denom. | GET|/virtengine/oracle/v1/aggregated_price/{denom}|
 
  <!-- end services -->

 
 
 <a name="virtengine/oracle/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/oracle/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.oracle.v1.Msg"></a>

 ### Msg
 Msg defines the oracle Msg service for managing price feeds

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `AddPriceEntry` | [MsgAddPriceEntry](#virtengine.oracle.v1.MsgAddPriceEntry) | [MsgAddPriceEntryResponse](#virtengine.oracle.v1.MsgAddPriceEntryResponse) | AddPriceEntry adds a new price entry for a denomination from an authorized source | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.oracle.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.oracle.v1.MsgUpdateParamsResponse) | UpdateParams defines a governance operation for updating the x/wasm module parameters. The authority is hard-coded to the x/gov module account.

Since: akash v2.0.0 | |
 
  <!-- end services -->

 
 
 <a name="virtengine/provider/v1beta4/event.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1beta4/event.proto
 

 
 <a name="virtengine.provider.v1beta4.EventProviderCreated"></a>

 ### EventProviderCreated
 EventProviderCreated defines an SDK message for provider created event.
It contains all the required information to identify a provider on-chain.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

 
 <a name="virtengine.provider.v1beta4.EventProviderDeleted"></a>

 ### EventProviderDeleted
 EventProviderDeleted defines an SDK message for provider deleted event.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

 
 <a name="virtengine.provider.v1beta4.EventProviderDomainVerificationStarted"></a>

 ### EventProviderDomainVerificationStarted
 EventProviderDomainVerificationStarted defines an SDK message for when domain verification begins

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  |  |
 | `domain` | [string](#string) |  |  |
 | `token` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.EventProviderDomainVerified"></a>

 ### EventProviderDomainVerified
 EventProviderDomainVerified defines an SDK message for provider domain verified event

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  |  |
 | `domain` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.EventProviderUpdated"></a>

 ### EventProviderUpdated
 EventProviderUpdated defines an SDK message for provider updated event.
It contains all the required information to identify a provider on-chain.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/provider/v1beta4/provider.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1beta4/provider.proto
 

 
 <a name="virtengine.provider.v1beta4.Info"></a>

 ### Info
 Info contains information on the provider.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `email` | [string](#string) |  | Email is the email address to contact the provider. |
 | `website` | [string](#string) |  | Website is the URL to the landing page or socials of the provider. |
 
 

 

 
 <a name="virtengine.provider.v1beta4.Provider"></a>

 ### Provider
 Provider stores owner and host details.
VirtEngine providers are entities that contribute computing resources to the network.
They can be individuals or organizations with underutilized computing resources, such as data centers or personal servers.
Providers participate in the network by running the VirtEngine node software and setting the price for their services.
Users can then choose a provider based on factors such as cost, performance, and location.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `host_uri` | [string](#string) |  | HostURI is the Uniform Resource Identifier for provider connection. This URI is used to directly connect to the provider to perform tasks such as sending the manifest. |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes is a list of arbitrary attribute key-value pairs. |
 | `info` | [Info](#virtengine.provider.v1beta4.Info) |  | Info contains additional provider information. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/provider/v1beta4/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1beta4/genesis.proto
 

 
 <a name="virtengine.provider.v1beta4.GenesisState"></a>

 ### GenesisState
 GenesisState defines the basic genesis state used by provider module.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `providers` | [Provider](#virtengine.provider.v1beta4.Provider) | repeated | Providers is a list of genesis providers. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/provider/v1beta4/msg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1beta4/msg.proto
 

 
 <a name="virtengine.provider.v1beta4.MsgCreateProvider"></a>

 ### MsgCreateProvider
 MsgCreateProvider defines an SDK message for creating a provider.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 | `host_uri` | [string](#string) |  | HostURI is the Uniform Resource Identifier for provider connection. This URI is used to directly connect to the provider to perform tasks such as sending the manifest. |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated | Attributes is a list of arbitrary attribute key-value pairs. |
 | `info` | [Info](#virtengine.provider.v1beta4.Info) |  | Info contains additional provider information. |
 
 

 

 
 <a name="virtengine.provider.v1beta4.MsgCreateProviderResponse"></a>

 ### MsgCreateProviderResponse
 MsgCreateProviderResponse defines the Msg/CreateProvider response type.

 

 

 
 <a name="virtengine.provider.v1beta4.MsgDeleteProvider"></a>

 ### MsgDeleteProvider
 MsgDeleteProvider defines an SDK message for deleting a provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.MsgDeleteProviderResponse"></a>

 ### MsgDeleteProviderResponse
 MsgDeleteProviderResponse defines the Msg/DeleteProvider response type.

 

 

 
 <a name="virtengine.provider.v1beta4.MsgGenerateDomainVerificationToken"></a>

 ### MsgGenerateDomainVerificationToken
 MsgGenerateDomainVerificationToken defines an SDK message for generating a domain verification token

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  |  |
 | `domain` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.MsgGenerateDomainVerificationTokenResponse"></a>

 ### MsgGenerateDomainVerificationTokenResponse
 MsgGenerateDomainVerificationTokenResponse defines the Msg/GenerateDomainVerificationToken response type.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `token` | [string](#string) |  |  |
 | `expires_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.MsgUpdateProvider"></a>

 ### MsgUpdateProvider
 MsgUpdateProvider defines an SDK message for updating a provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  |  |
 | `host_uri` | [string](#string) |  |  |
 | `attributes` | [virtengine.base.attributes.v1.Attribute](#virtengine.base.attributes.v1.Attribute) | repeated |  |
 | `info` | [Info](#virtengine.provider.v1beta4.Info) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.MsgUpdateProviderResponse"></a>

 ### MsgUpdateProviderResponse
 MsgUpdateProviderResponse defines the Msg/UpdateProvider response type.

 

 

 
 <a name="virtengine.provider.v1beta4.MsgVerifyProviderDomain"></a>

 ### MsgVerifyProviderDomain
 MsgVerifyProviderDomain defines an SDK message for verifying a provider's domain

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.provider.v1beta4.MsgVerifyProviderDomainResponse"></a>

 ### MsgVerifyProviderDomainResponse
 MsgVerifyProviderDomainResponse defines the Msg/VerifyProviderDomain response type.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `verified` | [bool](#bool) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/provider/v1beta4/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1beta4/query.proto
 

 
 <a name="virtengine.provider.v1beta4.QueryProviderRequest"></a>

 ### QueryProviderRequest
 QueryProviderRequest is request type for the Query/Provider RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "ve1..." |
 
 

 

 
 <a name="virtengine.provider.v1beta4.QueryProviderResponse"></a>

 ### QueryProviderResponse
 QueryProviderResponse is response type for the Query/Provider RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [Provider](#virtengine.provider.v1beta4.Provider) |  | Provider holds the representation of a provider on the network. |
 
 

 

 
 <a name="virtengine.provider.v1beta4.QueryProvidersRequest"></a>

 ### QueryProvidersRequest
 QueryProvidersRequest is request type for the Query/Providers RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.provider.v1beta4.QueryProvidersResponse"></a>

 ### QueryProvidersResponse
 QueryProvidersResponse is response type for the Query/Providers RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `providers` | [Provider](#virtengine.provider.v1beta4.Provider) | repeated | Providers is a list of providers on the network. |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination contains the information about response pagination. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.provider.v1beta4.Query"></a>

 ### Query
 Query defines the gRPC querier service for the provider package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Providers` | [QueryProvidersRequest](#virtengine.provider.v1beta4.QueryProvidersRequest) | [QueryProvidersResponse](#virtengine.provider.v1beta4.QueryProvidersResponse) | Providers queries providers | GET|/virtengine/provider/v1beta4/providers|
 | `Provider` | [QueryProviderRequest](#virtengine.provider.v1beta4.QueryProviderRequest) | [QueryProviderResponse](#virtengine.provider.v1beta4.QueryProviderResponse) | Provider queries provider details | GET|/virtengine/provider/v1beta4/providers/{owner}|
 
  <!-- end services -->

 
 
 <a name="virtengine/provider/v1beta4/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/provider/v1beta4/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.provider.v1beta4.Msg"></a>

 ### Msg
 Msg defines the provider Msg service.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `CreateProvider` | [MsgCreateProvider](#virtengine.provider.v1beta4.MsgCreateProvider) | [MsgCreateProviderResponse](#virtengine.provider.v1beta4.MsgCreateProviderResponse) | CreateProvider defines a method that creates a provider given the proper inputs. | |
 | `UpdateProvider` | [MsgUpdateProvider](#virtengine.provider.v1beta4.MsgUpdateProvider) | [MsgUpdateProviderResponse](#virtengine.provider.v1beta4.MsgUpdateProviderResponse) | UpdateProvider defines a method that updates a provider given the proper inputs. | |
 | `DeleteProvider` | [MsgDeleteProvider](#virtengine.provider.v1beta4.MsgDeleteProvider) | [MsgDeleteProviderResponse](#virtengine.provider.v1beta4.MsgDeleteProviderResponse) | DeleteProvider defines a method that deletes a provider given the proper inputs. | |
 | `GenerateDomainVerificationToken` | [MsgGenerateDomainVerificationToken](#virtengine.provider.v1beta4.MsgGenerateDomainVerificationToken) | [MsgGenerateDomainVerificationTokenResponse](#virtengine.provider.v1beta4.MsgGenerateDomainVerificationTokenResponse) | GenerateDomainVerificationToken generates a verification token for a provider's domain. | |
 | `VerifyProviderDomain` | [MsgVerifyProviderDomain](#virtengine.provider.v1beta4.MsgVerifyProviderDomain) | [MsgVerifyProviderDomainResponse](#virtengine.provider.v1beta4.MsgVerifyProviderDomainResponse) | VerifyProviderDomain verifies a provider's domain via DNS TXT record. | |
 
  <!-- end services -->

 
 
 <a name="virtengine/review/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/review/v1/tx.proto
 

 
 <a name="virtengine.review.v1.GenesisState"></a>

 ### GenesisState
 GenesisState is the genesis state for the review module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.review.v1.Params) |  |  |
 | `reviews` | [Review](#virtengine.review.v1.Review) | repeated |  |
 | `review_sequence` | [uint64](#uint64) |  |  |
 
 

 

 
 <a name="virtengine.review.v1.MsgDeleteReview"></a>

 ### MsgDeleteReview
 MsgDeleteReview deletes a review

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `review_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.review.v1.MsgDeleteReviewResponse"></a>

 ### MsgDeleteReviewResponse
 MsgDeleteReviewResponse is the response for MsgDeleteReview

 

 

 
 <a name="virtengine.review.v1.MsgSubmitReview"></a>

 ### MsgSubmitReview
 MsgSubmitReview submits a new review

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reviewer` | [string](#string) |  |  |
 | `subject_address` | [string](#string) |  |  |
 | `subject_type` | [string](#string) |  |  |
 | `order_id` | [string](#string) |  |  |
 | `rating` | [uint32](#uint32) |  |  |
 | `comment` | [string](#string) |  |  |
 | `lease_id` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.review.v1.MsgSubmitReviewResponse"></a>

 ### MsgSubmitReviewResponse
 MsgSubmitReviewResponse is the response for MsgSubmitReview

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `review_id` | [string](#string) |  |  |
 | `submitted_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.review.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams updates module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  |  |
 | `params` | [Params](#virtengine.review.v1.Params) |  |  |
 
 

 

 
 <a name="virtengine.review.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

 
 <a name="virtengine.review.v1.Params"></a>

 ### Params
 Params defines the parameters for the review module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `min_review_interval` | [uint64](#uint64) |  |  |
 | `max_comment_length` | [uint64](#uint64) |  |  |
 | `require_completed_order` | [bool](#bool) |  |  |
 | `review_window` | [uint64](#uint64) |  |  |
 | `min_rating` | [uint32](#uint32) |  |  |
 | `max_rating` | [uint32](#uint32) |  |  |
 
 

 

 
 <a name="virtengine.review.v1.Review"></a>

 ### Review
 Review represents a review

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `review_id` | [string](#string) |  |  |
 | `reviewer` | [string](#string) |  |  |
 | `subject_address` | [string](#string) |  |  |
 | `subject_type` | [string](#string) |  |  |
 | `order_id` | [string](#string) |  |  |
 | `lease_id` | [string](#string) |  |  |
 | `rating` | [uint32](#uint32) |  |  |
 | `comment` | [string](#string) |  |  |
 | `submitted_at` | [int64](#int64) |  |  |
 | `block_height` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.review.v1.Msg"></a>

 ### Msg
 Msg defines the review Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `SubmitReview` | [MsgSubmitReview](#virtengine.review.v1.MsgSubmitReview) | [MsgSubmitReviewResponse](#virtengine.review.v1.MsgSubmitReviewResponse) | SubmitReview submits a new review | |
 | `DeleteReview` | [MsgDeleteReview](#virtengine.review.v1.MsgDeleteReview) | [MsgDeleteReviewResponse](#virtengine.review.v1.MsgDeleteReviewResponse) | DeleteReview deletes a review | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.review.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.review.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/roles/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/roles/v1/types.proto
 

 
 <a name="virtengine.roles.v1.AccountStateRecord"></a>

 ### AccountStateRecord
 AccountStateRecord represents the stored state of an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address |
 | `state` | [AccountState](#virtengine.roles.v1.AccountState) |  | State is the current account state |
 | `reason` | [string](#string) |  | Reason is the reason for the current state |
 | `modified_by` | [string](#string) |  | ModifiedBy is the address that last modified the state |
 | `modified_at` | [int64](#int64) |  | ModifiedAt is the Unix timestamp when the state was last modified |
 | `previous_state` | [AccountState](#virtengine.roles.v1.AccountState) |  | PreviousState is the previous account state |
 
 

 

 
 <a name="virtengine.roles.v1.EventAccountStateChanged"></a>

 ### EventAccountStateChanged
 EventAccountStateChanged is emitted when an account state changes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account whose state changed |
 | `previous_state` | [string](#string) |  | PreviousState is the previous state |
 | `new_state` | [string](#string) |  | NewState is the new state |
 | `modified_by` | [string](#string) |  | ModifiedBy is the address that modified the state |
 | `reason` | [string](#string) |  | Reason is the reason for the state change |
 
 

 

 
 <a name="virtengine.roles.v1.EventAdminNominated"></a>

 ### EventAdminNominated
 EventAdminNominated is emitted when a new administrator is nominated

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the nominated administrator |
 | `nominated_by` | [string](#string) |  | NominatedBy is the address that made the nomination |
 
 

 

 
 <a name="virtengine.roles.v1.EventRoleAssigned"></a>

 ### EventRoleAssigned
 EventRoleAssigned is emitted when a role is assigned to an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account that received the role |
 | `role` | [string](#string) |  | Role is the assigned role |
 | `assigned_by` | [string](#string) |  | AssignedBy is the address that assigned the role |
 
 

 

 
 <a name="virtengine.roles.v1.EventRoleRevoked"></a>

 ### EventRoleRevoked
 EventRoleRevoked is emitted when a role is revoked from an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account that lost the role |
 | `role` | [string](#string) |  | Role is the revoked role |
 | `revoked_by` | [string](#string) |  | RevokedBy is the address that revoked the role |
 
 

 

 
 <a name="virtengine.roles.v1.Params"></a>

 ### Params
 Params defines the parameters for the roles module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_roles_per_account` | [uint32](#uint32) |  | MaxRolesPerAccount is the maximum number of roles an account can have |
 | `allow_self_revoke` | [bool](#bool) |  | AllowSelfRevoke determines if accounts can revoke their own roles |
 
 

 

 
 <a name="virtengine.roles.v1.RoleAssignment"></a>

 ### RoleAssignment
 RoleAssignment represents a role assigned to an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address that has the role |
 | `role` | [Role](#virtengine.roles.v1.Role) |  | Role is the assigned role |
 | `assigned_by` | [string](#string) |  | AssignedBy is the address that assigned this role |
 | `assigned_at` | [int64](#int64) |  | AssignedAt is the Unix timestamp when the role was assigned |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.roles.v1.AccountState"></a>

 ### AccountState
 AccountState represents the state of an account

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | ACCOUNT_STATE_UNSPECIFIED | 0 | ACCOUNT_STATE_UNSPECIFIED is the default/invalid state |
 | ACCOUNT_STATE_ACTIVE | 1 | ACCOUNT_STATE_ACTIVE represents an active account |
 | ACCOUNT_STATE_SUSPENDED | 2 | ACCOUNT_STATE_SUSPENDED represents a temporarily suspended account |
 | ACCOUNT_STATE_TERMINATED | 3 | ACCOUNT_STATE_TERMINATED represents a permanently terminated account |
 

 
 <a name="virtengine.roles.v1.Role"></a>

 ### Role
 Role represents the different roles in the VirtEngine system

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | ROLE_UNSPECIFIED | 0 | ROLE_UNSPECIFIED is the default/invalid role |
 | ROLE_GENESIS_ACCOUNT | 1 | ROLE_GENESIS_ACCOUNT represents the highest privilege role - initial chain authority |
 | ROLE_ADMINISTRATOR | 2 | ROLE_ADMINISTRATOR represents platform operations with high trust level |
 | ROLE_MODERATOR | 3 | ROLE_MODERATOR represents content/user moderation with medium-high trust level |
 | ROLE_VALIDATOR | 4 | ROLE_VALIDATOR represents consensus participants with high trust level |
 | ROLE_SERVICE_PROVIDER | 5 | ROLE_SERVICE_PROVIDER represents infrastructure operators with medium trust level |
 | ROLE_CUSTOMER | 6 | ROLE_CUSTOMER represents end users with standard trust level |
 | ROLE_SUPPORT_AGENT | 7 | ROLE_SUPPORT_AGENT represents customer support with medium trust level |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/roles/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/roles/v1/genesis.proto
 

 
 <a name="virtengine.roles.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the roles module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `genesis_accounts` | [string](#string) | repeated | GenesisAccounts are the accounts with GenesisAccount role |
 | `role_assignments` | [RoleAssignment](#virtengine.roles.v1.RoleAssignment) | repeated | RoleAssignments are the initial role assignments |
 | `account_states` | [AccountStateRecord](#virtengine.roles.v1.AccountStateRecord) | repeated | AccountStates are the initial account states |
 | `params` | [Params](#virtengine.roles.v1.Params) |  | Params are the module parameters |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/roles/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/roles/v1/query.proto
 

 
 <a name="virtengine.roles.v1.QueryAccountRolesRequest"></a>

 ### QueryAccountRolesRequest
 QueryAccountRolesRequest is the request for QueryAccountRoles

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query roles for |
 
 

 

 
 <a name="virtengine.roles.v1.QueryAccountRolesResponse"></a>

 ### QueryAccountRolesResponse
 QueryAccountRolesResponse is the response for QueryAccountRoles

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the queried account address |
 | `roles` | [RoleAssignment](#virtengine.roles.v1.RoleAssignment) | repeated | Roles are the role assignments for this account |
 
 

 

 
 <a name="virtengine.roles.v1.QueryAccountStateRequest"></a>

 ### QueryAccountStateRequest
 QueryAccountStateRequest is the request for QueryAccountState

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to query state for |
 
 

 

 
 <a name="virtengine.roles.v1.QueryAccountStateResponse"></a>

 ### QueryAccountStateResponse
 QueryAccountStateResponse is the response for QueryAccountState

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_state` | [AccountStateRecord](#virtengine.roles.v1.AccountStateRecord) |  | AccountState is the account state record |
 | `found` | [bool](#bool) |  | Found indicates if the account state was found |
 
 

 

 
 <a name="virtengine.roles.v1.QueryGenesisAccountsRequest"></a>

 ### QueryGenesisAccountsRequest
 QueryGenesisAccountsRequest is the request for QueryGenesisAccounts

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.roles.v1.QueryGenesisAccountsResponse"></a>

 ### QueryGenesisAccountsResponse
 QueryGenesisAccountsResponse is the response for QueryGenesisAccounts

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `addresses` | [string](#string) | repeated | Addresses are the genesis account addresses |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.roles.v1.QueryHasRoleRequest"></a>

 ### QueryHasRoleRequest
 QueryHasRoleRequest is the request for QueryHasRole

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `address` | [string](#string) |  | Address is the account address to check |
 | `role` | [string](#string) |  | Role is the role to check for |
 
 

 

 
 <a name="virtengine.roles.v1.QueryHasRoleResponse"></a>

 ### QueryHasRoleResponse
 QueryHasRoleResponse is the response for QueryHasRole

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `has_role` | [bool](#bool) |  | HasRole indicates if the account has the specified role |
 | `assignment` | [RoleAssignment](#virtengine.roles.v1.RoleAssignment) |  | Assignment is the role assignment if found |
 
 

 

 
 <a name="virtengine.roles.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for QueryParams

 

 

 
 <a name="virtengine.roles.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for QueryParams

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.roles.v1.Params) |  | Params are the module parameters |
 
 

 

 
 <a name="virtengine.roles.v1.QueryRoleMembersRequest"></a>

 ### QueryRoleMembersRequest
 QueryRoleMembersRequest is the request for QueryRoleMembers

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `role` | [string](#string) |  | Role is the role to query members for |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.roles.v1.QueryRoleMembersResponse"></a>

 ### QueryRoleMembersResponse
 QueryRoleMembersResponse is the response for QueryRoleMembers

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `role` | [string](#string) |  | Role is the queried role |
 | `members` | [RoleAssignment](#virtengine.roles.v1.RoleAssignment) | repeated | Members are the role assignments for this role |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.roles.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the roles module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `AccountRoles` | [QueryAccountRolesRequest](#virtengine.roles.v1.QueryAccountRolesRequest) | [QueryAccountRolesResponse](#virtengine.roles.v1.QueryAccountRolesResponse) | AccountRoles queries all roles for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/roles/v1/account/{address}/roles|
 | `RoleMembers` | [QueryRoleMembersRequest](#virtengine.roles.v1.QueryRoleMembersRequest) | [QueryRoleMembersResponse](#virtengine.roles.v1.QueryRoleMembersResponse) | RoleMembers queries all members of a specific role buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/roles/v1/role/{role}/members|
 | `AccountState` | [QueryAccountStateRequest](#virtengine.roles.v1.QueryAccountStateRequest) | [QueryAccountStateResponse](#virtengine.roles.v1.QueryAccountStateResponse) | AccountState queries the state of an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/roles/v1/account/{address}/state|
 | `GenesisAccounts` | [QueryGenesisAccountsRequest](#virtengine.roles.v1.QueryGenesisAccountsRequest) | [QueryGenesisAccountsResponse](#virtengine.roles.v1.QueryGenesisAccountsResponse) | GenesisAccounts queries all genesis accounts buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/roles/v1/genesis_accounts|
 | `Params` | [QueryParamsRequest](#virtengine.roles.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.roles.v1.QueryParamsResponse) | Params queries the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/roles/v1/params|
 | `HasRole` | [QueryHasRoleRequest](#virtengine.roles.v1.QueryHasRoleRequest) | [QueryHasRoleResponse](#virtengine.roles.v1.QueryHasRoleResponse) | HasRole checks if an account has a specific role buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/roles/v1/account/{address}/has_role/{role}|
 
  <!-- end services -->

 
 
 <a name="virtengine/roles/v1/store.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/roles/v1/store.proto
 

 
 <a name="virtengine.roles.v1.AccountStateStore"></a>

 ### AccountStateStore
 AccountStateStore is the internal storage format for account states

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [uint32](#uint32) |  | State is the current account state (as uint32 for storage) |
 | `reason` | [string](#string) |  | Reason is the reason for the current state |
 | `modified_by` | [string](#string) |  | ModifiedBy is the address that last modified the state |
 | `modified_at` | [int64](#int64) |  | ModifiedAt is the Unix timestamp when the state was last modified |
 | `previous_state` | [uint32](#uint32) |  | PreviousState is the previous account state (as uint32 for storage) |
 
 

 

 
 <a name="virtengine.roles.v1.ParamsStore"></a>

 ### ParamsStore
 ParamsStore is the internal storage format for module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_roles_per_account` | [uint32](#uint32) |  | MaxRolesPerAccount is the maximum number of roles an account can have |
 | `allow_self_revoke` | [bool](#bool) |  | AllowSelfRevoke determines if accounts can revoke their own roles |
 
 

 

 
 <a name="virtengine.roles.v1.RoleAssignmentStore"></a>

 ### RoleAssignmentStore
 RoleAssignmentStore is the internal storage format for role assignments

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `assigned_by` | [string](#string) |  | AssignedBy is the address that assigned this role |
 | `assigned_at` | [int64](#int64) |  | AssignedAt is the Unix timestamp when the role was assigned |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/roles/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/roles/v1/tx.proto
 

 
 <a name="virtengine.roles.v1.MsgAssignRole"></a>

 ### MsgAssignRole
 MsgAssignRole is the message for assigning a role to an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account assigning the role |
 | `address` | [string](#string) |  | Address is the target account to receive the role |
 | `role` | [string](#string) |  | Role is the role to assign |
 
 

 

 
 <a name="virtengine.roles.v1.MsgAssignRoleResponse"></a>

 ### MsgAssignRoleResponse
 MsgAssignRoleResponse is the response for MsgAssignRole

 

 

 
 <a name="virtengine.roles.v1.MsgNominateAdmin"></a>

 ### MsgNominateAdmin
 MsgNominateAdmin is the message for nominating an administrator (GenesisAccount only)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the genesis account making the nomination |
 | `address` | [string](#string) |  | Address is the account to be nominated as administrator |
 
 

 

 
 <a name="virtengine.roles.v1.MsgNominateAdminResponse"></a>

 ### MsgNominateAdminResponse
 MsgNominateAdminResponse is the response for MsgNominateAdmin

 

 

 
 <a name="virtengine.roles.v1.MsgRevokeRole"></a>

 ### MsgRevokeRole
 MsgRevokeRole is the message for revoking a role from an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account revoking the role |
 | `address` | [string](#string) |  | Address is the target account to lose the role |
 | `role` | [string](#string) |  | Role is the role to revoke |
 
 

 

 
 <a name="virtengine.roles.v1.MsgRevokeRoleResponse"></a>

 ### MsgRevokeRoleResponse
 MsgRevokeRoleResponse is the response for MsgRevokeRole

 

 

 
 <a name="virtengine.roles.v1.MsgSetAccountState"></a>

 ### MsgSetAccountState
 MsgSetAccountState is the message for setting an account's state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account setting the state |
 | `address` | [string](#string) |  | Address is the target account whose state is being set |
 | `state` | [string](#string) |  | State is the new state for the account |
 | `reason` | [string](#string) |  | Reason is the reason for the state change |
 | `mfa_proof` | [virtengine.mfa.v1.MFAProof](#virtengine.mfa.v1.MFAProof) |  | MFAProof is optional proof of MFA for sensitive account recovery operations |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is the optional client device fingerprint |
 
 

 

 
 <a name="virtengine.roles.v1.MsgSetAccountStateResponse"></a>

 ### MsgSetAccountStateResponse
 MsgSetAccountStateResponse is the response for MsgSetAccountState

 

 

 
 <a name="virtengine.roles.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module (x/gov module account) |
 | `params` | [Params](#virtengine.roles.v1.Params) |  | Params are the new module parameters |
 
 

 

 
 <a name="virtengine.roles.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.roles.v1.Msg"></a>

 ### Msg
 Msg defines the roles Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `AssignRole` | [MsgAssignRole](#virtengine.roles.v1.MsgAssignRole) | [MsgAssignRoleResponse](#virtengine.roles.v1.MsgAssignRoleResponse) | AssignRole assigns a role to an account | |
 | `RevokeRole` | [MsgRevokeRole](#virtengine.roles.v1.MsgRevokeRole) | [MsgRevokeRoleResponse](#virtengine.roles.v1.MsgRevokeRoleResponse) | RevokeRole revokes a role from an account | |
 | `SetAccountState` | [MsgSetAccountState](#virtengine.roles.v1.MsgSetAccountState) | [MsgSetAccountStateResponse](#virtengine.roles.v1.MsgSetAccountStateResponse) | SetAccountState sets the state of an account | |
 | `NominateAdmin` | [MsgNominateAdmin](#virtengine.roles.v1.MsgNominateAdmin) | [MsgNominateAdminResponse](#virtengine.roles.v1.MsgNominateAdminResponse) | NominateAdmin nominates an administrator (GenesisAccount only) | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.roles.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.roles.v1.MsgUpdateParamsResponse) | UpdateParams updates the module parameters (governance only) | |
 
  <!-- end services -->

 
 
 <a name="virtengine/settlement/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/settlement/v1/tx.proto
 

 
 <a name="virtengine.settlement.v1.MsgAcknowledgeUsage"></a>

 ### MsgAcknowledgeUsage
 MsgAcknowledgeUsage acknowledges a usage record

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `usage_id` | [string](#string) |  |  |
 | `signature` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgAcknowledgeUsageResponse"></a>

 ### MsgAcknowledgeUsageResponse
 MsgAcknowledgeUsageResponse is the response for MsgAcknowledgeUsage

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `acknowledged_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgActivateEscrow"></a>

 ### MsgActivateEscrow
 MsgActivateEscrow activates an escrow when a lease is created

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `escrow_id` | [string](#string) |  |  |
 | `lease_id` | [string](#string) |  |  |
 | `recipient` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgActivateEscrowResponse"></a>

 ### MsgActivateEscrowResponse
 MsgActivateEscrowResponse is the response for MsgActivateEscrow

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `activated_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgClaimRewards"></a>

 ### MsgClaimRewards
 MsgClaimRewards claims accumulated rewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `source` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgClaimRewardsResponse"></a>

 ### MsgClaimRewardsResponse
 MsgClaimRewardsResponse is the response for MsgClaimRewards

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `claimed_amount` | [string](#string) |  |  |
 | `claimed_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgCreateEscrow"></a>

 ### MsgCreateEscrow
 MsgCreateEscrow creates a new escrow account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `order_id` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `expires_in` | [uint64](#uint64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgCreateEscrowResponse"></a>

 ### MsgCreateEscrowResponse
 MsgCreateEscrowResponse is the response for MsgCreateEscrow

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `escrow_id` | [string](#string) |  |  |
 | `created_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgDisputeEscrow"></a>

 ### MsgDisputeEscrow
 MsgDisputeEscrow marks an escrow as disputed

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `escrow_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 | `evidence` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgDisputeEscrowResponse"></a>

 ### MsgDisputeEscrowResponse
 MsgDisputeEscrowResponse is the response for MsgDisputeEscrow

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `disputed_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgRecordUsage"></a>

 ### MsgRecordUsage
 MsgRecordUsage records usage from a provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `order_id` | [string](#string) |  |  |
 | `lease_id` | [string](#string) |  |  |
 | `usage_units` | [uint64](#uint64) |  |  |
 | `usage_type` | [string](#string) |  |  |
 | `period_start` | [int64](#int64) |  |  |
 | `period_end` | [int64](#int64) |  |  |
 | `unit_price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
 | `signature` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgRecordUsageResponse"></a>

 ### MsgRecordUsageResponse
 MsgRecordUsageResponse is the response for MsgRecordUsage

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `usage_id` | [string](#string) |  |  |
 | `total_cost` | [string](#string) |  |  |
 | `recorded_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgRefundEscrow"></a>

 ### MsgRefundEscrow
 MsgRefundEscrow refunds escrow funds to the depositor

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `escrow_id` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgRefundEscrowResponse"></a>

 ### MsgRefundEscrowResponse
 MsgRefundEscrowResponse is the response for MsgRefundEscrow

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `refunded_amount` | [string](#string) |  |  |
 | `refunded_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgReleaseEscrow"></a>

 ### MsgReleaseEscrow
 MsgReleaseEscrow releases escrow funds to the recipient

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `escrow_id` | [string](#string) |  |  |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
 | `reason` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgReleaseEscrowResponse"></a>

 ### MsgReleaseEscrowResponse
 MsgReleaseEscrowResponse is the response for MsgReleaseEscrow

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `released_amount` | [string](#string) |  |  |
 | `released_at` | [int64](#int64) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgSettleOrder"></a>

 ### MsgSettleOrder
 MsgSettleOrder settles an order based on usage records

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  |  |
 | `order_id` | [string](#string) |  |  |
 | `usage_record_ids` | [string](#string) | repeated |  |
 | `is_final` | [bool](#bool) |  |  |
 
 

 

 
 <a name="virtengine.settlement.v1.MsgSettleOrderResponse"></a>

 ### MsgSettleOrderResponse
 MsgSettleOrderResponse is the response for MsgSettleOrder

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `settlement_id` | [string](#string) |  |  |
 | `total_amount` | [string](#string) |  |  |
 | `provider_share` | [string](#string) |  |  |
 | `platform_fee` | [string](#string) |  |  |
 | `settled_at` | [int64](#int64) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.settlement.v1.Msg"></a>

 ### Msg
 Msg defines the settlement Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `CreateEscrow` | [MsgCreateEscrow](#virtengine.settlement.v1.MsgCreateEscrow) | [MsgCreateEscrowResponse](#virtengine.settlement.v1.MsgCreateEscrowResponse) | CreateEscrow creates a new escrow account | |
 | `ActivateEscrow` | [MsgActivateEscrow](#virtengine.settlement.v1.MsgActivateEscrow) | [MsgActivateEscrowResponse](#virtengine.settlement.v1.MsgActivateEscrowResponse) | ActivateEscrow activates an escrow when a lease is created | |
 | `ReleaseEscrow` | [MsgReleaseEscrow](#virtengine.settlement.v1.MsgReleaseEscrow) | [MsgReleaseEscrowResponse](#virtengine.settlement.v1.MsgReleaseEscrowResponse) | ReleaseEscrow releases escrow funds to the recipient | |
 | `RefundEscrow` | [MsgRefundEscrow](#virtengine.settlement.v1.MsgRefundEscrow) | [MsgRefundEscrowResponse](#virtengine.settlement.v1.MsgRefundEscrowResponse) | RefundEscrow refunds escrow funds to the depositor | |
 | `DisputeEscrow` | [MsgDisputeEscrow](#virtengine.settlement.v1.MsgDisputeEscrow) | [MsgDisputeEscrowResponse](#virtengine.settlement.v1.MsgDisputeEscrowResponse) | DisputeEscrow marks an escrow as disputed | |
 | `SettleOrder` | [MsgSettleOrder](#virtengine.settlement.v1.MsgSettleOrder) | [MsgSettleOrderResponse](#virtengine.settlement.v1.MsgSettleOrderResponse) | SettleOrder settles an order based on usage records | |
 | `RecordUsage` | [MsgRecordUsage](#virtengine.settlement.v1.MsgRecordUsage) | [MsgRecordUsageResponse](#virtengine.settlement.v1.MsgRecordUsageResponse) | RecordUsage records usage from a provider | |
 | `AcknowledgeUsage` | [MsgAcknowledgeUsage](#virtengine.settlement.v1.MsgAcknowledgeUsage) | [MsgAcknowledgeUsageResponse](#virtengine.settlement.v1.MsgAcknowledgeUsageResponse) | AcknowledgeUsage acknowledges a usage record | |
 | `ClaimRewards` | [MsgClaimRewards](#virtengine.settlement.v1.MsgClaimRewards) | [MsgClaimRewardsResponse](#virtengine.settlement.v1.MsgClaimRewardsResponse) | ClaimRewards claims accumulated rewards | |
 
  <!-- end services -->

 
 
 <a name="virtengine/staking/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/staking/v1/params.proto
 

 
 <a name="virtengine.staking.v1.Params"></a>

 ### Params
 Params defines the parameters for the staking module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `epoch_length` | [uint64](#uint64) |  | EpochLength is the number of blocks per reward epoch |
 | `base_reward_per_block` | [int64](#int64) |  | BaseRewardPerBlock is the base reward per block in smallest unit |
 | `veid_reward_pool` | [int64](#int64) |  | VEIDRewardPool is the VEID verification reward pool per epoch |
 | `identity_network_reward_pool` | [int64](#int64) |  | IdentityNetworkRewardPool is the identity network reward pool per epoch |
 | `downtime_threshold` | [int64](#int64) |  | DowntimeThreshold is the consecutive missed blocks before slashing |
 | `signed_blocks_window` | [int64](#int64) |  | SignedBlocksWindow is the window size for tracking missed blocks |
 | `min_signed_per_window` | [int64](#int64) |  | MinSignedPerWindow is the minimum percentage that must be signed (fixed-point) |
 | `slash_fraction_double_sign` | [int64](#int64) |  | SlashFractionDoubleSign is the slash fraction for double signing (fixed-point) |
 | `slash_fraction_downtime` | [int64](#int64) |  | SlashFractionDowntime is the slash fraction for downtime (fixed-point) |
 | `slash_fraction_invalid_attestation` | [int64](#int64) |  | SlashFractionInvalidAttestation is the slash fraction for invalid VEID attestation (fixed-point) |
 | `jail_duration_downtime` | [int64](#int64) |  | JailDurationDowntime is the jail duration for downtime (seconds) |
 | `jail_duration_double_sign` | [int64](#int64) |  | JailDurationDoubleSign is the jail duration for double signing (seconds) |
 | `jail_duration_invalid_attestation` | [int64](#int64) |  | JailDurationInvalidAttestation is the jail duration for invalid attestation (seconds) |
 | `score_tolerance` | [int64](#int64) |  | ScoreTolerance is the allowed score difference from consensus (fixed-point) |
 | `max_missed_veid_recomputations` | [int64](#int64) |  | MaxMissedVEIDRecomputations is max missed recomputations before slash |
 | `reward_denom` | [string](#string) |  | RewardDenom is the denomination for rewards |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/staking/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/staking/v1/types.proto
 

 
 <a name="virtengine.staking.v1.DoubleSignEvidence"></a>

 ### DoubleSignEvidence
 DoubleSignEvidence represents evidence of double signing

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `evidence_id` | [string](#string) |  | EvidenceID is the unique identifier |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator who double signed |
 | `height_1` | [int64](#int64) |  | Height1 is the first height of the double sign |
 | `height_2` | [int64](#int64) |  | Height2 is the second height of the double sign |
 | `vote_hash_1` | [string](#string) |  | VoteHash1 is the hash of the first vote |
 | `vote_hash_2` | [string](#string) |  | VoteHash2 is the hash of the second vote |
 | `detected_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | DetectedAt is when the double sign was detected |
 | `detected_height` | [int64](#int64) |  | DetectedHeight is the block height when detected |
 | `processed` | [bool](#bool) |  | Processed indicates if this evidence has been processed |
 
 

 

 
 <a name="virtengine.staking.v1.InvalidVEIDAttestation"></a>

 ### InvalidVEIDAttestation
 InvalidVEIDAttestation represents evidence of invalid VEID attestation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `record_id` | [string](#string) |  | RecordID is the unique identifier |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator with invalid attestation |
 | `attestation_id` | [string](#string) |  | AttestationID is the ID of the invalid attestation |
 | `reason` | [string](#string) |  | Reason is why the attestation is invalid |
 | `expected_score` | [int64](#int64) |  | ExpectedScore is the expected score from consensus |
 | `actual_score` | [int64](#int64) |  | ActualScore is the score the validator reported |
 | `score_difference` | [int64](#int64) |  | ScoreDifference is the difference between expected and actual |
 | `detected_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | DetectedAt is when the issue was detected |
 | `detected_height` | [int64](#int64) |  | DetectedHeight is the block height when detected |
 | `processed` | [bool](#bool) |  | Processed indicates if this evidence has been processed |
 
 

 

 
 <a name="virtengine.staking.v1.RewardEpoch"></a>

 ### RewardEpoch
 RewardEpoch represents a reward epoch

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `epoch_number` | [uint64](#uint64) |  | EpochNumber is the epoch identifier |
 | `start_height` | [int64](#int64) |  | StartHeight is the starting block height |
 | `end_height` | [int64](#int64) |  | EndHeight is the ending block height |
 | `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | StartTime is when the epoch started |
 | `end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | EndTime is when the epoch ended (zero if current) |
 | `total_rewards_distributed` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | TotalRewardsDistributed is the total rewards distributed |
 | `block_proposal_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | BlockProposalRewards is rewards from block proposals |
 | `veid_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | VEIDRewards is rewards from VEID verification work |
 | `uptime_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | UptimeRewards is rewards from uptime |
 | `validator_count` | [int64](#int64) |  | ValidatorCount is the number of validators in this epoch |
 | `total_stake` | [string](#string) |  | TotalStake is the total stake in this epoch |
 | `finalized` | [bool](#bool) |  | Finalized indicates if this epoch is finalized |
 
 

 

 
 <a name="virtengine.staking.v1.SlashConfig"></a>

 ### SlashConfig
 SlashConfig defines the slashing configuration for a reason

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reason` | [SlashReason](#virtengine.staking.v1.SlashReason) |  | Reason is the slash reason |
 | `slash_percent` | [int64](#int64) |  | SlashPercent is the base slash percentage (fixed-point, 1e6 scale) |
 | `jail_duration` | [int64](#int64) |  | JailDuration is the jail duration in seconds |
 | `is_tombstone` | [bool](#bool) |  | IsTombstone indicates if this should tombstone the validator |
 | `escalation_multiplier` | [int64](#int64) |  | EscalationMultiplier is the multiplier for repeat offenses |
 
 

 

 
 <a name="virtengine.staking.v1.SlashRecord"></a>

 ### SlashRecord
 SlashRecord represents a slashing record

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `slash_id` | [string](#string) |  | SlashID is the unique identifier for this slashing event |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator being slashed |
 | `reason` | [SlashReason](#virtengine.staking.v1.SlashReason) |  | Reason is the reason for slashing |
 | `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Amount is the amount slashed |
 | `slash_percent` | [int64](#int64) |  | SlashPercent is the percentage slashed (fixed-point, 1e6 scale) |
 | `infraction_height` | [int64](#int64) |  | InfractionHeight is the block height of the infraction |
 | `slash_height` | [int64](#int64) |  | SlashHeight is the block height when slash was executed |
 | `slash_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | SlashTime is when the slash was executed |
 | `jailed` | [bool](#bool) |  | Jailed indicates if the validator was jailed |
 | `jail_duration` | [int64](#int64) |  | JailDuration is how long the validator is jailed |
 | `jailed_until` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | JailedUntil is when the jail period ends |
 | `tombstoned` | [bool](#bool) |  | Tombstoned indicates if validator is permanently banned |
 | `evidence` | [string](#string) |  | Evidence contains the infraction evidence |
 | `evidence_hash` | [string](#string) |  | EvidenceHash is the hash of the evidence |
 | `reporter_address` | [string](#string) |  | ReporterAddress is who reported the infraction (if any) |
 
 

 

 
 <a name="virtengine.staking.v1.ValidatorPerformance"></a>

 ### ValidatorPerformance
 ValidatorPerformance represents a validator's performance metrics

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator's blockchain address |
 | `blocks_proposed` | [int64](#int64) |  | BlocksProposed is the number of blocks proposed in the current epoch |
 | `blocks_expected` | [int64](#int64) |  | BlocksExpected is the expected number of blocks based on stake weight |
 | `blocks_missed` | [int64](#int64) |  | BlocksMissed is the number of missed blocks (when expected to sign) |
 | `total_signatures` | [int64](#int64) |  | TotalSignatures is the total number of blocks signed |
 | `veid_verifications_completed` | [int64](#int64) |  | VEIDVerificationsCompleted is the number of VEID verifications completed |
 | `veid_verifications_expected` | [int64](#int64) |  | VEIDVerificationsExpected is the expected VEID verifications based on committee selection |
 | `veid_verification_score` | [int64](#int64) |  | VEIDVerificationScore is the quality score for VEID verifications (0-10000) |
 | `uptime_seconds` | [int64](#int64) |  | UptimeSeconds is the total uptime in seconds |
 | `downtime_seconds` | [int64](#int64) |  | DowntimeSeconds is the total downtime in seconds |
 | `consecutive_missed_blocks` | [int64](#int64) |  | ConsecutiveMissedBlocks is the current streak of missed blocks |
 | `last_proposed_height` | [int64](#int64) |  | LastProposedHeight is the last height where this validator proposed a block |
 | `last_signed_height` | [int64](#int64) |  | LastSignedHeight is the last height where this validator signed |
 | `epoch_number` | [uint64](#uint64) |  | EpochNumber is the epoch this performance record belongs to |
 | `updated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | UpdatedAt is when this record was last updated |
 | `overall_score` | [int64](#int64) |  | OverallScore is the computed overall performance score (0-10000) |
 
 

 

 
 <a name="virtengine.staking.v1.ValidatorReward"></a>

 ### ValidatorReward
 ValidatorReward represents a validator's rewards for an epoch

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator's address |
 | `epoch_number` | [uint64](#uint64) |  | EpochNumber is the epoch this reward belongs to |
 | `total_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | TotalReward is the total reward amount |
 | `block_proposal_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | BlockProposalReward is reward from block proposals |
 | `veid_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | VEIDReward is reward from VEID verification |
 | `uptime_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | UptimeReward is reward from uptime |
 | `identity_network_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | IdentityNetworkReward is reward from identity network participation |
 | `performance_score` | [int64](#int64) |  | PerformanceScore is the performance score used for calculation |
 | `stake_weight` | [string](#string) |  | StakeWeight is the stake weight used for calculation |
 | `calculated_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | CalculatedAt is when the reward was calculated |
 | `block_height` | [int64](#int64) |  | BlockHeight is when the reward was recorded |
 | `claimed` | [bool](#bool) |  | Claimed indicates if the reward has been claimed |
 | `claimed_at` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | ClaimedAt is when the reward was claimed |
 
 

 

 
 <a name="virtengine.staking.v1.ValidatorSigningInfo"></a>

 ### ValidatorSigningInfo
 ValidatorSigningInfo contains validator signing information for slashing

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator's blockchain address |
 | `start_height` | [int64](#int64) |  | StartHeight is the height at which validator started signing |
 | `index_offset` | [int64](#int64) |  | IndexOffset is the current index offset into the signed blocks window |
 | `jailed_until` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | JailedUntil is the time until which the validator is jailed |
 | `tombstoned` | [bool](#bool) |  | Tombstoned indicates if the validator has been tombstoned (permanently banned) |
 | `missed_blocks_counter` | [int64](#int64) |  | MissedBlocksCounter is the counter for missed blocks in the current window |
 | `infraction_count` | [int64](#int64) |  | InfractionCount is the total number of infractions |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.staking.v1.RewardType"></a>

 ### RewardType
 RewardType indicates the type of reward

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | REWARD_TYPE_UNSPECIFIED | 0 | REWARD_TYPE_UNSPECIFIED is the default/invalid reward type |
 | REWARD_TYPE_BLOCK_PROPOSAL | 1 | REWARD_TYPE_BLOCK_PROPOSAL is for block proposal rewards |
 | REWARD_TYPE_VEID_VERIFICATION | 2 | REWARD_TYPE_VEID_VERIFICATION is for VEID verification rewards |
 | REWARD_TYPE_UPTIME | 3 | REWARD_TYPE_UPTIME is for uptime-based rewards |
 | REWARD_TYPE_IDENTITY_NETWORK | 4 | REWARD_TYPE_IDENTITY_NETWORK is for identity network participation rewards |
 | REWARD_TYPE_STAKING | 5 | REWARD_TYPE_STAKING is for base staking rewards |
 

 
 <a name="virtengine.staking.v1.SlashReason"></a>

 ### SlashReason
 SlashReason indicates the reason for slashing

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | SLASH_REASON_UNSPECIFIED | 0 | SLASH_REASON_UNSPECIFIED is the default/invalid reason |
 | SLASH_REASON_DOUBLE_SIGNING | 1 | SLASH_REASON_DOUBLE_SIGNING is for double signing infractions |
 | SLASH_REASON_DOWNTIME | 2 | SLASH_REASON_DOWNTIME is for excessive downtime |
 | SLASH_REASON_INVALID_VEID_ATTESTATION | 3 | SLASH_REASON_INVALID_VEID_ATTESTATION is for invalid VEID attestations |
 | SLASH_REASON_MISSED_RECOMPUTATION | 4 | SLASH_REASON_MISSED_RECOMPUTATION is for missing VEID recomputation |
 | SLASH_REASON_INCONSISTENT_SCORE | 5 | SLASH_REASON_INCONSISTENT_SCORE is for scores differing from consensus |
 | SLASH_REASON_EXPIRED_ATTESTATION | 6 | SLASH_REASON_EXPIRED_ATTESTATION is for expired attestation |
 | SLASH_REASON_DEBUG_MODE_ENABLED | 7 | SLASH_REASON_DEBUG_MODE_ENABLED is for enclave debug mode |
 | SLASH_REASON_NON_ALLOWLISTED_MEASUREMENT | 8 | SLASH_REASON_NON_ALLOWLISTED_MEASUREMENT is for non-allowlisted enclave measurement |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/staking/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/staking/v1/genesis.proto
 

 
 <a name="virtengine.staking.v1.GenesisState"></a>

 ### GenesisState
 GenesisState is the genesis state for the staking module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.staking.v1.Params) |  | Params are the module parameters |
 | `validator_performances` | [ValidatorPerformance](#virtengine.staking.v1.ValidatorPerformance) | repeated | ValidatorPerformances are the initial validator performances |
 | `slash_records` | [SlashRecord](#virtengine.staking.v1.SlashRecord) | repeated | SlashRecords are the initial slashing records |
 | `reward_epochs` | [RewardEpoch](#virtengine.staking.v1.RewardEpoch) | repeated | RewardEpochs are the initial reward epochs |
 | `validator_rewards` | [ValidatorReward](#virtengine.staking.v1.ValidatorReward) | repeated | ValidatorRewards are the initial validator rewards |
 | `validator_signing_infos` | [ValidatorSigningInfo](#virtengine.staking.v1.ValidatorSigningInfo) | repeated | ValidatorSigningInfos are the initial signing infos |
 | `current_epoch` | [uint64](#uint64) |  | CurrentEpoch is the current epoch number |
 | `slash_sequence` | [uint64](#uint64) |  | SlashSequence is the next slashing sequence number |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/staking/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/staking/v1/tx.proto
 

 
 <a name="virtengine.staking.v1.MsgRecordPerformance"></a>

 ### MsgRecordPerformance
 MsgRecordPerformance is the message for recording validator performance

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator |
 | `blocks_proposed` | [int64](#int64) |  | BlocksProposed is the number of blocks proposed |
 | `blocks_signed` | [int64](#int64) |  | BlocksSigned is the number of blocks signed |
 | `veid_verifications_completed` | [int64](#int64) |  | VEIDVerificationsCompleted is VEID verifications completed |
 | `veid_verification_score` | [int64](#int64) |  | VEIDVerificationScore is the VEID verification quality score |
 
 

 

 
 <a name="virtengine.staking.v1.MsgRecordPerformanceResponse"></a>

 ### MsgRecordPerformanceResponse
 MsgRecordPerformanceResponse is the response for MsgRecordPerformance

 

 

 
 <a name="virtengine.staking.v1.MsgSlashValidator"></a>

 ### MsgSlashValidator
 MsgSlashValidator is the message for slashing a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator to slash |
 | `reason` | [SlashReason](#virtengine.staking.v1.SlashReason) |  | Reason is the slashing reason |
 | `infraction_height` | [int64](#int64) |  | InfractionHeight is when the infraction occurred |
 | `evidence` | [string](#string) |  | Evidence is the evidence supporting the slash |
 
 

 

 
 <a name="virtengine.staking.v1.MsgSlashValidatorResponse"></a>

 ### MsgSlashValidatorResponse
 MsgSlashValidatorResponse is the response for MsgSlashValidator

 

 

 
 <a name="virtengine.staking.v1.MsgUnjailValidator"></a>

 ### MsgUnjailValidator
 MsgUnjailValidator is the message for unjailing a validator

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator to unjail |
 
 

 

 
 <a name="virtengine.staking.v1.MsgUnjailValidatorResponse"></a>

 ### MsgUnjailValidatorResponse
 MsgUnjailValidatorResponse is the response for MsgUnjailValidator

 

 

 
 <a name="virtengine.staking.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that controls the module (x/gov module account) |
 | `params` | [Params](#virtengine.staking.v1.Params) |  | Params are the new module parameters |
 
 

 

 
 <a name="virtengine.staking.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.staking.v1.Msg"></a>

 ### Msg
 Msg defines the staking Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.staking.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.staking.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 | `SlashValidator` | [MsgSlashValidator](#virtengine.staking.v1.MsgSlashValidator) | [MsgSlashValidatorResponse](#virtengine.staking.v1.MsgSlashValidatorResponse) | SlashValidator slashes a validator for misbehavior | |
 | `UnjailValidator` | [MsgUnjailValidator](#virtengine.staking.v1.MsgUnjailValidator) | [MsgUnjailValidatorResponse](#virtengine.staking.v1.MsgUnjailValidatorResponse) | UnjailValidator unjails a validator | |
 | `RecordPerformance` | [MsgRecordPerformance](#virtengine.staking.v1.MsgRecordPerformance) | [MsgRecordPerformanceResponse](#virtengine.staking.v1.MsgRecordPerformanceResponse) | RecordPerformance records validator performance metrics | |
 
  <!-- end services -->

 
 
 <a name="virtengine/take/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/take/v1/params.proto
 

 
 <a name="virtengine.take.v1.DenomTakeRate"></a>

 ### DenomTakeRate
 DenomTakeRate describes take rate for specified denom.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | Denom is the denomination of the take rate (uve, usdc, etc.). |
 | `rate` | [uint32](#uint32) |  | Rate is the value of the take rate. |
 
 

 

 
 <a name="virtengine.take.v1.Params"></a>

 ### Params
 Params defines the parameters for the x/take package.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom_take_rates` | [DenomTakeRate](#virtengine.take.v1.DenomTakeRate) | repeated | DenomTakeRates is a list of configured take rates. |
 | `default_take_rate` | [uint32](#uint32) |  | DefaultTakeRate holds the default take rate. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/take/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/take/v1/genesis.proto
 

 
 <a name="virtengine.take.v1.GenesisState"></a>

 ### GenesisState
 GenesisState stores slice of genesis staking parameters.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.take.v1.Params) |  | Params holds parameters of the genesis of staking. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/take/v1/paramsmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/take/v1/paramsmsg.proto
 

 
 <a name="virtengine.take.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the Msg/UpdateParams request type.

Since: akash v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.take.v1.Params) |  | params defines the x/deployment parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.take.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: akash v1.0.0

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/take/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/take/v1/query.proto
 

 
 <a name="virtengine.take.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request type for the Query/Params RPC method.

 

 

 
 <a name="virtengine.take.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response type for the Query/Params RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.take.v1.Params) |  | params defines the parameters of the module. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.take.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service of the take package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Params` | [QueryParamsRequest](#virtengine.take.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.take.v1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/virtengine/take/v1/params|
 
  <!-- end services -->

 
 
 <a name="virtengine/take/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/take/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.take.v1.Msg"></a>

 ### Msg
 Msg defines the take Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.take.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.take.v1.MsgUpdateParamsResponse) | UpdateParams defines a governance operation for updating the x/market module parameters. The authority is hard-coded to the x/gov module account.

Since: akash v1.0.0 | |
 
  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/appeal.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/appeal.proto
 

 
 <a name="virtengine.veid.v1.AppealParams"></a>

 ### AppealParams
 AppealParams defines the parameters for the appeal system

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_window_blocks` | [int64](#int64) |  | AppealWindowBlocks is how long after rejection can user appeal |
 | `max_appeals_per_scope` | [uint32](#uint32) |  | MaxAppealsPerScope is the maximum appeals allowed per scope |
 | `min_appeal_reason_length` | [uint32](#uint32) |  | MinAppealReasonLength is the minimum characters for appeal reason |
 | `review_timeout_blocks` | [int64](#int64) |  | ReviewTimeoutBlocks is how long an appeal can stay in reviewing status |
 | `enabled` | [bool](#bool) |  | Enabled indicates whether the appeal system is active |
 | `require_escrow_deposit` | [bool](#bool) |  | RequireEscrowDeposit indicates whether appeals require a deposit |
 | `escrow_deposit_amount` | [int64](#int64) |  | EscrowDepositAmount is the deposit amount required (in base units) |
 
 

 

 
 <a name="virtengine.veid.v1.AppealRecord"></a>

 ### AppealRecord
 AppealRecord tracks an appeal against a verification decision

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_id` | [string](#string) |  | AppealID is the unique identifier for this appeal |
 | `account_address` | [string](#string) |  | AccountAddress is the address of the account filing the appeal |
 | `scope_id` | [string](#string) |  | ScopeID is the scope whose verification decision is being appealed |
 | `original_status` | [string](#string) |  | OriginalStatus is the verification status that prompted the appeal |
 | `original_score` | [uint32](#uint32) |  | OriginalScore is the verification score at the time of appeal |
 | `appeal_reason` | [string](#string) |  | AppealReason is the user's explanation for why they are appealing |
 | `evidence_hashes` | [string](#string) | repeated | EvidenceHashes are hashes of supporting evidence documents |
 | `submitted_at` | [int64](#int64) |  | SubmittedAt is the block height when the appeal was submitted |
 | `submitted_at_time` | [int64](#int64) |  | SubmittedAtTime is the Unix timestamp when the appeal was submitted |
 | `status` | [AppealStatus](#virtengine.veid.v1.AppealStatus) |  | Status is the current status of the appeal |
 | `reviewer_address` | [string](#string) |  | ReviewerAddress is the address of the arbitrator reviewing the appeal |
 | `claimed_at` | [int64](#int64) |  | ClaimedAt is the block height when the appeal was claimed for review |
 | `reviewed_at` | [int64](#int64) |  | ReviewedAt is the block height when the appeal was resolved |
 | `reviewed_at_time` | [int64](#int64) |  | ReviewedAtTime is the Unix timestamp when the appeal was resolved |
 | `resolution_reason` | [string](#string) |  | ResolutionReason is the arbitrator's explanation for the decision |
 | `score_adjustment` | [int32](#int32) |  | ScoreAdjustment is the adjustment to the verification score |
 | `appeal_number` | [uint32](#uint32) |  | AppealNumber tracks which appeal this is for the scope (1st, 2nd, 3rd) |
 
 

 

 
 <a name="virtengine.veid.v1.AppealSummary"></a>

 ### AppealSummary
 AppealSummary provides a summary of appeals for an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `total_appeals` | [uint32](#uint32) |  | TotalAppeals is the total number of appeals |
 | `pending_appeals` | [uint32](#uint32) |  | PendingAppeals is the number of pending appeals |
 | `approved_appeals` | [uint32](#uint32) |  | ApprovedAppeals is the number of approved appeals |
 | `rejected_appeals` | [uint32](#uint32) |  | RejectedAppeals is the number of rejected appeals |
 | `withdrawn_appeals` | [uint32](#uint32) |  | WithdrawnAppeals is the number of withdrawn appeals |
 
 

 

 
 <a name="virtengine.veid.v1.MsgClaimAppeal"></a>

 ### MsgClaimAppeal
 MsgClaimAppeal is the message for an arbitrator to claim an appeal for review

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `reviewer` | [string](#string) |  | Reviewer is the arbitrator claiming the appeal |
 | `appeal_id` | [string](#string) |  | AppealID is the appeal being claimed |
 
 

 

 
 <a name="virtengine.veid.v1.MsgClaimAppealResponse"></a>

 ### MsgClaimAppealResponse
 MsgClaimAppealResponse is the response for MsgClaimAppeal

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_id` | [string](#string) |  | AppealID is the claimed appeal ID |
 | `claimed_at` | [int64](#int64) |  | ClaimedAt is the block height when claimed |
 
 

 

 
 <a name="virtengine.veid.v1.MsgResolveAppeal"></a>

 ### MsgResolveAppeal
 MsgResolveAppeal is the message to resolve an appeal

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `resolver` | [string](#string) |  | Resolver is the arbitrator resolving the appeal |
 | `appeal_id` | [string](#string) |  | AppealID is the appeal being resolved |
 | `resolution` | [AppealStatus](#virtengine.veid.v1.AppealStatus) |  | Resolution is the resolution status (approved or rejected) |
 | `reason` | [string](#string) |  | Reason is the explanation for the resolution decision |
 | `score_adjustment` | [int32](#int32) |  | ScoreAdjustment is the adjustment to the verification score |
 
 

 

 
 <a name="virtengine.veid.v1.MsgResolveAppealResponse"></a>

 ### MsgResolveAppealResponse
 MsgResolveAppealResponse is the response for MsgResolveAppeal

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_id` | [string](#string) |  | AppealID is the resolved appeal ID |
 | `resolution` | [AppealStatus](#virtengine.veid.v1.AppealStatus) |  | Resolution is the final resolution status |
 | `resolved_at` | [int64](#int64) |  | ResolvedAt is the block height when resolved |
 
 

 

 
 <a name="virtengine.veid.v1.MsgSubmitAppeal"></a>

 ### MsgSubmitAppeal
 MsgSubmitAppeal is the message to submit an appeal against a verification decision

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `submitter` | [string](#string) |  | Submitter is the account address submitting the appeal (must own the scope) |
 | `scope_id` | [string](#string) |  | ScopeID is the scope whose verification decision is being appealed |
 | `reason` | [string](#string) |  | Reason is the explanation for why the submitter is appealing |
 | `evidence_hashes` | [string](#string) | repeated | EvidenceHashes are hashes of supporting evidence documents |
 
 

 

 
 <a name="virtengine.veid.v1.MsgSubmitAppealResponse"></a>

 ### MsgSubmitAppealResponse
 MsgSubmitAppealResponse is the response for MsgSubmitAppeal

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_id` | [string](#string) |  | AppealID is the unique identifier for the created appeal |
 | `status` | [AppealStatus](#virtengine.veid.v1.AppealStatus) |  | Status is the initial status of the appeal |
 | `appeal_number` | [uint32](#uint32) |  | AppealNumber is which appeal this is for the scope |
 | `submitted_at` | [int64](#int64) |  | SubmittedAt is the block height when submitted |
 
 

 

 
 <a name="virtengine.veid.v1.MsgWithdrawAppeal"></a>

 ### MsgWithdrawAppeal
 MsgWithdrawAppeal allows the submitter to withdraw their appeal

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `submitter` | [string](#string) |  | Submitter is the original appeal submitter |
 | `appeal_id` | [string](#string) |  | AppealID is the appeal to withdraw |
 
 

 

 
 <a name="virtengine.veid.v1.MsgWithdrawAppealResponse"></a>

 ### MsgWithdrawAppealResponse
 MsgWithdrawAppealResponse is the response for MsgWithdrawAppeal

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_id` | [string](#string) |  | AppealID is the withdrawn appeal ID |
 | `withdrawn_at` | [int64](#int64) |  | WithdrawnAt is the block height when withdrawn |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.veid.v1.AppealStatus"></a>

 ### AppealStatus
 AppealStatus represents the current state of an appeal

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | APPEAL_STATUS_UNSPECIFIED | 0 | APPEAL_STATUS_UNSPECIFIED is the default unspecified status |
 | APPEAL_STATUS_PENDING | 1 | APPEAL_STATUS_PENDING indicates the appeal has been submitted and awaits review |
 | APPEAL_STATUS_REVIEWING | 2 | APPEAL_STATUS_REVIEWING indicates an arbitrator has claimed the appeal for review |
 | APPEAL_STATUS_APPROVED | 3 | APPEAL_STATUS_APPROVED indicates the appeal was approved |
 | APPEAL_STATUS_REJECTED | 4 | APPEAL_STATUS_REJECTED indicates the appeal was rejected |
 | APPEAL_STATUS_WITHDRAWN | 5 | APPEAL_STATUS_WITHDRAWN indicates the submitter withdrew their appeal |
 | APPEAL_STATUS_EXPIRED | 6 | APPEAL_STATUS_EXPIRED indicates the appeal expired without resolution |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/compliance.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/compliance.proto
 

 
 <a name="virtengine.veid.v1.ComplianceAttestation"></a>

 ### ComplianceAttestation
 ComplianceAttestation is a validator attestation of compliance status

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the address of the attesting validator |
 | `attested_at` | [int64](#int64) |  | AttestedAt is the Unix timestamp when attestation was made |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is the Unix timestamp when this attestation expires |
 | `attestation_type` | [string](#string) |  | AttestationType describes what is being attested |
 | `attestation_hash` | [string](#string) |  | AttestationHash is a hash of the attestation data for verification |
 
 

 

 
 <a name="virtengine.veid.v1.ComplianceCheckResult"></a>

 ### ComplianceCheckResult
 ComplianceCheckResult stores result of a single compliance check

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `check_type` | [ComplianceCheckType](#virtengine.veid.v1.ComplianceCheckType) |  | CheckType indicates what kind of check was performed |
 | `passed` | [bool](#bool) |  | Passed indicates whether the check passed |
 | `details` | [string](#string) |  | Details provides additional context about the check result |
 | `match_score` | [int32](#int32) |  | MatchScore indicates confidence of any matches found (0-100) |
 | `checked_at` | [int64](#int64) |  | CheckedAt is the Unix timestamp when the check was performed |
 | `provider_id` | [string](#string) |  | ProviderID identifies the compliance provider that performed the check |
 | `reference_id` | [string](#string) |  | ReferenceID is the provider's reference for this check |
 
 

 

 
 <a name="virtengine.veid.v1.ComplianceParams"></a>

 ### ComplianceParams
 ComplianceParams configures the compliance module behavior

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `require_sanction_check` | [bool](#bool) |  | RequireSanctionCheck indicates if sanction list check is mandatory |
 | `require_pep_check` | [bool](#bool) |  | RequirePEPCheck indicates if PEP check is mandatory |
 | `check_expiry_blocks` | [int64](#int64) |  | CheckExpiryBlocks is how long compliance checks remain valid |
 | `risk_score_threshold` | [int32](#int32) |  | RiskScoreThreshold is the maximum allowed risk score (0-100) |
 | `restricted_countries` | [string](#string) | repeated | RestrictedCountries is list of ISO country codes that are restricted |
 | `min_attestations_required` | [int32](#int32) |  | MinAttestationsRequired is minimum validator attestations needed |
 | `enable_auto_expiry` | [bool](#bool) |  | EnableAutoExpiry enables automatic expiration of compliance records |
 | `require_document_verification` | [bool](#bool) |  | RequireDocumentVerification indicates if document verification is mandatory |
 
 

 

 
 <a name="virtengine.veid.v1.ComplianceProvider"></a>

 ### ComplianceProvider
 ComplianceProvider represents an authorized external compliance provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_id` | [string](#string) |  | ProviderID is the unique identifier for this provider |
 | `name` | [string](#string) |  | Name is the human-readable name of the provider |
 | `provider_address` | [string](#string) |  | ProviderAddress is the blockchain address authorized to submit checks |
 | `supported_check_types` | [ComplianceCheckType](#virtengine.veid.v1.ComplianceCheckType) | repeated | SupportedCheckTypes lists which check types this provider can perform |
 | `is_active` | [bool](#bool) |  | IsActive indicates if this provider is currently active |
 | `registered_at` | [int64](#int64) |  | RegisteredAt is when this provider was registered |
 | `last_active_at` | [int64](#int64) |  | LastActiveAt is when this provider last submitted a check |
 
 

 

 
 <a name="virtengine.veid.v1.ComplianceRecord"></a>

 ### ComplianceRecord
 ComplianceRecord stores the complete compliance status for an identity

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the blockchain address of the identity |
 | `status` | [ComplianceStatus](#virtengine.veid.v1.ComplianceStatus) |  | Status is the overall compliance status |
 | `check_results` | [ComplianceCheckResult](#virtengine.veid.v1.ComplianceCheckResult) | repeated | CheckResults contains results of individual compliance checks |
 | `last_checked_at` | [int64](#int64) |  | LastCheckedAt is the Unix timestamp of the last compliance check |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is the Unix timestamp when the compliance record expires |
 | `risk_score` | [int32](#int32) |  | RiskScore is the overall risk score (0-100, lower is better) |
 | `restricted_regions` | [string](#string) | repeated | RestrictedRegions lists regions where this identity is restricted |
 | `attestations` | [ComplianceAttestation](#virtengine.veid.v1.ComplianceAttestation) | repeated | Attestations contains validator attestations of compliance |
 | `created_at` | [int64](#int64) |  | CreatedAt is when this record was first created |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when this record was last updated |
 | `notes` | [string](#string) |  | Notes contains any additional notes about the compliance status |
 
 

 

 
 <a name="virtengine.veid.v1.MsgAttestCompliance"></a>

 ### MsgAttestCompliance
 MsgAttestCompliance allows validators to attest compliance status

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the address of the attesting validator |
 | `target_address` | [string](#string) |  | TargetAddress is the address being attested |
 | `attestation_type` | [string](#string) |  | AttestationType describes what is being attested |
 | `expiry_blocks` | [int64](#int64) |  | ExpiryBlocks is how long until this attestation expires (in blocks) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgAttestComplianceResponse"></a>

 ### MsgAttestComplianceResponse
 MsgAttestComplianceResponse is the response for MsgAttestCompliance

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `attested_at` | [int64](#int64) |  | AttestedAt is the block height when attested |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the attestation expires |
 
 

 

 
 <a name="virtengine.veid.v1.MsgDeactivateComplianceProvider"></a>

 ### MsgDeactivateComplianceProvider
 MsgDeactivateComplianceProvider deactivates a compliance provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that is authorized to deactivate providers (x/gov) |
 | `provider_id` | [string](#string) |  | ProviderID is the ID of the provider to deactivate |
 | `reason` | [string](#string) |  | Reason is the reason for deactivation |
 
 

 

 
 <a name="virtengine.veid.v1.MsgDeactivateComplianceProviderResponse"></a>

 ### MsgDeactivateComplianceProviderResponse
 MsgDeactivateComplianceProviderResponse is the response for MsgDeactivateComplianceProvider

 

 

 
 <a name="virtengine.veid.v1.MsgRegisterComplianceProvider"></a>

 ### MsgRegisterComplianceProvider
 MsgRegisterComplianceProvider registers a new compliance provider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that is authorized to register providers (x/gov) |
 | `provider` | [ComplianceProvider](#virtengine.veid.v1.ComplianceProvider) |  | Provider is the compliance provider to register |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRegisterComplianceProviderResponse"></a>

 ### MsgRegisterComplianceProviderResponse
 MsgRegisterComplianceProviderResponse is the response for MsgRegisterComplianceProvider

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_id` | [string](#string) |  | ProviderID is the registered provider's ID |
 
 

 

 
 <a name="virtengine.veid.v1.MsgSubmitComplianceCheck"></a>

 ### MsgSubmitComplianceCheck
 MsgSubmitComplianceCheck submits external compliance check results

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_address` | [string](#string) |  | ProviderAddress is the address of the compliance provider |
 | `target_address` | [string](#string) |  | TargetAddress is the address being checked |
 | `check_results` | [ComplianceCheckResult](#virtengine.veid.v1.ComplianceCheckResult) | repeated | CheckResults contains the compliance check results |
 | `provider_id` | [string](#string) |  | ProviderID is the ID of the compliance provider |
 
 

 

 
 <a name="virtengine.veid.v1.MsgSubmitComplianceCheckResponse"></a>

 ### MsgSubmitComplianceCheckResponse
 MsgSubmitComplianceCheckResponse is the response for MsgSubmitComplianceCheck

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `status` | [ComplianceStatus](#virtengine.veid.v1.ComplianceStatus) |  | Status is the updated compliance status |
 | `risk_score` | [int32](#int32) |  | RiskScore is the updated risk score |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateComplianceParams"></a>

 ### MsgUpdateComplianceParams
 MsgUpdateComplianceParams updates compliance configuration (gov only)

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address that is authorized to update params (x/gov) |
 | `params` | [ComplianceParams](#virtengine.veid.v1.ComplianceParams) |  | Params are the new compliance parameters |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateComplianceParamsResponse"></a>

 ### MsgUpdateComplianceParamsResponse
 MsgUpdateComplianceParamsResponse is the response for MsgUpdateComplianceParams

 

 

  <!-- end messages -->

 
 <a name="virtengine.veid.v1.ComplianceCheckType"></a>

 ### ComplianceCheckType
 ComplianceCheckType defines what type of compliance check

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | COMPLIANCE_CHECK_SANCTION_LIST | 0 | COMPLIANCE_CHECK_SANCTION_LIST checks against global sanction lists |
 | COMPLIANCE_CHECK_PEP | 1 | COMPLIANCE_CHECK_PEP checks for Politically Exposed Person status |
 | COMPLIANCE_CHECK_ADVERSE_MEDIA | 2 | COMPLIANCE_CHECK_ADVERSE_MEDIA checks for negative news coverage |
 | COMPLIANCE_CHECK_GEOGRAPHIC | 3 | COMPLIANCE_CHECK_GEOGRAPHIC checks geographic restrictions |
 | COMPLIANCE_CHECK_WATCHLIST | 4 | COMPLIANCE_CHECK_WATCHLIST checks against custom watchlists |
 | COMPLIANCE_CHECK_DOCUMENT_VERIFICATION | 5 | COMPLIANCE_CHECK_DOCUMENT_VERIFICATION verifies identity documents |
 | COMPLIANCE_CHECK_AML_RISK | 6 | COMPLIANCE_CHECK_AML_RISK assesses anti-money laundering risk |
 

 
 <a name="virtengine.veid.v1.ComplianceStatus"></a>

 ### ComplianceStatus
 ComplianceStatus represents the compliance state of an identity

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | COMPLIANCE_STATUS_UNKNOWN | 0 | COMPLIANCE_STATUS_UNKNOWN indicates no compliance check has been performed |
 | COMPLIANCE_STATUS_PENDING | 1 | COMPLIANCE_STATUS_PENDING indicates compliance check is in progress |
 | COMPLIANCE_STATUS_CLEARED | 2 | COMPLIANCE_STATUS_CLEARED indicates identity passed all compliance checks |
 | COMPLIANCE_STATUS_FLAGGED | 3 | COMPLIANCE_STATUS_FLAGGED indicates identity has been flagged for review |
 | COMPLIANCE_STATUS_BLOCKED | 4 | COMPLIANCE_STATUS_BLOCKED indicates identity is blocked from transactions |
 | COMPLIANCE_STATUS_EXPIRED | 5 | COMPLIANCE_STATUS_EXPIRED indicates compliance check has expired |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/types.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/types.proto
 

 
 <a name="virtengine.veid.v1.ApprovedClient"></a>

 ### ApprovedClient
 ApprovedClient represents an approved client application

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `client_id` | [string](#string) |  | ClientID is the unique identifier for the client |
 | `name` | [string](#string) |  | Name is the human-readable name of the client |
 | `public_key` | [bytes](#bytes) |  | PublicKey is the client's public key for signature verification |
 | `active` | [bool](#bool) |  | Active indicates if the client is currently active |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the client was registered (Unix timestamp) |
 | `deactivated_at` | [int64](#int64) |  | DeactivatedAt is when the client was deactivated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.BorderlineParams"></a>

 ### BorderlineParams
 BorderlineParams contains parameters for borderline score handling

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `lower_threshold` | [uint32](#uint32) |  | LowerThreshold is the lower threshold for borderline scores |
 | `upper_threshold` | [uint32](#uint32) |  | UpperThreshold is the upper threshold for borderline scores |
 | `mfa_timeout_blocks` | [int64](#int64) |  | MfaTimeoutBlocks is how long MFA challenge is valid |
 | `required_factors` | [uint32](#uint32) |  | RequiredFactors is the number of MFA factors required |
 
 

 

 
 <a name="virtengine.veid.v1.ConsentSettings"></a>

 ### ConsentSettings
 ConsentSettings represents consent configuration for an identity wallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `share_with_providers` | [bool](#bool) |  | ShareWithProviders allows providers to access non-sensitive identity metadata |
 | `share_for_verification` | [bool](#bool) |  | ShareForVerification allows the identity to be used for verification requests |
 | `allow_re_verification` | [bool](#bool) |  | AllowReVerification allows the identity to be re-verified without explicit request |
 | `allow_derived_feature_sharing` | [bool](#bool) |  | AllowDerivedFeatureSharing allows sharing of derived feature hashes |
 | `consent_version` | [uint32](#uint32) |  | ConsentVersion tracks consent settings version for audit |
 | `last_updated_at` | [int64](#int64) |  | LastUpdatedAt is when consent was last updated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.DerivedFeatures"></a>

 ### DerivedFeatures
 DerivedFeatures contains hashes of derived features from identity scopes.
These are used for verification matching without revealing the underlying data.
All hashes are SHA-256 (32 bytes) for consistency.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `face_embedding_hash` | [bytes](#bytes) |  | FaceEmbeddingHash is the SHA-256 hash of the face embedding vector. The actual embedding is stored encrypted; this hash allows matching. |
 | `doc_field_hashes` | [DerivedFeatures.DocFieldHashesEntry](#virtengine.veid.v1.DerivedFeatures.DocFieldHashesEntry) | repeated | DocFieldHashes contains hashes of extracted document fields. Keys are field names: "name_hash", "dob_hash", "doc_number_hash", etc. |
 | `biometric_hash` | [bytes](#bytes) |  | BiometricHash is the hash of biometric data (fingerprint, voice, etc.) |
 | `liveness_proof_hash` | [bytes](#bytes) |  | LivenessProofHash is the hash of liveness detection proof |
 | `last_computed_at` | [int64](#int64) |  | LastComputedAt is when these features were last computed (Unix timestamp) |
 | `model_version` | [string](#string) |  | ModelVersion is the ML model version used to compute these features |
 | `computed_by` | [string](#string) |  | ComputedBy is the validator address that computed these features |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height when features were computed |
 | `feature_version` | [uint32](#uint32) |  | FeatureVersion tracks the derived features schema version |
 
 

 

 
 <a name="virtengine.veid.v1.DerivedFeatures.DocFieldHashesEntry"></a>

 ### DerivedFeatures.DocFieldHashesEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.EncryptedPayloadEnvelope"></a>

 ### EncryptedPayloadEnvelope
 EncryptedPayloadEnvelope is the canonical encrypted payload structure
for all sensitive fields stored on-chain.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `version` | [uint32](#uint32) |  | Version is the envelope format version for future compatibility |
 | `algorithm_id` | [string](#string) |  | AlgorithmID identifies the encryption algorithm used |
 | `algorithm_version` | [uint32](#uint32) |  | AlgorithmVersion is the version of the algorithm used |
 | `recipient_key_ids` | [string](#string) | repeated | RecipientKeyIDs are the fingerprints of intended recipients' public keys |
 | `recipient_public_keys` | [bytes](#bytes) | repeated | RecipientPublicKeys are the public keys for intended recipients |
 | `encrypted_keys` | [bytes](#bytes) | repeated | EncryptedKeys contains the data encryption key encrypted for each recipient |
 | `nonce` | [bytes](#bytes) |  | Nonce is the initialization vector / nonce for encryption |
 | `ciphertext` | [bytes](#bytes) |  | Ciphertext is the encrypted payload data |
 | `sender_signature` | [bytes](#bytes) |  | SenderSignature is the signature over the envelope contents |
 | `sender_pub_key` | [bytes](#bytes) |  | SenderPubKey is the sender's public key for signature verification |
 | `metadata` | [EncryptedPayloadEnvelope.MetadataEntry](#virtengine.veid.v1.EncryptedPayloadEnvelope.MetadataEntry) | repeated | Metadata contains optional public or encrypted metadata |
 
 

 

 
 <a name="virtengine.veid.v1.EncryptedPayloadEnvelope.MetadataEntry"></a>

 ### EncryptedPayloadEnvelope.MetadataEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.GlobalConsentUpdate"></a>

 ### GlobalConsentUpdate
 GlobalConsentUpdate represents an update to global consent settings

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `share_with_providers` | [bool](#bool) |  | ShareWithProviders update |
 | `share_for_verification` | [bool](#bool) |  | ShareForVerification update |
 | `allow_re_verification` | [bool](#bool) |  | AllowReVerification update |
 | `allow_derived_feature_sharing` | [bool](#bool) |  | AllowDerivedFeatureSharing update |
 
 

 

 
 <a name="virtengine.veid.v1.IdentityRecord"></a>

 ### IdentityRecord
 IdentityRecord represents a user's complete identity record on-chain

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the blockchain address that owns this identity |
 | `scope_refs` | [ScopeRef](#virtengine.veid.v1.ScopeRef) | repeated | ScopeRefs are lightweight references to the scopes owned by this identity |
 | `current_score` | [uint32](#uint32) |  | CurrentScore is the current identity score (0-100) |
 | `score_version` | [string](#string) |  | ScoreVersion is the ML model version used to compute the current score |
 | `last_verified_at` | [int64](#int64) |  | LastVerifiedAt is when the identity was last verified (Unix timestamp) |
 | `created_at` | [int64](#int64) |  | CreatedAt is when this identity record was created (Unix timestamp) |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when this identity record was last updated (Unix timestamp) |
 | `tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | Tier is the current identity tier based on score |
 | `flags` | [string](#string) | repeated | Flags contains any flags on this identity |
 | `locked` | [bool](#bool) |  | Locked indicates if the identity is locked |
 | `locked_reason` | [string](#string) |  | LockedReason is the reason for locking |
 
 

 

 
 <a name="virtengine.veid.v1.IdentityScope"></a>

 ### IdentityScope
 IdentityScope represents a single piece of identity information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier for this scope |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType indicates what kind of identity data this scope contains |
 | `version` | [uint32](#uint32) |  | Version is the schema version for this scope |
 | `encrypted_payload` | [EncryptedPayloadEnvelope](#virtengine.veid.v1.EncryptedPayloadEnvelope) |  | EncryptedPayload contains the encrypted identity data |
 | `upload_metadata` | [UploadMetadata](#virtengine.veid.v1.UploadMetadata) |  | UploadMetadata contains metadata about the upload |
 | `status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | Status is the current verification status |
 | `uploaded_at` | [int64](#int64) |  | UploadedAt is when this scope was uploaded (Unix timestamp) |
 | `verified_at` | [int64](#int64) |  | VerifiedAt is when this scope was verified (Unix timestamp) |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the verification expires (Unix timestamp) |
 | `owner_address` | [string](#string) |  | OwnerAddress is the blockchain address that owns this scope |
 
 

 

 
 <a name="virtengine.veid.v1.IdentityScore"></a>

 ### IdentityScore
 IdentityScore represents the current identity score for an account

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the blockchain address this score belongs to |
 | `score` | [uint32](#uint32) |  | Score is the current identity score (0-100) |
 | `status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | Status is the current account verification status |
 | `tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | Tier is the identity tier based on score |
 | `model_version` | [string](#string) |  | ModelVersion is the ML model version used |
 | `last_updated_at` | [int64](#int64) |  | LastUpdatedAt is when the score was last updated (Unix timestamp) |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height when score was computed |
 
 

 

 
 <a name="virtengine.veid.v1.IdentityWallet"></a>

 ### IdentityWallet
 IdentityWallet represents a user-controlled identity container.
This is the first-class on-chain identity primitive that references
encrypted scopes and derived features, bound to the user's account key(s).

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `wallet_id` | [string](#string) |  | WalletID is the unique identifier for this wallet. Typically derived from the account address. |
 | `account_address` | [string](#string) |  | AccountAddress is the blockchain address bound to this wallet |
 | `created_at` | [int64](#int64) |  | CreatedAt is when this wallet was created (Unix timestamp) |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when this wallet was last updated (Unix timestamp) |
 | `status` | [WalletStatus](#virtengine.veid.v1.WalletStatus) |  | Status is the current status of the wallet |
 | `scope_refs` | [ScopeReference](#virtengine.veid.v1.ScopeReference) | repeated | ScopeRefs are references to encrypted scope envelopes. These are opaque references - the actual encrypted data is stored separately. |
 | `derived_features` | [DerivedFeatures](#virtengine.veid.v1.DerivedFeatures) |  | DerivedFeatures contains hashes of derived features for verification matching |
 | `current_score` | [uint32](#uint32) |  | CurrentScore is the current identity verification score (0-100) |
 | `score_status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | ScoreStatus is the current verification status |
 | `verification_history` | [VerificationHistoryEntry](#virtengine.veid.v1.VerificationHistoryEntry) | repeated | VerificationHistory contains the history of verification events |
 | `consent_settings` | [ConsentSettings](#virtengine.veid.v1.ConsentSettings) |  | ConsentSettings contains the consent configuration for this wallet |
 | `scope_consents` | [IdentityWallet.ScopeConsentsEntry](#virtengine.veid.v1.IdentityWallet.ScopeConsentsEntry) | repeated | ScopeConsents contains per-scope consent settings (key is scopeID) |
 | `binding_signature` | [bytes](#bytes) |  | BindingSignature is the user's signature over (WalletID + AccountAddress). This cryptographically binds the wallet to the user's account. |
 | `binding_pub_key` | [bytes](#bytes) |  | BindingPubKey is the public key used to create the binding signature. This is captured at wallet creation and updated on key rotation. |
 | `last_binding_at` | [int64](#int64) |  | LastBindingAt is when the wallet was last bound/rebound (Unix timestamp) |
 | `tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | Tier is the current identity tier based on score |
 | `metadata` | [IdentityWallet.MetadataEntry](#virtengine.veid.v1.IdentityWallet.MetadataEntry) | repeated | Metadata contains additional wallet metadata |
 
 

 

 
 <a name="virtengine.veid.v1.IdentityWallet.MetadataEntry"></a>

 ### IdentityWallet.MetadataEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.IdentityWallet.ScopeConsentsEntry"></a>

 ### IdentityWallet.ScopeConsentsEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [ScopeConsent](#virtengine.veid.v1.ScopeConsent) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.Params"></a>

 ### Params
 Params defines the parameters for the veid module

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `max_scopes_per_account` | [uint32](#uint32) |  | MaxScopesPerAccount is the maximum number of scopes per account |
 | `max_scopes_per_type` | [uint32](#uint32) |  | MaxScopesPerType is the maximum number of scopes per type per account |
 | `salt_min_bytes` | [uint32](#uint32) |  | SaltMinBytes is the minimum salt size in bytes |
 | `salt_max_bytes` | [uint32](#uint32) |  | SaltMaxBytes is the maximum salt size in bytes |
 | `require_client_signature` | [bool](#bool) |  | RequireClientSignature determines if client signatures are mandatory |
 | `require_user_signature` | [bool](#bool) |  | RequireUserSignature determines if user signatures are mandatory |
 | `verification_expiry_days` | [uint32](#uint32) |  | VerificationExpiryDays is how long a verification is valid (in days) |
 
 

 

 
 <a name="virtengine.veid.v1.ScopeConsent"></a>

 ### ScopeConsent
 ScopeConsent represents consent configuration for a specific scope.
This tracks per-scope consent settings within a wallet.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the identifier of the scope this consent applies to |
 | `granted` | [bool](#bool) |  | Granted indicates if consent is currently granted |
 | `granted_at` | [int64](#int64) |  | GrantedAt is when consent was granted (Unix timestamp, 0 if never granted) |
 | `revoked_at` | [int64](#int64) |  | RevokedAt is when consent was revoked (Unix timestamp, 0 if not revoked) |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when this consent expires (Unix timestamp, 0 for no expiration) |
 | `purpose` | [string](#string) |  | Purpose describes the purpose for which consent was granted |
 | `granted_to_providers` | [string](#string) | repeated | GrantedToProviders lists specific providers consent was granted to. Empty means consent is general (not provider-specific). |
 | `restrictions` | [string](#string) | repeated | Restrictions contains any restrictions on this consent |
 
 

 

 
 <a name="virtengine.veid.v1.ScopeRef"></a>

 ### ScopeRef
 ScopeRef is a lightweight reference to an identity scope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier for the scope |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType indicates what kind of identity data this scope contains |
 | `status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | Status is the current verification status |
 | `uploaded_at` | [int64](#int64) |  | UploadedAt is when the scope was uploaded (Unix timestamp) |
 | `verified_at` | [int64](#int64) |  | VerifiedAt is when the scope was verified (Unix timestamp) |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the verification expires (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.ScopeReference"></a>

 ### ScopeReference
 ScopeReference represents a reference to an encrypted scope within a wallet.
This is a more detailed reference than ScopeRef, containing wallet-specific metadata.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier of the scope |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType indicates what kind of identity data this scope contains |
 | `envelope_hash` | [bytes](#bytes) |  | EnvelopeHash is the SHA-256 hash of the encrypted envelope. This allows verification without exposing the encrypted content. |
 | `added_at` | [int64](#int64) |  | AddedAt is when this scope was added to the wallet (Unix timestamp) |
 | `status` | [ScopeRefStatus](#virtengine.veid.v1.ScopeRefStatus) |  | Status is the current status of this scope reference |
 | `consent_granted` | [bool](#bool) |  | ConsentGranted indicates if consent has been granted for this scope |
 | `revocation_reason` | [string](#string) |  | RevocationReason is the reason for revocation (if revoked) |
 | `revoked_at` | [int64](#int64) |  | RevokedAt is when this scope was revoked (Unix timestamp, 0 if not revoked) |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when this scope reference expires (Unix timestamp, 0 for no expiry) |
 
 

 

 
 <a name="virtengine.veid.v1.UploadMetadata"></a>

 ### UploadMetadata
 UploadMetadata contains metadata about an identity scope upload

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `salt` | [bytes](#bytes) |  | Salt is a per-upload unique salt for cryptographic binding |
 | `salt_hash` | [bytes](#bytes) |  | SaltHash is the SHA256 hash of the salt |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is a hash/identifier of the device used for upload |
 | `client_id` | [string](#string) |  | ClientID is the identifier of the approved client |
 | `client_signature` | [bytes](#bytes) |  | ClientSignature is the cryptographic signature from the approved client |
 | `user_signature` | [bytes](#bytes) |  | UserSignature is the cryptographic signature from the user's account |
 | `payload_hash` | [bytes](#bytes) |  | PayloadHash is the SHA256 hash of the encrypted payload |
 | `upload_nonce` | [bytes](#bytes) |  | UploadNonce is a unique nonce for this upload session |
 | `capture_timestamp` | [int64](#int64) |  | CaptureTimestamp is when the data was captured (Unix timestamp) |
 | `geo_hint` | [string](#string) |  | GeoHint is an optional geographic hint (coarse location) |
 
 

 

 
 <a name="virtengine.veid.v1.VerificationHistoryEntry"></a>

 ### VerificationHistoryEntry
 VerificationHistoryEntry represents a single verification event in the wallet's history.
This tracks score and status changes over time for audit purposes.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `entry_id` | [string](#string) |  | EntryID is a unique identifier for this history entry |
 | `timestamp` | [int64](#int64) |  | Timestamp is when this verification occurred (Unix timestamp) |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height when this was recorded |
 | `previous_score` | [uint32](#uint32) |  | PreviousScore is the score before this verification |
 | `new_score` | [uint32](#uint32) |  | NewScore is the score after this verification |
 | `previous_status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | PreviousStatus is the status before this verification |
 | `new_status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | NewStatus is the status after this verification |
 | `scopes_evaluated` | [string](#string) | repeated | ScopesEvaluated lists the scope IDs that were evaluated |
 | `model_version` | [string](#string) |  | ModelVersion is the ML model version used for this verification |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the address of the validator that performed this verification |
 | `reason` | [string](#string) |  | Reason is an optional reason/description for this verification |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.veid.v1.AccountStatus"></a>

 ### AccountStatus
 AccountStatus represents the overall account verification status

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | ACCOUNT_STATUS_UNKNOWN | 0 | ACCOUNT_STATUS_UNKNOWN indicates an uninitialized status |
 | ACCOUNT_STATUS_PENDING | 1 | ACCOUNT_STATUS_PENDING indicates verification is in progress |
 | ACCOUNT_STATUS_IN_PROGRESS | 2 | ACCOUNT_STATUS_IN_PROGRESS indicates active ML scoring |
 | ACCOUNT_STATUS_VERIFIED | 3 | ACCOUNT_STATUS_VERIFIED indicates account is verified |
 | ACCOUNT_STATUS_REJECTED | 4 | ACCOUNT_STATUS_REJECTED indicates verification was rejected |
 | ACCOUNT_STATUS_EXPIRED | 5 | ACCOUNT_STATUS_EXPIRED indicates verification has expired |
 | ACCOUNT_STATUS_NEEDS_ADDITIONAL_FACTOR | 6 | ACCOUNT_STATUS_NEEDS_ADDITIONAL_FACTOR indicates additional verification needed |
 

 
 <a name="virtengine.veid.v1.IdentityTier"></a>

 ### IdentityTier
 IdentityTier represents the verification tier of an identity

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | IDENTITY_TIER_UNVERIFIED | 0 | IDENTITY_TIER_UNVERIFIED is the initial state with no verification |
 | IDENTITY_TIER_BASIC | 1 | IDENTITY_TIER_BASIC is for minimally verified identities (score 50-69) |
 | IDENTITY_TIER_STANDARD | 2 | IDENTITY_TIER_STANDARD is for standard verified identities (score 70-84) |
 | IDENTITY_TIER_PREMIUM | 3 | IDENTITY_TIER_PREMIUM is for premium verified identities (score 85-100) |
 

 
 <a name="virtengine.veid.v1.ScopeRefStatus"></a>

 ### ScopeRefStatus
 ScopeRefStatus represents the status of a scope reference within a wallet

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | SCOPE_REF_STATUS_UNSPECIFIED | 0 | SCOPE_REF_STATUS_UNSPECIFIED indicates an unspecified status |
 | SCOPE_REF_STATUS_ACTIVE | 1 | SCOPE_REF_STATUS_ACTIVE indicates the scope reference is active |
 | SCOPE_REF_STATUS_REVOKED | 2 | SCOPE_REF_STATUS_REVOKED indicates the scope reference has been revoked |
 | SCOPE_REF_STATUS_EXPIRED | 3 | SCOPE_REF_STATUS_EXPIRED indicates the scope reference has expired |
 | SCOPE_REF_STATUS_PENDING | 4 | SCOPE_REF_STATUS_PENDING indicates the scope is pending verification |
 

 
 <a name="virtengine.veid.v1.ScopeType"></a>

 ### ScopeType
 ScopeType represents the type of identity scope

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | SCOPE_TYPE_UNSPECIFIED | 0 | SCOPE_TYPE_UNSPECIFIED represents an unspecified scope type |
 | SCOPE_TYPE_ID_DOCUMENT | 1 | SCOPE_TYPE_ID_DOCUMENT represents government-issued ID documents |
 | SCOPE_TYPE_SELFIE | 2 | SCOPE_TYPE_SELFIE represents a selfie photo for face verification |
 | SCOPE_TYPE_FACE_VIDEO | 3 | SCOPE_TYPE_FACE_VIDEO represents a video for liveness detection |
 | SCOPE_TYPE_BIOMETRIC | 4 | SCOPE_TYPE_BIOMETRIC represents biometric data (fingerprint, voice, etc.) |
 | SCOPE_TYPE_SSO_METADATA | 5 | SCOPE_TYPE_SSO_METADATA represents SSO provider metadata pointers |
 | SCOPE_TYPE_EMAIL_PROOF | 6 | SCOPE_TYPE_EMAIL_PROOF represents email verification proof |
 | SCOPE_TYPE_SMS_PROOF | 7 | SCOPE_TYPE_SMS_PROOF represents SMS/phone verification proof |
 | SCOPE_TYPE_DOMAIN_VERIFY | 8 | SCOPE_TYPE_DOMAIN_VERIFY represents domain ownership verification |
 | SCOPE_TYPE_AD_SSO | 9 | SCOPE_TYPE_AD_SSO represents Active Directory SSO verification |
 

 
 <a name="virtengine.veid.v1.VerificationStatus"></a>

 ### VerificationStatus
 VerificationStatus represents the verification state of an identity scope

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | VERIFICATION_STATUS_UNKNOWN | 0 | VERIFICATION_STATUS_UNKNOWN indicates an uninitialized or unknown status |
 | VERIFICATION_STATUS_PENDING | 1 | VERIFICATION_STATUS_PENDING indicates the scope is awaiting verification |
 | VERIFICATION_STATUS_IN_PROGRESS | 2 | VERIFICATION_STATUS_IN_PROGRESS indicates verification is actively being processed |
 | VERIFICATION_STATUS_VERIFIED | 3 | VERIFICATION_STATUS_VERIFIED indicates the scope has been successfully verified |
 | VERIFICATION_STATUS_REJECTED | 4 | VERIFICATION_STATUS_REJECTED indicates the scope failed verification |
 | VERIFICATION_STATUS_EXPIRED | 5 | VERIFICATION_STATUS_EXPIRED indicates the verification has expired |
 | VERIFICATION_STATUS_NEEDS_ADDITIONAL_FACTOR | 6 | VERIFICATION_STATUS_NEEDS_ADDITIONAL_FACTOR indicates borderline score requires MFA |
 | VERIFICATION_STATUS_ADDITIONAL_FACTOR_PENDING | 7 | VERIFICATION_STATUS_ADDITIONAL_FACTOR_PENDING indicates MFA challenge is in progress |
 

 
 <a name="virtengine.veid.v1.WalletStatus"></a>

 ### WalletStatus
 WalletStatus represents the overall status of an identity wallet

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | WALLET_STATUS_UNSPECIFIED | 0 | WALLET_STATUS_UNSPECIFIED indicates an unspecified status |
 | WALLET_STATUS_ACTIVE | 1 | WALLET_STATUS_ACTIVE indicates the wallet is active and usable |
 | WALLET_STATUS_SUSPENDED | 2 | WALLET_STATUS_SUSPENDED indicates the wallet is temporarily suspended |
 | WALLET_STATUS_REVOKED | 3 | WALLET_STATUS_REVOKED indicates the wallet has been revoked |
 | WALLET_STATUS_EXPIRED | 4 | WALLET_STATUS_EXPIRED indicates the wallet verification has expired |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/model.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/model.proto
 

 
 <a name="virtengine.veid.v1.MLModelInfo"></a>

 ### MLModelInfo
 MLModelInfo describes a registered ML model for version tracking

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `model_id` | [string](#string) |  | ModelID is the unique identifier for this model |
 | `name` | [string](#string) |  | Name is the human-readable name of the model |
 | `version` | [string](#string) |  | Version is the semantic version of the model (e.g., "1.0.0") |
 | `model_type` | [string](#string) |  | ModelType is the type of model (e.g., "face_verification") |
 | `sha256_hash` | [string](#string) |  | SHA256Hash is the SHA256 hash of the model binary |
 | `description` | [string](#string) |  | Description is a human-readable description of the model |
 | `activated_at` | [int64](#int64) |  | ActivatedAt is the block height when this model was activated |
 | `registered_at` | [int64](#int64) |  | RegisteredAt is the block height when this model was registered |
 | `registered_by` | [string](#string) |  | RegisteredBy is the address that registered this model |
 | `governance_id` | [uint64](#uint64) |  | GovernanceID is the governance proposal ID that approved this model |
 | `status` | [ModelStatus](#virtengine.veid.v1.ModelStatus) |  | Status is the current status of the model |
 
 

 

 
 <a name="virtengine.veid.v1.ModelParams"></a>

 ### ModelParams
 ModelParams contains parameters for model management

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `required_model_types` | [string](#string) | repeated | RequiredModelTypes lists model types that must be registered |
 | `activation_delay_blocks` | [int64](#int64) |  | ActivationDelayBlocks is the default delay for model activation |
 | `max_model_age_days` | [int32](#int32) |  | MaxModelAgeDays is the maximum age of a model before requiring update |
 | `allowed_registrars` | [string](#string) | repeated | AllowedRegistrars lists addresses allowed to register models |
 | `validator_sync_grace_period` | [int64](#int64) |  | ValidatorSyncGracePeriod is blocks allowed for validators to sync |
 | `model_update_quorum` | [uint32](#uint32) |  | ModelUpdateQuorum is the minimum voting power for model updates |
 | `enable_governance_updates` | [bool](#bool) |  | EnableGovernanceUpdates enables governance-controlled updates |
 
 

 

 
 <a name="virtengine.veid.v1.ModelUpdateProposal"></a>

 ### ModelUpdateProposal
 ModelUpdateProposal for governance-controlled model updates

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `title` | [string](#string) |  | Title is the proposal title |
 | `description` | [string](#string) |  | Description is the detailed proposal description |
 | `model_type` | [string](#string) |  | ModelType is the type of model being updated |
 | `new_model_id` | [string](#string) |  | NewModelID is the ID of the new model to activate |
 | `new_model_hash` | [string](#string) |  | NewModelHash is the SHA256 hash of the new model |
 | `activation_delay` | [int64](#int64) |  | ActivationDelay is the number of blocks to wait after approval |
 | `proposed_at` | [int64](#int64) |  | ProposedAt is the block height when proposal was submitted |
 | `proposer_address` | [string](#string) |  | ProposerAddress is the address that submitted the proposal |
 | `status` | [ModelProposalStatus](#virtengine.veid.v1.ModelProposalStatus) |  | Status is the current status of the proposal |
 | `governance_id` | [uint64](#uint64) |  | GovernanceID is the governance proposal ID (set when submitted to gov) |
 | `activation_height` | [int64](#int64) |  | ActivationHeight is the block height when activation should occur |
 
 

 

 
 <a name="virtengine.veid.v1.ModelVersionHistory"></a>

 ### ModelVersionHistory
 ModelVersionHistory tracks model version changes

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `history_id` | [string](#string) |  | HistoryID is the unique identifier for this history entry |
 | `model_type` | [string](#string) |  | ModelType is the type of model that was changed |
 | `old_model_id` | [string](#string) |  | OldModelID is the previous model ID (empty if first model) |
 | `new_model_id` | [string](#string) |  | NewModelID is the new model ID |
 | `old_model_hash` | [string](#string) |  | OldModelHash is the previous model hash |
 | `new_model_hash` | [string](#string) |  | NewModelHash is the new model hash |
 | `changed_at` | [int64](#int64) |  | ChangedAt is the block height when the change occurred |
 | `governance_id` | [uint64](#uint64) |  | GovernanceID is the governance proposal ID that approved this change |
 | `proposer_address` | [string](#string) |  | ProposerAddress is the address that proposed this change |
 | `reason` | [string](#string) |  | Reason is the reason for the change |
 
 

 

 
 <a name="virtengine.veid.v1.ModelVersionState"></a>

 ### ModelVersionState
 ModelVersionState tracks the current active model versions for consensus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `trust_score_model` | [string](#string) |  | TrustScoreModel is the active model ID for trust scoring |
 | `face_verification_model` | [string](#string) |  | FaceVerificationModel is the active model ID for face verification |
 | `liveness_model` | [string](#string) |  | LivenessModel is the active model ID for liveness detection |
 | `gan_detection_model` | [string](#string) |  | GANDetectionModel is the active model ID for GAN detection |
 | `ocr_model` | [string](#string) |  | OCRModel is the active model ID for OCR extraction |
 | `last_updated` | [int64](#int64) |  | LastUpdated is the block height when state was last updated |
 
 

 

 
 <a name="virtengine.veid.v1.MsgActivateModel"></a>

 ### MsgActivateModel
 MsgActivateModel activates a pending model after governance approval

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address authorized to activate models |
 | `model_type` | [string](#string) |  | ModelType is the type of model to activate |
 | `model_id` | [string](#string) |  | ModelID is the ID of the model to activate |
 | `governance_id` | [uint64](#uint64) |  | GovernanceID is the governance proposal ID |
 
 

 

 
 <a name="virtengine.veid.v1.MsgActivateModelResponse"></a>

 ### MsgActivateModelResponse
 MsgActivateModelResponse is the response for MsgActivateModel

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `activated_at` | [int64](#int64) |  | ActivatedAt is the block height when activated |
 
 

 

 
 <a name="virtengine.veid.v1.MsgDeprecateModel"></a>

 ### MsgDeprecateModel
 MsgDeprecateModel deprecates a model

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address authorized to deprecate models |
 | `model_id` | [string](#string) |  | ModelID is the ID of the model to deprecate |
 | `reason` | [string](#string) |  | Reason is the reason for deprecation |
 
 

 

 
 <a name="virtengine.veid.v1.MsgDeprecateModelResponse"></a>

 ### MsgDeprecateModelResponse
 MsgDeprecateModelResponse is the response for MsgDeprecateModel

 

 

 
 <a name="virtengine.veid.v1.MsgProposeModelUpdate"></a>

 ### MsgProposeModelUpdate
 MsgProposeModelUpdate proposes updating active model via governance

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `proposer` | [string](#string) |  | Proposer is the address proposing the update |
 | `proposal` | [ModelUpdateProposal](#virtengine.veid.v1.ModelUpdateProposal) |  | Proposal contains the update proposal details |
 
 

 

 
 <a name="virtengine.veid.v1.MsgProposeModelUpdateResponse"></a>

 ### MsgProposeModelUpdateResponse
 MsgProposeModelUpdateResponse is the response for MsgProposeModelUpdate

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `proposal_id` | [uint64](#uint64) |  | ProposalID is the governance proposal ID |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRegisterModel"></a>

 ### MsgRegisterModel
 MsgRegisterModel registers a new ML model

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address authorized to register models |
 | `model_info` | [MLModelInfo](#virtengine.veid.v1.MLModelInfo) |  | ModelInfo contains the model information |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRegisterModelResponse"></a>

 ### MsgRegisterModelResponse
 MsgRegisterModelResponse is the response for MsgRegisterModel

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `model_id` | [string](#string) |  | ModelID is the registered model's ID |
 | `status` | [ModelStatus](#virtengine.veid.v1.ModelStatus) |  | Status is the initial model status |
 
 

 

 
 <a name="virtengine.veid.v1.MsgReportModelVersion"></a>

 ### MsgReportModelVersion
 MsgReportModelVersion reports validator's model versions

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator's operator address |
 | `model_versions` | [MsgReportModelVersion.ModelVersionsEntry](#virtengine.veid.v1.MsgReportModelVersion.ModelVersionsEntry) | repeated | ModelVersions maps model type to SHA256 hash |
 
 

 

 
 <a name="virtengine.veid.v1.MsgReportModelVersion.ModelVersionsEntry"></a>

 ### MsgReportModelVersion.ModelVersionsEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.MsgReportModelVersionResponse"></a>

 ### MsgReportModelVersionResponse
 MsgReportModelVersionResponse is the response for MsgReportModelVersion

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `is_synced` | [bool](#bool) |  | IsSynced indicates if all models match consensus |
 | `mismatched_models` | [string](#string) | repeated | MismatchedModels lists any models that don't match |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRevokeModel"></a>

 ### MsgRevokeModel
 MsgRevokeModel revokes a model

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address authorized to revoke models |
 | `model_id` | [string](#string) |  | ModelID is the ID of the model to revoke |
 | `reason` | [string](#string) |  | Reason is the reason for revocation |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRevokeModelResponse"></a>

 ### MsgRevokeModelResponse
 MsgRevokeModelResponse is the response for MsgRevokeModel

 

 

 
 <a name="virtengine.veid.v1.ValidatorModelReport"></a>

 ### ValidatorModelReport
 ValidatorModelReport represents a validator's reported model versions

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator's operator address |
 | `model_versions` | [ValidatorModelReport.ModelVersionsEntry](#virtengine.veid.v1.ValidatorModelReport.ModelVersionsEntry) | repeated | ModelVersions maps model type to SHA256 hash |
 | `reported_at` | [int64](#int64) |  | ReportedAt is the block height when this report was submitted |
 | `last_verified` | [int64](#int64) |  | LastVerified is the block height when versions were last verified |
 | `is_synced` | [bool](#bool) |  | IsSynced indicates if all model versions match consensus |
 | `mismatched_models` | [string](#string) | repeated | MismatchedModels lists model types with version mismatches |
 
 

 

 
 <a name="virtengine.veid.v1.ValidatorModelReport.ModelVersionsEntry"></a>

 ### ValidatorModelReport.ModelVersionsEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

 
 <a name="virtengine.veid.v1.ModelProposalStatus"></a>

 ### ModelProposalStatus
 ModelProposalStatus represents the status of a model update proposal

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | MODEL_PROPOSAL_STATUS_UNSPECIFIED | 0 | MODEL_PROPOSAL_STATUS_UNSPECIFIED is the default unspecified status |
 | MODEL_PROPOSAL_STATUS_PENDING | 1 | MODEL_PROPOSAL_STATUS_PENDING indicates the proposal is pending |
 | MODEL_PROPOSAL_STATUS_APPROVED | 2 | MODEL_PROPOSAL_STATUS_APPROVED indicates the proposal was approved |
 | MODEL_PROPOSAL_STATUS_REJECTED | 3 | MODEL_PROPOSAL_STATUS_REJECTED indicates the proposal was rejected |
 | MODEL_PROPOSAL_STATUS_ACTIVATED | 4 | MODEL_PROPOSAL_STATUS_ACTIVATED indicates the model has been activated |
 | MODEL_PROPOSAL_STATUS_EXPIRED | 5 | MODEL_PROPOSAL_STATUS_EXPIRED indicates the proposal has expired |
 

 
 <a name="virtengine.veid.v1.ModelStatus"></a>

 ### ModelStatus
 ModelStatus represents the lifecycle status of a model

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | MODEL_STATUS_UNSPECIFIED | 0 | MODEL_STATUS_UNSPECIFIED is the default unspecified status |
 | MODEL_STATUS_PENDING | 1 | MODEL_STATUS_PENDING indicates the model is pending activation |
 | MODEL_STATUS_ACTIVE | 2 | MODEL_STATUS_ACTIVE indicates the model is currently active |
 | MODEL_STATUS_DEPRECATED | 3 | MODEL_STATUS_DEPRECATED indicates the model has been deprecated |
 | MODEL_STATUS_REVOKED | 4 | MODEL_STATUS_REVOKED indicates the model has been revoked |
 

 
 <a name="virtengine.veid.v1.ModelType"></a>

 ### ModelType
 ModelType represents the type of ML model

 | Name | Number | Description |
 | ---- | ------ | ----------- |
 | MODEL_TYPE_UNSPECIFIED | 0 | MODEL_TYPE_UNSPECIFIED is the default unspecified type |
 | MODEL_TYPE_TRUST_SCORE | 1 | MODEL_TYPE_TRUST_SCORE is the trust score calculation model |
 | MODEL_TYPE_FACE_VERIFICATION | 2 | MODEL_TYPE_FACE_VERIFICATION is the facial verification model |
 | MODEL_TYPE_LIVENESS | 3 | MODEL_TYPE_LIVENESS is the liveness detection model |
 | MODEL_TYPE_GAN_DETECTION | 4 | MODEL_TYPE_GAN_DETECTION is the GAN-generated image detection model |
 | MODEL_TYPE_OCR | 5 | MODEL_TYPE_OCR is the OCR extraction model |
 

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/genesis.proto
 

 
 <a name="virtengine.veid.v1.GenesisState"></a>

 ### GenesisState
 GenesisState defines the veid module's genesis state

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `identity_records` | [IdentityRecord](#virtengine.veid.v1.IdentityRecord) | repeated | IdentityRecords are the initial identity records |
 | `scopes` | [IdentityScope](#virtengine.veid.v1.IdentityScope) | repeated | Scopes are the initial identity scopes |
 | `approved_clients` | [ApprovedClient](#virtengine.veid.v1.ApprovedClient) | repeated | ApprovedClients are the initially approved clients |
 | `params` | [Params](#virtengine.veid.v1.Params) |  | Params are the module parameters |
 | `scores` | [IdentityScore](#virtengine.veid.v1.IdentityScore) | repeated | Scores are the initial identity scores |
 | `borderline_params` | [BorderlineParams](#virtengine.veid.v1.BorderlineParams) |  | BorderlineParams are the borderline fallback parameters |
 | `appeal_records` | [AppealRecord](#virtengine.veid.v1.AppealRecord) | repeated | AppealRecords are the initial appeal records |
 | `appeal_params` | [AppealParams](#virtengine.veid.v1.AppealParams) |  | AppealParams are the appeal system parameters |
 | `compliance_records` | [ComplianceRecord](#virtengine.veid.v1.ComplianceRecord) | repeated | ComplianceRecords are the initial compliance records |
 | `compliance_providers` | [ComplianceProvider](#virtengine.veid.v1.ComplianceProvider) | repeated | ComplianceProviders are the registered compliance providers |
 | `compliance_params` | [ComplianceParams](#virtengine.veid.v1.ComplianceParams) |  | ComplianceParams are the compliance configuration parameters |
 | `ml_models` | [MLModelInfo](#virtengine.veid.v1.MLModelInfo) | repeated | MLModels are the registered ML models |
 | `model_version_state` | [ModelVersionState](#virtengine.veid.v1.ModelVersionState) |  | ModelVersionState is the current active model versions |
 | `model_version_history` | [ModelVersionHistory](#virtengine.veid.v1.ModelVersionHistory) | repeated | ModelVersionHistory is the model version change history |
 | `model_params` | [ModelParams](#virtengine.veid.v1.ModelParams) |  | ModelParams are the model management parameters |
 | `pending_model_proposals` | [ModelUpdateProposal](#virtengine.veid.v1.ModelUpdateProposal) | repeated | PendingModelProposals are the pending model update proposals |
 | `validator_model_reports` | [ValidatorModelReport](#virtengine.veid.v1.ValidatorModelReport) | repeated | ValidatorModelReports are the validator model version reports |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/query.proto
 

 
 <a name="virtengine.veid.v1.PublicConsentInfo"></a>

 ### PublicConsentInfo
 PublicConsentInfo represents non-sensitive consent information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the scope identifier |
 | `granted` | [bool](#bool) |  | Granted indicates if consent is granted |
 | `is_active` | [bool](#bool) |  | IsActive indicates if consent is currently active |
 | `purpose` | [string](#string) |  | Purpose is the consent purpose |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when consent expires (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.PublicDerivedFeaturesInfo"></a>

 ### PublicDerivedFeaturesInfo
 PublicDerivedFeaturesInfo contains non-sensitive derived features metadata

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `has_face_embedding` | [bool](#bool) |  | HasFaceEmbedding indicates if face embedding hash exists |
 | `has_biometric` | [bool](#bool) |  | HasBiometric indicates if biometric hash exists |
 | `has_liveness_proof` | [bool](#bool) |  | HasLivenessProof indicates if liveness proof hash exists |
 | `doc_field_keys` | [string](#string) | repeated | DocFieldKeys lists which document fields have hashes |
 | `last_computed_at` | [int64](#int64) |  | LastComputedAt is when features were last computed (Unix timestamp) |
 | `model_version` | [string](#string) |  | ModelVersion is the model version used |
 | `feature_version` | [uint32](#uint32) |  | FeatureVersion is the schema version |
 
 

 

 
 <a name="virtengine.veid.v1.PublicVerificationHistoryEntry"></a>

 ### PublicVerificationHistoryEntry
 PublicVerificationHistoryEntry represents non-sensitive verification history

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `entry_id` | [string](#string) |  | EntryID is the entry identifier |
 | `timestamp` | [int64](#int64) |  | Timestamp is when this verification occurred (Unix timestamp) |
 | `block_height` | [int64](#int64) |  | BlockHeight is the block height |
 | `previous_score` | [uint32](#uint32) |  | PreviousScore is the score before verification |
 | `new_score` | [uint32](#uint32) |  | NewScore is the score after verification |
 | `previous_status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | PreviousStatus is the status before verification |
 | `new_status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | NewStatus is the status after verification |
 | `scope_count` | [int32](#int32) |  | ScopeCount is the number of scopes evaluated |
 | `model_version` | [string](#string) |  | ModelVersion is the model version used |
 
 

 

 
 <a name="virtengine.veid.v1.PublicWalletInfo"></a>

 ### PublicWalletInfo
 PublicWalletInfo contains non-sensitive wallet information

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `wallet_id` | [string](#string) |  | WalletID is the wallet identifier |
 | `account_address` | [string](#string) |  | AccountAddress is the owner's address |
 | `status` | [WalletStatus](#virtengine.veid.v1.WalletStatus) |  | Status is the wallet status |
 | `score` | [uint32](#uint32) |  | Score is the current identity score |
 | `tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | Tier is the current identity tier |
 | `scope_count` | [int32](#int32) |  | ScopeCount is the number of scopes in the wallet |
 | `verified_scope_count` | [int32](#int32) |  | VerifiedScopeCount is the number of verified scopes |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the wallet was created (Unix timestamp) |
 | `last_updated_at` | [int64](#int64) |  | LastUpdatedAt is when the wallet was last updated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.QueryActiveModelsRequest"></a>

 ### QueryActiveModelsRequest
 QueryActiveModelsRequest is the request for the ActiveModels query

 

 

 
 <a name="virtengine.veid.v1.QueryActiveModelsResponse"></a>

 ### QueryActiveModelsResponse
 QueryActiveModelsResponse is the response for the ActiveModels query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `state` | [ModelVersionState](#virtengine.veid.v1.ModelVersionState) |  | State is the current model version state |
 | `models` | [MLModelInfo](#virtengine.veid.v1.MLModelInfo) | repeated | Models is the list of active ML models |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealParamsRequest"></a>

 ### QueryAppealParamsRequest
 QueryAppealParamsRequest is the request for the AppealParams query

 

 

 
 <a name="virtengine.veid.v1.QueryAppealParamsResponse"></a>

 ### QueryAppealParamsResponse
 QueryAppealParamsResponse is the response for the AppealParams query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [AppealParams](#virtengine.veid.v1.AppealParams) |  | Params are the appeal parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealRequest"></a>

 ### QueryAppealRequest
 QueryAppealRequest is the request for the Appeal query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal_id` | [string](#string) |  | AppealID is the appeal to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealResponse"></a>

 ### QueryAppealResponse
 QueryAppealResponse is the response for the Appeal query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeal` | [AppealRecord](#virtengine.veid.v1.AppealRecord) |  | Appeal is the appeal record |
 | `found` | [bool](#bool) |  | Found indicates if the appeal was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealsByScopeRequest"></a>

 ### QueryAppealsByScopeRequest
 QueryAppealsByScopeRequest is the request for the AppealsByScope query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the scope to query appeals for |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealsByScopeResponse"></a>

 ### QueryAppealsByScopeResponse
 QueryAppealsByScopeResponse is the response for the AppealsByScope query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeals` | [AppealRecord](#virtengine.veid.v1.AppealRecord) | repeated | Appeals is the list of appeal records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealsRequest"></a>

 ### QueryAppealsRequest
 QueryAppealsRequest is the request for the Appeals query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query appeals for |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryAppealsResponse"></a>

 ### QueryAppealsResponse
 QueryAppealsResponse is the response for the Appeals query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `appeals` | [AppealRecord](#virtengine.veid.v1.AppealRecord) | repeated | Appeals is the list of appeal records |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.veid.v1.QueryApprovedClientsRequest"></a>

 ### QueryApprovedClientsRequest
 QueryApprovedClientsRequest is the request for the ApprovedClients query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `active_only` | [bool](#bool) |  | ActiveOnly filters to only active clients |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is the pagination options |
 
 

 

 
 <a name="virtengine.veid.v1.QueryApprovedClientsResponse"></a>

 ### QueryApprovedClientsResponse
 QueryApprovedClientsResponse is the response for the ApprovedClients query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `clients` | [ApprovedClient](#virtengine.veid.v1.ApprovedClient) | repeated | Clients is the list of approved clients |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.veid.v1.QueryBorderlineParamsRequest"></a>

 ### QueryBorderlineParamsRequest
 QueryBorderlineParamsRequest is the request for the BorderlineParams query

 

 

 
 <a name="virtengine.veid.v1.QueryBorderlineParamsResponse"></a>

 ### QueryBorderlineParamsResponse
 QueryBorderlineParamsResponse is the response for the BorderlineParams query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [BorderlineParams](#virtengine.veid.v1.BorderlineParams) |  | Params are the borderline parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceParamsRequest"></a>

 ### QueryComplianceParamsRequest
 QueryComplianceParamsRequest is the request for the ComplianceParams query

 

 

 
 <a name="virtengine.veid.v1.QueryComplianceParamsResponse"></a>

 ### QueryComplianceParamsResponse
 QueryComplianceParamsResponse is the response for the ComplianceParams query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [ComplianceParams](#virtengine.veid.v1.ComplianceParams) |  | Params are the compliance parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceProviderRequest"></a>

 ### QueryComplianceProviderRequest
 QueryComplianceProviderRequest is the request for the ComplianceProvider query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider_id` | [string](#string) |  | ProviderID is the provider to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceProviderResponse"></a>

 ### QueryComplianceProviderResponse
 QueryComplianceProviderResponse is the response for the ComplianceProvider query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `provider` | [ComplianceProvider](#virtengine.veid.v1.ComplianceProvider) |  | Provider is the compliance provider |
 | `found` | [bool](#bool) |  | Found indicates if the provider was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceProvidersRequest"></a>

 ### QueryComplianceProvidersRequest
 QueryComplianceProvidersRequest is the request for the ComplianceProviders query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `active_only` | [bool](#bool) |  | ActiveOnly filters to only return active providers |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceProvidersResponse"></a>

 ### QueryComplianceProvidersResponse
 QueryComplianceProvidersResponse is the response for the ComplianceProviders query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `providers` | [ComplianceProvider](#virtengine.veid.v1.ComplianceProvider) | repeated | Providers is the list of compliance providers |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceStatusRequest"></a>

 ### QueryComplianceStatusRequest
 QueryComplianceStatusRequest is the request for the ComplianceStatus query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query compliance for |
 
 

 

 
 <a name="virtengine.veid.v1.QueryComplianceStatusResponse"></a>

 ### QueryComplianceStatusResponse
 QueryComplianceStatusResponse is the response for the ComplianceStatus query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `record` | [ComplianceRecord](#virtengine.veid.v1.ComplianceRecord) |  | Record is the compliance record |
 | `found` | [bool](#bool) |  | Found indicates if the record was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryConsentSettingsRequest"></a>

 ### QueryConsentSettingsRequest
 QueryConsentSettingsRequest is the request for the ConsentSettings query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query consent for |
 | `scope_id` | [string](#string) |  | ScopeID is an optional filter for specific scope consent |
 
 

 

 
 <a name="virtengine.veid.v1.QueryConsentSettingsResponse"></a>

 ### QueryConsentSettingsResponse
 QueryConsentSettingsResponse is the response for the ConsentSettings query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `global_settings` | [ConsentSettings](#virtengine.veid.v1.ConsentSettings) |  | GlobalSettings contains global consent settings |
 | `scope_consents` | [PublicConsentInfo](#virtengine.veid.v1.PublicConsentInfo) | repeated | ScopeConsents contains per-scope consent info |
 | `consent_version` | [uint32](#uint32) |  | ConsentVersion is the current consent version |
 | `last_updated_at` | [int64](#int64) |  | LastUpdatedAt is when consent was last updated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.QueryDerivedFeatureHashesRequest"></a>

 ### QueryDerivedFeatureHashesRequest
 QueryDerivedFeatureHashesRequest is the request for the DerivedFeatureHashes query
This is used for verification matching by authorized parties

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query hashes for |
 | `requester` | [string](#string) |  | Requester is the address requesting the hashes |
 | `purpose` | [string](#string) |  | Purpose describes why the hashes are being requested |
 
 

 

 
 <a name="virtengine.veid.v1.QueryDerivedFeatureHashesResponse"></a>

 ### QueryDerivedFeatureHashesResponse
 QueryDerivedFeatureHashesResponse is the response for the DerivedFeatureHashes query
Only returned if consent allows sharing

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `allowed` | [bool](#bool) |  | Allowed indicates if the request was allowed based on consent |
 | `denial_reason` | [string](#string) |  | DenialReason is set if Allowed is false |
 | `face_embedding_hash` | [bytes](#bytes) |  | FaceEmbeddingHash is the face embedding hash (if consented) |
 | `doc_field_hashes` | [QueryDerivedFeatureHashesResponse.DocFieldHashesEntry](#virtengine.veid.v1.QueryDerivedFeatureHashesResponse.DocFieldHashesEntry) | repeated | DocFieldHashes are document field hashes (if consented) |
 | `biometric_hash` | [bytes](#bytes) |  | BiometricHash is the biometric hash (if consented) |
 | `model_version` | [string](#string) |  | ModelVersion is the model version used |
 
 

 

 
 <a name="virtengine.veid.v1.QueryDerivedFeatureHashesResponse.DocFieldHashesEntry"></a>

 ### QueryDerivedFeatureHashesResponse.DocFieldHashesEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.QueryDerivedFeaturesRequest"></a>

 ### QueryDerivedFeaturesRequest
 QueryDerivedFeaturesRequest is the request for the DerivedFeatures query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query features for |
 
 

 

 
 <a name="virtengine.veid.v1.QueryDerivedFeaturesResponse"></a>

 ### QueryDerivedFeaturesResponse
 QueryDerivedFeaturesResponse is the response for the DerivedFeatures query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `features` | [PublicDerivedFeaturesInfo](#virtengine.veid.v1.PublicDerivedFeaturesInfo) |  | Features contains non-sensitive derived features information |
 | `found` | [bool](#bool) |  | Found indicates if features were found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityRecordRequest"></a>

 ### QueryIdentityRecordRequest
 QueryIdentityRecordRequest is the request for the IdentityRecord query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityRecordResponse"></a>

 ### QueryIdentityRecordResponse
 QueryIdentityRecordResponse is the response for the IdentityRecord query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `record` | [IdentityRecord](#virtengine.veid.v1.IdentityRecord) |  | Record is the identity record |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityRequest"></a>

 ### QueryIdentityRequest
 QueryIdentityRequest is the request for the Identity query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityResponse"></a>

 ### QueryIdentityResponse
 QueryIdentityResponse is the response for the Identity query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `identity` | [IdentityRecord](#virtengine.veid.v1.IdentityRecord) |  | Identity is the identity record |
 | `found` | [bool](#bool) |  | Found indicates if the identity was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityScoreRequest"></a>

 ### QueryIdentityScoreRequest
 QueryIdentityScoreRequest is the request for the IdentityScore query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query the score for |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityScoreResponse"></a>

 ### QueryIdentityScoreResponse
 QueryIdentityScoreResponse is the response for the IdentityScore query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `score` | [IdentityScore](#virtengine.veid.v1.IdentityScore) |  | Score is the identity score details |
 | `found` | [bool](#bool) |  | Found indicates if a score was found for the account |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityStatusRequest"></a>

 ### QueryIdentityStatusRequest
 QueryIdentityStatusRequest is the request for the IdentityStatus query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query the status for |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityStatusResponse"></a>

 ### QueryIdentityStatusResponse
 QueryIdentityStatusResponse is the response for the IdentityStatus query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the queried address |
 | `status` | [AccountStatus](#virtengine.veid.v1.AccountStatus) |  | Status is the current verification status |
 | `tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | Tier is the current identity tier |
 | `score` | [uint32](#uint32) |  | Score is the current score |
 | `model_version` | [string](#string) |  | ModelVersion is the ML model version used |
 | `last_updated_at` | [int64](#int64) |  | LastUpdatedAt is when the status was last updated (Unix timestamp) |
 | `found` | [bool](#bool) |  | Found indicates if the account exists |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityWalletRequest"></a>

 ### QueryIdentityWalletRequest
 QueryIdentityWalletRequest is the request for the IdentityWallet query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query the wallet for |
 
 

 

 
 <a name="virtengine.veid.v1.QueryIdentityWalletResponse"></a>

 ### QueryIdentityWalletResponse
 QueryIdentityWalletResponse is the response for the IdentityWallet query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `wallet` | [PublicWalletInfo](#virtengine.veid.v1.PublicWalletInfo) |  | Wallet contains non-sensitive wallet information |
 | `found` | [bool](#bool) |  | Found indicates if the wallet was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryModelHistoryRequest"></a>

 ### QueryModelHistoryRequest
 QueryModelHistoryRequest is the request for the ModelHistory query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `model_type` | [string](#string) |  | ModelType is the type of model to query history for |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is optional pagination parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryModelHistoryResponse"></a>

 ### QueryModelHistoryResponse
 QueryModelHistoryResponse is the response for the ModelHistory query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `history` | [ModelVersionHistory](#virtengine.veid.v1.ModelVersionHistory) | repeated | History is the list of model version changes |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.veid.v1.QueryModelParamsRequest"></a>

 ### QueryModelParamsRequest
 QueryModelParamsRequest is the request for the ModelParams query

 

 

 
 <a name="virtengine.veid.v1.QueryModelParamsResponse"></a>

 ### QueryModelParamsResponse
 QueryModelParamsResponse is the response for the ModelParams query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [ModelParams](#virtengine.veid.v1.ModelParams) |  | Params are the model management parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryModelVersionRequest"></a>

 ### QueryModelVersionRequest
 QueryModelVersionRequest is the request for the ModelVersion query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `model_type` | [string](#string) |  | ModelType is the type of model to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryModelVersionResponse"></a>

 ### QueryModelVersionResponse
 QueryModelVersionResponse is the response for the ModelVersion query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `model_info` | [MLModelInfo](#virtengine.veid.v1.MLModelInfo) |  | ModelInfo is the model information |
 | `found` | [bool](#bool) |  | Found indicates if the model was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request for the Params query

 

 

 
 <a name="virtengine.veid.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response for the Params query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.veid.v1.Params) |  | Params are the module parameters |
 
 

 

 
 <a name="virtengine.veid.v1.QueryScopeRequest"></a>

 ### QueryScopeRequest
 QueryScopeRequest is the request for the Scope query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account that owns the scope |
 | `scope_id` | [string](#string) |  | ScopeID is the scope to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryScopeResponse"></a>

 ### QueryScopeResponse
 QueryScopeResponse is the response for the Scope query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope` | [IdentityScope](#virtengine.veid.v1.IdentityScope) |  | Scope is the identity scope |
 | `found` | [bool](#bool) |  | Found indicates if the scope was found |
 
 

 

 
 <a name="virtengine.veid.v1.QueryScopesByTypeRequest"></a>

 ### QueryScopesByTypeRequest
 QueryScopesByTypeRequest is the request for the ScopesByType query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query scopes for |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType is the type of scopes to filter by |
 
 

 

 
 <a name="virtengine.veid.v1.QueryScopesByTypeResponse"></a>

 ### QueryScopesByTypeResponse
 QueryScopesByTypeResponse is the response for the ScopesByType query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scopes` | [IdentityScope](#virtengine.veid.v1.IdentityScope) | repeated | Scopes is the list of identity scopes of the specified type |
 
 

 

 
 <a name="virtengine.veid.v1.QueryScopesRequest"></a>

 ### QueryScopesRequest
 QueryScopesRequest is the request for the Scopes query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query scopes for |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType is an optional filter by scope type |
 | `status_filter` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | StatusFilter is an optional filter by verification status |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is the pagination options |
 
 

 

 
 <a name="virtengine.veid.v1.QueryScopesResponse"></a>

 ### QueryScopesResponse
 QueryScopesResponse is the response for the Scopes query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scopes` | [IdentityScope](#virtengine.veid.v1.IdentityScope) | repeated | Scopes is the list of identity scopes |
 | `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination is the pagination response |
 
 

 

 
 <a name="virtengine.veid.v1.QueryValidatorModelSyncRequest"></a>

 ### QueryValidatorModelSyncRequest
 QueryValidatorModelSyncRequest is the request for the ValidatorModelSync query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `validator_address` | [string](#string) |  | ValidatorAddress is the validator to query |
 
 

 

 
 <a name="virtengine.veid.v1.QueryValidatorModelSyncResponse"></a>

 ### QueryValidatorModelSyncResponse
 QueryValidatorModelSyncResponse is the response for the ValidatorModelSync query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `report` | [ValidatorModelReport](#virtengine.veid.v1.ValidatorModelReport) |  | Report is the validator's model version report |
 | `is_synced` | [bool](#bool) |  | IsSynced indicates if the validator is synced with consensus |
 
 

 

 
 <a name="virtengine.veid.v1.QueryVerificationHistoryRequest"></a>

 ### QueryVerificationHistoryRequest
 QueryVerificationHistoryRequest is the request for the VerificationHistory query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query history for |
 | `limit` | [uint32](#uint32) |  | Limit is the maximum number of entries to return |
 | `offset` | [uint32](#uint32) |  | Offset is the number of entries to skip |
 
 

 

 
 <a name="virtengine.veid.v1.QueryVerificationHistoryResponse"></a>

 ### QueryVerificationHistoryResponse
 QueryVerificationHistoryResponse is the response for the VerificationHistory query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `entries` | [PublicVerificationHistoryEntry](#virtengine.veid.v1.PublicVerificationHistoryEntry) | repeated | Entries contains verification history entries |
 | `total_count` | [int32](#int32) |  | TotalCount is the total number of entries |
 
 

 

 
 <a name="virtengine.veid.v1.QueryWalletScopesRequest"></a>

 ### QueryWalletScopesRequest
 QueryWalletScopesRequest is the request for the WalletScopes query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account to query scopes for |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType is an optional filter by scope type |
 | `status_filter` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | StatusFilter is an optional filter by verification status |
 | `active_only` | [bool](#bool) |  | ActiveOnly filters to only active scopes |
 
 

 

 
 <a name="virtengine.veid.v1.QueryWalletScopesResponse"></a>

 ### QueryWalletScopesResponse
 QueryWalletScopesResponse is the response for the WalletScopes query

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scopes` | [WalletScopeInfo](#virtengine.veid.v1.WalletScopeInfo) | repeated | Scopes contains non-sensitive scope information |
 | `total_count` | [int32](#int32) |  | TotalCount is the total number of scopes in the wallet |
 | `active_count` | [int32](#int32) |  | ActiveCount is the number of active scopes |
 
 

 

 
 <a name="virtengine.veid.v1.WalletScopeInfo"></a>

 ### WalletScopeInfo
 WalletScopeInfo represents non-sensitive scope information in a wallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the scope identifier |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType indicates what kind of identity data this scope contains |
 | `status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | Status is the current verification status |
 | `added_at` | [int64](#int64) |  | AddedAt is when the scope was added (Unix timestamp) |
 | `verified_at` | [int64](#int64) |  | VerifiedAt is when the scope was verified (Unix timestamp) |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the verification expires (Unix timestamp) |
 | `consent_granted` | [bool](#bool) |  | ConsentGranted indicates if consent is granted for this scope |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.veid.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service for the veid module

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `IdentityRecord` | [QueryIdentityRecordRequest](#virtengine.veid.v1.QueryIdentityRecordRequest) | [QueryIdentityRecordResponse](#virtengine.veid.v1.QueryIdentityRecordResponse) | IdentityRecord queries an identity record by account address buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/identity_record/{account_address}|
 | `Identity` | [QueryIdentityRequest](#virtengine.veid.v1.QueryIdentityRequest) | [QueryIdentityResponse](#virtengine.veid.v1.QueryIdentityResponse) | Identity queries an identity record by account address (alias for IdentityRecord) buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/identity/{account_address}|
 | `Scope` | [QueryScopeRequest](#virtengine.veid.v1.QueryScopeRequest) | [QueryScopeResponse](#virtengine.veid.v1.QueryScopeResponse) | Scope queries a specific scope by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/scope/{account_address}/{scope_id}|
 | `Scopes` | [QueryScopesRequest](#virtengine.veid.v1.QueryScopesRequest) | [QueryScopesResponse](#virtengine.veid.v1.QueryScopesResponse) | Scopes queries all scopes for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/scopes/{account_address}|
 | `ScopesByType` | [QueryScopesByTypeRequest](#virtengine.veid.v1.QueryScopesByTypeRequest) | [QueryScopesByTypeResponse](#virtengine.veid.v1.QueryScopesByTypeResponse) | ScopesByType queries all scopes of a specific type for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/scopes/{account_address}/type/{scope_type}|
 | `IdentityScore` | [QueryIdentityScoreRequest](#virtengine.veid.v1.QueryIdentityScoreRequest) | [QueryIdentityScoreResponse](#virtengine.veid.v1.QueryIdentityScoreResponse) | IdentityScore queries the identity score for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/score/{account_address}|
 | `IdentityStatus` | [QueryIdentityStatusRequest](#virtengine.veid.v1.QueryIdentityStatusRequest) | [QueryIdentityStatusResponse](#virtengine.veid.v1.QueryIdentityStatusResponse) | IdentityStatus queries the identity status for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/status/{account_address}|
 | `IdentityWallet` | [QueryIdentityWalletRequest](#virtengine.veid.v1.QueryIdentityWalletRequest) | [QueryIdentityWalletResponse](#virtengine.veid.v1.QueryIdentityWalletResponse) | IdentityWallet queries an identity wallet by account address buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/wallet/{account_address}|
 | `WalletScopes` | [QueryWalletScopesRequest](#virtengine.veid.v1.QueryWalletScopesRequest) | [QueryWalletScopesResponse](#virtengine.veid.v1.QueryWalletScopesResponse) | WalletScopes queries all scope references in a wallet buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/wallet/{account_address}/scopes|
 | `ConsentSettings` | [QueryConsentSettingsRequest](#virtengine.veid.v1.QueryConsentSettingsRequest) | [QueryConsentSettingsResponse](#virtengine.veid.v1.QueryConsentSettingsResponse) | ConsentSettings queries consent settings for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/consent/{account_address}|
 | `DerivedFeatures` | [QueryDerivedFeaturesRequest](#virtengine.veid.v1.QueryDerivedFeaturesRequest) | [QueryDerivedFeaturesResponse](#virtengine.veid.v1.QueryDerivedFeaturesResponse) | DerivedFeatures queries derived features metadata for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/derived_features/{account_address}|
 | `DerivedFeatureHashes` | [QueryDerivedFeatureHashesRequest](#virtengine.veid.v1.QueryDerivedFeatureHashesRequest) | [QueryDerivedFeatureHashesResponse](#virtengine.veid.v1.QueryDerivedFeatureHashesResponse) | DerivedFeatureHashes queries derived feature hashes for an account (consent-gated) buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/derived_feature_hashes/{account_address}|
 | `VerificationHistory` | [QueryVerificationHistoryRequest](#virtengine.veid.v1.QueryVerificationHistoryRequest) | [QueryVerificationHistoryResponse](#virtengine.veid.v1.QueryVerificationHistoryResponse) | VerificationHistory queries verification history for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/history/{account_address}|
 | `ApprovedClients` | [QueryApprovedClientsRequest](#virtengine.veid.v1.QueryApprovedClientsRequest) | [QueryApprovedClientsResponse](#virtengine.veid.v1.QueryApprovedClientsResponse) | ApprovedClients queries all approved clients buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/clients|
 | `Params` | [QueryParamsRequest](#virtengine.veid.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.veid.v1.QueryParamsResponse) | Params queries the module parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/params|
 | `BorderlineParams` | [QueryBorderlineParamsRequest](#virtengine.veid.v1.QueryBorderlineParamsRequest) | [QueryBorderlineParamsResponse](#virtengine.veid.v1.QueryBorderlineParamsResponse) | BorderlineParams queries the borderline parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/borderline_params|
 | `Appeal` | [QueryAppealRequest](#virtengine.veid.v1.QueryAppealRequest) | [QueryAppealResponse](#virtengine.veid.v1.QueryAppealResponse) | Appeal queries a specific appeal by ID buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/appeal/{appeal_id}|
 | `Appeals` | [QueryAppealsRequest](#virtengine.veid.v1.QueryAppealsRequest) | [QueryAppealsResponse](#virtengine.veid.v1.QueryAppealsResponse) | Appeals queries all appeals for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/appeals/{account_address}|
 | `AppealsByScope` | [QueryAppealsByScopeRequest](#virtengine.veid.v1.QueryAppealsByScopeRequest) | [QueryAppealsByScopeResponse](#virtengine.veid.v1.QueryAppealsByScopeResponse) | AppealsByScope queries all appeals for a specific scope buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/appeals/scope/{scope_id}|
 | `AppealParams` | [QueryAppealParamsRequest](#virtengine.veid.v1.QueryAppealParamsRequest) | [QueryAppealParamsResponse](#virtengine.veid.v1.QueryAppealParamsResponse) | AppealParams queries the appeal system parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/appeal_params|
 | `ComplianceStatus` | [QueryComplianceStatusRequest](#virtengine.veid.v1.QueryComplianceStatusRequest) | [QueryComplianceStatusResponse](#virtengine.veid.v1.QueryComplianceStatusResponse) | ComplianceStatus queries the compliance status for an account buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/compliance/{account_address}|
 | `ComplianceProvider` | [QueryComplianceProviderRequest](#virtengine.veid.v1.QueryComplianceProviderRequest) | [QueryComplianceProviderResponse](#virtengine.veid.v1.QueryComplianceProviderResponse) | ComplianceProvider queries a specific compliance provider buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/compliance_provider/{provider_id}|
 | `ComplianceProviders` | [QueryComplianceProvidersRequest](#virtengine.veid.v1.QueryComplianceProvidersRequest) | [QueryComplianceProvidersResponse](#virtengine.veid.v1.QueryComplianceProvidersResponse) | ComplianceProviders queries all compliance providers buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/compliance_providers|
 | `ComplianceParams` | [QueryComplianceParamsRequest](#virtengine.veid.v1.QueryComplianceParamsRequest) | [QueryComplianceParamsResponse](#virtengine.veid.v1.QueryComplianceParamsResponse) | ComplianceParams queries the compliance parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/compliance_params|
 | `ModelVersion` | [QueryModelVersionRequest](#virtengine.veid.v1.QueryModelVersionRequest) | [QueryModelVersionResponse](#virtengine.veid.v1.QueryModelVersionResponse) | ModelVersion queries the active model for a given type buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/model/{model_type}|
 | `ActiveModels` | [QueryActiveModelsRequest](#virtengine.veid.v1.QueryActiveModelsRequest) | [QueryActiveModelsResponse](#virtengine.veid.v1.QueryActiveModelsResponse) | ActiveModels queries all active models buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/models|
 | `ModelHistory` | [QueryModelHistoryRequest](#virtengine.veid.v1.QueryModelHistoryRequest) | [QueryModelHistoryResponse](#virtengine.veid.v1.QueryModelHistoryResponse) | ModelHistory queries the version history for a model type buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/model/{model_type}/history|
 | `ValidatorModelSync` | [QueryValidatorModelSyncRequest](#virtengine.veid.v1.QueryValidatorModelSyncRequest) | [QueryValidatorModelSyncResponse](#virtengine.veid.v1.QueryValidatorModelSyncResponse) | ValidatorModelSync queries a validator's model sync status buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/validator_model_sync/{validator_address}|
 | `ModelParams` | [QueryModelParamsRequest](#virtengine.veid.v1.QueryModelParamsRequest) | [QueryModelParamsResponse](#virtengine.veid.v1.QueryModelParamsResponse) | ModelParams queries the model management parameters buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE buf:lint:ignore RPC_RESPONSE_STANDARD_NAME | GET|/virtengine/veid/v1/model_params|
 
  <!-- end services -->

 
 
 <a name="virtengine/veid/v1/tx.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/veid/v1/tx.proto
 

 
 <a name="virtengine.veid.v1.MsgAddScopeToWallet"></a>

 ### MsgAddScopeToWallet
 MsgAddScopeToWallet is the message for adding a scope reference to a wallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account that owns the wallet |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier of the scope to add |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType is the type of scope being added |
 | `envelope_hash` | [bytes](#bytes) |  | EnvelopeHash is the hash of the encrypted scope envelope |
 | `user_signature` | [bytes](#bytes) |  | UserSignature authorizes adding this scope to the wallet |
 | `grant_consent` | [bool](#bool) |  | GrantConsent indicates if consent should be granted for this scope |
 | `consent_purpose` | [string](#string) |  | ConsentPurpose describes the purpose for consent |
 | `consent_expires_at` | [int64](#int64) |  | ConsentExpiresAt is when consent expires (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgAddScopeToWalletResponse"></a>

 ### MsgAddScopeToWalletResponse
 MsgAddScopeToWalletResponse is the response for MsgAddScopeToWallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the ID of the added scope |
 | `added_at` | [int64](#int64) |  | AddedAt is when the scope was added (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgCompleteBorderlineFallback"></a>

 ### MsgCompleteBorderlineFallback
 MsgCompleteBorderlineFallback is the message for completing a borderline fallback

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account completing the fallback |
 | `challenge_id` | [string](#string) |  | ChallengeID is the MFA challenge ID that was satisfied |
 | `factors_satisfied` | [string](#string) | repeated | FactorsSatisfied are the factor types that were successfully verified |
 
 

 

 
 <a name="virtengine.veid.v1.MsgCompleteBorderlineFallbackResponse"></a>

 ### MsgCompleteBorderlineFallbackResponse
 MsgCompleteBorderlineFallbackResponse is the response for MsgCompleteBorderlineFallback

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `fallback_id` | [string](#string) |  | FallbackID is the ID of the completed fallback |
 | `final_status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | FinalStatus is the resulting verification status |
 | `factor_class` | [string](#string) |  | FactorClass is the security class of the satisfied factors |
 
 

 

 
 <a name="virtengine.veid.v1.MsgCreateIdentityWallet"></a>

 ### MsgCreateIdentityWallet
 MsgCreateIdentityWallet is the message for creating an identity wallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account creating the wallet |
 | `binding_signature` | [bytes](#bytes) |  | BindingSignature is the signature proving ownership of the account |
 | `binding_pub_key` | [bytes](#bytes) |  | BindingPubKey is the public key used for the binding signature |
 | `initial_consent` | [ConsentSettings](#virtengine.veid.v1.ConsentSettings) |  | InitialConsent contains optional initial consent settings |
 | `metadata` | [MsgCreateIdentityWallet.MetadataEntry](#virtengine.veid.v1.MsgCreateIdentityWallet.MetadataEntry) | repeated | Metadata contains optional wallet metadata |
 
 

 

 
 <a name="virtengine.veid.v1.MsgCreateIdentityWallet.MetadataEntry"></a>

 ### MsgCreateIdentityWallet.MetadataEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [string](#string) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.MsgCreateIdentityWalletResponse"></a>

 ### MsgCreateIdentityWalletResponse
 MsgCreateIdentityWalletResponse is the response for MsgCreateIdentityWallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `wallet_id` | [string](#string) |  | WalletID is the ID of the created wallet |
 | `created_at` | [int64](#int64) |  | CreatedAt is when the wallet was created (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRebindWallet"></a>

 ### MsgRebindWallet
 MsgRebindWallet is the message for rebinding a wallet during key rotation

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account that owns the wallet |
 | `new_binding_signature` | [bytes](#bytes) |  | NewBindingSignature is the new binding signature with the new key |
 | `new_binding_pub_key` | [bytes](#bytes) |  | NewBindingPubKey is the new public key |
 | `old_signature` | [bytes](#bytes) |  | OldSignature proves ownership of the old key |
 | `mfa_proof` | [bytes](#bytes) |  | MFAProof is proof of MFA for key rotation (serialized) |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is the client device fingerprint |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRebindWalletResponse"></a>

 ### MsgRebindWalletResponse
 MsgRebindWalletResponse is the response for MsgRebindWallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `rebound_at` | [int64](#int64) |  | ReboundAt is when the wallet was rebound (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRequestVerification"></a>

 ### MsgRequestVerification
 MsgRequestVerification is the message for requesting verification of a scope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account requesting verification |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier of the scope to verify |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRequestVerificationResponse"></a>

 ### MsgRequestVerificationResponse
 MsgRequestVerificationResponse is the response for MsgRequestVerification

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the ID of the scope |
 | `status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | Status is the new verification status |
 | `requested_at` | [int64](#int64) |  | RequestedAt is when verification was requested (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRevokeScope"></a>

 ### MsgRevokeScope
 MsgRevokeScope is the message for revoking an identity scope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account revoking the scope |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier of the scope to revoke |
 | `reason` | [string](#string) |  | Reason is the reason for revocation |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRevokeScopeFromWallet"></a>

 ### MsgRevokeScopeFromWallet
 MsgRevokeScopeFromWallet is the message for revoking a scope from a wallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account that owns the wallet |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier of the scope to revoke |
 | `reason` | [string](#string) |  | Reason is the reason for revocation |
 | `user_signature` | [bytes](#bytes) |  | UserSignature authorizes revoking this scope |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRevokeScopeFromWalletResponse"></a>

 ### MsgRevokeScopeFromWalletResponse
 MsgRevokeScopeFromWalletResponse is the response for MsgRevokeScopeFromWallet

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the ID of the revoked scope |
 | `revoked_at` | [int64](#int64) |  | RevokedAt is when the scope was revoked (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgRevokeScopeResponse"></a>

 ### MsgRevokeScopeResponse
 MsgRevokeScopeResponse is the response for MsgRevokeScope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the ID of the revoked scope |
 | `revoked_at` | [int64](#int64) |  | RevokedAt is when the scope was revoked (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateBorderlineParams"></a>

 ### MsgUpdateBorderlineParams
 MsgUpdateBorderlineParams is the message for updating borderline parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the governance module account address |
 | `params` | [BorderlineParams](#virtengine.veid.v1.BorderlineParams) |  | Params are the new borderline parameters |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateBorderlineParamsResponse"></a>

 ### MsgUpdateBorderlineParamsResponse
 MsgUpdateBorderlineParamsResponse is the response for MsgUpdateBorderlineParams

 

 

 
 <a name="virtengine.veid.v1.MsgUpdateConsentSettings"></a>

 ### MsgUpdateConsentSettings
 MsgUpdateConsentSettings is the message for updating consent settings

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account that owns the wallet |
 | `scope_id` | [string](#string) |  | ScopeID is the scope to update consent for (empty for global settings) |
 | `grant_consent` | [bool](#bool) |  | GrantConsent indicates whether to grant or revoke consent |
 | `purpose` | [string](#string) |  | Purpose is the purpose for granting consent |
 | `expires_at` | [int64](#int64) |  | ExpiresAt is when the consent should expire (Unix timestamp) |
 | `global_settings` | [GlobalConsentUpdate](#virtengine.veid.v1.GlobalConsentUpdate) |  | GlobalSettings contains global settings updates |
 | `user_signature` | [bytes](#bytes) |  | UserSignature authorizes this consent update |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateConsentSettingsResponse"></a>

 ### MsgUpdateConsentSettingsResponse
 MsgUpdateConsentSettingsResponse is the response for MsgUpdateConsentSettings

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when consent was updated (Unix timestamp) |
 | `consent_version` | [uint32](#uint32) |  | ConsentVersion is the new consent version |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateDerivedFeatures"></a>

 ### MsgUpdateDerivedFeatures
 MsgUpdateDerivedFeatures is the message for updating derived features

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the validator submitting the update |
 | `account_address` | [string](#string) |  | AccountAddress is the account to update features for |
 | `face_embedding_hash` | [bytes](#bytes) |  | FaceEmbeddingHash is the new face embedding hash |
 | `doc_field_hashes` | [MsgUpdateDerivedFeatures.DocFieldHashesEntry](#virtengine.veid.v1.MsgUpdateDerivedFeatures.DocFieldHashesEntry) | repeated | DocFieldHashes are new document field hashes |
 | `biometric_hash` | [bytes](#bytes) |  | BiometricHash is the new biometric hash |
 | `liveness_proof_hash` | [bytes](#bytes) |  | LivenessProofHash is the new liveness proof hash |
 | `model_version` | [string](#string) |  | ModelVersion is the ML model version used |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateDerivedFeatures.DocFieldHashesEntry"></a>

 ### MsgUpdateDerivedFeatures.DocFieldHashesEntry
 

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `key` | [string](#string) |  |  |
 | `value` | [bytes](#bytes) |  |  |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateDerivedFeaturesResponse"></a>

 ### MsgUpdateDerivedFeaturesResponse
 MsgUpdateDerivedFeaturesResponse is the response for MsgUpdateDerivedFeatures

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when the features were updated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the message for updating module parameters

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the governance module account address |
 | `params` | [Params](#virtengine.veid.v1.Params) |  | Params are the new module parameters |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse is the response for MsgUpdateParams

 

 

 
 <a name="virtengine.veid.v1.MsgUpdateScore"></a>

 ### MsgUpdateScore
 MsgUpdateScore is the message for validators to update identity score

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the validator updating the score |
 | `account_address` | [string](#string) |  | AccountAddress is the account whose score is being updated |
 | `new_score` | [uint32](#uint32) |  | NewScore is the new identity score (0-100) |
 | `score_version` | [string](#string) |  | ScoreVersion is the ML model version used |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateScoreResponse"></a>

 ### MsgUpdateScoreResponse
 MsgUpdateScoreResponse is the response for MsgUpdateScore

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `account_address` | [string](#string) |  | AccountAddress is the account whose score was updated |
 | `previous_score` | [uint32](#uint32) |  | PreviousScore is the previous identity score |
 | `new_score` | [uint32](#uint32) |  | NewScore is the new identity score |
 | `previous_tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | PreviousTier is the previous identity tier |
 | `new_tier` | [IdentityTier](#virtengine.veid.v1.IdentityTier) |  | NewTier is the new identity tier |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when the score was updated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateVerificationStatus"></a>

 ### MsgUpdateVerificationStatus
 MsgUpdateVerificationStatus is the message for validators to update verification status

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the validator updating the status |
 | `account_address` | [string](#string) |  | AccountAddress is the account whose scope is being updated |
 | `scope_id` | [string](#string) |  | ScopeID is the unique identifier of the scope |
 | `new_status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | NewStatus is the new verification status |
 | `reason` | [string](#string) |  | Reason is the reason for the status update |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUpdateVerificationStatusResponse"></a>

 ### MsgUpdateVerificationStatusResponse
 MsgUpdateVerificationStatusResponse is the response for MsgUpdateVerificationStatus

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the ID of the scope |
 | `previous_status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | PreviousStatus is the previous verification status |
 | `new_status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | NewStatus is the new verification status |
 | `updated_at` | [int64](#int64) |  | UpdatedAt is when the status was updated (Unix timestamp) |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUploadScope"></a>

 ### MsgUploadScope
 MsgUploadScope is the message for uploading an identity scope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `sender` | [string](#string) |  | Sender is the account uploading the scope |
 | `scope_id` | [string](#string) |  | ScopeID is a unique identifier for this scope |
 | `scope_type` | [ScopeType](#virtengine.veid.v1.ScopeType) |  | ScopeType indicates what kind of identity data this scope contains |
 | `encrypted_payload` | [EncryptedPayloadEnvelope](#virtengine.veid.v1.EncryptedPayloadEnvelope) |  | EncryptedPayload is the encrypted identity data |
 | `salt` | [bytes](#bytes) |  | Salt is a per-upload unique salt for cryptographic binding |
 | `device_fingerprint` | [string](#string) |  | DeviceFingerprint is the device that captured/uploaded this data |
 | `client_id` | [string](#string) |  | ClientID is the approved client that facilitated this upload |
 | `client_signature` | [bytes](#bytes) |  | ClientSignature is the signature from the approved client |
 | `user_signature` | [bytes](#bytes) |  | UserSignature is the signature from the user authorizing this upload |
 | `payload_hash` | [bytes](#bytes) |  | PayloadHash is the hash of the encrypted payload for integrity |
 | `capture_timestamp` | [int64](#int64) |  | CaptureTimestamp is when the data was captured (Unix timestamp) |
 | `geo_hint` | [string](#string) |  | GeoHint is an optional coarse geographic hint |
 
 

 

 
 <a name="virtengine.veid.v1.MsgUploadScopeResponse"></a>

 ### MsgUploadScopeResponse
 MsgUploadScopeResponse is the response for MsgUploadScope

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `scope_id` | [string](#string) |  | ScopeID is the ID of the uploaded scope |
 | `status` | [VerificationStatus](#virtengine.veid.v1.VerificationStatus) |  | Status is the initial verification status |
 | `uploaded_at` | [int64](#int64) |  | UploadedAt is when the scope was uploaded (Unix timestamp) |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.veid.v1.Msg"></a>

 ### Msg
 Msg defines the veid Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `UploadScope` | [MsgUploadScope](#virtengine.veid.v1.MsgUploadScope) | [MsgUploadScopeResponse](#virtengine.veid.v1.MsgUploadScopeResponse) | UploadScope uploads an identity scope | |
 | `RevokeScope` | [MsgRevokeScope](#virtengine.veid.v1.MsgRevokeScope) | [MsgRevokeScopeResponse](#virtengine.veid.v1.MsgRevokeScopeResponse) | RevokeScope revokes an identity scope | |
 | `RequestVerification` | [MsgRequestVerification](#virtengine.veid.v1.MsgRequestVerification) | [MsgRequestVerificationResponse](#virtengine.veid.v1.MsgRequestVerificationResponse) | RequestVerification requests verification of a scope | |
 | `UpdateVerificationStatus` | [MsgUpdateVerificationStatus](#virtengine.veid.v1.MsgUpdateVerificationStatus) | [MsgUpdateVerificationStatusResponse](#virtengine.veid.v1.MsgUpdateVerificationStatusResponse) | UpdateVerificationStatus updates the verification status (validator only) | |
 | `UpdateScore` | [MsgUpdateScore](#virtengine.veid.v1.MsgUpdateScore) | [MsgUpdateScoreResponse](#virtengine.veid.v1.MsgUpdateScoreResponse) | UpdateScore updates the identity score (validator only) | |
 | `CreateIdentityWallet` | [MsgCreateIdentityWallet](#virtengine.veid.v1.MsgCreateIdentityWallet) | [MsgCreateIdentityWalletResponse](#virtengine.veid.v1.MsgCreateIdentityWalletResponse) | CreateIdentityWallet creates an identity wallet | |
 | `AddScopeToWallet` | [MsgAddScopeToWallet](#virtengine.veid.v1.MsgAddScopeToWallet) | [MsgAddScopeToWalletResponse](#virtengine.veid.v1.MsgAddScopeToWalletResponse) | AddScopeToWallet adds a scope reference to a wallet | |
 | `RevokeScopeFromWallet` | [MsgRevokeScopeFromWallet](#virtengine.veid.v1.MsgRevokeScopeFromWallet) | [MsgRevokeScopeFromWalletResponse](#virtengine.veid.v1.MsgRevokeScopeFromWalletResponse) | RevokeScopeFromWallet revokes a scope from a wallet | |
 | `UpdateConsentSettings` | [MsgUpdateConsentSettings](#virtengine.veid.v1.MsgUpdateConsentSettings) | [MsgUpdateConsentSettingsResponse](#virtengine.veid.v1.MsgUpdateConsentSettingsResponse) | UpdateConsentSettings updates consent settings | |
 | `RebindWallet` | [MsgRebindWallet](#virtengine.veid.v1.MsgRebindWallet) | [MsgRebindWalletResponse](#virtengine.veid.v1.MsgRebindWalletResponse) | RebindWallet rebinds a wallet during key rotation | |
 | `UpdateDerivedFeatures` | [MsgUpdateDerivedFeatures](#virtengine.veid.v1.MsgUpdateDerivedFeatures) | [MsgUpdateDerivedFeaturesResponse](#virtengine.veid.v1.MsgUpdateDerivedFeaturesResponse) | UpdateDerivedFeatures updates derived features (validator only) | |
 | `CompleteBorderlineFallback` | [MsgCompleteBorderlineFallback](#virtengine.veid.v1.MsgCompleteBorderlineFallback) | [MsgCompleteBorderlineFallbackResponse](#virtengine.veid.v1.MsgCompleteBorderlineFallbackResponse) | CompleteBorderlineFallback completes a borderline fallback after MFA | |
 | `UpdateBorderlineParams` | [MsgUpdateBorderlineParams](#virtengine.veid.v1.MsgUpdateBorderlineParams) | [MsgUpdateBorderlineParamsResponse](#virtengine.veid.v1.MsgUpdateBorderlineParamsResponse) | UpdateBorderlineParams updates borderline parameters (governance only) | |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.veid.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.veid.v1.MsgUpdateParamsResponse) | UpdateParams updates module parameters (governance only) | |
 | `SubmitAppeal` | [MsgSubmitAppeal](#virtengine.veid.v1.MsgSubmitAppeal) | [MsgSubmitAppealResponse](#virtengine.veid.v1.MsgSubmitAppealResponse) | SubmitAppeal submits an appeal against a verification decision | |
 | `ClaimAppeal` | [MsgClaimAppeal](#virtengine.veid.v1.MsgClaimAppeal) | [MsgClaimAppealResponse](#virtengine.veid.v1.MsgClaimAppealResponse) | ClaimAppeal allows an arbitrator to claim an appeal for review | |
 | `ResolveAppeal` | [MsgResolveAppeal](#virtengine.veid.v1.MsgResolveAppeal) | [MsgResolveAppealResponse](#virtengine.veid.v1.MsgResolveAppealResponse) | ResolveAppeal resolves an appeal (arbitrator/governance only) | |
 | `WithdrawAppeal` | [MsgWithdrawAppeal](#virtengine.veid.v1.MsgWithdrawAppeal) | [MsgWithdrawAppealResponse](#virtengine.veid.v1.MsgWithdrawAppealResponse) | WithdrawAppeal allows the submitter to withdraw their appeal | |
 | `SubmitComplianceCheck` | [MsgSubmitComplianceCheck](#virtengine.veid.v1.MsgSubmitComplianceCheck) | [MsgSubmitComplianceCheckResponse](#virtengine.veid.v1.MsgSubmitComplianceCheckResponse) | SubmitComplianceCheck submits external compliance check results | |
 | `AttestCompliance` | [MsgAttestCompliance](#virtengine.veid.v1.MsgAttestCompliance) | [MsgAttestComplianceResponse](#virtengine.veid.v1.MsgAttestComplianceResponse) | AttestCompliance allows validators to attest compliance status | |
 | `UpdateComplianceParams` | [MsgUpdateComplianceParams](#virtengine.veid.v1.MsgUpdateComplianceParams) | [MsgUpdateComplianceParamsResponse](#virtengine.veid.v1.MsgUpdateComplianceParamsResponse) | UpdateComplianceParams updates compliance configuration (gov only) | |
 | `RegisterComplianceProvider` | [MsgRegisterComplianceProvider](#virtengine.veid.v1.MsgRegisterComplianceProvider) | [MsgRegisterComplianceProviderResponse](#virtengine.veid.v1.MsgRegisterComplianceProviderResponse) | RegisterComplianceProvider registers a new compliance provider (gov only) | |
 | `DeactivateComplianceProvider` | [MsgDeactivateComplianceProvider](#virtengine.veid.v1.MsgDeactivateComplianceProvider) | [MsgDeactivateComplianceProviderResponse](#virtengine.veid.v1.MsgDeactivateComplianceProviderResponse) | DeactivateComplianceProvider deactivates a compliance provider (gov only) | |
 | `RegisterModel` | [MsgRegisterModel](#virtengine.veid.v1.MsgRegisterModel) | [MsgRegisterModelResponse](#virtengine.veid.v1.MsgRegisterModelResponse) | RegisterModel registers a new ML model (authorized only) | |
 | `ProposeModelUpdate` | [MsgProposeModelUpdate](#virtengine.veid.v1.MsgProposeModelUpdate) | [MsgProposeModelUpdateResponse](#virtengine.veid.v1.MsgProposeModelUpdateResponse) | ProposeModelUpdate proposes updating active model via governance | |
 | `ReportModelVersion` | [MsgReportModelVersion](#virtengine.veid.v1.MsgReportModelVersion) | [MsgReportModelVersionResponse](#virtengine.veid.v1.MsgReportModelVersionResponse) | ReportModelVersion reports validator's model versions | |
 | `ActivateModel` | [MsgActivateModel](#virtengine.veid.v1.MsgActivateModel) | [MsgActivateModelResponse](#virtengine.veid.v1.MsgActivateModelResponse) | ActivateModel activates a pending model after governance approval | |
 | `DeprecateModel` | [MsgDeprecateModel](#virtengine.veid.v1.MsgDeprecateModel) | [MsgDeprecateModelResponse](#virtengine.veid.v1.MsgDeprecateModelResponse) | DeprecateModel deprecates a model | |
 | `RevokeModel` | [MsgRevokeModel](#virtengine.veid.v1.MsgRevokeModel) | [MsgRevokeModelResponse](#virtengine.veid.v1.MsgRevokeModelResponse) | RevokeModel revokes a model | |
 
  <!-- end services -->

 
 
 <a name="virtengine/wasm/v1/event.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/wasm/v1/event.proto
 

 
 <a name="virtengine.wasm.v1.EventMsgBlocked"></a>

 ### EventMsgBlocked
 EventMsgBlocked is triggered when smart contract does not
pass message filter

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `contract_address` | [string](#string) |  |  |
 | `msg_type` | [string](#string) |  |  |
 | `reason` | [string](#string) |  |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/wasm/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/wasm/v1/params.proto
 

 
 <a name="virtengine.wasm.v1.Params"></a>

 ### Params
 Params defines the parameters for the x/wasm package.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `blocked_addresses` | [string](#string) | repeated |  |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/wasm/v1/genesis.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/wasm/v1/genesis.proto
 

 
 <a name="virtengine.wasm.v1.GenesisState"></a>

 ### GenesisState
 GenesisState stores slice of genesis wasm parameters.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.wasm.v1.Params) |  | Params holds parameters of the genesis of virtengine wasm. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/wasm/v1/paramsmsg.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/wasm/v1/paramsmsg.proto
 

 
 <a name="virtengine.wasm.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the Msg/UpdateParams request type.

Since: akash v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.wasm.v1.Params) |  | params defines the x/wasm parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.wasm.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: akash v1.0.0

 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

  <!-- end services -->

 
 
 <a name="virtengine/wasm/v1/query.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/wasm/v1/query.proto
 

 
 <a name="virtengine.wasm.v1.QueryParamsRequest"></a>

 ### QueryParamsRequest
 QueryParamsRequest is the request type for the Query/Params RPC method.

 

 

 
 <a name="virtengine.wasm.v1.QueryParamsResponse"></a>

 ### QueryParamsResponse
 QueryParamsResponse is the response type for the Query/Params RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `params` | [Params](#virtengine.wasm.v1.Params) |  | params defines the parameters of the module. |
 
 

 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.wasm.v1.Query"></a>

 ### Query
 Query defines the gRPC querier service of the wasm package.

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `Params` | [QueryParamsRequest](#virtengine.wasm.v1.QueryParamsRequest) | [QueryParamsResponse](#virtengine.wasm.v1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/virtengine/wasm/v1/params|
 
  <!-- end services -->

 
 
 <a name="virtengine/wasm/v1/service.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/wasm/v1/service.proto
 

  <!-- end messages -->

  <!-- end enums -->

  <!-- end HasExtensions -->

 
 <a name="virtengine.wasm.v1.Msg"></a>

 ### Msg
 Msg defines the wasm Msg service

 | Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
 | ----------- | ------------ | ------------- | ------------| ------- | -------- |
 | `UpdateParams` | [MsgUpdateParams](#virtengine.wasm.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#virtengine.wasm.v1.MsgUpdateParamsResponse) | UpdateParams defines a governance operation for updating the x/wasm module parameters. The authority is hard-coded to the x/gov module account.

Since: akash v2.0.0 | |
 
  <!-- end services -->

 

 ## Scalar Value Types

 | .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
 | ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
 | <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
 | <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
 | <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
 | <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
 | <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
 | <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
 | <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
 | <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
 | <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
 | <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
 | <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
 | <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
 | <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
 | <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
 | <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
 

