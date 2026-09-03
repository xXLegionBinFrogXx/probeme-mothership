package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	readHeaderTimeout = 10 * time.Second
	idleTimeout       = 60 * time.Second
	maxHeaderBytes    = 8 << 10
)

type Server struct {
	http.Server
}

func New(addr string, reg *prometheus.Registry, ready func() bool, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", getOrHeadHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !getOrHead(w, r) {
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		if !getOrHead(w, r) {
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if ready() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "ready")
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "not ready")
	})

	return &Server{http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}}
}

func getOrHeadHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getOrHead(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	return true
}
