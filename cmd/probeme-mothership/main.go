// Command probeme-mothership exposes node_* / DCGM_FI_DEV_* / pme_* on /metrics.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/buildinfo"
	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/config"
	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/metrics"
	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/provider"
	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/server"
)

func main() { os.Exit(run()) }

func run() int {
	cfg, showVersion, err := config.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		return 1
	}
	if showVersion {
		fmt.Println(buildinfo.String())
		return 0
	}

	logger := newLogger(cfg.LogLevel)
	logger.Info("starting probeme-mothership",
		"version", buildinfo.Version,
		"commit", buildinfo.Commit,
		"providers", strings.Join(cfg.Providers, ","),
		"provider_dir", cfg.ProviderDir,
		"probe.interval", cfg.ProbeInterval,
		"probe.timeout", cfg.ProbeTimeout,
		"web.listen-address", cfg.ListenAddress)

	loaded, err := provider.Load(provider.Options{
		Explicit: cfg.Providers,
		Dir:      cfg.ProviderDir,
		Logger:   logger,
	})
	if err != nil {
		logger.Error("provider load failed", "err", err)
		return 1
	}
	defer func() {
		for _, l := range loaded {
			l.Provider.Destroy()
			l.Provider.Close()
		}
	}()

	if len(loaded) == 0 && cfg.RequireProvider {
		logger.Error("no usable provider and --require-provider set")
		return 2
	}
	if len(loaded) == 0 {
		logger.Warn("running without providers; only pme_* self metrics will be exported")
	}

	pub := probe.NewPublished()
	self := probe.NewSelfMetrics(pub)
	ticker, err := probe.NewTicker(loaded, pub, self, cfg.ProbeInterval, cfg.ProbeTimeout, logger)
	if err != nil {
		logger.Error("probe ticker init failed", "err", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	tickerDone := make(chan struct{})
	go func() {
		defer close(tickerDone)
		ticker.Run(ctx)
	}()

	reg := prometheus.NewRegistry()
	reg.MustRegister(self.Collectors()...)
	reg.MustRegister(metrics.NewCollector(pub, metrics.NewFilters(
		cfg.Filters.MountPointsExclude,
		cfg.Filters.FSTypesExclude,
		cfg.Filters.NetDevicesExclude,
		cfg.Filters.DiskDevicesExclude,
	), self.Dropped))

	srv := server.New(cfg.ListenAddress, reg, pub.Ready, logger)
	srvErr := make(chan error, 1)
	go func() { srvErr <- srv.ListenAndServe() }()
	logger.Info("listening", "addr", cfg.ListenAddress)

	select {
	case err := <-srvErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "err", err)
			return 1
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("http shutdown", "err", err)
	}
	stop()
	<-tickerDone
	logger.Info("stopped")
	return 0
}

func newLogger(level string) *slog.Logger {
	var lv slog.Level
	switch level {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lv}))
}
