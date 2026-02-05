package enclave

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	enclave_runtime "github.com/virtengine/virtengine/pkg/enclave_runtime"
)

// MockService is a lightweight mock of the EnclaveService interface for tests.
// It avoids hardware dependencies while keeping deterministic behavior.
type MockService struct {
	Measurement    []byte
	EncryptionKey  []byte
	SigningKey     []byte
	Attestation    []byte
	Status         enclave_runtime.EnclaveStatus
	ScoreResult    *enclave_runtime.ScoringResult
	ScoreErr       error
	Initialized    bool
	RotateCalled   bool
	ShutdownCalled bool
}

// NewMockService returns a mock with deterministic defaults.
func NewMockService() *MockService {
	measurement := sha256.Sum256([]byte("mock-enclave-measurement"))
	encKey := sha256.Sum256([]byte("mock-encryption-key"))
	signKey := sha256.Sum256([]byte("mock-signing-key"))
	att := []byte("mock-attestation")

	return &MockService{
		Measurement:   measurement[:],
		EncryptionKey: encKey[:],
		SigningKey:    signKey[:],
		Attestation:   att,
		Status: enclave_runtime.EnclaveStatus{
			Initialized: false,
			Available:   false,
		},
	}
}

func (m *MockService) Initialize(config enclave_runtime.RuntimeConfig) error {
	m.Initialized = true
	m.Status.Initialized = true
	m.Status.Available = true

	if len(m.EncryptionKey) == 0 {
		m.EncryptionKey = make([]byte, 32)
		_, _ = rand.Read(m.EncryptionKey)
	}

	if len(m.SigningKey) == 0 {
		m.SigningKey = make([]byte, 32)
		_, _ = rand.Read(m.SigningKey)
	}

	if len(m.Measurement) == 0 {
		sum := sha256.Sum256([]byte(config.ModelPath))
		m.Measurement = sum[:]
	}

	return nil
}

func (m *MockService) Score(_ context.Context, request *enclave_runtime.ScoringRequest) (*enclave_runtime.ScoringResult, error) {
	if m.ScoreErr != nil {
		return nil, m.ScoreErr
	}
	if !m.Status.Available {
		return nil, enclave_runtime.ErrEnclaveUnavailable
	}
	if request == nil {
		return nil, errors.New("nil scoring request")
	}

	if m.ScoreResult != nil {
		return m.ScoreResult, nil
	}

	return &enclave_runtime.ScoringResult{
		Score:            75,
		Status:           "mock",
		EnclaveSignature: []byte("mock-signature"),
		MeasurementHash:  m.Measurement,
	}, nil
}

func (m *MockService) GetMeasurement() ([]byte, error) {
	return m.Measurement, nil
}

func (m *MockService) GetEncryptionPubKey() ([]byte, error) {
	return m.EncryptionKey, nil
}

func (m *MockService) GetSigningPubKey() ([]byte, error) {
	return m.SigningKey, nil
}

func (m *MockService) GenerateAttestation(reportData []byte) ([]byte, error) {
	if len(reportData) == 0 {
		return m.Attestation, nil
	}

	att := append([]byte("mock-attestation-"), reportData...)
	return att, nil
}

func (m *MockService) RotateKeys() error {
	m.RotateCalled = true
	m.EncryptionKey = nil
	m.SigningKey = nil
	return nil
}

func (m *MockService) GetStatus() enclave_runtime.EnclaveStatus {
	return m.Status
}

func (m *MockService) Shutdown() error {
	m.ShutdownCalled = true
	m.Status.Available = false
	return nil
}
