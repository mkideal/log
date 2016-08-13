package main

import (
	"flag"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/mkideal/log"
)

var (
	flInterval = flag.Int("i", 0, "interval of writting logs")
)

func main() {
	defer log.Uninit(log.InitFile("./log/app.log"))
	log.SetLevel(log.TRACE)

	flag.Parse()
	d := time.Duration(*flInterval) * time.Second

	running := int32(1)
	go func() {
		for atomic.LoadInt32(&running) != 0 {
			log.Trace("hello %s", "Trace")
			log.Debug("hello %s", "Debug")
			log.Info("hello %s", "Info")
			log.Warn("hello %s", "Warn")
			log.Error("hello %s", "Error")
			if d > 0 {
				time.Sleep(d)
			}
		}
	}()

	quit := make(chan struct{})
	ListenSignal(func(sig os.Signal) bool {
		atomic.StoreInt32(&running, 0)
		quit <- struct{}{}
		return true
	})
	<-quit
	log.Info("app exit")
}

func ListenSignal(handler func(sig os.Signal) (ret bool), signals ...os.Signal) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	go func() {
		for !handler(<-sigChan) {
		}
	}()
}
