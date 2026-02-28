import { type AnyDesc, create, type DescFile, type DescMessage, fromBinary, fromJson, type Message, type MessageInitShape, type MessageShape, toBinary, toJson } from "@bufbuild/protobuf";
import type { GenMessage, GenService, GenServiceMethods } from "@bufbuild/protobuf/codegenv1";
import { BinaryWriter } from "@bufbuild/protobuf/wire";
import assert from "assert";
import { createHash } from "crypto";
import { mkdir, writeFile } from "fs/promises";
import { join as joinPath, relative as relativePath } from "path";

import type { MessageDesc, MethodDesc } from "../../src/sdk/client/types.ts";
import { runBufGenerate } from "./runBufGenerate.ts";
const PWD = joinPath(__dirname, "..", "..", "..");

const cache: Record<string, DescFileDefinition> = Object.create(null);
export async function proto(strings: TemplateStringsArray, ...values: unknown[]): Promise<DescFileDefinition> {
  const content = strings.reduce((result, string, index) => {
    return result + string + (values[index] ?? "");
  }, "");
  const fileContent = [
    "syntax = \"proto3\";",
    "package virtengine.test.unit;",
    content,
  ].join("\n");
  const hash = createHash("sha256").update(fileContent).digest("hex");
  if (cache[hash]) return cache[hash];

  const baseDir = joinPath(PWD, "ts", "node_modules", ".tmp", "virtengine");
  const outputDir = joinPath(baseDir, "generated");
  const protoDir = joinPath(baseDir, "proto");
  const filePath = joinPath(protoDir, `${hash}.proto`);
  await mkdir(protoDir, { recursive: true });
  await writeFile(filePath, fileContent);

  await runBufGenerate({
    config: {
      version: "v2",
      modules: [
        { path: "go/vendor/github.com/cosmos/gogoproto" },
        { path: relativePath(PWD, protoDir) },
      ],
    },
    template: {
      version: "v2",
      plugins: [
        {
          local: "protoc-gen-es",
          strategy: "all",
          out: ".",
          include_imports: true,
          opt: [
            "target=js",
            "js_import_style=legacy_commonjs",
          ],
        },
      ],
    },
    cwd: PWD,
    outputDir,
    paths: [filePath],
    inputs: [relativePath(PWD, protoDir)],
  });

  const module = await import(joinPath(outputDir, `${hash}_pb`)) as Record<string, AnyDesc>;
  cache[hash] = new DescFileDefinition(Object.values(module).find((value) => value?.kind === "file")!);

  return cache[hash];
}

class DescFileDefinition {
  constructor(public readonly file: DescFile) {}

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getMessage<Type extends string, TShape = Record<string, any>>(name: Type): GenMessage<Message<`virtengine.test.unit.${Type}`> & TShape> {
    const message = this.file.messages.find((type) => type.name === name);
    assert(message, `Message with name ${name} not found in this proto file`);
    return message as GenMessage<Message<`virtengine.test.unit.${Type}`> & TShape>;
  }

  getService<T extends GenServiceMethods>(name: string): GenService<T> {
    const service = this.file.services.find((type) => type.name === name);
    assert(service, `Service with name ${name} not found in this proto file`);
    return service as GenService<T>;
  }

  /**
   * Service representation generated from ts-proto generator types
   */
  getTsProtoService<T extends GenServiceMethods>(name: string): ServiceDescFrom<T> {
    const service = this.getService(name);
    const serviceDesc = { typeName: service.typeName, methods: {} as Record<string, unknown> };

    service.methods.forEach((method) => {
      serviceDesc.methods[method.localName] = {
        kind: method.methodKind,
        name: method.name,
        input: createMessageDesc(method.input),
        output: createMessageDesc(method.output),
        parent: serviceDesc,
      };
    });
    return serviceDesc as ServiceDescFrom<T>;
  }
}

type ServiceDescFrom<T extends GenServiceMethods> = {
  typeName: string;
  methods: {
    [K in keyof T]: MethodDesc<T[K]["methodKind"], MessageDesc<MessageInitShape<T[K]["input"]>, T[K]["input"]["typeName"]>, MessageDesc<MessageInitShape<T[K]["output"]>, T[K]["output"]["typeName"]>>;
  };
};

/**
 * Creates a ts-proto message desc from a bufbuild desc message.
 */
function createMessageDesc<T extends DescMessage>(schema: T): MessageDesc<MessageInitShape<T>> {
  return {
    $type: schema.typeName,
    encode(message, writer = new BinaryWriter()) {
      const object: MessageShape<T> = message.$typeName ? message as MessageShape<T> : create(schema, message as MessageInitShape<T>);
      const bytes = toBinary(schema, object);
      writer.raw(bytes);
      return writer;
    },
    decode(input) {
      const bytes = input instanceof Uint8Array ? input : (input as unknown as { buf: Uint8Array }).buf;
      return fromBinary(schema, bytes);
    },
    fromPartial(message) {
      const { $typeName: _typeName, ...rest } = create(schema, message as MessageInitShape<T>);
      return rest as unknown as MessageInitShape<T>;
    },
    toJSON(message) {
      return toJson(schema, create(schema, message as MessageInitShape<T>));
    },
    fromJSON(message) {
      return fromJson(schema, message);
    },
  };
}
