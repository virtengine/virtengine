# Task 85A Provider Mutation Inventory - 2026-07-23

This inventory classifies provider-daemon chain interactions after Task 85A. Production provider-originated writes must enter `ProviderMutationSubmitter`; query-only paths remain read-only and unsupported/customer-signed mutations fail closed.

## Durable Provider Writes

| Class | Registered kind | SDK message | Production entry point |
| --- | --- | --- | --- |
| Market bid | `market.create_bid` | `virtengine.market.v1beta5.MsgCreateBid` | `pkg/provider_daemon/chain_client.go:354` |
| Market bid | `market.close_bid` | `virtengine.market.v1beta5.MsgCloseBid` | Registry-ready; no direct production caller found |
| Market lease | `market.withdraw_lease` | `virtengine.market.v1beta5.MsgWithdrawLease` | Registry-ready; no direct production caller found |
| HPC cluster | `hpc.register_cluster` | `virtengine.hpc.v1.MsgRegisterCluster` | Registry-ready; no direct production caller found |
| HPC cluster | `hpc.update_cluster` | `virtengine.hpc.v1.MsgUpdateCluster` | Registry-ready; no direct production caller found |
| HPC cluster | `hpc.deregister_cluster` | `virtengine.hpc.v1.MsgDeregisterCluster` | Registry-ready; no direct production caller found |
| HPC offering | `hpc.create_offering` | `virtengine.hpc.v1.MsgCreateOffering` | Registry-ready; no direct production caller found |
| HPC offering | `hpc.update_offering` | `virtengine.hpc.v1.MsgUpdateOffering` | Registry-ready; no direct production caller found |
| HPC status/accounting | `hpc.report_job_status` | `virtengine.hpc.v1.MsgReportJobStatus` | `pkg/provider_daemon/chain_client.go:499`, `pkg/provider_daemon/chain_client.go:566`, `pkg/provider_daemon/chain_client.go:591`, `pkg/provider_daemon/chain_client.go:621` |
| HPC node metadata | `hpc.update_node_metadata` | `virtengine.hpc.v1.MsgUpdateNodeMetadata` | `pkg/provider_daemon/chain_client.go:558` |
| Resources inventory | `resources.provider_heartbeat` | `virtengine.resources.v1.MsgProviderHeartbeat` | `pkg/provider_daemon/chain_client.go:518`, `pkg/provider_daemon/resource_sync.go:151` |
| Resources allocation | `resources.activate_allocation` | `virtengine.resources.v1.MsgActivateAllocation` | Registry-ready; callback/future reconciler only |
| Resources allocation | `resources.release_allocation` | `virtengine.resources.v1.MsgReleaseAllocation` | Registry-ready; callback/future reconciler only |
| Usage | `settlement.record_usage` | `virtengine.settlement.v1.MsgRecordUsage` | `pkg/provider_daemon/chain_submitter.go:329` |
| Settlement | `settlement.settle_order` | `virtengine.settlement.v1.MsgSettleOrder` | `pkg/provider_daemon/chain_submitter.go:295` |
| Provider lifecycle | `provider.create` | `virtengine.provider.v1beta4.MsgCreateProvider` | Registry-ready; operator flow only |
| Provider lifecycle | `provider.update` | `virtengine.provider.v1beta4.MsgUpdateProvider` | Registry-ready; operator flow only |
| Provider lifecycle | `provider.delete` | `virtengine.provider.v1beta4.MsgDeleteProvider` | Registry-ready; operator flow only |
| Provider domain | `provider.generate_domain_token` | `virtengine.provider.v1beta4.MsgGenerateDomainVerificationToken` | Registry-ready; operator flow only |
| Provider domain | `provider.verify_domain` | `virtengine.provider.v1beta4.MsgVerifyProviderDomain` | Registry-ready; operator flow only |
| Provider domain | `provider.request_domain_verification` | `virtengine.provider.v1beta4.MsgRequestDomainVerification` | Registry-ready; operator flow only |
| Provider domain | `provider.confirm_domain_verification` | `virtengine.provider.v1beta4.MsgConfirmDomainVerification` | `pkg/provider_daemon/chain_client.go:113`, `pkg/provider_daemon/domain_verification_checker.go:222` |
| Provider domain | `provider.revoke_domain_verification` | `virtengine.provider.v1beta4.MsgRevokeDomainVerification` | Registry-ready; operator flow only |
| Provider key | `provider.set_signing_key` | `virtengine.provider.v1beta4.MsgSetProviderSigningKey` | Registry-ready; operator flow only |
| Provider key | `provider.rotate_signing_key` | `virtengine.provider.v1beta4.MsgRotateProviderSigningKey` | Registry-ready; operator flow only |
| Provider key | `provider.revoke_signing_key` | `virtengine.provider.v1beta4.MsgRevokeProviderSigningKey` | Registry-ready; operator flow only |
| Marketplace callback | `marketplace.waldur_callback` | `virtengine.marketplace.v1.MsgWaldurCallback` | `pkg/provider_daemon/chain_callback_sink.go:61` |
| Support request | `support.update_request` | `virtengine.support.v1.MsgUpdateSupportRequest` | `pkg/provider_daemon/support_chain.go:70` |
| Support response | `support.add_response` | `virtengine.support.v1.MsgAddSupportResponse` | `pkg/provider_daemon/support_chain.go:85` |
| Support external ref | `support.register_external` | `virtengine.support.v1.MsgRegisterExternalTicket` | `pkg/provider_daemon/support_chain.go:111` |
| Support external ref | `support.update_external` | `virtengine.support.v1.MsgUpdateExternalTicket` | `pkg/provider_daemon/support_chain.go:149` |

## Query-Only Chain Reads

The following paths are read-only and intentionally stay outside the mutation queue:

- Open orders, provider bids, provider config, allocations, reservations, billing rules and current block height in `pkg/provider_daemon/chain_client.go`.
- Provider signing-key epoch, usage stream state and account sequence reads used by the signed submitters.
- Store queries used for resource inventory, domain verification records and support external refs.

## Unsupported Or Non-Provider Writes

- Generated `MsgClient` mutation calls are not used by provider production paths.
- `MsgCreateLease` and `MsgCloseLease` are customer/owner-signed by current market schemas. They are not registered in the provider-originated queue; attempts to submit them with provider queue kinds return `ErrUnknownProviderMutation`.
- The old JSON `ChainSubmitterClient` path remains only behind `AllowTestLegacyChainClient`; production submitter construction requires the generalized mutation submitter.
- Waldur/provisioning callbacks fail startup without the durable chain callback sink. File callback sinks are not a production substitute for Task 85A.
- HPC node metadata fails startup/submission without the injected `HPCNodeChainReporter`; the raw HTTP and no-op fallback paths are removed.
- Domain verification requires the production chain client generalized mutation backend; the legacy standalone RPC confirmation backend is not used.
