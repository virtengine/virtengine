// Package nitro provides AWS Nitro Enclave integration for VirtEngine TEE.
//
// AWS Nitro Enclaves are isolated compute environments that run on Amazon EC2
// instances, providing strong isolation for sensitive data processing. This
// package implements comprehensive Nitro Enclave support including enclave
// lifecycle management, attestation document handling, NSM device interaction,
// and attestation verification.
//
// # Components
//
// The package consists of four main components:
//
//   - NitroEnclave: Manages enclave lifecycle (build, run, describe, terminate)
//   - NSMDevice: Interacts with the Nitro Security Module for attestation
//   - AttestationDocument: Handles CBOR/COSE encoded attestation documents
//   - Verifier: Validates attestation documents and certificate chains
//
// # Simulation Mode
//
// When running on non-Nitro hardware, the package automatically falls back to
// simulation mode. Simulation mode is suitable for development and testing but
// does not provide real security guarantees. Use NewNitroEnclaveWithMode or
// NSMDevice.OpenWithMode to explicitly control hardware requirements.
//
// # Hardware Mode
//
// Real Nitro Enclave operations require:
//   - EC2 instance with Nitro Enclave support (c5.xlarge, m5.xlarge, etc.)
//   - nitro-cli installed and configured
//   - /dev/nitro_enclaves device available
//   - Enclave allocator configured in /etc/nitro_enclaves/allocator.yaml
//
// # PCR Measurements
//
// Nitro Enclaves use Platform Configuration Registers (PCRs) for attestation:
//   - PCR0: Enclave Image File (EIF) measurement
//   - PCR1: Linux kernel and boot ramfs measurement
//   - PCR2: User application measurement
//   - PCR3: IAM role attached to parent instance (if any)
//   - PCR4: Instance ID of the parent EC2 instance
//   - PCR8: Signing certificate for signed enclave images
//
// # Example Usage
//
//	// Create enclave manager
//	ne, err := nitro.NewNitroEnclave()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Build enclave from Docker image
//	if err := ne.BuildEnclave("myapp:latest", "/tmp/myapp.eif"); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Run enclave
//	info, err := ne.RunEnclave("/tmp/myapp.eif", 2, 2048)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer ne.TerminateEnclave(info.EnclaveID)
//
//	// Get attestation from inside enclave (using NSM)
//	nsm, _ := nitro.NewNSMSession()
//	defer nsm.Close()
//
//	attestation, err := nsm.GetAttestation(userData, nonce, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Verify attestation
//	verifier := nitro.NewVerifier()
//	result, err := verifier.VerifyRaw(attestation)
//	if err != nil || !result.Valid {
//	    log.Fatal("Verification failed")
//	}
//
// # Thread Safety
//
// All exported types in this package are safe for concurrent use.
//
// # Build Tags
//
// The package uses build tags to control behavior:
//   - Default: Simulation mode for development/testing
//   - nitro_hardware: Enables real hardware operations (not implemented)
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package nitro
