import { beforeEach, describe, expect, it } from "@jest/globals";

import type { ChainInfo } from "./types.ts";
import { WalletManager } from "./WalletManager.ts";

describe("WalletManager", () => {
  let manager: WalletManager;

  const mockChainInfo: ChainInfo = {
    chainId: "virtengine-testnet-1",
    chainName: "VirtEngine Testnet",
    rpc: "https://rpc.testnet.virtengine.io",
    rest: "https://api.testnet.virtengine.io",
    bip44: { coinType: 118 },
    bech32Config: {
      bech32PrefixAccAddr: "virt",
      bech32PrefixAccPub: "virtpub",
      bech32PrefixValAddr: "virtvaloper",
      bech32PrefixValPub: "virtvaloperpub",
      bech32PrefixConsAddr: "virtvalcons",
      bech32PrefixConsPub: "virtvalconspub",
    },
    currencies: [
      {
        coinDenom: "VIRT",
        coinMinimalDenom: "uvirt",
        coinDecimals: 6,
      },
    ],
    feeCurrencies: [
      {
        coinDenom: "VIRT",
        coinMinimalDenom: "uvirt",
        coinDecimals: 6,
      },
    ],
    stakeCurrency: {
      coinDenom: "VIRT",
      coinMinimalDenom: "uvirt",
      coinDecimals: 6,
    },
  };

  beforeEach(() => {
    manager = new WalletManager();
  });

  describe("getAvailableWallets", () => {
    it("should return list of supported wallets", () => {
      const wallets = manager.getAvailableWallets();
      // In Node.js environment, no wallets will be available (browser extensions)
      expect(Array.isArray(wallets)).toBe(true);
    });

    it("should return empty array in Node.js environment", () => {
      // Browser extensions are not available in Node.js
      const wallets = manager.getAvailableWallets();
      expect(wallets).toHaveLength(0);
    });
  });

  describe("getAdapter", () => {
    it("should return adapter for supported wallet", () => {
      const adapter = manager.getAdapter("keplr");
      expect(adapter).toBeDefined();
      expect(adapter?.name).toBe("Keplr");
    });

    it("should return undefined for unsupported wallet", () => {
      const adapter = manager.getAdapter("unknown" as never);
      expect(adapter).toBeUndefined();
    });
  });

  describe("getCurrentAdapter", () => {
    it("should return null when no wallet connected", () => {
      expect(manager.getCurrentAdapter()).toBeNull();
    });
  });

  describe("getChainInfo", () => {
    it("should return null when no wallet connected", () => {
      expect(manager.getChainInfo()).toBeNull();
    });
  });

  describe("connect", () => {
    it("should throw error when wallet not available", async () => {
      await expect(manager.connect("keplr", mockChainInfo)).rejects.toThrow();
    });
  });

  describe("disconnect", () => {
    it("should not throw when no wallet connected", async () => {
      await expect(manager.disconnect()).resolves.not.toThrow();
    });
  });

  describe("getAddress", () => {
    it("should throw error when no wallet connected", () => {
      expect(() => manager.getAddress()).toThrow("No wallet connected");
    });
  });

  describe("isConnected", () => {
    it("should return false when no wallet connected", () => {
      expect(manager.isConnected()).toBe(false);
    });
  });
});
