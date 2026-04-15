<!--
  GaugeRing.svelte — Animated circular progress ring.

  Usage:
    <GaugeRing value={75} color="var(--cyan)">
      <span>75%</span>
    </GaugeRing>

  The ring rotates from the 12-o'clock position clockwise.
  Content passed via <slot> is centred inside the ring.
-->
<script>
  /** Current fill level, 0–100. Clamped internally so out-of-range values are safe. */
  export let value = 0;

  /** Outer diameter of the SVG in pixels. */
  export let size = 140;

  /** Stroke width of the ring in pixels. */
  export let thickness = 10;

  /** Stroke colour — any valid CSS colour or variable. */
  export let color = 'var(--cyan)';

  // ─── Geometry ──────────────────────────────────────────────────────────────
  // The radius sits in the middle of the stroke, so we inset it by half
  // the stroke width to keep the ring fully inside the SVG viewport.

  $: radius        = (size - thickness) / 2;
  $: centre        = size / 2;
  $: circumference = 2 * Math.PI * radius;

  // stroke-dashoffset controls how much of the ring is "filled".
  // At 0 the ring is fully drawn; at circumference it is empty.
  $: clampedValue  = Math.min(Math.max(value, 0), 100);
  $: dashOffset    = circumference * (1 - clampedValue / 100);
</script>

<!--
  The outer div establishes the stacking context so the SVG (absolute) and
  the slot content (relative) can overlap at the same centre point.
-->
<div class="gauge" style="width:{size}px; height:{size}px" role="img" aria-label="{clampedValue.toFixed(1)}%">
  <!--
    The SVG is rotated -90° so progress starts at the top rather than the
    right — the natural starting point for a clock-like gauge.
  -->
  <svg
    width={size}
    height={size}
    style="transform: rotate(-90deg)"
    aria-hidden="true"
  >
    <!-- Background track (full circle, low opacity) -->
    <circle
      cx={centre}
      cy={centre}
      r={radius}
      fill="none"
      stroke="rgba(255, 255, 255, 0.07)"
      stroke-width={thickness}
    />

    <!-- Foreground arc (animated fill) -->
    <circle
      cx={centre}
      cy={centre}
      r={radius}
      fill="none"
      stroke={color}
      stroke-width={thickness}
      stroke-linecap="round"
      stroke-dasharray={circumference}
      stroke-dashoffset={dashOffset}
      style="
        transition: stroke-dashoffset 0.7s cubic-bezier(0.4, 0, 0.2, 1),
                    stroke 0.4s ease;
        filter: drop-shadow(0 0 5px {color});
      "
    />
  </svg>

  <!-- Slot content is centred over the ring via absolute positioning. -->
  <div class="centre-content">
    <slot />
  </div>
</div>

<style>
  .gauge {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0; /* Prevent the gauge from squishing inside flex parents. */
  }

  svg {
    position: absolute;
    inset: 0;
  }

  .centre-content {
    position: relative; /* Sits above the SVG in stacking order. */
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    pointer-events: none; /* Allow clicks to pass through to the parent. */
  }
</style>
