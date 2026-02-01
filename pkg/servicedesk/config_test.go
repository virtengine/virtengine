package servicedesk

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected Enabled = false by default")
	}

	if cfg.MappingSchema == nil {
		t.Error("expected MappingSchema to be set")
	}

	if cfg.SyncConfig.SyncInterval != 30*time.Second {
		t.Errorf("expected SyncInterval = 30s, got %v", cfg.SyncConfig.SyncInterval)
	}

	if cfg.RetryConfig.MaxRetries != 5 {
		t.Errorf("expected MaxRetries = 5, got %d", cfg.RetryConfig.MaxRetries)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "disabled config is valid",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "enabled without service desk",
			cfg: &Config{
				Enabled:     true,
				SyncConfig:  DefaultSyncConfig(),
				RetryConfig: DefaultRetryConfig(),
			},
			wantErr: true,
		},
		{
			name: "enabled with jira",
			cfg: &Config{
				Enabled: true,
				JiraConfig: &JiraConfig{
					BaseURL:    "https://jira.example.com",
					Username:   "user",
					APIToken:   "token",
					ProjectKey: "PROJ",
				},
				SyncConfig:  DefaultSyncConfig(),
				RetryConfig: DefaultRetryConfig(),
			},
			wantErr: false,
		},
		{
			name: "enabled with waldur",
			cfg: &Config{
				Enabled: true,
				WaldurConfig: &WaldurConfig{
					BaseURL:          "https://waldur.example.com",
					Token:            "token",
					OrganizationUUID: "org-uuid",
				},
				SyncConfig:  DefaultSyncConfig(),
				RetryConfig: DefaultRetryConfig(),
			},
			wantErr: false,
		},
		{
			name: "jira missing base url",
			cfg: &Config{
				Enabled: true,
				JiraConfig: &JiraConfig{
					Username:   "user",
					APIToken:   "token",
					ProjectKey: "PROJ",
				},
				SyncConfig:  DefaultSyncConfig(),
				RetryConfig: DefaultRetryConfig(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJiraConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *JiraConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &JiraConfig{
				BaseURL:    "https://jira.example.com",
				Username:   "user",
				APIToken:   "token",
				ProjectKey: "PROJ",
			},
			wantErr: false,
		},
		{
			name: "missing base url",
			cfg: &JiraConfig{
				Username:   "user",
				APIToken:   "token",
				ProjectKey: "PROJ",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			cfg: &JiraConfig{
				BaseURL:    "https://jira.example.com",
				APIToken:   "token",
				ProjectKey: "PROJ",
			},
			wantErr: true,
		},
		{
			name: "missing api token",
			cfg: &JiraConfig{
				BaseURL:    "https://jira.example.com",
				Username:   "user",
				ProjectKey: "PROJ",
			},
			wantErr: true,
		},
		{
			name: "missing project key",
			cfg: &JiraConfig{
				BaseURL:  "https://jira.example.com",
				Username: "user",
				APIToken: "token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWaldurConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *WaldurConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &WaldurConfig{
				BaseURL:          "https://waldur.example.com",
				Token:            "token",
				OrganizationUUID: "org-uuid",
			},
			wantErr: false,
		},
		{
			name: "missing base url",
			cfg: &WaldurConfig{
				Token:            "token",
				OrganizationUUID: "org-uuid",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			cfg: &WaldurConfig{
				BaseURL:          "https://waldur.example.com",
				OrganizationUUID: "org-uuid",
			},
			wantErr: true,
		},
		{
			name: "missing org uuid",
			cfg: &WaldurConfig{
				BaseURL: "https://waldur.example.com",
				Token:   "token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyncConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SyncConfig
		wantErr bool
	}{
		{
			name:    "default config is valid",
			cfg:     DefaultSyncConfig(),
			wantErr: false,
		},
		{
			name: "sync interval too short",
			cfg: SyncConfig{
				SyncInterval: 1 * time.Second,
				BatchSize:    50,
			},
			wantErr: true,
		},
		{
			name: "batch size too small",
			cfg: SyncConfig{
				SyncInterval: 30 * time.Second,
				BatchSize:    0,
			},
			wantErr: true,
		},
		{
			name: "batch size too large",
			cfg: SyncConfig{
				SyncInterval: 30 * time.Second,
				BatchSize:    1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRetryConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RetryConfig
		wantErr bool
	}{
		{
			name:    "default config is valid",
			cfg:     DefaultRetryConfig(),
			wantErr: false,
		},
		{
			name: "negative max retries",
			cfg: RetryConfig{
				MaxRetries:        -1,
				InitialBackoff:    1 * time.Second,
				MaxBackoff:        1 * time.Minute,
				BackoffMultiplier: 2.0,
			},
			wantErr: true,
		},
		{
			name: "max backoff less than initial",
			cfg: RetryConfig{
				MaxRetries:        5,
				InitialBackoff:    1 * time.Minute,
				MaxBackoff:        1 * time.Second,
				BackoffMultiplier: 2.0,
			},
			wantErr: true,
		},
		{
			name: "multiplier less than 1",
			cfg: RetryConfig{
				MaxRetries:        5,
				InitialBackoff:    1 * time.Second,
				MaxBackoff:        1 * time.Minute,
				BackoffMultiplier: 0.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	cfg := RetryConfig{
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        1 * time.Minute,
		BackoffMultiplier: 2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 32 * time.Second},
		{6, 1 * time.Minute}, // capped at max
		{10, 1 * time.Minute},
	}

	for _, tt := range tests {
		backoff := cfg.CalculateBackoff(tt.attempt)
		if backoff != tt.expected {
			t.Errorf("CalculateBackoff(%d) = %v, want %v", tt.attempt, backoff, tt.expected)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	cfg := DefaultRetryConfig()

	tests := []struct {
		statusCode int
		retryable  bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},  // Rate limited
		{500, true},  // Server error
		{502, true},  // Bad gateway
		{503, true},  // Service unavailable
		{504, true},  // Gateway timeout
	}

	for _, tt := range tests {
		if cfg.IsRetryable(tt.statusCode) != tt.retryable {
			t.Errorf("IsRetryable(%d) = %v, want %v", tt.statusCode, cfg.IsRetryable(tt.statusCode), tt.retryable)
		}
	}
}

