// Package main implements the VirtEngine HPC Node Agent.
//
// Message handling for inter-agent communication
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// MessageHandler handles incoming agent messages
type MessageHandler struct {
	agent     *Agent
	inbox     chan *hpctypes.AgentMessage
	outbox    chan *hpctypes.AgentMessage
	pending   map[string]*PendingMessage
	pendingMu sync.RWMutex
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// PendingMessage tracks messages awaiting responses
type PendingMessage struct {
	Message    *hpctypes.AgentMessage
	ResponseCh chan *hpctypes.AgentMessage
	Timeout    time.Time
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(agent *Agent) *MessageHandler {
	return &MessageHandler{
		agent:   agent,
		inbox:   make(chan *hpctypes.AgentMessage, 100),
		outbox:  make(chan *hpctypes.AgentMessage, 100),
		pending: make(map[string]*PendingMessage),
		stopCh:  make(chan struct{}),
	}
}

// Start begins processing messages
func (m *MessageHandler) Start(ctx context.Context) {
	m.wg.Add(2)
	go m.processInbox(ctx)
	go m.cleanupPending(ctx)
}

// Stop halts message processing
func (m *MessageHandler) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// SendHandoffRequest sends a handoff request to another agent
func (m *MessageHandler) SendHandoffRequest(ctx context.Context, targetNodeID string, req *hpctypes.HandoffRequest) (*hpctypes.HandoffResponse, error) {
	msgID := uuid.New().String()
	msg := &hpctypes.AgentMessage{
		MessageID:  msgID,
		Type:       hpctypes.MessageTypeHandoffRequest,
		FromNodeID: m.agent.config.NodeID,
		ToNodeID:   targetNodeID,
		ClusterID:  m.agent.config.ClusterID,
		Priority:   req.Priority,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}

	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Send and wait for response
	respCh := make(chan *hpctypes.AgentMessage, 1)
	m.trackPending(msgID, msg, respCh)
	defer m.removePending(msgID)

	if err := m.sendMessage(ctx, msg, req); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		if resp.Type != hpctypes.MessageTypeHandoffResponse {
			return nil, fmt.Errorf("unexpected response type: %s", resp.Type)
		}
		handoffResp, ok := resp.Payload.(*hpctypes.HandoffResponse)
		if !ok {
			return nil, fmt.Errorf("invalid response payload")
		}
		return handoffResp, nil
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("handoff request timed out")
	}
}

// SendNeedMoreRequest sends a request for more tasks
func (m *MessageHandler) SendNeedMoreRequest(ctx context.Context, req *hpctypes.NeedMoreRequest) (*hpctypes.NeedMoreResponse, error) {
	msgID := uuid.New().String()
	msg := &hpctypes.AgentMessage{
		MessageID:  msgID,
		Type:       hpctypes.MessageTypeNeedMoreRequest,
		FromNodeID: m.agent.config.NodeID,
		ToNodeID:   "", // Broadcast to provider daemon
		ClusterID:  m.agent.config.ClusterID,
		Priority:   req.PreferredPriority,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Minute),
	}

	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Send and wait for response
	respCh := make(chan *hpctypes.AgentMessage, 1)
	m.trackPending(msgID, msg, respCh)
	defer m.removePending(msgID)

	if err := m.sendMessage(ctx, msg, req); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		if resp.Type != hpctypes.MessageTypeNeedMoreResponse {
			return nil, fmt.Errorf("unexpected response type: %s", resp.Type)
		}
		needMoreResp, ok := resp.Payload.(*hpctypes.NeedMoreResponse)
		if !ok {
			return nil, fmt.Errorf("invalid response payload")
		}
		return needMoreResp, nil
	case <-time.After(1 * time.Minute):
		return nil, fmt.Errorf("needmore request timed out")
	}
}

// HandleIncomingMessage processes an incoming message
func (m *MessageHandler) HandleIncomingMessage(msg *hpctypes.AgentMessage) error {
	select {
	case m.inbox <- msg:
		return nil
	case <-m.stopCh:
		return fmt.Errorf("message handler stopped")
	default:
		return fmt.Errorf("inbox full")
	}
}

func (m *MessageHandler) processInbox(ctx context.Context) {
	defer m.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case msg := <-m.inbox:
			if err := m.handleMessage(msg); err != nil {
				fmt.Printf("[MESSAGE] Error handling message %s: %v\n", msg.MessageID, err)
			}
		}
	}
}

func (m *MessageHandler) handleMessage(msg *hpctypes.AgentMessage) error {
	fmt.Printf("[MESSAGE] Received %s from %s\n", msg.Type, msg.FromNodeID)

	// Check if this is a response to a pending message
	if m.isResponse(msg.Type) {
		return m.deliverResponse(msg)
	}

	// Handle requests
	switch msg.Type {
	case hpctypes.MessageTypeHandoffRequest:
		return m.handleHandoffRequest(msg)
	case hpctypes.MessageTypeNeedMoreRequest:
		return m.handleNeedMoreRequest(msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (m *MessageHandler) handleHandoffRequest(msg *hpctypes.AgentMessage) error {
	// Unmarshal payload
	var req hpctypes.HandoffRequest
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return fmt.Errorf("failed to unmarshal handoff request: %w", err)
	}

	if err := req.Validate(); err != nil {
		return m.sendRejection(msg, hpctypes.RejectionCodeIncompatible, fmt.Sprintf("Invalid request: %v", err))
	}

	// Check if agent can accept the handoff
	accepted, rejectionCode, reason := m.evaluateHandoff(&req)

	resp := &hpctypes.HandoffResponse{
		RequestMessageID: msg.MessageID,
		Accepted:         accepted,
		Reason:           reason,
	}

	if !accepted {
		resp.RejectionCode = rejectionCode
	} else {
		estimatedStart := time.Now().Add(30 * time.Second)
		resp.EstimatedStartTime = &estimatedStart
	}

	return m.sendResponse(msg, hpctypes.MessageTypeHandoffResponse, resp)
}

func (m *MessageHandler) evaluateHandoff(req *hpctypes.HandoffRequest) (accepted bool, code hpctypes.RejectionCode, reason string) {
	// Check if agent is healthy
	health, err := m.agent.metricsCollector.CollectHealth()
	if err != nil {
		return false, hpctypes.RejectionCodeUnhealthy, "Failed to collect health metrics"
	}

	if health.Status != "healthy" {
		return false, hpctypes.RejectionCodeUnhealthy, fmt.Sprintf("Agent is %s", health.Status)
	}

	// Check capacity
	capacity, err := m.agent.metricsCollector.CollectCapacity()
	if err != nil {
		return false, hpctypes.RejectionCodeUnhealthy, "Failed to collect capacity"
	}

	// Get agent capabilities
	agentCaps := hpctypes.AgentCapabilities{
		MinMemoryGB:                int32(capacity.MemoryGBAvailable),
		MinCPUCores:                int32(capacity.CPUCoresAvailable),
		MinGPUs:                    int32(capacity.GPUsAvailable),
		SupportedContainerRuntimes: []string{"docker", "singularity"},
		MaxTaskDurationSeconds:     86400, // 24 hours
	}

	if capacity.GPUType != "" {
		agentCaps.GPUTypes = []string{capacity.GPUType}
	}

	// Check capability match
	if !agentCaps.Matches(req.RequiredCapabilities) {
		return false, hpctypes.RejectionCodeIncompatible, "Insufficient capabilities"
	}

	// Check if overloaded
	if capacity.CPUCoresAvailable < 2 || capacity.MemoryGBAvailable < 4 {
		return false, hpctypes.RejectionCodeOverloaded, "Insufficient available resources"
	}

	// Check priority threshold
	if req.Priority < hpctypes.MessagePriorityNormal {
		jobs := m.agent.metricsCollector.CollectJobs()
		if jobs.RunningCount >= 5 {
			return false, hpctypes.RejectionCodeLowPriority, "Too many running jobs, only accepting high priority"
		}
	}

	return true, "", "Capacity available"
}

func (m *MessageHandler) handleNeedMoreRequest(msg *hpctypes.AgentMessage) error {
	// Only provider daemon should send this, but agents can request from daemon
	return fmt.Errorf("needmore request should be sent to provider daemon")
}

func (m *MessageHandler) sendRejection(msg *hpctypes.AgentMessage, code hpctypes.RejectionCode, reason string) error {
	resp := &hpctypes.HandoffResponse{
		RequestMessageID: msg.MessageID,
		Accepted:         false,
		RejectionCode:    code,
		Reason:           reason,
	}
	return m.sendResponse(msg, hpctypes.MessageTypeHandoffResponse, resp)
}

func (m *MessageHandler) sendResponse(reqMsg *hpctypes.AgentMessage, respType hpctypes.MessageType, payload interface{}) error {
	resp := &hpctypes.AgentMessage{
		MessageID:  uuid.New().String(),
		Type:       respType,
		FromNodeID: m.agent.config.NodeID,
		ToNodeID:   reqMsg.FromNodeID,
		ClusterID:  m.agent.config.ClusterID,
		Priority:   reqMsg.Priority,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.sendMessage(ctx, resp, payload)
}

func (m *MessageHandler) sendMessage(ctx context.Context, msg *hpctypes.AgentMessage, payload interface{}) error {
	// Construct message envelope
	envelope := map[string]interface{}{
		"message_id":   msg.MessageID,
		"type":         msg.Type,
		"from_node_id": msg.FromNodeID,
		"to_node_id":   msg.ToNodeID,
		"cluster_id":   msg.ClusterID,
		"priority":     msg.Priority,
		"created_at":   msg.CreatedAt.Format(time.RFC3339),
		"expires_at":   msg.ExpiresAt.Format(time.RFC3339),
		"payload":      payload,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send via provider daemon API
	url := fmt.Sprintf("%s/api/v1/hpc/messages", m.agent.config.ProviderDaemonURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.agent.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

func (m *MessageHandler) isResponse(msgType hpctypes.MessageType) bool {
	return msgType == hpctypes.MessageTypeHandoffResponse || msgType == hpctypes.MessageTypeNeedMoreResponse
}

func (m *MessageHandler) deliverResponse(msg *hpctypes.AgentMessage) error {
	// Find the pending request this is responding to
	var reqMsgID string

	// Unmarshal payload to get request_message_id
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	switch msg.Type {
	case hpctypes.MessageTypeHandoffResponse:
		var resp hpctypes.HandoffResponse
		if err := json.Unmarshal(payloadBytes, &resp); err == nil {
			reqMsgID = resp.RequestMessageID
		}
	case hpctypes.MessageTypeNeedMoreResponse:
		var resp hpctypes.NeedMoreResponse
		if err := json.Unmarshal(payloadBytes, &resp); err == nil {
			reqMsgID = resp.RequestMessageID
		}
	}

	if reqMsgID == "" {
		return fmt.Errorf("response missing request_message_id")
	}

	m.pendingMu.RLock()
	pending, exists := m.pending[reqMsgID]
	m.pendingMu.RUnlock()

	if !exists {
		fmt.Printf("[MESSAGE] Received response for unknown request: %s\n", reqMsgID)
		return nil
	}

	// Unmarshal the full payload
	msg.Payload, err = m.unmarshalPayload(msg.Type, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	select {
	case pending.ResponseCh <- msg:
		fmt.Printf("[MESSAGE] Delivered response for request %s\n", reqMsgID)
	default:
		fmt.Printf("[MESSAGE] Response channel full for request %s\n", reqMsgID)
	}

	return nil
}

func (m *MessageHandler) unmarshalPayload(msgType hpctypes.MessageType, data []byte) (interface{}, error) {
	switch msgType {
	case hpctypes.MessageTypeHandoffResponse:
		var resp hpctypes.HandoffResponse
		err := json.Unmarshal(data, &resp)
		return &resp, err
	case hpctypes.MessageTypeNeedMoreResponse:
		var resp hpctypes.NeedMoreResponse
		err := json.Unmarshal(data, &resp)
		return &resp, err
	default:
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}
}

func (m *MessageHandler) trackPending(msgID string, msg *hpctypes.AgentMessage, responseCh chan *hpctypes.AgentMessage) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()

	m.pending[msgID] = &PendingMessage{
		Message:    msg,
		ResponseCh: responseCh,
		Timeout:    msg.ExpiresAt,
	}
}

func (m *MessageHandler) removePending(msgID string) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	delete(m.pending, msgID)
}

func (m *MessageHandler) cleanupPending(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.cleanupExpired()
		}
	}
}

func (m *MessageHandler) cleanupExpired() {
	now := time.Now()

	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()

	for msgID, pending := range m.pending {
		if now.After(pending.Timeout) {
			fmt.Printf("[MESSAGE] Timeout for request %s\n", msgID)
			close(pending.ResponseCh)
			delete(m.pending, msgID)
		}
	}
}
