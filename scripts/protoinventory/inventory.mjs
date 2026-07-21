// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { readFile, readdir, writeFile } from "node:fs/promises";
import { dirname, join, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");

function normalizePath(path) {
  return path.split(sep).join("/");
}

function qualifiedType(packageName, typeName) {
  if (typeName.startsWith(".")) {
    return typeName.slice(1);
  }
  if (typeName.includes(".")) {
    return typeName;
  }
  return packageName ? `${packageName}.${typeName}` : typeName;
}

function findBlocks(source, expression) {
  const blocks = [];
  for (const match of source.matchAll(expression)) {
    const open = source.indexOf("{", match.index);
    let depth = 0;
    let close = -1;
    for (let index = open; index < source.length; index += 1) {
      if (source[index] === "{") depth += 1;
      if (source[index] === "}") {
        depth -= 1;
        if (depth === 0) {
          close = index;
          break;
        }
      }
    }
    if (close < 0) throw new Error(`unterminated block beginning at byte ${match.index}`);
    blocks.push({ name: match.groups.name, body: source.slice(open + 1, close) });
  }
  return blocks;
}

function parseHTTPBindings(methodBody) {
  const bindings = [];
  const direct = /option\s+\(google\.api\.http\)\.(?<verb>get|put|post|delete|patch)\s*=\s*"(?<path>[^"]+)"\s*;/gi;
  for (const match of methodBody.matchAll(direct)) {
    bindings.push({ body: "", method: match.groups.verb.toUpperCase(), path: match.groups.path });
  }

  const blocks = /option\s+\(google\.api\.http\)\s*=\s*\{(?<body>[\s\S]*?)\}\s*;/gi;
  for (const match of methodBody.matchAll(blocks)) {
    const verb = match.groups.body.match(/\b(get|put|post|delete|patch)\s*:\s*"([^"]+)"/i);
    if (!verb) continue;
    const body = match.groups.body.match(/\bbody\s*:\s*"([^"]*)"/i);
    bindings.push({ body: body?.[1] ?? "", method: verb[1].toUpperCase(), path: verb[2] });
  }

  return bindings.sort((left, right) => `${left.method} ${left.path}`.localeCompare(`${right.method} ${right.path}`));
}

export function parseProto(file, source) {
  const packageName = source.match(/\bpackage\s+([\w.]+)\s*;/)?.[1] ?? "";
  const goPackage = source.match(/\boption\s+go_package\s*=\s*"([^"]+)"\s*;/)?.[1] ?? "";
  const services = [];

  for (const service of findBlocks(source, /\bservice\s+(?<name>[A-Za-z_]\w*)\s*\{/g)) {
    const methods = [];
    const rpcPattern = /\brpc\s+(?<name>[A-Za-z_]\w*)\s*\(\s*(?:stream\s+)?(?<request>[.\w]+)\s*\)\s*returns\s*\(\s*(?:stream\s+)?(?<response>[.\w]+)\s*\)\s*(?<tail>[;{])/g;
    for (const rpc of service.body.matchAll(rpcPattern)) {
      let body = "";
      if (rpc.groups.tail === "{") {
        const open = service.body.indexOf("{", rpc.index);
        let depth = 0;
        for (let index = open; index < service.body.length; index += 1) {
          if (service.body[index] === "{") depth += 1;
          if (service.body[index] === "}") {
            depth -= 1;
            if (depth === 0) {
              body = service.body.slice(open + 1, index);
              break;
            }
          }
        }
      }
      const serviceFullName = packageName ? `${packageName}.${service.name}` : service.name;
      methods.push({
        name: rpc.groups.name,
        fullName: `${serviceFullName}.${rpc.groups.name}`,
        grpcPath: `/${serviceFullName}/${rpc.groups.name}`,
        requestType: qualifiedType(packageName, rpc.groups.request),
        responseType: qualifiedType(packageName, rpc.groups.response),
        http: parseHTTPBindings(body),
      });
    }
    const fullName = packageName ? `${packageName}.${service.name}` : service.name;
    services.push({
      name: service.name,
      fullName,
      kind: service.name.startsWith("Query") ? "query" : service.name === "Msg" || service.name === "MsgService" ? "msg" : "service",
      methods,
    });
  }

  return { file: normalizePath(file), package: packageName, goPackage, services };
}

async function walk(root, predicate, excludedDirectories = new Set()) {
  const files = [];
  async function visit(directory) {
    const entries = await readdir(directory, { withFileTypes: true });
    entries.sort((left, right) => left.name.localeCompare(right.name));
    for (const entry of entries) {
      const absolute = join(directory, entry.name);
      if (entry.isDirectory() && !excludedDirectories.has(entry.name)) await visit(absolute);
      else if (predicate(absolute)) files.push(absolute);
    }
  }
  await visit(root);
  return files;
}

function summarizeProto(files) {
  const services = files.flatMap((file) => file.services);
  const methods = services.flatMap((service) => service.methods);
  const httpBindings = methods.flatMap((method) => method.http);
  return {
    files: files.length,
    services: services.length,
    queryServices: services.filter((service) => service.kind === "query").length,
    msgServices: services.filter((service) => service.kind === "msg").length,
    methods: methods.length,
    httpBindings: httpBindings.length,
  };
}

async function moduleInventory() {
  const moduleFiles = await walk(
    repositoryRoot,
    (file) => file.endsWith(`${sep}go.mod`),
    new Set([".cache", ".git", "node_modules", "vendor"]),
  );
  const modules = [];
  for (const moduleFile of moduleFiles) {
    const source = await readFile(moduleFile, "utf8");
    const moduleDirectory = dirname(moduleFile);
    const moduleJSON = JSON.parse(execFileSync("go", ["mod", "edit", "-json"], {
      cwd: moduleDirectory,
      encoding: "utf8",
      env: { ...process.env, GOWORK: "off" },
    }));
    const replaces = (moduleJSON.Replace ?? []).map((replacement) => ({
      old: replacement.Old.Path,
      oldVersion: replacement.Old.Version ?? "",
      new: replacement.New.Path,
      version: replacement.New.Version ?? "",
    }));
    modules.push({
      path: normalizePath(relative(repositoryRoot, moduleFile) || "go.mod"),
      module: source.match(/^module\s+([^\s]+)/m)?.[1] ?? "",
      go: source.match(/^go\s+([^\s]+)/m)?.[1] ?? "",
      workspace: normalizePath(relative(repositoryRoot, moduleFile)) !== "sdk/generation/go.mod",
      hasGoSum: await readFile(join(dirname(moduleFile), "go.sum"), "utf8").then(() => true, () => false),
      vendorPolicy: normalizePath(relative(repositoryRoot, moduleFile)) === "sdk/go/go.mod" ? "generated-and-ignored" : "not-checked-in",
      replaces: replaces.sort((left, right) => left.old.localeCompare(right.old)),
    });
  }
  return modules.sort((left, right) => left.path.localeCompare(right.path));
}

async function generatedInventory() {
  const ignoredGoDirectories = new Set(["vendor"]);
  const goFiles = await walk(join(repositoryRoot, "sdk/go"), (file) => file.endsWith(".pb.go"), ignoredGoDirectories);
  const gateways = await walk(join(repositoryRoot, "sdk/go"), (file) => file.endsWith(".pb.gw.go"), ignoredGoDirectories);
  const stubs = await walk(join(repositoryRoot, "sdk/go"), (file) => file.endsWith(`${sep}gateway_stub.go`), ignoredGoDirectories);
  const tsFiles = await walk(join(repositoryRoot, "sdk/ts/src/generated"), (file) => file.endsWith(".ts"));
  return {
    goMessages: goFiles.map((file) => normalizePath(relative(repositoryRoot, file))),
    gateways: gateways.map((file) => normalizePath(relative(repositoryRoot, file))),
    gatewayStubs: stubs.map((file) => normalizePath(relative(repositoryRoot, file))),
    typescript: tsFiles.map((file) => normalizePath(relative(repositoryRoot, file))),
    python: { supported: false, policy: "fail-closed: no published Python proto contract is declared" },
    rust: { supported: false, policy: "fail-closed: Rust generation config is experimental and not a release output" },
  };
}

export function validateInventory(inventory) {
  const routes = new Map();
  for (const file of inventory.proto.files) {
    for (const service of file.services) {
      for (const method of service.methods) {
        for (const binding of method.http) {
          const key = `${binding.method} ${binding.path}`;
          const owner = routes.get(key);
          if (owner && owner !== method.fullName) throw new Error(`duplicate HTTP binding ${key}: ${owner}, ${method.fullName}`);
          routes.set(key, method.fullName);
        }
      }
    }
  }
  if (inventory.generated?.gatewayStubs?.length) throw new Error(`production gateway stubs remain: ${inventory.generated.gatewayStubs.join(", ")}`);

  const protoFiles = new Set(inventory.proto.files.map((file) => file.file));
  const tsProtoFiles = new Set(inventory.generated.typescript
    .filter((file) => file.startsWith("sdk/ts/src/generated/protos/"))
    .map((file) => file.slice("sdk/ts/src/generated/protos/".length)));
  for (const protoFile of protoFiles) {
    const expectedTS = protoFile.replace(/^(node|provider)\//, "").replace(/\.proto$/, ".ts");
    if (!tsProtoFiles.has(expectedTS)) throw new Error(`missing generated TypeScript contract for ${protoFile}: ${expectedTS}`);
  }
  return inventory;
}

export async function buildInventory() {
  const protoRoot = join(repositoryRoot, "sdk/proto");
  const protoPaths = await walk(protoRoot, (file) => file.endsWith(".proto"));
  const protoFiles = [];
  for (const path of protoPaths) protoFiles.push(parseProto(relative(protoRoot, path), await readFile(path, "utf8")));
  const modules = await moduleInventory();
  const generated = await generatedInventory();
  const inventory = {
    schemaVersion: 1,
    sourceGraph: {
      root: "sdk/proto",
      bufWorkspace: "sdk/buf.yaml",
      lock: "sdk/buf.lock",
      nodeModule: "sdk/proto/node",
      providerModule: "sdk/proto/provider",
    },
    modules,
    proto: { summary: summarizeProto(protoFiles), files: protoFiles },
    generated,
    summaries: {
      modules: modules.filter((module) => module.workspace).length,
      toolModules: modules.filter((module) => !module.workspace).length,
      replaces: modules.reduce((sum, module) => sum + module.replaces.length, 0),
      goMessages: generated.goMessages.length,
      gateways: generated.gateways.length,
      gatewayStubs: generated.gatewayStubs.length,
      typescript: generated.typescript.length,
    },
  };
  const canonical = `${JSON.stringify(validateInventory(inventory), null, 2)}\n`;
  return { canonical, sha256: createHash("sha256").update(canonical).digest("hex") };
}

async function main() {
  const output = process.argv[2] ? resolve(repositoryRoot, process.argv[2]) : join(repositoryRoot, "sdk/artifacts/proto/inventory.json");
  const digestOutput = `${output}.sha256`;
  const { canonical, sha256 } = await buildInventory();
  await writeFile(output, canonical);
  await writeFile(digestOutput, `${sha256}  ${normalizePath(relative(repositoryRoot, output))}\n`);
  process.stdout.write(`wrote ${normalizePath(relative(repositoryRoot, output))} (${sha256})\n`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    console.error(error.stack ?? error);
    process.exitCode = 1;
  });
}
