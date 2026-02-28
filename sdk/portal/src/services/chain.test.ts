import { describe, expect, it } from "vitest";

import { resolveRuntimeConfig } from "./chain";

describe("resolveRuntimeConfig", () => {
  it("prefers VITE variables and derives testnet defaults", () => {
    const config = resolveRuntimeConfig({
      VITE_CHAIN_ID: "virtengine-testnet-1",
      VITE_APP_URL: "https://staging.portal.virtengine.com/",
      VITE_SUPPORTED_WALLETS: "keplr, leap ,cosmostation",
    });

    expect(config.chainId).toBe("virtengine-testnet-1");
    expect(config.rpc).toBe("https://rpc.testnet.virtengine.com");
    expect(config.rest).toBe("https://api.testnet.virtengine.com");
    expect(config.ws).toBe("wss://ws.testnet.virtengine.com");
    expect(config.appUrl).toBe("https://staging.portal.virtengine.com");
    expect(config.supportedWallets).toEqual(["keplr", "leap", "cosmostation"]);
    expect(config.chainLabel).toBe("Testnet ready");
  });

  it("accepts NEXT_PUBLIC variables for parity with the main portal", () => {
    const config = resolveRuntimeConfig({
      NEXT_PUBLIC_CHAIN_ID: "virtengine-1",
      NEXT_PUBLIC_CHAIN_RPC: "https://rpc.override.virtengine.com/",
      NEXT_PUBLIC_CHAIN_REST: "https://api.override.virtengine.com/",
      NEXT_PUBLIC_CHAIN_WS: "wss://ws.override.virtengine.com/",
      NEXT_PUBLIC_PROVIDER_DAEMON_URL: "https://provider.virtengine.com/",
      NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID: "project-123",
    });

    expect(config.chainId).toBe("virtengine-1");
    expect(config.rpc).toBe("https://rpc.override.virtengine.com");
    expect(config.rest).toBe("https://api.override.virtengine.com");
    expect(config.ws).toBe("wss://ws.override.virtengine.com");
    expect(config.providerDaemonUrl).toBe("https://provider.virtengine.com");
    expect(config.walletConnectProjectId).toBe("project-123");
    expect(config.chainLabel).toBe("Mainnet ready");
  });
});
