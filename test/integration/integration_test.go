//go:build integration

// Boots the built binary with the fake provider and asserts /metrics,
// /health and /-/ready. Run via `make integration`.
package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const fakeSo = "../../test/fakeprovider/build/libprobeme_fake.so"

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestBinaryWithFakeProvider(t *testing.T) {
	bin := envOr("PME_INTEGRATION_BIN", "../../bin/probeme-mothership")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("binary %s not built; run make build", bin)
	}
	if _, err := os.Stat(fakeSo); err != nil {
		t.Skip("fake provider not built; run make fakeprovider")
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, bin,
		"--providers="+fakeSo,
		"--probe.interval=100ms",
		"--probe.timeout=80ms",
		"--web.listen-address="+addr,
		"--log.level=debug")
	cmd.Env = append(os.Environ(), "PME_FAKE_SLEEP_MS=0")
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	})

	base := fmt.Sprintf("http://%s", addr)

	require.Eventually(t, func() bool {
		resp, err := http.Get(base + "/-/ready")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 50*time.Millisecond)

	resp, err := http.Get(base + "/health")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "ok\n", string(body))

	resp, err = http.Post(base+"/metrics", "text/plain", strings.NewReader(""))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	resp, err = http.Get(base + "/metrics")
	require.NoError(t, err)
	metricsBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	metricsText := string(metricsBody)

	for _, want := range []string{
		"node_cpu_seconds_total{cpu=\"0\",mode=\"user\"} 10",
		"node_memory_MemTotal_bytes 3.4359738368e+10",
		"node_load1 0.5",
		"node_boot_time_seconds 1.7e+09",
		"node_disk_reads_completed_total{device=\"nvme0n1\"} 1000",
		"node_filesystem_size_bytes{device=\"/dev/nvme0n1p2\",fstype=\"ext4\",mountpoint=\"/\"} 5.00107862016e+11",
		"node_network_receive_bytes_total{device=\"eth0\"} 1.23456789e+08",
		"node_thermal_zone_temp{type=\"cpu\",zone=\"0\"} 45",
		"DCGM_FI_DEV_GPU_TEMP{UUID=\"GPU-fake-0000\",device=\"nvidia0\",gpu=\"0\",modelName=\"Fake GPU 0\"} 55",
		"DCGM_FI_DEV_POWER_USAGE{UUID=\"GPU-fake-0000\",device=\"nvidia0\",gpu=\"0\",modelName=\"Fake GPU 0\"} 120",
		"pme_build_info{",
		"pme_probe_age_seconds ",
		"pme_provider_duration_seconds{provider=\"fake\"}",
	} {
		require.Contains(t, metricsText, want, "missing series in /metrics")
	}

	require.NotContains(t, metricsText, "mountpoint=\"/proc\"")

	// promtool validates the exposition when available. Exit code 3 means
	// lint warnings only (node_exporter-style camelCase names trip the same
	// linter), which is acceptable — compatibility wins over style.
	if promtool, lookErr := exec.LookPath("promtool"); lookErr == nil {
		cmd := exec.Command(promtool, "check", "metrics")
		cmd.Stdin = strings.NewReader(metricsText)
		out, err := cmd.CombinedOutput()
		if err != nil {
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr)
			require.Equal(t, 3, exitErr.ExitCode(),
				"promtool failed (not just lint):\n%s", out)
			t.Logf("promtool lint warnings (expected):\n%s", out)
		}
	} else {
		t.Log("promtool not in PATH; skipping exposition validation")
	}
}

func TestBinaryZeroProvidersReadyStays503(t *testing.T) {
	bin := envOr("PME_INTEGRATION_BIN", "../../bin/probeme-mothership")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("binary %s not built; run make build", bin)
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, bin,
		"--provider.dir="+t.TempDir(), // empty dir → zero providers
		"--web.listen-address="+addr)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	})

	base := fmt.Sprintf("http://%s", addr)

	require.Eventually(t, func() bool {
		resp, err := http.Get(base + "/metrics")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode == http.StatusOK && strings.Contains(string(body), "pme_build_info")
	}, 10*time.Second, 50*time.Millisecond)

	resp, err := http.Get(base + "/-/ready")
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}
