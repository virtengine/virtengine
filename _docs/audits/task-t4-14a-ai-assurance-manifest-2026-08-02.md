# T4-14A AI Assurance Manifest Evidence

Status: diagnostic-green, production-blocked

The core-RC manifest generator now derives a closed AI assurance section from
committed model provenance, production policy, and feature parity fixtures. It
binds model, runtime, schema, and license digests; all four canonical feature
vector hashes; evaluation state; the current non-LSH uniqueness implementation
class; process-memory vault/key custody; allow-all consent fallback; and the
uncertified retention implementation state.

Blocked artifacts retain null digests and their source blocker identifiers.
The section explicitly keeps production certification false and names model,
runtime, evaluation, biometric uniqueness, durable KMS/vault, consent, and
retention/legal-hold/erasure profiles as uncertified.

This checkpoint adds generator and schema controls only. It does not certify
the current implementations or remove any T4-12 production policy finding.