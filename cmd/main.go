package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/scraper"

	log "github.com/cjohnhelms/sentinel/pkg/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	log.Info("Service starting")
	log.Debug(fmt.Sprintf("Config: %+v", cfg))

	ctx, cancel := context.WithCancel(context.Background())

	wg := new(sync.WaitGroup)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	wg.Add(1)
	go scraper.Run(ctx, cfg, wg)

	// start metrics server
	server := &http.Server{
		Addr: ":2112",
	}
	http.Handle("/metrics", promhttp.Handler())
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error: %v", err)
		}
		log.Info("Metrics server stopped serving new connections.")
	}(wg)

	<-sig
	log.Info("Main routine recieved signal, waiting")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP shutdown error: %v", err)
	}

	cancel()
	wg.Wait()

	log.Info("All routines finished")
}
