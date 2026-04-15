// Package main is the entry point for SysMon, a lightweight system resource
// monitor built with Wails v3 (Go backend) and Svelte (frontend).
package main

// CPUStats holds a point-in-time snapshot of CPU usage and static hardware info.
type CPUStats struct {
	// Usage is the average utilisation across all logical cores (0–100).
	Usage float64 `json:"usage"`

	// PerCore holds the individual utilisation for each logical core (0–100).
	PerCore []float64 `json:"perCore"`

	// Cores is the number of physical CPU cores.
	Cores int `json:"cores"`

	// LogicalCores is cores × hardware threads (what the OS schedules on).
	LogicalCores int `json:"logicalCores"`

	// FreqMHz is the base clock speed reported by the OS, in megahertz.
	FreqMHz float64 `json:"freqMHz"`

	// ModelName is the CPU brand string, e.g. "Intel Core i9-13900K".
	ModelName string `json:"modelName"`
}

// MemStats holds a point-in-time snapshot of system memory usage.
type MemStats struct {
	// Total is the total physical RAM in bytes.
	Total uint64 `json:"total"`

	// Used is the amount of RAM currently in use, in bytes.
	Used uint64 `json:"used"`

	// Available is the RAM that can be given to processes immediately, in bytes.
	Available uint64 `json:"available"`

	// UsedPct is Used / Total expressed as a percentage (0–100).
	UsedPct float64 `json:"usedPct"`

	// SwapTotal is the total swap/page-file space in bytes (0 if none configured).
	SwapTotal uint64 `json:"swapTotal"`

	// SwapUsed is the swap currently committed, in bytes.
	SwapUsed uint64 `json:"swapUsed"`
}

// DiskStats holds a point-in-time snapshot of a single mounted filesystem.
type DiskStats struct {
	// Mountpoint is the path where this filesystem is mounted (e.g. "C:\" or "/").
	Mountpoint string `json:"mountpoint"`

	// Device is the underlying block device (e.g. "/dev/sda1").
	Device string `json:"device"`

	// Fstype is the filesystem type (e.g. "ntfs", "ext4", "apfs").
	Fstype string `json:"fstype"`

	// Total, Used, and Free are all in bytes.
	Total uint64  `json:"total"`
	Used  uint64  `json:"used"`
	Free  uint64  `json:"free"`

	// UsedPct is Used / Total expressed as a percentage (0–100).
	UsedPct float64 `json:"usedPct"`
}

// NetStats holds a point-in-time snapshot of cumulative network I/O counters
// and the derived per-second throughput rates.
type NetStats struct {
	// BytesSent and BytesRecv are cumulative totals since the OS booted.
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`

	// PacketsSent and PacketsRecv are the packet-level equivalents.
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`

	// SendRate and RecvRate are the current throughput in bytes per second,
	// derived by comparing two successive counter readings.
	SendRate uint64 `json:"sendRate"`
	RecvRate uint64 `json:"recvRate"`
}

// OSInfo holds host and operating-system metadata that changes rarely
// (except Uptime, which increments every second).
type OSInfo struct {
	// OS is the operating system family (e.g. "windows", "linux", "darwin").
	OS string `json:"os"`

	// Platform is the specific distribution or edition with version
	// (e.g. "Windows 11 Pro 23H2", "ubuntu 22.04").
	Platform string `json:"platform"`

	// Hostname is the machine's network name.
	Hostname string `json:"hostname"`

	// Uptime is the number of seconds since the system was last booted.
	Uptime uint64 `json:"uptime"`
}

// SystemStats is the top-level payload emitted to the frontend once per second.
// It bundles all sub-snapshots together so the frontend receives a single
// coherent update instead of multiple partial events.
type SystemStats struct {
	CPU     CPUStats    `json:"cpu"`
	Memory  MemStats    `json:"memory"`
	Disks   []DiskStats `json:"disks"`
	Network NetStats    `json:"network"`
	OS      OSInfo      `json:"os"`

	// Timestamp is Unix milliseconds at the moment the snapshot was taken.
	Timestamp int64 `json:"timestamp"`
}
