#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { manifestRelativePath, validateManifest } = require("./generate-core-rc-manifest.cjs");

const root = resolve(__dirname, "..");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/core-rc-manifest.schema.json");

function exactKeys(value, keys, label) {
  assert.deepEqual(Object.keys(value).sort(), [...keys].sort(), `${label} has unknown or missing fields`);
}

function assertUnique(values, label) {
  assert.equal(new Set(values).size, values.length, `${label} must be unique`);
}

function validateSchemaNode(node, label) {
  assert.ok(node && typeof node === "object" && !Array.isArray(node), `${label} must be an object`);
  const allowed = new Set(["$ref", "type", "const", "enum", "pattern", "title", "description", "additionalProperties", "required", "properties", "items", "oneOf", "minItems", "maxItems", "uniqueItems", "minimum", "minLength", "minProperties"]);
  for (const key of Object.keys(node)) assert.ok(allowed.has(key), `${label} has unknown schema keyword ${key}`);
  if (node.enum) assertUnique(node.enum, `${label} enum`);
  if (Object.hasOwn(node, "uniqueItems")) assert.equal(node.uniqueItems, true, `${label} uniqueItems must fail closed`);
  if (node.type === "object" && node.properties) {
    assert.equal(node.additionalProperties, false, `${label} object must reject unknown properties`);
    assert.ok(Array.isArray(node.required), `${label} object must declare required fields`);
    assertUnique(node.required, `${label} required fields`);
    assert.deepEqual([...node.required].sort(), Object.keys(node.properties).sort(), `${label} required fields must exactly match properties`);
    for (const [key, child] of Object.entries(node.properties)) validateSchemaNode(child, `${label}.${key}`);
  } else if (node.additionalProperties && typeof node.additionalProperties === "object") {
    validateSchemaNode(node.additionalProperties, `${label}.additionalProperties`);
  }
  if (node.items) validateSchemaNode(node.items, `${label}.items`);
  if (node.oneOf) node.oneOf.forEach((child, index) => validateSchemaNode(child, `${label}.oneOf[${index}]`));
}

function validateSchema(schema) {
  exactKeys(schema, ["$schema", "$id", "title", "type", "additionalProperties", "required", "properties", "$defs"], "schema root");
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.type, "object");
  validateSchemaNode({ type: schema.type, additionalProperties: schema.additionalProperties, required: schema.required, properties: schema.properties }, "schema root");
  assertUnique(schema.required, "schema root required fields");
  for (const [name, definition] of Object.entries(schema.$defs)) validateSchemaNode(definition, `$defs.${name}`);
  assert.equal(schema.properties.schema_version.const, "virtengine.core-rc.prototype/v0");
  assert.equal(schema.properties.authoritative.const, false);
  assert.equal(schema.properties.planned_functionality_complete.const, false);
  assert.equal(schema.properties.milestone_m_eligible.const, false);
  assert.equal(schema.properties.status.const, "dependency_blocked");
  assert.ok(schema.required.includes("blockers"));
  assert.equal(schema.$defs.artifactGroup.additionalProperties, false);
  assert.deepEqual(schema.$defs.toolingArtifact.properties.id.enum, ["generator", "schema", "validator"]);
  assert.deepEqual(schema.$defs.artifactGroup.properties.id.enum, ["lockfiles", "modules", "proto", "openapi", "sdk", "model", "chart"]);
  assert.equal(schema.$defs.producerCheckpoints.properties.accepted.maxItems, 0);
  return true;
}

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

module.exports = { validateSchema };

if (require.main === module) {
  validateSchema(loadJson(schemaPath));
  validateManifest(loadJson(resolve(root, manifestRelativePath)), { rootDir: root });
  console.log("core RC prototype manifest: valid and dependency-blocked");
}