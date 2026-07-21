// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import { readFile, readdir, writeFile } from "node:fs/promises";
import { dirname, join, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");

function normalize(value) {
  if (Array.isArray(value)) return value.map(normalize);
  if (value && typeof value === "object") {
    return Object.fromEntries(Object.entries(value).sort(([left], [right]) => left.localeCompare(right)).map(([key, nested]) => [key, normalize(nested)]));
  }
  return value;
}

async function fragments(root) {
  const paths = [];
  async function visit(directory) {
    const entries = await readdir(directory, { withFileTypes: true });
    entries.sort((left, right) => left.name.localeCompare(right.name));
    for (const entry of entries) {
      const path = join(directory, entry.name);
      if (entry.isDirectory()) await visit(path);
      else if (entry.name.endsWith(".swagger.json")) paths.push(path);
    }
  }
  await visit(root);
  return paths;
}

export async function composeOpenAPI(fragmentRoot) {
  const result = {
    swagger: "2.0",
    info: { title: "VirtEngine protobuf API", version: "1.0.0", description: "Generated from sdk/proto descriptors. Manual provider portal APIs remain in api/openapi/portal_api.yaml." },
    consumes: ["application/json"],
    produces: ["application/json"],
    paths: {},
    definitions: {},
    "x-virtengine-source": "sdk/proto",
  };
  const operationIds = new Map();
  for (const fragmentPath of await fragments(fragmentRoot)) {
    const fragment = JSON.parse(await readFile(fragmentPath, "utf8"));
    const source = relative(fragmentRoot, fragmentPath).split(sep).join("/");
    const protoSource = fragment.info?.title;
    if (typeof protoSource !== "string" || !protoSource.endsWith(".proto")) {
      throw new Error(`OpenAPI fragment ${source} does not identify its protobuf source`);
    }
    const packageName = protoSource.slice(0, protoSource.lastIndexOf("/")).replaceAll("/", ".");
    for (const [path, item] of Object.entries(fragment.paths ?? {})) {
      if (result.paths[path]) throw new Error(`duplicate OpenAPI path ${path}`);
      for (const [verb, operation] of Object.entries(item)) {
        if (!operation.operationId) throw new Error(`missing operationId for ${verb.toUpperCase()} ${path}`);
        const stableOperationId = `${packageName}.${operation.operationId}`;
        const owner = operationIds.get(stableOperationId);
        if (owner) throw new Error(`duplicate operationId ${stableOperationId}: ${owner}, ${verb.toUpperCase()} ${path}`);
        operation.operationId = stableOperationId;
        operationIds.set(stableOperationId, `${verb.toUpperCase()} ${path}`);
        operation["x-virtengine-proto-fragment"] = source;
        operation["x-virtengine-proto-source"] = protoSource;
      }
      result.paths[path] = item;
    }
    for (const [name, schema] of Object.entries(fragment.definitions ?? {})) {
      const existing = result.definitions[name];
      const normalizedSchema = normalize(schema);
      if (existing && JSON.stringify(normalize(existing)) !== JSON.stringify(normalizedSchema)) {
        const existingSize = JSON.stringify(existing).length;
        const candidateSize = JSON.stringify(normalizedSchema).length;
        if (existingSize === candidateSize) throw new Error(`conflicting OpenAPI definition ${name}`);
        result.definitions[name] = candidateSize > existingSize ? normalizedSchema : existing;
      } else {
        result.definitions[name] = normalizedSchema;
      }
    }
  }
  return normalize(result);
}

async function main() {
  const fragmentRoot = resolve(repositoryRoot, process.argv[2] ?? "sdk/.generation/openapi/fragments");
  const output = resolve(repositoryRoot, process.argv[3] ?? "api/openapi/virtengine-proto.swagger.json");
  const document = await composeOpenAPI(fragmentRoot);
  await writeFile(output, `${JSON.stringify(document, null, 2)}\n`);
  console.log(`wrote ${relative(repositoryRoot, output).split(sep).join("/")} (${Object.keys(document.paths).length} paths)`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    console.error(error.stack ?? error);
    process.exitCode = 1;
  });
}
