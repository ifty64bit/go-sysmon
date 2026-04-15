/**
 * utils.js — Pure helper functions shared across all card components.
 *
 * Rules for this file:
 *  - No imports, no side effects, no Svelte/DOM references.
 *  - Every function is a pure transformation: same input → same output.
 *  - This makes them trivial to test and easy to reason about.
 */

// ─── Byte formatting ──────────────────────────────────────────────────────────

/**
 * Converts a raw byte count into a human-readable string.
 *
 * @param {number} bytes     - Raw byte count (0 or positive integer).
 * @param {number} [decimals=1] - Number of decimal places in the output.
 * @returns {string} e.g. "0 B", "1.5 KB", "3.2 GB"
 */
export function formatBytes(bytes, decimals = 1) {
  if (!bytes || bytes <= 0) return '0 B';

  const UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const K = 1024;

  // Find the largest unit where the value is ≥ 1.
  const unitIndex = Math.min(
    Math.floor(Math.log(bytes) / Math.log(K)),
    UNITS.length - 1
  );

  const value = (bytes / Math.pow(K, unitIndex)).toFixed(decimals);
  return `${value} ${UNITS[unitIndex]}`;
}

/**
 * Formats a bytes-per-second rate as a human-readable string.
 *
 * @param {number} bytesPerSec
 * @returns {string} e.g. "0 B/s", "1.2 MB/s"
 */
export function formatRate(bytesPerSec) {
  return `${formatBytes(bytesPerSec)}/s`;
}

// ─── Time formatting ──────────────────────────────────────────────────────────

/**
 * Converts an uptime in seconds to a concise human-readable string.
 * Shows the two largest non-zero units so the display stays compact.
 *
 * @param {number} seconds - Uptime in seconds.
 * @returns {string} e.g. "2d 4h", "45m 30s", "—" if seconds is falsy.
 */
export function formatUptime(seconds) {
  if (!seconds) return '—';

  const days    = Math.floor(seconds / 86400);
  const hours   = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs    = seconds % 60;

  if (days  > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m ${secs}s`;
}

// ─── Color thresholds ─────────────────────────────────────────────────────────

/**
 * Returns a CSS variable name for a usage percentage.
 *
 * Thresholds:
 *   < 60  → green  (healthy)
 *   60–84 → orange (elevated)
 *   ≥ 85  → red    (high / warning)
 *
 * @param {number} pct - Usage percentage (0–100).
 * @returns {string} A CSS custom property reference, e.g. "var(--green)".
 */
export function usageColor(pct) {
  if (pct < 60) return 'var(--green)';
  if (pct < 85) return 'var(--orange)';
  return 'var(--red)';
}

/**
 * Returns a CSS variable name for a disk fill percentage.
 *
 * Thresholds:
 *   < 70  → cyan   (plenty of space)
 *   70–89 → orange (getting full)
 *   ≥ 90  → red    (critically full)
 *
 * @param {number} pct - Disk usage percentage (0–100).
 * @returns {string} A CSS custom property reference.
 */
export function diskColor(pct) {
  if (pct < 70) return 'var(--cyan)';
  if (pct < 90) return 'var(--orange)';
  return 'var(--red)';
}

// ─── Disk display ─────────────────────────────────────────────────────────────

/**
 * Derives a short display label from a disk mountpoint.
 *
 * Examples:
 *   "C:\"        → "C:"   (Windows drive letter)
 *   "/"          → "/"    (Unix root)
 *   "/home/user" → "user" (last path segment)
 *
 * @param {{ mountpoint: string }} disk
 * @returns {string}
 */
export function diskLabel(disk) {
  const mp = disk.mountpoint;

  // Windows drive letters like "C:\" or "D:/" → strip trailing slash.
  if (/^[A-Za-z]:[\\\/]?$/.test(mp)) {
    return mp.replace(/[\\\/]$/, '');
  }

  // Short paths (root "/", or anything ≤ 5 chars) → show as-is.
  if (mp.length <= 5) return mp;

  // Unix paths → show the last non-empty segment.
  const segments = mp.split('/').filter(Boolean);
  return segments.at(-1) ?? mp;
}
