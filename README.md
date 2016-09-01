LOG
===
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/mkideal/log/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mkideal/log)](https://goreportcard.com/report/github.com/mkideal/log)
[![Travis branch](https://img.shields.io/travis/mkideal/log/master.svg)](https://travis-ci.org/mkideal/log)
[![Coverage Status](https://coveralls.io/repos/github/mkideal/log/badge.svg?branch=master)](https://coveralls.io/github/mkideal/log?branch=master)
[![GoDoc](https://godoc.org/github.com/mkideal/log?status.svg)](https://godoc.org/github.com/mkideal/log)

`log` package inspired from [golang/glog](https://github.com/golang/glog). We have following key features:

* **lightweight** - `log` package very lightweight, and so easy to use.
* **highly customizable** - You can customize `Provider`,`Logger`.
* **fast** - Write logs to a buffer queue.

## Getting started

```go
package main

import (
	"github.com/mkideal/log"
)

func main() {
	// Init and defer Uninit
	defer log.Uninit(log.InitFile("./log/app.log"))

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

	log.If(true).Info("%v should be printed", true)

	iq := 250
	log.If(iq < 250).Info("IQ less than 250").
		ElseIf(iq > 250).Info("IQ greater than 250").
		Else().Info("IQ equal to 250")

	log.With("hello").Info("With a string field")
	log.With(1).Info("With an int field")
	log.With(true).Info("With a bool field")
	log.With(1, "2", false).Info("With 3 fields")
	log.With(log.M{"a":1}).Info("With a map")
	log.WithJSON(log.M{"a":1}).Info("With a map and using JSONFormatter")

	// don't print message header
	log.NoHeader()

	log.Info("This message have no header")

	log.Fatal("%s should be printed and exit program with status code 1", "FATAL")

	log.Info("You cannot see me")
}
```

Run it:

```sh
go run main.go
```

Now, current directory should have a subdirectory `log`

	.
	├── log
	│   ├── app.log -> app.log.20160814
	│   └── app.log.20160814
	└── main.go

The log file is `./log/app.log.20160814`, and `./log/app.log` link to it.

NOTE: You can remove the line `defer log.Uninit(log.InitFile("./log/app.log"))`. In this case, the log package use standard log package

## Log level

There are 6 log levels: `Fatal`,`Error`,`Warn`,`Info`,`Debug`,`Trace`

Default log level is `Info` if log level isn't specified.

`Logger` define methods `GetLevel() Level` and `SetLevel(Level)`.

## Init/Uninit


You should call one of the following `Init` functions before writing logs:

* **InitWithLogger**(*logger.Logger*) - init log with a `Logger`
* **InitWithProvider**(*logger.Provider*) - init log with a `Provider`
* **Init**(*providerType, opts string*) - init log by specified provider type and opts
* **InitFile**(*fullpath string*) - wrap the `Init` for `file` provider
* **InitConsole**(*toStderrLevel logger.Level*) - wrap the `Init` for `consle` provider
* **InitFileAndConsole**(*fullpath string, toStderrLevel logger.Level*) - combine `file` and `console` providers
* **InitMultiFile**(*rootdir, filename string*) - wrap the `Init` for `multifile` provider
* **InitMultiFileAndConsole**(*rootdir, filename string*) - combine `multifile` and `console` providers

And call the `Uninit` before exit program.

Examples:

```go
// InitWithLogger
func main() {
	defer log.Uninit(log.InitWithLogger(logger))
	...
}
```

```go
// InitWithProvider
func main() {
	defer log.Uninit(log.InitWithProvider(provider))
	...
}
```

```go
// Init
func main() {
	defer log.Uninit(log.Init("file", `{"dir":"./log","filename":"app.log"}`))
	// OR
	// defer log.Uninit(log.Init("file", `{"dir":"./log","filename":"app.log","maxsize":67108864}`))
	// NOTE: 67108864 = 1 << 26 = 64M
}
```

```go
// InitFile
func main() {
	defer log.Uninit(log.InitFile("./log/app.log"))
	...
}
```

```go
// InitConsole
func main() {
	defer log.Uninit(log.InitConsole(log.LvWARN))
	...
}
```

```go
// InitFileAndConsole
func main() {
	defer log.Uninit(log.InitFileAndConsole("./log/app.log", log.LvERROR))
	...
}
```

```go
// InitMultiFile
func main() {
	defer log.Uninit(log.InitMultiFile("./log", "app.log"))
	...
}
```

```go
// InitMultiFileAndConsole
func main() {
	defer log.Uninit(log.InitMultiFileAndConsole("./log", "app.log", log.LvERROR))
	...
}
```

## Print functions

* **Fatal**(*format string, args ...interface{}*) - Print and exit program with code 1
* **Error**(*format string, args ...interface{}*) - Print if level >= ERROR
* **Warn**(*format string, args ...interface{}*) - Print if level >= WARN
* **Info**(*format string, args ...interface{}*) - Print if level >= INFO
* **Debug**(*format string, args ...interface{}*) - Print if level >= DEBUG
* **Trace**(*format string, args ...interface{}*) - Print if level >= TRACE

Examples:

```go
log.Trace("hello %s", "TRACE")
log.Debug("hello %s", "DEBUG")
log.Info("hello %s", "INFO")
log.Warn("hello %s", "WARN")
log.Error("hello %s", "ERROR")
log.Fatal("bye bye %s", "FATAL")
```

## Provider

`Provider` interface defined in `log/logger/provider.go`:

```
type Provider interface {
	Write(level Level, headerLength int, data []byte) error
	Close() error
}
```

You can implement your `Provider`, then use `InitWithProvider`(see example [provider](https://github.com/mkideal/log/tree/master/_examples/provider/main.go)).

Or register your provider first, and use `Init`(see example [register_provider](https://github.com/mkideal/log/tree/master/_examples/register_provider/main.go))

Here are 4 builtin providers: `console`,`file`,`multifile`,`mix`

### console

*Creator:*

```go
// opts should be a JSON string or empty
func NewConsole(opts string) logger.Provider
```

*Opts:*

```go
type ConsoleOpts struct {
	ToStderrLevel logger.Level `json:"tostderrlevel"` // level which write to stderr from
}
```

### file

*Creator:*

```go
// opts should be a JSON string or empty
func NewFile(opts string) logger.Provider
```

*Opts:*

```go
type FileOpts struct {
	Dir       string `json:"dir"`       // log directory(default: .)
	Filename  string `json:"filename"`  // log filename(default: <appName>.log)
	NoSymlink bool   `json:"nosymlink"` // doesn't create symlink to latest log file(default: false)
	MaxSize   int    `json:"maxsize"`   // max bytes number of every log file(default: 64M)
}
```

### multifile

*Creator:*

```go
// opts should be a JSON string or empty
func NewMultiFile(opts string) logger.Provider
```

*Opts:*

```go
type MultiFileOpts struct {
	RootDir   string `json:"rootdir"`   // log directory(default: .)
	ErrorDir  string `json:"errordir"`  // error subdirectory(default: error)
	WarnDir   string `json:"warndir"`   // warn subdirectory(default: warn)
	InfoDir   string `json:"infodir"`   // info subdirectory(default: info)
	DebugDir  string `json:"debugdir"`  // debug subdirectory(default: debug)
	TraceDir  string `json:"tracedir"`  // trace subdirectory(default: trace)
	Filename  string `json:"filename"`  // log filename(default: <appName>.log)
	NoSymlink bool   `json:"nosymlink"` // doesn't create symlink to latest log file(default: false)
	MaxSize   int    `json:"maxsize"`   // max bytes number of every log file(default: 64M)
}
```

### mix

*Creator:*

```go
func NewMixProvider(first logger.Provider, others ...logger.Provider) logger.Provider
```

## Logger

`Logger` interface defined in log/logger/logger.go:

```go
// Logger is the top-level object of log package
type Logger interface {
	Run()
	Quit()
	NoHeader()
	GetLevel() Level
	SetLevel(level Level)
	Trace(calldepth int, format string, args ...interface{})
	Debug(calldepth int, format string, args ...interface{})
	Info(calldepth int, format string, args ...interface{})
	Warn(calldepth int, format string, args ...interface{})
	Error(calldepth int, format string, args ...interface{})
	Fatal(calldepth int, format string, args ...interface{})
}

// Entry represents a logging entry
type Entry interface {
	Level() Level
	Timestamp() int64
	Body() []byte
	Desc() []byte
	Clone() Entry
}

// Handler handle the logging entry
type Handler interface {
	Handle(entry Entry) error
}

// HookableLogger is a logger which can hook handlers
type HookableLogger interface {
	Logger
	Hook(Handler)
}
```

You can use your `Logger` by calling `InitWithLogger`.

## If(ElseIf,Else)

`log.If(bool)` returns a bool value which type is `IfLogger`. `IfLogger` implements `Trace`,`Debug`,`Info`,`Warn`,`Error`,`Fatal` methods, and all methods return the IfLogger.

`IfLogger` has methods `ElseIf` and `Else`.

```go
func (il IfLogger) Else() IfLogger          { return !il }
func (il IfLogger) ElseIf(ok bool) IfLogger { return IfLogger(ok) }
```

Here is an example demonstrates how to use `IfLogger`:

```go
iq := 250
log.If(iq < 250).Info("IQ less than 250").
	ElseIf(iq > 250).Info("IQ greater than 250").
	Else().Info("IQ equal to 250")
```

## With structured fields

```go
func With(objs ...interface{}) ContextLogger
func WithJSON(objs ...interface{}) ContextLogger
```

`ContextLogger` defined as following:

```go
type Context interface {
	With(values ...interface{}) ContextLogger
	WithJSON(values ...interface{}) ContextLogger
	SetFormatter(f Formatter) ContextLogger
}

type ContextLogger interface {
	Context
	Trace(format string, args ...interface{}) ContextLogger
	Debug(format string, args ...interface{}) ContextLogger
	Info(format string, args ...interface{}) ContextLogger
	Warn(format string, args ...interface{}) ContextLogger
	Error(format string, args ...interface{}) ContextLogger
	Fatal(format string, args ...interface{}) ContextLogger
}
```

`Formatter` is an interface that used to format data in ContextLogger

```go
type Formatter interface {
	Format(v interface{}) []byte
}
```

There are 2 builtin Formatter: nil(default), JSONFormatter
