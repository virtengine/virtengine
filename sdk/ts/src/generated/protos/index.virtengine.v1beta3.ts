import { patched } from "./nodePatchMessage.ts";

export { Attribute, SignedBy, PlacementRequirements } from "./virtengine/base/v1beta3/attribute.ts";
export { Provider, AuditedAttributes, AttributesResponse, AttributesFilters, MsgSignProviderAttributes, MsgSignProviderAttributesResponse, MsgDeleteProviderAttributes, MsgDeleteProviderAttributesResponse } from "./virtengine/audit/v1beta3/audit.ts";
export { GenesisState } from "./virtengine/audit/v1beta3/genesis.ts";
export { ResourceValue } from "./virtengine/base/v1beta3/resourcevalue.ts";
export { CPU } from "./virtengine/base/v1beta3/cpu.ts";
export { Endpoint, Endpoint_Kind } from "./virtengine/base/v1beta3/endpoint.ts";
export { GPU } from "./virtengine/base/v1beta3/gpu.ts";
export { Memory } from "./virtengine/base/v1beta3/memory.ts";
export { Storage } from "./virtengine/base/v1beta3/storage.ts";
export { Resources } from "./virtengine/base/v1beta3/resources.ts";
export { CertificateID, Certificate, Certificate_State, CertificateFilter, MsgCreateCertificate, MsgCreateCertificateResponse, MsgRevokeCertificate, MsgRevokeCertificateResponse } from "./virtengine/cert/v1beta3/cert.ts";
export { GenesisCertificate, GenesisState as Cert_GenesisState } from "./virtengine/cert/v1beta3/genesis.ts";
export { DepositDeploymentAuthorization } from "./virtengine/deployment/v1beta3/authz.ts";
export { DeploymentID, Deployment, Deployment_State, DeploymentFilters } from "./virtengine/deployment/v1beta3/deployment.ts";

import { ResourceUnit as _ResourceUnit } from "./virtengine/deployment/v1beta3/resourceunit.ts";
export const ResourceUnit = patched(_ResourceUnit);
export type ResourceUnit = _ResourceUnit

import { GroupSpec as _GroupSpec } from "./virtengine/deployment/v1beta3/groupspec.ts";
export const GroupSpec = patched(_GroupSpec);
export type GroupSpec = _GroupSpec
export { MsgCreateDeploymentResponse, MsgDepositDeployment, MsgDepositDeploymentResponse, MsgUpdateDeployment, MsgUpdateDeploymentResponse, MsgCloseDeployment, MsgCloseDeploymentResponse } from "./virtengine/deployment/v1beta3/deploymentmsg.ts";

import { MsgCreateDeployment as _MsgCreateDeployment } from "./virtengine/deployment/v1beta3/deploymentmsg.ts";
export const MsgCreateDeployment = patched(_MsgCreateDeployment);
export type MsgCreateDeployment = _MsgCreateDeployment
export { GroupID } from "./virtengine/deployment/v1beta3/groupid.ts";
export { Group_State } from "./virtengine/deployment/v1beta3/group.ts";

import { Group as _Group } from "./virtengine/deployment/v1beta3/group.ts";
export const Group = patched(_Group);
export type Group = _Group
export { Params } from "./virtengine/deployment/v1beta3/params.ts";

import { GenesisDeployment as _GenesisDeployment, GenesisState as _Deployment_GenesisState } from "./virtengine/deployment/v1beta3/genesis.ts";
export const GenesisDeployment = patched(_GenesisDeployment);
export type GenesisDeployment = _GenesisDeployment
export const Deployment_GenesisState = patched(_Deployment_GenesisState);
export type Deployment_GenesisState = _Deployment_GenesisState
export { MsgCloseGroup, MsgCloseGroupResponse, MsgPauseGroup, MsgPauseGroupResponse, MsgStartGroup, MsgStartGroupResponse } from "./virtengine/deployment/v1beta3/groupmsg.ts";
export { AccountID, Account_State, FractionalPayment_State } from "./virtengine/escrow/v1beta3/types.ts";

import { Account as _Account, FractionalPayment as _FractionalPayment } from "./virtengine/escrow/v1beta3/types.ts";
export const Account = patched(_Account);
export type Account = _Account
export const FractionalPayment = patched(_FractionalPayment);
export type FractionalPayment = _FractionalPayment

import { GenesisState as _Escrow_GenesisState } from "./virtengine/escrow/v1beta3/genesis.ts";
export const Escrow_GenesisState = patched(_Escrow_GenesisState);
export type Escrow_GenesisState = _Escrow_GenesisState
export { QueryAccountsRequest, QueryPaymentsRequest } from "./virtengine/escrow/v1beta3/query.ts";

import { QueryAccountsResponse as _QueryAccountsResponse, QueryPaymentsResponse as _QueryPaymentsResponse } from "./virtengine/escrow/v1beta3/query.ts";
export const QueryAccountsResponse = patched(_QueryAccountsResponse);
export type QueryAccountsResponse = _QueryAccountsResponse
export const QueryPaymentsResponse = patched(_QueryPaymentsResponse);
export type QueryPaymentsResponse = _QueryPaymentsResponse
export { ProviderInfo, MsgCreateProvider, MsgCreateProviderResponse, MsgUpdateProvider, MsgUpdateProviderResponse, MsgDeleteProvider, MsgDeleteProviderResponse, Provider as Provider_Provider } from "./virtengine/provider/v1beta3/provider.ts";
export { GenesisState as Provider_GenesisState } from "./virtengine/provider/v1beta3/genesis.ts";
export { DenomTakeRate, Params as Take_Params } from "./virtengine/take/v1beta3/params.ts";
export { GenesisState as Take_GenesisState } from "./virtengine/take/v1beta3/genesis.ts";
