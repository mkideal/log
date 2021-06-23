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

func main() {
	defer log.Uninit(log.InitWithProvider(newMyProvider("")))
	log.SetLevel(log.LvTRACE)
	log.Trace("hello myProvider")
}
