import { patched } from "./nodePatchMessage.ts";

export { ResourceValue } from "./virtengine/base/resources/v1beta4/resourcevalue.ts";
export { CPU } from "./virtengine/base/resources/v1beta4/cpu.ts";
export { Endpoint, Endpoint_Kind } from "./virtengine/base/resources/v1beta4/endpoint.ts";
export { GPU } from "./virtengine/base/resources/v1beta4/gpu.ts";
export { Memory } from "./virtengine/base/resources/v1beta4/memory.ts";
export { Storage } from "./virtengine/base/resources/v1beta4/storage.ts";
export { Resources } from "./virtengine/base/resources/v1beta4/resources.ts";

import { ResourceUnit as _ResourceUnit } from "./virtengine/deployment/v1beta4/resourceunit.ts";
export const ResourceUnit = patched(_ResourceUnit);
export type ResourceUnit = _ResourceUnit

import { GroupSpec as _GroupSpec } from "./virtengine/deployment/v1beta4/groupspec.ts";
export const GroupSpec = patched(_GroupSpec);
export type GroupSpec = _GroupSpec
export { MsgCreateDeploymentResponse, MsgUpdateDeployment, MsgUpdateDeploymentResponse, MsgCloseDeployment, MsgCloseDeploymentResponse } from "./virtengine/deployment/v1beta4/deploymentmsg.ts";

import { MsgCreateDeployment as _MsgCreateDeployment } from "./virtengine/deployment/v1beta4/deploymentmsg.ts";
export const MsgCreateDeployment = patched(_MsgCreateDeployment);
export type MsgCreateDeployment = _MsgCreateDeployment
export { DeploymentFilters, GroupFilters } from "./virtengine/deployment/v1beta4/filters.ts";
export { Group_State } from "./virtengine/deployment/v1beta4/group.ts";

import { Group as _Group } from "./virtengine/deployment/v1beta4/group.ts";
export const Group = patched(_Group);
export type Group = _Group
export { Params } from "./virtengine/deployment/v1beta4/params.ts";

import { GenesisDeployment as _GenesisDeployment, GenesisState as _GenesisState } from "./virtengine/deployment/v1beta4/genesis.ts";
export const GenesisDeployment = patched(_GenesisDeployment);
export type GenesisDeployment = _GenesisDeployment
export const GenesisState = patched(_GenesisState);
export type GenesisState = _GenesisState
export { MsgCloseGroup, MsgCloseGroupResponse, MsgPauseGroup, MsgPauseGroupResponse, MsgStartGroup, MsgStartGroupResponse } from "./virtengine/deployment/v1beta4/groupmsg.ts";
export { MsgUpdateParams, MsgUpdateParamsResponse } from "./virtengine/deployment/v1beta4/paramsmsg.ts";
export { QueryDeploymentsRequest, QueryDeploymentRequest, QueryGroupRequest, QueryParamsRequest, QueryParamsResponse } from "./virtengine/deployment/v1beta4/query.ts";

import { QueryDeploymentsResponse as _QueryDeploymentsResponse, QueryDeploymentResponse as _QueryDeploymentResponse, QueryGroupResponse as _QueryGroupResponse } from "./virtengine/deployment/v1beta4/query.ts";
export const QueryDeploymentsResponse = patched(_QueryDeploymentsResponse);
export type QueryDeploymentsResponse = _QueryDeploymentsResponse
export const QueryDeploymentResponse = patched(_QueryDeploymentResponse);
export type QueryDeploymentResponse = _QueryDeploymentResponse
export const QueryGroupResponse = patched(_QueryGroupResponse);
export type QueryGroupResponse = _QueryGroupResponse
export { OrderID, Order_State, OrderFilters } from "./virtengine/market/v1beta4/order.ts";

import { Order as _Order } from "./virtengine/market/v1beta4/order.ts";
export const Order = patched(_Order);
export type Order = _Order
export { ResourceOffer, MsgCreateBidResponse, MsgCloseBid, MsgCloseBidResponse, BidID, Bid_State, BidFilters } from "./virtengine/market/v1beta4/bid.ts";

import { MsgCreateBid as _MsgCreateBid, Bid as _Bid } from "./virtengine/market/v1beta4/bid.ts";
export const MsgCreateBid = patched(_MsgCreateBid);
export type MsgCreateBid = _MsgCreateBid
export const Bid = patched(_Bid);
export type Bid = _Bid
export { LeaseID, Lease_State, LeaseFilters, MsgCreateLease, MsgCreateLeaseResponse, MsgWithdrawLease, MsgWithdrawLeaseResponse, MsgCloseLease, MsgCloseLeaseResponse } from "./virtengine/market/v1beta4/lease.ts";

import { Lease as _Lease } from "./virtengine/market/v1beta4/lease.ts";
export const Lease = patched(_Lease);
export type Lease = _Lease
export { Params as Market_Params } from "./virtengine/market/v1beta4/params.ts";

import { GenesisState as _Market_GenesisState } from "./virtengine/market/v1beta4/genesis.ts";
export const Market_GenesisState = patched(_Market_GenesisState);
export type Market_GenesisState = _Market_GenesisState
export { EventProviderCreated, EventProviderUpdated, EventProviderDeleted } from "./virtengine/provider/v1beta4/event.ts";
export { Info, Provider } from "./virtengine/provider/v1beta4/provider.ts";
export { GenesisState as Provider_GenesisState } from "./virtengine/provider/v1beta4/genesis.ts";
export { MsgCreateProvider, MsgCreateProviderResponse, MsgUpdateProvider, MsgUpdateProviderResponse, MsgDeleteProvider, MsgDeleteProviderResponse } from "./virtengine/provider/v1beta4/msg.ts";
export { QueryProvidersRequest, QueryProvidersResponse, QueryProviderRequest, QueryProviderResponse } from "./virtengine/provider/v1beta4/query.ts";
