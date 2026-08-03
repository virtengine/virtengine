// Fixture-only vector. This module has no runtime dependencies.
export const requestV1Vector = {
  name: "request-v1",
  domain: "virtengine/federation/request/v1",
  providerId: "provider-1",
  serviceId: "orders",
  method: "post",
  requestTarget: "/v1/orders?z=two&a=one",
  bodyBase64: "eyJpZCI6Im9yZGVyLTEifQ==",
  contentType: "application/json; charset=utf-8",
  timestampUnix: 1800000000,
  nonce: "vector-nonce-1",
  signingKeyEpoch: 7,
  discoveryDigestHex:
    "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
  chainId: "virtengine-1",
  canonicalBase64:
    "AAAAIHZpcnRlbmdpbmUvZmVkZXJhdGlvbi9yZXF1ZXN0L3YxAAAAAQAAAApwcm92aWRlci0xAAAABm9yZGVycwAAAARQT1NUAAAACi92MS9vcmRlcnMAAAALYT1vbmUmej10d2/vsWkfMxeOtZppfoQVPZ1bwvZK5gOMS26uOVSx+UJmbwAAAB9hcHBsaWNhdGlvbi9qc29uOyBjaGFyc2V0PXV0Zi04AAAAAGtJ0gAAAAAOdmVjdG9yLW5vbmNlLTEAAAAAAAAABwECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gAAAADHZpcnRlbmdpbmUtMQ==",
  digestHex: "c578be4fa829c04d0222cc4375e24c437d959f86049507796353d99cf16c7dc0",
} as const;