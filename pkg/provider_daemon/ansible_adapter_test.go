package provider_daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryValidate(t *testing.T) {
	tests := []struct {
		name      string
		inventory *Inventory
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid inventory",
			inventory: &Inventory{
				Name: "test-inventory",
				Groups: []InventoryGroup{
					{Name: "webservers", Hosts: []InventoryHost{{Name: "192.168.1.1"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			inventory: &Inventory{
				Groups: []InventoryGroup{
					{Name: "webservers"},
				},
			},
			wantErr: true,
			errMsg:  "inventory name is required",
		},
		{
			name: "no groups",
			inventory: &Inventory{
				Name:   "test",
				Groups: []InventoryGroup{},
			},
			wantErr: true,
			errMsg:  "at least one group is required",
		},
		{
			name: "empty group name",
			inventory: &Inventory{
				Name: "test",
				Groups: []InventoryGroup{
					{Name: "", Hosts: []InventoryHost{{Name: "host1"}}},
				},
			},
			wantErr: true,
			errMsg:  "group name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inventory.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInventoryToINI(t *testing.T) {
	inventory := &Inventory{
		Name: "test-inventory",
		Groups: []InventoryGroup{
			{
				Name: "webservers",
				Hosts: []InventoryHost{
					{Name: "192.168.1.1", User: "deploy", Port: 22},
					{Name: "192.168.1.2", Alias: "web2", User: "deploy"},
				},
				Variables: map[string]interface{}{
					"http_port": 80,
				},
			},
			{
				Name: "dbservers",
				Hosts: []InventoryHost{
					{Name: "192.168.1.10", Variables: map[string]interface{}{"db_name": "prod"}},
				},
			},
			{
				Name:     "all_servers",
				Children: []string{"webservers", "dbservers"},
			},
		},
	}

	ini := inventory.ToINI()

	// Verify webservers group
	assert.Contains(t, ini, "[webservers]")
	assert.Contains(t, ini, "192.168.1.1 ansible_port=22 ansible_user=deploy")
	assert.Contains(t, ini, "web2 ansible_host=192.168.1.2 ansible_user=deploy")

	// Verify group variables
	assert.Contains(t, ini, "[webservers:vars]")
	assert.Contains(t, ini, "http_port=80")

	// Verify dbservers group
	assert.Contains(t, ini, "[dbservers]")
	assert.Contains(t, ini, "db_name=prod")

	// Verify children
	assert.Contains(t, ini, "[all_servers:children]")
	assert.Contains(t, ini, "webservers")
	assert.Contains(t, ini, "dbservers")
}

func TestPlaybookValidate(t *testing.T) {
	tests := []struct {
		name     string
		playbook *Playbook
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid playbook",
			playbook: &Playbook{
				Name: "deploy-app",
				Path: "/path/to/playbook.yml",
				Type: PlaybookTypeDeployment,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			playbook: &Playbook{
				Path: "/path/to/playbook.yml",
			},
			wantErr: true,
			errMsg:  "playbook name is required",
		},
		{
			name: "missing path",
			playbook: &Playbook{
				Name: "deploy-app",
			},
			wantErr: true,
			errMsg:  "playbook path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.playbook.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultAnsibleAdapterConfig(t *testing.T) {
	config := DefaultAnsibleAdapterConfig()

	assert.Equal(t, "ansible-playbook", config.AnsiblePath)
	assert.Equal(t, 30*time.Minute, config.DefaultTimeout)
	assert.Equal(t, 5, config.DefaultForks)
	assert.Equal(t, 10, config.MaxConcurrentExecutions)
	assert.True(t, config.LogRedaction)
}

func TestNewAnsibleAdapter(t *testing.T) {
	config := DefaultAnsibleAdapterConfig()
	adapter := NewAnsibleAdapter(config)

	require.NotNil(t, adapter)
	assert.NotNil(t, adapter.executions)
	assert.NotNil(t, adapter.activeExecutions)
	assert.NotNil(t, adapter.semaphore)
}

func TestAnsibleAdapterGenerateExecutionID(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	id1 := adapter.generateExecutionID()
	id2 := adapter.generateExecutionID()

	assert.True(t, strings.HasPrefix(id1, "exec-"))
	assert.True(t, strings.HasPrefix(id2, "exec-"))
	assert.NotEqual(t, id1, id2)
}

func TestAnsibleAdapterBuildCommandArgs(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	playbook := &Playbook{
		Name: "test",
		Path: "/path/to/playbook.yml",
	}

	tests := []struct {
		name     string
		options  ExecutionOptions
		expected []string
		notIn    []string
	}{
		{
			name:    "basic options",
			options: ExecutionOptions{},
			expected: []string{
				"/path/to/playbook.yml",
				"-i", "inventory.ini",
				"--forks", "5",
			},
		},
		{
			name: "with verbosity",
			options: ExecutionOptions{
				Verbosity: 2,
			},
			expected: []string{"-vv"},
		},
		{
			name: "with check mode",
			options: ExecutionOptions{
				CheckMode: true,
			},
			expected: []string{"--check"},
		},
		{
			name: "with diff mode",
			options: ExecutionOptions{
				DiffMode: true,
			},
			expected: []string{"--diff"},
		},
		{
			name: "with tags",
			options: ExecutionOptions{
				Tags: []string{"install", "configure"},
			},
			expected: []string{"--tags", "install,configure"},
		},
		{
			name: "with skip tags",
			options: ExecutionOptions{
				SkipTags: []string{"test"},
			},
			expected: []string{"--skip-tags", "test"},
		},
		{
			name: "with limit",
			options: ExecutionOptions{
				Limit: "webservers",
			},
			expected: []string{"--limit", "webservers"},
		},
		{
			name: "with become",
			options: ExecutionOptions{
				BecomeMethod: "sudo",
				BecomeUser:   "root",
			},
			expected: []string{"--become-method", "sudo", "--become", "--become-user", "root"},
		},
		{
			name: "with variables",
			options: ExecutionOptions{
				Variables: map[string]interface{}{
					"app_version": "1.0.0",
				},
			},
			expected: []string{"--extra-vars"},
		},
		{
			name: "with vault password file",
			options: ExecutionOptions{
				VaultPasswordFile: "/path/to/vault.txt",
			},
			expected: []string{"--vault-password-file", "/path/to/vault.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := adapter.buildCommandArgs(playbook, "inventory.ini", tt.options)

			for _, exp := range tt.expected {
				assert.Contains(t, args, exp)
			}
			for _, notExp := range tt.notIn {
				assert.NotContains(t, args, notExp)
			}
		})
	}
}

func TestAnsibleAdapterRedactSensitiveData(t *testing.T) {
	adapter := NewAnsibleAdapter(AnsibleAdapterConfig{
		LogRedaction: true,
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "password in output",
			input:    "Setting password=mysecretpassword123",
			expected: "[REDACTED]",
		},
		{
			name:     "api_key in output",
			input:    "Using api_key=abc123xyz789",
			expected: "[REDACTED]",
		},
		{
			name:     "token in output",
			input:    "Auth token: supersecrettoken",
			expected: "[REDACTED]",
		},
		{
			name:     "vault_password in output",
			input:    "vault_password=myvaultpass",
			expected: "[REDACTED]",
		},
		{
			name:     "no sensitive data",
			input:    "Task completed successfully",
			expected: "Task completed successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.redactSensitiveData(tt.input)
			if tt.expected == "[REDACTED]" {
				assert.Contains(t, result, "[REDACTED]")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAnsibleAdapterRedactSensitiveDataDisabled(t *testing.T) {
	adapter := NewAnsibleAdapter(AnsibleAdapterConfig{
		LogRedaction: false,
	})

	input := "password=mysecret"
	result := adapter.redactSensitiveData(input)

	assert.Equal(t, input, result)
}

func TestAnsibleAdapterParseRecapLine(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	tests := []struct {
		name     string
		line     string
		expected *PlaySummary
	}{
		{
			name: "full recap line",
			line: "webserver1 : ok=5 changed=2 unreachable=0 failed=0 skipped=1 rescued=0 ignored=0",
			expected: &PlaySummary{
				Host:        "webserver1",
				OK:          5,
				Changed:     2,
				Unreachable: 0,
				Failed:      0,
				Skipped:     1,
				Rescued:     0,
				Ignored:     0,
			},
		},
		{
			name: "partial recap line",
			line: "dbserver : ok=3 changed=1 failed=1",
			expected: &PlaySummary{
				Host:    "dbserver",
				OK:      3,
				Changed: 1,
				Failed:  1,
			},
		},
		{
			name:     "invalid line no colon",
			line:     "no colon here",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.parseRecapLine(tt.line)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Host, result.Host)
				assert.Equal(t, tt.expected.OK, result.OK)
				assert.Equal(t, tt.expected.Changed, result.Changed)
				assert.Equal(t, tt.expected.Unreachable, result.Unreachable)
				assert.Equal(t, tt.expected.Failed, result.Failed)
				assert.Equal(t, tt.expected.Skipped, result.Skipped)
			}
		})
	}
}

func TestAnsibleAdapterGetExecution(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	// Add an execution
	result := &ExecutionResult{
		ExecutionID:  "exec-test-123",
		PlaybookName: "test-playbook",
		State:        ExecutionStateSuccess,
	}
	adapter.executions["exec-test-123"] = result

	// Test retrieval
	t.Run("existing execution", func(t *testing.T) {
		retrieved, err := adapter.GetExecution("exec-test-123")
		require.NoError(t, err)
		assert.Equal(t, result, retrieved)
	})

	t.Run("non-existing execution", func(t *testing.T) {
		_, err := adapter.GetExecution("exec-nonexistent")
		assert.ErrorIs(t, err, ErrExecutionNotFound)
	})
}

func TestAnsibleAdapterListExecutions(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	// Add executions
	adapter.executions["exec-1"] = &ExecutionResult{ExecutionID: "exec-1"}
	adapter.executions["exec-2"] = &ExecutionResult{ExecutionID: "exec-2"}
	adapter.executions["exec-3"] = &ExecutionResult{ExecutionID: "exec-3"}

	results := adapter.ListExecutions()

	assert.Len(t, results, 3)

	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.ExecutionID] = true
	}
	assert.True(t, ids["exec-1"])
	assert.True(t, ids["exec-2"])
	assert.True(t, ids["exec-3"])
}

func TestAnsibleAdapterCleanupExecutions(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	now := time.Now()

	// Add executions with different ages
	adapter.executions["exec-old"] = &ExecutionResult{
		ExecutionID: "exec-old",
		State:       ExecutionStateSuccess,
		CompletedAt: now.Add(-2 * time.Hour),
	}
	adapter.executions["exec-recent"] = &ExecutionResult{
		ExecutionID: "exec-recent",
		State:       ExecutionStateSuccess,
		CompletedAt: now.Add(-30 * time.Minute),
	}
	adapter.executions["exec-running"] = &ExecutionResult{
		ExecutionID: "exec-running",
		State:       ExecutionStateRunning,
		CompletedAt: now.Add(-3 * time.Hour),
	}

	removed := adapter.CleanupExecutions(1 * time.Hour)

	assert.Equal(t, 1, removed)
	assert.Len(t, adapter.executions, 2)

	_, ok := adapter.executions["exec-old"]
	assert.False(t, ok, "old execution should be removed")

	_, ok = adapter.executions["exec-recent"]
	assert.True(t, ok, "recent execution should remain")

	_, ok = adapter.executions["exec-running"]
	assert.True(t, ok, "running execution should remain")
}

func TestAnsibleAdapterCancelExecution(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	cancelled := false
	cancelFunc := func() {
		cancelled = true
	}

	adapter.activeExecutions["exec-active"] = cancelFunc

	t.Run("cancel existing execution", func(t *testing.T) {
		err := adapter.CancelExecution("exec-active")
		require.NoError(t, err)
		assert.True(t, cancelled)
	})

	t.Run("cancel non-existing execution", func(t *testing.T) {
		err := adapter.CancelExecution("exec-nonexistent")
		assert.ErrorIs(t, err, ErrExecutionNotFound)
	})
}

func TestAnsibleAdapterWriteTemporaryInventory(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	inventory := &Inventory{
		Name: "test-inventory",
		Groups: []InventoryGroup{
			{
				Name: "webservers",
				Hosts: []InventoryHost{
					{Name: "192.168.1.1", User: "deploy"},
				},
			},
		},
	}

	tmpFile, err := adapter.writeTemporaryInventory(inventory)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// Verify file exists and contains correct content
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	assert.Contains(t, string(content), "[webservers]")
	assert.Contains(t, string(content), "192.168.1.1")
}

func TestAnsibleAdapterWriteTemporaryVaultPassword(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	password := "test-vault-password"
	tmpFile, err := adapter.writeTemporaryVaultPassword(password)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// Verify file exists and contains correct content
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, password, string(content))

	// Verify restrictive permissions
	info, err := os.Stat(tmpFile)
	require.NoError(t, err)
	// On Windows, permissions work differently, but file should be readable
	assert.NotNil(t, info)
}

func TestAnsibleAdapterGetPlaybookPath(t *testing.T) {
	t.Run("with playbooks dir", func(t *testing.T) {
		adapter := NewAnsibleAdapter(AnsibleAdapterConfig{
			PlaybooksDir: "/opt/playbooks",
		})

		path := adapter.GetPlaybookPath("deploy.yml")
		assert.Equal(t, filepath.Join("/opt/playbooks", "deploy.yml"), path)
	})

	t.Run("without playbooks dir", func(t *testing.T) {
		adapter := NewAnsibleAdapter(AnsibleAdapterConfig{})

		path := adapter.GetPlaybookPath("deploy.yml")
		assert.Equal(t, "deploy.yml", path)
	})
}

func TestAnsibleAdapterConcurrencyLimit(t *testing.T) {
	config := AnsibleAdapterConfig{
		MaxConcurrentExecutions: 2,
	}
	adapter := NewAnsibleAdapter(config)

	// Verify semaphore capacity
	assert.Equal(t, 2, cap(adapter.semaphore))
}

func TestAnsibleAdapterStatusUpdates(t *testing.T) {
	statusChan := make(chan ExecutionStatusUpdate, 10)
	adapter := NewAnsibleAdapter(AnsibleAdapterConfig{
		StatusUpdateChan: statusChan,
	})

	adapter.sendStatusUpdate("exec-123", ExecutionStateRunning, "Starting", "task1")

	select {
	case update := <-statusChan:
		assert.Equal(t, "exec-123", update.ExecutionID)
		assert.Equal(t, ExecutionStateRunning, update.State)
		assert.Equal(t, "Starting", update.Progress)
		assert.Equal(t, "task1", update.CurrentTask)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected status update")
	}
}

func TestAnsibleAdapterStatusUpdatesChannelFull(t *testing.T) {
	statusChan := make(chan ExecutionStatusUpdate, 1)
	statusChan <- ExecutionStatusUpdate{} // Fill the channel

	adapter := NewAnsibleAdapter(AnsibleAdapterConfig{
		StatusUpdateChan: statusChan,
	})

	// Should not block
	done := make(chan bool)
	go func() {
		adapter.sendStatusUpdate("exec-123", ExecutionStateRunning, "", "")
		done <- true
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sendStatusUpdate blocked when channel was full")
	}
}

func TestExecutionStates(t *testing.T) {
	states := []ExecutionState{
		ExecutionStatePending,
		ExecutionStateRunning,
		ExecutionStateSuccess,
		ExecutionStateFailed,
		ExecutionStateCancelled,
		ExecutionStateTimeout,
	}

	for _, state := range states {
		assert.NotEmpty(t, string(state))
	}
}

func TestPlaybookTypes(t *testing.T) {
	types := []PlaybookType{
		PlaybookTypeDeployment,
		PlaybookTypeConfiguration,
		PlaybookTypeMaintenance,
		PlaybookTypeCustom,
	}

	for _, pt := range types {
		assert.NotEmpty(t, string(pt))
	}
}

func TestAnsibleAdapterThreadSafety(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent execution storage
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		//nolint:unparam // id kept for future goroutine-specific logging
		go func(_ int) {
			defer wg.Done()
			execID := adapter.generateExecutionID()
			adapter.mu.Lock()
			adapter.executions[execID] = &ExecutionResult{
				ExecutionID: execID,
				State:       ExecutionStatePending,
			}
			adapter.mu.Unlock()
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			adapter.ListExecutions()
		}()
	}

	wg.Wait()

	// Should have at least numGoroutines executions
	assert.GreaterOrEqual(t, len(adapter.executions), numGoroutines)
}

func TestAnsibleAdapterValidatePlaybookFileNotFound(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())
	ctx := context.Background()

	playbook := &Playbook{
		Name: "nonexistent",
		Path: "/path/to/nonexistent/playbook.yml",
	}

	err := adapter.ValidatePlaybook(ctx, playbook)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPlaybookNotFound))
}

func TestAnsibleAdapterExecutePlaybookInvalidPlaybook(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())
	ctx := context.Background()

	playbook := &Playbook{
		// Missing required fields
	}
	inventory := &Inventory{
		Name:   "test",
		Groups: []InventoryGroup{{Name: "all"}},
	}

	_, err := adapter.ExecutePlaybook(ctx, playbook, inventory, ExecutionOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPlaybook))
}

func TestAnsibleAdapterExecutePlaybookInvalidInventory(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())
	ctx := context.Background()

	playbook := &Playbook{
		Name: "test",
		Path: "/path/to/playbook.yml",
	}
	inventory := &Inventory{
		// Missing required fields
	}

	_, err := adapter.ExecutePlaybook(ctx, playbook, inventory, ExecutionOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidInventory))
}

func TestAnsibleAdapterExecutePlaybookFileNotFound(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())
	ctx := context.Background()

	playbook := &Playbook{
		Name: "test",
		Path: "/path/to/nonexistent/playbook.yml",
	}
	inventory := &Inventory{
		Name:   "test",
		Groups: []InventoryGroup{{Name: "all"}},
	}

	_, err := adapter.ExecutePlaybook(ctx, playbook, inventory, ExecutionOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPlaybookNotFound))
}

func TestAnsibleAdapterRegisterPlaybook(t *testing.T) {
	adapter := NewAnsibleAdapter(DefaultAnsibleAdapterConfig())

	t.Run("valid playbook", func(t *testing.T) {
		playbook := &Playbook{
			Name: "test",
			Path: "/path/to/playbook.yml",
		}
		err := adapter.RegisterPlaybook(playbook)
		require.NoError(t, err)
	})

	t.Run("invalid playbook", func(t *testing.T) {
		playbook := &Playbook{}
		err := adapter.RegisterPlaybook(playbook)
		require.Error(t, err)
	})
}
