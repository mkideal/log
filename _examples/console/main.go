package main

import (
	"github.com/mkideal/log"
)

func main() {
	defer log.Uninit(log.InitConsole(log.LvWARN))
	log.SetLevel(log.LvDEBUG)

	log.Trace("%s cannot be printed", "TRACE")

	log.Debug("%s should be printed into stdout", "DEBUG")
	log.Info("%s should be printed into stdout", "INFO")

	log.Warn("%s should be printed into stderr", "WARN")
	log.Error("%s should be printed into stderr", "ERROR")
	log.Fatal("%s should be printed into stderr", "FATAL")
}
