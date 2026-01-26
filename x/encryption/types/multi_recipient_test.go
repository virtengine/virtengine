package types

import (
	"testing"
)

func TestMultiRecipientEnvelope_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*MultiRecipientEnvelope)
		wantErr bool
	}{
		{
			name:    "valid envelope",
			modify:  func(e *MultiRecipientEnvelope) {},
			wantErr: false,
		},
		{
			name: "zero version",
			modify: func(e *MultiRecipientEnvelope) {
				e.Version = 0
			},
			wantErr: true,
		},
		{
			name: "unsupported version",
			modify: func(e *MultiRecipientEnvelope) {
				e.Version = 999
			},
			wantErr: true,
		},
		{
			name: "empty payload ciphertext",
			modify: func(e *MultiRecipientEnvelope) {
				e.PayloadCiphertext = nil
			},
			wantErr: true,
		},
		{
			name: "empty payload nonce",
			modify: func(e *MultiRecipientEnvelope) {
				e.PayloadNonce = nil
			},
			wantErr: true,
		},
		{
			name: "no wrapped keys",
			modify: func(e *MultiRecipientEnvelope) {
				e.WrappedKeys = nil
			},
			wantErr: true,
		},
		{
			name: "empty recipient ID in wrapped key",
			modify: func(e *MultiRecipientEnvelope) {
				e.WrappedKeys[0].RecipientID = ""
			},
			wantErr: true,
		},
		{
			name: "empty wrapped key",
			modify: func(e *MultiRecipientEnvelope) {
				e.WrappedKeys[0].WrappedKey = nil
			},
			wantErr: true,
		},
		{
			name: "duplicate recipient IDs",
			modify: func(e *MultiRecipientEnvelope) {
				e.WrappedKeys = append(e.WrappedKeys, WrappedKeyEntry{
					RecipientID: "recipient1",
					WrappedKey:  []byte("wrapped_key_duplicate"),
				})
			},
			wantErr: true,
		},
		{
			name: "empty client signature",
			modify: func(e *MultiRecipientEnvelope) {
				e.ClientSignature = nil
			},
			wantErr: true,
		},
		{
			name: "empty client ID",
			modify: func(e *MultiRecipientEnvelope) {
				e.ClientID = ""
			},
			wantErr: true,
		},
		{
			name: "empty user signature",
			modify: func(e *MultiRecipientEnvelope) {
				e.UserSignature = nil
			},
			wantErr: true,
		},
		{
			name: "empty user public key",
			modify: func(e *MultiRecipientEnvelope) {
				e.UserPubKey = nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := createValidMultiRecipientEnvelope()
			tt.modify(envelope)

			err := envelope.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultiRecipientEnvelope_GetWrappedKey(t *testing.T) {
	envelope := createValidMultiRecipientEnvelope()

	// Test finding existing recipient
	key, found := envelope.GetWrappedKey("recipient1")
	if !found {
		t.Error("expected to find recipient1")
	}
	if string(key) != "wrapped_key_1" {
		t.Errorf("unexpected wrapped key: %s", string(key))
	}

	// Test not finding non-existent recipient
	_, found = envelope.GetWrappedKey("nonexistent")
	if found {
		t.Error("expected not to find nonexistent recipient")
	}
}

func TestMultiRecipientEnvelope_RecipientIDs(t *testing.T) {
	envelope := createValidMultiRecipientEnvelope()
	envelope.WrappedKeys = append(envelope.WrappedKeys, WrappedKeyEntry{
		RecipientID: "recipient2",
		WrappedKey:  []byte("wrapped_key_2"),
	})
	envelope.WrappedKeys = append(envelope.WrappedKeys, WrappedKeyEntry{
		RecipientID: "recipient0",
		WrappedKey:  []byte("wrapped_key_0"),
	})

	ids := envelope.RecipientIDs()
	if len(ids) != 3 {
		t.Fatalf("expected 3 IDs, got %d", len(ids))
	}

	// Should be sorted
	expected := []string{"recipient0", "recipient1", "recipient2"}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, id)
		}
	}
}

func TestMultiRecipientEnvelope_DeterministicBytes(t *testing.T) {
	envelope1 := createValidMultiRecipientEnvelope()
	envelope2 := createValidMultiRecipientEnvelope()

	// Add recipients in different orders
	envelope1.WrappedKeys = append(envelope1.WrappedKeys, WrappedKeyEntry{
		RecipientID: "recipient2",
		WrappedKey:  []byte("wrapped_key_2"),
	})
	envelope1.WrappedKeys = append(envelope1.WrappedKeys, WrappedKeyEntry{
		RecipientID: "recipient3",
		WrappedKey:  []byte("wrapped_key_3"),
	})

	envelope2.WrappedKeys = append(envelope2.WrappedKeys, WrappedKeyEntry{
		RecipientID: "recipient3",
		WrappedKey:  []byte("wrapped_key_3"),
	})
	envelope2.WrappedKeys = append(envelope2.WrappedKeys, WrappedKeyEntry{
		RecipientID: "recipient2",
		WrappedKey:  []byte("wrapped_key_2"),
	})

	bytes1, err := envelope1.DeterministicBytes()
	if err != nil {
		t.Fatalf("DeterministicBytes() error: %v", err)
	}

	bytes2, err := envelope2.DeterministicBytes()
	if err != nil {
		t.Fatalf("DeterministicBytes() error: %v", err)
	}

	// Should produce identical bytes regardless of order
	if string(bytes1) != string(bytes2) {
		t.Error("DeterministicBytes() should produce identical bytes for same content in different order")
	}
}

func TestMultiRecipientEnvelopeBuilder(t *testing.T) {
	envelope, err := NewMultiRecipientEnvelopeBuilder().
		WithAlgorithm(DefaultAlgorithm()).
		WithRecipientMode(RecipientModeFullValidatorSet).
		WithPayload([]byte("ciphertext"), []byte("nonce12345678901234567890123456")).
		AddRecipient("validator1", []byte("wrapped_key_1")).
		AddRecipient("validator2", []byte("wrapped_key_2")).
		WithClientSignature("approved_client_1", []byte("client_sig")).
		WithUserSignature([]byte("user_pubkey_32bytes_exactly_now!"), []byte("user_sig")).
		WithMetadata("scope_type", "id_document").
		Build()

	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if envelope.RecipientCount() != 2 {
		t.Errorf("expected 2 recipients, got %d", envelope.RecipientCount())
	}

	if !envelope.HasRecipient("validator1") {
		t.Error("expected validator1 to be a recipient")
	}

	if envelope.Metadata["scope_type"] != "id_document" {
		t.Error("expected metadata to be set")
	}
}

func createValidMultiRecipientEnvelope() *MultiRecipientEnvelope {
	return &MultiRecipientEnvelope{
		Version:           MultiRecipientEnvelopeVersion,
		AlgorithmID:       DefaultAlgorithm(),
		RecipientMode:     RecipientModeFullValidatorSet,
		PayloadCiphertext: []byte("encrypted_payload_data"),
		PayloadNonce:      []byte("nonce12345678901234567890123456"),
		WrappedKeys: []WrappedKeyEntry{
			{
				RecipientID: "recipient1",
				WrappedKey:  []byte("wrapped_key_1"),
			},
		},
		ClientSignature: []byte("client_signature"),
		ClientID:        "approved_client_1",
		UserSignature:   []byte("user_signature"),
		UserPubKey:      []byte("user_pubkey_32bytes_exactly_now!"),
		Metadata:        make(map[string]string),
	}
}
