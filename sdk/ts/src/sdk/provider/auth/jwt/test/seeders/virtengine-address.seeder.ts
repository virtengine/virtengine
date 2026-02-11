import { faker } from "@faker-js/faker";

export function createVirtEngineAddress(): string {
  return `virtengine1${faker.string.alphanumeric({ length: 38, casing: "lower" })}`;
}
