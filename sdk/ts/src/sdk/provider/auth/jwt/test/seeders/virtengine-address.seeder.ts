import { faker } from "@faker-js/faker";

export function createVirtEngineAddress(): string {
  return `virt1${faker.string.alphanumeric({ length: 38, casing: "lower" })}`;
}
