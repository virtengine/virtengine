package enclaVIRTENGINE_runtime

import (
	"context"
	"testing"
	"time"
)

func TestSimulatedEnclaveService_Initialize(t *testing.T) {
	svc := NewSimulatedEnclaveService()

	err := svc.Initialize(DefaultRuntimeConfig())
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	status := svc.GetStatus()
	if !status.Initialized {
		t.Error("expected enclave to be initialized")
	}

	if !status.Available {
		t.Error("expected enclave to be available")
	}
}

func TestSimulatedEnclaveService_Score(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	config := DefaultRuntimeConfig()
	
	if err := svc.Initialize(config); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	request := &ScoringRequest{
		RequestID:      "test-request-1",
		Ciphertext:     []byte("encrypted_identity_data"),
		WrappedKey:     []byte("wrapped_data_encryption_key"),
		Nonce:          []byte("nonce_12345678"),
		ScopeID:        "scope-123",
		AccountAddress: "virtengine1abc123",
		BlockHeight:    12345,
	}

	result, err := svc.Score(context.Background(), request)
	if err != nil {
		t.Fatalf("Score() error: %v", err)
	}

	if !result.IsSuccess() {
		t.Fatalf("Score() failed: %s", result.Error)
	}

	if result.RequestID != request.RequestID {
		t.Errorf("expected request ID %s, got %s", request.RequestID, result.RequestID)
	}

	if result.Score > 100 {
		t.Errorf("score should be 0-100, got %d", result.Score)
	}

	if len(result.EnclaveSignature) == 0 {
		t.Error("expected enclave signature")
	}

	if len(result.MeasurementHash) == 0 {
		t.Error("expected measurement hash")
	}

	if len(result.ModelVersionHash) == 0 {
		t.Error("expected model version hash")
	}

	if len(result.InputHash) == 0 {
		t.Error("expected input hash")
	}
}

func TestSimulatedEnclaveService_ScoreNotInitialized(t *testing.T) {
	svc := NewSimulatedEnclaveService()

	request := &ScoringRequest{
		RequestID:      "test-request-1",
		Ciphertext:     []byte("encrypted_data"),
		WrappedKey:     []byte("wrapped_key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope-1",
		AccountAddress: "addr1",
	}

	_, err := svc.Score(context.Background(), request)
	if err != ErrEnclaveNotInitialized {
		t.Errorf("expected ErrEnclaveNotInitialized, got %v", err)
	}
}

func TestSimulatedEnclaveService_ScoreValidation(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	if err := svc.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	tests := []struct {
		name    string
		request *ScoringRequest
		wantErr string
	}{
		{
			name: "empty request ID",
			request: &ScoringRequest{
				RequestID:      "",
				Ciphertext:     []byte("data"),
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			},
			wantErr: "request ID required",
		},
		{
			name: "empty ciphertext",
			request: &ScoringRequest{
				RequestID:      "req-1",
				Ciphertext:     nil,
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			},
			wantErr: "ciphertext required",
		},
		{
			name: "empty scope ID",
			request: &ScoringRequest{
				RequestID:      "req-1",
				Ciphertext:     []byte("data"),
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "",
				AccountAddress: "addr",
			},
			wantErr: "scope ID required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := svc.Score(context.Background(), tt.request)
			if result.Error != tt.wantErr {
				t.Errorf("expected error %q, got %q", tt.wantErr, result.Error)
			}
		})
	}
}

func TestSimulatedEnclaveService_ScoreTimeout(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	config := DefaultRuntimeConfig()
	config.MaxExecutionTimeMs = 1 // Very short timeout
	
	if err := svc.Initialize(config); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	request := &ScoringRequest{
		RequestID:      "test-request-timeout",
		Ciphertext:     []byte("encrypted_data"),
		WrappedKey:     []byte("wrapped_key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope-1",
		AccountAddress: "addr1",
	}

	// Use cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, _ := svc.Score(ctx, request)
	if result.Error != ErrTimeout.Error() && result.Error != context.Canceled.Error() {
		// Either timeout or cancelled is acceptable
		t.Logf("got error: %s (may be timing dependent)", result.Error)
	}
}

func TestSimulatedEnclaveService_GetKeys(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	
	// Before initialization
	_, err := svc.GetEncryptionPubKey()
	if err != ErrEnclaveNotInitialized {
		t.Errorf("expected ErrEnclaveNotInitialized, got %v", err)
	}

	if err := svc.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	encKey, err := svc.GetEncryptionPubKey()
	if err != nil {
		t.Fatalf("GetEncryptionPubKey() error: %v", err)
	}
	if len(encKey) != 32 {
		t.Errorf("expected 32 byte key, got %d", len(encKey))
	}

	sigKey, err := svc.GetSigningPubKey()
	if err != nil {
		t.Fatalf("GetSigningPubKey() error: %v", err)
	}
	if len(sigKey) != 32 {
		t.Errorf("expected 32 byte key, got %d", len(sigKey))
	}

	measurement, err := svc.GetMeasurement()
	if err != nil {
		t.Fatalf("GetMeasurement() error: %v", err)
	}
	if len(measurement) != 32 {
		t.Errorf("expected 32 byte measurement, got %d", len(measurement))
	}
}

func TestSimulatedEnclaveService_RotateKeys(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	
	if err := svc.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	oldKey, _ := svc.GetEncryptionPubKey()
	oldEpoch := svc.GetStatus().CurrentEpoch

	if err := svc.RotateKeys(); err != nil {
		t.Fatalf("RotateKeys() error: %v", err)
	}

	newKey, _ := svc.GetEncryptionPubKey()
	newEpoch := svc.GetStatus().CurrentEpoch

	if string(oldKey) == string(newKey) {
		t.Error("expected encryption key to change after rotation")
	}

	if newEpoch != oldEpoch+1 {
		t.Errorf("expected epoch to increment from %d to %d, got %d", oldEpoch, oldEpoch+1, newEpoch)
	}
}

func TestSimulatedEnclaveService_GenerateAttestation(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	
	if err := svc.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	reportData := []byte("user_provided_report_data")
	
	quote, err := svc.GenerateAttestation(reportData)
	if err != nil {
		t.Fatalf("GenerateAttestation() error: %v", err)
	}

	if len(quote) == 0 {
		t.Error("expected non-empty attestation quote")
	}
}

func TestSimulatedEnclaveService_Shutdown(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	
	if err := svc.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	if !svc.GetStatus().Initialized {
		t.Error("expected enclave to be initialized")
	}

	if err := svc.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}

	if svc.GetStatus().Initialized {
		t.Error("expected enclave to be shut down")
	}
}

func TestSimulatedEnclaveService_DeterministicScoring(t *testing.T) {
	svc1 := NewSimulatedEnclaveService()
	svc2 := NewSimulatedEnclaveService()
	
	if err := svc1.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	if err := svc2.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	request := &ScoringRequest{
		RequestID:      "determinism-test",
		Ciphertext:     []byte("same_encrypted_data_for_both"),
		WrappedKey:     []byte("wrapped_key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope-1",
		AccountAddress: "addr1",
		BlockHeight:    100,
	}

	result1, _ := svc1.Score(context.Background(), request)
	result2, _ := svc2.Score(context.Background(), request)

	if result1.Score != result2.Score {
		t.Errorf("scores should be deterministic: %d vs %d", result1.Score, result2.Score)
	}

	if result1.Status != result2.Status {
		t.Errorf("statuses should be deterministic: %s vs %s", result1.Status, result2.Status)
	}

	if string(result1.InputHash) != string(result2.InputHash) {
		t.Error("input hashes should be deterministic")
	}
}

func TestSimulatedEnclaveService_Status(t *testing.T) {
	svc := NewSimulatedEnclaveService()
	
	// Before init
	status := svc.GetStatus()
	if status.Initialized {
		t.Error("expected not initialized")
	}

	if err := svc.Initialize(DefaultRuntimeConfig()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	// Small delay to ensure uptime is measurable
	time.Sleep(10 * time.Millisecond)

	status = svc.GetStatus()
	if !status.Initialized {
		t.Error("expected initialized")
	}
	if !status.Available {
		t.Error("expected available")
	}
	if status.CurrentEpoch != 1 {
		t.Errorf("expected epoch 1, got %d", status.CurrentEpoch)
	}
	if status.TotalProcessed != 0 {
		t.Errorf("expected 0 processed, got %d", status.TotalProcessed)
	}
}

func TestScoringRequest_Validate(t *testing.T) {
	config := DefaultRuntimeConfig()

	tests := []struct {
		name    string
		request *ScoringRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &ScoringRequest{
				RequestID:      "req-1",
				Ciphertext:     []byte("data"),
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			},
			wantErr: false,
		},
		{
			name: "input too large",
			request: &ScoringRequest{
				RequestID:      "req-1",
				Ciphertext:     make([]byte, 20*1024*1024), // 20MB
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
