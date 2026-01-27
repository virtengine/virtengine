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
     - [MsgUpdateProvider](#virtengine.provider.v1beta4.MsgUpdateProvider)
     - [MsgUpdateProviderResponse](#virtengine.provider.v1beta4.MsgUpdateProviderResponse)
   
 - [virtengine/provider/v1beta4/query.proto](#virtengine/provider/v1beta4/query.proto)
     - [QueryProviderRequest](#virtengine.provider.v1beta4.QueryProviderRequest)
     - [QueryProviderResponse](#virtengine.provider.v1beta4.QueryProviderResponse)
     - [QueryProvidersRequest](#virtengine.provider.v1beta4.QueryProvidersRequest)
     - [QueryProvidersResponse](#virtengine.provider.v1beta4.QueryProvidersResponse)
   
     - [Query](#virtengine.provider.v1beta4.Query)
   
 - [virtengine/provider/v1beta4/service.proto](#virtengine/provider/v1beta4/service.proto)
     - [Msg](#virtengine.provider.v1beta4.Msg)
   
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

Example: "virt1..." |
 | `auditor` | [string](#string) |  | Auditor is the account bech32 address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
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

Example: "virt1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
 
 

 

 
 <a name="virtengine.audit.v1.EventTrustedAuditorDeleted"></a>

 ### EventTrustedAuditorDeleted
 EventTrustedAuditorDeleted defines an event for when a trusted auditor is deleted.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "virt1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
 
 

 

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

Example: "virt1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
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

Example: "virt1..." |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
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

Example: "virt1..." |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProviderAttributesRequest"></a>

 ### QueryProviderAttributesRequest
 QueryProviderAttributesRequest is request type for the Query/Provider RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "virt1..." |
 | `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination is used to paginate the request. |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProviderAuditorRequest"></a>

 ### QueryProviderAuditorRequest
 QueryProviderAuditorRequest is request type for the Query/Providers RPC method.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "virt1..." |
 
 

 

 
 <a name="virtengine.audit.v1.QueryProviderRequest"></a>

 ### QueryProviderRequest
 QueryProviderRequest is request type for the Query/Provider RPC method

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `auditor` | [string](#string) |  | Auditor is the account address of the auditor. It is a string representing a valid account address.

Example: "virt1..." |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the provider. It is a string representing a valid account address.

Example: "virt1..." |
 
 

 

 
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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `state` | [string](#string) |  | State defines the sate of the deployment. A deployment can be either active or inactive. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.GroupFilters"></a>

 ### GroupFilters
 GroupFilters defines filters used to filter groups

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account address of the user who owns the group. It is a string representing a valid account address.

Example: "virt1..." |
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

Since: VirtEngine v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | Authority is the address of the governance account. |
 | `params` | [Params](#virtengine.deployment.v1beta4.Params) |  | Params defines the x/deployment parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.deployment.v1beta4.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: VirtEngine v1.0.0

 

 

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

Example: "virt1..." If depositor is same as the owner, then any incoming coins are added to the Balance. If depositor isn't same as the owner, then any incoming coins are added to the Funds. |
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

Example: "virt1..." |
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

Since: VirtEngine v1.0.0 | |
 
  <!-- end services -->

 
 
 <a name="virtengine/discovery/v1/client_info.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/discovery/v1/client_info.proto
 

 
 <a name="virtengine.discovery.v1.ClientInfo"></a>

 ### ClientInfo
 ClientInfo is the VirtEngine specific client info.

 
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
 VirtEngine VirtEngine specific RPC parameters.

 
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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "virt1..." |
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

Example: "virt1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "virt1..." |
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

Example: "virt1..." |
 | `dseq` | [uint64](#uint64) |  | Dseq (deployment sequence number) is a unique numeric identifier for the deployment. It is used to differentiate deployments created by the same owner. |
 | `gseq` | [uint32](#uint32) |  | Gseq (group sequence number) is a unique numeric identifier for the group. It is used to differentiate groups created by the same owner in a deployment. |
 | `oseq` | [uint32](#uint32) |  | Oseq (order sequence) distinguishes multiple orders associated with a single deployment. Oseq is incremented when a lease associated with an existing deployment is closed, and a new order is generated. |
 | `provider` | [string](#string) |  | Provider is the account bech32 address of the provider making the bid. It is a string representing a valid account bech32 address.

Example: "virt1..." |
 | `state` | [string](#string) |  | State represents the state of the lease. |
 | `bseq` | [uint32](#uint32) |  | BSeq (bid sequence) distinguishes multiple bids associated with a single deployment from same provider. |
 
 

 

 
 <a name="virtengine.market.v1beta5.OrderFilters"></a>

 ### OrderFilters
 OrderFilters defines flags for order list filter

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the account bech32 address of the user who owns the deployment. It is a string representing a valid bech32 account address.

Example: "virt1..." |
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

Since: VirtEngine v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.market.v1beta5.Params) |  | params defines the x/deployment parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.market.v1beta5.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: VirtEngine v1.0.0

 

 

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

Since: VirtEngine v1.0.0 | |
 
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
 | `denom` | [string](#string) |  | denom is the asset denomination (e.g., "uakt") |
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

Example: "virt1..." |
 | `id` | [DataID](#virtengine.oracle.v1.DataID) |  | id uniquely identifies the price data by denomination and base denomination |
 | `price` | [PriceDataState](#virtengine.oracle.v1.PriceDataState) |  | price contains the price value and timestamp for this entry |
 
 

 

 
 <a name="virtengine.oracle.v1.MsgAddPriceEntryResponse"></a>

 ### MsgAddPriceEntryResponse
 MsgAddPriceEntryResponse defines the Msg/MsgAddDPriceEntry response type.

 

 

 
 <a name="virtengine.oracle.v1.MsgUpdateParams"></a>

 ### MsgUpdateParams
 MsgUpdateParams is the Msg/UpdateParams request type.

Since: VirtEngine v2.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.oracle.v1.Params) |  | params defines the x/oracle parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.oracle.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: VirtEngine v2.0.0

 

 

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

Since: VirtEngine v2.0.0 | |
 
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

Example: "virt1..." |
 
 

 

 
 <a name="virtengine.provider.v1beta4.EventProviderDeleted"></a>

 ### EventProviderDeleted
 EventProviderDeleted defines an SDK message for provider deleted event.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "virt1..." |
 
 

 

 
 <a name="virtengine.provider.v1beta4.EventProviderUpdated"></a>

 ### EventProviderUpdated
 EventProviderUpdated defines an SDK message for provider updated event.
It contains all the required information to identify a provider on-chain.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `owner` | [string](#string) |  | Owner is the bech32 address of the account of the provider. It is a string representing a valid account address.

Example: "virt1..." |
 
 

 

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

Example: "virt1..." |
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

Example: "virt1..." |
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

Example: "virt1..." |
 
 

 

 
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
 
  <!-- end services -->

 
 
 <a name="virtengine/take/v1/params.proto"></a>
 <p align="right"><a href="#top">Top</a></p>

 ## virtengine/take/v1/params.proto
 

 
 <a name="virtengine.take.v1.DenomTakeRate"></a>

 ### DenomTakeRate
 DenomTakeRate describes take rate for specified denom.

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `denom` | [string](#string) |  | Denom is the denomination of the take rate (uakt, usdc, etc.). |
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

Since: VirtEngine v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.take.v1.Params) |  | params defines the x/deployment parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.take.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: VirtEngine v1.0.0

 

 

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

Since: VirtEngine v1.0.0 | |
 
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
 | `params` | [Params](#virtengine.wasm.v1.Params) |  | Params holds parameters of the genesis of VirtEngine wasm. |
 
 

 

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

Since: VirtEngine v1.0.0

 
 | Field | Type | Label | Description |
 | ----- | ---- | ----- | ----------- |
 | `authority` | [string](#string) |  | authority is the address of the governance account. |
 | `params` | [Params](#virtengine.wasm.v1.Params) |  | params defines the x/wasm parameters to update.

NOTE: All parameters must be supplied. |
 
 

 

 
 <a name="virtengine.wasm.v1.MsgUpdateParamsResponse"></a>

 ### MsgUpdateParamsResponse
 MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: VirtEngine v1.0.0

 

 

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

Since: VirtEngine v2.0.0 | |
 
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
 
