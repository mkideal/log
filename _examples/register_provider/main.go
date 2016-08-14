package main

import (
	"os"

	"github.com/mkideal/log"
	"github.com/mkideal/log/logger"
)

type myProvider struct {
}

func newMyProvider(opts string) logger.Provider {
	return &myProvider{}
}

func (p *myProvider) Write(level logger.Level, headerLength int, data []byte) error {
	// ignore the header
	_, err := os.Stdout.Write(data[headerLength:])
	return err
}

func (p *myProvider) Close() error { return nil }

func init() {
	logger.Register("myProvider", newMyProvider)
}

func main() {
	defer log.Uninit(log.Init("myProvider", ""))
	log.SetLevel(log.LvTRACE)
	log.Trace("hello my registered Provider")
}
