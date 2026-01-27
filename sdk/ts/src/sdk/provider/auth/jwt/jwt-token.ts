import type { OfflineAminoSigner } from "@cosmjs/amino";
import { default as stableStringify } from "json-stable-stringify";

import { base64UrlDecode, base64UrlEncode, toBase64Url } from "./base64.ts";
import { JwtValidator } from "./jwt-validator.ts";
import type { JwtTokenPayload } from "./types.ts";
import type { OfflineDataSigner } from "./wallet-utils.ts";
import { createOfflineDataSigner } from "./wallet-utils.ts";

export class JwtTokenManager {
  private readonly validator: JwtValidator;
  private readonly signer: OfflineDataSigner;

  constructor(signer: OfflineDataSigner | OfflineAminoSigner) {
    this.validator = new JwtValidator();
    this.signer = "signAmino" in signer ? createOfflineDataSigner(signer) : signer;
  }

  /**
   * Creates a new JWT token with ES256K signature using a custom signArbitrary method with the current wallet
   * @param options - JWT token options
   * @returns The signed JWT token
   * @example
   * const wallet = await Secp256k1HdWallet.fromMnemonic(jwtMnemonic, {
   *   prefix: "virtengine"
   * });
   * const jwtToken = new JwtTokenManager(wallet);
   * // OR ON FRONTEND
   * const wallet = useSelectedChain();
   * const jwt = new JwtTokenManager(wallet);
   * const token = await jwtToken.generateToken({
   *   version: "v1",
   *   iss: wallet.address, // virt1...
   *   exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
   *   iat: Math.floor(Date.now() / 1000), // current timestamp
   * });
   * console.log(token);
   */
  async generateToken(options: CreateJWTOptions): Promise<string> {
    const now = Math.floor(Date.now() / 1000);
    const inputPayload: JwtTokenPayload = {
      iss: options.iss,
      exp: options.exp ? options.exp : now + 3600, // Default to 1 hour expiration
      nbf: options.nbf || now,
      iat: options.iat || now,
      version: options.version,
      leases: options.leases || { access: "full" },
    };
    if (options.jti) inputPayload.jti = options.jti;

    const validationResult = this.validatePayload(inputPayload);
    if (!validationResult.isValid) {
      throw new Error(`Invalid payload: ${validationResult.errors?.join(", ")}`);
    }

    const header = base64UrlEncode(stableStringify({ alg: this.signer.algorithm || "ES256KADR36", typ: "JWT" })!);
    const stringPayload = base64UrlEncode(stableStringify(inputPayload)!);
    const { signature } = await this.signer.signArbitrary(options.iss, `${header}.${stringPayload}`);
    const token = `${header}.${stringPayload}.${toBase64Url(signature)}`;

    return token;
  }

  /**
   * Decodes a JWT token
   * @param token - The JWT token to decode
   * @returns The decoded JWT payload
   * @throws Error if the token is malformed
   */
  decodeToken(token: string): JwtTokenPayload {
    const parts = token.split(".");
    if (parts.length !== 3) {
      throw new Error("Invalid JWT format");
    }

    try {
      const [, payload] = parts;
      const json = base64UrlDecode(payload);
      return JSON.parse(json);
    } catch (error) {
      throw new Error("Failed to decode JWT token", { cause: error });
    }
  }

  /**
   * Validates a JWT payload against the schema and time-based constraints
   * @param payload - The JWT payload to validate
   * @returns A boolean indicating whether the payload is valid
   */
  public validatePayload(payload: JwtTokenPayload): { isValid: boolean; errors?: string[] } {
    const result = this.validator.validateToken(payload);
    if (!result.isValid) {
      return { isValid: false, errors: result.errors };
    }

    const now = Math.floor(Date.now() / 1000);
    const errors: string[] = [];

    if (payload.exp <= now) {
      errors.push("Token has expired");
    }

    if (payload.nbf > now) {
      errors.push("Token is not yet valid (nbf check failed)");
    }

    return {
      isValid: errors.length === 0,
      errors: errors.length > 0 ? errors : undefined,
    };
  }
}

export interface CreateJWTOptions extends Partial<Omit<JwtTokenPayload, "iss" | "version">> {
  version: JwtTokenPayload["version"];
  iss: JwtTokenPayload["iss"];
}
