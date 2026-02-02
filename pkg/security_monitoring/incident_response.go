package security_monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// IncidentResponder handles automated incident response
type IncidentResponder struct {
	logger       zerolog.Logger
	metrics      *SecurityMetrics
	playbooks    map[string]*Playbook
	playbookPath string
	mu           sync.RWMutex
}

// Playbook defines an automated incident response playbook
type Playbook struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	TriggerTypes    []string              `json:"trigger_types"`
	MinSeverity     SecurityEventSeverity `json:"min_severity"`
	Steps           []PlaybookStep        `json:"steps"`
	Enabled         bool                  `json:"enabled"`
	CooldownMinutes int                   `json:"cooldown_minutes"`
	MaxExecutions   int                   `json:"max_executions_per_hour"`
	NotifyOnSuccess bool                  `json:"notify_on_success"`
	NotifyOnFailure bool                  `json:"notify_on_failure"`
}

// PlaybookStep defines a single step in a playbook
type PlaybookStep struct {
	Name              string            `json:"name"`
	Action            string            `json:"action"`
	Parameters        map[string]string `json:"parameters,omitempty"`
	Timeout           int               `json:"timeout_seconds"`
	ContinueOnFailure bool              `json:"continue_on_failure"`
	Condition         string            `json:"condition,omitempty"`
}

// PlaybookAction represents types of playbook actions
type PlaybookAction string

const (
	ActionLogEvent         PlaybookAction = "log_event"
	ActionSendAlert        PlaybookAction = "send_alert"
	ActionBlockIP          PlaybookAction = "block_ip"
	ActionRevokeKey        PlaybookAction = "revoke_key"
	ActionSuspendAccount   PlaybookAction = "suspend_account"
	ActionSuspendProvider  PlaybookAction = "suspend_provider"
	ActionIncreaseSeverity PlaybookAction = "increase_severity"
	ActionTriggerBackup    PlaybookAction = "trigger_backup"
	ActionNotifyTeam       PlaybookAction = "notify_team"
	ActionRunScript        PlaybookAction = "run_script"
	ActionUpdateFirewall   PlaybookAction = "update_firewall"
	ActionCollectEvidence  PlaybookAction = "collect_evidence"
	ActionEscalate         PlaybookAction = "escalate"
)

// PlaybookExecution tracks playbook execution
type PlaybookExecution struct {
	ID            string          `json:"id"`
	PlaybookID    string          `json:"playbook_id"`
	IncidentID    string          `json:"incident_id"`
	StartedAt     time.Time       `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	Status        string          `json:"status"` // running, completed, failed
	StepsExecuted []StepExecution `json:"steps_executed"`
	Error         string          `json:"error,omitempty"`
}

// StepExecution tracks individual step execution
type StepExecution struct {
	StepName    string     `json:"step_name"`
	Action      string     `json:"action"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Success     bool       `json:"success"`
	Output      string     `json:"output,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// NewIncidentResponder creates a new incident responder
func NewIncidentResponder(playbookPath string, logger zerolog.Logger) (*IncidentResponder, error) {
	ir := &IncidentResponder{
		logger:       logger.With().Str("component", "incident-responder").Logger(),
		metrics:      GetSecurityMetrics(),
		playbooks:    make(map[string]*Playbook),
		playbookPath: playbookPath,
	}

	// Load default playbooks
	ir.loadDefaultPlaybooks()

	return ir, nil
}

// loadDefaultPlaybooks loads the built-in default playbooks
func (ir *IncidentResponder) loadDefaultPlaybooks() {
	defaults := []*Playbook{
		{
			ID:          "ddos-response",
			Name:        "DDoS Attack Response",
			Description: "Automated response to DDoS indicators",
			TriggerTypes: []string{
				"rate_limit_breach",
				string(ProviderIndicatorRapidActivity),
			},
			MinSeverity:     SeverityHigh,
			CooldownMinutes: 5,
			MaxExecutions:   10,
			Enabled:         true,
			NotifyOnSuccess: true,
			NotifyOnFailure: true,
			Steps: []PlaybookStep{
				{Name: "Log Event", Action: string(ActionLogEvent), Timeout: 5},
				{Name: "Block Source IP", Action: string(ActionBlockIP), Timeout: 10,
					Parameters: map[string]string{"duration": "3600"}},
				{Name: "Notify Security Team", Action: string(ActionNotifyTeam), Timeout: 30,
					Parameters: map[string]string{"channel": "security-alerts"}},
			},
		},
		{
			ID:          "key-compromise-response",
			Name:        "Key Compromise Response",
			Description: "Automated response to key compromise indicators",
			TriggerTypes: []string{
				string(ProviderIndicatorKeyCompromise),
				string(CryptoAnomalySignatureFailure),
				string(CryptoAnomalyKeyReuse),
			},
			MinSeverity:     SeverityCritical,
			CooldownMinutes: 0, // No cooldown for critical
			MaxExecutions:   100,
			Enabled:         true,
			NotifyOnSuccess: true,
			NotifyOnFailure: true,
			Steps: []PlaybookStep{
				{Name: "Log Event", Action: string(ActionLogEvent), Timeout: 5},
				{Name: "Collect Evidence", Action: string(ActionCollectEvidence), Timeout: 60},
				{Name: "Revoke Compromised Key", Action: string(ActionRevokeKey), Timeout: 30},
				{Name: "Suspend Provider", Action: string(ActionSuspendProvider), Timeout: 30,
					ContinueOnFailure: true},
				{Name: "Notify Security Team", Action: string(ActionNotifyTeam), Timeout: 30,
					Parameters: map[string]string{"channel": "security-critical", "priority": "high"}},
				{Name: "Escalate to On-Call", Action: string(ActionEscalate), Timeout: 60},
			},
		},
		{
			ID:          "fraud-response",
			Name:        "Identity Fraud Response",
			Description: "Automated response to identity verification fraud",
			TriggerTypes: []string{
				string(FraudIndicatorDocumentTampering),
				string(FraudIndicatorReplayAttack),
				string(FraudIndicatorSyntheticIdentity),
				string(FraudIndicatorMultipleIdentities),
			},
			MinSeverity:     SeverityCritical,
			CooldownMinutes: 0,
			MaxExecutions:   50,
			Enabled:         true,
			NotifyOnSuccess: true,
			NotifyOnFailure: true,
			Steps: []PlaybookStep{
				{Name: "Log Event", Action: string(ActionLogEvent), Timeout: 5},
				{Name: "Collect Evidence", Action: string(ActionCollectEvidence), Timeout: 60},
				{Name: "Suspend Account", Action: string(ActionSuspendAccount), Timeout: 30},
				{Name: "Block Source IP", Action: string(ActionBlockIP), Timeout: 10,
					Parameters: map[string]string{"duration": "86400"}},
				{Name: "Notify Compliance Team", Action: string(ActionNotifyTeam), Timeout: 30,
					Parameters: map[string]string{"channel": "compliance-alerts"}},
			},
		},
		{
			ID:          "crypto-anomaly-response",
			Name:        "Cryptographic Anomaly Response",
			Description: "Response to cryptographic operation anomalies",
			TriggerTypes: []string{
				string(CryptoAnomalyWeakEntropy),
				string(CryptoAnomalyDeprecatedAlgorithm),
			},
			MinSeverity:     SeverityHigh,
			CooldownMinutes: 15,
			MaxExecutions:   20,
			Enabled:         true,
			NotifyOnSuccess: false,
			NotifyOnFailure: true,
			Steps: []PlaybookStep{
				{Name: "Log Event", Action: string(ActionLogEvent), Timeout: 5},
				{Name: "Increase Severity", Action: string(ActionIncreaseSeverity), Timeout: 5},
				{Name: "Notify Security Team", Action: string(ActionNotifyTeam), Timeout: 30,
					Parameters: map[string]string{"channel": "security-alerts"}},
			},
		},
		{
			ID:          "transaction-anomaly-response",
			Name:        "Transaction Anomaly Response",
			Description: "Response to suspicious transaction patterns",
			TriggerTypes: []string{
				"tx_replay_attempt",
				"tx_rapid_fire",
				"tx_velocity_violation",
			},
			MinSeverity:     SeverityHigh,
			CooldownMinutes: 5,
			MaxExecutions:   30,
			Enabled:         true,
			NotifyOnSuccess: false,
			NotifyOnFailure: true,
			Steps: []PlaybookStep{
				{Name: "Log Event", Action: string(ActionLogEvent), Timeout: 5},
				{Name: "Block Source IP", Action: string(ActionBlockIP), Timeout: 10,
					Parameters: map[string]string{"duration": "1800"}, ContinueOnFailure: true},
				{Name: "Send Alert", Action: string(ActionSendAlert), Timeout: 30},
			},
		},
	}

	for _, playbook := range defaults {
		ir.playbooks[playbook.ID] = playbook
	}

	ir.logger.Info().Int("count", len(defaults)).Msg("loaded default playbooks")
}

// HandleIncident handles a security incident by executing appropriate playbooks
func (ir *IncidentResponder) HandleIncident(ctx context.Context, incident *SecurityIncident) {
	ir.mu.RLock()
	playbooks := ir.findMatchingPlaybooks(incident)
	ir.mu.RUnlock()

	if len(playbooks) == 0 {
		ir.logger.Debug().
			Str("incident_id", incident.ID).
			Str("type", incident.Type).
			Msg("no matching playbooks for incident")
		return
	}

	for _, playbook := range playbooks {
		go ir.executePlaybook(ctx, playbook, incident)
	}
}

// findMatchingPlaybooks finds playbooks that match the incident
func (ir *IncidentResponder) findMatchingPlaybooks(incident *SecurityIncident) []*Playbook {
	var matches []*Playbook

	for _, playbook := range ir.playbooks {
		if !playbook.Enabled {
			continue
		}

		if incident.Severity < playbook.MinSeverity {
			continue
		}

		for _, triggerType := range playbook.TriggerTypes {
			if incident.Type == triggerType {
				matches = append(matches, playbook)
				break
			}
		}
	}

	return matches
}

// executePlaybook executes a playbook for an incident
func (ir *IncidentResponder) executePlaybook(ctx context.Context, playbook *Playbook, incident *SecurityIncident) {
	execution := &PlaybookExecution{
		ID:            generateExecutionID(),
		PlaybookID:    playbook.ID,
		IncidentID:    incident.ID,
		StartedAt:     time.Now(),
		Status:        "running",
		StepsExecuted: make([]StepExecution, 0),
	}

	ir.logger.Info().
		Str("playbook_id", playbook.ID).
		Str("incident_id", incident.ID).
		Str("execution_id", execution.ID).
		Msg("starting playbook execution")

	stepNames := make([]string, 0, len(playbook.Steps))
	var hasFailure bool

	for _, step := range playbook.Steps {
		stepExec := ir.executeStep(ctx, step, incident)
		execution.StepsExecuted = append(execution.StepsExecuted, stepExec)
		stepNames = append(stepNames, step.Name)

		if !stepExec.Success {
			hasFailure = true
			if !step.ContinueOnFailure {
				execution.Status = "failed"
				execution.Error = stepExec.Error
				break
			}
		}
	}

	if execution.Status == "running" {
		execution.Status = "completed"
	}

	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Update metrics
	ir.metrics.IncidentResponseActions.WithLabelValues(
		execution.Status, playbook.ID).Inc()

	ir.logger.Info().
		Str("playbook_id", playbook.ID).
		Str("incident_id", incident.ID).
		Str("execution_id", execution.ID).
		Str("status", execution.Status).
		Dur("duration", completedAt.Sub(execution.StartedAt)).
		Bool("has_failure", hasFailure).
		Msg("playbook execution completed")

	// Update incident with playbook info
	incident.PlaybookID = playbook.ID
}

// executeStep executes a single playbook step
func (ir *IncidentResponder) executeStep(ctx context.Context, step PlaybookStep, incident *SecurityIncident) StepExecution {
	exec := StepExecution{
		StepName:  step.Name,
		Action:    step.Action,
		StartedAt: time.Now(),
	}

	ir.logger.Debug().
		Str("step", step.Name).
		Str("action", step.Action).
		Msg("executing playbook step")

	// Create timeout context
	stepCtx, cancel := context.WithTimeout(ctx, time.Duration(step.Timeout)*time.Second)
	defer cancel()

	// Execute the action
	err := ir.executeAction(stepCtx, PlaybookAction(step.Action), step.Parameters, incident)

	completedAt := time.Now()
	exec.CompletedAt = &completedAt

	if err != nil {
		exec.Success = false
		exec.Error = err.Error()
		ir.logger.Error().
			Err(err).
			Str("step", step.Name).
			Str("action", step.Action).
			Msg("step execution failed")
	} else {
		exec.Success = true
		exec.Output = "success"
		ir.logger.Debug().
			Str("step", step.Name).
			Str("action", step.Action).
			Msg("step execution completed")
	}

	return exec
}

// executeAction executes a specific action
//
//nolint:unparam // ctx kept for future async action execution
func (ir *IncidentResponder) executeAction(
	_ context.Context,
	action PlaybookAction,
	params map[string]string,
	incident *SecurityIncident,
) error {
	switch action {
	case ActionLogEvent:
		return ir.actionLogEvent(incident)
	case ActionSendAlert:
		return ir.actionSendAlert(incident, params)
	case ActionBlockIP:
		return ir.actionBlockIP(incident, params)
	case ActionRevokeKey:
		return ir.actionRevokeKey(incident, params)
	case ActionSuspendAccount:
		return ir.actionSuspendAccount(incident, params)
	case ActionSuspendProvider:
		return ir.actionSuspendProvider(incident, params)
	case ActionIncreaseSeverity:
		return ir.actionIncreaseSeverity(incident)
	case ActionNotifyTeam:
		return ir.actionNotifyTeam(incident, params)
	case ActionCollectEvidence:
		return ir.actionCollectEvidence(incident)
	case ActionEscalate:
		return ir.actionEscalate(incident, params)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// Action implementations

func (ir *IncidentResponder) actionLogEvent(incident *SecurityIncident) error {
	ir.logger.Info().
		Str("incident_id", incident.ID).
		Str("type", incident.Type).
		Str("severity", string(incident.Severity)).
		Msg("INCIDENT_LOGGED")
	return nil
}

func (ir *IncidentResponder) actionSendAlert(incident *SecurityIncident, params map[string]string) error {
	ir.logger.Info().
		Str("incident_id", incident.ID).
		Str("type", incident.Type).
		Interface("params", params).
		Msg("ALERT_SENT")
	return nil
}

func (ir *IncidentResponder) actionBlockIP(incident *SecurityIncident, params map[string]string) error {
	// In production, this would call firewall/WAF APIs
	duration := params["duration"]
	ir.logger.Warn().
		Str("incident_id", incident.ID).
		Strs("assets", incident.AffectedAssets).
		Str("duration", duration).
		Msg("IP_BLOCK_REQUESTED")
	return nil
}

//nolint:unparam // params kept for future key ID extraction
func (ir *IncidentResponder) actionRevokeKey(incident *SecurityIncident, _ map[string]string) error {
	// In production, this would call key management APIs
	ir.logger.Warn().
		Str("incident_id", incident.ID).
		Strs("assets", incident.AffectedAssets).
		Msg("KEY_REVOCATION_REQUESTED")
	return nil
}

//nolint:unparam // params kept for future account ID extraction
func (ir *IncidentResponder) actionSuspendAccount(incident *SecurityIncident, _ map[string]string) error {
	// In production, this would call account management APIs
	ir.logger.Warn().
		Str("incident_id", incident.ID).
		Strs("assets", incident.AffectedAssets).
		Msg("ACCOUNT_SUSPENSION_REQUESTED")
	return nil
}

//nolint:unparam // params kept for future provider ID extraction
func (ir *IncidentResponder) actionSuspendProvider(incident *SecurityIncident, _ map[string]string) error {
	// In production, this would call provider management APIs
	ir.logger.Warn().
		Str("incident_id", incident.ID).
		Strs("assets", incident.AffectedAssets).
		Msg("PROVIDER_SUSPENSION_REQUESTED")
	return nil
}

func (ir *IncidentResponder) actionIncreaseSeverity(incident *SecurityIncident) error {
	if incident.Severity < SeverityCritical {
		switch incident.Severity {
		case SeverityLow:
			incident.Severity = SeverityMedium
		case SeverityMedium:
			incident.Severity = SeverityHigh
		case SeverityHigh:
			incident.Severity = SeverityCritical
		}
		ir.logger.Info().
			Str("incident_id", incident.ID).
			Str("new_severity", string(incident.Severity)).
			Msg("SEVERITY_INCREASED")
	}
	return nil
}

func (ir *IncidentResponder) actionNotifyTeam(incident *SecurityIncident, params map[string]string) error {
	channel := params["channel"]
	priority := params["priority"]
	ir.logger.Info().
		Str("incident_id", incident.ID).
		Str("channel", channel).
		Str("priority", priority).
		Msg("TEAM_NOTIFIED")
	return nil
}

func (ir *IncidentResponder) actionCollectEvidence(incident *SecurityIncident) error {
	ir.logger.Info().
		Str("incident_id", incident.ID).
		Msg("EVIDENCE_COLLECTED")
	return nil
}

//nolint:unparam // params kept for future escalation configuration
func (ir *IncidentResponder) actionEscalate(incident *SecurityIncident, _ map[string]string) error {
	ir.logger.Warn().
		Str("incident_id", incident.ID).
		Str("severity", string(incident.Severity)).
		Msg("INCIDENT_ESCALATED")
	return nil
}

// GetPlaybook retrieves a playbook by ID
func (ir *IncidentResponder) GetPlaybook(id string) (*Playbook, bool) {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	playbook, exists := ir.playbooks[id]
	return playbook, exists
}

// GetAllPlaybooks returns all playbooks
func (ir *IncidentResponder) GetAllPlaybooks() []*Playbook {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	playbooks := make([]*Playbook, 0, len(ir.playbooks))
	for _, p := range ir.playbooks {
		playbooks = append(playbooks, p)
	}
	return playbooks
}

// AddPlaybook adds a new playbook
func (ir *IncidentResponder) AddPlaybook(playbook *Playbook) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.playbooks[playbook.ID] = playbook
	ir.logger.Info().Str("playbook_id", playbook.ID).Msg("playbook added")
}

// RemovePlaybook removes a playbook
func (ir *IncidentResponder) RemovePlaybook(id string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	delete(ir.playbooks, id)
	ir.logger.Info().Str("playbook_id", id).Msg("playbook removed")
}

// EnablePlaybook enables a playbook
func (ir *IncidentResponder) EnablePlaybook(id string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if playbook, exists := ir.playbooks[id]; exists {
		playbook.Enabled = true
	}
}

// DisablePlaybook disables a playbook
func (ir *IncidentResponder) DisablePlaybook(id string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if playbook, exists := ir.playbooks[id]; exists {
		playbook.Enabled = false
	}
}
