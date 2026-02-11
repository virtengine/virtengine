import type { ChatContextSnapshot, ChatToolContext } from "./types";
import type { ChainState } from "../../types/chain";
import type { WalletState } from "../../src/wallet/types";

export interface ChatContextInput {
  wallet?: WalletState | null;
  chain?: ChainState | null;
  chainConfig?: {
    restEndpoint?: string;
    wsEndpoint?: string;
    chainId?: string;
  };
  identity?: { score?: number; status?: string } | null;
  roles?: string[];
  permissions?: string[];
}

export const buildChatContext = (input: ChatContextInput): ChatToolContext => {
  const walletAddress =
    input.wallet?.accounts[input.wallet.activeAccountIndex]?.address;

  return {
    walletAddress,
    chainId:
      input.chain?.chainId ??
      input.chainConfig?.chainId ??
      input.wallet?.chainId ??
      undefined,
    networkName: input.chain?.networkName ?? undefined,
    chainRestEndpoint: input.chainConfig?.restEndpoint ?? undefined,
    chainEndpoint: input.chainConfig?.wsEndpoint ?? undefined,
    balances: undefined,
    identity: input.identity ?? undefined,
    roles: input.roles ?? [],
    permissions: input.permissions ?? [],
  };
};

export const createChatSnapshot = (
  context: ChatToolContext,
): ChatContextSnapshot => ({
  walletAddress: context.walletAddress,
  chainId: context.chainId,
  networkName: context.networkName,
  balances: context.balances,
  identity: context.identity ?? null,
  roles: context.roles ?? [],
  permissions: context.permissions ?? [],
  generatedAt: Date.now(),
});
