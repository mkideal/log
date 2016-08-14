package main

import (
	"github.com/mkideal/log"
)

func main() {
	// Init and defer Uninit
	defer log.Uninit(log.InitMultiFile("./log", "app.log"))
	log.SetLevel(log.LvDEBUG)

	log.Trace("%s cannot be printed, and trace subdirectory not created", "TRACE")
	log.Debug("%s should be printed", "DEBUG")
	log.Info("%s should be printed", "INFO")
	log.Warn("%s should be printed", "WARN")
	log.Error("%s should be printed", "ERROR")

	// log again
	log.Trace("%s cannot be printed, and trace subdirectory not created", "TRACE")
	log.Debug("%s should be printed", "DEBUG")
	log.Info("%s should be printed", "INFO")
	log.Warn("%s should be printed", "WARN")
	log.Error("%s should be printed", "ERROR")

	log.Fatal("%s should be printed and exit program with status code 1", "FATAL")

	log.Info("You cannot see me")
}
