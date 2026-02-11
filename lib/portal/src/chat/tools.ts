import type { ChatToolHandler } from "./types";
import { createDeploymentTools } from "./chain-tools/deployments";
import { createMarketplaceTools } from "./chain-tools/marketplace";
import { createIdentityTools } from "./chain-tools/identity";
import { createGovernanceTools } from "./chain-tools/governance";
import { createWalletTools } from "./chain-tools/wallet";

export const createDefaultChatTools = (): ChatToolHandler[] => [
  ...createDeploymentTools(),
  ...createMarketplaceTools(),
  ...createIdentityTools(),
  ...createGovernanceTools(),
  ...createWalletTools(),
];
