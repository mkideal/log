package main

import (
	"encoding/json"

	"github.com/mkideal/log"
	"github.com/mkideal/log/provider"
)

func main() {
	// Init and defer Uninit
	//defer log.Uninit(log.InitMultiFile("./log", "app.log"))
	config := provider.NewMultiFileOpts()
	config.DebugDir = config.InfoDir
	config.RootDir = "log"
	b, _ := json.Marshal(config)
	defer log.Uninit(log.Init("multifile", string(b)))
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
