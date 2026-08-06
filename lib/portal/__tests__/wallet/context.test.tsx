import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { WalletAccount, WalletContextValue, WalletProviderConfig } from "../../src/wallet/types";
import { WalletError, WalletErrorCode } from "../../src/wallet/errors";
import {
  WALLET_ARBITRARY_SIGNING_SCOPE,
  WALLET_TRANSACTION_SIGNING_SCOPE,
  type WalletAuthorizationBinding,
  type WalletSigningAuthorizationRequest,
} from "../../src/wallet/session";

const adapters = vi.hoisted(() => {
  const createAdapter = (type: string, name: string) => ({
    type,
    name,
    isAvailable: vi.fn(() => true),
    connect: vi.fn(),
    disconnect: vi.fn(async () => undefined),
    getAccounts: vi.fn(),
    signAmino: vi.fn(),
    signDirect: vi.fn(),
    signArbitrary: vi.fn(),
  });
  return {
    keplr: createAdapter("keplr", "Keplr"),
    leap: createAdapter("leap", "Leap"),
    cosmostation: createAdapter("cosmostation", "Cosmostation"),
  };
});

vi.mock("../../src/wallet/adapters/keplr", () => ({
  KeplrAdapter: class {
    constructor() {
      return adapters.keplr;
    }
  },
}));
vi.mock("../../src/wallet/adapters/leap", () => ({
  LeapAdapter: class {
    constructor() {
      return adapters.leap;
    }
  },
}));
vi.mock("../../src/wallet/adapters/cosmostation", () => ({
  CosmostationAdapter: class {
    constructor() {
      return adapters.cosmostation;
    }
  },
}));
vi.mock("../../src/wallet/adapters/walletconnect", () => ({
  WalletConnectAdapter: class {},
}));

import { WalletProvider, useWallet } from "../../src/wallet/context";

const account = (address: string): WalletAccount => ({
  address,
  pubKey: new Uint8Array([1, 2, 3]),
  algo: "secp256k1",
});

const aminoSignDoc = {
  chain_id: "virtengine-1",
  account_number: "0",
  sequence: "0",
  fee: { gas: "1", amount: [] },
  msgs: [],
  memo: "",
};

const directSignDoc = {
  bodyBytes: new Uint8Array(),
  authInfoBytes: new Uint8Array(),
  chainId: "virtengine-1",
  accountNumber: 0,
};

const authorizationFor = (
  request: WalletSigningAuthorizationRequest,
  overrides: Partial<WalletAuthorizationBinding> = {},
): WalletAuthorizationBinding => {
  const now = Date.now();
  return {
    chainId: request.chainId,
    account: request.account,
    publicKey: request.publicKey,
    walletType: request.walletType,
    deviceId: "device-1",
    sessionId: "session-1",
    issuedAt: now - 1_000,
    expiresAt: now + 60_000,
    mfa: { scopes: [request.requiredScope], expiresAt: now + 30_000 },
    ...overrides,
  };
};

const chainInfo: WalletProviderConfig["chainInfo"] = {
  chainId: "virtengine-1",
  chainName: "VirtEngine",
  rpcEndpoint: "https://rpc.example.test",
  restEndpoint: "https://rest.example.test",
  bech32Config: {
    bech32PrefixAccAddr: "virtengine",
    bech32PrefixAccPub: "virtenginepub",
    bech32PrefixValAddr: "virtenginevaloper",
    bech32PrefixValPub: "virtenginevaloperpub",
    bech32PrefixConsAddr: "virtenginevalcons",
    bech32PrefixConsPub: "virtenginevalconspub",
  },
  stakeCurrency: {
    coinDenom: "VE",
    coinMinimalDenom: "uve",
    coinDecimals: 6,
  },
  currencies: [{ coinDenom: "VE", coinMinimalDenom: "uve", coinDecimals: 6 }],
  feeCurrencies: [{ coinDenom: "VE", coinMinimalDenom: "uve", coinDecimals: 6 }],
};

const storeReconnectMetadata = (persistKey: string, address: string): void => {
  const now = Date.now();
  localStorage.setItem(persistKey, JSON.stringify({
    version: 2,
    walletType: "keplr",
    address,
    chainId: chainInfo.chainId,
    connectedAt: now,
    lastActiveAt: now,
    expiresAt: now + 60_000,
    autoReconnect: true,
  }));
};

describe("WalletProvider session integration", () => {
  let container: HTMLDivElement;
  let root: Root;
  let wallet: WalletContextValue;

  function Probe(): JSX.Element | null {
    wallet = useWallet();
    return null;
  }

  async function renderProvider(config: Partial<WalletProviderConfig> = {}): Promise<void> {
    await act(async () => {
      root.render(
        <WalletProvider config={{ chainInfo, autoConnect: false, ...config }}>
          <Probe />
        </WalletProvider>,
      );
    });
  }

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    localStorage.clear();
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
    adapters.keplr.isAvailable.mockReturnValue(true);
    adapters.keplr.connect.mockReset();
    adapters.keplr.disconnect.mockReset().mockResolvedValue(undefined);
    adapters.keplr.getAccounts.mockReset();
    adapters.keplr.signAmino.mockReset();
    adapters.keplr.signDirect.mockReset();
    adapters.keplr.signArbitrary.mockReset();
  });

  afterEach(async () => {
    await act(async () => root.unmount());
    container.remove();
    vi.restoreAllMocks();
  });

  it("rejects unavailable and adapter failures with typed errors after setting UI error", async () => {
    await renderProvider();
    let rejected: unknown;
    await act(async () => {
      try {
        await wallet.connect("walletconnect");
      } catch (error) {
        rejected = error;
      }
    });
    expect(rejected).toMatchObject({
      code: WalletErrorCode.UNKNOWN,
      message: "Unsupported wallet type",
    });
    expect(wallet.error).toBeInstanceOf(WalletError);

    adapters.keplr.isAvailable.mockReturnValue(false);
    await act(async () => {
      try {
        await wallet.connect("keplr");
      } catch (error) {
        rejected = error;
      }
    });
    expect(rejected).toMatchObject({
      code: WalletErrorCode.WALLET_NOT_INSTALLED,
    });
    expect(wallet.error).toBeInstanceOf(WalletError);

    adapters.keplr.isAvailable.mockReturnValue(true);
    adapters.keplr.connect.mockRejectedValueOnce(new Error("request rejected"));
    await act(async () => {
      try {
        await wallet.connect("keplr");
      } catch (error) {
        rejected = error;
      }
    });
    expect(rejected).toMatchObject({
      code: WalletErrorCode.WALLET_CONNECTION_REJECTED,
    });
    expect(wallet.status).toBe("error");
  });

  it("persists only manager reconnect metadata after a verified account is returned", async () => {
    await renderProvider({ persistKey: "provider-session" });
    expect(localStorage.getItem("provider-session")).toBeNull();
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);

    await act(async () => wallet.connect("keplr"));

    const stored = JSON.parse(localStorage.getItem("provider-session") ?? "null");
    expect(Object.keys(stored).sort()).toEqual([
      "address",
      "autoReconnect",
      "chainId",
      "connectedAt",
      "expiresAt",
      "lastActiveAt",
      "version",
      "walletType",
    ]);
    expect(stored).toMatchObject({
      version: 2,
      walletType: "keplr",
      address: "virtengine1first",
      chainId: "virtengine-1",
    });
  });

  it("fails closed when auto-reconnect returns a different stored account", async () => {
    const onError = vi.fn();
    storeReconnectMetadata("provider-session", "virtengine1stored");
    adapters.keplr.connect.mockResolvedValue([account("virtengine1different")]);

    await renderProvider({ persistKey: "provider-session", autoConnect: true, onError });

    expect(adapters.keplr.connect).toHaveBeenCalledTimes(1);
    expect(wallet.status).toBe("error");
    expect(wallet.error).toBeInstanceOf(WalletError);
    expect(wallet.error).toMatchObject({ code: WalletErrorCode.SESSION_EXPIRED });
    expect(wallet.walletType).toBeNull();
    expect(wallet.accounts).toEqual([]);
    expect(onError).toHaveBeenCalledTimes(1);
    expect(localStorage.getItem("provider-session")).toBeNull();
    await expect(wallet.signDirect({
      bodyBytes: new Uint8Array(),
      authInfoBytes: new Uint8Array(),
      chainId: "virtengine-1",
      accountNumber: 0,
    })).rejects.toThrow("No wallet connected");

    await act(async () => root.render(
      <WalletProvider config={{ chainInfo, persistKey: "provider-session", autoConnect: true, onError }}>
        <Probe />
      </WalletProvider>,
    ));
    expect(adapters.keplr.connect).toHaveBeenCalledTimes(1);
  });

  it("auto-reconnects once when the stored binding matches", async () => {
    storeReconnectMetadata("provider-session", "virtengine1stored");
    adapters.keplr.connect.mockResolvedValue([account("virtengine1stored")]);

    await renderProvider({ persistKey: "provider-session", autoConnect: true });

    expect(adapters.keplr.connect).toHaveBeenCalledTimes(1);
    expect(wallet.status).toBe("connected");
    expect(wallet.walletType).toBe("keplr");
    expect(wallet.chainId).toBe("virtengine-1");
    expect(wallet.accounts[0]?.address).toBe("virtengine1stored");
    expect(JSON.parse(localStorage.getItem("provider-session") ?? "null").address).toBe(
      "virtengine1stored",
    );

    await act(async () => root.render(
      <WalletProvider config={{ chainInfo, persistKey: "provider-session", autoConnect: true }}>
        <Probe />
      </WalletProvider>,
    ));
    expect(adapters.keplr.connect).toHaveBeenCalledTimes(1);
  });

  it("clears provider and manager state even when adapter disconnect fails", async () => {
    await renderProvider({ persistKey: "provider-session" });
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
    await act(async () => wallet.connect("keplr"));
    adapters.keplr.disconnect.mockRejectedValueOnce(new Error("disconnect failed"));

    let rejected: unknown;
    await act(async () => {
      try {
        await wallet.disconnect();
      } catch (error) {
        rejected = error;
      }
    });

    expect(rejected).toMatchObject({ message: "disconnect failed" });
    expect(wallet.status).toBe("idle");
    expect(wallet.accounts).toEqual([]);
    expect(localStorage.getItem("provider-session")).toBeNull();
  });

  it("updates verified reconnect metadata after account refresh and selection", async () => {
    await renderProvider({ persistKey: "provider-session" });
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
    await act(async () => wallet.connect("keplr"));
    adapters.keplr.getAccounts.mockResolvedValueOnce([
      account("virtengine1first"),
      account("virtengine1second"),
    ]);

    await act(async () => wallet.refreshAccounts());
    act(() => wallet.selectAccount(1));

    expect(JSON.parse(localStorage.getItem("provider-session") ?? "null").address).toBe(
      "virtengine1second",
    );
  });

  it("fails all signing methods closed when no live authorization authority is configured", async () => {
    await renderProvider();
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
    await act(async () => wallet.connect("keplr"));

    const attempts = [
      () => wallet.signAmino(aminoSignDoc),
      () => wallet.signDirect(directSignDoc),
      () => wallet.signArbitrary("proof"),
    ];
    for (const attempt of attempts) {
      await expect(attempt()).rejects.toMatchObject({ code: WalletErrorCode.SESSION_EXPIRED });
    }

    expect(adapters.keplr.signAmino).not.toHaveBeenCalled();
    expect(adapters.keplr.signDirect).not.toHaveBeenCalled();
    expect(adapters.keplr.signArbitrary).not.toHaveBeenCalled();
  });

  it.each(["mismatched", "expired"] as const)(
    "does not invoke any signing adapter with %s authorization",
    async (kind) => {
      const authorize = vi.fn(async (request: WalletSigningAuthorizationRequest) =>
        authorizationFor(request, kind === "mismatched"
          ? { account: "virtengine1other" }
          : {
              issuedAt: Date.now() - 2_000,
              expiresAt: Date.now() - 1_000,
              mfa: { scopes: [request.requiredScope], expiresAt: Date.now() - 1_000 },
            }),
      );
      await renderProvider({ signingAuthorization: { authorize } });
      adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
      await act(async () => wallet.connect("keplr"));

      await expect(wallet.signAmino(aminoSignDoc)).rejects.toMatchObject({
        code: WalletErrorCode.SESSION_EXPIRED,
      });
      await expect(wallet.signDirect(directSignDoc)).rejects.toMatchObject({
        code: WalletErrorCode.SESSION_EXPIRED,
      });
      await expect(wallet.signArbitrary("proof")).rejects.toMatchObject({
        code: WalletErrorCode.SESSION_EXPIRED,
      });

      expect(adapters.keplr.signAmino).not.toHaveBeenCalled();
      expect(adapters.keplr.signDirect).not.toHaveBeenCalled();
      expect(adapters.keplr.signArbitrary).not.toHaveBeenCalled();
    },
  );

  it("invokes each adapter once with an exact live operation-scoped binding", async () => {
    const authorize = vi.fn(async (request: WalletSigningAuthorizationRequest) =>
      authorizationFor(request),
    );
    adapters.keplr.signAmino.mockResolvedValueOnce({ signed: aminoSignDoc, signature: {} });
    adapters.keplr.signDirect.mockResolvedValueOnce({ signed: directSignDoc, signature: {} });
    adapters.keplr.signArbitrary.mockResolvedValueOnce({
      signature: "signature",
      pubKey: new Uint8Array([1, 2, 3]),
    });
    await renderProvider({ persistKey: "provider-session", signingAuthorization: { authorize } });
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
    await act(async () => wallet.connect("keplr"));

    await wallet.signAmino(aminoSignDoc);
    await wallet.signDirect(directSignDoc);
    await wallet.signArbitrary("proof");

    expect(adapters.keplr.signAmino).toHaveBeenCalledTimes(1);
    expect(adapters.keplr.signDirect).toHaveBeenCalledTimes(1);
    expect(adapters.keplr.signArbitrary).toHaveBeenCalledTimes(1);
    expect(authorize.mock.calls.map(([request]) => [request.operation, request.requiredScope])).toEqual([
      ["amino", WALLET_TRANSACTION_SIGNING_SCOPE],
      ["direct", WALLET_TRANSACTION_SIGNING_SCOPE],
      ["arbitrary", WALLET_ARBITRARY_SIGNING_SCOPE],
    ]);
    expect(localStorage.getItem("provider-session")).not.toContain("device-1");
    expect(localStorage.getItem("provider-session")).not.toContain("session-1");
  });

  it.each(["account", "public key"] as const)(
    "invalidates authorization when the active %s changes",
    async (changedIdentity) => {
      let originalBinding: WalletAuthorizationBinding | null = null;
      const authorize = vi.fn(async (request: WalletSigningAuthorizationRequest) => {
        originalBinding ??= authorizationFor(request);
        return originalBinding;
      });
      await renderProvider({ signingAuthorization: { authorize } });
      adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
      await act(async () => wallet.connect("keplr"));

      await wallet.signDirect(directSignDoc);
      adapters.keplr.signDirect.mockClear();
      if (changedIdentity === "account") {
        adapters.keplr.getAccounts.mockResolvedValueOnce([
          account("virtengine1first"),
          account("virtengine1second"),
        ]);
        await act(async () => wallet.refreshAccounts());
        act(() => wallet.selectAccount(1));
      } else {
        adapters.keplr.getAccounts.mockResolvedValueOnce([{
          ...account("virtengine1first"),
          pubKey: new Uint8Array([4, 5, 6]),
        }]);
        await act(async () => wallet.refreshAccounts());
      }

      await expect(wallet.signDirect(directSignDoc)).rejects.toMatchObject({
        code: WalletErrorCode.SESSION_EXPIRED,
      });
      expect(adapters.keplr.signDirect).not.toHaveBeenCalled();
    },
  );

  it("does not let arbitrary signing reuse transaction MFA authorization", async () => {
    const authorize = vi.fn(async (request: WalletSigningAuthorizationRequest) =>
      authorizationFor(request, {
        mfa: { scopes: [WALLET_TRANSACTION_SIGNING_SCOPE], expiresAt: Date.now() + 30_000 },
      }),
    );
    await renderProvider({ signingAuthorization: { authorize } });
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
    await act(async () => wallet.connect("keplr"));

    await expect(wallet.signArbitrary("proof")).rejects.toMatchObject({
      code: WalletErrorCode.SESSION_EXPIRED,
    });
    expect(authorize).toHaveBeenCalledWith(expect.objectContaining({
      operation: "arbitrary",
      requiredScope: WALLET_ARBITRARY_SIGNING_SCOPE,
    }));
    expect(adapters.keplr.signArbitrary).not.toHaveBeenCalled();
  });

  it("immediately clears provider state and signing ability on cross-tab removal", async () => {
    await renderProvider({ persistKey: "provider-session" });
    adapters.keplr.connect.mockResolvedValueOnce([account("virtengine1first")]);
    await act(async () => wallet.connect("keplr"));

    await act(async () => {
      window.dispatchEvent(
        new StorageEvent("storage", { key: "provider-session", newValue: null }),
      );
    });

    expect(wallet.status).toBe("idle");
    expect(wallet.walletType).toBeNull();
    await expect(wallet.signDirect({
      bodyBytes: new Uint8Array(),
      authInfoBytes: new Uint8Array(),
      chainId: "virtengine-1",
      accountNumber: 0,
    })).rejects.toThrow("No wallet connected");
  });
});