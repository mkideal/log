package adaptor

import (
	"github.com/mkideal/log"
)

// AdaptorLogger implements generic logger:
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
type AdaptorLogger byte

func (l AdaptorLogger) Trace(msg ...interface{}) {
	log.Print(3, log.LvTRACE, msg...)
}

func (l AdaptorLogger) Tracef(format string, args ...interface{}) {
	log.Printf(3, log.LvTRACE, format, args...)
}

func (l AdaptorLogger) Debug(msg ...interface{}) {
	log.Print(3, log.LvDEBUG, msg...)
}

func (l AdaptorLogger) Debugf(format string, args ...interface{}) {
	log.Printf(3, log.LvDEBUG, format, args...)
}

func (l AdaptorLogger) Info(msg ...interface{}) {
	log.Print(3, log.LvINFO, msg...)
}

func (l AdaptorLogger) Infof(format string, args ...interface{}) {
	log.Printf(3, log.LvINFO, format, args...)
}

func (l AdaptorLogger) Warn(msg ...interface{}) {
	log.Print(3, log.LvWARN, msg...)
}

func (l AdaptorLogger) Warnf(format string, args ...interface{}) {
	log.Printf(3, log.LvWARN, format, args...)
}

func (l AdaptorLogger) Error(msg ...interface{}) {
	log.Print(3, log.LvERROR, msg...)
}

func (l AdaptorLogger) Errorf(format string, args ...interface{}) {
	log.Printf(3, log.LvERROR, format, args...)
}

func (l AdaptorLogger) Fatal(msg ...interface{}) {
	log.Print(3, log.LvFATAL, msg...)
}

func (l AdaptorLogger) Fatalf(format string, args ...interface{}) {
	log.Printf(3, log.LvFATAL, format, args...)
}
