#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = resolve(fileURLToPath(new URL(".", import.meta.url)));

function hasArg(flag) {
  return process.argv.includes(flag);
}

function getRegistryUrl() {
  const raw =
    process.env.NPM_REGISTRY_URL ||
    process.env.npm_config_registry ||
    "https://registry.npmjs.org/";
  const parsed = new URL(raw);
  if (!parsed.pathname.endsWith("/")) {
    parsed.pathname = `${parsed.pathname}/`;
  }
  return parsed.toString();
}

function createEphemeralNpmrc(registryUrl, token) {
  const folder = mkdtempSync(join(tmpdir(), "codex-monitor-npmrc-"));
  const npmrcPath = join(folder, ".npmrc");
  const parsed = new URL(registryUrl);
  const authPath = parsed.pathname || "/";
  const authLine = `//${parsed.host}${authPath}:_authToken=${token}`;
  const content = [
    `registry=${registryUrl}`,
    "always-auth=true",
    authLine,
  ].join("\n");
  writeFileSync(npmrcPath, `${content}\n`, "utf8");
  return { folder, npmrcPath };
}

function run(command, args, env) {
  const result = spawnSync(command, args, {
    stdio: "inherit",
    cwd: SCRIPT_DIR,
    env,
    shell: true,
  });
  if (result.error) {
    console.error(`[publish] Failed to execute ${command}: ${result.error.message}`);
    return 1;
  }
  return result.status ?? 1;
}

const NPM_BIN = "npm";

function main() {
  const dryRun = hasArg("--dry-run");
  const tag = process.env.NPM_PUBLISH_TAG || "latest";
  const otp = process.env.NPM_OTP || "";
  const access = process.env.NPM_PUBLISH_ACCESS || "public";
  const registry = getRegistryUrl();
  const token = process.env.NPM_ACCESS_TOKEN || process.env.NODE_AUTH_TOKEN || "";

  if (!dryRun && !token) {
    console.error(
      "[publish] Missing token. Set NPM_ACCESS_TOKEN (or NODE_AUTH_TOKEN) in environment. - use a 2FA Permissionless token",
    );
    process.exit(1);
  }

  console.log(
    `[publish] Running prepublish checks (${dryRun ? "dry-run" : "publish"})...`,
  );
  const checkStatus = run(NPM_BIN, ["run", "prepublishOnly"], process.env);
  if (checkStatus !== 0) {
    process.exit(checkStatus);
  }

  let tempConfig = null;
  try {
    const env = { ...process.env };
    if (!dryRun) {
      tempConfig = createEphemeralNpmrc(registry, token);
      env.NPM_CONFIG_USERCONFIG = tempConfig.npmrcPath;
      env.NODE_AUTH_TOKEN = token;
    }

    const publishArgs = [
      "publish",
      "--registry",
      registry,
      "--access",
      access,
      "--tag",
      tag,
    ];

    if (dryRun) {
      publishArgs.push("--dry-run");
    }
    if (otp) {
      publishArgs.push("--otp", otp);
    }

    console.log(
      `[publish] npm ${publishArgs.join(" ")} (token via env/userconfig, redacted)`,
    );
    const status = run(NPM_BIN, publishArgs, env);
    process.exit(status);
  } finally {
    if (tempConfig?.folder) {
      rmSync(tempConfig.folder, { recursive: true, force: true });
    }
  }
}

main();
