package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

// fixture mirrors test/fakeprovider canned values.
func fixture() *probe.Snapshot {
	s := &probe.Snapshot{
		Valid: probe.CapCPU | probe.CapMemory | probe.CapLoadAvg | probe.CapUptime |
			probe.CapDiskIO | probe.CapFilesystem | probe.CapNetDev |
			probe.CapThermal | probe.CapGPU,
	}
	s.CPU.ClkTck = 100
	s.CPU.Cores = []probe.CPUCore{
		{User: 1000, Nice: 200, System: 300, Idle: 400, Iowait: 500, IRQ: 600, SoftIRQ: 700, Steal: 800},
		{User: 1100, Nice: 300, System: 400, Idle: 500, Iowait: 600, IRQ: 700, SoftIRQ: 800, Steal: 900},
	}
	s.Memory = probe.MemorySnapshot{
		Total: 34359738368, Free: 17179869184, Available: 21474836480,
		Buffers: 1073741824, Cached: 4294967296,
		SwapTotal: 2147483648, SwapFree: 1073741824,
	}
	s.LoadAvg = probe.LoadAvgSnapshot{Load1: 0.5, Load5: 0.25, Load15: 0.125, Running: 2, Total: 16}
	s.Uptime = probe.UptimeSnapshot{UptimeS: 86400, BootTimeUnixS: 1700000000}
	s.DiskIO.Disks = []probe.Disk{
		{Name: "nvme0n1", Reads: 1000, ReadBytes: 2048000, ReadTimeMS: 500,
			Writes: 2000, WriteBytes: 4096000, WriteTimeMS: 700, IOInProgress: 1, IOTimeMS: 1200},
		{Name: "nvme1n1", Reads: 10, ReadBytes: 20, ReadTimeMS: 30,
			Writes: 40, WriteBytes: 50, WriteTimeMS: 60, IOTimeMS: 70},
	}
	s.Filesystem.Mounts = []probe.Mount{
		{Device: "/dev/nvme0n1p2", Mountpoint: "/", FSType: "ext4",
			SizeBytes: 500107862016, FreeBytes: 250053931008, AvailBytes: 200000000000,
			Files: 30541877, FilesFree: 15432807},
		{Device: "/dev/nvme0n1p1", Mountpoint: "/boot", FSType: "vfat",
			SizeBytes: 536870912, FreeBytes: 268435456, AvailBytes: 268435456},
		{Device: "proc", Mountpoint: "/proc", FSType: "proc", Flags: probe.MountSkipped},
	}
	s.NetDev.Ifaces = []probe.Iface{
		{Name: "eth0", RxBytes: 123456789, RxPackets: 100000, RxErrs: 3, RxDrop: 7,
			TxBytes: 987654321, TxPackets: 200000, TxErrs: 1, TxDrop: 2},
		{Name: "lo", RxBytes: 1, RxPackets: 1, TxBytes: 1, TxPackets: 1},
	}
	s.Thermal.Zones = []probe.Zone{
		{Type: "cpu", TempMC: 45000},
		{Type: "acpitz", TempMC: 38000},
	}
	s.GPU.GPUs = []probe.GPUDev{
		{UUID: "GPU-fake-0000", Name: "Fake GPU 0", TempC: 55, PowerMW: 120000, SMClockMHz: 1200, UtilPct: 37, Pstate: 2},
		{UUID: "GPU-fake-0001", Name: "Fake GPU 1", TempC: 61, PowerMW: 230000, SMClockMHz: 2100},
	}
	return s
}

func testCollector(t *testing.T, s *probe.Snapshot, f Filters) (*Collector, *prometheus.CounterVec) {
	t.Helper()
	dropped := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pme_series_dropped_total",
	}, []string{"family"})
	pub := probe.NewPublished()
	if s != nil {
		pub.Store(s)
	}
	return NewCollector(pub, f, dropped), dropped
}

func TestCollectCPU(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.
# TYPE node_cpu_seconds_total counter
node_cpu_seconds_total{cpu="0",mode="user"} 10
node_cpu_seconds_total{cpu="0",mode="nice"} 2
node_cpu_seconds_total{cpu="0",mode="system"} 3
node_cpu_seconds_total{cpu="0",mode="idle"} 4
node_cpu_seconds_total{cpu="0",mode="iowait"} 5
node_cpu_seconds_total{cpu="0",mode="irq"} 6
node_cpu_seconds_total{cpu="0",mode="softirq"} 7
node_cpu_seconds_total{cpu="0",mode="steal"} 8
node_cpu_seconds_total{cpu="1",mode="user"} 11
node_cpu_seconds_total{cpu="1",mode="nice"} 3
node_cpu_seconds_total{cpu="1",mode="system"} 4
node_cpu_seconds_total{cpu="1",mode="idle"} 5
node_cpu_seconds_total{cpu="1",mode="iowait"} 6
node_cpu_seconds_total{cpu="1",mode="irq"} 7
node_cpu_seconds_total{cpu="1",mode="softirq"} 8
node_cpu_seconds_total{cpu="1",mode="steal"} 9
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected), "node_cpu_seconds_total"))
}

func TestCollectMemoryLoadUptime(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP node_memory_MemTotal_bytes Total memory in bytes.
# TYPE node_memory_MemTotal_bytes gauge
node_memory_MemTotal_bytes 3.4359738368e+10
# HELP node_memory_MemFree_bytes Free memory in bytes.
# TYPE node_memory_MemFree_bytes gauge
node_memory_MemFree_bytes 1.7179869184e+10
# HELP node_memory_MemAvailable_bytes Available memory in bytes.
# TYPE node_memory_MemAvailable_bytes gauge
node_memory_MemAvailable_bytes 2.147483648e+10
# HELP node_memory_Buffers_bytes Buffer memory in bytes.
# TYPE node_memory_Buffers_bytes gauge
node_memory_Buffers_bytes 1.073741824e+09
# HELP node_memory_Cached_bytes Page cache memory in bytes.
# TYPE node_memory_Cached_bytes gauge
node_memory_Cached_bytes 4.294967296e+09
# HELP node_memory_SwapTotal_bytes Total swap in bytes.
# TYPE node_memory_SwapTotal_bytes gauge
node_memory_SwapTotal_bytes 2.147483648e+09
# HELP node_memory_SwapFree_bytes Free swap in bytes.
# TYPE node_memory_SwapFree_bytes gauge
node_memory_SwapFree_bytes 1.073741824e+09
# HELP node_load1 1m load average.
# TYPE node_load1 gauge
node_load1 0.5
# HELP node_load5 5m load average.
# TYPE node_load5 gauge
node_load5 0.25
# HELP node_load15 15m load average.
# TYPE node_load15 gauge
node_load15 0.125
# HELP node_boot_time_seconds Node boot time, in unixtime.
# TYPE node_boot_time_seconds gauge
node_boot_time_seconds 1.7e+09
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected),
		"node_memory_MemTotal_bytes", "node_memory_MemFree_bytes", "node_memory_MemAvailable_bytes",
		"node_memory_Buffers_bytes", "node_memory_Cached_bytes",
		"node_memory_SwapTotal_bytes", "node_memory_SwapFree_bytes",
		"node_load1", "node_load5", "node_load15", "node_boot_time_seconds"))
}

func TestCollectFilesystem(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP node_filesystem_size_bytes Filesystem size in bytes.
# TYPE node_filesystem_size_bytes gauge
node_filesystem_size_bytes{device="/dev/nvme0n1p1",fstype="vfat",mountpoint="/boot"} 5.36870912e+08
node_filesystem_size_bytes{device="/dev/nvme0n1p2",fstype="ext4",mountpoint="/"} 5.00107862016e+11
# HELP node_filesystem_free_bytes Filesystem free space in bytes.
# TYPE node_filesystem_free_bytes gauge
node_filesystem_free_bytes{device="/dev/nvme0n1p1",fstype="vfat",mountpoint="/boot"} 2.68435456e+08
node_filesystem_free_bytes{device="/dev/nvme0n1p2",fstype="ext4",mountpoint="/"} 2.50053931008e+11
# HELP node_filesystem_avail_bytes Filesystem space available to non-root users in bytes.
# TYPE node_filesystem_avail_bytes gauge
node_filesystem_avail_bytes{device="/dev/nvme0n1p1",fstype="vfat",mountpoint="/boot"} 2.68435456e+08
node_filesystem_avail_bytes{device="/dev/nvme0n1p2",fstype="ext4",mountpoint="/"} 2e+11
# HELP node_filesystem_files Filesystem total file nodes.
# TYPE node_filesystem_files gauge
node_filesystem_files{device="/dev/nvme0n1p1",fstype="vfat",mountpoint="/boot"} 0
node_filesystem_files{device="/dev/nvme0n1p2",fstype="ext4",mountpoint="/"} 3.0541877e+07
# HELP node_filesystem_files_free Filesystem total free file nodes.
# TYPE node_filesystem_files_free gauge
node_filesystem_files_free{device="/dev/nvme0n1p1",fstype="vfat",mountpoint="/boot"} 0
node_filesystem_files_free{device="/dev/nvme0n1p2",fstype="ext4",mountpoint="/"} 1.5432807e+07
# HELP node_filesystem_readonly Filesystem read-only status: 1 yes, 0 no.
# TYPE node_filesystem_readonly gauge
node_filesystem_readonly{device="/dev/nvme0n1p1",fstype="vfat",mountpoint="/boot"} 0
node_filesystem_readonly{device="/dev/nvme0n1p2",fstype="ext4",mountpoint="/"} 0
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected),
		"node_filesystem_size_bytes", "node_filesystem_free_bytes", "node_filesystem_avail_bytes",
		"node_filesystem_files", "node_filesystem_files_free", "node_filesystem_readonly"))
}

func TestCollectFilesystemFilters(t *testing.T) {
	// /proc is skipped at the source; /boot and vfat by filters.
	f := Filters{MountPoints: []string{"/boot"}, FSTypes: []string{"vfat"}}
	c, _ := testCollector(t, fixture(), f)

	count := testutil.CollectAndCount(c, "node_filesystem_size_bytes")
	require.Equal(t, 1, count, "only the / mount must survive filters")
}

func TestCollectDisk(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP node_disk_reads_completed_total Total successful reads.
# TYPE node_disk_reads_completed_total counter
node_disk_reads_completed_total{device="nvme0n1"} 1000
node_disk_reads_completed_total{device="nvme1n1"} 10
# HELP node_disk_read_bytes_total Total bytes read successfully.
# TYPE node_disk_read_bytes_total counter
node_disk_read_bytes_total{device="nvme0n1"} 2.048e+06
node_disk_read_bytes_total{device="nvme1n1"} 20
# HELP node_disk_read_time_seconds_total Total seconds spent on all reads.
# TYPE node_disk_read_time_seconds_total counter
node_disk_read_time_seconds_total{device="nvme0n1"} 0.5
node_disk_read_time_seconds_total{device="nvme1n1"} 0.03
# HELP node_disk_writes_completed_total Total successful writes.
# TYPE node_disk_writes_completed_total counter
node_disk_writes_completed_total{device="nvme0n1"} 2000
node_disk_writes_completed_total{device="nvme1n1"} 40
# HELP node_disk_written_bytes_total Total bytes written successfully.
# TYPE node_disk_written_bytes_total counter
node_disk_written_bytes_total{device="nvme0n1"} 4.096e+06
node_disk_written_bytes_total{device="nvme1n1"} 50
# HELP node_disk_write_time_seconds_total Total seconds spent on all writes.
# TYPE node_disk_write_time_seconds_total counter
node_disk_write_time_seconds_total{device="nvme0n1"} 0.7
node_disk_write_time_seconds_total{device="nvme1n1"} 0.06
# HELP node_disk_io_now I/Os currently in progress.
# TYPE node_disk_io_now gauge
node_disk_io_now{device="nvme0n1"} 1
node_disk_io_now{device="nvme1n1"} 0
# HELP node_disk_io_time_seconds_total Total seconds spent doing I/O.
# TYPE node_disk_io_time_seconds_total counter
node_disk_io_time_seconds_total{device="nvme0n1"} 1.2
node_disk_io_time_seconds_total{device="nvme1n1"} 0.07
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected),
		"node_disk_reads_completed_total", "node_disk_read_bytes_total", "node_disk_read_time_seconds_total",
		"node_disk_writes_completed_total", "node_disk_written_bytes_total", "node_disk_write_time_seconds_total",
		"node_disk_io_now", "node_disk_io_time_seconds_total"))
}

func TestCollectDiskFilter(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{DiskDevices: []string{"nvme1n1"}})
	require.Equal(t, 1, testutil.CollectAndCount(c, "node_disk_reads_completed_total"))
}

func TestCollectNetDev(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP node_network_receive_bytes_total Received bytes.
# TYPE node_network_receive_bytes_total counter
node_network_receive_bytes_total{device="eth0"} 1.23456789e+08
node_network_receive_bytes_total{device="lo"} 1
# HELP node_network_receive_packets_total Received packets.
# TYPE node_network_receive_packets_total counter
node_network_receive_packets_total{device="eth0"} 100000
node_network_receive_packets_total{device="lo"} 1
# HELP node_network_receive_errs_total Receive errors.
# TYPE node_network_receive_errs_total counter
node_network_receive_errs_total{device="eth0"} 3
node_network_receive_errs_total{device="lo"} 0
# HELP node_network_receive_drop_total Received dropped packets.
# TYPE node_network_receive_drop_total counter
node_network_receive_drop_total{device="eth0"} 7
node_network_receive_drop_total{device="lo"} 0
# HELP node_network_transmit_bytes_total Transmitted bytes.
# TYPE node_network_transmit_bytes_total counter
node_network_transmit_bytes_total{device="eth0"} 9.87654321e+08
node_network_transmit_bytes_total{device="lo"} 1
# HELP node_network_transmit_packets_total Transmitted packets.
# TYPE node_network_transmit_packets_total counter
node_network_transmit_packets_total{device="eth0"} 200000
node_network_transmit_packets_total{device="lo"} 1
# HELP node_network_transmit_errs_total Transmit errors.
# TYPE node_network_transmit_errs_total counter
node_network_transmit_errs_total{device="eth0"} 1
node_network_transmit_errs_total{device="lo"} 0
# HELP node_network_transmit_drop_total Transmit dropped packets.
# TYPE node_network_transmit_drop_total counter
node_network_transmit_drop_total{device="eth0"} 2
node_network_transmit_drop_total{device="lo"} 0
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected),
		"node_network_receive_bytes_total", "node_network_receive_packets_total",
		"node_network_receive_errs_total", "node_network_receive_drop_total",
		"node_network_transmit_bytes_total", "node_network_transmit_packets_total",
		"node_network_transmit_errs_total", "node_network_transmit_drop_total"))
}

func TestCollectNetDevFilter(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{NetDevices: []string{"lo"}})
	require.Equal(t, 1, testutil.CollectAndCount(c, "node_network_receive_bytes_total"))
}

func TestCollectThermal(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP node_thermal_zone_temp Thermal zone temperature in celsius.
# TYPE node_thermal_zone_temp gauge
node_thermal_zone_temp{type="acpitz",zone="1"} 38
node_thermal_zone_temp{type="cpu",zone="0"} 45
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected), "node_thermal_zone_temp"))
}

func TestCollectGPU(t *testing.T) {
	c, _ := testCollector(t, fixture(), Filters{})
	expected := `
# HELP DCGM_FI_DEV_GPU_TEMP GPU temperature (in C).
# TYPE DCGM_FI_DEV_GPU_TEMP gauge
DCGM_FI_DEV_GPU_TEMP{UUID="GPU-fake-0000",device="nvidia0",gpu="0",modelName="Fake GPU 0"} 55
DCGM_FI_DEV_GPU_TEMP{UUID="GPU-fake-0001",device="nvidia1",gpu="1",modelName="Fake GPU 1"} 61
# HELP DCGM_FI_DEV_POWER_USAGE GPU power usage (in W).
# TYPE DCGM_FI_DEV_POWER_USAGE gauge
DCGM_FI_DEV_POWER_USAGE{UUID="GPU-fake-0000",device="nvidia0",gpu="0",modelName="Fake GPU 0"} 120
DCGM_FI_DEV_POWER_USAGE{UUID="GPU-fake-0001",device="nvidia1",gpu="1",modelName="Fake GPU 1"} 230
# HELP DCGM_FI_DEV_SM_CLOCK SM clock frequency (in MHz).
# TYPE DCGM_FI_DEV_SM_CLOCK gauge
DCGM_FI_DEV_SM_CLOCK{UUID="GPU-fake-0000",device="nvidia0",gpu="0",modelName="Fake GPU 0"} 1200
DCGM_FI_DEV_SM_CLOCK{UUID="GPU-fake-0001",device="nvidia1",gpu="1",modelName="Fake GPU 1"} 2100
# HELP DCGM_FI_DEV_GPU_UTIL GPU utilization (in %).
# TYPE DCGM_FI_DEV_GPU_UTIL gauge
DCGM_FI_DEV_GPU_UTIL{UUID="GPU-fake-0000",device="nvidia0",gpu="0",modelName="Fake GPU 0"} 37
DCGM_FI_DEV_GPU_UTIL{UUID="GPU-fake-0001",device="nvidia1",gpu="1",modelName="Fake GPU 1"} 0
# HELP DCGM_FI_DEV_PSTATE PState of the GPU (0 to 15, 0 is maximum performance).
# TYPE DCGM_FI_DEV_PSTATE gauge
DCGM_FI_DEV_PSTATE{UUID="GPU-fake-0000",device="nvidia0",gpu="0",modelName="Fake GPU 0"} 2
DCGM_FI_DEV_PSTATE{UUID="GPU-fake-0001",device="nvidia1",gpu="1",modelName="Fake GPU 1"} 0
`
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected),
		"DCGM_FI_DEV_GPU_TEMP", "DCGM_FI_DEV_POWER_USAGE", "DCGM_FI_DEV_SM_CLOCK",
		"DCGM_FI_DEV_GPU_UTIL", "DCGM_FI_DEV_PSTATE"))
}

func TestCollectNothingPublished(t *testing.T) {
	c, _ := testCollector(t, nil, Filters{})
	require.NoError(t, testutil.GatherAndCompare(prometheus.NewRegistry(), strings.NewReader("")))
	count := testutil.CollectAndCount(c)
	require.Equal(t, 0, count)
}

func TestCollectSkipsInvalidCapabilities(t *testing.T) {
	s := fixture()
	s.Valid &^= probe.CapGPU | probe.CapThermal
	c, _ := testCollector(t, s, Filters{})
	require.Equal(t, 0, testutil.CollectAndCount(c, "DCGM_FI_DEV_GPU_TEMP"))
	require.Equal(t, 0, testutil.CollectAndCount(c, "node_thermal_zone_temp"))
	require.Equal(t, 16, testutil.CollectAndCount(c, "node_cpu_seconds_total"))
}

func TestFamilyCapDropsAndCounts(t *testing.T) {
	dropped := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pme_series_dropped_total",
	}, []string{"family"})
	fc := &familyCap{family: "test", remaining: 1, dropped: dropped}

	require.True(t, fc.allow())
	require.False(t, fc.allow())
	require.False(t, fc.allow())
	require.Equal(t, 2.0, testutil.ToFloat64(dropped.WithLabelValues("test")))
}
