// Copyright 2026 VirtEngine Authors
import { spawn } from "node:child_process";
import { createWriteStream, existsSync, mkdirSync } from "node:fs";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));
const repoRoot = resolve(__dirname, "..");
const logDir = resolve(repoRoot, ".cache", "telegram-bot");

const telegramToken = process.env.TELEGRAM_BOT_TOKEN;
const allowedChatId = process.env.TELEGRAM_CHAT_ID;
const pollTimeoutSec = Number(process.env.TELEGRAM_POLL_TIMEOUT_SEC || "25");
const pollIntervalMs = Number(process.env.TELEGRAM_POLL_INTERVAL_MS || "1500");

if (!telegramToken) {
  console.error("Missing TELEGRAM_BOT_TOKEN environment variable.");
  process.exit(1);
}

if (!existsSync(logDir)) {
  mkdirSync(logDir, { recursive: true });
}

const apiBase = `https://api.telegram.org/bot${telegramToken}`;
const jobs = new Map();
let updateOffset = 0;

function nowStamp() {
  return new Date().toISOString().replace(/[:.]/g, "-");
}

function buildJobId() {
  const rand = Math.random().toString(36).slice(2, 6);
  return `bg-${nowStamp()}-${rand}`;
}

function truncate(text, maxLen) {
  if (!text) return "";
  if (text.length <= maxLen) return text;
  return `${text.slice(0, maxLen - 12)}\n...(truncated)`;
}

async function sendMessage(chatId, text) {
  const payload = {
    chat_id: chatId,
    text: truncate(text, 3800),
    disable_web_page_preview: true,
  };
  try {
    const res = await fetch(`${apiBase}/sendMessage`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const body = await res.json();
    if (!body.ok) {
      console.warn(`[telegram-bot] send failed: ${body.description || "unknown"}`);
    }
  } catch (err) {
    console.warn(`[telegram-bot] send failed: ${err.message || err}`);
  }
}

async function readOptional(path) {
  try {
    return await readFile(path, "utf8");
  } catch {
    return "";
  }
}

async function getTail(path, maxChars) {
  const content = await readOptional(path);
  if (!content) return "";
  return content.length > maxChars ? content.slice(-maxChars) : content;
}

function spawnBackgroundJob(prompt, chatId) {
  const id = buildJobId();
  const logPath = resolve(logDir, `${id}.log`);
  const lastMessagePath = resolve(logDir, `${id}.last.txt`);
  const startedAt = Date.now();

  const args = [
    "exec",
    "--full-auto",
    "-C",
    repoRoot,
    "--output-last-message",
    lastMessagePath,
    prompt,
  ];

  const child = spawn("codex", args, {
    cwd: repoRoot,
    env: { ...process.env },
    shell: true,
    detached: true,
    stdio: ["ignore", "pipe", "pipe"],
  });

  const logStream = createWriteStream(logPath, { flags: "a" });
  child.stdout.on("data", (chunk) => {
    logStream.write(chunk);
  });
  child.stderr.on("data", (chunk) => {
    logStream.write(chunk);
  });

  const job = {
    id,
    chatId,
    prompt,
    pid: child.pid,
    logPath,
    lastMessagePath,
    startedAt,
    status: "running",
  };
  jobs.set(id, job);

  child.on("error", async (err) => {
    job.status = "failed";
    await sendMessage(
      chatId,
      `Background job ${id} failed to start: ${err.message || err}`,
    );
    jobs.delete(id);
  });

  child.on("exit", async (code, signal) => {
    job.status = code === 0 ? "completed" : "failed";
    const durationSec = Math.max(1, Math.round((Date.now() - startedAt) / 1000));
    const lastMessage = (await readOptional(lastMessagePath)).trim();
    const tail = lastMessage ? "" : (await getTail(logPath, 2400)).trim();
    const statusLine =
      code === 0
        ? `completed in ${durationSec}s`
        : `failed in ${durationSec}s (exit ${code ?? "?"}${signal ? ` signal ${signal}` : ""})`;
    const body = lastMessage || tail || "(no output captured)";
    await sendMessage(chatId, `Background job ${id} ${statusLine}.\n\n${body}`);
    jobs.delete(id);
  });

  return job;
}

function normalizeCommand(text) {
  const trimmed = text.trim();
  if (!trimmed.startsWith("/")) return null;
  const parts = trimmed.split(/\s+/);
  const command = parts[0].split("@")[0];
  const prompt = parts.slice(1).join(" ").trim();
  return { command, prompt };
}

async function handleMessage(message) {
  if (!message?.text) return;
  const chatId = message.chat?.id;
  if (!chatId) return;
  if (allowedChatId && String(chatId) !== String(allowedChatId)) {
    return;
  }

  const cmd = normalizeCommand(message.text);
  if (!cmd) return;

  if (cmd.command === "/background") {
    if (!cmd.prompt) {
      await sendMessage(chatId, "Usage: /background <prompt>");
      return;
    }
    const job = spawnBackgroundJob(cmd.prompt, chatId);
    await sendMessage(
      chatId,
      `Background job ${job.id} started (pid ${job.pid ?? "unknown"}).`,
    );
    return;
  }

  if (cmd.command === "/help") {
    await sendMessage(chatId, "Commands:\n/background <prompt>");
    return;
  }

  await sendMessage(chatId, `Unknown command: ${cmd.command}`);
}

async function pollOnce() {
  const payload = {
    offset: updateOffset || undefined,
    timeout: pollTimeoutSec,
    allowed_updates: ["message"],
  };
  try {
    const res = await fetch(`${apiBase}/getUpdates`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const body = await res.json();
    if (!body.ok) {
      console.warn(`[telegram-bot] getUpdates failed: ${body.description || "unknown"}`);
      return;
    }
    for (const update of body.result || []) {
      updateOffset = Math.max(updateOffset, update.update_id + 1);
      await handleMessage(update.message);
    }
  } catch (err) {
    console.warn(`[telegram-bot] polling error: ${err.message || err}`);
  }
}

async function run() {
  console.log("[telegram-bot] started.");
  for (;;) {
    await pollOnce();
    await new Promise((resolve) => setTimeout(resolve, pollIntervalMs));
  }
}

run();
