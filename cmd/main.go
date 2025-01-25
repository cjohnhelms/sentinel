package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/display"
	"github.com/cjohnhelms/sentinel/pkg/notify"
	"github.com/cjohnhelms/sentinel/pkg/scraper"

	log "github.com/cjohnhelms/sentinel/pkg/logging"
)

func main() {
	cfg := config.New()

	log.Info("Service starting")
	log.Debug(fmt.Sprintf("Config: %+v", cfg))

	wg := new(sync.WaitGroup)
	wg.Add(3)

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	data := make(chan scraper.Event, 1)
	go scraper.FetchEvents(ctx, wg, data)
	go display.Update(ctx, wg, data)
	go notify.Notify(ctx, wg, data, cfg)

	<-sig
	log.Error("Cancel recieved, killing routines")
	cancel()

	wg.Wait()
	log.Info("All routines finished")
}
