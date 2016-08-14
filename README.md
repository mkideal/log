LOG
===

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
	│   ├── app.log -> app.log.20160814-1449.028439.000
	│   └── app.log.20160814-1449.028439.000
	└── main.go

The log file is `./log/app.log.20160814-1449.028439.000`, and `./log/app.log` link to it.

* `20160814-1449` represents datetime `2016/08/14 14:49:00`
* `028439` is the pid
* `000` is a sequence number

NOTE: You can remove the line `defer log.Uninit(log.InitFile("./log/app.log"))`. In this case, the log package use standard log package

## Log level

There are 6 log levels: `Fatal`,`Error`,`Warn`,`Info`,`Debug`,`Trace`

Default log level is `Info` if log level isn't specified.

`Logger` define methods `GetLevel() Level` and `SetLevel(Level)`.

## Init/Uninit


You should call one of the following `Init` functions before writting logs:

* **InitWithLogger**(*logger.Logger*) - init log with a `Logger`
* **InitWithProvider**(*logger.Provider*) - init log with a `Provider`
* **Init**(*providerType, opts string*) - init log by specified provider type and opts
* **InitFile**(*fullpath string*) - wrap the `Init` for `file` provider
* **InitConsole**(*toStderrLevel logger.Level*) - wrap the `Init` for `consle` provider
* **InitFileAndConsole**(*fullpath string, toStderrLevel logger.Level*) - combine `file` and `console`

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

You can implements a your `Provider`, then use `InitWithProvider`(see example [provider](https://github.com/mkideal/log/tree/master/_examples/provider/main.go)).

Or register your provider first, and use `Init`(see example [register_provider](https://github.com/mkideal/log/tree/master/_examples/register_provider/main.go))

Here are 3 builtin providers: `console`,`file`,`mix`

### console

*Creator:*

```go
// opts should be a JSON string or empty
func NewFile(opts string) logger.Provider
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
func NewConsole(opts string) logger.Provider
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

### mix

*Creator:*

```go
func NewMixProvider(first logger.Provider, others ...logger.Provider) logger.Provider
```

## Logger

`Logger` interface defined in log/logger/logger.go:

```go
type Logger interface {
	Run()
	Quit()
	GetLevel() Level
	SetLevel(level Level)
	Trace(calldepth int, format string, args ...interface{})
	Debug(calldepth int, format string, args ...interface{})
	Info(calldepth int, format string, args ...interface{})
	Warn(calldepth int, format string, args ...interface{})
	Error(calldepth int, format string, args ...interface{})
	Fatal(calldepth int, format string, args ...interface{})
}
```

You can use your `Logger` by calling `InitWithLogger`.
