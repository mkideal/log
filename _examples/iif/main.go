package main

import (
	"github.com/mkideal/log"
)

func main() {
	defer log.Uninit(log.InitConsole(log.LvERROR))

	log.If(true).Info("true")
	log.If(false).Info("false")
}
