# probeme-mothership

<p align="center">
  <img src="assets/logo.png" alt="Mothership — probe me" width="320">
</p>

One consumer of [`libprobeme`](https://github.com/xXLegionBinFrogXx/libprobeme):
loads `libprobeme_*.so` providers, ticks `collect_all` and exposes
`node_*` / `DCGM_FI_DEV_*` / `pme_*` series on `/metrics`.

## Metrics

| Family | Series | Conversion |
|---|---|---|
| `node_cpu_seconds_total` | labels `cpu`, `mode` | jiffies ÷ `clk_tck` |
| `node_memory_*_bytes` | MemTotal, MemFree, MemAvailable, Buffers, Cached, SwapTotal, SwapFree | as-is |
| `node_load1/5/15`, `node_boot_time_seconds` | — | as-is |
| `node_disk_*` | reads/writes completed + bytes + `time_seconds`, `io_now`, `io_time_seconds` | ms ÷ 1000 |
| `node_filesystem_*` | size/free/avail `_bytes`, files, files_free, readonly | skipped mounts excluded |
| `node_network_{receive,transmit}_*_total` | bytes, packets, errs, drop | as-is |
| `node_thermal_zone_temp` | labels `zone`, `type` | m°C ÷ 1000 |
| `DCGM_FI_DEV_*` | GPU_TEMP, POWER_USAGE, SM_CLOCK, GPU_UTIL, PSTATE; labels `gpu`, `UUID`, `device`, `modelName` | mW ÷ 1000 |
| `pme_*` | build_info, probe_age/skipped/timeouts, provider_duration/errors, series_dropped | self-monitoring |

- `node_*` is scrape-compatible with node_exporter (Grafana dashboard 1860);
  `DCGM_FI_DEV_*` with dcgm-exporter (dashboard 12239). FB/VRAM fields are never exported.
- Filters are Go-side prefix excludes (`--collector.filesystem.mountpoints-exclude`,
  `--collector.filesystem.fstypes-exclude`, `--collector.netdev.devices-exclude`,
  `--collector.disk.devices-exclude`, env `PME_*` equivalents).
- A 256-entry per-family cap guards cardinality; drops bump `pme_series_dropped_total{family}`.

## Design notes

- `internal/provider/cgo.go` is the only file that imports "C"; everything above sees plain Go structs. The libprobeme ABI (major 1) is frozen.
- Providers are `dlopen`ed at startup (`--providers=a.so,b.so`, else glob `${PME_PROVIDER_DIR:-/usr/lib/probeme}/libprobeme_*.so`); first provider to claim a capability wins.
- A 5s timer tick calls every provider into one C buffer (`--probe.timeout=3s`); ticks never overlap — a stalled call is abandoned and ticks skip until it returns. The result is copied into a Go snapshot and published via `atomic.Pointer` (latest wins).
- `/metrics` serves the last published snapshot as const metrics — no collect-on-scrape, no sample timestamps. `/-/ready` returns 503 until the first successful publish.
- Exit codes: 0 clean, 1 config/startup error, 2 no provider with `--require-provider`.
- Unit conversions live only in `internal/metrics`; metric name strings only in `internal/metrics` and `internal/probe/selfmetrics.go`.

## Requirements

- Linux x86_64 or aarch64, Go 1.23+, cgo toolchain
- `libprobeme` v1.0.0 installed (`cmake --install`); `pkg-config --cflags --libs probeme` must succeed

## Quick start

```
make build          # builds the fake test provider too
./bin/probeme-mothership --providers=/usr/lib/probeme/libprobeme_linux.so
curl -s localhost:9167/metrics | head
```

## Install

```
sudo make install               # /usr/local/bin + systemd unit (/usr/local/lib/systemd/system)
sudo systemctl daemon-reload
sudo systemctl enable --now probeme-mothership
```

`make install PREFIX=/usr/local DESTDIR=/tmp/pkg` supports staged installs;
`sudo make uninstall` removes both files.

## Development

```
make test           # go test -race ./... + CGO_ENABLED=0 build check
make integration    # boots the binary with the fake provider, promtool check
make run PROVIDERS=/usr/lib/probeme/libprobeme_linux.so
```
