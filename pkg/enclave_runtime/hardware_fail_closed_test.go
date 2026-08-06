package enclave_runtime

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSGXHardwareSelectedRuntimeFailsClosed(t *testing.T) {
	t.Run("service initialization rejects unsupported hardware load", func(t *testing.T) {
		enclavePath := filepath.Join(t.TempDir(), "test.signed.so")
		if err := os.WriteFile(enclavePath, []byte("dummy enclave"), 0o600); err != nil {
			t.Fatalf("failed to create test enclave file: %v", err)
		}

		detector := &SGXHardwareDetector{available: true}
		loader := NewSGXEnclaveLoader(detector)
		loader.simulated = false

		backend := &SGXHardwareBackend{
			detector:    detector,
			loader:      loader,
			initialized: true,
			enclavePath: enclavePath,
			sealer:      NewSGXSealingService(loader),
			quoter:      NewSGXQuoteGenerator(detector, loader),
			ecaller:     NewSGXECallInterface(loader),
		}

		svc := &SGXEnclaveServiceImpl{
			config:          SGXEnclaveConfig{EnclavePath: enclavePath},
			hardwareBackend: backend,
		}

		err := svc.Initialize(DefaultRuntimeConfig())
		if !errors.Is(err, ErrHardwareOperationUnsupported) && !errors.Is(err, ErrHardwareNotAvailable) {
			t.Fatalf("expected unsupported or unavailable hardware error, got %v", err)
		}
	})

	t.Run("low-level SGX hardware helpers reject synthetic execution", func(t *testing.T) {
		loader := &SGXEnclaveLoader{
			detector:  &SGXHardwareDetector{available: true},
			loaded:    true,
			simulated: false,
		}

		var reportData [64]byte
		copy(reportData[:], []byte("sgx-fail-closed"))

		reportGen := NewSGXReportGenerator(loader)
		if _, err := reportGen.generateHardwareReport(reportData, nil); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SGX report error, got %v", err)
		}

		quoter := NewSGXQuoteGenerator(loader.detector, loader)
		if _, err := quoter.generateHardwareQuote(reportData); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SGX quote error, got %v", err)
		}

		sealer := NewSGXSealingService(loader)
		if _, err := sealer.sealHardware([]byte("secret")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SGX seal error, got %v", err)
		}
		if _, err := sealer.unsealHardware([]byte("sealed")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SGX unseal error, got %v", err)
		}

		ecaller := NewSGXECallInterface(loader)
		if _, err := ecaller.callHardware(1, []byte("payload")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SGX ecall error, got %v", err)
		}
	})
}

func TestSEVHardwareSelectedRuntimeFailsClosed(t *testing.T) {
	device := &SEVGuestDevice{
		detector:   &SEVHardwareDetector{available: true},
		devicePath: SEVGuestDevicePath,
		opened:     true,
		simulated:  false,
	}
	backend := &SEVHardwareBackend{
		detector:     device.detector,
		device:       device,
		reportReq:    NewSNPReportRequester(device),
		keyReq:       NewSNPDerivedKeyRequester(device),
		extReportReq: NewSNPExtendedReportRequester(device),
		initialized:  true,
	}

	t.Run("service initialization rejects unsupported SNP ioctls", func(t *testing.T) {
		svc := &SEVSNPEnclaveServiceImpl{
			config:          SEVSNPConfig{Endpoint: "localhost:8443"},
			hardwareBackend: backend,
			currentEpoch:    1,
		}

		err := svc.Initialize(DefaultRuntimeConfig())
		if !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SEV initialization error, got %v", err)
		}
	})

	t.Run("low-level SEV hardware requesters reject synthetic execution", func(t *testing.T) {
		var userData [64]byte
		copy(userData[:], []byte("sev-fail-closed"))

		if _, err := backend.reportReq.requestHardwareReport(userData, 0); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SNP report error, got %v", err)
		}
		if _, err := backend.keyReq.requestHardwareKey(SNP_KEY_ROOT_VCEK, SNP_KEY_GUEST_FIELD, 0); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SNP derived key error, got %v", err)
		}
		if _, err := backend.extReportReq.requestHardwareExtendedReport(userData, 0); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SNP extended report error, got %v", err)
		}
	})

	t.Run("service extended report does not fall back to simulation", func(t *testing.T) {
		svc := &SEVSNPEnclaveServiceImpl{
			config:          SEVSNPConfig{Endpoint: "localhost:8443"},
			hardwareBackend: backend,
			initialized:     true,
		}

		if _, _, err := svc.GenerateExtendedReport([]byte("report")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported SEV extended report error, got %v", err)
		}
	})
}

func TestNitroHardwareSelectedRuntimeFailsClosed(t *testing.T) {
	t.Run("service initialization rejects incomplete Nitro runtime", func(t *testing.T) {
		svc := &NitroEnclaveServiceImpl{
			config: NitroEnclaveConfig{
				EnclaveImagePath: "/tmp/test.eif",
				CPUCount:         2,
				MemoryMB:         512,
			},
			hardwareBackend: &NitroHardwareBackend{initialized: true},
		}

		err := svc.Initialize(DefaultRuntimeConfig())
		if !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported Nitro initialization error, got %v", err)
		}
	})

	t.Run("low-level Nitro NSM helpers reject synthetic execution", func(t *testing.T) {
		client := &NitroNSMClient{
			devicePath: NitroNSMDevPath,
			opened:     true,
			simulated:  false,
		}

		if _, err := client.getHardwareAttestationDocument([]byte("user"), []byte("nonce"), []byte("pub")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported Nitro attestation error, got %v", err)
		}
		if _, err := client.DescribePCRs(); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported Nitro describe PCRs error, got %v", err)
		}
		if err := client.ExtendPCR(1, []byte("extend")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported Nitro extend PCR error, got %v", err)
		}
		if err := client.LockPCR(1); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported Nitro lock PCR error, got %v", err)
		}
	})

	t.Run("service attestation does not fall back to simulation", func(t *testing.T) {
		client := &NitroNSMClient{
			devicePath: NitroNSMDevPath,
			opened:     true,
			simulated:  false,
		}
		backend := &NitroHardwareBackend{
			nsmClient:   client,
			initialized: true,
		}
		svc := &NitroEnclaveServiceImpl{
			config: NitroEnclaveConfig{
				EnclaveImagePath: "/tmp/test.eif",
				CPUCount:         2,
				MemoryMB:         512,
			},
			hardwareBackend: backend,
			initialized:     true,
		}

		if _, err := svc.GenerateAttestation([]byte("nonce")); !errors.Is(err, ErrHardwareOperationUnsupported) {
			t.Fatalf("expected unsupported Nitro attestation error, got %v", err)
		}
	})
}
