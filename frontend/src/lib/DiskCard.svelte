<!--
  DiskCard.svelte — A single disk partition card.

  Used inside the scrollable storage row in App.svelte. One card per
  partition returned by the Go backend.

  Displays:
    - Mountpoint label (short, e.g. "C:" or "/home").
    - Usage percentage (colour-coded).
    - Animated fill bar.
    - Used / free byte counts and filesystem type.
-->
<script>
  import { formatBytes, diskColor, diskLabel } from './utils.js';

  /**
   * Disk snapshot from the Go backend.
   * @type {{
   *   mountpoint: string,
   *   device:     string,
   *   fstype:     string,
   *   total:      number,
   *   used:       number,
   *   free:       number,
   *   usedPct:    number
   * }}
   */
  export let disk;

  // Reactive helpers so the template stays clean.
  $: color = diskColor(disk.usedPct);
  $: label = diskLabel(disk);
</script>

<article
  class="disk-card"
  role="listitem"
  aria-label="{label} — {disk.usedPct.toFixed(1)}% full"
>
  <!-- ── Header: label + percentage ───────────────────────────────── -->
  <div class="header">
    <span class="mount-label">{label}</span>
    <span class="pct-label" style="color: {color}">
      {disk.usedPct.toFixed(1)}%
    </span>
  </div>

  <!-- ── Usage bar ─────────────────────────────────────────────────── -->
  <div
    class="bar-track"
    role="progressbar"
    aria-valuenow={disk.usedPct}
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <div
      class="bar-fill"
      style="width: {disk.usedPct}%; background: {color};"
    ></div>
  </div>

  <!-- ── Byte details ───────────────────────────────────────────────── -->
  <div class="byte-row">
    <span class="dim">{formatBytes(disk.used)} used</span>
    <span class="dim">{formatBytes(disk.free)} free</span>
  </div>

  <!-- ── Footer: total capacity + filesystem type ─────────────────── -->
  <p class="footer dim">{formatBytes(disk.total)} · {disk.fstype}</p>
</article>

<style>
  .disk-card {
    flex: 1 1 180px;  /* Grow to fill space, shrink no smaller than 180px. */
    max-width: 260px; /* Cap width so cards don't balloon when there are few. */
    min-width: 0;     /* Allow shrinking below 180px if the parent forces it. */
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    transition: border-color 0.2s ease;
  }
  .disk-card:hover {
    border-color: var(--border-bright);
  }

  /* ── Header ──────────────────────────────────────────────────────── */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }
  .mount-label {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text);
  }
  .pct-label {
    font-size: 0.85rem;
    font-weight: 700;
    transition: color 0.3s ease;
  }

  /* ── Bar ─────────────────────────────────────────────────────────── */
  .bar-track {
    height: 5px;
    background: rgba(255, 255, 255, 0.07);
    border-radius: 99px;
    overflow: hidden;
  }
  .bar-fill {
    height: 100%;
    border-radius: 99px;
    transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  }

  /* ── Details ─────────────────────────────────────────────────────── */
  .byte-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.7rem;
  }
  .footer {
    font-size: 0.68rem;
    margin: 0;
  }
  .dim { color: var(--text-dim); }
</style>
