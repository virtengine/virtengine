import { describe, expect, it } from "@jest/globals";
import { exec } from "child_process";
import { access, constants as fsConst, readFile, rmdir } from "fs/promises";
import { tmpdir } from "os";
import { join as joinPath } from "path";
import { promisify } from "util";

const execAsync = promisify(exec);

describe("protoc-gen-customtype-patches plugin", () => {
  const config = {
    version: "v2",
    clean: true,
    plugins: [
      {
        local: "ts/script/protoc-gen-customtype-patches.ts",
        strategy: "all",
        out: ".",
        opt: [
          "target=ts",
          "import_extension=ts"
        ],
      },
    ],
  };

  it("generates `Set` instance with all the types that have reference to fields with custom type option", async () => {
    const outputDir = joinPath(tmpdir(), `ts-bufplugin-${process.pid.toString()}`);
    const protoDir = "./ts/test/functional/proto";
    const command = [
      `buf generate`,
      `--config '${JSON.stringify({
        version: "v2",
        modules: [
          { path: "go/vendor/github.com/cosmos/gogoproto" },
          { path: "./ts/test/functional/proto" },
        ],
      })}'`,
      `--template '${JSON.stringify(config)}'`,
      `-o '${outputDir}'`,
      `--path ${protoDir}/customtype.proto`,
      protoDir,
    ].join(" ");

    try {
      await execAsync(command, {
        cwd: joinPath(__dirname, "..", "..", ".."),
        env: {
          ...process.env,
          BUF_PLUGIN_CUSTOMTYPE_TYPES_PATCHES_OUTPUT_FILE: "customPatches.ts",
        },
      });

      expect(await readFile(joinPath(outputDir, "customPatches.ts"), "utf-8")).toMatchSnapshot();
    } finally {
      if (await access(outputDir, fsConst.W_OK).catch(() => false)) {
        await rmdir(outputDir, { recursive: true });
      }
    }
  });
});
