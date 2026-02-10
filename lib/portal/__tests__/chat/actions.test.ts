import { describe, it, expect } from "vitest";
import { transferTokens } from "../../src/chat/chain-tools/wallet";

const runtime = {
  walletAddress: "virtengine1abc",
  tokenDenom: "uve",
};

describe("chat action confirmation flow", () => {
  it("should create a confirmation-required transfer action", async () => {
    const result = await transferTokens(
      { toAddress: "virtengine1xyz", amount: "100", denom: "uve" },
      runtime,
    );

    expect(result.action).toBeDefined();
    expect(result.action?.confirmationRequired).toBe(true);
    expect(result.action?.preview?.severity).toBe("danger");
  });
});
