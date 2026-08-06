# VEID Capture App (React Native)

React Native reference implementation for the VEID mobile capture flow.
This app implements the capture pipeline described in the VirtEngine mobile capture SDK
and aligns with AU2024203136A1 requirements for document + biometric verification.

## Scope

- Document capture (front/back)
- Selfie capture with face detection guidance
- Active liveness challenges (blink, head turn, smile)
- Biometric hardware capture (fingerprint / iris) with liveness + anti-spoofing signals
- Document OCR extraction + field parsing
- Secure payload packaging hooks (encryption + signing adapters)
- Device integrity attestation hooks (Play Integrity / App Attest)

## Requirements Mapping (AU2024203136A1)

- Document capture: multi-side capture with guided framing
- Biometric capture: selfie + liveness challenge-response
- Biometric hardware attestation: fingerprint/iris sensor verification
- Quality checks: face confidence + liveness gating
- OCR: extracted fields from document image
- Secure transport: encryption + device attestation hooks

## Development

```bash
cd mobile/veid-capture-app
npm install
npm run typecheck
npm test
```

## Notes

- Native camera, biometric, attestation, encryption, and signing modules are pluggable.
- Production paths fail closed when a native provider, device, permission, or capture binding
  is unavailable. Mock providers and insecure fixtures require explicit injection or configuration
  for tests and development.
