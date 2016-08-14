package main

import (
	"github.com/mkideal/log"
)

func main() {
	// Init and defer Uninit
	defer log.Uninit(log.InitFile("./log/app.log"))

	// Default log level is log.LvINFO, you can change the level as following:
	//
	//	log.SetLevel(log.LvTRACE)
	// 	log.SetLevel(log.LvDEBUG)
	// 	log.SetLevel(log.LvINFO)
	// 	log.SetLevel(log.LvWARN)
	// 	log.SetLevel(log.LvERROR)
	// 	log.SetLevel(log.LvFATAL)

	log.Trace("%s cannot be printed", "TRACE")
	log.Debug("%s cannot be printed", "DEBUG")

	log.Info("%s should be printed", "INFO")
	log.Warn("%s should be printed", "WARN")
	log.Error("%s should be printed", "ERROR")

	log.If(true).Info("%v should be printed", true)

	x := 5
	log.If(x < 5).Info("x less than 5").
		ElseIf(x > 5).Info("x greater than 5").
		Else().Info("x equal to 5")

	log.Fatal("%s should be printed and exit program with status code 1", "FATAL")

	log.Info("You cannot see me")
}
