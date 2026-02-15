/* ─────────────────────────────────────────────────────────────
 *  VirtEngine Control Center – Shared UI Components
 *  Card, Badge, StatCard, Modal, Toast, EmptyState, etc.
 * ────────────────────────────────────────────────────────────── */

import { h } from "https://esm.sh/preact@10.25.4";
import {
  useState,
  useEffect,
  useRef,
} from "https://esm.sh/preact@10.25.4/hooks";
import htm from "https://esm.sh/htm@3.1.1";

const html = htm.bind(h);

import { ICONS } from "../modules/icons.js";
import { toasts, showToast } from "../modules/state.js";
import {
  haptic,
  showBackButton,
  hideBackButton,
  getTg,
} from "../modules/telegram.js";
import { classNames } from "../modules/utils.js";

/* ═══════════════════════════════════════════════
 *  Card
 * ═══════════════════════════════════════════════ */

/**
 * Card container with optional title / subtitle.
 * @param {{title?: string, subtitle?: string, children?: any, className?: string, onClick?: () => void}} props
 */
export function Card({ title, subtitle, children, className = "", onClick }) {
  return html`
    <div class="card ${className}" onClick=${onClick}>
      ${title ? html`<div class="card-title">${title}</div>` : null}
      ${subtitle ? html`<div class="card-subtitle">${subtitle}</div>` : null}
      ${children}
    </div>
  `;
}

/* ═══════════════════════════════════════════════
 *  Badge
 * ═══════════════════════════════════════════════ */

const BADGE_STATUS_MAP = new Set([
  "todo",
  "inprogress",
  "inreview",
  "done",
  "error",
  "cancelled",
  "critical",
  "high",
  "medium",
  "low",
  "log",
  "info",
  "warning",
]);

/**
 * Status badge pill.
 * @param {{status?: string, text?: string, className?: string}} props
 */
export function Badge({ status, text, className = "" }) {
  const label = text || status || "";
  const normalized = (status || "").toLowerCase().replace(/\s+/g, "");
  const statusClass = BADGE_STATUS_MAP.has(normalized)
    ? `badge-${normalized}`
    : "";
  return html`<span class="badge ${statusClass} ${className}">${label}</span>`;
}

/* ═══════════════════════════════════════════════
 *  StatCard
 * ═══════════════════════════════════════════════ */

/**
 * Stat display card with large value and small label.
 * @param {{value: any, label: string, trend?: 'up'|'down', color?: string}} props
 */
export function StatCard({ value, label, trend, color }) {
  const valueStyle = color ? `color: ${color}` : "";
  const trendIcon =
    trend === "up"
      ? html`<span class="stat-trend stat-trend-up">↑</span>`
      : trend === "down"
        ? html`<span class="stat-trend stat-trend-down">↓</span>`
        : null;

  return html`
    <div class="stat-card">
      <div class="stat-value" style=${valueStyle}>
        ${value ?? "—"}${trendIcon}
      </div>
      <div class="stat-label">${label}</div>
    </div>
  `;
}

/* ═══════════════════════════════════════════════
 *  SkeletonCard
 * ═══════════════════════════════════════════════ */

/**
 * Animated loading placeholder.
 * @param {{height?: string, className?: string}} props
 */
export function SkeletonCard({ height = "80px", className = "" }) {
  return html`
    <div
      class="skeleton skeleton-card ${className}"
      style="height: ${height}"
    ></div>
  `;
}

/* ═══════════════════════════════════════════════
 *  Modal (Bottom Sheet)
 * ═══════════════════════════════════════════════ */

/**
 * Bottom-sheet modal with drag handle, title, and TG BackButton integration.
 * @param {{title?: string, open?: boolean, onClose: () => void, children?: any}} props
 */
export function Modal({ title, open = true, onClose, children }) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    // Animate in after mount
    requestAnimationFrame(() => setVisible(open));
  }, [open]);

  // BackButton integration
  useEffect(() => {
    const tg = getTg();
    if (!tg?.BackButton) return;

    const handler = () => {
      onClose();
      tg.BackButton.hide();
      tg.BackButton.offClick(handler);
    };
    tg.BackButton.show();
    tg.BackButton.onClick(handler);

    return () => {
      tg.BackButton.hide();
      tg.BackButton.offClick(handler);
    };
  }, [onClose]);

  if (!open) return null;

  return html`
    <div
      class="modal-overlay ${visible ? "modal-overlay-visible" : ""}"
      onClick=${(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        class="modal-content ${visible ? "modal-content-visible" : ""}"
        onClick=${(e) => e.stopPropagation()}
      >
        <div class="modal-handle"></div>
        ${title ? html`<div class="modal-title">${title}</div>` : null}
        ${children}
      </div>
    </div>
  `;
}

/* ═══════════════════════════════════════════════
 *  Toast / ToastContainer
 * ═══════════════════════════════════════════════ */

/**
 * Renders all active toasts from the toasts signal.
 * Each toast auto-dismisses (handled by showToast in state.js).
 */
export function ToastContainer() {
  const items = toasts.value;
  if (!items.length) return null;

  return html`
    <div class="toast-container">
      ${items.map(
        (t) => html`
          <div key=${t.id} class="toast toast-${t.type}">
            <span class="toast-message">${t.message}</span>
            <button
              class="toast-close"
              onClick=${() => {
                toasts.value = toasts.value.filter((x) => x.id !== t.id);
              }}
            >
              ×
            </button>
          </div>
        `,
      )}
    </div>
  `;
}

/* ═══════════════════════════════════════════════
 *  EmptyState
 * ═══════════════════════════════════════════════ */

/**
 * Empty state display.
 * @param {{icon?: string, title?: string, description?: string, action?: {label: string, onClick: () => void}}} props
 */
export function EmptyState({ icon, title, description, action }) {
  const iconSvg = icon && ICONS[icon] ? ICONS[icon] : null;
  return html`
    <div class="empty-state">
      ${iconSvg ? html`<div class="empty-state-icon">${iconSvg}</div>` : null}
      ${title ? html`<div class="empty-state-title">${title}</div>` : null}
      ${description
        ? html`<div class="empty-state-description">${description}</div>`
        : null}
      ${action
        ? html`<button class="btn btn-primary btn-sm" onClick=${action.onClick}>
            ${action.label}
          </button>`
        : null}
    </div>
  `;
}

/* ═══════════════════════════════════════════════
 *  Divider
 * ═══════════════════════════════════════════════ */

/**
 * Section divider with optional centered label.
 * @param {{label?: string}} props
 */
export function Divider({ label }) {
  if (!label) return html`<div class="divider"></div>`;
  return html`
    <div class="divider divider-label">
      <span>${label}</span>
    </div>
  `;
}

/* ═══════════════════════════════════════════════
 *  Avatar
 * ═══════════════════════════════════════════════ */

/**
 * Circle avatar with initials fallback.
 * @param {{name?: string, size?: number, src?: string}} props
 */
export function Avatar({ name = "", size = 36, src }) {
  const initials = name
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w.charAt(0).toUpperCase())
    .join("");

  const style = `width:${size}px;height:${size}px;border-radius:50%;overflow:hidden;
    display:flex;align-items:center;justify-content:center;
    background:var(--accent,#5b6eae);color:var(--accent-text,#fff);
    font-size:${Math.round(size * 0.4)}px;font-weight:600;flex-shrink:0`;

  if (src) {
    return html`
      <div style=${style}>
        <img
          src=${src}
          alt=${name}
          style="width:100%;height:100%;object-fit:cover"
          onError=${(e) => {
            e.target.style.display = "none";
          }}
        />
      </div>
    `;
  }

  return html`<div style=${style}>${initials || "?"}</div>`;
}

/* ═══════════════════════════════════════════════
 *  ListItem
 * ═══════════════════════════════════════════════ */

/**
 * Generic list item for settings-style lists.
 * @param {{title: string, subtitle?: string, trailing?: any, onClick?: () => void, icon?: string}} props
 */
export function ListItem({ title, subtitle, trailing, onClick, icon }) {
  const iconSvg = icon && ICONS[icon] ? ICONS[icon] : null;
  return html`
    <div
      class=${classNames("list-item", { "list-item-clickable": !!onClick })}
      onClick=${onClick}
    >
      ${iconSvg ? html`<div class="list-item-icon">${iconSvg}</div>` : null}
      <div class="list-item-body">
        <div class="list-item-title">${title}</div>
        ${subtitle
          ? html`<div class="list-item-subtitle">${subtitle}</div>`
          : null}
      </div>
      ${trailing != null
        ? html`<div class="list-item-trailing">${trailing}</div>`
        : null}
    </div>
  `;
}
