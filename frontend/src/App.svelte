<!--
  App.svelte — Root layout shell.

  Responsibilities (and nothing more):
    1. Subscribe to the "stats" Wails event emitted by the Go backend.
    2. Maintain a 60-second rolling window of network rates for sparklines.
    3. Pass the received data down to the four panel components.

  All business logic, formatting, and component-level styles live inside
  the individual card components and src/lib/utils.js — not here.
-->
<script>
  import { onMount, onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';

  import { formatUptime } from './lib/utils.js';
  import CpuCard     from './lib/CpuCard.svelte';
  import MemoryCard  from './lib/MemoryCard.svelte';
  import NetworkCard from './lib/NetworkCard.svelte';
  import DiskCard    from './lib/DiskCard.svelte';

  // ─── State ─────────────────────────────────────────────────────────────────

  /**
   * The most recent system snapshot from the Go backend.
   * Null until the first "stats" event arrives (usually within ~1 second).
   * @type {object | null}
   */
  let stats = null;

  /**
   * Rolling 60-second history of upload / download rates (bytes/sec).
   * New values are appended and the oldest trimmed on every stats event.
   * Pre-filled with zeros so sparklines render immediately with a flat baseline.
   */
  let sendHistory = /** @type {number[]} */ (new Array(60).fill(0));
  let recvHistory = /** @type {number[]} */ (new Array(60).fill(0));

  // Wails v3 Events.On returns a cancel function — store it for cleanup.
  let cancelListener = /** @type {(() => void) | undefined} */ (undefined);

  // ─── Lifecycle ─────────────────────────────────────────────────────────────

  onMount(() => {
    cancelListener = Events.On('stats', (event) => {
      stats = event.data;

      // Append the latest rates and discard values older than 60 seconds.
      sendHistory = [...sendHistory.slice(-59), stats.network.sendRate ?? 0];
      recvHistory = [...recvHistory.slice(-59), stats.network.recvRate ?? 0];
    });
  });

  onDestroy(() => {
    // Clean up the event listener to prevent memory leaks if the component
    // is ever unmounted (e.g. during hot-module replacement in dev mode).
    cancelListener?.();
  });
</script>

<!-- ─── App shell ─────────────────────────────────────────────────────────── -->

<div class="app">

  <!-- ── Sticky header ─────────────────────────────────────────────── -->
  <header class="header">
    <div class="brand">
      <!--
        The live dot pulses to give a visual heartbeat signal.
        It uses a CSS animation defined in style.css rather than JS.
      -->
      <span class="live-dot" aria-hidden="true"></span>
      <span class="brand-name">SysMon</span>
    </div>

    {#if stats}
      <div class="meta" aria-label="System info">
        <span>{stats.os.hostname}</span>
        <span class="sep" aria-hidden="true">·</span>
        <span>{stats.os.platform || stats.os.os}</span>
        <span class="sep" aria-hidden="true">·</span>
        <span>up {formatUptime(stats.os.uptime)}</span>
      </div>
    {:else}
      <span class="connecting">Connecting…</span>
    {/if}
  </header>

  <!-- ── Content ───────────────────────────────────────────────────── -->

  {#if !stats}
    <!--
      Loading state: the Go backend takes up to ~1 second to emit the first
      event (CPU sampler warm-up + initial disk scan). Show a spinner.
    -->
    <div class="loading" role="status" aria-live="polite">
      <div class="spinner" aria-hidden="true"></div>
      <p>Reading system stats…</p>
    </div>

  {:else}
    <main class="main">

      <!-- Top row: three equal panels side by side -->
      <div class="top-grid">
        <CpuCard     cpu={stats.cpu} />
        <MemoryCard  memory={stats.memory} />
        <NetworkCard
          network={stats.network}
          {sendHistory}
          {recvHistory}
        />
      </div>

      <!-- Storage row: one DiskCard per partition, scrolls horizontally -->
      {#if stats.disks?.length}
        <section class="storage-section" aria-label="Storage">
          <h2 class="section-heading">Storage</h2>
          <div class="disk-row" role="list">
            <!--
              (disk.mountpoint) is a Svelte keyed-each hint.
              It ensures Svelte reuses existing DOM nodes when the list order
              changes, avoiding unnecessary layout flashes.
            -->
            {#each stats.disks as disk (disk.mountpoint)}
              <DiskCard {disk} />
            {/each}
          </div>
        </section>
      {/if}

    </main>
  {/if}

</div>

<style>
  /* ── App root ────────────────────────────────────────────────────── */
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  /* ── Header ──────────────────────────────────────────────────────── */
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 24px;
    border-bottom: 1px solid var(--border);
    background: rgba(10, 10, 15, 0.8);
    backdrop-filter: blur(12px);
    /* Stays pinned at top while the content scrolls. */
    position: sticky;
    top: 0;
    z-index: 10;
  }

  /* ── Brand ───────────────────────────────────────────────────────── */
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .live-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--cyan);
    box-shadow: 0 0 8px var(--cyan);
    animation: live-pulse 2s ease-in-out infinite;
  }
  @keyframes live-pulse {
    0%, 100% { opacity: 1; }
    50%       { opacity: 0.3; }
  }
  .brand-name {
    font-size: 1rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--cyan);
  }

  /* ── Meta / connecting ───────────────────────────────────────────── */
  .meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.78rem;
    color: var(--text-dim);
  }
  .sep        { color: var(--border-bright); }
  .connecting { font-size: 0.78rem; color: var(--text-dim); }

  /* ── Loading spinner ─────────────────────────────────────────────── */
  .loading {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 20px;
    color: var(--text-dim);
  }
  .spinner {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 2px solid rgba(0, 212, 255, 0.2);
    border-top-color: var(--cyan);
    animation: spin 1s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* ── Main content area ───────────────────────────────────────────── */
  .main {
    flex: 1;
    padding: 20px 24px 28px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* ── Top grid ────────────────────────────────────────────────────── */
  .top-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
  }

  /* ── Storage section ─────────────────────────────────────────────── */
  .storage-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .section-heading {
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: var(--text-dim);
    margin: 0;
    padding-left: 2px;
  }
  .disk-row {
    display: flex;
    gap: 12px;
    overflow-x: auto;
    padding-bottom: 6px; /* Space for the scrollbar. */
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
  }
</style>
