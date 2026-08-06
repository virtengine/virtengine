import { Registry, type OfflineSigner } from "@cosmjs/proto-signing";
import { SigningStargateClient, type Coin } from "@cosmjs/stargate";
import type { AssetList, Chain } from "@chain-registry/types";

import type { ChainStatus, VeidStatus } from "../types/chain";

type RuntimeEnv = Record<string, string | boolean | undefined>;

export interface RuntimeConfig {
  chainId: string;
  rpc: string;
  rest: string;
  ws: string;
  appUrl: string;
  providerDaemonUrl: string;
  walletConnectProjectId: string;
  supportedWallets: string[];
  chainLabel: string;
}

const MAINNET = {
  chainId: "virtengine-1",
  rpc: "https://rpc.virtengine.com",
  rest: "https://api.virtengine.com",
  ws: "wss://ws.virtengine.com",
  label: "Mainnet ready",
};

const TESTNET = {
  chainId: "virtengine-testnet-1",
  rpc: "https://rpc.testnet.virtengine.com",
  rest: "https://api.testnet.virtengine.com",
  ws: "wss://ws.testnet.virtengine.com",
  label: "Testnet ready",
};

const DEVNET = {
  chainId: "virtengine-devnet-1",
  rpc: "http://localhost:26657",
  rest: "http://localhost:1317",
  ws: "ws://localhost:26657/websocket",
  label: "Devnet override",
};

const LOCALNET = {
  chainId: "virtengine-localnet-1",
  rpc: "http://localhost:26657",
  rest: "http://localhost:1317",
  ws: "ws://localhost:26657/websocket",
  label: "Localnet override",
};

function getEnvValue(env: RuntimeEnv, ...keys: string[]): string | undefined {
  for (const key of keys) {
    const value = env[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return undefined;
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

function deriveNetworkDefaults(chainId: string) {
  const normalized = chainId.toLowerCase();
  if (normalized.includes("localnet")) return LOCALNET;
  if (normalized.includes("devnet")) return DEVNET;
  if (normalized.includes("testnet")) return TESTNET;
  return MAINNET;
}

export function resolveRuntimeConfig(env: RuntimeEnv = import.meta.env): RuntimeConfig {
  const requestedChainId =
    getEnvValue(env, "VITE_CHAIN_ID", "NEXT_PUBLIC_CHAIN_ID") ?? MAINNET.chainId;
  const defaults = deriveNetworkDefaults(requestedChainId);
  const appUrl =
    getEnvValue(env, "VITE_APP_URL", "NEXT_PUBLIC_APP_URL") ??
    "https://portal.virtengine.com";
  const providerDaemonUrl =
    getEnvValue(env, "VITE_PROVIDER_DAEMON_URL", "NEXT_PUBLIC_PROVIDER_DAEMON_URL") ?? "";
  const supportedWallets = (
    getEnvValue(env, "VITE_SUPPORTED_WALLETS", "NEXT_PUBLIC_SUPPORTED_WALLETS") ??
    "keplr,leap,cosmostation"
  )
    .split(",")
    .map((wallet) => wallet.trim())
    .filter(Boolean);

  return {
    chainId: requestedChainId,
    rpc: trimTrailingSlash(
      getEnvValue(env, "VITE_CHAIN_RPC", "NEXT_PUBLIC_CHAIN_RPC") ?? defaults.rpc
    ),
    rest: trimTrailingSlash(
      getEnvValue(env, "VITE_CHAIN_REST", "NEXT_PUBLIC_CHAIN_REST") ?? defaults.rest
    ),
    ws: trimTrailingSlash(
      getEnvValue(env, "VITE_CHAIN_WS", "NEXT_PUBLIC_CHAIN_WS") ?? defaults.ws
    ),
    appUrl: trimTrailingSlash(appUrl),
    providerDaemonUrl: trimTrailingSlash(providerDaemonUrl),
    walletConnectProjectId:
      getEnvValue(
        env,
        "VITE_WALLET_CONNECT_PROJECT_ID",
        "NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID"
      ) ?? "",
    supportedWallets,
    chainLabel: defaults.label,
  };
}

export const runtimeConfig = resolveRuntimeConfig();

export const chainName = "virtengine";
export const chainConfig = {
  chainId: runtimeConfig.chainId,
  rpc: runtimeConfig.rpc,
  rest: runtimeConfig.rest,
  ws: runtimeConfig.ws,
};

export const virtengineChain: Chain = {
  chain_id: chainConfig.chainId,
  chain_name: chainName,
  chain_type: "cosmos",
  pretty_name: "VirtEngine",
  status: "live",
  network_type: runtimeConfig.chainId.includes("testnet") ? "testnet" : "mainnet",
  bech32_prefix: "virtengine",
  slip44: 118,
  fees: {
    fee_tokens: [
      {
        denom: "uve",
        fixed_min_gas_price: 0.01,
        low_gas_price: 0.01,
        average_gas_price: 0.025,
        high_gas_price: 0.04,
      },
    ],
  },
  staking: {
    staking_tokens: [{ denom: "uve" }],
  },
  apis: {
    rpc: [{ address: chainConfig.rpc }],
    rest: [{ address: chainConfig.rest }],
  },
};

export const virtengineAssets: AssetList = {
  chain_name: chainName,
  assets: [
    {
      base: "uve",
      name: "VirtEngine",
      display: "VE",
      symbol: "VE",
      type_asset: "sdk.coin",
      denom_units: [
        { denom: "uve", exponent: 0 },
        { denom: "mve", exponent: 3 },
        { denom: "virtengine", exponent: 6 },
        { denom: "VE", exponent: 6 },
      ],
    },
  ],
};

const jsonFetch = async <T>(url: string): Promise<T> => {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Request failed (${response.status}) for ${url}`);
  }
  return (await response.json()) as T;
};

export const fetchChainStatus = async (): Promise<ChainStatus> => {
  const rest = chainConfig.rest;
  const [blockData, validatorsData] = await Promise.all([
    jsonFetch<{ block?: { header?: { chain_id?: string; height?: string } } }>(
      `${rest}/cosmos/base/tendermint/v1beta1/blocks/latest`
    ),
    jsonFetch<{ validators?: unknown[] }>(
      `${rest}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED`
    ),
  ]);

  const rawHeight = blockData?.block?.header?.height ?? "0";
  const latestHeight = Number(rawHeight);

  return {
    chainId: blockData?.block?.header?.chain_id ?? chainConfig.chainId,
    latestHeight: Number.isNaN(latestHeight) ? null : latestHeight,
    validatorCount: validatorsData?.validators?.length ?? 0,
  };
};

export const fetchBalances = async (address: string): Promise<readonly Coin[]> => {
  if (!address.trim()) {
    return [];
  }

  const payload = await jsonFetch<{ balances?: Array<{ denom?: string; amount?: string }> }>(
    `${chainConfig.rest}/cosmos/bank/v1beta1/balances/${address}`
  );

  return (payload.balances ?? []).map((coin) => ({
    denom: coin.denom ?? "",
    amount: coin.amount ?? "0",
  }));
};

export const createSigningClient = async (signer: OfflineSigner) => {
  const registry = new Registry();
  return SigningStargateClient.connectWithSigner(chainConfig.rpc, signer, {
    registry,
  });
};

export const fetchVeidStatus = async (address: string): Promise<VeidStatus> => {
  const rest = chainConfig.rest;
  const endpoints = [
    `${rest}/virtengine/veid/v1/identity_record/${address}`,
    `${rest}/virtengine/veid/v1/identity-record/${address}`,
    `${rest}/virtengine/veid/v1/identity_records/${address}`,
  ];

  for (const endpoint of endpoints) {
    try {
      const data = await jsonFetch<{ record?: { status?: string; state?: string } }>(
        endpoint
      );
      const status = data?.record?.status ?? data?.record?.state ?? "Not verified";
      return { status };
    } catch {
      // Try the next endpoint.
    }
  }

  return {
    status: "Unknown",
    detail: "VEID query not available on this node.",
  };
};

export const formatCoin = (coin: Coin): string => {
  if (coin.denom === "uve") {
    const value = Number(coin.amount) / 1_000_000;
    if (Number.isFinite(value)) {
      return `${value.toLocaleString(undefined, {
        maximumFractionDigits: 6,
      })} VE`;
    }
  }

  return `${coin.amount} ${coin.denom}`;
};
