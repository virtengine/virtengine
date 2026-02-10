package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	TicketID            string                    `json:"ticket_id"`
	Version             uint32                    `json:"version"`
	Category            string                    `json:"category"`
	Priority            string                    `json:"priority"`
	Tags                []string                  `json:"tags,omitempty"`
	Status              string                    `json:"status"`
	CreatedAt           int64                     `json:"created_at"`
	UpdatedAt           int64                     `json:"updated_at"`
	ClosedAt            int64                     `json:"closed_at,omitempty"`
	CustomerAddress     string                    `json:"customer_address"`
	AssignedAgentAddr   string                    `json:"assigned_agent_addr,omitempty"`
	OrderRef            *OrderReference           `json:"order_ref,omitempty"`
	ProviderRef         *ProviderReference        `json:"provider_ref,omitempty"`
	DeploymentRef       *DeploymentReference      `json:"deployment_ref,omitempty"`
	RelatedTickets      []string                  `json:"related_tickets,omitempty"`
	Subject             string                    `json:"subject"`
	DescriptionEnvelope *EncryptedPayloadEnvelope `json:"description_envelope"`
	Messages            []TicketMessage           `json:"messages,omitempty"`
	Attachments         []AttachmentRef           `json:"attachments,omitempty"`
	ContactEnvelope     *EncryptedPayloadEnvelope `json:"contact_envelope,omitempty"`
	RetentionPolicy     RetentionPolicyRef        `json:"retention_policy"`
	JurisdictionCode    string                    `json:"jurisdiction_code"`
	ConsentRecordID     string                    `json:"consent_record_id,omitempty"`
	Metadata            map[string]string         `json:"metadata,omitempty"`
}

type OrderReference struct {
	OrderID   string `json:"order_id"`
	OwnerAddr string `json:"owner_addr"`
	DSeq      uint64 `json:"dseq"`
	GSeq      uint32 `json:"gseq"`
	OSeq      uint32 `json:"oseq"`
}

type ProviderReference struct {
	ProviderAddr string `json:"provider_addr"`
	ProviderName string `json:"provider_name,omitempty"`
	LeaseID      string `json:"lease_id,omitempty"`
}

type DeploymentReference struct {
	DeploymentID string `json:"deployment_id"`
	OwnerAddr    string `json:"owner_addr"`
	DSeq         uint64 `json:"dseq"`
}

type EncryptedPayloadEnvelope struct {
	Version             uint32            `json:"version"`
	AlgorithmID         string            `json:"algorithm_id"`
	AlgorithmVersion    uint32            `json:"algorithm_version"`
	RecipientKeyIDs     []string          `json:"recipient_key_ids"`
	RecipientPublicKeys []string          `json:"recipient_public_keys,omitempty"`
	EncryptedKeys       []string          `json:"encrypted_keys,omitempty"`
	WrappedKeys         []WrappedKeyEntry `json:"wrapped_keys,omitempty"`
	Nonce               string            `json:"nonce"`
	Ciphertext          string            `json:"ciphertext"`
	SenderSignature     string            `json:"sender_signature"`
	SenderPubKey        string            `json:"sender_pub_key"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

type WrappedKeyEntry struct {
	RecipientID     string `json:"recipient_id"`
	WrappedKey      string `json:"wrapped_key"`
	Algorithm       string `json:"algorithm,omitempty"`
	EphemeralPubKey string `json:"ephemeral_pub_key,omitempty"`
}

type TicketMessage struct {
	MessageID       string                    `json:"message_id"`
	SenderAddr      string                    `json:"sender_addr"`
	SenderRole      string                    `json:"sender_role"`
	CreatedAt       int64                     `json:"created_at"`
	ContentEnvelope *EncryptedPayloadEnvelope `json:"content_envelope"`
	IsInternal      bool                      `json:"is_internal"`
}

type AttachmentRef struct {
	AttachmentID    string                    `json:"attachment_id"`
	FileName        string                    `json:"file_name"`
	ContentType     string                    `json:"content_type"`
	SizeBytes       int64                     `json:"size_bytes"`
	StorageLocation string                    `json:"storage_location"`
	UploadedAt      int64                     `json:"uploaded_at"`
	UploadedBy      string                    `json:"uploaded_by"`
	ChecksumSHA256  string                    `json:"checksum_sha256"`
	KeyEnvelope     *EncryptedPayloadEnvelope `json:"key_envelope"`
}

type RetentionPolicyRef struct {
	PolicyID          string `json:"policy_id"`
	RetentionDays     int    `json:"retention_days"`
	DeleteAfterClosed bool   `json:"delete_after_closed"`
	ArchiveAfterDays  int    `json:"archive_after_days,omitempty"`
}

// Valid values
var validStatuses = []string{
	"OPEN", "ASSIGNED", "IN_PROGRESS", "AWAITING_CUSTOMER",
	"AWAITING_PROVIDER", "ESCALATED", "ON_HOLD", "RESOLVED",
	"CLOSED", "CANCELLED",
}

var validCategories = []string{
	"billing", "technical", "identity", "marketplace", "security", "other",
}

var validPriorities = []string{
	"low", "medium", "high", "critical",
}

var validSenderRoles = []string{
	"customer", "support_agent", "moderator", "administrator", "service_provider", "system",
}

var validPolicyIDs = []string{
	"gdpr", "ccpa", "us_general", "uk_gdpr", "global", "legal_hold",
}

var validAlgorithms = []string{
	"X25519-XSALSA20-POLY1305", "AGE-X25519",
}

var validContentTypes = []string{
	"image/png", "image/jpeg", "image/gif", "image/webp",
	"application/pdf", "text/plain", "text/csv", "application/json",
	"application/zip", "application/gzip", "video/mp4", "video/webm",
}

// State transition rules
var allowedTransitions = map[string][]string{
	"OPEN":              {"ASSIGNED", "CANCELLED"},
	"ASSIGNED":          {"IN_PROGRESS"},
	"IN_PROGRESS":       {"AWAITING_CUSTOMER", "AWAITING_PROVIDER", "ESCALATED", "ON_HOLD", "RESOLVED"},
	"AWAITING_CUSTOMER": {"IN_PROGRESS"},
	"AWAITING_PROVIDER": {"IN_PROGRESS"},
	"ESCALATED":         {"IN_PROGRESS", "RESOLVED"},
	"ON_HOLD":           {"IN_PROGRESS"},
	"RESOLVED":          {"CLOSED", "IN_PROGRESS"},
	"CLOSED":            {},
	"CANCELLED":         {},
}

// Regex patterns
var uuidV7Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var virtengineAddrPattern = regexp.MustCompile(`^virtengine1[a-z0-9]{32,58}$`)

//nolint:unused // Reserved for SHA256 validation in tests
var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var jurisdictionPattern = regexp.MustCompile(`^[A-Z]{2}(-[A-Z0-9]{1,3})?$`)

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateTicket validates a support ticket against the schema
func ValidateTicket(t *SupportTicket) []string {
	var errors []string

	// Required field validations
	if t.TicketID == "" {
		errors = append(errors, "ticket_id is required")
	} else if !uuidV7Pattern.MatchString(t.TicketID) {
		errors = append(errors, "ticket_id must be a valid UUID v7")
	}

	if t.Version != 1 {
		errors = append(errors, "version must be 1")
	}

	if !contains(validCategories, t.Category) {
		errors = append(errors, "invalid category: "+t.Category)
	}

	if !contains(validPriorities, t.Priority) {
		errors = append(errors, "invalid priority: "+t.Priority)
	}

	if !contains(validStatuses, t.Status) {
		errors = append(errors, "invalid status: "+t.Status)
	}

	if t.CreatedAt <= 0 {
		errors = append(errors, "created_at must be a positive Unix timestamp")
	}

	if t.UpdatedAt <= 0 {
		errors = append(errors, "updated_at must be a positive Unix timestamp")
	}

	if t.UpdatedAt < t.CreatedAt {
		errors = append(errors, "updated_at must be >= created_at")
	}

	// Terminal states require closed_at
	if t.Status == "CLOSED" || t.Status == "CANCELLED" {
		if t.ClosedAt <= 0 {
			errors = append(errors, "closed_at is required for terminal states")
		}
	}

	if t.CustomerAddress == "" {
		errors = append(errors, "customer_address is required")
	} else if !virtengineAddrPattern.MatchString(t.CustomerAddress) {
		errors = append(errors, "customer_address must be a valid VirtEngine address")
	}

	if t.AssignedAgentAddr != "" && !virtengineAddrPattern.MatchString(t.AssignedAgentAddr) {
		errors = append(errors, "assigned_agent_addr must be a valid VirtEngine address")
	}

	if t.Subject == "" {
		errors = append(errors, "subject is required")
	} else if len(t.Subject) < 5 {
		errors = append(errors, "subject must be at least 5 characters")
	} else if len(t.Subject) > 200 {
		errors = append(errors, "subject must not exceed 200 characters")
	}

	if t.DescriptionEnvelope == nil {
		errors = append(errors, "description_envelope is required")
	} else {
		envErrors := validateEnvelope(t.DescriptionEnvelope, "description_envelope")
		errors = append(errors, envErrors...)
	}

	// Validate retention policy
	if t.RetentionPolicy.PolicyID == "" {
		errors = append(errors, "retention_policy.policy_id is required")
	} else if !contains(validPolicyIDs, t.RetentionPolicy.PolicyID) {
		errors = append(errors, "invalid retention_policy.policy_id: "+t.RetentionPolicy.PolicyID)
	}

	if t.RetentionPolicy.RetentionDays < 0 || t.RetentionPolicy.RetentionDays > 3650 {
		errors = append(errors, "retention_policy.retention_days must be between 0 and 3650")
	}

	// Validate jurisdiction code
	if t.JurisdictionCode == "" {
		errors = append(errors, "jurisdiction_code is required")
	} else if !jurisdictionPattern.MatchString(t.JurisdictionCode) {
		errors = append(errors, "jurisdiction_code must be a valid ISO 3166-1 alpha-2 code")
	}

	// Validate messages
	for i, msg := range t.Messages {
		msgErrors := validateMessage(&msg, i)
		errors = append(errors, msgErrors...)
	}

	// Validate attachments
	for i, att := range t.Attachments {
		attErrors := validateAttachment(&att, i)
		errors = append(errors, attErrors...)
	}

	// Validate contact envelope if present
	if t.ContactEnvelope != nil {
		envErrors := validateEnvelope(t.ContactEnvelope, "contact_envelope")
		errors = append(errors, envErrors...)
	}

	// Validate tags
	if len(t.Tags) > 10 {
		errors = append(errors, "tags must not exceed 10 items")
	}
	for i, tag := range t.Tags {
		if len(tag) == 0 || len(tag) > 50 {
			errors = append(errors, "tags["+string(rune('0'+i))+"] must be 1-50 characters")
		}
	}

	// Validate related tickets
	if len(t.RelatedTickets) > 20 {
		errors = append(errors, "related_tickets must not exceed 20 items")
	}
	for i, rt := range t.RelatedTickets {
		if !uuidV7Pattern.MatchString(rt) {
			errors = append(errors, "related_tickets["+string(rune('0'+i))+"] must be a valid UUID v7")
		}
	}

	return errors
}

func validateEnvelope(e *EncryptedPayloadEnvelope, prefix string) []string {
	var errors []string

	if e.Version != 1 {
		errors = append(errors, prefix+".version must be 1")
	}

	if !contains(validAlgorithms, e.AlgorithmID) {
		errors = append(errors, prefix+".algorithm_id must be a supported algorithm")
	}

	if e.AlgorithmVersion < 1 {
		errors = append(errors, prefix+".algorithm_version must be >= 1")
	}

	if len(e.RecipientKeyIDs) == 0 {
		errors = append(errors, prefix+".recipient_key_ids must have at least one entry")
	}

	if len(e.RecipientKeyIDs) > 10 {
		errors = append(errors, prefix+".recipient_key_ids must not exceed 10 entries")
	}

	if e.Nonce == "" {
		errors = append(errors, prefix+".nonce is required")
	}

	if e.Ciphertext == "" {
		errors = append(errors, prefix+".ciphertext is required")
	}

	if e.SenderSignature == "" {
		errors = append(errors, prefix+".sender_signature is required")
	}

	if e.SenderPubKey == "" {
		errors = append(errors, prefix+".sender_pub_key is required")
	}

	return errors
}

func validateMessage(m *TicketMessage, index int) []string {
	var errors []string
	prefix := "messages[" + string(rune('0'+index)) + "]"

	if !uuidV7Pattern.MatchString(m.MessageID) {
		errors = append(errors, prefix+".message_id must be a valid UUID v7")
	}

	if !virtengineAddrPattern.MatchString(m.SenderAddr) {
		errors = append(errors, prefix+".sender_addr must be a valid VirtEngine address")
	}

	if !contains(validSenderRoles, m.SenderRole) {
		errors = append(errors, prefix+".sender_role is invalid: "+m.SenderRole)
	}

	if m.CreatedAt <= 0 {
		errors = append(errors, prefix+".created_at must be a positive Unix timestamp")
	}

	if m.ContentEnvelope == nil {
		errors = append(errors, prefix+".content_envelope is required")
	} else {
		envErrors := validateEnvelope(m.ContentEnvelope, prefix+".content_envelope")
		errors = append(errors, envErrors...)
	}

	return errors
}

func validateAttachment(a *AttachmentRef, index int) []string {
	var errors []string
	prefix := "attachments[" + string(rune('0'+index)) + "]"

	if !uuidV7Pattern.MatchString(a.AttachmentID) {
		errors = append(errors, prefix+".attachment_id must be a valid UUID v7")
	}

	if a.FileName == "" || len(a.FileName) > 255 {
		errors = append(errors, prefix+".file_name must be 1-255 characters")
	}

	if !contains(validContentTypes, a.ContentType) {
		errors = append(errors, prefix+".content_type is not allowed: "+a.ContentType)
	}

	if a.SizeBytes < 1 || a.SizeBytes > 52428800 {
		errors = append(errors, prefix+".size_bytes must be 1-52428800 (50MB)")
	}

	if a.StorageLocation == "" {
		errors = append(errors, prefix+".storage_location is required")
	}

	if a.UploadedAt <= 0 {
		errors = append(errors, prefix+".uploaded_at must be a positive Unix timestamp")
	}

	if !virtengineAddrPattern.MatchString(a.UploadedBy) {
		errors = append(errors, prefix+".uploaded_by must be a valid VirtEngine address")
	}

	if !sha256Pattern.MatchString(a.ChecksumSHA256) {
		errors = append(errors, prefix+".checksum_sha256 must be a valid SHA-256 hash")
	}

	if a.KeyEnvelope == nil {
		errors = append(errors, prefix+".key_envelope is required")
	} else {
		envErrors := validateEnvelope(a.KeyEnvelope, prefix+".key_envelope")
		errors = append(errors, envErrors...)
	}

	return errors
}

// IsValidTransition checks if a state transition is allowed
func IsValidTransition(from, to string) bool {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return contains(allowed, to)
}

// GetRetentionDaysForJurisdiction returns the default retention days for a jurisdiction
func GetRetentionDaysForJurisdiction(code string) int {
	// GDPR countries
	gdprCountries := []string{
		"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR",
		"DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL",
		"PL", "PT", "RO", "SK", "SI", "ES", "SE",
	}

	// Extract country code (first 2 chars)
	if len(code) >= 2 {
		countryCode := code[:2]

		if countryCode == "GB" {
			return 730 // UK GDPR
		}

		for _, c := range gdprCountries {
			if c == countryCode {
				return 730 // GDPR - 2 years
			}
		}

		if countryCode == "US" {
			if len(code) > 2 && code[2:] == "-CA" {
				return 1095 // CCPA - 3 years
			}
			return 1825 // US General - 5 years
		}
	}

	return 1095 // Global default - 3 years
}

// Tests

func TestValidateBasicTicket(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "support", "examples", "basic_ticket.json"))
	if err != nil {
		t.Skipf("Example file not found: %v", err)
	}

	var ticket SupportTicket
	if err := json.Unmarshal(data, &ticket); err != nil {
		t.Fatalf("Failed to unmarshal basic_ticket.json: %v", err)
	}

	errors := ValidateTicket(&ticket)
	if len(errors) > 0 {
		t.Errorf("basic_ticket.json validation failed:\n%v", errors)
	}
}

func TestValidateMarketplaceTicket(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "support", "examples", "marketplace_ticket.json"))
	if err != nil {
		t.Skipf("Example file not found: %v", err)
	}

	var ticket SupportTicket
	if err := json.Unmarshal(data, &ticket); err != nil {
		t.Fatalf("Failed to unmarshal marketplace_ticket.json: %v", err)
	}

	errors := ValidateTicket(&ticket)
	if len(errors) > 0 {
		t.Errorf("marketplace_ticket.json validation failed:\n%v", errors)
	}
}

func TestValidateIdentityTicket(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "support", "examples", "identity_ticket.json"))
	if err != nil {
		t.Skipf("Example file not found: %v", err)
	}

	var ticket SupportTicket
	if err := json.Unmarshal(data, &ticket); err != nil {
		t.Fatalf("Failed to unmarshal identity_ticket.json: %v", err)
	}

	errors := ValidateTicket(&ticket)
	if len(errors) > 0 {
		t.Errorf("identity_ticket.json validation failed:\n%v", errors)
	}
}

func TestTicketIDFormat(t *testing.T) {
	testCases := []struct {
		id      string
		isValid bool
	}{
		{"019479b8-7c3d-7000-8000-000000000001", true},  // Valid UUID v7
		{"019479b8-7c3d-7000-9000-000000000001", true},  // Valid UUID v7 (variant 9)
		{"019479b8-7c3d-7000-a000-000000000001", true},  // Valid UUID v7 (variant a)
		{"019479b8-7c3d-7000-b000-000000000001", true},  // Valid UUID v7 (variant b)
		{"019479b8-7c3d-4000-8000-000000000001", false}, // UUID v4, not v7
		{"not-a-uuid", false},
		{"", false},
		{"019479b8-7c3d-7000-c000-000000000001", false}, // Invalid variant
	}

	for _, tc := range testCases {
		result := uuidV7Pattern.MatchString(tc.id)
		if result != tc.isValid {
			t.Errorf("UUID v7 validation for %q: got %v, want %v", tc.id, result, tc.isValid)
		}
	}
}

func TestVirtEngineAddressFormat(t *testing.T) {
	testCases := []struct {
		addr    string
		isValid bool
	}{
		{"virtengine1abc123def456ghi789jkl012mno345pq", true},                  // 32 chars after prefix
		{"virtengine1abcdefghijklmnopqrstuvwxyz012345", true},                  // 32 chars after prefix
		{"virtengine1abc123def456ghi789jkl012mno345pqrstuvwxyz01234567", true}, // 50 chars
		{"cosmos1abc123def456ghi789jkl012mno345pqr678st", false},               // Wrong prefix
		{"virtengine1ABC", false},                                              // Uppercase not allowed
		{"", false},
	}

	for _, tc := range testCases {
		result := virtengineAddrPattern.MatchString(tc.addr)
		if result != tc.isValid {
			t.Errorf("VirtEngine address validation for %q: got %v, want %v", tc.addr, result, tc.isValid)
		}
	}
}

func TestStateTransitions(t *testing.T) {
	testCases := []struct {
		from    string
		to      string
		allowed bool
	}{
		{"OPEN", "ASSIGNED", true},
		{"OPEN", "CANCELLED", true},
		{"OPEN", "CLOSED", false},
		{"ASSIGNED", "IN_PROGRESS", true},
		{"ASSIGNED", "CLOSED", false},
		{"IN_PROGRESS", "AWAITING_CUSTOMER", true},
		{"IN_PROGRESS", "RESOLVED", true},
		{"RESOLVED", "CLOSED", true},
		{"RESOLVED", "IN_PROGRESS", true}, // Reopen
		{"CLOSED", "OPEN", false},         // Terminal state
		{"CANCELLED", "OPEN", false},      // Terminal state
	}

	for _, tc := range testCases {
		result := IsValidTransition(tc.from, tc.to)
		if result != tc.allowed {
			t.Errorf("Transition %s -> %s: got %v, want %v", tc.from, tc.to, result, tc.allowed)
		}
	}
}

func TestRetentionPolicyByJurisdiction(t *testing.T) {
	testCases := []struct {
		code     string
		expected int
	}{
		{"DE", 730},     // Germany - GDPR
		{"FR", 730},     // France - GDPR
		{"GB", 730},     // UK - UK GDPR
		{"US-CA", 1095}, // California - CCPA
		{"US-NY", 1825}, // New York - US General
		{"US", 1825},    // US General
		{"AU", 1095},    // Australia - Global default
		{"JP", 1095},    // Japan - Global default
	}

	for _, tc := range testCases {
		result := GetRetentionDaysForJurisdiction(tc.code)
		if result != tc.expected {
			t.Errorf("Retention for %s: got %d, want %d", tc.code, result, tc.expected)
		}
	}
}

func TestTerminalStatesRequireClosedAt(t *testing.T) {
	ticket := &SupportTicket{
		TicketID:        "019479b8-7c3d-7000-8000-000000000001",
		Version:         1,
		Category:        "technical",
		Priority:        "medium",
		Status:          "CLOSED",
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
		CustomerAddress: "virtengine1abc123def456ghi789jkl012mno345pqr678",
		Subject:         "Test ticket",
		DescriptionEnvelope: &EncryptedPayloadEnvelope{
			Version:          1,
			AlgorithmID:      "X25519-XSALSA20-POLY1305",
			AlgorithmVersion: 1,
			RecipientKeyIDs:  []string{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
			Nonce:            "dGVzdE5vbmNl",
			Ciphertext:       "dGVzdENpcGhlcnRleHQ=",
			SenderSignature:  "dGVzdFNpZw==",
			SenderPubKey:     "dGVzdFB1YktleQ==",
		},
		RetentionPolicy: RetentionPolicyRef{
			PolicyID:          "global",
			RetentionDays:     1095,
			DeleteAfterClosed: true,
		},
		JurisdictionCode: "US",
		// Note: ClosedAt is missing
	}

	errors := ValidateTicket(ticket)
	found := false
	for _, e := range errors {
		if e == "closed_at is required for terminal states" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected validation error for missing closed_at in CLOSED state")
	}
}

func TestEnvelopeValidation(t *testing.T) {
	// Test empty envelope
	emptyEnv := &EncryptedPayloadEnvelope{}
	errors := validateEnvelope(emptyEnv, "test")
	if len(errors) == 0 {
		t.Error("Expected validation errors for empty envelope")
	}

	// Test valid envelope
	validEnv := &EncryptedPayloadEnvelope{
		Version:          1,
		AlgorithmID:      "X25519-XSALSA20-POLY1305",
		AlgorithmVersion: 1,
		RecipientKeyIDs:  []string{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
		Nonce:            "dGVzdE5vbmNl",
		Ciphertext:       "dGVzdENpcGhlcnRleHQ=",
		SenderSignature:  "dGVzdFNpZw==",
		SenderPubKey:     "dGVzdFB1YktleQ==",
	}
	errors = validateEnvelope(validEnv, "test")
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors for valid envelope, got: %v", errors)
	}

	// Test invalid algorithm
	invalidAlgEnv := &EncryptedPayloadEnvelope{
		Version:          1,
		AlgorithmID:      "INVALID-ALGORITHM",
		AlgorithmVersion: 1,
		RecipientKeyIDs:  []string{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
		Nonce:            "dGVzdE5vbmNl",
		Ciphertext:       "dGVzdENpcGhlcnRleHQ=",
		SenderSignature:  "dGVzdFNpZw==",
		SenderPubKey:     "dGVzdFB1YktleQ==",
	}
	errors = validateEnvelope(invalidAlgEnv, "test")
	found := false
	for _, e := range errors {
		if e == "test.algorithm_id must be a supported algorithm" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected validation error for invalid algorithm")
	}
}

func TestAttachmentSizeLimit(t *testing.T) {
	attachment := &AttachmentRef{
		AttachmentID:    "019479b8-7c3d-7002-8000-000000000001",
		FileName:        "large_file.zip",
		ContentType:     "application/zip",
		SizeBytes:       60 * 1024 * 1024, // 60MB - exceeds 50MB limit
		StorageLocation: "ipfs://Qm...",
		UploadedAt:      time.Now().Unix(),
		UploadedBy:      "virtengine1abc123def456ghi789jkl012mno345pqr678",
		ChecksumSHA256:  "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
		KeyEnvelope: &EncryptedPayloadEnvelope{
			Version:          1,
			AlgorithmID:      "X25519-XSALSA20-POLY1305",
			AlgorithmVersion: 1,
			RecipientKeyIDs:  []string{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
			Nonce:            "dGVzdE5vbmNl",
			Ciphertext:       "dGVzdENpcGhlcnRleHQ=",
			SenderSignature:  "dGVzdFNpZw==",
			SenderPubKey:     "dGVzdFB1YktleQ==",
		},
	}

	errors := validateAttachment(attachment, 0)
	found := false
	for _, e := range errors {
		if e == "attachments[0].size_bytes must be 1-52428800 (50MB)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected validation error for oversized attachment")
	}
}

func TestInvalidContentType(t *testing.T) {
	attachment := &AttachmentRef{
		AttachmentID:    "019479b8-7c3d-7002-8000-000000000001",
		FileName:        "malware.exe",
		ContentType:     "application/x-msdownload", // Executable - not allowed
		SizeBytes:       1024,
		StorageLocation: "ipfs://Qm...",
		UploadedAt:      time.Now().Unix(),
		UploadedBy:      "virtengine1abc123def456ghi789jkl012mno345pqr678",
		ChecksumSHA256:  "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
		KeyEnvelope: &EncryptedPayloadEnvelope{
			Version:          1,
			AlgorithmID:      "X25519-XSALSA20-POLY1305",
			AlgorithmVersion: 1,
			RecipientKeyIDs:  []string{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
			Nonce:            "dGVzdE5vbmNl",
			Ciphertext:       "dGVzdENpcGhlcnRleHQ=",
			SenderSignature:  "dGVzdFNpZw==",
			SenderPubKey:     "dGVzdFB1YktleQ==",
		},
	}

	errors := validateAttachment(attachment, 0)
	found := false
	for _, e := range errors {
		if e == "attachments[0].content_type is not allowed: application/x-msdownload" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected validation error for disallowed content type")
	}
}
