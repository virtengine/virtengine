import { describe, it, expect } from "vitest";
import { buildChatContext } from "../../src/chat/context";

const wallet = {
  accounts: [{ address: "virtengine1abcd" }],
  activeAccountIndex: 0,
  chainId: "virtengine-1",
} as any;

const chain = {
  chainId: "virtengine-1",
  networkName: "VirtEngine",
} as any;
const chainConfig = {
  restEndpoint: "https://api.virtengine.com",
  wsEndpoint: "wss://ws.virtengine.com",
};

describe("chat context", () => {
  it("builds context from wallet and chain state", () => {
    const context = buildChatContext({ wallet, chain, chainConfig });
    expect(context.walletAddress).toBe("virtengine1abcd");
    expect(context.chainId).toBe("virtengine-1");
    expect(context.chainRestEndpoint).toBe("https://api.virtengine.com");
  });
});
