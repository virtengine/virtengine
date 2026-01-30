// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SECURITY-007: Access Controls and Audit Logging
// This file provides access controls and comprehensive audit logging for key management.
package provider_daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// Sentinel errors for audit and access control
var (
	// ErrAccessDenied is returned when access is denied
	ErrAccessDenied = verrors.ErrUnauthorized

	// ErrInvalidCredentials is returned for invalid credentials
	ErrInvalidCredentials = verrors.ErrUnauthorized

	// ErrSessionExpired is returned when a session has expired
	ErrSessionExpired = verrors.ErrExpired

	// ErrInsufficientPermissions is returned for insufficient permissions
	ErrInsufficientPermissions = verrors.ErrForbidden
)

// Permission represents a key management permission
type Permission string

const (
	// PermissionKeyCreate allows creating keys
	PermissionKeyCreate Permission = "key:create"

	// PermissionKeyRead allows reading key information
	PermissionKeyRead Permission = "key:read"

	// PermissionKeySign allows signing with keys
	PermissionKeySign Permission = "key:sign"

	// PermissionKeyRotate allows rotating keys
	PermissionKeyRotate Permission = "key:rotate"

	// PermissionKeyRevoke allows revoking keys
	PermissionKeyRevoke Permission = "key:revoke"

	// PermissionKeyDestroy allows destroying keys
	PermissionKeyDestroy Permission = "key:destroy"

	// PermissionKeyBackup allows backing up keys
	PermissionKeyBackup Permission = "key:backup"

	// PermissionKeyRestore allows restoring keys
	PermissionKeyRestore Permission = "key:restore"

	// PermissionKeyImport allows importing keys
	PermissionKeyImport Permission = "key:import"

	// PermissionKeyExportPub allows exporting public keys
	PermissionKeyExportPub Permission = "key:export_public"

	// PermissionAuditRead allows reading audit logs
	PermissionAuditRead Permission = "audit:read"

	// PermissionAuditExport allows exporting audit logs
	PermissionAuditExport Permission = "audit:export"

	// PermissionPolicyManage allows managing policies
	PermissionPolicyManage Permission = "policy:manage"

	// PermissionHSMManage allows managing HSM configuration
	PermissionHSMManage Permission = "hsm:manage"

	// PermissionMultiSigManage allows managing multisig configurations
	PermissionMultiSigManage Permission = "multisig:manage"

	// PermissionAdmin grants all permissions
	PermissionAdmin Permission = "admin:*"
)

// AllPermissions returns all defined permissions
func AllPermissions() []Permission {
	return []Permission{
		PermissionKeyCreate, PermissionKeyRead, PermissionKeySign,
		PermissionKeyRotate, PermissionKeyRevoke, PermissionKeyDestroy,
		PermissionKeyBackup, PermissionKeyRestore, PermissionKeyImport,
		PermissionKeyExportPub, PermissionAuditRead, PermissionAuditExport,
		PermissionPolicyManage, PermissionHSMManage, PermissionMultiSigManage,
		PermissionAdmin,
	}
}

// Role represents an access control role
type Role struct {
	// Name is the role name
	Name string `json:"name"`

	// Description is the role description
	Description string `json:"description"`

	// Permissions is the list of permissions
	Permissions []Permission `json:"permissions"`

	// CreatedAt is when the role was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the role was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// HasPermission checks if the role has a permission
func (r *Role) HasPermission(perm Permission) bool {
	for _, p := range r.Permissions {
		if p == PermissionAdmin || p == perm {
			return true
		}
	}
	return false
}

// Principal represents a security principal (user/service)
type Principal struct {
	// ID is the principal ID
	ID string `json:"id"`

	// Type is the principal type (user, service, api_key)
	Type string `json:"type"`

	// Name is the principal name
	Name string `json:"name"`

	// Roles is the list of assigned roles
	Roles []string `json:"roles"`

	// DirectPermissions is any directly assigned permissions
	DirectPermissions []Permission `json:"direct_permissions,omitempty"`

	// CreatedAt is when the principal was created
	CreatedAt time.Time `json:"created_at"`

	// LastActiveAt is when the principal was last active
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`

	// Enabled indicates if the principal is enabled
	Enabled bool `json:"enabled"`

	// Metadata contains additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Session represents an authenticated session
type Session struct {
	// ID is the session ID
	ID string `json:"id"`

	// PrincipalID is the authenticated principal
	PrincipalID string `json:"principal_id"`

	// CreatedAt is when the session was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the session expires
	ExpiresAt time.Time `json:"expires_at"`

	// LastActivityAt is the last activity time
	LastActivityAt time.Time `json:"last_activity_at"`

	// SourceIP is the source IP address
	SourceIP string `json:"source_ip,omitempty"`

	// UserAgent is the user agent
	UserAgent string `json:"user_agent,omitempty"`

	// Valid indicates if the session is valid
	Valid bool `json:"valid"`
}

// AccessControlConfig configures access control
type AccessControlConfig struct {
	// SessionTimeoutMinutes is the session timeout
	SessionTimeoutMinutes int `json:"session_timeout_minutes"`

	// MaxSessionsPerPrincipal is the max concurrent sessions
	MaxSessionsPerPrincipal int `json:"max_sessions_per_principal"`

	// RequireMFA requires MFA for sensitive operations
	RequireMFA bool `json:"require_mfa"`

	// MFARequiredOperations lists operations requiring MFA
	MFARequiredOperations []Permission `json:"mfa_required_operations"`

	// EnableIPRestriction enables IP-based restrictions
	EnableIPRestriction bool `json:"enable_ip_restriction"`

	// AllowedIPRanges is the list of allowed IP ranges
	AllowedIPRanges []string `json:"allowed_ip_ranges,omitempty"`
}

// DefaultAccessControlConfig returns the default configuration
func DefaultAccessControlConfig() *AccessControlConfig {
	return &AccessControlConfig{
		SessionTimeoutMinutes:   60,
		MaxSessionsPerPrincipal: 5,
		RequireMFA:              true,
		MFARequiredOperations: []Permission{
			PermissionKeyDestroy,
			PermissionKeyBackup,
			PermissionKeyRestore,
			PermissionPolicyManage,
		},
		EnableIPRestriction: false,
	}
}

// AccessController manages access control
type AccessController struct {
	config     *AccessControlConfig
	roles      map[string]*Role
	principals map[string]*Principal
	sessions   map[string]*Session
	mu         sync.RWMutex
}

// NewAccessController creates a new access controller
func NewAccessController(config *AccessControlConfig) *AccessController {
	if config == nil {
		config = DefaultAccessControlConfig()
	}

	ac := &AccessController{
		config:     config,
		roles:      make(map[string]*Role),
		principals: make(map[string]*Principal),
		sessions:   make(map[string]*Session),
	}

	// Create default roles
	ac.createDefaultRoles()

	return ac
}

// createDefaultRoles creates default roles
func (ac *AccessController) createDefaultRoles() {
	now := time.Now().UTC()

	// Admin role
	ac.roles["admin"] = &Role{
		Name:        "admin",
		Description: "Full administrative access",
		Permissions: []Permission{PermissionAdmin},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Operator role
	ac.roles["operator"] = &Role{
		Name:        "operator",
		Description: "Key operations access",
		Permissions: []Permission{
			PermissionKeyCreate, PermissionKeyRead, PermissionKeySign,
			PermissionKeyRotate, PermissionKeyExportPub,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Auditor role
	ac.roles["auditor"] = &Role{
		Name:        "auditor",
		Description: "Read-only audit access",
		Permissions: []Permission{
			PermissionKeyRead, PermissionAuditRead, PermissionAuditExport,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Signer role
	ac.roles["signer"] = &Role{
		Name:        "signer",
		Description: "Signing-only access",
		Permissions: []Permission{
			PermissionKeyRead, PermissionKeySign,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CreateRole creates a new role
func (ac *AccessController) CreateRole(role *Role) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, exists := ac.roles[role.Name]; exists {
		return fmt.Errorf("role already exists: %s", role.Name)
	}

	now := time.Now().UTC()
	role.CreatedAt = now
	role.UpdatedAt = now

	ac.roles[role.Name] = role
	return nil
}

// GetRole retrieves a role
func (ac *AccessController) GetRole(name string) (*Role, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	role, exists := ac.roles[name]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", name)
	}

	return role, nil
}

// CreatePrincipal creates a new principal
func (ac *AccessController) CreatePrincipal(principal *Principal) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, exists := ac.principals[principal.ID]; exists {
		return fmt.Errorf("principal already exists: %s", principal.ID)
	}

	principal.CreatedAt = time.Now().UTC()
	principal.Enabled = true

	ac.principals[principal.ID] = principal
	return nil
}

// GetPrincipal retrieves a principal
func (ac *AccessController) GetPrincipal(id string) (*Principal, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	principal, exists := ac.principals[id]
	if !exists {
		return nil, fmt.Errorf("principal not found: %s", id)
	}

	return principal, nil
}

// CreateSession creates a new session for a principal
func (ac *AccessController) CreateSession(principalID, sourceIP, userAgent string) (*Session, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	principal, exists := ac.principals[principalID]
	if !exists {
		return nil, ErrInvalidCredentials
	}

	if !principal.Enabled {
		return nil, ErrAccessDenied
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(ac.config.SessionTimeoutMinutes) * time.Minute)

	sessionID := generateSessionID(principalID, now)

	session := &Session{
		ID:             sessionID,
		PrincipalID:    principalID,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
		LastActivityAt: now,
		SourceIP:       sourceIP,
		UserAgent:      userAgent,
		Valid:          true,
	}

	ac.sessions[sessionID] = session

	// Update principal's last active time
	principal.LastActiveAt = &now

	return session, nil
}

// ValidateSession validates a session
func (ac *AccessController) ValidateSession(sessionID string) (*Session, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	session, exists := ac.sessions[sessionID]
	if !exists {
		return nil, ErrSessionExpired
	}

	if !session.Valid {
		return nil, ErrSessionExpired
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// RevokeSession revokes a session
func (ac *AccessController) RevokeSession(sessionID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	session, exists := ac.sessions[sessionID]
	if !exists {
		return nil
	}

	session.Valid = false
	return nil
}

// CheckPermission checks if a session has a permission
func (ac *AccessController) CheckPermission(sessionID string, permission Permission) error {
	session, err := ac.ValidateSession(sessionID)
	if err != nil {
		return err
	}

	ac.mu.RLock()
	defer ac.mu.RUnlock()

	principal, exists := ac.principals[session.PrincipalID]
	if !exists || !principal.Enabled {
		return ErrAccessDenied
	}

	// Check direct permissions
	for _, p := range principal.DirectPermissions {
		if p == PermissionAdmin || p == permission {
			return nil
		}
	}

	// Check role permissions
	for _, roleName := range principal.Roles {
		role, exists := ac.roles[roleName]
		if exists && role.HasPermission(permission) {
			return nil
		}
	}

	return ErrInsufficientPermissions
}

// Authorize is a convenience function to check permission and return error
func (ac *AccessController) Authorize(sessionID string, permission Permission) error {
	return ac.CheckPermission(sessionID, permission)
}

// generateSessionID generates a unique session ID
func generateSessionID(principalID string, timestamp time.Time) string {
	data := fmt.Sprintf("%s-%d", principalID, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// AuditEventKeyCreated indicates a key was created
	AuditEventKeyCreated AuditEventType = "key_created"

	// AuditEventKeyRead indicates a key was read
	AuditEventKeyRead AuditEventType = "key_read"

	// AuditEventKeySigned indicates a key was used for signing
	AuditEventKeySigned AuditEventType = "key_signed"

	// AuditEventKeyRotated indicates a key was rotated
	AuditEventKeyRotated AuditEventType = "key_rotated"

	// AuditEventKeyRevoked indicates a key was revoked
	AuditEventKeyRevoked AuditEventType = "key_revoked"

	// AuditEventKeyDestroyed indicates a key was destroyed
	AuditEventKeyDestroyed AuditEventType = "key_destroyed"

	// AuditEventKeyBackup indicates a key backup was created
	AuditEventKeyBackup AuditEventType = "key_backup"

	// AuditEventKeyRestore indicates keys were restored
	AuditEventKeyRestore AuditEventType = "key_restore"

	// AuditEventKeyImport indicates a key was imported
	AuditEventKeyImport AuditEventType = "key_import"

	// AuditEventSessionCreated indicates a session was created
	AuditEventSessionCreated AuditEventType = "session_created"

	// AuditEventSessionRevoked indicates a session was revoked
	AuditEventSessionRevoked AuditEventType = "session_revoked"

	// AuditEventAccessDenied indicates access was denied
	AuditEventAccessDenied AuditEventType = "access_denied"

	// AuditEventPolicyChanged indicates a policy was changed
	AuditEventPolicyChanged AuditEventType = "policy_changed"

	// AuditEventCompromiseDetected indicates compromise was detected
	AuditEventCompromiseDetected AuditEventType = "compromise_detected"

	// AuditEventHSMOperation indicates an HSM operation
	AuditEventHSMOperation AuditEventType = "hsm_operation"

	// AuditEventMultiSigOperation indicates a multisig operation
	AuditEventMultiSigOperation AuditEventType = "multisig_operation"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	// ID is the unique event ID
	ID string `json:"id"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Type is the event type
	Type AuditEventType `json:"type"`

	// SessionID is the session that triggered the event
	SessionID string `json:"session_id,omitempty"`

	// PrincipalID is the principal that triggered the event
	PrincipalID string `json:"principal_id,omitempty"`

	// PrincipalName is the principal name
	PrincipalName string `json:"principal_name,omitempty"`

	// KeyID is the affected key (if applicable)
	KeyID string `json:"key_id,omitempty"`

	// Operation is the operation performed
	Operation string `json:"operation"`

	// Success indicates if the operation succeeded
	Success bool `json:"success"`

	// ErrorMessage is the error message (if failed)
	ErrorMessage string `json:"error_message,omitempty"`

	// SourceIP is the source IP address
	SourceIP string `json:"source_ip,omitempty"`

	// UserAgent is the user agent
	UserAgent string `json:"user_agent,omitempty"`

	// Details contains additional event details
	Details map[string]interface{} `json:"details,omitempty"`

	// Hash is the cryptographic hash for integrity
	Hash string `json:"hash"`

	// PreviousHash is the hash of the previous event
	PreviousHash string `json:"previous_hash,omitempty"`
}

// AuditLogConfig configures audit logging
type AuditLogConfig struct {
	// Enabled indicates if audit logging is enabled
	Enabled bool `json:"enabled"`

	// LogFile is the audit log file path
	LogFile string `json:"log_file"`

	// MaxFileSizeMB is the maximum log file size in MB
	MaxFileSizeMB int `json:"max_file_size_mb"`

	// MaxBackups is the maximum number of backup files
	MaxBackups int `json:"max_backups"`

	// RetentionDays is how long to retain audit logs
	RetentionDays int `json:"retention_days"`

	// IncludeDetails includes detailed event information
	IncludeDetails bool `json:"include_details"`

	// EnableChaining enables hash chaining for integrity
	EnableChaining bool `json:"enable_chaining"`

	// SignEvents cryptographically signs events
	SignEvents bool `json:"sign_events"`

	// RemoteEndpoint sends events to remote endpoint
	RemoteEndpoint string `json:"remote_endpoint,omitempty"`
}

// DefaultAuditLogConfig returns the default audit configuration
func DefaultAuditLogConfig() *AuditLogConfig {
	return &AuditLogConfig{
		Enabled:        true,
		LogFile:        "audit.log",
		MaxFileSizeMB:  100,
		MaxBackups:     10,
		RetentionDays:  365,
		IncludeDetails: true,
		EnableChaining: true,
		SignEvents:     false,
	}
}

// AuditLogger logs audit events
type AuditLogger struct {
	config       *AuditLogConfig
	events       []*AuditEvent
	lastHash     string
	file         *os.File
	mu           sync.RWMutex
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config *AuditLogConfig) (*AuditLogger, error) {
	if config == nil {
		config = DefaultAuditLogConfig()
	}

	logger := &AuditLogger{
		config:   config,
		events:   make([]*AuditEvent, 0),
		lastHash: "",
	}

	if config.LogFile != "" {
		dir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log: %w", err)
		}
		logger.file = file
	}

	return logger, nil
}

// Close closes the audit logger
func (l *AuditLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Log logs an audit event
func (l *AuditLogger) Log(event *AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.config.Enabled {
		return nil
	}

	// Set timestamp and ID
	event.Timestamp = time.Now().UTC()
	event.ID = generateAuditEventID(event)

	// Add hash chaining
	if l.config.EnableChaining {
		event.PreviousHash = l.lastHash
		event.Hash = computeEventHash(event)
		l.lastHash = event.Hash
	} else {
		event.Hash = computeEventHash(event)
	}

	// Store in memory
	l.events = append(l.events, event)

	// Write to file
	if l.file != nil {
		if err := l.writeEvent(event); err != nil {
			return err
		}
	}

	// Send to remote endpoint if configured with panic recovery
	if l.config.RemoteEndpoint != "" {
		verrors.SafeGo("provider-daemon:audit-remote", func() {
			l.sendToRemote(event)
		})
	}

	return nil
}

// writeEvent writes an event to the log file
func (l *AuditLogger) writeEvent(event *AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = l.file.Write(append(data, '\n'))
	return err
}

// sendToRemote sends an event to a remote endpoint
func (l *AuditLogger) sendToRemote(event *AuditEvent) {
	// In production, would send HTTP POST to remote endpoint
	_ = event
}

// LogKeyOperation logs a key operation
func (l *AuditLogger) LogKeyOperation(eventType AuditEventType, sessionID, principalID, principalName, keyID, operation string, success bool, errorMsg string, details map[string]interface{}) error {
	event := &AuditEvent{
		Type:          eventType,
		SessionID:     sessionID,
		PrincipalID:   principalID,
		PrincipalName: principalName,
		KeyID:         keyID,
		Operation:     operation,
		Success:       success,
		ErrorMessage:  errorMsg,
		Details:       details,
	}

	return l.Log(event)
}

// LogAccessDenied logs an access denied event
func (l *AuditLogger) LogAccessDenied(sessionID, principalID, operation, reason, sourceIP string) error {
	event := &AuditEvent{
		Type:         AuditEventAccessDenied,
		SessionID:    sessionID,
		PrincipalID:  principalID,
		Operation:    operation,
		Success:      false,
		ErrorMessage: reason,
		SourceIP:     sourceIP,
	}

	return l.Log(event)
}

// LogSessionEvent logs a session event
func (l *AuditLogger) LogSessionEvent(eventType AuditEventType, sessionID, principalID, sourceIP, userAgent string) error {
	event := &AuditEvent{
		Type:        eventType,
		SessionID:   sessionID,
		PrincipalID: principalID,
		Operation:   string(eventType),
		Success:     true,
		SourceIP:    sourceIP,
		UserAgent:   userAgent,
	}

	return l.Log(event)
}

// GetEvents retrieves audit events
func (l *AuditLogger) GetEvents(since time.Time, eventType *AuditEventType) []*AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	events := make([]*AuditEvent, 0)

	for _, event := range l.events {
		if event.Timestamp.Before(since) {
			continue
		}
		if eventType != nil && event.Type != *eventType {
			continue
		}
		events = append(events, event)
	}

	return events
}

// GetEventsByKey retrieves events for a specific key
func (l *AuditLogger) GetEventsByKey(keyID string, since time.Time) []*AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	events := make([]*AuditEvent, 0)

	for _, event := range l.events {
		if event.KeyID != keyID {
			continue
		}
		if event.Timestamp.Before(since) {
			continue
		}
		events = append(events, event)
	}

	return events
}

// GetEventsByPrincipal retrieves events for a specific principal
func (l *AuditLogger) GetEventsByPrincipal(principalID string, since time.Time) []*AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	events := make([]*AuditEvent, 0)

	for _, event := range l.events {
		if event.PrincipalID != principalID {
			continue
		}
		if event.Timestamp.Before(since) {
			continue
		}
		events = append(events, event)
	}

	return events
}

// ExportEvents exports audit events to a writer
func (l *AuditLogger) ExportEvents(writer io.Writer, since time.Time, format string) error {
	events := l.GetEvents(since, nil)

	// Sort by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(events)
	case "jsonl":
		for _, event := range events {
			data, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err := writer.Write(append(data, '\n')); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// VerifyIntegrity verifies the integrity of the audit log
func (l *AuditLogger) VerifyIntegrity() (bool, []string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	errors := make([]string, 0)

	if !l.config.EnableChaining {
		return true, errors
	}

	var previousHash string
	for i, event := range l.events {
		// Verify previous hash
		if event.PreviousHash != previousHash {
			errors = append(errors, fmt.Sprintf("event %d: previous hash mismatch", i))
		}

		// Verify event hash
		computedHash := computeEventHash(event)
		if computedHash != event.Hash {
			errors = append(errors, fmt.Sprintf("event %d: hash mismatch", i))
		}

		previousHash = event.Hash
	}

	return len(errors) == 0, errors
}

// Cleanup removes old audit events based on retention policy
func (l *AuditLogger) Cleanup() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -l.config.RetentionDays)
	count := 0

	newEvents := make([]*AuditEvent, 0)
	for _, event := range l.events {
		if event.Timestamp.Before(cutoff) {
			count++
		} else {
			newEvents = append(newEvents, event)
		}
	}

	l.events = newEvents
	return count
}

// generateAuditEventID generates a unique audit event ID
func generateAuditEventID(event *AuditEvent) string {
	data := fmt.Sprintf("%s-%s-%d", event.Type, event.KeyID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// computeEventHash computes the hash of an audit event
func computeEventHash(event *AuditEvent) string {
	// Create a copy without the hash fields
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%t|%s",
		event.ID, event.Timestamp.Format(time.RFC3339Nano),
		event.Type, event.PrincipalID, event.KeyID,
		event.Operation, event.Success, event.PreviousHash)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateAuditReport generates an audit report
func (l *AuditLogger) GenerateAuditReport(since time.Time) *AuditReport {
	events := l.GetEvents(since, nil)

	report := &AuditReport{
		GeneratedAt: time.Now().UTC(),
		PeriodStart: since,
		PeriodEnd:   time.Now().UTC(),
		TotalEvents: len(events),
		ByType:      make(map[string]int),
		ByPrincipal: make(map[string]int),
		ByKey:       make(map[string]int),
	}

	for _, event := range events {
		report.ByType[string(event.Type)]++
		if event.PrincipalID != "" {
			report.ByPrincipal[event.PrincipalID]++
		}
		if event.KeyID != "" {
			report.ByKey[event.KeyID]++
		}
		if event.Success {
			report.SuccessCount++
		} else {
			report.FailureCount++
		}
	}

	return report
}

// AuditReport contains audit statistics
type AuditReport struct {
	GeneratedAt  time.Time      `json:"generated_at"`
	PeriodStart  time.Time      `json:"period_start"`
	PeriodEnd    time.Time      `json:"period_end"`
	TotalEvents  int            `json:"total_events"`
	SuccessCount int            `json:"success_count"`
	FailureCount int            `json:"failure_count"`
	ByType       map[string]int `json:"by_type"`
	ByPrincipal  map[string]int `json:"by_principal"`
	ByKey        map[string]int `json:"by_key"`
}
