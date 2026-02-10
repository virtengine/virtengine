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

- Native ML modules are pluggable. The default adapters fall back to mock services
  when native modules are not present.
- Encryption, biometric hardware capture, and device attestation are implemented as stubs
  with clear extension points for the production mobile SDK.
