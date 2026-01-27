import { ScalarType } from "@bufbuild/protobuf";
import { faker } from "@faker-js/faker";

import { encodeBinary } from "../../src/encoding/binaryEncoding.ts";

export interface Field {
  name: string;
  kind: string;
  scalarType?: number;
  customType?: string;
  enum?: string[];
  mapKeyType?: number;
  message?: MessageSchema;
};

export type MessageSchema = {
  fields: Field[],
  type: {
    fromPartial(attrs: Record<string, unknown>): unknown,
    $type: string,
  },
};

/**
 * Used in generated tests for custom types patches
 */
export function generateMessage(typeName: string, messageToFields: Record<string, MessageSchema>) {
  const messageSchema = messageToFields[typeName];
  if (!messageSchema) {
    throw new Error(`Message ${typeName} not found`);
  }
  const attrs = messageSchema.fields.reduce<Record<string, unknown>>((acc, field) => {
    acc[field.name] = generateField(field, messageToFields);
    return acc;
  }, {});

  return messageSchema.type.fromPartial(attrs);
}

function generateField(field: Field, messageToFields: Record<string, MessageSchema>) {
  switch (field.kind) {
    case "scalar":
      return generateScalar(field, field.scalarType!);
    case "enum":
      return faker.number.int({ min: 0, max: field.enum!.length - 1 });
    case "message":
      return generateMessage(field.message!.type.$type, messageToFields);
    case "list":
      return Array.from({ length: faker.number.int({ min: 0, max: 10 }) }, () => generateMessage(field.message!.type.$type, messageToFields));
    case "map":
      return Array.from({ length: faker.number.int({ min: 0, max: 10 }) }).reduce<Record<PropertyKey, unknown>>((map) => {
        const key = generateScalar(field, field.mapKeyType!);
        map[key as string] = generateMessage(field.message!.type.$type, messageToFields);
        return map;
      }, {});
    default:
      throw new Error(`Unknown field kind: ${field.kind}`);
  }
}

function generateScalar(field: Field, scalarType: ScalarType) {
  switch (scalarType) {
    case ScalarType.STRING:
      return guessFakeValue(field);
    case ScalarType.BYTES:
      return encodeBinary(String(guessFakeValue(field)));
    case ScalarType.SFIXED32:
    case ScalarType.SINT32:
    case ScalarType.INT32:
      return faker.number.int({ min: -1000000, max: 1000000 });
    case ScalarType.SFIXED64:
    case ScalarType.SINT64:
    case ScalarType.INT64:
      return faker.number.bigInt({ min: -1000000000n, max: 100000000000n });
    case ScalarType.UINT32:
      return faker.number.int({ min: 0, max: 1000000 });
    case ScalarType.UINT64:
      return faker.number.bigInt({ min: 0, max: 1000000n });
    case ScalarType.FLOAT:
      return faker.number.float({ min: -1000000, max: 1000000 });
    case ScalarType.DOUBLE:
      return faker.number.float({ min: -1000000, max: 1000000 });
    case ScalarType.BOOL:
      return faker.datatype.boolean();
    default:
      throw new Error(`Unknown scalar type: ${field.scalarType}`);
  }
}

function guessFakeValue(field: Field): unknown {
  const lowerName = field.name.toLowerCase();

  if (field.customType) {
    const value = guessFakeValueForCustomType(field.customType);
    if (value !== null) return value;
  }
  if (lowerName.includes("name")) return faker.person.fullName();
  if (lowerName.includes("first")) return faker.person.firstName();
  if (lowerName.includes("last")) return faker.person.lastName();
  if (lowerName.includes("email")) return faker.internet.email();
  if (lowerName.includes("phone")) return faker.phone.number();
  if (lowerName.includes("address")) return faker.location.streetAddress();
  if (lowerName.includes("city")) return faker.location.city();
  if (lowerName.includes("zip") || lowerName.includes("postal")) return faker.location.zipCode();
  if (lowerName.includes("country")) return faker.location.country();
  if (lowerName.includes("date")) return faker.date.past().toISOString();
  if (lowerName.includes("id")) return faker.string.uuid();
  if (lowerName.includes("username")) return faker.internet.userName();
  if (lowerName.includes("password")) return faker.internet.password();
  if (lowerName.includes("avatar")) return faker.image.avatar();
  if (lowerName.includes("denom")) return "uakt";
  if (lowerName.includes("amount") || lowerName.includes("commission") || lowerName.includes("price") || lowerName.includes("rate")) {
    return faker.number.float({ min: 0, max: 1000000 }).toString();
  }

  return faker.lorem.word();
}

function guessFakeValueForCustomType(shortName: string | undefined) {
  switch (shortName) {
    case "LegacyDec":
      return faker.number.float({ min: 0, max: 1000000 }).toString();
    default:
      return null;
  }
}
