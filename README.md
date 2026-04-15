# SysMon

A lightweight, cross-platform system resource monitor built with **Go + Wails v3** (backend) and **Svelte** (frontend). Ships as a single portable binary — no runtime or installer needed.

Displays live CPU usage (per-core), memory, network throughput with sparklines, and disk usage per partition. Updates every second.

---

## Table of Contents

1. [What You Need](#1-what-you-need)
2. [Getting the Code Running](#2-getting-the-code-running)
3. [Project Structure](#3-project-structure)
4. [How the App Works](#4-how-the-app-works)
5. [Making Changes](#5-making-changes)
6. [Building a Release Binary](#6-building-a-release-binary)
7. [Key Dependencies](#7-key-dependencies)
8. [Common Issues](#8-common-issues)

---

## 1. What You Need

Install these before anything else.

| Tool | Minimum version | Why |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.25 | Backend language |
| [Wails v3 CLI](https://v3alpha.wails.io/) | alpha.74 | Dev server, build tooling |
| [Bun](https://bun.sh/) | 1.0 | Frontend package manager |
| [Node.js](https://nodejs.org/) | 18 | Required by Wails internals even if you use Bun |
| A C compiler | any | Wails needs CGo on Windows (comes with MSYS2 / TDM-GCC) |

**Install Wails v3 CLI** (run once after Go is installed):

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

Verify everything works:

```bash
go version        # go1.25+
wails3 version    # v3.0.0-alpha.74
bun --version     # 1.x
```

---

## 2. Getting the Code Running

```bash
# 1. Clone the repo
git clone <repo-url>
cd sysmon

# 2. Install Go dependencies
go mod download

# 3. Install frontend dependencies
cd frontend && bun install && cd ..

# 4. Generate Wails bindings
#    (creates frontend/bindings/ — needed before building or running)
wails3 generate bindings

# 5. Start the dev server with hot reload
wails3 dev
```

The app window opens automatically. Any change to a `.go` file restarts the backend; any change to a `.svelte` or `.js` file hot-reloads the frontend without restarting.

> **Note:** `wails3 dev` uses `npm` internally for its task runner even if you installed deps with Bun. Both lockfiles (`bun.lock` and `package-lock.json`) are kept for this reason.

---

## 3. Project Structure

```
sysmon/
│
├── main.go          # Entry point. Creates the Wails app, window, and starts polling.
├── types.go         # All shared data types (CPUStats, MemStats, etc.) with doc comments.
├── service.go       # SysInfoService — collects metrics and exposes them to the frontend.
│
├── go.mod / go.sum  # Go module definition and checksums.
│
└── frontend/
    ├── src/
    │   ├── App.svelte              # Root shell. Subscribes to events, passes data to cards.
    │   ├── main.js                 # Svelte entry point (mounts App).
    │   └── lib/
    │       ├── utils.js            # Pure helpers: formatBytes, formatUptime, usageColor, etc.
    │       ├── GaugeRing.svelte    # Reusable animated SVG ring gauge.
    │       ├── Sparkline.svelte    # Reusable SVG area chart.
    │       ├── CpuCard.svelte      # CPU panel (gauge + per-core bars).
    │       ├── MemoryCard.svelte   # RAM + swap panel.
    │       ├── NetworkCard.svelte  # Upload/download rates + sparklines.
    │       └── DiskCard.svelte     # Single partition card (used in a list).
    │
    ├── public/
    │   └── style.css     # Global CSS variables (colours, typography, scrollbars).
    │
    ├── index.html        # HTML shell loaded by the Wails webview.
    ├── package.json      # Frontend dependencies.
    └── vite.config.js    # Vite build configuration.
```

Generated directories (not committed, safe to delete and regenerate):

```
frontend/dist/        # Built frontend. Created by `bun run build`.
frontend/bindings/    # Type-safe JS bindings. Created by `wails3 generate bindings`.
frontend/node_modules/
bin/                  # Compiled binary. Created by `go build`.
```

---

## 4. How the App Works

Understanding the data flow end-to-end is the fastest way to get productive.

### Backend → Frontend: event-based push

The Go backend does **not** wait for the frontend to ask for data. Instead, it pushes a snapshot every second:

```
main.go
  └─ go svc.StartPolling(app)        ← runs in a goroutine
       └─ every 1s: app.Event.Emit("stats", svc.GetStats())
```

`GetStats()` assembles a `SystemStats` struct (defined in `types.go`) by calling five collector methods in `service.go`:

```
GetStats()
  ├─ collectCPU()      uses gopsutil/cpu
  ├─ collectMemory()   uses gopsutil/mem
  ├─ collectDisks()    uses gopsutil/disk
  ├─ collectNetwork()  uses gopsutil/net  (rate = delta / elapsed time)
  └─ collectOS()       uses gopsutil/host
```

### Frontend: reactive state

`App.svelte` subscribes to the `"stats"` event once on mount:

```js
Events.On('stats', (event) => {
  stats = event.data;          // triggers Svelte reactivity
  sendHistory = [...sendHistory.slice(-59), stats.network.sendRate];
  recvHistory = [...recvHistory.slice(-59), stats.network.recvRate];
});
```

When `stats` changes, Svelte re-renders only the parts of the DOM that depend on it. The four card components (`CpuCard`, `MemoryCard`, `NetworkCard`, `DiskCard`) are purely presentational — they receive props and render, with no internal state.

### Static vs dynamic data

Querying CPU model name or OS info over WMI/sysctl can take hundreds of milliseconds. To avoid slowing down every poll tick, `service.go` separates the work:

- **Cached at startup** (`setup()`): CPU model, core count, clock speed, OS name, hostname.
- **Read every second**: CPU %, memory usage, disk usage, network I/O counters.

---

## 5. Making Changes

### Add a new metric to the backend

1. **Add a field** to the appropriate struct in `types.go` with a `json:"..."` tag and a doc comment.
2. **Populate it** in the matching `collect*()` function in `service.go`.
3. **Re-run bindings** so the frontend type definitions stay in sync:
   ```bash
   wails3 generate bindings
   ```

### Add a new frontend panel

1. Create `frontend/src/lib/YourCard.svelte`.
2. Accept your data as a typed `export let` prop.
3. Import and use helpers from `utils.js` for formatting.
4. Import and render it in `App.svelte`.

No build step needed during dev — Vite picks up the new file instantly.

### Change the colour theme

All colour tokens live in `frontend/public/style.css` under `:root`. Change a variable there and every component that uses it updates automatically:

```css
:root {
  --cyan:   #00d4ff;   /* primary accent */
  --purple: #a855f7;   /* secondary accent */
  --green:  #10b981;   /* healthy usage */
  --orange: #f59e0b;   /* elevated usage */
  --red:    #ef4444;   /* high / warning */
}
```

Usage thresholds (when green → orange → red) are in `utils.js` in `usageColor()` and `diskColor()`.

### Change the poll rate

Edit the constant at the top of `service.go`:

```go
const PollInterval = time.Second  // change to e.g. 2*time.Second
```

Also update the history window in `App.svelte` if you change the interval:

```js
// 60 samples at 1s = 60 seconds of history.
// At 2s interval, 30 samples covers the same 60 seconds.
let sendHistory = new Array(60).fill(0);
```

---

## 6. Building a Release Binary

```bash
# Build the frontend first (output goes to frontend/dist/)
cd frontend && bun run build && cd ..

# Compile the Go binary (embeds frontend/dist/ inside the .exe)
go build -o bin/sysmon.exe .
```

The result is a single self-contained executable. Copy `bin/sysmon.exe` to any Windows machine — no installer, no runtime, no DLLs required.

For other platforms, cross-compile with:

```bash
GOOS=linux   GOARCH=amd64 go build -o bin/sysmon-linux .
GOOS=darwin  GOARCH=arm64 go build -o bin/sysmon-mac .
```

> **Wails build command:** `wails3 build` handles the above steps automatically and also embeds platform icons and version info. Use it when preparing an official release.

---

## 7. Key Dependencies

### Go

| Package | Purpose |
|---|---|
| `github.com/wailsapp/wails/v3` | Desktop window, webview, event bus, RPC bridge |
| `github.com/shirou/gopsutil/v4` | CPU, memory, disk, network stats — cross-platform |

### Frontend

| Package | Purpose |
|---|---|
| `svelte` | Reactive UI framework (v4) |
| `vite` | Dev server and production bundler |
| `@wailsio/runtime` | Wails JS runtime — `Events.On`, RPC calls |

---

## 8. Common Issues

**`wails3: command not found`**
Your Go `bin` directory is not on `PATH`. Add `$(go env GOPATH)/bin` to your shell profile.

**`frontend/dist not found` during `go build`**
The frontend must be built before the Go binary because `main.go` embeds `frontend/dist/`. Run `cd frontend && bun run build` first.

**`event bindings module not found` during `bun run build`**
Wails bindings have not been generated yet. Run `wails3 generate bindings` from the project root.

**First CPU reading is 0%**
Expected. `gopsutil` needs two readings to compute a percentage. The first sample after startup is always 0; the second (after ~1 second) is accurate.

**Blank window on launch**
Usually means the frontend was built with a different set of asset hashes than what the binary embedded. Do a clean rebuild:
```bash
rm -rf frontend/dist
cd frontend && bun run build && cd ..
go build -o bin/sysmon.exe .
```
