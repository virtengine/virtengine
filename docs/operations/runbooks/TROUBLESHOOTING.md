# VirtEngine Troubleshooting Runbook

**Owner:** SRE Team
**Last Updated:** 2026-04-11

This runbook is the public operator entry point for the reliability alerts
defined in the repository monitoring configuration. Each section lists the exact
alert names it covers and the commands operators are expected to run.

## First Five Minutes

For any alert:

1. acknowledge the page,
2. capture the exact alert name and firing labels,
3. open the relevant Grafana dashboard,
4. record current rollout state,
5. decide whether to stop ongoing deploys or upgrades immediately.

Baseline commands:

```bash
virtengine status | jq
curl -s http://localhost:26657/status | jq '.result.sync_info'
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
journalctl -u virtengine -n 100 --no-pager
kubectl get pods -A
```

## Chain Halts and Consensus Loss

Covered alerts:

- `ChainHalted`
- `BlockHeightStalled`
- `BlockProductionStalled`
- `ConsensusStalled`
- `LowVotingPowerParticipation`
- `LowValidatorCount`

Diagnosis:

```bash
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state'
virtengine query staking validators --status bonded
virtengine query slashing signing-info "$(virtengine tendermint show-validator)"
```

Actions:

1. If more than one third of voting power is absent, freeze all upgrades and
   contact validator operators immediately.
2. If only one validator is degraded, recover the node first and do not invoke
   rollback yet.
3. If consensus cannot recover, move to the DR workflow in
   `_docs/disaster-recovery.md`.
4. If the incident started during an upgrade, follow the rollback path in
   `UPGRADE_PROCEDURES.md`.

## Slow Consensus and Block Latency

Covered alerts:

- `BlockTimeSlow`
- `SlowBlockProduction`
- `HighConsensusRounds`
- `ConsensusMultipleRounds`
- `ConsensusTimeoutRateHigh`

Diagnosis:

```bash
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state'
curl -s http://localhost:26657/net_info | jq '.result.peers[].remote_ip'
journalctl -u virtengine --since "15 minutes ago" | grep -Ei "timeout|prevote|precommit|proposal"
```

Actions:

1. Check whether the cause is peer latency, validator CPU pressure, or a recent
   config change.
2. If only specific validators are timing out, coordinate targeted restarts.
3. If latency rose during rollout, halt the rollout and move to the upgrade
   rollback decision path.
4. If the chain remains live, prefer remediation over rollback.

## P2P, Sync, and State Sync Failures

Covered alerts:

- `LowPeerCount`
- `NoPeers`
- `HighPeerDisconnectionRate`
- `HighPeerLatency`
- `NodeOutOfSync`
- `NodeBehind`
- `StateSyncBehind`
- `StateSyncLatencyHigh`
- `StateSyncProviderDown`
- `StateSyncSnapshotStale`

Diagnosis:

```bash
curl -s http://localhost:26657/net_info | jq
curl -s http://localhost:26657/status | jq '.result.sync_info'
ss -tulpn | grep 26656
journalctl -u virtengine --since "30 minutes ago" | grep -Ei "peer|dial|statesync|snapshot"
```

Actions:

1. Verify seed and persistent-peer config on the affected node.
2. Check firewall rules before changing application config.
3. If the state-sync provider is down or snapshots are stale, recover that
   service before onboarding new validators.
4. If a node is too far behind to recover naturally, rejoin it from a verified
   state-sync or backup path.

## Node Resource and Storage Pressure

Covered alerts:

- `NodeDown`
- `NodeHighCPU`
- `NodeCPUHigh`
- `NodeHighMemoryUsage`
- `NodeMemoryHigh`
- `DiskSpaceLow`
- `DiskSpaceCritical`
- `StateDBLargeSize`
- `ClusterNodePressure`

Diagnosis:

```bash
top -bn1 | head -20
free -h
df -h
iostat -x 1 5
kubectl describe nodes
```

Actions:

1. If disk is critical, stop optional workloads and recover free space before
   restarting the node.
2. If memory pressure is caused by VEID or enclave work, reduce concurrency or
   move work off the node.
3. If the Kubernetes cluster reports pressure, scale nodes before scaling
   replicas.
4. If a single process is leaking, capture logs and restart only that service.

## Transaction, Query, and API Degradation

Covered alerts:

- `HighTransactionLatency`
- `HighTxFailureRate`
- `TransactionFailureRateHigh`
- `HighTxPoolSize`
- `TxPoolCritical`
- `MempoolBacklog`
- `LowTPS`
- `HighBlockTxCount`
- `TransactionThroughputLow`
- `SLOTxConfirmationLatencyBreach`
- `TransactionConfirmationSLOViolation`
- `SLOQueryLatencyBreach`
- `QueryResponseSLOViolation`
- `SLOAPIAvailabilityBreach`
- `SLOAPILatencyBreach`
- `APIAvailabilitySLOViolation`
- `APIErrorRateSLOViolation`
- `APILatencySLOViolation`
- `SLAUptimeBreach`
- `SLALatencyBreach`
- `SLAUptimeWarning`

Diagnosis:

```bash
curl -s http://localhost:26657/num_unconfirmed_txs | jq
curl -s http://localhost:26657/unconfirmed_txs?limit=10 | jq
journalctl -u virtengine --since "15 minutes ago" | grep -Ei "failed|timeout|latency|rpc"
```

Actions:

1. If mempool depth is the main symptom, check whether the issue is congestion
   or failed transaction retries.
2. If query or API latency is high, inspect full-node and RPC scaling before
   changing consensus parameters.
3. For SLA breaches tied to a lease or provider, coordinate with provider ops
   and prepare credit or settlement review.
4. If the alert fired during rollout, stop traffic promotion and validate the
   previous revision.

## Module, Panics, Timeouts, and Retryable Error Spikes

Covered alerts:

- `HighErrorRate`
- `CriticalErrorRate`
- `HighCriticalErrorRate`
- `PanicRecovered`
- `FrequentPanics`
- `VEIDVerificationTimeout`
- `MFAHighFailureRate`
- `MFAMaxAttemptsExceeded`
- `EncryptionFailureRate`
- `MarketHighOrderFailureRate`
- `ProviderDaemonExternalServiceFailures`
- `InferenceHighFailureRate`
- `InferenceNonDeterministicResults`
- `HighValidationErrorRate`
- `HighUnauthorizedErrorRate`
- `HighExternalServiceErrorRate`
- `HighTimeoutRate`
- `RateLimitErrorsDetected`
- `HighRetryableErrorRate`

Diagnosis:

```bash
journalctl -u virtengine --since "15 minutes ago" | grep -Ei "panic|error|timeout|retry"
journalctl -u provider-daemon --since "15 minutes ago" | grep -Ei "error|timeout|external"
curl -s http://localhost:26660/metrics | grep -E "virtengine_errors_total|virtengine_retryable_errors_total"
```

Actions:

1. Use the module label on the alert to route to the owning surface first.
2. Treat panic frequency and non-deterministic inference as rollback-class
   issues.
3. Treat MFA, unauthorized, and encryption spikes as joint ops and security
   incidents.
4. If external-service or retryable errors surge after a deploy, stop the
   rollout and restore the last known good revision before tuning retries.

## Marketplace, Provider, and Escrow Failures

Covered alerts:

- `MarketplaceUnavailable`
- `MarketModuleErrors`
- `MarketOrderBacklog`
- `MarketOrderBacklogCritical`
- `OrderFulfillmentTimeLong`
- `OrderFailureRateHigh`
- `NoOrdersCreated`
- `NoBidsReceived`
- `BidResponseTimeSlow`
- `BidRejectionRateHigh`
- `LeaseCreationFailures`
- `LeaseTerminationRateHigh`
- `LeaseDisputeRate`
- `ActiveProviderCountLow`
- `ProviderCapacityUtilizationHigh`
- `NoProvidersForResourceType`
- `ProviderHighFailureRate`
- `ProviderDaemonHighPendingOrders`
- `ProviderDaemonClaimConflictsHigh`
- `ProviderDaemonBidLatencyHigh`
- `EscrowBalanceLow`
- `EscrowBalanceIssues`
- `EscrowWithdrawalFailures`
- `EscrowSettlementDelayed`
- `SLOOrderMatchingSuccess`
- `SLOMarketFillRate`
- `SLOMarketEfficiency`
- `SLOProviderUptimeBudgetBurning`
- `SLODeploymentSuccessRateBreach`
- `SLOProvisioningLatencyBreach`
- `ProviderDaemonUptimeSLOViolation`
- `DeploymentSuccessRateSLOViolation`
- `ProvisioningLatencySLOViolation`

Diagnosis:

```bash
virtengine query market orders --limit 20
virtengine query provider list
virtengine query bank balances "$(virtengine keys show provider -a 2>/dev/null || true)"
journalctl -u provider-daemon --since "30 minutes ago" | tail -100
kubectl get pods -n virtengine -l app=provider-daemon
```

Actions:

1. If orders are pending with no bids, check provider availability and bid-path
   latency before modifying order policy.
2. If claim conflicts are high, treat that as a scaling or deduplication issue,
   not a marketplace issue alone.
3. If escrow alerts are active, freeze payouts or withdrawals until balances and
   settlement state are verified.
4. If provider failure rate rises after a deploy, roll back the provider-daemon
   revision before widening blast radius.

## VEID Scoring, Model, and Verification Failures

Covered alerts:

- `VEIDVerificationQueue`
- `VEIDScoringUnavailable`
- `VEIDScoringSuccessRateLow`
- `VEIDScoringLatencyHigh`
- `VEIDScoringLatencyWarning`
- `VEIDMLInferenceFailureRate`
- `MLInferenceTimeout`
- `VEIDNonDeterministicInference`
- `VEIDInferenceNonDeterministic`
- `VEIDModelVersionMismatch`
- `VEIDModelNotLoaded`
- `VEIDQueueBacklog`
- `VEIDQueueCritical`
- `VEIDScoreAgreementLow`
- `VEIDHighScoreDifference`
- `VEIDDecryptionFailureRate`
- `VEIDSignatureVerificationFailures`
- `VEIDGPUMemoryHigh`
- `VEIDInferenceLatencySpike`
- `SLOVEIDResponseTime`
- `SLOVEIDVerificationSuccessRate`
- `SLOVEIDVerificationLatency`

Diagnosis:

```bash
virtengine query veid model-status
curl -s http://localhost:26660/metrics | grep -E "veid_|inference_"
journalctl -u ml-inference --since "30 minutes ago" | tail -100
sha256sum models/trust_score/* 2>/dev/null || true
```

Actions:

1. Treat any non-deterministic inference or model version mismatch as a
   consensus-safety incident.
2. Refuse to serve traffic from nodes with missing or mismatched model bundles.
3. If queue depth is growing but model hashes are correct, scale capacity or
   shed load before changing model state.
4. If decryption or signature failures spike, verify key, recipient, and
   manifest state before restarting validators.

## HPC, SLURM, and GPU Incidents

Covered alerts:

- `HPCSchedulerUnavailable`
- `HPCSubmissionAvailabilityLow`
- `HPCSchedulingLatencyHigh`
- `HPCJobQueueBacklog`
- `HPCJobQueueCritical`
- `HPCJobFailureRateHigh`
- `HPCJobTimeoutRate`
- `HPCJobCompletionReliabilityLow`
- `HPCProviderHighFailureRate`
- `NoHPCProvidersAvailable`
- `HPCProviderCapacityLow`
- `SLURMControllerUnavailable`
- `SLURMNodeDown`
- `SLURMQueueWaitTimeHigh`
- `HPCGPUUtilizationLow`
- `HPCGPUMemoryPressure`
- `HPCGPUTemperatureHigh`
- `HPCResourceOverAllocation`
- `HPCResourceWaste`

Diagnosis:

```bash
virtengine query hpc jobs --limit 20
sinfo || true
squeue -a || true
nvidia-smi || true
journalctl -u hpc-node-agent --since "30 minutes ago" | tail -100
```

Actions:

1. If the scheduler is down, stop accepting new HPC submissions until health is
   restored.
2. If queue depth is critical, check provider capacity before changing
   scheduling weights.
3. If SLURM nodes are down or queue wait time is high, drain unhealthy nodes
   and rebalance workloads.
4. If GPU memory or temperature alerts fire, reduce concurrency and validate
   hardware before re-enabling placement.

## TEE, Attestation, and Enclave Failures

Covered alerts:

- `TEEEnclaveReplicasUnavailable`
- `TEEEnclaveSignatureFailures`
- `TEEEnclaveAttestationFailures`
- `TEEEnclaveStaleAttestations`
- `TEEEnclaveUnhealthyValidators`
- `EnclaveInitializationFailed`
- `EnclaveUnhealthy`
- `EnclaveNoHardwareInProduction`
- `EnclaveHighRestartRate`
- `AllEnclavesDown`
- `EnclaveReplicasBelowMinimum`
- `AttestationVerificationFailed`
- `AttestationFailureRateHigh`
- `RemoteAttestationServiceDown`
- `UnknownMeasurementDetected`
- `AttestationQuoteGenerationFailed`
- `AttestationCertificateFetchFailed`
- `TCBVersionOutOfDate`
- `TCBUpdateAvailable`
- `DebugEnclaveInProduction`
- `NoTEENodesAvailable`
- `NitroNodeCountLow`
- `SEVSNPNodeCountLow`
- `SGXNodeCountLow`
- `TEEHardwareError`
- `EnclaveHighLatency`
- `EnclaveQueueFull`
- `EnclaveMemoryPressure`
- `PrimaryTEEPlatformDown`
- `TEEFailoverInProgress`
- `TEEFailoverFailed`

Diagnosis:

```bash
kubectl -n virtengine get deploy,pods -l app.kubernetes.io/name=tee-enclave
kubectl -n virtengine logs deploy/tee-enclave --tail=100
kubectl -n virtengine get nodes -L virtengine.io/tee-platform,virtengine.io/enclave-ready
```

Actions:

1. If hardware TEE is absent in production or an unknown measurement is seen,
   isolate the affected deployment immediately.
2. If attestation is failing but replicas are healthy, verify PCCS or KDS
   access, certificate fetches, and TCB level before restarting.
3. If failover is in progress, do not issue a second manual cutover unless the
   current failover has failed.
4. Use the DR and TEE recovery playbooks for any platform failover or stale
   attestation incident.

## Cluster Capacity, Scaling, and Regional Degradation

Covered alerts:

- `ProviderDaemonHPAMaxedOut`
- `ProviderDaemonScalingFailure`
- `FullNodeHPAMaxedOut`
- `FullNodeRPCLatencyHigh`
- `FullNodeConnectionsHigh`
- `ClusterCapacityInsufficient`
- `ClusterNodePressure`
- `RegionUnhealthy`
- `CrossRegionLatencyHigh`

Diagnosis:

```bash
kubectl get hpa -A
kubectl get pods -n virtengine
kubectl describe hpa -n virtengine
kubectl get nodes -o wide
```

Actions:

1. If HPA is maxed but nodes are healthy, raise max replicas only after
   checking upstream dependency limits.
2. If scaling is failing, repair capacity or quotas before retrying rollouts.
3. If region health drops below threshold, move to the DR region-failover path.

## SLO and Error Budget Alerts

Covered alerts:

- `ErrorBudgetFastBurn`
- `ErrorBudgetSlowBurn`
- `ErrorBudgetDepleted`
- `ErrorBudgetCritical`
- `ErrorBudgetWarning`
- `ErrorBudgetBurnRateCritical`
- `ErrorBudgetBurnRateHigh`
- `ErrorBudgetBurnRateElevated`
- `SLOMultiWindowAlert`

Actions:

1. Identify the primary symptom alert that is consuming the budget.
2. Stop rollout, promotion, or launch activity on that surface.
3. If the alert is `critical` or `depleted`, prepare rollback immediately.
4. Close the loop in `UPGRADE_PROCEDURES.md` before resuming changes.

## Escalation Boundary

Escalate to DR immediately when:

- the incident spans regions,
- consensus cannot be restored locally,
- storage replication threatens the RPO,
- enclave failover fails,
- rollback is required to protect the error budget.
