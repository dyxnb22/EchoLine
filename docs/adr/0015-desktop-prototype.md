# ADR 0015: Desktop Client Prototype Strategy (Electron vs Tauri)

## Status

Accepted (design; desktop implementation deferred to extension phase)

## Context

Telegram and Discord both offer desktop clients that are integral to their power-user experience. EchoLine's web app works in a browser, but a dedicated desktop app offers:

1. **System tray + background process**: Notifications when the browser is closed.
2. **Global keyboard shortcuts**: Platform-native hotkeys for jump-to-conversation.
3. **OS-level notifications**: Richer notification payloads (image thumbnails, quick-reply actions on macOS/Windows).
4. **Native file system access**: Drag-and-drop file sending directly from the desktop.
5. **Auto-launch on login**: Users expect a messaging app to always be running.

The question is: **Electron vs Tauri**, both of which can wrap the existing React web frontend.

## Decision

Use **Tauri** for the desktop prototype.

### Comparison

| Factor | Electron | Tauri |
|--------|----------|-------|
| Bundle size | ~100–150 MB (ships Chromium) | ~3–10 MB (uses system WebView) |
| Memory usage | ~100–300 MB per instance | ~20–50 MB (no bundled Chromium) |
| Performance | Good (V8 is fast) | Good (system WebView) |
| Frontend reuse | Full (same React build) | Full (same React build) |
| Native backend | Node.js | Rust (safer, no GC pauses) |
| Distribution | Electron builder (mature) | Tauri bundler (good, improving) |
| Security | Larger attack surface (Node.js in main process) | Smaller attack surface (Rust, IPC allowlist) |
| macOS notarization | Well-documented | Supported, requires Apple certificate |

Tauri's smaller bundle size and memory footprint are meaningful for a messaging app that runs persistently in the background. The Rust backend is more appropriate for system-level operations (tray icon, OS notifications, file access) than Node.js.

### Architecture

```
┌─────────────────────────────────────────────────────┐
│  Tauri Desktop App                                   │
│  ┌─────────────────────────────────────────────┐    │
│  │  WebView (system: WKWebView / WebView2)     │    │
│  │  React frontend (same build as web)         │    │
│  │  WS connection to EchoLine API              │    │
│  └─────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────┐    │
│  │  Tauri Rust Core                            │    │
│  │  - OS notifications (tauri-plugin-notify)   │    │
│  │  - System tray (tauri::SystemTray)          │    │
│  │  - Auto-launch (tauri-plugin-autostart)     │    │
│  │  - File system access (tauri::fs)           │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

### Frontend Integration Points

The Tauri IPC bridge allows the React frontend to invoke Rust commands:

```typescript
// From React (web layer)
import { invoke } from '@tauri-apps/api/tauri';

// Show OS notification
await invoke('show_notification', { title: 'Alice', body: 'Hello!' });

// Open file picker
const file = await invoke('pick_file');
```

The frontend build is identical to the web build, with a feature flag `import.meta.env.TAURI_ENV` to enable Tauri-specific IPC calls.

### Push Notifications (Desktop)

Desktop push uses the same WS connection as the web client — no separate push system needed. When the app is in the background (tray), the WS connection is maintained and incoming messages trigger OS notifications via the Rust `tauri-plugin-notify`.

## Implementation Files

- `desktop/` _(planned)_ — Tauri project (`src-tauri/`, shares `frontend/` build output)
- `desktop/src-tauri/src/main.rs` _(planned)_ — Tauri entry, tray, notification commands
- `frontend/src/lib/tauri.ts` _(planned)_ — Tauri IPC wrapper (no-ops in browser mode)
- `frontend/src/hooks/useNotifications.ts` _(planned)_ — platform-aware notification hook

## Consequences

**Positive:**
- Minimal bundle size; no shipping Chromium.
- Rust backend provides memory-safe system integration.
- Frontend code is exactly the same as the web app.

**Negative:**
- System WebView inconsistencies: WKWebView (macOS) and WebView2 (Windows) differ slightly; CSS/JS quirks possible.
- Tauri ecosystem is younger than Electron; some plugins are less mature.
- Requires Rust build toolchain in CI.

## Interview Talking Points

- **Why Tauri over Electron?** "Telegram Desktop ships ~100 MB with bundled Chromium. Tauri ships 3–10 MB using the OS WebView, which users already have. For a messaging app running in the background all day, that memory footprint matters."
- **WebView inconsistency risk**: "We mitigate this with a shared React build and a CI step that tests the Tauri app on macOS and Windows. Most CSS quirks appear in WebView2 (Windows)."
- **Auto-launch behavior**: "The Rust `autostart` plugin adds the app to the OS startup list. This is a user opt-in setting. We default it to false and prompt after the third session."
