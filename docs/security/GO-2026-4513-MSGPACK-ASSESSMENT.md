# GO-2026-4513 — `shamaton/msgpack` out-of-bounds read: reachability assessment and accepted risk

**Advisory:** GO-2026-4513 · CVE-2026-32284
**Severity:** High as published (the Go vulnerability DB carries no CVSS vector; the sibling
GHSA-h9q6-hc68-35rp for the same decoder is `AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H`)
**Affected module:** `github.com/shamaton/msgpack` — the OSV record lists three affected packages
(`github.com/shamaton/msgpack`, `/v2`, `/v3`); this repository pins `/v2`
**Vulnerable symbols:** `github.com/shamaton/msgpack/v2`: `Unmarshal`, `UnmarshalAsArray`,
`UnmarshalAsMap` (`/v3` adds `DecodeStructAsArray`, `DecodeStructAsMap`)
**Patched versions:** none — the OSV range is a single `{introduced: "0"}` event with **no `fixed`
event**, on all three module paths, so no version bump clears it
**Assessed on:** `develop` @ `336ce39b`, 2026-09-29
**Toolchain:** Go 1.26.8, `govulncheck v1.1.4`, `GOWORK=off`
**Verdict:** accepted, time-boxed residual risk. This is the **same call path as
[GO-2026-4740](GO-2026-4740-MSGPACK-ASSESSMENT.md)** — different vulnerable symbols in the same
decoder — so the reachability argument and the compensating controls are identical. Tracked in
[issue #872](https://github.com/virtengine/virtengine/issues/872) and
`.vulnerability-allowlist.yaml`.

---

## 1. What the advisory says

The msgpack decoder does not validate the input buffer length when it processes truncated `fixext`
data (format codes `0xd4`–`0xd8`), producing an out-of-bounds read and a runtime panic — a denial
of service on hostile msgpack input. It affects the `Unmarshal` family in `v2` and `v3`, and the
older `github.com/shamaton/msgpack` path as well.

There is no version to move to: the OSV record for `GO-2026-4513` carries
`{"type": "SEMVER", "events": [{"introduced": "0"}]}` for each of the three module paths and no
`fixed` event anywhere.

## 2. Dependency route

Identical to GO-2026-4740, and documented in full there:

```
cmd/virtengine/cmd → sdk/go/cli → wasmd/x/wasm/types → wasmvm/v3/types → shamaton/msgpack/v2
```

`shamaton/msgpack` is declared `// indirect` in the root `go.mod`, and a repository-wide search of
`*.go` finds **no first-party file** importing it.

## 3. Why the vulnerable symbols are on no reachable call path

### 3.1 The only call site in the entire module graph

```
$ grep -rn 'msgpack\.' "$(go env GOMODCACHE)"/github.com/!cosm!wasm/wasmvm/v3@*/ --include=*.go | grep -v _test
.../wasmvm/v3@v3.0.2/types/types.go:215:	return msgpack.UnmarshalAsArray(data, pm)
```

One line, inside `PinnedMetrics.UnmarshalMessagePack`, reached only via
`wasmvm/internal/api.GetPinnedMetrics` → `VM.GetPinnedMetrics` →
`wasmd/x/wasm/keeper.WasmVMMetricsCollector.Collect`.

### 3.2 That call site's package is not in this repository's build

```
$ go list -deps ./... | grep -cE 'wasmd/x/wasm/keeper|wasmvm/v3/internal/api'
0
```

VirtEngine consumes wasmd only through `x/wasm/types` and `x/wasm/ioutils` (message and query
types for the SDK/CLI). It mounts no wasm module, never instantiates a `wasmvm.VM`, and never
constructs `WasmVMMetricsCollector`, so the Prometheus collector that owns the single `msgpack`
call site has no entry point here. The `Unmarshal*` functions the advisory names are therefore not
reachable from any shipped binary.

### 3.3 The decoded bytes are not attacker-controlled even in principle

The one call site decodes a buffer produced **in-process** by libwasmvm's Rust
`get_pinned_metrics` over FFI. A truncated-`fixext` payload crafted by a remote party has no route
into that buffer. Defence in depth; §3.2 is the primary argument.

### 3.4 Why `govulncheck` still reports it

This advisory **does** name symbols, so `govulncheck` reports it at package granularity rather than
module granularity: because `wasmvm/v3/types` imports msgpack, the msgpack packages are linked and
their `init` functions run, which the tool treats as reachability. That is a statement about
linkage, not about the vulnerable call path — the distinction is argued in §3 of the GO-2026-4740
assessment.

## 4. Decision

**Accepted residual risk, time-boxed**, under the same 30-day policy and with the same compensating
controls as GO-2026-4740:

1. `wasmd/x/wasm/keeper` and `wasmvm/v3/internal/api` are absent from `go list -deps ./...`; the
   single msgpack call site has no entry point in this repository.
2. VirtEngine mounts no CosmWasm module and never instantiates a wasmvm VM, so no network-supplied
   msgpack is decoded anywhere.
3. Even if the call site were wired up, it decodes in-process bytes produced by libwasmvm rather
   than attacker-controlled input.
4. Time-boxed to 30 days; `validate_security_policies.py` fails `Policy Validation` when the entry
   expires or exceeds `policy.max_allowlist_age_days`.

## 5. Review trigger

Re-assess on or before the allowlist expiry (**2026-10-28**), or immediately if:

- wasmvm or wasmd change their `shamaton/msgpack` pin, or publish a release that drops it;
- VirtEngine mounts the wasm module, instantiates a `wasmvm.VM`, or registers
  `WasmVMMetricsCollector`;
- a patched `shamaton/msgpack` release appears (then prefer the bump and delete this exception).

## 6. Reproducing this assessment

```bash
export GOWORK=off

grep -rn 'shamaton/msgpack' --include=*.go . | grep -v /vendor/          # no first-party file
go list -deps ./... | grep -cE 'wasmd/x/wasm/keeper|wasmvm/v3/internal/api'   # 0
grep -rn 'msgpack\.' "$(go env GOMODCACHE)"/github.com/!cosm!wasm/wasmvm/v3@*/ --include=*.go \
  | grep -v _test                                                       # one call site
govulncheck -show traces ./pkg/... 2>&1 | grep -A 20 'GO-2026-4513'
```

If the `go list -deps` step ever returns a non-zero count, this assessment is void and the
exception must be withdrawn.
