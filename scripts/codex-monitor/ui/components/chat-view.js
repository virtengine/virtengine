/* ─────────────────────────────────────────────────────────────
 *  Component: Chat View — ChatGPT-style message interface
 * ────────────────────────────────────────────────────────────── */
import { h } from "preact";
import { useState, useEffect, useRef, useCallback } from "preact/hooks";
import htm from "htm";
import { apiFetch } from "../modules/api.js";
import { showToast } from "../modules/state.js";
import { formatRelative, truncate } from "../modules/utils.js";
import {
  sessionMessages,
  loadSessionMessages,
  selectedSessionId,
  sessionsData,
} from "./session-list.js";

const html = htm.bind(h);

/* ─── Inline markdown formatting ─── */
function applyInline(text) {
  text = text.replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>');
  text = text.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  text = text.replace(/__(.+?)__/g, '<strong>$1</strong>');
  text = text.replace(/\*(.+?)\*/g, '<em>$1</em>');
  text = text.replace(/(?<![a-zA-Z0-9])_(.+?)_(?![a-zA-Z0-9])/g, '<em>$1</em>');
  text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label, url) => {
    if (/^(https?:|mailto:|\/|#)/.test(url)) {
      return `<a href="${url}" target="_blank" rel="noopener" class="md-link">${label}</a>`;
    }
    return `${label} (${url})`;
  });
  return text;
}

/* ─── Convert markdown text to HTML ─── */
function renderMarkdown(text) {
  const codes = [];
  let s = text.replace(/`([^`\n]+)`/g, (_, c) => {
    codes.push(c);
    return `%%ICODE${codes.length - 1}%%`;
  });

  s = s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

  const lines = s.split('\n');
  const out = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    let m;

    if ((m = line.match(/^(#{1,3}) (.+)$/))) {
      const lvl = m[1].length;
      out.push(`<div class="md-heading md-h${lvl}">${applyInline(m[2])}</div>`);
      i++; continue;
    }

    if (/^-{3,}\s*$/.test(line.trim())) {
      out.push('<hr class="md-hr"/>');
      i++; continue;
    }

    if (/^&gt;\s?/.test(line)) {
      const q = [];
      while (i < lines.length && /^&gt;\s?/.test(lines[i])) {
        q.push(applyInline(lines[i].replace(/^&gt;\s?/, '')));
        i++;
      }
      out.push(`<div class="md-blockquote">${q.join('<br/>')}</div>`);
      continue;
    }

    if (/^[-*] /.test(line)) {
      const items = [];
      while (i < lines.length && /^[-*] /.test(lines[i])) {
        items.push(`<li>${applyInline(lines[i].replace(/^[-*] /, ''))}</li>`);
        i++;
      }
      out.push(`<ul class="md-list">${items.join('')}</ul>`);
      continue;
    }

    if (/^\d+\.\s/.test(line)) {
      const items = [];
      while (i < lines.length && /^\d+\.\s/.test(lines[i])) {
        items.push(`<li>${applyInline(lines[i].replace(/^\d+\.\s/, ''))}</li>`);
        i++;
      }
      out.push(`<ol class="md-list md-ol">${items.join('')}</ol>`);
      continue;
    }

    out.push(applyInline(line));
    i++;
  }

  let result = out.join('\n').replace(/\n/g, '<br/>');

  result = result.replace(/%%ICODE(\d+)%%/g, (_, idx) => {
    const c = codes[parseInt(idx)]
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return `<span class="md-inline-code">${c}</span>`;
  });

  return result;
}

/* ─── Code block copy button ─── */
function CodeBlock({ code }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = useCallback(() => {
    try {
      navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch { /* noop */ }
  }, [code]);

  return html`
    <div class="chat-code-block">
      <button class="chat-code-copy" onClick=${handleCopy}>
        ${copied ? "✓" : "📋"}
      </button>
      <pre><code>${code}</code></pre>
    </div>
  `;
}

/* ─── Render message content with code block + markdown support ─── */
function MessageContent({ text }) {
  if (!text) return null;
  const parts = text.split(/(```[\s\S]*?```)/g);
  return html`${parts.map((part, i) => {
    if (part.startsWith("```") && part.endsWith("```")) {
      const code = part.slice(3, -3).replace(/^\w+\n/, "");
      return html`<${CodeBlock} key=${i} code=${code} />`;
    }
    return html`<div key=${i} class="md-rendered" dangerouslySetInnerHTML=${{ __html: renderMarkdown(part) }} />`;
  })}`;
}

/* ─── Chat View component ─── */
export function ChatView({ sessionId }) {
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(false);
  const messagesRef = useRef(null);
  const inputRef = useRef(null);
  const messages = sessionMessages.value || [];

  const session = (sessionsData.value || []).find((s) => s.id === sessionId);
  const isActive =
    session?.status === "active" || session?.status === "running";

  /* Load messages on mount and poll while active */
  useEffect(() => {
    if (!sessionId) return;
    let active = true;
    setLoading(true);
    loadSessionMessages(sessionId).finally(() => {
      if (active) setLoading(false);
    });

    const interval = setInterval(() => {
      if (active) loadSessionMessages(sessionId);
    }, 3000);

    return () => {
      active = false;
      clearInterval(interval);
    };
  }, [sessionId]);

  /* Auto-scroll to bottom */
  useEffect(() => {
    if (messagesRef.current) {
      messagesRef.current.scrollTop = messagesRef.current.scrollHeight;
    }
  }, [messages.length]);

  const handleSend = useCallback(async () => {
    const text = input.trim();
    if (!text || sending) return;

    /* Optimistically add user message */
    const optimistic = {
      id: `opt-${Date.now()}`,
      role: "user",
      content: text,
      timestamp: new Date().toISOString(),
    };
    sessionMessages.value = [...sessionMessages.value, optimistic];
    setInput("");
    setSending(true);

    try {
      await apiFetch(`/api/sessions/${sessionId}/message`, {
        method: "POST",
        body: JSON.stringify({ content: text }),
      });
      await loadSessionMessages(sessionId);
    } catch {
      showToast("Failed to send message", "error");
    } finally {
      setSending(false);
    }
  }, [input, sending, sessionId]);

  const handleResume = useCallback(async () => {
    try {
      await apiFetch(`/api/sessions/${sessionId}/resume`, { method: "POST" });
      showToast("Session resumed", "success");
      await loadSessionMessages(sessionId);
    } catch {
      showToast("Failed to resume session", "error");
    }
  }, [sessionId]);

  const handleKeyDown = useCallback(
    (e) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend],
  );

  const handleInput = useCallback((e) => {
    setInput(e.target.value);
    const el = e.target;
    el.style.height = 'auto';
    el.style.height = Math.min(el.scrollHeight, 100) + 'px';
  }, []);

  if (!sessionId) {
    return html`
      <div class="chat-view chat-empty-state">
        <div class="session-empty-icon">💬</div>
        <div class="session-empty-text">Select a session to view messages</div>
      </div>
    `;
  }

  return html`
    <div class="chat-view">
      <div class="chat-header">
        <div class="chat-header-title">
          ${session?.title || session?.taskId || "Session"}
        </div>
        <div class="chat-header-meta">
          ${session?.type || "manual"} · ${session?.status || "unknown"}
        </div>
      </div>

      <div class="chat-messages" ref=${messagesRef}>
        ${loading && messages.length === 0 && html`
          <div class="chat-loading">Loading messages…</div>
        `}
        ${messages.map(
          (msg) => html`
            <div
              key=${msg.id || msg.timestamp}
              class="chat-bubble ${msg.role === "user"
                ? "user"
                : msg.role === "system"
                  ? "system"
                  : "assistant"}"
            >
              ${msg.role === "system"
                ? html`<div class="chat-system-text">${msg.content}</div>`
                : html`
                    <div class="chat-bubble-content">
                      <${MessageContent} text=${msg.content} />
                    </div>
                    <div class="chat-bubble-time">
                      ${formatRelative(msg.timestamp)}
                    </div>
                  `}
            </div>
          `,
        )}
        ${sending && html`
          <div class="chat-bubble assistant">
            <div class="chat-typing">
              <span class="chat-typing-dot"></span>
              <span class="chat-typing-dot"></span>
              <span class="chat-typing-dot"></span>
            </div>
          </div>
        `}
      </div>

      <div class="chat-input-bar">
        ${!isActive && session?.status &&
        html`
          <button class="btn btn-primary btn-sm chat-resume-btn" onClick=${handleResume}>
            ▶ Resume Session
          </button>
        `}
        <div class="chat-input-row">
          <textarea
            ref=${inputRef}
            class="input chat-input"
            placeholder="Send a message…"
            rows="1"
            value=${input}
            onInput=${handleInput}
            onKeyDown=${handleKeyDown}
          />
          <button
            class="btn btn-primary chat-send-btn"
            disabled=${!input.trim() || sending}
            onClick=${handleSend}
          >
            ➤
          </button>
        </div>
      </div>
    </div>
  `;
}
