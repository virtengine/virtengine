import { describe, expect, it, jest } from "@jest/globals";
import { mock } from "jest-mock-extended";

import type { TxClient } from "../transport/tx/TxClient.ts";
import {
  ChainCapability,
  ChainCapabilityController,
  ChainCapabilityError,
  ChainCapabilityErrorReason,
} from "./ChainCapability.ts";

describe(ChainCapabilityController.name, () => {
  const validAuthorization = () => ({
    authorizationId: "mfa-authorization-1",
    expiresAt: Date.now() + 60_000,
  });

  it("starts query-only without a signer and signing-ready with one", () => {
    expect(new ChainCapabilityController().state).toBe(ChainCapability.QueryOnly);
    expect(new ChainCapabilityController(mock<TxClient>()).state).toBe(ChainCapability.SigningReady);
  });

  it("transitions through every capability state", () => {
    const controller = new ChainCapabilityController();
    const signer = mock<TxClient>();

    expect(controller.state).toBe(ChainCapability.QueryOnly);
    controller.setSigner(signer);
    expect(controller.state).toBe(ChainCapability.SigningReady);
    controller.authorizeMFA(validAuthorization());
    expect(controller.state).toBe(ChainCapability.MfaAuthorized);
    controller.revokeMFA();
    expect(controller.state).toBe(ChainCapability.SigningReady);
    controller.clearSigner();
    expect(controller.state).toBe(ChainCapability.QueryOnly);
    controller.disconnect();
    expect(controller.state).toBe(ChainCapability.Disconnected);
    controller.connect();
    expect(controller.state).toBe(ChainCapability.QueryOnly);
  });

  it("clears signing authority when disconnected", () => {
    const controller = new ChainCapabilityController(mock<TxClient>());

    controller.disconnect();
    controller.connect();

    expect(() => controller.requireSigner("createProvider")).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.SigningRequired,
      currentCapability: ChainCapability.QueryOnly,
      requiredCapability: ChainCapability.SigningReady,
      operation: "createProvider",
    }));
  });

  it("rejects invalid MFA authorization with a typed capability error", () => {
    const controller = new ChainCapabilityController();

    expect(() => controller.authorizeMFA(validAuthorization())).toThrow(ChainCapabilityError);
    expect(() => controller.authorizeMFA(validAuthorization())).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.InvalidTransition,
      currentCapability: ChainCapability.QueryOnly,
      requiredCapability: ChainCapability.SigningReady,
    }));
  });

  it.each([
    ["missing authorization ID", { authorizationId: "", expiresAt: Date.now() + 60_000 }],
    ["expired authorization", { authorizationId: "mfa-authorization-1", expiresAt: Date.now() - 1 }],
    ["invalid expiry", { authorizationId: "mfa-authorization-1", expiresAt: Number.NaN }],
  ])("rejects %s metadata", (_name, metadata) => {
    const controller = new ChainCapabilityController(mock<TxClient>());

    expect(() => controller.authorizeMFA(metadata)).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.InvalidMfaAuthorization,
      currentCapability: ChainCapability.SigningReady,
      requiredCapability: ChainCapability.MfaAuthorized,
      operation: "authorizeMFA",
    }));
    expect(controller.state).toBe(ChainCapability.SigningReady);
  });

  it("rejects MFA use after authorization expires", () => {
    const controller = new ChainCapabilityController(mock<TxClient>());
    const now = Date.now();
    const dateNow = jest.spyOn(Date, "now").mockReturnValueOnce(now).mockReturnValueOnce(now + 2);
    controller.authorizeMFA({ authorizationId: "mfa-authorization-1", expiresAt: now + 1 });

    expect(() => controller.assertMFAAuthorized("submitBid")).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.MfaRequired,
      currentCapability: ChainCapability.SigningReady,
    }));
    expect(controller.state).toBe(ChainCapability.SigningReady);
    dateNow.mockRestore();
  });

  it("replaces and clears signers while revoking MFA authorization", () => {
    const firstSigner = mock<TxClient>();
    const secondSigner = mock<TxClient>();
    const controller = new ChainCapabilityController(firstSigner);
    controller.authorizeMFA(validAuthorization());

    controller.setSigner(secondSigner);
    expect(controller.state).toBe(ChainCapability.SigningReady);
    expect(controller.requireSigner("createProvider")).toBe(secondSigner);
    expect(() => controller.assertMFAAuthorized("createProvider")).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.MfaRequired,
    }));

    controller.clearSigner();
    expect(controller.state).toBe(ChainCapability.QueryOnly);
    expect(() => controller.requireSigner("createProvider")).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.SigningRequired,
    }));
  });

  it("rejects every operation while disconnected", () => {
    const controller = new ChainCapabilityController(mock<TxClient>());
    controller.disconnect();

    expect(() => controller.assertCanQuery("getNodeInfo")).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.Disconnected,
    }));
    expect(() => controller.requireSigner("createProvider")).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.Disconnected,
    }));
    expect(() => controller.setSigner(mock<TxClient>())).toThrow(expect.objectContaining({
      reason: ChainCapabilityErrorReason.InvalidTransition,
    }));
  });
});
