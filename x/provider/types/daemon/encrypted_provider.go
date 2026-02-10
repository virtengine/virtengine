package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// EncryptedProviderPayload wraps encrypted provider attribute data
// Sensitive provider information is encrypted and only accessible to:
// - The provider owner
// - Validators (for verification)
// - Authorized auditors
type EncryptedProviderPayload struct {
	// Envelope contains the encrypted provider payload
	Envelope *encryptiontypes.EncryptedPayloadEnvelope `json:"envelope,omitempty"`

	// EnvelopeRef optionally points to an off-chain payload location
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash,omitempty"`

	// PayloadSize is the encrypted payload size in bytes
	PayloadSize uint32 `json:"payload_size,omitempty"`

	// ProviderKeyID is the provider's key fingerprint that can decrypt
	ProviderKeyID string `json:"provider_key_id,omitempty"`
}

// Validate validates the encrypted provider payload
func (p *EncryptedProviderPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.Envelope == nil {
		return fmt.Errorf("payload envelope is required")
	}
	if err := p.Envelope.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}
	if len(p.EnvelopeHash) > 0 && len(p.EnvelopeHash) != 32 {
		return fmt.Errorf("invalid envelope_hash length: %d", len(p.EnvelopeHash))
	}

	// Validate key ID is in recipients if provided
	if p.ProviderKeyID != "" && !p.Envelope.IsRecipient(p.ProviderKeyID) {
		return fmt.Errorf("provider key id not present in envelope recipients")
	}

	return nil
}

// EnsureEnvelopeHash sets the envelope hash if missing
func (p *EncryptedProviderPayload) EnsureEnvelopeHash() {
	if p == nil || p.Envelope == nil {
		return
	}
	if len(p.EnvelopeHash) == 0 {
		p.EnvelopeHash = p.Envelope.Hash()
	}
	if p.PayloadSize == 0 {
		p.PayloadSize = safeUint32FromInt(len(p.Envelope.Ciphertext))
	}
}

// EnvelopeHashHex returns the envelope hash as hex string
func (p *EncryptedProviderPayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload
func (p *EncryptedProviderPayload) CloneWithoutEnvelope() *EncryptedProviderPayload {
	if p == nil {
		return nil
	}
	return &EncryptedProviderPayload{
		Envelope:      nil,
		EnvelopeRef:   p.EnvelopeRef,
		EnvelopeHash:  append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:   p.PayloadSize,
		ProviderKeyID: p.ProviderKeyID,
	}
}

// HasEnvelope returns true if the envelope is present
func (p *EncryptedProviderPayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging
func (p *EncryptedProviderPayload) String() string {
	if p == nil {
		return "EncryptedProviderPayload<nil>"
	}
	return fmt.Sprintf("EncryptedProviderPayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
}

// ProviderPayload is the decrypted provider payload structure
// This is what gets encrypted into the envelope
type ProviderPayload struct {
	// Infrastructure details
	DatacenterLocation string            `json:"datacenter_location,omitempty"`
	HardwareSpecs      map[string]string `json:"hardware_specs,omitempty"`
	NetworkCapacity    uint64            `json:"network_capacity,omitempty"`

	// Connectivity
	APIEndpoints []string `json:"api_endpoints,omitempty"`
	RPCEndpoints []string `json:"rpc_endpoints,omitempty"`

	// Credentials and secrets
	AdminCredentials string            `json:"admin_credentials,omitempty"` // Encrypted credentials
	APIKeys          map[string]string `json:"api_keys,omitempty"`

	// Compliance and certifications
	Certifications []string          `json:"certifications,omitempty"`
	ComplianceInfo map[string]string `json:"compliance_info,omitempty"`

	// Custom attributes
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Validate validates the provider payload
func (p *ProviderPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	// Add specific validation rules as needed
	return nil
}

// EncryptedConfigPayload wraps encrypted daemon configuration data
// Sensitive daemon configuration is encrypted and only accessible to:
// - The provider daemon
// - The provider owner
type EncryptedConfigPayload struct {
	// Envelope contains the encrypted config payload
	Envelope *encryptiontypes.EncryptedPayloadEnvelope `json:"envelope,omitempty"`

	// EnvelopeRef optionally points to an off-chain payload location
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash,omitempty"`

	// PayloadSize is the encrypted payload size in bytes
	PayloadSize uint32 `json:"payload_size,omitempty"`

	// DaemonKeyID is the daemon's key fingerprint that can decrypt
	DaemonKeyID string `json:"daemon_key_id,omitempty"`

	// OwnerKeyID is the owner's key fingerprint that can decrypt
	OwnerKeyID string `json:"owner_key_id,omitempty"`
}

// Validate validates the encrypted config payload
func (p *EncryptedConfigPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.Envelope == nil {
		return fmt.Errorf("payload envelope is required")
	}
	if err := p.Envelope.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}
	if len(p.EnvelopeHash) > 0 && len(p.EnvelopeHash) != 32 {
		return fmt.Errorf("invalid envelope_hash length: %d", len(p.EnvelopeHash))
	}

	// Validate key IDs are in recipients if provided
	if p.DaemonKeyID != "" && !p.Envelope.IsRecipient(p.DaemonKeyID) {
		return fmt.Errorf("daemon key id not present in envelope recipients")
	}
	if p.OwnerKeyID != "" && !p.Envelope.IsRecipient(p.OwnerKeyID) {
		return fmt.Errorf("owner key id not present in envelope recipients")
	}

	return nil
}

// EnsureEnvelopeHash sets the envelope hash if missing
func (p *EncryptedConfigPayload) EnsureEnvelopeHash() {
	if p == nil || p.Envelope == nil {
		return
	}
	if len(p.EnvelopeHash) == 0 {
		p.EnvelopeHash = p.Envelope.Hash()
	}
	if p.PayloadSize == 0 {
		p.PayloadSize = safeUint32FromInt(len(p.Envelope.Ciphertext))
	}
}

// EnvelopeHashHex returns the envelope hash as hex string
func (p *EncryptedConfigPayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload
func (p *EncryptedConfigPayload) CloneWithoutEnvelope() *EncryptedConfigPayload {
	if p == nil {
		return nil
	}
	return &EncryptedConfigPayload{
		Envelope:     nil,
		EnvelopeRef:  p.EnvelopeRef,
		EnvelopeHash: append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:  p.PayloadSize,
		DaemonKeyID:  p.DaemonKeyID,
		OwnerKeyID:   p.OwnerKeyID,
	}
}

// HasEnvelope returns true if the envelope is present
func (p *EncryptedConfigPayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging
func (p *EncryptedConfigPayload) String() string {
	if p == nil {
		return "EncryptedConfigPayload<nil>"
	}
	return fmt.Sprintf("EncryptedConfigPayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
}

// DaemonConfigPayload is the decrypted daemon configuration payload
// This is what gets encrypted into the envelope
type DaemonConfigPayload struct {
	// Kubernetes configuration
	KubernetesConfig *K8sConfigData `json:"kubernetes_config,omitempty"`

	// Waldur integration
	WaldurConfig *WaldurConfigData `json:"waldur_config,omitempty"`

	// Ansible configuration
	AnsibleConfig *AnsibleConfigData `json:"ansible_config,omitempty"`

	// Cloud provider configs
	CloudConfigs map[string]interface{} `json:"cloud_configs,omitempty"`

	// Secrets
	Secrets map[string]string `json:"secrets,omitempty"`
}

// K8sConfigData holds Kubernetes-specific configuration
type K8sConfigData struct {
	Kubeconfig   string   `json:"kubeconfig,omitempty"`
	Namespaces   []string `json:"namespaces,omitempty"`
	IngressClass string   `json:"ingress_class,omitempty"`
	StorageClass string   `json:"storage_class,omitempty"`
}

// WaldurConfigData holds Waldur integration configuration
type WaldurConfigData struct {
	APIURL      string `json:"api_url,omitempty"`
	APIToken    string `json:"api_token,omitempty"`
	ProjectUUID string `json:"project_uuid,omitempty"`
}

// AnsibleConfigData holds Ansible playbook configuration
type AnsibleConfigData struct {
	PlaybookPath string            `json:"playbook_path,omitempty"`
	Inventory    string            `json:"inventory,omitempty"`
	VaultPass    string            `json:"vault_pass,omitempty"`
	ExtraVars    map[string]string `json:"extra_vars,omitempty"`
}

// Validate validates the daemon config payload
func (p *DaemonConfigPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	// Add specific validation rules as needed
	return nil
}

// VerifyProviderEnvelopeHash verifies that the envelope hash matches the envelope content
func VerifyProviderEnvelopeHash(envelope *encryptiontypes.EncryptedPayloadEnvelope, expectedHash []byte) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}
	if len(expectedHash) != 32 {
		return fmt.Errorf("invalid hash length: %d", len(expectedHash))
	}

	actualHash := envelope.Hash()
	if len(actualHash) != 32 {
		return fmt.Errorf("computed hash has invalid length: %d", len(actualHash))
	}

	// Constant-time comparison
	match := true
	for i := 0; i < 32; i++ {
		if actualHash[i] != expectedHash[i] {
			match = false
		}
	}

	if !match {
		return fmt.Errorf("envelope hash mismatch: expected %s, got %s",
			hex.EncodeToString(expectedHash), hex.EncodeToString(actualHash))
	}

	return nil
}

// ComputeProviderPayloadHash computes the SHA-256 hash of a payload
func ComputeProviderPayloadHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
}
