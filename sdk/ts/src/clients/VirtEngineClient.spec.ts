import { describe, expect, it } from "@jest/globals";

import {
  VIRTENGINE_MAINNET,
  VIRTENGINE_TESTNET,
} from "./VirtEngineClient.ts";

describe("VirtEngineClient", () => {
  describe("Chain configs", () => {
    describe("VIRTENGINE_MAINNET", () => {
      it("should have correct chain ID", () => {
        expect(VIRTENGINE_MAINNET.chainId).toBe("virtengine-1");
      });

      it("should have correct chain name", () => {
        expect(VIRTENGINE_MAINNET.chainName).toBe("VirtEngine");
      });

      it("should have correct bech32 prefix", () => {
        expect(VIRTENGINE_MAINNET.bech32Config.bech32PrefixAccAddr).toBe("virt");
      });

      it("should have correct coin type", () => {
        expect(VIRTENGINE_MAINNET.bip44.coinType).toBe(118);
      });

      it("should have native currency configured", () => {
        expect(VIRTENGINE_MAINNET.currencies.length).toBeGreaterThanOrEqual(1);
        expect(VIRTENGINE_MAINNET.currencies[0].coinDenom).toBe("VIRT");
        expect(VIRTENGINE_MAINNET.currencies[0].coinMinimalDenom).toBe("uvirt");
        expect(VIRTENGINE_MAINNET.currencies[0].coinDecimals).toBe(6);
      });

      it("should have fee currency configured", () => {
        expect(VIRTENGINE_MAINNET.feeCurrencies).toHaveLength(1);
        expect(VIRTENGINE_MAINNET.feeCurrencies[0].coinMinimalDenom).toBe("uvirt");
      });

      it("should have stake currency configured", () => {
        expect(VIRTENGINE_MAINNET.stakeCurrency.coinMinimalDenom).toBe("uvirt");
      });
    });

    describe("VIRTENGINE_TESTNET", () => {
      it("should have correct chain ID", () => {
        expect(VIRTENGINE_TESTNET.chainId).toBe("virtengine-testnet-1");
      });

      it("should have correct chain name", () => {
        expect(VIRTENGINE_TESTNET.chainName).toBe("VirtEngine Testnet");
      });

      it("should have correct bech32 prefix", () => {
        expect(VIRTENGINE_TESTNET.bech32Config.bech32PrefixAccAddr).toBe("virt");
      });

      it("should use testnet RPC endpoints", () => {
        expect(VIRTENGINE_TESTNET.rpc).toContain("testnet");
      });

      it("should use testnet REST endpoints", () => {
        expect(VIRTENGINE_TESTNET.rest).toContain("testnet");
      });
    });
  });

  describe("Chain config structure", () => {
    it("should have all required bech32 prefixes", () => {
      const config = VIRTENGINE_MAINNET;
      expect(config.bech32Config.bech32PrefixAccAddr).toBeDefined();
      expect(config.bech32Config.bech32PrefixAccPub).toBeDefined();
      expect(config.bech32Config.bech32PrefixValAddr).toBeDefined();
      expect(config.bech32Config.bech32PrefixValPub).toBeDefined();
      expect(config.bech32Config.bech32PrefixConsAddr).toBeDefined();
      expect(config.bech32Config.bech32PrefixConsPub).toBeDefined();
    });

    it("should have consistent prefix naming", () => {
      const config = VIRTENGINE_MAINNET;
      expect(config.bech32Config.bech32PrefixValAddr).toBe("virtvaloper");
      expect(config.bech32Config.bech32PrefixConsAddr).toBe("virtvalcons");
    });
  });
});
