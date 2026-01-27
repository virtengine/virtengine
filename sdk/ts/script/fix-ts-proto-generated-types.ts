#!/usr/bin/env -S node --experimental-strip-types --no-warnings

import { promises as fs } from "node:fs";
import { dirname, relative as relativePath, resolve as resolvePath } from "node:path";

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
for await (const path of fs.glob(`${ROOT_DIR}/generated/protos/**/*.ts`)) {
  const source = await fs.readFile(path, "utf8");
  let newSource = source;

  // Remove the `create` method from message objects
  newSource = newSource.replace(/^\s*create\(base\?:\s*DeepPartial<\w+>\):\s*\w+\s*\{\s*return\s*\w+\.fromPartial\(base \?\? \{\}\);\s*\},?\n?/gm, "");
  newSource = injectOwnHelpers(newSource, path);

  if (newSource !== source) {
    await fs.writeFile(path, newSource);
  }
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
    ? `import { ${Array.from(foundHelperNames).join(", ")} } from "${relativePath(dirname(path), `${ROOT_DIR}/encoding/typeEncodingHelpers.ts`)}"\n`
    : "";
  const importTypeHelpers = foundTypeHelperNames.size
    ? `import type { ${Array.from(foundTypeHelperNames).join(", ")} } from "${relativePath(dirname(path), `${ROOT_DIR}/encoding/typeEncodingHelpers.ts`)}"\n`
    : "";

  return importHelpers + importTypeHelpers + source;
}
