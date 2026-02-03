import { patched } from "./nodePatchMessage.ts";

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
export { Params } from "./virtengine/market/v1beta5/params.ts";
export { Order_State } from "./virtengine/market/v1beta5/order.ts";

import { Order as _Order } from "./virtengine/market/v1beta5/order.ts";
export const Order = patched(_Order);
export type Order = _Order

import { GenesisState as _GenesisState } from "./virtengine/market/v1beta5/genesis.ts";
export const GenesisState = patched(_GenesisState);
export type GenesisState = _GenesisState
export { MsgCreateLease, MsgCreateLeaseResponse, MsgWithdrawLease, MsgWithdrawLeaseResponse, MsgCloseLease, MsgCloseLeaseResponse } from "./virtengine/market/v1beta5/leasemsg.ts";
export { MsgUpdateParams, MsgUpdateParamsResponse } from "./virtengine/market/v1beta5/paramsmsg.ts";
export { QueryOrdersRequest, QueryOrderRequest, QueryBidsRequest, QueryBidRequest, QueryLeasesRequest, QueryLeaseRequest, QueryParamsRequest, QueryParamsResponse } from "./virtengine/market/v1beta5/query.ts";

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
