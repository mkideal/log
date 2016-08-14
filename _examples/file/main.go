package main

import (
	"github.com/mkideal/log"
)

func main() {
	defer log.Uninit(log.InitFile("./log/app.log"))
	// Default log level is log.LvINFO

	log.Trace("%s cannot be printed", "TRACE")
	log.Debug("%s cannot be printed", "DEBUG")

	log.Info("%s should be printed", "INFO")
	log.Warn("%s should be printed", "WARN")
	log.Error("%s should be printed", "ERROR")
	log.Fatal("%s should be printed", "FATAL")
}
