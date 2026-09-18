# GO-2026-4740 — `shamaton/msgpack` DoS: reachability assessment and accepted risk

**Advisory:** GO-2026-4740 · CVE-2026-32284 · GHSA-h9q6-hc68-35rp
**Severity:** High (CVSS 3.1 7.5, `AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H`)
**Affected module:** `github.com/shamaton/msgpack/v2` (also `/v3`)
**Patched versions:** none — the Go vuln DB records `Fixed in: N/A`, and the OSV record carries a
single `{introduced: "0"}` range with no `fixed` event.
**Assessed on:** `main` @ `63349afe071bbba878ccd36a5c8e096117e970e4`, 2026-09-19
**Toolchain:** Go 1.25.8, `govulncheck v1.1.4`
**Verdict:** the vulnerable decode path is **not reachable from any shipped binary** in this repo.
Time-boxed residual risk accepted; tracked in
[issue #872](https://github.com/virtengine/virtengine/issues/872) and
`.vulnerability-allowlist.yaml`.

---

## 1. What the advisory says

The msgpack decoder fails to validate the input buffer length when processing truncated `fixext`
data (format codes `0xd4`–`0xd8`), leading to an out-of-bounds read and a runtime panic — a denial
of service triggered by hostile msgpack input.

The finding cannot be cleared by a version bump. This was tested, not assumed:

```
go get github.com/shamaton/msgpack/v2@v2.4.2     # newest v2 release
go mod tidy
govulncheck -show traces ./...
```

still reports GO-2026-4740 against `v2.4.2` with `Fixed in: N/A`. Upstream `github.com/CosmWasm/wasmd`
`v0.61.14` (newest 0.61.x) and `github.com/CosmWasm/wasmvm/v3` `v3.0.7` (newest) **both still require
`shamaton/msgpack/v2`**, so upgrading the CosmWasm stack does not drop or replace it either. There is
no released version of this module, in any major line, that is marked fixed.

## 2. Dependency route

`go mod why -m github.com/shamaton/msgpack/v2`:

```
github.com/virtengine/virtengine/cmd/virtengine/cmd
github.com/virtengine/virtengine/sdk/go/cli
github.com/CosmWasm/wasmd/x/wasm/types
github.com/CosmWasm/wasmvm/v3/types
github.com/shamaton/msgpack/v2
```

Both `go.mod` files declare it `// indirect`.

## 3. Why it is unreachable

### 3.1 No VirtEngine-owned code imports it

A repository-wide search of `*.go` for `shamaton/msgpack` and `msgpack/v2` returns **zero** files.
The module enters the build only as a transitive dependency.

### 3.2 Exactly one call site exists in the entire module graph

In wasmvm `v3.0.2`, every reference to the msgpack package in non-test Go code is one line —
`types/types.go:215`:

```go
func (pm *PinnedMetrics) UnmarshalMessagePack(data []byte) error {
	return msgpack.UnmarshalAsArray(data, pm)
}
```

### 3.3 That call site sits behind a package that is not in this repo's build

The only callers, verified by grep over the module cache, are:

```
wasmvm/internal/api/lib.go:195   GetPinnedMetrics(cache Cache)          // builds the value
wasmvm/internal/api/lib.go:203   pinnedMetrics.UnmarshalMessagePack(..) // the msgpack call
wasmvm/lib_libwasmvm.go:143      func (vm *VM) GetPinnedMetrics()
wasmd/x/wasm/keeper/metrics.go:76  p.source.GetPinnedMetrics()          // inside Collect()
```

`go list -deps ./...` in this repository contains **neither**
`github.com/CosmWasm/wasmd/x/wasm/keeper` **nor** `github.com/CosmWasm/wasmvm/v3/internal/api`.
VirtEngine does not mount the wasm module in `app/`, never creates a VM, and has no reference to
`WasmVMMetricsCollector` / `NewWasmVMMetricsCollector` anywhere. The type that owns `Collect()` is a
Prometheus collector that is never constructed, so the chain has no entry point in this codebase.

The repo consumes wasmd only through `x/wasm/types` and `x/wasm/ioutils` (SDK/CLI message and query
types, e.g. `sdk/go/cli/wasm_tx.go`).

### 3.4 `govulncheck`'s own traces agree

`govulncheck -show traces ./...` reports GO-2026-4740 with 14 example traces, and **every one**
terminates in an `init` or package-variable initialisation symbol:

```
#1: for function github.com/shamaton/msgpack/v2/internal/common.init
#3: for function github.com/shamaton/msgpack/v2/internal/decoding.init
#11: for function github.com/shamaton/msgpack/v2/time.timeDecoder.Code
#12: for function github.com/shamaton/msgpack/v2/time.timeEncoder.Type
...
init @ github.com/CosmWasm/wasmd/x/wasm/types/errors.go:4:2
init @ github.com/CosmWasm/wasmvm/v3/types/types.go:8:2
```

No decode function (`UnmarshalAsArray`, `Decode`, `Unmarshal`, …) appears on any reachable path.

**Why the scan still fails.** The OSV record for GO-2026-4740 has an empty
`ecosystem_specific: {}` — the advisory names no vulnerable symbols. `govulncheck` therefore reports
the finding at *package* granularity: because `wasmvm/v3/types` imports msgpack, the msgpack packages
are linked and their `init` functions run, so the package is considered reachable. That is a
statement about linkage, not about the vulnerable call path. This is the specific reason the finding
cannot be reasoned away by the scanner and has to be closed by a documented human decision.

### 3.5 The decoded input is not attacker-controlled even in principle

The single call site decodes bytes produced **in-process** by libwasmvm's Rust
`get_pinned_metrics` over FFI — the VM serialising its own pinned-contract metrics for a monitoring
endpoint. A truncated-`fixext` payload crafted by a remote attacker has no route into that buffer.
This is a defence-in-depth observation; the primary argument is §3.3 (no entry point).

## 4. Decision

**Accepted residual risk, time-boxed.** Recorded in `.vulnerability-allowlist.yaml` with a 30-day
expiry under the repository's existing policy (`max_allowlist_age_days: 30`,
`require_compensating_controls: true`).

Compensating controls:

1. The vulnerable decode path has no call site in the shipped dependency graph (proved with
   `go list -deps ./...`); the owner package `wasmd/x/wasm/keeper` is absent.
2. VirtEngine does not mount the CosmWasm module or instantiate a VM, so no wasm contract input —
   the only realistic source of hostile msgpack — is processed.
3. The one remaining call site consumes in-process, trusted bytes from libwasmvm.
4. The exception is time-boxed to 30 days and re-reviewed at PR time; expiry is enforced by
   `.github/scripts/validate_security_policies.py`, which fails the `Policy Validation` job if the
   entry lapses or the age limit is exceeded.

**This entry does not turn `Go Vulnerability Scan` green.** `.vulnerability-allowlist.yaml` is
validated but is not consumed by any scanner in this repository. That job also fails on unrelated
stdlib findings that require a Go toolchain of 1.25.13 or newer, and on the `golang.org/x/crypto`
v0.56.0 work. Those are tracked separately.

## 5. Review trigger

Re-assess on or before the allowlist expiry (**2026-10-19**), or immediately if any of these change:

- wasmvm or wasmd change their `shamaton/msgpack` pin, or publish a release that drops it;
- VirtEngine mounts the wasm module, instantiates a `wasmvm.VM`, or registers
  `WasmVMMetricsCollector`;
- a patched `shamaton/msgpack` release appears (then prefer the bump and delete this exception).

## 6. Reproducing this assessment

```bash
# 1. Confirm the module is indirect and find its route
grep -n 'shamaton/msgpack' go.mod
go mod why -m github.com/shamaton/msgpack/v2

# 2. Confirm no shipped package reaches the decode call site
go list -deps ./... | grep -E 'wasmd/x/wasm/keeper|wasmvm/v3/internal/api'   # no output

# 3. Confirm the only msgpack call site in wasmvm
grep -rn 'msgpack\.' "$(go env GOMODCACHE)"/github.com/!cosm!wasm/wasmvm/v3@*/ --include=*.go

# 4. Confirm the traces are init-only
govulncheck -show traces ./... 2>&1 | grep -A 20 'GO-2026-4740'
```

If step 2 ever produces output, this assessment is void and the exception must be withdrawn.
