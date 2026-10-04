/**
 * Parity tests for the repo-owned resolver that replaced `ts-jest-resolver@2.0.1`.
 *
 * These run as part of the sdk/ts jest suites, so a resolver change that alters
 * module resolution behaviour fails CI rather than silently changing what the
 * test suite can load.
 *
 * Each case is asserted against an explicit `defaultResolver` spy that records
 * every request and throws on anything it does not know, so the resolver's
 * candidate ORDER and its fallbacks are observable rather than assumed.
 *
 * Paths are plain POSIX literals on purpose: the resolver only rewrites the
 * extension of the string Jest hands it, so no host path normalisation should
 * enter the expectation (`path.resolve` would make these assertions pass on one
 * OS and fail on another).
 */

"use strict";

const test = require("node:test");
const assert = require("node:assert/strict");

const resolverUnderTest = require("./jest-resolver.cjs");

/**
 * Build a `defaultResolver` spy that only accepts a known set of paths.
 * Throwing on anything else is what lets the resolver's fallbacks be observed.
 */
function makeDefaultResolver(known = []) {
	const calls = [];
	const fn = (request) => {
		calls.push(request);
		if (!known.includes(request)) {
			throw new Error(`Cannot find module '${request}'`);
		}
		return request;
	};
	fn.calls = calls;
	return fn;
}

test("maps a .js request to the .ts source when it exists", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/dep.ts"]);
	const result = resolverUnderTest("/proj/src/dep.js", { defaultResolver });
	assert.equal(result, "/proj/src/dep.ts");
	assert.deepEqual(defaultResolver.calls, ["/proj/src/dep.ts"]);
});

test("tries .ts first for a .js request (upstream candidate order)", () => {
	// Upstream order for a `.js` request is ['.ts', '.tsx'] (ts-jest-resolver
	// dist/index.cjs), and the loop RETURNS on the first candidate that resolves
	// rather than exhausting the list — so a resolvable `.ts` must short-circuit
	// before `.tsx` is ever attempted.
	const defaultResolver = makeDefaultResolver(["/proj/src/a.tsx", "/proj/src/a.ts"]);
	const result = resolverUnderTest("/proj/src/a.js", { defaultResolver });
	assert.equal(result, "/proj/src/a.ts");
	assert.deepEqual(defaultResolver.calls, ["/proj/src/a.ts"]);
});

test("falls through to .tsx when no .ts candidate resolves", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/a.tsx"]);
	const result = resolverUnderTest("/proj/src/a.js", { defaultResolver });
	assert.equal(result, "/proj/src/a.tsx");
	assert.deepEqual(defaultResolver.calls, ["/proj/src/a.ts", "/proj/src/a.tsx"]);
});

test("falls back to the original path when no TypeScript candidate resolves", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/plain.js"]);
	const result = resolverUnderTest("/proj/src/plain.js", { defaultResolver });
	assert.equal(result, "/proj/src/plain.js");
	assert.deepEqual(defaultResolver.calls, [
		"/proj/src/plain.ts",
		"/proj/src/plain.tsx",
		"/proj/src/plain.js",
	]);
});

test("maps .cjs to .cts", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/x.cts"]);
	assert.equal(resolverUnderTest("/proj/src/x.cjs", { defaultResolver }), "/proj/src/x.cts");
});

test("maps .mjs to .mts", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/y.mts"]);
	assert.equal(resolverUnderTest("/proj/src/y.mjs", { defaultResolver }), "/proj/src/y.mts");
});

test("leaves a .jsx request trying .ts, .tsx then .js in order", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/c.js"]);
	resolverUnderTest("/proj/src/c.jsx", { defaultResolver });
	assert.deepEqual(defaultResolver.calls, [
		"/proj/src/c.ts",
		"/proj/src/c.tsx",
		"/proj/src/c.js",
	]);
});

test("passes a path with no JavaScript extension straight through", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/native.node"]);
	const result = resolverUnderTest("/proj/src/native.node", { defaultResolver });
	assert.equal(result, "/proj/src/native.node");
	assert.deepEqual(defaultResolver.calls, ["/proj/src/native.node"]);
});

test("propagates the original error when even the fallback path does not resolve", () => {
	const defaultResolver = makeDefaultResolver([]);
	assert.throws(
		() => resolverUnderTest("/proj/src/missing.js", { defaultResolver }),
		/Cannot find module/,
	);
});

test("is case-insensitive on the extension, like the package it replaces", () => {
	const defaultResolver = makeDefaultResolver(["/proj/src/z.ts"]);
	// `.JS` must still match the /\.js$/i rule.
	assert.equal(resolverUnderTest("/proj/src/z.JS", { defaultResolver }), "/proj/src/z.ts");
});

test("RESOLUTIONS is frozen so a caller cannot mutate shared state", () => {
	assert.equal(Object.isFrozen(resolverUnderTest.RESOLUTIONS), true);
});