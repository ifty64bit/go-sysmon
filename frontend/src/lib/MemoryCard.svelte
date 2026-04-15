<!--
  MemoryCard.svelte — RAM and swap memory panel.

  Displays:
    - A ring gauge showing overall RAM usage %.
    - A detail table with used / free / total byte values.
    - An animated progress bar for RAM.
    - An animated progress bar for swap (hidden when no swap is configured).
-->
<script>
  import GaugeRing from './GaugeRing.svelte';
  import { formatBytes, usageColor } from './utils.js';

  /**
   * Memory snapshot from the Go backend.
   * @type {{
   *   total:     number,
   *   used:      number,
   *   available: number,
   *   usedPct:   number,
   *   swapTotal: number,
   *   swapUsed:  number
   * }}
   */
  export let memory;

  // Swap percentage is derived here rather than in Go to keep the backend simple.
  $: swapPct = memory.swapTotal > 0
    ? (memory.swapUsed / memory.swapTotal) * 100
    : 0;
</script>

<article class="card" aria-label="Memory">
  <h2 class="card-title">Memory</h2>

  <!-- ── Ring gauge + detail table side by side ───────────────────── -->
  <div class="gauge-row">
    <GaugeRing
      value={memory.usedPct}
      size={120}
      thickness={9}
      color={usageColor(memory.usedPct)}
    >
      <span class="gauge-value" style="color: {usageColor(memory.usedPct)}">
        {memory.usedPct.toFixed(1)}%
      </span>
    </GaugeRing>

    <dl class="detail-table">
      <div class="detail-row">
        <dt>Used</dt>
        <dd>{formatBytes(memory.used, 2)}</dd>
      </div>
      <div class="detail-row">
        <dt>Free</dt>
        <dd>{formatBytes(memory.available, 2)}</dd>
      </div>
      <div class="detail-row">
        <dt>Total</dt>
        <dd>{formatBytes(memory.total, 1)}</dd>
      </div>
    </dl>
  </div>

  <!-- ── RAM progress bar ────────────────────────────────────────── -->
  <div class="bar-group">
    <div class="bar-header">
      <span class="bar-name">RAM</span>
      <span class="bar-numbers">
        {formatBytes(memory.used)} / {formatBytes(memory.total)}
      </span>
    </div>
    <div class="bar-track" role="progressbar" aria-valuenow={memory.usedPct} aria-valuemin="0" aria-valuemax="100">
      <div
        class="bar-fill"
        style="width: {memory.usedPct}%; background: {usageColor(memory.usedPct)};"
      ></div>
    </div>
  </div>

  <!-- ── Swap progress bar (only rendered when swap is configured) ── -->
  {#if memory.swapTotal > 0}
    <div class="bar-group">
      <div class="bar-header">
        <span class="bar-name">Swap</span>
        <span class="bar-numbers">
          {formatBytes(memory.swapUsed)} / {formatBytes(memory.swapTotal)}
        </span>
      </div>
      <div class="bar-track swap" role="progressbar" aria-valuenow={swapPct} aria-valuemin="0" aria-valuemax="100">
        <div
          class="bar-fill"
          style="width: {swapPct}%; background: var(--purple);"
        ></div>
      </div>
    </div>
  {/if}
</article>

<style>
  /* ── Card shell ────────────────────────────────────────────────────── */
  .card {
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    transition: border-color 0.2s ease;
  }
  .card:hover {
    border-color: var(--border-bright);
  }
  .card-title {
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: var(--text-dim);
    margin: 0;
  }

  /* ── Gauge + detail row ────────────────────────────────────────────── */
  .gauge-row {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .gauge-value {
    font-size: 1rem;
    font-weight: 700;
    line-height: 1;
    transition: color 0.4s ease;
  }

  /* ── Detail table ──────────────────────────────────────────────────── */
  .detail-table {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 0;
  }
  .detail-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.78rem;
  }
  dt { color: var(--text-dim); }
  dd { margin: 0; color: var(--text); }

  /* ── Progress bars ─────────────────────────────────────────────────── */
  .bar-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .bar-header {
    display: flex;
    justify-content: space-between;
    font-size: 0.72rem;
  }
  .bar-name    { color: var(--text-dim); }
  .bar-numbers { color: var(--text); }

  .bar-track {
    height: 6px;
    background: rgba(255, 255, 255, 0.07);
    border-radius: 99px;
    overflow: hidden;
  }
  .bar-track.swap { height: 5px; }

  .bar-fill {
    height: 100%;
    border-radius: 99px;
    transition: width 0.7s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 0 6px currentColor;
  }
</style>
