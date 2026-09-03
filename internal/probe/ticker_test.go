//go:build cgo

package probe

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/provider"
)

const fakeSo = "../../test/fakeprovider/build/libprobeme_fake.so"

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newFakeProvider(t *testing.T) *provider.Loaded {
	t.Helper()
	if _, err := os.Stat(fakeSo); err != nil {
		t.Skip("fake provider not built; run make fakeprovider")
	}
	p, err := provider.Open(fakeSo)
	require.NoError(t, err)
	t.Cleanup(p.Close)
	require.NoError(t, p.Init(0))
	return &provider.Loaded{Provider: p, Caps: p.Capabilities()}
}

func newTestTicker(t *testing.T, loaded []*provider.Loaded, interval, timeout time.Duration) (*Ticker, *Published, *SelfMetrics) {
	t.Helper()
	pub := NewPublished()
	self := NewSelfMetrics(pub)
	tk, err := NewTicker(loaded, pub, self, interval, timeout, testLogger(t))
	require.NoError(t, err)
	return tk, pub, self
}

func TestTickerPublishesSnapshot(t *testing.T) {
	l := newFakeProvider(t)
	tk, pub, _ := newTestTicker(t, []*provider.Loaded{l}, time.Hour, time.Second)

	tk.tickOnce()

	s := pub.Load()
	require.NotNil(t, s)
	require.Equal(t, uint64(1), s.Generation)
	require.WithinDuration(t, time.Now(), s.PublishedAt, time.Second)
	require.True(t, pub.Ready())

	tk.tickOnce()
	require.Equal(t, uint64(2), pub.Load().Generation)
}

func TestTickerTimeoutPath(t *testing.T) {
	t.Setenv("PME_FAKE_SLEEP_MS", "50")
	l := newFakeProvider(t)
	tk, pub, self := newTestTicker(t, []*provider.Loaded{l}, time.Hour, 10*time.Millisecond)

	base := runtime.NumGoroutine()
	tk.tickOnce() // returns at the 10ms timeout, collect still running

	require.True(t, tk.inFlight.Load())
	require.Equal(t, 1.0, testutil.ToFloat64(self.Timeouts))
	require.Nil(t, pub.Load(), "timed-out tick must not publish")

	require.Eventually(t, func() bool { return !tk.inFlight.Load() }, 2*time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool { return runtime.NumGoroutine() <= base+1 }, 2*time.Second, 5*time.Millisecond)

	require.NoError(t, os.Unsetenv("PME_FAKE_SLEEP_MS"))
	require.NoError(t, l.Provider.Init(0))
	tk.tickOnce()
	require.NotNil(t, pub.Load())
}

func TestTickerSkippedTick(t *testing.T) {
	l := newFakeProvider(t)
	tk, _, self := newTestTicker(t, []*provider.Loaded{l}, time.Hour, time.Second)

	tk.inFlight.Store(true)
	tk.tickOnce()

	require.Equal(t, 1.0, testutil.ToFloat64(self.Skipped))
	require.True(t, tk.inFlight.Load(), "skip must not clear the in-flight guard")
	tk.inFlight.Store(false)
}

func TestTickerRunStopsOnContextCancel(t *testing.T) {
	l := newFakeProvider(t)
	tk, _, _ := newTestTicker(t, []*provider.Loaded{l}, 5*time.Millisecond, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		tk.Run(ctx)
	}()

	require.Eventually(t, func() bool { return tk.generation.Load() >= 1 }, 2*time.Second, 5*time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}
