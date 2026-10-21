# Logging

Final provides a Session.Logger object, and helper words, that provide the application logging interface. Logging is based on levels of severity, with the logger operating at a given level. Logs targeted for the active Level or lower get logged, other logs get dropped. A conditional word to test for each level is provided, to enable conditional preparation of logging messages that might entail formatting costs.

The underlying logging mechanism is supplied by a LogProvider object. This can be supplied by an embedder, during Session creation, or in Final, on Session initialization. Final provides 2 types of default LogProviders: a simple File-based logger, based on a TextWriter, and a StdOut or StdErrr logger.

### Logging Levels

The following Levels are provided, and must be supported by any LogProvider.

```
Levels:
0 Off
1 Fatal
2 Error
3 Warn
4 Info
5 Debug
6 Trace
```



## Session.Logger

Session holds an object of type Logger via sess.logger that is the logger shared by all Pips in the Session. It has these principle members (use `help Logger` for full description):

```
Logger
  Int level readonly const property
  String service property
  String instanceId property
  
  // ( msg:s data:a level:i -- )
  log
```

The `service`, optional `instanceId` and `level` are typically set once during startup. Here are the logging levels

## Final Logging Words

In your final code, you will typically use the following words to log. These abstract all the details, and enable concise logging logs in your code.

Even if you send a log, it will only get written to the underlying Provider if the level of the log is less than or equal to the current Logger.level, and logging is enabled (level <> 0). It is typical for an application to operate at Info level, or at a minimum, Warn. Conditional test words for all log levels are provided, that can be used to put your log statment in an `if`/`then` if preparation of the message or data will entail a significant runtime cost. For most eg Info or below messages, with a simple unformatted string message, and no data, the extra cost of the test is usually not worth the complexity.

All the words below delegate to a single word `logAt` which also includes the level at which to log. You will typically never use `logAt`, and will use one of these, or the test then then the log word:

```
Words:
// logs, only if level <= Logger.level, and level <> 0 
// ( msg:s data:a level:i -- )
logAt

// ( -- b ) ( msg:s data:a -- )
logFatal?   logFatal
logError?   logError
logWarn?    logWarn
logInfo?    logInfo
logDebug?   logDebug
logTrace?   logTrace
```

Typical logging scenarios, simple, and with JSON data:

```forth
// simple log, no data
"Something simple happened" "" logInfo

// providing an existing object as data context
"Something else happened" myObj >json logInfo

// a log, with constructed JSON data
"Something happened" { 1 "param1" 2 "param2" } >json logInfo
```
Note: use the compact `>json` operator, not the pretty-print one `>jsonp` to avoid spurious newlines and unnecessary whitespace.

Conditional logging:

```forth
// costly/elaborate Trace message formatting
logTrace? if makeTraceMessage makeTraceData logTrace then
```



## Logger Initialization



### In Final

Initializing logging takes 2 simple steps:

1. Obtain a LogProvider, from a LogProvider factory method;
2. Call logger.init with your service/instanceId/level and the provider.

See the example code examples/app-svcs/logging.fnl for details.

### By an Embedder

Implement a LogProvider that conforms to the following exported interface.

```go
type LogProvider interface {
    Log(service, instanceId string, level int, msg string, data any)
}
```

The Log method will only be called if the LogProvider is enabled, and the `level` is at or below the current LogProvider level.
The LogProvider must be threadsafe, and must serialize logs.
It should not block in normal operation; blocking under unusual conditions of backpressure by the underlying log writing mechanism is acceptable. Logs should not be dropped.

The `data` parameter will typically contain:

- an empty string; for database logging, this can be treated as NULL
- a string, which can be assumed to be valid, compact JSON
- a Go `nil` interface value; can be treated as NULL or empty

The LogProvider may apply additional interpretations of another type of value in data, for its own native log rendering. For example, it can check for map or [] and call json.Marshall on the value.
    