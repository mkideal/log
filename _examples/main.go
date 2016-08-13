package main

import (
	"os"
	"os/signal"
	"sync/atomic"

	"github.com/mkideal/log"
)

func main() {
	defer log.Uninit(log.InitFile("./log/app.log"))
	log.SetLevel(log.TRACE)

	running := int32(1)
	go func() {
		for atomic.LoadInt32(&running) != 0 {
			log.Trace("hello %s", "Trace")
			log.Debug("hello %s", "Debug")
			log.Info("hello %s", "Info")
			log.Warn("hello %s", "Warn")
			log.Error("hello %s", "Error")
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
