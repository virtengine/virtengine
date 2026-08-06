import { describe, expect, it } from "@jest/globals";

import { ChainCapability, ChainCapabilityError, ChainCapabilityErrorReason } from "../sdk/chain/ChainCapability.ts";
import { BaseClient } from "./BaseClient.ts";

class TestClient extends BaseClient {
  handle(error: unknown): never {
    return this.handleQueryError(error, "testQuery");
  }
}

describe(BaseClient.name, () => {
  it("preserves ChainCapabilityError from high-level query methods", () => {
    const error = new ChainCapabilityError(
      ChainCapabilityErrorReason.Disconnected,
      ChainCapability.Disconnected,
      ChainCapability.QueryOnly,
      "testQuery",
    );

    expect(() => new TestClient().handle(error)).toThrow(error);
  });
});
