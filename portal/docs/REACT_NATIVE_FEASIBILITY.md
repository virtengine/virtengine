# React Native Companion App — Feasibility Assessment

## Executive Summary

A React Native companion app for VirtEngine is **feasible and recommended** as a future phase
(v2.0+). The responsive web portal covers 90% of mobile use-cases today; a native app adds value
through push notifications, biometric auth, offline capability, and camera-quality improvements for
VEID identity verification.

## Architecture

### Recommended Stack

| Layer         | Technology                               | Rationale                                       |
| ------------- | ---------------------------------------- | ----------------------------------------------- |
| Framework     | React Native + Expo                      | Shared JS/TS ecosystem with portal, OTA updates |
| Navigation    | React Navigation v7                      | Standard for RN, deep-linking support           |
| State         | Zustand                                  | Already used in portal — share store patterns   |
| Wallet        | CosmJS + WalletConnect v2                | Same chain interaction layer as web portal      |
| Camera        | `expo-camera` + `expo-image-manipulator` | Native camera access for VEID capture           |
| Auth          | `expo-local-authentication`              | FaceID / TouchID for biometric unlock           |
| Notifications | `expo-notifications` + FCM/APNs          | Push for order status, bid updates, governance  |

### Code Sharing Strategy

The portal already separates logic from UI via `lib/portal` (SDK adapter) and `stores/`. These can
be extracted into a shared package:

```
packages/
  shared/           # Shared between web and native
    stores/         # Zustand stores (identical)
    hooks/          # Business logic hooks
    types/          # TypeScript types
    config/         # Chain config, wallet config
  portal-web/       # Next.js portal (current portal/)
  portal-native/    # React Native app
```

**Estimated code reuse: 60-70%** of business logic, 0% of UI components.

## Native Advantages

### 1. Push Notifications

- **Order status changes**: Bid accepted, lease active, deployment ready
- **Governance alerts**: New proposals, vote reminders, quorum thresholds
- **Identity verification**: Approval/rejection notifications
- **Escrow events**: Deposit confirmations, settlement notices

### 2. Camera Integration for VEID

- Native camera pipeline offers higher quality captures than `getUserMedia`
- Direct access to auto-focus, exposure, HDR
- Real-time ML processing via TensorFlow Lite for liveness detection
- Better document edge detection with native vision frameworks

### 3. Offline Capability

- Cache active orders and deployment status for offline viewing
- Queue transactions for submission when connectivity returns
- Persist wallet credentials securely via Keychain/Keystore

### 4. Biometric Authentication

- FaceID / TouchID / fingerprint for wallet unlock
- Eliminate password entry for transaction signing
- Hardware-backed key storage via Secure Enclave / StrongBox

## Effort Estimate

| Phase   | Scope                                            | Duration  | Team        |
| ------- | ------------------------------------------------ | --------- | ----------- |
| Phase 1 | Core shell + wallet connect + marketplace browse | 4-6 weeks | 2 engineers |
| Phase 2 | Orders + dashboard + push notifications          | 3-4 weeks | 2 engineers |
| Phase 3 | VEID camera capture + biometric auth             | 3-4 weeks | 2 engineers |
| Phase 4 | Offline support + polish + app store submission  | 3-4 weeks | 2 engineers |

**Total: ~14-18 weeks** for a production-ready companion app.

## Risks

| Risk                            | Mitigation                                                  |
| ------------------------------- | ----------------------------------------------------------- |
| App store review delays         | Submit early builds to TestFlight/Play Console for feedback |
| WalletConnect mobile UX         | Test with Keplr Mobile, Leap Mobile, Cosmostation Mobile    |
| Camera permission UX on iOS     | Clear permission prompts with rationale strings             |
| Bundle size                     | Use Expo's tree-shaking; lazy-load CosmJS modules           |
| Chain RPC reliability on mobile | Implement retry logic, multiple RPC fallbacks               |

## Recommendation

**Phase the native app after portal v1.5** (post-marketplace + identity flows are stable). The
responsive web portal with WalletConnect support provides adequate mobile coverage for initial
launch. Invest in native when user demand for push notifications and camera quality becomes a
measurable retention driver.

## Prerequisites

Before starting native development:

1. Extract shared stores/hooks into `packages/shared/` (monorepo refactor)
2. Stabilize VEID identity flow on web (camera capture + ML scoring)
3. WalletConnect v2 integration complete on web portal
4. Set up CI/CD for mobile builds (EAS Build + Fastlane)
5. Apple Developer + Google Play Console accounts provisioned
