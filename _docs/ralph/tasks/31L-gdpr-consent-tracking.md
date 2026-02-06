# Task 31L: GDPR Consent Tracking System

**vibe-kanban ID:** `5802be8d-1c37-410b-a05b-64b70f340c01`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31L |
| **Title** | feat(compliance): GDPR consent tracking |
| **Priority** | P1 |
| **Wave** | 2 |
| **Estimated LOC** | 3500 |
| **Duration** | 3 weeks |
| **Dependencies** | VEID module |
| **Blocking** | None |

---

## Problem Statement

GDPR compliance requires explicit consent tracking for:
- Biometric data processing (VEID facial/liveness scopes)
- Data retention agreements
- Marketing communications
- Third-party data sharing

Current state:
- GDPR_COMPLIANCE.md exists as policy documentation
- No programmatic consent tracking
- No consent withdrawal mechanism
- No data export/deletion capabilities

### Current State Analysis

```
GDPR_COMPLIANCE.md              ✅ Policy documentation
CONSENT_FRAMEWORK.md            ✅ Framework documentation
x/veid/                         ⚠️  No consent linkage
pkg/gdpr/                       ❌ Does not exist
portal/consent/                 ❌ No consent UI
```

---

## Acceptance Criteria

### AC-1: Consent Management Module
- [ ] On-chain consent records (hash references)
- [ ] Off-chain detailed consent storage
- [ ] Consent version tracking
- [ ] Consent withdrawal mechanism
- [ ] Consent proof generation

### AC-2: VEID Consent Integration
- [ ] Require biometric consent before VEID submission
- [ ] Link VEID records to consent records
- [ ] Block processing without valid consent
- [ ] Handle consent withdrawal (data deletion cascade)

### AC-3: Data Subject Rights
- [ ] Right to Access (data export)
- [ ] Right to be Forgotten (deletion request)
- [ ] Right to Rectification (data correction)
- [ ] Right to Data Portability (machine-readable export)
- [ ] Processing audit trail

### AC-4: Consent UI
- [ ] Consent collection forms
- [ ] Consent management dashboard
- [ ] Withdrawal interface
- [ ] Data export request interface
- [ ] Privacy preference center

---

## Technical Requirements

### Consent Module Types

```go
// x/consent/types/consent.go

package types

import (
    "time"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type ConsentPurpose string

const (
    PurposeBiometricProcessing ConsentPurpose = "biometric_processing"
    PurposeDataRetention       ConsentPurpose = "data_retention"
    PurposeThirdPartySharing   ConsentPurpose = "third_party_sharing"
    PurposeMarketing           ConsentPurpose = "marketing"
    PurposeAnalytics           ConsentPurpose = "analytics"
)

type ConsentRecord struct {
    ID              string         `json:"id"`
    DataSubject     string         `json:"data_subject"`  // Account address
    Purpose         ConsentPurpose `json:"purpose"`
    Version         string         `json:"version"`       // Policy version
    Status          ConsentStatus  `json:"status"`
    GrantedAt       time.Time      `json:"granted_at"`
    ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
    WithdrawnAt     *time.Time     `json:"withdrawn_at,omitempty"`
    
    // Evidence
    ConsentHash     []byte         `json:"consent_hash"`      // Hash of full consent text
    SignatureHash   []byte         `json:"signature_hash"`    // Hash of user signature
    IPAddressHash   []byte         `json:"ip_address_hash"`   // HMAC of IP (for proof, not tracking)
    
    // Off-chain reference
    DetailedRecordRef string       `json:"detailed_record_ref"`
}

type ConsentStatus string

const (
    StatusActive     ConsentStatus = "active"
    StatusWithdrawn  ConsentStatus = "withdrawn"
    StatusExpired    ConsentStatus = "expired"
)

// ConsentProof provides verifiable proof of consent
type ConsentProof struct {
    ConsentID     string
    DataSubject   string
    Purpose       ConsentPurpose
    Version       string
    GrantedAt     time.Time
    ConsentHash   []byte
    MerkleRoot    []byte
    MerklePath    [][]byte
    BlockHeight   int64
    TxHash        string
}
```

### Consent Keeper

```go
// x/consent/keeper/keeper.go

package keeper

import (
    "context"
    "crypto/sha256"
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    storetypes "github.com/cosmos/cosmos-sdk/store/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "github.com/virtengine/virtengine/x/consent/types"
)

type Keeper struct {
    cdc       codec.BinaryCodec
    storeKey  storetypes.StoreKey
    veidKeeper VEIDKeeper
}

type VEIDKeeper interface {
    HasActiveIdentity(ctx sdk.Context, addr string) bool
    DeleteUserData(ctx sdk.Context, addr string) error
}

func (k Keeper) GrantConsent(ctx context.Context, msg *types.MsgGrantConsent) (*types.MsgGrantConsentResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    
    // Verify signature
    if !k.verifyConsentSignature(msg) {
        return nil, types.ErrInvalidSignature
    }
    
    // Create consent record
    consentID := generateConsentID(msg.DataSubject, msg.Purpose, sdkCtx.BlockHeight())
    
    record := types.ConsentRecord{
        ID:          consentID,
        DataSubject: msg.DataSubject,
        Purpose:     msg.Purpose,
        Version:     msg.PolicyVersion,
        Status:      types.StatusActive,
        GrantedAt:   sdkCtx.BlockTime(),
        ExpiresAt:   calculateExpiry(msg.Purpose),
        ConsentHash: sha256Hash(msg.ConsentText),
        SignatureHash: sha256Hash(msg.Signature),
        IPAddressHash: hmacHash(msg.IPAddress, k.getHMACKey()),
    }
    
    // Store on-chain
    store := sdkCtx.KVStore(k.storeKey)
    bz := k.cdc.MustMarshal(&record)
    store.Set(types.ConsentKey(consentID), bz)
    
    // Index by data subject and purpose
    k.setConsentIndex(sdkCtx, msg.DataSubject, msg.Purpose, consentID)
    
    // Emit event
    sdkCtx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeConsentGranted,
            sdk.NewAttribute(types.AttributeKeyConsentID, consentID),
            sdk.NewAttribute(types.AttributeKeyDataSubject, msg.DataSubject),
            sdk.NewAttribute(types.AttributeKeyPurpose, string(msg.Purpose)),
        ),
    )
    
    return &types.MsgGrantConsentResponse{ConsentId: consentID}, nil
}

func (k Keeper) WithdrawConsent(ctx context.Context, msg *types.MsgWithdrawConsent) (*types.MsgWithdrawConsentResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    
    // Get existing consent
    record, found := k.GetConsent(sdkCtx, msg.ConsentId)
    if !found {
        return nil, types.ErrConsentNotFound
    }
    
    // Verify ownership
    if record.DataSubject != msg.DataSubject {
        return nil, types.ErrUnauthorized
    }
    
    // Update status
    now := sdkCtx.BlockTime()
    record.Status = types.StatusWithdrawn
    record.WithdrawnAt = &now
    
    // Save
    store := sdkCtx.KVStore(k.storeKey)
    bz := k.cdc.MustMarshal(&record)
    store.Set(types.ConsentKey(msg.ConsentId), bz)
    
    // Trigger data cleanup if biometric consent withdrawn
    if record.Purpose == types.PurposeBiometricProcessing {
        go k.initiateDataDeletion(record.DataSubject)
    }
    
    return &types.MsgWithdrawConsentResponse{}, nil
}

func (k Keeper) HasValidConsent(ctx sdk.Context, dataSubject string, purpose types.ConsentPurpose) bool {
    consentID, found := k.getConsentBySubjectPurpose(ctx, dataSubject, purpose)
    if !found {
        return false
    }
    
    record, found := k.GetConsent(ctx, consentID)
    if !found {
        return false
    }
    
    if record.Status != types.StatusActive {
        return false
    }
    
    if record.ExpiresAt != nil && ctx.BlockTime().After(*record.ExpiresAt) {
        return false
    }
    
    return true
}

func (k Keeper) GenerateConsentProof(ctx sdk.Context, consentID string) (*types.ConsentProof, error) {
    record, found := k.GetConsent(ctx, consentID)
    if !found {
        return nil, types.ErrConsentNotFound
    }
    
    // Generate Merkle proof from consent store
    merkleRoot, merklePath := k.generateMerkleProof(ctx, consentID)
    
    return &types.ConsentProof{
        ConsentID:   consentID,
        DataSubject: record.DataSubject,
        Purpose:     record.Purpose,
        Version:     record.Version,
        GrantedAt:   record.GrantedAt,
        ConsentHash: record.ConsentHash,
        MerkleRoot:  merkleRoot,
        MerklePath:  merklePath,
        BlockHeight: ctx.BlockHeight(),
    }, nil
}
```

### VEID Consent Integration

```go
// x/veid/keeper/consent.go

package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    consenttypes "github.com/virtengine/virtengine/x/consent/types"
    "github.com/virtengine/virtengine/x/veid/types"
)

// PreSubmitScopeHook validates consent before scope submission
func (k Keeper) PreSubmitScopeHook(ctx sdk.Context, msg *types.MsgSubmitScope) error {
    // Check biometric processing consent
    if !k.consentKeeper.HasValidConsent(ctx, msg.Signer, consenttypes.PurposeBiometricProcessing) {
        return types.ErrMissingBiometricConsent
    }
    
    // Check data retention consent
    if !k.consentKeeper.HasValidConsent(ctx, msg.Signer, consenttypes.PurposeDataRetention) {
        return types.ErrMissingRetentionConsent
    }
    
    return nil
}

// HandleConsentWithdrawal processes consent withdrawal cascade
func (k Keeper) HandleConsentWithdrawal(ctx context.Context, dataSubject string) error {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    
    // Get all user data
    identity, found := k.GetIdentityRecord(sdkCtx, dataSubject)
    if !found {
        return nil // No data to delete
    }
    
    // Create deletion request record (for audit)
    deletionID := k.createDeletionRequest(sdkCtx, dataSubject)
    
    // Delete VEID data
    if err := k.deleteIdentityData(sdkCtx, dataSubject); err != nil {
        return err
    }
    
    // Delete scopes
    if err := k.deleteAllScopes(sdkCtx, dataSubject); err != nil {
        return err
    }
    
    // Emit deletion event
    sdkCtx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeDataDeleted,
            sdk.NewAttribute(types.AttributeKeyAddress, dataSubject),
            sdk.NewAttribute(types.AttributeKeyDeletionID, deletionID),
        ),
    )
    
    return nil
}
```

### Data Rights Service

```go
// pkg/gdpr/rights.go

package gdpr

import (
    "context"
    "encoding/json"
    "time"
)

type DataRightsService struct {
    consentKeeper ConsentKeeper
    veidKeeper    VEIDKeeper
    escrowKeeper  EscrowKeeper
    exportStore   ExportStore
}

type DataExportRequest struct {
    ID            string
    DataSubject   string
    RequestedAt   time.Time
    Status        ExportStatus
    Format        ExportFormat
    DownloadURL   string
    ExpiresAt     time.Time
}

type ExportStatus string

const (
    ExportPending   ExportStatus = "pending"
    ExportProcessing ExportStatus = "processing"
    ExportReady     ExportStatus = "ready"
    ExportExpired   ExportStatus = "expired"
)

type ExportFormat string

const (
    FormatJSON ExportFormat = "json"
    FormatCSV  ExportFormat = "csv"
)

// RequestDataExport initiates a data export (Right to Access / Portability)
func (s *DataRightsService) RequestDataExport(ctx context.Context, dataSubject string, format ExportFormat) (*DataExportRequest, error) {
    request := &DataExportRequest{
        ID:          generateRequestID(),
        DataSubject: dataSubject,
        RequestedAt: time.Now(),
        Status:      ExportPending,
        Format:      format,
    }
    
    // Store request
    if err := s.exportStore.SaveRequest(ctx, request); err != nil {
        return nil, err
    }
    
    // Process asynchronously
    go s.processExport(ctx, request)
    
    return request, nil
}

func (s *DataRightsService) processExport(ctx context.Context, request *DataExportRequest) {
    request.Status = ExportProcessing
    s.exportStore.SaveRequest(ctx, request)
    
    // Collect all user data
    data := &UserDataExport{
        ExportedAt:  time.Now(),
        DataSubject: request.DataSubject,
    }
    
    // VEID data
    identity, _ := s.veidKeeper.GetIdentityRecord(ctx, request.DataSubject)
    if identity != nil {
        data.Identity = &IdentityExport{
            Address:      identity.Address,
            TrustScore:   identity.TrustScore,
            TierLevel:    identity.TierLevel,
            CreatedAt:    identity.CreatedAt,
            Scopes:       s.exportScopes(ctx, request.DataSubject),
        }
    }
    
    // Consent history
    data.Consents = s.consentKeeper.GetAllConsents(ctx, request.DataSubject)
    
    // Transaction history
    data.Transactions = s.getTransactionHistory(ctx, request.DataSubject)
    
    // Escrow records
    data.EscrowRecords = s.escrowKeeper.GetUserEscrows(ctx, request.DataSubject)
    
    // Serialize and store
    var exportData []byte
    var err error
    switch request.Format {
    case FormatJSON:
        exportData, err = json.MarshalIndent(data, "", "  ")
    case FormatCSV:
        exportData, err = s.convertToCSV(data)
    }
    
    if err != nil {
        request.Status = "failed"
        s.exportStore.SaveRequest(ctx, request)
        return
    }
    
    // Upload to secure storage
    downloadURL, err := s.uploadExport(ctx, request.ID, exportData)
    if err != nil {
        request.Status = "failed"
        s.exportStore.SaveRequest(ctx, request)
        return
    }
    
    request.Status = ExportReady
    request.DownloadURL = downloadURL
    request.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
    s.exportStore.SaveRequest(ctx, request)
    
    // Notify user
    s.notifyExportReady(request)
}

type UserDataExport struct {
    ExportedAt    time.Time
    DataSubject   string
    Identity      *IdentityExport
    Consents      []ConsentExport
    Transactions  []TransactionExport
    EscrowRecords []EscrowExport
}

// RequestDataDeletion initiates a deletion request (Right to be Forgotten)
func (s *DataRightsService) RequestDataDeletion(ctx context.Context, dataSubject string) (*DeletionRequest, error) {
    request := &DeletionRequest{
        ID:          generateRequestID(),
        DataSubject: dataSubject,
        RequestedAt: time.Now(),
        Status:      DeletionPending,
    }
    
    // Check for blockers (active leases, pending escrow, etc.)
    blockers, err := s.checkDeletionBlockers(ctx, dataSubject)
    if err != nil {
        return nil, err
    }
    
    if len(blockers) > 0 {
        request.Status = DeletionBlocked
        request.Blockers = blockers
        s.exportStore.SaveDeletionRequest(ctx, request)
        return request, nil
    }
    
    // Process deletion
    if err := s.executeDataDeletion(ctx, dataSubject); err != nil {
        request.Status = DeletionFailed
        request.Error = err.Error()
        s.exportStore.SaveDeletionRequest(ctx, request)
        return request, err
    }
    
    request.Status = DeletionComplete
    request.CompletedAt = timePtr(time.Now())
    s.exportStore.SaveDeletionRequest(ctx, request)
    
    return request, nil
}
```

### Portal Consent Components

```tsx
// portal/src/components/consent/ConsentManager.tsx

'use client';

import { useState, useEffect } from 'react';
import { useVirtEngine } from '@virtengine/portal';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';

interface ConsentItem {
  id: string;
  purpose: string;
  status: 'active' | 'withdrawn';
  grantedAt: string;
  version: string;
}

const CONSENT_PURPOSES = [
  {
    key: 'biometric_processing',
    title: 'Biometric Data Processing',
    description: 'Allow processing of facial and liveness data for identity verification.',
    required: true,
  },
  {
    key: 'data_retention',
    title: 'Data Retention',
    description: 'Allow retention of verification data for the legally required period.',
    required: true,
  },
  {
    key: 'analytics',
    title: 'Analytics',
    description: 'Allow aggregated, anonymized data for service improvement.',
    required: false,
  },
  {
    key: 'marketing',
    title: 'Marketing Communications',
    description: 'Receive updates about new features and services.',
    required: false,
  },
];

export function ConsentManager() {
  const { address, signMessage } = useVirtEngine();
  const [consents, setConsents] = useState<ConsentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);

  useEffect(() => {
    loadConsents();
  }, [address]);

  const loadConsents = async () => {
    const res = await fetch(`/api/consent/${address}`);
    const data = await res.json();
    setConsents(data.consents);
    setLoading(false);
  };

  const handleConsentToggle = async (purpose: string, enabled: boolean) => {
    setUpdating(purpose);
    try {
      if (enabled) {
        // Grant consent
        const consentText = getConsentText(purpose);
        const signature = await signMessage(consentText);
        
        await fetch('/api/consent/grant', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            dataSubject: address,
            purpose,
            consentText,
            signature,
          }),
        });
      } else {
        // Withdraw consent
        const consent = consents.find(c => c.purpose === purpose);
        if (consent) {
          await fetch('/api/consent/withdraw', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              consentId: consent.id,
              dataSubject: address,
            }),
          });
        }
      }
      await loadConsents();
    } catch (error) {
      console.error('Consent update failed:', error);
    } finally {
      setUpdating(null);
    }
  };

  const isConsentActive = (purpose: string) => {
    return consents.some(c => c.purpose === purpose && c.status === 'active');
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Privacy Preferences</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {CONSENT_PURPOSES.map(purpose => (
            <div key={purpose.key} className="flex items-start justify-between pb-4 border-b last:border-0">
              <div className="flex-1 pr-4">
                <div className="flex items-center gap-2">
                  <h4 className="font-medium">{purpose.title}</h4>
                  {purpose.required && (
                    <span className="text-xs text-red-500">Required</span>
                  )}
                </div>
                <p className="text-sm text-gray-500 mt-1">{purpose.description}</p>
              </div>
              <Switch
                checked={isConsentActive(purpose.key)}
                onCheckedChange={(checked) => handleConsentToggle(purpose.key, checked)}
                disabled={updating === purpose.key || (purpose.required && isConsentActive(purpose.key))}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Your Data Rights</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h4 className="font-medium">Download Your Data</h4>
              <p className="text-sm text-gray-500">Get a copy of all your data</p>
            </div>
            <Button variant="outline" onClick={() => requestDataExport()}>
              Request Export
            </Button>
          </div>
          
          <div className="flex justify-between items-center">
            <div>
              <h4 className="font-medium">Delete Your Account</h4>
              <p className="text-sm text-gray-500">Permanently delete all your data</p>
            </div>
            <Button variant="destructive" onClick={() => requestDataDeletion()}>
              Request Deletion
            </Button>
          </div>
        </CardContent>
      </Card>

      {consents.some(c => c.status === 'withdrawn') && (
        <Alert>
          <AlertDescription>
            Some consents have been withdrawn. This may limit certain features.
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
```

---

## Directory Structure

```
x/consent/
├── types/
│   ├── consent.go
│   ├── msgs.go
│   ├── keys.go
│   └── errors.go
├── keeper/
│   ├── keeper.go
│   ├── msg_server.go
│   └── query_server.go
└── module.go

pkg/gdpr/
├── rights.go             # Data rights service
├── export.go             # Data export functionality
├── deletion.go           # Data deletion handler
└── audit.go              # Processing audit trail

portal/src/
├── app/privacy/
│   └── page.tsx          # Privacy center
└── components/consent/
    ├── ConsentManager.tsx
    ├── ConsentHistory.tsx
    └── DataExportStatus.tsx
```

---

## Testing Requirements

### Unit Tests
- Consent grant/withdrawal flow
- Consent validation
- Data export generation

### Integration Tests
- VEID submission with consent check
- Consent withdrawal cascade
- Data export download

### Compliance Tests
- 30-day response requirement
- Complete data inclusion
- Proper data anonymization

---

## Security Considerations

1. **Consent Proof**: On-chain hash ensures consent cannot be forged
2. **IP Privacy**: Hash IP addresses, never store raw
3. **Export Security**: Encrypted exports, time-limited URLs
4. **Deletion Verification**: Audit trail of deletions
5. **Access Control**: Only data subject can manage their consent
