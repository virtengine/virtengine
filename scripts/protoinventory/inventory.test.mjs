// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import assert from "node:assert/strict";
import test from "node:test";

import { buildInventory, parseProto, validateInventory } from "./inventory.mjs";

test("parseProto extracts services, methods, types, and HTTP bindings", () => {
  const source = `
    syntax = "proto3";
    package virtengine.example.v1;
    option go_package = "github.com/virtengine/virtengine/sdk/go/node/example/v1";
    service Query {
      rpc Item(QueryItemRequest) returns (QueryItemResponse) {
        option (google.api.http).get = "/virtengine/example/v1/items/{item_id}";
      }
      rpc Search(QuerySearchRequest) returns (QuerySearchResponse) {
        option (google.api.http) = {
          post: "/virtengine/example/v1/search"
          body: "*"
        };
      }
    }
  `;

  assert.deepEqual(parseProto("virtengine/example/v1/query.proto", source), {
    file: "virtengine/example/v1/query.proto",
    package: "virtengine.example.v1",
    goPackage: "github.com/virtengine/virtengine/sdk/go/node/example/v1",
    services: [
      {
        name: "Query",
        fullName: "virtengine.example.v1.Query",
        kind: "query",
        methods: [
          {
            name: "Item",
            fullName: "virtengine.example.v1.Query.Item",
            grpcPath: "/virtengine.example.v1.Query/Item",
            requestType: "virtengine.example.v1.QueryItemRequest",
            responseType: "virtengine.example.v1.QueryItemResponse",
            http: [{ body: "", method: "GET", path: "/virtengine/example/v1/items/{item_id}" }],
          },
          {
            name: "Search",
            fullName: "virtengine.example.v1.Query.Search",
            grpcPath: "/virtengine.example.v1.Query/Search",
            requestType: "virtengine.example.v1.QuerySearchRequest",
            responseType: "virtengine.example.v1.QuerySearchResponse",
            http: [{ body: "*", method: "POST", path: "/virtengine/example/v1/search" }],
          },
        ],
      },
    ],
  });
});

test("validateInventory rejects duplicate HTTP verb and path pairs", () => {
  const inventory = {
    proto: {
      files: [
        {
          file: "one.proto",
          services: [{ methods: [{ fullName: "a.Query.One", http: [{ method: "GET", path: "/same" }] }] }],
        },
        {
          file: "two.proto",
          services: [{ methods: [{ fullName: "b.Query.Two", http: [{ method: "GET", path: "/same" }] }] }],
        },
      ],
    },
  };

  assert.throws(() => validateInventory(inventory), /duplicate HTTP binding GET \/same/);
});

test("repository inventory has exact Go module replaces and TypeScript proto parity", async () => {
  const { canonical } = await buildInventory();
  const inventory = JSON.parse(canonical);

  assert.equal(inventory.modules.some((module) => module.replaces.some((replacement) => replacement.old === "(")), false);

  // Exact pin, deliberately a literal: this is the canary that forces a human
  // to look at a replace-set change before it ships. Update it in the same PR
  // that changes go.mod, and say in that PR which directive was added or
  // repointed.
  //
  // 23 as of 2026-10-03: 15 in go.mod + 1 in sdk/generation/go.mod + 7 in
  // sdk/go/go.mod. It was 24 until #1200 dropped the
  // `github.com/CosmWasm/wasmd => github.com/CosmWasm/wasmd v0.53.4` replace
  // from sdk/go/go.mod, which is what forced wasmvm v2 to be pulled in
  // alongside v3. #1200 kept the `require` line (now v0.61.7), so only the
  // replace went away -- this constant and the regenerated inventory were
  // updated in the same change.
  assert.equal(inventory.summaries.replaces, 23);

  // The removed directive must stay removed: #1200 exists precisely to stop
  // sdk/go from resolving wasmvm v2 alongside v3, and the count above cannot
  // catch a regenerate that quietly reintroduces it -- any other replace being
  // added or dropped in the same change would still total 23. Assert the
  // absence of the directive itself, not just its number. This is the same
  // shape as the virtengine/ledger-go assertion below: assert the thing, do
  // not trust a count to notice its absence.
  assert.equal(
    inventory.modules.some((module) =>
      module.replaces.some(
        (replacement) => replacement.old === "github.com/CosmWasm/wasmd" && replacement.version === "v0.53.4",
      ),
    ),
    false,
  );

  // A replace must never point at the deleted virtengine/ledger-go repository
  // (#848); assert the target rather than trusting the count to catch it.
  assert.equal(
    inventory.modules.some((module) => module.replaces.some((replacement) => String(replacement.new).includes("virtengine/ledger-go"))),
    false,
  );

  // Every replace must name a concrete source, not an unresolved placeholder:
  // a replace whose `old` is not a real module path silently stops applying and
  // the inventory keeps reporting a count that no longer describes the build.
  for (const module of inventory.modules) {
    for (const replacement of module.replaces) {
      assert.match(replacement.old, /^[^(\s]+\/[^(\s]+$/);
    }
  }

  assert.equal(inventory.generated.gatewayStubs.length, 0);
});