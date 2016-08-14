package main

import (
	"fmt"
	"github.com/mkideal/log"
	"github.com/mkideal/log/provider"
)

func main() {
	opts := fmt.Sprintf(`{
		"tostderrlevel": %d,
		"dir": "./log",
		"filename": "app.log",
		"nosymlink": true
	}`, log.LvERROR)
	p := provider.NewMixProvider(provider.NewConsole(opts), provider.NewFile(opts))
	defer log.Uninit(log.InitWithProvider(p))

	// NOTE: The above equivalent to following:
	// defer log.Uninit(log.InitFileAndConsole(log.LvERROR, "./log/app.log"))

	log.SetLevel(log.LvTRACE)

	log.Trace("hello %s", "TRACE")
	log.Debug("hello %s", "DEBUG")
	log.Info("hello %s", "INFO")
	log.Warn("hello %s", "WARN")
	log.Error("hello %s", "ERROR")
}
