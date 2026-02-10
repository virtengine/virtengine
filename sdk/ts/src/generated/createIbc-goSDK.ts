import { createServiceLoader } from "../sdk/client/createServiceLoader.ts";
import { SDKOptions } from "../sdk/types.ts";

import type * as ibc_applications_interchain_accounts_controller_v1_query from "./protos/ibc/applications/interchain_accounts/controller/v1/query.ts";
import type * as ibc_applications_interchain_accounts_controller_v1_tx from "./protos/ibc/applications/interchain_accounts/controller/v1/tx.ts";
import type * as ibc_applications_interchain_accounts_host_v1_query from "./protos/ibc/applications/interchain_accounts/host/v1/query.ts";
import type * as ibc_applications_interchain_accounts_host_v1_tx from "./protos/ibc/applications/interchain_accounts/host/v1/tx.ts";
import type * as ibc_applications_transfer_v1_query from "./protos/ibc/applications/transfer/v1/query.ts";
import type * as ibc_applications_transfer_v1_tx from "./protos/ibc/applications/transfer/v1/tx.ts";
import type * as ibc_core_channel_v1_query from "./protos/ibc/core/channel/v1/query.ts";
import type * as ibc_core_channel_v1_tx from "./protos/ibc/core/channel/v1/tx.ts";
import type * as ibc_core_channel_v2_query from "./protos/ibc/core/channel/v2/query.ts";
import type * as ibc_core_channel_v2_tx from "./protos/ibc/core/channel/v2/tx.ts";
import type * as ibc_core_client_v1_query from "./protos/ibc/core/client/v1/query.ts";
import type * as ibc_core_client_v1_tx from "./protos/ibc/core/client/v1/tx.ts";
import type * as ibc_core_client_v2_query from "./protos/ibc/core/client/v2/query.ts";
import type * as ibc_core_client_v2_tx from "./protos/ibc/core/client/v2/tx.ts";
import type * as ibc_core_connection_v1_query from "./protos/ibc/core/connection/v1/query.ts";
import type * as ibc_core_connection_v1_tx from "./protos/ibc/core/connection/v1/tx.ts";
import type * as ibc_lightclients_wasm_v1_query from "./protos/ibc/lightclients/wasm/v1/query.ts";
import type * as ibc_lightclients_wasm_v1_tx from "./protos/ibc/lightclients/wasm/v1/tx.ts";
import { createClientFactory } from "../sdk/client/createClientFactory.ts";
import type { Transport, CallOptions, TxCallOptions } from "../sdk/transport/types.ts";
import { withMetadata } from "../sdk/client/sdkMetadata.ts";
import type { DeepPartial, DeepSimplify } from "../encoding/typeEncodingHelpers.ts";


export const serviceLoader= createServiceLoader([
  () => import("./protos/ibc/applications/interchain_accounts/controller/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/applications/interchain_accounts/controller/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/applications/interchain_accounts/host/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/applications/interchain_accounts/host/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/applications/transfer/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/applications/transfer/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/core/channel/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/core/channel/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/core/channel/v2/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/core/channel/v2/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/core/client/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/core/client/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/core/client/v2/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/core/client/v2/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/core/connection/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/core/connection/v1/tx_virtengine.ts").then(m => m.Msg),
  () => import("./protos/ibc/lightclients/wasm/v1/query_virtengine.ts").then(m => m.Query),
  () => import("./protos/ibc/lightclients/wasm/v1/tx_virtengine.ts").then(m => m.Msg)
] as const);
export function createSDK(queryTransport: Transport, txTransport: Transport, options?: SDKOptions) {
  const getClient = createClientFactory<CallOptions>(queryTransport, options?.clientOptions);
  const getMsgClient = createClientFactory<TxCallOptions>(txTransport, options?.clientOptions);
  return {
    ibc: {
      applications: {
        interchain_accounts: {
          controller: {
            v1: {
              /**
               * getInterchainAccount returns the interchain account address for a given owner address on a given connection
               */
              getInterchainAccount: withMetadata(async function getInterchainAccount(input: DeepPartial<ibc_applications_interchain_accounts_controller_v1_query.QueryInterchainAccountRequest>, options?: CallOptions) {
                const service = await serviceLoader.loadAt(0);
                return getClient(service).interchainAccount(input, options);
              }, { path: [0, 0] }),
              /**
               * getParams queries all parameters of the ICA controller submodule.
               */
              getParams: withMetadata(async function getParams(input: DeepPartial<ibc_applications_interchain_accounts_controller_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
                const service = await serviceLoader.loadAt(0);
                return getClient(service).params(input, options);
              }, { path: [0, 1] }),
              /**
               * registerInterchainAccount defines a rpc handler for MsgRegisterInterchainAccount.
               */
              registerInterchainAccount: withMetadata(async function registerInterchainAccount(input: DeepSimplify<ibc_applications_interchain_accounts_controller_v1_tx.MsgRegisterInterchainAccount>, options?: TxCallOptions) {
                const service = await serviceLoader.loadAt(1);
                return getMsgClient(service).registerInterchainAccount(input, options);
              }, { path: [1, 0] }),
              /**
               * sendTx defines a rpc handler for MsgSendTx.
               */
              sendTx: withMetadata(async function sendTx(input: DeepSimplify<ibc_applications_interchain_accounts_controller_v1_tx.MsgSendTx>, options?: TxCallOptions) {
                const service = await serviceLoader.loadAt(1);
                return getMsgClient(service).sendTx(input, options);
              }, { path: [1, 1] }),
              /**
               * updateParams defines a rpc handler for MsgUpdateParams.
               */
              updateParams: withMetadata(async function updateParams(input: DeepSimplify<ibc_applications_interchain_accounts_controller_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
                const service = await serviceLoader.loadAt(1);
                return getMsgClient(service).updateParams(input, options);
              }, { path: [1, 2] })
            }
          },
          host: {
            v1: {
              /**
               * getParams queries all parameters of the ICA host submodule.
               */
              getParams: withMetadata(async function getParams(input: DeepPartial<ibc_applications_interchain_accounts_host_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
                const service = await serviceLoader.loadAt(2);
                return getClient(service).params(input, options);
              }, { path: [2, 0] }),
              /**
               * updateParams defines a rpc handler for MsgUpdateParams.
               */
              updateParams: withMetadata(async function updateParams(input: DeepSimplify<ibc_applications_interchain_accounts_host_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
                const service = await serviceLoader.loadAt(3);
                return getMsgClient(service).updateParams(input, options);
              }, { path: [3, 0] }),
              /**
               * moduleQuerySafe defines a rpc handler for MsgModuleQuerySafe.
               */
              moduleQuerySafe: withMetadata(async function moduleQuerySafe(input: DeepSimplify<ibc_applications_interchain_accounts_host_v1_tx.MsgModuleQuerySafe>, options?: TxCallOptions) {
                const service = await serviceLoader.loadAt(3);
                return getMsgClient(service).moduleQuerySafe(input, options);
              }, { path: [3, 1] })
            }
          }
        },
        transfer: {
          v1: {
            /**
             * getParams queries all parameters of the ibc-transfer module.
             */
            getParams: withMetadata(async function getParams(input: DeepPartial<ibc_applications_transfer_v1_query.QueryParamsRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(4);
              return getClient(service).params(input, options);
            }, { path: [4, 0] }),
            /**
             * getDenoms queries all denominations
             */
            getDenoms: withMetadata(async function getDenoms(input: DeepPartial<ibc_applications_transfer_v1_query.QueryDenomsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(4);
              return getClient(service).denoms(input, options);
            }, { path: [4, 1] }),
            /**
             * getDenom queries a denomination
             */
            getDenom: withMetadata(async function getDenom(input: DeepPartial<ibc_applications_transfer_v1_query.QueryDenomRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(4);
              return getClient(service).denom(input, options);
            }, { path: [4, 2] }),
            /**
             * getDenomHash queries a denomination hash information.
             */
            getDenomHash: withMetadata(async function getDenomHash(input: DeepPartial<ibc_applications_transfer_v1_query.QueryDenomHashRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(4);
              return getClient(service).denomHash(input, options);
            }, { path: [4, 3] }),
            /**
             * getEscrowAddress returns the escrow address for a particular port and channel id.
             */
            getEscrowAddress: withMetadata(async function getEscrowAddress(input: DeepPartial<ibc_applications_transfer_v1_query.QueryEscrowAddressRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(4);
              return getClient(service).escrowAddress(input, options);
            }, { path: [4, 4] }),
            /**
             * getTotalEscrowForDenom returns the total amount of tokens in escrow based on the denom.
             */
            getTotalEscrowForDenom: withMetadata(async function getTotalEscrowForDenom(input: DeepPartial<ibc_applications_transfer_v1_query.QueryTotalEscrowForDenomRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(4);
              return getClient(service).totalEscrowForDenom(input, options);
            }, { path: [4, 5] }),
            /**
             * transfer defines a rpc handler method for MsgTransfer.
             */
            transfer: withMetadata(async function transfer(input: DeepSimplify<ibc_applications_transfer_v1_tx.MsgTransfer>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(5);
              return getMsgClient(service).transfer(input, options);
            }, { path: [5, 0] }),
            /**
             * updateParams defines a rpc handler for MsgUpdateParams.
             */
            updateParams: withMetadata(async function updateParams(input: DeepSimplify<ibc_applications_transfer_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(5);
              return getMsgClient(service).updateParams(input, options);
            }, { path: [5, 1] })
          }
        }
      },
      core: {
        channel: {
          v1: {
            /**
             * getChannel queries an IBC getChannel.
             */
            getChannel: withMetadata(async function getChannel(input: DeepPartial<ibc_core_channel_v1_query.QueryChannelRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).channel(input, options);
            }, { path: [6, 0] }),
            /**
             * getChannels queries all the IBC channels of a chain.
             */
            getChannels: withMetadata(async function getChannels(input: DeepPartial<ibc_core_channel_v1_query.QueryChannelsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).channels(input, options);
            }, { path: [6, 1] }),
            /**
             * getConnectionChannels queries all the channels associated with a connection
             * end.
             */
            getConnectionChannels: withMetadata(async function getConnectionChannels(input: DeepPartial<ibc_core_channel_v1_query.QueryConnectionChannelsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).connectionChannels(input, options);
            }, { path: [6, 2] }),
            /**
             * getChannelClientState queries for the client state for the channel associated
             * with the provided channel identifiers.
             */
            getChannelClientState: withMetadata(async function getChannelClientState(input: DeepPartial<ibc_core_channel_v1_query.QueryChannelClientStateRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).channelClientState(input, options);
            }, { path: [6, 3] }),
            /**
             * getChannelConsensusState queries for the consensus state for the channel
             * associated with the provided channel identifiers.
             */
            getChannelConsensusState: withMetadata(async function getChannelConsensusState(input: DeepPartial<ibc_core_channel_v1_query.QueryChannelConsensusStateRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).channelConsensusState(input, options);
            }, { path: [6, 4] }),
            /**
             * getPacketCommitment queries a stored packet commitment hash.
             */
            getPacketCommitment: withMetadata(async function getPacketCommitment(input: DeepPartial<ibc_core_channel_v1_query.QueryPacketCommitmentRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).packetCommitment(input, options);
            }, { path: [6, 5] }),
            /**
             * getPacketCommitments returns all the packet commitments hashes associated
             * with a channel.
             */
            getPacketCommitments: withMetadata(async function getPacketCommitments(input: DeepPartial<ibc_core_channel_v1_query.QueryPacketCommitmentsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).packetCommitments(input, options);
            }, { path: [6, 6] }),
            /**
             * getPacketReceipt queries if a given packet sequence has been received on the
             * queried chain
             */
            getPacketReceipt: withMetadata(async function getPacketReceipt(input: DeepPartial<ibc_core_channel_v1_query.QueryPacketReceiptRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).packetReceipt(input, options);
            }, { path: [6, 7] }),
            /**
             * getPacketAcknowledgement queries a stored packet acknowledgement hash.
             */
            getPacketAcknowledgement: withMetadata(async function getPacketAcknowledgement(input: DeepPartial<ibc_core_channel_v1_query.QueryPacketAcknowledgementRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).packetAcknowledgement(input, options);
            }, { path: [6, 8] }),
            /**
             * getPacketAcknowledgements returns all the packet acknowledgements associated
             * with a channel.
             */
            getPacketAcknowledgements: withMetadata(async function getPacketAcknowledgements(input: DeepPartial<ibc_core_channel_v1_query.QueryPacketAcknowledgementsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).packetAcknowledgements(input, options);
            }, { path: [6, 9] }),
            /**
             * getUnreceivedPackets returns all the unreceived IBC packets associated with a
             * channel and sequences.
             */
            getUnreceivedPackets: withMetadata(async function getUnreceivedPackets(input: DeepPartial<ibc_core_channel_v1_query.QueryUnreceivedPacketsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).unreceivedPackets(input, options);
            }, { path: [6, 10] }),
            /**
             * getUnreceivedAcks returns all the unreceived IBC acknowledgements associated
             * with a channel and sequences.
             */
            getUnreceivedAcks: withMetadata(async function getUnreceivedAcks(input: DeepPartial<ibc_core_channel_v1_query.QueryUnreceivedAcksRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).unreceivedAcks(input, options);
            }, { path: [6, 11] }),
            /**
             * getNextSequenceReceive returns the next receive sequence for a given channel.
             */
            getNextSequenceReceive: withMetadata(async function getNextSequenceReceive(input: DeepPartial<ibc_core_channel_v1_query.QueryNextSequenceReceiveRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).nextSequenceReceive(input, options);
            }, { path: [6, 12] }),
            /**
             * getNextSequenceSend returns the next send sequence for a given channel.
             */
            getNextSequenceSend: withMetadata(async function getNextSequenceSend(input: DeepPartial<ibc_core_channel_v1_query.QueryNextSequenceSendRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(6);
              return getClient(service).nextSequenceSend(input, options);
            }, { path: [6, 13] }),
            /**
             * channelOpenInit defines a rpc handler method for MsgChannelOpenInit.
             */
            channelOpenInit: withMetadata(async function channelOpenInit(input: DeepSimplify<ibc_core_channel_v1_tx.MsgChannelOpenInit>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).channelOpenInit(input, options);
            }, { path: [7, 0] }),
            /**
             * channelOpenTry defines a rpc handler method for MsgChannelOpenTry.
             */
            channelOpenTry: withMetadata(async function channelOpenTry(input: DeepSimplify<ibc_core_channel_v1_tx.MsgChannelOpenTry>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).channelOpenTry(input, options);
            }, { path: [7, 1] }),
            /**
             * channelOpenAck defines a rpc handler method for MsgChannelOpenAck.
             */
            channelOpenAck: withMetadata(async function channelOpenAck(input: DeepSimplify<ibc_core_channel_v1_tx.MsgChannelOpenAck>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).channelOpenAck(input, options);
            }, { path: [7, 2] }),
            /**
             * channelOpenConfirm defines a rpc handler method for MsgChannelOpenConfirm.
             */
            channelOpenConfirm: withMetadata(async function channelOpenConfirm(input: DeepSimplify<ibc_core_channel_v1_tx.MsgChannelOpenConfirm>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).channelOpenConfirm(input, options);
            }, { path: [7, 3] }),
            /**
             * channelCloseInit defines a rpc handler method for MsgChannelCloseInit.
             */
            channelCloseInit: withMetadata(async function channelCloseInit(input: DeepSimplify<ibc_core_channel_v1_tx.MsgChannelCloseInit>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).channelCloseInit(input, options);
            }, { path: [7, 4] }),
            /**
             * channelCloseConfirm defines a rpc handler method for
             * MsgChannelCloseConfirm.
             */
            channelCloseConfirm: withMetadata(async function channelCloseConfirm(input: DeepSimplify<ibc_core_channel_v1_tx.MsgChannelCloseConfirm>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).channelCloseConfirm(input, options);
            }, { path: [7, 5] }),
            /**
             * recvPacket defines a rpc handler method for MsgRecvPacket.
             */
            recvPacket: withMetadata(async function recvPacket(input: DeepSimplify<ibc_core_channel_v1_tx.MsgRecvPacket>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).recvPacket(input, options);
            }, { path: [7, 6] }),
            /**
             * timeout defines a rpc handler method for MsgTimeout.
             */
            timeout: withMetadata(async function timeout(input: DeepSimplify<ibc_core_channel_v1_tx.MsgTimeout>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).timeout(input, options);
            }, { path: [7, 7] }),
            /**
             * timeoutOnClose defines a rpc handler method for MsgTimeoutOnClose.
             */
            timeoutOnClose: withMetadata(async function timeoutOnClose(input: DeepSimplify<ibc_core_channel_v1_tx.MsgTimeoutOnClose>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).timeoutOnClose(input, options);
            }, { path: [7, 8] }),
            /**
             * acknowledgement defines a rpc handler method for MsgAcknowledgement.
             */
            acknowledgement: withMetadata(async function acknowledgement(input: DeepSimplify<ibc_core_channel_v1_tx.MsgAcknowledgement>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(7);
              return getMsgClient(service).acknowledgement(input, options);
            }, { path: [7, 9] })
          },
          v2: {
            /**
             * getNextSequenceSend returns the next send sequence for a given channel.
             */
            getNextSequenceSend: withMetadata(async function getNextSequenceSend(input: DeepPartial<ibc_core_channel_v2_query.QueryNextSequenceSendRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).nextSequenceSend(input, options);
            }, { path: [8, 0] }),
            /**
             * getPacketCommitment queries a stored packet commitment hash.
             */
            getPacketCommitment: withMetadata(async function getPacketCommitment(input: DeepPartial<ibc_core_channel_v2_query.QueryPacketCommitmentRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).packetCommitment(input, options);
            }, { path: [8, 1] }),
            /**
             * getPacketCommitments queries a stored packet commitment hash.
             */
            getPacketCommitments: withMetadata(async function getPacketCommitments(input: DeepPartial<ibc_core_channel_v2_query.QueryPacketCommitmentsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).packetCommitments(input, options);
            }, { path: [8, 2] }),
            /**
             * getPacketAcknowledgement queries a stored acknowledgement commitment hash.
             */
            getPacketAcknowledgement: withMetadata(async function getPacketAcknowledgement(input: DeepPartial<ibc_core_channel_v2_query.QueryPacketAcknowledgementRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).packetAcknowledgement(input, options);
            }, { path: [8, 3] }),
            /**
             * getPacketAcknowledgements returns all packet acknowledgements associated with a channel.
             */
            getPacketAcknowledgements: withMetadata(async function getPacketAcknowledgements(input: DeepPartial<ibc_core_channel_v2_query.QueryPacketAcknowledgementsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).packetAcknowledgements(input, options);
            }, { path: [8, 4] }),
            /**
             * getPacketReceipt queries a stored packet receipt.
             */
            getPacketReceipt: withMetadata(async function getPacketReceipt(input: DeepPartial<ibc_core_channel_v2_query.QueryPacketReceiptRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).packetReceipt(input, options);
            }, { path: [8, 5] }),
            /**
             * getUnreceivedPackets returns all the unreceived IBC packets associated with a channel and sequences.
             */
            getUnreceivedPackets: withMetadata(async function getUnreceivedPackets(input: DeepPartial<ibc_core_channel_v2_query.QueryUnreceivedPacketsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).unreceivedPackets(input, options);
            }, { path: [8, 6] }),
            /**
             * getUnreceivedAcks returns all the unreceived IBC acknowledgements associated with a channel and sequences.
             */
            getUnreceivedAcks: withMetadata(async function getUnreceivedAcks(input: DeepPartial<ibc_core_channel_v2_query.QueryUnreceivedAcksRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(8);
              return getClient(service).unreceivedAcks(input, options);
            }, { path: [8, 7] }),
            /**
             * sendPacket defines a rpc handler method for MsgSendPacket.
             */
            sendPacket: withMetadata(async function sendPacket(input: DeepSimplify<ibc_core_channel_v2_tx.MsgSendPacket>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(9);
              return getMsgClient(service).sendPacket(input, options);
            }, { path: [9, 0] }),
            /**
             * recvPacket defines a rpc handler method for MsgRecvPacket.
             */
            recvPacket: withMetadata(async function recvPacket(input: DeepSimplify<ibc_core_channel_v2_tx.MsgRecvPacket>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(9);
              return getMsgClient(service).recvPacket(input, options);
            }, { path: [9, 1] }),
            /**
             * timeout defines a rpc handler method for MsgTimeout.
             */
            timeout: withMetadata(async function timeout(input: DeepSimplify<ibc_core_channel_v2_tx.MsgTimeout>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(9);
              return getMsgClient(service).timeout(input, options);
            }, { path: [9, 2] }),
            /**
             * acknowledgement defines a rpc handler method for MsgAcknowledgement.
             */
            acknowledgement: withMetadata(async function acknowledgement(input: DeepSimplify<ibc_core_channel_v2_tx.MsgAcknowledgement>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(9);
              return getMsgClient(service).acknowledgement(input, options);
            }, { path: [9, 3] })
          }
        },
        client: {
          v1: {
            /**
             * getClientState queries an IBC light client.
             */
            getClientState: withMetadata(async function getClientState(input: DeepPartial<ibc_core_client_v1_query.QueryClientStateRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).clientState(input, options);
            }, { path: [10, 0] }),
            /**
             * getClientStates queries all the IBC light clients of a chain.
             */
            getClientStates: withMetadata(async function getClientStates(input: DeepPartial<ibc_core_client_v1_query.QueryClientStatesRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).clientStates(input, options);
            }, { path: [10, 1] }),
            /**
             * getConsensusState queries a consensus state associated with a client state at
             * a given height.
             */
            getConsensusState: withMetadata(async function getConsensusState(input: DeepPartial<ibc_core_client_v1_query.QueryConsensusStateRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).consensusState(input, options);
            }, { path: [10, 2] }),
            /**
             * getConsensusStates queries all the consensus state associated with a given
             * client.
             */
            getConsensusStates: withMetadata(async function getConsensusStates(input: DeepPartial<ibc_core_client_v1_query.QueryConsensusStatesRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).consensusStates(input, options);
            }, { path: [10, 3] }),
            /**
             * getConsensusStateHeights queries the height of every consensus states associated with a given client.
             */
            getConsensusStateHeights: withMetadata(async function getConsensusStateHeights(input: DeepPartial<ibc_core_client_v1_query.QueryConsensusStateHeightsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).consensusStateHeights(input, options);
            }, { path: [10, 4] }),
            /**
             * Status queries the status of an IBC client.
             */
            getClientStatus: withMetadata(async function getClientStatus(input: DeepPartial<ibc_core_client_v1_query.QueryClientStatusRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).clientStatus(input, options);
            }, { path: [10, 5] }),
            /**
             * getClientParams queries all parameters of the ibc client submodule.
             */
            getClientParams: withMetadata(async function getClientParams(input: DeepPartial<ibc_core_client_v1_query.QueryClientParamsRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).clientParams(input, options);
            }, { path: [10, 6] }),
            /**
             * getClientCreator queries the creator of a given client.
             */
            getClientCreator: withMetadata(async function getClientCreator(input: DeepPartial<ibc_core_client_v1_query.QueryClientCreatorRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).clientCreator(input, options);
            }, { path: [10, 7] }),
            /**
             * getUpgradedClientState queries an Upgraded IBC light client.
             */
            getUpgradedClientState: withMetadata(async function getUpgradedClientState(input: DeepPartial<ibc_core_client_v1_query.QueryUpgradedClientStateRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).upgradedClientState(input, options);
            }, { path: [10, 8] }),
            /**
             * getUpgradedConsensusState queries an Upgraded IBC consensus state.
             */
            getUpgradedConsensusState: withMetadata(async function getUpgradedConsensusState(input: DeepPartial<ibc_core_client_v1_query.QueryUpgradedConsensusStateRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).upgradedConsensusState(input, options);
            }, { path: [10, 9] }),
            /**
             * getVerifyMembership queries an IBC light client for proof verification of a value at a given key path.
             */
            getVerifyMembership: withMetadata(async function getVerifyMembership(input: DeepPartial<ibc_core_client_v1_query.QueryVerifyMembershipRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(10);
              return getClient(service).verifyMembership(input, options);
            }, { path: [10, 10] }),
            /**
             * createClient defines a rpc handler method for MsgCreateClient.
             */
            createClient: withMetadata(async function createClient(input: DeepSimplify<ibc_core_client_v1_tx.MsgCreateClient>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).createClient(input, options);
            }, { path: [11, 0] }),
            /**
             * updateClient defines a rpc handler method for MsgUpdateClient.
             */
            updateClient: withMetadata(async function updateClient(input: DeepSimplify<ibc_core_client_v1_tx.MsgUpdateClient>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).updateClient(input, options);
            }, { path: [11, 1] }),
            /**
             * upgradeClient defines a rpc handler method for MsgUpgradeClient.
             */
            upgradeClient: withMetadata(async function upgradeClient(input: DeepSimplify<ibc_core_client_v1_tx.MsgUpgradeClient>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).upgradeClient(input, options);
            }, { path: [11, 2] }),
            /**
             * submitMisbehaviour defines a rpc handler method for MsgSubmitMisbehaviour.
             */
            submitMisbehaviour: withMetadata(async function submitMisbehaviour(input: DeepSimplify<ibc_core_client_v1_tx.MsgSubmitMisbehaviour>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).submitMisbehaviour(input, options);
            }, { path: [11, 3] }),
            /**
             * recoverClient defines a rpc handler method for MsgRecoverClient.
             */
            recoverClient: withMetadata(async function recoverClient(input: DeepSimplify<ibc_core_client_v1_tx.MsgRecoverClient>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).recoverClient(input, options);
            }, { path: [11, 4] }),
            /**
             * iBCSoftwareUpgrade defines a rpc handler method for MsgIBCSoftwareUpgrade.
             */
            iBCSoftwareUpgrade: withMetadata(async function iBCSoftwareUpgrade(input: DeepSimplify<ibc_core_client_v1_tx.MsgIBCSoftwareUpgrade>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).iBCSoftwareUpgrade(input, options);
            }, { path: [11, 5] }),
            /**
             * updateClientParams defines a rpc handler method for MsgUpdateParams.
             */
            updateClientParams: withMetadata(async function updateClientParams(input: DeepSimplify<ibc_core_client_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).updateClientParams(input, options);
            }, { path: [11, 6] }),
            /**
             * deleteClientCreator defines a rpc handler method for MsgDeleteClientCreator.
             */
            deleteClientCreator: withMetadata(async function deleteClientCreator(input: DeepSimplify<ibc_core_client_v1_tx.MsgDeleteClientCreator>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(11);
              return getMsgClient(service).deleteClientCreator(input, options);
            }, { path: [11, 7] })
          },
          v2: {
            /**
             * getCounterpartyInfo queries an IBC light counter party info.
             */
            getCounterpartyInfo: withMetadata(async function getCounterpartyInfo(input: DeepPartial<ibc_core_client_v2_query.QueryCounterpartyInfoRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).counterpartyInfo(input, options);
            }, { path: [12, 0] }),
            /**
             * getConfig queries the IBC client v2 configuration for a given client.
             */
            getConfig: withMetadata(async function getConfig(input: DeepPartial<ibc_core_client_v2_query.QueryConfigRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(12);
              return getClient(service).config(input, options);
            }, { path: [12, 1] }),
            /**
             * registerCounterparty defines a rpc handler method for MsgRegisterCounterparty.
             */
            registerCounterparty: withMetadata(async function registerCounterparty(input: DeepSimplify<ibc_core_client_v2_tx.MsgRegisterCounterparty>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(13);
              return getMsgClient(service).registerCounterparty(input, options);
            }, { path: [13, 0] }),
            /**
             * updateClientConfig defines a rpc handler method for MsgUpdateClientConfig.
             */
            updateClientConfig: withMetadata(async function updateClientConfig(input: DeepSimplify<ibc_core_client_v2_tx.MsgUpdateClientConfig>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(13);
              return getMsgClient(service).updateClientConfig(input, options);
            }, { path: [13, 1] })
          }
        },
        connection: {
          v1: {
            /**
             * getConnection queries an IBC connection end.
             */
            getConnection: withMetadata(async function getConnection(input: DeepPartial<ibc_core_connection_v1_query.QueryConnectionRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(14);
              return getClient(service).connection(input, options);
            }, { path: [14, 0] }),
            /**
             * getConnections queries all the IBC connections of a chain.
             */
            getConnections: withMetadata(async function getConnections(input: DeepPartial<ibc_core_connection_v1_query.QueryConnectionsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(14);
              return getClient(service).connections(input, options);
            }, { path: [14, 1] }),
            /**
             * getClientConnections queries the connection paths associated with a client
             * state.
             */
            getClientConnections: withMetadata(async function getClientConnections(input: DeepPartial<ibc_core_connection_v1_query.QueryClientConnectionsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(14);
              return getClient(service).clientConnections(input, options);
            }, { path: [14, 2] }),
            /**
             * getConnectionClientState queries the client state associated with the
             * connection.
             */
            getConnectionClientState: withMetadata(async function getConnectionClientState(input: DeepPartial<ibc_core_connection_v1_query.QueryConnectionClientStateRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(14);
              return getClient(service).connectionClientState(input, options);
            }, { path: [14, 3] }),
            /**
             * getConnectionConsensusState queries the consensus state associated with the
             * connection.
             */
            getConnectionConsensusState: withMetadata(async function getConnectionConsensusState(input: DeepPartial<ibc_core_connection_v1_query.QueryConnectionConsensusStateRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(14);
              return getClient(service).connectionConsensusState(input, options);
            }, { path: [14, 4] }),
            /**
             * getConnectionParams queries all parameters of the ibc connection submodule.
             */
            getConnectionParams: withMetadata(async function getConnectionParams(input: DeepPartial<ibc_core_connection_v1_query.QueryConnectionParamsRequest> = {}, options?: CallOptions) {
              const service = await serviceLoader.loadAt(14);
              return getClient(service).connectionParams(input, options);
            }, { path: [14, 5] }),
            /**
             * connectionOpenInit defines a rpc handler method for MsgConnectionOpenInit.
             */
            connectionOpenInit: withMetadata(async function connectionOpenInit(input: DeepSimplify<ibc_core_connection_v1_tx.MsgConnectionOpenInit>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(15);
              return getMsgClient(service).connectionOpenInit(input, options);
            }, { path: [15, 0] }),
            /**
             * connectionOpenTry defines a rpc handler method for MsgConnectionOpenTry.
             */
            connectionOpenTry: withMetadata(async function connectionOpenTry(input: DeepSimplify<ibc_core_connection_v1_tx.MsgConnectionOpenTry>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(15);
              return getMsgClient(service).connectionOpenTry(input, options);
            }, { path: [15, 1] }),
            /**
             * connectionOpenAck defines a rpc handler method for MsgConnectionOpenAck.
             */
            connectionOpenAck: withMetadata(async function connectionOpenAck(input: DeepSimplify<ibc_core_connection_v1_tx.MsgConnectionOpenAck>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(15);
              return getMsgClient(service).connectionOpenAck(input, options);
            }, { path: [15, 2] }),
            /**
             * connectionOpenConfirm defines a rpc handler method for
             * MsgConnectionOpenConfirm.
             */
            connectionOpenConfirm: withMetadata(async function connectionOpenConfirm(input: DeepSimplify<ibc_core_connection_v1_tx.MsgConnectionOpenConfirm>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(15);
              return getMsgClient(service).connectionOpenConfirm(input, options);
            }, { path: [15, 3] }),
            /**
             * updateConnectionParams defines a rpc handler method for
             * MsgUpdateParams.
             */
            updateConnectionParams: withMetadata(async function updateConnectionParams(input: DeepSimplify<ibc_core_connection_v1_tx.MsgUpdateParams>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(15);
              return getMsgClient(service).updateConnectionParams(input, options);
            }, { path: [15, 4] })
          }
        }
      },
      lightclients: {
        wasm: {
          v1: {
            /**
             * Get all Wasm checksums
             */
            getChecksums: withMetadata(async function getChecksums(input: DeepPartial<ibc_lightclients_wasm_v1_query.QueryChecksumsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(16);
              return getClient(service).checksums(input, options);
            }, { path: [16, 0] }),
            /**
             * Get Wasm code for given checksum
             */
            getCode: withMetadata(async function getCode(input: DeepPartial<ibc_lightclients_wasm_v1_query.QueryCodeRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(16);
              return getClient(service).code(input, options);
            }, { path: [16, 1] }),
            /**
             * storeCode defines a rpc handler method for MsgStoreCode.
             */
            storeCode: withMetadata(async function storeCode(input: DeepSimplify<ibc_lightclients_wasm_v1_tx.MsgStoreCode>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(17);
              return getMsgClient(service).storeCode(input, options);
            }, { path: [17, 0] }),
            /**
             * removeChecksum defines a rpc handler method for MsgRemoveChecksum.
             */
            removeChecksum: withMetadata(async function removeChecksum(input: DeepSimplify<ibc_lightclients_wasm_v1_tx.MsgRemoveChecksum>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(17);
              return getMsgClient(service).removeChecksum(input, options);
            }, { path: [17, 1] }),
            /**
             * migrateContract defines a rpc handler method for MsgMigrateContract.
             */
            migrateContract: withMetadata(async function migrateContract(input: DeepSimplify<ibc_lightclients_wasm_v1_tx.MsgMigrateContract>, options?: TxCallOptions) {
              const service = await serviceLoader.loadAt(17);
              return getMsgClient(service).migrateContract(input, options);
            }, { path: [17, 2] })
          }
        }
      }
    }
  };
}
