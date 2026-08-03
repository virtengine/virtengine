import type { ChatToolHandler } from "./types";
import { createDeploymentTools } from "./chain-tools/deployments";
import { createMarketplaceTools } from "./chain-tools/marketplace";
import { createIdentityTools } from "./chain-tools/identity";
import { createGovernanceTools } from "./chain-tools/governance";
import { createWalletTools } from "./chain-tools/wallet";

export interface DefaultChatToolOptions {
  chatEnabled?: boolean;
  mutationsEnabled?: boolean;
}

export const createDefaultChatTools = (
  options: DefaultChatToolOptions = {},
): ChatToolHandler[] => {
  if (!options.chatEnabled) {
    return [];
  }

  const tools = [
    ...createDeploymentTools(),
    ...createMarketplaceTools(),
    ...createIdentityTools(),
    ...createGovernanceTools(),
    ...createWalletTools(),
  ];

  return tools.filter(
    (tool) =>
      options.mutationsEnabled === true ||
      (tool.definition.kind === "query" &&
        tool.definition.destructive !== true &&
        tool.execute === undefined),
  );
};
