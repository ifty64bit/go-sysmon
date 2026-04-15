<!--
  Sparkline.svelte — A minimal time-series area chart.

  Renders an SVG polyline + filled area from an array of numbers.
  The Y axis auto-scales to the maximum value in the current data window,
  with a minimum scale floor (minScale) to avoid jitter on near-zero data.

  Sizing model:
    - `width` and `height` define the internal SVG coordinate space (viewBox),
      NOT the rendered pixel size. This keeps the point math simple and stable.
    - The rendered width is always 100% of the parent container via CSS.
    - `preserveAspectRatio="none"` lets the SVG stretch horizontally to fill
      whatever space it is given without letterboxing.
    - Only `height` controls the rendered pixel height.

  Usage:
    <Sparkline data={sendHistory} color="var(--cyan)" height={48} />
-->
<script>
  /** Array of numeric values to plot (oldest → newest, left → right). */
  export let data = /** @type {number[]} */ ([]);

  /**
   * Internal coordinate width used for point calculations.
   * Does NOT set the rendered pixel width — CSS handles that (width: 100%).
   */
  export let width = 200;

  /** Rendered pixel height of the chart. */
  export let height = 48;

  /** Stroke and fill colour. */
  export let color = 'var(--cyan)';

  /**
   * Minimum Y-axis ceiling in the same units as `data`.
   * Prevents the chart from going wild on near-zero values (e.g. idle network).
   * Default: 1 KB so the chart looks calm when nothing is happening.
   */
  export let minScale = 1024;

  // ─── Derived geometry ──────────────────────────────────────────────────────

  // Scale the Y axis to the largest value seen, but never below minScale.
  $: yMax = Math.max(...data, minScale);

  // Map each data point to an (x, y) coordinate pair.
  $: points = data.length < 2
    ? ''
    : data.map((value, index) => {
        const x = (index / (data.length - 1)) * width;
        // Leave 1px padding top and bottom so the line is never clipped.
        const y = height - (value / yMax) * (height - 2) - 1;
        return `${x.toFixed(1)},${y.toFixed(1)}`;
      }).join(' ');

  // Close the shape along the bottom for the filled area polygon.
  $: fillPoints = points ? `${points} ${width},${height} 0,${height}` : '';
</script>

<svg
  viewBox="0 0 {width} {height}"
  {height}
  preserveAspectRatio="none"
  class="sparkline"
  aria-hidden="true"
>
  <!-- Filled area beneath the line (low opacity for subtlety) -->
  {#if fillPoints}
    <polygon
      points={fillPoints}
      fill={color}
      opacity="0.12"
    />
  {/if}

  <!-- The line itself -->
  {#if points}
    <polyline
      {points}
      fill="none"
      stroke={color}
      stroke-width="1.5"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  {/if}
</svg>

<style>
  .sparkline {
    display: block;
    width: 100%;   /* Stretch to fill the parent container. */
    overflow: visible;
  }
</style>
