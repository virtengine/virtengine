import { execFile } from "child_process";
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
  const args = [
    "generate",
    "--config",
    JSON.stringify(options.config),
    "--template",
    JSON.stringify(options.template),
    "-o",
    options.outputDir,
    ...options.paths.flatMap((path) => ["--path", path]),
    ...options.inputs,
  ];

  await execFileAsync("buf", args, {
    cwd: options.cwd,
    env: options.env ?? process.env,
  });
}
