import { createServiceLoader } from "../sdk/client/createServiceLoader.ts";
import { SDKOptions } from "../sdk/types.ts";

import type * as cosmos_app_v1alpha1_query from "./protos/cosmos/app/v1alpha1/query.ts";
import type * as cosmos_auth_v1beta1_query from "./protos/cosmos/auth/v1beta1/query.ts";
import type * as cosmos_auth_v1beta1_tx from "./protos/cosmos/auth/v1beta1/tx.ts";
import type * as cosmos_authz_v1beta1_query from "./protos/cosmos/authz/v1beta1/query.ts";
import type * as cosmos_authz_v1beta1_tx from "./protos/cosmos/authz/v1beta1/tx.ts";
import type * as cosmos_autocli_v1_query from "./protos/cosmos/autocli/v1/query.ts";
import type * as cosmos_bank_v1beta1_query from "./protos/cosmos/bank/v1beta1/query.ts";
import type * as cosmos_bank_v1beta1_tx from "./protos/cosmos/bank/v1beta1/tx.ts";
import type * as tendermint_abci_types from "./protos/tendermint/abci/types.ts";
import type * as cosmos_base_node_v1beta1_query from "./protos/cosmos/base/node/v1beta1/query.ts";
import type * as cosmos_base_reflection_v1beta1_reflection from "./protos/cosmos/base/reflection/v1beta1/reflection.ts";
import type * as cosmos_base_reflection_v2alpha1_reflection from "./protos/cosmos/base/reflection/v2alpha1/reflection.ts";
import type * as cosmos_base_tendermint_v1beta1_query from "./protos/cosmos/base/tendermint/v1beta1/query.ts";
import type * as cosmos_benchmark_v1_tx from "./protos/cosmos/benchmark/v1/tx.ts";
import type * as cosmos_circuit_v1_query from "./protos/cosmos/circuit/v1/query.ts";
import type * as cosmos_circuit_v1_tx from "./protos/cosmos/circuit/v1/tx.ts";
import type * as cosmos_consensus_v1_query from "./protos/cosmos/consensus/v1/query.ts";
import type * as cosmos_consensus_v1_tx from "./protos/cosmos/consensus/v1/tx.ts";
import type * as cosmos_counter_v1_query from "./protos/cosmos/counter/v1/query.ts";
import type * as cosmos_counter_v1_tx from "./protos/cosmos/counter/v1/tx.ts";
import type * as cosmos_crisis_v1beta1_tx from "./protos/cosmos/crisis/v1beta1/tx.ts";
import type * as cosmos_distribution_v1beta1_query from "./protos/cosmos/distribution/v1beta1/query.ts";
import type * as cosmos_distribution_v1beta1_tx from "./protos/cosmos/distribution/v1beta1/tx.ts";
import type * as cosmos_epochs_v1beta1_query from "./protos/cosmos/epochs/v1beta1/query.ts";
import type * as cosmos_evidence_v1beta1_query from "./protos/cosmos/evidence/v1beta1/query.ts";
import type * as cosmos_evidence_v1beta1_tx from "./protos/cosmos/evidence/v1beta1/tx.ts";
import type * as cosmos_feegrant_v1beta1_query from "./protos/cosmos/feegrant/v1beta1/query.ts";
import type * as cosmos_feegrant_v1beta1_tx from "./protos/cosmos/feegrant/v1beta1/tx.ts";
import type * as cosmos_gov_v1_query from "./protos/cosmos/gov/v1/query.ts";
import type * as cosmos_gov_v1_tx from "./protos/cosmos/gov/v1/tx.ts";
import type * as cosmos_gov_v1beta1_query from "./protos/cosmos/gov/v1beta1/query.ts";
import type * as cosmos_gov_v1beta1_tx from "./protos/cosmos/gov/v1beta1/tx.ts";
import type * as cosmos_group_v1_query from "./protos/cosmos/group/v1/query.ts";
import type * as cosmos_group_v1_tx from "./protos/cosmos/group/v1/tx.ts";
import type * as cosmos_mint_v1beta1_query from "./protos/cosmos/mint/v1beta1/query.ts";
import type * as cosmos_mint_v1beta1_tx from "./protos/cosmos/mint/v1beta1/tx.ts";
import type * as cosmos_nft_v1beta1_query from "./protos/cosmos/nft/v1beta1/query.ts";
import type * as cosmos_nft_v1beta1_tx from "./protos/cosmos/nft/v1beta1/tx.ts";
import type * as cosmos_params_v1beta1_query from "./protos/cosmos/params/v1beta1/query.ts";
import type * as cosmos_protocolpool_v1_query from "./protos/cosmos/protocolpool/v1/query.ts";
import type * as cosmos_protocolpool_v1_tx from "./protos/cosmos/protocolpool/v1/tx.ts";
import type * as cosmos_reflection_v1_reflection from "./protos/cosmos/reflection/v1/reflection.ts";
import type * as cosmos_slashing_v1beta1_query from "./protos/cosmos/slashing/v1beta1/query.ts";
import type * as cosmos_slashing_v1beta1_tx from "./protos/cosmos/slashing/v1beta1/tx.ts";
import type * as cosmos_staking_v1beta1_query from "./protos/cosmos/staking/v1beta1/query.ts";
import type * as cosmos_staking_v1beta1_tx from "./protos/cosmos/staking/v1beta1/tx.ts";
import type * as cosmos_store_streaming_abci_grpc from "./protos/cosmos/store/streaming/abci/grpc.ts";
import type * as cosmos_tx_v1beta1_service from "./protos/cosmos/tx/v1beta1/service.ts";
import type * as cosmos_upgrade_v1beta1_query from "./protos/cosmos/upgrade/v1beta1/query.ts";
import type * as cosmos_upgrade_v1beta1_tx from "./protos/cosmos/upgrade/v1beta1/tx.ts";
import type * as cosmos_vesting_v1beta1_tx from "./protos/cosmos/vesting/v1beta1/tx.ts";
import { createClientFactory } from "../sdk/client/createClientFactory.ts";
import type { Transport, CallOptions, TxCallOptions } from "../sdk/transport/types.ts";
import { withMetadata } from "../sdk/client/sdkMetadata.ts";
import type { DeepPartial, DeepSimplify } from "../encoding/typeEncodingHelpers.ts";


export const serviceLoader= createServiceLoader([
  () => import("./protos/cosmos/app/v1alpha1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/auth/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/auth/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/authz/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/authz/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/autocli/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/bank/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/bank/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/tendermint/abci/types_virtengine.ts").then(m => m.ABCI),
  () => import("./protos/cosmos/base/node/v1beta1/query_virtengine.ts").then(m => m.Service),
  () => import("./protos/cosmos/base/reflection/v1beta1/reflection_virtengine.ts").then(m => m.ReflectionService),
  () => import("./protos/cosmos/base/reflection/v2alpha1/reflection_virtengine.ts").then(m => m.ReflectionService),
  () => import("./protos/cosmos/base/tendermint/v1beta1/query_virtengine.ts").then(m => m.Service),
  () => import("./protos/cosmos/benchmark/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/circuit/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/circuit/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/consensus/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/consensus/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/counter/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/counter/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/crisis/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/distribution/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/distribution/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/epochs/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/evidence/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/evidence/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/feegrant/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/feegrant/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/gov/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/gov/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/gov/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/gov/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/group/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/group/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/mint/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/mint/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/nft/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/nft/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/params/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/protocolpool/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/protocolpool/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/reflection/v1/reflection_virtengine.ts").then(m => m.ReflectionService),
  () => import("./protos/cosmos/slashing/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/slashing/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/staking/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/staking/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/store/streaming/abci/grpc_virtengine.ts").then(m => m.ABCIListenerService),
  () => import("./protos/cosmos/tx/v1beta1/service_virtengine.ts").then(m => m.Service),
  () => import("./protos/cosmos/upgrade/v1beta1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/cosmos/upgrade/v1beta1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/cosmos/vesting/v1beta1/tx_virtengine.ts").then(m => m.Msg)
] as const);
export function createSDK(queryTransport: Transport, txTransport: Transport, options?: SDKOptions) {
  const getClient = createClientFactory<CallOptions>(queryTransport, options?.clientOptions);
  const getMsgClient = createClientFactory<TxCallOptions>(txTransport, options?.clientOptions);
  return {
    cosmos: {
      app: {
        v1alpha1: {
          /**
           * getConfig returns the current app config.
           * @deprecated
           */
          getConfig: withMetadata(async function getConfig(input: DeepPartial<cosmos_app_v1alpha1_query.QueryConfigRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).config(input, options);
          }, { path: [0, 0] })
        }
      },
      auth: {
        v1beta1: {
          /**
           * getAccounts returns all the existing accounts.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<cosmos_auth_v1beta1_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).accounts(input, options);
          }, { path: [1, 0] }),
          /**
           * getAccount returns account details based on address.
           */
          getAccount: withMetadata(async function getAccount(input: DeepPartial<cosmos_auth_v1beta1_query.QueryAccountRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).account(input, options);
          }, { path: [1, 1] }),
          /**
           * getAccountAddressByID returns account address based on account number.
           */
          getAccountAddressByID: withMetadata(async function getAccountAddressByID(input: DeepPartial<cosmos_auth_v1beta1_query.QueryAccountAddressByIDRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).accountAddressByID(input, options);
          }, { path: [1, 2] }),
          /**
           * getParams queries all parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_auth_v1beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).params(input, options);
          }, { path: [1, 3] }),
          /**
           * getModuleAccounts returns all the existing module accounts.
           */
          getModuleAccounts: withMetadata(async function getModuleAccounts(input: DeepPartial<cosmos_auth_v1beta1_query.QueryModuleAccountsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).moduleAccounts(input, options);
          }, { path: [1, 4] }),
          /**
           * getModuleAccountByName returns the module account info by module name
           */
          getModuleAccountByName: withMetadata(async function getModuleAccountByName(input: DeepPartial<cosmos_auth_v1beta1_query.QueryModuleAccountByNameRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).moduleAccountByName(input, options);
          }, { path: [1, 5] }),
          /**
           * getBech32Prefix queries bech32Prefix
           */
          getBech32Prefix: withMetadata(async function getBech32Prefix(input: DeepPartial<cosmos_auth_v1beta1_query.Bech32PrefixRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).bech32Prefix(input, options);
          }, { path: [1, 6] }),
          /**
           * getAddressBytesToString converts Account Address bytes to string
           */
          getAddressBytesToString: withMetadata(async function getAddressBytesToString(input: DeepPartial<cosmos_auth_v1beta1_query.AddressBytesToStringRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).addressBytesToString(input, options);
          }, { path: [1, 7] }),
          /**
           * getAddressStringToBytes converts Address string to bytes
           */
          getAddressStringToBytes: withMetadata(async function getAddressStringToBytes(input: DeepPartial<cosmos_auth_v1beta1_query.AddressStringToBytesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).addressStringToBytes(input, options);
          }, { path: [1, 8] }),
          /**
           * getAccountInfo queries account info which is common to all account types.
           */
          getAccountInfo: withMetadata(async function getAccountInfo(input: DeepPartial<cosmos_auth_v1beta1_query.QueryAccountInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).accountInfo(input, options);
          }, { path: [1, 9] }),
          /**
           * updateParams defines a (governance) operation for updating the x/auth module
           * parameters. The authority defaults to the x/gov module account.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_auth_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(2);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [2, 0] })
        }
      },
      authz: {
        v1beta1: {
          /**
           * Returns list of `Authorization`, granted to the grantee by the granter.
           */
          getGrants: withMetadata(async function getGrants(input: DeepPartial<cosmos_authz_v1beta1_query.QueryGrantsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).grants(input, options);
          }, { path: [3, 0] }),
          /**
           * getGranterGrants returns list of `GrantAuthorization`, granted by granter.
           */
          getGranterGrants: withMetadata(async function getGranterGrants(input: DeepPartial<cosmos_authz_v1beta1_query.QueryGranterGrantsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).granterGrants(input, options);
          }, { path: [3, 1] }),
          /**
           * getGranteeGrants returns a list of `GrantAuthorization` by grantee.
           */
          getGranteeGrants: withMetadata(async function getGranteeGrants(input: DeepPartial<cosmos_authz_v1beta1_query.QueryGranteeGrantsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).granteeGrants(input, options);
          }, { path: [3, 2] }),
          /**
           * grant grants the provided authorization to the grantee on the granter's
           * account with the provided expiration time. If there is already a grant
           * for the given (granter, grantee, Authorization) triple, then the grant
           * will be overwritten.
           */
          grant: withMetadata(async function grant(input: DeepSimplify<cosmos_authz_v1beta1_tx.MsgGrant>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).grant(input, options);
          }, { path: [4, 0] }),
          /**
           * exec attempts to execute the provided messages using
           * authorizations granted to the grantee. Each message should have only
           * one signer corresponding to the granter of the authorization.
           */
          exec: withMetadata(async function exec(input: DeepSimplify<cosmos_authz_v1beta1_tx.MsgExec>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).exec(input, options);
          }, { path: [4, 1] }),
          /**
           * revoke revokes any authorization corresponding to the provided method name on the
           * granter's account that has been granted to the grantee.
           */
          revoke: withMetadata(async function revoke(input: DeepSimplify<cosmos_authz_v1beta1_tx.MsgRevoke>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(4);
            return getMsgClient(service).revoke(input, options);
          }, { path: [4, 2] })
        }
      },
      autocli: {
        v1: {
          /**
           * getAppOptions returns the autocli options for all of the modules in an app.
           */
          getAppOptions: withMetadata(async function getAppOptions(input: DeepPartial<cosmos_autocli_v1_query.AppOptionsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(5);
            return getClient(service).appOptions(input, options);
          }, { path: [5, 0] })
        }
      },
      bank: {
        v1beta1: {
          /**
           * getBalance queries the balance of a single coin for a single account.
           */
          getBalance: withMetadata(async function getBalance(input: DeepPartial<cosmos_bank_v1beta1_query.QueryBalanceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).balance(input, options);
          }, { path: [6, 0] }),
          /**
           * getAllBalances queries the balance of all coins for a single account.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getAllBalances: withMetadata(async function getAllBalances(input: DeepPartial<cosmos_bank_v1beta1_query.QueryAllBalancesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).allBalances(input, options);
          }, { path: [6, 1] }),
          /**
           * getSpendableBalances queries the spendable balance of all coins for a single
           * account.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getSpendableBalances: withMetadata(async function getSpendableBalances(input: DeepPartial<cosmos_bank_v1beta1_query.QuerySpendableBalancesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).spendableBalances(input, options);
          }, { path: [6, 2] }),
          /**
           * getSpendableBalanceByDenom queries the spendable balance of a single denom for
           * a single account.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getSpendableBalanceByDenom: withMetadata(async function getSpendableBalanceByDenom(input: DeepPartial<cosmos_bank_v1beta1_query.QuerySpendableBalanceByDenomRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).spendableBalanceByDenom(input, options);
          }, { path: [6, 3] }),
          /**
           * getTotalSupply queries the total supply of all coins.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getTotalSupply: withMetadata(async function getTotalSupply(input: DeepPartial<cosmos_bank_v1beta1_query.QueryTotalSupplyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).totalSupply(input, options);
          }, { path: [6, 4] }),
          /**
           * getSupplyOf queries the supply of a single coin.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getSupplyOf: withMetadata(async function getSupplyOf(input: DeepPartial<cosmos_bank_v1beta1_query.QuerySupplyOfRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).supplyOf(input, options);
          }, { path: [6, 5] }),
          /**
           * getParams queries the parameters of x/bank module.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_bank_v1beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).params(input, options);
          }, { path: [6, 6] }),
          /**
           * getDenomsMetadata queries the client metadata for all registered coin
           * denominations.
           */
          getDenomsMetadata: withMetadata(async function getDenomsMetadata(input: DeepPartial<cosmos_bank_v1beta1_query.QueryDenomsMetadataRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).denomsMetadata(input, options);
          }, { path: [6, 7] }),
          /**
           * getDenomMetadata queries the client metadata of a given coin denomination.
           */
          getDenomMetadata: withMetadata(async function getDenomMetadata(input: DeepPartial<cosmos_bank_v1beta1_query.QueryDenomMetadataRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).denomMetadata(input, options);
          }, { path: [6, 8] }),
          /**
           * getDenomMetadataByQueryString queries the client metadata of a given coin denomination.
           */
          getDenomMetadataByQueryString: withMetadata(async function getDenomMetadataByQueryString(input: DeepPartial<cosmos_bank_v1beta1_query.QueryDenomMetadataByQueryStringRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).denomMetadataByQueryString(input, options);
          }, { path: [6, 9] }),
          /**
           * getDenomOwners queries for all account addresses that own a particular token
           * denomination.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getDenomOwners: withMetadata(async function getDenomOwners(input: DeepPartial<cosmos_bank_v1beta1_query.QueryDenomOwnersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).denomOwners(input, options);
          }, { path: [6, 10] }),
          /**
           * getDenomOwnersByQuery queries for all account addresses that own a particular token
           * denomination.
           */
          getDenomOwnersByQuery: withMetadata(async function getDenomOwnersByQuery(input: DeepPartial<cosmos_bank_v1beta1_query.QueryDenomOwnersByQueryRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).denomOwnersByQuery(input, options);
          }, { path: [6, 11] }),
          /**
           * getSendEnabled queries for getSendEnabled entries.
           *
           * This query only returns denominations that have specific getSendEnabled settings.
           * Any denomination that does not have a specific setting will use the default
           * params.default_send_enabled, and will not be returned by this query.
           */
          getSendEnabled: withMetadata(async function getSendEnabled(input: DeepPartial<cosmos_bank_v1beta1_query.QuerySendEnabledRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(6);
            return getClient(service).sendEnabled(input, options);
          }, { path: [6, 12] }),
          /**
           * send defines a method for sending coins from one account to another account.
           */
          send: withMetadata(async function send(input: DeepSimplify<cosmos_bank_v1beta1_tx.MsgSend>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).send(input, options);
          }, { path: [7, 0] }),
          /**
           * multiSend defines a method for sending coins from some accounts to other accounts.
           */
          multiSend: withMetadata(async function multiSend(input: DeepSimplify<cosmos_bank_v1beta1_tx.MsgMultiSend>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).multiSend(input, options);
          }, { path: [7, 1] }),
          /**
           * updateParams defines a governance operation for updating the x/bank module parameters.
           * The authority is defined in the keeper.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_bank_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [7, 2] }),
          /**
           * setSendEnabled is a governance operation for setting the SendEnabled flag
           * on any number of Denoms. Only the entries to add or update should be
           * included. Entries that already exist in the store, but that aren't
           * included in this message, will be left unchanged.
           */
          setSendEnabled: withMetadata(async function setSendEnabled(input: DeepSimplify<cosmos_bank_v1beta1_tx.MsgSetSendEnabled>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(7);
            return getMsgClient(service).setSendEnabled(input, options);
          }, { path: [7, 3] })
        }
      },
      base: {
        node: {
          v1beta1: {
            /**
             * getConfig queries for the operator configuration.
             */
            getConfig: withMetadata(async function getConfig(input: DeepPartial<cosmos_base_node_v1beta1_query.ConfigRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(9);
              return getClient(service).config(input, options);
            }, { path: [9, 0] }),
            /**
             * getStatus queries for the node status.
             */
            getStatus: withMetadata(async function getStatus(input: DeepPartial<cosmos_base_node_v1beta1_query.StatusRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(9);
              return getClient(service).status(input, options);
            }, { path: [9, 1] })
          }
        },
        reflection: {
          v1beta1: {
            /**
             * getListAllInterfaces lists all the interfaces registered in the interface
             * registry.
             */
            getListAllInterfaces: withMetadata(async function getListAllInterfaces(input: DeepPartial<cosmos_base_reflection_v1beta1_reflection.ListAllInterfacesRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).listAllInterfaces(input, options);
            }, { path: [10, 0] }),
            /**
             * getListImplementations list all the concrete types that implement a given
             * interface.
             */
            getListImplementations: withMetadata(async function getListImplementations(input: DeepPartial<cosmos_base_reflection_v1beta1_reflection.ListImplementationsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).listImplementations(input, options);
            }, { path: [10, 1] })
          },
          v2alpha1: {
            /**
             * getAuthnDescriptor returns information on how to authenticate transactions in the application
             * NOTE: this RPC is still experimental and might be subject to breaking changes or removal in
             * future releases of the cosmos-sdk.
             */
            getAuthnDescriptor: withMetadata(async function getAuthnDescriptor(input: DeepPartial<cosmos_base_reflection_v2alpha1_reflection.GetAuthnDescriptorRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getClient(service).getAuthnDescriptor(input, options);
            }, { path: [11, 0] }),
            /**
             * getChainDescriptor returns the description of the chain
             */
            getChainDescriptor: withMetadata(async function getChainDescriptor(input: DeepPartial<cosmos_base_reflection_v2alpha1_reflection.GetChainDescriptorRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getClient(service).getChainDescriptor(input, options);
            }, { path: [11, 1] }),
            /**
             * getCodecDescriptor returns the descriptor of the codec of the application
             */
            getCodecDescriptor: withMetadata(async function getCodecDescriptor(input: DeepPartial<cosmos_base_reflection_v2alpha1_reflection.GetCodecDescriptorRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getClient(service).getCodecDescriptor(input, options);
            }, { path: [11, 2] }),
            /**
             * getConfigurationDescriptor returns the descriptor for the sdk.Config of the application
             */
            getConfigurationDescriptor: withMetadata(async function getConfigurationDescriptor(input: DeepPartial<cosmos_base_reflection_v2alpha1_reflection.GetConfigurationDescriptorRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getClient(service).getConfigurationDescriptor(input, options);
            }, { path: [11, 3] }),
            /**
             * getQueryServicesDescriptor returns the available gRPC queryable services of the application
             */
            getQueryServicesDescriptor: withMetadata(async function getQueryServicesDescriptor(input: DeepPartial<cosmos_base_reflection_v2alpha1_reflection.GetQueryServicesDescriptorRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getClient(service).getQueryServicesDescriptor(input, options);
            }, { path: [11, 4] }),
            /**
             * getTxDescriptor returns information on the used transaction object and available msgs that can be used
             */
            getTxDescriptor: withMetadata(async function getTxDescriptor(input: DeepPartial<cosmos_base_reflection_v2alpha1_reflection.GetTxDescriptorRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getClient(service).getTxDescriptor(input, options);
            }, { path: [11, 5] })
          }
        },
        tendermint: {
          v1beta1: {
            /**
             * getNodeInfo queries the current node info.
             */
            getNodeInfo: withMetadata(async function getNodeInfo(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.GetNodeInfoRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).getNodeInfo(input, options);
            }, { path: [12, 0] }),
            /**
             * getSyncing queries node syncing.
             */
            getSyncing: withMetadata(async function getSyncing(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.GetSyncingRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).getSyncing(input, options);
            }, { path: [12, 1] }),
            /**
             * getLatestBlock returns the latest block.
             */
            getLatestBlock: withMetadata(async function getLatestBlock(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.GetLatestBlockRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).getLatestBlock(input, options);
            }, { path: [12, 2] }),
            /**
             * getBlockByHeight queries block for given height.
             */
            getBlockByHeight: withMetadata(async function getBlockByHeight(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.GetBlockByHeightRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).getBlockByHeight(input, options);
            }, { path: [12, 3] }),
            /**
             * getLatestValidatorSet queries latest validator-set.
             */
            getLatestValidatorSet: withMetadata(async function getLatestValidatorSet(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.GetLatestValidatorSetRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).getLatestValidatorSet(input, options);
            }, { path: [12, 4] }),
            /**
             * getValidatorSetByHeight queries validator-set at a given height.
             */
            getValidatorSetByHeight: withMetadata(async function getValidatorSetByHeight(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.GetValidatorSetByHeightRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).getValidatorSetByHeight(input, options);
            }, { path: [12, 5] }),
            /**
             * getABCIQuery defines a query handler that supports ABCI queries directly to the
             * application, bypassing Tendermint completely. The ABCI query must contain
             * a valid and supported path, including app, custom, p2p, and store.
             */
            getABCIQuery: withMetadata(async function getABCIQuery(input: DeepPartial<cosmos_base_tendermint_v1beta1_query.ABCIQueryRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).aBCIQuery(input, options);
            }, { path: [12, 6] })
          }
        }
      },
      benchmark: {
        v1: {
          /**
           * loadTest defines a method for executing a sequence of load test operations.
           */
          loadTest: withMetadata(async function loadTest(input: DeepSimplify<cosmos_benchmark_v1_tx.MsgLoadTest>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(13);
            return getMsgClient(service).loadTest(input, options);
          }, { path: [13, 0] })
        }
      },
      circuit: {
        v1: {
          /**
           * getAccount returns account permissions.
           */
          getAccount: withMetadata(async function getAccount(input: DeepPartial<cosmos_circuit_v1_query.QueryAccountRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getClient(service).account(input, options);
          }, { path: [14, 0] }),
          /**
           * Account returns account permissions.
           */
          getAccounts: withMetadata(async function getAccounts(input: DeepPartial<cosmos_circuit_v1_query.QueryAccountsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getClient(service).accounts(input, options);
          }, { path: [14, 1] }),
          /**
           * getDisabledList returns a list of disabled message urls
           */
          getDisabledList: withMetadata(async function getDisabledList(input: DeepPartial<cosmos_circuit_v1_query.QueryDisabledListRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(14);
            return getClient(service).disabledList(input, options);
          }, { path: [14, 2] }),
          /**
           * authorizeCircuitBreaker allows a super-admin to grant (or revoke) another
           * account's circuit breaker permissions.
           */
          authorizeCircuitBreaker: withMetadata(async function authorizeCircuitBreaker(input: DeepSimplify<cosmos_circuit_v1_tx.MsgAuthorizeCircuitBreaker>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getMsgClient(service).authorizeCircuitBreaker(input, options);
          }, { path: [15, 0] }),
          /**
           * tripCircuitBreaker pauses processing of Msg's in the state machine.
           */
          tripCircuitBreaker: withMetadata(async function tripCircuitBreaker(input: DeepSimplify<cosmos_circuit_v1_tx.MsgTripCircuitBreaker>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getMsgClient(service).tripCircuitBreaker(input, options);
          }, { path: [15, 1] }),
          /**
           * resetCircuitBreaker resumes processing of Msg's in the state machine that
           * have been been paused using TripCircuitBreaker.
           */
          resetCircuitBreaker: withMetadata(async function resetCircuitBreaker(input: DeepSimplify<cosmos_circuit_v1_tx.MsgResetCircuitBreaker>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(15);
            return getMsgClient(service).resetCircuitBreaker(input, options);
          }, { path: [15, 2] })
        }
      },
      consensus: {
        v1: {
          /**
           * getParams queries the parameters of x/consensus module.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_consensus_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(16);
            return getClient(service).params(input, options);
          }, { path: [16, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/consensus module parameters.
           * The authority is defined in the keeper.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_consensus_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(17);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [17, 0] })
        }
      },
      counter: {
        v1: {
          /**
           * getCount queries the parameters of x/Counter module.
           */
          getCount: withMetadata(async function getCount(input: DeepPartial<cosmos_counter_v1_query.QueryGetCountRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(18);
            return getClient(service).getCount(input, options);
          }, { path: [18, 0] }),
          /**
           * increaseCount increments the counter by the specified amount.
           */
          increaseCount: withMetadata(async function increaseCount(input: DeepSimplify<cosmos_counter_v1_tx.MsgIncreaseCounter>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(19);
            return getMsgClient(service).increaseCount(input, options);
          }, { path: [19, 0] })
        }
      },
      crisis: {
        v1beta1: {
          /**
           * verifyInvariant defines a method to verify a particular invariant.
           */
          verifyInvariant: withMetadata(async function verifyInvariant(input: DeepSimplify<cosmos_crisis_v1beta1_tx.MsgVerifyInvariant>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getMsgClient(service).verifyInvariant(input, options);
          }, { path: [20, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/crisis module
           * parameters. The authority is defined in the keeper.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_crisis_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(20);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [20, 1] })
        }
      },
      distribution: {
        v1beta1: {
          /**
           * getParams queries params of the distribution module.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).params(input, options);
          }, { path: [21, 0] }),
          /**
           * getValidatorDistributionInfo queries validator commission and self-delegation rewards for validator
           */
          getValidatorDistributionInfo: withMetadata(async function getValidatorDistributionInfo(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryValidatorDistributionInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).validatorDistributionInfo(input, options);
          }, { path: [21, 1] }),
          /**
           * getValidatorOutstandingRewards queries rewards of a validator address.
           */
          getValidatorOutstandingRewards: withMetadata(async function getValidatorOutstandingRewards(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryValidatorOutstandingRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).validatorOutstandingRewards(input, options);
          }, { path: [21, 2] }),
          /**
           * getValidatorCommission queries accumulated commission for a validator.
           */
          getValidatorCommission: withMetadata(async function getValidatorCommission(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryValidatorCommissionRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).validatorCommission(input, options);
          }, { path: [21, 3] }),
          /**
           * getValidatorSlashes queries slash events of a validator.
           */
          getValidatorSlashes: withMetadata(async function getValidatorSlashes(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryValidatorSlashesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).validatorSlashes(input, options);
          }, { path: [21, 4] }),
          /**
           * getDelegationRewards queries the total rewards accrued by a delegation.
           */
          getDelegationRewards: withMetadata(async function getDelegationRewards(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryDelegationRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).delegationRewards(input, options);
          }, { path: [21, 5] }),
          /**
           * getDelegationTotalRewards queries the total rewards accrued by each
           * validator.
           */
          getDelegationTotalRewards: withMetadata(async function getDelegationTotalRewards(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryDelegationTotalRewardsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).delegationTotalRewards(input, options);
          }, { path: [21, 6] }),
          /**
           * getDelegatorValidators queries the validators of a delegator.
           */
          getDelegatorValidators: withMetadata(async function getDelegatorValidators(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryDelegatorValidatorsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).delegatorValidators(input, options);
          }, { path: [21, 7] }),
          /**
           * getDelegatorWithdrawAddress queries withdraw address of a delegator.
           */
          getDelegatorWithdrawAddress: withMetadata(async function getDelegatorWithdrawAddress(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryDelegatorWithdrawAddressRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).delegatorWithdrawAddress(input, options);
          }, { path: [21, 8] }),
          /**
           * getCommunityPool queries the community pool coins.
           *
           * WARNING: This query will fail if an external community pool is used.
           */
          getCommunityPool: withMetadata(async function getCommunityPool(input: DeepPartial<cosmos_distribution_v1beta1_query.QueryCommunityPoolRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(21);
            return getClient(service).communityPool(input, options);
          }, { path: [21, 9] }),
          /**
           * setWithdrawAddress defines a method to change the withdraw address
           * for a delegator (or validator self-delegation).
           */
          setWithdrawAddress: withMetadata(async function setWithdrawAddress(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgSetWithdrawAddress>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).setWithdrawAddress(input, options);
          }, { path: [22, 0] }),
          /**
           * withdrawDelegatorReward defines a method to withdraw rewards of delegator
           * from a single validator.
           */
          withdrawDelegatorReward: withMetadata(async function withdrawDelegatorReward(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgWithdrawDelegatorReward>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).withdrawDelegatorReward(input, options);
          }, { path: [22, 1] }),
          /**
           * withdrawValidatorCommission defines a method to withdraw the
           * full commission to the validator address.
           */
          withdrawValidatorCommission: withMetadata(async function withdrawValidatorCommission(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgWithdrawValidatorCommission>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).withdrawValidatorCommission(input, options);
          }, { path: [22, 2] }),
          /**
           * fundCommunityPool defines a method to allow an account to directly
           * fund the community pool.
           *
           * WARNING: This method will fail if an external community pool is used.
           */
          fundCommunityPool: withMetadata(async function fundCommunityPool(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgFundCommunityPool>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).fundCommunityPool(input, options);
          }, { path: [22, 3] }),
          /**
           * updateParams defines a governance operation for updating the x/distribution
           * module parameters. The authority is defined in the keeper.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [22, 4] }),
          /**
           * communityPoolSpend defines a governance operation for sending tokens from
           * the community pool in the x/distribution module to another account, which
           * could be the governance module itself. The authority is defined in the
           * keeper.
           *
           * WARNING: This method will fail if an external community pool is used.
           */
          communityPoolSpend: withMetadata(async function communityPoolSpend(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgCommunityPoolSpend>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).communityPoolSpend(input, options);
          }, { path: [22, 5] }),
          /**
           * depositValidatorRewardsPool defines a method to provide additional rewards
           * to delegators to a specific validator.
           */
          depositValidatorRewardsPool: withMetadata(async function depositValidatorRewardsPool(input: DeepSimplify<cosmos_distribution_v1beta1_tx.MsgDepositValidatorRewardsPool>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(22);
            return getMsgClient(service).depositValidatorRewardsPool(input, options);
          }, { path: [22, 6] })
        }
      },
      epochs: {
        v1beta1: {
          /**
           * getEpochInfos provide running epochInfos
           */
          getEpochInfos: withMetadata(async function getEpochInfos(input: DeepPartial<cosmos_epochs_v1beta1_query.QueryEpochInfosRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).epochInfos(input, options);
          }, { path: [23, 0] }),
          /**
           * getCurrentEpoch provide current epoch of specified identifier
           */
          getCurrentEpoch: withMetadata(async function getCurrentEpoch(input: DeepPartial<cosmos_epochs_v1beta1_query.QueryCurrentEpochRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(23);
            return getClient(service).currentEpoch(input, options);
          }, { path: [23, 1] })
        }
      },
      evidence: {
        v1beta1: {
          /**
           * getEvidence queries evidence based on evidence hash.
           */
          getEvidence: withMetadata(async function getEvidence(input: DeepPartial<cosmos_evidence_v1beta1_query.QueryEvidenceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).evidence(input, options);
          }, { path: [24, 0] }),
          /**
           * getAllEvidence queries all evidence.
           */
          getAllEvidence: withMetadata(async function getAllEvidence(input: DeepPartial<cosmos_evidence_v1beta1_query.QueryAllEvidenceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(24);
            return getClient(service).allEvidence(input, options);
          }, { path: [24, 1] }),
          /**
           * submitEvidence submits an arbitrary Evidence of misbehavior such as equivocation or
           * counterfactual signing.
           */
          submitEvidence: withMetadata(async function submitEvidence(input: DeepSimplify<cosmos_evidence_v1beta1_tx.MsgSubmitEvidence>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(25);
            return getMsgClient(service).submitEvidence(input, options);
          }, { path: [25, 0] })
        }
      },
      feegrant: {
        v1beta1: {
          /**
           * getAllowance returns granted allwance to the grantee by the granter.
           */
          getAllowance: withMetadata(async function getAllowance(input: DeepPartial<cosmos_feegrant_v1beta1_query.QueryAllowanceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).allowance(input, options);
          }, { path: [26, 0] }),
          /**
           * getAllowances returns all the grants for the given grantee address.
           */
          getAllowances: withMetadata(async function getAllowances(input: DeepPartial<cosmos_feegrant_v1beta1_query.QueryAllowancesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).allowances(input, options);
          }, { path: [26, 1] }),
          /**
           * getAllowancesByGranter returns all the grants given by an address
           */
          getAllowancesByGranter: withMetadata(async function getAllowancesByGranter(input: DeepPartial<cosmos_feegrant_v1beta1_query.QueryAllowancesByGranterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(26);
            return getClient(service).allowancesByGranter(input, options);
          }, { path: [26, 2] }),
          /**
           * grantAllowance grants fee allowance to the grantee on the granter's
           * account with the provided expiration time.
           */
          grantAllowance: withMetadata(async function grantAllowance(input: DeepSimplify<cosmos_feegrant_v1beta1_tx.MsgGrantAllowance>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).grantAllowance(input, options);
          }, { path: [27, 0] }),
          /**
           * revokeAllowance revokes any fee allowance of granter's account that
           * has been granted to the grantee.
           */
          revokeAllowance: withMetadata(async function revokeAllowance(input: DeepSimplify<cosmos_feegrant_v1beta1_tx.MsgRevokeAllowance>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).revokeAllowance(input, options);
          }, { path: [27, 1] }),
          /**
           * pruneAllowances prunes expired fee allowances, currently up to 75 at a time.
           */
          pruneAllowances: withMetadata(async function pruneAllowances(input: DeepSimplify<cosmos_feegrant_v1beta1_tx.MsgPruneAllowances>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(27);
            return getMsgClient(service).pruneAllowances(input, options);
          }, { path: [27, 2] })
        }
      },
      gov: {
        v1: {
          /**
           * getConstitution queries the chain's constitution.
           */
          getConstitution: withMetadata(async function getConstitution(input: DeepPartial<cosmos_gov_v1_query.QueryConstitutionRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).constitution(input, options);
          }, { path: [28, 0] }),
          /**
           * getProposal queries proposal details based on ProposalID.
           */
          getProposal: withMetadata(async function getProposal(input: DeepPartial<cosmos_gov_v1_query.QueryProposalRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).proposal(input, options);
          }, { path: [28, 1] }),
          /**
           * getProposals queries all proposals based on given status.
           */
          getProposals: withMetadata(async function getProposals(input: DeepPartial<cosmos_gov_v1_query.QueryProposalsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).proposals(input, options);
          }, { path: [28, 2] }),
          /**
           * getVote queries voted information based on proposalID, voterAddr.
           */
          getVote: withMetadata(async function getVote(input: DeepPartial<cosmos_gov_v1_query.QueryVoteRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).vote(input, options);
          }, { path: [28, 3] }),
          /**
           * getVotes queries votes of a given proposal.
           */
          getVotes: withMetadata(async function getVotes(input: DeepPartial<cosmos_gov_v1_query.QueryVotesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).votes(input, options);
          }, { path: [28, 4] }),
          /**
           * getParams queries all parameters of the gov module.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_gov_v1_query.QueryParamsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).params(input, options);
          }, { path: [28, 5] }),
          /**
           * getDeposit queries single deposit information based on proposalID, depositAddr.
           */
          getDeposit: withMetadata(async function getDeposit(input: DeepPartial<cosmos_gov_v1_query.QueryDepositRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).deposit(input, options);
          }, { path: [28, 6] }),
          /**
           * getDeposits queries all deposits of a single proposal.
           */
          getDeposits: withMetadata(async function getDeposits(input: DeepPartial<cosmos_gov_v1_query.QueryDepositsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).deposits(input, options);
          }, { path: [28, 7] }),
          /**
           * getTallyResult queries the tally of a proposal vote.
           */
          getTallyResult: withMetadata(async function getTallyResult(input: DeepPartial<cosmos_gov_v1_query.QueryTallyResultRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(28);
            return getClient(service).tallyResult(input, options);
          }, { path: [28, 8] }),
          /**
           * submitProposal defines a method to create new proposal given the messages.
           */
          submitProposal: withMetadata(async function submitProposal(input: DeepSimplify<cosmos_gov_v1_tx.MsgSubmitProposal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).submitProposal(input, options);
          }, { path: [29, 0] }),
          /**
           * execLegacyContent defines a Msg to be in included in a MsgSubmitProposal
           * to execute a legacy content-based proposal.
           */
          execLegacyContent: withMetadata(async function execLegacyContent(input: DeepSimplify<cosmos_gov_v1_tx.MsgExecLegacyContent>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).execLegacyContent(input, options);
          }, { path: [29, 1] }),
          /**
           * vote defines a method to add a vote on a specific proposal.
           */
          vote: withMetadata(async function vote(input: DeepSimplify<cosmos_gov_v1_tx.MsgVote>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).vote(input, options);
          }, { path: [29, 2] }),
          /**
           * voteWeighted defines a method to add a weighted vote on a specific proposal.
           */
          voteWeighted: withMetadata(async function voteWeighted(input: DeepSimplify<cosmos_gov_v1_tx.MsgVoteWeighted>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).voteWeighted(input, options);
          }, { path: [29, 3] }),
          /**
           * deposit defines a method to add deposit on a specific proposal.
           */
          deposit: withMetadata(async function deposit(input: DeepSimplify<cosmos_gov_v1_tx.MsgDeposit>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).deposit(input, options);
          }, { path: [29, 4] }),
          /**
           * updateParams defines a governance operation for updating the x/gov module
           * parameters. The authority is defined in the keeper.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_gov_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [29, 5] }),
          /**
           * cancelProposal defines a method to cancel governance proposal
           */
          cancelProposal: withMetadata(async function cancelProposal(input: DeepSimplify<cosmos_gov_v1_tx.MsgCancelProposal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(29);
            return getMsgClient(service).cancelProposal(input, options);
          }, { path: [29, 6] })
        },
        v1beta1: {
          /**
           * getProposal queries proposal details based on ProposalID.
           */
          getProposal: withMetadata(async function getProposal(input: DeepPartial<cosmos_gov_v1beta1_query.QueryProposalRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).proposal(input, options);
          }, { path: [30, 0] }),
          /**
           * getProposals queries all proposals based on given status.
           */
          getProposals: withMetadata(async function getProposals(input: DeepPartial<cosmos_gov_v1beta1_query.QueryProposalsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).proposals(input, options);
          }, { path: [30, 1] }),
          /**
           * getVote queries voted information based on proposalID, voterAddr.
           */
          getVote: withMetadata(async function getVote(input: DeepPartial<cosmos_gov_v1beta1_query.QueryVoteRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).vote(input, options);
          }, { path: [30, 2] }),
          /**
           * getVotes queries votes of a given proposal.
           */
          getVotes: withMetadata(async function getVotes(input: DeepPartial<cosmos_gov_v1beta1_query.QueryVotesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).votes(input, options);
          }, { path: [30, 3] }),
          /**
           * getParams queries all parameters of the gov module.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_gov_v1beta1_query.QueryParamsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).params(input, options);
          }, { path: [30, 4] }),
          /**
           * getDeposit queries single deposit information based on proposalID, depositor address.
           */
          getDeposit: withMetadata(async function getDeposit(input: DeepPartial<cosmos_gov_v1beta1_query.QueryDepositRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).deposit(input, options);
          }, { path: [30, 5] }),
          /**
           * getDeposits queries all deposits of a single proposal.
           */
          getDeposits: withMetadata(async function getDeposits(input: DeepPartial<cosmos_gov_v1beta1_query.QueryDepositsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).deposits(input, options);
          }, { path: [30, 6] }),
          /**
           * getTallyResult queries the tally of a proposal vote.
           */
          getTallyResult: withMetadata(async function getTallyResult(input: DeepPartial<cosmos_gov_v1beta1_query.QueryTallyResultRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(30);
            return getClient(service).tallyResult(input, options);
          }, { path: [30, 7] }),
          /**
           * submitProposal defines a method to create new proposal given a content.
           */
          submitProposal: withMetadata(async function submitProposal(input: DeepSimplify<cosmos_gov_v1beta1_tx.MsgSubmitProposal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).submitProposal(input, options);
          }, { path: [31, 0] }),
          /**
           * vote defines a method to add a vote on a specific proposal.
           */
          vote: withMetadata(async function vote(input: DeepSimplify<cosmos_gov_v1beta1_tx.MsgVote>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).vote(input, options);
          }, { path: [31, 1] }),
          /**
           * voteWeighted defines a method to add a weighted vote on a specific proposal.
           */
          voteWeighted: withMetadata(async function voteWeighted(input: DeepSimplify<cosmos_gov_v1beta1_tx.MsgVoteWeighted>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).voteWeighted(input, options);
          }, { path: [31, 2] }),
          /**
           * deposit defines a method to add deposit on a specific proposal.
           */
          deposit: withMetadata(async function deposit(input: DeepSimplify<cosmos_gov_v1beta1_tx.MsgDeposit>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(31);
            return getMsgClient(service).deposit(input, options);
          }, { path: [31, 3] })
        }
      },
      group: {
        v1: {
          /**
           * getGroupInfo queries group info based on group id.
           */
          getGroupInfo: withMetadata(async function getGroupInfo(input: DeepPartial<cosmos_group_v1_query.QueryGroupInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupInfo(input, options);
          }, { path: [32, 0] }),
          /**
           * getGroupPolicyInfo queries group policy info based on account address of group policy.
           */
          getGroupPolicyInfo: withMetadata(async function getGroupPolicyInfo(input: DeepPartial<cosmos_group_v1_query.QueryGroupPolicyInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupPolicyInfo(input, options);
          }, { path: [32, 1] }),
          /**
           * getGroupMembers queries members of a group by group id.
           */
          getGroupMembers: withMetadata(async function getGroupMembers(input: DeepPartial<cosmos_group_v1_query.QueryGroupMembersRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupMembers(input, options);
          }, { path: [32, 2] }),
          /**
           * getGroupsByAdmin queries groups by admin address.
           */
          getGroupsByAdmin: withMetadata(async function getGroupsByAdmin(input: DeepPartial<cosmos_group_v1_query.QueryGroupsByAdminRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupsByAdmin(input, options);
          }, { path: [32, 3] }),
          /**
           * getGroupPoliciesByGroup queries group policies by group id.
           */
          getGroupPoliciesByGroup: withMetadata(async function getGroupPoliciesByGroup(input: DeepPartial<cosmos_group_v1_query.QueryGroupPoliciesByGroupRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupPoliciesByGroup(input, options);
          }, { path: [32, 4] }),
          /**
           * getGroupPoliciesByAdmin queries group policies by admin address.
           */
          getGroupPoliciesByAdmin: withMetadata(async function getGroupPoliciesByAdmin(input: DeepPartial<cosmos_group_v1_query.QueryGroupPoliciesByAdminRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupPoliciesByAdmin(input, options);
          }, { path: [32, 5] }),
          /**
           * getProposal queries a proposal based on proposal id.
           */
          getProposal: withMetadata(async function getProposal(input: DeepPartial<cosmos_group_v1_query.QueryProposalRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).proposal(input, options);
          }, { path: [32, 6] }),
          /**
           * getProposalsByGroupPolicy queries proposals based on account address of group policy.
           */
          getProposalsByGroupPolicy: withMetadata(async function getProposalsByGroupPolicy(input: DeepPartial<cosmos_group_v1_query.QueryProposalsByGroupPolicyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).proposalsByGroupPolicy(input, options);
          }, { path: [32, 7] }),
          /**
           * getVoteByProposalVoter queries a vote by proposal id and voter.
           */
          getVoteByProposalVoter: withMetadata(async function getVoteByProposalVoter(input: DeepPartial<cosmos_group_v1_query.QueryVoteByProposalVoterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).voteByProposalVoter(input, options);
          }, { path: [32, 8] }),
          /**
           * getVotesByProposal queries a vote by proposal id.
           */
          getVotesByProposal: withMetadata(async function getVotesByProposal(input: DeepPartial<cosmos_group_v1_query.QueryVotesByProposalRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).votesByProposal(input, options);
          }, { path: [32, 9] }),
          /**
           * getVotesByVoter queries a vote by voter.
           */
          getVotesByVoter: withMetadata(async function getVotesByVoter(input: DeepPartial<cosmos_group_v1_query.QueryVotesByVoterRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).votesByVoter(input, options);
          }, { path: [32, 10] }),
          /**
           * getGroupsByMember queries groups by member address.
           */
          getGroupsByMember: withMetadata(async function getGroupsByMember(input: DeepPartial<cosmos_group_v1_query.QueryGroupsByMemberRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groupsByMember(input, options);
          }, { path: [32, 11] }),
          /**
           * getTallyResult returns the tally result of a proposal. If the proposal is
           * still in voting period, then this query computes the current tally state,
           * which might not be final. On the other hand, if the proposal is final,
           * then it simply returns the `final_tally_result` state stored in the
           * proposal itself.
           */
          getTallyResult: withMetadata(async function getTallyResult(input: DeepPartial<cosmos_group_v1_query.QueryTallyResultRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).tallyResult(input, options);
          }, { path: [32, 12] }),
          /**
           * getGroups queries all groups in state.
           */
          getGroups: withMetadata(async function getGroups(input: DeepPartial<cosmos_group_v1_query.QueryGroupsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(32);
            return getClient(service).groups(input, options);
          }, { path: [32, 13] }),
          /**
           * createGroup creates a new group with an admin account address, a list of members and some optional metadata.
           */
          createGroup: withMetadata(async function createGroup(input: DeepSimplify<cosmos_group_v1_tx.MsgCreateGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).createGroup(input, options);
          }, { path: [33, 0] }),
          /**
           * updateGroupMembers updates the group members with given group id and admin address.
           */
          updateGroupMembers: withMetadata(async function updateGroupMembers(input: DeepSimplify<cosmos_group_v1_tx.MsgUpdateGroupMembers>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateGroupMembers(input, options);
          }, { path: [33, 1] }),
          /**
           * updateGroupAdmin updates the group admin with given group id and previous admin address.
           */
          updateGroupAdmin: withMetadata(async function updateGroupAdmin(input: DeepSimplify<cosmos_group_v1_tx.MsgUpdateGroupAdmin>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateGroupAdmin(input, options);
          }, { path: [33, 2] }),
          /**
           * updateGroupMetadata updates the group metadata with given group id and admin address.
           */
          updateGroupMetadata: withMetadata(async function updateGroupMetadata(input: DeepSimplify<cosmos_group_v1_tx.MsgUpdateGroupMetadata>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateGroupMetadata(input, options);
          }, { path: [33, 3] }),
          /**
           * createGroupPolicy creates a new group policy using given DecisionPolicy.
           */
          createGroupPolicy: withMetadata(async function createGroupPolicy(input: DeepSimplify<cosmos_group_v1_tx.MsgCreateGroupPolicy>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).createGroupPolicy(input, options);
          }, { path: [33, 4] }),
          /**
           * createGroupWithPolicy creates a new group with policy.
           */
          createGroupWithPolicy: withMetadata(async function createGroupWithPolicy(input: DeepSimplify<cosmos_group_v1_tx.MsgCreateGroupWithPolicy>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).createGroupWithPolicy(input, options);
          }, { path: [33, 5] }),
          /**
           * updateGroupPolicyAdmin updates a group policy admin.
           */
          updateGroupPolicyAdmin: withMetadata(async function updateGroupPolicyAdmin(input: DeepSimplify<cosmos_group_v1_tx.MsgUpdateGroupPolicyAdmin>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateGroupPolicyAdmin(input, options);
          }, { path: [33, 6] }),
          /**
           * updateGroupPolicyDecisionPolicy allows a group policy's decision policy to be updated.
           */
          updateGroupPolicyDecisionPolicy: withMetadata(async function updateGroupPolicyDecisionPolicy(input: DeepSimplify<cosmos_group_v1_tx.MsgUpdateGroupPolicyDecisionPolicy>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateGroupPolicyDecisionPolicy(input, options);
          }, { path: [33, 7] }),
          /**
           * updateGroupPolicyMetadata updates a group policy metadata.
           */
          updateGroupPolicyMetadata: withMetadata(async function updateGroupPolicyMetadata(input: DeepSimplify<cosmos_group_v1_tx.MsgUpdateGroupPolicyMetadata>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).updateGroupPolicyMetadata(input, options);
          }, { path: [33, 8] }),
          /**
           * submitProposal submits a new proposal.
           */
          submitProposal: withMetadata(async function submitProposal(input: DeepSimplify<cosmos_group_v1_tx.MsgSubmitProposal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).submitProposal(input, options);
          }, { path: [33, 9] }),
          /**
           * withdrawProposal withdraws a proposal.
           */
          withdrawProposal: withMetadata(async function withdrawProposal(input: DeepSimplify<cosmos_group_v1_tx.MsgWithdrawProposal>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).withdrawProposal(input, options);
          }, { path: [33, 10] }),
          /**
           * vote allows a voter to vote on a proposal.
           */
          vote: withMetadata(async function vote(input: DeepSimplify<cosmos_group_v1_tx.MsgVote>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).vote(input, options);
          }, { path: [33, 11] }),
          /**
           * exec executes a proposal.
           */
          exec: withMetadata(async function exec(input: DeepSimplify<cosmos_group_v1_tx.MsgExec>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).exec(input, options);
          }, { path: [33, 12] }),
          /**
           * leaveGroup allows a group member to leave the group.
           */
          leaveGroup: withMetadata(async function leaveGroup(input: DeepSimplify<cosmos_group_v1_tx.MsgLeaveGroup>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(33);
            return getMsgClient(service).leaveGroup(input, options);
          }, { path: [33, 13] })
        }
      },
      mint: {
        v1beta1: {
          /**
           * getParams returns the total set of minting parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_mint_v1beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).params(input, options);
          }, { path: [34, 0] }),
          /**
           * getInflation returns the current minting inflation value.
           */
          getInflation: withMetadata(async function getInflation(input: DeepPartial<cosmos_mint_v1beta1_query.QueryInflationRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).inflation(input, options);
          }, { path: [34, 1] }),
          /**
           * getAnnualProvisions current minting annual provisions value.
           */
          getAnnualProvisions: withMetadata(async function getAnnualProvisions(input: DeepPartial<cosmos_mint_v1beta1_query.QueryAnnualProvisionsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(34);
            return getClient(service).annualProvisions(input, options);
          }, { path: [34, 2] }),
          /**
           * updateParams defines a governance operation for updating the x/mint module
           * parameters. The authority is defaults to the x/gov module account.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_mint_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(35);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [35, 0] })
        }
      },
      nft: {
        v1beta1: {
          /**
           * getBalance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
           */
          getBalance: withMetadata(async function getBalance(input: DeepPartial<cosmos_nft_v1beta1_query.QueryBalanceRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).balance(input, options);
          }, { path: [36, 0] }),
          /**
           * getOwner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721
           */
          getOwner: withMetadata(async function getOwner(input: DeepPartial<cosmos_nft_v1beta1_query.QueryOwnerRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).owner(input, options);
          }, { path: [36, 1] }),
          /**
           * getSupply queries the number of NFTs from the given class, same as totalSupply of ERC721.
           */
          getSupply: withMetadata(async function getSupply(input: DeepPartial<cosmos_nft_v1beta1_query.QuerySupplyRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).supply(input, options);
          }, { path: [36, 2] }),
          /**
           * getNFTs queries all getNFTs of a given class or owner,choose at least one of the two, similar to tokenByIndex in
           * ERC721Enumerable
           */
          getNFTs: withMetadata(async function getNFTs(input: DeepPartial<cosmos_nft_v1beta1_query.QueryNFTsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).nFTs(input, options);
          }, { path: [36, 3] }),
          /**
           * getNFT queries an getNFT based on its class and id.
           */
          getNFT: withMetadata(async function getNFT(input: DeepPartial<cosmos_nft_v1beta1_query.QueryNFTRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).nFT(input, options);
          }, { path: [36, 4] }),
          /**
           * getClass queries an NFT class based on its id
           */
          getClass: withMetadata(async function getClass(input: DeepPartial<cosmos_nft_v1beta1_query.QueryClassRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).class(input, options);
          }, { path: [36, 5] }),
          /**
           * getClasses queries all NFT classes
           */
          getClasses: withMetadata(async function getClasses(input: DeepPartial<cosmos_nft_v1beta1_query.QueryClassesRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(36);
            return getClient(service).classes(input, options);
          }, { path: [36, 6] }),
          /**
           * send defines a method to send a nft from one account to another account.
           */
          send: withMetadata(async function send(input: DeepSimplify<cosmos_nft_v1beta1_tx.MsgSend>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(37);
            return getMsgClient(service).send(input, options);
          }, { path: [37, 0] })
        }
      },
      params: {
        v1beta1: {
          /**
           * getParams queries a specific parameter of a module, given its subspace and
           * key.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_params_v1beta1_query.QueryParamsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).params(input, options);
          }, { path: [38, 0] }),
          /**
           * getSubspaces queries for all registered subspaces and all keys for a subspace.
           */
          getSubspaces: withMetadata(async function getSubspaces(input: DeepPartial<cosmos_params_v1beta1_query.QuerySubspacesRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(38);
            return getClient(service).subspaces(input, options);
          }, { path: [38, 1] })
        }
      },
      protocolpool: {
        v1: {
          /**
           * getCommunityPool queries the community pool coins.
           */
          getCommunityPool: withMetadata(async function getCommunityPool(input: DeepPartial<cosmos_protocolpool_v1_query.QueryCommunityPoolRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getClient(service).communityPool(input, options);
          }, { path: [39, 0] }),
          /**
           * getContinuousFund queries a continuous fund by the recipient is is associated with.
           */
          getContinuousFund: withMetadata(async function getContinuousFund(input: DeepPartial<cosmos_protocolpool_v1_query.QueryContinuousFundRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getClient(service).continuousFund(input, options);
          }, { path: [39, 1] }),
          /**
           * getContinuousFunds queries all continuous funds in the store.
           */
          getContinuousFunds: withMetadata(async function getContinuousFunds(input: DeepPartial<cosmos_protocolpool_v1_query.QueryContinuousFundsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getClient(service).continuousFunds(input, options);
          }, { path: [39, 2] }),
          /**
           * getParams returns the total set of x/protocolpool parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_protocolpool_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(39);
            return getClient(service).params(input, options);
          }, { path: [39, 3] }),
          /**
           * fundCommunityPool defines a method to allow an account to directly
           * fund the community pool.
           */
          fundCommunityPool: withMetadata(async function fundCommunityPool(input: DeepSimplify<cosmos_protocolpool_v1_tx.MsgFundCommunityPool>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getMsgClient(service).fundCommunityPool(input, options);
          }, { path: [40, 0] }),
          /**
           * communityPoolSpend defines a governance operation for sending tokens from
           * the community pool in the x/protocolpool module to another account, which
           * could be the governance module itself. The authority is defined in the
           * keeper.
           */
          communityPoolSpend: withMetadata(async function communityPoolSpend(input: DeepSimplify<cosmos_protocolpool_v1_tx.MsgCommunityPoolSpend>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getMsgClient(service).communityPoolSpend(input, options);
          }, { path: [40, 1] }),
          /**
           * createContinuousFund defines a method to distribute a percentage of funds to an address continuously.
           * This ContinuousFund can be indefinite or run until a given expiry time.
           * Funds come from validator block rewards from x/distribution, but may also come from
           * any user who funds the ProtocolPoolEscrow module account directly through x/bank.
           */
          createContinuousFund: withMetadata(async function createContinuousFund(input: DeepSimplify<cosmos_protocolpool_v1_tx.MsgCreateContinuousFund>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getMsgClient(service).createContinuousFund(input, options);
          }, { path: [40, 2] }),
          /**
           * cancelContinuousFund defines a method for cancelling continuous fund.
           */
          cancelContinuousFund: withMetadata(async function cancelContinuousFund(input: DeepSimplify<cosmos_protocolpool_v1_tx.MsgCancelContinuousFund>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getMsgClient(service).cancelContinuousFund(input, options);
          }, { path: [40, 3] }),
          /**
           * updateParams defines a governance operation for updating the x/protocolpool module parameters.
           * The authority is defined in the keeper.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_protocolpool_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(40);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [40, 4] })
        }
      },
      reflection: {
        v1: {
          /**
           * getFileDescriptors queries all the file descriptors in the app in order
           * to enable easier generation of dynamic clients.
           */
          getFileDescriptors: withMetadata(async function getFileDescriptors(input: DeepPartial<cosmos_reflection_v1_reflection.FileDescriptorsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(41);
            return getClient(service).fileDescriptors(input, options);
          }, { path: [41, 0] })
        }
      },
      slashing: {
        v1beta1: {
          /**
           * getParams queries the parameters of slashing module
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_slashing_v1beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).params(input, options);
          }, { path: [42, 0] }),
          /**
           * getSigningInfo queries the signing info of given cons address
           */
          getSigningInfo: withMetadata(async function getSigningInfo(input: DeepPartial<cosmos_slashing_v1beta1_query.QuerySigningInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).signingInfo(input, options);
          }, { path: [42, 1] }),
          /**
           * getSigningInfos queries signing info of all validators
           */
          getSigningInfos: withMetadata(async function getSigningInfos(input: DeepPartial<cosmos_slashing_v1beta1_query.QuerySigningInfosRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(42);
            return getClient(service).signingInfos(input, options);
          }, { path: [42, 2] }),
          /**
           * unjail defines a method for unjailing a jailed validator, thus returning
           * them into the bonded validator set, so they can begin receiving provisions
           * and rewards again.
           */
          unjail: withMetadata(async function unjail(input: DeepSimplify<cosmos_slashing_v1beta1_tx.MsgUnjail>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).unjail(input, options);
          }, { path: [43, 0] }),
          /**
           * updateParams defines a governance operation for updating the x/slashing module
           * parameters. The authority defaults to the x/gov module account.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_slashing_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(43);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [43, 1] })
        }
      },
      staking: {
        v1beta1: {
          /**
           * getValidators queries all validators that match the given status.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getValidators: withMetadata(async function getValidators(input: DeepPartial<cosmos_staking_v1beta1_query.QueryValidatorsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).validators(input, options);
          }, { path: [44, 0] }),
          /**
           * getValidator queries validator info for given validator address.
           */
          getValidator: withMetadata(async function getValidator(input: DeepPartial<cosmos_staking_v1beta1_query.QueryValidatorRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).validator(input, options);
          }, { path: [44, 1] }),
          /**
           * getValidatorDelegations queries delegate info for given validator.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getValidatorDelegations: withMetadata(async function getValidatorDelegations(input: DeepPartial<cosmos_staking_v1beta1_query.QueryValidatorDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).validatorDelegations(input, options);
          }, { path: [44, 2] }),
          /**
           * getValidatorUnbondingDelegations queries unbonding delegations of a validator.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getValidatorUnbondingDelegations: withMetadata(async function getValidatorUnbondingDelegations(input: DeepPartial<cosmos_staking_v1beta1_query.QueryValidatorUnbondingDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).validatorUnbondingDelegations(input, options);
          }, { path: [44, 3] }),
          /**
           * getDelegation queries delegate info for given validator delegator pair.
           */
          getDelegation: withMetadata(async function getDelegation(input: DeepPartial<cosmos_staking_v1beta1_query.QueryDelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).delegation(input, options);
          }, { path: [44, 4] }),
          /**
           * getUnbondingDelegation queries unbonding info for given validator delegator
           * pair.
           */
          getUnbondingDelegation: withMetadata(async function getUnbondingDelegation(input: DeepPartial<cosmos_staking_v1beta1_query.QueryUnbondingDelegationRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).unbondingDelegation(input, options);
          }, { path: [44, 5] }),
          /**
           * getDelegatorDelegations queries all delegations of a given delegator address.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getDelegatorDelegations: withMetadata(async function getDelegatorDelegations(input: DeepPartial<cosmos_staking_v1beta1_query.QueryDelegatorDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).delegatorDelegations(input, options);
          }, { path: [44, 6] }),
          /**
           * getDelegatorUnbondingDelegations queries all unbonding delegations of a given
           * delegator address.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getDelegatorUnbondingDelegations: withMetadata(async function getDelegatorUnbondingDelegations(input: DeepPartial<cosmos_staking_v1beta1_query.QueryDelegatorUnbondingDelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).delegatorUnbondingDelegations(input, options);
          }, { path: [44, 7] }),
          /**
           * getRedelegations queries redelegations of given address.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getRedelegations: withMetadata(async function getRedelegations(input: DeepPartial<cosmos_staking_v1beta1_query.QueryRedelegationsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).redelegations(input, options);
          }, { path: [44, 8] }),
          /**
           * getDelegatorValidators queries all validators info for given delegator
           * address.
           *
           * When called from another module, this query might consume a high amount of
           * gas if the pagination field is incorrectly set.
           */
          getDelegatorValidators: withMetadata(async function getDelegatorValidators(input: DeepPartial<cosmos_staking_v1beta1_query.QueryDelegatorValidatorsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).delegatorValidators(input, options);
          }, { path: [44, 9] }),
          /**
           * getDelegatorValidator queries validator info for given delegator validator
           * pair.
           */
          getDelegatorValidator: withMetadata(async function getDelegatorValidator(input: DeepPartial<cosmos_staking_v1beta1_query.QueryDelegatorValidatorRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).delegatorValidator(input, options);
          }, { path: [44, 10] }),
          /**
           * getHistoricalInfo queries the historical info for given height.
           */
          getHistoricalInfo: withMetadata(async function getHistoricalInfo(input: DeepPartial<cosmos_staking_v1beta1_query.QueryHistoricalInfoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).historicalInfo(input, options);
          }, { path: [44, 11] }),
          /**
           * getPool queries the pool info.
           */
          getPool: withMetadata(async function getPool(input: DeepPartial<cosmos_staking_v1beta1_query.QueryPoolRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).pool(input, options);
          }, { path: [44, 12] }),
          /**
           * Parameters queries the staking parameters.
           */
          getParams: withMetadata(async function getParams(input: DeepPartial<cosmos_staking_v1beta1_query.QueryParamsRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(44);
            return getClient(service).params(input, options);
          }, { path: [44, 13] }),
          /**
           * createValidator defines a method for creating a new validator.
           */
          createValidator: withMetadata(async function createValidator(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgCreateValidator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).createValidator(input, options);
          }, { path: [45, 0] }),
          /**
           * editValidator defines a method for editing an existing validator.
           */
          editValidator: withMetadata(async function editValidator(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgEditValidator>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).editValidator(input, options);
          }, { path: [45, 1] }),
          /**
           * delegate defines a method for performing a delegation of coins
           * from a delegator to a validator.
           */
          delegate: withMetadata(async function delegate(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgDelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).delegate(input, options);
          }, { path: [45, 2] }),
          /**
           * beginRedelegate defines a method for performing a redelegation
           * of coins from a delegator and source validator to a destination validator.
           */
          beginRedelegate: withMetadata(async function beginRedelegate(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgBeginRedelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).beginRedelegate(input, options);
          }, { path: [45, 3] }),
          /**
           * undelegate defines a method for performing an undelegation from a
           * delegate and a validator.
           */
          undelegate: withMetadata(async function undelegate(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgUndelegate>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).undelegate(input, options);
          }, { path: [45, 4] }),
          /**
           * cancelUnbondingDelegation defines a method for performing canceling the unbonding delegation
           * and delegate back to previous validator.
           */
          cancelUnbondingDelegation: withMetadata(async function cancelUnbondingDelegation(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgCancelUnbondingDelegation>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).cancelUnbondingDelegation(input, options);
          }, { path: [45, 5] }),
          /**
           * updateParams defines an operation for updating the x/staking module
           * parameters.
           */
          updateParams: withMetadata(async function updateParams(input: DeepSimplify<cosmos_staking_v1beta1_tx.MsgUpdateParams>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(45);
            return getMsgClient(service).updateParams(input, options);
          }, { path: [45, 6] })
        }
      },
      store: {
        streaming: {
          abci: {
            /**
             * getListenFinalizeBlock is the corresponding endpoint for ABCIListener.ListenEndBlock
             */
            getListenFinalizeBlock: withMetadata(async function getListenFinalizeBlock(input: DeepPartial<cosmos_store_streaming_abci_grpc.ListenFinalizeBlockRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(46);
              return getClient(service).listenFinalizeBlock(input, options);
            }, { path: [46, 0] }),
            /**
             * getListenCommit is the corresponding endpoint for ABCIListener.getListenCommit
             */
            getListenCommit: withMetadata(async function getListenCommit(input: DeepPartial<cosmos_store_streaming_abci_grpc.ListenCommitRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(46);
              return getClient(service).listenCommit(input, options);
            }, { path: [46, 1] })
          }
        }
      },
      tx: {
        v1beta1: {
          /**
           * getSimulate simulates executing a transaction for estimating gas usage.
           */
          getSimulate: withMetadata(async function getSimulate(input: DeepPartial<cosmos_tx_v1beta1_service.SimulateRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).simulate(input, options);
          }, { path: [47, 0] }),
          /**
           * getTx fetches a tx by hash.
           */
          getTx: withMetadata(async function getTx(input: DeepPartial<cosmos_tx_v1beta1_service.GetTxRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).getTx(input, options);
          }, { path: [47, 1] }),
          /**
           * getBroadcastTx broadcast transaction.
           */
          getBroadcastTx: withMetadata(async function getBroadcastTx(input: DeepPartial<cosmos_tx_v1beta1_service.BroadcastTxRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).broadcastTx(input, options);
          }, { path: [47, 2] }),
          /**
           * getTxsEvent fetches txs by event.
           */
          getTxsEvent: withMetadata(async function getTxsEvent(input: DeepPartial<cosmos_tx_v1beta1_service.GetTxsEventRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).getTxsEvent(input, options);
          }, { path: [47, 3] }),
          /**
           * getBlockWithTxs fetches a block with decoded txs.
           */
          getBlockWithTxs: withMetadata(async function getBlockWithTxs(input: DeepPartial<cosmos_tx_v1beta1_service.GetBlockWithTxsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).getBlockWithTxs(input, options);
          }, { path: [47, 4] }),
          /**
           * getTxDecode decodes the transaction.
           */
          getTxDecode: withMetadata(async function getTxDecode(input: DeepPartial<cosmos_tx_v1beta1_service.TxDecodeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).txDecode(input, options);
          }, { path: [47, 5] }),
          /**
           * getTxEncode encodes the transaction.
           */
          getTxEncode: withMetadata(async function getTxEncode(input: DeepPartial<cosmos_tx_v1beta1_service.TxEncodeRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).txEncode(input, options);
          }, { path: [47, 6] }),
          /**
           * getTxEncodeAmino encodes an Amino transaction from JSON to encoded bytes.
           */
          getTxEncodeAmino: withMetadata(async function getTxEncodeAmino(input: DeepPartial<cosmos_tx_v1beta1_service.TxEncodeAminoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).txEncodeAmino(input, options);
          }, { path: [47, 7] }),
          /**
           * getTxDecodeAmino decodes an Amino transaction from encoded bytes to JSON.
           */
          getTxDecodeAmino: withMetadata(async function getTxDecodeAmino(input: DeepPartial<cosmos_tx_v1beta1_service.TxDecodeAminoRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(47);
            return getClient(service).txDecodeAmino(input, options);
          }, { path: [47, 8] })
        }
      },
      upgrade: {
        v1beta1: {
          /**
           * getCurrentPlan queries the current upgrade plan.
           */
          getCurrentPlan: withMetadata(async function getCurrentPlan(input: DeepPartial<cosmos_upgrade_v1beta1_query.QueryCurrentPlanRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).currentPlan(input, options);
          }, { path: [48, 0] }),
          /**
           * getAppliedPlan queries a previously applied upgrade plan by its name.
           */
          getAppliedPlan: withMetadata(async function getAppliedPlan(input: DeepPartial<cosmos_upgrade_v1beta1_query.QueryAppliedPlanRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).appliedPlan(input, options);
          }, { path: [48, 1] }),
          /**
           * getUpgradedConsensusState queries the consensus state that will serve
           * as a trusted kernel for the next version of this chain. It will only be
           * stored at the last height of this chain.
           * getUpgradedConsensusState RPC not supported with legacy querier
           * This rpc is deprecated now that IBC has its own replacement
           * (https://github.com/cosmos/ibc-go/blob/2c880a22e9f9cc75f62b527ca94aa75ce1106001/proto/ibc/core/client/v1/query.proto#L54)
           * @deprecated
           */
          getUpgradedConsensusState: withMetadata(async function getUpgradedConsensusState(input: DeepPartial<cosmos_upgrade_v1beta1_query.QueryUpgradedConsensusStateRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).upgradedConsensusState(input, options);
          }, { path: [48, 2] }),
          /**
           * getModuleVersions queries the list of module versions from state.
           */
          getModuleVersions: withMetadata(async function getModuleVersions(input: DeepPartial<cosmos_upgrade_v1beta1_query.QueryModuleVersionsRequest>, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).moduleVersions(input, options);
          }, { path: [48, 3] }),
          /**
           * Returns the account with authority to conduct upgrades
           */
          getAuthority: withMetadata(async function getAuthority(input: DeepPartial<cosmos_upgrade_v1beta1_query.QueryAuthorityRequest> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(48);
            return getClient(service).authority(input, options);
          }, { path: [48, 4] }),
          /**
           * softwareUpgrade is a governance operation for initiating a software upgrade.
           */
          softwareUpgrade: withMetadata(async function softwareUpgrade(input: DeepSimplify<cosmos_upgrade_v1beta1_tx.MsgSoftwareUpgrade>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).softwareUpgrade(input, options);
          }, { path: [49, 0] }),
          /**
           * cancelUpgrade is a governance operation for cancelling a previously
           * approved software upgrade.
           */
          cancelUpgrade: withMetadata(async function cancelUpgrade(input: DeepSimplify<cosmos_upgrade_v1beta1_tx.MsgCancelUpgrade>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(49);
            return getMsgClient(service).cancelUpgrade(input, options);
          }, { path: [49, 1] })
        }
      },
      vesting: {
        v1beta1: {
          /**
           * createVestingAccount defines a method that enables creating a vesting
           * account.
           */
          createVestingAccount: withMetadata(async function createVestingAccount(input: DeepSimplify<cosmos_vesting_v1beta1_tx.MsgCreateVestingAccount>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getMsgClient(service).createVestingAccount(input, options);
          }, { path: [50, 0] }),
          /**
           * createPermanentLockedAccount defines a method that enables creating a permanent
           * locked account.
           */
          createPermanentLockedAccount: withMetadata(async function createPermanentLockedAccount(input: DeepSimplify<cosmos_vesting_v1beta1_tx.MsgCreatePermanentLockedAccount>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getMsgClient(service).createPermanentLockedAccount(input, options);
          }, { path: [50, 1] }),
          /**
           * createPeriodicVestingAccount defines a method that enables creating a
           * periodic vesting account.
           */
          createPeriodicVestingAccount: withMetadata(async function createPeriodicVestingAccount(input: DeepSimplify<cosmos_vesting_v1beta1_tx.MsgCreatePeriodicVestingAccount>, options?: TxCallOptions) {
            const service = await serviceLoader.loadAt(50);
            return getMsgClient(service).createPeriodicVestingAccount(input, options);
          }, { path: [50, 2] })
        }
      }
    },
    tendermint: {
      abci: {
        getEcho: withMetadata(async function getEcho(input: DeepPartial<tendermint_abci_types.RequestEcho>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).echo(input, options);
        }, { path: [8, 0] }),
        getFlush: withMetadata(async function getFlush(input: DeepPartial<tendermint_abci_types.RequestFlush> = {}, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).flush(input, options);
        }, { path: [8, 1] }),
        getInfo: withMetadata(async function getInfo(input: DeepPartial<tendermint_abci_types.RequestInfo>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).info(input, options);
        }, { path: [8, 2] }),
        getCheckTx: withMetadata(async function getCheckTx(input: DeepPartial<tendermint_abci_types.RequestCheckTx>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).checkTx(input, options);
        }, { path: [8, 3] }),
        getQuery: withMetadata(async function getQuery(input: DeepPartial<tendermint_abci_types.RequestQuery>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).query(input, options);
        }, { path: [8, 4] }),
        getCommit: withMetadata(async function getCommit(input: DeepPartial<tendermint_abci_types.RequestCommit> = {}, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).commit(input, options);
        }, { path: [8, 5] }),
        getInitChain: withMetadata(async function getInitChain(input: DeepPartial<tendermint_abci_types.RequestInitChain>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).initChain(input, options);
        }, { path: [8, 6] }),
        getListSnapshots: withMetadata(async function getListSnapshots(input: DeepPartial<tendermint_abci_types.RequestListSnapshots> = {}, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).listSnapshots(input, options);
        }, { path: [8, 7] }),
        getOfferSnapshot: withMetadata(async function getOfferSnapshot(input: DeepPartial<tendermint_abci_types.RequestOfferSnapshot>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).offerSnapshot(input, options);
        }, { path: [8, 8] }),
        getLoadSnapshotChunk: withMetadata(async function getLoadSnapshotChunk(input: DeepPartial<tendermint_abci_types.RequestLoadSnapshotChunk>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).loadSnapshotChunk(input, options);
        }, { path: [8, 9] }),
        getApplySnapshotChunk: withMetadata(async function getApplySnapshotChunk(input: DeepPartial<tendermint_abci_types.RequestApplySnapshotChunk>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).applySnapshotChunk(input, options);
        }, { path: [8, 10] }),
        getPrepareProposal: withMetadata(async function getPrepareProposal(input: DeepPartial<tendermint_abci_types.RequestPrepareProposal>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).prepareProposal(input, options);
        }, { path: [8, 11] }),
        getProcessProposal: withMetadata(async function getProcessProposal(input: DeepPartial<tendermint_abci_types.RequestProcessProposal>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).processProposal(input, options);
        }, { path: [8, 12] }),
        getExtendVote: withMetadata(async function getExtendVote(input: DeepPartial<tendermint_abci_types.RequestExtendVote>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).extendVote(input, options);
        }, { path: [8, 13] }),
        getVerifyVoteExtension: withMetadata(async function getVerifyVoteExtension(input: DeepPartial<tendermint_abci_types.RequestVerifyVoteExtension>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).verifyVoteExtension(input, options);
        }, { path: [8, 14] }),
        getFinalizeBlock: withMetadata(async function getFinalizeBlock(input: DeepPartial<tendermint_abci_types.RequestFinalizeBlock>, options?: CallOptions) {
          const service = await serviceLoader.loadAt(8);
          return getClient(service).finalizeBlock(input, options);
        }, { path: [8, 15] })
      }
    }
  };
}
