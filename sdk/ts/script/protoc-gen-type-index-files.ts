#!/usr/bin/env -S node --experimental-strip-types --no-warnings

import { type DescEnum, type DescField, type DescMessage } from "@bufbuild/protobuf";
import {
  createEcmaScriptPlugin,
  type GeneratedFile,
  runNodeJs,
  type Schema
} from "@bufbuild/protoplugin";
import { findPathsToCustomField } from "../src/encoding/customTypes/utils.ts";

runNodeJs(
  createEcmaScriptPlugin({
    name: "protoc-gen-type-index-files",
    version: "v1",
    generateTs,
  }),
);

function generateTs(schema: Schema): void {
  const allCustomTypeFieldPaths: DescField[][] = [];

  schema.files.forEach((file) => {
    file.messages.forEach((message) => {
      const paths = findPathsToCustomField(message, () => true);
      allCustomTypeFieldPaths.push(...paths);
    });
  });

  const typesNamesToPatch = new Set<string>();
  allCustomTypeFieldPaths.forEach((path) => {
    path.forEach((field) => {
      typesNamesToPatch.add(field.parent.typeName);
    });
  });
  const protoSource = process.env.PROTO_SOURCE;
  if (!protoSource) {
    throw new Error("PROTO_SOURCE is required and should be set to 'node', 'provider', or 'cosmos'");
  }
  const patchesFileName = `${protoSource}PatchMessage.ts`;

  if (typesNamesToPatch.size > 0) {
    const patchesFile = schema.generateFile(patchesFileName);
    patchesFile.print(`import { patches } from "../patches/${process.env.PROTO_SOURCE}CustomTypePatches.ts";`);
    patchesFile.print(`import type { MessageDesc } from "../../sdk/client/types.ts";`);
    patchesFile.print(`export const patched = <T extends MessageDesc>(messageDesc: T): T => {`);
    patchesFile.print(`  const patchMessage = patches[messageDesc.$type as keyof typeof patches] as any;`);
    patchesFile.print(`  if (!patchMessage) return messageDesc;`);
    patchesFile.print(`  return {`);
    patchesFile.print(`    ...messageDesc,`);
    patchesFile.print(`    encode(message, writer) {`);
    patchesFile.print(`      return messageDesc.encode(patchMessage(message, 'encode'), writer);`);
    patchesFile.print(`    },`);
    patchesFile.print(`    decode(input, length) {`);
    patchesFile.print(`      return patchMessage(messageDesc.decode(input, length), 'decode');`);
    patchesFile.print(`    },`);
    patchesFile.print(`  };`);
    patchesFile.print(`};`);
  }

  const indexFiles: Record<string, {
    file: GeneratedFile;
    symbols: Set<string>;
  }> = {};
  const namespacePrefix = protoSource === 'provider' ? 'provider.' : '';
  schema.files.forEach((file) => {
    const packageParts = file.proto.package.split('.');
    const namespace = namespacePrefix + packageParts[0];
    const version = packageParts.at(-1);
    const path = `index.${namespace}.${version}.ts`;
    indexFiles[path] ??= {
      file: schema.generateFile(path),
      symbols: new Set(),
    };
    const {file: indexFile, symbols: fileSymbols} = indexFiles[path];

    const typesToPatch: Array<{ exportedName: string; name: string }> = [];
    const typesToExport: Array<{ exportedName: string; name: string }> = [];
    for (const type of schema.typesInFile(file)) {
      if (type.kind === 'service' || type.kind === 'extension') continue;

      const name = genName(type);
      const exportedName = fileSymbols.has(name) ? genUniqueName(type, fileSymbols) : name;
      fileSymbols.add(exportedName);

      if (type.kind === "message" && typesNamesToPatch.has(type.typeName)) {
        typesToPatch.push({ exportedName, name });
      } else {
        typesToExport.push({ exportedName, name });
      }
    }

    if (typesToExport.length > 0) {
      const symbolsToExport = typesToExport.map(type => type.exportedName === type.name ? type.exportedName : `${type.name} as ${type.exportedName}`).join(", ");
      indexFile.print(`export { ${symbolsToExport} } from "./${file.name}.ts";`);
    }

    if (typesToPatch.length > 0) {
      const symbolsToPatch = typesToPatch.map((type) => `${type.name} as _${type.exportedName}`).join(", ");
      indexFile.print('');
      indexFile.print(`import { ${symbolsToPatch} } from "./${file.name}.ts";`);
      for (const type of typesToPatch) {
        indexFile.print(`export const ${type.exportedName} = `, indexFile.import('patched', `./${patchesFileName}`),`(_${type.exportedName});`);
        indexFile.print(`export type ${type.exportedName} = _${type.exportedName}`);
      }
    }
  });
}

function genName(type: DescMessage | DescEnum): string {
  return type.typeName.slice(type.file.proto.package.length + 1).replace(/\./g, "_");
}

let uniqueNameCounter = 0;
function genUniqueName(type: DescMessage | DescEnum, allSymbols: Set<string>, attempt = 0): string {
  const name = genName(type);
  if (allSymbols.has(name)) {
    const packageParts = type.file.proto.package.split('.');
    const prefix = packageParts.slice(-2 - attempt, -1).map(capitalize).join('_');
    let newName = `${prefix}_${name}`;
    if (newName === name) {
      newName = `${prefix}_${name}_${uniqueNameCounter++}`;
    }
    return allSymbols.has(newName) ? genUniqueName(type, allSymbols, attempt + 1) : newName;
  }
  return name;
}

function capitalize(str: string): string {
  return str[0].toUpperCase() + str.slice(1);
}
