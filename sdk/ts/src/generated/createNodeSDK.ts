import { createServiceLoader } from "../sdk/client/createServiceLoader.ts";
import { SDKOptions } from "../sdk/types.ts";

import type * as virtengine_audit_v1_msg_audit_log from "./protos/virtengine/audit/v1/msg_audit_log.ts";
import type * as virtengine_audit_v1_query from "./protos/virtengine/audit/v1/query.ts";
import type * as virtengine_audit_v1_query_audit_log from "./protos/virtengine/audit/v1/query_audit_log.ts";
import type * as virtengine_audit_v1_msg from "./protos/virtengine/audit/v1/msg.ts";
import type * as virtengine_benchmark_v1_query from "./protos/virtengine/benchmark/v1/query.ts";
import type * as virtengine_benchmark_v1_tx from "./protos/virtengine/benchmark/v1/tx.ts";
import type * as virtengine_bme_v1_query from "./protos/virtengine/bme/v1/query.ts";
import type * as virtengine_bme_v1_msgs from "./protos/virtengine/bme/v1/msgs.ts";
import type * as virtengine_cert_v1_query from "./protos/virtengine/cert/v1/query.ts";
import type * as virtengine_cert_v1_msg from "./protos/virtengine/cert/v1/msg.ts";
import type * as virtengine_config_v1_tx from "./protos/virtengine/config/v1/tx.ts";
import type * as virtengine_delegation_v1_tx from "./protos/virtengine/delegation/v1/tx.ts";
import type * as virtengine_delegation_v1_query from "./protos/virtengine/delegation/v1/query.ts";
import type * as virtengine_deployment_v1beta4_query from "./protos/virtengine/deployment/v1beta4/query.ts";
import type * as virtengine_deployment_v1beta4_deploymentmsg from "./protos/virtengine/deployment/v1beta4/deploymentmsg.ts";
import type * as virtengine_deployment_v1beta4_groupmsg from "./protos/virtengine/deployment/v1beta4/groupmsg.ts";
import type * as virtengine_deployment_v1beta4_paramsmsg from "./protos/virtengine/deployment/v1beta4/paramsmsg.ts";
import type * as virtengine_deployment_v1beta5_query from "./protos/virtengine/deployment/v1beta5/query.ts";
import type * as virtengine_deployment_v1beta5_deploymentmsg from "./protos/virtengine/deployment/v1beta5/deploymentmsg.ts";
import type * as virtengine_deployment_v1beta5_groupmsg from "./protos/virtengine/deployment/v1beta5/groupmsg.ts";
import type * as virtengine_deployment_v1beta5_paramsmsg from "./protos/virtengine/deployment/v1beta5/paramsmsg.ts";
import type * as virtengine_downtimedetector_v1beta1_query from "./protos/virtengine/downtimedetector/v1beta1/query.ts";
import type * as virtengine_enclave_v1_query from "./protos/virtengine/enclave/v1/query.ts";
import type * as virtengine_enclave_v1_tx from "./protos/virtengine/enclave/v1/tx.ts";
import type * as virtengine_encryption_v1_query from "./protos/virtengine/encryption/v1/query.ts";
import type * as virtengine_encryption_v1_tx from "./protos/virtengine/encryption/v1/tx.ts";
import type * as virtengine_epochs_v1beta1_query from "./protos/virtengine/epochs/v1beta1/query.ts";
import type * as virtengine_escrow_v1_query from "./protos/virtengine/escrow/v1/query.ts";
import type * as virtengine_escrow_v1_msg from "./protos/virtengine/escrow/v1/msg.ts";
import type * as virtengine_escrow_v1beta3_query from "./protos/virtengine/escrow/v1beta3/query.ts";
import type * as virtengine_fraud_v1_query from "./protos/virtengine/fraud/v1/query.ts";
import type * as virtengine_fraud_v1_tx from "./protos/virtengine/fraud/v1/tx.ts";
import type * as virtengine_hpc_v1_query from "./protos/virtengine/hpc/v1/query.ts";
import type * as virtengine_hpc_v1_tx from "./protos/virtengine/hpc/v1/tx.ts";
import type * as virtengine_market_v1beta5_query from "./protos/virtengine/market/v1beta5/query.ts";
import type * as virtengine_market_v1beta5_bidmsg from "./protos/virtengine/market/v1beta5/bidmsg.ts";
import type * as virtengine_market_v1beta5_leasemsg from "./protos/virtengine/market/v1beta5/leasemsg.ts";
import type * as virtengine_market_v1beta5_paramsmsg from "./protos/virtengine/market/v1beta5/paramsmsg.ts";
import type * as virtengine_market_v2beta1_query from "./protos/virtengine/market/v2beta1/query.ts";
import type * as virtengine_market_v2beta1_bidmsg from "./protos/virtengine/market/v2beta1/bidmsg.ts";
import type * as virtengine_market_v2beta1_leasemsg from "./protos/virtengine/market/v2beta1/leasemsg.ts";
import type * as virtengine_market_v2beta1_paramsmsg from "./protos/virtengine/market/v2beta1/paramsmsg.ts";
import type * as virtengine_marketplace_v1_query from "./protos/virtengine/marketplace/v1/query.ts";
import type * as virtengine_marketplace_v1_tx from "./protos/virtengine/marketplace/v1/tx.ts";
import type * as virtengine_mfa_v1_query from "./protos/virtengine/mfa/v1/query.ts";
import type * as virtengine_mfa_v1_tx from "./protos/virtengine/mfa/v1/tx.ts";
import type * as virtengine_oracle_v1_prices from "./protos/virtengine/oracle/v1/prices.ts";
import type * as virtengine_oracle_v1_query from "./protos/virtengine/oracle/v1/query.ts";
import type * as virtengine_oracle_v1_msgs from "./protos/virtengine/oracle/v1/msgs.ts";
import type * as virtengine_provider_v1beta4_query from "./protos/virtengine/provider/v1beta4/query.ts";
import type * as virtengine_provider_v1beta4_msg from "./protos/virtengine/provider/v1beta4/msg.ts";
import type * as virtengine_resources_v1_query from "./protos/virtengine/resources/v1/query.ts";
import type * as virtengine_resources_v1_tx from "./protos/virtengine/resources/v1/tx.ts";
import type * as virtengine_review_v1_tx from "./protos/virtengine/review/v1/tx.ts";
import type * as virtengine_review_v1_query from "./protos/virtengine/review/v1/query.ts";
import type * as virtengine_roles_v1_query from "./protos/virtengine/roles/v1/query.ts";
import type * as virtengine_roles_v1_tx from "./protos/virtengine/roles/v1/tx.ts";
import type * as virtengine_settlement_v1_query from "./protos/virtengine/settlement/v1/query.ts";
import type * as virtengine_settlement_v1_tx from "./protos/virtengine/settlement/v1/tx.ts";
import type * as virtengine_staking_v1_query from "./protos/virtengine/staking/v1/query.ts";
import type * as virtengine_staking_v1_tx from "./protos/virtengine/staking/v1/tx.ts";
import type * as virtengine_support_v1_query from "./protos/virtengine/support/v1/query.ts";
import type * as virtengine_support_v1_tx from "./protos/virtengine/support/v1/tx.ts";
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
  () => import("./protos/virtengine/audit/v1/msg_audit_log_virtengine.ts").then(m => m.MsgService),
  () => import("./protos/virtengine/audit/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/audit/v1/query_audit_log_virtengine.ts").then(m => m.QueryAuditLog),
  () => import("./protos/virtengine/audit/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/benchmark/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/benchmark/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/bme/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/bme/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/cert/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/cert/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/config/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/delegation/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/delegation/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/deployment/v1beta4/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/deployment/v1beta4/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/deployment/v1beta5/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/deployment/v1beta5/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/downtimedetector/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/enclave/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/enclave/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/encryption/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/encryption/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/epochs/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/escrow/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/escrow/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/escrow/v1beta3/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/fraud/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/fraud/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/hpc/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/hpc/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/market/v1beta5/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/market/v1beta5/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/market/v2beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/market/v2beta1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/marketplace/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/marketplace/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/mfa/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/mfa/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/oracle/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/oracle/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/provider/v1beta4/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/provider/v1beta4/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/resources/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/resources/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/review/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/review/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/roles/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/roles/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/settlement/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/settlement/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/staking/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/staking/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/support/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/support/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/take/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/virtengine/take/v1/service_virtengine.ts").then(m => m.Msg),
  () => import("./protos/virtengine/take/v1beta3/query_virtengine.ts").then(m => m.Query),
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
           * createExportJob creates a new export job for audit logs
           */
          createExportJob: withMetadata(async function createExportJob(input: DeepSimplify<virtengine_audit_v1_msg_audit_log.MsgCreateExportJob>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getMsgClient(service).createExportJob(input, options);
          }, { path: [0, 0] }),
          /**
           * updateParams updates the audit log module parameters
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_audit_v1_msg_audit_log.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [0, 1] }),
          /**
           * getAllProvidersAttributes queries all providers.
           */
          getAllProvidersAttributes: withMetadata(async function getAllProvidersAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryAllProvidersAttributesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).allProvidersAttributes(input, options);
          }, { path: [1, 0] }),
          /**
           * getProviderAttributes queries all provider signed attributes.
           */
          getProviderAttributes: withMetadata(async function getProviderAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryProviderAttributesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).providerAttributes(input, options);
          }, { path: [1, 1] }),
          /**
           * getProviderAuditorAttributes queries provider signed attributes by specific auditor.
           */
          getProviderAuditorAttributes: withMetadata(async function getProviderAuditorAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryProviderAuditorRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).providerAuditorAttributes(input, options);
          }, { path: [1, 2] }),
          /**
           * getAuditorAttributes queries all providers signed by this auditor.
           */
          getAuditorAttributes: withMetadata(async function getAuditorAttributes(input: DeepPartial<virtengine_audit_v1_query.QueryAuditorAttributesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).auditorAttributes(input, options);
          }, { path: [1, 3] }),
          /**
           * getQueryLogEntries queries all audit log entries with optional filters
           */
          getQueryLogEntries: withMetadata(async function getQueryLogEntries(input: DeepPartial<virtengine_audit_v1_query_audit_log.QueryLogEntriesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).queryLogEntries(input, options);
          }, { path: [2, 0] }),
          /**
           * getQueryLogEntry queries a specific audit log entry by ID
           */
          getQueryLogEntry: withMetadata(async function getQueryLogEntry(input: DeepPartial<virtengine_audit_v1_query_audit_log.QueryLogEntryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).queryLogEntry(input, options);
          }, { path: [2, 1] }),
          /**
           * getQueryExportJobs queries all export jobs
           */
          getQueryExportJobs: withMetadata(async function getQueryExportJobs(input: DeepPartial<virtengine_audit_v1_query_audit_log.QueryExportJobsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).queryExportJobs(input, options);
          }, { path: [2, 2] }),
          /**
           * getQueryExportJob queries a specific export job by ID
           */
          getQueryExportJob: withMetadata(async function getQueryExportJob(input: DeepPartial<virtengine_audit_v1_query_audit_log.QueryExportJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).queryExportJob(input, options);
          }, { path: [2, 3] }),
          /**
           * getQueryParams queries the audit log module parameters
           */
          getQueryParams: withMetadata(async function getQueryParams(input: DeepPartial<virtengine_audit_v1_query_audit_log.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getClient(service).queryParams(input, options);
          }, { path: [2, 4] }),
          /**
           * signProviderAttributes defines a method that signs provider attributes.
           */
          signProviderAttributes: withMetadata(async function signProviderAttributes(input: DeepSimplify<virtengine_audit_v1_msg.MsgSignProviderAttributes>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getMsgClient(service).signProviderAttributes(input, options);
          }, { path: [3, 0] }),
          /**
           * deleteProviderAttributes defines a method that deletes provider attributes.
           */
          deleteProviderAttributes: withMetadata(async function deleteProviderAttributes(input: DeepSimplify<virtengine_audit_v1_msg.MsgDeleteProviderAttributes>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getMsgClient(service).deleteProviderAttributes(input, options);
          }, { path: [3, 1] })
        }
      },
      benchmark: {
        v1: {
          /**
           * getBenchmark queries a single benchmark report by ID.
           */
          getBenchmark: withMetadata(async function getBenchmark(input: DeepPartial<virtengine_benchmark_v1_query.QueryBenchmarkRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getClient(service).benchmark(input, options);
          }, { path: [4, 0] }),
          /**
           * getBenchmarksByProvider queries benchmark reports by provider address.
           */
          getBenchmarksByProvider: withMetadata(async function getBenchmarksByProvider(input: DeepPartial<virtengine_benchmark_v1_query.QueryBenchmarksByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getClient(service).benchmarksByProvider(input, options);
          }, { path: [4, 1] }),
          /**
           * getParams returns the module parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_benchmark_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getClient(service).params(input, options);
          }, { path: [4, 2] }),
          /**
           * submitBenchmarks submits benchmark results
           */
          submitBenchmarks: withMetadata(async function submitBenchmarks(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgSubmitBenchmarks>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).submitBenchmarks(input, options);
          }, { path: [5, 0] }),
          /**
           * requestChallenge requests a benchmark challenge
           */
          requestChallenge: withMetadata(async function requestChallenge(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgRequestChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).requestChallenge(input, options);
          }, { path: [5, 1] }),
          /**
           * respondChallenge responds to a benchmark challenge
           */
          respondChallenge: withMetadata(async function respondChallenge(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgRespondChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).respondChallenge(input, options);
          }, { path: [5, 2] }),
          /**
           * flagProvider flags a provider for anomaly
           */
          flagProvider: withMetadata(async function flagProvider(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgFlagProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).flagProvider(input, options);
          }, { path: [5, 3] }),
          /**
           * unflagProvider removes a provider flag
           */
          unflagProvider: withMetadata(async function unflagProvider(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgUnflagProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).unflagProvider(input, options);
          }, { path: [5, 4] }),
          /**
           * resolveAnomalyFlag resolves an anomaly flag
           */
          resolveAnomalyFlag: withMetadata(async function resolveAnomalyFlag(input: DeepSimplify<virtengine_benchmark_v1_tx.MsgResolveAnomalyFlag>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getMsgClient(service).resolveAnomalyFlag(input, options);
          }, { path: [5, 5] })
        }
      },
      bme: {
        v1: {
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_bme_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).params(input, options);
          }, { path: [6, 0] }),
          /**
           * getVaultState returns the current vault state
           */
          getVaultState: withMetadata(async function getVaultState(input: DeepPartial<virtengine_bme_v1_query.QueryVaultStateRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).vaultState(input, options);
          }, { path: [6, 1] }),
          /**
           * getStatus returns the current circuit breaker status
           */
          getStatus: withMetadata(async function getStatus(input: DeepPartial<virtengine_bme_v1_query.QueryStatusRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).status(input, options);
          }, { path: [6, 2] }),
          /**
           * updateParams updates the module parameters.
           * This operation can only be performed through governance proposals.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_bme_v1_msgs.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [7, 0] }),
          /**
           * burnMint allows users to burn one token and mint another at current oracle prices.
           * Typically used to burn unused ACT tokens back to AKT.
           * The operation may be delayed or rejected based on circuit breaker status.
           */
          burnMint: withMetadata(async function burnMint(input: DeepSimplify<virtengine_bme_v1_msgs.MsgBurnMint>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).burnMint(input, options);
          }, { path: [7, 1] }),
          /**
           * mintACT mints ACT tokens by burning the specified source token.
           * The mint amount is calculated based on current oracle prices and
           * the collateral ratio. May be halted if circuit breaker is triggered.
           */
          mintACT: withMetadata(async function mintACT(input: DeepSimplify<virtengine_bme_v1_msgs.MsgMintACT>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).mintACT(input, options);
          }, { path: [7, 2] }),
          /**
           * burnACT burns ACT tokens and mints the specified destination token.
           * The burn operation uses remint credits when available, otherwise
           * requires adequate collateral backing based on oracle prices.
           */
          burnACT: withMetadata(async function burnACT(input: DeepSimplify<virtengine_bme_v1_msgs.MsgBurnACT>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).burnACT(input, options);
          }, { path: [7, 3] })
        }
      },
      cert: {
        v1: {
          /**
           * getCertificates queries certificates on-chain.
           */
          getCertificates: withMetadata(async function getCertificates(input: DeepPartial<virtengine_cert_v1_query.QueryCertificatesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(8);
            return getClient(service).certificates(input, options);
          }, { path: [8, 0] }),
          /**
           * createCertificate defines a method to create new certificate given proper inputs.
           */
          createCertificate: withMetadata(async function createCertificate(input: DeepSimplify<virtengine_cert_v1_msg.MsgCreateCertificate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).createCertificate(input, options);
          }, { path: [9, 0] }),
          /**
           * revokeCertificate defines a method to revoke the certificate.
           */
          revokeCertificate: withMetadata(async function revokeCertificate(input: DeepSimplify<virtengine_cert_v1_msg.MsgRevokeCertificate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(9);
            return getMsgClient(service).revokeCertificate(input, options);
          }, { path: [9, 1] })
        }
      },
      config: {
        v1: {
          /**
           * registerApprovedClient registers a new approved client
           */
          registerApprovedClient: withMetadata(async function registerApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgRegisterApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getMsgClient(service).registerApprovedClient(input, options);
          }, { path: [10, 0] }),
          /**
           * updateApprovedClient updates an approved client
           */
          updateApprovedClient: withMetadata(async function updateApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgUpdateApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getMsgClient(service).updateApprovedClient(input, options);
          }, { path: [10, 1] }),
          /**
           * suspendApprovedClient suspends an approved client
           */
          suspendApprovedClient: withMetadata(async function suspendApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgSuspendApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getMsgClient(service).suspendApprovedClient(input, options);
          }, { path: [10, 2] }),
          /**
           * revokeApprovedClient revokes an approved client
           */
          revokeApprovedClient: withMetadata(async function revokeApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgRevokeApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getMsgClient(service).revokeApprovedClient(input, options);
          }, { path: [10, 3] }),
          /**
           * reactivateApprovedClient reactivates a suspended client
           */
          reactivateApprovedClient: withMetadata(async function reactivateApprovedClient(input: DeepSimplify<virtengine_config_v1_tx.MsgReactivateApprovedClient>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getMsgClient(service).reactivateApprovedClient(input, options);
          }, { path: [10, 4] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_config_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(10);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [10, 5] })
        }
      },
      delegation: {
        v1: {
          /**
           * delegate delegates tokens to a validator
           */
          delegate: withMetadata(async function delegate(input: DeepSimplify<virtengine_delegation_v1_tx.MsgDelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).delegate(input, options);
          }, { path: [11, 0] }),
          /**
           * undelegate undelegates tokens from a validator
           */
          undelegate: withMetadata(async function undelegate(input: DeepSimplify<virtengine_delegation_v1_tx.MsgUndelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).undelegate(input, options);
          }, { path: [11, 1] }),
          /**
           * redelegate redelegates tokens between validators
           */
          redelegate: withMetadata(async function redelegate(input: DeepSimplify<virtengine_delegation_v1_tx.MsgRedelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).redelegate(input, options);
          }, { path: [11, 2] }),
          /**
           * claimRewards claims rewards from a specific validator
           */
          claimRewards: withMetadata(async function claimRewards(input: DeepSimplify<virtengine_delegation_v1_tx.MsgClaimRewards>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).claimRewards(input, options);
          }, { path: [11, 3] }),
          /**
           * claimAllRewards claims rewards from all validators
           */
          claimAllRewards: withMetadata(async function claimAllRewards(input: DeepSimplify<virtengine_delegation_v1_tx.MsgClaimAllRewards>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).claimAllRewards(input, options);
          }, { path: [11, 4] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_delegation_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(11);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [11, 5] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_delegation_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).params(input, options);
          }, { path: [12, 0] }),
          /**
           * getDelegation queries a specific delegation
           */
          getDelegation: withMetadata(async function getDelegation(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).delegation(input, options);
          }, { path: [12, 1] }),
          /**
           * getDelegatorDelegations queries all delegations for a delegator
           */
          getDelegatorDelegations: withMetadata(async function getDelegatorDelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).delegatorDelegations(input, options);
          }, { path: [12, 2] }),
          /**
           * getValidatorDelegations queries all delegations for a validator
           */
          getValidatorDelegations: withMetadata(async function getValidatorDelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryValidatorDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).validatorDelegations(input, options);
          }, { path: [12, 3] }),
          /**
           * getUnbondingDelegation queries a specific unbonding delegation
           */
          getUnbondingDelegation: withMetadata(async function getUnbondingDelegation(input: DeepPartial<virtengine_delegation_v1_query.QueryUnbondingDelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).unbondingDelegation(input, options);
          }, { path: [12, 4] }),
          /**
           * getDelegatorUnbondingDelegations queries all unbonding delegations for a delegator
           */
          getDelegatorUnbondingDelegations: withMetadata(async function getDelegatorUnbondingDelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorUnbondingDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).delegatorUnbondingDelegations(input, options);
          }, { path: [12, 5] }),
          /**
           * getRedelegation queries a specific redelegation
           */
          getRedelegation: withMetadata(async function getRedelegation(input: DeepPartial<virtengine_delegation_v1_query.QueryRedelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).redelegation(input, options);
          }, { path: [12, 6] }),
          /**
           * getDelegatorRedelegations queries all redelegations for a delegator
           */
          getDelegatorRedelegations: withMetadata(async function getDelegatorRedelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorRedelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).delegatorRedelegations(input, options);
          }, { path: [12, 7] }),
          /**
           * getRedelegations queries all active redelegations
           */
          getRedelegations: withMetadata(async function getRedelegations(input: DeepPartial<virtengine_delegation_v1_query.QueryRedelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).redelegations(input, options);
          }, { path: [12, 8] }),
          /**
           * getDelegatorRewards queries unclaimed rewards for a delegator from a specific validator
           */
          getDelegatorRewards: withMetadata(async function getDelegatorRewards(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).delegatorRewards(input, options);
          }, { path: [12, 9] }),
          /**
           * getHistoricalRewards queries historical rewards for a delegator from a validator
           */
          getHistoricalRewards: withMetadata(async function getHistoricalRewards(input: DeepPartial<virtengine_delegation_v1_query.QueryHistoricalRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).historicalRewards(input, options);
          }, { path: [12, 10] }),
          /**
           * getSlashingEvents queries slashing events for a delegator
           */
          getSlashingEvents: withMetadata(async function getSlashingEvents(input: DeepPartial<virtengine_delegation_v1_query.QuerySlashingEventsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).slashingEvents(input, options);
          }, { path: [12, 11] }),
          /**
           * getDelegatorAllRewards queries all unclaimed rewards for a delegator
           */
          getDelegatorAllRewards: withMetadata(async function getDelegatorAllRewards(input: DeepPartial<virtengine_delegation_v1_query.QueryDelegatorAllRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).delegatorAllRewards(input, options);
          }, { path: [12, 12] }),
          /**
           * getValidatorShares queries the total shares for a validator
           */
          getValidatorShares: withMetadata(async function getValidatorShares(input: DeepPartial<virtengine_delegation_v1_query.QueryValidatorSharesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(12);
            return getClient(service).validatorShares(input, options);
          }, { path: [12, 13] })
        }
      },
      deployment: {
        v1beta4: {
          /**
           * getDeployments queries deployments.
           */
          getDeployments: withMetadata(async function getDeployments(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryDeploymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).deployments(input, options);
          }, { path: [13, 0] }),
          /**
           * getDeployment queries deployment details.
           */
          getDeployment: withMetadata(async function getDeployment(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryDeploymentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).deployment(input, options);
          }, { path: [13, 1] }),
          /**
           * getGroup queries group details.
           */
          getGroup: withMetadata(async function getGroup(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryGroupRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).group(input, options);
          }, { path: [13, 2] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_deployment_v1beta4_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getClient(service).params(input, options);
          }, { path: [13, 3] }),
          /**
           * createDeployment defines a method to create new deployment given proper inputs.
           */
          createDeployment: withMetadata(async function createDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgCreateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).createDeployment(input, options);
          }, { path: [14, 0] }),
          /**
           * updateDeployment defines a method to update a deployment given proper inputs.
           */
          updateDeployment: withMetadata(async function updateDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgUpdateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).updateDeployment(input, options);
          }, { path: [14, 1] }),
          /**
           * closeDeployment defines a method to close a deployment given proper inputs.
           */
          closeDeployment: withMetadata(async function closeDeployment(input: DeepSimplify<virtengine_deployment_v1beta4_deploymentmsg.MsgCloseDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).closeDeployment(input, options);
          }, { path: [14, 2] }),
          /**
           * closeGroup defines a method to close a group of a deployment given proper inputs.
           */
          closeGroup: withMetadata(async function closeGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgCloseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).closeGroup(input, options);
          }, { path: [14, 3] }),
          /**
           * pauseGroup defines a method to pause a group of a deployment given proper inputs.
           */
          pauseGroup: withMetadata(async function pauseGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgPauseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).pauseGroup(input, options);
          }, { path: [14, 4] }),
          /**
           * startGroup defines a method to start a group of a deployment given proper inputs.
           */
          startGroup: withMetadata(async function startGroup(input: DeepSimplify<virtengine_deployment_v1beta4_groupmsg.MsgStartGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).startGroup(input, options);
          }, { path: [14, 5] }),
          /**
           * updateParams defines a governance operation for updating the x/deployment module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_deployment_v1beta4_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [14, 6] })
        },
        v1beta5: {
          /**
           * getDeployments queries deployments.
           */
          getDeployments: withMetadata(async function getDeployments(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryDeploymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).deployments(input, options);
          }, { path: [15, 0] }),
          /**
           * getDeployment queries deployment details.
           */
          getDeployment: withMetadata(async function getDeployment(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryDeploymentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).deployment(input, options);
          }, { path: [15, 1] }),
          /**
           * getGroup queries group details.
           */
          getGroup: withMetadata(async function getGroup(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryGroupRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).group(input, options);
          }, { path: [15, 2] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_deployment_v1beta5_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getClient(service).params(input, options);
          }, { path: [15, 3] }),
          /**
           * createDeployment defines a method to create new deployment given proper inputs.
           */
          createDeployment: withMetadata(async function createDeployment(input: DeepSimplify<virtengine_deployment_v1beta5_deploymentmsg.MsgCreateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).createDeployment(input, options);
          }, { path: [16, 0] }),
          /**
           * updateDeployment defines a method to update a deployment given proper inputs.
           */
          updateDeployment: withMetadata(async function updateDeployment(input: DeepSimplify<virtengine_deployment_v1beta5_deploymentmsg.MsgUpdateDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).updateDeployment(input, options);
          }, { path: [16, 1] }),
          /**
           * closeDeployment defines a method to close a deployment given proper inputs.
           */
          closeDeployment: withMetadata(async function closeDeployment(input: DeepSimplify<virtengine_deployment_v1beta5_deploymentmsg.MsgCloseDeployment>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).closeDeployment(input, options);
          }, { path: [16, 2] }),
          /**
           * closeGroup defines a method to close a group of a deployment given proper inputs.
           */
          closeGroup: withMetadata(async function closeGroup(input: DeepSimplify<virtengine_deployment_v1beta5_groupmsg.MsgCloseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).closeGroup(input, options);
          }, { path: [16, 3] }),
          /**
           * pauseGroup defines a method to pause a group of a deployment given proper inputs.
           */
          pauseGroup: withMetadata(async function pauseGroup(input: DeepSimplify<virtengine_deployment_v1beta5_groupmsg.MsgPauseGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).pauseGroup(input, options);
          }, { path: [16, 4] }),
          /**
           * startGroup defines a method to start a group of a deployment given proper inputs.
           */
          startGroup: withMetadata(async function startGroup(input: DeepSimplify<virtengine_deployment_v1beta5_groupmsg.MsgStartGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).startGroup(input, options);
          }, { path: [16, 5] }),
          /**
           * updateParams defines a governance operation for updating the x/deployment module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_deployment_v1beta5_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [16, 6] })
        }
      },
      downtimedetector: {
        v1beta1: {
          /**
           * getRecoveredSinceDowntimeOfLength queries if the chain has recovered for a specified duration
           * since experiencing downtime of a given length
           */
          getRecoveredSinceDowntimeOfLength: withMetadata(async function getRecoveredSinceDowntimeOfLength(input: DeepPartial<virtengine_downtimedetector_v1beta1_query.RecoveredSinceDowntimeOfLengthRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getClient(service).recoveredSinceDowntimeOfLength(input, options);
          }, { path: [17, 0] })
        }
      },
      enclave: {
        v1: {
          /**
           * getEnclaveIdentity queries an enclave identity for a validator
           */
          getEnclaveIdentity: withMetadata(async function getEnclaveIdentity(input: DeepPartial<virtengine_enclave_v1_query.QueryEnclaveIdentityRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).enclaveIdentity(input, options);
          }, { path: [18, 0] }),
          /**
           * getActiveValidatorEnclaveKeys queries all active validator enclave keys
           */
          getActiveValidatorEnclaveKeys: withMetadata(async function getActiveValidatorEnclaveKeys(input: DeepPartial<virtengine_enclave_v1_query.QueryActiveValidatorEnclaveKeysRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).activeValidatorEnclaveKeys(input, options);
          }, { path: [18, 1] }),
          /**
           * getCommitteeEnclaveKeys queries committee enclave keys for an epoch
           */
          getCommitteeEnclaveKeys: withMetadata(async function getCommitteeEnclaveKeys(input: DeepPartial<virtengine_enclave_v1_query.QueryCommitteeEnclaveKeysRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).committeeEnclaveKeys(input, options);
          }, { path: [18, 2] }),
          /**
           * getMeasurementAllowlist queries the measurement allowlist
           */
          getMeasurementAllowlist: withMetadata(async function getMeasurementAllowlist(input: DeepPartial<virtengine_enclave_v1_query.QueryMeasurementAllowlistRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).measurementAllowlist(input, options);
          }, { path: [18, 3] }),
          /**
           * getMeasurement queries a specific measurement
           */
          getMeasurement: withMetadata(async function getMeasurement(input: DeepPartial<virtengine_enclave_v1_query.QueryMeasurementRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).measurement(input, options);
          }, { path: [18, 4] }),
          /**
           * getKeyRotation queries key rotation status for a validator
           */
          getKeyRotation: withMetadata(async function getKeyRotation(input: DeepPartial<virtengine_enclave_v1_query.QueryKeyRotationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).keyRotation(input, options);
          }, { path: [18, 5] }),
          /**
           * getValidKeySet queries the current valid key set
           */
          getValidKeySet: withMetadata(async function getValidKeySet(input: DeepPartial<virtengine_enclave_v1_query.QueryValidKeySetRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).validKeySet(input, options);
          }, { path: [18, 6] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_enclave_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).params(input, options);
          }, { path: [18, 7] }),
          /**
           * getAttestedResult queries an attested scoring result
           */
          getAttestedResult: withMetadata(async function getAttestedResult(input: DeepPartial<virtengine_enclave_v1_query.QueryAttestedResultRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).attestedResult(input, options);
          }, { path: [18, 8] }),
          /**
           * registerEnclaveIdentity registers a new enclave identity for a validator
           */
          registerEnclaveIdentity: withMetadata(async function registerEnclaveIdentity(input: DeepSimplify<virtengine_enclave_v1_tx.MsgRegisterEnclaveIdentity>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).registerEnclaveIdentity(input, options);
          }, { path: [19, 0] }),
          /**
           * rotateEnclaveIdentity initiates a key rotation for a validator's enclave
           */
          rotateEnclaveIdentity: withMetadata(async function rotateEnclaveIdentity(input: DeepSimplify<virtengine_enclave_v1_tx.MsgRotateEnclaveIdentity>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).rotateEnclaveIdentity(input, options);
          }, { path: [19, 1] }),
          /**
           * proposeMeasurement proposes a new enclave measurement for the allowlist
           */
          proposeMeasurement: withMetadata(async function proposeMeasurement(input: DeepSimplify<virtengine_enclave_v1_tx.MsgProposeMeasurement>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).proposeMeasurement(input, options);
          }, { path: [19, 2] }),
          /**
           * revokeMeasurement revokes an enclave measurement from the allowlist
           */
          revokeMeasurement: withMetadata(async function revokeMeasurement(input: DeepSimplify<virtengine_enclave_v1_tx.MsgRevokeMeasurement>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).revokeMeasurement(input, options);
          }, { path: [19, 3] }),
          /**
           * updateParams updates the module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_enclave_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [19, 4] })
        }
      },
      encryption: {
        v1: {
          /**
           * getRecipientKey returns the recipient keys for an account
           */
          getRecipientKey: withMetadata(async function getRecipientKey(input: DeepPartial<virtengine_encryption_v1_query.QueryRecipientKeyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).recipientKey(input, options);
          }, { path: [20, 0] }),
          /**
           * getKeyByFingerprint returns a key by its fingerprint
           */
          getKeyByFingerprint: withMetadata(async function getKeyByFingerprint(input: DeepPartial<virtengine_encryption_v1_query.QueryKeyByFingerprintRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).keyByFingerprint(input, options);
          }, { path: [20, 1] }),
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_encryption_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).params(input, options);
          }, { path: [20, 2] }),
          /**
           * getAlgorithms returns the supported encryption algorithms
           */
          getAlgorithms: withMetadata(async function getAlgorithms(input: DeepPartial<virtengine_encryption_v1_query.QueryAlgorithmsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).algorithms(input, options);
          }, { path: [20, 3] }),
          /**
           * getValidateEnvelope validates an encrypted payload envelope
           */
          getValidateEnvelope: withMetadata(async function getValidateEnvelope(input: DeepPartial<virtengine_encryption_v1_query.QueryValidateEnvelopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getClient(service).validateEnvelope(input, options);
          }, { path: [20, 4] }),
          /**
           * registerRecipientKey registers a new recipient public key
           */
          registerRecipientKey: withMetadata(async function registerRecipientKey(input: DeepSimplify<virtengine_encryption_v1_tx.MsgRegisterRecipientKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).registerRecipientKey(input, options);
          }, { path: [21, 0] }),
          /**
           * revokeRecipientKey revokes an existing recipient public key
           */
          revokeRecipientKey: withMetadata(async function revokeRecipientKey(input: DeepSimplify<virtengine_encryption_v1_tx.MsgRevokeRecipientKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).revokeRecipientKey(input, options);
          }, { path: [21, 1] }),
          /**
           * updateKeyLabel updates the label of an existing recipient key
           */
          updateKeyLabel: withMetadata(async function updateKeyLabel(input: DeepSimplify<virtengine_encryption_v1_tx.MsgUpdateKeyLabel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).updateKeyLabel(input, options);
          }, { path: [21, 2] }),
          /**
           * rotateKey rotates a recipient key with a new public key
           */
          rotateKey: withMetadata(async function rotateKey(input: DeepSimplify<virtengine_encryption_v1_tx.MsgRotateKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getMsgClient(service).rotateKey(input, options);
          }, { path: [21, 3] })
        }
      },
      epochs: {
        v1beta1: {
          /**
           * getEpochInfos provide running epochInfos
           */
          getEpochInfos: withMetadata(async function getEpochInfos(input: DeepPartial<virtengine_epochs_v1beta1_query.QueryEpochInfosRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).epochInfos(input, options);
          }, { path: [22, 0] }),
          /**
           * getCurrentEpoch provide current epoch of specified identifier
           */
          getCurrentEpoch: withMetadata(async function getCurrentEpoch(input: DeepPartial<virtengine_epochs_v1beta1_query.QueryCurrentEpochRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getClient(service).currentEpoch(input, options);
          }, { path: [22, 1] })
        }
      },
      escrow: {
        v1: {
          /**
           * getAccounts queries all accounts.
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<virtengine_escrow_v1_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).accounts(input, options);
          }, { path: [23, 0] }),
          /**
           * getPayments queries all payments.
           */
          getPayments: withMetadata(async function getPayments(input: DeepPartial<virtengine_escrow_v1_query.QueryPaymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).payments(input, options);
          }, { path: [23, 1] }),
          /**
           * getInvoice queries an invoice ledger record by ID.
           */
          getInvoice: withMetadata(async function getInvoice(input: DeepPartial<virtengine_escrow_v1_query.QueryInvoiceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).invoice(input, options);
          }, { path: [23, 2] }),
          /**
           * getInvoicesByProvider queries invoices by provider.
           */
          getInvoicesByProvider: withMetadata(async function getInvoicesByProvider(input: DeepPartial<virtengine_escrow_v1_query.QueryInvoicesByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).invoicesByProvider(input, options);
          }, { path: [23, 3] }),
          /**
           * getInvoicesByCustomer queries invoices by customer.
           */
          getInvoicesByCustomer: withMetadata(async function getInvoicesByCustomer(input: DeepPartial<virtengine_escrow_v1_query.QueryInvoicesByCustomerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).invoicesByCustomer(input, options);
          }, { path: [23, 4] }),
          /**
           * getInvoiceLedger queries invoice ledger entries by invoice ID.
           */
          getInvoiceLedger: withMetadata(async function getInvoiceLedger(input: DeepPartial<virtengine_escrow_v1_query.QueryInvoiceLedgerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).invoiceLedger(input, options);
          }, { path: [23, 5] }),
          /**
           * accountDeposit deposits more funds into the escrow account.
           */
          accountDeposit: withMetadata(async function accountDeposit(input: DeepSimplify<virtengine_escrow_v1_msg.MsgAccountDeposit>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getMsgClient(service).accountDeposit(input, options);
          }, { path: [24, 0] })
        },
        v1beta3: {
          /**
           * getAccounts queries all accounts
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<virtengine_escrow_v1beta3_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getClient(service).accounts(input, options);
          }, { path: [25, 0] }),
          /**
           * getPayments queries all payments
           */
          getPayments: withMetadata(async function getPayments(input: DeepPartial<virtengine_escrow_v1beta3_query.QueryPaymentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getClient(service).payments(input, options);
          }, { path: [25, 1] })
        }
      },
      fraud: {
        v1: {
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_fraud_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).params(input, options);
          }, { path: [26, 0] }),
          /**
           * getFraudReport returns a fraud report by ID
           */
          getFraudReport: withMetadata(async function getFraudReport(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).fraudReport(input, options);
          }, { path: [26, 1] }),
          /**
           * getFraudReports returns all fraud reports with optional filters
           */
          getFraudReports: withMetadata(async function getFraudReports(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).fraudReports(input, options);
          }, { path: [26, 2] }),
          /**
           * getFraudReportsByReporter returns fraud reports submitted by a reporter
           */
          getFraudReportsByReporter: withMetadata(async function getFraudReportsByReporter(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportsByReporterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).fraudReportsByReporter(input, options);
          }, { path: [26, 3] }),
          /**
           * getFraudReportsByReportedParty returns fraud reports against a reported party
           */
          getFraudReportsByReportedParty: withMetadata(async function getFraudReportsByReportedParty(input: DeepPartial<virtengine_fraud_v1_query.QueryFraudReportsByReportedPartyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).fraudReportsByReportedParty(input, options);
          }, { path: [26, 4] }),
          /**
           * getAuditLog returns the audit log for a report
           */
          getAuditLog: withMetadata(async function getAuditLog(input: DeepPartial<virtengine_fraud_v1_query.QueryAuditLogRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).auditLog(input, options);
          }, { path: [26, 5] }),
          /**
           * getModeratorQueue returns the moderator queue entries
           */
          getModeratorQueue: withMetadata(async function getModeratorQueue(input: DeepPartial<virtengine_fraud_v1_query.QueryModeratorQueueRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).moderatorQueue(input, options);
          }, { path: [26, 6] }),
          /**
           * submitFraudReport submits a new fraud report
           */
          submitFraudReport: withMetadata(async function submitFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgSubmitFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).submitFraudReport(input, options);
          }, { path: [27, 0] }),
          /**
           * assignModerator assigns a moderator to a fraud report
           */
          assignModerator: withMetadata(async function assignModerator(input: DeepSimplify<virtengine_fraud_v1_tx.MsgAssignModerator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).assignModerator(input, options);
          }, { path: [27, 1] }),
          /**
           * updateReportStatus updates the status of a fraud report
           */
          updateReportStatus: withMetadata(async function updateReportStatus(input: DeepSimplify<virtengine_fraud_v1_tx.MsgUpdateReportStatus>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).updateReportStatus(input, options);
          }, { path: [27, 2] }),
          /**
           * resolveFraudReport resolves a fraud report with action
           */
          resolveFraudReport: withMetadata(async function resolveFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgResolveFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).resolveFraudReport(input, options);
          }, { path: [27, 3] }),
          /**
           * rejectFraudReport rejects a fraud report
           */
          rejectFraudReport: withMetadata(async function rejectFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgRejectFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).rejectFraudReport(input, options);
          }, { path: [27, 4] }),
          /**
           * escalateFraudReport escalates a fraud report to admin
           */
          escalateFraudReport: withMetadata(async function escalateFraudReport(input: DeepSimplify<virtengine_fraud_v1_tx.MsgEscalateFraudReport>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).escalateFraudReport(input, options);
          }, { path: [27, 5] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_fraud_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [27, 6] })
        }
      },
      hpc: {
        v1: {
          /**
           * getCluster returns a cluster by ID
           */
          getCluster: withMetadata(async function getCluster(input: DeepPartial<virtengine_hpc_v1_query.QueryClusterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).cluster(input, options);
          }, { path: [28, 0] }),
          /**
           * getClusters returns all clusters with optional filters
           */
          getClusters: withMetadata(async function getClusters(input: DeepPartial<virtengine_hpc_v1_query.QueryClustersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).clusters(input, options);
          }, { path: [28, 1] }),
          /**
           * getClustersByProvider returns clusters owned by a provider
           */
          getClustersByProvider: withMetadata(async function getClustersByProvider(input: DeepPartial<virtengine_hpc_v1_query.QueryClustersByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).clustersByProvider(input, options);
          }, { path: [28, 2] }),
          /**
           * getOffering returns an offering by ID
           */
          getOffering: withMetadata(async function getOffering(input: DeepPartial<virtengine_hpc_v1_query.QueryOfferingRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).offering(input, options);
          }, { path: [28, 3] }),
          /**
           * getOfferings returns all offerings with optional filters
           */
          getOfferings: withMetadata(async function getOfferings(input: DeepPartial<virtengine_hpc_v1_query.QueryOfferingsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).offerings(input, options);
          }, { path: [28, 4] }),
          /**
           * getOfferingsByCluster returns offerings for a cluster
           */
          getOfferingsByCluster: withMetadata(async function getOfferingsByCluster(input: DeepPartial<virtengine_hpc_v1_query.QueryOfferingsByClusterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).offeringsByCluster(input, options);
          }, { path: [28, 5] }),
          /**
           * getJob returns a job by ID
           */
          getJob: withMetadata(async function getJob(input: DeepPartial<virtengine_hpc_v1_query.QueryJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).job(input, options);
          }, { path: [28, 6] }),
          /**
           * getJobs returns all jobs with optional filters
           */
          getJobs: withMetadata(async function getJobs(input: DeepPartial<virtengine_hpc_v1_query.QueryJobsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).jobs(input, options);
          }, { path: [28, 7] }),
          /**
           * getJobsByCustomer returns jobs submitted by a customer
           */
          getJobsByCustomer: withMetadata(async function getJobsByCustomer(input: DeepPartial<virtengine_hpc_v1_query.QueryJobsByCustomerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).jobsByCustomer(input, options);
          }, { path: [28, 8] }),
          /**
           * getJobsByProvider returns jobs handled by a provider
           */
          getJobsByProvider: withMetadata(async function getJobsByProvider(input: DeepPartial<virtengine_hpc_v1_query.QueryJobsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).jobsByProvider(input, options);
          }, { path: [28, 9] }),
          /**
           * getJobAccounting returns accounting data for a job
           */
          getJobAccounting: withMetadata(async function getJobAccounting(input: DeepPartial<virtengine_hpc_v1_query.QueryJobAccountingRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).jobAccounting(input, options);
          }, { path: [28, 10] }),
          /**
           * getNodeMetadata returns metadata for a node
           */
          getNodeMetadata: withMetadata(async function getNodeMetadata(input: DeepPartial<virtengine_hpc_v1_query.QueryNodeMetadataRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).nodeMetadata(input, options);
          }, { path: [28, 11] }),
          /**
           * getNodesByCluster returns nodes in a cluster
           */
          getNodesByCluster: withMetadata(async function getNodesByCluster(input: DeepPartial<virtengine_hpc_v1_query.QueryNodesByClusterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).nodesByCluster(input, options);
          }, { path: [28, 12] }),
          /**
           * getSchedulingDecision returns a scheduling decision by ID
           */
          getSchedulingDecision: withMetadata(async function getSchedulingDecision(input: DeepPartial<virtengine_hpc_v1_query.QuerySchedulingDecisionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).schedulingDecision(input, options);
          }, { path: [28, 13] }),
          /**
           * getSchedulingDecisionByJob returns the scheduling decision for a job
           */
          getSchedulingDecisionByJob: withMetadata(async function getSchedulingDecisionByJob(input: DeepPartial<virtengine_hpc_v1_query.QuerySchedulingDecisionByJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).schedulingDecisionByJob(input, options);
          }, { path: [28, 14] }),
          /**
           * getSchedulingMetrics returns scheduling metrics for a cluster and queue
           */
          getSchedulingMetrics: withMetadata(async function getSchedulingMetrics(input: DeepPartial<virtengine_hpc_v1_query.QuerySchedulingMetricsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).schedulingMetrics(input, options);
          }, { path: [28, 15] }),
          /**
           * getReward returns a reward record by ID
           */
          getReward: withMetadata(async function getReward(input: DeepPartial<virtengine_hpc_v1_query.QueryRewardRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).reward(input, options);
          }, { path: [28, 16] }),
          /**
           * getRewardsByJob returns rewards for a job
           */
          getRewardsByJob: withMetadata(async function getRewardsByJob(input: DeepPartial<virtengine_hpc_v1_query.QueryRewardsByJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).rewardsByJob(input, options);
          }, { path: [28, 17] }),
          /**
           * getDispute returns a dispute by ID
           */
          getDispute: withMetadata(async function getDispute(input: DeepPartial<virtengine_hpc_v1_query.QueryDisputeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).dispute(input, options);
          }, { path: [28, 18] }),
          /**
           * getDisputes returns all disputes with optional filters
           */
          getDisputes: withMetadata(async function getDisputes(input: DeepPartial<virtengine_hpc_v1_query.QueryDisputesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).disputes(input, options);
          }, { path: [28, 19] }),
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_hpc_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).params(input, options);
          }, { path: [28, 20] }),
          /**
           * registerCluster registers a new HPC cluster
           */
          registerCluster: withMetadata(async function registerCluster(input: DeepSimplify<virtengine_hpc_v1_tx.MsgRegisterCluster>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).registerCluster(input, options);
          }, { path: [29, 0] }),
          /**
           * updateCluster updates an existing HPC cluster
           */
          updateCluster: withMetadata(async function updateCluster(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateCluster>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).updateCluster(input, options);
          }, { path: [29, 1] }),
          /**
           * deregisterCluster deregisters an HPC cluster
           */
          deregisterCluster: withMetadata(async function deregisterCluster(input: DeepSimplify<virtengine_hpc_v1_tx.MsgDeregisterCluster>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).deregisterCluster(input, options);
          }, { path: [29, 2] }),
          /**
           * createOffering creates a new HPC offering
           */
          createOffering: withMetadata(async function createOffering(input: DeepSimplify<virtengine_hpc_v1_tx.MsgCreateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).createOffering(input, options);
          }, { path: [29, 3] }),
          /**
           * updateOffering updates an existing HPC offering
           */
          updateOffering: withMetadata(async function updateOffering(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).updateOffering(input, options);
          }, { path: [29, 4] }),
          /**
           * submitJob submits a new HPC job
           */
          submitJob: withMetadata(async function submitJob(input: DeepSimplify<virtengine_hpc_v1_tx.MsgSubmitJob>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).submitJob(input, options);
          }, { path: [29, 5] }),
          /**
           * cancelJob cancels an HPC job
           */
          cancelJob: withMetadata(async function cancelJob(input: DeepSimplify<virtengine_hpc_v1_tx.MsgCancelJob>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).cancelJob(input, options);
          }, { path: [29, 6] }),
          /**
           * reportJobStatus reports job status from the provider daemon
           */
          reportJobStatus: withMetadata(async function reportJobStatus(input: DeepSimplify<virtengine_hpc_v1_tx.MsgReportJobStatus>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).reportJobStatus(input, options);
          }, { path: [29, 7] }),
          /**
           * updateNodeMetadata updates node metadata
           */
          updateNodeMetadata: withMetadata(async function updateNodeMetadata(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateNodeMetadata>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).updateNodeMetadata(input, options);
          }, { path: [29, 8] }),
          /**
           * flagDispute flags a dispute for moderation
           */
          flagDispute: withMetadata(async function flagDispute(input: DeepSimplify<virtengine_hpc_v1_tx.MsgFlagDispute>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).flagDispute(input, options);
          }, { path: [29, 9] }),
          /**
           * resolveDispute resolves a dispute (moderator only)
           */
          resolveDispute: withMetadata(async function resolveDispute(input: DeepSimplify<virtengine_hpc_v1_tx.MsgResolveDispute>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).resolveDispute(input, options);
          }, { path: [29, 10] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_hpc_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [29, 11] })
        }
      },
      market: {
        v1beta5: {
          /**
           * getOrders queries orders with filters.
           */
          getOrders: withMetadata(async function getOrders(input: DeepPartial<virtengine_market_v1beta5_query.QueryOrdersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).orders(input, options);
          }, { path: [30, 0] }),
          /**
           * getOrder queries order details.
           */
          getOrder: withMetadata(async function getOrder(input: DeepPartial<virtengine_market_v1beta5_query.QueryOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).order(input, options);
          }, { path: [30, 1] }),
          /**
           * getBids queries bids with filters.
           */
          getBids: withMetadata(async function getBids(input: DeepPartial<virtengine_market_v1beta5_query.QueryBidsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).bids(input, options);
          }, { path: [30, 2] }),
          /**
           * getBid queries bid details.
           */
          getBid: withMetadata(async function getBid(input: DeepPartial<virtengine_market_v1beta5_query.QueryBidRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).bid(input, options);
          }, { path: [30, 3] }),
          /**
           * getLeases queries leases with filters.
           */
          getLeases: withMetadata(async function getLeases(input: DeepPartial<virtengine_market_v1beta5_query.QueryLeasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).leases(input, options);
          }, { path: [30, 4] }),
          /**
           * getLease queries lease details.
           */
          getLease: withMetadata(async function getLease(input: DeepPartial<virtengine_market_v1beta5_query.QueryLeaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).lease(input, options);
          }, { path: [30, 5] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_market_v1beta5_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).params(input, options);
          }, { path: [30, 6] }),
          /**
           * createBid defines a method to create a bid given proper inputs.
           */
          createBid: withMetadata(async function createBid(input: DeepSimplify<virtengine_market_v1beta5_bidmsg.MsgCreateBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).createBid(input, options);
          }, { path: [31, 0] }),
          /**
           * closeBid defines a method to close a bid given proper inputs.
           */
          closeBid: withMetadata(async function closeBid(input: DeepSimplify<virtengine_market_v1beta5_bidmsg.MsgCloseBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).closeBid(input, options);
          }, { path: [31, 1] }),
          /**
           * withdrawLease withdraws accrued funds from the lease payment
           */
          withdrawLease: withMetadata(async function withdrawLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgWithdrawLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).withdrawLease(input, options);
          }, { path: [31, 2] }),
          /**
           * createLease creates a new lease
           */
          createLease: withMetadata(async function createLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgCreateLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).createLease(input, options);
          }, { path: [31, 3] }),
          /**
           * closeLease defines a method to close an order given proper inputs.
           */
          closeLease: withMetadata(async function closeLease(input: DeepSimplify<virtengine_market_v1beta5_leasemsg.MsgCloseLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).closeLease(input, options);
          }, { path: [31, 4] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: virtengine v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_market_v1beta5_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [31, 5] })
        },
        v2beta1: {
          /**
           * getOrders queries orders with filters.
           */
          getOrders: withMetadata(async function getOrders(input: DeepPartial<virtengine_market_v2beta1_query.QueryOrdersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).orders(input, options);
          }, { path: [32, 0] }),
          /**
           * getOrder queries order details.
           */
          getOrder: withMetadata(async function getOrder(input: DeepPartial<virtengine_market_v2beta1_query.QueryOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).order(input, options);
          }, { path: [32, 1] }),
          /**
           * getBids queries bids with filters.
           */
          getBids: withMetadata(async function getBids(input: DeepPartial<virtengine_market_v2beta1_query.QueryBidsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).bids(input, options);
          }, { path: [32, 2] }),
          /**
           * getBid queries bid details.
           */
          getBid: withMetadata(async function getBid(input: DeepPartial<virtengine_market_v2beta1_query.QueryBidRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).bid(input, options);
          }, { path: [32, 3] }),
          /**
           * getLeases queries leases with filters.
           */
          getLeases: withMetadata(async function getLeases(input: DeepPartial<virtengine_market_v2beta1_query.QueryLeasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).leases(input, options);
          }, { path: [32, 4] }),
          /**
           * getLease queries lease details.
           */
          getLease: withMetadata(async function getLease(input: DeepPartial<virtengine_market_v2beta1_query.QueryLeaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).lease(input, options);
          }, { path: [32, 5] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_market_v2beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).params(input, options);
          }, { path: [32, 6] }),
          /**
           * createBid defines a method to create a bid given proper inputs.
           */
          createBid: withMetadata(async function createBid(input: DeepSimplify<virtengine_market_v2beta1_bidmsg.MsgCreateBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).createBid(input, options);
          }, { path: [33, 0] }),
          /**
           * closeBid defines a method to close a bid given proper inputs.
           */
          closeBid: withMetadata(async function closeBid(input: DeepSimplify<virtengine_market_v2beta1_bidmsg.MsgCloseBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).closeBid(input, options);
          }, { path: [33, 1] }),
          /**
           * withdrawLease withdraws accrued funds from the lease payment
           */
          withdrawLease: withMetadata(async function withdrawLease(input: DeepSimplify<virtengine_market_v2beta1_leasemsg.MsgWithdrawLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).withdrawLease(input, options);
          }, { path: [33, 2] }),
          /**
           * createLease creates a new lease
           */
          createLease: withMetadata(async function createLease(input: DeepSimplify<virtengine_market_v2beta1_leasemsg.MsgCreateLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).createLease(input, options);
          }, { path: [33, 3] }),
          /**
           * closeLease defines a method to close an order given proper inputs.
           */
          closeLease: withMetadata(async function closeLease(input: DeepSimplify<virtengine_market_v2beta1_leasemsg.MsgCloseLease>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).closeLease(input, options);
          }, { path: [33, 4] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: virtengine v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_market_v2beta1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [33, 5] })
        }
      },
      marketplace: {
        v1: {
          /**
           * getOfferingPrice calculates pricing for a specific offering.
           */
          getOfferingPrice: withMetadata(async function getOfferingPrice(input: DeepPartial<virtengine_marketplace_v1_query.QueryOfferingPriceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).offeringPrice(input, options);
          }, { path: [34, 0] }),
          /**
           * getAllocationsByCustomer returns allocations for a customer.
           */
          getAllocationsByCustomer: withMetadata(async function getAllocationsByCustomer(input: DeepPartial<virtengine_marketplace_v1_query.QueryAllocationsByCustomerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).allocationsByCustomer(input, options);
          }, { path: [34, 1] }),
          /**
           * getAllocationsByProvider returns allocations for a provider.
           */
          getAllocationsByProvider: withMetadata(async function getAllocationsByProvider(input: DeepPartial<virtengine_marketplace_v1_query.QueryAllocationsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).allocationsByProvider(input, options);
          }, { path: [34, 2] }),
          /**
           * createOffering creates a new marketplace offering
           */
          createOffering: withMetadata(async function createOffering(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgCreateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).createOffering(input, options);
          }, { path: [35, 0] }),
          /**
           * updateOffering updates an existing marketplace offering
           */
          updateOffering: withMetadata(async function updateOffering(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgUpdateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).updateOffering(input, options);
          }, { path: [35, 1] }),
          /**
           * deactivateOffering deactivates an existing marketplace offering
           */
          deactivateOffering: withMetadata(async function deactivateOffering(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgDeactivateOffering>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).deactivateOffering(input, options);
          }, { path: [35, 2] }),
          /**
           * acceptBid accepts a provider bid on an order
           */
          acceptBid: withMetadata(async function acceptBid(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgAcceptBid>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).acceptBid(input, options);
          }, { path: [35, 3] }),
          /**
           * terminateAllocation terminates an allocation
           */
          terminateAllocation: withMetadata(async function terminateAllocation(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgTerminateAllocation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).terminateAllocation(input, options);
          }, { path: [35, 4] }),
          /**
           * resizeAllocation resizes an existing allocation
           */
          resizeAllocation: withMetadata(async function resizeAllocation(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgResizeAllocation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).resizeAllocation(input, options);
          }, { path: [35, 5] }),
          /**
           * pauseAllocation pauses an active allocation
           */
          pauseAllocation: withMetadata(async function pauseAllocation(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgPauseAllocation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).pauseAllocation(input, options);
          }, { path: [35, 6] }),
          /**
           * waldurCallback handles callbacks from Waldur integration
           */
          waldurCallback: withMetadata(async function waldurCallback(input: DeepSimplify<virtengine_marketplace_v1_tx.MsgWaldurCallback>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).waldurCallback(input, options);
          }, { path: [35, 7] })
        }
      },
      mfa: {
        v1: {
          /**
           * getMFAPolicy returns the MFA policy for an account
           */
          getMFAPolicy: withMetadata(async function getMFAPolicy(input: DeepPartial<virtengine_mfa_v1_query.QueryMFAPolicyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).mFAPolicy(input, options);
          }, { path: [36, 0] }),
          /**
           * getFactorEnrollments returns all factor enrollments for an account
           */
          getFactorEnrollments: withMetadata(async function getFactorEnrollments(input: DeepPartial<virtengine_mfa_v1_query.QueryFactorEnrollmentsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).factorEnrollments(input, options);
          }, { path: [36, 1] }),
          /**
           * getFactorEnrollment returns a specific factor enrollment
           */
          getFactorEnrollment: withMetadata(async function getFactorEnrollment(input: DeepPartial<virtengine_mfa_v1_query.QueryFactorEnrollmentRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).factorEnrollment(input, options);
          }, { path: [36, 2] }),
          /**
           * getChallenge returns a challenge by ID
           */
          getChallenge: withMetadata(async function getChallenge(input: DeepPartial<virtengine_mfa_v1_query.QueryChallengeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).challenge(input, options);
          }, { path: [36, 3] }),
          /**
           * getPendingChallenges returns pending challenges for an account
           */
          getPendingChallenges: withMetadata(async function getPendingChallenges(input: DeepPartial<virtengine_mfa_v1_query.QueryPendingChallengesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).pendingChallenges(input, options);
          }, { path: [36, 4] }),
          /**
           * getAuthorizationSession returns an authorization session by ID
           */
          getAuthorizationSession: withMetadata(async function getAuthorizationSession(input: DeepPartial<virtengine_mfa_v1_query.QueryAuthorizationSessionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).authorizationSession(input, options);
          }, { path: [36, 5] }),
          /**
           * getAuthorizationSessions returns authorization sessions for an account
           */
          getAuthorizationSessions: withMetadata(async function getAuthorizationSessions(input: DeepPartial<virtengine_mfa_v1_query.QueryAuthorizationSessionsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).authorizationSessions(input, options);
          }, { path: [36, 6] }),
          /**
           * getTrustedDevices returns trusted devices for an account
           */
          getTrustedDevices: withMetadata(async function getTrustedDevices(input: DeepPartial<virtengine_mfa_v1_query.QueryTrustedDevicesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).trustedDevices(input, options);
          }, { path: [36, 7] }),
          /**
           * getTrustedDevice returns a trusted device by fingerprint
           */
          getTrustedDevice: withMetadata(async function getTrustedDevice(input: DeepPartial<virtengine_mfa_v1_query.QueryTrustedDeviceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).trustedDevice(input, options);
          }, { path: [36, 8] }),
          /**
           * getSensitiveTxConfig returns the configuration for a sensitive tx type
           */
          getSensitiveTxConfig: withMetadata(async function getSensitiveTxConfig(input: DeepPartial<virtengine_mfa_v1_query.QuerySensitiveTxConfigRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).sensitiveTxConfig(input, options);
          }, { path: [36, 9] }),
          /**
           * getAllSensitiveTxConfigs returns all sensitive tx configurations
           */
          getAllSensitiveTxConfigs: withMetadata(async function getAllSensitiveTxConfigs(input: DeepPartial<virtengine_mfa_v1_query.QueryAllSensitiveTxConfigsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).allSensitiveTxConfigs(input, options);
          }, { path: [36, 10] }),
          /**
           * getMFARequired checks if MFA is required for a transaction
           */
          getMFARequired: withMetadata(async function getMFARequired(input: DeepPartial<virtengine_mfa_v1_query.QueryMFARequiredRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).mFARequired(input, options);
          }, { path: [36, 11] }),
          /**
           * getParams returns the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_mfa_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).params(input, options);
          }, { path: [36, 12] }),
          /**
           * enrollFactor enrolls a new MFA factor
           */
          enrollFactor: withMetadata(async function enrollFactor(input: DeepSimplify<virtengine_mfa_v1_tx.MsgEnrollFactor>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).enrollFactor(input, options);
          }, { path: [37, 0] }),
          /**
           * revokeFactor revokes an enrolled factor
           */
          revokeFactor: withMetadata(async function revokeFactor(input: DeepSimplify<virtengine_mfa_v1_tx.MsgRevokeFactor>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).revokeFactor(input, options);
          }, { path: [37, 1] }),
          /**
           * setMFAPolicy sets the MFA policy for an account
           */
          setMFAPolicy: withMetadata(async function setMFAPolicy(input: DeepSimplify<virtengine_mfa_v1_tx.MsgSetMFAPolicy>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).setMFAPolicy(input, options);
          }, { path: [37, 2] }),
          /**
           * createChallenge creates an MFA challenge
           */
          createChallenge: withMetadata(async function createChallenge(input: DeepSimplify<virtengine_mfa_v1_tx.MsgCreateChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).createChallenge(input, options);
          }, { path: [37, 3] }),
          /**
           * verifyChallenge verifies an MFA challenge response
           */
          verifyChallenge: withMetadata(async function verifyChallenge(input: DeepSimplify<virtengine_mfa_v1_tx.MsgVerifyChallenge>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).verifyChallenge(input, options);
          }, { path: [37, 4] }),
          /**
           * addTrustedDevice adds a trusted device
           */
          addTrustedDevice: withMetadata(async function addTrustedDevice(input: DeepSimplify<virtengine_mfa_v1_tx.MsgAddTrustedDevice>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).addTrustedDevice(input, options);
          }, { path: [37, 5] }),
          /**
           * removeTrustedDevice removes a trusted device
           */
          removeTrustedDevice: withMetadata(async function removeTrustedDevice(input: DeepSimplify<virtengine_mfa_v1_tx.MsgRemoveTrustedDevice>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).removeTrustedDevice(input, options);
          }, { path: [37, 6] }),
          /**
           * updateSensitiveTxConfig updates sensitive transaction configuration (governance only)
           */
          updateSensitiveTxConfig: withMetadata(async function updateSensitiveTxConfig(input: DeepSimplify<virtengine_mfa_v1_tx.MsgUpdateSensitiveTxConfig>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).updateSensitiveTxConfig(input, options);
          }, { path: [37, 7] }),
          /**
           * issueSession issues an authorization session
           */
          issueSession: withMetadata(async function issueSession(input: DeepSimplify<virtengine_mfa_v1_tx.MsgIssueSession>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).issueSession(input, options);
          }, { path: [37, 8] }),
          /**
           * refreshSession refreshes an authorization session
           */
          refreshSession: withMetadata(async function refreshSession(input: DeepSimplify<virtengine_mfa_v1_tx.MsgRefreshSession>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).refreshSession(input, options);
          }, { path: [37, 9] }),
          /**
           * revokeSession revokes an authorization session
           */
          revokeSession: withMetadata(async function revokeSession(input: DeepSimplify<virtengine_mfa_v1_tx.MsgRevokeSession>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).revokeSession(input, options);
          }, { path: [37, 10] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_mfa_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [37, 11] })
        }
      },
      oracle: {
        v1: {
          /**
           * getPrices query prices for specific denom
           */
          getPrices: withMetadata(async function getPrices(input: DeepPartial<virtengine_oracle_v1_prices.QueryPricesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).prices(input, options);
          }, { path: [38, 0] }),
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_oracle_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).params(input, options);
          }, { path: [38, 1] }),
          /**
           * getPriceFeedConfig queries the price feed configuration for a given denom.
           */
          getPriceFeedConfig: withMetadata(async function getPriceFeedConfig(input: DeepPartial<virtengine_oracle_v1_query.QueryPriceFeedConfigRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).priceFeedConfig(input, options);
          }, { path: [38, 2] }),
          /**
           * getAggregatedPrice queries the aggregated price for a given denom.
           */
          getAggregatedPrice: withMetadata(async function getAggregatedPrice(input: DeepPartial<virtengine_oracle_v1_query.QueryAggregatedPriceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).aggregatedPrice(input, options);
          }, { path: [38, 3] }),
          /**
           * addPriceEntry adds a new price entry for a denomination from an authorized source
           */
          addPriceEntry: withMetadata(async function addPriceEntry(input: DeepSimplify<virtengine_oracle_v1_msgs.MsgAddPriceEntry>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getMsgClient(service).addPriceEntry(input, options);
          }, { path: [39, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/wasm module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v2.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_oracle_v1_msgs.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [39, 1] })
        }
      },
      provider: {
        v1beta4: {
          /**
           * getProviders queries providers
           */
          getProviders: withMetadata(async function getProviders(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProvidersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).providers(input, options);
          }, { path: [40, 0] }),
          /**
           * getProvider queries provider details
           */
          getProvider: withMetadata(async function getProvider(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).provider(input, options);
          }, { path: [40, 1] }),
          getProviderSigningKey: withMetadata(async function getProviderSigningKey(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProviderSigningKeyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).providerSigningKey(input, options);
          }, { path: [40, 2] }),
          getProviderSigningKeyEpochs: withMetadata(async function getProviderSigningKeyEpochs(input: DeepPartial<virtengine_provider_v1beta4_query.QueryProviderSigningKeyEpochsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getClient(service).providerSigningKeyEpochs(input, options);
          }, { path: [40, 3] }),
          /**
           * createProvider defines a method that creates a provider given the proper inputs.
           */
          createProvider: withMetadata(async function createProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgCreateProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).createProvider(input, options);
          }, { path: [41, 0] }),
          /**
           * updateProvider defines a method that updates a provider given the proper inputs.
           */
          updateProvider: withMetadata(async function updateProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgUpdateProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).updateProvider(input, options);
          }, { path: [41, 1] }),
          /**
           * deleteProvider defines a method that deletes a provider given the proper inputs.
           */
          deleteProvider: withMetadata(async function deleteProvider(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgDeleteProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).deleteProvider(input, options);
          }, { path: [41, 2] }),
          /**
           * generateDomainVerificationToken generates a verification token for a provider's domain.
           */
          generateDomainVerificationToken: withMetadata(async function generateDomainVerificationToken(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgGenerateDomainVerificationToken>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).generateDomainVerificationToken(input, options);
          }, { path: [41, 3] }),
          /**
           * verifyProviderDomain verifies a provider's domain via DNS TXT record.
           */
          verifyProviderDomain: withMetadata(async function verifyProviderDomain(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgVerifyProviderDomain>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).verifyProviderDomain(input, options);
          }, { path: [41, 4] }),
          /**
           * requestDomainVerification requests domain verification with specified method.
           */
          requestDomainVerification: withMetadata(async function requestDomainVerification(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgRequestDomainVerification>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).requestDomainVerification(input, options);
          }, { path: [41, 5] }),
          /**
           * confirmDomainVerification confirms domain verification with off-chain proof.
           */
          confirmDomainVerification: withMetadata(async function confirmDomainVerification(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgConfirmDomainVerification>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).confirmDomainVerification(input, options);
          }, { path: [41, 6] }),
          /**
           * revokeDomainVerification revokes a provider's domain verification.
           */
          revokeDomainVerification: withMetadata(async function revokeDomainVerification(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgRevokeDomainVerification>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).revokeDomainVerification(input, options);
          }, { path: [41, 7] }),
          /**
           * setProviderSigningKey registers the first detached-signature key epoch.
           */
          setProviderSigningKey: withMetadata(async function setProviderSigningKey(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgSetProviderSigningKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).setProviderSigningKey(input, options);
          }, { path: [41, 8] }),
          /**
           * rotateProviderSigningKey rotates through a proof signed by the old key.
           */
          rotateProviderSigningKey: withMetadata(async function rotateProviderSigningKey(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgRotateProviderSigningKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).rotateProviderSigningKey(input, options);
          }, { path: [41, 9] }),
          /**
           * revokeProviderSigningKey permanently revokes the active key epoch.
           */
          revokeProviderSigningKey: withMetadata(async function revokeProviderSigningKey(input: DeepSimplify<virtengine_provider_v1beta4_msg.MsgRevokeProviderSigningKey>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getMsgClient(service).revokeProviderSigningKey(input, options);
          }, { path: [41, 10] })
        }
      },
      resources: {
        v1: {
          /**
           * getAvailableResources returns eligible inventories for a request.
           */
          getAvailableResources: withMetadata(async function getAvailableResources(input: DeepPartial<virtengine_resources_v1_query.QueryAvailableResourcesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).availableResources(input, options);
          }, { path: [42, 0] }),
          /**
           * getAllocation returns an allocation by ID.
           */
          getAllocation: withMetadata(async function getAllocation(input: DeepPartial<virtengine_resources_v1_query.QueryAllocationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).allocation(input, options);
          }, { path: [42, 1] }),
          /**
           * getAllocationHistory returns lifecycle events for an allocation.
           */
          getAllocationHistory: withMetadata(async function getAllocationHistory(input: DeepPartial<virtengine_resources_v1_query.QueryAllocationHistoryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).allocationHistory(input, options);
          }, { path: [42, 2] }),
          /**
           * getAllocationsByProvider returns allocations for a provider.
           */
          getAllocationsByProvider: withMetadata(async function getAllocationsByProvider(input: DeepPartial<virtengine_resources_v1_query.QueryAllocationsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).allocationsByProvider(input, options);
          }, { path: [42, 3] }),
          getReservation: withMetadata(async function getReservation(input: DeepPartial<virtengine_resources_v1_query.QueryReservationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservation(input, options);
          }, { path: [42, 4] }),
          getReservationByOrder: withMetadata(async function getReservationByOrder(input: DeepPartial<virtengine_resources_v1_query.QueryReservationByOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationByOrder(input, options);
          }, { path: [42, 5] }),
          getReservationByBid: withMetadata(async function getReservationByBid(input: DeepPartial<virtengine_resources_v1_query.QueryReservationByBidRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationByBid(input, options);
          }, { path: [42, 6] }),
          getReservationByLease: withMetadata(async function getReservationByLease(input: DeepPartial<virtengine_resources_v1_query.QueryReservationByLeaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationByLease(input, options);
          }, { path: [42, 7] }),
          getReservationByJob: withMetadata(async function getReservationByJob(input: DeepPartial<virtengine_resources_v1_query.QueryReservationByJobRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationByJob(input, options);
          }, { path: [42, 8] }),
          getReservationByConsumer: withMetadata(async function getReservationByConsumer(input: DeepPartial<virtengine_resources_v1_query.QueryReservationByConsumerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationByConsumer(input, options);
          }, { path: [42, 9] }),
          getReservationsByProvider: withMetadata(async function getReservationsByProvider(input: DeepPartial<virtengine_resources_v1_query.QueryReservationsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationsByProvider(input, options);
          }, { path: [42, 10] }),
          getReservationLineage: withMetadata(async function getReservationLineage(input: DeepPartial<virtengine_resources_v1_query.QueryReservationLineageRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).reservationLineage(input, options);
          }, { path: [42, 11] }),
          /**
           * getParams returns module parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_resources_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).params(input, options);
          }, { path: [42, 12] }),
          /**
           * providerHeartbeat updates resource inventory from a provider heartbeat.
           */
          providerHeartbeat: withMetadata(async function providerHeartbeat(input: DeepSimplify<virtengine_resources_v1_tx.MsgProviderHeartbeat>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).providerHeartbeat(input, options);
          }, { path: [43, 0] }),
          /**
           * allocateResources selects a provider and creates a pending allocation.
           */
          allocateResources: withMetadata(async function allocateResources(input: DeepSimplify<virtengine_resources_v1_tx.MsgAllocateResources>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).allocateResources(input, options);
          }, { path: [43, 1] }),
          /**
           * activateAllocation acknowledges and activates an allocation.
           */
          activateAllocation: withMetadata(async function activateAllocation(input: DeepSimplify<virtengine_resources_v1_tx.MsgActivateAllocation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).activateAllocation(input, options);
          }, { path: [43, 2] }),
          /**
           * releaseAllocation releases an allocation and returns capacity.
           */
          releaseAllocation: withMetadata(async function releaseAllocation(input: DeepSimplify<virtengine_resources_v1_tx.MsgReleaseAllocation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).releaseAllocation(input, options);
          }, { path: [43, 3] }),
          /**
           * updateParams updates module parameters (governance only).
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_resources_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [43, 4] })
        }
      },
      review: {
        v1: {
          /**
           * submitReview submits a new review
           */
          submitReview: withMetadata(async function submitReview(input: DeepSimplify<virtengine_review_v1_tx.MsgSubmitReview>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getMsgClient(service).submitReview(input, options);
          }, { path: [44, 0] }),
          /**
           * deleteReview deletes a review
           */
          deleteReview: withMetadata(async function deleteReview(input: DeepSimplify<virtengine_review_v1_tx.MsgDeleteReview>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getMsgClient(service).deleteReview(input, options);
          }, { path: [44, 1] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_review_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [44, 2] }),
          /**
           * getReview queries a single review by ID.
           */
          getReview: withMetadata(async function getReview(input: DeepPartial<virtengine_review_v1_query.QueryReviewRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getClient(service).review(input, options);
          }, { path: [45, 0] }),
          /**
           * getReviewsByUser queries reviews by reviewer address.
           */
          getReviewsByUser: withMetadata(async function getReviewsByUser(input: DeepPartial<virtengine_review_v1_query.QueryReviewsByUserRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getClient(service).reviewsByUser(input, options);
          }, { path: [45, 1] }),
          /**
           * getParams returns the module parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_review_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getClient(service).params(input, options);
          }, { path: [45, 2] })
        }
      },
      roles: {
        v1: {
          /**
           * getAccountRoles queries all roles for an account
           */
          getAccountRoles: withMetadata(async function getAccountRoles(input: DeepPartial<virtengine_roles_v1_query.QueryAccountRolesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(46);
            return getClient(service).accountRoles(input, options);
          }, { path: [46, 0] }),
          /**
           * getRoleMembers queries all members of a specific role
           */
          getRoleMembers: withMetadata(async function getRoleMembers(input: DeepPartial<virtengine_roles_v1_query.QueryRoleMembersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(46);
            return getClient(service).roleMembers(input, options);
          }, { path: [46, 1] }),
          /**
           * getAccountState queries the state of an account
           */
          getAccountState: withMetadata(async function getAccountState(input: DeepPartial<virtengine_roles_v1_query.QueryAccountStateRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(46);
            return getClient(service).accountState(input, options);
          }, { path: [46, 2] }),
          /**
           * getGenesisAccounts queries all genesis accounts
           */
          getGenesisAccounts: withMetadata(async function getGenesisAccounts(input: DeepPartial<virtengine_roles_v1_query.QueryGenesisAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(46);
            return getClient(service).genesisAccounts(input, options);
          }, { path: [46, 3] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_roles_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(46);
            return getClient(service).params(input, options);
          }, { path: [46, 4] }),
          /**
           * getHasRole checks if an account has a specific role
           */
          getHasRole: withMetadata(async function getHasRole(input: DeepPartial<virtengine_roles_v1_query.QueryHasRoleRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(46);
            return getClient(service).hasRole(input, options);
          }, { path: [46, 5] }),
          /**
           * assignRole assigns a role to an account
           */
          assignRole: withMetadata(async function assignRole(input: DeepSimplify<virtengine_roles_v1_tx.MsgAssignRole>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getMsgClient(service).assignRole(input, options);
          }, { path: [47, 0] }),
          /**
           * revokeRole revokes a role from an account
           */
          revokeRole: withMetadata(async function revokeRole(input: DeepSimplify<virtengine_roles_v1_tx.MsgRevokeRole>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getMsgClient(service).revokeRole(input, options);
          }, { path: [47, 1] }),
          /**
           * setAccountState sets the state of an account
           */
          setAccountState: withMetadata(async function setAccountState(input: DeepSimplify<virtengine_roles_v1_tx.MsgSetAccountState>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getMsgClient(service).setAccountState(input, options);
          }, { path: [47, 2] }),
          /**
           * nominateAdmin nominates an administrator (GenesisAccount only)
           */
          nominateAdmin: withMetadata(async function nominateAdmin(input: DeepSimplify<virtengine_roles_v1_tx.MsgNominateAdmin>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getMsgClient(service).nominateAdmin(input, options);
          }, { path: [47, 3] }),
          /**
           * updateParams updates the module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_roles_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [47, 4] })
        }
      },
      settlement: {
        v1: {
          getEscrow: withMetadata(async function getEscrow(input: DeepPartial<virtengine_settlement_v1_query.QueryEscrowRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).escrow(input, options);
          }, { path: [48, 0] }),
          getEscrowsByOrder: withMetadata(async function getEscrowsByOrder(input: DeepPartial<virtengine_settlement_v1_query.QueryEscrowsByOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).escrowsByOrder(input, options);
          }, { path: [48, 1] }),
          getEscrowsByState: withMetadata(async function getEscrowsByState(input: DeepPartial<virtengine_settlement_v1_query.QueryEscrowsByStateRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).escrowsByState(input, options);
          }, { path: [48, 2] }),
          getSettlement: withMetadata(async function getSettlement(input: DeepPartial<virtengine_settlement_v1_query.QuerySettlementRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).settlement(input, options);
          }, { path: [48, 3] }),
          getSettlementsByOrder: withMetadata(async function getSettlementsByOrder(input: DeepPartial<virtengine_settlement_v1_query.QuerySettlementsByOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).settlementsByOrder(input, options);
          }, { path: [48, 4] }),
          getUsageRecord: withMetadata(async function getUsageRecord(input: DeepPartial<virtengine_settlement_v1_query.QueryUsageRecordRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).usageRecord(input, options);
          }, { path: [48, 5] }),
          getUsageRecordsByOrder: withMetadata(async function getUsageRecordsByOrder(input: DeepPartial<virtengine_settlement_v1_query.QueryUsageRecordsByOrderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).usageRecordsByOrder(input, options);
          }, { path: [48, 6] }),
          getUsageStreamState: withMetadata(async function getUsageStreamState(input: DeepPartial<virtengine_settlement_v1_query.QueryUsageStreamStateRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).usageStreamState(input, options);
          }, { path: [48, 7] }),
          getUsageSummary: withMetadata(async function getUsageSummary(input: DeepPartial<virtengine_settlement_v1_query.QueryUsageSummaryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).usageSummary(input, options);
          }, { path: [48, 8] }),
          getRewardDistribution: withMetadata(async function getRewardDistribution(input: DeepPartial<virtengine_settlement_v1_query.QueryRewardDistributionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).rewardDistribution(input, options);
          }, { path: [48, 9] }),
          getRewardsByEpoch: withMetadata(async function getRewardsByEpoch(input: DeepPartial<virtengine_settlement_v1_query.QueryRewardsByEpochRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).rewardsByEpoch(input, options);
          }, { path: [48, 10] }),
          getRewardHistory: withMetadata(async function getRewardHistory(input: DeepPartial<virtengine_settlement_v1_query.QueryRewardHistoryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).rewardHistory(input, options);
          }, { path: [48, 11] }),
          getClaimableRewards: withMetadata(async function getClaimableRewards(input: DeepPartial<virtengine_settlement_v1_query.QueryClaimableRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).claimableRewards(input, options);
          }, { path: [48, 12] }),
          getPayout: withMetadata(async function getPayout(input: DeepPartial<virtengine_settlement_v1_query.QueryPayoutRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).payout(input, options);
          }, { path: [48, 13] }),
          getPayoutsByProvider: withMetadata(async function getPayoutsByProvider(input: DeepPartial<virtengine_settlement_v1_query.QueryPayoutsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).payoutsByProvider(input, options);
          }, { path: [48, 14] }),
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_settlement_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).params(input, options);
          }, { path: [48, 15] }),
          getFiatConversion: withMetadata(async function getFiatConversion(input: DeepPartial<virtengine_settlement_v1_query.QueryFiatConversionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).fiatConversion(input, options);
          }, { path: [48, 16] }),
          getFiatConversionsByProvider: withMetadata(async function getFiatConversionsByProvider(input: DeepPartial<virtengine_settlement_v1_query.QueryFiatConversionsByProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).fiatConversionsByProvider(input, options);
          }, { path: [48, 17] }),
          getFiatPayoutPreference: withMetadata(async function getFiatPayoutPreference(input: DeepPartial<virtengine_settlement_v1_query.QueryFiatPayoutPreferenceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).fiatPayoutPreference(input, options);
          }, { path: [48, 18] }),
          getFinancialCase: withMetadata(async function getFinancialCase(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCaseRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCase(input, options);
          }, { path: [48, 19] }),
          getFinancialCaseBySubject: withMetadata(async function getFinancialCaseBySubject(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCaseBySubjectRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCaseBySubject(input, options);
          }, { path: [48, 20] }),
          getFinancialCasesByOrder: withMetadata(async function getFinancialCasesByOrder(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByOrder(input, options);
          }, { path: [48, 21] }),
          getFinancialCasesByInvoice: withMetadata(async function getFinancialCasesByInvoice(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByInvoice(input, options);
          }, { path: [48, 22] }),
          getFinancialCasesByUsage: withMetadata(async function getFinancialCasesByUsage(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByUsage(input, options);
          }, { path: [48, 23] }),
          getFinancialCasesByJob: withMetadata(async function getFinancialCasesByJob(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByJob(input, options);
          }, { path: [48, 24] }),
          getFinancialCasesByEscrow: withMetadata(async function getFinancialCasesByEscrow(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByEscrow(input, options);
          }, { path: [48, 25] }),
          getFinancialCasesByStatus: withMetadata(async function getFinancialCasesByStatus(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByStatus(input, options);
          }, { path: [48, 26] }),
          getFinancialCasesByParty: withMetadata(async function getFinancialCasesByParty(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCasesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCasesByParty(input, options);
          }, { path: [48, 27] }),
          getFinancialCaseLineage: withMetadata(async function getFinancialCaseLineage(input: DeepPartial<virtengine_settlement_v1_query.QueryFinancialCaseLineageRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).financialCaseLineage(input, options);
          }, { path: [48, 28] }),
          /**
           * createEscrow creates a new escrow account
           */
          createEscrow: withMetadata(async function createEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgCreateEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).createEscrow(input, options);
          }, { path: [49, 0] }),
          /**
           * activateEscrow activates an escrow when a lease is created
           */
          activateEscrow: withMetadata(async function activateEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgActivateEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).activateEscrow(input, options);
          }, { path: [49, 1] }),
          /**
           * releaseEscrow releases escrow funds to the recipient
           */
          releaseEscrow: withMetadata(async function releaseEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgReleaseEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).releaseEscrow(input, options);
          }, { path: [49, 2] }),
          /**
           * refundEscrow refunds escrow funds to the depositor
           */
          refundEscrow: withMetadata(async function refundEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgRefundEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).refundEscrow(input, options);
          }, { path: [49, 3] }),
          /**
           * disputeEscrow marks an escrow as disputed
           */
          disputeEscrow: withMetadata(async function disputeEscrow(input: DeepSimplify<virtengine_settlement_v1_tx.MsgDisputeEscrow>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).disputeEscrow(input, options);
          }, { path: [49, 4] }),
          /**
           * settleOrder settles an order based on usage records
           */
          settleOrder: withMetadata(async function settleOrder(input: DeepSimplify<virtengine_settlement_v1_tx.MsgSettleOrder>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).settleOrder(input, options);
          }, { path: [49, 5] }),
          /**
           * recordUsage records usage from a provider
           */
          recordUsage: withMetadata(async function recordUsage(input: DeepSimplify<virtengine_settlement_v1_tx.MsgRecordUsage>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).recordUsage(input, options);
          }, { path: [49, 6] }),
          /**
           * acknowledgeUsage acknowledges a usage record
           */
          acknowledgeUsage: withMetadata(async function acknowledgeUsage(input: DeepSimplify<virtengine_settlement_v1_tx.MsgAcknowledgeUsage>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).acknowledgeUsage(input, options);
          }, { path: [49, 7] }),
          /**
           * claimRewards claims accumulated rewards
           */
          claimRewards: withMetadata(async function claimRewards(input: DeepSimplify<virtengine_settlement_v1_tx.MsgClaimRewards>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).claimRewards(input, options);
          }, { path: [49, 8] }),
          /**
           * openFinancialCase atomically opens or merges a canonical financial case.
           */
          openFinancialCase: withMetadata(async function openFinancialCase(input: DeepSimplify<virtengine_settlement_v1_tx.MsgOpenFinancialCase>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).openFinancialCase(input, options);
          }, { path: [49, 9] }),
          /**
           * addFinancialClaim adds a privacy-safe typed claim to an active case.
           */
          addFinancialClaim: withMetadata(async function addFinancialClaim(input: DeepSimplify<virtengine_settlement_v1_tx.MsgAddFinancialClaim>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).addFinancialClaim(input, options);
          }, { path: [49, 10] }),
          /**
           * submitFinancialCaseForReview closes evidence collection.
           */
          submitFinancialCaseForReview: withMetadata(async function submitFinancialCaseForReview(input: DeepSimplify<virtengine_settlement_v1_tx.MsgSubmitFinancialCaseForReview>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).submitFinancialCaseForReview(input, options);
          }, { path: [49, 11] }),
          /**
           * escalateFinancialCase escalates a case to governance/arbitration.
           */
          escalateFinancialCase: withMetadata(async function escalateFinancialCase(input: DeepSimplify<virtengine_settlement_v1_tx.MsgEscalateFinancialCase>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).escalateFinancialCase(input, options);
          }, { path: [49, 12] }),
          /**
           * resolveFinancialCase records a conserved allocation pending appeal.
           */
          resolveFinancialCase: withMetadata(async function resolveFinancialCase(input: DeepSimplify<virtengine_settlement_v1_tx.MsgResolveFinancialCase>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).resolveFinancialCase(input, options);
          }, { path: [49, 13] }),
          /**
           * appealFinancialCase reopens review while all holds remain active.
           */
          appealFinancialCase: withMetadata(async function appealFinancialCase(input: DeepSimplify<virtengine_settlement_v1_tx.MsgAppealFinancialCase>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).appealFinancialCase(input, options);
          }, { path: [49, 14] }),
          /**
           * cancelFinancialCase cancels an unreviewed case and restores its holds.
           */
          cancelFinancialCase: withMetadata(async function cancelFinancialCase(input: DeepSimplify<virtengine_settlement_v1_tx.MsgCancelFinancialCase>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).cancelFinancialCase(input, options);
          }, { path: [49, 15] }),
          /**
           * finalizeFinancialCase applies the pending allocation exactly once.
           */
          finalizeFinancialCase: withMetadata(async function finalizeFinancialCase(input: DeepSimplify<virtengine_settlement_v1_tx.MsgFinalizeFinancialCase>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).finalizeFinancialCase(input, options);
          }, { path: [49, 16] }),
          /**
           * recordFiatConversionObservation records one authenticated, replay-safe
           * observation produced by the provider's off-chain conversion orchestrator.
           */
          recordFiatConversionObservation: withMetadata(async function recordFiatConversionObservation(input: DeepSimplify<virtengine_settlement_v1_tx.MsgRecordFiatConversionObservation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).recordFiatConversionObservation(input, options);
          }, { path: [49, 17] }),
          /**
           * updateParams updates settlement parameters through the configured x/gov authority.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_settlement_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [49, 18] })
        }
      },
      staking: {
        v1: {
          /**
           * getParams queries the module parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_staking_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).params(input, options);
          }, { path: [50, 0] }),
          /**
           * getValidatorPerformance queries a validator performance record.
           */
          getValidatorPerformance: withMetadata(async function getValidatorPerformance(input: DeepPartial<virtengine_staking_v1_query.QueryValidatorPerformanceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).validatorPerformance(input, options);
          }, { path: [50, 1] }),
          /**
           * getValidatorPerformances queries performance records for an epoch.
           */
          getValidatorPerformances: withMetadata(async function getValidatorPerformances(input: DeepPartial<virtengine_staking_v1_query.QueryValidatorPerformancesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).validatorPerformances(input, options);
          }, { path: [50, 2] }),
          /**
           * getValidatorReward queries a validator reward for an epoch.
           */
          getValidatorReward: withMetadata(async function getValidatorReward(input: DeepPartial<virtengine_staking_v1_query.QueryValidatorRewardRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).validatorReward(input, options);
          }, { path: [50, 3] }),
          /**
           * getValidatorRewards queries all rewards for a validator.
           */
          getValidatorRewards: withMetadata(async function getValidatorRewards(input: DeepPartial<virtengine_staking_v1_query.QueryValidatorRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).validatorRewards(input, options);
          }, { path: [50, 4] }),
          /**
           * getRewardEpoch queries a reward epoch by number.
           */
          getRewardEpoch: withMetadata(async function getRewardEpoch(input: DeepPartial<virtengine_staking_v1_query.QueryRewardEpochRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).rewardEpoch(input, options);
          }, { path: [50, 5] }),
          /**
           * getSlashRecords queries slash records for a validator.
           */
          getSlashRecords: withMetadata(async function getSlashRecords(input: DeepPartial<virtengine_staking_v1_query.QuerySlashRecordsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).slashRecords(input, options);
          }, { path: [50, 6] }),
          /**
           * getSigningInfo queries validator signing info.
           */
          getSigningInfo: withMetadata(async function getSigningInfo(input: DeepPartial<virtengine_staking_v1_query.QuerySigningInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).signingInfo(input, options);
          }, { path: [50, 7] }),
          /**
           * getCurrentEpoch queries the current epoch.
           */
          getCurrentEpoch: withMetadata(async function getCurrentEpoch(input: DeepPartial<virtengine_staking_v1_query.QueryCurrentEpochRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getClient(service).currentEpoch(input, options);
          }, { path: [50, 8] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_staking_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(51);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [51, 0] }),
          /**
           * slashValidator slashes a validator for misbehavior
           */
          slashValidator: withMetadata(async function slashValidator(input: DeepSimplify<virtengine_staking_v1_tx.MsgSlashValidator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(51);
            return getMsgClient(service).slashValidator(input, options);
          }, { path: [51, 1] }),
          /**
           * unjailValidator unjails a validator
           */
          unjailValidator: withMetadata(async function unjailValidator(input: DeepSimplify<virtengine_staking_v1_tx.MsgUnjailValidator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(51);
            return getMsgClient(service).unjailValidator(input, options);
          }, { path: [51, 2] }),
          /**
           * recordPerformance records validator performance metrics
           */
          recordPerformance: withMetadata(async function recordPerformance(input: DeepSimplify<virtengine_staking_v1_tx.MsgRecordPerformance>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(51);
            return getMsgClient(service).recordPerformance(input, options);
          }, { path: [51, 3] })
        }
      },
      support: {
        v1: {
          getSupportRequest: withMetadata(async function getSupportRequest(input: DeepPartial<virtengine_support_v1_query.QuerySupportRequestRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(52);
            return getClient(service).supportRequest(input, options);
          }, { path: [52, 0] }),
          getSupportRequestsBySubmitter: withMetadata(async function getSupportRequestsBySubmitter(input: DeepPartial<virtengine_support_v1_query.QuerySupportRequestsBySubmitterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(52);
            return getClient(service).supportRequestsBySubmitter(input, options);
          }, { path: [52, 1] }),
          getSupportResponsesByRequest: withMetadata(async function getSupportResponsesByRequest(input: DeepPartial<virtengine_support_v1_query.QuerySupportResponsesByRequestRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(52);
            return getClient(service).supportResponsesByRequest(input, options);
          }, { path: [52, 2] }),
          getExternalRef: withMetadata(async function getExternalRef(input: DeepPartial<virtengine_support_v1_query.QueryExternalRefRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(52);
            return getClient(service).externalRef(input, options);
          }, { path: [52, 3] }),
          getExternalRefsByOwner: withMetadata(async function getExternalRefsByOwner(input: DeepPartial<virtengine_support_v1_query.QueryExternalRefsByOwnerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(52);
            return getClient(service).externalRefsByOwner(input, options);
          }, { path: [52, 4] }),
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_support_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(52);
            return getClient(service).params(input, options);
          }, { path: [52, 5] }),
          createSupportRequest: withMetadata(async function createSupportRequest(input: DeepSimplify<virtengine_support_v1_tx.MsgCreateSupportRequest>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).createSupportRequest(input, options);
          }, { path: [53, 0] }),
          updateSupportRequest: withMetadata(async function updateSupportRequest(input: DeepSimplify<virtengine_support_v1_tx.MsgUpdateSupportRequest>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).updateSupportRequest(input, options);
          }, { path: [53, 1] }),
          addSupportResponse: withMetadata(async function addSupportResponse(input: DeepSimplify<virtengine_support_v1_tx.MsgAddSupportResponse>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).addSupportResponse(input, options);
          }, { path: [53, 2] }),
          archiveSupportRequest: withMetadata(async function archiveSupportRequest(input: DeepSimplify<virtengine_support_v1_tx.MsgArchiveSupportRequest>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).archiveSupportRequest(input, options);
          }, { path: [53, 3] }),
          registerExternalTicket: withMetadata(async function registerExternalTicket(input: DeepSimplify<virtengine_support_v1_tx.MsgRegisterExternalTicket>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).registerExternalTicket(input, options);
          }, { path: [53, 4] }),
          updateExternalTicket: withMetadata(async function updateExternalTicket(input: DeepSimplify<virtengine_support_v1_tx.MsgUpdateExternalTicket>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).updateExternalTicket(input, options);
          }, { path: [53, 5] }),
          removeExternalTicket: withMetadata(async function removeExternalTicket(input: DeepSimplify<virtengine_support_v1_tx.MsgRemoveExternalTicket>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).removeExternalTicket(input, options);
          }, { path: [53, 6] }),
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_support_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(53);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [53, 7] })
        }
      },
      take: {
        v1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_take_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(54);
            return getClient(service).params(input, options);
          }, { path: [54, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/market module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v1.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_take_v1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(55);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [55, 0] })
        }
      },
      veid: {
        v1: {
          /**
           * getIdentityRecord queries an identity record by account address
           */
          getIdentityRecord: withMetadata(async function getIdentityRecord(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityRecordRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).identityRecord(input, options);
          }, { path: [57, 0] }),
          /**
           * getIdentity queries an identity record by account address (alias for IdentityRecord)
           */
          getIdentity: withMetadata(async function getIdentity(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).identity(input, options);
          }, { path: [57, 1] }),
          /**
           * getScope queries a specific scope by ID
           */
          getScope: withMetadata(async function getScope(input: DeepPartial<virtengine_veid_v1_query.QueryScopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).scope(input, options);
          }, { path: [57, 2] }),
          /**
           * getScopes queries all scopes for an account
           */
          getScopes: withMetadata(async function getScopes(input: DeepPartial<virtengine_veid_v1_query.QueryScopesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).scopes(input, options);
          }, { path: [57, 3] }),
          /**
           * getScopesByType queries all scopes of a specific type for an account
           */
          getScopesByType: withMetadata(async function getScopesByType(input: DeepPartial<virtengine_veid_v1_query.QueryScopesByTypeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).scopesByType(input, options);
          }, { path: [57, 4] }),
          /**
           * getIdentityScore queries the identity score for an account
           */
          getIdentityScore: withMetadata(async function getIdentityScore(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityScoreRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).identityScore(input, options);
          }, { path: [57, 5] }),
          /**
           * getIdentityStatus queries the identity status for an account
           */
          getIdentityStatus: withMetadata(async function getIdentityStatus(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityStatusRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).identityStatus(input, options);
          }, { path: [57, 6] }),
          /**
           * getIdentityWallet queries an identity wallet by account address
           */
          getIdentityWallet: withMetadata(async function getIdentityWallet(input: DeepPartial<virtengine_veid_v1_query.QueryIdentityWalletRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).identityWallet(input, options);
          }, { path: [57, 7] }),
          /**
           * getWalletScopes queries all scope references in a wallet
           */
          getWalletScopes: withMetadata(async function getWalletScopes(input: DeepPartial<virtengine_veid_v1_query.QueryWalletScopesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).walletScopes(input, options);
          }, { path: [57, 8] }),
          /**
           * getConsentSettings queries consent settings for an account
           */
          getConsentSettings: withMetadata(async function getConsentSettings(input: DeepPartial<virtengine_veid_v1_query.QueryConsentSettingsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).consentSettings(input, options);
          }, { path: [57, 9] }),
          /**
           * getDerivedFeatures queries derived features metadata for an account
           */
          getDerivedFeatures: withMetadata(async function getDerivedFeatures(input: DeepPartial<virtengine_veid_v1_query.QueryDerivedFeaturesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).derivedFeatures(input, options);
          }, { path: [57, 10] }),
          /**
           * getDerivedFeatureHashes queries derived feature hashes for an account (consent-gated)
           */
          getDerivedFeatureHashes: withMetadata(async function getDerivedFeatureHashes(input: DeepPartial<virtengine_veid_v1_query.QueryDerivedFeatureHashesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).derivedFeatureHashes(input, options);
          }, { path: [57, 11] }),
          /**
           * getVerificationHistory queries verification history for an account
           */
          getVerificationHistory: withMetadata(async function getVerificationHistory(input: DeepPartial<virtengine_veid_v1_query.QueryVerificationHistoryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).verificationHistory(input, options);
          }, { path: [57, 12] }),
          /**
           * getApprovedClients queries all approved clients
           */
          getApprovedClients: withMetadata(async function getApprovedClients(input: DeepPartial<virtengine_veid_v1_query.QueryApprovedClientsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).approvedClients(input, options);
          }, { path: [57, 13] }),
          /**
           * getParams queries the module parameters
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_veid_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).params(input, options);
          }, { path: [57, 14] }),
          /**
           * getSSOLinkage queries SSO linkage metadata
           */
          getSSOLinkage: withMetadata(async function getSSOLinkage(input: DeepPartial<virtengine_veid_v1_query.QuerySSOLinkageRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).sSOLinkage(input, options);
          }, { path: [57, 15] }),
          /**
           * getEmailVerification queries an email verification record
           */
          getEmailVerification: withMetadata(async function getEmailVerification(input: DeepPartial<virtengine_veid_v1_query.QueryEmailVerificationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).emailVerification(input, options);
          }, { path: [57, 16] }),
          /**
           * getSMSVerification queries an SMS verification record
           */
          getSMSVerification: withMetadata(async function getSMSVerification(input: DeepPartial<virtengine_veid_v1_query.QuerySMSVerificationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).sMSVerification(input, options);
          }, { path: [57, 17] }),
          /**
           * getSocialMediaScope queries a social media scope by scope ID
           */
          getSocialMediaScope: withMetadata(async function getSocialMediaScope(input: DeepPartial<virtengine_veid_v1_query.QuerySocialMediaScopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).socialMediaScope(input, options);
          }, { path: [57, 18] }),
          /**
           * getSocialMediaScopes queries social media scopes for an account
           */
          getSocialMediaScopes: withMetadata(async function getSocialMediaScopes(input: DeepPartial<virtengine_veid_v1_query.QuerySocialMediaScopesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).socialMediaScopes(input, options);
          }, { path: [57, 19] }),
          /**
           * getBorderlineParams queries the borderline parameters
           */
          getBorderlineParams: withMetadata(async function getBorderlineParams(input: DeepPartial<virtengine_veid_v1_query.QueryBorderlineParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).borderlineParams(input, options);
          }, { path: [57, 20] }),
          /**
           * getAppeal queries a specific appeal by ID
           */
          getAppeal: withMetadata(async function getAppeal(input: DeepPartial<virtengine_veid_v1_query.QueryAppealRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).appeal(input, options);
          }, { path: [57, 21] }),
          /**
           * getAppeals queries all appeals for an account
           */
          getAppeals: withMetadata(async function getAppeals(input: DeepPartial<virtengine_veid_v1_query.QueryAppealsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).appeals(input, options);
          }, { path: [57, 22] }),
          /**
           * getAppealsByScope queries all appeals for a specific scope
           */
          getAppealsByScope: withMetadata(async function getAppealsByScope(input: DeepPartial<virtengine_veid_v1_query.QueryAppealsByScopeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).appealsByScope(input, options);
          }, { path: [57, 23] }),
          /**
           * getAppealParams queries the appeal system parameters
           */
          getAppealParams: withMetadata(async function getAppealParams(input: DeepPartial<virtengine_veid_v1_query.QueryAppealParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).appealParams(input, options);
          }, { path: [57, 24] }),
          /**
           * getComplianceStatus queries the compliance status for an account
           */
          getComplianceStatus: withMetadata(async function getComplianceStatus(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceStatusRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).complianceStatus(input, options);
          }, { path: [57, 25] }),
          /**
           * getComplianceProvider queries a specific compliance provider
           */
          getComplianceProvider: withMetadata(async function getComplianceProvider(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceProviderRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).complianceProvider(input, options);
          }, { path: [57, 26] }),
          /**
           * getComplianceProviders queries all compliance providers
           */
          getComplianceProviders: withMetadata(async function getComplianceProviders(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceProvidersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).complianceProviders(input, options);
          }, { path: [57, 27] }),
          /**
           * getComplianceParams queries the compliance parameters
           */
          getComplianceParams: withMetadata(async function getComplianceParams(input: DeepPartial<virtengine_veid_v1_query.QueryComplianceParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).complianceParams(input, options);
          }, { path: [57, 28] }),
          /**
           * getModelVersion queries the active model for a given type
           */
          getModelVersion: withMetadata(async function getModelVersion(input: DeepPartial<virtengine_veid_v1_query.QueryModelVersionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).modelVersion(input, options);
          }, { path: [57, 29] }),
          /**
           * getActiveModels queries all active models
           */
          getActiveModels: withMetadata(async function getActiveModels(input: DeepPartial<virtengine_veid_v1_query.QueryActiveModelsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).activeModels(input, options);
          }, { path: [57, 30] }),
          /**
           * getModelHistory queries the version history for a model type
           */
          getModelHistory: withMetadata(async function getModelHistory(input: DeepPartial<virtengine_veid_v1_query.QueryModelHistoryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).modelHistory(input, options);
          }, { path: [57, 31] }),
          /**
           * getValidatorModelSync queries a validator's model sync status
           */
          getValidatorModelSync: withMetadata(async function getValidatorModelSync(input: DeepPartial<virtengine_veid_v1_query.QueryValidatorModelSyncRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).validatorModelSync(input, options);
          }, { path: [57, 32] }),
          /**
           * getModelParams queries the model management parameters
           */
          getModelParams: withMetadata(async function getModelParams(input: DeepPartial<virtengine_veid_v1_query.QueryModelParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(57);
            return getClient(service).modelParams(input, options);
          }, { path: [57, 33] }),
          /**
           * uploadScope uploads an identity scope
           */
          uploadScope: withMetadata(async function uploadScope(input: DeepSimplify<virtengine_veid_v1_tx.MsgUploadScope>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).uploadScope(input, options);
          }, { path: [58, 0] }),
          /**
           * revokeScope revokes an identity scope
           */
          revokeScope: withMetadata(async function revokeScope(input: DeepSimplify<virtengine_veid_v1_tx.MsgRevokeScope>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).revokeScope(input, options);
          }, { path: [58, 1] }),
          /**
           * requestVerification requests verification of a scope
           */
          requestVerification: withMetadata(async function requestVerification(input: DeepSimplify<virtengine_veid_v1_tx.MsgRequestVerification>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).requestVerification(input, options);
          }, { path: [58, 2] }),
          /**
           * updateVerificationStatus updates the verification status (validator only)
           */
          updateVerificationStatus: withMetadata(async function updateVerificationStatus(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateVerificationStatus>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateVerificationStatus(input, options);
          }, { path: [58, 3] }),
          /**
           * updateScore updates the identity score (validator only)
           */
          updateScore: withMetadata(async function updateScore(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateScore>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateScore(input, options);
          }, { path: [58, 4] }),
          /**
           * submitConsensusVerification applies the canonical aggregate carried by the
           * index-zero system transaction. It has no ordinary transaction signer and
           * is admitted only by the application proposal/finalization boundary.
           */
          submitConsensusVerification: withMetadata(async function submitConsensusVerification(input: DeepSimplify<virtengine_veid_v1_tx.MsgSubmitConsensusVerification>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitConsensusVerification(input, options);
          }, { path: [58, 5] }),
          /**
           * createIdentityWallet creates an identity wallet
           */
          createIdentityWallet: withMetadata(async function createIdentityWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgCreateIdentityWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).createIdentityWallet(input, options);
          }, { path: [58, 6] }),
          /**
           * addScopeToWallet adds a scope reference to a wallet
           */
          addScopeToWallet: withMetadata(async function addScopeToWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgAddScopeToWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).addScopeToWallet(input, options);
          }, { path: [58, 7] }),
          /**
           * revokeScopeFromWallet revokes a scope from a wallet
           */
          revokeScopeFromWallet: withMetadata(async function revokeScopeFromWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgRevokeScopeFromWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).revokeScopeFromWallet(input, options);
          }, { path: [58, 8] }),
          /**
           * updateConsentSettings updates consent settings
           */
          updateConsentSettings: withMetadata(async function updateConsentSettings(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateConsentSettings>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateConsentSettings(input, options);
          }, { path: [58, 9] }),
          /**
           * rebindWallet rebinds a wallet during key rotation
           */
          rebindWallet: withMetadata(async function rebindWallet(input: DeepSimplify<virtengine_veid_v1_tx.MsgRebindWallet>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).rebindWallet(input, options);
          }, { path: [58, 10] }),
          /**
           * updateDerivedFeatures updates derived features (validator only)
           */
          updateDerivedFeatures: withMetadata(async function updateDerivedFeatures(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateDerivedFeatures>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateDerivedFeatures(input, options);
          }, { path: [58, 11] }),
          /**
           * completeBorderlineFallback completes a borderline fallback after MFA
           */
          completeBorderlineFallback: withMetadata(async function completeBorderlineFallback(input: DeepSimplify<virtengine_veid_v1_tx.MsgCompleteBorderlineFallback>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).completeBorderlineFallback(input, options);
          }, { path: [58, 12] }),
          /**
           * updateBorderlineParams updates borderline parameters (governance only)
           */
          updateBorderlineParams: withMetadata(async function updateBorderlineParams(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateBorderlineParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateBorderlineParams(input, options);
          }, { path: [58, 13] }),
          /**
           * updateParams updates module parameters (governance only)
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_veid_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [58, 14] }),
          /**
           * submitSSOVerificationProof submits an SSO verification proof
           */
          submitSSOVerificationProof: withMetadata(async function submitSSOVerificationProof(input: DeepSimplify<virtengine_veid_v1_tx.MsgSubmitSSOVerificationProof>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitSSOVerificationProof(input, options);
          }, { path: [58, 15] }),
          /**
           * submitEmailVerificationProof submits an email OTP verification proof
           */
          submitEmailVerificationProof: withMetadata(async function submitEmailVerificationProof(input: DeepSimplify<virtengine_veid_v1_tx.MsgSubmitEmailVerificationProof>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitEmailVerificationProof(input, options);
          }, { path: [58, 16] }),
          /**
           * submitSMSVerificationProof submits an SMS OTP verification proof
           */
          submitSMSVerificationProof: withMetadata(async function submitSMSVerificationProof(input: DeepSimplify<virtengine_veid_v1_tx.MsgSubmitSMSVerificationProof>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitSMSVerificationProof(input, options);
          }, { path: [58, 17] }),
          /**
           * submitSocialMediaScope submits a social media profile scope
           */
          submitSocialMediaScope: withMetadata(async function submitSocialMediaScope(input: DeepSimplify<virtengine_veid_v1_tx.MsgSubmitSocialMediaScope>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitSocialMediaScope(input, options);
          }, { path: [58, 18] }),
          /**
           * submitAppeal submits an appeal against a verification decision
           */
          submitAppeal: withMetadata(async function submitAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgSubmitAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitAppeal(input, options);
          }, { path: [58, 19] }),
          /**
           * claimAppeal allows an arbitrator to claim an appeal for review
           */
          claimAppeal: withMetadata(async function claimAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgClaimAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).claimAppeal(input, options);
          }, { path: [58, 20] }),
          /**
           * resolveAppeal resolves an appeal (arbitrator/governance only)
           */
          resolveAppeal: withMetadata(async function resolveAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgResolveAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).resolveAppeal(input, options);
          }, { path: [58, 21] }),
          /**
           * withdrawAppeal allows the submitter to withdraw their appeal
           */
          withdrawAppeal: withMetadata(async function withdrawAppeal(input: DeepSimplify<virtengine_veid_v1_appeal.MsgWithdrawAppeal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).withdrawAppeal(input, options);
          }, { path: [58, 22] }),
          /**
           * submitComplianceCheck submits external compliance check results
           */
          submitComplianceCheck: withMetadata(async function submitComplianceCheck(input: DeepSimplify<virtengine_veid_v1_compliance.MsgSubmitComplianceCheck>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).submitComplianceCheck(input, options);
          }, { path: [58, 23] }),
          /**
           * attestCompliance allows validators to attest compliance status
           */
          attestCompliance: withMetadata(async function attestCompliance(input: DeepSimplify<virtengine_veid_v1_compliance.MsgAttestCompliance>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).attestCompliance(input, options);
          }, { path: [58, 24] }),
          /**
           * updateComplianceParams updates compliance configuration (gov only)
           */
          updateComplianceParams: withMetadata(async function updateComplianceParams(input: DeepSimplify<virtengine_veid_v1_compliance.MsgUpdateComplianceParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).updateComplianceParams(input, options);
          }, { path: [58, 25] }),
          /**
           * registerComplianceProvider registers a new compliance provider (gov only)
           */
          registerComplianceProvider: withMetadata(async function registerComplianceProvider(input: DeepSimplify<virtengine_veid_v1_compliance.MsgRegisterComplianceProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).registerComplianceProvider(input, options);
          }, { path: [58, 26] }),
          /**
           * deactivateComplianceProvider deactivates a compliance provider (gov only)
           */
          deactivateComplianceProvider: withMetadata(async function deactivateComplianceProvider(input: DeepSimplify<virtengine_veid_v1_compliance.MsgDeactivateComplianceProvider>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).deactivateComplianceProvider(input, options);
          }, { path: [58, 27] }),
          /**
           * registerModel registers a new ML model (authorized only)
           */
          registerModel: withMetadata(async function registerModel(input: DeepSimplify<virtengine_veid_v1_model.MsgRegisterModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).registerModel(input, options);
          }, { path: [58, 28] }),
          /**
           * proposeModelUpdate proposes updating active model via governance
           */
          proposeModelUpdate: withMetadata(async function proposeModelUpdate(input: DeepSimplify<virtengine_veid_v1_model.MsgProposeModelUpdate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).proposeModelUpdate(input, options);
          }, { path: [58, 29] }),
          /**
           * reportModelVersion reports validator's model versions
           */
          reportModelVersion: withMetadata(async function reportModelVersion(input: DeepSimplify<virtengine_veid_v1_model.MsgReportModelVersion>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).reportModelVersion(input, options);
          }, { path: [58, 30] }),
          /**
           * activateModel activates a pending model after governance approval
           */
          activateModel: withMetadata(async function activateModel(input: DeepSimplify<virtengine_veid_v1_model.MsgActivateModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).activateModel(input, options);
          }, { path: [58, 31] }),
          /**
           * deprecateModel deprecates a model
           */
          deprecateModel: withMetadata(async function deprecateModel(input: DeepSimplify<virtengine_veid_v1_model.MsgDeprecateModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).deprecateModel(input, options);
          }, { path: [58, 32] }),
          /**
           * revokeModel revokes a model
           */
          revokeModel: withMetadata(async function revokeModel(input: DeepSimplify<virtengine_veid_v1_model.MsgRevokeModel>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(58);
            return getMsgClient(service).revokeModel(input, options);
          }, { path: [58, 33] })
        }
      },
      wasm: {
        v1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<virtengine_wasm_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(59);
            return getClient(service).params(input, options);
          }, { path: [59, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/wasm module
           * parameters. The authority is hard-coded to the x/gov module account.
           *
           * Since: akash v2.0.0
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<virtengine_wasm_v1_paramsmsg.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(60);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [60, 0] })
        }
      }
    }
  };
}
