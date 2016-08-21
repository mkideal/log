package main

import (
	"github.com/mkideal/log"
	"github.com/mkideal/log/provider"
)

func main() {
	log.InitSyncWithProvider(provider.NewConsole(""))

	// (FIXME): try replace first line with following:
	//	log.InitWithProvider(provider.NewConsole(""))
	//
	// (FIXME): try replace first line with following:
	//	defer log.Uninit(log.InitWithProvider(provider.NewConsole("")))

	log.SetLevel(log.LvTRACE)

	log.Trace("sync logging %v message", log.LvTRACE)
	log.Debug("sync logging %v message", log.LvDEBUG)
	log.Info("sync logging %v message", log.LvINFO)
	log.Warn("sync logging %v message", log.LvWARN)
	log.Error("sync logging %v message", log.LvERROR)
}
