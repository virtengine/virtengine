import { faker } from "@faker-js/faker";

// The chain's real bech32 account HRP is "ve" (sdk/go/sdkutil.Bech32PrefixAccAddr),
// so a valid address is "ve1" + 38 alphanumeric chars. Emitting the old
// "virtengine1" prefix produced addresses the JWT schema correctly rejects.
export function createVirtEngineAddress(): string {
  return `ve1${faker.string.alphanumeric({ length: 38, casing: "lower" })}`;
}
