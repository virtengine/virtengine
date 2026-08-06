import { patched } from "./nodePatchMessage.ts";


import { ResourceUnit as _ResourceUnit } from "./virtengine/deployment/v1beta5/resourceunit.ts";
export const ResourceUnit = patched(_ResourceUnit);
export type ResourceUnit = _ResourceUnit

import { GroupSpec as _GroupSpec } from "./virtengine/deployment/v1beta5/groupspec.ts";
export const GroupSpec = patched(_GroupSpec);
export type GroupSpec = _GroupSpec
export { MsgCreateDeploymentResponse, MsgUpdateDeployment, MsgUpdateDeploymentResponse, MsgCloseDeployment, MsgCloseDeploymentResponse } from "./virtengine/deployment/v1beta5/deploymentmsg.ts";

import { MsgCreateDeployment as _MsgCreateDeployment } from "./virtengine/deployment/v1beta5/deploymentmsg.ts";
export const MsgCreateDeployment = patched(_MsgCreateDeployment);
export type MsgCreateDeployment = _MsgCreateDeployment
export { DeploymentFilters, GroupFilters } from "./virtengine/deployment/v1beta5/filters.ts";
export { Group_State } from "./virtengine/deployment/v1beta5/group.ts";

import { Group as _Group } from "./virtengine/deployment/v1beta5/group.ts";
export const Group = patched(_Group);
export type Group = _Group
export { Params } from "./virtengine/deployment/v1beta5/params.ts";

import { GenesisDeployment as _GenesisDeployment, GenesisState as _GenesisState } from "./virtengine/deployment/v1beta5/genesis.ts";
export const GenesisDeployment = patched(_GenesisDeployment);
export type GenesisDeployment = _GenesisDeployment
export const GenesisState = patched(_GenesisState);
export type GenesisState = _GenesisState
export { MsgCloseGroup, MsgCloseGroupResponse, MsgPauseGroup, MsgPauseGroupResponse, MsgStartGroup, MsgStartGroupResponse } from "./virtengine/deployment/v1beta5/groupmsg.ts";
export { MsgUpdateParams, MsgUpdateParamsResponse } from "./virtengine/deployment/v1beta5/paramsmsg.ts";
export { QueryDeploymentsRequest, QueryDeploymentRequest, QueryGroupRequest, QueryParamsRequest, QueryParamsResponse } from "./virtengine/deployment/v1beta5/query.ts";

import { QueryDeploymentsResponse as _QueryDeploymentsResponse, QueryDeploymentResponse as _QueryDeploymentResponse, QueryGroupResponse as _QueryGroupResponse } from "./virtengine/deployment/v1beta5/query.ts";
export const QueryDeploymentsResponse = patched(_QueryDeploymentsResponse);
export type QueryDeploymentsResponse = _QueryDeploymentsResponse
export const QueryDeploymentResponse = patched(_QueryDeploymentResponse);
export type QueryDeploymentResponse = _QueryDeploymentResponse
export const QueryGroupResponse = patched(_QueryGroupResponse);
export type QueryGroupResponse = _QueryGroupResponse
export { ResourceOffer } from "./virtengine/market/v1beta5/resourcesoffer.ts";
export { Bid_State } from "./virtengine/market/v1beta5/bid.ts";

import { Bid as _Bid } from "./virtengine/market/v1beta5/bid.ts";
export const Bid = patched(_Bid);
export type Bid = _Bid
export { MsgCreateBidResponse, MsgCloseBid, MsgCloseBidResponse } from "./virtengine/market/v1beta5/bidmsg.ts";

import { MsgCreateBid as _MsgCreateBid } from "./virtengine/market/v1beta5/bidmsg.ts";
export const MsgCreateBid = patched(_MsgCreateBid);
export type MsgCreateBid = _MsgCreateBid
export { BidFilters, OrderFilters } from "./virtengine/market/v1beta5/filters.ts";
export { Params as Market_Params } from "./virtengine/market/v1beta5/params.ts";
export { Order_State } from "./virtengine/market/v1beta5/order.ts";

import { Order as _Order } from "./virtengine/market/v1beta5/order.ts";
export const Order = patched(_Order);
export type Order = _Order

import { GenesisState as _Market_GenesisState } from "./virtengine/market/v1beta5/genesis.ts";
export const Market_GenesisState = patched(_Market_GenesisState);
export type Market_GenesisState = _Market_GenesisState
export { MsgCreateLease, MsgCreateLeaseResponse, MsgWithdrawLease, MsgWithdrawLeaseResponse, MsgCloseLease, MsgCloseLeaseResponse } from "./virtengine/market/v1beta5/leasemsg.ts";
export { MsgUpdateParams as Market_MsgUpdateParams, MsgUpdateParamsResponse as Market_MsgUpdateParamsResponse } from "./virtengine/market/v1beta5/paramsmsg.ts";
export { QueryOrdersRequest, QueryOrderRequest, QueryBidsRequest, QueryBidRequest, QueryLeasesRequest, QueryLeaseRequest, QueryParamsRequest as Market_QueryParamsRequest, QueryParamsResponse as Market_QueryParamsResponse } from "./virtengine/market/v1beta5/query.ts";

import { QueryOrdersResponse as _QueryOrdersResponse, QueryOrderResponse as _QueryOrderResponse, QueryBidsResponse as _QueryBidsResponse, QueryBidResponse as _QueryBidResponse, QueryLeasesResponse as _QueryLeasesResponse, QueryLeaseResponse as _QueryLeaseResponse } from "./virtengine/market/v1beta5/query.ts";
export const QueryOrdersResponse = patched(_QueryOrdersResponse);
export type QueryOrdersResponse = _QueryOrdersResponse
export const QueryOrderResponse = patched(_QueryOrderResponse);
export type QueryOrderResponse = _QueryOrderResponse
export const QueryBidsResponse = patched(_QueryBidsResponse);
export type QueryBidsResponse = _QueryBidsResponse
export const QueryBidResponse = patched(_QueryBidResponse);
export type QueryBidResponse = _QueryBidResponse
export const QueryLeasesResponse = patched(_QueryLeasesResponse);
export type QueryLeasesResponse = _QueryLeasesResponse
export const QueryLeaseResponse = patched(_QueryLeaseResponse);
export type QueryLeaseResponse = _QueryLeaseResponse
