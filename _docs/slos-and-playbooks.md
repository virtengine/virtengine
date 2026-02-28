# VirtEngine SLOs and Operational Playbooks

**Owner:** SRE + Operations
**Last Updated:** 2026-04-11

This document is the authoritative reliability index for the repo-owned
monitoring and runbook surface. It aligns the live alert rules under
`deploy/monitoring`, `deploy/tee`, and `deploy/cockroachdb` with the operator
actions in the public runbooks.

## Source of Truth

The monitored reliability surface in this repository is defined by:

- `deploy/monitoring/prometheus/rules/chain_alerts.yaml`
- `deploy/monitoring/prometheus/rules/virtengine_alerts.yaml`
- `deploy/monitoring/prometheus/rules/slo_alerts.yaml`
- `deploy/monitoring/prometheus/rules/scaling_alerts.yaml`
- `deploy/monitoring/prometheus/rules/sla.yaml`
- `deploy/monitoring/alerts/chain-health.yaml`
- `deploy/monitoring/alerts/marketplace.yaml`
- `deploy/monitoring/alerts/veid-scoring.yaml`
- `deploy/monitoring/alerts/hpc-scheduling.yaml`
- `deploy/monitoring/alerts/enclave-health.yaml`
- `deploy/monitoring/alerts/slo-compliance.yaml`
- `deploy/tee/monitoring.yaml`
- `deploy/cockroachdb/multi-region-values.yaml`

If an alert is added to those files, this document and the linked runbooks must
be updated in the same change.

## Reliability Objectives

These objectives match the thresholds used by the checked-in SLO and alert rules.

| SLO ID | Service | Objective |
| --- | --- | --- |
| `SLO-NODE-001` | Node availability | `99.95%` node uptime over `28d` |
| `SLO-NODE-005` | Chain transactions | P95 confirmation latency under `10s` |
| `SLO-NODE-006` | Query surface | P95 query latency under `2s` |
| `SLO-PROV-001` | Provider daemon | `99.90%` uptime over `28d` |
| `SLO-PROV-003` | Provider deployments | Deployment success rate at or above `99%` |
| `SLO-PROV-005` | Provisioning | P95 provisioning latency under `300s` |
| `SLO-API-001` | API availability | `99.90%` successful non-5xx responses |
| `SLO-API-003` | API latency | P95 request latency under `2s` |
| `SLO-VEID-001` | VEID verification | Verification success rate at or above `95%` |
| `SLO-VEID-002` | VEID latency | P95 verification latency under `30s` |
| `SLO-VEID-DETERMINISM` | VEID consensus safety | Zero non-deterministic inference events |
| `SLO-MARKET-001` | Marketplace fill rate | Fill rate at or above `90%` |
| `SLO-MARKET-002` | Marketplace efficiency | Efficiency score above `60` |
| `SLO-HPC-001` | HPC submission | Submission success rate at or above `99%` |
| `SLO-HPC-002` | HPC scheduling | P95 scheduling latency under `15m` |
| `SLO-HPC-003` | HPC completion | Completion reliability at or above `99%` |
| `SLO-TEE-001` | TEE readiness | At least `2` healthy enclave replicas and no stale attestation breach |
| `SLO-DB-001` | CockroachDB replication | Replication lag under `300s` |

## Error Budget Policy

- `warning`: investigate in the current on-call shift and record the likely
  burn driver.
- `high`: pause discretionary rollout work for the affected service until the
  slope is understood.
- `critical`: stop progressive rollout or promotion, prepare rollback, and
  page the owning lead.
- `depleted`: feature-freeze the affected surface until the incident is closed
  and a mitigation plan is approved.

These rules apply to:

- `SLONodeUptimeBudgetBurning`
- `SLONodeUptimeBudgetCritical`
- `SLOProviderUptimeBudgetBurning`
- `ErrorBudgetFastBurn`
- `ErrorBudgetSlowBurn`
- `ErrorBudgetDepleted`
- `ErrorBudgetCritical`
- `ErrorBudgetWarning`
- `ErrorBudgetBurnRateCritical`
- `ErrorBudgetBurnRateHigh`
- `ErrorBudgetBurnRateElevated`
- `SLOMultiWindowAlert`

Primary operator actions for those alerts live in
`docs/operations/runbooks/UPGRADE_PROCEDURES.md` under the rollout abort and
rollback sections, plus `docs/operations/runbooks/TROUBLESHOOTING.md` under
the SLO and burn-rate section.

## Alert Coverage Matrix

Every operational alert family currently checked into monitoring maps to a real
runbook section below.

| Alert family | Exact alert names covered | Primary runbook |
| --- | --- | --- |
| Chain halt and stalled consensus | `ChainHalted`, `BlockHeightStalled`, `BlockProductionStalled`, `ConsensusStalled`, `LowVotingPowerParticipation`, `LowValidatorCount` | `docs/operations/runbooks/TROUBLESHOOTING.md#chain-halts-and-consensus-loss` |
| Slow consensus and slow blocks | `BlockTimeSlow`, `SlowBlockProduction`, `HighConsensusRounds`, `ConsensusMultipleRounds`, `ConsensusTimeoutRateHigh` | `docs/operations/runbooks/TROUBLESHOOTING.md#slow-consensus-and-block-latency` |
| Validator participation | `ValidatorDown`, `ValidatorMissingBlocks`, `ValidatorMissedBlocks`, `ValidatorLowUptime`, `MissedBlocksHigh`, `NodeUptimeSLOViolation`, `SLONodeUptimeBudgetBurning`, `SLONodeUptimeBudgetCritical` | `docs/operations/runbooks/VALIDATOR_SETUP.md#validator-monitoring-and-response` |
| P2P, sync, and state sync | `LowPeerCount`, `NoPeers`, `HighPeerDisconnectionRate`, `HighPeerLatency`, `NodeOutOfSync`, `NodeBehind`, `StateSyncBehind`, `StateSyncLatencyHigh`, `StateSyncProviderDown`, `StateSyncSnapshotStale` | `docs/operations/runbooks/TROUBLESHOOTING.md#p2p-sync-and-state-sync-failures` |
| Node resources and storage | `NodeDown`, `NodeHighCPU`, `NodeCPUHigh`, `NodeHighMemoryUsage`, `NodeMemoryHigh`, `DiskSpaceLow`, `DiskSpaceCritical`, `StateDBLargeSize` | `docs/operations/runbooks/TROUBLESHOOTING.md#node-resource-and-storage-pressure` |
| Transaction, query, and API latency | `HighTransactionLatency`, `HighTxFailureRate`, `TransactionFailureRateHigh`, `HighTxPoolSize`, `TxPoolCritical`, `MempoolBacklog`, `LowTPS`, `HighBlockTxCount`, `TransactionThroughputLow`, `SLOTxConfirmationLatencyBreach`, `TransactionConfirmationSLOViolation`, `SLOQueryLatencyBreach`, `QueryResponseSLOViolation`, `SLOAPIAvailabilityBreach`, `SLOAPILatencyBreach`, `APIAvailabilitySLOViolation`, `APIErrorRateSLOViolation`, `APILatencySLOViolation`, `SLAUptimeBreach`, `SLALatencyBreach`, `SLAUptimeWarning` | `docs/operations/runbooks/TROUBLESHOOTING.md#transaction-query-and-api-degradation` |
| Marketplace, provider, and escrow | `MarketplaceUnavailable`, `MarketModuleErrors`, `MarketOrderBacklog`, `MarketOrderBacklogCritical`, `NoBidsReceived`, `BidResponseTimeSlow`, `BidRejectionRateHigh`, `OrderFulfillmentTimeLong`, `OrderFailureRateHigh`, `LeaseCreationFailures`, `LeaseTerminationRateHigh`, `LeaseDisputeRate`, `ActiveProviderCountLow`, `ProviderCapacityUtilizationHigh`, `NoProvidersForResourceType`, `ProviderHighFailureRate`, `ProviderDaemonHighPendingOrders`, `ProviderDaemonClaimConflictsHigh`, `ProviderDaemonBidLatencyHigh`, `EscrowBalanceLow`, `EscrowBalanceIssues`, `EscrowWithdrawalFailures`, `EscrowSettlementDelayed`, `SLOOrderMatchingSuccess`, `SLOMarketFillRate`, `SLOMarketEfficiency`, `SLOProviderUptimeBudgetBurning`, `SLODeploymentSuccessRateBreach`, `SLOProvisioningLatencyBreach`, `ProviderDaemonUptimeSLOViolation`, `DeploymentSuccessRateSLOViolation`, `ProvisioningLatencySLOViolation` | `docs/operations/runbooks/TROUBLESHOOTING.md#marketplace-provider-and-escrow-failures` |
| VEID scoring and model integrity | `VEIDVerificationQueue`, `VEIDScoringUnavailable`, `VEIDScoringSuccessRateLow`, `VEIDScoringLatencyHigh`, `VEIDScoringLatencyWarning`, `VEIDMLInferenceFailureRate`, `MLInferenceTimeout`, `VEIDNonDeterministicInference`, `VEIDInferenceNonDeterministic`, `VEIDModelVersionMismatch`, `VEIDModelNotLoaded`, `VEIDQueueBacklog`, `VEIDQueueCritical`, `VEIDScoreAgreementLow`, `VEIDHighScoreDifference`, `VEIDDecryptionFailureRate`, `VEIDSignatureVerificationFailures`, `VEIDGPUMemoryHigh`, `VEIDInferenceLatencySpike`, `SLOVEIDResponseTime`, `SLOVEIDVerificationSuccessRate`, `SLOVEIDVerificationLatency` | `docs/operations/runbooks/TROUBLESHOOTING.md#veid-scoring-model-and-verification-failures` |
| HPC, SLURM, and GPU | `HPCSchedulerUnavailable`, `HPCSubmissionAvailabilityLow`, `HPCSchedulingLatencyHigh`, `HPCJobQueueBacklog`, `HPCJobQueueCritical`, `HPCJobFailureRateHigh`, `HPCJobTimeoutRate`, `HPCJobCompletionReliabilityLow`, `HPCProviderHighFailureRate`, `NoHPCProvidersAvailable`, `HPCProviderCapacityLow`, `SLURMControllerUnavailable`, `SLURMNodeDown`, `SLURMQueueWaitTimeHigh`, `HPCGPUUtilizationLow`, `HPCGPUMemoryPressure`, `HPCGPUTemperatureHigh`, `HPCResourceOverAllocation`, `HPCResourceWaste` | `docs/operations/runbooks/TROUBLESHOOTING.md#hpc-slurm-and-gpu-incidents` |
| TEE readiness and attestation | `TEEEnclaveReplicasUnavailable`, `TEEEnclaveSignatureFailures`, `TEEEnclaveAttestationFailures`, `TEEEnclaveStaleAttestations`, `TEEEnclaveUnhealthyValidators`, `EnclaveInitializationFailed`, `EnclaveUnhealthy`, `EnclaveNoHardwareInProduction`, `EnclaveHighRestartRate`, `AllEnclavesDown`, `EnclaveReplicasBelowMinimum`, `AttestationVerificationFailed`, `AttestationFailureRateHigh`, `RemoteAttestationServiceDown`, `UnknownMeasurementDetected`, `AttestationQuoteGenerationFailed`, `AttestationCertificateFetchFailed`, `TCBVersionOutOfDate`, `TCBUpdateAvailable`, `DebugEnclaveInProduction`, `NoTEENodesAvailable`, `NitroNodeCountLow`, `SEVSNPNodeCountLow`, `SGXNodeCountLow`, `TEEHardwareError`, `EnclaveHighLatency`, `EnclaveQueueFull`, `EnclaveMemoryPressure`, `PrimaryTEEPlatformDown`, `TEEFailoverInProgress`, `TEEFailoverFailed` | `docs/operations/runbooks/TROUBLESHOOTING.md#tee-attestation-and-enclave-failures` and `_docs/disaster-recovery.md#tee-failover-and-attestation-recovery` |
| Scaling, RPC, and cluster capacity | `ProviderDaemonHPAMaxedOut`, `ProviderDaemonScalingFailure`, `FullNodeHPAMaxedOut`, `FullNodeRPCLatencyHigh`, `FullNodeConnectionsHigh`, `ClusterCapacityInsufficient`, `ClusterNodePressure`, `RegionUnhealthy`, `CrossRegionLatencyHigh` | `docs/operations/runbooks/TROUBLESHOOTING.md#cluster-capacity-scaling-and-regional-degradation` and `_docs/disaster-recovery.md#region-failover` |
| CockroachDB and replicated state | `CockroachDBNodeDown`, `CockroachDBHighReplicationLag`, `CockroachDBLowDiskSpace`, `CockroachDBHighMemoryUsage` | `_docs/disaster-recovery.md#cockroachdb-and-replicated-state-recovery` |
| Module and retryable error spikes | `HighErrorRate`, `CriticalErrorRate`, `HighCriticalErrorRate`, `PanicRecovered`, `FrequentPanics`, `VEIDVerificationTimeout`, `MFAHighFailureRate`, `MFAMaxAttemptsExceeded`, `EncryptionFailureRate`, `MarketHighOrderFailureRate`, `ProviderDaemonExternalServiceFailures`, `InferenceHighFailureRate`, `InferenceNonDeterministicResults`, `HighValidationErrorRate`, `HighUnauthorizedErrorRate`, `HighExternalServiceErrorRate`, `HighTimeoutRate`, `RateLimitErrorsDetected`, `HighRetryableErrorRate` | `docs/operations/runbooks/TROUBLESHOOTING.md#module-panics-timeouts-and-retryable-error-spikes` |
| Error budget and burn-rate policy | `ErrorBudgetFastBurn`, `ErrorBudgetSlowBurn`, `ErrorBudgetDepleted`, `ErrorBudgetCritical`, `ErrorBudgetWarning`, `ErrorBudgetBurnRateCritical`, `ErrorBudgetBurnRateHigh`, `ErrorBudgetBurnRateElevated`, `SLOMultiWindowAlert` | `docs/operations/runbooks/TROUBLESHOOTING.md#slo-and-error-budget-alerts` and `docs/operations/runbooks/UPGRADE_PROCEDURES.md#rollback-decision-rules` |

## Operator Response Rules

### Tier 0

- Acknowledge within `15m`.
- Freeze rollout or promotion work on the affected surface.
- Start evidence capture immediately:
  - active alerts,
  - relevant Grafana dashboard URL,
  - recent `journalctl` or `kubectl logs`,
  - current rollback readiness.

### Tier 1

- Acknowledge within `30m`.
- Continue only the rollout work that is not causally related to the alert.
- If the alert persists longer than one review window, escalate to the service
  owner.

### Tier 2

- Acknowledge within `4h`.
- If customer-facing latency or queue depth crosses the critical threshold,
  reclassify as high severity and follow the Tier 1 path.

## Required Evidence for Any Incident

Record all of the following in the incident ticket or drill bundle:

1. alert name and first-fire timestamp,
2. affected region, cluster, validator, or provider IDs,
3. exact commands run,
4. artifact paths created by repo scripts,
5. rollback decision and execution timestamp if rollback was used,
6. the cleared-alert time and residual risk.

## Change Rule

No reliability alert may remain in monitoring without one of the following:

- a direct mapping in the matrix above,
- a concrete response section in the linked runbook,
- a repo-owned command or script that operators can execute as written.
