package provider

import (
	"github.com/mkideal/log/logger"
)

type mixProvider struct {
	providers []logger.Provider
}

// NewMixProvider creates a mixProvider
func NewMixProvider(first logger.Provider, others ...logger.Provider) logger.Provider {
	p := new(mixProvider)
	p.providers = make([]logger.Provider, 0, len(others)+1)
	p.providers = append(p.providers, first)
	for _, other := range others {
		p.providers = append(p.providers, other)
	}
	return p
}

// Write writes log to all inner providers
func (p *mixProvider) Write(level logger.Level, headerLength int, data []byte) error {
	var err errorList
	for _, op := range p.providers {
		err.tryPush(op.Write(level, headerLength, data))
	}
	return err.err()
}

// Close close all inner providers
func (p *mixProvider) Close() error {
	var err errorList
	for _, op := range p.providers {
		err.tryPush(op.Close())
	}
	return err.err()
}
