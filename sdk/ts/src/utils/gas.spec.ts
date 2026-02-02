import { describe, expect, it } from "@jest/globals";

import {
  adjustGas,
  calculateFee,
  createAutoFee,
  createFeeForMessage,
  createFeeForMessages,
  DEFAULT_GAS_CONFIG,
  estimateGas,
  estimateGasForMessages,
  GAS_ESTIMATES,
  isValidGasPrice,
  parseGasPrice,
} from "./gas.ts";

describe("GAS_ESTIMATES", () => {
  it("should have VEID message types defined", () => {
    expect(GAS_ESTIMATES["veid/MsgUploadScope"]).toBe(250000);
    expect(GAS_ESTIMATES["veid/MsgRequestVerification"]).toBe(150000);
    expect(GAS_ESTIMATES["veid/MsgCreateIdentityWallet"]).toBe(200000);
  });

  it("should have MFA message types defined", () => {
    expect(GAS_ESTIMATES["mfa/MsgEnrollFactor"]).toBe(180000);
    expect(GAS_ESTIMATES["mfa/MsgVerifyChallenge"]).toBe(150000);
  });

  it("should have Market message types defined", () => {
    expect(GAS_ESTIMATES["market/MsgCreateBid"]).toBe(200000);
    expect(GAS_ESTIMATES["market/MsgCloseBid"]).toBe(150000);
    expect(GAS_ESTIMATES["market/MsgCloseLease"]).toBe(180000);
  });

  it("should have Escrow message types defined", () => {
    expect(GAS_ESTIMATES["escrow/MsgDeposit"]).toBe(150000);
    expect(GAS_ESTIMATES["escrow/MsgWithdraw"]).toBe(150000);
  });

  it("should have HPC message types defined", () => {
    expect(GAS_ESTIMATES["hpc/MsgSubmitJob"]).toBe(300000);
    expect(GAS_ESTIMATES["hpc/MsgCancelJob"]).toBe(120000);
  });

  it("should have Cosmos message types defined", () => {
    expect(GAS_ESTIMATES["cosmos.bank.v1beta1/MsgSend"]).toBe(100000);
    expect(GAS_ESTIMATES["cosmos.staking.v1beta1/MsgDelegate"]).toBe(200000);
  });

  it("should have a default fallback", () => {
    expect(GAS_ESTIMATES.default).toBe(200000);
  });
});

describe("DEFAULT_GAS_CONFIG", () => {
  it("should have reasonable default values", () => {
    expect(DEFAULT_GAS_CONFIG.gasPrice).toBe("0.025uakt");
    expect(DEFAULT_GAS_CONFIG.gasAdjustment).toBe(1.3);
    expect(DEFAULT_GAS_CONFIG.defaultGasLimit).toBe(200000);
  });
});

describe("parseGasPrice", () => {
  it("should parse valid gas price string", () => {
    const result = parseGasPrice("0.025uakt");
    expect(result.amount).toBe("0.025");
    expect(result.denom).toBe("uakt");
  });

  it("should parse integer gas price", () => {
    const result = parseGasPrice("100uvirt");
    expect(result.amount).toBe("100");
    expect(result.denom).toBe("uvirt");
  });

  it("should throw for invalid format", () => {
    expect(() => parseGasPrice("invalid")).toThrow("Invalid gas price format");
    expect(() => parseGasPrice("")).toThrow("Invalid gas price format");
  });
});

describe("calculateFee", () => {
  it("should calculate fee from gas limit and price", () => {
    const fee = calculateFee(100000, "0.025uakt");
    expect(fee.gas).toBe("100000");
    expect(fee.amount[0].denom).toBe("uakt");
    expect(fee.amount[0].amount).toBe("2500");
  });

  it("should round up fee amount", () => {
    const fee = calculateFee(100001, "0.025uakt");
    expect(fee.amount[0].amount).toBe("2501");
  });
});

describe("adjustGas", () => {
  it("should apply default adjustment", () => {
    const result = adjustGas(100000);
    expect(result).toBe(130000);
  });

  it("should apply custom adjustment", () => {
    const result = adjustGas(100000, 1.5);
    expect(result).toBe(150000);
  });

  it("should round up to integer", () => {
    const result = adjustGas(100001, 1.3);
    expect(result).toBe(Math.ceil(100001 * 1.3));
  });
});

describe("createAutoFee", () => {
  it("should return 'auto'", () => {
    expect(createAutoFee()).toBe("auto");
  });
});

describe("estimateGas", () => {
  it("should return estimate for known message type", () => {
    const estimate = estimateGas("veid/MsgUploadScope");
    expect(estimate).toBe(250000);
  });

  it("should return default for unknown message type", () => {
    const estimate = estimateGas("unknown/MsgDoSomething");
    expect(estimate).toBe(GAS_ESTIMATES.default);
  });
});

describe("estimateGasForMessages", () => {
  it("should add base overhead to single message", () => {
    const estimate = estimateGasForMessages(["veid/MsgUploadScope"]);
    // Base 80000 + 250000
    expect(estimate).toBe(330000);
  });

  it("should sum gas for multiple messages", () => {
    const estimate = estimateGasForMessages([
      "cosmos.bank.v1beta1/MsgSend",
      "cosmos.staking.v1beta1/MsgDelegate",
    ]);
    // Base 80000 + 100000 + 200000
    expect(estimate).toBe(380000);
  });

  it("should return base overhead for empty array", () => {
    const estimate = estimateGasForMessages([]);
    expect(estimate).toBe(80000);
  });
});

describe("createFeeForMessage", () => {
  it("should create fee with defaults", () => {
    const fee = createFeeForMessage("cosmos.bank.v1beta1/MsgSend");
    // 100000 * 1.3 = 130000 gas, 130000 * 0.025 = 3250 uakt
    expect(fee.gas).toBe("130000");
    expect(fee.amount[0].amount).toBe("3250");
    expect(fee.amount[0].denom).toBe("uakt");
  });

  it("should use custom gas price", () => {
    const fee = createFeeForMessage("cosmos.bank.v1beta1/MsgSend", "0.1uvirt");
    expect(fee.amount[0].denom).toBe("uvirt");
  });
});

describe("createFeeForMessages", () => {
  it("should create fee for multiple messages", () => {
    const fee = createFeeForMessages([
      "cosmos.bank.v1beta1/MsgSend",
      "cosmos.bank.v1beta1/MsgSend",
    ]);
    // Base 80000 + 100000 + 100000 = 280000, * 1.3 = 364000
    expect(fee.gas).toBe("364000");
  });
});

describe("isValidGasPrice", () => {
  it("should return true for valid gas price", () => {
    expect(isValidGasPrice("0.025uakt")).toBe(true);
    expect(isValidGasPrice("100uvirt")).toBe(true);
  });

  it("should return false for invalid gas price", () => {
    expect(isValidGasPrice("invalid")).toBe(false);
    expect(isValidGasPrice("")).toBe(false);
  });
});
