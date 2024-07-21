package http

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

func MustRun(
	ctx context.Context,
	shutdownDur time.Duration,
	addr string,
) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()

		log.Printf("Shutting down server with duration %0.3fs\n", shutdownDur.Seconds())
		<-time.After(shutdownDur)

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP handler Shutdown: %s\n", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Printf("HTTP server ListenAndServe: %s\n", err)
	}
}
