import { Secp256k1HdWallet } from "@cosmjs/amino";
import type { AccountData } from "@cosmjs/proto-signing";
import { beforeAll, describe, expect, it } from "@jest/globals";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

import { toBase64Url } from "./base64.ts";
import type { CreateJWTOptions } from "./jwt-token.ts";
import { JwtTokenManager } from "./jwt-token.ts";
import type { ClaimsTestCase, SigningTestCase } from "./test/test-utils.ts";
import { replaceTemplateValues } from "./test/test-utils.ts";
import { createOfflineDataSigner } from "./wallet-utils.ts";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe("JWT Claims Validation", () => {
  const testdataPath = path.join(__dirname, "../../../../../..", "testdata", "jwt");
  const jwtMnemonic = fs.readFileSync(path.join(testdataPath, "mnemonic"), "utf-8").trim();
  const jwtSigningTestCases = JSON.parse(fs.readFileSync(path.join(testdataPath, "cases_es256k.json"), "utf-8")) as SigningTestCase[];
  const jwtClaimsTestCases = JSON.parse(fs.readFileSync(path.join(testdataPath, "cases_jwt.json.tmpl"), "utf-8")) as ClaimsTestCase[];

  let testWallet: Secp256k1HdWallet;
  let signingWallet: Secp256k1HdWallet;
  let jwtToken: JwtTokenManager;
  let testAccount: AccountData;

  beforeAll(async () => {
    testWallet = await Secp256k1HdWallet.fromMnemonic(jwtMnemonic, {
      prefix: "ve",
    });
    signingWallet = await Secp256k1HdWallet.fromMnemonic(jwtMnemonic, {
      prefix: "akash",
    });
    const [account] = await testWallet.getAccounts();
    testAccount = account;
    jwtToken = new JwtTokenManager(testWallet);
  });

  it.each(jwtClaimsTestCases.filter(isSigningWithES256KADR36))("$description", async (testCase) => {
    const { claims, tokenString } = replaceTemplateValues(testCase, { iss: testAccount.address });

    // For test cases that should fail, we need to validate the payload first
    if (testCase.expected.signFail || testCase.expected.verifyFail) {
      const validationResult = jwtToken.validatePayload(claims);
      expect(validationResult.isValid).toBe(false);

      if (validationResult.isValid) {
        throw new Error("Validation should have failed", { cause: testCase });
      }

      return;
    }

    // For test cases that should pass, create and verify the token
    const token = await jwtToken.generateToken(claims as CreateJWTOptions);
    const decoded = jwtToken.decodeToken(token);
    expect(decoded).toBeDefined();

    // If the test case has a token string, compare it with the generated token
    if (tokenString) {
      expect(token).toEqual(tokenString);
    }
  });

  it.each(jwtSigningTestCases.filter(isSigningWithES256KADR36))("$description", async (testCase) => {
    const [expectedHeader, expectedPayload, expectedSignature] = testCase.tokenString.split(".");
    expect(expectedHeader).toBeDefined();
    expect(expectedPayload).toBeDefined();
    expect(expectedSignature).toBeDefined();

    const signingString = `${expectedHeader}.${expectedPayload}`;

    const signer = createOfflineDataSigner(signingWallet);
    const [account] = await signingWallet.getAccounts();
    const signResponse = await signer.signArbitrary(account.address, signingString);
    const signature = toBase64Url(signResponse.signature);

    if (!testCase.mustFail) {
      expect(signature).toBe(expectedSignature);
    } else {
      expect(signature).not.toBe(expectedSignature);
    }
  });

  function isSigningWithES256KADR36(testCase: SigningTestCase | ClaimsTestCase): boolean {
    return !testCase.expected.alg || testCase.expected.alg === "ES256KADR36";
  }
});
