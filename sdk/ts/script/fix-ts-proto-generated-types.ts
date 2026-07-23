#!/usr/bin/env -S node --experimental-strip-types --no-warnings

import { promises as fs } from "node:fs";
import { dirname, posix, resolve as resolvePath } from "node:path";

const helperNames = ["isSet", "bytesFromBase64", "base64FromBytes", "toTimestamp", "fromTimestamp", "fromJsonTimestamp", "numberToLong", "isObject"];
const helperRegex = new RegExp(
  `^(function|const)\\s+(${helperNames.join("|")})\\b`,
  "gm",
);
const typeHelpers = ["MessageFns", "DeepPartial"];
const helperTypeRegex = new RegExp(
  `^(interface|type)\\s+(${typeHelpers.join("|")})\\b`,
  "gm",
);

const ROOT_DIR = resolvePath(import.meta.dirname, "..", "src");
const paths: string[] = [];
for await (const path of fs.glob(`${ROOT_DIR}/generated/protos/**/*.ts`)) {
  paths.push(path);
}
paths.sort();

for (const path of paths) {
  const source = await fs.readFile(path, "utf8");
  let newSource = source;

  // Remove the `create` method from message objects
  newSource = newSource.replace(/^\s*create\(base\?:\s*DeepPartial<\w+>\):\s*\w+\s*\{\s*return\s*\w+\.fromPartial\(base \?\? \{\}\);\s*\},?\n?/gm, "");
  newSource = injectOwnHelpers(newSource, path);
	newSource = restoreCosmosNumericWrappers(newSource, path);

  if (newSource !== source) {
    const temporaryPath = `${path}.tmp`;
    await fs.writeFile(temporaryPath, newSource);
    await fs.rename(temporaryPath, path);
  }
}

function restoreCosmosNumericWrappers(source: string, path: string) {
	const normalizedPath = path.replace(/\\/g, "/");
	if (!normalizedPath.endsWith("/generated/protos/cosmos/base/v1beta1/coin.ts") || source.includes("export interface IntProto")) {
		return source;
	}

	const insertionPoint = "type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;";
	if (!source.includes(insertionPoint)) {
		throw new Error(`cannot restore Cosmos numeric wrappers in ${path}`);
	}
	const wrappers = `/** Numeric wrapper retained for Cosmos SDK wire compatibility. */
export interface IntProto {
  int: string;
}

/** Decimal wrapper retained for Cosmos SDK wire compatibility. */
export interface DecProto {
  dec: string;
}

function createBaseIntProto(): IntProto {
  return { int: "" };
}

export const IntProto: MessageFns<IntProto, "cosmos.base.v1beta1.IntProto"> = {
  $type: "cosmos.base.v1beta1.IntProto" as const,
  encode(message: IntProto, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.int !== "") writer.uint32(10).string(message.int);
    return writer;
  },
  decode(input: BinaryReader | Uint8Array, length?: number): IntProto {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseIntProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      if (tag === 10) message.int = reader.string();
      else if ((tag & 7) === 4 || tag === 0) break;
      else reader.skip(tag & 7);
    }
    return message;
  },
  fromJSON(object: any): IntProto {
    return { int: isSet(object.int) ? globalThis.String(object.int) : "" };
  },
  toJSON(message: IntProto): unknown {
    const obj: any = {};
    if (message.int !== "") obj.int = message.int;
    return obj;
  },
  fromPartial(object: DeepPartial<IntProto>): IntProto {
    const message = createBaseIntProto();
    message.int = object.int ?? "";
    return message;
  },
};

function createBaseDecProto(): DecProto {
  return { dec: "" };
}

export const DecProto: MessageFns<DecProto, "cosmos.base.v1beta1.DecProto"> = {
  $type: "cosmos.base.v1beta1.DecProto" as const,
  encode(message: DecProto, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.dec !== "") writer.uint32(10).string(message.dec);
    return writer;
  },
  decode(input: BinaryReader | Uint8Array, length?: number): DecProto {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseDecProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      if (tag === 10) message.dec = reader.string();
      else if ((tag & 7) === 4 || tag === 0) break;
      else reader.skip(tag & 7);
    }
    return message;
  },
  fromJSON(object: any): DecProto {
    return { dec: isSet(object.dec) ? globalThis.String(object.dec) : "" };
  },
  toJSON(message: DecProto): unknown {
    const obj: any = {};
    if (message.dec !== "") obj.dec = message.dec;
    return obj;
  },
  fromPartial(object: DeepPartial<DecProto>): DecProto {
    const message = createBaseDecProto();
    message.dec = object.dec ?? "";
    return message;
  },
};

`;
	return source.replace(insertionPoint, wrappers + insertionPoint);
}

function injectOwnHelpers(source: string, path: string) {
  const foundHelperNames = new Set<string>();
  source = source.replace(helperRegex, (_, kind, name) => {
    foundHelperNames.add(name);
    return `${kind} _unused_${name}`;
  });

  const foundTypeHelperNames = new Set<string>();
  source = source.replace(helperTypeRegex, (_, kind, name) => {
    foundTypeHelperNames.add(name);
    return `${kind} _unused_${name}`;
  });

  const importHelpers = foundHelperNames.size
    ? `import { ${Array.from(foundHelperNames).join(", ")} } from "${relativeImportPath(dirname(path), `${ROOT_DIR}/encoding/typeEncodingHelpers.ts`)}"\n`
    : "";
  const importTypeHelpers = foundTypeHelperNames.size
    ? `import type { ${Array.from(foundTypeHelperNames).join(", ")} } from "${relativeImportPath(dirname(path), `${ROOT_DIR}/encoding/typeEncodingHelpers.ts`)}"\n`
    : "";

  return importHelpers + importTypeHelpers + source;
}

function relativeImportPath(fromPath: string, toPath: string) {
  const normalizedFrom = fromPath.replace(/\\/g, "/");
  const normalizedTo = toPath.replace(/\\/g, "/");
  return posix.relative(normalizedFrom, normalizedTo);
}
