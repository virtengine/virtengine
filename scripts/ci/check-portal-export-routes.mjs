/**
 * Enumerate every App Router API route that cannot build under `output: 'export'`.
 *
 * Next.js refuses a route handler on a static-export build unless it declares
 * `export const dynamic = 'force-static'` (or a matching `revalidate`). This
 * script reads the route tree from source and reports which routes lack the
 * declaration, so the count is measured rather than inferred from whichever
 * route the build happens to abort on first.
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const portalRoot = path.join(repoRoot, 'portal');
const apiDir = path.join(portalRoot, 'src', 'app', 'api');

if (!fs.existsSync(apiDir)) {
  console.error(`no API route tree at ${apiDir}`);
  process.exit(2);
}

function routeFiles(dir) {
  const out = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) out.push(...routeFiles(full));
    else if (/^route\.(ts|js|tsx|jsx)$/.test(entry.name)) out.push(full);
  }
  return out;
}

const findings = [];
for (const file of routeFiles(apiDir)) {
  const source = fs.readFileSync(file, 'utf8');
  const route = `/${path
    .relative(apiDir, file)
    .replace(/[\\/]route\.(ts|js|tsx|jsx)$/, '')
    .split(path.sep)
    .filter((seg) => seg && !seg.startsWith('(') && seg !== '[[...slug]]')
    .join('/')}`;
  const dynamicMatch = source.match(/^\s*export\s+const\s+dynamic\s*=\s*['"]([^'"]+)['"]/m);
  const revalidateMatch = source.match(/^\s*export\s+const\s+revalidate\s*=\s*(\S+)/m);
  const dynamic = dynamicMatch ? dynamicMatch[1] : null;
  const revalidate = revalidateMatch ? revalidateMatch[1] : null;
  const exportable = dynamic === 'force-static' || revalidate !== null;
  findings.push({ route, dynamic, revalidate, exportable, file: path.relative(portalRoot, file).split(path.sep).join('/') });
}

findings.sort((a, b) => a.route.localeCompare(b.route));

const blocking = findings.filter((f) => !f.exportable);
for (const f of findings) {
  const state = f.exportable ? 'OK  ' : 'BLOCK';
  const why = f.exportable
    ? `dynamic=${f.dynamic ?? '-'} revalidate=${f.revalidate ?? '-'}`
    : 'no export const dynamic=force-static and no revalidate';
  console.log(`${state} ${f.route.padEnd(42)} ${why}`);
}

console.log(
  `\ncheck-portal-export-routes: ${findings.length} route(s) total, ` +
    `${blocking.length} block a static export`
);

// Fail closed on a tree that suddenly has no routes: that means the walk broke
// (a moved directory, a glob typo) or the routes were deleted, and reporting
// "0 block a static export" in either case would be a vacuous pass. A portal
// with no /api tree at all is itself a finding, not a clean bill of health.
if (findings.length === 0) {
  console.error('check-portal-export-routes: no App Router /api routes found - refusing to pass vacuously');
  process.exit(1);
}

process.exit(blocking.length === 0 ? 0 : 1);