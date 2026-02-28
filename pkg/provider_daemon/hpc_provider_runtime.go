// Package provider_daemon implements the VirtEngine provider daemon.
//
// Runtime wiring helpers for the HPC provider aggregate.
package provider_daemon

import (
	"fmt"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCProviderDeps allows the daemon runtime and tests to inject concrete HPC dependencies.
type HPCProviderDeps struct {
	CredentialManager *HPCCredentialManager
	Signer            HPCSchedulerSigner
	BackendClients    *HPCBackendClients
}

// NewHPCProviderWithDeps creates an HPC provider with optional injected runtime dependencies.
func NewHPCProviderWithDeps(
	config HPCProviderConfig,
	chainClient HPCChainClient,
	auditor HPCAuditLogger,
	deps *HPCProviderDeps,
) (*HPCProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	credManager, err := resolveHPCCredentialManager(config, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential manager: %w", err)
	}

	signer, err := resolveHPCProviderSigner(config, deps, credManager)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve HPC provider signer: %w", err)
	}

	var backendFactory *HPCBackendFactory
	if deps != nil && deps.BackendClients != nil {
		backendFactory, err = newHPCBackendFactory(config.HPC, credManager, signer, deps.BackendClients)
	} else {
		backendFactory, err = NewHPCBackendFactory(config.HPC, credManager, signer)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create backend factory: %w", err)
	}

	scheduler := backendFactory.GetScheduler()
	jobService := NewHPCJobService(config.HPC, scheduler, chainClient, auditor)

	chainSubscriber, err := NewHPCChainSubscriberWithStats(
		config.Chain,
		config.HPC.ClusterID,
		signer.GetProviderAddress(),
		chainClient,
		jobService,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain subscriber: %w", err)
	}

	usageReporter := NewHPCUsageReporter(config.HPC.UsageReporting, config.HPC.ClusterID, signer)
	settlementPipeline := NewHPCBatchSettlementPipeline(config.Settlement, chainClient, signer)

	var routingEnforcer *RoutingEnforcer
	if config.HPC.Routing.Enabled {
		routingEnforcerConfig := RoutingEnforcerConfig{
			EnforcementMode:              hpctypes.RoutingEnforcementMode(config.HPC.Routing.EnforcementMode),
			MaxDecisionAgeBlocks:         config.HPC.Routing.MaxDecisionAgeBlocks,
			MaxDecisionAgeSeconds:        config.HPC.Routing.MaxDecisionAgeSeconds,
			AllowAutomaticFallback:       config.HPC.Routing.AllowAutomaticFallback,
			RequireDecisionForSubmission: config.HPC.Routing.RequireDecisionForSubmission,
			AutoRefreshStaleDecisions:    config.HPC.Routing.AutoRefreshStaleDecisions,
			ViolationAlertThreshold:      config.HPC.Routing.ViolationAlertThreshold,
		}
		routingEnforcer = NewRoutingEnforcer(routingEnforcerConfig, nil, nil, auditor)
		jobService.SetRoutingEnforcer(routingEnforcer)
	}

	var nodeAggregator *HPCNodeAggregator
	if config.HPC.NodeAggregator.Enabled {
		nodeCfg := config.HPC.NodeAggregator
		if nodeCfg.ProviderAddress == "" {
			nodeCfg.ProviderAddress = config.HPC.ProviderAddress
		}
		if nodeCfg.ClusterID == "" {
			nodeCfg.ClusterID = config.HPC.ClusterID
		}
		if nodeCfg.NodeDiscoverer == nil {
			if slurmWrapper, ok := scheduler.(*SLURMSchedulerWrapper); ok {
				nodeCfg.NodeDiscoverer = NewHPCSLURMNodeDiscoverer(
					slurmWrapper.adapter,
					nodeCfg.ClusterID,
					nodeCfg.DefaultRegion,
					nodeCfg.DefaultDatacenter,
				)
			}
		}
		if nodeCfg.ChainReporter == nil {
			if reporter, ok := chainClient.(HPCNodeChainReporter); ok {
				nodeCfg.ChainReporter = reporter
			}
		}

		aggregator, err := NewHPCNodeAggregator(nodeCfg, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create node aggregator: %w", err)
		}
		nodeAggregator = aggregator
	}

	slurmManager, err := NewHPCSlurmK8sManager(
		config.HPC.SlurmK8s,
		config.HPC.ClusterID,
		config.HPC.ProviderAddress,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create slurm_k8s manager: %w", err)
	}

	return &HPCProvider{
		config:             config,
		backendFactory:     backendFactory,
		jobService:         jobService,
		chainSubscriber:    chainSubscriber,
		settlementPipeline: settlementPipeline,
		usageReporter:      usageReporter,
		credManager:        credManager,
		routingEnforcer:    routingEnforcer,
		nodeAggregator:     nodeAggregator,
		slurmK8sManager:    slurmManager,
		auditor:            auditor,
	}, nil
}

func resolveHPCCredentialManager(config HPCProviderConfig, deps *HPCProviderDeps) (*HPCCredentialManager, error) {
	if deps != nil && deps.CredentialManager != nil {
		return deps.CredentialManager, nil
	}
	return NewHPCCredentialManager(config.Credentials.Manager)
}

func resolveHPCProviderSigner(
	config HPCProviderConfig,
	deps *HPCProviderDeps,
	credManager *HPCCredentialManager,
) (HPCSchedulerSigner, error) {
	if deps != nil && deps.Signer != nil {
		return deps.Signer, nil
	}

	if credManager == nil {
		return nil, fmt.Errorf("credential manager is required when no signer override is provided")
	}

	if credManager.IsLocked() {
		if err := credManager.Unlock(""); err != nil {
			return nil, fmt.Errorf(
				"default HPC signer requires an unlocked credential manager; configure hpc_provider.credentials.manager or inject a signer: %w",
				err,
			)
		}
	}

	if _, err := credManager.GetPublicKey(); err != nil {
		if err := credManager.GenerateSigningKey(); err != nil {
			return nil, fmt.Errorf("failed to initialize default HPC signing key: %w", err)
		}
	}

	return &providerCredentialSigner{
		credManager:     credManager,
		providerAddress: config.HPC.ProviderAddress,
	}, nil
}
