package provider

import (
	"bytes"
	"errors"

	"github.com/mkideal/log/logger"
)

type mixProvider struct {
	providers []logger.Provider
}

func NewMixProvider(first logger.Provider, others ...logger.Provider) logger.Provider {
	p := new(mixProvider)
	p.providers = make([]logger.Provider, 0, len(others)+1)
	p.providers = append(p.providers, first)
	for _, other := range others {
		p.providers = append(p.providers, other)
	}
	return p
}

func (p *mixProvider) Write(level logger.Level, headerLength int, data []byte) error {
	var err errorList
	for _, op := range p.providers {
		err.tryPush(op.Write(level, headerLength, data))
	}
	return err.Err()
}

func (p *mixProvider) Close() {
	for _, op := range p.providers {
		op.Close()
	}
}

type errorList struct {
	errs []error
}

func (e *errorList) tryPush(err error) {
	if err != nil {
		if e.errs == nil {
			e.errs = make([]error, 0, 2)
		}
		e.errs = append(e.errs, err)
	}
}

func (e errorList) Err() error {
	if e.errs == nil {
		return nil
	}
	b := new(bytes.Buffer)
	for _, err := range e.errs {
		b.WriteString(err.Error())
		b.WriteByte(';')
	}
	return errors.New(b.String())
}
