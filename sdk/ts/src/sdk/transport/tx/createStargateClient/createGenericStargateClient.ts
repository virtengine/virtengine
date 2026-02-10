import type {
  AccountData,
  DirectSecp256k1HdWalletOptions,
  EncodeObject,
  GeneratedType,
  OfflineSigner,
} from "@cosmjs/proto-signing";
import {
  DirectSecp256k1HdWallet,
  Registry,
} from "@cosmjs/proto-signing";
import type {
  DeliverTxResponse,
  HttpEndpoint,
  SignerData,
  SigningStargateClientOptions,
  StdFee,
} from "@cosmjs/stargate";
import {
  calculateFee,
  GasPrice,
  SigningStargateClient,
} from "@cosmjs/stargate";

import type { TxClient, TxRaw } from "../TxClient.ts";

const DEFAULT_AVERAGE_GAS_PRICE = "0.025uve";
const DEFAULT_GAS_MULTIPLIER = 1.3;

export function createGenericStargateClient(options: WithSigner<BaseGenericStargateClientOptions>): StargateTxClient {
  const builtInTypes = options.builtInTypes?.map((type) => [type.typeUrl, type] as [string, GeneratedType]) || [];
  const registry = new Registry(builtInTypes);
  const createStargateClient = options.createClient ?? SigningStargateClient.connectWithSigner;

  let offlineSignerPromise: Promise<OfflineSigner> | undefined;
  const getOfflineSigner = () => offlineSignerPromise ??= createOfflineSigner(options);

  let stargateClientPromise: Promise<SigningStargateClient> | undefined;
  const getStargateClient = () => stargateClientPromise ??= getOfflineSigner().then((signer) => createStargateClient(
    options.baseUrl,
    signer,
    {
      ...options.stargateOptions,
      registry,
    },
  ));

  const getAccount = () => getOfflineSigner().then((signer) => (options.getAccount ?? getDefaultAccount)(signer));
  const gasMultiplier = options.gasMultiplier ?? DEFAULT_GAS_MULTIPLIER;
  const ensureMessageTypesRegistered = (messages: EncodeObject[]) => {
    for (const message of messages) {
      if (registry.lookupType(message.typeUrl)) continue;
      const type = options.getMessageType(message.typeUrl);
      if (!type) {
        throw new Error(`Cannot find message type ${message.typeUrl} in type registry. Probably it's not loaded yet.`);
      }
      registry.register(message.typeUrl, type);
    }
    return messages;
  };
  const gasPrice = GasPrice.fromString(options.defaultGasPrice ?? DEFAULT_AVERAGE_GAS_PRICE);

  return {
    getAccount,

    async signAndBroadcast(messages, options) {
      let fee: StdFee;
      const providedFee = options?.fee;
      if (!providedFee?.amount || !providedFee?.gas) {
        const estimatedFee = await this.estimateFee(messages, options?.memo);
        fee = providedFee ? { ...estimatedFee, ...providedFee } : estimatedFee;
      } else {
        fee = providedFee as StdFee;
      }

      const txRaw = await this.sign(messages, fee, options?.memo || "", undefined, options?.timeoutHeight);
      options?.afterSign?.(txRaw);
      const txResponse = await this.broadcast(txRaw);
      return txResponse;
    },
    async estimateFee(messages, memo) {
      ensureMessageTypesRegistered(messages);
      const account = await getAccount();
      const client = await getStargateClient();
      const estimatedGas = await client.simulate(account.address, messages, memo);
      const minGas = Math.floor(gasMultiplier * estimatedGas);
      const fee = calculateFee(minGas, gasPrice);

      return fee;
    },
    async sign(messages, fee, memo, explicitSignerData, timeoutHeight) {
      ensureMessageTypesRegistered(messages);
      const account = await getAccount();
      const client = await getStargateClient();
      return client.sign(account.address, messages, fee, memo, explicitSignerData, timeoutHeight);
    },
    async broadcast(txRaw) {
      const txTypeUrl = "/cosmos.tx.v1beta1.TxRaw";
      const TxRawType = registry.lookupType(txTypeUrl) || options.getMessageType(txTypeUrl);
      if (!TxRawType) {
        throw new Error("Cannot broadcast transaction: TxRaw type is not registered in transaction client");
      }
      const client = await getStargateClient();
      return client.broadcastTx(
        TxRawType.encode(txRaw).finish(),
        options.stargateOptions?.broadcastTimeoutMs,
        options.stargateOptions?.broadcastPollIntervalMs,
      );
    },
    async disconnect() {
      if (!stargateClientPromise) return;

      const client = await stargateClientPromise;
      client.disconnect();
      stargateClientPromise = undefined;
      offlineSignerPromise = undefined;
    },
  };
}

export interface StargateTxClient extends TxClient {
  estimateFee(messages: EncodeObject[], memo?: string): Promise<StdFee>;
  sign(messages: EncodeObject[], fee: StdFee, memo: string, explicitSignerData?: SignerData, timeoutHeight?: bigint): Promise<TxRaw>;
  broadcast(signedMessages: TxRaw): Promise<DeliverTxResponse>;
  getAccount(): Promise<AccountData>;
  disconnect(): Promise<void>;
}

export type WithSigner<T> = T & (
  | {
    /**
       * Signer to use for transactions signing
       */
    signer: OfflineSigner;
  }
  | {
    signer?: never;
    /**
       * Uses the mnemonic to create a `DirectSecp256k1HdWallet` to use for transactions signing
       */
    signerMnemonic: string;
    /**
       * Options to pass to the `DirectSecp256k1HdWallet`
       */
    signerOptions?: Partial<Omit<DirectSecp256k1HdWalletOptions, "prefix">>;
  }
);

export interface BaseGenericStargateClientOptions {
  /**
   * Blockchain RPC endpoint
   */
  baseUrl: string;
  /**
   * Gas multiplier
   * @default 1.3
   */
  gasMultiplier?: number;
  /**
   * @default "0.025uve"
   */
  defaultGasPrice?: string;
  /**
   * Retrieves the account to use for transactions
   * @default returns the first account from the signer
   */
  getAccount?(signer: OfflineSigner): Promise<AccountData>;
  stargateOptions?: Omit<SigningStargateClientOptions, "registry">;
  /**
   * Additional protobuf message types to register with the transaction transport
   */
  builtInTypes?: Array<GeneratedType & { typeUrl: string }>;
  getMessageType: (typeUrl: string) => GeneratedType | undefined;
  /**
   * Allows to use a custom Stargate client implementation.
   * @default `SigningStargateClient.connectWithSigner`
   */
  createClient?: (endpoint: string | HttpEndpoint, signer: OfflineSigner, options?: SigningStargateClientOptions) => Promise<SigningStargateClient>;
}

async function getDefaultAccount(signer: OfflineSigner) {
  const accounts = await signer.getAccounts();
  if (accounts.length === 0) {
    throw new Error("provided offline signer has no accounts");
  }
  return accounts[0];
}

function createOfflineSigner(options: WithSigner<BaseGenericStargateClientOptions>) {
  if ("signer" in options && options.signer) return Promise.resolve(options.signer);

  return DirectSecp256k1HdWallet.fromMnemonic(options.signerMnemonic, {
    ...options.signerOptions,
    prefix: "virtengine",
  });
}
