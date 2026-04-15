package main

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// PollInterval is how often system metrics are collected and pushed to the
// frontend. One second gives smooth animations without noticeable CPU overhead.
const PollInterval = time.Second

// ─────────────────────────────────────────────────────────────────────────────
// SysInfoService
//
// This struct is registered as a Wails service. All exported methods
// automatically become RPC calls that the Svelte frontend can invoke.
//
// Design notes:
//   - Static data (CPU model, OS info) is queried once at startup and cached,
//     because those WMI / sysctl calls can take hundreds of milliseconds.
//   - Dynamic data (CPU %, memory, disk, network) is read on every poll tick.
//   - Network throughput is calculated by comparing two successive I/O counter
//     readings and dividing the delta by elapsed time.
// ─────────────────────────────────────────────────────────────────────────────

// SysInfoService collects and exposes system resource metrics.
type SysInfoService struct {
	// cpuCache stores hardware constants that never change at runtime.
	cpuCache cpuStaticInfo

	// osCache stores host metadata. Only Uptime is refreshed each tick.
	osCache OSInfo

	// netSnapshot is the previous network I/O reading used to compute rates.
	netSnapshot netSnapshot
}

// cpuStaticInfo holds CPU properties that are queried once and reused.
type cpuStaticInfo struct {
	modelName     string
	physicalCores int
	logicalCores  int
	freqMHz       float64
}

// netSnapshot records a single network I/O counter reading with its timestamp,
// so the next reading can derive bytes-per-second rates.
type netSnapshot struct {
	bytesSent uint64
	bytesRecv uint64
	takenAt   time.Time
}

// ─── Startup ─────────────────────────────────────────────────────────────────

// setup runs once before polling begins. It populates the caches and warms up
// the gopsutil CPU sampler so the first real reading is immediately accurate.
func (s *SysInfoService) setup() {
	s.cacheCPUInfo()
	s.cacheOSInfo()
	s.warmUpCPUSampler()
	s.warmUpNetSampler()
}

// cacheCPUInfo queries slow/static CPU properties and stores them in s.cpuCache.
func (s *SysInfoService) cacheCPUInfo() {
	if info, err := cpu.Info(); err == nil && len(info) > 0 {
		s.cpuCache.modelName = info[0].ModelName
		s.cpuCache.freqMHz = info[0].Mhz
		s.cpuCache.physicalCores = int(info[0].Cores)
	}

	s.cpuCache.logicalCores, _ = cpu.Counts(true)

	// Some systems report 0 physical cores via Info(); fall back to Counts().
	if s.cpuCache.physicalCores == 0 {
		s.cpuCache.physicalCores, _ = cpu.Counts(false)
	}
}

// cacheOSInfo queries host metadata and stores it in s.osCache.
func (s *SysInfoService) cacheOSInfo() {
	hi, err := host.Info()
	if err != nil {
		// Degrade gracefully: only OS family is guaranteed via the standard library.
		s.osCache = OSInfo{OS: runtime.GOOS}
		return
	}

	s.osCache = OSInfo{
		OS:       hi.OS,
		Platform: hi.Platform + " " + hi.PlatformVersion,
		Hostname: hi.Hostname,
		Uptime:   hi.Uptime,
	}
}

// warmUpCPUSampler makes the first cpu.Percent() call, which sets gopsutil's
// internal baseline. Without this, the first real poll would return 0%.
func (s *SysInfoService) warmUpCPUSampler() {
	_, _ = cpu.Percent(0, true)
}

// warmUpNetSampler records the initial network counter so the first poll can
// produce a meaningful rate instead of zero.
func (s *SysInfoService) warmUpNetSampler() {
	counters, err := psnet.IOCounters(false) // false = aggregate all interfaces
	if err != nil || len(counters) == 0 {
		return
	}

	s.netSnapshot = netSnapshot{
		bytesSent: counters[0].BytesSent,
		bytesRecv: counters[0].BytesRecv,
		takenAt:   time.Now(),
	}
}

// ─── Per-subsystem collectors ─────────────────────────────────────────────────

// collectCPU returns the current CPU snapshot, blending cached static info
// with a fresh per-core utilisation reading.
func (s *SysInfoService) collectCPU() CPUStats {
	stats := CPUStats{
		ModelName:    s.cpuCache.modelName,
		Cores:        s.cpuCache.physicalCores,
		LogicalCores: s.cpuCache.logicalCores,
		FreqMHz:      s.cpuCache.freqMHz,
	}

	// Percent(0, true) uses the elapsed time since the last call as the sample
	// window — perfect for a regular 1-second polling loop.
	perCore, err := cpu.Percent(0, true)
	if err != nil || len(perCore) == 0 {
		return stats
	}

	// Compute the average ourselves so we also have the per-core slice.
	var sum float64
	for _, p := range perCore {
		sum += p
	}

	stats.Usage = sum / float64(len(perCore))
	stats.PerCore = perCore
	return stats
}

// collectMemory returns a snapshot of physical and swap memory usage.
func (s *SysInfoService) collectMemory() MemStats {
	stats := MemStats{}

	if v, err := mem.VirtualMemory(); err == nil {
		stats.Total = v.Total
		stats.Used = v.Used
		stats.Available = v.Available
		stats.UsedPct = v.UsedPercent
	}

	if sw, err := mem.SwapMemory(); err == nil {
		stats.SwapTotal = sw.Total
		stats.SwapUsed = sw.Used
	}

	return stats
}

// collectDisks enumerates physical partitions and returns usage for each.
// Partitions with zero total size (e.g. pseudo-filesystems) are skipped.
func (s *SysInfoService) collectDisks() []DiskStats {
	// false = skip pseudo/virtual filesystems (proc, tmpfs, etc.)
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	disks := make([]DiskStats, 0, len(partitions))

	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		disks = append(disks, DiskStats{
			Mountpoint: p.Mountpoint,
			Device:     p.Device,
			Fstype:     p.Fstype,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			UsedPct:    usage.UsedPercent,
		})
	}

	return disks
}

// collectNetwork returns current I/O counters and computes throughput rates
// by comparing against the previous snapshot stored in s.netSnapshot.
func (s *SysInfoService) collectNetwork() NetStats {
	// false = return one aggregate entry across all interfaces
	counters, err := psnet.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return NetStats{}
	}

	current := counters[0]
	now := time.Now()

	stats := NetStats{
		BytesSent:   current.BytesSent,
		BytesRecv:   current.BytesRecv,
		PacketsSent: current.PacketsSent,
		PacketsRecv: current.PacketsRecv,
	}

	// Calculate per-second rates only when we have a previous sample to diff against.
	prev := s.netSnapshot
	if !prev.takenAt.IsZero() {
		elapsed := now.Sub(prev.takenAt).Seconds()

		if elapsed > 0 {
			// Guard against counter resets (e.g. after a network adapter change).
			if current.BytesSent >= prev.bytesSent {
				stats.SendRate = uint64(float64(current.BytesSent-prev.bytesSent) / elapsed)
			}
			if current.BytesRecv >= prev.bytesRecv {
				stats.RecvRate = uint64(float64(current.BytesRecv-prev.bytesRecv) / elapsed)
			}
		}
	}

	// Update snapshot for the next call.
	s.netSnapshot = netSnapshot{
		bytesSent: current.BytesSent,
		bytesRecv: current.BytesRecv,
		takenAt:   now,
	}

	return stats
}

// collectOS returns OS metadata. Uptime is re-queried each call since it
// changes every second; the rest comes from the cache populated at startup.
func (s *SysInfoService) collectOS() OSInfo {
	// Take a copy so we don't mutate the cache.
	osInfo := s.osCache

	if hi, err := host.Info(); err == nil {
		osInfo.Uptime = hi.Uptime
	}

	return osInfo
}

// ─── Public API ──────────────────────────────────────────────────────────────

// GetStats assembles and returns a complete system snapshot.
// This is an exported Wails RPC method — the frontend can call it on demand,
// though in practice the polling loop (StartPolling) makes that unnecessary.
func (s *SysInfoService) GetStats() SystemStats {
	return SystemStats{
		CPU:       s.collectCPU(),
		Memory:    s.collectMemory(),
		Disks:     s.collectDisks(),
		Network:   s.collectNetwork(),
		OS:        s.collectOS(),
		Timestamp: time.Now().UnixMilli(),
	}
}

// StartPolling runs a blocking loop that emits a "stats" event to the frontend
// once per PollInterval. It must be called in a goroutine (see main.go).
func (s *SysInfoService) StartPolling(app *application.App) {
	s.setup()

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		app.Event.Emit("stats", s.GetStats())
	}
}
