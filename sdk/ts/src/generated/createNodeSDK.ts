import { createServiceLoader } from "../sdk/client/createServiceLoader.ts";
import { SDKOptions } from "../sdk/types.ts";

import type * as virtengine_audit_v1_query from "./protos/virtengine/audit/v1/query.ts";
import type * as virtengine_audit_v1_msg from "./protos/virtengine/audit/v1/msg.ts";
import type * as virtengine_benchmark_v1_tx from "./protos/virtengine/benchmark/v1/tx.ts";
import type * as virtengine_bme_v1_query from "./protos/virtengine/bme/v1/query.ts";
import type * as virtengine_bme_v1_msgs from "./protos/virtengine/bme/v1/msgs.ts";
import type * as virtengine_cert_v1_query from "./protos/virtengine/cert/v1/query.ts";
import type * as virtengine_cert_v1_msg from "./protos/virtengine/cert/v1/msg.ts";
import type * as virtengine_config_v1_tx from "./protos/virtengine/config/v1/tx.ts";
import type * as virtengine_delegation_v1_query from "./protos/virtengine/delegation/v1/query.ts";
import type * as virtengine_delegation_v1_tx from "./protos/virtengine/delegation/v1/tx.ts";
import type * as virtengine_deployment_v1beta4_query from "./protos/virtengine/deployment/v1beta4/query.ts";
import type * as virtengine_deployment_v1beta4_deploymentmsg from "./protos/virtengine/deployment/v1beta4/deploymentmsg.ts";
import type * as virtengine_deployment_v1beta4_groupmsg from "./protos/virtengine/deployment/v1beta4/groupmsg.ts";
import type * as virtengine_deployment_v1beta4_paramsmsg from "./protos/virtengine/deployment/v1beta4/paramsmsg.ts";
import type * as virtengine_downtimedetector_v1beta1_query from "./protos/virtengine/downtimedetector/v1beta1/query.ts";
import type * as virtengine_enclave_v1_query from "./protos/virtengine/enclave/v1/query.ts";
import type * as virtengine_enclave_v1_tx from "./protos/virtengine/enclave/v1/tx.ts";
import type * as virtengine_encryption_v1_query from "./protos/virtengine/encryption/v1/query.ts";
import type * as virtengine_encryption_v1_tx from "./protos/virtengine/encryption/v1/tx.ts";
import type * as virtengine_epochs_v1beta1_query from "./protos/virtengine/epochs/v1beta1/query.ts";
import type * as virtengine_escrow_v1_query from "./protos/virtengine/escrow/v1/query.ts";
import type * as virtengine_escrow_v1_msg from "./protos/virtengine/escrow/v1/msg.ts";
import type * as virtengine_fraud_v1_query from "./protos/virtengine/fraud/v1/query.ts";
import type * as virtengine_fraud_v1_tx from "./protos/virtengine/fraud/v1/tx.ts";
import type * as virtengine_hpc_v1_query from "./protos/virtengine/hpc/v1/query.ts";
import type * as virtengine_hpc_v1_tx from "./protos/virtengine/hpc/v1/tx.ts";
import type * as virtengine_market_v1beta5_query from "./protos/virtengine/market/v1beta5/query.ts";
import type * as virtengine_market_v1beta5_bidmsg from "./protos/virtengine/market/v1beta5/bidmsg.ts";
import type * as virtengine_market_v1beta5_leasemsg from "./protos/virtengine/market/v1beta5/leasemsg.ts";
import type * as virtengine_market_v1beta5_paramsmsg from "./protos/virtengine/market/v1beta5/paramsmsg.ts";
import type * as virtengine_marketplace_v1_tx from "./protos/virtengine/marketplace/v1/tx.ts";
import type * as virtengine_mfa_v1_query from "./protos/virtengine/mfa/v1/query.ts";
import type * as virtengine_mfa_v1_tx from "./protos/virtengine/mfa/v1/tx.ts";
import type * as virtengine_oracle_v1_prices from "./protos/virtengine/oracle/v1/prices.ts";
import type * as virtengine_oracle_v1_query from "./protos/virtengine/oracle/v1/query.ts";
import type * as virtengine_oracle_v1_msgs from "./protos/virtengine/oracle/v1/msgs.ts";
import type * as virtengine_provider_v1beta4_query from "./protos/virtengine/provider/v1beta4/query.ts";
import type * as virtengine_provider_v1beta4_msg from "./protos/virtengine/provider/v1beta4/msg.ts";
import type * as virtengine_review_v1_tx from "./protos/virtengine/review/v1/tx.ts";
import type * as virtengine_roles_v1_query from "./protos/virtengine/roles/v1/query.ts";
import type * as virtengine_roles_v1_tx from "./protos/virtengine/roles/v1/tx.ts";
import type * as virtengine_settlement_v1_tx from "./protos/virtengine/settlement/v1/tx.ts";
import type * as virtengine_staking_v1_tx from "./protos/virtengine/staking/v1/tx.ts";
import type * as virtengine_take_v1_query from "./protos/virtengine/take/v1/query.ts";
import type * as virtengine_take_v1_paramsmsg from "./protos/virtengine/take/v1/paramsmsg.ts";
import type * as virtengine_veid_v1_query from "./protos/virtengine/veid/v1/query.ts";
import type * as virtengine_veid_v1_tx from "./protos/virtengine/veid/v1/tx.ts";
import type * as virtengine_veid_v1_appeal from "./protos/virtengine/veid/v1/appeal.ts";
import type * as virtengine_veid_v1_compliance from "./protos/virtengine/veid/v1/compliance.ts";
import type * as virtengine_veid_v1_model from "./protos/virtengine/veid/v1/model.ts";
import type * as virtengine_wasm_v1_query from "./protos/virtengine/wasm/v1/query.ts";
import type * as virtengine_wasm_v1_paramsmsg from "./protos/virtengine/wasm/v1/paramsmsg.ts";
import { createClientFactory } from "../sdk/client/createClientFactory.ts";
import type { Transport, CallOptions, TxCallOptions } from "../sdk/transport/types.ts";
import { withMetadata } from "../sdk/client/sdkMetadata.ts";
import type { DeepPartial, DeepSimplify } from "../encoding/typeEncodingHelpers.ts";


export const serviceLoader= createServiceLoader([
  () => import("./protos/virtengine/audit/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/audit/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/benchmark/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/bme/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/bme/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/cert/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/cert/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/config/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/delegation/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/delegation/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/deployment/v1beta4/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/deployment/v1beta4/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/downtimedetector/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/enclave/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/enclave/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/encryption/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/encryption/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/epochs/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/escrow/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/escrow/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/fraud/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/fraud/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/hpc/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/hpc/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/market/v1beta5/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/market/v1beta5/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/marketplace/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/mfa/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/mfa/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/oracle/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/oracle/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/provider/v1beta4/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/provider/v1beta4/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/review/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/roles/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/roles/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/settlement/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/staking/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/take/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/take/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/veid/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/veid/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/wasm/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/wasm/v1/service_virtengine.ts").then(m => m.Msg)
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
      benchmark: {
        v1: {
          /**
           * submitBenchmarks submits benchmark results
           */
          submitBenchmarks: withMetadata(async function submitBenchmarks(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgSubmitBenchmarks>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).submitBenchmarks(input, options);
          }, { path: [2, 0] }),
          /**
           * requestChallenge requests a benchmark challenge
           */
          requestChallenge: withMetadata(async function requestChallenge(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgRequestChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).requestChallenge(input, options);
          }, { path: [2, 1] }),
          /**
           * respondChallenge responds to a benchmark challenge
           */
          respondChallenge: withMetadata(async function respondChallenge(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgRespondChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).respondChallenge(input, options);
          }, { path: [2, 2] }),
          /**
           * flagProvider flags a provider for anomaly
           */
          flagProvider: withMetadata(async function flagProvider(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgFlagProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).flagProvider(input, options);
          }, { path: [2, 3] }),
          /**
           * unflagProvider removes a provider flag
           */
          unflagProvider: withMetadata(async function unflagProvider(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgUnflagProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).unflagProvider(input, options);
          }, { path: [2, 4] }),
          /**
           * resolveAnomalyFlag resolves an anomaly flag
           */
          resolveAnomalyFlag: withMetadata(async function resolveAnomalyFlag(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgResolveAnomalyFlag>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).resolveAnomalyFlag(input, options);
          }, { path: [2, 5] })
        }
      },
      bme: {
        v1: {
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_bme_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).params(input, options);
          }, { path: [3, 0] }),
          /**
           * getVaultState returns the current vault state
           */
          getVaultState: withMetadata(async function getVaultState(input: DeepPartial<virtengine_bme_v1_query.QueryVaultStateRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).vaultState(input, options);
          }, { path: [3, 1] }),
          /**
           * getStatus returns the current circuit breaker status
           */
          getStatus: withMetadata(async function getStatus(input: DeepPartial<virtengine_bme_v1_query.QueryStatusRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).status(input, options);
          }, { path: [3, 2] }),
          /**
           * updateParams updates the module parameters.
           * This operation can only be performed through governance proposals.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_bme_v1_msgs.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [4, 0] }),
          /**
           * burnMint allows users to burn one token and mint another at current oracle prices.
           * Typically used to burn unused ACT tokens back to AKT.
           * The operation may be delayed or rejected based on circuit breaker status.
           */
          burnMint: withMetadata(async function burnMint(input: DeepSimplify<virtengine_bme_v1_msgs.MsgBurnMint>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).burnMint(input, options);
          }, { path: [4, 1] }),
          /**
           * mintACT mints ACT tokens by burning the specified source token.
           * The mint amount is calculated based on current oracle prices and
           * the collateral ratio. May be halted if circuit breaker is triggered.
           */
          mintACT: withMetadata(async function mintACT(input: DeepSimplify<virtengine_bme_v1_msgs.MsgMintACT>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).mintACT(input, options);
          }, { path: [4, 2] }),
          /**
           * burnACT burns ACT tokens and mints the specified destination token.
           * The burn operation uses remint credits when available, otherwise
           * requires adequate collateral backing based on oracle prices.
           */
          burnACT: withMetadata(async function burnACT(input: DeepSimplify<virtengine_bme_v1_msgs.MsgBurnACT>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).burnACT(input, options);
          }, { path: [4, 3] })
        }
      },
      cert: {
        v1: {
          /**
           * getCertificates queries certificates on-chain.
           */
          getCertificates: withMetadata(async function getCertificates(input: DeepPartial<virtengine_cert_v1_query.QueryCertificatesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getClient(service).certificates(input, options);
          }, { path: [5, 0] }),
          /**
           * createCertificate defines a method to create new certificate given proper inputs.
           */
          createCertificate: withMetadata(async function createCertificate(input: DeepSimplify<virtengine_cert_v1_msg.MsgCreateCertificate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getMsgClient(service).createCertificate(input, options);
          }, { path: [6, 0] }),
          /**
           * revokeCertificate defines a method to revoke the certificate.
           */
          revokeCertificate: withMetadata(async function revokeCertificate(input: DeepSimplify<virtengine_cert_v1_msg.MsgRevokeCertificate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getMsgClient(service).revokeCertificate(input, options);
          }, { path: [6, 1] })
        }
      },
      config: {
        v1: {
          /**
           * registerApprovedClient registers a new approved client
           */
          registerApprovedClient: withMetadata(async function registerApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgRegisterApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).registerApprovedClient(input, options);
          }, { path: [7, 0] }),
          /**
           * updateApprovedClient updates an approved client
           */
          updateApprovedClient: withMetadata(async function updateApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgUpdateApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).updateApprovedClient(input, options);
          }, { path: [7, 1] }),
          /**
           * suspendApprovedClient suspends an approved client
           */
          suspendApprovedClient: withMetadata(async function suspendApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgSuspendApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).suspendApprovedClient(input, options);
          }, { path: [7, 2] }),
          /**
           * revokeApprovedClient revokes an approved client
           */
          revokeApprovedClient: withMetadata(async function revokeApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgRevokeApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).revokeApprovedClient(input, options);
          }, { path: [7, 3] }),
          /**
           * reactivateApprovedClient reactivates a suspended client
           */
          reactivateApprovedClient: withMetadata(async function reactivateApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgReactivateApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).reactivateApprovedClient(input, options);
          }, { path: [7, 4] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_config_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [7, 5] })
        }
      },
      delegation: {
        v1: {
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_delegation_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).params(input, options);
          }, { path: [8, 0] }),
          /**
           * getDelegation queries a specific delegation
           */
          getDelegation: withMetadata(async function getDelegation(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).delegation(input, options);
          }, { path: [8, 1] }),
          /**
           * getDelegatorDelegations queries all delegations for a delegator
           */
          getDelegatorDelegations: withMetadata(async function getDelegatorDelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).delegatorDelegations(input, options);
          }, { path: [8, 2] }),
          /**
           * getValidatorDelegations queries all delegations for a validator
           */
          getValidatorDelegations: withMetadata(async function getValidatorDelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryValidatorDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).validatorDelegations(input, options);
          }, { path: [8, 3] }),
          /**
           * getUnbondingDelegation queries a specific unbonding delegation
           */
          getUnbondingDelegation: withMetadata(async function getUnbondingDelegation(input: DeepPartial<virtengine_delegation_v1_query.QueryUnbondingDelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).unbondingDelegation(input, options);
          }, { path: [8, 4] }),
          /**
           * getDelegatorUnbondingDelegations queries all unbonding delegations for a delegator
           */
          getDelegatorUnbondingDelegations: withMetadata(async function getDelegatorUnbondingDelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorUnbondingDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).delegatorUnbondingDelegations(input, options);
          }, { path: [8, 5] }),
          /**
           * getRedelegation queries a specific redelegation
           */
          getRedelegation: withMetadata(async function getRedelegation(input: DeepPartial<virtengine_delegation_v1_query.QueryRedelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).redelegation(input, options);
          }, { path: [8, 6] }),
          /**
           * getDelegatorRedelegations queries all redelegations for a delegator
           */
          getDelegatorRedelegations: withMetadata(async function getDelegatorRedelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorRedelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).delegatorRedelegations(input, options);
          }, { path: [8, 7] }),
          /**
           * getDelegatorRewards queries unclaimed rewards for a delegator from a specific validator
           */
          getDelegatorRewards: withMetadata(async function getDelegatorRewards(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).delegatorRewards(input, options);
          }, { path: [8, 8] }),
          /**
           * getDelegatorAllRewards queries all unclaimed rewards for a delegator
           */
          getDelegatorAllRewards: withMetadata(async function getDelegatorAllRewards(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorAllRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).delegatorAllRewards(input, options);
          }, { path: [8, 9] }),
          /**
           * getValidatorShares queries the total shares for a validator
           */
          getValidatorShares: withMetadata(async function getValidatorShares(input: DeepPartial<virtengine_delegation_v1_query.QueryValidatorSharesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).validatorShares(input, options);
          }, { path: [8, 10] }),
          /**
           * delegate delegates tokens to a validator
           */
          delegate: withMetadata(async function delegate(input: DeepSimplify<virtengine_delegation_v1_tx.MsgDelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).delegate(input, options);
          }, { path: [9, 0] }),
          /**
           * undelegate undelegates tokens from a validator
           */
          undelegate: withMetadata(async function undelegate(input: DeepSimplify<virtengine_delegation_v1_tx.MsgUndelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).undelegate(input, options);
          }, { path: [9, 1] }),
          /**
           * redelegate redelegates tokens between validators
           */
          redelegate: withMetadata(async function redelegate(input: DeepSimplify<virtengine_delegation_v1_tx.MsgRedelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).redelegate(input, options);
          }, { path: [9, 2] }),
          /**
           * claimRewards claims rewards from a specific validator
           */
          claimRewards: withMetadata(async function claimRewards(input: DeepSimplify<virtengine_delegation_v1_tx.MsgClaimRewards>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).claimRewards(input, options);
          }, { path: [9, 3] }),
          /**
           * claimAllRewards claims rewards from all validators
           */
          claimAllRewards: withMetadata(async function claimAllRewards(input: DeepSimplify<virtengine_delegation_v1_tx.MsgClaimAllRewards>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).claimAllRewards(input, options);
          }, { path: [9, 4] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_delegation_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [9, 5] })
        }
      },
      deployment: {
        v1beta4: {
          /**
           * getDeployments queries deployments.
           */
          getDeployments: withMetadata(async function getDeployments(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryDeploymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getClient(service).deployments(input, options);
          }, { path: [10, 0] }),
          /**
           * getDeployment queries deployment details.
           */
          getDeployment: withMetadata(async function getDeployment(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryDeploymentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getClient(service).deployment(input, options);
          }, { path: [10, 1] }),
          /**
           * getGroup queries group details.
           */
          getGroup: withMetadata(async function getGroup(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryGroupRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getClient(service).group(input, options);
          }, { path: [10, 2] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getClient(service).params(input, options);
          }, { path: [10, 3] }),
          /**
           * createDeployment defines a method to create new deployment given proper inputs.
           */
          createDeployment: withMetadata(async function createDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgCreateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).createDeployment(input, options);
          }, { path: [11, 0] }),
          /**
           * updateDeployment defines a method to update a deployment given proper inputs.
           */
          updateDeployment: withMetadata(async function updateDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgUpdateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).updateDeployment(input, options);
          }, { path: [11, 1] }),
          /**
           * closeDeployment defines a method to close a deployment given proper inputs.
           */
          closeDeployment: withMetadata(async function closeDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgCloseDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).closeDeployment(input, options);
          }, { path: [11, 2] }),
          /**
           * closeGroup defines a method to close a group of a deployment given proper inputs.
           */
          closeGroup: withMetadata(async function closeGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgCloseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).closeGroup(input, options);
          }, { path: [11, 3] }),
          /**
           * pauseGroup defines a method to pause a group of a deployment given proper inputs.
           */
          pauseGroup: withMetadata(async function pauseGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgPauseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).pauseGroup(input, options);
          }, { path: [11, 4] }),
          /**
           * startGroup defines a method to start a group of a deployment given proper inputs.
           */
          startGroup: withMetadata(async function startGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgStartGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).startGroup(input, options);
          }, { path: [11, 5] }),
          /**
           * updateParams defines a governance operation for updating the x/deployment module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_deployment_v1beta4_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [11, 6] })
        }
      },
      downtimedetector: {
        v1beta1: {
          /**
           * getRecoveredSinceDowntimeOfLength queries if the chain has recovered for a specified duration
           * since experiencing downtime of a given length
           */
          getRecoveredSinceDowntimeOfLength: withMetadata(async function getRecoveredSinceDowntimeOfLength(input: DeepPartial<virtengine_downtimedetector_v1beta1_query.RecoveredSinceDowntimeOfLengthRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).recoveredSinceDowntimeOfLength(input, options);
          }, { path: [12, 0] })
        }
      },
      enclave: {
        v1: {
          /**
           * getEnclaveIdentity queries an enclave identity for a validator
           */
          getEnclaveIdentity: withMetadata(async function getEnclaveIdentity(input: DeepPartial<virtengine_enclave_v1_query.QueryEnclaveIdentityRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).enclaveIdentity(input, options);
          }, { path: [13, 0] }),
          /**
           * getActiveValidatorEnclaveKeys queries all active validator enclave keys
           */
          getActiveValidatorEnclaveKeys: withMetadata(async function getActiveValidatorEnclaveKeys(input: DeepPartial<virtengine_enclave_v1_query.QueryActiveValidatorEnclaveKeysRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).activeValidatorEnclaveKeys(input, options);
          }, { path: [13, 1] }),
          /**
           * getCommitteeEnclaveKeys queries committee enclave keys for an epoch
           */
          getCommitteeEnclaveKeys: withMetadata(async function getCommitteeEnclaveKeys(input: DeepPartial<virtengine_enclave_v1_query.QueryCommitteeEnclaveKeysRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).committeeEnclaveKeys(input, options);
          }, { path: [13, 2] }),
          /**
           * getMeasurementAllowlist queries the measurement allowlist
           */
          getMeasurementAllowlist: withMetadata(async function getMeasurementAllowlist(input: DeepPartial<virtengine_enclave_v1_query.QueryMeasurementAllowlistRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).measurementAllowlist(input, options);
          }, { path: [13, 3] }),
          /**
           * getMeasurement queries a specific measurement
           */
          getMeasurement: withMetadata(async function getMeasurement(input: DeepPartial<virtengine_enclave_v1_query.QueryMeasurementRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).measurement(input, options);
          }, { path: [13, 4] }),
          /**
           * getKeyRotation queries key rotation status for a validator
           */
          getKeyRotation: withMetadata(async function getKeyRotation(input: DeepPartial<virtengine_enclave_v1_query.QueryKeyRotationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).keyRotation(input, options);
          }, { path: [13, 5] }),
          /**
           * getValidKeySet queries the current valid key set
           */
          getValidKeySet: withMetadata(async function getValidKeySet(input: DeepPartial<virtengine_enclave_v1_query.QueryValidKeySetRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).validKeySet(input, options);
          }, { path: [13, 6] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_enclave_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).params(input, options);
          }, { path: [13, 7] }),
          /**
           * getAttestedResult queries an attested scoring result
           */
          getAttestedResult: withMetadata(async function getAttestedResult(input: DeepPartial<virtengine_enclave_v1_query.QueryAttestedResultRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).attestedResult(input, options);
          }, { path: [13, 8] }),
          /**
           * registerEnclaveIdentity registers a new enclave identity for a validator
           */
          registerEnclaveIdentity: withMetadata(async function registerEnclaveIdentity(input: DeepSimplify<virtengine_enclave_v1_tx.MsgRegisterEnclaveIdentity>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).registerEnclaveIdentity(input, options);
          }, { path: [14, 0] }),
          /**
           * rotateEnclaveIdentity initiates a key rotation for a validator's enclave
           */
          rotateEnclaveIdentity: withMetadata(async function rotateEnclaveIdentity(input: DeepSimplify<virtengine_enclave_v1_tx.MsgRotateEnclaveIdentity>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).rotateEnclaveIdentity(input, options);
          }, { path: [14, 1] }),
          /**
           * proposeMeasurement proposes a new enclave measurement for the allowlist
           */
          proposeMeasurement: withMetadata(async function proposeMeasurement(input: DeepSimplify<virtengine_enclave_v1_tx.MsgProposeMeasurement>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).proposeMeasurement(input, options);
          }, { path: [14, 2] }),
          /**
           * revokeMeasurement revokes an enclave measurement from the allowlist
           */
          revokeMeasurement: withMetadata(async function revokeMeasurement(input: DeepSimplify<virtengine_enclave_v1_tx.MsgRevokeMeasurement>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).revokeMeasurement(input, options);
          }, { path: [14, 3] }),
          /**
           * updateParams updates the module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_enclave_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [14, 4] })
        }
      },
      encryption: {
        v1: {
          /**
           * getRecipientKey returns the recipient keys for an account
           */
          getRecipientKey: withMetadata(async function getRecipientKey(input: DeepPartial<virtengine_encryption_v1_query.QueryRecipientKeyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).recipientKey(input, options);
          }, { path: [15, 0] }),
          /**
           * getKeyByFingerprint returns a key by its fingerprint
           */
          getKeyByFingerprint: withMetadata(async function getKeyByFingerprint(input: DeepPartial<virtengine_encryption_v1_query.QueryKeyByFingerprintRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).keyByFingerprint(input, options);
          }, { path: [15, 1] }),
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_encryption_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).params(input, options);
          }, { path: [15, 2] }),
          /**
           * getAlgorithms returns the supported encryption algorithms
           */
          getAlgorithms: withMetadata(async function getAlgorithms(input: DeepPartial<virtengine_encryption_v1_query.QueryAlgorithmsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).algorithms(input, options);
          }, { path: [15, 3] }),
          /**
           * getValidateEnvelope validates an encrypted payload envelope
           */
          getValidateEnvelope: withMetadata(async function getValidateEnvelope(input: DeepPartial<virtengine_encryption_v1_query.QueryValidateEnvelopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).validateEnvelope(input, options);
          }, { path: [15, 4] }),
          /**
           * registerRecipientKey registers a new recipient public key
           */
          registerRecipientKey: withMetadata(async function registerRecipientKey(input: DeepSimplify<virtengine_encryption_v1_tx.MsgRegisterRecipientKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).registerRecipientKey(input, options);
          }, { path: [16, 0] }),
          /**
           * revokeRecipientKey revokes an existing recipient public key
           */
          revokeRecipientKey: withMetadata(async function revokeRecipientKey(input: DeepSimplify<virtengine_encryption_v1_tx.MsgRevokeRecipientKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).revokeRecipientKey(input, options);
          }, { path: [16, 1] }),
          /**
           * updateKeyLabel updates the label of an existing recipient key
           */
          updateKeyLabel: withMetadata(async function updateKeyLabel(input: DeepSimplify<virtengine_encryption_v1_tx.MsgUpdateKeyLabel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).updateKeyLabel(input, options);
          }, { path: [16, 2] })
        }
      },
      epochs: {
        v1beta1: {
          /**
           * getEpochInfos provide running epochInfos
           */
          getEpochInfos: withMetadata(async function getEpochInfos(input: DeepPartial<virtengine_epochs_v1beta1_query.QueryEpochInfosRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).epochInfos(input, options);
          }, { path: [17, 0] }),
          /**
           * getCurrentEpoch provide current epoch of specified identifier
           */
          getCurrentEpoch: withMetadata(async function getCurrentEpoch(input: DeepPartial<virtengine_epochs_v1beta1_query.QueryCurrentEpochRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).currentEpoch(input, options);
          }, { path: [17, 1] })
        }
      },
      escrow: {
        v1: {
          /**
           * getAccounts queries all accounts.
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<virtengine_escrow_v1_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).accounts(input, options);
          }, { path: [18, 0] }),
          /**
           * getPayments queries all payments.
           */
          getPayments: withMetadata(async function getPayments(input: DeepPartial<virtengine_escrow_v1_query.QueryPaymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).payments(input, options);
          }, { path: [18, 1] }),
          /**
           * accountDeposit deposits more funds into the escrow account.
           */
          accountDeposit: withMetadata(async function accountDeposit(input: DeepSimplify<virtengine_escrow_v1_msg.MsgAccountDeposit>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).accountDeposit(input, options);
          }, { path: [19, 0] })
        }
      },
      fraud: {
        v1: {
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_fraud_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).params(input, options);
          }, { path: [20, 0] }),
          /**
           * getFraudReport returns a fraud report by ID
           */
          getFraudReport: withMetadata(async function getFraudReport(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).fraudReport(input, options);
          }, { path: [20, 1] }),
          /**
           * getFraudReports returns all fraud reports with optional filters
           */
          getFraudReports: withMetadata(async function getFraudReports(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).fraudReports(input, options);
          }, { path: [20, 2] }),
          /**
           * getFraudReportsByReporter returns fraud reports submitted by a reporter
           */
          getFraudReportsByReporter: withMetadata(async function getFraudReportsByReporter(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportsByReporterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).fraudReportsByReporter(input, options);
          }, { path: [20, 3] }),
          /**
           * getFraudReportsByReportedParty returns fraud reports against a reported party
           */
          getFraudReportsByReportedParty: withMetadata(async function getFraudReportsByReportedParty(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportsByReportedPartyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).fraudReportsByReportedParty(input, options);
          }, { path: [20, 4] }),
          /**
           * getAuditLog returns the audit log for a report
           */
          getAuditLog: withMetadata(async function getAuditLog(input: DeepPartial<virtengine_fraud_v1_query.QueryAuditLogRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).auditLog(input, options);
          }, { path: [20, 5] }),
          /**
           * getModeratorQueue returns the moderator queue entries
           */
          getModeratorQueue: withMetadata(async function getModeratorQueue(input: DeepPartial<virtengine_fraud_v1_query.QueryModeratorQueueRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).moderatorQueue(input, options);
          }, { path: [20, 6] }),
          /**
           * submitFraudReport submits a new fraud report
           */
          submitFraudReport: withMetadata(async function submitFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgSubmitFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).submitFraudReport(input, options);
          }, { path: [21, 0] }),
          /**
           * assignModerator assigns a moderator to a fraud report
           */
          assignModerator: withMetadata(async function assignModerator(input: DeepSimplify<virtengine_fraud_v1_tx.MsgAssignModerator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).assignModerator(input, options);
          }, { path: [21, 1] }),
          /**
           * updateReportStatus updates the status of a fraud report
           */
          updateReportStatus: withMetadata(async function updateReportStatus(input: DeepSimplify<virtengine_fraud_v1_tx.MsgUpdateReportStatus>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).updateReportStatus(input, options);
          }, { path: [21, 2] }),
          /**
           * resolveFraudReport resolves a fraud report with action
           */
          resolveFraudReport: withMetadata(async function resolveFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgResolveFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).resolveFraudReport(input, options);
          }, { path: [21, 3] }),
          /**
           * rejectFraudReport rejects a fraud report
           */
          rejectFraudReport: withMetadata(async function rejectFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgRejectFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).rejectFraudReport(input, options);
          }, { path: [21, 4] }),
          /**
           * escalateFraudReport escalates a fraud report to admin
           */
          escalateFraudReport: withMetadata(async function escalateFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgEscalateFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).escalateFraudReport(input, options);
          }, { path: [21, 5] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_fraud_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [21, 6] })
        }
      },
      hpc: {
        v1: {
          /**
           * getCluster returns a cluster by ID
           */
          getCluster: withMetadata(async function getCluster(input: DeepPartial<virtengine_hpc_v1_query.QueryClusterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).cluster(input, options);
          }, { path: [22, 0] }),
          /**
           * getClusters returns all clusters with optional filters
           */
          getClusters: withMetadata(async function getClusters(input: DeepPartial<virtengine_hpc_v1_query.QueryClustersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).clusters(input, options);
          }, { path: [22, 1] }),
          /**
           * getClustersByProvider returns clusters owned by a provider
           */
          getClustersByProvider: withMetadata(async function getClustersByProvider(input: DeepPartial<virtengine_hpc_v1_query.QueryClustersByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).clustersByProvider(input, options);
          }, { path: [22, 2] }),
          /**
           * getOffering returns an offering by ID
           */
          getOffering: withMetadata(async function getOffering(input: DeepPartial<virtengine_hpc_v1_query.QueryOfferingRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).offering(input, options);
          }, { path: [22, 3] }),
          /**
           * getOfferings returns all offerings with optional filters
           */
          getOfferings: withMetadata(async function getOfferings(input: DeepPartial<virtengine_hpc_v1_query.QueryOfferingsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).offerings(input, options);
          }, { path: [22, 4] }),
          /**
           * getOfferingsByCluster returns offerings for a cluster
           */
          getOfferingsByCluster: withMetadata(async function getOfferingsByCluster(input: DeepPartial<virtengine_hpc_v1_query.QueryOfferingsByClusterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).offeringsByCluster(input, options);
          }, { path: [22, 5] }),
          /**
           * getJob returns a job by ID
           */
          getJob: withMetadata(async function getJob(input: DeepPartial<virtengine_hpc_v1_query.QueryJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).job(input, options);
          }, { path: [22, 6] }),
          /**
           * getJobs returns all jobs with optional filters
           */
          getJobs: withMetadata(async function getJobs(input: DeepPartial<virtengine_hpc_v1_query.QueryJobsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).jobs(input, options);
          }, { path: [22, 7] }),
          /**
           * getJobsByCustomer returns jobs submitted by a customer
           */
          getJobsByCustomer: withMetadata(async function getJobsByCustomer(input: DeepPartial<virtengine_hpc_v1_query.QueryJobsByCustomerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).jobsByCustomer(input, options);
          }, { path: [22, 8] }),
          /**
           * getJobsByProvider returns jobs handled by a provider
           */
          getJobsByProvider: withMetadata(async function getJobsByProvider(input: DeepPartial<virtengine_hpc_v1_query.QueryJobsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).jobsByProvider(input, options);
          }, { path: [22, 9] }),
          /**
           * getJobAccounting returns accounting data for a job
           */
          getJobAccounting: withMetadata(async function getJobAccounting(input: DeepPartial<virtengine_hpc_v1_query.QueryJobAccountingRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).jobAccounting(input, options);
          }, { path: [22, 10] }),
          /**
           * getNodeMetadata returns metadata for a node
           */
          getNodeMetadata: withMetadata(async function getNodeMetadata(input: DeepPartial<virtengine_hpc_v1_query.QueryNodeMetadataRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).nodeMetadata(input, options);
          }, { path: [22, 11] }),
          /**
           * getNodesByCluster returns nodes in a cluster
           */
          getNodesByCluster: withMetadata(async function getNodesByCluster(input: DeepPartial<virtengine_hpc_v1_query.QueryNodesByClusterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).nodesByCluster(input, options);
          }, { path: [22, 12] }),
          /**
           * getSchedulingDecision returns a scheduling decision by ID
           */
          getSchedulingDecision: withMetadata(async function getSchedulingDecision(input: DeepPartial<virtengine_hpc_v1_query.QuerySchedulingDecisionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).schedulingDecision(input, options);
          }, { path: [22, 13] }),
          /**
           * getSchedulingDecisionByJob returns the scheduling decision for a job
           */
          getSchedulingDecisionByJob: withMetadata(async function getSchedulingDecisionByJob(input: DeepPartial<virtengine_hpc_v1_query.QuerySchedulingDecisionByJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).schedulingDecisionByJob(input, options);
          }, { path: [22, 14] }),
          /**
           * getReward returns a reward record by ID
           */
          getReward: withMetadata(async function getReward(input: DeepPartial<virtengine_hpc_v1_query.QueryRewardRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).reward(input, options);
          }, { path: [22, 15] }),
          /**
           * getRewardsByJob returns rewards for a job
           */
          getRewardsByJob: withMetadata(async function getRewardsByJob(input: DeepPartial<virtengine_hpc_v1_query.QueryRewardsByJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).rewardsByJob(input, options);
          }, { path: [22, 16] }),
          /**
           * getDispute returns a dispute by ID
           */
          getDispute: withMetadata(async function getDispute(input: DeepPartial<virtengine_hpc_v1_query.QueryDisputeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).dispute(input, options);
          }, { path: [22, 17] }),
          /**
           * getDisputes returns all disputes with optional filters
           */
          getDisputes: withMetadata(async function getDisputes(input: DeepPartial<virtengine_hpc_v1_query.QueryDisputesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).disputes(input, options);
          }, { path: [22, 18] }),
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_hpc_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).params(input, options);
          }, { path: [22, 19] }),
          /**
           * registerCluster registers a new HPC cluster
           */
          registerCluster: withMetadata(async function registerCluster(input: DeepSimplify<virtengine_hpc_v1_tx.MsgRegisterCluster>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).registerCluster(input, options);
          }, { path: [23, 0] }),
          /**
           * updateCluster updates an existing HPC cluster
           */
          updateCluster: withMetadata(async function updateCluster(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateCluster>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).updateCluster(input, options);
          }, { path: [23, 1] }),
          /**
           * deregisterCluster deregisters an HPC cluster
           */
          deregisterCluster: withMetadata(async function deregisterCluster(input: DeepSimplify<virtengine_hpc_v1_tx.MsgDeregisterCluster>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).deregisterCluster(input, options);
          }, { path: [23, 2] }),
          /**
           * createOffering creates a new HPC offering
           */
          createOffering: withMetadata(async function createOffering(input: DeepSimplify<virtengine_hpc_v1_tx.MsgCreateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).createOffering(input, options);
          }, { path: [23, 3] }),
          /**
           * updateOffering updates an existing HPC offering
           */
          updateOffering: withMetadata(async function updateOffering(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).updateOffering(input, options);
          }, { path: [23, 4] }),
          /**
           * submitJob submits a new HPC job
           */
          submitJob: withMetadata(async function submitJob(input: DeepSimplify<virtengine_hpc_v1_tx.MsgSubmitJob>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).submitJob(input, options);
          }, { path: [23, 5] }),
          /**
           * cancelJob cancels an HPC job
           */
          cancelJob: withMetadata(async function cancelJob(input: DeepSimplify<virtengine_hpc_v1_tx.MsgCancelJob>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).cancelJob(input, options);
          }, { path: [23, 6] }),
          /**
           * reportJobStatus reports job status from the provider daemon
           */
          reportJobStatus: withMetadata(async function reportJobStatus(input: DeepSimplify<virtengine_hpc_v1_tx.MsgReportJobStatus>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).reportJobStatus(input, options);
          }, { path: [23, 7] }),
          /**
           * updateNodeMetadata updates node metadata
           */
          updateNodeMetadata: withMetadata(async function updateNodeMetadata(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateNodeMetadata>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).updateNodeMetadata(input, options);
          }, { path: [23, 8] }),
          /**
           * flagDispute flags a dispute for moderation
           */
          flagDispute: withMetadata(async function flagDispute(input: DeepSimplify<virtengine_hpc_v1_tx.MsgFlagDispute>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).flagDispute(input, options);
          }, { path: [23, 9] }),
          /**
           * resolveDispute resolves a dispute (moderator only)
           */
          resolveDispute: withMetadata(async function resolveDispute(input: DeepSimplify<virtengine_hpc_v1_tx.MsgResolveDispute>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).resolveDispute(input, options);
          }, { path: [23, 10] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [23, 11] })
        }
      },
      market: {
        v1beta5: {
          /**
           * getOrders queries orders with filters.
           */
          getOrders: withMetadata(async function getOrders(input: DeepPartial<virtengine_market_v1beta5_query.QueryOrdersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).orders(input, options);
          }, { path: [24, 0] }),
          /**
           * getOrder queries order details.
           */
          getOrder: withMetadata(async function getOrder(input: DeepPartial<virtengine_market_v1beta5_query.QueryOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).order(input, options);
          }, { path: [24, 1] }),
          /**
           * getBids queries bids with filters.
           */
          getBids: withMetadata(async function getBids(input: DeepPartial<virtengine_market_v1beta5_query.QueryBidsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).bids(input, options);
          }, { path: [24, 2] }),
          /**
           * getBid queries bid details.
           */
          getBid: withMetadata(async function getBid(input: DeepPartial<virtengine_market_v1beta5_query.QueryBidRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).bid(input, options);
          }, { path: [24, 3] }),
          /**
           * getLeases queries leases with filters.
           */
          getLeases: withMetadata(async function getLeases(input: DeepPartial<virtengine_market_v1beta5_query.QueryLeasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).leases(input, options);
          }, { path: [24, 4] }),
          /**
           * getLease queries lease details.
           */
          getLease: withMetadata(async function getLease(input: DeepPartial<virtengine_market_v1beta5_query.QueryLeaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).lease(input, options);
          }, { path: [24, 5] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_market_v1beta5_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).params(input, options);
          }, { path: [24, 6] }),
          /**
           * createBid defines a method to create a bid given proper inputs.
           */
          createBid: withMetadata(async function createBid(input: DeepSimplify<virtengine_market_v1beta5_bidmsg.MsgCreateBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).createBid(input, options);
          }, { path: [25, 0] }),
          /**
           * closeBid defines a method to close a bid given proper inputs.
           */
          closeBid: withMetadata(async function closeBid(input: DeepSimplify<virtengine_market_v1beta5_bidmsg.MsgCloseBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).closeBid(input, options);
          }, { path: [25, 1] }),
          /**
           * withdrawLease withdraws accrued funds from the lease payment
           */
          withdrawLease: withMetadata(async function withdrawLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgWithdrawLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).withdrawLease(input, options);
          }, { path: [25, 2] }),
          /**
           * createLease creates a new lease
           */
          createLease: withMetadata(async function createLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgCreateLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).createLease(input, options);
          }, { path: [25, 3] }),
          /**
           * closeLease defines a method to close an order given proper inputs.
           */
          closeLease: withMetadata(async function closeLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgCloseLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).closeLease(input, options);
          }, { path: [25, 4] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: virtengine v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_market_v1beta5_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [25, 5] })
        }
      },
      marketplace: {
        v1: {
          /**
           * waldurCallback handles callbacks from Waldur integration
           */
          waldurCallback: withMetadata(async function waldurCallback(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgWaldurCallback>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getMsgClient(service).waldurCallback(input, options);
          }, { path: [26, 0] })
        }
      },
      mfa: {
        v1: {
          /**
           * getMFAPolicy returns the MFA policy for an account
           */
          getMFAPolicy: withMetadata(async function getMFAPolicy(input: DeepPartial<virtengine_mfa_v1_query.QueryMFAPolicyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).mFAPolicy(input, options);
          }, { path: [27, 0] }),
          /**
           * getFactorEnrollments returns all factor enrollments for an account
           */
          getFactorEnrollments: withMetadata(async function getFactorEnrollments(input: DeepPartial<virtengine_mfa_v1_query.QueryFactorEnrollmentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).factorEnrollments(input, options);
          }, { path: [27, 1] }),
          /**
           * getFactorEnrollment returns a specific factor enrollment
           */
          getFactorEnrollment: withMetadata(async function getFactorEnrollment(input: DeepPartial<virtengine_mfa_v1_query.QueryFactorEnrollmentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).factorEnrollment(input, options);
          }, { path: [27, 2] }),
          /**
           * getChallenge returns a challenge by ID
           */
          getChallenge: withMetadata(async function getChallenge(input: DeepPartial<virtengine_mfa_v1_query.QueryChallengeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).challenge(input, options);
          }, { path: [27, 3] }),
          /**
           * getPendingChallenges returns pending challenges for an account
           */
          getPendingChallenges: withMetadata(async function getPendingChallenges(input: DeepPartial<virtengine_mfa_v1_query.QueryPendingChallengesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).pendingChallenges(input, options);
          }, { path: [27, 4] }),
          /**
           * getAuthorizationSession returns an authorization session by ID
           */
          getAuthorizationSession: withMetadata(async function getAuthorizationSession(input: DeepPartial<virtengine_mfa_v1_query.QueryAuthorizationSessionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).authorizationSession(input, options);
          }, { path: [27, 5] }),
          /**
           * getTrustedDevices returns trusted devices for an account
           */
          getTrustedDevices: withMetadata(async function getTrustedDevices(input: DeepPartial<virtengine_mfa_v1_query.QueryTrustedDevicesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).trustedDevices(input, options);
          }, { path: [27, 6] }),
          /**
           * getSensitiveTxConfig returns the configuration for a sensitive tx type
           */
          getSensitiveTxConfig: withMetadata(async function getSensitiveTxConfig(input: DeepPartial<virtengine_mfa_v1_query.QuerySensitiveTxConfigRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).sensitiveTxConfig(input, options);
          }, { path: [27, 7] }),
          /**
           * getAllSensitiveTxConfigs returns all sensitive tx configurations
           */
          getAllSensitiveTxConfigs: withMetadata(async function getAllSensitiveTxConfigs(input: DeepPartial<virtengine_mfa_v1_query.QueryAllSensitiveTxConfigsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).allSensitiveTxConfigs(input, options);
          }, { path: [27, 8] }),
          /**
           * getMFARequired checks if MFA is required for a transaction
           */
          getMFARequired: withMetadata(async function getMFARequired(input: DeepPartial<virtengine_mfa_v1_query.QueryMFARequiredRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).mFARequired(input, options);
          }, { path: [27, 9] }),
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_mfa_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getClient(service).params(input, options);
          }, { path: [27, 10] }),
          /**
           * enrollFactor enrolls a new MFA factor
           */
          enrollFactor: withMetadata(async function enrollFactor(input: DeepSimplify<virtengine_mfa_v1_tx.MsgEnrollFactor>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).enrollFactor(input, options);
          }, { path: [28, 0] }),
          /**
           * revokeFactor revokes an enrolled factor
           */
          revokeFactor: withMetadata(async function revokeFactor(input: DeepSimplify<virtengine_mfa_v1_tx.MsgRevokeFactor>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).revokeFactor(input, options);
          }, { path: [28, 1] }),
          /**
           * setMFAPolicy sets the MFA policy for an account
           */
          setMFAPolicy: withMetadata(async function setMFAPolicy(input: DeepSimplify<virtengine_mfa_v1_tx.MsgSetMFAPolicy>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).setMFAPolicy(input, options);
          }, { path: [28, 2] }),
          /**
           * createChallenge creates an MFA challenge
           */
          createChallenge: withMetadata(async function createChallenge(input: DeepSimplify<virtengine_mfa_v1_tx.MsgCreateChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).createChallenge(input, options);
          }, { path: [28, 3] }),
          /**
           * verifyChallenge verifies an MFA challenge response
           */
          verifyChallenge: withMetadata(async function verifyChallenge(input: DeepSimplify<virtengine_mfa_v1_tx.MsgVerifyChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).verifyChallenge(input, options);
          }, { path: [28, 4] }),
          /**
           * addTrustedDevice adds a trusted device
           */
          addTrustedDevice: withMetadata(async function addTrustedDevice(input: DeepSimplify<virtengine_mfa_v1_tx.MsgAddTrustedDevice>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).addTrustedDevice(input, options);
          }, { path: [28, 5] }),
          /**
           * removeTrustedDevice removes a trusted device
           */
          removeTrustedDevice: withMetadata(async function removeTrustedDevice(input: DeepSimplify<virtengine_mfa_v1_tx.MsgRemoveTrustedDevice>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).removeTrustedDevice(input, options);
          }, { path: [28, 6] }),
          /**
           * updateSensitiveTxConfig updates sensitive transaction configuration (governance only)
           */
          updateSensitiveTxConfig: withMetadata(async function updateSensitiveTxConfig(input: DeepSimplify<virtengine_mfa_v1_tx.MsgUpdateSensitiveTxConfig>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).updateSensitiveTxConfig(input, options);
          }, { path: [28, 7] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_mfa_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [28, 8] })
        }
      },
      oracle: {
        v1: {
          /**
           * getPrices query prices for specific denom
           */
          getPrices: withMetadata(async function getPrices(input: DeepPartial<virtengine_oracle_v1_prices.QueryPricesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getClient(service).prices(input, options);
          }, { path: [29, 0] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_oracle_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getClient(service).params(input, options);
          }, { path: [29, 1] }),
          /**
           * getPriceFeedConfig queries the price feed configuration for a given denom.
           */
          getPriceFeedConfig: withMetadata(async function getPriceFeedConfig(input: DeepPartial<virtengine_oracle_v1_query.QueryPriceFeedConfigRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getClient(service).priceFeedConfig(input, options);
          }, { path: [29, 2] }),
          /**
           * getAggregatedPrice queries the aggregated price for a given denom.
           */
          getAggregatedPrice: withMetadata(async function getAggregatedPrice(input: DeepPartial<virtengine_oracle_v1_query.QueryAggregatedPriceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getClient(service).aggregatedPrice(input, options);
          }, { path: [29, 3] }),
          /**
           * addPriceEntry adds a new price entry for a denomination from an authorized source
           */
          addPriceEntry: withMetadata(async function addPriceEntry(input: DeepSimplify<virtengine_oracle_v1_msgs.MsgAddPriceEntry>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getMsgClient(service).addPriceEntry(input, options);
          }, { path: [30, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/wasm module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v2.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_oracle_v1_msgs.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [30, 1] })
        }
      },
      provider: {
        v1beta4: {
          /**
           * getProviders queries providers
           */
          getProviders: withMetadata(async function getProviders(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProvidersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getClient(service).providers(input, options);
          }, { path: [31, 0] }),
          /**
           * getProvider queries provider details
           */
          getProvider: withMetadata(async function getProvider(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getClient(service).provider(input, options);
          }, { path: [31, 1] }),
          /**
           * createProvider defines a method that creates a provider given the proper inputs.
           */
          createProvider: withMetadata(async function createProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgCreateProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getMsgClient(service).createProvider(input, options);
          }, { path: [32, 0] }),
          /**
           * updateProvider defines a method that updates a provider given the proper inputs.
           */
          updateProvider: withMetadata(async function updateProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgUpdateProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getMsgClient(service).updateProvider(input, options);
          }, { path: [32, 1] }),
          /**
           * deleteProvider defines a method that deletes a provider given the proper inputs.
           */
          deleteProvider: withMetadata(async function deleteProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgDeleteProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getMsgClient(service).deleteProvider(input, options);
          }, { path: [32, 2] }),
          /**
           * generateDomainVerificationToken generates a verification token for a provider's domain.
           */
          generateDomainVerificationToken: withMetadata(async function generateDomainVerificationToken(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgGenerateDomainVerificationToken>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getMsgClient(service).generateDomainVerificationToken(input, options);
          }, { path: [32, 3] }),
          /**
           * verifyProviderDomain verifies a provider's domain via DNS TXT record.
           */
          verifyProviderDomain: withMetadata(async function verifyProviderDomain(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgVerifyProviderDomain>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getMsgClient(service).verifyProviderDomain(input, options);
          }, { path: [32, 4] })
        }
      },
      review: {
        v1: {
          /**
           * submitReview submits a new review
           */
          submitReview: withMetadata(async function submitReview(input: DeepSimplify<virtengine_review_v1_tx.MsgSubmitReview>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).submitReview(input, options);
          }, { path: [33, 0] }),
          /**
           * deleteReview deletes a review
           */
          deleteReview: withMetadata(async function deleteReview(input: DeepSimplify<virtengine_review_v1_tx.MsgDeleteReview>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).deleteReview(input, options);
          }, { path: [33, 1] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_review_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [33, 2] })
        }
      },
      roles: {
        v1: {
          /**
           * getAccountRoles queries all roles for an account
           */
          getAccountRoles: withMetadata(async function getAccountRoles(input: DeepPartial<virtengine_roles_v1_query.QueryAccountRolesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).accountRoles(input, options);
          }, { path: [34, 0] }),
          /**
           * getRoleMembers queries all members of a specific role
           */
          getRoleMembers: withMetadata(async function getRoleMembers(input: DeepPartial<virtengine_roles_v1_query.QueryRoleMembersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).roleMembers(input, options);
          }, { path: [34, 1] }),
          /**
           * getAccountState queries the state of an account
           */
          getAccountState: withMetadata(async function getAccountState(input: DeepPartial<virtengine_roles_v1_query.QueryAccountStateRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).accountState(input, options);
          }, { path: [34, 2] }),
          /**
           * getGenesisAccounts queries all genesis accounts
           */
          getGenesisAccounts: withMetadata(async function getGenesisAccounts(input: DeepPartial<virtengine_roles_v1_query.QueryGenesisAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).genesisAccounts(input, options);
          }, { path: [34, 3] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_roles_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).params(input, options);
          }, { path: [34, 4] }),
          /**
           * getHasRole checks if an account has a specific role
           */
          getHasRole: withMetadata(async function getHasRole(input: DeepPartial<virtengine_roles_v1_query.QueryHasRoleRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).hasRole(input, options);
          }, { path: [34, 5] }),
          /**
           * assignRole assigns a role to an account
           */
          assignRole: withMetadata(async function assignRole(input: DeepSimplify<virtengine_roles_v1_tx.MsgAssignRole>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).assignRole(input, options);
          }, { path: [35, 0] }),
          /**
           * revokeRole revokes a role from an account
           */
          revokeRole: withMetadata(async function revokeRole(input: DeepSimplify<virtengine_roles_v1_tx.MsgRevokeRole>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).revokeRole(input, options);
          }, { path: [35, 1] }),
          /**
           * setAccountState sets the state of an account
           */
          setAccountState: withMetadata(async function setAccountState(input: DeepSimplify<virtengine_roles_v1_tx.MsgSetAccountState>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).setAccountState(input, options);
          }, { path: [35, 2] }),
          /**
           * nominateAdmin nominates an administrator (GenesisAccount only)
           */
          nominateAdmin: withMetadata(async function nominateAdmin(input: DeepSimplify<virtengine_roles_v1_tx.MsgNominateAdmin>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).nominateAdmin(input, options);
          }, { path: [35, 3] }),
          /**
           * updateParams updates the module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_roles_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [35, 4] })
        }
      },
      settlement: {
        v1: {
          /**
           * createEscrow creates a new escrow account
           */
          createEscrow: withMetadata(async function createEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgCreateEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).createEscrow(input, options);
          }, { path: [36, 0] }),
          /**
           * activateEscrow activates an escrow when a lease is created
           */
          activateEscrow: withMetadata(async function activateEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgActivateEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).activateEscrow(input, options);
          }, { path: [36, 1] }),
          /**
           * releaseEscrow releases escrow funds to the recipient
           */
          releaseEscrow: withMetadata(async function releaseEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgReleaseEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).releaseEscrow(input, options);
          }, { path: [36, 2] }),
          /**
           * refundEscrow refunds escrow funds to the depositor
           */
          refundEscrow: withMetadata(async function refundEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgRefundEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).refundEscrow(input, options);
          }, { path: [36, 3] }),
          /**
           * disputeEscrow marks an escrow as disputed
           */
          disputeEscrow: withMetadata(async function disputeEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgDisputeEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).disputeEscrow(input, options);
          }, { path: [36, 4] }),
          /**
           * settleOrder settles an order based on usage records
           */
          settleOrder: withMetadata(async function settleOrder(input: DeepSimplify<virtengine_settlement_v1_tx.MsgSettleOrder>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).settleOrder(input, options);
          }, { path: [36, 5] }),
          /**
           * recordUsage records usage from a provider
           */
          recordUsage: withMetadata(async function recordUsage(input: DeepSimplify<virtengine_settlement_v1_tx.MsgRecordUsage>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).recordUsage(input, options);
          }, { path: [36, 6] }),
          /**
           * acknowledgeUsage acknowledges a usage record
           */
          acknowledgeUsage: withMetadata(async function acknowledgeUsage(input: DeepSimplify<virtengine_settlement_v1_tx.MsgAcknowledgeUsage>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).acknowledgeUsage(input, options);
          }, { path: [36, 7] }),
          /**
           * claimRewards claims accumulated rewards
           */
          claimRewards: withMetadata(async function claimRewards(input: DeepSimplify<virtengine_settlement_v1_tx.MsgClaimRewards>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getMsgClient(service).claimRewards(input, options);
          }, { path: [36, 8] })
        }
      },
      staking: {
        v1: {
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_staking_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [37, 0] }),
          /**
           * slashValidator slashes a validator for misbehavior
           */
          slashValidator: withMetadata(async function slashValidator(input: DeepSimplify<virtengine_staking_v1_tx.MsgSlashValidator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).slashValidator(input, options);
          }, { path: [37, 1] }),
          /**
           * unjailValidator unjails a validator
           */
          unjailValidator: withMetadata(async function unjailValidator(input: DeepSimplify<virtengine_staking_v1_tx.MsgUnjailValidator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).unjailValidator(input, options);
          }, { path: [37, 2] }),
          /**
           * recordPerformance records validator performance metrics
           */
          recordPerformance: withMetadata(async function recordPerformance(input: DeepSimplify<virtengine_staking_v1_tx.MsgRecordPerformance>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).recordPerformance(input, options);
          }, { path: [37, 3] })
        }
      },
      take: {
        v1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_take_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).params(input, options);
          }, { path: [38, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_take_v1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [39, 0] })
        }
      },
      veid: {
        v1: {
          /**
           * getIdentityRecord queries an identity record by account address
           */
          getIdentityRecord: withMetadata(async function getIdentityRecord(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityRecordRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).identityRecord(input, options);
          }, { path: [40, 0] }),
          /**
           * getIdentity queries an identity record by account address (alias for IdentityRecord)
           */
          getIdentity: withMetadata(async function getIdentity(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).identity(input, options);
          }, { path: [40, 1] }),
          /**
           * getScope queries a specific scope by ID
           */
          getScope: withMetadata(async function getScope(input: DeepPartial<virtengine_veid_v1_query.QueryScopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).scope(input, options);
          }, { path: [40, 2] }),
          /**
           * getScopes queries all scopes for an account
           */
          getScopes: withMetadata(async function getScopes(input: DeepPartial<virtengine_veid_v1_query.QueryScopesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).scopes(input, options);
          }, { path: [40, 3] }),
          /**
           * getScopesByType queries all scopes of a specific type for an account
           */
          getScopesByType: withMetadata(async function getScopesByType(input: DeepPartial<virtengine_veid_v1_query.QueryScopesByTypeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).scopesByType(input, options);
          }, { path: [40, 4] }),
          /**
           * getIdentityScore queries the identity score for an account
           */
          getIdentityScore: withMetadata(async function getIdentityScore(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityScoreRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).identityScore(input, options);
          }, { path: [40, 5] }),
          /**
           * getIdentityStatus queries the identity status for an account
           */
          getIdentityStatus: withMetadata(async function getIdentityStatus(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityStatusRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).identityStatus(input, options);
          }, { path: [40, 6] }),
          /**
           * getIdentityWallet queries an identity wallet by account address
           */
          getIdentityWallet: withMetadata(async function getIdentityWallet(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityWalletRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).identityWallet(input, options);
          }, { path: [40, 7] }),
          /**
           * getWalletScopes queries all scope references in a wallet
           */
          getWalletScopes: withMetadata(async function getWalletScopes(input: DeepPartial<virtengine_veid_v1_query.QueryWalletScopesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).walletScopes(input, options);
          }, { path: [40, 8] }),
          /**
           * getConsentSettings queries consent settings for an account
           */
          getConsentSettings: withMetadata(async function getConsentSettings(input: DeepPartial<virtengine_veid_v1_query.QueryConsentSettingsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).consentSettings(input, options);
          }, { path: [40, 9] }),
          /**
           * getDerivedFeatures queries derived features metadata for an account
           */
          getDerivedFeatures: withMetadata(async function getDerivedFeatures(input: DeepPartial<virtengine_veid_v1_query.QueryDerivedFeaturesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).derivedFeatures(input, options);
          }, { path: [40, 10] }),
          /**
           * getDerivedFeatureHashes queries derived feature hashes for an account (consent-gated)
           */
          getDerivedFeatureHashes: withMetadata(async function getDerivedFeatureHashes(input: DeepPartial<virtengine_veid_v1_query.QueryDerivedFeatureHashesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).derivedFeatureHashes(input, options);
          }, { path: [40, 11] }),
          /**
           * getVerificationHistory queries verification history for an account
           */
          getVerificationHistory: withMetadata(async function getVerificationHistory(input: DeepPartial<virtengine_veid_v1_query.QueryVerificationHistoryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).verificationHistory(input, options);
          }, { path: [40, 12] }),
          /**
           * getApprovedClients queries all approved clients
           */
          getApprovedClients: withMetadata(async function getApprovedClients(input: DeepPartial<virtengine_veid_v1_query.QueryApprovedClientsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).approvedClients(input, options);
          }, { path: [40, 13] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_veid_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).params(input, options);
          }, { path: [40, 14] }),
          /**
           * getBorderlineParams queries the borderline parameters
           */
          getBorderlineParams: withMetadata(async function getBorderlineParams(input: DeepPartial<virtengine_veid_v1_query.QueryBorderlineParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).borderlineParams(input, options);
          }, { path: [40, 15] }),
          /**
           * getAppeal queries a specific appeal by ID
           */
          getAppeal: withMetadata(async function getAppeal(input: DeepPartial<virtengine_veid_v1_query.QueryAppealRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).appeal(input, options);
          }, { path: [40, 16] }),
          /**
           * getAppeals queries all appeals for an account
           */
          getAppeals: withMetadata(async function getAppeals(input: DeepPartial<virtengine_veid_v1_query.QueryAppealsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).appeals(input, options);
          }, { path: [40, 17] }),
          /**
           * getAppealsByScope queries all appeals for a specific scope
           */
          getAppealsByScope: withMetadata(async function getAppealsByScope(input: DeepPartial<virtengine_veid_v1_query.QueryAppealsByScopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).appealsByScope(input, options);
          }, { path: [40, 18] }),
          /**
           * getAppealParams queries the appeal system parameters
           */
          getAppealParams: withMetadata(async function getAppealParams(input: DeepPartial<virtengine_veid_v1_query.QueryAppealParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).appealParams(input, options);
          }, { path: [40, 19] }),
          /**
           * getComplianceStatus queries the compliance status for an account
           */
          getComplianceStatus: withMetadata(async function getComplianceStatus(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceStatusRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).complianceStatus(input, options);
          }, { path: [40, 20] }),
          /**
           * getComplianceProvider queries a specific compliance provider
           */
          getComplianceProvider: withMetadata(async function getComplianceProvider(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).complianceProvider(input, options);
          }, { path: [40, 21] }),
          /**
           * getComplianceProviders queries all compliance providers
           */
          getComplianceProviders: withMetadata(async function getComplianceProviders(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceProvidersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).complianceProviders(input, options);
          }, { path: [40, 22] }),
          /**
           * getComplianceParams queries the compliance parameters
           */
          getComplianceParams: withMetadata(async function getComplianceParams(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).complianceParams(input, options);
          }, { path: [40, 23] }),
          /**
           * getModelVersion queries the active model for a given type
           */
          getModelVersion: withMetadata(async function getModelVersion(input: DeepPartial<virtengine_veid_v1_query.QueryModelVersionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).modelVersion(input, options);
          }, { path: [40, 24] }),
          /**
           * getActiveModels queries all active models
           */
          getActiveModels: withMetadata(async function getActiveModels(input: DeepPartial<virtengine_veid_v1_query.QueryActiveModelsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).activeModels(input, options);
          }, { path: [40, 25] }),
          /**
           * getModelHistory queries the version history for a model type
           */
          getModelHistory: withMetadata(async function getModelHistory(input: DeepPartial<virtengine_veid_v1_query.QueryModelHistoryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).modelHistory(input, options);
          }, { path: [40, 26] }),
          /**
           * getValidatorModelSync queries a validator's model sync status
           */
          getValidatorModelSync: withMetadata(async function getValidatorModelSync(input: DeepPartial<virtengine_veid_v1_query.QueryValidatorModelSyncRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).validatorModelSync(input, options);
          }, { path: [40, 27] }),
          /**
           * getModelParams queries the model management parameters
           */
          getModelParams: withMetadata(async function getModelParams(input: DeepPartial<virtengine_veid_v1_query.QueryModelParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).modelParams(input, options);
          }, { path: [40, 28] }),
          /**
           * uploadScope uploads an identity scope
           */
          uploadScope: withMetadata(async function uploadScope(input: DeepSimplify<virtengine_veid_v1_tx.MsgUploadScope>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).uploadScope(input, options);
          }, { path: [41, 0] }),
          /**
           * revokeScope revokes an identity scope
           */
          revokeScope: withMetadata(async function revokeScope(input: DeepSimplify<virtengine_veid_v1_tx.MsgRevokeScope>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).revokeScope(input, options);
          }, { path: [41, 1] }),
          /**
           * requestVerification requests verification of a scope
           */
          requestVerification: withMetadata(async function requestVerification(input: DeepSimplify<virtengine_veid_v1_tx.MsgRequestVerification>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).requestVerification(input, options);
          }, { path: [41, 2] }),
          /**
           * updateVerificationStatus updates the verification status (validator only)
           */
          updateVerificationStatus: withMetadata(async function updateVerificationStatus(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateVerificationStatus>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateVerificationStatus(input, options);
          }, { path: [41, 3] }),
          /**
           * updateScore updates the identity score (validator only)
           */
          updateScore: withMetadata(async function updateScore(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateScore>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateScore(input, options);
          }, { path: [41, 4] }),
          /**
           * createIdentityWallet creates an identity wallet
           */
          createIdentityWallet: withMetadata(async function createIdentityWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgCreateIdentityWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).createIdentityWallet(input, options);
          }, { path: [41, 5] }),
          /**
           * addScopeToWallet adds a scope reference to a wallet
           */
          addScopeToWallet: withMetadata(async function addScopeToWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgAddScopeToWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).addScopeToWallet(input, options);
          }, { path: [41, 6] }),
          /**
           * revokeScopeFromWallet revokes a scope from a wallet
           */
          revokeScopeFromWallet: withMetadata(async function revokeScopeFromWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgRevokeScopeFromWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).revokeScopeFromWallet(input, options);
          }, { path: [41, 7] }),
          /**
           * updateConsentSettings updates consent settings
           */
          updateConsentSettings: withMetadata(async function updateConsentSettings(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateConsentSettings>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateConsentSettings(input, options);
          }, { path: [41, 8] }),
          /**
           * rebindWallet rebinds a wallet during key rotation
           */
          rebindWallet: withMetadata(async function rebindWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgRebindWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).rebindWallet(input, options);
          }, { path: [41, 9] }),
          /**
           * updateDerivedFeatures updates derived features (validator only)
           */
          updateDerivedFeatures: withMetadata(async function updateDerivedFeatures(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateDerivedFeatures>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateDerivedFeatures(input, options);
          }, { path: [41, 10] }),
          /**
           * completeBorderlineFallback completes a borderline fallback after MFA
           */
          completeBorderlineFallback: withMetadata(async function completeBorderlineFallback(input: DeepSimplify<virtengine_veid_v1_tx.MsgCompleteBorderlineFallback>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).completeBorderlineFallback(input, options);
          }, { path: [41, 11] }),
          /**
           * updateBorderlineParams updates borderline parameters (governance only)
           */
          updateBorderlineParams: withMetadata(async function updateBorderlineParams(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateBorderlineParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateBorderlineParams(input, options);
          }, { path: [41, 12] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [41, 13] }),
          /**
           * submitAppeal submits an appeal against a verification decision
           */
          submitAppeal: withMetadata(async function submitAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgSubmitAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).submitAppeal(input, options);
          }, { path: [41, 14] }),
          /**
           * claimAppeal allows an arbitrator to claim an appeal for review
           */
          claimAppeal: withMetadata(async function claimAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgClaimAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).claimAppeal(input, options);
          }, { path: [41, 15] }),
          /**
           * resolveAppeal resolves an appeal (arbitrator/governance only)
           */
          resolveAppeal: withMetadata(async function resolveAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgResolveAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).resolveAppeal(input, options);
          }, { path: [41, 16] }),
          /**
           * withdrawAppeal allows the submitter to withdraw their appeal
           */
          withdrawAppeal: withMetadata(async function withdrawAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgWithdrawAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).withdrawAppeal(input, options);
          }, { path: [41, 17] }),
          /**
           * submitComplianceCheck submits external compliance check results
           */
          submitComplianceCheck: withMetadata(async function submitComplianceCheck(input: DeepSimplify<virtengine_veid_v1_compliance.MsgSubmitComplianceCheck>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).submitComplianceCheck(input, options);
          }, { path: [41, 18] }),
          /**
           * attestCompliance allows validators to attest compliance status
           */
          attestCompliance: withMetadata(async function attestCompliance(input: DeepSimplify<virtengine_veid_v1_compliance.MsgAttestCompliance>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).attestCompliance(input, options);
          }, { path: [41, 19] }),
          /**
           * updateComplianceParams updates compliance configuration (gov only)
           */
          updateComplianceParams: withMetadata(async function updateComplianceParams(input: DeepSimplify<virtengine_veid_v1_compliance.MsgUpdateComplianceParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateComplianceParams(input, options);
          }, { path: [41, 20] }),
          /**
           * registerComplianceProvider registers a new compliance provider (gov only)
           */
          registerComplianceProvider: withMetadata(async function registerComplianceProvider(input: DeepSimplify<virtengine_veid_v1_compliance.MsgRegisterComplianceProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).registerComplianceProvider(input, options);
          }, { path: [41, 21] }),
          /**
           * deactivateComplianceProvider deactivates a compliance provider (gov only)
           */
          deactivateComplianceProvider: withMetadata(async function deactivateComplianceProvider(input: DeepSimplify<virtengine_veid_v1_compliance.MsgDeactivateComplianceProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).deactivateComplianceProvider(input, options);
          }, { path: [41, 22] }),
          /**
           * registerModel registers a new ML model (authorized only)
           */
          registerModel: withMetadata(async function registerModel(input: DeepSimplify<virtengine_veid_v1_model.MsgRegisterModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).registerModel(input, options);
          }, { path: [41, 23] }),
          /**
           * proposeModelUpdate proposes updating active model via governance
           */
          proposeModelUpdate: withMetadata(async function proposeModelUpdate(input: DeepSimplify<virtengine_veid_v1_model.MsgProposeModelUpdate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).proposeModelUpdate(input, options);
          }, { path: [41, 24] }),
          /**
           * reportModelVersion reports validator's model versions
           */
          reportModelVersion: withMetadata(async function reportModelVersion(input: DeepSimplify<virtengine_veid_v1_model.MsgReportModelVersion>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).reportModelVersion(input, options);
          }, { path: [41, 25] }),
          /**
           * activateModel activates a pending model after governance approval
           */
          activateModel: withMetadata(async function activateModel(input: DeepSimplify<virtengine_veid_v1_model.MsgActivateModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).activateModel(input, options);
          }, { path: [41, 26] }),
          /**
           * deprecateModel deprecates a model
           */
          deprecateModel: withMetadata(async function deprecateModel(input: DeepSimplify<virtengine_veid_v1_model.MsgDeprecateModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).deprecateModel(input, options);
          }, { path: [41, 27] }),
          /**
           * revokeModel revokes a model
           */
          revokeModel: withMetadata(async function revokeModel(input: DeepSimplify<virtengine_veid_v1_model.MsgRevokeModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).revokeModel(input, options);
          }, { path: [41, 28] })
        }
      },
      wasm: {
        v1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_wasm_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).params(input, options);
          }, { path: [42, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/wasm module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v2.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_wasm_v1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [43, 0] })
        }
      }
    }
  };
}
