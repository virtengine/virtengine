# HPC workload-template approval: the intended route does not exist

**Task:** t_f3da965c
**Verdict:** the answer is **neither (a) nor (b)**. The card's premise — "either the gov
route is intended, or the in-module `WorkloadGovernanceProposal` path is intended" — is
wrong, because *both* routes are structurally impossible on this chain. The real defect is
one layer deeper and larger than the card describes.

Verified against `origin/develop` @ `60767240e77d900b92d118f801985d056fc49196` (the merge of
this doc), worktree `wt/t_f3da965c`, Go 1.26.8, `GOWORK=off GOFLAGS=-mod=mod`.
The probe output below was re-measured on that exact tree during review, not carried over.

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

The strongest form of the evidence is not "the router has no handler" — it is that the
message **cannot even be put on the wire**. Because `ProtoMessage()` is hand-implemented and
there is no `Descriptor()`, `sdk.MsgTypeURL` collapses to the empty string:

```
MsgTypeURL(MsgApproveWorkloadTemplate)        = "/"
codectypes.NewAnyWithValue(approve).TypeUrl   = "/"
MsgTypeURL(bank MsgSend, control)             = "/cosmos.bank.v1beta1.MsgSend"   <- control
```

So the failure sits at the **tx parse layer**, strictly upstream of the router and of every
x/gov gate — the bytes encode and then fail to decode:

```
TxEncoder -> 11 bytes, err=<nil>
TxDecoder -> err=unable to resolve type URL /: tx parse error
             [virtengine/cosmos-sdk@v0.53.4-virtengine.2/x/auth/tx/decoder.go:44]
```

Full probe output (source inlined below, so this is re-runnable rather than a claim):

```
=== RUN   TestProbeTemplateApprovalRouteEstablishes
    MsgTypeURL(MsgApproveWorkloadTemplate) = "/"
    codectypes.NewAnyWithValue(approve).TypeUrl = "/"
    MsgTypeURL(bank MsgSend, control)      = "/cosmos.bank.v1beta1.MsgSend"
    cdc.GetMsgV1Signers(approve) -> signers=[] err=protoFiles does not have descriptor /: proto: not found
    MsgServiceRouter.Handler(approve) = <nil>
    InterfaceRegistry.Resolve("/") err=unable to resolve type URL /
    TxEncoder -> 11 bytes, err=<nil>
    TxDecoder -> err=unable to resolve type URL /: tx parse error [virtengine/cosmos-sdk@v0.53.4-virtengine.2/x/auth/tx/decoder.go:44]
    LegacyAmino().MarshalJSON(approve) -> {"authority":"ve1074y0eqvtepqxlahkp04rknfeuagyjcd4g5yl7","template_id":"tmpl-1","version":"1.0.0"} err=<nil>
--- PASS: TestProbeTemplateApprovalRouteEstablishes (0.51s)
=== RUN   TestProbeTemplateSubsystemIsInert
    *types.MsgApproveWorkloadTemplate router handler = <nil>
    *types.MsgRejectWorkloadTemplate router handler = <nil>
    *types.MsgDeprecateWorkloadTemplate router handler = <nil>
    *types.MsgRevokeWorkloadTemplate router handler = <nil>
    *types.MsgSubmitJobFromTemplate router handler = <nil>
    *types.MsgCreateWorkloadTemplate router handler = <nil>
    *types.MsgUpdateWorkloadTemplate router handler = <nil>
--- PASS: TestProbeTemplateSubsystemIsInert (0.08s)
```

Note the `LegacyAmino` line, not the last one: the **legacy amino** JSON encoding succeeds.
That is a decoy — it buys nothing on the wire, which is why the amino registration at
`codec.go:42-48` without a matching `RegisterInterfaces` entry is not evidence the messages work.

**All seven template messages are unroutable. The entire template subsystem is inert — not
just approval.** No template can be *created*, *updated*, *approved*, *rejected*, *deprecated*,
*revoked*, or *consumed via `SubmitJobFromTemplate`* by any transaction.

---

## Why route (a) — an x/gov proposal carrying `MsgApproveWorkloadTemplate` — cannot work

x/gov v1 `MsgSubmitProposal` is available (the app wires the v1 keeper:
`app/types/app.go:418` `govkeeper.NewKeeper(..., bApp.MsgServiceRouter(), govConfig,
authtypes.NewModuleAddress(govtypes.ModuleName).String())`). But the message never reaches
that keeper, because it cannot be decoded out of a transaction in the first place
(proof above: `TxDecoder -> unable to resolve type URL /: tx parse error`). Submission is
therefore impossible for a reason that is **upstream of all of x/gov's own gates**.

For completeness, those gates would each reject it anyway. In
`x/gov/keeper/proposal.go` (SDK `virtengine/cosmos-sdk@v0.53.4-virtengine.2`):

1. **`signers, _, err := k.cdc.GetMsgV1Signers(msg)`** (`:53`) — fails:
   `protoFiles does not have descriptor /: proto: not found`. This is *submission-time*, so the
   proposal can never be recorded, let alone pass.
2. **sole signer must be the gov module account** (`:58`, `:62-63`) — unreachable, because gate 1
   already returned an error.
3. **`handler := k.router.Handler(msg); if handler == nil`** (`:67-69`) — `nil`, proven above.

There is **no** legacy fallback either: `git grep "RegisterLegacyProposalHandler"` across
`app/` and `x/hpc` returns nothing, so `MsgExecLegacyContent` cannot rescue these messages
either.

Consequence: the card's route (a) cannot be made to work by writing a proposal JSON. The
message would need a protobuf descriptor first. **There is no `--authority` invocation to
document — and this is not an omission in the documentation, it is a property of the chain:**
with `MsgTypeURL` collapsing to `"/"`, no proposal payload expressing this message can even be
parsed.

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

## How to re-derive this (the probe is inlined below)

There is **no in-tree test for this finding**, deliberately: a test asserting `router.Handler(msg) == nil`
passes *because* the bug exists, so committing it would mean shipping a green test that pins the
defect in place. The probe below is the artifact instead — it is complete and self-contained, so
anyone can materialize it and re-derive every number above.

**Do not just run the `go test` line against the repo as-is.** Without the probe file present that
command exits 0 with `testing: warning: no tests to run` — a *false green* that verifies nothing.

Materialize the probe, then run it:

```bash
# 1. Save the probe below as tests/integration/hpc/zz_probe_scratch_test.go
# 2. Run it:
GOWORK=off GOFLAGS=-mod=mod go test -tags="e2e.integration" -run 'TestProbeTemplate' \
  -v -timeout 900s ./tests/integration/hpc/
# 3. Remove it again (untracked scratch; do not commit it):
rm tests/integration/hpc/zz_probe_scratch_test.go
```

Expect `--- PASS` on both tests. **A PASS here means the bug is still present.** If someone
fixes the wiring, these assertions will start failing — that is the intended signal.

The `LegacyAmino` line prints a randomly generated address that differs on every run; the
doc's captured value is one sample, not a fixed expectation.

Verified on `origin/develop` @ `60767240e7`, Go 1.26.8, `GOWORK=off GOFLAGS=-mod=mod`.
The build tag matters: without `-tags="e2e.integration"` the file compiles to nothing.

### The static half (no probe needed)

These greps re-derive the root cause on their own and need no app:

```bash
grep -rn "WorkloadTemplate" sdk/proto/node/virtengine/hpc/v1/*.proto   # -> 0 hits (root cause)
grep -rn "ProcessWorkloadProposals" --include=*.go .                   # -> definition only
grep -rn "AddTxCommands" --include=*.go .                             # -> 0 hits (CLI never wired)
grep -rn "RegisterInterfaces" -A 20 x/hpc/types/codec.go               # -> none of the 7 msgs listed
```

### The probe source

```go
//go:build e2e.integration

package hpc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestProbeTemplateApprovalRouteEstablishes(t *testing.T) {
	a := app.Setup(app.WithChainID("virtengine-probe-1"))
	t.Cleanup(func() { _ = a.Close() })

	govAuthority := sdktestutil.AccAddress(t).String()
	approve := hpctypes.NewMsgApproveWorkloadTemplate(govAuthority, "tmpl-1", "1.0.0")

	// Proof 1: no protobuf descriptor => sdk.MsgTypeURL collapses to "/".
	t.Logf("MsgTypeURL(MsgApproveWorkloadTemplate) = %q", sdk.MsgTypeURL(approve))
	anyMsg, err := codectypes.NewAnyWithValue(approve)
	require.NoError(t, err)
	t.Logf("codectypes.NewAnyWithValue(approve).TypeUrl = %q", anyMsg.TypeUrl)

	// Control: a generated message has a real type URL.
	control := banktypes.NewMsgSend(
		sdktestutil.AccAddress(t), sdktestutil.AccAddress(t),
		sdk.NewCoins(sdk.NewCoin("uvve", math.NewInt(1))),
	)
	t.Logf("MsgTypeURL(bank MsgSend, control)      = %q", sdk.MsgTypeURL(control))
	require.NotEqual(t, "/", sdk.MsgTypeURL(control))

	// Proof 2: the x/gov submission-time signer gate cannot resolve it.
	signers, _, err := a.AppCodec().GetMsgV1Signers(approve)
	t.Logf("cdc.GetMsgV1Signers(approve) -> signers=%v err=%v", signers, err)
	require.Error(t, err)

	// Proof 3: no MsgServiceRouter handler.
	require.Nil(t, a.MsgServiceRouter().Handler(approve))
	t.Logf("MsgServiceRouter.Handler(approve) = %v", a.MsgServiceRouter().Handler(approve))

	// Proof 4: not resolvable in the interface registry either.
	_, err = a.InterfaceRegistry().Resolve("/")
	t.Logf("InterfaceRegistry.Resolve(\"/\") err=%v", err)

	// Proof 5 (STRONGEST): the bytes encode and then fail to decode.
	builder := a.TxConfig().NewTxBuilder()
	require.NoError(t, builder.SetMsgs(approve))
	rawTx, encErr := a.TxConfig().TxEncoder()(builder.GetTx())
	t.Logf("TxEncoder -> %d bytes, err=%v", len(rawTx), encErr)
	require.NoError(t, encErr)
	_, decErr := a.TxConfig().TxDecoder()(rawTx)
	t.Logf("TxDecoder -> err=%v", decErr)
	require.Error(t, decErr)

	// Decoy: legacy amino JSON encodes fine and buys nothing on the wire.
	aminoJSON, aminoErr := a.LegacyAmino().MarshalJSON(approve)
	t.Logf("LegacyAmino().MarshalJSON(approve) -> %s err=%v", string(aminoJSON), aminoErr)
}

func TestProbeTemplateSubsystemIsInert(t *testing.T) {
	a := app.Setup(app.WithChainID("virtengine-probe-2"))
	t.Cleanup(func() { _ = a.Close() })

	authority := sdktestutil.AccAddress(t).String()
	tpl := &hpctypes.WorkloadTemplate{TemplateID: "tmpl-1", Version: "1.0.0"}

	msgs := []sdk.Msg{
		hpctypes.NewMsgApproveWorkloadTemplate(authority, "tmpl-1", "1.0.0"),
		hpctypes.NewMsgRejectWorkloadTemplate(authority, "tmpl-1", "1.0.0", "no"),
		hpctypes.NewMsgDeprecateWorkloadTemplate(authority, "tmpl-1", "1.0.0", "old"),
		hpctypes.NewMsgRevokeWorkloadTemplate(authority, "tmpl-1", "1.0.0", "bad"),
		hpctypes.NewMsgSubmitJobFromTemplate(authority, "tmpl-1", "1.0.0", nil),
		hpctypes.NewMsgCreateWorkloadTemplate(authority, tpl),
		hpctypes.NewMsgUpdateWorkloadTemplate(authority, tpl),
	}

	for _, m := range msgs {
		h := a.MsgServiceRouter().Handler(m)
		t.Logf("%-30T router handler = %v", m, h)
		require.Nil(t, h, "%T must have no router handler", m)
		_, _, sErr := a.AppCodec().GetMsgV1Signers(m)
		require.Error(t, sErr, "%T must have no resolvable signers", m)
	}
}
```

### A note on where this probe must not live

`tests/integration/hpc/` is swept by `make test-integration` and by the CI `integration` job
(`.github/workflows/ci.yaml:856`), both of which run `-tags="e2e.integration"` over
`./tests/integration/...`. If the probe were committed there it would become a permanent green
test pinning the defect. Keep it as a materialized-then-deleted scratch file, as above.