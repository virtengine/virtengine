package provider_daemon

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/data_vault"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

type providerConsentResolver struct{}

func (providerConsentResolver) HasConsent(context.Context, data_vault.ConsentRequest) (bool, error) {
	return true, nil
}

func TestDefaultVaultServiceConfigIsDisabled(t *testing.T) {
	cfg := DefaultVaultServiceConfig()
	require.False(t, cfg.Enabled)
	service, err := NewVaultService(cfg)
	require.NoError(t, err)
	require.Nil(t, service)
}

func TestVaultServiceProductionRejectsUnsafeDependencies(t *testing.T) {
	cfg := VaultServiceConfig{
		Enabled: true, Environment: "production", Profile: "production", Backend: "memory",
		KeyManager: keys.NewKeyManager(), AuditStore: data_vault.NewMemoryAuditStore(),
	}
	_, err := NewVaultService(cfg)
	require.Error(t, err)
	for _, expected := range []string{
		"production durable artifact backend", "production-safe KMS envelope cipher", "role resolver",
		"organization resolver", "consent resolver", "durable audit store", "non-exportable KMS operation interface",
	} {
		require.True(t, strings.Contains(err.Error(), expected), "error %q should contain %q", err, expected)
	}
}

func TestVaultServiceMemoryRequiresExplicitDevelopmentOnly(t *testing.T) {
	_, err := NewVaultService(VaultServiceConfig{Enabled: true, Backend: "memory"})
	require.ErrorContains(t, err, "DevelopmentOnly")
}

func TestVaultServiceFixtureFilesystemCreateAndRestore(t *testing.T) {
	root := t.TempDir()
	base := VaultServiceConfig{
		Enabled: true, Environment: "fixture", Profile: "fixture", Backend: "fixture-filesystem",
		ArtifactPath: filepath.Join(root, "artifacts"), KeyStatePath: filepath.Join(root, "keys.state"),
		AuditPath: filepath.Join(root, "audit.state"), FixtureWrappingKey: []byte("0123456789abcdef0123456789abcdef"),
		RevisionAnchor: data_vault.NewProcessRevisionAnchor(), UnsafeWindowsDevelopment: true,
		InitializeFixtureKeys: true, ConsentResolver: providerConsentResolver{},
		RoleResolver: data_vault.StaticRoleResolver{Roles: map[string]map[string]bool{}},
		OrgResolver:  data_vault.StaticOrgResolver{Members: map[string]map[string]bool{}},
	}
	service, err := NewVaultService(base)
	require.NoError(t, err)
	require.NotNil(t, service)

	base.InitializeFixtureKeys = false
	_, err = NewVaultService(base)
	require.ErrorContains(t, err, "already open")
	require.NoError(t, service.Close())
	restored, err := NewVaultService(base)
	require.NoError(t, err)
	require.NotNil(t, restored)
	require.NoError(t, restored.Close())
}
