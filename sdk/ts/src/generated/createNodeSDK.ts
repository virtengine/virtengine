import { createServiceLoader } from "../sdk/client/createServiceLoader.ts";
import { SDKOptions } from "../sdk/types.ts";

import type * as virtengine_audit_v1_query from "./protos\virtengine\audit\v1\query.ts";
import type * as virtengine_audit_v1_msg from "./protos\virtengine\audit\v1\msg.ts";
import type * as virtengine_bme_v1_query from "./protos\virtengine\bme\v1\query.ts";
import type * as virtengine_bme_v1_msgs from "./protos\virtengine\bme\v1\msgs.ts";
import type * as virtengine_cert_v1_query from "./protos\virtengine\cert\v1\query.ts";
import type * as virtengine_cert_v1_msg from "./protos\virtengine\cert\v1\msg.ts";
import type * as virtengine_deployment_v1beta4_query from "./protos\virtengine\deployment\v1beta4\query.ts";
import type * as virtengine_deployment_v1beta4_deploymentmsg from "./protos\virtengine\deployment\v1beta4\deploymentmsg.ts";
import type * as virtengine_deployment_v1beta4_groupmsg from "./protos\virtengine\deployment\v1beta4\groupmsg.ts";
import type * as virtengine_deployment_v1beta4_paramsmsg from "./protos\virtengine\deployment\v1beta4\paramsmsg.ts";
import type * as virtengine_deployment_v1beta5_query from "./protos\virtengine\deployment\v1beta5\query.ts";
import type * as virtengine_deployment_v1beta5_deploymentmsg from "./protos\virtengine\deployment\v1beta5\deploymentmsg.ts";
import type * as virtengine_deployment_v1beta5_groupmsg from "./protos\virtengine\deployment\v1beta5\groupmsg.ts";
import type * as virtengine_deployment_v1beta5_paramsmsg from "./protos\virtengine\deployment\v1beta5\paramsmsg.ts";
import type * as virtengine_downtimedetector_v1beta1_query from "./protos\virtengine\downtimedetector\v1beta1\query.ts";
import type * as virtengine_epochs_v1beta1_query from "./protos\virtengine\epochs\v1beta1\query.ts";
import type * as virtengine_escrow_v1_query from "./protos\virtengine\escrow\v1\query.ts";
import type * as virtengine_escrow_v1_msg from "./protos\virtengine\escrow\v1\msg.ts";
import type * as virtengine_escrow_v1beta3_query from "./protos\virtengine\escrow\v1beta3\query.ts";
import type * as virtengine_market_v1beta5_query from "./protos\virtengine\market\v1beta5\query.ts";
import type * as virtengine_market_v1beta5_bidmsg from "./protos\virtengine\market\v1beta5\bidmsg.ts";
import type * as virtengine_market_v1beta5_leasemsg from "./protos\virtengine\market\v1beta5\leasemsg.ts";
import type * as virtengine_market_v1beta5_paramsmsg from "./protos\virtengine\market\v1beta5\paramsmsg.ts";
import type * as virtengine_market_v2beta1_query from "./protos\virtengine\market\v2beta1\query.ts";
import type * as virtengine_market_v2beta1_bidmsg from "./protos\virtengine\market\v2beta1\bidmsg.ts";
import type * as virtengine_market_v2beta1_leasemsg from "./protos\virtengine\market\v2beta1\leasemsg.ts";
import type * as virtengine_market_v2beta1_paramsmsg from "./protos\virtengine\market\v2beta1\paramsmsg.ts";
import type * as virtengine_oracle_v1_prices from "./protos\virtengine\oracle\v1\prices.ts";
import type * as virtengine_oracle_v1_query from "./protos\virtengine\oracle\v1\query.ts";
import type * as virtengine_oracle_v1_msgs from "./protos\virtengine\oracle\v1\msgs.ts";
import type * as virtengine_provider_v1beta4_query from "./protos\virtengine\provider\v1beta4\query.ts";
import type * as virtengine_provider_v1beta4_msg from "./protos\virtengine\provider\v1beta4\msg.ts";
import type * as virtengine_take_v1_query from "./protos\virtengine\take\v1\query.ts";
import type * as virtengine_take_v1_paramsmsg from "./protos\virtengine\take\v1\paramsmsg.ts";
import type * as virtengine_wasm_v1_query from "./protos\virtengine\wasm\v1\query.ts";
import type * as virtengine_wasm_v1_paramsmsg from "./protos\virtengine\wasm\v1\paramsmsg.ts";
import { createClientFactory } from "../sdk/client/createClientFactory.ts";
import type { Transport, CallOptions, TxCallOptions } from "../sdk/transport/types.ts";
import { withMetadata } from "../sdk/client/sdkMetadata.ts";
import type { DeepPartial, DeepSimplify } from "../encoding/typeEncodingHelpers.ts";


export const serviceLoader= createServiceLoader([
  () => import("./protos/virtengine\audit\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\audit\v1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\bme\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\bme\v1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\cert\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\cert\v1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\deployment\v1beta4\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\deployment\v1beta4\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\deployment\v1beta5\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\deployment\v1beta5\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\downtimedetector\v1beta1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\epochs\v1beta1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\escrow\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\escrow\v1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\escrow\v1beta3\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\market\v1beta5\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\market\v1beta5\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\market\v2beta1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\market\v2beta1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\oracle\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\oracle\v1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\provider\v1beta4\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\provider\v1beta4\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\take\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\take\v1\service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine\take\v1beta3\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\wasm\v1\query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine\wasm\v1\service_virtengine.ts").then(m => m.Msg)
] as const);
export function createSDK(queryTransport: Transport, txTransport: Transport, options?: SDKOptions) {
  const getClient = createClientFactory<CallOptions>(queryTransport, options?.clientOptions);
  const getMsgClient = createClientFactory<TxCallOptions>(txTransport, options?.clientOptions);
  return {
    virtengine: {
      audit: {
        v1: {
          /**
           * getAllProvidersAttributes queries all providers.
           */
          getAllProvidersAttributes: withMetadata(async function getAllProvidersAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryAllProvidersAttributesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).allProvidersAttributes(input, options);
          }, { path: [0, 0] }),
          /**
           * getProviderAttributes queries all provider signed attributes.
           */
          getProviderAttributes: withMetadata(async function getProviderAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryProviderAttributesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).providerAttributes(input, options);
          }, { path: [0, 1] }),
          /**
           * getProviderAuditorAttributes queries provider signed attributes by specific auditor.
           */
          getProviderAuditorAttributes: withMetadata(async function getProviderAuditorAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryProviderAuditorRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).providerAuditorAttributes(input, options);
          }, { path: [0, 2] }),
          /**
           * getAuditorAttributes queries all providers signed by this auditor.
           */
          getAuditorAttributes: withMetadata(async function getAuditorAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryAuditorAttributesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).auditorAttributes(input, options);
          }, { path: [0, 3] }),
          /**
           * signProviderAttributes defines a method that signs provider attributes.
           */
          signProviderAttributes: withMetadata(async function signProviderAttributes(input: DeepSimplify<virtengine_audit_v1_msg.MsgSignProviderAttributes>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getMsgClient(service).signProviderAttributes(input, options);
          }, { path: [1, 0] }),
          /**
           * deleteProviderAttributes defines a method that deletes provider attributes.
           */
          deleteProviderAttributes: withMetadata(async function deleteProviderAttributes(input: DeepSimplify<virtengine_audit_v1_msg.MsgDeleteProviderAttributes>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getMsgClient(service).deleteProviderAttributes(input, options);
          }, { path: [1, 1] })
        }
      },
      bme: {
        v1: {
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_bme_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).params(input, options);
          }, { path: [2, 0] }),
          /**
           * getVaultState returns the current vault state
           */
          getVaultState: withMetadata(async function getVaultState(input: DeepPartial<virtengine_bme_v1_query.QueryVaultStateRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).vaultState(input, options);
          }, { path: [2, 1] }),
          /**
           * getStatus returns the current circuit breaker status
           */
          getStatus: withMetadata(async function getStatus(input: DeepPartial<virtengine_bme_v1_query.QueryStatusRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).status(input, options);
          }, { path: [2, 2] }),
          /**
           * updateParams updates the module parameters.
           * This operation can only be performed through governance proposals.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_bme_v1_msgs.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [3, 0] }),
          /**
           * burnMint allows users to burn one token and mint another at current oracle prices.
           * Typically used to burn unused ACT tokens back to AKT.
           * The operation may be delayed or rejected based on circuit breaker status.
           */
          burnMint: withMetadata(async function burnMint(input: DeepSimplify<virtengine_bme_v1_msgs.MsgBurnMint>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getMsgClient(service).burnMint(input, options);
          }, { path: [3, 1] }),
          /**
           * mintACT mints ACT tokens by burning the specified source token.
           * The mint amount is calculated based on current oracle prices and
           * the collateral ratio. May be halted if circuit breaker is triggered.
           */
          mintACT: withMetadata(async function mintACT(input: DeepSimplify<virtengine_bme_v1_msgs.MsgMintACT>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getMsgClient(service).mintACT(input, options);
          }, { path: [3, 2] }),
          /**
           * burnACT burns ACT tokens and mints the specified destination token.
           * The burn operation uses remint credits when available, otherwise
           * requires adequate collateral backing based on oracle prices.
           */
          burnACT: withMetadata(async function burnACT(input: DeepSimplify<virtengine_bme_v1_msgs.MsgBurnACT>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getMsgClient(service).burnACT(input, options);
          }, { path: [3, 3] })
        }
      },
      cert: {
        v1: {
          /**
           * getCertificates queries certificates on-chain.
           */
          getCertificates: withMetadata(async function getCertificates(input: DeepPartial<virtengine_cert_v1_query.QueryCertificatesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getClient(service).certificates(input, options);
          }, { path: [4, 0] }),
          /**
           * createCertificate defines a method to create new certificate given proper inputs.
           */
          createCertificate: withMetadata(async function createCertificate(input: DeepSimplify<virtengine_cert_v1_msg.MsgCreateCertificate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).createCertificate(input, options);
          }, { path: [5, 0] }),
          /**
           * revokeCertificate defines a method to revoke the certificate.
           */
          revokeCertificate: withMetadata(async function revokeCertificate(input: DeepSimplify<virtengine_cert_v1_msg.MsgRevokeCertificate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).revokeCertificate(input, options);
          }, { path: [5, 1] })
        }
      },
      deployment: {
        v1beta4: {
          /**
           * getDeployments queries deployments.
           */
          getDeployments: withMetadata(async function getDeployments(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryDeploymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).deployments(input, options);
          }, { path: [6, 0] }),
          /**
           * getDeployment queries deployment details.
           */
          getDeployment: withMetadata(async function getDeployment(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryDeploymentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).deployment(input, options);
          }, { path: [6, 1] }),
          /**
           * getGroup queries group details.
           */
          getGroup: withMetadata(async function getGroup(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryGroupRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).group(input, options);
          }, { path: [6, 2] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).params(input, options);
          }, { path: [6, 3] }),
          /**
           * createDeployment defines a method to create new deployment given proper inputs.
           */
          createDeployment: withMetadata(async function createDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgCreateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).createDeployment(input, options);
          }, { path: [7, 0] }),
          /**
           * updateDeployment defines a method to update a deployment given proper inputs.
           */
          updateDeployment: withMetadata(async function updateDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgUpdateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).updateDeployment(input, options);
          }, { path: [7, 1] }),
          /**
           * closeDeployment defines a method to close a deployment given proper inputs.
           */
          closeDeployment: withMetadata(async function closeDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgCloseDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).closeDeployment(input, options);
          }, { path: [7, 2] }),
          /**
           * closeGroup defines a method to close a group of a deployment given proper inputs.
           */
          closeGroup: withMetadata(async function closeGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgCloseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).closeGroup(input, options);
          }, { path: [7, 3] }),
          /**
           * pauseGroup defines a method to pause a group of a deployment given proper inputs.
           */
          pauseGroup: withMetadata(async function pauseGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgPauseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).pauseGroup(input, options);
          }, { path: [7, 4] }),
          /**
           * startGroup defines a method to start a group of a deployment given proper inputs.
           */
          startGroup: withMetadata(async function startGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgStartGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).startGroup(input, options);
          }, { path: [7, 5] }),
          /**
           * updateParams defines a governance operation for updating the x/deployment module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_deployment_v1beta4_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [7, 6] })
        },
        v1beta5: {
          /**
           * getDeployments queries deployments.
           */
          getDeployments: withMetadata(async function getDeployments(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryDeploymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).deployments(input, options);
          }, { path: [8, 0] }),
          /**
           * getDeployment queries deployment details.
           */
          getDeployment: withMetadata(async function getDeployment(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryDeploymentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).deployment(input, options);
          }, { path: [8, 1] }),
          /**
           * getGroup queries group details.
           */
          getGroup: withMetadata(async function getGroup(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryGroupRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).group(input, options);
          }, { path: [8, 2] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).params(input, options);
          }, { path: [8, 3] }),
          /**
           * createDeployment defines a method to create new deployment given proper inputs.
           */
          createDeployment: withMetadata(async function createDeployment(input: DeepSimplify<virtengine_deployment_v1beta5_deploymentmsg.MsgCreateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).createDeployment(input, options);
          }, { path: [9, 0] }),
          /**
           * updateDeployment defines a method to update a deployment given proper inputs.
           */
          updateDeployment: withMetadata(async function updateDeployment(input: DeepSimplify<virtengine_deployment_v1beta5_deploymentmsg.MsgUpdateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).updateDeployment(input, options);
          }, { path: [9, 1] }),
          /**
           * closeDeployment defines a method to close a deployment given proper inputs.
           */
          closeDeployment: withMetadata(async function closeDeployment(input: DeepSimplify<virtengine_deployment_v1beta5_deploymentmsg.MsgCloseDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).closeDeployment(input, options);
          }, { path: [9, 2] }),
          /**
           * closeGroup defines a method to close a group of a deployment given proper inputs.
           */
          closeGroup: withMetadata(async function closeGroup(input: DeepSimplify<virtengine_deployment_v1beta5_groupmsg.MsgCloseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).closeGroup(input, options);
          }, { path: [9, 3] }),
          /**
           * pauseGroup defines a method to pause a group of a deployment given proper inputs.
           */
          pauseGroup: withMetadata(async function pauseGroup(input: DeepSimplify<virtengine_deployment_v1beta5_groupmsg.MsgPauseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).pauseGroup(input, options);
          }, { path: [9, 4] }),
          /**
           * startGroup defines a method to start a group of a deployment given proper inputs.
           */
          startGroup: withMetadata(async function startGroup(input: DeepSimplify<virtengine_deployment_v1beta5_groupmsg.MsgStartGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).startGroup(input, options);
          }, { path: [9, 5] }),
          /**
           * updateParams defines a governance operation for updating the x/deployment module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_deployment_v1beta5_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [9, 6] })
        }
      },
      downtimedetector: {
        v1beta1: {
          /**
           * getRecoveredSinceDowntimeOfLength queries if the chain has recovered for a specified duration
           * since experiencing downtime of a given length
           */
          getRecoveredSinceDowntimeOfLength: withMetadata(async function getRecoveredSinceDowntimeOfLength(input: DeepPartial<virtengine_downtimedetector_v1beta1_query.RecoveredSinceDowntimeOfLengthRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getClient(service).recoveredSinceDowntimeOfLength(input, options);
          }, { path: [10, 0] })
        }
      },
      epochs: {
        v1beta1: {
          /**
           * getEpochInfos provide running epochInfos
           */
          getEpochInfos: withMetadata(async function getEpochInfos(input: DeepPartial<virtengine_epochs_v1beta1_query.QueryEpochInfosRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getClient(service).epochInfos(input, options);
          }, { path: [11, 0] }),
          /**
           * getCurrentEpoch provide current epoch of specified identifier
           */
          getCurrentEpoch: withMetadata(async function getCurrentEpoch(input: DeepPartial<virtengine_epochs_v1beta1_query.QueryCurrentEpochRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getClient(service).currentEpoch(input, options);
          }, { path: [11, 1] })
        }
      },
      escrow: {
        v1: {
          /**
           * getAccounts queries all accounts.
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<virtengine_escrow_v1_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).accounts(input, options);
          }, { path: [12, 0] }),
          /**
           * getPayments queries all payments.
           */
          getPayments: withMetadata(async function getPayments(input: DeepPartial<virtengine_escrow_v1_query.QueryPaymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).payments(input, options);
          }, { path: [12, 1] }),
          /**
           * accountDeposit deposits more funds into the escrow account.
           */
          accountDeposit: withMetadata(async function accountDeposit(input: DeepSimplify<virtengine_escrow_v1_msg.MsgAccountDeposit>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getMsgClient(service).accountDeposit(input, options);
          }, { path: [13, 0] })
        },
        v1beta3: {
          /**
           * getAccounts queries all accounts
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<virtengine_escrow_v1beta3_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getClient(service).accounts(input, options);
          }, { path: [14, 0] }),
          /**
           * getPayments queries all payments
           */
          getPayments: withMetadata(async function getPayments(input: DeepPartial<virtengine_escrow_v1beta3_query.QueryPaymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getClient(service).payments(input, options);
          }, { path: [14, 1] })
        }
      },
      market: {
        v1beta5: {
          /**
           * getOrders queries orders with filters.
           */
          getOrders: withMetadata(async function getOrders(input: DeepPartial<virtengine_market_v1beta5_query.QueryOrdersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).orders(input, options);
          }, { path: [15, 0] }),
          /**
           * getOrder queries order details.
           */
          getOrder: withMetadata(async function getOrder(input: DeepPartial<virtengine_market_v1beta5_query.QueryOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).order(input, options);
          }, { path: [15, 1] }),
          /**
           * getBids queries bids with filters.
           */
          getBids: withMetadata(async function getBids(input: DeepPartial<virtengine_market_v1beta5_query.QueryBidsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).bids(input, options);
          }, { path: [15, 2] }),
          /**
           * getBid queries bid details.
           */
          getBid: withMetadata(async function getBid(input: DeepPartial<virtengine_market_v1beta5_query.QueryBidRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).bid(input, options);
          }, { path: [15, 3] }),
          /**
           * getLeases queries leases with filters.
           */
          getLeases: withMetadata(async function getLeases(input: DeepPartial<virtengine_market_v1beta5_query.QueryLeasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).leases(input, options);
          }, { path: [15, 4] }),
          /**
           * getLease queries lease details.
           */
          getLease: withMetadata(async function getLease(input: DeepPartial<virtengine_market_v1beta5_query.QueryLeaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).lease(input, options);
          }, { path: [15, 5] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_market_v1beta5_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).params(input, options);
          }, { path: [15, 6] }),
          /**
           * createBid defines a method to create a bid given proper inputs.
           */
          createBid: withMetadata(async function createBid(input: DeepSimplify<virtengine_market_v1beta5_bidmsg.MsgCreateBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).createBid(input, options);
          }, { path: [16, 0] }),
          /**
           * closeBid defines a method to close a bid given proper inputs.
           */
          closeBid: withMetadata(async function closeBid(input: DeepSimplify<virtengine_market_v1beta5_bidmsg.MsgCloseBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).closeBid(input, options);
          }, { path: [16, 1] }),
          /**
           * withdrawLease withdraws accrued funds from the lease payment
           */
          withdrawLease: withMetadata(async function withdrawLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgWithdrawLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).withdrawLease(input, options);
          }, { path: [16, 2] }),
          /**
           * createLease creates a new lease
           */
          createLease: withMetadata(async function createLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgCreateLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).createLease(input, options);
          }, { path: [16, 3] }),
          /**
           * closeLease defines a method to close an order given proper inputs.
           */
          closeLease: withMetadata(async function closeLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgCloseLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).closeLease(input, options);
          }, { path: [16, 4] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_market_v1beta5_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [16, 5] })
        },
        v2beta1: {
          /**
           * getOrders queries orders with filters.
           */
          getOrders: withMetadata(async function getOrders(input: DeepPartial<virtengine_market_v2beta1_query.QueryOrdersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).orders(input, options);
          }, { path: [17, 0] }),
          /**
           * getOrder queries order details.
           */
          getOrder: withMetadata(async function getOrder(input: DeepPartial<virtengine_market_v2beta1_query.QueryOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).order(input, options);
          }, { path: [17, 1] }),
          /**
           * getBids queries bids with filters.
           */
          getBids: withMetadata(async function getBids(input: DeepPartial<virtengine_market_v2beta1_query.QueryBidsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).bids(input, options);
          }, { path: [17, 2] }),
          /**
           * getBid queries bid details.
           */
          getBid: withMetadata(async function getBid(input: DeepPartial<virtengine_market_v2beta1_query.QueryBidRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).bid(input, options);
          }, { path: [17, 3] }),
          /**
           * getLeases queries leases with filters.
           */
          getLeases: withMetadata(async function getLeases(input: DeepPartial<virtengine_market_v2beta1_query.QueryLeasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).leases(input, options);
          }, { path: [17, 4] }),
          /**
           * getLease queries lease details.
           */
          getLease: withMetadata(async function getLease(input: DeepPartial<virtengine_market_v2beta1_query.QueryLeaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).lease(input, options);
          }, { path: [17, 5] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_market_v2beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).params(input, options);
          }, { path: [17, 6] }),
          /**
           * createBid defines a method to create a bid given proper inputs.
           */
          createBid: withMetadata(async function createBid(input: DeepSimplify<virtengine_market_v2beta1_bidmsg.MsgCreateBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getMsgClient(service).createBid(input, options);
          }, { path: [18, 0] }),
          /**
           * closeBid defines a method to close a bid given proper inputs.
           */
          closeBid: withMetadata(async function closeBid(input: DeepSimplify<virtengine_market_v2beta1_bidmsg.MsgCloseBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getMsgClient(service).closeBid(input, options);
          }, { path: [18, 1] }),
          /**
           * withdrawLease withdraws accrued funds from the lease payment
           */
          withdrawLease: withMetadata(async function withdrawLease(input: DeepSimplify<virtengine_market_v2beta1_leasemsg.MsgWithdrawLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getMsgClient(service).withdrawLease(input, options);
          }, { path: [18, 2] }),
          /**
           * createLease creates a new lease
           */
          createLease: withMetadata(async function createLease(input: DeepSimplify<virtengine_market_v2beta1_leasemsg.MsgCreateLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getMsgClient(service).createLease(input, options);
          }, { path: [18, 3] }),
          /**
           * closeLease defines a method to close an order given proper inputs.
           */
          closeLease: withMetadata(async function closeLease(input: DeepSimplify<virtengine_market_v2beta1_leasemsg.MsgCloseLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getMsgClient(service).closeLease(input, options);
          }, { path: [18, 4] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_market_v2beta1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [18, 5] })
        }
      },
      oracle: {
        v1: {
          /**
           * getPrices query prices for specific denom
           */
          getPrices: withMetadata(async function getPrices(input: DeepPartial<virtengine_oracle_v1_prices.QueryPricesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getClient(service).prices(input, options);
          }, { path: [19, 0] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_oracle_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getClient(service).params(input, options);
          }, { path: [19, 1] }),
          /**
           * getPriceFeedConfig queries the price feed configuration for a given denom.
           */
          getPriceFeedConfig: withMetadata(async function getPriceFeedConfig(input: DeepPartial<virtengine_oracle_v1_query.QueryPriceFeedConfigRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getClient(service).priceFeedConfig(input, options);
          }, { path: [19, 2] }),
          /**
           * getAggregatedPrice queries the aggregated price for a given denom.
           */
          getAggregatedPrice: withMetadata(async function getAggregatedPrice(input: DeepPartial<virtengine_oracle_v1_query.QueryAggregatedPriceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getClient(service).aggregatedPrice(input, options);
          }, { path: [19, 3] }),
          /**
           * addPriceEntry adds a new price entry for a denomination from an authorized source
           */
          addPriceEntry: withMetadata(async function addPriceEntry(input: DeepSimplify<virtengine_oracle_v1_msgs.MsgAddPriceEntry>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getMsgClient(service).addPriceEntry(input, options);
          }, { path: [20, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/wasm module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v2.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_oracle_v1_msgs.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [20, 1] })
        }
      },
      provider: {
        v1beta4: {
          /**
           * getProviders queries providers
           */
          getProviders: withMetadata(async function getProviders(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProvidersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).providers(input, options);
          }, { path: [21, 0] }),
          /**
           * getProvider queries provider details
           */
          getProvider: withMetadata(async function getProvider(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).provider(input, options);
          }, { path: [21, 1] }),
          /**
           * createProvider defines a method that creates a provider given the proper inputs.
           */
          createProvider: withMetadata(async function createProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgCreateProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).createProvider(input, options);
          }, { path: [22, 0] }),
          /**
           * updateProvider defines a method that updates a provider given the proper inputs.
           */
          updateProvider: withMetadata(async function updateProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgUpdateProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).updateProvider(input, options);
          }, { path: [22, 1] }),
          /**
           * deleteProvider defines a method that deletes a provider given the proper inputs.
           */
          deleteProvider: withMetadata(async function deleteProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgDeleteProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).deleteProvider(input, options);
          }, { path: [22, 2] })
        }
      },
      take: {
        v1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_take_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).params(input, options);
          }, { path: [23, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_take_v1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [24, 0] })
        }
      },
      wasm: {
        v1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_wasm_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).params(input, options);
          }, { path: [26, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/wasm module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v2.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_wasm_v1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [27, 0] })
        }
      }
    }
  };
}
