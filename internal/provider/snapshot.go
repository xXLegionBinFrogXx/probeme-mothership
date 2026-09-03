package provider

import "time"

const (
	MaxCPU          = 256
	MaxDisks        = 64
	MaxMounts       = 64
	MaxIfaces       = 64
	MaxThermalZones = 32
	MaxGPUs         = 8
)

const (
	CapCPU        uint64 = 1 << 0
	CapMemory     uint64 = 1 << 1
	CapLoadAvg    uint64 = 1 << 2
	CapUptime     uint64 = 1 << 3
	CapDiskIO     uint64 = 1 << 4
	CapFilesystem uint64 = 1 << 5
	CapNetDev     uint64 = 1 << 6
	CapThermal    uint64 = 1 << 7
	CapGPU        uint64 = 1 << 8

	MountRO      uint32 = 1 << 0
	MountSkipped uint32 = 1 << 1
)

type capName struct {
	bit  uint64
	name string
}

var capNames = []capName{
	{CapCPU, "cpu"},
	{CapMemory, "memory"},
	{CapLoadAvg, "loadavg"},
	{CapUptime, "uptime"},
	{CapDiskIO, "disk_io"},
	{CapFilesystem, "filesystem"},
	{CapNetDev, "netdev"},
	{CapThermal, "thermal"},
	{CapGPU, "gpu"},
}

// Snapshot is the Go-owned copy of struct pme_snapshot: units are the
// library's (jiffies, bytes, m°C, mW); conversions happen in internal/metrics.
type Snapshot struct {
	Generation  uint64
	PublishedAt time.Time
	Valid       uint64
	Truncated   uint64

	CPU        CPUSnapshot
	Memory     MemorySnapshot
	LoadAvg    LoadAvgSnapshot
	Uptime     UptimeSnapshot
	DiskIO     DiskIOSnapshot
	Filesystem FilesystemSnapshot
	NetDev     NetDevSnapshot
	Thermal    ThermalSnapshot
	GPU        GPUSnapshot
}

func (s *Snapshot) Has(cap uint64) bool { return s.Valid&cap != 0 }

type CPUSnapshot struct {
	ClkTck uint32
	ReadAt time.Time
	Cores  []CPUCore
}

type CPUCore struct {
	User, Nice, System, Idle, Iowait, IRQ, SoftIRQ, Steal uint64
}

type MemorySnapshot struct {
	ReadAt    time.Time
	Total     uint64
	Free      uint64
	Available uint64
	Buffers   uint64
	Cached    uint64
	SwapTotal uint64
	SwapFree  uint64
}

type LoadAvgSnapshot struct {
	ReadAt  time.Time
	Load1   float64
	Load5   float64
	Load15  float64
	Running uint32
	Total   uint32
}

type UptimeSnapshot struct {
	ReadAt        time.Time
	UptimeS       uint64
	BootTimeUnixS uint64
}

type Disk struct {
	Name         string
	Reads        uint64
	ReadBytes    uint64
	ReadTimeMS   uint64
	Writes       uint64
	WriteBytes   uint64
	WriteTimeMS  uint64
	IOInProgress uint64
	IOTimeMS     uint64
}

type DiskIOSnapshot struct {
	ReadAt time.Time
	Disks  []Disk
}

type Mount struct {
	Device     string
	Mountpoint string
	FSType     string
	Flags      uint32
	SizeBytes  uint64
	FreeBytes  uint64
	AvailBytes uint64
	Files      uint64
	FilesFree  uint64
}

type FilesystemSnapshot struct {
	ReadAt time.Time
	Mounts []Mount
}

type Iface struct {
	Name      string
	RxBytes   uint64
	RxPackets uint64
	RxErrs    uint64
	RxDrop    uint64
	TxBytes   uint64
	TxPackets uint64
	TxErrs    uint64
	TxDrop    uint64
}

type NetDevSnapshot struct {
	ReadAt time.Time
	Ifaces []Iface
}

type Zone struct {
	Type   string
	TempMC int64
}

type ThermalSnapshot struct {
	ReadAt time.Time
	Zones  []Zone
}

type GPUDev struct {
	UUID       string
	Name       string
	TempC      uint32
	PowerMW    uint32
	SMClockMHz uint32
	UtilPct    uint32
	Pstate     uint32
}

type GPUSnapshot struct {
	ReadAt time.Time
	GPUs   []GPUDev
}

func CapNamesFor(mask uint64) string {
	s := ""
	for _, cn := range capNames {
		if mask&cn.bit != 0 {
			if s != "" {
				s += ","
			}
			s += cn.name
		}
	}
	return s
}
