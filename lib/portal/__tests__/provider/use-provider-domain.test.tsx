import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ProviderDomainVerificationError,
  type ProviderDomainChallenge,
  type ProviderDomainVerifier,
  validateProviderDomainVerification,
} from "../../components/provider/domain-verification";
import { ProviderProvider, useProvider } from "../../hooks/useProvider";
import type { QueryClient } from "../../types/chain";

const binding = {
  chainId: "virtengine-1",
  accountAddress: "virtengine1provider",
} as const;
const now = Date.now();

function verifier(
  overrides: Partial<ProviderDomainVerifier> = {},
): ProviderDomainVerifier {
  return {
    ...binding,
    issueChallenge: vi.fn((domain, method) =>
      Promise.resolve({
        ...binding,
        challengeId: "challenge-1",
        domain,
        method,
        challengeValue: "ve-verification=authoritative",
        dnsRecordName:
          method === "dns_txt" ? `_virtengine.${domain}` : undefined,
        httpFilePath:
          method === "http_file"
            ? "/.well-known/virtengine-verification"
            : undefined,
        expiresAt: now + 60_000,
        instructions: "Publish the authoritative challenge",
      }),
    ),
    verifyChallenge: vi.fn((challenge: ProviderDomainChallenge) =>
      Promise.resolve({
        ...binding,
        challengeId: challenge.challengeId,
        evidenceId: "evidence-1",
        domain: challenge.domain,
        method: challenge.method,
        status: "verified",
        verifiedAt: now,
        expiresAt: now + 86_400_000,
      }),
    ),
    ...overrides,
  };
}

describe("provider domain verification authority", () => {
  let container: HTMLDivElement;
  let root: Root;
  let provider: ReturnType<typeof useProvider>;

  const Consumer = () => {
    provider = useProvider();
    return null;
  };

  const renderProvider = async (
    domainVerifier?: ProviderDomainVerifier,
    accountAddress: string | null = binding.accountAddress,
  ) => {
    await act(async () =>
      root.render(
        <ProviderProvider
          queryClient={
            {
              queryProvider: vi.fn().mockResolvedValue(null),
            } as unknown as QueryClient
          }
          chainId={binding.chainId}
          accountAddress={accountAddress}
          domainVerifier={domainVerifier}
        >
          <Consumer />
        </ProviderProvider>,
      ),
    );
  };

  const startRegistration = async () => {
    await act(async () => provider.actions.startRegistration());
    await act(async () =>
      provider.actions.updateRegistrationData({
        primaryDomain: "provider.example.com",
      }),
    );
  };

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(async () => {
    await act(async () => root.unmount());
    container.remove();
  });

  it("fails closed without exact verifier authority or a challenge", async () => {
    await renderProvider();
    await expect(
      provider.actions.startDomainVerification(
        "provider.example.com",
        "dns_txt",
      ),
    ).rejects.toMatchObject({ code: "feature_unavailable" });
    expect(provider.state.domainVerifications).toEqual([]);

    const verifyChallenge = vi.fn();
    await renderProvider(
      verifier({ accountAddress: "virtengine1other", verifyChallenge }),
    );
    await expect(
      provider.actions.checkDomainVerification("provider.example.com"),
    ).rejects.toBeInstanceOf(ProviderDomainVerificationError);
    expect(verifyChallenge).not.toHaveBeenCalled();
  });

  it("rejects registration actions while disconnected", async () => {
    await renderProvider(undefined, null);
    expect(() => provider.actions.startRegistration()).toThrow(
      ProviderDomainVerificationError,
    );
    expect(() =>
      provider.actions.updateRegistrationData({
        primaryDomain: "provider.example.com",
      }),
    ).toThrow(ProviderDomainVerificationError);
    await expect(provider.actions.submitRegistration()).rejects.toMatchObject({
      code: "authority_changed",
    });
    expect(provider.state.registration).toBeNull();
    expect(provider.state.isRegistered).toBe(false);
  });

  it("advances registration only after exact authoritative evidence", async () => {
    const domainVerifier = verifier();
    await renderProvider(domainVerifier);
    await startRegistration();
    let challenge!: Awaited<
      ReturnType<typeof provider.actions.startDomainVerification>
    >;
    await act(async () => {
      challenge = await provider.actions.startDomainVerification(
        "Provider.Example.com.",
        "dns_txt",
      );
    });
    expect(challenge.domain).toBe("provider.example.com");
    expect(Object.isFrozen(challenge)).toBe(true);

    let verification!: Awaited<
      ReturnType<typeof provider.actions.checkDomainVerification>
    >;
    await act(async () => {
      verification = await provider.actions.checkDomainVerification(
        "provider.example.com",
      );
    });
    expect(verification).toMatchObject({
      domain: "provider.example.com",
      status: "verified",
    });
    expect(Object.isFrozen(verification)).toBe(true);
    expect(provider.state.registration).toMatchObject({
      domainVerified: true,
      step: "stake_deposit",
    });
    expect(provider.state.domainVerifications).toHaveLength(1);
    await expect(
      provider.actions.checkDomainVerification("provider.example.com"),
    ).rejects.toMatchObject({ code: "invalid_challenge" });
  });

  it("rejects malformed evidence without changing registration", async () => {
    const domainVerifier = verifier({
      verifyChallenge: vi.fn().mockResolvedValue({
        ...binding,
        challengeId: "other",
        evidenceId: "evidence-1",
        domain: "provider.example.com",
        method: "dns_txt",
        status: "verified",
        verifiedAt: now,
        expiresAt: now + 86_400_000,
      }),
    });
    await renderProvider(domainVerifier);
    await startRegistration();
    await act(async () => {
      await provider.actions.startDomainVerification(
        "provider.example.com",
        "dns_txt",
      );
    });

    await expect(
      provider.actions.checkDomainVerification("provider.example.com"),
    ).rejects.toMatchObject({ code: "invalid_verification" });
    expect(provider.state.registration?.domainVerified).toBe(false);
    expect(provider.state.domainVerifications).toEqual([]);
  });

  it("blocks duplicate verification and rejects late evidence after authority changes", async () => {
    let resolveVerification!: (value: unknown) => void;
    const verifyChallenge = vi.fn(
      () => new Promise((resolve) => (resolveVerification = resolve)),
    );
    const first = verifier({ verifyChallenge });
    await renderProvider(first);
    await startRegistration();
    await act(async () => {
      await provider.actions.startDomainVerification(
        "provider.example.com",
        "dns_txt",
      );
    });
    let pending!: Promise<unknown>;
    await act(async () => {
      pending = provider.actions
        .checkDomainVerification("provider.example.com")
        .catch((error) => error);
      await Promise.resolve();
    });
    await expect(
      provider.actions.checkDomainVerification("provider.example.com"),
    ).rejects.toMatchObject({ code: "verification_in_progress" });

    await renderProvider(verifier(), null);
    resolveVerification({
      ...binding,
      challengeId: "challenge-1",
      evidenceId: "evidence-1",
      domain: "provider.example.com",
      method: "dns_txt",
      status: "verified",
      verifiedAt: now,
      expiresAt: now + 86_400_000,
    });

    await expect(pending).resolves.toMatchObject({ code: "authority_changed" });
    expect(provider.state.domainVerifications).toEqual([]);
  });

  it("clears account-scoped verification state on account change", async () => {
    await renderProvider(verifier());
    await startRegistration();
    await act(async () => {
      await provider.actions.startDomainVerification(
        "provider.example.com",
        "dns_txt",
      );
      await provider.actions.checkDomainVerification("provider.example.com");
    });
    expect(provider.state.domainVerifications).toHaveLength(1);

    await renderProvider(
      verifier({ accountAddress: "virtengine1other" }),
      "virtengine1other",
    );
    expect(provider.state.domainVerifications).toEqual([]);
    expect(provider.state.registration).toBeNull();
  });

  it("invalidates challenges when the registration primary domain changes", async () => {
    await renderProvider(verifier());
    await startRegistration();
    await act(async () => {
      await provider.actions.startDomainVerification(
        "provider.example.com",
        "dns_txt",
      );
    });
    await act(async () =>
      provider.actions.updateRegistrationData({
        primaryDomain: "other.example.com",
      }),
    );
    await act(async () =>
      provider.actions.updateRegistrationData({
        primaryDomain: "provider.example.com",
      }),
    );

    await expect(
      provider.actions.checkDomainVerification("provider.example.com"),
    ).rejects.toMatchObject({ code: "invalid_challenge" });
  });

  it("blocks duplicate challenge issuance", async () => {
    let resolveChallenge!: (value: unknown) => void;
    const issueChallenge = vi.fn(
      () => new Promise((resolve) => (resolveChallenge = resolve)),
    );
    await renderProvider(verifier({ issueChallenge }));
    await startRegistration();
    let pending!: Promise<unknown>;
    await act(async () => {
      pending = provider.actions
        .startDomainVerification("provider.example.com", "dns_txt")
        .catch((error) => error);
      await Promise.resolve();
    });
    await expect(
      provider.actions.startDomainVerification(
        "provider.example.com",
        "dns_txt",
      ),
    ).rejects.toMatchObject({ code: "challenge_in_progress" });
    await act(async () =>
      resolveChallenge({
        ...binding,
        challengeId: "challenge-1",
        domain: "provider.example.com",
        method: "dns_txt",
        challengeValue: "authoritative",
        dnsRecordName: "_virtengine.provider.example.com",
        expiresAt: now + 60_000,
        instructions: "Publish challenge",
      }),
    );
    await expect(pending).resolves.toMatchObject({
      challengeId: "challenge-1",
    });
  });

  it("rejects expired challenges in the public verification validator", () => {
    const challenge: ProviderDomainChallenge = {
      ...binding,
      challengeId: "challenge-1",
      domain: "provider.example.com",
      method: "dns_txt",
      challengeValue: "authoritative",
      dnsRecordName: "_virtengine.provider.example.com",
      expiresAt: now - 1,
      instructions: "Expired",
    };
    expect(() =>
      validateProviderDomainVerification(
        {
          ...binding,
          challengeId: challenge.challengeId,
          evidenceId: "evidence-1",
          domain: challenge.domain,
          method: challenge.method,
          status: "verified",
          verifiedAt: now - 2,
          expiresAt: now + 60_000,
        },
        binding,
        challenge,
        now,
      ),
    ).toThrow(ProviderDomainVerificationError);
  });

  it("ignores a stale provider query after account change", async () => {
    let resolveProvider!: (value: unknown) => void;
    const queryProvider = vi.fn(
      () => new Promise((resolve) => (resolveProvider = resolve)),
    );
    await act(async () =>
      root.render(
        <ProviderProvider
          queryClient={{ queryProvider } as unknown as QueryClient}
          chainId={binding.chainId}
          accountAddress={binding.accountAddress}
          domainVerifier={verifier()}
        >
          <Consumer />
        </ProviderProvider>,
      ),
    );
    await renderProvider(
      verifier({ accountAddress: "virtengine1other" }),
      "virtengine1other",
    );
    await act(async () =>
      resolveProvider({
        address: binding.accountAddress,
        status: "active",
        reliabilityScore: 99,
        registeredAt: now,
      }),
    );
    expect(provider.state.profile).toBeNull();
  });

  it("rejects retained registration mutations after account change", async () => {
    await renderProvider(verifier());
    const staleStart = provider.actions.startRegistration;
    const staleUpdate = provider.actions.updateRegistrationData;
    const staleSubmit = provider.actions.submitRegistration;
    const staleCreateOffering = provider.actions.createOffering;
    const staleGetIncomingOrders = provider.actions.getIncomingOrders;

    await renderProvider(
      verifier({ accountAddress: "virtengine1other" }),
      "virtengine1other",
    );

    expect(() => staleStart()).toThrow(ProviderDomainVerificationError);
    expect(() => staleUpdate({ primaryDomain: "stale.example.com" })).toThrow(
      ProviderDomainVerificationError,
    );
    await expect(staleSubmit()).rejects.toMatchObject({
      code: "authority_changed",
    });
    await expect(
      staleCreateOffering({
        title: "Stale offering",
        type: "compute",
        autoPublish: false,
      } as Parameters<typeof staleCreateOffering>[0]),
    ).rejects.toMatchObject({ code: "authority_changed" });
    await expect(staleGetIncomingOrders()).rejects.toMatchObject({
      code: "authority_changed",
    });
    expect(provider.state.registration).toBeNull();
    expect(provider.state.isRegistered).toBe(false);
    expect(provider.state.offerings).toEqual([]);
  });
});
