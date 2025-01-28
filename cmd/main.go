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

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type loggerKey struct{}

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	logger := cfg.Logger

	logger.Info("Service starting")
	logger.Debug(fmt.Sprintf("Config: %+v", cfg))

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
		logger.Info("Starting metrics server")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error: %v", err.Error())
		}
		logger.Info("Metrics server stopped serving new connections.")
	}(wg)

	<-sig
	logger.Info("Main routine recieved signal, waiting")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown error: %v", err)
	}

	cancel()
	wg.Wait()

	logger.Info("All routines finished")
}
