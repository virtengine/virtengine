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
  assert.equal(inventory.summaries.replaces, 22);
  assert.equal(inventory.generated.gatewayStubs.length, 0);
});