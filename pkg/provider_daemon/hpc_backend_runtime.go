// Package provider_daemon implements the VirtEngine provider daemon.
//
// Runtime construction helpers for HPC scheduler backends.
package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/virtengine/virtengine/pkg/moab_adapter"
	"github.com/virtengine/virtengine/pkg/ood_adapter"
	"github.com/virtengine/virtengine/pkg/slurm_adapter"
)

// HPCBackendClients allows tests and custom runtimes to inject concrete backend clients.
// When unset, the backend factory creates production scheduler clients from configuration.
type HPCBackendClients struct {
	SLURMClient     slurm_adapter.SLURMClient
	MOABClient      moab_adapter.MOABClient
	OODClient       ood_adapter.OODClient
	OODAuthProvider ood_adapter.VEIDAuthProvider
}

func newHPCBackendFactory(
	config HPCConfig,
	credManager *HPCCredentialManager,
	signer HPCSchedulerSigner,
	clientOverrides *HPCBackendClients,
) (*HPCBackendFactory, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid HPC config: %w", err)
	}

	if !config.Enabled {
		return nil, errors.New("HPC is not enabled in configuration")
	}

	if signer == nil {
		return nil, errors.New("signer is required")
	}

	factory := &HPCBackendFactory{
		config:          config,
		credManager:     credManager,
		signer:          signer,
		clientOverrides: clientOverrides,
		callbacks:       make([]HPCJobLifecycleCallback, 0),
	}

	scheduler, err := factory.createScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	factory.scheduler = scheduler

	return factory, nil
}

func (f *HPCBackendFactory) createProductionSLURMClient() (slurm_adapter.SLURMClient, error) {
	if f.clientOverrides != nil && f.clientOverrides.SLURMClient != nil {
		return f.clientOverrides.SLURMClient, nil
	}

	creds, _ := f.loadClusterCredentials(context.Background(), CredentialTypeSLURM)
	sshConfig, err := resolveSLURMSSHConfig(f.config.SLURM, creds)
	if err != nil {
		return nil, err
	}

	client, err := slurm_adapter.NewSSHSLURMClient(
		sshConfig,
		f.config.SLURM.ClusterName,
		f.config.SLURM.DefaultPartition,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH SLURM client: %w", err)
	}

	return client, nil
}

func (f *HPCBackendFactory) createProductionMOABClient() (moab_adapter.MOABClient, error) {
	if f.clientOverrides != nil && f.clientOverrides.MOABClient != nil {
		return f.clientOverrides.MOABClient, nil
	}

	creds, _ := f.loadClusterCredentials(context.Background(), CredentialTypeMOAB)
	cfg := f.resolvedMOABConfig(creds)

	client, err := moab_adapter.NewProductionMOABClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create production MOAB client: %w", err)
	}

	return client, nil
}

func (f *HPCBackendFactory) createProductionOODClient() (ood_adapter.OODClient, ood_adapter.VEIDAuthProvider, error) {
	if f.clientOverrides != nil && f.clientOverrides.OODClient != nil {
		if f.clientOverrides.OODAuthProvider == nil {
			return nil, nil, errors.New("OOD auth provider override is required when an OOD client override is supplied")
		}
		return f.clientOverrides.OODClient, f.clientOverrides.OODAuthProvider, nil
	}

	creds, _ := f.loadClusterCredentials(context.Background(), CredentialTypeOOD)
	cfg := f.resolvedOODConfig(creds)

	if cfg.OIDCIssuer == "" {
		return nil, nil, errors.New("ood.oidc_issuer is required")
	}
	if cfg.OIDCClientID == "" {
		return nil, nil, errors.New("ood.oidc_client_id is required")
	}
	if cfg.OIDCClientSecret == "" {
		return nil, nil, errors.New("ood.oidc_client_secret is required")
	}

	client, err := ood_adapter.NewOODProductionClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create production Open OnDemand client: %w", err)
	}

	authProvider := ood_adapter.NewVEIDAuthClient(cfg.OIDCIssuer, cfg.OIDCClientID, cfg.OIDCClientSecret)
	timeout := cfg.ConnectionTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := authProvider.Initialize(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Open OnDemand VEID auth provider: %w", err)
	}

	return client, authProvider, nil
}

func (f *HPCBackendFactory) loadClusterCredentials(ctx context.Context, credType CredentialType) (*HPCCredentials, error) {
	if f.credManager == nil || f.credManager.IsLocked() {
		return nil, errors.New("credential manager unavailable")
	}
	return f.credManager.GetCredentials(ctx, f.config.ClusterID, credType)
}

func (f *HPCBackendFactory) runtimeCredentialsConfigured() bool {
	switch f.config.SchedulerType {
	case HPCSchedulerTypeSLURM:
		if f.clientOverrides != nil && f.clientOverrides.SLURMClient != nil {
			return true
		}
		creds, _ := f.loadClusterCredentials(context.Background(), CredentialTypeSLURM)
		_, err := resolveSLURMSSHConfig(f.config.SLURM, creds)
		return err == nil
	case HPCSchedulerTypeMOAB:
		if f.clientOverrides != nil && f.clientOverrides.MOABClient != nil {
			return true
		}
		cfg := f.resolvedMOABConfig(f.mustLoadCredentials(CredentialTypeMOAB))
		return cfg.Username != "" && (cfg.Password != "" || cfg.SSHPrivateKey != "" || cfg.SSHPrivateKeyPath != "")
	case HPCSchedulerTypeOOD:
		if f.clientOverrides != nil && f.clientOverrides.OODClient != nil && f.clientOverrides.OODAuthProvider != nil {
			return true
		}
		cfg := f.resolvedOODConfig(f.mustLoadCredentials(CredentialTypeOOD))
		return cfg.BaseURL != "" && cfg.OIDCIssuer != "" && cfg.OIDCClientID != "" && cfg.OIDCClientSecret != ""
	default:
		return false
	}
}

func (f *HPCBackendFactory) mustLoadCredentials(credType CredentialType) *HPCCredentials {
	creds, _ := f.loadClusterCredentials(context.Background(), credType)
	return creds
}

func resolveSLURMSSHConfig(cfg slurm_adapter.SLURMConfig, creds *HPCCredentials) (slurm_adapter.SSHConfig, error) {
	sshConfig := slurm_adapter.DefaultSSHConfig()
	sshConfig.Host = firstNonEmpty(cfg.SSHHost, cfg.ControllerHost)
	sshConfig.Port = nonZeroInt(cfg.SSHPort, sshConfig.Port)
	sshConfig.User = firstNonEmpty(credentialString(creds, func(c *HPCCredentials) string { return c.Username }), cfg.SSHUser)
	sshConfig.PrivateKey = firstNonEmpty(credentialString(creds, func(c *HPCCredentials) string { return c.SSHPrivateKey }), cfg.SSHPrivateKey)
	sshConfig.PrivateKeyPath = firstNonEmpty(credentialString(creds, func(c *HPCCredentials) string { return c.SSHPrivateKeyPath }), cfg.SSHPrivateKeyPath)
	sshConfig.Password = firstNonEmpty(credentialString(creds, func(c *HPCCredentials) string { return c.Password }), cfg.SSHPassword)
	sshConfig.Passphrase = firstNonEmpty(credentialString(creds, func(c *HPCCredentials) string { return c.SSHPassphrase }), cfg.SSHPassphrase)
	sshConfig.HostKeyCallback = firstNonEmpty(cfg.SSHHostKeyCallback, sshConfig.HostKeyCallback)
	sshConfig.KnownHostsPath = cfg.SSHKnownHostsPath
	if cfg.ConnectionTimeout > 0 {
		sshConfig.Timeout = cfg.ConnectionTimeout
	}
	if cfg.MaxRetries > 0 {
		sshConfig.MaxRetries = cfg.MaxRetries
	}

	if sshConfig.Host == "" {
		return slurm_adapter.SSHConfig{}, errors.New("slurm.controller_host or slurm.ssh_host is required")
	}
	if sshConfig.User == "" {
		return slurm_adapter.SSHConfig{}, errors.New("slurm.ssh_user or stored SLURM username is required")
	}
	if sshConfig.PrivateKey == "" && sshConfig.PrivateKeyPath == "" && sshConfig.Password == "" {
		return slurm_adapter.SSHConfig{}, errors.New("SLURM SSH credentials require ssh_private_key, ssh_private_key_path, or ssh_password")
	}

	return sshConfig, nil
}

func (f *HPCBackendFactory) resolvedMOABConfig(creds *HPCCredentials) moab_adapter.MOABConfig {
	cfg := f.config.MOAB
	if creds != nil {
		if cfg.Username == "" {
			cfg.Username = creds.Username
		}
		if cfg.Password == "" {
			cfg.Password = creds.Password
		}
		if cfg.SSHPrivateKey == "" {
			cfg.SSHPrivateKey = creds.SSHPrivateKey
		}
		if cfg.SSHPrivateKeyPath == "" {
			cfg.SSHPrivateKeyPath = creds.SSHPrivateKeyPath
		}
		if cfg.SSHPassphrase == "" {
			cfg.SSHPassphrase = creds.SSHPassphrase
		}
	}

	if cfg.Username == "" {
		cfg.Username = os.Getenv("MOAB_USERNAME")
	}
	if cfg.Password == "" {
		cfg.Password = os.Getenv("MOAB_PASSWORD")
	}
	if cfg.SSHPrivateKeyPath == "" {
		cfg.SSHPrivateKeyPath = os.Getenv("MOAB_SSH_KEY")
	}
	if cfg.SSHPassphrase == "" {
		cfg.SSHPassphrase = os.Getenv("MOAB_SSH_PASSPHRASE")
	}

	return cfg
}

func (f *HPCBackendFactory) resolvedOODConfig(creds *HPCCredentials) ood_adapter.OODConfig {
	cfg := f.config.OOD
	if creds != nil {
		if cfg.OIDCClientID == "" {
			cfg.OIDCClientID = creds.Username
		}
		if cfg.OIDCClientSecret == "" {
			cfg.OIDCClientSecret = creds.Password
		}
		if cfg.OIDCIssuer == "" && creds.Metadata != nil {
			cfg.OIDCIssuer = creds.Metadata["oidc_issuer"]
		}
	}

	if cfg.OIDCClientSecret == "" {
		cfg.OIDCClientSecret = os.Getenv("OOD_OIDC_CLIENT_SECRET")
	}
	if cfg.OIDCClientID == "" {
		cfg.OIDCClientID = os.Getenv("OOD_OIDC_CLIENT_ID")
	}
	if cfg.OIDCIssuer == "" {
		cfg.OIDCIssuer = os.Getenv("OOD_OIDC_ISSUER")
	}

	return cfg
}

func credentialString(creds *HPCCredentials, selectFn func(*HPCCredentials) string) string {
	if creds == nil || selectFn == nil {
		return ""
	}
	return selectFn(creds)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func nonZeroInt(value int, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}
