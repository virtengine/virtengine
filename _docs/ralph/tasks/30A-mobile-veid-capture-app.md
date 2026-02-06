# Task 30A: React Native Mobile App for VEID Capture

**vibe-kanban ID:** `4106db7d-8d00-4f29-98d0-9d164e4cbb72`

## Problem Statement

The patent specification explicitly requires mobile capture capability for VEID identity verification. Currently, the portal is PWA-only which limits:
- High-resolution camera access
- Native liveness detection integration
- Device attestation (Play Integrity, App Attest)
- Offline capability
- App store distribution

A native mobile app provides better UX for document/selfie capture and stronger security guarantees.

## Acceptance Criteria

### AC-1: Document Capture Module
- [ ] High-resolution document capture (4K support)
- [ ] Real-time quality validation (blur, lighting, framing)
- [ ] Guided capture with visual feedback
- [ ] Edge detection and auto-crop
- [ ] Multiple document types (passport, ID, driver license)

### AC-2: Selfie Capture with Liveness
- [ ] Front camera selfie capture
- [ ] Liveness detection integration (turn head, blink, smile)
- [ ] Anti-spoofing checks (screen detection)
- [ ] Quality validation
- [ ] Guided instructions

### AC-3: Wallet Integration
- [ ] WalletConnect v2 integration
- [ ] Keplr mobile support
- [ ] Leap mobile support
- [ ] Transaction signing
- [ ] VEID status display

### AC-4: Provider API Integration
- [ ] ProviderAPIClient (from 29D) integration
- [ ] Encrypted upload to provider
- [ ] Progress tracking
- [ ] Retry logic

### AC-5: Security
- [ ] Device attestation (Play Integrity / App Attest)
- [ ] Certificate pinning
- [ ] Secure storage for keys
- [ ] Anti-tampering detection
- [ ] Biometric app lock

### AC-6: App Store Compliance
- [ ] iOS App Store guidelines compliance
- [ ] Android Play Store guidelines compliance
- [ ] Privacy policy integration
- [ ] Permission handling

## Technical Requirements

### Project Structure

```
mobile/
├── app.json                        # Expo config
├── package.json
├── tsconfig.json
├── babel.config.js
├── metro.config.js
├── eas.json                        # EAS Build config
├── ios/
├── android/
├── assets/
│   ├── images/
│   └── fonts/
├── src/
│   ├── App.tsx
│   ├── navigation/
│   │   ├── AppNavigator.tsx
│   │   ├── AuthNavigator.tsx
│   │   └── types.ts
│   ├── screens/
│   │   ├── HomeScreen.tsx
│   │   ├── WalletScreen.tsx
│   │   ├── DocumentCaptureScreen.tsx
│   │   ├── SelfieCaptureScreen.tsx
│   │   ├── LivenessScreen.tsx
│   │   ├── ReviewScreen.tsx
│   │   ├── VerificationStatusScreen.tsx
│   │   └── SettingsScreen.tsx
│   ├── components/
│   │   ├── Camera/
│   │   │   ├── CameraView.tsx
│   │   │   ├── DocumentOverlay.tsx
│   │   │   ├── SelfieOverlay.tsx
│   │   │   ├── QualityIndicator.tsx
│   │   │   └── CaptureButton.tsx
│   │   ├── Liveness/
│   │   │   ├── LivenessChallenge.tsx
│   │   │   ├── LivenessProgress.tsx
│   │   │   └── LivenessInstructions.tsx
│   │   ├── Wallet/
│   │   │   ├── WalletConnectButton.tsx
│   │   │   ├── WalletInfo.tsx
│   │   │   └── TransactionModal.tsx
│   │   └── UI/
│   │       ├── Button.tsx
│   │       ├── Card.tsx
│   │       └── Progress.tsx
│   ├── hooks/
│   │   ├── useCamera.ts
│   │   ├── useLiveness.ts
│   │   ├── useWallet.ts
│   │   ├── useProviderAPI.ts
│   │   └── useDeviceAttestation.ts
│   ├── services/
│   │   ├── provider-api.ts
│   │   ├── wallet-connect.ts
│   │   ├── image-processing.ts
│   │   └── secure-storage.ts
│   ├── stores/
│   │   ├── verification-store.ts
│   │   ├── wallet-store.ts
│   │   └── app-store.ts
│   └── utils/
│       ├── crypto.ts
│       └── permissions.ts
└── __tests__/
```

### Camera Module Implementation

```typescript
// src/hooks/useCamera.ts
import { Camera, CameraType, FlashMode } from 'expo-camera';
import * as ImageManipulator from 'expo-image-manipulator';

export interface CaptureOptions {
  quality: number;
  maxWidth: number;
  maxHeight: number;
  exif: boolean;
}

export interface CaptureResult {
  uri: string;
  base64?: string;
  width: number;
  height: number;
  exif?: Record<string, unknown>;
}

export function useCamera() {
  const [permission, requestPermission] = Camera.useCameraPermissions();
  const cameraRef = useRef<Camera>(null);
  const [isProcessing, setIsProcessing] = useState(false);

  const capture = async (options: CaptureOptions): Promise<CaptureResult> => {
    if (!cameraRef.current) {
      throw new Error('Camera not ready');
    }

    setIsProcessing(true);
    try {
      const photo = await cameraRef.current.takePictureAsync({
        quality: options.quality,
        exif: options.exif,
        base64: true,
        skipProcessing: false,
      });

      // Resize if needed
      const manipulated = await ImageManipulator.manipulateAsync(
        photo.uri,
        [
          {
            resize: {
              width: Math.min(photo.width, options.maxWidth),
              height: Math.min(photo.height, options.maxHeight),
            },
          },
        ],
        { compress: options.quality, format: ImageManipulator.SaveFormat.JPEG }
      );

      return {
        uri: manipulated.uri,
        base64: manipulated.base64,
        width: manipulated.width,
        height: manipulated.height,
        exif: photo.exif,
      };
    } finally {
      setIsProcessing(false);
    }
  };

  return {
    permission,
    requestPermission,
    cameraRef,
    capture,
    isProcessing,
  };
}
```

### Quality Validation

```typescript
// src/services/image-processing.ts
export interface QualityResult {
  isValid: boolean;
  issues: QualityIssue[];
  scores: {
    blur: number;      // 0-100, higher is better
    lighting: number;  // 0-100
    framing: number;   // 0-100
  };
}

export enum QualityIssue {
  TooBlurry = 'too_blurry',
  TooDark = 'too_dark',
  TooBright = 'too_bright',
  BadFraming = 'bad_framing',
  GlareDetected = 'glare_detected',
  FaceNotFound = 'face_not_found',
}

export async function validateDocumentImage(
  imageUri: string
): Promise<QualityResult> {
  // Use TensorFlow Lite or ML Kit for quality checks
  const scores = await runQualityModel(imageUri);
  
  const issues: QualityIssue[] = [];
  
  if (scores.blur < 70) issues.push(QualityIssue.TooBlurry);
  if (scores.lighting < 40) issues.push(QualityIssue.TooDark);
  if (scores.lighting > 90) issues.push(QualityIssue.TooBright);
  if (scores.framing < 60) issues.push(QualityIssue.BadFraming);

  return {
    isValid: issues.length === 0,
    issues,
    scores,
  };
}
```

### Liveness Detection

```typescript
// src/hooks/useLiveness.ts
export interface LivenessChallenge {
  type: 'turn_left' | 'turn_right' | 'blink' | 'smile' | 'nod';
  duration: number;  // milliseconds
}

export interface LivenessResult {
  passed: boolean;
  challenges: {
    type: LivenessChallenge['type'];
    passed: boolean;
    confidence: number;
  }[];
  frames: string[];  // Base64 frames for verification
}

export function useLiveness(challenges: LivenessChallenge[]) {
  const [currentChallenge, setCurrentChallenge] = useState(0);
  const [results, setResults] = useState<LivenessResult['challenges']>([]);
  const [frames, setFrames] = useState<string[]>([]);
  
  // Use ML Kit Face Detection for challenge validation
  const validateChallenge = async (
    challenge: LivenessChallenge,
    faceData: FaceDetectionResult
  ): Promise<boolean> => {
    switch (challenge.type) {
      case 'turn_left':
        return faceData.headEulerAngleY < -20;
      case 'turn_right':
        return faceData.headEulerAngleY > 20;
      case 'blink':
        return faceData.leftEyeOpenProbability < 0.3 && 
               faceData.rightEyeOpenProbability < 0.3;
      case 'smile':
        return faceData.smilingProbability > 0.7;
      case 'nod':
        return faceData.headEulerAngleX < -15 || 
               faceData.headEulerAngleX > 15;
      default:
        return false;
    }
  };

  return {
    currentChallenge,
    challenges,
    progress: (currentChallenge / challenges.length) * 100,
    validateChallenge,
    results,
    frames,
  };
}
```

### WalletConnect Integration

```typescript
// src/services/wallet-connect.ts
import { Core } from '@walletconnect/core';
import { Web3Wallet, IWeb3Wallet } from '@walletconnect/web3wallet';

const WALLET_CONNECT_PROJECT_ID = process.env.EXPO_PUBLIC_WC_PROJECT_ID;

export class WalletConnectService {
  private web3wallet: IWeb3Wallet | null = null;

  async initialize(): Promise<void> {
    const core = new Core({
      projectId: WALLET_CONNECT_PROJECT_ID,
    });

    this.web3wallet = await Web3Wallet.init({
      core,
      metadata: {
        name: 'VirtEngine Mobile',
        description: 'VEID Identity Verification App',
        url: 'https://virtengine.io',
        icons: ['https://virtengine.io/icon.png'],
      },
    });
  }

  async connect(uri: string): Promise<void> {
    if (!this.web3wallet) throw new Error('Not initialized');
    await this.web3wallet.core.pairing.pair({ uri });
  }

  async signArbitrary(
    signerAddress: string,
    data: string
  ): Promise<string> {
    // Cosmos arbitrary signing (ADR-036)
    const request = {
      method: 'cosmos_signArbitrary',
      params: {
        signerAddress,
        data,
      },
    };
    
    // Request signature from connected wallet
    return await this.requestSignature(request);
  }
}
```

### Device Attestation

```typescript
// src/hooks/useDeviceAttestation.ts
import * as Application from 'expo-application';
import { Platform } from 'react-native';

export interface AttestationResult {
  platform: 'ios' | 'android';
  attestation: string;  // Base64 encoded
  timestamp: number;
}

export async function getDeviceAttestation(
  challenge: string
): Promise<AttestationResult> {
  if (Platform.OS === 'ios') {
    // Use App Attest
    const attestation = await generateAppAttest(challenge);
    return {
      platform: 'ios',
      attestation: attestation,
      timestamp: Date.now(),
    };
  } else {
    // Use Play Integrity API
    const attestation = await generatePlayIntegrity(challenge);
    return {
      platform: 'android',
      attestation: attestation,
      timestamp: Date.now(),
    };
  }
}

// iOS: App Attest via native module
async function generateAppAttest(challenge: string): Promise<string> {
  // Native module calls DCAppAttestService
  return NativeModules.AppAttest.generateAssertion(challenge);
}

// Android: Play Integrity API
async function generatePlayIntegrity(challenge: string): Promise<string> {
  // Use @react-native-google-signin/google-signin or custom native module
  return NativeModules.PlayIntegrity.requestIntegrityToken(challenge);
}
```

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `mobile/` | React Native project root | N/A |
| `mobile/src/screens/*.tsx` | Screen components | 2,500 |
| `mobile/src/components/Camera/*.tsx` | Camera components | 1,500 |
| `mobile/src/components/Liveness/*.tsx` | Liveness components | 800 |
| `mobile/src/components/Wallet/*.tsx` | Wallet components | 600 |
| `mobile/src/hooks/*.ts` | React hooks | 1,200 |
| `mobile/src/services/*.ts` | Services | 1,500 |
| `mobile/src/stores/*.ts` | Zustand stores | 500 |
| `mobile/__tests__/*.ts` | Tests | 1,500 |

**Total Estimated:** 10,000-12,000 lines

## Dependencies

- **29D:** Provider API TypeScript client (required)
- **29E:** Wallet-signed authentication (required)
- **lib/portal:** Shared types (import)

## Validation Checklist

- [ ] Camera permissions working (iOS + Android)
- [ ] Document capture quality validation
- [ ] Selfie capture working
- [ ] Liveness detection passing
- [ ] WalletConnect connecting to Keplr/Leap
- [ ] Transaction signing working
- [ ] Provider API upload successful
- [ ] Device attestation generating
- [ ] Certificate pinning configured
- [ ] App builds for iOS simulator
- [ ] App builds for Android emulator
- [ ] E2E tests passing
- [ ] App Store submission checklist complete
- [ ] Play Store submission checklist complete

## Testing Strategy

1. **Unit Tests:** Camera hooks, quality validation, wallet service
2. **Integration Tests:** Full capture → upload flow
3. **E2E Tests:** Detox/Maestro test suites
4. **Manual Tests:** Device-specific testing matrix

## App Store Checklist

### iOS
- [ ] Privacy policy URL
- [ ] App Store Connect listing
- [ ] Screenshots (6.5", 5.5")
- [ ] App icon
- [ ] Camera usage description
- [ ] Face ID usage description
- [ ] App Tracking Transparency

### Android
- [ ] Privacy policy URL
- [ ] Play Console listing
- [ ] Screenshots
- [ ] Feature graphic
- [ ] Camera permission rationale
- [ ] Biometric permission rationale
- [ ] Data safety form
