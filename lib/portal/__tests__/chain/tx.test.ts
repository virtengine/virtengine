import { describe, it, expect } from "vitest";
import {
  buildMsgCreateOrder,
  buildMsgCloseOrder,
  buildMsgCreateBid,
} from "../../src/chain/tx/market";
import {
  buildMsgUploadScope,
  buildMsgRequestVerification,
} from "../../src/chain/tx/veid";

describe("chain tx builders", () => {
  it("builds market order messages", () => {
    const msg = buildMsgCreateOrder({
      cpu: 4,
      memoryGb: 16,
      storageGb: 200,
      price: "1000",
      customer: "ve1",
    });

    expect(msg.typeUrl).toBe("/virtengine.market.v1.MsgCreateOrder");
    expect(msg.value).toMatchObject({
      cpu: 4,
      memory_gb: 16,
      storage_gb: 200,
      price: "1000",
      customer: "ve1",
    });
  });

  it("builds market close order message", () => {
    const msg = buildMsgCloseOrder({ orderId: "order-1", reason: "done" });
    expect(msg.typeUrl).toBe("/virtengine.market.v1.MsgCloseOrder");
    expect(msg.value).toMatchObject({ order_id: "order-1", reason: "done" });
  });

  it("builds market bid message", () => {
    const msg = buildMsgCreateBid({
      orderId: "order-1",
      provider: "ve1",
      price: "1200",
    });
    expect(msg.typeUrl).toBe("/virtengine.market.v1.MsgCreateBid");
    expect(msg.value).toMatchObject({
      order_id: "order-1",
      provider: "ve1",
      price: "1200",
    });
  });

  it("builds veid upload scope message", () => {
    const msg = buildMsgUploadScope({
      sender: "ve1",
      scopeId: "scope-1",
      scopeType: 1,
      encryptedPayload: { payload: "data" },
      salt: "salt",
      deviceFingerprint: "device",
      clientId: "client",
      clientSignature: "sig",
      userSignature: "user-sig",
      payloadHash: "hash",
    });

    expect(msg.typeUrl).toBe("/virtengine.veid.v1.MsgUploadScope");
    expect(msg.value).toMatchObject({
      sender: "ve1",
      scope_id: "scope-1",
      scope_type: 1,
      client_id: "client",
    });
  });

  it("builds veid request verification message", () => {
    const msg = buildMsgRequestVerification({
      sender: "ve1",
      scopeId: "scope-1",
    });
    expect(msg.typeUrl).toBe("/virtengine.veid.v1.MsgRequestVerification");
    expect(msg.value).toMatchObject({ sender: "ve1", scope_id: "scope-1" });
  });
});
