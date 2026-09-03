// Package provider is the only package in the repo that imports "C";
// all C code lives in this single file because each cgo file only sees
// its own preamble.
package provider

/*
#cgo pkg-config: probeme
#include <stdlib.h>
#include <string.h>
#include "probeme.h"
#include "shim.h"

static inline int pme_trampoline_init(const struct pme_provider *p, const struct pme_config *c) {
	return p->init(c);
}

static inline int pme_trampoline_collect_all(const struct pme_provider *p, struct pme_snapshot *s) {
	return p->collect_all(s);
}

static inline void pme_trampoline_destroy(const struct pme_provider *p) {
	p->destroy();
}

static inline const char *pme_zone_type(const struct pme_zone *z) {
	return z->type;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"time"
	"unsafe"
)

const ABIMajor = C.PME_ABI_MAJOR

var (
	ErrStructTooSmall = errors.New("provider struct smaller than expected")
	ErrABIMismatch    = errors.New("provider ABI major mismatch")
	ErrNotSupported   = errors.New("operation not supported")
	ErrIO             = errors.New("I/O error")
	ErrInvalid        = errors.New("invalid argument")
	ErrNotInit        = errors.New("provider not initialized")
)

type Provider struct {
	path         string
	name         string
	capabilities uint64
	abiVersion   uint32
	handle       unsafe.Pointer
	meta         *C.struct_pme_provider
}

func Open(path string) (*Provider, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var cerr *C.char
	h := C.pme_shim_open(cpath, &cerr)
	if h == nil {
		return nil, fmt.Errorf("dlopen %q: %s", path, cStr(cerr))
	}

	var gerr *C.char
	meta := C.pme_shim_get(h, &gerr)
	if meta == nil {
		C.pme_shim_close(h)
		if gerr != nil {
			return nil, fmt.Errorf("dlsym pme_provider_get in %q: %s", path, cStr(gerr))
		}
		return nil, fmt.Errorf("%q: pme_provider_get returned NULL", path)
	}

	declaredSize := uint64(meta.size)
	abiVersion := uint32(meta.abi_version)
	if declaredSize < uint64(C.sizeof_struct_pme_provider) {
		C.pme_shim_close(h)
		return nil, fmt.Errorf("%q: provider struct too small (%d < %d): %w",
			path, declaredSize, C.sizeof_struct_pme_provider, ErrStructTooSmall)
	}
	if abiVersion>>16 != ABIMajor {
		C.pme_shim_close(h)
		return nil, fmt.Errorf("%q: ABI major %d, want %d: %w",
			path, abiVersion>>16, ABIMajor, ErrABIMismatch)
	}

	return &Provider{
		path:         path,
		name:         C.GoString(meta.name),
		capabilities: uint64(meta.capabilities),
		abiVersion:   uint32(meta.abi_version),
		handle:       h,
		meta:         meta,
	}, nil
}

func (p *Provider) Init(flags uint64) error {
	var cfg C.struct_pme_config
	cfg.size = C.uint32_t(unsafe.Sizeof(cfg))
	cfg.flags = C.uint64_t(flags)
	return codeError(C.pme_trampoline_init(p.meta, &cfg), p.name, "init")
}

func (p *Provider) CollectAll(buf *CBuffer) error {
	return codeError(C.pme_trampoline_collect_all(p.meta, buf.c), p.name, "collect_all")
}

func (p *Provider) Destroy() {
	C.pme_trampoline_destroy(p.meta)
}

func (p *Provider) Close() {
	if p.handle != nil {
		C.pme_shim_close(p.handle)
		p.handle = nil
	}
}

func (p *Provider) Name() string         { return p.name }
func (p *Provider) Path() string         { return p.path }
func (p *Provider) Capabilities() uint64 { return p.capabilities }
func (p *Provider) ABIVersion() uint32   { return p.abiVersion }

type CBuffer struct {
	c *C.struct_pme_snapshot
}

func NewCBuffer() (*CBuffer, error) {
	c := (*C.struct_pme_snapshot)(C.malloc(C.size_t(unsafe.Sizeof(C.struct_pme_snapshot{}))))
	if c == nil {
		return nil, errors.New("probeme: failed to allocate snapshot buffer")
	}
	return &CBuffer{c: c}, nil
}

// Reset zeroes the buffer before a collect pass, then stamps the snapshot
// header: providers validate size before writing (libprobeme provider.c).
func (b *CBuffer) Reset() {
	C.memset(unsafe.Pointer(b.c), 0, C.size_t(unsafe.Sizeof(*b.c)))
	b.c.size = C.uint32_t(unsafe.Sizeof(*b.c))
	b.c.abi_version = (C.PME_ABI_MAJOR << 16) | C.PME_ABI_MINOR
}

func (b *CBuffer) Snapshot() *Snapshot {
	c := b.c
	s := &Snapshot{
		Valid:       uint64(c.valid),
		Truncated:   uint64(c.truncated),
		PublishedAt: time.Now(),
	}

	s.CPU.ClkTck = uint32(c.cpu.clk_tck)
	s.CPU.ReadAt = nsToTime(c.cpu.read_at_ns)
	if n := clamp(c.cpu.n, MaxCPU); n > 0 {
		s.CPU.Cores = make([]CPUCore, n)
		for i := range s.CPU.Cores {
			core := &c.cpu.cpu[i]
			s.CPU.Cores[i] = CPUCore{
				User:    uint64(core.user),
				Nice:    uint64(core.nice),
				System:  uint64(core.system),
				Idle:    uint64(core.idle),
				Iowait:  uint64(core.iowait),
				IRQ:     uint64(core.irq),
				SoftIRQ: uint64(core.softirq),
				Steal:   uint64(core.steal),
			}
		}
	}

	s.Memory.ReadAt = nsToTime(c.memory.read_at_ns)
	s.Memory.Total = uint64(c.memory.total)
	s.Memory.Free = uint64(c.memory.free)
	s.Memory.Available = uint64(c.memory.available)
	s.Memory.Buffers = uint64(c.memory.buffers)
	s.Memory.Cached = uint64(c.memory.cached)
	s.Memory.SwapTotal = uint64(c.memory.swap_total)
	s.Memory.SwapFree = uint64(c.memory.swap_free)

	s.LoadAvg.ReadAt = nsToTime(c.loadavg.read_at_ns)
	s.LoadAvg.Load1 = float64(c.loadavg.load1)
	s.LoadAvg.Load5 = float64(c.loadavg.load5)
	s.LoadAvg.Load15 = float64(c.loadavg.load15)
	s.LoadAvg.Running = uint32(c.loadavg.running)
	s.LoadAvg.Total = uint32(c.loadavg.total)

	s.Uptime.ReadAt = nsToTime(c.uptime.read_at_ns)
	s.Uptime.UptimeS = uint64(c.uptime.uptime_s)
	s.Uptime.BootTimeUnixS = uint64(c.uptime.boot_time_unix_s)

	s.DiskIO.ReadAt = nsToTime(c.disk_io.read_at_ns)
	if n := clamp(c.disk_io.n, MaxDisks); n > 0 {
		s.DiskIO.Disks = make([]Disk, n)
		for i := range s.DiskIO.Disks {
			d := &c.disk_io.disks[i]
			s.DiskIO.Disks[i] = Disk{
				Name:         C.GoString(&d.name[0]),
				Reads:        uint64(d.reads),
				ReadBytes:    uint64(d.read_bytes),
				ReadTimeMS:   uint64(d.read_time_ms),
				Writes:       uint64(d.writes),
				WriteBytes:   uint64(d.write_bytes),
				WriteTimeMS:  uint64(d.write_time_ms),
				IOInProgress: uint64(d.io_in_progress),
				IOTimeMS:     uint64(d.io_time_ms),
			}
		}
	}

	s.Filesystem.ReadAt = nsToTime(c.filesystem.read_at_ns)
	if n := clamp(c.filesystem.n, MaxMounts); n > 0 {
		s.Filesystem.Mounts = make([]Mount, n)
		for i := range s.Filesystem.Mounts {
			m := &c.filesystem.mounts[i]
			s.Filesystem.Mounts[i] = Mount{
				Device:     C.GoString(&m.device[0]),
				Mountpoint: C.GoString(&m.mountpoint[0]),
				FSType:     C.GoString(&m.fstype[0]),
				Flags:      uint32(m.flags),
				SizeBytes:  uint64(m.size_bytes),
				FreeBytes:  uint64(m.free_bytes),
				AvailBytes: uint64(m.avail_bytes),
				Files:      uint64(m.files),
				FilesFree:  uint64(m.files_free),
			}
		}
	}

	s.NetDev.ReadAt = nsToTime(c.netdev.read_at_ns)
	if n := clamp(c.netdev.n, MaxIfaces); n > 0 {
		s.NetDev.Ifaces = make([]Iface, n)
		for i := range s.NetDev.Ifaces {
			f := &c.netdev.ifaces[i]
			s.NetDev.Ifaces[i] = Iface{
				Name:      C.GoString(&f.name[0]),
				RxBytes:   uint64(f.rx_bytes),
				RxPackets: uint64(f.rx_packets),
				RxErrs:    uint64(f.rx_errs),
				RxDrop:    uint64(f.rx_drop),
				TxBytes:   uint64(f.tx_bytes),
				TxPackets: uint64(f.tx_packets),
				TxErrs:    uint64(f.tx_errs),
				TxDrop:    uint64(f.tx_drop),
			}
		}
	}

	s.Thermal.ReadAt = nsToTime(c.thermal.read_at_ns)
	if n := clamp(c.thermal.n, MaxThermalZones); n > 0 {
		s.Thermal.Zones = make([]Zone, n)
		for i := range s.Thermal.Zones {
			z := &c.thermal.zones[i]
			s.Thermal.Zones[i] = Zone{
				Type:   C.GoString(C.pme_zone_type(z)),
				TempMC: int64(z.temp_mc),
			}
		}
	}

	s.GPU.ReadAt = nsToTime(c.gpu.read_at_ns)
	if n := clamp(c.gpu.n, MaxGPUs); n > 0 {
		s.GPU.GPUs = make([]GPUDev, n)
		for i := range s.GPU.GPUs {
			g := &c.gpu.gpus[i]
			s.GPU.GPUs[i] = GPUDev{
				UUID:       C.GoString(&g.uuid[0]),
				Name:       C.GoString(&g.name[0]),
				TempC:      uint32(g.temp_c),
				PowerMW:    uint32(g.power_mw),
				SMClockMHz: uint32(g.sm_clock_mhz),
				UtilPct:    uint32(g.util_pct),
				Pstate:     uint32(g.pstate),
			}
		}
	}

	return s
}

func cStr(cs *C.char) string {
	if cs == nil {
		return ""
	}
	return C.GoString(cs)
}

func codeError(rc C.int, name, op string) error {
	switch rc {
	case C.PME_OK:
		return nil
	case C.PME_ENOTSUP:
		return fmt.Errorf("%s: %s: %w", name, op, ErrNotSupported)
	case C.PME_EIO:
		return fmt.Errorf("%s: %s: %w", name, op, ErrIO)
	case C.PME_EINVAL:
		return fmt.Errorf("%s: %s: %w", name, op, ErrInvalid)
	case C.PME_ENOINIT:
		return fmt.Errorf("%s: %s: %w", name, op, ErrNotInit)
	default:
		return fmt.Errorf("%s: %s: unknown error code %d", name, op, rc)
	}
}

func nsToTime(ns C.uint64_t) time.Time {
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(ns))
}

func clamp(n C.uint32_t, max int) int {
	if int(n) > max {
		return max
	}
	return int(n)
}
