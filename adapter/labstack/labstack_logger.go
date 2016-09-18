package labstack

import (
	"io"

	glog "github.com/labstack/gommon/log"
	"github.com/mkideal/log"
	"github.com/mkideal/log/adapter"
	"github.com/mkideal/log/logger"
)

type LabstackLogger struct {
	adapter.GenericLogger
}

func (l LabstackLogger) SetOutput(w io.Writer) {
	// do nothing
}

func (l LabstackLogger) SetLevel(lv glog.Lvl) {
	level := logger.Level(glog.FATAL - lv)
	log.SetLevel(level)
}

func (l LabstackLogger) Print(msg ...interface{}) {
	log.Print(3, log.LvINFO, msg...)
}

func (l LabstackLogger) Printf(format string, args ...interface{}) {
	log.Printf(3, log.LvINFO, format, args...)
}

func (l LabstackLogger) Printj(m glog.JSON) {
	log.Print(3, log.LvINFO, log.M(m).JSON())
}

func (l LabstackLogger) Debugj(m glog.JSON) {
	if log.GetLevel() <= log.LvDEBUG {
		log.Print(3, log.LvDEBUG, log.M(m).JSON())
	}
}

func (l LabstackLogger) Infoj(m glog.JSON) {
	if log.GetLevel() <= log.LvINFO {
		log.Print(3, log.LvINFO, log.M(m).JSON())
	}
}

func (l LabstackLogger) Warnj(m glog.JSON) {
	if log.GetLevel() <= log.LvWARN {
		log.Print(3, log.LvWARN, log.M(m).JSON())
	}
}

func (l LabstackLogger) Errorj(m glog.JSON) {
	if log.GetLevel() <= log.LvERROR {
		log.Print(3, log.LvERROR, log.M(m).JSON())
	}
}

func (l LabstackLogger) Fatalj(m glog.JSON) {
	if log.GetLevel() <= log.LvFATAL {
		log.Print(3, log.LvFATAL, log.M(m).JSON())
	}
}
