# VEID Biometric Hardware + Device Attestation

This document describes biometric hardware capture and device integrity attestation support in VEID.
It complements the VEID Flow Specification and the Biometric Data Addendum.

## Supported Devices

### Android
- **Play Integrity API** (primary): device integrity + strong integrity verdicts
- **SafetyNet Attestation** (fallback for legacy devices)
- Supported sensors: fingerprint (optical/capacitive/ultrasonic), iris (if OEM hardware supports)

### iOS
- **App Attest** (primary)
- **DeviceCheck** (fallback for legacy devices)
- Supported sensors: Touch ID / Face ID / Iris (where available)

## Enrollment Flow (End-to-End)

1. **Consent**: user accepts biometric data notice and privacy terms.
2. **Document capture**: front/back with OCR extraction.
3. **Selfie + liveness**: active liveness challenges and anti-spoofing checks.
4. **Biometric hardware capture**: fingerprint/iris template capture via platform-secure APIs.
5. **Device attestation**: Play Integrity / App Attest token issued and verified.
6. **Encrypted payload**: biometric templates + attestation payloads encrypted into the vault.
7. **On-chain scope creation**: new scopes are created referencing encrypted vault payloads.
8. **Verification**: verification service checks attestation and stores signed results on-chain.

## On-Chain Scopes

New scope types are used for biometric hardware and device attestation payloads:
- `biometric_hardware`
- `device_attestation`

The encrypted payload stored in the identity scope references the vault (32A). The
verification attestation references the scope IDs for auditability.

## Privacy Considerations

- **Data minimization**: store only templates + integrity metadata required for verification.
- **Encryption**: all biometric templates and attestation payloads are encrypted using the
  VEID envelope and stored in the vault.
- **Consent**: biometric capture requires explicit consent (see `BIOMETRIC_DATA_ADDENDUM.md`).
- **Retention**: attestation payloads follow VEID lifecycle policies (expiration + revocation).
- **Auditability**: attestation failures are recorded on-chain for review and dispute handling.

## Fallback Behavior

- If the platform does not support attestation APIs, the capture flow continues with a
  `device_attestation` scope marked as unsupported.
- Verification pipelines should enforce **attestation-required** policies for high-risk actions
  (e.g., validator onboarding) and allow fallback for low-risk flows.
