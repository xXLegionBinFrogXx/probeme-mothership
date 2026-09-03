package probe

import (
	"context"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/provider"
)

// Ticker publishes a snapshot per tick. Ticks never overlap: a timed-out
// tick leaves the stalled call running and skips ticks until it returns.
type Ticker struct {
	providers []*provider.Loaded
	buf       *provider.CBuffer
	pub       *Published
	self      *SelfMetrics
	interval  time.Duration
	timeout   time.Duration
	logger    *slog.Logger

	inFlight   atomic.Bool
	generation atomic.Uint64
}

func NewTicker(loaded []*provider.Loaded, pub *Published, self *SelfMetrics,
	interval, timeout time.Duration, logger *slog.Logger,
) (*Ticker, error) {
	buf, err := provider.NewCBuffer()
	if err != nil {
		return nil, err
	}
	return &Ticker{
		providers: loaded,
		buf:       buf,
		pub:       pub,
		self:      self,
		interval:  interval,
		timeout:   timeout,
		logger:    logger,
	}, nil
}

func (t *Ticker) Run(ctx context.Context) {
	tk := time.NewTicker(t.interval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			t.tickOnce()
		}
	}
}

func (t *Ticker) tickOnce() {
	if !t.inFlight.CompareAndSwap(false, true) {
		t.self.Skipped.Inc()
		return
	}

	done := make(chan bool, 1)
	go func() {
		// Some drivers (NVML) are thread-affine.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		t.buf.Reset()
		for _, l := range t.providers {
			start := time.Now()
			err := l.Provider.CollectAll(t.buf)
			t.self.Duration.WithLabelValues(l.Provider.Name()).Set(time.Since(start).Seconds())
			if err != nil {
				t.self.Errors.WithLabelValues(l.Provider.Name()).Inc()
				t.logger.Warn("provider collect failed", "provider", l.Provider.Name(), "err", err)
				continue
			}
		}
		done <- true
	}()

	timer := time.NewTimer(t.timeout)
	defer timer.Stop()
	select {
	case <-done:
		t.publish()
		t.inFlight.Store(false)
	case <-timer.C:
		t.self.Timeouts.Inc()
		t.logger.Warn("probe tick timed out; collect still running in background", "timeout", t.timeout)
		go func() {
			<-done
			t.inFlight.Store(false)
		}()
	}
}

func (t *Ticker) publish() {
	snap := t.buf.Snapshot()
	snap.Generation = t.generation.Add(1)
	snap.PublishedAt = time.Now()
	t.pub.Store(snap)
}
