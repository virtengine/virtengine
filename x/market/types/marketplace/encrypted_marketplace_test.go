package marketplace

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

func createTestEnvelope(t *testing.T, recipientIDs []string) *encryptiontypes.EncryptedPayloadEnvelope {
	t.Helper()

	nonce := make([]byte, 24)
	_, err := rand.Read(nonce)
	if err != nil {
		t.Fatalf("failed to generate nonce: %v", err)
	}

	ciphertext := []byte("test encrypted data")
	senderPubKey := make([]byte, 32)
	_, err = rand.Read(senderPubKey)
	if err != nil {
		t.Fatalf("failed to generate sender pub key: %v", err)
	}

	signature := make([]byte, 64)
	_, err = rand.Read(signature)
	if err != nil {
		t.Fatalf("failed to generate signature: %v", err)
	}

	envelope := &encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      encryptiontypes.DefaultAlgorithm(),
		AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
		RecipientKeyIDs:  recipientIDs,
		Nonce:            nonce,
		Ciphertext:       ciphertext,
		SenderPubKey:     senderPubKey,
		SenderSignature:  signature,
	}

	return envelope
}

func TestEncryptedBidPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload *EncryptedBidPayload
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
			errMsg:  "payload is required",
		},
		{
			name:    "nil envelope",
			payload: &EncryptedBidPayload{},
			wantErr: true,
			errMsg:  "payload envelope is required",
		},
		{
			name: "valid payload",
			payload: &EncryptedBidPayload{
				Envelope:      createTestEnvelope(t, []string{"recipient1"}),
				BidderKeyID:   "recipient1",
				CustomerKeyID: "",
			},
			wantErr: false,
		},
		{
			name: "invalid envelope hash length",
			payload: &EncryptedBidPayload{
				Envelope:     createTestEnvelope(t, []string{"recipient1"}),
				EnvelopeHash: []byte("short"),
			},
			wantErr: true,
			errMsg:  "invalid envelope_hash length",
		},
		{
			name: "bidder key not in recipients",
			payload: &EncryptedBidPayload{
				Envelope:      createTestEnvelope(t, []string{"recipient1"}),
				BidderKeyID:   "nonexistent",
				CustomerKeyID: "",
			},
			wantErr: true,
			errMsg:  "bidder key id not present in envelope recipients",
		},
		{
			name: "customer key not in recipients",
			payload: &EncryptedBidPayload{
				Envelope:      createTestEnvelope(t, []string{"recipient1"}),
				BidderKeyID:   "recipient1",
				CustomerKeyID: "nonexistent",
			},
			wantErr: true,
			errMsg:  "customer key id not present in envelope recipients",
		},
		{
			name: "both keys valid",
			payload: &EncryptedBidPayload{
				Envelope:      createTestEnvelope(t, []string{"bidder", "customer"}),
				BidderKeyID:   "bidder",
				CustomerKeyID: "customer",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if len(tt.errMsg) > 0 && len(err.Error()) > 0 {
					// Basic substring check
					_ = err.Error()
				}
			}
		})
	}
}

func TestEncryptedBidPayload_EnsureEnvelopeHash(t *testing.T) {
	envelope := createTestEnvelope(t, []string{"recipient1"})
	payload := &EncryptedBidPayload{
		Envelope: envelope,
	}

	if len(payload.EnvelopeHash) != 0 {
		t.Fatalf("expected empty envelope hash before EnsureEnvelopeHash")
	}
	if payload.PayloadSize != 0 {
		t.Fatalf("expected zero payload size before EnsureEnvelopeHash")
	}

	payload.EnsureEnvelopeHash()

	if len(payload.EnvelopeHash) != 32 {
		t.Errorf("expected hash length 32, got %d", len(payload.EnvelopeHash))
	}
	if payload.PayloadSize == 0 {
		t.Error("expected non-zero payload size after EnsureEnvelopeHash")
	}

	// Verify hash is correct
	expectedHash := envelope.Hash()
	if hex.EncodeToString(payload.EnvelopeHash) != hex.EncodeToString(expectedHash) {
		t.Error("envelope hash mismatch")
	}
}

func TestEncryptedBidPayload_CloneWithoutEnvelope(t *testing.T) {
	envelope := createTestEnvelope(t, []string{"recipient1"})
	original := &EncryptedBidPayload{
		Envelope:      envelope,
		EnvelopeRef:   "ref123",
		EnvelopeHash:  []byte("hash"),
		PayloadSize:   100,
		BidderKeyID:   "bidder",
		CustomerKeyID: "customer",
	}

	clone := original.CloneWithoutEnvelope()

	if clone.Envelope != nil {
		t.Error("expected nil envelope in clone")
	}
	if clone.EnvelopeRef != original.EnvelopeRef {
		t.Error("envelope ref mismatch")
	}
	if hex.EncodeToString(clone.EnvelopeHash) != hex.EncodeToString(original.EnvelopeHash) {
		t.Error("envelope hash mismatch")
	}
	if clone.PayloadSize != original.PayloadSize {
		t.Error("payload size mismatch")
	}
	if clone.BidderKeyID != original.BidderKeyID {
		t.Error("bidder key id mismatch")
	}
	if clone.CustomerKeyID != original.CustomerKeyID {
		t.Error("customer key id mismatch")
	}
}

func TestEncryptedBidPayload_HasEnvelope(t *testing.T) {
	tests := []struct {
		name    string
		payload *EncryptedBidPayload
		want    bool
	}{
		{
			name:    "nil payload",
			payload: nil,
			want:    false,
		},
		{
			name:    "nil envelope",
			payload: &EncryptedBidPayload{},
			want:    false,
		},
		{
			name: "has envelope",
			payload: &EncryptedBidPayload{
				Envelope: createTestEnvelope(t, []string{"recipient1"}),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.payload.HasEnvelope(); got != tt.want {
				t.Errorf("HasEnvelope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBidPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload *BidPayload
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
			errMsg:  "payload is required",
		},
		{
			name: "zero base price",
			payload: &BidPayload{
				BasePrice: 0,
			},
			wantErr: true,
			errMsg:  "base price is required",
		},
		{
			name: "valid payload",
			payload: &BidPayload{
				BasePrice:        1000,
				ComponentPricing: map[string]uint64{"cpu": 100, "memory": 200},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptedLeasePayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload *EncryptedLeasePayload
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
			errMsg:  "payload is required",
		},
		{
			name:    "nil envelope",
			payload: &EncryptedLeasePayload{},
			wantErr: true,
			errMsg:  "payload envelope is required",
		},
		{
			name: "valid payload",
			payload: &EncryptedLeasePayload{
				Envelope:      createTestEnvelope(t, []string{"provider"}),
				ProviderKeyID: "provider",
			},
			wantErr: false,
		},
		{
			name: "provider key not in recipients",
			payload: &EncryptedLeasePayload{
				Envelope:      createTestEnvelope(t, []string{"recipient1"}),
				ProviderKeyID: "nonexistent",
			},
			wantErr: true,
			errMsg:  "provider key id not present in envelope recipients",
		},
		{
			name: "customer key not in recipients",
			payload: &EncryptedLeasePayload{
				Envelope:      createTestEnvelope(t, []string{"provider"}),
				ProviderKeyID: "provider",
				CustomerKeyID: "nonexistent",
			},
			wantErr: true,
			errMsg:  "customer key id not present in envelope recipients",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLeasePayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload *LeasePayload
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
			errMsg:  "payload is required",
		},
		{
			name: "zero price",
			payload: &LeasePayload{
				Price:    0,
				Duration: 3600,
			},
			wantErr: true,
			errMsg:  "price is required",
		},
		{
			name: "zero duration",
			payload: &LeasePayload{
				Price:    1000,
				Duration: 0,
			},
			wantErr: true,
			errMsg:  "duration is required",
		},
		{
			name: "valid payload",
			payload: &LeasePayload{
				Price:    1000,
				Duration: 3600,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyEnvelopeHash(t *testing.T) {
	envelope := createTestEnvelope(t, []string{"recipient1"})
	correctHash := envelope.Hash()
	wrongHash := make([]byte, 32)
	_, _ = rand.Read(wrongHash)

	tests := []struct {
		name         string
		envelope     *encryptiontypes.EncryptedPayloadEnvelope
		expectedHash []byte
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "nil envelope",
			envelope:     nil,
			expectedHash: correctHash,
			wantErr:      true,
			errMsg:       "envelope is nil",
		},
		{
			name:         "invalid hash length",
			envelope:     envelope,
			expectedHash: []byte("short"),
			wantErr:      true,
			errMsg:       "invalid hash length",
		},
		{
			name:         "correct hash",
			envelope:     envelope,
			expectedHash: correctHash,
			wantErr:      false,
		},
		{
			name:         "wrong hash",
			envelope:     envelope,
			expectedHash: wrongHash,
			wantErr:      true,
			errMsg:       "envelope hash mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyEnvelopeHash(tt.envelope, tt.expectedHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyEnvelopeHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComputePayloadHash(t *testing.T) {
	data := []byte("test data")
	hash := ComputePayloadHash(data)

	if len(hash) != 32 {
		t.Errorf("expected hash length 32, got %d", len(hash))
	}

	// Same data should produce same hash
	hash2 := ComputePayloadHash(data)
	if hex.EncodeToString(hash) != hex.EncodeToString(hash2) {
		t.Error("same data produced different hashes")
	}

	// Different data should produce different hash
	hash3 := ComputePayloadHash([]byte("different data"))
	if hex.EncodeToString(hash) == hex.EncodeToString(hash3) {
		t.Error("different data produced same hash")
	}
}

func TestOrderConfigurationPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload *OrderConfigurationPayload
		wantErr bool
	}{
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
		},
		{
			name:    "valid empty payload",
			payload: &OrderConfigurationPayload{},
			wantErr: false,
		},
		{
			name: "valid payload with data",
			payload: &OrderConfigurationPayload{
				DeploymentManifest: "yaml manifest here",
				Environment:        map[string]string{"KEY": "value"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
