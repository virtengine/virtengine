import { describe, expect, it, vi } from "vitest";
import {
  ActionGate,
  type ActionGateAuthorities,
  type ActionPreview,
  type CapabilityDecision,
  type ExecuteActionRequest,
  type GatedAction,
  type PolicyDecision,
} from "../../src/chat/action-gate";

const NOW = 1_800_000_000_000;
const action: GatedAction = {
  actionId: "action-1",
  toolName: "delete-deployments",
  requiredCapability: "deployment.delete",
  payload: { deploymentIds: ["d-1"] },
};

const policy = (overrides: Partial<PolicyDecision> = {}): PolicyDecision => ({
  actionId: action.actionId,
  toolName: action.toolName,
  allowed: true,
  expiresAt: NOW + 120_000,
  ...overrides,
});

const capability = (
  overrides: Partial<CapabilityDecision> = {},
): CapabilityDecision => ({
  actionId: action.actionId,
  toolName: action.toolName,
  capability: action.requiredCapability,
  allowed: true,
  expiresAt: NOW + 120_000,
  ...overrides,
});

const createAuthorities = (
  overrides: Partial<ActionGateAuthorities> = {},
): ActionGateAuthorities => ({
  decidePolicy: vi.fn(async () => policy()),
  decideCapability: vi.fn(async () => capability()),
  simulate: vi.fn(async () => ({
    actionId: action.actionId,
    toolName: action.toolName,
    stateDigest: "state-1",
    impact: { affected: ["d-1"], operation: "delete" },
  })),
  readCurrentStateDigest: vi.fn(async () => "state-1"),
  ...overrides,
});

const expectPreview = async (gate: ActionGate): Promise<ActionPreview> => {
  const result = await gate.prepare(action);
  expect(result.status).toBe("preview");
  if (result.status !== "preview") throw new Error(result.reason);
  return result.preview;
};

const validRequest = (preview: ActionPreview): ExecuteActionRequest => ({
  preview,
  confirmation: {
    confirmed: true,
    previewDigest: preview.previewDigest,
    nonce: preview.nonce,
    confirmedAt: NOW,
    expiresAt: NOW + 30_000,
  },
  signer: {
    actionId: action.actionId,
    toolName: action.toolName,
    previewDigest: preview.previewDigest,
    stateDigest: preview.stateDigest,
    walletAddress: "virt1wallet",
    accountId: "account-1",
    chainId: "virt-1",
    signature: "signed-evidence",
    expiresAt: NOW + 30_000,
    mfa: { scope: "deployment.delete", expiresAt: NOW + 30_000 },
  },
  context: {
    walletAddress: "virt1wallet",
    accountId: "account-1",
    chainId: "virt-1",
    mfaScope: "deployment.delete",
  },
});

const expectDenial = async (
  result: Promise<{ status: string; reason?: string }>,
  reason: string,
) => {
  await expect(result).resolves.toMatchObject({ status: "denied", reason });
};

describe("ActionGate preparation", () => {
  it.each([
    ["policy_missing", { decidePolicy: vi.fn(async () => undefined) }],
    [
      "policy_denied",
      { decidePolicy: vi.fn(async () => policy({ allowed: false })) },
    ],
    [
      "policy_mismatch",
      { decidePolicy: vi.fn(async () => policy({ toolName: "other" })) },
    ],
    [
      "policy_expired",
      { decidePolicy: vi.fn(async () => policy({ expiresAt: NOW })) },
    ],
    ["capability_missing", { decideCapability: vi.fn(async () => undefined) }],
    [
      "capability_denied",
      { decideCapability: vi.fn(async () => capability({ allowed: false })) },
    ],
    [
      "capability_mismatch",
      {
        decideCapability: vi.fn(async () =>
          capability({ capability: "other" }),
        ),
      },
    ],
    [
      "capability_expired",
      { decideCapability: vi.fn(async () => capability({ expiresAt: NOW })) },
    ],
  ])("denies %s before simulation", async (reason, overrides) => {
    const authorities = createAuthorities(overrides);
    const gate = new ActionGate(authorities, { now: () => NOW });

    await expectDenial(gate.prepare(action), reason);
    expect(authorities.simulate).not.toHaveBeenCalled();
  });

  it.each([
    [
      "simulation_failed",
      vi.fn(async () => Promise.reject(new Error("offline"))),
    ],
    [
      "simulation_mismatch",
      vi.fn(async () => ({
        actionId: "other",
        toolName: action.toolName,
        stateDigest: "state-1",
        impact: {},
      })),
    ],
    [
      "simulation_invalid",
      vi.fn(async () => ({
        actionId: action.actionId,
        toolName: action.toolName,
        stateDigest: "",
        impact: {},
      })),
    ],
  ])("denies %s without creating a preview", async (reason, simulate) => {
    const gate = new ActionGate(createAuthorities({ simulate }), {
      now: () => NOW,
    });
    await expectDenial(gate.prepare(action), reason);
  });

  it("returns only a canonical preview and never invokes model-attached execution", async () => {
    const modelExecutor = vi.fn();
    const modelSelectedTool = {
      run: vi.fn(async () => action),
      execute: modelExecutor,
    };
    const gate = new ActionGate(createAuthorities(), { now: () => NOW });

    const result = await gate.prepare(await modelSelectedTool.run());

    expect(result.status).toBe("preview");
    expect(result).not.toHaveProperty("ok");
    expect(result).not.toHaveProperty("result");
    expect(modelExecutor).not.toHaveBeenCalled();
  });
});

describe("ActionGate execution evidence", () => {
  it.each([
    [
      "confirmation_missing",
      (request: ExecuteActionRequest) => ({
        ...request,
        confirmation: undefined,
      }),
    ],
    [
      "confirmation_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        confirmation: { ...request.confirmation!, previewDigest: "wrong" },
      }),
    ],
    [
      "confirmation_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        confirmation: { ...request.confirmation!, nonce: "wrong" },
      }),
    ],
    [
      "confirmation_expired",
      (request: ExecuteActionRequest) => ({
        ...request,
        confirmation: { ...request.confirmation!, expiresAt: NOW },
      }),
    ],
    [
      "signer_missing",
      (request: ExecuteActionRequest) => ({ ...request, signer: undefined }),
    ],
    [
      "signer_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, actionId: "wrong" },
      }),
    ],
    [
      "signer_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, previewDigest: "wrong" },
      }),
    ],
    [
      "signer_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, stateDigest: "wrong" },
      }),
    ],
    [
      "signer_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, walletAddress: "wrong" },
      }),
    ],
    [
      "signer_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, accountId: "wrong" },
      }),
    ],
    [
      "signer_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, chainId: "wrong" },
      }),
    ],
    [
      "signer_expired",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, expiresAt: NOW },
      }),
    ],
    [
      "signature_missing",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, signature: " " },
      }),
    ],
    [
      "mfa_missing",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: { ...request.signer!, mfa: undefined },
      }),
    ],
    [
      "mfa_scope_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: {
          ...request.signer!,
          mfa: { ...request.signer!.mfa!, scope: "other" },
        },
      }),
    ],
    [
      "mfa_scope_mismatch",
      (request: ExecuteActionRequest) => ({
        ...request,
        context: { ...request.context, mfaScope: "deployment.read" },
        signer: {
          ...request.signer!,
          mfa: { ...request.signer!.mfa!, scope: "deployment.read" },
        },
      }),
    ],
    [
      "mfa_expired",
      (request: ExecuteActionRequest) => ({
        ...request,
        signer: {
          ...request.signer!,
          mfa: { ...request.signer!.mfa!, expiresAt: NOW },
        },
      }),
    ],
  ])("denies %s without calling the executor", async (reason, mutate) => {
    const gate = new ActionGate(createAuthorities(), { now: () => NOW });
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => ({ transactionHash: "tx-1" }));

    await expectDenial(
      gate.execute(mutate(validRequest(preview)), executor),
      reason,
    );
    expect(executor).not.toHaveBeenCalled();
  });

  it("rejects unknown, tampered, and expired previews", async () => {
    let now = NOW;
    const gate = new ActionGate(createAuthorities(), { now: () => now });
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => "executed");

    await expectDenial(
      gate.execute(
        { ...validRequest(preview), preview: { ...preview, nonce: "unknown" } },
        executor,
      ),
      "preview_unknown",
    );
    await expectDenial(
      gate.execute(
        {
          ...validRequest(preview),
          preview: { ...preview, stateDigest: "tampered" },
        },
        executor,
      ),
      "preview_digest_mismatch",
    );
    now = preview.expiresAt;
    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "preview_expired",
    );
    expect(executor).not.toHaveBeenCalled();
  });

  it("consumes drifted evidence and never calls the executor", async () => {
    const gate = new ActionGate(
      createAuthorities({
        readCurrentStateDigest: vi.fn(async () => "state-2"),
      }),
      { now: () => NOW },
    );
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => "executed");

    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "state_drift",
    );
    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "nonce_replayed",
    );
    expect(executor).not.toHaveBeenCalled();
  });

  it("consumes evidence when authoritative state cannot be re-read", async () => {
    const gate = new ActionGate(
      createAuthorities({
        readCurrentStateDigest: vi.fn(async () => {
          throw new Error("unavailable");
        }),
      }),
      { now: () => NOW },
    );
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => "executed");

    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "authoritative_state_unavailable",
    );
    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "nonce_replayed",
    );
    expect(executor).not.toHaveBeenCalled();
  });

  it("calls the final injected executor exactly once for fully valid evidence", async () => {
    const gate = new ActionGate(createAuthorities(), { now: () => NOW });
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => ({ transactionHash: "tx-1" }));

    await expect(
      gate.execute(validRequest(preview), executor),
    ).resolves.toEqual({
      status: "executed",
      result: { transactionHash: "tx-1" },
    });
    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "nonce_replayed",
    );
    expect(executor).toHaveBeenCalledTimes(1);
    expect(executor.mock.calls[0][0]).toMatchObject({
      action,
      previewDigest: preview.previewDigest,
      stateDigest: preview.stateDigest,
      nonce: preview.nonce,
    });
  });

  it("consumes evidence when the final executor fails", async () => {
    const gate = new ActionGate(createAuthorities(), { now: () => NOW });
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => {
      throw new Error("failed");
    });

    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "execution_failed",
    );
    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "nonce_replayed",
    );
    expect(executor).toHaveBeenCalledTimes(1);
  });

  it("single-flights concurrent use of one confirmation", async () => {
    let releaseStateRead: ((digest: string) => void) | undefined;
    const stateRead = new Promise<string>((resolve) => {
      releaseStateRead = resolve;
    });
    const readCurrentStateDigest = vi.fn(() => stateRead);
    const gate = new ActionGate(createAuthorities({ readCurrentStateDigest }), {
      now: () => NOW,
    });
    const preview = await expectPreview(gate);
    const executor = vi.fn(async () => "executed");

    const first = gate.execute(validRequest(preview), executor);
    await vi.waitFor(() =>
      expect(readCurrentStateDigest).toHaveBeenCalledOnce(),
    );
    await expectDenial(
      gate.execute(validRequest(preview), executor),
      "nonce_in_use",
    );
    releaseStateRead?.("state-1");
    await expect(first).resolves.toMatchObject({ status: "executed" });
    expect(executor).toHaveBeenCalledTimes(1);
  });
});
