import { describe, expect, it } from "@jest/globals";
import { exec } from "child_process";
import { access, constants as fsConst, mkdir, readFile, rmdir, writeFile } from "fs/promises";
import { tmpdir } from "os";
import { delimiter as pathDelimiter, dirname, join as joinPath } from "path";
import { fileURLToPath } from "url";
import { promisify } from "util";

const __dirname = dirname(fileURLToPath(import.meta.url));
const execAsync = promisify(exec);
const BIN_PATH = joinPath(__dirname, "..", "..", "node_modules", ".bin");
const PLUGIN_SCRIPT = joinPath(__dirname, "..", "..", "script", "protoc-gen-sdk-object.ts");

describe("protoc-sdk-objec plugin", () => {
  it("generates SDK object from proto files", async () => {
    const outputDir = joinPath(tmpdir(), `ts-bufplugin-${process.pid.toString()}`);
    const protoDir = "./ts/test/functional/proto";
    const configPath = joinPath(outputDir, "buf.config.json");
    const templatePath = joinPath(outputDir, "buf.template.json");
    await mkdir(outputDir, { recursive: true });
    await writeFile(configPath, JSON.stringify({
      version: "v2",
      modules: [
        { path: "go/vendor/github.com/cosmos/cosmos-sdk/proto" },
        { path: "./ts/test/functional/proto" },
      ],
    }));
    const pluginPath = joinPath(outputDir, "protoc-gen-sdk-object.cmd");
    await writeFile(
      pluginPath,
      `@echo off\r\n"${process.execPath}" --experimental-strip-types --no-warnings "${PLUGIN_SCRIPT}" %*\r\n`,
    );
    const config = {
      version: "v2",
      clean: false,
      plugins: [
        {
          local: pluginPath,
          strategy: "all",
          out: ".",
          opt: [
            "target=ts",
            "import_extension=ts",
          ],
        },
      ],
    };
    await writeFile(templatePath, JSON.stringify(config));
    const command = [
      "buf generate",
      `--config "${configPath}"`,
      `--template "${templatePath}"`,
      `-o "${outputDir}"`,
      `--path "${protoDir}/msg.proto"`,
      `--path "${protoDir}/query.proto"`,
      protoDir,
    ].join(" ");

    try {
      await execAsync(command, {
        cwd: joinPath(__dirname, "..", "..", ".."),
        env: {
          ...process.env,
          BUF_PLUGIN_SDK_OBJECT_OUTPUT_FILE: "sdk.ts",
          PATH: `${BIN_PATH}${pathDelimiter}${process.env.PATH ?? ""}`,
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
