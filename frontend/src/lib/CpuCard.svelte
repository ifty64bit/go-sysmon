<!--
  CpuCard.svelte — CPU metrics panel.

  Displays:
    - An animated ring gauge showing overall usage %.
    - Model name, physical/logical core count, and base clock speed.
    - A row of per-core vertical bars, colour-coded by load level.
-->
<script>
  import GaugeRing from './GaugeRing.svelte';
  import { usageColor } from './utils.js';

  /**
   * CPU snapshot from the Go backend.
   * @type {{
   *   usage:        number,
   *   perCore:      number[],
   *   cores:        number,
   *   logicalCores: number,
   *   freqMHz:      number,
   *   modelName:    string
   * }}
   */
  export let cpu;

  // Derive a display string for clock speed (convert MHz → GHz).
  $: freqLabel = cpu.freqMHz > 0
    ? `${(cpu.freqMHz / 1000).toFixed(2)} GHz`
    : null;
</script>

<article class="card" aria-label="CPU">
  <h2 class="card-title">CPU</h2>

  <!-- ── Top section: gauge + static info ─────────────────────────── -->
  <div class="top-row">
    <GaugeRing
      value={cpu.usage}
      size={140}
      thickness={10}
      color={usageColor(cpu.usage)}
    >
      <span class="gauge-value" style="color: {usageColor(cpu.usage)}">
        {cpu.usage.toFixed(1)}%
      </span>
      <span class="gauge-label">usage</span>
    </GaugeRing>

    <div class="info">
      <p class="model-name" title={cpu.modelName}>
        {cpu.modelName || 'Unknown CPU'}
      </p>
      <div class="badges">
        <span class="badge">{cpu.cores}C / {cpu.logicalCores}T</span>
        {#if freqLabel}
          <span class="badge">{freqLabel}</span>
        {/if}
      </div>
    </div>
  </div>

  <!-- ── Per-core bars ─────────────────────────────────────────────── -->
  <!--
    Each bar is a small vertical column that fills from the bottom.
    The height is driven by a CSS transition so changes animate smoothly.
  -->
  {#if cpu.perCore?.length}
    <div class="cores" title="Per-core utilisation" role="list" aria-label="Per-core CPU usage">
      {#each cpu.perCore as pct, index}
        <div class="core-col" role="listitem" aria-label="Core {index}: {pct.toFixed(0)}%">
          <div class="core-track">
            <div
              class="core-fill"
              style="height: {pct.toFixed(0)}%; background: {usageColor(pct)};"
            ></div>
          </div>
          <span class="core-index" aria-hidden="true">{index}</span>
        </div>
      {/each}
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
    gap: 16px;
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

  /* ── Top row ───────────────────────────────────────────────────────── */
  .top-row {
    display: flex;
    align-items: center;
    gap: 18px;
  }
  .info {
    flex: 1;
    min-width: 0; /* Allow text truncation inside a flex container. */
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .model-name {
    font-size: 0.78rem;
    color: var(--text);
    line-height: 1.4;
    margin: 0;
    /* Truncate long CPU brand strings with an ellipsis. */
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .badges {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .badge {
    font-size: 0.68rem;
    padding: 2px 8px;
    border-radius: 99px;
    background: rgba(255, 255, 255, 0.06);
    color: var(--text-dim);
    border: 1px solid var(--border);
  }

  /* ── Gauge centre text ─────────────────────────────────────────────── */
  .gauge-value {
    font-size: 1.15rem;
    font-weight: 700;
    line-height: 1;
    transition: color 0.4s ease;
  }
  .gauge-label {
    font-size: 0.6rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-top: 2px;
  }

  /* ── Per-core bars ─────────────────────────────────────────────────── */
  .cores {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .core-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
  }
  .core-track {
    width: 14px;
    height: 36px;
    background: rgba(255, 255, 255, 0.06);
    border-radius: 3px;
    position: relative;
    overflow: hidden;
  }
  .core-fill {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    min-height: 2px; /* Always show a sliver so "0%" cores are visible. */
    border-radius: 3px;
    transition: height 0.7s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .core-index {
    font-size: 0.5rem;
    color: var(--text-muted);
  }
</style>
