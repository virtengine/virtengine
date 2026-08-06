import { execFile } from "child_process";
import { existsSync } from "fs";
import { cp, mkdtemp, rm, writeFile } from "fs/promises";
import { tmpdir } from "os";
import { isAbsolute, join as joinPath, relative as relativePath, resolve as resolvePath } from "path";
import { promisify } from "util";

const execFileAsync = promisify(execFile);

interface BufModule {
  path: string;
}

interface BufGenerateConfig {
  version: string;
  modules: BufModule[];
}

interface BufGenerateTemplate {
  version: string;
  clean?: boolean;
  plugins: Array<{
    local: string | string[];
    strategy: string;
    out: string;
    include_imports?: boolean;
    opt?: string[];
  }>;
}

interface RunBufGenerateOptions {
  config: BufGenerateConfig;
  template: BufGenerateTemplate;
  cwd: string;
  outputDir: string;
  paths: string[];
  inputs: string[];
  env?: NodeJS.ProcessEnv;
}

export async function runBufGenerate(options: RunBufGenerateOptions) {
  // Inline configs still inherit a parent buf.lock. That caused the harness's
  // local Cosmos/gogoproto modules to collide with the same BSR dependencies
  // from sdk/buf.lock. A lock-free temporary workspace isolates each harness
  // while preserving paths relative to the repository working directory.
  const configDir = await mkdtemp(joinPath(tmpdir(), "virtengine-buf-test-"));
  const configPath = joinPath(configDir, "buf.yaml");
  const templatePath = joinPath(configDir, "buf.gen.yaml");

  try {
    const stagedModules = await Promise.all(options.config.modules.map(async (module, index) => {
      const source = resolvePath(options.cwd, module.path);
      const destination = joinPath(configDir, "modules", index.toString());
      await cp(source, destination, { recursive: true });
      await Promise.all([
        rm(joinPath(destination, "buf.yaml"), { force: true }),
        rm(joinPath(destination, "buf.lock"), { force: true }),
      ]);
      return { source, destination };
    }));
    const stagedConfig = {
      ...options.config,
      modules: stagedModules.map((module) => ({ path: relativePath(configDir, module.destination) })),
    };
    const stagedTemplate = {
      ...options.template,
      plugins: options.template.plugins.map((plugin) => ({
        ...plugin,
        local: Array.isArray(plugin.local)
          ? plugin.local.map((argument) => {
              const candidate = resolvePath(options.cwd, argument);
              return existsSync(candidate) ? candidate : argument;
            })
          : plugin.local,
      })),
    };
    const stagedInputs = options.inputs.map((input) => {
      const absoluteInput = resolvePath(options.cwd, input);
      const module = stagedModules.find((candidate) => candidate.source === absoluteInput);
      if (!module) throw new Error(`Buf test input ${input} is not one of the configured local modules`);
      return relativePath(configDir, module.destination);
    });
    const stagedPaths = options.paths.map((path) => {
      const absolutePath = isAbsolute(path) ? path : resolvePath(options.cwd, path);
      const input = options.inputs
        .map((value) => resolvePath(options.cwd, value))
        .find((value) => {
          const pathFromInput = relativePath(value, absolutePath);
          return pathFromInput === "" || (!pathFromInput.startsWith("..") && !isAbsolute(pathFromInput));
        });
      if (!input) throw new Error(`Buf test path ${path} is outside the requested inputs`);
      const module = stagedModules.find((candidate) => candidate.source === input);
      if (!module) throw new Error(`Buf test path ${path} is outside the configured local modules`);
      return relativePath(configDir, joinPath(module.destination, relativePath(input, absolutePath)));
    });

    await Promise.all([
      writeFile(configPath, JSON.stringify(stagedConfig)),
      writeFile(templatePath, JSON.stringify(stagedTemplate)),
    ]);

    const args = [
      "generate",
      "--config",
      configPath,
      "--template",
      templatePath,
      "-o",
      options.outputDir,
      ...stagedPaths.flatMap((path) => ["--path", path]),
      ...stagedInputs,
    ];

    await execFileAsync("buf", args, {
      cwd: configDir,
      env: options.env ?? process.env,
    });
  } finally {
    await rm(configDir, { recursive: true, force: true });
  }
}
