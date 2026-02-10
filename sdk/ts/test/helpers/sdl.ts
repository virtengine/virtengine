import dot from "dot-object";
import type { CustomCommands, Spec } from "immutability-helper";
import update from "immutability-helper";
import { dump as dumpYaml } from "js-yaml";

import type { v2Sdl } from "../../src/sdl/types.ts";
// @ts-expect-error ts-jest doesn't process import tags
import groupsBasic from "../fixtures/sdl/groups-basic-snapshot.json" with { type: "json" };
// @ts-expect-error ts-jest doesn't process import tags
import manifestBasic from "../fixtures/sdl/manifest-basic-snapshot.json" with { type: "json" };
// @ts-expect-error ts-jest doesn't process import tags
import sdlBasic from "../fixtures/sdl/sdl-basic.json" with { type: "json" };

type AnySpec = Spec<object, CustomCommands<object>>;

export const createSdlJson = ($spec: AnySpec = {}): v2Sdl => {
  return update(sdlBasic, dot.object($spec)) as unknown as v2Sdl;
};

export const createSdlYml = ($spec: AnySpec = {}): string => {
  return dumpYaml(createSdlJson($spec), { forceQuotes: true, quotingType: "\"" });
};

export const createManifest = ($spec: AnySpec = {}) => {
  return update(manifestBasic, dot.object($spec) as AnySpec) as Record<string, unknown>;
};

export const createGroups = ($spec: AnySpec = {}) => {
  return update(groupsBasic, dot.object($spec) as AnySpec) as Record<string, unknown>;
};
