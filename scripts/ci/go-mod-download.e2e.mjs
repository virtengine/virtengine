// End-to-end proof for the module-proxy TRANSPORT-retry card, using the REAL go
// toolchain and a REAL module proxy that kills the connection on the first zip
// request:
//   1. bare `go mod download`         -> fails (the card's red)
//   2. scripts/ci/go-mod-download.sh  -> passes (the fix)
// Both runs use a cold, per-run GOMODCACHE so step 1 genuinely fetches instead
// of reading a warm local cache.
//
// The proxy serves a SYNTHETIC module built in-process, so this proof needs no
// egress to proxy.golang.org and cannot be made unrunnable by an upstream proxy
// hiccup - which is the very class of fault under test.
//
// Run: node scripts/ci/go-mod-download.e2e.mjs
// Not wired into CI: it needs a real `go` on PATH and a loopback listener, which
// is fine here but heavier than a unit test. The CI-facing pins live in
// go-mod-download.test.mjs, which the quality-gate `vet` job runs.
import fs from "node:fs";
import http from "node:http";
import os from "node:os";
import path from "node:path";
import { spawn, spawnSync } from "node:child_process";

// `bash` on PATH on a Windows host is ambiguous: `where bash` lists MSYS's bash
// FIRST but WSL's bash.exe SECOND, and spawn resolves the bare name against
// whichever PATH order applies - which resolved WSL's launcher here and made it
// report "No such file or directory" for a C:/... path it cannot see. So do not
// trust a bare `bash`: resolve it, and prefer a Git-Bash/MSYS bash (a real one
// that understands the path form below) over the WSL launcher.
const resolveBash = () => {
  const probe = spawnSync(process.env.ComSpec ?? "cmd.exe", ["/c", "where", "bash"], {
    encoding: "utf8",
  });
  const candidates = (probe.stdout ?? "")
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
  const preferred = candidates.find((p) => /git|msys|usr[\\/]bin/i.test(p));
  const chosen = preferred ?? candidates[0];
  if (!chosen) throw new Error("no bash on PATH; cannot verify the wrapper");
  return chosen;
};

const bashExe = resolveBash();
const scriptUnderTest = path.resolve(import.meta.dirname, "go-mod-download.sh");

// All-lowercase module path and version, so Go's case-escaping (!x for an
// upper-case letter) never engages and the zip entry prefix is the literal path.
const MODULE = "example.com/synthetic";
const VERSION = "v1.0.0";
const PREFIX = `${MODULE}@${VERSION}`;

const GOMOD_TEXT = `module ${MODULE}\n\ngo 1.21\n`;
const GO_TEXT = "package synthetic\n\nfunc Value() int { return 42 }\n";

// --- minimal store-only zip writer -----------------------------------------
// archive/zip reads a STORE-only archive, so no deflate is needed. Node's stdlib
// has no zip writer, so build the three record types directly.
const CRC_TABLE = (() => {
  const table = new Int32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    table[n] = c;
  }
  return table;
})();

const crc32 = (buf) => {
  let c = -1;
  for (let i = 0; i < buf.length; i++) c = CRC_TABLE[(c ^ buf[i]) & 0xff] ^ (c >>> 8);
  return (c ^ -1) >>> 0;
};

// DOS date 0x0021 = 1980-01-01, the earliest the format can express.
const DOS_DATE = 0x0021;

const buildZip = (entries) => {
  const parts = [];
  const centrals = [];
  let offset = 0;

  for (const { name, data } of entries) {
    const nameBuf = Buffer.from(name, "utf8");
    const crc = crc32(data);

    const local = Buffer.alloc(30 + nameBuf.length);
    local.writeUInt32LE(0x04034b50, 0); // local file header signature
    local.writeUInt16LE(20, 4); // version needed to extract
    local.writeUInt16LE(0, 6); // general purpose flags
    local.writeUInt16LE(0, 8); // method: store
    local.writeUInt16LE(0, 10); // last mod time
    local.writeUInt16LE(DOS_DATE, 12); // last mod date
    local.writeUInt32LE(crc, 14);
    local.writeUInt32LE(data.length, 18); // compressed size
    local.writeUInt32LE(data.length, 22); // uncompressed size
    local.writeUInt16LE(nameBuf.length, 26);
    local.writeUInt16LE(0, 28); // extra field length
    nameBuf.copy(local, 30);
    parts.push(local, data);

    const central = Buffer.alloc(46 + nameBuf.length);
    central.writeUInt32LE(0x02014b50, 0); // central directory header signature
    central.writeUInt16LE(20, 4); // version made by
    central.writeUInt16LE(20, 6); // version needed to extract
    central.writeUInt16LE(0, 8); // general purpose flags
    central.writeUInt16LE(0, 10); // compression method
    central.writeUInt16LE(0, 12); // last mod time
    central.writeUInt16LE(DOS_DATE, 14); // last mod date
    central.writeUInt32LE(crc, 16);
    central.writeUInt32LE(data.length, 20); // compressed size
    central.writeUInt32LE(data.length, 24); // uncompressed size
    central.writeUInt16LE(nameBuf.length, 28);
    central.writeUInt16LE(0, 30); // extra field length
    central.writeUInt16LE(0, 32); // comment length
    central.writeUInt16LE(0, 34); // disk number start
    central.writeUInt16LE(0, 36); // internal attributes
    central.writeUInt32LE(0, 38); // external attributes
    central.writeUInt32LE(offset, 42); // relative offset of local header
    nameBuf.copy(central, 46);
    centrals.push(central);

    offset += local.length + data.length;
  }

  const centralBuf = Buffer.concat(centrals);
  const eocd = Buffer.alloc(22);
  eocd.writeUInt32LE(0x06054b50, 0); // end of central directory signature
  eocd.writeUInt16LE(0, 4); // this disk number
  eocd.writeUInt16LE(0, 6); // disk with the central directory
  eocd.writeUInt16LE(entries.length, 8); // entries on this disk
  eocd.writeUInt16LE(entries.length, 10); // total entries
  eocd.writeUInt32LE(centralBuf.length, 12); // size of the central directory
  eocd.writeUInt32LE(offset, 16); // offset of the central directory
  eocd.writeUInt16LE(0, 20); // comment length

  return Buffer.concat([...parts, centralBuf, eocd]);
};

const zipBuffer = buildZip([
  { name: `${PREFIX}/go.mod`, data: Buffer.from(GOMOD_TEXT, "utf8") },
  { name: `${PREFIX}/synthetic.go`, data: Buffer.from(GO_TEXT, "utf8") },
]);

const state = { resets: 0, zipRequests: 0, served: 0, modRequests: 0 };

// PRECISION ABOUT WHAT IS INJECTED: `http.createServer` is HTTP/1.1, so
// `res.socket.destroy()` mid-body is a TCP connection reset, NOT an HTTP/2
// RST_STREAM frame. It is a member of the same TRANSPORT-failure class the
// wrapper exists to absorb and lands on the same `go mod download` code path, so
// it is a valid end-to-end fault - but it does not reproduce the exact
// RST_STREAM spelling from run 36915824143. That spelling is pinned verbatim by
// the TRANSPORT_ERROR / RESET_ERRORS fixtures in go-mod-download.test.mjs, which
// drive the real script through the classifier. Do not read this log as proof of
// HTTP/2 framing; it is proof that a real cold-cache module fetch, aborted
// mid-stream, is retried into a pass.
const server = http.createServer((req, res) => {
  const url = req.url ?? "";

  if (url.endsWith(".mod")) {
    state.modRequests += 1;
    res.writeHead(200, { "content-type": "text/plain; charset=UTF-8" });
    res.end(GOMOD_TEXT);
    return;
  }

  if (url.endsWith(".info")) {
    res.writeHead(200, { "content-type": "application/json" });
    res.end(JSON.stringify({ Version: VERSION, Time: "2020-01-01T00:00:00Z" }));
    return;
  }

  if (url.endsWith(".zip")) {
    state.zipRequests += 1;
    if (state.resets === 0) {
      state.resets += 1;
      res.writeHead(200, { "content-type": "application/zip" });
      res.flushHeaders();
      res.socket.destroy();
      return;
    }
    state.served += 1;
    res.writeHead(200, { "content-type": "application/zip" });
    res.end(zipBuffer);
    return;
  }

  res.writeHead(404, { "content-type": "text/plain" });
  res.end("not found\n");
});

const listen = () =>
  new Promise((resolve) => server.listen(0, "127.0.0.1", () => resolve(server.address().port)));

// spawnSync does not apply Windows PATHEXT resolution, so a bare "go" fails with
// ENOENT (status null) instead of running. Resolve it here, and treat an
// unresolvable command as a HARNESS error rather than as a reproduced failure.
const resolveWindowsCommand = (command) => {
  if (process.platform !== "win32" || path.extname(command)) return command;
  const probe = spawnSync(process.env.ComSpec ?? "cmd.exe", ["/c", "where", command], {
    encoding: "utf8",
  });
  if (probe.status !== 0) return command;
  return probe.stdout.split(/\r?\n/)[0].trim();
};

// THE EVENT LOOP IS THE POINT. This proxy lives IN this process, so the child
// must be spawned ASYNCHRONOUSLY: spawnSync blocks the single JS thread, the
// listening socket is never serviced, and `go` waits forever for a response from
// a proxy that is structurally unable to answer. That deadlock is neither a red
// reproduction nor a product defect - it is this harness lying about which of
// the two it found. Hence spawn + a promise, with the socket free to serve
// throughout.
const runAsync = (command, args, { cwd, env, timeoutMs = 120_000 }) =>
  new Promise((resolve) => {
    const child = spawn(resolveWindowsCommand(command), args, {
      cwd,
      env,
      windowsHide: true,
      shell: false,
    });

    let stdout = "";
    let stderr = "";
    let timedOut = false;
    child.stdout.setEncoding("utf8");
    child.stderr.setEncoding("utf8");
    child.stdout.on("data", (chunk) => {
      stdout += chunk;
    });
    child.stderr.on("data", (chunk) => {
      stderr += chunk;
    });

    const timer = setTimeout(() => {
      timedOut = true;
      child.kill("SIGKILL");
    }, timeoutMs);

    child.on("error", (err) => {
      clearTimeout(timer);
      resolve({ status: null, signal: null, stdout, stderr, error: err, timedOut });
    });
    child.on("close", (code, signal) => {
      clearTimeout(timer);
      resolve({ status: code, signal, stdout, stderr, error: null, timedOut });
    });
  });

// Prefer this run's scratch dir over the system temp dir: under the Windows temp
// tree a cold GOMODCACHE intermittently returns EPERM on cleanup (an indexer or
// scanner holding a handle), which crashed the proof at cleanup time AFTER it had
// already printed a verdict - a failure that says nothing about the wrapper.
const scratchRoot = fs.mkdtempSync(
  path.join(process.env.TMPDIR && fs.existsSync(process.env.TMPDIR) ? process.env.TMPDIR : os.tmpdir(), "gomoddl-"),
);

let harnessFailure = false;

// Atomics.wait is a real synchronous sleep, unlike a busy loop - needed only in
// the cleanup path, where blocking is acceptable because the proxy is idle.
const sleepSync = (ms) => {
  Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, ms);
};

// Windows keeps transient locks on freshly-written trees; retry briefly rather
// than letting cleanup decide the verdict.
const removeTree = (target) => {
  for (let attempt = 0; attempt < 5; attempt++) {
    try {
      fs.rmSync(target, { recursive: true, force: true, maxRetries: 3, retryDelay: 200 });
      return true;
    } catch {
      sleepSync(300);
    }
  }
  console.log(`warning: could not remove ${target}; leaving it behind`);
  return false;
};

const runGo = async (label, commandArgs, extraEnv = {}) => {
  const cacheDir = fs.mkdtempSync(path.join(scratchRoot, "coldcache-"));
  const [command, ...rest] = commandArgs;
  const result = await runAsync(command, rest, {
    cwd: fixtureDir,
    env: {
      ...process.env,
      GOMODCACHE: cacheDir,
      GOFLAGS: "-mod=mod",
      GOWORK: "off",
      GOSUMDB: "off",
      GONOSUMDB: "*",
      GOTOOLCHAIN: "local", // the fixture declares go 1.21; never switch toolchains
      VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS: "1",
      ...extraEnv,
    },
  });
  removeTree(cacheDir);
  const output = `${result.stdout}${result.stderr}`;
  console.log(`--- ${label}: exit=${result.status}${result.timedOut ? " (TIMED OUT)" : ""}`);
  console.log(
    output
      .split("\n")
      .filter((line) => line.trim())
      .slice(0, 12)
      .join("\n"),
  );
  // A KILLED child (timeout, or a signal such as ENOENT on an unresolvable
  // command) never ran to a verdict at all. Counting that as a failure would make
  // this verifier manufacture a "the bare command reproduces the card's red"
  // claim out of a harness bug, which is the exact vacuous-pass shape PR #1137
  // exists to kill. Surface it loudly instead.
  if (result.status === null || result.timedOut) {
    console.log(
      `HARNESS ERROR: \`${command}\` produced no usable exit status (${
        result.timedOut ? "timed out" : result.signal ?? "no signal"
      }; ${result.error?.message ?? result.error?.code ?? "unknown"}). The fault was never exercised.`,
    );
    harnessFailure = true;
  }
  return { status: result.status, output };
};

const fixtureDir = fs.mkdtempSync(path.join(scratchRoot, "fixture-"));
fs.writeFileSync(path.join(fixtureDir, "go.mod"), GOMOD_TEXT);

const main = async () => {
  const port = await listen();
  const proxy = `http://127.0.0.1:${port}`;
  const target = `${MODULE}@${VERSION}`;

  console.log(`fault-injecting proxy on ${proxy}, module ${target}`);
  console.log(`scratch root ${scratchRoot}\n`);

  const bare = await runGo("bare `go mod download` (expect FAIL)", ["go", "mod", "download", target], {
    GOPROXY: proxy,
  });

  // A `null` status is a harness failure, not a red reproduction: require a REAL
  // non-zero exit before claiming the fault was reproduced.
  const bareRan = bare.status !== null;
  const bareFailed = bareRan && bare.status !== 0;
  console.log(
    bareFailed
      ? `\nOK: the bare command reproduces the card's red (exit ${bare.status}).`
      : `\nPROBLEM: the bare command did NOT reproduce a failure - bareRan=${bareRan} status=${bare.status}. The fault was never exercised.`,
  );

  state.resets = 0;
  state.zipRequests = 0;
  state.served = 0;
  state.modRequests = 0;

  const wrapped = await runGo("scripts/ci/go-mod-download.sh (expect PASS)", [bashExe, scriptUnderTest, target], {
    GOPROXY: proxy,
  });

  console.log(
    `\nzip requests=${state.zipRequests} served=${state.served} resets injected=${state.resets} mod requests=${state.modRequests}`,
  );

  // The wrapper's output must SHOW the retry, or a pass could come from a warm
  // cache or a proxy that simply never served the reset - both would make this a
  // green that proves nothing.
  const showsRetry = /transport error/i.test(wrapped.output);
  // And the fault must actually have been injected on THIS run.
  const retried = state.resets === 1 && state.zipRequests >= 2 && state.served >= 1;
  const passed = wrapped.status === 0;
  const ok = !harnessFailure && bareFailed && passed && retried && showsRetry;

  console.log(
    ok
      ? "RESULT: PASS - cold cache + one injected mid-stream reset: bare go fails, the wrapper retries and passes."
      : `RESULT: FAIL - harnessFailure=${harnessFailure} bareFailed=${bareFailed} passed=${passed} retried=${retried} showsRetry=${showsRetry}`,
  );
  removeTree(fixtureDir);
  removeTree(scratchRoot);
  server.close();
  process.exit(ok ? 0 : 1);
};

await main();
