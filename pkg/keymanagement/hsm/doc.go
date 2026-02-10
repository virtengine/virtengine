// Package hsm provides Hardware Security Module (HSM) integration for
// VirtEngine key management. It supports PKCS#11 backends, cloud HSM services
// (AWS CloudHSM, GCP Cloud HSM, Azure Dedicated HSM), and Ledger hardware
// wallets.
//
// The package is organised around the [HSMProvider] interface which all
// backends must implement. A [Manager] coordinates provider lifecycle,
// health-checks and audit logging.
//
// # Security Properties
//
//   - Private keys are created as non-extractable inside the HSM.
//   - PINs/passwords are never logged and are cleared from memory after use.
//   - All signing operations are audited.
//   - Sessions are cleaned up on error or timeout.
package hsm
