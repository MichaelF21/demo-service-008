package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serviceName = "demo-service-008"

var requestCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "demo_service_008_http_requests_total",
		Help: "Total HTTP requests handled by " + serviceName,
	},
	[]string{"path", "status"},
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"service": serviceName,
		"message": fmt.Sprintf("Hello from %s!", serviceName),
	})
	requestCounter.WithLabelValues("/", "200").Inc()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
	requestCounter.WithLabelValues("/healthz", "200").Inc()
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ready")
	requestCounter.WithLabelValues("/readyz", "200").Inc()
}

// newMux wires the HTTP routes. Exposed so tests can exercise the same router
// the binary serves at runtime.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readyHandler)
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newMux(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		logger.Info("server starting", "service", serviceName, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown initiated")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("shutdown complete")
}
