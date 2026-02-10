const DEFINITIONS_PREFIX = "#/definitions/";

export function humanizeErrors(errors: ValidationError[] | undefined, schema: Record<string, unknown>, errorMessages?: ErrorMessages): ValidationError[] {
  if (!errors) return [];

  const messages = new Map<string, ValidationError>();
  errors.forEach((error) => {
    let getErrorMessage = getDefaultErrorMessage;

    if (errorMessages && Object.hasOwn(errorMessages, error.schemaPath)) {
      getErrorMessage = errorMessages[error.schemaPath];
    }

    if (errorMessages && error.schemaPath.startsWith(DEFINITIONS_PREFIX)) {
      const endIndex = error.schemaPath.indexOf("/", DEFINITIONS_PREFIX.length);
      const key = endIndex === -1 ? error.schemaPath : error.schemaPath.slice(0, endIndex);
      if (Object.hasOwn(errorMessages, key)) {
        getErrorMessage = errorMessages[key];
      }
    }

    const message = getErrorMessage(error, schema);
    if (message) messages.set(message, { ...error, message });
  });

  return Array.from(messages.values());
}

export interface ValidationError {
  schemaPath: string;
  instancePath: string;
  keyword: string;
  /** keyword parameters */
  params: Record<string, any>; // eslint-disable-line @typescript-eslint/no-explicit-any
  message: string;
}

function getDefaultErrorMessage(error: ValidationError, schema: Record<string, unknown>): string | undefined {
  if (error.keyword === "anyOf" || error.keyword === "oneOf") {
    // ignore anyOf and oneOf, since they are expressed by inner schema errors
    return;
  }

  if (error.keyword === "required") {
    return `Missing required field: "${error.params.missingProperty}"${getErrorLocation(error.instancePath)}.`;
  }
  if (error.keyword === "pattern") {
    return `Invalid format: "${getFieldName(error.instancePath)}"${getErrorLocation(dirname(error.instancePath))} does not match pattern "${error.params.pattern}"`;
  }
  if (error.keyword === "additionalProperties") {
    const patternProperties = getSchemaFieldByPath<Record<string, unknown> | undefined>(`${dirname(error.schemaPath)}/patternProperties`, schema);
    if (patternProperties) {
      return `Field "${error.params.additionalProperty}"${getErrorLocation(error.instancePath)} doesn't satisfy any of the allowed patterns: ${Object.keys(patternProperties).join(", ")}.`;
    }

    return `Additional property "${error.params.additionalProperty}" is not allowed${getErrorLocation(error.instancePath)}.`;
  }
  if (error.keyword === "type") {
    return `"${getFieldName(error.instancePath)}"${getErrorLocation(dirname(error.instancePath))} should be ${getSchemaFieldByPath(error.schemaPath, schema)}.`;
  }
  if (error.keyword === "enum") {
    return `"${getFieldName(error.instancePath)}"${getErrorLocation(dirname(error.instancePath))} should be one of: ${error.params.allowedValues.join(", ")}.`;
  }
  if (error.keyword === "minLength") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must be at least ${getSchemaFieldByPath(error.schemaPath, schema)} characters long.`;
  }
  if (error.keyword === "maxLength") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must be at most ${getSchemaFieldByPath(error.schemaPath, schema)} characters long.`;
  }
  if (error.keyword === "minItems") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must have at least ${getSchemaFieldByPath(error.schemaPath, schema)} items.`;
  }
  if (error.keyword === "maxItems") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must have at most ${getSchemaFieldByPath(error.schemaPath, schema)} items.`;
  }
  if (error.keyword === "minimum") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must be at least ${getSchemaFieldByPath(error.schemaPath, schema)}.`;
  }
  if (error.keyword === "exclusiveMinimum") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must be greater than ${getSchemaFieldByPath(error.schemaPath, schema)}.`;
  }
  if (error.keyword === "exclusiveMaximum") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must be less than ${getSchemaFieldByPath(error.schemaPath, schema)}.`;
  }
  if (error.keyword === "maximum") {
    return `"${getFieldName(error.instancePath)}" at "${dirname(error.instancePath)}" must be at most ${getSchemaFieldByPath(error.schemaPath, schema)}.`;
  }
  if (error.keyword === "minProperties") {
    const suffix = error.params.limit === 1 ? "property" : "properties";
    return `"${getFieldName(error.instancePath)}"${getErrorLocation(dirname(error.instancePath))} must have at least ${error.params.limit} ${suffix}.`;
  }
  if (error.keyword === "const") {
    return `"${getFieldName(error.instancePath)}"${getErrorLocation(dirname(error.instancePath))} must be ${error.params.allowedValue}.`;
  }

  return `"${getFieldName(error.instancePath)}"${getErrorLocation(dirname(error.instancePath))} ${error.message}.`;
}

function getFieldName(instanceLocation: string): string {
  return basename(instanceLocation);
}

export function basename(path: string): string {
  const lastPartIndex = path.lastIndexOf("/");
  if (lastPartIndex === -1) return path;
  return path.slice(lastPartIndex + 1);
}

export function dirname(path: string): string {
  const lastPartIndex = path.lastIndexOf("/");
  if (lastPartIndex === -1) return path;
  return path.slice(0, lastPartIndex);
}

function getSchemaFieldByPath<T = string>(keywordLocation: string, schema?: Record<string, unknown>): T {
  return keywordLocation.split("/").slice(1)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .reduce<T>((schema, key) => (schema as Record<string, any>)[key], schema as any);
}

export type ValidationFunction = {
  (data: unknown): boolean;
  errors?: ValidationError[];
};

export function getErrorLocation(path: string): string {
  return path ? ` at "${path}"` : "";
}

export type ErrorMessages = Record<string, (error: ValidationError, schema: Record<string, unknown>) => string | undefined>;
