#!/usr/bin/env python3
"""Unit-check audit_node_runtime's error classifier and manifest probe order.

The retry logic is only correct if it can tell three things apart:
  404            -> this layout does not exist, move on to the next candidate
  ratelimit      -> the call did not happen, retry with backoff
  success        -> stop

Getting this wrong in either direction is a real defect: classify a 404 as
retriable and every 404-bearing pin burns 4 sleeps; classify a rate limit as
404 and the audit reports "no such manifest" for a pin that plainly exists.

These run against the SHIPPED functions by importing the module, so they test
the code that runs, not a copy of it.
"""

from __future__ import annotations

import importlib.util
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
AUDIT = os.path.join(HERE, "audit_node_runtime.py")

spec = importlib.util.spec_from_file_location("audit_node_runtime", AUDIT)
mod = importlib.util.module_from_spec(spec)
assert spec.loader is not None
# Register BEFORE exec: @dataclass resolves its own module out of sys.modules,
# and an unregistered module makes the decorator blow up on first use.
sys.modules["audit_node_runtime"] = mod
spec.loader.exec_module(mod)


class FakeProc:
    def __init__(self, rc, out="", err=""):
        self.returncode = rc
        self.stdout = out
        self.stderr = err


def fail(rc, err):
    """What `subprocess.run(check=True)` raises for a non-zero exit.

    The real subprocess.run raises internally; a stub that only sets
    returncode does not, so the audit would decode the empty stdout as a
    SUCCESS and every failure case would silently test nothing. The stub has to
    raise, or the harness is decoration.
    """
    return subprocess.CalledProcessError(rc, ["gh", "api"], output="", stderr=err)


def patch_gh(handler):
    """Swap _gh_contents' subprocess call for a scripted responder.

    The handler may RETURN a failure (a `fail(...)` exception object) to mean
    "this call failed". The stub then RAISES it, because that is what
    `subprocess.run(check=True)` does. Returning it instead would make every
    failed call look like a success carrying empty stdout, and the whole harness
    would pass while testing nothing.

    `**kwargs` is forwarded to the handler, not swallowed: the audit passes
    `env=` on its anonymous read, and a handler that cannot see it cannot tell
    an authenticated call from an anonymous one.
    """
    calls = []

    def fake_run(cmd, **kwargs):
        calls.append(" ".join(cmd) if isinstance(cmd, (list, tuple)) else str(cmd))
        result = handler(cmd, **kwargs)
        if isinstance(result, BaseException):
            raise result
        return result

    real = subprocess.run
    mod.subprocess.run = fake_run
    return calls, lambda: setattr(mod.subprocess, "run", real)


def main() -> int:
    failures: list[str] = []
    mod.time.sleep = lambda *_: None  # keep the suite fast

    # --- hermetic network guard -------------------------------------------
    # A 403 from the authenticated read now triggers a REAL anonymous HTTPS
    # read, so without this every forbidden-classification case would reach
    # out to api.github.com. Worse, a fabricated repo like "o/r" answers 404
    # there, so the suite would assert against whatever the live API said
    # rather than against the classifier. The default below refuses the
    # anonymous read with 403 -- "still forbidden" -- which is the answer that
    # preserves the intent of the pre-existing cases. Individual cases
    # override it; the one that must reach a URL asserts on the URL instead.
    real_urlopen = mod.urllib.request.urlopen

    def default_refuse(req, timeout=None):
        raise mod.urllib.error.HTTPError(req.full_url, 403, "Forbidden", {}, None)

    mod.urllib.request.urlopen = default_refuse

    # --- classification of real GitHub error text -------------------------
    cases = [
        ("gh: Not Found (HTTP 404)", "404"),
        ('{"message":"Not Found","status":"404"}', "404"),
        ("API rate limit exceeded for user", "ratelimit"),
        ("HTTP 403: Resource not accessible by integration", "forbidden"),
        ("HTTP 500: internal error", "api-error"),
    ]
    for text, expected in cases:
        mod.subprocess.run = subprocess.run
        def handler(cmd, _t=text, **kw):
            raise subprocess.CalledProcessError(1, cmd, output="", stderr=_t)
        _, restore = patch_gh(handler)
        try:
            _, err = mod._gh_contents("o/r", "action.yml", "v1")
        finally:
            restore()
        if err != expected:
            failures.append(f"classify({text[:34]!r}) -> {err!r}, expected {expected!r}")
        print(f"  {'FAIL' if err != expected else 'ok  '} classify {text[:40]!r} -> {err}")

    # --- probe order: subdir, then action.yml, then action.yaml -----------
    seen: list[str] = []
    def handler2(cmd, **kw):
        seen.append(" ".join(cmd) if isinstance(cmd, (list, tuple)) else str(cmd))
        # The path is argv[2] (`gh api <path> --jq <filter>`), NOT the last
        # element -- argv[-1] is the --jq filter, so matching on it tests
        # nothing and every probe looked like a miss.
        if "action.yaml" in cmd[2]:
            return FakeProc(0, out="aW5wdXRzOjpjb21wb3NpdGUK")  # b64 "inputs::composite"
        return fail(1, "gh: Not Found (HTTP 404)")
    _, restore = patch_gh(handler2)
    try:
        mod._FETCH_CACHE.clear(); mod._READ_SEEN.clear(); mod._FETCH_ERRORS.clear()
        content = mod.fetch_action_yml("github/codeql-action/init", "v4")
    finally:
        restore()
    tried = [c.split("/contents/")[1].split("?")[0] for c in seen]
    if content != "inputs::composite\n":
        failures.append(f"probe-order: expected the action.yaml hit, got {content!r}")
    if tried[:2] != ["init/action.yml", "init/action.yaml"]:
        failures.append(f"probe-order: wrong candidate order {tried}")
    print(f"  {'FAIL' if failures else 'ok  '} probe-order tried {tried[:3]}")

    # --- a 404 pin is terminal, not retried --------------------------------
    seen2: list[str] = []
    def handler3(cmd, **kw):
        seen2.append(" ".join(cmd) if isinstance(cmd, (list, tuple)) else str(cmd))
        return fail(1, "gh: Not Found (HTTP 404)")
    _, restore = patch_gh(handler3)
    try:
        mod._FETCH_CACHE.clear(); mod._READ_SEEN.clear(); mod._FETCH_ERRORS.clear()
        content = mod.fetch_action_yml("virtengine/nope", "v1")
    finally:
        restore()
    # 2 candidates (action.yml, action.yaml) x 1 attempt = 2 calls, not 8.
    if len(seen2) != 2:
        failures.append(f"404-retries: expected 2 calls (terminal), made {len(seen2)}")
    if content is not None:
        failures.append(f"404-retries: expected None, got {content!r}")
    print(f"  {'FAIL' if len(seen2) != 2 else 'ok  '} 404-is-terminal ({len(seen2)} calls)")

    # --- a rate-limited pin IS retried -------------------------------------
    state = {"n": 0}
    def handler4(cmd, **kw):
        state["n"] += 1
        if state["n"] < 3:
            return fail(1, "API rate limit exceeded")
        return FakeProc(0, out="aW5wdXRzOjpjb21wb3NpdGUK")
    _, restore = patch_gh(handler4)
    try:
        mod._FETCH_CACHE.clear(); mod._READ_SEEN.clear(); mod._FETCH_ERRORS.clear()
        content = mod.fetch_action_yml("o/r", "v1")
    finally:
        restore()
    if content != "inputs::composite\n":
        failures.append(f"ratelimit-retries: expected a recovered read, got {content!r} after {state['n']} calls")
    print(f"  {'FAIL' if content != 'inputs::composite\n' else 'ok  '} ratelimit-is-retried (recovered on call {state['n']})")

    # --- a permission failure is NOT "no such manifest" ---------------------
    # This is the regression the third CI run exposed: the ranking compared a
    # class's attempt budget against the sweep number, so `forbidden` (budget
    # 1) seen on attempt 1 never displaced the "404" default and the audit
    # confidently reported a repo as having no manifest when it simply was not
    # readable. Stating a confident falsehood is worse than staying silent.
    def handler5(cmd, **kw):
        return fail(1, "HTTP 403: Resource not accessible by integration")

    _, restore = patch_gh(handler5)
    try:
        mod._FETCH_CACHE.clear(); mod._READ_SEEN.clear(); mod._FETCH_ERRORS.clear()
        content = mod.fetch_action_yml("o/r", "v1")
    finally:
        restore()
    ok = content == mod.TRANSIENT_FAILURE and mod._FETCH_ERRORS.get(("o/r", "v1", "manifest")) == "forbidden"
    if not ok:
        failures.append(
            f"forbidden-is-not-a-404: expected TRANSIENT/forbidden, got "
            f"{content!r} class={mod._FETCH_ERRORS.get(('o/r', 'v1', 'manifest'))!r}"
        )
    print(f"  {'FAIL' if not ok else 'ok  '} forbidden-is-not-a-404")

    # --- a 403 on a PUBLIC manifest is retried ANONYMOUSLY ----------------
    # This is the defect the fourth CI run exposed. The CI job passes the
    # job's own `github.token`, which is not scoped to other repositories, so
    # GitHub answers 403 "Resource not accessible by integration" for pins
    # that are plainly public. The audit believed that verdict, called itself
    # INCOMPLETE, and failed the build on a tree it had already proven clean
    # on all 38 other pins.
    #
    # The retry must NOT go through `gh`. An earlier version re-ran `gh` with
    # the token stripped from the environment, and CI moved the 5 pins from
    # `forbidden` to `api-error` instead of clearing them: on a runner `gh`
    # refuses an unauthenticated API call outright, so the "fallback" was a
    # guaranteed failure. It passed every local test because a workstation's
    # `gh` has a logged-in account to fall back on. So the assertion is on the
    # NETWORK path, not on a second subprocess call.
    anon_urls: list[str] = []
    gh_calls: list[str] = []

    class FakeResp:
        def __init__(self, body: bytes):
            self._body = body

        def read(self):
            return self._body

        def __enter__(self):
            return self

        def __exit__(self, *a):
            return False

    payload = json.dumps({"content": "aW5wdXRzOjpjb21wb3NpdGUK"}).encode()
    mod.urllib.request.urlopen = lambda req, timeout=None: (
        anon_urls.append(req.full_url) or FakeResp(payload)
    )

    def handler6(cmd, **kw):
        gh_calls.append(" ".join(cmd) if isinstance(cmd, (list, tuple)) else str(cmd))
        return fail(1, "HTTP 403: Resource not accessible by integration")

    _, restore = patch_gh(handler6)
    try:
        mod._FETCH_CACHE.clear(); mod._READ_SEEN.clear(); mod._FETCH_ERRORS.clear()
        content = mod._gh_contents("o/r", "action.yml", "v1")
    finally:
        restore()
        mod.urllib.request.urlopen = default_refuse

    if content != ("inputs::composite\n", ""):
        failures.append(f"anon-fallback: expected an anonymous read to succeed, got {content!r}")
    if len(gh_calls) != 1:
        failures.append(f"anon-fallback: expected exactly 1 authenticated gh call, made {len(gh_calls)}")
    if len(anon_urls) != 1:
        failures.append(f"anon-fallback: expected 1 anonymous HTTPS read, made {len(anon_urls)}")
    elif "repos/o/r/contents/action.yml?ref=v1" not in anon_urls[0]:
        failures.append(f"anon-fallback: anonymous URL is wrong: {anon_urls[0]!r}")
    anon_ok = content == ("inputs::composite\n", "") and len(anon_urls) == 1
    print(f"  {'ok  ' if anon_ok else 'FAIL'} "
          f"anon-fallback (auth 403 -> anonymous HTTPS read)")

    # --- an HTTP status maps to the same classes as gh's stderr -----------
    http_cases = [(404, "404"), (429, "ratelimit"), (403, "forbidden"), (500, "api-error")]
    for status, expected in http_cases:
        got = mod._classify_http(status)
        if got != expected:
            failures.append(f"classify_http({status}) -> {got!r}, expected {expected!r}")
    http_ok = all(mod._classify_http(s) == e for s, e in http_cases)
    print(f"  {'ok  ' if http_ok else 'FAIL'} classify_http {[s for s, _ in http_cases]}")

    # --- a 403 that survives the anonymous read is still a real failure ---
    def handler7(cmd, **kw):
        return fail(1, "HTTP 403: Resource not accessible by integration")

    def refuse(req, timeout=None):
        raise mod.urllib.error.HTTPError(req.full_url, 403, "Forbidden", {}, None)

    mod.urllib.request.urlopen = refuse
    _, restore = patch_gh(handler7)
    try:
        mod._FETCH_CACHE.clear(); mod._READ_SEEN.clear(); mod._FETCH_ERRORS.clear()
        content = mod.fetch_action_yml("o/r", "v1")
    finally:
        restore()
        mod.urllib.request.urlopen = default_refuse
    ok = content == mod.TRANSIENT_FAILURE and mod._FETCH_ERRORS.get(("o/r", "v1", "manifest")) == "forbidden"
    if not ok:
        failures.append(
            f"anon-fallback-hard-fail: expected TRANSIENT/forbidden when the "
            f"anonymous read is refused too, got {content!r} "
            f"class={mod._FETCH_ERRORS.get(('o/r', 'v1', 'manifest'))!r}"
        )
    print(f"  {'FAIL' if not ok else 'ok  '} anon-fallback-hard-fail")

    mod.urllib.request.urlopen = real_urlopen
    for f in failures:
        print("FAIL:", f)
    total = 9
    print(f"{total - len({f.split(':')[0] for f in failures})}/{total} checks passed")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
