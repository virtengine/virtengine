import { patched } from "./nodePatchMessage.ts";

export { Attribute, SignedBy, PlacementRequirements } from "./virtengine/base/attributes/v1/attribute.ts";
export { AuditedProvider, AuditedAttributesStore, AttributesFilters } from "./virtengine/audit/v1/audit.ts";
export { EventTrustedAuditorCreated, EventTrustedAuditorDeleted } from "./virtengine/audit/v1/event.ts";
export { GenesisState } from "./virtengine/audit/v1/genesis.ts";
export { MsgSignProviderAttributes, MsgSignProviderAttributesResponse, MsgDeleteProviderAttributes, MsgDeleteProviderAttributesResponse } from "./virtengine/audit/v1/msg.ts";
export { QueryProvidersResponse, QueryProviderRequest, QueryAllProvidersAttributesRequest, QueryProviderAttributesRequest, QueryProviderAuditorRequest, QueryAuditorAttributesRequest } from "./virtengine/audit/v1/query.ts";
export { Deposit, Source } from "./virtengine/base/deposit/v1/deposit.ts";
export { MsgSignData } from "./virtengine/base/offchain/sign/v1/sign.ts";
export { LedgerID, State, LedgerRecordID, LedgerPendingRecord, Status, MintEpoch, MintStatus, LedgerRecordStatus } from "./virtengine/bme/v1/types.ts";

import { CollateralRatio as _CollateralRatio, CoinPrice as _CoinPrice, BurnMintPair as _BurnMintPair, LedgerRecord as _LedgerRecord } from "./virtengine/bme/v1/types.ts";
export const CollateralRatio = patched(_CollateralRatio);
export type CollateralRatio = _CollateralRatio
export const CoinPrice = patched(_CoinPrice);
export type CoinPrice = _CoinPrice
export const BurnMintPair = patched(_BurnMintPair);
export type BurnMintPair = _BurnMintPair
export const LedgerRecord = patched(_LedgerRecord);
export type LedgerRecord = _LedgerRecord
export { EventVaultSeeded, EventLedgerRecordExecuted } from "./virtengine/bme/v1/events.ts";

import { EventMintStatusChange as _EventMintStatusChange } from "./virtengine/bme/v1/events.ts";
export const EventMintStatusChange = patched(_EventMintStatusChange);
export type EventMintStatusChange = _EventMintStatusChange
export { Params } from "./virtengine/bme/v1/params.ts";
export { GenesisLedgerPendingRecord, GenesisVaultState } from "./virtengine/bme/v1/genesis.ts";

import { GenesisLedgerRecord as _GenesisLedgerRecord, GenesisLedgerState as _GenesisLedgerState, GenesisState as _Bme_GenesisState } from "./virtengine/bme/v1/genesis.ts";
export const GenesisLedgerRecord = patched(_GenesisLedgerRecord);
export type GenesisLedgerRecord = _GenesisLedgerRecord
export const GenesisLedgerState = patched(_GenesisLedgerState);
export type GenesisLedgerState = _GenesisLedgerState
export const Bme_GenesisState = patched(_Bme_GenesisState);
export type Bme_GenesisState = _Bme_GenesisState
export { MsgUpdateParams, MsgUpdateParamsResponse, MsgSeedVault, MsgSeedVaultResponse, MsgBurnMint, MsgMintACT, MsgBurnACT, MsgBurnMintResponse, MsgMintACTResponse, MsgBurnACTResponse } from "./virtengine/bme/v1/msgs.ts";
export { QueryParamsRequest, QueryParamsResponse, QueryVaultStateRequest, QueryVaultStateResponse, QueryStatusRequest } from "./virtengine/bme/v1/query.ts";

import { QueryStatusResponse as _QueryStatusResponse } from "./virtengine/bme/v1/query.ts";
export const QueryStatusResponse = patched(_QueryStatusResponse);
export type QueryStatusResponse = _QueryStatusResponse
export { ID, Certificate, State as Cert_State } from "./virtengine/cert/v1/cert.ts";
export { CertificateFilter } from "./virtengine/cert/v1/filters.ts";
export { GenesisCertificate, GenesisState as Cert_GenesisState } from "./virtengine/cert/v1/genesis.ts";
export { MsgCreateCertificate, MsgCreateCertificateResponse, MsgRevokeCertificate, MsgRevokeCertificateResponse } from "./virtengine/cert/v1/msg.ts";
export { CertificateResponse, QueryCertificatesRequest, QueryCertificatesResponse } from "./virtengine/cert/v1/query.ts";
export { DeploymentID, Deployment, Deployment_State } from "./virtengine/deployment/v1/deployment.ts";
export { GroupID } from "./virtengine/deployment/v1/group.ts";
export { EventDeploymentCreated, EventDeploymentUpdated, EventDeploymentClosed, EventGroupStarted, EventGroupPaused, EventGroupClosed } from "./virtengine/deployment/v1/event.ts";
export { Account, Payment, Scope } from "./virtengine/escrow/id/v1/id.ts";

import { Balance as _Balance } from "./virtengine/escrow/types/v1/balance.ts";
export const Balance = patched(_Balance);
export type Balance = _Balance

import { Depositor as _Depositor } from "./virtengine/escrow/types/v1/deposit.ts";
export const Depositor = patched(_Depositor);
export type Depositor = _Depositor
export { State as Types_State } from "./virtengine/escrow/types/v1/state.ts";

import { AccountState as _AccountState, Account as _Types_Account } from "./virtengine/escrow/types/v1/account.ts";
export const AccountState = patched(_AccountState);
export type AccountState = _AccountState
export const Types_Account = patched(_Types_Account);
export type Types_Account = _Types_Account
export { ClientInfo } from "./virtengine/discovery/v1/client_info.ts";
export { Akash } from "./virtengine/discovery/v1/akash.ts";

import { PaymentState as _PaymentState, Payment as _Types_Payment } from "./virtengine/escrow/types/v1/payment.ts";
export const PaymentState = patched(_PaymentState);
export type PaymentState = _PaymentState
export const Types_Payment = patched(_Types_Payment);
export type Types_Payment = _Types_Payment
export { DepositAuthorization, DepositAuthorization_Scope } from "./virtengine/escrow/v1/authz.ts";

import { GenesisState as _Escrow_GenesisState } from "./virtengine/escrow/v1/genesis.ts";
export const Escrow_GenesisState = patched(_Escrow_GenesisState);
export type Escrow_GenesisState = _Escrow_GenesisState
export { MsgAccountDeposit, MsgAccountDepositResponse } from "./virtengine/escrow/v1/msg.ts";
export { QueryAccountsRequest, QueryPaymentsRequest } from "./virtengine/escrow/v1/query.ts";

import { QueryAccountsResponse as _QueryAccountsResponse, QueryPaymentsResponse as _QueryPaymentsResponse } from "./virtengine/escrow/v1/query.ts";
export const QueryAccountsResponse = patched(_QueryAccountsResponse);
export type QueryAccountsResponse = _QueryAccountsResponse
export const QueryPaymentsResponse = patched(_QueryPaymentsResponse);
export type QueryPaymentsResponse = _QueryPaymentsResponse
export { BidID } from "./virtengine/market/v1/bid.ts";
export { OrderID } from "./virtengine/market/v1/order.ts";
export { LeaseClosedReason } from "./virtengine/market/v1/types.ts";
export { LeaseID, Lease_State } from "./virtengine/market/v1/lease.ts";

import { Lease as _Lease } from "./virtengine/market/v1/lease.ts";
export const Lease = patched(_Lease);
export type Lease = _Lease
export { EventOrderCreated, EventOrderClosed, EventBidClosed, EventLeaseClosed } from "./virtengine/market/v1/event.ts";

import { EventBidCreated as _EventBidCreated, EventLeaseCreated as _EventLeaseCreated } from "./virtengine/market/v1/event.ts";
export const EventBidCreated = patched(_EventBidCreated);
export type EventBidCreated = _EventBidCreated
export const EventLeaseCreated = patched(_EventLeaseCreated);
export type EventLeaseCreated = _EventLeaseCreated
export { LeaseFilters } from "./virtengine/market/v1/filters.ts";
export { DataID, PriceDataID, PriceDataRecordID, PriceHealth, PricesFilter, QueryPricesRequest } from "./virtengine/oracle/v1/prices.ts";

import { PriceDataState as _PriceDataState, PriceData as _PriceData, AggregatedPrice as _AggregatedPrice, QueryPricesResponse as _QueryPricesResponse } from "./virtengine/oracle/v1/prices.ts";
export const PriceDataState = patched(_PriceDataState);
export type PriceDataState = _PriceDataState
export const PriceData = patched(_PriceData);
export type PriceData = _PriceData
export const AggregatedPrice = patched(_AggregatedPrice);
export type AggregatedPrice = _AggregatedPrice
export const QueryPricesResponse = patched(_QueryPricesResponse);
export type QueryPricesResponse = _QueryPricesResponse
export { EventPriceStaleWarning, EventPriceStaled, EventPriceRecovered } from "./virtengine/oracle/v1/events.ts";

import { EventPriceData as _EventPriceData } from "./virtengine/oracle/v1/events.ts";
export const EventPriceData = patched(_EventPriceData);
export type EventPriceData = _EventPriceData
export { PythContractParams, Params as Oracle_Params } from "./virtengine/oracle/v1/params.ts";

import { GenesisState as _Oracle_GenesisState } from "./virtengine/oracle/v1/genesis.ts";
export const Oracle_GenesisState = patched(_Oracle_GenesisState);
export type Oracle_GenesisState = _Oracle_GenesisState
export { MsgAddPriceEntryResponse, MsgUpdateParams as Oracle_MsgUpdateParams, MsgUpdateParamsResponse as Oracle_MsgUpdateParamsResponse } from "./virtengine/oracle/v1/msgs.ts";

import { MsgAddPriceEntry as _MsgAddPriceEntry } from "./virtengine/oracle/v1/msgs.ts";
export const MsgAddPriceEntry = patched(_MsgAddPriceEntry);
export type MsgAddPriceEntry = _MsgAddPriceEntry
export { QueryParamsRequest as Oracle_QueryParamsRequest, QueryParamsResponse as Oracle_QueryParamsResponse, QueryPriceFeedConfigRequest, QueryPriceFeedConfigResponse, QueryAggregatedPriceRequest } from "./virtengine/oracle/v1/query.ts";

import { QueryAggregatedPriceResponse as _QueryAggregatedPriceResponse } from "./virtengine/oracle/v1/query.ts";
export const QueryAggregatedPriceResponse = patched(_QueryAggregatedPriceResponse);
export type QueryAggregatedPriceResponse = _QueryAggregatedPriceResponse
export { DenomTakeRate, Params as Take_Params } from "./virtengine/take/v1/params.ts";
export { GenesisState as Take_GenesisState } from "./virtengine/take/v1/genesis.ts";
export { MsgUpdateParams as Take_MsgUpdateParams, MsgUpdateParamsResponse as Take_MsgUpdateParamsResponse } from "./virtengine/take/v1/paramsmsg.ts";
export { QueryParamsRequest as Take_QueryParamsRequest, QueryParamsResponse as Take_QueryParamsResponse } from "./virtengine/take/v1/query.ts";
export { EventMsgBlocked } from "./virtengine/wasm/v1/event.ts";
export { Params as Wasm_Params } from "./virtengine/wasm/v1/params.ts";
export { GenesisState as Wasm_GenesisState } from "./virtengine/wasm/v1/genesis.ts";
export { MsgUpdateParams as Wasm_MsgUpdateParams, MsgUpdateParamsResponse as Wasm_MsgUpdateParamsResponse } from "./virtengine/wasm/v1/paramsmsg.ts";
export { QueryParamsRequest as Wasm_QueryParamsRequest, QueryParamsResponse as Wasm_QueryParamsResponse } from "./virtengine/wasm/v1/query.ts";
