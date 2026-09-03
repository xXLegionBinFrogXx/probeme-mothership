/* libprobeme_fake — canned provider for tests.
 *
 * Env knobs (read at init()/pme_provider_get() time):
 *   PME_FAKE_CAPABILITIES  hex mask override; default all PME_CAP_*
 *   PME_FAKE_ABI_MAJOR     override reported ABI major (mismatch tests)
 *   PME_FAKE_ENOTSUP=1     collect_all returns PME_ENOTSUP
 *   PME_FAKE_EIO=1         collect_all returns PME_EIO
 *   PME_FAKE_TRUNCATE=1    set all truncated bits
 *   PME_FAKE_SLEEP_MS=N    collect_all sleeps N ms (timeout tests)
 *   PME_FAKE_CLK_TCK=N     clk_tck value (default 100)
 */

#include <probeme.h>

#include <stdlib.h>
#include <string.h>
#include <time.h>

static struct pme_provider fake_provider = {
    .size = sizeof(struct pme_provider),
    .abi_version = (PME_ABI_MAJOR << 16) | PME_ABI_MINOR,
    .name = "fake",
    .capabilities = PME_CAP_CPU | PME_CAP_MEMORY | PME_CAP_LOADAVG |
                    PME_CAP_UPTIME | PME_CAP_DISK_IO | PME_CAP_FILESYSTEM |
                    PME_CAP_NETDEV | PME_CAP_THERMAL | PME_CAP_GPU,
};

static unsigned int enotsup_flag;
static unsigned int eio_flag;
static unsigned int truncate_flag;
static long sleep_ms;
static unsigned int clk_tck = 100;

static uint64_t now_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ull + (uint64_t)ts.tv_nsec;
}

static void fill(struct pme_snapshot *s) {
    const uint64_t caps = fake_provider.capabilities;

    s->valid = 0;
    s->truncated = 0;

    if (caps & PME_CAP_CPU) {
        s->valid |= PME_CAP_CPU;
        s->cpu.n = 4;
        s->cpu.clk_tck = clk_tck;
        s->cpu.read_at_ns = now_ns();
        for (uint32_t i = 0; i < 4; i++) {
            s->cpu.cpu[i].user    = 1000 + 100 * i;
            s->cpu.cpu[i].nice    = 200 + 100 * i;
            s->cpu.cpu[i].system  = 300 + 100 * i;
            s->cpu.cpu[i].idle    = 400 + 100 * i;
            s->cpu.cpu[i].iowait  = 500 + 100 * i;
            s->cpu.cpu[i].irq     = 600 + 100 * i;
            s->cpu.cpu[i].softirq = 700 + 100 * i;
            s->cpu.cpu[i].steal   = 800 + 100 * i;
        }
    }

    if (caps & PME_CAP_MEMORY) {
        s->valid |= PME_CAP_MEMORY;
        s->memory.read_at_ns = now_ns();
        s->memory.total      = 34359738368ull; /* 32 GiB */
        s->memory.free       = 17179869184ull; /* 16 GiB */
        s->memory.available  = 21474836480ull; /* 20 GiB */
        s->memory.buffers    = 1073741824ull;  /*  1 GiB */
        s->memory.cached     = 4294967296ull;  /*  4 GiB */
        s->memory.swap_total = 2147483648ull;  /*  2 GiB */
        s->memory.swap_free  = 1073741824ull;  /*  1 GiB */
    }

    if (caps & PME_CAP_LOADAVG) {
        s->valid |= PME_CAP_LOADAVG;
        s->loadavg.read_at_ns = now_ns();
        s->loadavg.load1 = 0.5;
        s->loadavg.load5 = 0.25;
        s->loadavg.load15 = 0.125;
        s->loadavg.running = 2;
        s->loadavg.total = 16;
    }

    if (caps & PME_CAP_UPTIME) {
        s->valid |= PME_CAP_UPTIME;
        s->uptime.read_at_ns = now_ns();
        s->uptime.uptime_s = 86400;
        s->uptime.boot_time_unix_s = 1700000000;
    }

    if (caps & PME_CAP_DISK_IO) {
        s->valid |= PME_CAP_DISK_IO;
        s->disk_io.n = 2;
        s->disk_io.read_at_ns = now_ns();
        strcpy(s->disk_io.disks[0].name, "nvme0n1");
        s->disk_io.disks[0].reads = 1000;
        s->disk_io.disks[0].read_bytes = 2048000;
        s->disk_io.disks[0].read_time_ms = 500;
        s->disk_io.disks[0].writes = 2000;
        s->disk_io.disks[0].write_bytes = 4096000;
        s->disk_io.disks[0].write_time_ms = 700;
        s->disk_io.disks[0].io_in_progress = 1;
        s->disk_io.disks[0].io_time_ms = 1200;
        strcpy(s->disk_io.disks[1].name, "nvme1n1");
        s->disk_io.disks[1].reads = 10;
        s->disk_io.disks[1].read_bytes = 20;
        s->disk_io.disks[1].read_time_ms = 30;
        s->disk_io.disks[1].writes = 40;
        s->disk_io.disks[1].write_bytes = 50;
        s->disk_io.disks[1].write_time_ms = 60;
        s->disk_io.disks[1].io_in_progress = 0;
        s->disk_io.disks[1].io_time_ms = 70;
    }

    if (caps & PME_CAP_FILESYSTEM) {
        s->valid |= PME_CAP_FILESYSTEM;
        s->filesystem.n = 3;
        s->filesystem.read_at_ns = now_ns();
        strcpy(s->filesystem.mounts[0].device, "/dev/nvme0n1p2");
        strcpy(s->filesystem.mounts[0].mountpoint, "/");
        strcpy(s->filesystem.mounts[0].fstype, "ext4");
        s->filesystem.mounts[0].size_bytes = 500107862016ull;
        s->filesystem.mounts[0].free_bytes = 250053931008ull;
        s->filesystem.mounts[0].avail_bytes = 200000000000ull;
        s->filesystem.mounts[0].files = 30541877;
        s->filesystem.mounts[0].files_free = 15432807;
        strcpy(s->filesystem.mounts[1].device, "/dev/nvme0n1p1");
        strcpy(s->filesystem.mounts[1].mountpoint, "/boot");
        strcpy(s->filesystem.mounts[1].fstype, "vfat");
        s->filesystem.mounts[1].size_bytes = 536870912;
        s->filesystem.mounts[1].free_bytes = 268435456;
        s->filesystem.mounts[1].avail_bytes = 268435456;
        strcpy(s->filesystem.mounts[2].device, "proc");
        strcpy(s->filesystem.mounts[2].mountpoint, "/proc");
        strcpy(s->filesystem.mounts[2].fstype, "proc");
        s->filesystem.mounts[2].flags = PME_MOUNT_SKIPPED;
    }

    if (caps & PME_CAP_NETDEV) {
        s->valid |= PME_CAP_NETDEV;
        s->netdev.n = 2;
        s->netdev.read_at_ns = now_ns();
        strcpy(s->netdev.ifaces[0].name, "eth0");
        s->netdev.ifaces[0].rx_bytes = 123456789;
        s->netdev.ifaces[0].rx_packets = 100000;
        s->netdev.ifaces[0].rx_errs = 3;
        s->netdev.ifaces[0].rx_drop = 7;
        s->netdev.ifaces[0].tx_bytes = 987654321;
        s->netdev.ifaces[0].tx_packets = 200000;
        s->netdev.ifaces[0].tx_errs = 1;
        s->netdev.ifaces[0].tx_drop = 2;
        strcpy(s->netdev.ifaces[1].name, "lo");
        s->netdev.ifaces[1].rx_bytes = 1;
        s->netdev.ifaces[1].rx_packets = 1;
        s->netdev.ifaces[1].tx_bytes = 1;
        s->netdev.ifaces[1].tx_packets = 1;
    }

    if (caps & PME_CAP_THERMAL) {
        s->valid |= PME_CAP_THERMAL;
        s->thermal.n = 2;
        s->thermal.read_at_ns = now_ns();
        strcpy(s->thermal.zones[0].type, "cpu");
        s->thermal.zones[0].temp_mc = 45000;
        strcpy(s->thermal.zones[1].type, "acpitz");
        s->thermal.zones[1].temp_mc = 38000;
    }

    if (caps & PME_CAP_GPU) {
        s->valid |= PME_CAP_GPU;
        s->gpu.n = 2;
        s->gpu.read_at_ns = now_ns();
        strcpy(s->gpu.gpus[0].uuid, "GPU-fake-0000");
        strcpy(s->gpu.gpus[0].name, "Fake GPU 0");
        s->gpu.gpus[0].temp_c = 55;
        s->gpu.gpus[0].power_mw = 120000;
        s->gpu.gpus[0].sm_clock_mhz = 1200;
        s->gpu.gpus[0].util_pct = 37;
        s->gpu.gpus[0].pstate = 2;
        strcpy(s->gpu.gpus[1].uuid, "GPU-fake-0001");
        strcpy(s->gpu.gpus[1].name, "Fake GPU 1");
        s->gpu.gpus[1].temp_c = 61;
        s->gpu.gpus[1].power_mw = 230000;
        s->gpu.gpus[1].sm_clock_mhz = 2100;
        s->gpu.gpus[1].util_pct = 0;
        s->gpu.gpus[1].pstate = 0;
    }

    if (truncate_flag) {
        s->truncated = s->valid;
    }
}

static int init_fn(const struct pme_config *cfg) {
    (void)cfg;
    enotsup_flag = getenv("PME_FAKE_ENOTSUP") ? 1u : 0u;
    eio_flag = getenv("PME_FAKE_EIO") ? 1u : 0u;
    truncate_flag = getenv("PME_FAKE_TRUNCATE") ? 1u : 0u;
    const char *sleep = getenv("PME_FAKE_SLEEP_MS");
    sleep_ms = sleep ? strtol(sleep, NULL, 10) : 0;
    const char *hz = getenv("PME_FAKE_CLK_TCK");
    if (hz) {
        clk_tck = (unsigned int)strtoul(hz, NULL, 10);
    }
    return PME_OK;
}

static int collect_all_fn(struct pme_snapshot *s) {
    if (enotsup_flag) {
        return PME_ENOTSUP;
    }
    if (eio_flag) {
        return PME_EIO;
    }
    if (sleep_ms > 0) {
        struct timespec req = {
            .tv_sec = sleep_ms / 1000,
            .tv_nsec = (sleep_ms % 1000) * 1000000L,
        };
        nanosleep(&req, NULL);
    }
    fill(s);
    return PME_OK;
}

static void destroy_fn(void) {}

const struct pme_provider *pme_provider_get(void) {
    fake_provider.init = init_fn;
    fake_provider.collect_all = collect_all_fn;
    fake_provider.destroy = destroy_fn;

    const char *abi = getenv("PME_FAKE_ABI_MAJOR");
    if (abi) {
        fake_provider.abi_version =
            ((uint32_t)strtoul(abi, NULL, 10) << 16) | PME_ABI_MINOR;
    } else {
        fake_provider.abi_version = (PME_ABI_MAJOR << 16) | PME_ABI_MINOR;
    }
    const char *caps_env = getenv("PME_FAKE_CAPABILITIES");
    if (caps_env) {
        fake_provider.capabilities = strtoull(caps_env, NULL, 0);
    }
    return &fake_provider;
}
