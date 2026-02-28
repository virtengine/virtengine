import { describe, expect, it } from "@jest/globals";
import { access, constants as fsConst, readFile, rmdir } from "fs/promises";
import { tmpdir } from "os";
import { join as joinPath } from "path";

import { runBufGenerate } from "../helpers/runBufGenerate.ts";

describe("protoc-sdk-objec plugin", () => {
  const config = {
    version: "v2",
    clean: true,
    plugins: [
      {
        local: [
          "node",
          "--experimental-strip-types",
          "--no-warnings",
          "ts/script/protoc-gen-sdk-object.ts",
        ],
        strategy: "all",
        out: ".",
        opt: [
          "target=ts",
          "import_extension=ts",
        ],
      },
    ],
  };

  it("generates SDK object from proto files", async () => {
    const outputDir = joinPath(tmpdir(), `ts-bufplugin-${process.pid.toString()}`);
    const protoDir = "./ts/test/functional/proto";

    try {
      await runBufGenerate({
        config: {
          version: "v2",
          modules: [
            { path: "go/vendor/github.com/cosmos/cosmos-sdk/proto" },
            { path: protoDir },
          ],
        },
        template: config,
        cwd: joinPath(__dirname, "..", "..", ".."),
        outputDir,
        paths: [`${protoDir}/msg.proto`, `${protoDir}/query.proto`],
        inputs: [protoDir],
        env: {
          ...process.env,
          BUF_PLUGIN_SDK_OBJECT_OUTPUT_FILE: "sdk.ts",
        },
      });

      expect(await readFile(joinPath(outputDir, "sdk.ts"), "utf-8")).toMatchSnapshot();
      expect(await readFile(joinPath(outputDir, "protos", "msg_virtengine.ts"), "utf-8")).toMatchSnapshot();
      expect(await readFile(joinPath(outputDir, "protos", "query_virtengine.ts"), "utf-8")).toMatchSnapshot();
    } finally {
      if (await access(outputDir, fsConst.W_OK).catch(() => false)) {
        await rmdir(outputDir, { recursive: true });
      }
    }
  });
});
