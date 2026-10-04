# HPC workload-template approval: the intended route does not exist

**Task:** t_f3da965c
**Verdict:** the answer is **neither (a) nor (b)**. The card's premise — "either the gov
route is intended, or the in-module `WorkloadGovernanceProposal` path is intended" — is
wrong, because *both* routes are structurally impossible on this chain. The real defect is
one layer deeper and larger than the card describes.

Verified against `origin/develop` @ `0c558acf888ed536b973a42641838e8031757938`, worktree
`wt/t_f3da965c`, Go 1.26.8, `GOWORK=off GOFLAGS=-mod=mod`.

---

## THE ROOT CAUSE: the template messages were never put on the protobuf wire

All seven workload-template messages are **hand-written Go structs**, not gogoproto-generated
types. They have no `Descriptor()`, no `XXX_MessageName`, and no `proto.RegisterType` entry.

- `x/hpc/types/msg_template.go` — the messages themselves. Line 23 states the intent outright:
  `// NOTE: In production, these messages would be generated from protobuf. // For now,
  implementing minimal interface methods for testing.`
- `sdk/proto/node/virtengine/hpc/v1/tx.proto` — the real proto. It declares 12 rpcs
  (`RegisterCluster` … `UpdateParams`). **Zero** mention of `WorkloadTemplate`:
  ```
  $ grep -rn "WorkloadTemplate" sdk/proto/node/virtengine/hpc/v1/*.proto
  (no output)
  ```
- Consequently `sdk/go/node/hpc/v1/tx.pb.go`'s `MsgServer` interface has no template methods,
  and `Msg_serviceDesc` (`tx.pb.go:2459`) — the only thing registered with the router in
  `x/hpc/module.go:124` — carries no template route.

### Proof (real `app.Setup()` full app, e2e.integration build tag)

```
=== RUN   TestProbeTemplateApprovalRouteEstablishes
    HPC keeper authority = ve10d07y265gmmuvt4z0w9aw880jnsr700jxlg3em
    GetMsgV1Signers(MsgApproveWorkloadTemplate) -> signers=[] err=protoFiles does not have descriptor /: proto: not found
    MsgServiceRouter handler for MsgApproveWorkloadTemplate = <nil>
    user-signed variant GetMsgV1Signers err=protoFiles does not have descriptor /: proto: not found
--- PASS

=== RUN   TestProbeTemplateSubsystemIsInert
    MsgApproveWorkloadTemplate       router handler = <nil>
    MsgRejectWorkloadTemplate        router handler = <nil>
    MsgDeprecateWorkloadTemplate     router handler = <nil>
    MsgRevokeWorkloadTemplate        router handler = <nil>
    MsgSubmitJobFromTemplate         router handler = <nil>
    MsgCreateWorkloadTemplate        router handler = <nil>
    MsgUpdateWorkloadTemplate        router handler = <nil>
--- PASS
```

**All seven template messages are unroutable. The entire template subsystem is inert — not
just approval.** No template can be *created*, *updated*, *approved*, *rejected*, *deprecated*,
*revoked*, or *consumed via `SubmitJobFromTemplate`* by any transaction.

---

## Why route (a) — an x/gov proposal carrying `MsgApproveWorkloadTemplate` — cannot work

x/gov v1 `MsgSubmitProposal` is available (the app wires the v1 keeper:
`app/types/app.go:418` `govkeeper.NewKeeper(..., bApp.MsgServiceRouter(), govConfig,
authtypes.NewModuleAddress(govtypes.ModuleName).String())`). But submission runs three hard
gates per carried message, in `x/gov/keeper/proposal.go` (SDK
`v0.53.4-virtengine.2`), and `MsgApproveWorkloadTemplate` fails the first two:

1. **`cdc.GetMsgV1Signers(msg)`** (proposal.go:56) — resolves the signer via the proto
   descriptor. Fails: `protoFiles does not have descriptor /: proto: not found`. This is
   *submission-time*, so the tx is rejected in `ValidateBasic`/`SubmitProposal` — it can
   never be recorded, let alone pass.
2. **sole-signer must be the gov module account** (proposal.go:61-65) — unreachable,
   because gate 1 already returned an error.
3. **`k.router.Handler(msg) != nil`** (proposal.go:68-71) — `nil`, proven above.

There is **no** legacy fallback either: `git grep "RegisterLegacyProposalHandler"` across
`app/` and `x/hpc` returns nothing, so `MsgExecLegacyContent` cannot rescue these messages
either.

Consequence: the card's route (a) cannot be made to work by writing a proposal JSON. The
message would need a protobuf descriptor first. **There is no working `--authority`
invocation to document.**

## Why route (b) — the in-module `WorkloadGovernanceProposal` path — also cannot work

Option (b) is the code the card identified as dead, and it is dead as described:

| Function | Location | Callers in repo |
|---|---|---|
| `CreateWorkloadProposal` | `workload_template.go:344` | **none** (definition only) |
| `VoteOnWorkloadProposal` | `workload_template.go:381` | **none** (definition only) |
| `TallyWorkloadProposal` | `workload_template.go:425` | only `ProcessWorkloadProposals:579` |
| `ProcessWorkloadProposals` | `workload_template.go:565` | **none** — never wired to BeginBlocker/EndBlocker |

`x/hpc/module.go` BeginBlock (`:161`) runs only `CheckClusterHealth` + `CheckStaleNodes`;
EndBlock (`:178`) runs only `ProcessExpiredJobs`. So nothing ever tallies a vote, and no user
can create or cast one. `vote.Weight` is also taken from the caller with no stake lookup
(`VoteOnWorkloadProposal:404-411`), so even if wired, "simple majority" would be a raw
caller-supplied integer — not a vote.

Additionally these types are unreachable from the outside *even for reads*: no query server,
no CLI command, no genesis field (`x/hpc/genesis.go` never touches `WorkloadTemplate` or
`WorkloadGovernanceProposal`), so `GetWorkloadProposal` state could never be populated or read.

## The CLI commands are also not actually registered

The card says "CLI command is gov-only". It is worse — **the command is not in the binary at
all.** Two separate CLI trees exist:

- `x/hpc/client/cli/tx_template.go:19 GetTxCmdWorkloadTemplates()` (which *does* contain
  `GetCmdApproveWorkloadTemplate`, setting `Authority` to the signer at `:205-209`) is
  **never called by anything**. Its only importer is `x/hpc/module.go:90`
  `AppModuleBasic.GetTxCmd()`, and the app never calls `AddTxCommands`:
  `grep -rn "AddTxCommands" .` → **0 hits**.
- The registered root commands are `sdk/go/cli/hpc_tx.go:29 GetTxHPCCmd()`
  (`sdk/go/cli/tx.go:98`) and `sdk/go/cli/hpc_query.go:13 GetQueryHPCCmd()`
  (`sdk/go/cli/query.go:82`) — neither has any template subcommand.

Verified against the real built binary:

```
$ ve-hpcprobe.exe tx hpc --help
Available Commands: cancel-job, create-offering, deregister-cluster,
                    register-cluster, submit-job, update-cluster
$ ve-hpcprobe.exe tx hpc template approve foo 1.0.0 --help
HPC high-performance computing transaction subcommands      <-- fell back to parent; no such subcommand
```

The only template CLI that ships is `virtengine hpc template`
(`cmd/virtengine/cmd/hpc/`, registered at `cmd/virtengine/cmd/root.go:113-114`), and its
`create`/`update`/`deprecate` subcommands are **off-chain manifest linters** — they read a
YAML/JSON file and print a payload. `cmd/virtengine/cmd/hpc/template_tx.go:39-72` never
broadcasts anything. So there is no `approve` in the CLI at all.

Note also `x/hpc/client/cli/tx_provider.go:319 NewCmdAddTemplate()` and `:304 NewCmdUpdateParams()`
are explicit stubs returning `"... is not supported by the current HPC module build"`.

---

## The consensus-level consequence

`x/hpc/keeper/workload_template.go:151-184 ApproveWorkloadTemplate` is authority-gated to
`k.authority`, which `app/types/app.go:602-607` sets to
`authtypes.NewModuleAddress(govtypes.ModuleName).String()`. Since no message can reach that
code path (all template Msgs unroutable; `ProcessWorkloadProposals` uncalled), the state
transition `WorkloadApprovalPending → WorkloadApprovalApproved` is **unreachable in
production**.

This is not a docs problem. The docs' claim at
`providers/hpc-workload-templates.mdx:100-101` ("New templates require an on-chain proposal
and vote") describes a flow that does not exist. The `UNVERIFIED` marker at `:103` was
correct to be there.

Only a chain upgrade that seeds `x/hpc` store keys could ever produce an approved template —
and `InitGenesis` has no template fields, so even genesis cannot.

---

## ANSWER TO THE CARD

**The intended approval route cannot be established from source because it was never
implemented.** Both candidate routes are structurally impossible:

- (a) requires the messages to have a protobuf descriptor and a router entry — they have neither.
- (b) requires a Msg + CLI + BeginBlocker/EndBlocker caller — none of the three exists.

Per the card's own framing, **the dead code is the bug**, and it is dead at a more basic
level than the card identified: the messages were never wired into the module's protobuf
surface at all. Fixing (b) alone would still not work, because a hand-written struct cannot
be routed regardless of who calls it.

## What a fix must do (consensus-critical — needs sign-off)

1. **Define the messages in proto.** Add `MsgCreateWorkloadTemplate`,
   `MsgUpdateWorkloadTemplate`, `MsgApproveWorkloadTemplate`, `MsgRejectWorkloadTemplate`,
   `MsgDeprecateWorkloadTemplate`, `MsgRevokeWorkloadTemplate`, `MsgSubmitJobFromTemplate`
   to `sdk/proto/node/virtengine/hpc/v1/tx.proto`, each with the matching rpc, and
   `option (cosmos.msg.v1.signer) = "authority"` (or `"creator"`) plus `amino.name`.
   **New proto fields/messages — additive only, never renumber.** Then regenerate.
2. **Delete the hand-written Go structs** in `msg_template.go` and switch to the generated
   types (or, at minimum, register them via `gogoproto.RegisterType` + a `Msg_serviceDesc`).
   Currently they are also registered in `RegisterLegacyAminoCodec` (`codec.go:42-48`) but
   **not** in `RegisterInterfaces` (`codec.go:52-68`) — that omission alone means they are
   not `sdk.Msg` implementations on the wire.
3. **Pick the governance route deliberately** and implement it end to end. The reference
   pattern for an authority-gated, gov-carried message already exists in this repo — see
   `sdk/proto/node/virtengine/benchmark/v1/tx.proto:96-108` (`MsgUnflagProvider`,
   `MsgResolveAnomalyFlag`), which have `option (cosmos.msg.v1.signer) = "authority"` and
   matching rpcs. With step 1 done, route (a) becomes viable with no further wiring, because
   x/gov v1 executes carried messages through the module's `MsgServiceRouter`.
4. **Add genesis fields** for templates if templates must survive export/import.
5. **Register the CLI** (`sdk/go/cli/hpc_tx.go`) or delete the unused
   `x/hpc/client/cli/tx_template.go` tree so the two CLI implementations stop diverging.
6. **Decide the in-module `WorkloadGovernanceProposal` route: implement or delete.**
   `ProcessWorkloadProposals` has no caller; the type has no query, no CLI, no genesis, and
   its `vote.Weight` is caller-supplied. Shipping it as-is would be a governance-weight
   vulnerability, not a feature.

Steps 1–3 are consensus-critical (they add messages to a live module's transaction surface)
and require explicit user sign-off per the chain-core standing rules. I have NOT implemented
them; this card asked for the route determination, and the determination is that neither
exists.

## Test command (reproduce)

```bash
git checkout wt/t_f3da965c
GOWORK=off GOFLAGS=-mod=mod go test -tags="e2e.integration" -run TestProbeTemplate \
  -v -timeout 900s ./tests/integration/hpc/
```
The probe asserted `require.Nil(router.Handler(msg))` and `require.Error(GetMsgV1Signers)`
for all seven messages; both passed. (Probe removed after measurement — it asserted a bug,
so it must not be left as a passing test in the tree.)