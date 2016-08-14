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
	log.SetLevel(log.LvTRACE)

	flag.Parse()
	d := time.Duration(*flInterval) * time.Millisecond

	running := int32(1)
	over := make(chan bool)
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
		over <- true
	}()

	listenSignal(func(sig os.Signal) bool {
		atomic.StoreInt32(&running, 0)
		return true
	})
	<-over
	log.Info("app exit")
}

func listenSignal(handler func(sig os.Signal) (ret bool), signals ...os.Signal) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	go func() {
		for !handler(<-sigChan) {
		}
	}()
}
