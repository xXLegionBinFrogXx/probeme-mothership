package probe

import (
	"sync/atomic"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/provider"
)

// Published hands the latest snapshot to readers: latest wins, readers
// never block writers.
type Published struct {
	snap atomic.Pointer[provider.Snapshot]
}

func NewPublished() *Published { return &Published{} }

func (p *Published) Store(s *provider.Snapshot) { p.snap.Store(s) }

func (p *Published) Load() *provider.Snapshot { return p.snap.Load() }

// Re-exports so downstream packages can reference these via probe.
type (
	Snapshot        = provider.Snapshot
	CPUSnapshot     = provider.CPUSnapshot
	CPUCore         = provider.CPUCore
	MemorySnapshot  = provider.MemorySnapshot
	LoadAvgSnapshot = provider.LoadAvgSnapshot
	UptimeSnapshot  = provider.UptimeSnapshot
	Disk            = provider.Disk
	DiskIOSnapshot  = provider.DiskIOSnapshot
	Mount           = provider.Mount
	Filesystem      = provider.FilesystemSnapshot
	NetDevSnapshot  = provider.NetDevSnapshot
	Iface           = provider.Iface
	Zone            = provider.Zone
	ThermalSnapshot = provider.ThermalSnapshot
	GPUDev          = provider.GPUDev
	GPUSnapshot     = provider.GPUSnapshot
)

const (
	CapCPU        = provider.CapCPU
	CapMemory     = provider.CapMemory
	CapLoadAvg    = provider.CapLoadAvg
	CapUptime     = provider.CapUptime
	CapDiskIO     = provider.CapDiskIO
	CapFilesystem = provider.CapFilesystem
	CapNetDev     = provider.CapNetDev
	CapThermal    = provider.CapThermal
	CapGPU        = provider.CapGPU

	MountRO      = provider.MountRO
	MountSkipped = provider.MountSkipped
)

func (p *Published) Ready() bool {
	s := p.snap.Load()
	return s != nil && s.Valid != 0
}
