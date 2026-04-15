<!--
  NetworkCard.svelte — Network throughput panel.

  Displays:
    - Current upload and download rates (bytes/sec).
    - Two 60-second sparkline charts: upload (cyan) and download (purple).
    - Cumulative total bytes sent and received since system boot.

  The history arrays are maintained in App.svelte and passed in as props,
  so this component stays stateless and purely presentational.
-->
<script>
  import Sparkline from './Sparkline.svelte';
  import { formatBytes, formatRate } from './utils.js';

  /**
   * Network snapshot from the Go backend.
   * @type {{
   *   bytesSent:   number,
   *   bytesRecv:   number,
   *   packetsSent: number,
   *   packetsRecv: number,
   *   sendRate:    number,
   *   recvRate:    number
   * }}
   */
  export let network;

  /**
   * Rolling 60-second history of upload rates (bytes/sec), newest last.
   * @type {number[]}
   */
  export let sendHistory = [];

  /**
   * Rolling 60-second history of download rates (bytes/sec), newest last.
   * @type {number[]}
   */
  export let recvHistory = [];
</script>

<article class="card" aria-label="Network">
  <h2 class="card-title">Network</h2>

  <!-- ── Live throughput rates ─────────────────────────────────────── -->
  <div class="rates">
    <div class="rate upload" aria-label="Upload rate">
      <span class="rate-arrow" aria-hidden="true">↑</span>
      <span class="rate-value">{formatRate(network.sendRate)}</span>
    </div>
    <div class="rate download" aria-label="Download rate">
      <span class="rate-arrow" aria-hidden="true">↓</span>
      <span class="rate-value">{formatRate(network.recvRate)}</span>
    </div>
  </div>

  <!-- ── Sparkline charts ───────────────────────────────────────────── -->
  <!--
    Each chart shows the last 60 seconds of data.
    The Y axis auto-scales to the peak value in the visible window.
  -->
  <div class="charts">
    <div class="chart-wrap" aria-label="Upload history">
      <Sparkline data={sendHistory} color="var(--cyan)" height={44} />
    </div>
    <div class="chart-wrap" aria-label="Download history">
      <Sparkline data={recvHistory} color="var(--purple)" height={44} />
    </div>
  </div>

  <!-- ── Cumulative totals ──────────────────────────────────────────── -->
  <dl class="totals">
    <div class="total-row">
      <dt>Total sent</dt>
      <dd>{formatBytes(network.bytesSent)}</dd>
    </div>
    <div class="total-row">
      <dt>Total recv</dt>
      <dd>{formatBytes(network.bytesRecv)}</dd>
    </div>
  </dl>
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
    gap: 10px;
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

  /* ── Live rates ────────────────────────────────────────────────────── */
  .rates {
    display: flex;
    gap: 20px;
  }
  .rate {
    display: flex;
    align-items: baseline;
    gap: 6px;
    font-weight: 600;
  }
  .rate.upload   { color: var(--cyan); }
  .rate.download { color: var(--purple); }
  .rate-arrow { font-size: 0.85rem; }
  .rate-value { font-size: 0.95rem; }

  /* ── Charts ────────────────────────────────────────────────────────── */
  .charts {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .chart-wrap {
    border-radius: 6px;
    overflow: hidden;
    background: rgba(255, 255, 255, 0.02);
  }

  /* ── Totals ────────────────────────────────────────────────────────── */
  .totals {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin: 0;
  }
  .total-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.72rem;
  }
  dt { color: var(--text-dim); }
  dd { margin: 0; color: var(--text); }
</style>
