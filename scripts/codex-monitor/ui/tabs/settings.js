/* ─────────────────────────────────────────────────────────────
 *  Tab: Settings — account, appearance, notifications, data,
 *       executor defaults, advanced, about
 * ────────────────────────────────────────────────────────────── */
import { h } from "https://esm.sh/preact@10.25.4";
import {
  useState,
  useEffect,
  useCallback,
} from "https://esm.sh/preact@10.25.4/hooks";
import htm from "https://esm.sh/htm@3.1.1";

const html = htm.bind(h);

import { haptic, showConfirm, showAlert } from "../modules/telegram.js";
import { apiFetch } from "../modules/api.js";
import {
  connected,
  statusData,
  executorData,
  showToast,
} from "../modules/state.js";
import { Card, Badge, ListItem } from "../components/shared.js";
import { SegmentedControl, Collapsible, Toggle } from "../components/forms.js";

/* ─── CloudStorage helpers ─── */
function cloudGet(key) {
  return new Promise((resolve) => {
    const tg = globalThis.Telegram?.WebApp;
    if (tg?.CloudStorage) {
      tg.CloudStorage.getItem(key, (err, val) => {
        if (err || val == null) resolve(null);
        else {
          try {
            resolve(JSON.parse(val));
          } catch {
            resolve(val);
          }
        }
      });
    } else {
      try {
        const v = localStorage.getItem("ve_settings_" + key);
        resolve(v != null ? JSON.parse(v) : null);
      } catch {
        resolve(null);
      }
    }
  });
}

function cloudSet(key, value) {
  const tg = globalThis.Telegram?.WebApp;
  const str = JSON.stringify(value);
  if (tg?.CloudStorage) {
    tg.CloudStorage.setItem(key, str, () => {});
  } else {
    try {
      localStorage.setItem("ve_settings_" + key, str);
    } catch {
      /* noop */
    }
  }
}

function cloudRemove(key) {
  const tg = globalThis.Telegram?.WebApp;
  if (tg?.CloudStorage) {
    tg.CloudStorage.removeItem(key, () => {});
  } else {
    try {
      localStorage.removeItem("ve_settings_" + key);
    } catch {
      /* noop */
    }
  }
}

/* ─── Version info ─── */
const APP_VERSION = "1.0.0";
const APP_NAME = "VirtEngine Control Center";

/* ─── SettingsTab ─── */
export function SettingsTab() {
  /* Telegram user info */
  const tg = globalThis.Telegram?.WebApp;
  const user = tg?.initDataUnsafe?.user;

  /* Preferences (loaded from CloudStorage) */
  const [fontSize, setFontSize] = useState("medium");
  const [notifyUpdates, setNotifyUpdates] = useState(true);
  const [notifyErrors, setNotifyErrors] = useState(true);
  const [notifyComplete, setNotifyComplete] = useState(true);
  const [debugMode, setDebugMode] = useState(false);
  const [defaultMaxParallel, setDefaultMaxParallel] = useState(4);
  const [defaultSdk, setDefaultSdk] = useState("auto");
  const [defaultRegion, setDefaultRegion] = useState("auto");
  const [showRawJson, setShowRawJson] = useState(false);
  const [loaded, setLoaded] = useState(false);

  /* Load prefs from CloudStorage on mount */
  useEffect(() => {
    (async () => {
      const [fs, nu, ne, nc, dm, dmp, ds, dr] = await Promise.all([
        cloudGet("fontSize"),
        cloudGet("notifyUpdates"),
        cloudGet("notifyErrors"),
        cloudGet("notifyComplete"),
        cloudGet("debugMode"),
        cloudGet("defaultMaxParallel"),
        cloudGet("defaultSdk"),
        cloudGet("defaultRegion"),
      ]);
      if (fs) setFontSize(fs);
      if (nu != null) setNotifyUpdates(nu);
      if (ne != null) setNotifyErrors(ne);
      if (nc != null) setNotifyComplete(nc);
      if (dm != null) setDebugMode(dm);
      if (dmp != null) setDefaultMaxParallel(dmp);
      if (ds) setDefaultSdk(ds);
      if (dr) setDefaultRegion(dr);
      setLoaded(true);
    })();
  }, []);

  /* Persist helpers */
  const toggle = useCallback((key, getter, setter) => {
    const next = !getter;
    setter(next);
    cloudSet(key, next);
    haptic();
  }, []);

  const handleFontSize = (v) => {
    setFontSize(v);
    cloudSet("fontSize", v);
    haptic();
    /* Apply font size to root */
    const map = { small: "13px", medium: "15px", large: "17px" };
    document.documentElement.style.setProperty(
      "--base-font-size",
      map[v] || "15px",
    );
  };

  const handleDefaultMaxParallel = (v) => {
    const val = Math.max(1, Math.min(20, Number(v)));
    setDefaultMaxParallel(val);
    cloudSet("defaultMaxParallel", val);
    haptic();
  };

  const handleDefaultSdk = (v) => {
    setDefaultSdk(v);
    cloudSet("defaultSdk", v);
    haptic();
  };

  const handleDefaultRegion = (v) => {
    setDefaultRegion(v);
    cloudSet("defaultRegion", v);
    haptic();
  };

  /* Clear cache */
  const handleClearCache = async () => {
    const ok = await showConfirm("Clear all cached data and preferences?");
    if (!ok) return;
    haptic("medium");
    const keys = [
      "fontSize",
      "notifyUpdates",
      "notifyErrors",
      "notifyComplete",
      "debugMode",
      "defaultMaxParallel",
      "defaultSdk",
      "defaultRegion",
    ];
    for (const k of keys) cloudRemove(k);
    showToast("Cache cleared — reload to apply", "success");
  };

  /* Reset all settings */
  const handleReset = async () => {
    const ok = await showConfirm("Reset ALL settings to defaults?");
    if (!ok) return;
    haptic("heavy");
    const keys = [
      "fontSize",
      "notifyUpdates",
      "notifyErrors",
      "notifyComplete",
      "debugMode",
      "defaultMaxParallel",
      "defaultSdk",
      "defaultRegion",
    ];
    for (const k of keys) cloudRemove(k);
    setFontSize("medium");
    setNotifyUpdates(true);
    setNotifyErrors(true);
    setNotifyComplete(true);
    setDebugMode(false);
    setDefaultMaxParallel(4);
    setDefaultSdk("auto");
    setDefaultRegion("auto");
    document.documentElement.style.removeProperty("--base-font-size");
    showToast("Settings reset", "success");
  };

  /* Raw status JSON */
  const rawJson =
    debugMode && showRawJson
      ? JSON.stringify(
          { status: statusData?.value, executor: executorData?.value },
          null,
          2,
        )
      : "";

  return html`
    <!-- ─── Account ─── -->
    <${Collapsible} title="👤 Account" defaultOpen=${true}>
      <${Card}>
        <div class="settings-row">
          ${user?.photo_url &&
          html`
            <img
              src=${user.photo_url}
              alt="avatar"
              class="settings-avatar"
              style="width:48px;height:48px;border-radius:50%;margin-right:12px"
            />
          `}
          <div>
            <div style="font-weight:600;font-size:15px">
              ${user?.first_name || "Unknown"} ${user?.last_name || ""}
            </div>
            ${user?.username &&
            html`<div class="meta-text">@${user.username}</div>`}
          </div>
        </div>
        <div class="meta-text mt-sm">App version: ${APP_VERSION}</div>
      <//>
    <//>

    <!-- ─── Appearance ─── -->
    <${Collapsible} title="🎨 Appearance" defaultOpen=${false}>
      <${Card}>
        <div class="card-subtitle mb-sm">Color Scheme</div>
        <div class="meta-text mb-md">
          Follows your Telegram theme automatically.
          ${tg?.colorScheme
            ? html` Current: <strong>${tg.colorScheme}</strong>`
            : ""}
        </div>
        <div class="card-subtitle mb-sm">Font Size</div>
        <${SegmentedControl}
          options=${[
            { value: "small", label: "Small" },
            { value: "medium", label: "Medium" },
            { value: "large", label: "Large" },
          ]}
          value=${fontSize}
          onChange=${handleFontSize}
        />
      <//>
    <//>

    <!-- ─── Notifications ─── -->
    <${Collapsible} title="🔔 Notifications" defaultOpen=${false}>
      <${Card}>
        <${ListItem}
          title="Real-time Updates"
          subtitle="Show live data refresh indicators"
          trailing=${html`
            <${Toggle}
              checked=${notifyUpdates}
              onChange=${() =>
                toggle("notifyUpdates", notifyUpdates, setNotifyUpdates)}
            />
          `}
        />
        <${ListItem}
          title="Error Alerts"
          subtitle="Toast notifications for errors"
          trailing=${html`
            <${Toggle}
              checked=${notifyErrors}
              onChange=${() =>
                toggle("notifyErrors", notifyErrors, setNotifyErrors)}
            />
          `}
        />
        <${ListItem}
          title="Task Completion"
          subtitle="Notify when tasks finish"
          trailing=${html`
            <${Toggle}
              checked=${notifyComplete}
              onChange=${() =>
                toggle("notifyComplete", notifyComplete, setNotifyComplete)}
            />
          `}
        />
      <//>
    <//>

    <!-- ─── Data & Storage ─── -->
    <${Collapsible} title="💾 Data & Storage" defaultOpen=${false}>
      <${Card}>
        <${ListItem}
          title="WebSocket"
          subtitle="Live connection status"
          trailing=${html`
            <${Badge}
              status=${connected?.value ? "done" : "error"}
              text=${connected?.value ? "Connected" : "Offline"}
            />
          `}
        />
        <${ListItem}
          title="API Endpoint"
          subtitle=${globalThis.location?.origin || "unknown"}
        />
        <${ListItem}
          title="Clear Cache"
          subtitle="Remove all stored preferences"
          trailing=${html`
            <button class="btn btn-ghost btn-sm" onClick=${handleClearCache}>
              🗑 Clear
            </button>
          `}
        />
      <//>
    <//>

    <!-- ─── Executor Defaults ─── -->
    <${Collapsible} title="⚙️ Executor Defaults" defaultOpen=${false}>
      <${Card}>
        <div class="card-subtitle mb-sm">Default Max Parallel</div>
        <div class="range-row mb-md">
          <input
            type="range"
            min="1"
            max="20"
            step="1"
            value=${defaultMaxParallel}
            onInput=${(e) => setDefaultMaxParallel(Number(e.target.value))}
            onChange=${(e) => handleDefaultMaxParallel(Number(e.target.value))}
          />
          <span class="pill">${defaultMaxParallel}</span>
        </div>

        <div class="card-subtitle mb-sm">Default SDK</div>
        <${SegmentedControl}
          options=${[
            { value: "codex", label: "Codex" },
            { value: "copilot", label: "Copilot" },
            { value: "claude", label: "Claude" },
            { value: "auto", label: "Auto" },
          ]}
          value=${defaultSdk}
          onChange=${handleDefaultSdk}
        />

        <div class="card-subtitle mt-md mb-sm">Default Region</div>
        <${SegmentedControl}
          options=${[
            { value: "us", label: "US" },
            { value: "sweden", label: "Sweden" },
            { value: "auto", label: "Auto" },
          ]}
          value=${defaultRegion}
          onChange=${handleDefaultRegion}
        />
      <//>
    <//>

    <!-- ─── Advanced ─── -->
    <${Collapsible} title="🔧 Advanced" defaultOpen=${false}>
      <${Card}>
        <${ListItem}
          title="Debug Mode"
          subtitle="Show raw data and extra diagnostics"
          trailing=${html`
            <${Toggle}
              checked=${debugMode}
              onChange=${() => toggle("debugMode", debugMode, setDebugMode)}
            />
          `}
        />

        ${debugMode &&
        html`
          <${ListItem}
            title="Raw Status JSON"
            subtitle="View raw API response data"
            trailing=${html`
              <button
                class="btn btn-ghost btn-sm"
                onClick=${() => {
                  setShowRawJson(!showRawJson);
                  haptic();
                }}
              >
                ${showRawJson ? "Hide" : "Show"}
              </button>
            `}
          />
          ${showRawJson &&
          html`
            <div class="log-box mt-sm" style="max-height:300px;font-size:11px">
              ${rawJson}
            </div>
          `}
        `}

        <${ListItem}
          title="Reset All Settings"
          subtitle="Restore defaults"
          trailing=${html`
            <button class="btn btn-danger btn-sm" onClick=${handleReset}>
              Reset
            </button>
          `}
        />
      <//>
    <//>

    <!-- ─── About ─── -->
    <${Collapsible} title="ℹ️ About" defaultOpen=${false}>
      <${Card}>
        <div style="text-align:center;padding:12px 0">
          <div style="font-size:18px;font-weight:700;margin-bottom:4px">
            ${APP_NAME}
          </div>
          <div class="meta-text">Version ${APP_VERSION}</div>
          <div class="meta-text mt-sm">
            Telegram Mini App for VirtEngine orchestrator management
          </div>
          <div class="meta-text mt-sm">
            Built with Preact + HTM · No build step
          </div>
          <div class="btn-row mt-md" style="justify-content:center">
            <button
              class="btn btn-ghost btn-sm"
              onClick=${() => {
                haptic();
                const tg = globalThis.Telegram?.WebApp;
                if (tg?.openLink)
                  tg.openLink("https://github.com/virtengine/virtengine");
                else
                  globalThis.open(
                    "https://github.com/virtengine/virtengine",
                    "_blank",
                  );
              }}
            >
              GitHub
            </button>
            <button
              class="btn btn-ghost btn-sm"
              onClick=${() => {
                haptic();
                const tg = globalThis.Telegram?.WebApp;
                if (tg?.openLink) tg.openLink("https://docs.virtengine.com");
                else globalThis.open("https://docs.virtengine.com", "_blank");
              }}
            >
              Docs
            </button>
          </div>
        </div>
      <//>
    <//>
  `;
}
