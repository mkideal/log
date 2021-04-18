LOG
===

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/mkideal/log/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mkideal/log)](https://goreportcard.com/report/github.com/mkideal/log)
[![Travis branch](https://img.shields.io/travis/mkideal/log/master.svg)](https://travis-ci.org/mkideal/log)
[![GoDoc](https://godoc.org/github.com/mkideal/log?status.svg)](https://godoc.org/github.com/mkideal/log)

`log` package inspired from [golang/glog](https://github.com/golang/glog). We have following key features:

-	**dependentless** - No any dependencies.
-	**lightweight** - Lightweight and easy to use.
-	**highly customizable** - You can customize `Writer`,`Printer`.
-	**fast** - Write logs to a buffer queue.

Getting started
---------------

```go
package main

import (
	"github.com/mkideal/log"
)

func main() {
	// Start and Shutdown
	log.Start(/* options ... */)
	defer log.Shutdown()

	// Default log level is log.LvINFO, you can change the level as following:
	//
	//	log.SetLevel(log.LvTRACE)
	// 	log.SetLevel(log.LvDEBUG)
	// 	log.SetLevel(log.LvINFO)
	// 	log.SetLevel(log.LvWARN)
	// 	log.SetLevel(log.LvERROR)
	// 	log.SetLevel(log.LvFATAL)

	log.Trace("%s cannot be printed", "TRACE")
	log.Debug("%s cannot be printed", "DEBUG")

	log.Info("%s should be printed", "INFO")
	log.Warn("%s should be printed", "WARN")
	log.Error("%s should be printed", "ERROR")

	log.Fatal("%s should be printed and exit program with status code 1", "FATAL")
	log.Info("You cannot see me")
}
```

Log level
---------

There are 6 log levels: `Fatal`,`Error`,`Warn`,`Info`,`Debug`,`Trace`

Default log level is `Info` if log level isn't specified.

Start options
-------------

Examples:

```go
// WithConsole
func main() {
	log.Start(log.WithConsole())
	defer log.Shutdown()
	...
}
```

```go
// WithFile
func main() {
	log.Start(log.WithFile(log.FileOptions{
		Dir: "./log",
		Filename: "app.log",
	}))
	defer log.Shutdown()
	...
}
```

```go
// WithMultiFile
func main() {
	log.Start(log.WithMultiFile(log.MultiFileOptions{
		RootDir: "./log",
		Filename: "app.log",
	}))
	defer log.Shutdown()
	...
}
```

```go
// WithWriter

type coloredConsole struct {}

func (c coloredConsole) Write(level log.Level, data []byte, headerLen int) error {
	// ...
	return nil
}

func (c coloredConsole) Close() error{
	return nil
}

func main() {
	log.Start(log.WithWriter(coloredConsole{}))
	// multi-writers supported, e.g.
	//
	//	log.Start(log.WithWriter(coloredConsole{}), log.WithFile(...))
	defer log.Shutdown()
	...
}
```

```go
// WithPrinter


// printer implements log.Printer
type printer struct {}

func main() {
	log.WithPrinter(new(printer))
	defer log.Shutdown()
	...
}
```

```go
// WithHTTPHandler
func main() {
	log.Start(log.WithHTTPHandler(true))
	defer log.Shutdown()
	...
}
```

```go
// WithLevel
func main() {
	log.Start(log.WithLevel(log.LvWARN))
	defer log.Shutdown()
	...
}
```

```go
// WithPrefix
func main() {
	log.Start(log.WithPrefix("name"))
	defer log.Shutdown()
	...
}
```
Print functions
---------------

-	**Fatal**\(*format string, args ...interface{}*) - Print and exit program with code 1
-	**Error**\(*format string, args ...interface{}*) - Print if level >= ERROR
-	**Warn**\(*format string, args ...interface{}*) - Print if level >= WARN
-	**Info**\(*format string, args ...interface{}*) - Print if level >= INFO
-	**Debug**\(*format string, args ...interface{}*) - Print if level >= DEBUG
-	**Trace**\(*format string, args ...interface{}*) - Print if level >= TRACE
