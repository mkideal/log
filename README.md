LOG
===

`log` package inspired from `google/glog`. We have following key features:

* **lightweight** - `log` package very lightweight, and so easy to use.
* **highly customizable** - You can customize `Provider`, even `Logger`.
* **fast** - Write logs to a buffer queue.

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
* **InitConsole**() - wrap the `Init` for `consle` provider

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
	defer log.Uninit(log.InitConsole())
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
log.Fatak("bye bye %s", "FATAL")
```
