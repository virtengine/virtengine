import { beforeEach, describe, expect, it } from "@jest/globals";

import { JwtValidator } from "./jwt-validator.ts";

const issuer = "virtengine1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63";
const provider = "virtengine18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4";

function toBase64Url(value: Record<string, unknown>) {
  return Buffer.from(JSON.stringify(value)).toString("base64url");
}

function createToken(header: Record<string, unknown>, payload: Record<string, unknown>) {
  return `${toBase64Url(header)}.${toBase64Url(payload)}.signature`;
}

describe("JwtValidator", () => {
  let validator: JwtValidator;

  beforeEach(() => {
    validator = new JwtValidator();
  });

  it("should validate a valid token", () => {
    const validToken = createToken(
      { alg: "ES256K", typ: "JWT" },
      {
        iss: issuer,
        iat: 1654000000,
        exp: 1654003600,
        nbf: 1654000000,
        version: "v1",
        leases: {
          access: "granular",
          permissions: [
            {
              provider,
              access: "scoped",
              scope: ["send-manifest"],
            },
          ],
        },
      },
    );
    const result = validator.validateToken(validToken);
    expect(result.isValid).toBe(true);
    expect(result.errors.length).toBe(0);
    expect(result.decodedToken).toBeDefined();
  });

  it("should reject a malformed token", () => {
    const result = validator.validateToken("not.a.valid.token");
    expect(result.isValid).toBe(false);
    expect(result.errors.length).toBeGreaterThan(0);
    expect(result.errors[0]).toContain("Error during JWT validation");
  });

  it("should validate required fields in header", () => {
    const result = validator.validateToken(createToken(
      { typ: "JWT" },
      {
        iss: issuer,
        iat: 1654000000,
        exp: 1654003600,
        nbf: 1654000000,
        version: "v1",
        leases: { access: "full" },
      },
    ));
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("Missing required field in header: alg");
  });

  it("should validate required fields in payload", () => {
    const result = validator.validateToken("eyJhbGciOiJFUzI1NksiLCJ0eXAiOiJKV1QifQ.eyJmb28iOiJiYXIifQ.signature");
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("Missing required field: \"iss\".");
    expect(result.errors).toContain("Missing required field: \"iat\".");
    expect(result.errors).toContain("Missing required field: \"exp\".");
    expect(result.errors).toContain("Missing required field: \"nbf\".");
    expect(result.errors).toContain("Missing required field: \"version\".");
    expect(result.errors).toContain("Missing required field: \"leases\".");
    expect(result.errors).toContain("Additional property \"foo\" is not allowed.");
  });

  it("should validate leases object when present", () => {
    const token = createToken(
      { alg: "ES256K", typ: "JWT" },
      {
        iss: issuer,
        iat: 1654000000,
        exp: 1654003600,
        nbf: 1654000000,
        version: "v1",
        leases: {},
      },
    );
    const result = validator.validateToken(token);
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("Missing required field: \"access\" at \"/leases\".");
  });

  it("should validate granular access requires permissions", () => {
    const token = createToken(
      { alg: "ES256K", typ: "JWT" },
      {
        iss: issuer,
        iat: 1654000000,
        exp: 1654003600,
        nbf: 1654000000,
        version: "v1",
        leases: { access: "granular" },
      },
    );
    const result = validator.validateToken(token);
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("Missing required field: permissions");
  });
});
