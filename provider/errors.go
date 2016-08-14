package provider

import (
	"bytes"
	"errors"
)

var (
	errWriterIsNil = errors.New("writer is nil")
	errOutOfRange  = errors.New("out of range")
)

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

func (e errorList) err() error {
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
