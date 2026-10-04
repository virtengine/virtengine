# secp256k1 signature-verifier equivalence corpus

These files record the accept/reject verdict that
`x/enclave/keeper.verifySigningKeySignature` produced **before** the
go-ethereum-to-decred secp256k1 swap, so that the swap can be proven not to change
which attestations the chain considers valid.

## Why this exists

`x/enclave` is consensus-critical. Its secp256k1 attestation verification
previously called `github.com/ethereum/go-ethereum/crypto`, which is licensed
GPL-3.0 and is denied by `scripts/supply-chain/go-module-policy.json`. The module
was reachable from exactly one non-test file, so it was replaced with
`github.com/decred/dcrd/dcrec/secp256k1/v4` (ISC), which this repository already
depended on directly.

A replacement crypto implementation in a consensus path is only safe if the
accept/reject set is identical. `signature_equivalence_test.go` replays these
recorded verdicts against the new implementation and fails on any divergence, in
either the boolean outcome or the reason for a rejection.

## The two build paths

go-ethereum ships two different `VerifySignature` implementations:

| Build path | Implementation | Non-32-byte payload |
| --- | --- | --- |
| cgo (Linux production) | libsecp256k1 via `crypto/secp256k1` | rejected outright |
| no-cgo | `decred/dcrec/secp256k1` via `ModNScalar.SetByteSlice` | silently zero-padded / truncated |

These **disagree with each other**. A node built with cgo and a node built without
it reached different verdicts on the same attestation, which is a latent consensus
hazard that exists independently of the licence work. Both corpora are recorded
here so the divergence is visible rather than inherited.

The new implementation resolves it by requiring exactly 32 bytes, matching the cgo
(production) behaviour. Every caller in `x/enclave` passes a `sha256.Sum256`
digest, so no reachable input changes behaviour. See
`TestSignatureVerifierNonCgoPayloadDivergenceIsDeliberate`.

## Files

- `signature_equivalence_cgo.json` — recorded with `CGO_ENABLED=1` on
  linux/amd64, go1.26.8, gcc 13.3.0. **Authoritative**: this is the path production
  Linux validators run.
- `signature_equivalence_noc.json` — recorded with `CGO_ENABLED=0`. Retained as
  evidence of the divergence described above.

Both were produced from `github.com/ethereum/go-ethereum v1.17.0`.

## Case coverage

25 cases per corpus: valid signatures from two independent keys; tampered R, S and
payload; wrong-key; high-S malleable (`S = N - S`, MSB set); over-half-order S with
a clear MSB (the window an MSB-only low-S check misses); R = 0; S = 0; R = N;
S = N; R = 0xFF..FF; off-curve public key; X >= field prime; hybrid 0x06 and 0x07
encodings; 0x02/0x00 prefixes; all-zero key; 31/33/64-byte payloads; and 63/65-byte
signatures.

## Regenerating

Only possible while go-ethereum is still a dependency, so this is a one-shot
operation tied to the original removal commit. The generator was
`x/enclave/keeper/zz_equiv_gen_test.go` (deleted in the same change, since keeping
it would have re-added the GPL import). It called the old geth implementation over
the case list above and emitted the JSON.

If the verifier is ever changed again, do **not** regenerate these files. That
would defeat their purpose: they are a record of pre-change behaviour, not a
snapshot of current behaviour. Add new assertions to
`signature_equivalence_test.go` instead.
