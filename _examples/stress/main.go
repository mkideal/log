package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
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

	var g sync.WaitGroup

	task := func() {
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
		g.Done()
	}

	n := 10
	g.Add(n)
	for i := 0; i < n; i++ {
		go task()
	}

	listenSignal(func(sig os.Signal) bool {
		atomic.StoreInt32(&running, 0)
		return true
	})

	g.Wait()
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
