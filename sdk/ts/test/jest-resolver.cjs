/**
 * Repo-owned TypeScript resolver for Jest.
 *
 * Why this replaces `ts-jest-resolver`
 * -----------------------------------
 * `ts-jest-resolver@2.0.1` declares a hard dependency on `jest-resolve@^29.5.0`,
 * which transitively pins `jest-haste-map@29` -> `micromatch@4` -> `braces@3.0.3`.
 * That whole subtree is the last remaining path to GHSA-vfj7-8cjw-p6xm
 * (CVE-2026-93687) in the `sdk/ts` dependency graph, and the node-vuln-scan gate
 * in `.github/workflows/security.yaml` audits `sdk/ts` INCLUDING devDependencies,
 * so it must leave the graph entirely — a devDependency cannot be waived.
 *
 * Measured facts this rests on (not inference):
 *   - `braces` has NO fixed release: `npm view braces dist-tags.latest` is
 *     3.0.3 and the advisory's `first_patched_version` is `null`, so no version
 *     bump of `braces` or `micromatch` can clear it. `braces` must not be present.
 *   - `ts-jest-resolver` NEVER imports `jest-resolve`. Grepping every file in
 *     its `dist/` for `require(` / `import` returns zero hits; the only external
 *     value it touches is `options.defaultResolver`, which Jest injects itself.
 *     Its declared `jest-resolve` dependency is therefore pure packaging weight,
 *     and removing it changes no runtime behaviour.
 *
 * Behaviour is a faithful port of `ts-jest-resolver@2.0.1`'s algorithm
 * (`dist/index.cjs`): try TypeScript extensions for a JavaScript-ish request and
 * fall back to the original path, delegating EVERY decision to the resolver Jest
 * supplies. `test/jest-resolver.test.cjs` pins that parity — run it with
 * `node --test test/jest-resolver.test.cjs`.
 *
 * Two spec suites (`src/sdk/provider/auth/jwt/jwt-{token,validator}.spec.ts`)
 * fail on `develop` independently of this tool chain: their fixtures use
 * `virtengine1…` account prefixes while the committed generated schema
 * `src/sdk/provider/auth/jwt/validateJwtPayload.ts` expects `^ve1[a-z0-9]{38}$`.
 * Regenerating that file from `specs/jwt-schema.json` is the fix; it is not
 * caused by, nor fixable by, this resolver.
 */

"use strict";

/** @type {ReadonlyArray<{ matcher: RegExp, extensions: readonly string[] }>} */
const RESOLUTIONS = Object.freeze([
	// `.js` requests may really be TypeScript sources (the ESM build convention).
	{ matcher: /\.js$/i, extensions: [".ts", ".tsx"] },
	{ matcher: /\.jsx$/i, extensions: [".ts", ".tsx", ".js"] },
	// CommonJS-marked request that may really be `.cts`.
	{ matcher: /\.cjs$/i, extensions: [".cts"] },
	// ES-module-marked request that may really be `.mts`.
	{ matcher: /\.mjs$/i, extensions: [".mts"] },
]);

/**
 * @param {string} path the module path Jest wants resolved
 * @param {object} options Jest's resolver options; `defaultResolver` is supplied by Jest
 * @returns {string} the first path that resolves, else `options.defaultResolver(path)`
 */
function resolveTsExtension(path, options) {
	const resolver = options.defaultResolver;
	const resolution = RESOLUTIONS.find(({ matcher }) => matcher.test(path));

	if (resolution) {
		for (const extension of resolution.extensions) {
			try {
				return resolver(path.replace(resolution.matcher, extension), options);
			} catch {
				// Try the next candidate extension, exactly like the package this replaces.
			}
		}
	}

	// No JS-ish extension, or no candidate resolved: defer to Jest unchanged.
	return resolver(path, options);
}

module.exports = resolveTsExtension;
module.exports.RESOLUTIONS = RESOLUTIONS;