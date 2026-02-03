package provider_daemon

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnsibleVault(t *testing.T) {
	vault := NewAnsibleVault()

	require.NotNil(t, vault)
	assert.NotNil(t, vault.passwordCache)
	assert.Equal(t, VaultSecretID("default"), vault.defaultSecretID)
}

func TestAnsibleVaultSetPassword(t *testing.T) {
	vault := NewAnsibleVault()

	t.Run("set valid password", func(t *testing.T) {
		err := vault.SetPassword("test-id", "mysecretpassword")
		require.NoError(t, err)
	})

	t.Run("set empty password", func(t *testing.T) {
		err := vault.SetPassword("test-id", "")
		assert.ErrorIs(t, err, ErrVaultPasswordEmpty)
	})

	t.Run("set default password", func(t *testing.T) {
		err := vault.SetDefaultPassword("defaultpass")
		require.NoError(t, err)
	})
}

func TestAnsibleVaultClearPassword(t *testing.T) {
	vault := NewAnsibleVault()

	// Set a password
	err := vault.SetPassword("test-id", "mysecretpassword")
	require.NoError(t, err)

	// Clear the password
	vault.ClearPassword("test-id")

	// Verify it's cleared
	_, err = vault.getPassword("test-id")
	assert.ErrorIs(t, err, ErrVaultPasswordRequired)
}

func TestAnsibleVaultClearAllPasswords(t *testing.T) {
	vault := NewAnsibleVault()

	// Set multiple passwords
	_ = vault.SetPassword("id1", "pass1")
	_ = vault.SetPassword("id2", "pass2")
	_ = vault.SetPassword("id3", "pass3")

	// Clear all
	vault.ClearAllPasswords()

	// Verify all are cleared
	assert.Empty(t, vault.passwordCache)
}

func TestAnsibleVaultEncryptDecrypt(t *testing.T) {
	vault := NewAnsibleVault()
	password := "test-vault-password-123"
	err := vault.SetDefaultPassword(password)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"short text", "hello"},
		{"longer text", "This is a longer piece of text that needs to be encrypted"},
		{"special characters", "password!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode text", "日本語テスト 🔐"},
		{"empty string", ""},
		{"multiline", "line1\nline2\nline3"},
		{"json content", `{"key": "value", "secret": "data"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := vault.EncryptString(tc.plaintext)
			require.NoError(t, err)

			// Verify encrypted format
			assert.True(t, strings.HasPrefix(encrypted, VaultHeader))
			assert.Contains(t, encrypted, VaultCipherAES256)

			// Decrypt and verify
			decrypted, err := vault.DecryptString(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}

func TestAnsibleVaultEncryptWithSecretID(t *testing.T) {
	vault := NewAnsibleVault()
	secretID := VaultSecretID("production")

	err := vault.SetPassword(secretID, "production-password")
	require.NoError(t, err)

	plaintext := []byte("sensitive production data")
	encrypted, err := vault.Encrypt(plaintext, secretID)
	require.NoError(t, err)

	// Verify vault version 1.2 format with secret ID
	assert.Contains(t, encrypted, VaultVersion12)
	assert.Contains(t, encrypted, string(secretID))

	// Decrypt
	decrypted, err := vault.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAnsibleVaultDecryptWrongPassword(t *testing.T) {
	vault1 := NewAnsibleVault()
	vault2 := NewAnsibleVault()

	_ = vault1.SetDefaultPassword("password1")
	_ = vault2.SetDefaultPassword("password2")

	plaintext := "secret data"
	encrypted, err := vault1.EncryptString(plaintext)
	require.NoError(t, err)

	// Try to decrypt with wrong password
	_, err = vault2.DecryptString(encrypted)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "HMAC") || strings.Contains(err.Error(), "decrypt"))
}

func TestAnsibleVaultDecryptInvalidFormat(t *testing.T) {
	vault := NewAnsibleVault()
	_ = vault.SetDefaultPassword("password")

	testCases := []struct {
		name      string
		vaultText string
	}{
		{"empty string", ""},
		{"random text", "not vault encrypted"},
		{"missing payload", "$ANSIBLE_VAULT;1.1;AES256"},
		{"invalid header", "$INVALID_VAULT;1.1;AES256\nabcdef"},
		{"invalid version", "$ANSIBLE_VAULT;9.9;AES256\nabcdef"},
		{"invalid cipher", "$ANSIBLE_VAULT;1.1;DES\nabcdef"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := vault.DecryptString(tc.vaultText)
			assert.Error(t, err)
		})
	}
}

func TestAnsibleVaultDecryptNoPassword(t *testing.T) {
	vault := NewAnsibleVault()
	// Don't set password

	vaultText := "$ANSIBLE_VAULT;1.1;AES256\n" +
		"6162636465666768696a6b6c6d6e6f70"

	_, err := vault.DecryptString(vaultText)
	assert.ErrorIs(t, err, ErrVaultPasswordRequired)
}

func TestIsVaultEncrypted(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected bool
	}{
		{"valid vault v1.1", "$ANSIBLE_VAULT;1.1;AES256\ndata", true},
		{"valid vault v1.2", "$ANSIBLE_VAULT;1.2;AES256;prod\ndata", true},
		{"with whitespace", "  $ANSIBLE_VAULT;1.1;AES256\ndata  ", true},
		{"plain text", "just plain text", false},
		{"empty string", "", false},
		{"partial header", "$ANSIBLE", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsVaultEncrypted(tc.text)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseVaultText(t *testing.T) {
	t.Run("valid v1.1", func(t *testing.T) {
		vaultText := "$ANSIBLE_VAULT;1.1;AES256\n" +
			"6162636465666768"
		header, payload, err := parseVaultText(vaultText)
		require.NoError(t, err)
		assert.Equal(t, "$ANSIBLE_VAULT;1.1;AES256", header)
		assert.Equal(t, "6162636465666768", payload)
	})

	t.Run("valid v1.2 with secret id", func(t *testing.T) {
		vaultText := "$ANSIBLE_VAULT;1.2;AES256;production\n" +
			"6162636465666768"
		header, payload, err := parseVaultText(vaultText)
		require.NoError(t, err)
		assert.Equal(t, "$ANSIBLE_VAULT;1.2;AES256;production", header)
		assert.Equal(t, "6162636465666768", payload)
	})

	t.Run("multiline payload", func(t *testing.T) {
		vaultText := "$ANSIBLE_VAULT;1.1;AES256\n" +
			"61626364\n" +
			"65666768"
		header, payload, err := parseVaultText(vaultText)
		require.NoError(t, err)
		assert.Equal(t, "$ANSIBLE_VAULT;1.1;AES256", header)
		assert.Equal(t, "6162636465666768", payload)
	})
}

func TestExtractSecretID(t *testing.T) {
	testCases := []struct {
		name     string
		header   string
		expected VaultSecretID
	}{
		{"v1.1 no secret id", "$ANSIBLE_VAULT;1.1;AES256", ""},
		{"v1.2 with secret id", "$ANSIBLE_VAULT;1.2;AES256;production", "production"},
		{"empty secret id", "$ANSIBLE_VAULT;1.2;AES256;", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractSecretID(tc.header)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPKCS7PadUnpad(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte{0x01}},
		{"block size", []byte("1234567890123456")}, // 16 bytes = AES block size
		{"multiple blocks", []byte("This is test data that spans multiple AES blocks for testing")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			padded := pkcs7Pad(tc.data, 16)

			// Verify padding is applied
			assert.Equal(t, 0, len(padded)%16, "padded length should be multiple of block size")
			assert.GreaterOrEqual(t, len(padded), len(tc.data))

			// Unpad and verify
			unpadded, err := pkcs7Unpad(padded)
			require.NoError(t, err)
			assert.Equal(t, tc.data, unpadded)
		})
	}
}

func TestPKCS7UnpadInvalid(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"invalid padding byte", []byte{0x01, 0x02, 0x03, 0x00}}, // padding byte is 0
		{"padding larger than length", []byte{0x10}},             // padding is 16 but only 1 byte
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pkcs7Unpad(tc.data)
			assert.Error(t, err)
		})
	}
}

func TestClearBytes(t *testing.T) {
	data := []byte("sensitive data")
	original := make([]byte, len(data))
	copy(original, data)

	clearBytes(data)

	// Verify all bytes are zeroed
	for i, b := range data {
		assert.Equal(t, byte(0), b, "byte %d should be zero", i)
	}
}

func TestFormatVaultOutput(t *testing.T) {
	header := "$ANSIBLE_VAULT;1.1;AES256"
	// Create a long payload
	payload := strings.Repeat("a", 200)

	output := formatVaultOutput(header, payload)

	lines := strings.Split(output, "\n")
	assert.Equal(t, header, lines[0])

	// Verify line length (except last line)
	for i := 1; i < len(lines)-1; i++ {
		assert.LessOrEqual(t, len(lines[i]), vaultLineLength)
	}
}

// VaultVariables tests

func TestNewVaultVariables(t *testing.T) {
	vault := NewAnsibleVault()
	vars := NewVaultVariables(vault)

	require.NotNil(t, vars)
	assert.NotNil(t, vars.variables)
	assert.NotNil(t, vars.encrypted)
	assert.Equal(t, vault, vars.vault)
}

func TestVaultVariablesSetGet(t *testing.T) {
	vault := NewAnsibleVault()
	vars := NewVaultVariables(vault)

	vars.Set("key1", "value1")
	vars.Set("key2", 42)
	vars.Set("key3", true)

	val1, err := vars.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val1)

	val2, err := vars.Get("key2")
	require.NoError(t, err)
	assert.Equal(t, 42, val2)

	val3, err := vars.Get("key3")
	require.NoError(t, err)
	assert.Equal(t, true, val3)
}

func TestVaultVariablesSetEncrypted(t *testing.T) {
	vault := NewAnsibleVault()
	_ = vault.SetDefaultPassword("test-password")
	vars := NewVaultVariables(vault)

	err := vars.SetEncrypted("secret_key", "secret_value")
	require.NoError(t, err)

	// Check it's marked as encrypted
	assert.True(t, vars.IsEncrypted("secret_key"))

	// Get decrypted value
	val, err := vars.Get("secret_key")
	require.NoError(t, err)
	assert.Equal(t, "secret_value", val)

	// Get encrypted form
	encrypted, ok := vars.GetEncrypted("secret_key")
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(encrypted, VaultHeader))
}

func TestVaultVariablesGetNonExistent(t *testing.T) {
	vault := NewAnsibleVault()
	vars := NewVaultVariables(vault)

	val, err := vars.Get("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestVaultVariablesKeys(t *testing.T) {
	vault := NewAnsibleVault()
	vars := NewVaultVariables(vault)

	vars.Set("key1", "value1")
	vars.Set("key2", "value2")
	vars.Set("key3", "value3")

	keys := vars.Keys()
	assert.Len(t, keys, 3)

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}
	assert.True(t, keyMap["key1"])
	assert.True(t, keyMap["key2"])
	assert.True(t, keyMap["key3"])
}

func TestVaultVariablesToMap(t *testing.T) {
	vault := NewAnsibleVault()
	_ = vault.SetDefaultPassword("test-password")
	vars := NewVaultVariables(vault)

	vars.Set("plain", "plain_value")
	_ = vars.SetEncrypted("secret", "secret_value")

	m := vars.ToMap()
	assert.Len(t, m, 2)
	assert.Equal(t, "plain_value", m["plain"])
	// Encrypted value should still be encrypted
	assert.True(t, strings.HasPrefix(m["secret"].(string), VaultHeader))
}

func TestVaultVariablesDecryptAll(t *testing.T) {
	vault := NewAnsibleVault()
	_ = vault.SetDefaultPassword("test-password")
	vars := NewVaultVariables(vault)

	vars.Set("plain", "plain_value")
	_ = vars.SetEncrypted("secret1", "secret_value1")
	_ = vars.SetEncrypted("secret2", "secret_value2")

	decrypted, err := vars.DecryptAll()
	require.NoError(t, err)

	assert.Len(t, decrypted, 3)
	assert.Equal(t, "plain_value", decrypted["plain"])
	assert.Equal(t, "secret_value1", decrypted["secret1"])
	assert.Equal(t, "secret_value2", decrypted["secret2"])
}

func TestVaultVariablesIsEncrypted(t *testing.T) {
	vault := NewAnsibleVault()
	_ = vault.SetDefaultPassword("test-password")
	vars := NewVaultVariables(vault)

	vars.Set("plain", "plain_value")
	_ = vars.SetEncrypted("secret", "secret_value")

	assert.False(t, vars.IsEncrypted("plain"))
	assert.True(t, vars.IsEncrypted("secret"))
	assert.False(t, vars.IsEncrypted("nonexistent"))
}

func TestVaultVariablesGetEncrypted(t *testing.T) {
	vault := NewAnsibleVault()
	_ = vault.SetDefaultPassword("test-password")
	vars := NewVaultVariables(vault)

	vars.Set("plain", "plain_value")
	_ = vars.SetEncrypted("secret", "secret_value")

	// Plain variable
	_, ok := vars.GetEncrypted("plain")
	assert.False(t, ok)

	// Encrypted variable
	encrypted, ok := vars.GetEncrypted("secret")
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(encrypted, VaultHeader))

	// Non-existent variable
	_, ok = vars.GetEncrypted("nonexistent")
	assert.False(t, ok)
}

func TestAnsibleVaultRoundTrip(t *testing.T) {
	// Test a complete round-trip with realistic data
	vault := NewAnsibleVault()
	password := "my-strong-vault-password-!@#$%"
	err := vault.SetDefaultPassword(password)
	require.NoError(t, err)

	// Test with realistic secrets
	secrets := map[string]string{
		"db_password":     "postgres_super_secret_123!",
		"api_key":         "sk_live_abcdef1234567890",
		"ssh_private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpA...",
		"aws_access_key":  "AKIAIOSFODNN7EXAMPLE",
		"aws_secret_key":  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"certificate":     "-----BEGIN CERTIFICATE-----\nMIIC+z...",
		"service_account": `{"type":"service_account","project_id":"my-project"}`,
	}

	for name, secret := range secrets {
		t.Run(name, func(t *testing.T) {
			encrypted, err := vault.EncryptString(secret)
			require.NoError(t, err)

			// Verify it's encrypted
			assert.True(t, IsVaultEncrypted(encrypted))
			assert.NotContains(t, encrypted, secret)

			// Decrypt and verify
			decrypted, err := vault.DecryptString(encrypted)
			require.NoError(t, err)
			assert.Equal(t, secret, decrypted)
		})
	}
}

func TestAnsibleVaultWithPasswordBytes(t *testing.T) {
	vault := NewAnsibleVault()

	plaintext := []byte("test data for encryption")
	password := []byte("password-bytes-test")

	encrypted, err := vault.EncryptWithPassword(plaintext, password, "")
	require.NoError(t, err)

	_, _ = vault.DecryptWithPassword(strings.TrimPrefix(encrypted, "$ANSIBLE_VAULT;1.1;AES256\n"), password)
	// This won't work directly because DecryptWithPassword expects the payload only
	// Let's use the full decrypt path instead

	// Set the password and use normal decrypt
	_ = vault.SetDefaultPassword(string(password))
	decrypted2, err := vault.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}
