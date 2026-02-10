import { faker } from "@faker-js/faker";
import { describe, expect, it } from "@jest/globals";

import { CertificateManager, dateToStr, strToDate } from "./CertificateManager.ts";

describe(CertificateManager.name, () => {
  const generateAddress = () => `virt1${faker.string.alpha({ length: 38 })}`;
  const ONE_DAY = 1000 * 60 * 60 * 24;

  describe("prototype.generateCertificate", () => {
    it("generates certificate PEMs with the default validity range", async () => {
      const manager = setup();

      const now = new Date();
      now.setMilliseconds(0);
      const inOneYear = getTime1yearFrom(now);
      const pem = await manager.generatePEM(generateAddress());
      const meta = await manager.parsePem(pem.cert);

      expect(pem).toMatchObject({
        cert: expect.stringMatching(/^-----BEGIN CERTIFICATE-----[\s\S]*-----END CERTIFICATE-----\r\n$/),
        publicKey: expect.stringMatching(/^-----BEGIN EC PUBLIC KEY-----[\s\S]*-----END EC PUBLIC KEY-----\r\n$/),
        privateKey: expect.stringMatching(/^-----BEGIN PRIVATE KEY-----[\s\S]*-----END PRIVATE KEY-----\r\n$/),
      });
      expect(new Date(meta.issuedOn).getTime()).toBeGreaterThanOrEqual(now.getTime());
      expect(new Date(meta.issuedOn).getTime()).toBeLessThan(new Date().getTime());
      expect(new Date(meta.expiresOn).getTime()).toBeGreaterThanOrEqual(inOneYear);
      expect(new Date(meta.issuedOn).getTime()).toBeLessThan(getTime1yearFrom(new Date()));
    });

    it("generates certificate PEMs with the provided validity range", async () => {
      const now = new Date();
      const inOneMonth = new Date(now.getTime() + 30 * ONE_DAY);
      const inTwoMonths = new Date(now.getTime() + 2 * 30 * ONE_DAY);
      const manager = setup();
      const pem = await manager.generatePEM(generateAddress(), {
        validFrom: inOneMonth,
        validTo: inTwoMonths,
      });
      const meta = await manager.parsePem(pem.cert);

      inOneMonth.setMilliseconds(0);
      inTwoMonths.setMilliseconds(0);

      expect(new Date(meta.issuedOn).getTime()).toBe(inOneMonth.getTime());
      expect(new Date(meta.expiresOn).getTime()).toBe(inTwoMonths.getTime());
    });
  });

  describe("prototype.parsePem", () => {
    it("extracts certificate details", async () => {
      const manager = setup();

      const pem = await manager.generatePEM(generateAddress());
      const cert = await manager.parsePem(pem.cert);

      expect(cert).toMatchObject({
        hSerial: expect.any(String),
        sIssuer: expect.any(String),
        sSubject: expect.any(String),
        sNotBefore: expect.any(String),
        sNotAfter: expect.any(String),
        issuedOn: expect.any(Date),
        expiresOn: expect.any(Date),
      });
    });
  });

  describe("strToDate", () => {
    it("should convert string to date", () => {
      const date = strToDate("240507122350Z");

      expect(date).toBeInstanceOf(Date);
      expect(date.toISOString()).toBe("2024-05-07T12:23:50.000Z");
    });
  });

  describe("dateToStr", () => {
    it("should convert date to string", () => {
      const date = new Date("2024-05-07T12:23:50.000Z");
      const str = dateToStr(date);

      expect(str).toBe("240507122350Z");
    });
  });

  function setup() {
    return new CertificateManager();
  }
});

function getTime1yearFrom(date = new Date()): number {
  const clone = new Date(date);
  const inOneYear = new Date(date);
  inOneYear.setFullYear(clone.getFullYear() + 1);
  return inOneYear.getTime();
}
