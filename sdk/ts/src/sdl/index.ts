/**
 * SDL (Stack Definition Language) module exports
 * Provides functionality for parsing and validating VirtEngine deployment manifests
 *
 * @example
 * ```ts
 * import { SDL } from './sdl';
 *
 * const yaml = `
 * version: "2.0"
 * services:
 *   web:
 *     image: nginx
 *     expose:
 *       - port: 80
 *         as: 80
 *         to:
 *           - global: true
 * `;
 *
 * const sdl = SDL.fromString(yaml);
 * const manifest = sdl.manifest();
 * ```
 */
export { SDL } from "./SDL/SDL.ts";

export { validateSDL, validationSDLSchema } from "./SDL/validateSDL/validateSDL.ts";
export type { SDLInput } from "./SDL/validateSDL/validateSDLInput.ts";
export type { ValidationError } from "../utils/jsonSchemaValidation.ts";

export * from "./types.ts";
export { SdlValidationError } from "./SDL/SdlValidationError.ts";
