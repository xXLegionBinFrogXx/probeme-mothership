package provider

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	fakeSo   = "../../test/fakeprovider/build/libprobeme_fake.so"
	brokenSo = "../../test/fakeprovider/build/libprobeme_broken.so"
	allCaps  = uint64(0x1ff)
)

func requireSo(t *testing.T, path string) string {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s not built; run make fakeprovider", filepath.Base(path))
	}
	return path
}

func openFake(t *testing.T) *Provider {
	t.Helper()
	p, err := Open(requireSo(t, fakeSo))
	require.NoError(t, err)
	t.Cleanup(p.Close)
	require.NoError(t, p.Init(0))
	return p
}

func TestOpenFake(t *testing.T) {
	p, err := Open(requireSo(t, fakeSo))
	require.NoError(t, err)
	defer p.Close()
	require.Equal(t, "fake", p.Name())
	require.Equal(t, allCaps, p.Capabilities())
	require.Equal(t, uint32(ABIMajor)<<16, p.ABIVersion())
}

func TestOpenMissing(t *testing.T) {
	_, err := Open(filepath.Join(t.TempDir(), "libprobeme_nope.so"))
	require.Error(t, err)
	require.ErrorContains(t, err, "dlopen")
}

func TestOpenBrokenNoSymbol(t *testing.T) {
	_, err := Open(requireSo(t, brokenSo))
	require.Error(t, err)
	require.ErrorContains(t, err, "pme_provider_get")
}

func TestOpenABIMismatch(t *testing.T) {
	t.Setenv("PME_FAKE_ABI_MAJOR", "2")
	_, err := Open(requireSo(t, fakeSo))
	require.ErrorIs(t, err, ErrABIMismatch)
}

func TestCollectAllSectionsRoundTrip(t *testing.T) {
	p := openFake(t)
	buf, err := NewCBuffer()
	require.NoError(t, err)
	buf.Reset()
	require.NoError(t, p.CollectAll(buf))

	s := buf.Snapshot()
	require.NotZero(t, s.Valid)
	require.Zero(t, s.Truncated)
	require.True(t, s.Has(CapCPU))
	require.True(t, s.Has(CapMemory))
	require.True(t, s.Has(CapLoadAvg))
	require.True(t, s.Has(CapUptime))
	require.True(t, s.Has(CapDiskIO))
	require.True(t, s.Has(CapFilesystem))
	require.True(t, s.Has(CapNetDev))
	require.True(t, s.Has(CapThermal))
	require.True(t, s.Has(CapGPU))

	require.Equal(t, uint32(100), s.CPU.ClkTck)
	require.Len(t, s.CPU.Cores, 4)
	require.Equal(t, uint64(1000), s.CPU.Cores[0].User)
	require.Equal(t, uint64(1100), s.CPU.Cores[3].Steal)
	require.Equal(t, uint64(800), s.CPU.Cores[1].SoftIRQ)

	require.Equal(t, uint64(34359738368), s.Memory.Total)
	require.Equal(t, uint64(21474836480), s.Memory.Available)
	require.Equal(t, uint64(1073741824), s.Memory.SwapFree)

	require.Equal(t, 0.5, s.LoadAvg.Load1)
	require.Equal(t, 0.125, s.LoadAvg.Load15)
	require.Equal(t, uint32(2), s.LoadAvg.Running)

	require.Equal(t, uint64(86400), s.Uptime.UptimeS)
	require.Equal(t, uint64(1700000000), s.Uptime.BootTimeUnixS)

	require.Len(t, s.DiskIO.Disks, 2)
	require.Equal(t, "nvme0n1", s.DiskIO.Disks[0].Name)
	require.Equal(t, uint64(2048000), s.DiskIO.Disks[0].ReadBytes)
	require.Equal(t, uint64(1200), s.DiskIO.Disks[0].IOTimeMS)
	require.Equal(t, "nvme1n1", s.DiskIO.Disks[1].Name)

	require.Len(t, s.Filesystem.Mounts, 3)
	require.Equal(t, "/", s.Filesystem.Mounts[0].Mountpoint)
	require.Equal(t, "ext4", s.Filesystem.Mounts[0].FSType)
	require.Equal(t, uint64(500107862016), s.Filesystem.Mounts[0].SizeBytes)
	require.Equal(t, uint32(0), s.Filesystem.Mounts[0].Flags)
	require.NotZero(t, s.Filesystem.Mounts[2].Flags&MountSkipped)

	require.Len(t, s.NetDev.Ifaces, 2)
	require.Equal(t, "eth0", s.NetDev.Ifaces[0].Name)
	require.Equal(t, uint64(123456789), s.NetDev.Ifaces[0].RxBytes)
	require.Equal(t, uint64(987654321), s.NetDev.Ifaces[0].TxBytes)

	require.Len(t, s.Thermal.Zones, 2)
	require.Equal(t, "cpu", s.Thermal.Zones[0].Type)
	require.Equal(t, int64(45000), s.Thermal.Zones[0].TempMC)
	require.Equal(t, int64(38000), s.Thermal.Zones[1].TempMC)

	require.Len(t, s.GPU.GPUs, 2)
	require.Equal(t, "GPU-fake-0000", s.GPU.GPUs[0].UUID)
	require.Equal(t, "Fake GPU 0", s.GPU.GPUs[0].Name)
	require.Equal(t, uint32(120000), s.GPU.GPUs[0].PowerMW)
	require.Equal(t, uint32(2), s.GPU.GPUs[0].Pstate)
}

func TestCollectNotSupported(t *testing.T) {
	t.Setenv("PME_FAKE_ENOTSUP", "1")
	p := openFake(t)
	buf, err := NewCBuffer()
	require.NoError(t, err)
	err = p.CollectAll(buf)
	require.ErrorIs(t, err, ErrNotSupported)
}

func TestCollectEIO(t *testing.T) {
	t.Setenv("PME_FAKE_EIO", "1")
	p := openFake(t)
	buf, err := NewCBuffer()
	require.NoError(t, err)
	err = p.CollectAll(buf)
	require.ErrorIs(t, err, ErrIO)
}

func TestCollectTruncated(t *testing.T) {
	t.Setenv("PME_FAKE_TRUNCATE", "1")
	p := openFake(t)
	buf, err := NewCBuffer()
	require.NoError(t, err)
	buf.Reset()
	require.NoError(t, p.CollectAll(buf))
	s := buf.Snapshot()
	require.NotZero(t, s.Truncated)
	require.Equal(t, s.Valid, s.Truncated)
}

func TestBufferReset(t *testing.T) {
	p := openFake(t)
	buf, err := NewCBuffer()
	require.NoError(t, err)
	require.NoError(t, p.CollectAll(buf))
	buf.Reset()
	s := buf.Snapshot()
	require.Zero(t, s.Valid)
	require.Empty(t, s.CPU.Cores)
	require.Empty(t, s.DiskIO.Disks)
}

func TestLoadExplicitAndDedupe(t *testing.T) {
	so := requireSo(t, fakeSo)
	loaded, err := Load(Options{Explicit: []string{so}, Logger: testLogger(t)})
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, allCaps, loaded[0].Caps)
	for _, l := range loaded {
		l.Provider.Destroy()
		l.Provider.Close()
	}

	// Same path twice in one Load: the first provider claims every
	// capability, the duplicate is masked to zero and skipped.
	loaded, err = Load(Options{Explicit: []string{so, so}, Logger: testLogger(t)})
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, allCaps, loaded[0].Caps)
	for _, l := range loaded {
		l.Provider.Destroy()
		l.Provider.Close()
	}
}

func TestLoadExplicitFailureIsFatal(t *testing.T) {
	_, err := Load(Options{Explicit: []string{"/nonexistent/libprobeme_x.so"}, Logger: testLogger(t)})
	require.Error(t, err)
}

func TestLoadEmptyGlob(t *testing.T) {
	loaded, err := Load(Options{Dir: t.TempDir(), Logger: testLogger(t)})
	require.NoError(t, err)
	require.Empty(t, loaded)
}

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
