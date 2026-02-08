package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsValidMessageType(t *testing.T) {
	tests := []struct {
		name  string
		mtype MessageType
		want  bool
	}{
		{"handoff request", MessageTypeHandoffRequest, true},
		{"handoff response", MessageTypeHandoffResponse, true},
		{"needmore request", MessageTypeNeedMoreRequest, true},
		{"needmore response", MessageTypeNeedMoreResponse, true},
		{"invalid", MessageType("invalid"), false},
		{"empty", MessageType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsValidMessageType(tt.mtype))
		})
	}
}

func TestIsValidMessagePriority(t *testing.T) {
	tests := []struct {
		name     string
		priority MessagePriority
		want     bool
	}{
		{"low", MessagePriorityLow, true},
		{"normal", MessagePriorityNormal, true},
		{"high", MessagePriorityHigh, true},
		{"critical", MessagePriorityCritical, true},
		{"too low", MessagePriority(0), false},
		{"too high", MessagePriority(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsValidMessagePriority(tt.priority))
		})
	}
}

func TestAgentMessage_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(5 * time.Minute)

	validMsg := &AgentMessage{
		MessageID:  "msg-123",
		Type:       MessageTypeHandoffRequest,
		FromNodeID: "node-1",
		ToNodeID:   "node-2",
		ClusterID:  "cluster-1",
		Priority:   MessagePriorityNormal,
		CreatedAt:  now,
		ExpiresAt:  future,
	}

	tests := []struct {
		name    string
		msg     *AgentMessage
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid message",
			msg:     validMsg,
			wantErr: false,
		},
		{
			name: "missing message_id",
			msg: &AgentMessage{
				Type:       MessageTypeHandoffRequest,
				FromNodeID: "node-1",
				ClusterID:  "cluster-1",
				Priority:   MessagePriorityNormal,
				CreatedAt:  now,
				ExpiresAt:  future,
			},
			wantErr: true,
			errMsg:  "message_id required",
		},
		{
			name: "invalid type",
			msg: &AgentMessage{
				MessageID:  "msg-123",
				Type:       MessageType("invalid"),
				FromNodeID: "node-1",
				ClusterID:  "cluster-1",
				Priority:   MessagePriorityNormal,
				CreatedAt:  now,
				ExpiresAt:  future,
			},
			wantErr: true,
			errMsg:  "invalid message type",
		},
		{
			name: "missing from_node_id",
			msg: &AgentMessage{
				MessageID: "msg-123",
				Type:      MessageTypeHandoffRequest,
				ClusterID: "cluster-1",
				Priority:  MessagePriorityNormal,
				CreatedAt: now,
				ExpiresAt: future,
			},
			wantErr: true,
			errMsg:  "from_node_id required",
		},
		{
			name: "missing cluster_id",
			msg: &AgentMessage{
				MessageID:  "msg-123",
				Type:       MessageTypeHandoffRequest,
				FromNodeID: "node-1",
				Priority:   MessagePriorityNormal,
				CreatedAt:  now,
				ExpiresAt:  future,
			},
			wantErr: true,
			errMsg:  "cluster_id required",
		},
		{
			name: "invalid priority",
			msg: &AgentMessage{
				MessageID:  "msg-123",
				Type:       MessageTypeHandoffRequest,
				FromNodeID: "node-1",
				ClusterID:  "cluster-1",
				Priority:   MessagePriority(100),
				CreatedAt:  now,
				ExpiresAt:  future,
			},
			wantErr: true,
			errMsg:  "invalid priority",
		},
		{
			name: "expires before created",
			msg: &AgentMessage{
				MessageID:  "msg-123",
				Type:       MessageTypeHandoffRequest,
				FromNodeID: "node-1",
				ClusterID:  "cluster-1",
				Priority:   MessagePriorityNormal,
				CreatedAt:  future,
				ExpiresAt:  now,
			},
			wantErr: true,
			errMsg:  "expires_at must be after created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAgentMessage_IsExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Minute)
	future := time.Now().Add(1 * time.Minute)

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"expired", past, true},
		{"not expired", future, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &AgentMessage{ExpiresAt: tt.expiresAt}
			require.Equal(t, tt.want, msg.IsExpired())
		})
	}
}

func TestHandoffRequest_Validate(t *testing.T) {
	validReq := &HandoffRequest{
		TaskID:                  "task-123",
		JobID:                   "job-456",
		Summary:                 "Test task",
		Priority:                MessagePriorityNormal,
		Reason:                  "Agent overloaded",
		EstimatedRuntimeSeconds: 3600,
		RequiredCapabilities: AgentCapabilities{
			MinMemoryGB:                4,
			MinCPUCores:                2,
			SupportedContainerRuntimes: []string{"docker"},
			MaxTaskDurationSeconds:     7200,
		},
	}

	tests := []struct {
		name    string
		req     *HandoffRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			req:     validReq,
			wantErr: false,
		},
		{
			name: "missing task_id",
			req: &HandoffRequest{
				JobID:                   "job-456",
				Summary:                 "Test task",
				Priority:                MessagePriorityNormal,
				Reason:                  "Overloaded",
				EstimatedRuntimeSeconds: 3600,
				RequiredCapabilities:    validReq.RequiredCapabilities,
			},
			wantErr: true,
			errMsg:  "task_id required",
		},
		{
			name: "summary too long",
			req: &HandoffRequest{
				TaskID:                  "task-123",
				JobID:                   "job-456",
				Summary:                 string(make([]byte, 501)),
				Priority:                MessagePriorityNormal,
				Reason:                  "Overloaded",
				EstimatedRuntimeSeconds: 3600,
				RequiredCapabilities:    validReq.RequiredCapabilities,
			},
			wantErr: true,
			errMsg:  "summary exceeds 500 characters",
		},
		{
			name: "reason too long",
			req: &HandoffRequest{
				TaskID:                  "task-123",
				JobID:                   "job-456",
				Summary:                 "Test",
				Priority:                MessagePriorityNormal,
				Reason:                  string(make([]byte, 201)),
				EstimatedRuntimeSeconds: 3600,
				RequiredCapabilities:    validReq.RequiredCapabilities,
			},
			wantErr: true,
			errMsg:  "reason exceeds 200 characters",
		},
		{
			name: "invalid runtime",
			req: &HandoffRequest{
				TaskID:                  "task-123",
				JobID:                   "job-456",
				Summary:                 "Test",
				Priority:                MessagePriorityNormal,
				Reason:                  "Overloaded",
				EstimatedRuntimeSeconds: 0,
				RequiredCapabilities:    validReq.RequiredCapabilities,
			},
			wantErr: true,
			errMsg:  "estimated_runtime_seconds must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandoffResponse_Validate(t *testing.T) {
	validResp := &HandoffResponse{
		RequestMessageID: "msg-123",
		Accepted:         true,
		Reason:           "Capacity available",
	}

	tests := []struct {
		name    string
		resp    *HandoffResponse
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid accepted",
			resp:    validResp,
			wantErr: false,
		},
		{
			name: "valid rejected",
			resp: &HandoffResponse{
				RequestMessageID: "msg-123",
				Accepted:         false,
				Reason:           "Overloaded",
				RejectionCode:    RejectionCodeOverloaded,
			},
			wantErr: false,
		},
		{
			name: "missing request_message_id",
			resp: &HandoffResponse{
				Accepted: true,
				Reason:   "OK",
			},
			wantErr: true,
			errMsg:  "request_message_id required",
		},
		{
			name: "reason too long",
			resp: &HandoffResponse{
				RequestMessageID: "msg-123",
				Accepted:         true,
				Reason:           string(make([]byte, 501)),
			},
			wantErr: true,
			errMsg:  "reason exceeds 500 characters",
		},
		{
			name: "rejected without code",
			resp: &HandoffResponse{
				RequestMessageID: "msg-123",
				Accepted:         false,
				Reason:           "No",
			},
			wantErr: true,
			errMsg:  "rejection_code required when not accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNeedMoreRequest_Validate(t *testing.T) {
	validReq := &NeedMoreRequest{
		AvailableCapacity: NodeCapacity{
			CPUCoresTotal:     8,
			CPUCoresAvailable: 4,
			MemoryGBTotal:     16,
			MemoryGBAvailable: 8,
		},
		Capabilities: AgentCapabilities{
			MinMemoryGB:                4,
			MinCPUCores:                2,
			SupportedContainerRuntimes: []string{"docker"},
			MaxTaskDurationSeconds:     7200,
		},
		MaxTasks:          10,
		PreferredPriority: MessagePriorityNormal,
	}

	tests := []struct {
		name    string
		req     *NeedMoreRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			req:     validReq,
			wantErr: false,
		},
		{
			name: "max_tasks too high",
			req: &NeedMoreRequest{
				AvailableCapacity: validReq.AvailableCapacity,
				Capabilities:      validReq.Capabilities,
				MaxTasks:          101,
				PreferredPriority: MessagePriorityNormal,
			},
			wantErr: true,
			errMsg:  "max_tasks must be 1-100",
		},
		{
			name: "max_tasks zero",
			req: &NeedMoreRequest{
				AvailableCapacity: validReq.AvailableCapacity,
				Capabilities:      validReq.Capabilities,
				MaxTasks:          0,
				PreferredPriority: MessagePriorityNormal,
			},
			wantErr: true,
			errMsg:  "max_tasks must be 1-100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNeedMoreResponse_Validate(t *testing.T) {
	validResp := &NeedMoreResponse{
		RequestMessageID: "msg-123",
		TaskIDs:          []string{"task-1", "task-2"},
	}

	tests := []struct {
		name    string
		resp    *NeedMoreResponse
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid with tasks",
			resp:    validResp,
			wantErr: false,
		},
		{
			name: "valid no tasks",
			resp: &NeedMoreResponse{
				RequestMessageID:  "msg-123",
				NoTasksAvailable:  true,
				RetryAfterSeconds: 60,
			},
			wantErr: false,
		},
		{
			name: "missing request_message_id",
			resp: &NeedMoreResponse{
				TaskIDs: []string{"task-1"},
			},
			wantErr: true,
			errMsg:  "request_message_id required",
		},
		{
			name: "no tasks but not flagged",
			resp: &NeedMoreResponse{
				RequestMessageID: "msg-123",
			},
			wantErr: true,
			errMsg:  "task_ids required when tasks available",
		},
		{
			name: "conflicting flags",
			resp: &NeedMoreResponse{
				RequestMessageID: "msg-123",
				TaskIDs:          []string{"task-1"},
				NoTasksAvailable: true,
			},
			wantErr: true,
			errMsg:  "cannot have both no_tasks_available and task_ids",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAgentCapabilities_Validate(t *testing.T) {
	valid := AgentCapabilities{
		MinMemoryGB:                4,
		MinCPUCores:                2,
		SupportedContainerRuntimes: []string{"docker"},
		MaxTaskDurationSeconds:     7200,
	}

	tests := []struct {
		name    string
		cap     AgentCapabilities
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid",
			cap:     valid,
			wantErr: false,
		},
		{
			name: "negative memory",
			cap: AgentCapabilities{
				MinMemoryGB:                -1,
				MinCPUCores:                2,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			wantErr: true,
			errMsg:  "min_memory_gb cannot be negative",
		},
		{
			name: "no runtimes",
			cap: AgentCapabilities{
				MinMemoryGB:            4,
				MinCPUCores:            2,
				MaxTaskDurationSeconds: 7200,
			},
			wantErr: true,
			errMsg:  "supported_container_runtimes required",
		},
		{
			name: "zero duration",
			cap: AgentCapabilities{
				MinMemoryGB:                4,
				MinCPUCores:                2,
				SupportedContainerRuntimes: []string{"docker"},
			},
			wantErr: true,
			errMsg:  "max_task_duration_seconds must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cap.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAgentCapabilities_Matches(t *testing.T) {
	tests := []struct {
		name     string
		have     AgentCapabilities
		required AgentCapabilities
		want     bool
	}{
		{
			name: "exact match",
			have: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				MinGPUs:                    1,
				GPUTypes:                   []string{"nvidia-a100"},
				SupportedContainerRuntimes: []string{"docker"},
				Features:                   []string{"cuda"},
				MaxTaskDurationSeconds:     7200,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				MinGPUs:                    1,
				GPUTypes:                   []string{"nvidia-a100"},
				SupportedContainerRuntimes: []string{"docker"},
				Features:                   []string{"cuda"},
				MaxTaskDurationSeconds:     7200,
			},
			want: true,
		},
		{
			name: "exceeds requirements",
			have: AgentCapabilities{
				MinMemoryGB:                16,
				MinCPUCores:                8,
				MinGPUs:                    2,
				SupportedContainerRuntimes: []string{"docker", "singularity"},
				MaxTaskDurationSeconds:     14400,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				MinGPUs:                    1,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			want: true,
		},
		{
			name: "insufficient memory",
			have: AgentCapabilities{
				MinMemoryGB:                4,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			want: false,
		},
		{
			name: "wrong GPU type",
			have: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				MinGPUs:                    1,
				GPUTypes:                   []string{"nvidia-v100"},
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				MinGPUs:                    1,
				GPUTypes:                   []string{"nvidia-a100"},
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			want: false,
		},
		{
			name: "missing feature",
			have: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				Features:                   []string{"avx2"},
				MaxTaskDurationSeconds:     7200,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				Features:                   []string{"cuda"},
				MaxTaskDurationSeconds:     7200,
			},
			want: false,
		},
		{
			name: "unsupported runtime",
			have: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"singularity"},
				MaxTaskDurationSeconds:     7200,
			},
			want: false,
		},
		{
			name: "insufficient duration",
			have: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     3600,
			},
			required: AgentCapabilities{
				MinMemoryGB:                8,
				MinCPUCores:                4,
				SupportedContainerRuntimes: []string{"docker"},
				MaxTaskDurationSeconds:     7200,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.have.Matches(tt.required))
		})
	}
}
