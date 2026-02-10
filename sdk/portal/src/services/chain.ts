import { Registry, type OfflineSigner } from "@cosmjs/proto-signing";
import { SigningStargateClient, StargateClient, type Coin } from "@cosmjs/stargate";
import type { AssetList, Chain } from "@chain-registry/types";

import type { ChainStatus, VeidStatus } from "../types/chain";

const defaultChainId = "virtengine-localnet-1";
const defaultRpc = "http://localhost:26657";
const defaultRest = "http://localhost:1317";

export const chainName = "virtengine";
export const chainConfig = {
  chainId: import.meta.env.VITE_CHAIN_ID ?? defaultChainId,
  rpc: import.meta.env.VITE_CHAIN_RPC ?? defaultRpc,
  rest: import.meta.env.VITE_CHAIN_REST ?? defaultRest,
};

export const virtengineChain: Chain = {
  chain_id: chainConfig.chainId,
  chain_name: chainName,
  chain_type: "cosmos",
  pretty_name: "VirtEngine",
  status: "live",
  network_type: "testnet",
  bech32_prefix: "ve",
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
  const rest = chainConfig.rest.replace(/\/$/, "");
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
  const client = await StargateClient.connect(chainConfig.rpc);
  return client.getAllBalances(address);
};

export const createSigningClient = async (signer: OfflineSigner) => {
  const registry = new Registry();
  return SigningStargateClient.connectWithSigner(chainConfig.rpc, signer, {
    registry,
  });
};

export const fetchVeidStatus = async (address: string): Promise<VeidStatus> => {
  const rest = chainConfig.rest.replace(/\/$/, "");
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
