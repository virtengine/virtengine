/**
 * Falsification for scripts/ci/check-portal-export-routes.mjs.
 *
 * A census that only reports the 15 routes already broken on develop would be
 * indistinguishable from a hardcoded list. These assertions prove the gate
 * reacts to a route ADDED after the fact, to a route that opts back in, and to
 * an emptied route tree — i.e. it reads the tree, it does not memorise it.
 *
 * Run: node scripts/ci/check-portal-export-routes.test.mjs
 */
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const checker = path.join(repoRoot, 'scripts', 'ci', 'check-portal-export-routes.mjs');
const realApiDir = path.join(repoRoot, 'portal', 'src', 'app', 'api');

function run(apiDir) {
  // The checker resolves the portal tree relative to its own location, so a
  // fixture must be a whole throwaway repo, not just a copied api/ dir.
  const fixtureRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'export-routes-'));
  const scriptsDir = path.join(fixtureRoot, 'scripts', 'ci');
  fs.mkdirSync(scriptsDir, { recursive: true });
  fs.copyFileSync(checker, path.join(scriptsDir, 'check-portal-export-routes.mjs'));
  fs.cpSync(apiDir, path.join(fixtureRoot, 'portal', 'src', 'app', 'api'), { recursive: true });
  try {
    return {
      status: execFileSync(process.execPath, [path.join(scriptsDir, 'check-portal-export-routes.mjs')], {
        encoding: 'utf8',
        stdio: ['ignore', 'pipe', 'pipe'],
      }),
      code: 0,
    };
  } catch (err) {
    return { status: `${err.stdout ?? ''}`, code: err.status ?? 1 };
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

let failures = 0;
function check(name, fn) {
  try {
    fn();
    console.log(`ok   ${name}`);
  } catch (err) {
    failures += 1;
    console.error(`FAIL ${name}\n     ${err.message.split('\n')[0]}`);
  }
}

// 1. The real tree is RED today: 16 routes, 15 blocking.
check('live tree is red with 15 blocking routes', () => {
  const { status, code } = run(realApiDir);
  assert.equal(code, 1, `expected exit 1, got ${code}`);
  assert.match(status, /16 route\(s\) total, 15 block a static export/);
});

// 2. A route added with no declaration is counted -- the census is not a
//    snapshot of today's routes. Adding one must take the count 15 -> 16.
check('a newly added dynamic route is counted', () => {
  const fixture = fs.mkdtempSync(path.join(os.tmpdir(), 'export-routes-add-'));
  fs.cpSync(realApiDir, fixture, { recursive: true });
  fs.mkdirSync(path.join(fixture, 'brand', 'new'), { recursive: true });
  fs.writeFileSync(
    path.join(fixture, 'brand', 'new', 'route.ts'),
    "import { NextResponse } from 'next/server';\nexport async function GET() { return NextResponse.json({}); }\n"
  );
  const { status, code } = run(fixture);
  fs.rmSync(fixture, { recursive: true, force: true });
  assert.equal(code, 1);
  assert.match(status, /BLOCK \/brand\/new/);
  assert.match(status, /17 route\(s\) total, 16 block a static export/);
});

// 3. A route that opts IN with dynamic='force-static' is not counted.
check('a force-static route is not counted as blocking', () => {
  const fixture = fs.mkdtempSync(path.join(os.tmpdir(), 'export-routes-optin-'));
  fs.cpSync(realApiDir, fixture, { recursive: true });
  fs.mkdirSync(path.join(fixture, 'brand', 'static'), { recursive: true });
  fs.writeFileSync(
    path.join(fixture, 'brand', 'static', 'route.ts'),
    "export const dynamic = 'force-static';\nexport const dynamicParams = false;\nexport async function GET() { return Response.json({}); }\n"
  );
  const { status, code } = run(fixture);
  fs.rmSync(fixture, { recursive: true, force: true });
  assert.equal(code, 1);
  assert.match(status, /OK   \/brand\/static/);
  assert.doesNotMatch(status, /BLOCK \/brand\/static/);
});

// 4. `revalidate` alone also satisfies the gate.
check('a revalidate declaration is accepted as exportable', () => {
  const fixture = fs.mkdtempSync(path.join(os.tmpdir(), 'export-routes-reval-'));
  fs.cpSync(realApiDir, fixture, { recursive: true });
  fs.mkdirSync(path.join(fixture, 'brand', 'revalidated'), { recursive: true });
  fs.writeFileSync(
    path.join(fixture, 'brand', 'revalidated', 'route.ts'),
    "export const revalidate = 3600;\nexport async function GET() { return Response.json({}); }\n"
  );
  const { status } = run(fixture);
  fs.rmSync(fixture, { recursive: true, force: true });
  assert.match(status, /OK   \/brand\/revalidated/);
});

// 5. An EMPTY tree is not a pass -- it would mean the census silently stopped
//    finding routes, which is the failure direction that matters.
check('an empty api tree fails closed rather than passing vacuously', () => {
  const fixture = fs.mkdtempSync(path.join(os.tmpdir(), 'export-routes-empty-'));
  fs.mkdirSync(fixture, { recursive: true });
  let code = 0;
  let status = '';
  try {
    const scriptsDir = path.join(fixture, 'scripts', 'ci');
    fs.mkdirSync(scriptsDir, { recursive: true });
    fs.copyFileSync(checker, path.join(scriptsDir, 'check-portal-export-routes.mjs'));
    fs.mkdirSync(path.join(fixture, 'portal', 'src', 'app', 'api'), { recursive: true });
    const out = execFileSync(process.execPath, [path.join(scriptsDir, 'check-portal-export-routes.mjs')], {
      encoding: 'utf8',
    });
    status = out;
  } catch (err) {
    code = err.status ?? 1;
    status = `${err.stdout ?? ''}${err.stderr ?? ''}`;
  } finally {
    fs.rmSync(fixture, { recursive: true, force: true });
  }
  assert.equal(code, 1, `empty tree must fail closed, got ${code}`);
  assert.match(status, /0 route\(s\) total, 0 block a static export/);
});

console.log(failures === 0 ? '\nexport-routes census: all assertions passed' : `\nexport-routes census: ${failures} assertion(s) FAILED`);
process.exit(failures === 0 ? 0 : 1);