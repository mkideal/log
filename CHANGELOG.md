Change Log
==========

## HEAD

* Add package-level function: `SetLevelFromString(s string) logger.Level`

## v0.1.0

* Add logging framework: `logger`,`provider`
* Add logger implementations: `logger.logger`, `logger.stdLogger`, `logger.testingLogger`
* Add provider implementations: `console`, `file`, `multifile`,`mix`,`level_filter`
* Add package-level function interface: `Trace/Debug/Info/Warn/Error/Fatal`,`If/With/WithJSON`,`Init*`,`Uninit`,`GetLevel/SetLevel/`
* Add `if_logger`, `context_logger`
