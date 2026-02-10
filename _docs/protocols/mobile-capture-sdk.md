# Mobile Capture SDK Documentation

## Overview

The VirtEngine Mobile Capture SDK enables secure capture of identity documents, selfies, and liveness videos for the VEID (VirtEngine Identity) verification system. All captured payloads are encrypted and signed before submission to the blockchain.

**Version:** 1.0.0  
**Protocol Version:** 1  
**Minimum Requirements:**
- Android 8.0 (API 26) or later
- iOS 13.0 or later

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Mobile Application                          │
├─────────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌────────────────┐  ┌────────────────────────┐  │
│  │ Capture Flow  │──│ Quality Checks │──│ Liveness Detection     │  │
│  │ Executor      │  │ (Blur, Face)   │  │ (Challenge-Response)   │  │
│  └───────────────┘  └────────────────┘  └────────────────────────┘  │
│          │                                         │                │
│          ▼                                         ▼                │
│  ┌───────────────┐  ┌────────────────┐  ┌────────────────────────┐  │
│  │ Compression   │──│ Encryption     │──│ Salt Binding +         │  │
│  │ (JPEG/HEIC)   │  │ (X25519-NaCl)  │  │ Dual Signatures        │  │
│  └───────────────┘  └────────────────┘  └────────────────────────┘  │
│          │                                         │                │
│          ▼                                         ▼                │
│  ┌───────────────┐                      ┌────────────────────────┐  │
│  │ Device        │                      │ Chunked Upload with    │  │
│  │ Attestation   │                      │ Retry/Resume           │  │
│  └───────────────┘                      └────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Backend Server                              │
├─────────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌────────────────┐  ┌────────────────────────┐  │
│  │ Upload        │──│ Validation     │──│ Tamper Detection       │  │
│  │ Handler       │  │ (Schema/Sig)   │  │ (Hash Verify)          │  │
│  └───────────────┘  └────────────────┘  └────────────────────────┘  │
│          │                                         │                │
│          ▼                                         ▼                │
│  ┌───────────────┐                      ┌────────────────────────┐  │
│  │ Scope         │                      │ Audit Trail            │  │
│  │ Submission    │──────────────────────│ Logger                 │  │
│  └───────────────┘                      └────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
                    ┌──────────────────────────┐
                    │   VirtEngine Blockchain  │
                    │   (x/veid module)        │
                    └──────────────────────────┘
```

## Quick Start

### Installation

**Android (Gradle):**
```gradle
dependencies {
    implementation 'io.virtengine:veid-capture-sdk:1.0.0'
}
```

**iOS (CocoaPods):**
```ruby
pod 'VirtEngineCaptureSDK', '~> 1.0.0'
```

**iOS (Swift Package Manager):**
```swift
.package(url: "https://github.com/virtengine/veid-capture-sdk-ios.git", from: "1.0.0")
```

### Basic Usage

#### Android (Kotlin)

```kotlin
import io.virtengine.veid.capture.*

class VerificationActivity : AppCompatActivity() {
    private lateinit var captureFlow: CaptureFlowExecutor
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Configure the capture flow
        val config = CaptureFlowConfig.Builder()
            .setClientId("your-approved-client-id")
            .setValidatorPublicKeys(listOf("validator-pubkey-1", "validator-pubkey-2"))
            .enableDocumentCapture(DocumentType.ID_CARD)
            .enableSelfieCapture()
            .enableLivenessDetection(LivenessMode.CHALLENGE_RESPONSE)
            .build()
        
        // Initialize the capture executor
        captureFlow = CaptureFlowExecutor(this, config)
    }
    
    fun startVerification() {
        lifecycleScope.launch {
            try {
                // Execute the capture flow
                val result = captureFlow.execute()
                
                if (result.isSuccess) {
                    // Upload encrypted payloads
                    uploadPayloads(result.encryptedPayloads)
                } else {
                    handleError(result.error)
                }
            } catch (e: CaptureException) {
                handleError(e)
            }
        }
    }
    
    private suspend fun uploadPayloads(payloads: List<EncryptedCapturePayload>) {
        val uploader = CaptureUploader(config.uploadEndpoint)
        
        for (payload in payloads) {
            val uploadResult = uploader.upload(payload) { progress ->
                updateProgress(progress)
            }
            
            if (!uploadResult.success) {
                handleUploadError(uploadResult.error)
                return
            }
        }
        
        showSuccess("Verification submitted successfully")
    }
}
```

#### iOS (Swift)

```swift
import VirtEngineCaptureSDK

class VerificationViewController: UIViewController {
    private var captureFlow: CaptureFlowExecutor!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        // Configure the capture flow
        let config = CaptureFlowConfig.Builder()
            .clientId("your-approved-client-id")
            .validatorPublicKeys(["validator-pubkey-1", "validator-pubkey-2"])
            .enableDocumentCapture(.idCard)
            .enableSelfieCapture()
            .enableLivenessDetection(.challengeResponse)
            .build()
        
        // Initialize the capture executor
        captureFlow = CaptureFlowExecutor(config: config)
    }
    
    func startVerification() {
        Task {
            do {
                // Execute the capture flow
                let result = try await captureFlow.execute()
                
                // Upload encrypted payloads
                try await uploadPayloads(result.encryptedPayloads)
                
                showSuccess("Verification submitted successfully")
            } catch {
                handleError(error)
            }
        }
    }
    
    private func uploadPayloads(_ payloads: [EncryptedCapturePayload]) async throws {
        let uploader = CaptureUploader(endpoint: config.uploadEndpoint)
        
        for payload in payloads {
            let result = try await uploader.upload(payload) { progress in
                updateProgress(progress)
            }
        }
    }
}
```

## Capture Flow

### Flow Steps

The capture flow consists of the following steps, which can be configured:

1. **Document Front** - Capture the front of an ID document
2. **Document Back** (optional) - Capture the back of an ID document
3. **Selfie** - Capture a selfie photo
4. **Liveness** - Complete liveness challenges (blink, turn head, etc.)

### Configuration Options

```go
// Go SDK configuration example
type CaptureFlowConfig struct {
    // Client identification
    ClientID        string   // Approved client identifier
    ClientVersion   string   // Client application version
    SDKVersion      string   // SDK version
    
    // Encryption recipients
    ValidatorPublicKeys []string  // Validator public keys for encryption
    
    // Steps to include
    IncludeDocumentFront  bool
    IncludeDocumentBack   bool
    IncludeSelfie         bool
    IncludeLiveness       bool
    
    // Document configuration
    DocumentType      DocumentType  // ID_CARD, PASSPORT, DRIVERS_LICENSE
    AllowedCountries  []string      // ISO 3166-1 alpha-2 codes
    
    // Quality requirements
    MinQualityScore   int     // Minimum quality score (0-100)
    MinBlurScore      int     // Minimum blur score (0-100)
    
    // Liveness configuration
    LivenessMode       LivenessMode  // PASSIVE, CHALLENGE_RESPONSE
    LivenessChallenges []Challenge   // BLINK, SMILE, TURN_LEFT, TURN_RIGHT
    LivenessTimeout    time.Duration // Timeout for liveness challenges
    
    // Compression
    CompressionQuality int  // JPEG quality (1-100)
    MaxFileSizeKB      int  // Maximum file size after compression
    
    // Timeouts
    CaptureTimeout     time.Duration
    FlowTimeout        time.Duration
    
    // Upload configuration
    UploadEndpoint     string
    ChunkSizeKB        int
    MaxRetries         int
}
```

### Document Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `ID_CARD` | National ID card | Front + Back |
| `PASSPORT` | Passport | Front only |
| `DRIVERS_LICENSE` | Driver's license | Front + Back |

### Quality Requirements

| Requirement | Threshold | Description |
|-------------|-----------|-------------|
| Blur Score | ≥ 50 | Image sharpness (0-100) |
| Brightness | 40-200 | Luminance level |
| Resolution | ≥ 720p | Minimum image size |
| Face Detection | Required | For selfie/liveness |
| Document Detection | Required | For ID documents |

## Security

### Encryption

All captured payloads are encrypted using the **X25519-XSalsa20-Poly1305** authenticated encryption scheme:

```
Algorithm: X25519-XSalsa20-Poly1305
Key Exchange: X25519 (Curve25519 ECDH)
Encryption: XSalsa20 stream cipher
Authentication: Poly1305 MAC
Nonce: 24 bytes (randomly generated)
Key: 32 bytes
```

**Encryption Flow:**
1. Generate ephemeral X25519 key pair
2. Derive shared secret using recipient's public key
3. Encrypt payload with XSalsa20-Poly1305
4. Package as `EncryptedPayloadEnvelope`

```go
type EncryptedPayloadEnvelope struct {
    RecipientFingerprint string  // Validator's key fingerprint
    Algorithm            string  // "X25519-XSalsa20-Poly1305"
    EphemeralPublicKey   []byte  // Sender's ephemeral public key
    Ciphertext           []byte  // Encrypted payload
    Nonce                []byte  // 24-byte nonce
    PayloadHash          []byte  // SHA256 of plaintext (for verification)
}
```

### Salt Binding

Every capture includes a cryptographic salt binding to prevent replay attacks:

```go
type SaltBinding struct {
    Salt              []byte    // 32-byte random salt
    DeviceFingerprint string    // Device identifier hash
    SessionID         string    // Capture session ID
    Timestamp         time.Time // Capture timestamp
    BindingHash       []byte    // SHA256(salt || device_id || session_id || timestamp)
}
```

**Verification:**
- Server recomputes the binding hash from components
- Mismatch indicates tampering or replay attempt
- Salt uniqueness is verified against replay database

### Dual Signatures

All payloads require two signatures:

1. **Client Signature** - Signed by the approved client application
2. **User Signature** - Signed by the user's wallet

```go
type CaptureSignaturePackage struct {
    ProtocolVersion  int              // Protocol version (currently 1)
    Salt             []byte           // Unique per-upload salt
    SaltBinding      SaltBinding      // Salt binding proof
    ClientSignature  SignatureProof   // Signed by approved client
    UserSignature    SignatureProof   // Signed by user wallet
    PayloadHash      []byte           // Hash of encrypted payload
    Timestamp        time.Time        // Signing timestamp
}
```

### Device Attestation

The SDK performs device integrity checks:

**Android:**
- Play Integrity API (preferred)
- SafetyNet Attestation (deprecated fallback)

**iOS:**
- App Attest (preferred)
- DeviceCheck (fallback)

**Attestation Checks:**
| Check | Description |
|-------|-------------|
| Basic Integrity | Device passes basic checks |
| Device Integrity | Device is not rooted/jailbroken |
| Strong Integrity | Hardware-backed security |
| App Recognized | App is legitimate (Play Store/App Store) |

## Upload Protocol

### Chunked Upload

Large payloads are uploaded in chunks with resume capability:

```
POST /api/v1/capture/upload/init
Content-Type: application/json

{
  "session_id": "session-uuid",
  "payload_size": 5242880,
  "chunk_size": 1048576,
  "payload_hash": "sha256-hex",
  "metadata": { ... }
}

Response:
{
  "upload_id": "upload-uuid",
  "chunk_urls": [
    "/api/v1/capture/upload/chunk/0",
    "/api/v1/capture/upload/chunk/1",
    ...
  ]
}
```

**Chunk Upload:**
```
PUT /api/v1/capture/upload/chunk/{index}
X-Upload-ID: upload-uuid
X-Chunk-Hash: sha256-of-chunk
Content-Type: application/octet-stream

[binary chunk data]
```

**Finalization:**
```
POST /api/v1/capture/upload/finalize
Content-Type: application/json

{
  "upload_id": "upload-uuid",
  "total_chunks": 5,
  "final_hash": "sha256-hex"
}
```

### Retry Strategy

```
Retry Strategy: Exponential backoff with jitter
Base Delay: 1 second
Max Delay: 30 seconds
Max Retries: 5 (configurable)
Jitter: ±20%

Retry Formula:
  delay = min(baseDelay * 2^attempt * (1 + random(-0.2, 0.2)), maxDelay)
```

### Resume Capability

If an upload is interrupted:
1. Client requests upload status with `upload_id`
2. Server returns list of completed chunk indices
3. Client resumes from first incomplete chunk

```
GET /api/v1/capture/upload/status/{upload_id}

Response:
{
  "upload_id": "upload-uuid",
  "status": "in_progress",
  "completed_chunks": [0, 1, 2],
  "total_chunks": 5,
  "expires_at": "2025-01-15T01:00:00Z"
}
```

## Error Handling

### Error Codes

| Code | Description | Action |
|------|-------------|--------|
| `CAMERA_PERMISSION_DENIED` | Camera access denied | Prompt user to enable permissions |
| `IMAGE_TOO_BLURRY` | Image quality too low | Retry capture |
| `FACE_NOT_DETECTED` | No face in frame | Adjust framing |
| `MULTIPLE_FACES_DETECTED` | Multiple faces | Ensure single face |
| `DOCUMENT_NOT_DETECTED` | Document not visible | Adjust framing |
| `LIVENESS_FAILED` | Liveness check failed | Retry liveness |
| `SPOOF_DETECTED` | Spoofing attempt | Security violation |
| `DEVICE_ROOTED` | Device integrity failed | Device not supported |
| `UPLOAD_RETRY_EXHAUSTED` | Max retries exceeded | Check network |
| `SESSION_EXPIRED` | Session timed out | Start new session |
| `SALT_BINDING_MISMATCH` | Replay detected | Security violation |

### User-Friendly Messages

The SDK provides user-friendly error messages:

```kotlin
// Android
when (error.code) {
    ErrorCode.IMAGE_TOO_BLURRY -> 
        "The image is blurry. Please hold your device steady."
    ErrorCode.FACE_NOT_DETECTED -> 
        "Unable to detect your face. Position clearly in frame."
    ErrorCode.DEVICE_ROOTED -> 
        "This app cannot run on modified devices."
}
```

```swift
// iOS
switch error.code {
case .imageTooBlurry:
    "The image is blurry. Please hold your device steady."
case .faceNotDetected:
    "Unable to detect your face. Position clearly in frame."
case .deviceRooted:
    "This app cannot run on modified devices."
}
```

## Testing

### Test Mode

Enable test mode for development:

```kotlin
val config = CaptureFlowConfig.Builder()
    .setTestMode(true)  // Enables mock captures
    .setMockDeviceAttestation(true)  // Skip device checks
    .build()
```

### Sample Payloads

Sample test payloads are provided in `pkg/capture_protocol/testdata/sample_payloads.json`:

- Valid document capture
- Valid selfie with liveness
- Invalid (blurry) document
- Invalid (no face) selfie
- Failed liveness (spoof)
- Salt binding mismatch
- Device integrity failure

### Integration Testing

```bash
# Run integration tests
go test -v ./pkg/capture_protocol/... -tags="integration"

# Run with test fixtures
go test -v ./pkg/capture_protocol/server/... -run TestValidation
```

## API Reference

### CaptureFlowExecutor

```go
// Start a new capture flow
func (e *CaptureFlowExecutor) Start(ctx context.Context) error

// Execute all capture steps
func (e *CaptureFlowExecutor) Execute(ctx context.Context) (*CaptureFlowResult, error)

// Abort the current flow
func (e *CaptureFlowExecutor) Abort() error

// Get current flow status
func (e *CaptureFlowExecutor) GetStatus() FlowStatus
```

### CaptureEncryptor

```go
// Encrypt a capture result
func (e *CaptureEncryptor) EncryptCapture(
    capture *CompressedPayload,
    metadata *CaptureMetadata,
) (*EncryptedCapturePayload, error)

// Encrypt with multiple recipients
func (e *CaptureEncryptor) EncryptForValidators(
    capture *CompressedPayload,
    validatorKeys [][]byte,
) ([]*EncryptedCapturePayload, error)
```

### DeviceIntegrityChecker

```go
// Check device integrity
func (c *DeviceIntegrityChecker) Check(
    ctx context.Context,
    nonce []byte,
) (*DeviceAttestationResult, error)

// Evaluate attestation result
func (c *DeviceIntegrityChecker) EvaluateAttestation(
    result *DeviceAttestationResult,
    requirements *IntegrityRequirements,
) (*IntegrityEvaluation, error)
```

### ServerValidator

```go
// Validate an upload request
func (v *ServerValidator) ValidateUploadRequest(
    request *CaptureUploadRequest,
    accountAddress string,
) *ValidationResult

// Validate signature package
func (v *ServerValidator) ValidateSignatures(
    signatures *CaptureSignaturePackage,
) (*SignatureValidationResult, error)
```

## Changelog

### v1.0.0 (2025-01-15)
- Initial release
- Document capture (ID card, passport, driver's license)
- Selfie capture with quality checks
- Liveness detection (challenge-response)
- X25519-XSalsa20-Poly1305 encryption
- Dual signature scheme
- Device attestation (Play Integrity, App Attest)
- Chunked upload with resume
- Salt binding for replay prevention

## Support

- **Documentation:** https://docs.virtengine.io/veid/mobile-sdk
- **GitHub Issues:** https://github.com/virtengine/veid-capture-sdk/issues
- **Discord:** https://discord.gg/virtengine
