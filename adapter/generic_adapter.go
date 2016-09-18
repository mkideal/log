package adapter

import (
	"github.com/mkideal/log"
)

// GenericLogger ixmplements generic logger:
//
//	type Logger interface {
//		Trace(...interface{})
//		Tracef(string,...interface{})
//
//		Debug(...interface{})
//		Debugf(string,...interface{})
//
//		Info(...interface{})
//		Infof(string,...interface{})
//
//		Warn(...interface{})
//		Warnf(string,...interface{})
//
//		Error(...interface{})
//		Errorf(string,...interface{})
//
//		Fatal(...interface{})
//		Fatalf(string,...interface{})
//	}
type GenericLogger byte

func (l GenericLogger) Trace(msg ...interface{}) {
	log.Print(3, log.LvTRACE, msg...)
}

func (l GenericLogger) Tracef(format string, args ...interface{}) {
	log.Printf(3, log.LvTRACE, format, args...)
}

func (l GenericLogger) Debug(msg ...interface{}) {
	log.Print(3, log.LvDEBUG, msg...)
}

func (l GenericLogger) Debugf(format string, args ...interface{}) {
	log.Printf(3, log.LvDEBUG, format, args...)
}

func (l GenericLogger) Info(msg ...interface{}) {
	log.Print(3, log.LvINFO, msg...)
}

func (l GenericLogger) Infof(format string, args ...interface{}) {
	log.Printf(3, log.LvINFO, format, args...)
}

func (l GenericLogger) Warn(msg ...interface{}) {
	log.Print(3, log.LvWARN, msg...)
}

func (l GenericLogger) Warnf(format string, args ...interface{}) {
	log.Printf(3, log.LvWARN, format, args...)
}

func (l GenericLogger) Error(msg ...interface{}) {
	log.Print(3, log.LvERROR, msg...)
}

func (l GenericLogger) Errorf(format string, args ...interface{}) {
	log.Printf(3, log.LvERROR, format, args...)
}

func (l GenericLogger) Fatal(msg ...interface{}) {
	log.Print(3, log.LvFATAL, msg...)
}

func (l GenericLogger) Fatalf(format string, args ...interface{}) {
	log.Printf(3, log.LvFATAL, format, args...)
}
